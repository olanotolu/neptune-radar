package connectors

import (
	"context"
	"encoding/json"
	"time"
)

// InstagramProber is satisfied by Bright Data / Apify clients. Declared
// narrowly here (rather than importing internal/ingest) so this package stays
// dependency-light and the interface is trivially fakeable in tests.
type InstagramProber interface {
	Available() bool
	FetchProfile(ctx context.Context, handle string) ([]json.RawMessage, error)
}

// ProfileStats is the subset of Apify's instagram-profile-scraper response
// this app persists. Field names match the actor's real JSON keys
// (followersCount, followsCount, postsCount, ...) — verified by inspecting a
// real response, not guessed from documentation.
type ProfileStats struct {
	FollowerCount  int
	FollowingCount int
	PostCount      int
	FullName       string
	ProfilePicURL  string
	Verified       bool
}

type apifyProfileFields struct {
	// Apify returns a 200 with a real dataset item even when the actor
	// couldn't read the profile — the item is an error payload
	// ({"error":"not_found","errorDescription":"..."}) instead of profile
	// fields. Checking len(items)>0 alone misses this and reports success
	// for a real failure — Error must be checked before trusting the rest.
	Error            string `json:"error"`
	ErrorDescription string `json:"errorDescription"`
	FollowersCount   int    `json:"followersCount"`
	FollowsCount     int    `json:"followsCount"`
	PostsCount       int    `json:"postsCount"`
	FullName         string `json:"fullName"`
	ProfilePicURL    string `json:"profilePicUrl"`
	Verified         bool   `json:"verified"`
}

// CheckInstagramHandle performs one real provider profile fetch as the health
// probe for a social connector — this is the same client the live worker
// uses, not a separate mock. If no provider token is configured, it honestly
// reports failure rather than pretending a check ran. On success, the real
// follower/following/post counts from that same fetch are returned too, so
// callers don't need a second API call to get them.
func CheckInstagramHandle(ctx context.Context, client InstagramProber, handle string) (CheckResult, *ProfileStats) {
	if client == nil || !client.Available() {
		return CheckResult{Status: "failure", ErrorMessage: "social provider not configured — no real check has run"}, nil
	}
	start := time.Now()
	items, err := client.FetchProfile(ctx, handle)
	elapsed := time.Since(start)
	if err != nil {
		return CheckResult{Status: "failure", ResponseTimeMs: int(elapsed.Milliseconds()), ErrorMessage: err.Error()}, nil
	}
	if len(items) == 0 {
		return CheckResult{Status: "failure", ResponseTimeMs: int(elapsed.Milliseconds()), ErrorMessage: "profile fetch returned no data"}, nil
	}
	var fields apifyProfileFields
	if err := json.Unmarshal(items[0], &fields); err != nil {
		return CheckResult{Status: "failure", ResponseTimeMs: int(elapsed.Milliseconds()), ErrorMessage: "unparseable profile response: " + err.Error()}, nil
	}
	if fields.Error != "" {
		msg := fields.Error
		if fields.ErrorDescription != "" {
			msg = fields.Error + ": " + fields.ErrorDescription
		}
		return CheckResult{Status: "failure", ResponseTimeMs: int(elapsed.Milliseconds()), ErrorMessage: msg}, nil
	}
	result := CheckResult{Status: "success", ResponseTimeMs: int(elapsed.Milliseconds())}
	return result, &ProfileStats{
		FollowerCount: fields.FollowersCount, FollowingCount: fields.FollowsCount, PostCount: fields.PostsCount,
		FullName: fields.FullName, ProfilePicURL: fields.ProfilePicURL, Verified: fields.Verified,
	}
}
