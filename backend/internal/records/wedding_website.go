package records

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

// WeddingWebsiteResult is one parsed wedding-website hit from The Knot / Zola / WeddingWire.
type WeddingWebsiteResult struct {
	URL          string   `json:"url"`
	Platform     string   `json:"platform"` // knot | zola | weddingwire
	WeddingDate  string   `json:"wedding_date,omitempty"`  // ISO yyyy-mm-dd when parseable
	VenueName    string   `json:"venue_name,omitempty"`
	VenueCity    string   `json:"venue_city,omitempty"`
	VenueState   string   `json:"venue_state,omitempty"`
	RegistryURLs []string `json:"registry_urls,omitempty"`
}

// WeddingWebsiteProvider searches The Knot, Zola, and WeddingWire for a couple's
// wedding website by name. Uses Bright Data Web Unlocker when
// BRIGHTDATA_UNLOCKER_ZONE is set (same pattern as truepeoplesearch.go),
// otherwise direct HTTP. Conservative: only returns results where BOTH names
// appear on the page — a single-name match is rejected as too risky.
type WeddingWebsiteProvider struct {
	Token  string
	Zone   string
	Client *http.Client

	resolvedZone string
	zoneMu       sync.Mutex
}

func (w *WeddingWebsiteProvider) Name() string    { return "wedding_website" }
func (w *WeddingWebsiteProvider) Available() bool { return true } // free public search pages

func (w *WeddingWebsiteProvider) client() *http.Client {
	if w.Client != nil {
		return w.Client
	}
	return &http.Client{Timeout: 30 * time.Second}
}

func (w *WeddingWebsiteProvider) token() string {
	if w.Token != "" {
		return w.Token
	}
	return strings.TrimSpace(os.Getenv("BRIGHTDATA_API_KEY"))
}

func (w *WeddingWebsiteProvider) zoneCandidates() []string {
	var out []string
	seen := map[string]bool{}
	add := func(z string) {
		z = strings.TrimSpace(z)
		if z == "" || seen[z] {
			return
		}
		seen[z] = true
		out = append(out, z)
	}
	add(w.Zone)
	add(os.Getenv("BRIGHTDATA_UNLOCKER_ZONE"))
	w.zoneMu.Lock()
	add(w.resolvedZone)
	w.zoneMu.Unlock()
	return out
}

// Search looks up a couple across all three platforms. first1/last1 + first2/last2
// are the two partners' names; city/state narrow the result. Never blocks the
// detective cascade — errors are swallowed per-platform and returned as a Result
// with Status ok/empty. Candidates are left empty (this provider writes to the
// couple row, not the address-candidate stream); RawJSON carries the parsed hits
// for audit.
func (w *WeddingWebsiteProvider) Search(ctx context.Context, q Query) (Result, error) {
	first1, last1 := strings.TrimSpace(q.FirstName), strings.TrimSpace(q.LastName)
	first2, last2 := strings.TrimSpace(q.PartnerFirst), strings.TrimSpace(q.PartnerLast)
	if first1 == "" || last1 == "" || first2 == "" || last2 == "" {
		return Result{Provider: "wedding_website", Status: "empty"}, nil
	}

	var hits []WeddingWebsiteResult
	platforms := []struct {
		name string
		url  string
	}{
		{"knot", fmt.Sprintf("https://www.theknot.com/us/search?query=%s+%s+%s+%s",
			url.QueryEscape(first1), url.QueryEscape(last1), url.QueryEscape(first2), url.QueryEscape(last2))},
		{"zola", fmt.Sprintf("https://www.zola.com/search?query=%s+%s+%s+%s",
			url.QueryEscape(first1), url.QueryEscape(last1), url.QueryEscape(first2), url.QueryEscape(last2))},
		{"weddingwire", fmt.Sprintf("https://www.weddingwire.com/couples/search?name=%s+%s",
			url.QueryEscape(first1), url.QueryEscape(last1))},
	}

	for _, p := range platforms {
		if err := ctx.Err(); err != nil {
			break
		}
		html, err := w.fetch(ctx, p.url)
		if err != nil {
			continue // never block the cascade
		}
		hits = append(hits, parseWeddingWebsiteHTML(html, p.name, p.url, first1, last1, first2, last2)...)
	}

	st := "ok"
	if len(hits) == 0 {
		st = "empty"
	}
	raw := ""
	if len(hits) > 0 {
		b, _ := json.Marshal(hits)
		raw = string(b)
	}
	return Result{Provider: "wedding_website", Status: st, RawJSON: truncate(raw, 8000)}, nil
}

