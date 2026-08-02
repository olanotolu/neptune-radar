package connectors

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// DiscoveredParish is one parish scraped from a diocese parish-finder
// directory. WebsiteURL is a link to the parish's own site (or an internal
// detail page when the directory does not link out directly); Address is set
// only when a street-style address is visible on the directory listing.
type DiscoveredParish struct {
	Name       string
	Address    string
	WebsiteURL string
}

// ParishDiscoveryConnector scrapes a diocese parish-finder directory and
// returns the parishes it lists. It handles the three common shapes these
// directories take: a server-rendered HTML table of parishes, a <ul>/<li> list
// of parish links, and a JSON API (ParishSoft and similar widgets sometimes
// expose one). It never guesses a parish that is not actually present on the
// fetched page.
//
// ponytail: HTML parsing uses regex over a full DOM parser because
// golang.org/x/net/html is not in go.mod and the rules say stdlib + existing
// deps only. This is fragile to markup drift — the upgrade path is to add
// x/net/html or a CSS-selector lib when a directory changes structure.
type ParishDiscoveryConnector struct {
	Client       *http.Client
	RequestDelay time.Duration // polite delay before each fetch; default 2s
}

func NewParishDiscoveryConnector() *ParishDiscoveryConnector {
	return &ParishDiscoveryConnector{
		Client:       &http.Client{Timeout: 20 * time.Second},
		RequestDelay: 2 * time.Second,
	}
}

// DiscoverParishes fetches the diocese directory URL and returns the parishes
// listed there. A failed fetch returns an error but never a partial list —
// callers treat a failed diocese as skipped, not fatal.
func (c *ParishDiscoveryConnector) DiscoverParishes(ctx context.Context, directoryURL string) ([]DiscoveredParish, error) {
	body, err := c.fetch(ctx, directoryURL)
	if err != nil {
		return nil, err
	}

	// JSON API first: some ParishSoft deployments and custom directories serve
	// a JSON array (or an object wrapping one) of parish records.
	if ps, ok := parseJSONParishes(body, directoryURL); ok {
		return ps, nil
	}

	return parseHTMLParishes(body, directoryURL), nil
}

// CheckHealth lets ParishDiscoveryConnector satisfy SourceConnector — a simple
// reachability GET delegated to HTTPHealthConnector. The real work is
// DiscoverParishes.
func (c *ParishDiscoveryConnector) CheckHealth(ctx context.Context, endpointURL string) CheckResult {
	return NewHTTPHealthConnector().CheckHealth(ctx, endpointURL)
}

func (c *ParishDiscoveryConnector) fetch(ctx context.Context, pageURL string) ([]byte, error) {
	if c.RequestDelay > 0 {
		select {
		case <-time.After(c.RequestDelay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pageURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "NeptuneRadar-ParishDiscovery/1.0 (+https://meetneptune.com; parish directory enumeration)")
	req.Header.Set("Accept", "text/html,application/json,*/*;q=0.8")
	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d from %s", resp.StatusCode, pageURL)
	}
	return io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024))
}

// --- JSON -------------------------------------------------------------------

// parseJSONParishes tries to read body as a JSON directory of parishes. It
// accepts a top-level array or an object whose values include arrays of
// records with a name-like field. Returns ok=true only when it found JSON it
// recognized as a parish list.
func parseJSONParishes(body []byte, base string) ([]DiscoveredParish, bool) {
	trimmed := strings.TrimSpace(string(body))
	if !strings.HasPrefix(trimmed, "{") && !strings.HasPrefix(trimmed, "[") {
		return nil, false
	}

	// Collect every JSON object anywhere in the tree that looks like a parish
	// record (has a name + at least one of url/website/address). This handles
	// both [{"name":...}] and {"parishes":[...]} / {"data":[...]} shapes.
	var found []DiscoveredParish
	walkJSONObjects(body, func(raw json.RawMessage) {
		var rec struct {
			Name       string `json:"name"`
			Parish     string `json:"parish"`
			Title      string `json:"title"`
			URL        string `json:"url"`
			Website    string `json:"website"`
			WebsiteURL string `json:"websiteUrl"`
			ParishURL  string `json:"parishUrl"`
			Address    string `json:"address"`
			Street     string `json:"street"`
			Location   string `json:"location"`
		}
		if json.Unmarshal(raw, &rec) != nil {
			return
		}
		name := firstNonEmpty(rec.Name, rec.Parish, rec.Title)
		if name == "" || !looksLikeParishName(name) {
			return
		}
		site := firstNonEmpty(rec.WebsiteURL, rec.Website, rec.URL, rec.ParishURL)
		addr := firstNonEmpty(rec.Address, rec.Street, rec.Location)
		found = append(found, DiscoveredParish{
			Name: name, Address: cleanAddress(addr), WebsiteURL: resolveLink(site, base),
		})
	})
	if len(found) == 0 {
		// Only claim "ok" when we actually parsed JSON and found parishes —
		// otherwise fall through to HTML parsing (some JSON is unrelated).
		// But still return true if it was valid JSON with no parishes, so the
		// caller doesn't misparse JSON as HTML. Distinguish by checking it
		// really was JSON.
		if _, err := decodeAny(body); err == nil {
			return nil, true
		}
		return nil, false
	}
	return dedupByName(found), true
}

