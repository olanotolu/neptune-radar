package ingest

import (
	"context"
	"encoding/json"
	"os"
	"strings"
)

// SocialProvider is the Instagram data plane: hashtag posts, account posts,
// profiles, and (optionally) following lists. Bright Data is primary;
// Apify remains a drop-in fallback when only APIFY_TOKEN is set.
type SocialProvider interface {
	Available() bool
	Name() string // "brightdata" | "apify" — used in usage accounting + status
	FetchHashtagPosts(ctx context.Context, hashtags []string, resultsLimit int) ([]json.RawMessage, error)
	FetchAccountPosts(ctx context.Context, handles []string, resultsLimit int) ([]json.RawMessage, error)
	FetchProfile(ctx context.Context, handle string) ([]json.RawMessage, error)
	FetchFollowing(ctx context.Context, handle string, resultsLimit int) ([]json.RawMessage, error)
}

// NewSocialProvider picks Bright Data when BRIGHTDATA_API_KEY is set,
// otherwise Apify only when APIFY_ENABLED=true (cost pause). Empty = idle watch loop.
func NewSocialProvider() SocialProvider {
	if key := os.Getenv("BRIGHTDATA_API_KEY"); key != "" {
		return NewBrightDataClient(key)
	}
	// Master pause: never fall back to Apify Instagram actors while APIFY_ENABLED is false.
	if !apifyIngestEnabled() {
		return NewApifyClient("") // Available() == false → watch loop idles on Apify path
	}
	if tok := os.Getenv("APIFY_TOKEN"); tok != "" {
		return NewApifyClient(tok)
	}
	return NewApifyClient("")
}

// apifyIngestEnabled requires APIFY_ENABLED=true (same master switch as TPS).
func apifyIngestEnabled() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("APIFY_ENABLED")))
	return v == "1" || v == "true" || v == "yes" || v == "on"
}

// Ensure *ApifyClient and *BrightDataClient satisfy the interface.
var (
	_ SocialProvider = (*ApifyClient)(nil)
	_ SocialProvider = (*BrightDataClient)(nil)
)
