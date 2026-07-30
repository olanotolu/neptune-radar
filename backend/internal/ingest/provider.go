package ingest

import (
	"context"
	"encoding/json"
	"os"
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
// otherwise Apify. Empty provider idles the watch loop.
func NewSocialProvider() SocialProvider {
	if key := os.Getenv("BRIGHTDATA_API_KEY"); key != "" {
		return NewBrightDataClient(key)
	}
	if tok := os.Getenv("APIFY_TOKEN"); tok != "" {
		return NewApifyClient(tok)
	}
	// Empty Apify client — Available() == false
	return NewApifyClient("")
}

// Ensure *ApifyClient and *BrightDataClient satisfy the interface.
var (
	_ SocialProvider = (*ApifyClient)(nil)
	_ SocialProvider = (*BrightDataClient)(nil)
)