// fetch returns the HTML for a URL via Bright Data Web Unlocker when configured,
// otherwise a direct GET. Errors are non-fatal — callers swallow them.
func (w *WeddingWebsiteProvider) fetch(ctx context.Context, target string) (string, error) {
	if w.token() != "" && len(w.zoneCandidates()) > 0 {
		html, _, err := w.unlock(ctx, target)
		if err == nil {
			return html, nil
		}
		// fall through to direct GET
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	resp, err := w.client().Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("wedding_website http %d", resp.StatusCode)
	}
	return string(raw), nil
}

// unlock tries zone candidates until one works; caches the winner.
func (w *WeddingWebsiteProvider) unlock(ctx context.Context, target string) (string, string, error) {
	w.zoneMu.Lock()
	cached := w.resolvedZone
	w.zoneMu.Unlock()
	if cached != "" {
		html, err := w.unlockWithZone(ctx, target, cached)
		if err == nil {
			return html, cached, nil
		}
	}
	var lastErr error
	for _, z := range w.zoneCandidates() {
		html, err := w.unlockWithZone(ctx, target, z)
		if err == nil {
			w.zoneMu.Lock()
			w.resolvedZone = z
			w.zoneMu.Unlock()
			return html, z, nil
		}
		lastErr = err
		if strings.Contains(err.Error(), "401") || strings.Contains(err.Error(), "403") {
			return "", "", err
		}
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("no Bright Data Web Unlocker zone configured")
	}
	return "", "", lastErr
}

