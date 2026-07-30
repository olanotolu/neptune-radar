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

// Bright Data Instagram dataset IDs (Web Scraper API library).
const (
	bdDatasetProfiles = "gd_l1vikfch901nx3by4" // Instagram - Profiles (includes recent posts)
	bdDatasetPosts    = "gd_lk5ns7kz21pck8jpis" // Instagram - Posts (by post URL, full tags)
)

// BrightDataClient implements SocialProvider against Bright Data's
// datasets/v3 trigger + progress + snapshot API.
type BrightDataClient struct {
	token  string
	client *http.Client
}

func NewBrightDataClient(token string) *BrightDataClient {
	return &BrightDataClient{
		token: strings.TrimSpace(token),
		client: &http.Client{
			Timeout: 180 * time.Second,
		},
	}
}

func (c *BrightDataClient) Available() bool { return c != nil && c.token != "" }
func (c *BrightDataClient) Name() string    { return "brightdata" }

// FetchProfile scrapes one Instagram profile and returns a single item shaped
// for ParseProfile (Apify-compatible field names).
func (c *BrightDataClient) FetchProfile(ctx context.Context, handle string) ([]json.RawMessage, error) {
	handle = strings.TrimPrefix(strings.TrimSpace(handle), "@")
	if handle == "" {
		return nil, fmt.Errorf("brightdata: empty handle")
	}
	items, err := c.collect(ctx, bdDatasetProfiles, []map[string]any{
		{"url": "https://www.instagram.com/" + handle + "/"},
	})
	if err != nil {
		return nil, err
	}
	out := make([]json.RawMessage, 0, len(items))
	for _, raw := range items {
		norm, err := normalizeBDProfile(raw)
		if err != nil {
			continue
		}
		out = append(out, norm)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("brightdata: no profile for @%s", handle)
	}
	return out, nil
}

// FetchAccountPosts uses the Profiles scraper (embedded posts) then optionally
// deepens a subset via the Posts scraper for tagged users.
func (c *BrightDataClient) FetchAccountPosts(ctx context.Context, handles []string, resultsLimit int) ([]json.RawMessage, error) {
	if resultsLimit <= 0 {
		resultsLimit = 20
	}
	var all []json.RawMessage
	perHandle := resultsLimit
	if len(handles) > 1 {
		perHandle = max(3, resultsLimit/len(handles))
	}
	for _, handle := range handles {
		handle = strings.TrimPrefix(strings.TrimSpace(handle), "@")
		if handle == "" {
			continue
		}
		items, err := c.collect(ctx, bdDatasetProfiles, []map[string]any{
			{"url": "https://www.instagram.com/" + handle + "/"},
		})
		if err != nil {
			return all, fmt.Errorf("brightdata posts @%s: %w", handle, err)
		}
		posts := extractBDProfilePosts(items, handle, perHandle)
		// Deepen first few posts for tagged_users when budget allows
		deepN := min(len(posts), 5)
		if deepN > 0 {
			var inputs []map[string]any
			urls := make([]string, 0, deepN)
			for i := 0; i < deepN; i++ {
				var p struct {
					URL string `json:"url"`
				}
				_ = json.Unmarshal(posts[i], &p)
				if p.URL != "" {
					inputs = append(inputs, map[string]any{"url": ensureTrailingSlashIG(p.URL)})
					urls = append(urls, p.URL)
				}
			}
			if len(inputs) > 0 {
				if deep, err := c.collect(ctx, bdDatasetPosts, inputs); err == nil {
					// Prefer deep posts when available
					byURL := map[string]json.RawMessage{}
					for _, d := range deep {
						norm, err := normalizeBDPost(d, handle)
						if err != nil {
							continue
						}
						var meta struct {
							URL string `json:"url"`
						}
						_ = json.Unmarshal(norm, &meta)
						if meta.URL != "" {
							byURL[strings.TrimRight(meta.URL, "/")] = norm
						}
					}
					for i, p := range posts {
						var meta struct {
							URL string `json:"url"`
						}
						_ = json.Unmarshal(p, &meta)
						key := strings.TrimRight(meta.URL, "/")
						if n, ok := byURL[key]; ok {
							posts[i] = n
						}
					}
				}
			}
		}
		all = append(all, posts...)
		if len(all) >= resultsLimit {
			return all[:resultsLimit], nil
		}
	}
	return all, nil
}

