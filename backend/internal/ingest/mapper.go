package ingest

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"neptune-social-radar/backend/internal/pipeline/watchtower"
)

// apifyPost is the defensive superset of what Apify's Instagram scrapers
// return for one post. Actor schemas drift between versions, so every field
// the pipeline cares about has a fallback parsed in normalizePost below.
type apifyPost struct {
	ID           string          `json:"id"`
	ShortCode    string          `json:"shortCode"`
	URL          string          `json:"url"`
	Caption      string          `json:"caption"`
	Owner        string          `json:"ownerUsername"`
	OwnerFull    string          `json:"ownerFullName"`
	Hashtags     json.RawMessage `json:"hashtags"`    // []string or [{"hashtag": ...}]
	Mentions     json.RawMessage `json:"mentions"`    // []string or [{"username": ...}]
	TaggedUsers  json.RawMessage `json:"taggedUsers"` // [{"username": ...}]
	LocationName string          `json:"locationName"`
	Location     string          `json:"location"`
	Timestamp    string          `json:"timestamp"`        // ISO 8601
	TakenAtUnix  json.RawMessage `json:"takenAtTimestamp"` // unix seconds (number or string)
	DisplayURL   string          `json:"displayUrl"`
	Images       json.RawMessage `json:"images"` // [{"url": ...}] or []string
	Type         string          `json:"type"`
	// Collaboration posts: Apify actors have emitted the co-author under
	// several field names over time — try each in normalizePost.
	CoauthorProducers json.RawMessage `json:"coauthorProducers"` // [{"username": ...}]
	Coauthors         json.RawMessage `json:"coauthors"`
	CollabWith        string          `json:"collabWith"`
}

// MapPost converts one provider dataset item into a watchtower.RawEvent of type "post".
// monitor identifies the watch source that surfaced it; imageURL is returned
// separately so the worker can decide (after the cheap pre-filter) whether to
// spend a vision-classifier call on it.
func MapPost(item json.RawMessage, monitor string) (watchtower.RawEvent, string, error) {
	var p apifyPost
	if err := json.Unmarshal(item, &p); err != nil {
		return watchtower.RawEvent{}, "", fmt.Errorf("map post: %w", err)
	}
	if p.Caption == "" && p.ID == "" && p.ShortCode == "" {
		return watchtower.RawEvent{}, "", fmt.Errorf("map post: item has neither caption nor id — not a post?")
	}

	extID := firstNonEmpty(p.ID, p.ShortCode, p.URL)
	if extID == "" {
		return watchtower.RawEvent{}, "", fmt.Errorf("map post: no id, shortCode, or url to key on")
	}
	occurredAt := parseApifyTime(p.Timestamp, p.TakenAtUnix)

	payload := map[string]any{
		"caption": p.Caption,
		"url":     p.URL,
	}
	// Persist the post image so the Sources UI and prospect cards can show it
	// without a second provider fetch. Vision still receives imageURL separately.
	if img := firstNonEmpty(p.DisplayURL, firstImageURL(p.Images)); img != "" {
		payload["image_url"] = img
	}
	if tags := parseUserRefs(p.TaggedUsers, "username"); len(tags) > 0 {
		payload["tags"] = tags
	}
	if hashtags := parseUserRefs(p.Hashtags, "hashtag"); len(hashtags) > 0 {
		payload["hashtags"] = hashtags
	}
	if mentions := parseUserRefs(p.Mentions, "username"); len(mentions) > 0 {
		// Caption @mentions are extracted by the signals package from the
		// caption text itself; some items omit them from the caption string,
		// so merge any provider-parsed mentions in explicitly.
		payload["provider_mentions"] = mentions
	}
	if loc := firstNonEmpty(p.LocationName, p.Location); loc != "" {
		payload["location"] = loc
	}
	// A collaboration post is co-authored by both accounts — the strongest
	// single pair signal available (spec: co-authors). The signals package
	// reads payload["collab_with"]; Apify has shipped the co-author under
	// different field names, so try them all.
	if collab := extractCoauthor(p); collab != "" && collab != p.Owner {
		payload["collab_with"] = collab
	}
	if p.OwnerFull != "" {
		payload["author_display_name"] = p.OwnerFull
	}

	return watchtower.RawEvent{
		Monitor:         monitor,
		Source:          "apify",
		ExternalEventID: extID,
		Handle:          p.Owner,
		Type:            "post",
		Payload:         payload,
		OccurredAt:      occurredAt,
	}, p.DisplayURL, nil
}