func (w *WeddingWebsiteProvider) unlockWithZone(ctx context.Context, target, zone string) (string, error) {
	if zone == "" {
		return "", fmt.Errorf("empty zone")
	}
	body, _ := json.Marshal(map[string]any{"zone": zone, "url": target, "format": "raw"})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.brightdata.com/request", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+w.token())
	req.Header.Set("Content-Type", "application/json")
	resp, err := w.client().Do(req)
	if err != nil {
		return "", fmt.Errorf("wedding_website unlock: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("wedding_website unlock http %d zone=%s", resp.StatusCode, zone)
	}
	s := string(raw)
	if strings.HasPrefix(strings.TrimSpace(s), "{") && strings.Contains(s, "not found") {
		return "", fmt.Errorf("zone %q not found", zone)
	}
	return s, nil
}

// --- HTML parsing -----------------------------------------------------------
// ponytail: each platform's selectors are a regex shortcut over fragile HTML.
// If a platform changes its markup these silently return empty — no crash.
// Upgrade path: switch to x/net/html tokenizer per platform.

var (
	// weddingDateRe matches "October 12, 2025", "Oct 12 2025", "2025-10-12".
	weddingDateRe = regexp.MustCompile(`(?i)\b((?:January|February|March|April|May|June|July|August|September|October|November|December|Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec)\.?\s+\d{1,2},?\s+\d{4}|\d{4}-\d{2}-\d{2})\b`)
	// isoDateRe pulls yyyy-mm-dd out of a longer match.
	isoDateRe = regexp.MustCompile(`\d{4}-\d{2}-\d{2}`)
	// registryLinkRe matches anchor hrefs containing registry domains.
	registryLinkRe = regexp.MustCompile(`(?i)href=["']([^"']*(?:registry|zola\.com/registry|theknot\.com/registry|amazon\.com/wedding|target\.com/wedding|crateandbarrel\.com/wedding|macys\.com/wedding|bedbathandbeyond\.com/wedding)[^"']*)["']`)
	// venueRe looks for "Venue: ..." or a venue-name label near the word venue.
	venueRe = regexp.MustCompile(`(?i)(?:venue|location|ceremony|reception)[\s\S]{0,40}?([^<\n]{3,80})`)
)

// parseWeddingWebsiteHTML extracts wedding-website hits from a platform's search
// results HTML. Conservative: requires BOTH partner names to appear on the page
// (case-insensitive) before returning any result — a single-name match is dropped.
func parseWeddingWebsiteHTML(html, platform, pageURL, first1, last1, first2, last2 string) []WeddingWebsiteResult {
	low := strings.ToLower(html)
	nameA := strings.ToLower(strings.TrimSpace(first1 + " " + last1))
	nameB := strings.ToLower(strings.TrimSpace(first2 + " " + last2))
	// Require both full names present — single-name matches are too risky.
	if nameA == "" || nameB == "" {
		return nil
	}
	if !strings.Contains(low, nameA) || !strings.Contains(low, nameB) {
		return nil
	}

	r := WeddingWebsiteResult{URL: pageURL, Platform: platform}

	if m := weddingDateRe.FindString(html); m != "" {
		r.WeddingDate = normalizeWeddingDate(m)
	}

	if m := venueRe.FindStringSubmatch(html); len(m) > 1 {
		r.VenueName = strings.TrimSpace(strings.Trim(m[1], " :,.;\t\n"))
	}

	// Registry links — dedupe, cap at 5.
	seen := map[string]bool{}
	for _, m := range registryLinkRe.FindAllStringSubmatch(html, -1) {
		u := strings.TrimSpace(m[1])
		if u == "" || seen[u] {
			continue
		}
		seen[u] = true
		if !strings.HasPrefix(u, "http") {
			u = "https://" + u
		}
		r.RegistryURLs = append(r.RegistryURLs, u)
		if len(r.RegistryURLs) >= 5 {
			break
		}
	}

	// Only return a hit if we extracted at least a date or a venue — a bare
	// name match with no wedding detail is not actionable.
	if r.WeddingDate == "" && r.VenueName == "" && len(r.RegistryURLs) == 0 {
		return nil
	}
	return []WeddingWebsiteResult{r}
}

// normalizeWeddingDate converts "October 12, 2025" → "2025-10-12". Passes through
// already-ISO dates. Returns "" on parse failure.
func normalizeWeddingDate(s string) string {
	s = strings.TrimSpace(s)
	if isoDateRe.MatchString(s) {
		return isoDateRe.FindString(s)
	}
	months := map[string]int{
		"january": 1, "february": 2, "march": 3, "april": 4, "may": 5, "june": 6,
		"july": 7, "august": 8, "september": 9, "october": 10, "november": 11, "december": 12,
		"jan": 1, "feb": 2, "mar": 3, "apr": 4, "jun": 6, "jul": 7, "aug": 8,
		"sep": 9, "oct": 10, "nov": 11, "dec": 12,
	}
	re := regexp.MustCompile(`(?i)([A-Za-z]+)\.?\s+(\d{1,2}),?\s+(\d{4})`)
	m := re.FindStringSubmatch(s)
	if len(m) != 4 {
		return ""
	}
	mo, ok := months[strings.ToLower(m[1])]
	if !ok {
		return ""
	}
	day := strings.TrimPrefix(m[2], "0")
	return fmt.Sprintf("%s-%02d-%02d", m[3], mo, atoiSafe(day))
}

func atoiSafe(s string) int {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			break
		}
		n = n*10 + int(c-'0')
	}
	return n
}