// FetchHashtagPosts uses watched vendor-style discovery: Bright Data's free
// hashtag "collection" datasets are often disabled; we approximate by
// scraping profile posts that mention the hashtag is not possible without a
// hashtag dataset. When the hashtag library supports collection we use
// keyword search; otherwise return empty with a clear error so the worker
// logs and continues vendor monitors.
func (c *BrightDataClient) FetchHashtagPosts(ctx context.Context, hashtags []string, resultsLimit int) ([]json.RawMessage, error) {
	// Instagram hashtag dataset (gd_lp9xgqvwrt2wc2jjy) currently returns
	// "does not support collection" on many accounts — try once with first tag
	// as a soft signal; on failure return empty (not hard error) so vendors still run.
	if len(hashtags) == 0 {
		return nil, nil
	}
	tag := strings.TrimPrefix(hashtags[0], "#")
	// Prefer Posts dataset only works with post URLs — skip hashtag path when
	// collection is unsupported. Worker will still poll vendors.
	_ = tag
	_ = resultsLimit
	_ = ctx
	return nil, fmt.Errorf("brightdata: hashtag collection not enabled on this account — using vendor/profile monitors only")
}

// FetchFollowing is not exposed as a cheap Bright Data dataset on free tiers.
// Reciprocal-follow evidence degrades gracefully (no +10) without it.
func (c *BrightDataClient) FetchFollowing(ctx context.Context, handle string, resultsLimit int) ([]json.RawMessage, error) {
	_ = ctx
	_ = handle
	_ = resultsLimit
	return nil, fmt.Errorf("brightdata: following list not configured")
}

// --- HTTP + polling ----------------------------------------------------------

func (c *BrightDataClient) collect(ctx context.Context, datasetID string, input []map[string]any) ([]json.RawMessage, error) {
	if !c.Available() {
		return nil, fmt.Errorf("BRIGHTDATA_API_KEY not set")
	}
	// Prefer sync scrape for small inputs; fall back to trigger+poll on 202.
	body, err := json.Marshal(map[string]any{"input": input})
	if err != nil {
		return nil, err
	}
	u := "https://api.brightdata.com/datasets/v3/scrape?dataset_id=" + url.QueryEscape(datasetID) + "&format=json"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("brightdata scrape: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == 200 {
		return decodeBDItems(raw)
	}
	if resp.StatusCode == 202 {
		var ack struct {
			SnapshotID string `json:"snapshot_id"`
		}
		_ = json.Unmarshal(raw, &ack)
		if ack.SnapshotID == "" {
			return nil, fmt.Errorf("brightdata scrape 202 without snapshot_id: %s", truncate(string(raw), 200))
		}
		return c.pollSnapshot(ctx, ack.SnapshotID)
	}
	// Sync may be disabled — trigger async
	if resp.StatusCode == 400 && bytes.Contains(raw, []byte("not active")) {
		return nil, fmt.Errorf("brightdata: customer not active — enable scrapers in the control panel")
	}
	// Always try trigger path as fallback
	return c.triggerAndWait(ctx, datasetID, input)
}