// ParseProfileBio extracts the biography from a profile-scraper item.
func ParseProfileBio(item json.RawMessage) (bio string, ok bool) {
	prof, ok := ParseProfile(item)
	if !ok {
		return "", false
	}
	return prof.Bio, true
}

// ProfileDetails is the subset of an Instagram profile we store on
// social_accounts / watched_sources for prospect cards and geo.
type ProfileDetails struct {
	Bio            string
	DisplayName    string
	ProfilePicURL  string
	FollowerCount  *int
	FollowingCount *int
	PostCount      *int
	IsPrivate      bool
	Verified       bool
	// Location from profile address fields and/or bio inference.
	City   string
	Region string // state / region
	// LocationSource: "profile" | "bio" | ""
	LocationSource string
	// Street address from Instagram business profile (rare but gold when present).
	StreetAddress  string
	BusinessCity   string
	BusinessState  string
	BusinessPostal string
}

// ParseProfile extracts bio + avatar + stats + location from a profile-scraper item.
func ParseProfile(item json.RawMessage) (ProfileDetails, bool) {
	var p struct {
		Biography       string `json:"biography"`
		Bio             string `json:"bio"`
		FullName        string `json:"fullName"`
		ProfilePicURL   string `json:"profilePicUrl"`
		ProfilePicURLHD string `json:"profilePicUrlHD"`
		FollowersCount  int    `json:"followersCount"`
		FollowsCount    int    `json:"followsCount"`
		PostsCount      int    `json:"postsCount"`
		Private         bool   `json:"private"`
		IsPrivate       bool   `json:"isPrivate"`
		Verified        bool   `json:"verified"`
		// Location-ish fields Apify has used across actor versions
		AddressCity     string          `json:"addressCity"`
		CityName        string          `json:"cityName"`
		City            string          `json:"city"`
		BusinessAddress json.RawMessage `json:"businessAddress"`
		LocationName    string          `json:"locationName"`
		Error           string          `json:"error"`
	}
	if err := json.Unmarshal(item, &p); err != nil || p.Error != "" {
		return ProfileDetails{}, false
	}
	bio := firstNonEmpty(p.Biography, p.Bio)
	pic := firstNonEmpty(p.ProfilePicURLHD, p.ProfilePicURL)
	out := ProfileDetails{
		Bio:           bio,
		DisplayName:   p.FullName,
		ProfilePicURL: pic,
		IsPrivate:     p.Private || p.IsPrivate,
		Verified:      p.Verified,
	}
	if strings.Contains(string(item), "followersCount") {
		n := p.FollowersCount
		out.FollowerCount = &n
	}
	if strings.Contains(string(item), "followsCount") {
		n := p.FollowsCount
		out.FollowingCount = &n
	}
	if strings.Contains(string(item), "postsCount") {
		n := p.PostsCount
		out.PostCount = &n
	}

	// Explicit profile address fields first, then bio inference.
	city := firstNonEmpty(p.AddressCity, p.CityName, p.City, p.LocationName)
	region := ""
	if city == "" && len(p.BusinessAddress) > 0 {
		var ba struct {
			City          string `json:"city"`
			CityName      string `json:"city_name"`
			Region        string `json:"region"`
			RegionCode    string `json:"region_code"`
			State         string `json:"state"`
			StreetAddress string `json:"street_address"`
			PostalCode    string `json:"postal_code"`
			ZipCode       string `json:"zip_code"`
		}
		if json.Unmarshal(p.BusinessAddress, &ba) == nil {
			city = firstNonEmpty(ba.City, ba.CityName)
			region = firstNonEmpty(ba.RegionCode, ba.State, ba.Region)
			if ba.StreetAddress != "" {
				out.StreetAddress = ba.StreetAddress
				out.BusinessCity = firstNonEmpty(ba.City, ba.CityName)
				out.BusinessState = firstNonEmpty(ba.RegionCode, ba.State)
				out.BusinessPostal = firstNonEmpty(ba.PostalCode, ba.ZipCode)
			}
		}
	}
	if city != "" {
		out.City = city
		out.Region = region
		out.LocationSource = "profile"
	} else if bio != "" {
		// signals package is imported by worker, not mapper — do light inline
		// city,ST parse here to avoid an import cycle.
		if c, r, ok := parseCityState(bio); ok {
			out.City, out.Region = c, r
			out.LocationSource = "bio"
		}
	}
	return out, true
}

