package ingest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// ApifyClient is the production data provider: it runs Apify actors
// (Instagram hashtag/post/profile scrapers) and returns their dataset items.
// Apify bills per result — every call site in the worker accounts for what
// it fetched via api_usage, and the daily cap is enforced before spending.
type ApifyClient struct {
	token  string
	client *http.Client
	// baseURL is overridable for tests; production always uses api.apify.com.
	baseURL string
	// Actor IDs are configurable because Apify actor names/versions change:
	HashtagActor string // hashtag monitor, default "apify/instagram-hashtag-scraper"
	ScraperActor string // vendor account posts, default "apify/instagram-scraper"
	ProfileActor string // profile/bio enrichment, default "apify/instagram-profile-scraper"
}

func NewApifyClient(token string) *ApifyClient {
	return &ApifyClient{
		token:        token,
		client:       &http.Client{Timeout: 120 * time.Second},
		baseURL:      "https://api.apify.com/v2/acts",
		HashtagActor: "apify/instagram-hashtag-scraper",
		ScraperActor: "apify/instagram-scraper",
		ProfileActor: "apify/instagram-profile-scraper",
	}
}

func (c *ApifyClient) Available() bool {
	// Master pause: token alone never enables spend.
	return c != nil && c.token != "" && apifyIngestEnabled()
}
func (c *ApifyClient) Name() string { return "apify" }

// runSync runs an actor with the given input and returns the raw dataset
// items, waiting for completion (run-sync-get-dataset-items). Transient
// failures (network errors, 429, 5xx) are retried with exponential backoff
// — a single Apify hiccup no longer drops an entire tick's worth of events.
func (c *ApifyClient) runSync(ctx context.Context, actorID string, input map[string]any) ([]json.RawMessage, error) {
	if c == nil || c.token == "" {
		return nil, fmt.Errorf("APIFY_TOKEN not set")
	}
	if !apifyIngestEnabled() {
		return nil, fmt.Errorf("APIFY_ENABLED=false — Apify paused, no API call")
	}
	body, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	endpoint := c.baseURL + "/" + url.PathEscape(actorID) + "/run-sync-get-dataset-items?token=" + url.QueryEscape(c.token)

	const maxRetries = 3
	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 2s, 4s, 8s — bounded so a down provider
			// doesn't stall the tick indefinitely.
			wait := time.Duration(1<<attempt) * time.Second
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(wait):
			}
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		req.Header.Set("content-type", "application/json")
		resp, err := c.client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("apify %s: %w", actorID, err)
			continue // network error — retry
		}
		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("apify %s: read body: %w", actorID, err)
			continue
		}
		// 429 (rate limit) and 5xx (server error) are retried; 4xx are not.
		if resp.StatusCode == 429 || resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("apify %s http %d: %s", actorID, resp.StatusCode, truncate(string(respBody), 300))
			continue
		}
		if resp.StatusCode >= 400 {
			return nil, fmt.Errorf("apify %s http %d: %s", actorID, resp.StatusCode, truncate(string(respBody), 300))
		}
		var items []json.RawMessage
		if err := json.Unmarshal(respBody, &items); err != nil {
			return nil, fmt.Errorf("apify %s: decode dataset: %w", actorID, err)
		}
		return items, nil
	}
	return nil, lastErr
}

// FetchHashtagPosts polls one batch of watched hashtags, most recent first.
// resultsLimit caps spend per run.
func (c *ApifyClient) FetchHashtagPosts(ctx context.Context, hashtags []string, resultsLimit int) ([]json.RawMessage, error) {
	return c.runSync(ctx, c.HashtagActor, map[string]any{
		"hashtags":     hashtags,
		"resultsLimit": resultsLimit,
		"resultsType":  "posts",
	})
}

// FetchAccountPosts polls recent posts for one or more watched accounts (vendors).
// Results from all handles are returned in one batch; each item's ownerUsername
// identifies which vendor it came from. resultsLimit caps the total across all
// handles.
func (c *ApifyClient) FetchAccountPosts(ctx context.Context, handles []string, resultsLimit int) ([]json.RawMessage, error) {
	return c.runSync(ctx, c.ScraperActor, map[string]any{
		"username":     handles,
		"resultsLimit": resultsLimit,
	})
}

// FetchProfile returns profile details (bio etc.) for one handle.
func (c *ApifyClient) FetchProfile(ctx context.Context, handle string) ([]json.RawMessage, error) {
	return c.runSync(ctx, c.ProfileActor, map[string]any{"usernames": []string{handle}})
}

// FetchFollowing returns the accounts a handle follows — used ONLY for lazy
// reciprocal-evidence enrichment on open hypotheses (it is one of the most
// expensive pulls per result, so the worker never runs it speculatively).
func (c *ApifyClient) FetchFollowing(ctx context.Context, handle string, resultsLimit int) ([]json.RawMessage, error) {
	return c.runSync(ctx, "apify/instagram-following-scraper", map[string]any{
		"usernames":    []string{handle},
		"resultsLimit": resultsLimit,
	})
}

func truncate(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) > n {
		return s[:n] + "..."
	}
	return s
}