func (c *BrightDataClient) triggerAndWait(ctx context.Context, datasetID string, input []map[string]any) ([]json.RawMessage, error) {
	body, err := json.Marshal(map[string]any{"input": input})
	if err != nil {
		return nil, err
	}
	u := "https://api.brightdata.com/datasets/v3/trigger?dataset_id=" + url.QueryEscape(datasetID) + "&format=json"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("brightdata trigger: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("brightdata trigger http %d: %s", resp.StatusCode, truncate(string(raw), 300))
	}
	var ack struct {
		SnapshotID string `json:"snapshot_id"`
		Error      string `json:"error"`
	}
	if err := json.Unmarshal(raw, &ack); err != nil {
		return nil, fmt.Errorf("brightdata trigger decode: %w (%s)", err, truncate(string(raw), 200))
	}
	if ack.Error != "" {
		return nil, fmt.Errorf("brightdata trigger: %s", ack.Error)
	}
	if ack.SnapshotID == "" {
		return nil, fmt.Errorf("brightdata trigger: no snapshot_id in %s", truncate(string(raw), 200))
	}
	return c.pollSnapshot(ctx, ack.SnapshotID)
}

func (c *BrightDataClient) pollSnapshot(ctx context.Context, snapshotID string) ([]json.RawMessage, error) {
	deadline := time.Now().Add(3 * time.Minute)
	for time.Now().Before(deadline) {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet,
			"https://api.brightdata.com/datasets/v3/progress/"+url.PathEscape(snapshotID), nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+c.token)
		resp, err := c.client.Do(req)
		if err != nil {
			return nil, err
		}
		raw, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		var prog struct {
			Status string `json:"status"`
			Error  string `json:"error"`
		}
		_ = json.Unmarshal(raw, &prog)
		switch prog.Status {
		case "ready":
			return c.downloadSnapshot(ctx, snapshotID)
		case "failed":
			return nil, fmt.Errorf("brightdata snapshot failed: %s", truncate(string(raw), 250))
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(3 * time.Second):
		}
	}
	return nil, fmt.Errorf("brightdata: timeout waiting for snapshot %s", snapshotID)
}

func (c *BrightDataClient) downloadSnapshot(ctx context.Context, snapshotID string) ([]json.RawMessage, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		"https://api.brightdata.com/datasets/v3/snapshot/"+url.PathEscape(snapshotID)+"?format=json", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("brightdata snapshot http %d: %s", resp.StatusCode, truncate(string(raw), 250))
	}
	return decodeBDItems(raw)
}

func decodeBDItems(raw []byte) ([]json.RawMessage, error) {
	// Array of objects
	var arr []json.RawMessage
	if err := json.Unmarshal(raw, &arr); err == nil {
		return arr, nil
	}
	// Single object
	var obj map[string]any
	if err := json.Unmarshal(raw, &obj); err == nil {
		// status wrapper when not ready
		if _, ok := obj["status"]; ok && obj["message"] != nil {
			return nil, fmt.Errorf("brightdata: %v", obj["message"])
		}
		b, _ := json.Marshal(obj)
		return []json.RawMessage{b}, nil
	}
	return nil, fmt.Errorf("brightdata: unexpected payload %s", truncate(string(raw), 200))
}

// --- Normalizers → Apify-shaped JSON for MapPost / ParseProfile --------------