// parseCityState is a tiny subset of signals.InferLocationFromText for the
// mapper (which must not import signals → store cycles via other packages).
func parseCityState(text string) (city, region string, ok bool) {
	// "City, ST" pattern
	re := regexp.MustCompile(`\b([A-Z][a-z]+(?:\s[A-Z][a-z]+)?),\s*([A-Z]{2})\b`)
	if m := re.FindStringSubmatch(text); len(m) >= 3 {
		return m[1], m[2], true
	}
	lower := strings.ToLower(text)
	known := []struct{ needle, city, region string }{
		{"columbus", "Columbus", "OH"},
		{"cleveland", "Cleveland", "OH"},
		{"cincinnati", "Cincinnati", "OH"},
		{"brooklyn", "Brooklyn", "NY"},
		{"manhattan", "Manhattan", "NY"},
		{"new york", "New York", "NY"},
		{"nyc", "New York", "NY"},
		{"dumbo", "Brooklyn", "NY"},
		{"los angeles", "Los Angeles", "CA"},
		{"san francisco", "San Francisco", "CA"},
		{"bay area", "San Francisco", "CA"},
		{"malibu", "Malibu", "CA"},
		{"beverly hills", "Beverly Hills", "CA"},
		{"napa", "Napa", "CA"},
		{"chicago", "Chicago", "IL"},
		{"miami", "Miami", "FL"},
		{"austin", "Austin", "TX"},
	}
	for _, k := range known {
		if strings.Contains(lower, k.needle) {
			return k.city, k.region, true
		}
	}
	return "", "", false
}

func firstImageURL(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var asStrings []string
	if err := json.Unmarshal(raw, &asStrings); err == nil && len(asStrings) > 0 {
		return asStrings[0]
	}
	var asObjs []map[string]any
	if err := json.Unmarshal(raw, &asObjs); err == nil {
		for _, o := range asObjs {
			for _, k := range []string{"url", "src", "displayUrl"} {
				if s, ok := o[k].(string); ok && s != "" {
					return s
				}
			}
		}
	}
	return ""
}

// ParseFollowingUsernames extracts followed handles from a following-list
// dataset item batch.
func ParseFollowingUsernames(items []json.RawMessage) []string {
	var out []string
	for _, item := range items {
		var u struct {
			Username string `json:"username"`
		}
		if err := json.Unmarshal(item, &u); err == nil && u.Username != "" {
			out = append(out, u.Username)
		}
	}
	return out
}

// extractCoauthor pulls the collaboration co-author handle from whichever
// field the actor version populated.
func extractCoauthor(p apifyPost) string {
	if p.CollabWith != "" {
		return strings.TrimPrefix(strings.TrimSpace(p.CollabWith), "@")
	}
	for _, raw := range []json.RawMessage{p.CoauthorProducers, p.Coauthors} {
		if handles := parseUserRefs(raw, "username"); len(handles) > 0 {
			return handles[0]
		}
	}
	return ""
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

// parseUserRefs tolerates both ["handle", ...] and [{"key": "handle"}, ...].
func parseUserRefs(raw json.RawMessage, key string) []string {
	if len(raw) == 0 {
		return nil
	}
	var asStrings []string
	if err := json.Unmarshal(raw, &asStrings); err == nil {
		var out []string
		for _, s := range asStrings {
			if s = strings.TrimSpace(strings.TrimPrefix(s, "@")); s != "" {
				out = append(out, s)
			}
		}
		return out
	}
	var asObjs []map[string]any
	if err := json.Unmarshal(raw, &asObjs); err == nil {
		var out []string
		for _, o := range asObjs {
			if s, ok := o[key].(string); ok && s != "" {
				out = append(out, s)
			}
		}
		return out
	}
	return nil
}

func parseApifyTime(iso string, unixRaw json.RawMessage) time.Time {
	if iso != "" {
		if t, err := time.Parse(time.RFC3339, iso); err == nil {
			return t.UTC()
		}
		// Some actors emit "2026-06-03T18:30:00.000Z" variants covered by
		// RFC3339; anything else falls through to unix.
	}
	if len(unixRaw) > 0 {
		var sec int64
		if err := json.Unmarshal(unixRaw, &sec); err != nil {
			var s string
			if json.Unmarshal(unixRaw, &s) == nil {
				sec, _ = strconv.ParseInt(s, 10, 64)
			}
		}
		if sec > 0 {
			return time.Unix(sec, 0).UTC()
		}
	}
	// A post with no usable timestamp is treated as observed now — better
	// than dropping it, and the recency evidence window is on the order of
	// days so the approximation is safe.
	return time.Now().UTC()
}