// walkJSONObjects decodes body as any JSON value and invokes fn for every JSON
// object encountered at any depth (including inside arrays and as object
// values).
func walkJSONObjects(body []byte, fn func(json.RawMessage)) {
	v, err := decodeAny(body)
	if err != nil {
		return
	}
	var walk func(val any)
	walk = func(val any) {
		switch v := val.(type) {
		case map[string]any:
			raw, _ := json.Marshal(v)
			fn(raw)
			for _, child := range v {
				walk(child)
			}
		case []any:
			for _, child := range v {
				walk(child)
			}
		}
	}
	walk(v)
}

func decodeAny(body []byte) (any, error) {
	var v any
	dec := json.NewDecoder(strings.NewReader(string(body)))
	dec.UseNumber()
	return v, dec.Decode(&v)
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if s := strings.TrimSpace(v); s != "" {
			return s
		}
	}
	return ""
}

// --- HTML -------------------------------------------------------------------

// itemRe splits the page into candidate parish "items": table rows and list
// items. Either is a common container for one parish on a diocese directory.
var itemRe = regexp.MustCompile(`(?is)<(?:tr|li)[^>]*>(.*?)</(?:tr|li)>`)

var anchorRe = regexp.MustCompile(`(?is)<a\s[^>]*href=["']([^"']*)["'][^>]*>(.*?)</a>`)

// parishNameRe is a permissive heuristic for "this text is a parish name":
// contains one of the common Catholic parish keywords. Good enough to skip
// nav links and "Search" buttons; the ceiling is unusual named chapels
// without these keywords (upgrade to a name DB if precision matters). No
// trailing \b: dotted abbreviations (st., sts., mt.) end on a non-word char.
var parishNameRe = regexp.MustCompile(`(?i)\b(saint|sts\.|st\.|mt\.|cathedral|our lady of|our lady|holy|sacred|divine|blessed|christ the|church of|chapel|parish|mission|shrine)`)

// addressRe matches a US street-style address (a leading number + a street
// suffix). Intentionally narrow to avoid grabbing phone numbers or zip codes.
var addressRe = regexp.MustCompile(`(?i)\b\d{1,6}\s+[A-Za-z0-9.\-']+(?:\s+[A-Za-z0-9.\-']+){0,5}\s+(?:st|street|ave|avenue|rd|road|dr|drive|blvd|boulevard|ln|lane|ct|court|pl|place|hwy|highway|way|pkwy|parkway|cir|circle)\b\.?`)

func parseHTMLParishes(body []byte, base string) []DiscoveredParish {
	var out []DiscoveredParish
	for _, m := range itemRe.FindAllSubmatch(body, -1) {
		chunk := m[1]
		// Find the first anchor in this item whose text looks like a parish.
		for _, a := range anchorRe.FindAllSubmatch(chunk, -1) {
			href := string(a[1])
			text := stripTags(string(a[2]))
			if text == "" || !looksLikeParishName(text) {
				continue
			}
			dp := DiscoveredParish{
				Name:       normalizeParishName(text),
				WebsiteURL: resolveLink(href, base),
				Address:    cleanAddress(findAddress(string(chunk))),
			}
			out = append(out, dp)
			break
		}
	}
	// Fallback: some directories are a flat page of parish links with no
	// table/list wrapper. Pick up anchors directly so we still discover them.
	if len(out) == 0 {
		for _, a := range anchorRe.FindAllSubmatch(body, -1) {
			href := string(a[1])
			text := stripTags(string(a[2]))
			if text == "" || !looksLikeParishName(text) {
				continue
			}
			out = append(out, DiscoveredParish{
				Name:       normalizeParishName(text),
				WebsiteURL: resolveLink(href, base),
			})
		}
	}
	return dedupByName(out)
}

func looksLikeParishName(s string) bool {
	return parishNameRe.MatchString(s)
}

// findAddress returns the first address-like substring in a chunk, or "".
func findAddress(chunk string) string {
	// Strip tags so an address split across inline elements still matches.
	plain := stripTags(chunk)
	if m := addressRe.FindString(plain); m != "" {
		return m
	}
	return ""
}

// normalizeParishName collapses whitespace and trims trailing punctuation the
// anchor text sometimes carries (e.g. "St. Mary -" or "St. Mary »").
func normalizeParishName(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimRight(s, " -–—»>,.")
	return wsRe.ReplaceAllString(s, " ")
}

func cleanAddress(s string) string {
	return strings.TrimSpace(s)
}

// resolveLink turns a possibly-relative href into an absolute URL against the
// directory page. Empty or script-only hrefs return "".
func resolveLink(href, base string) string {
	href = strings.TrimSpace(href)
	if href == "" || strings.HasPrefix(strings.ToLower(href), "javascript:") {
		return ""
	}
	if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
		return href
	}
	if u, err := url.Parse(base); err == nil {
		if ref, err := u.Parse(href); err == nil {
			return ref.String()
		}
	}
	return href
}

// dedupByName keeps the first occurrence of each (name, website) pair so a
// parish linked from both a table and a fallback scan isn't double-counted.
func dedupByName(ps []DiscoveredParish) []DiscoveredParish {
	seen := map[string]bool{}
	out := ps[:0]
	for _, p := range ps {
		key := strings.ToLower(p.Name) + "\x00" + p.WebsiteURL
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, p)
	}
	return out
}