func normalizeBDProfile(raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		Account      string `json:"account"`
		Biography    string `json:"biography"`
		Followers    int    `json:"followers"`
		Following    int    `json:"following"`
		PostsCount   int    `json:"posts_count"`
		IsVerified   bool   `json:"is_verified"`
		ProfileImage string `json:"profile_image_link"`
		ProfileURL   string `json:"profile_url"`
		// alternate keys
		FullName string `json:"full_name"`
		// some responses nest
		ProfileName string `json:"profile_name"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, err
	}
	// Also try to pull display name from generic map
	var m map[string]any
	_ = json.Unmarshal(raw, &m)
	fullName := firstString(m, "full_name", "profile_name", "name")
	if fullName == "" {
		fullName = p.FullName
	}
	if fullName == "" {
		fullName = p.ProfileName
	}
	pic := firstString(m, "profile_image_link", "profile_pic_url", "profilePicUrl", "avatar")
	if pic == "" {
		pic = p.ProfileImage
	}
	out := map[string]any{
		"biography":      p.Biography,
		"bio":            p.Biography,
		"fullName":       fullName,
		"profilePicUrl":  pic,
		"followersCount": p.Followers,
		"followsCount":   p.Following,
		"postsCount":     p.PostsCount,
		"verified":       p.IsVerified,
		"username":       p.Account,
	}
	// Pass through embedded posts for account-post extraction
	if posts, ok := m["posts"]; ok {
		out["_bd_posts"] = posts
		out["_bd_account"] = p.Account
	}
	b, err := json.Marshal(out)
	return b, err
}

func extractBDProfilePosts(profileItems []json.RawMessage, handle string, limit int) []json.RawMessage {
	var out []json.RawMessage
	for _, raw := range profileItems {
		var wrap struct {
			Posts []struct {
				Caption      string   `json:"caption"`
				ID           string   `json:"id"`
				URL          string   `json:"url"`
				PostHashtags []string `json:"post_hashtags"`
				ContentType  string   `json:"content_type"`
				VideoURL     string   `json:"video_url"`
				Datetime     string   `json:"datetime"`
				// image may be nested
			} `json:"posts"`
			Account string `json:"account"`
		}
		// After normalizeBDProfile, posts are under _bd_posts
		var m map[string]any
		if json.Unmarshal(raw, &m) == nil {
			if posts, ok := m["_bd_posts"]; ok {
				b, _ := json.Marshal(map[string]any{"posts": posts, "account": m["_bd_account"]})
				_ = json.Unmarshal(b, &wrap)
			} else {
				_ = json.Unmarshal(raw, &wrap)
			}
		}
		owner := wrap.Account
		if owner == "" {
			owner = handle
		}
		for _, p := range wrap.Posts {
			if p.URL == "" && p.ID == "" {
				continue
			}
			item := map[string]any{
				"id":            p.ID,
				"url":           p.URL,
				"caption":       p.Caption,
				"ownerUsername": owner,
				"hashtags":      p.PostHashtags,
				"timestamp":     p.Datetime,
			}
			if p.VideoURL != "" {
				item["displayUrl"] = p.VideoURL
			}
			// Mentions from caption are extracted by signals package
			b, _ := json.Marshal(item)
			out = append(out, b)
			if limit > 0 && len(out) >= limit {
				return out
			}
		}
	}
	return out
}

func normalizeBDPost(raw json.RawMessage, fallbackOwner string) (json.RawMessage, error) {
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, err
	}
	// Skip error-only rows (input echo without content)
	if _, hasInput := m["input"]; hasInput && m["url"] == nil && m["description"] == nil && m["caption"] == nil {
		return nil, fmt.Errorf("empty post row")
	}
	caption := firstString(m, "description", "caption", "post_content")
	owner := firstString(m, "user_posted", "user_name", "username", "ownerUsername", "account")
	if owner == "" {
		owner = fallbackOwner
	}
	postURL := firstString(m, "url", "post_url")
	id := firstString(m, "post_id", "pk", "id", "shortcode", "content_id", "short_code")
	img := firstString(m, "thumbnail", "display_url", "displayUrl", "image_url", "video_url")
	// photos: []string or []{url}
	if img == "" {
		if arr, ok := m["photos"].([]any); ok && len(arr) > 0 {
			switch p0 := arr[0].(type) {
			case string:
				img = p0
			case map[string]any:
				if s, ok := p0["url"].(string); ok {
					img = s
				}
			}
		}
	}
	if img == "" {
		if arr, ok := m["images"].([]any); ok && len(arr) > 0 {
			if p0, ok := arr[0].(map[string]any); ok {
				if s, ok := p0["url"].(string); ok {
					img = s
				}
			}
		}
	}
	if img == "" {
		if arr, ok := m["post_content"].([]any); ok && len(arr) > 0 {
			if p0, ok := arr[0].(map[string]any); ok {
				if s, ok := p0["url"].(string); ok {
					img = s
				}
			}
		}
	}
	hashtags := anyToStrings(m["hashtags"])
	if len(hashtags) == 0 {
		hashtags = anyToStrings(m["post_hashtags"])
	}
	// tagged_users: [{username, full_name, profile_url, ...}]
	tags := bdTaggedUsernames(m["tagged_users"])
	if len(tags) == 0 {
		tags = bdTaggedUsernames(m["tagged_user"])
	}
	// coauthor_producers: ["handle"] or objects
	if collabs := anyToUsernames(m["coauthor_producers"]); len(collabs) > 0 {
		outCollab := collabs[0]
		// store as collab for scorer
		_ = outCollab
	}
	ts := firstString(m, "date_posted", "datetime", "timestamp", "date")
	out := map[string]any{
		"id":            id,
		"shortCode":     firstString(m, "shortcode", "content_id"),
		"url":           postURL,
		"caption":       caption,
		"ownerUsername": owner,
		"hashtags":      hashtags,
		"displayUrl":    img,
		"timestamp":     ts,
	}
	if len(tags) > 0 {
		var tu []map[string]any
		for _, t := range tags {
			tu = append(tu, map[string]any{"username": t})
		}
		out["taggedUsers"] = tu
	}
	if collabs := anyToUsernames(m["coauthor_producers"]); len(collabs) > 0 && collabs[0] != owner {
		out["collabWith"] = collabs[0]
	}
	loc := firstString(m, "location", "post_location", "locationName")
	if loc != "" {
		out["location"] = loc
	}
	b, err := json.Marshal(out)
	return b, err
}

// bdTaggedUsernames extracts Instagram handles from Bright Data tagged_users.
func bdTaggedUsernames(v any) []string {
	arr, ok := v.([]any)
	if !ok {
		return anyToUsernames(v)
	}
	var out []string
	for _, x := range arr {
		switch t := x.(type) {
		case string:
			t = strings.TrimPrefix(strings.TrimSpace(t), "@")
			if t != "" {
				out = append(out, t)
			}
		case map[string]any:
			u := firstString(t, "username", "user_name", "account", "handle")
			if u == "" {
				// profile_url: https://www.instagram.com/foo
				if purl := firstString(t, "profile_url", "url"); purl != "" {
					u = handleFromIGURL(purl)
				}
			}
			u = strings.TrimPrefix(strings.TrimSpace(u), "@")
			if u != "" {
				out = append(out, u)
			}
		}
	}
	return out
}

func handleFromIGURL(u string) string {
	u = strings.TrimSpace(u)
	u = strings.TrimSuffix(u, "/")
	if i := strings.LastIndex(u, "/"); i >= 0 && i+1 < len(u) {
		return strings.TrimPrefix(u[i+1:], "@")
	}
	return ""
}

func firstString(m map[string]any, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			switch t := v.(type) {
			case string:
				if strings.TrimSpace(t) != "" {
					return t
				}
			case float64:
				return fmt.Sprintf("%.0f", t)
			case json.Number:
				return t.String()
			}
		}
	}
	return ""
}

func anyToStrings(v any) []string {
	switch t := v.(type) {
	case []string:
		return t
	case []any:
		var out []string
		for _, x := range t {
			if s, ok := x.(string); ok && s != "" {
				out = append(out, strings.TrimPrefix(s, "#"))
			}
		}
		return out
	default:
		return nil
	}
}

func anyToUsernames(v any) []string {
	switch t := v.(type) {
	case []string:
		var out []string
		for _, s := range t {
			s = strings.TrimPrefix(strings.TrimSpace(s), "@")
			if s != "" {
				out = append(out, s)
			}
		}
		return out
	case []any:
		var out []string
		for _, x := range t {
			switch u := x.(type) {
			case string:
				u = strings.TrimPrefix(strings.TrimSpace(u), "@")
				if u != "" {
					out = append(out, u)
				}
			case map[string]any:
				for _, k := range []string{"username", "user_name", "account", "handle"} {
					if s, ok := u[k].(string); ok && s != "" {
						out = append(out, strings.TrimPrefix(s, "@"))
						break
					}
				}
			}
		}
		return out
	default:
		return nil
	}
}

func ensureTrailingSlashIG(u string) string {
	if u == "" {
		return u
	}
	if !strings.HasSuffix(u, "/") {
		return u + "/"
	}
	return u
}
