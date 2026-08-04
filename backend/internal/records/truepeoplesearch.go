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

// TruePeopleSearch scrapes truepeoplesearch.com via Bright Data Web Unlocker.
// Requires a Web Unlocker *zone* on the Bright Data account (not just a dataset API key).
// Set BRIGHTDATA_UNLOCKER_ZONE to your zone name (Dashboard → Zones → Web Unlocker).
type TruePeopleSearch struct {
	Token  string
	Zone   string
	Client *http.Client

	// resolvedZone is cached after the first successful unlock.
	resolvedZone string
	zoneMu       sync.Mutex
}

func (t *TruePeopleSearch) Name() string { return "truepeoplesearch" }
func (t *TruePeopleSearch) Available() bool {
	return strings.TrimSpace(t.Token) != ""
}

func (t *TruePeopleSearch) client() *http.Client {
	if t.Client != nil {
		return t.Client
	}
	return &http.Client{Timeout: 90 * time.Second}
}

func (t *TruePeopleSearch) zoneCandidates() []string {
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
	add(t.Zone)
	add(os.Getenv("BRIGHTDATA_UNLOCKER_ZONE"))
	add(t.resolvedZone)
	// Common defaults when customers name the zone after the product
	for _, z := range []string{"web_unlocker", "web_unlocker1", "unlocker", "webunlocker", "tps_unlocker"} {
		add(z)
	}
	return out
}

// Search: name + city → results page → detail pages → street candidates.
// Person A / Person B should be called as separate Search queries by the detective.
func (t *TruePeopleSearch) Search(ctx context.Context, q Query) (Result, error) {
	if !t.Available() {
		return Result{Provider: "truepeoplesearch", Status: "empty",
			Error: "BRIGHTDATA_API_KEY not set"}, nil
	}

	name := strings.TrimSpace(q.FirstName + " " + q.LastName)
	if q.FirstName == "" || q.LastName == "" {
		// First+last required for usable TPS hits
		return Result{Provider: "truepeoplesearch", Status: "empty",
			Error: "first and last name required for TruePeopleSearch"}, nil
	}
	loc := strings.TrimSpace(strings.Trim(q.City+", "+q.Region, ", "))

	searchURL := "https://www.truepeoplesearch.com/results?name=" + url.QueryEscape(name)
	if loc != "" {
		searchURL += "&citystatezip=" + url.QueryEscape(loc)
	}

	html, zoneUsed, err := t.unlock(ctx, searchURL)
	if err != nil {
		// Clear setup message — this is why operators only see "tags"
		note := fmt.Sprintf(
			"TruePeopleSearch blocked: %s. Create a Web Unlocker zone in Bright Data (Zones → Add → Web Unlocker), then set BRIGHTDATA_UNLOCKER_ZONE=<zone_name>. Searched: %s",
			err.Error(), name)
		if loc != "" {
			note += " near " + loc
		}
		return Result{
			Provider: "truepeoplesearch",
			Status:   "error",
			Error:    err.Error(),
			Candidates: []Candidate{
				ResearchLink(searchURL, q.City, q.Region, name, "truepeoplesearch", note),
			},
			CostCents: 0,
		}, nil
	}

	// Prefer detail pages; also try card-level streets on results HTML
	var cands []Candidate
	cands = append(cands, tpsParseResultsCards(html, name, q)...)

	slugs := tpsDetailSlugs(html)
	for i, slug := range slugs {
		if len(cands) >= 6 {
			break
		}
		if i > 0 {
			select {
			case <-ctx.Done():
				return Result{Provider: "truepeoplesearch", Status: "error", Error: ctx.Err().Error()}, ctx.Err()
			case <-time.After(1500 * time.Millisecond):
			}
		}
		detailURL := "https://www.truepeoplesearch.com/find/person/" + slug
		detail, dErr := t.unlockWithZone(ctx, detailURL, zoneUsed)
		if dErr != nil {
			continue
		}
		for _, c := range tpsParseAllAddresses(detail, name) {
			if IsRealStreet(c.Line1) {
				c.Source = "truepeoplesearch"
				c.Country = "US"
				c.Confidence = rankTPSConfidence(c, q)
				c.Note = fmt.Sprintf("TruePeopleSearch detail · searched %s · zone=%s — verify before mail.", name, zoneUsed)
				cands = append(cands, c)
			}
		}
	}

	cands = dedupeCandidates(normalizeCandidates(cands))
	rankCandidates(cands)
	st := "ok"
	if len(cands) == 0 {
		st = "empty"
		// Still give research link so operator can finish by hand
		cands = append(cands, ResearchLink(searchURL, q.City, q.Region, name, "truepeoplesearch",
			fmt.Sprintf("TPS unlocked (zone=%s) but no street parsed for %s — open link and paste address.", zoneUsed, name)))
	}
	return Result{
		Provider:   "truepeoplesearch",
		Status:     st,
		Candidates: cands,
		RawJSON:    truncate(html, 4000),
		CostCents:  0,
	}, nil
}

func rankTPSConfidence(c Candidate, q Query) float64 {
	conf := 0.68
	if q.City != "" && c.City != "" && strings.EqualFold(c.City, q.City) {
		conf += 0.08
	}
	if q.Region != "" && c.Region != "" && strings.EqualFold(c.Region, q.Region) {
		conf += 0.04
	}
	if c.Postal != "" {
		conf += 0.03
	}
	if conf > 0.88 {
		conf = 0.88
	}
	return conf
}

// unlock tries zone candidates until one works; caches the winner.
func (t *TruePeopleSearch) unlock(ctx context.Context, target string) (html, zone string, err error) {
	t.zoneMu.Lock()
	cached := t.resolvedZone
	t.zoneMu.Unlock()
	if cached != "" {
		html, err = t.unlockWithZone(ctx, target, cached)
		if err == nil {
			return html, cached, nil
		}
	}
	var lastErr error
	for _, z := range t.zoneCandidates() {
		html, err = t.unlockWithZone(ctx, target, z)
		if err == nil {
			t.zoneMu.Lock()
			t.resolvedZone = z
			t.zoneMu.Unlock()
			return html, z, nil
		}
		lastErr = err
		// Permanent: zone not found — try next; other errors may still try next zone
		if !strings.Contains(err.Error(), "not found") && !strings.Contains(err.Error(), "zone") {
			// auth / rate limit — stop
			if strings.Contains(err.Error(), "401") || strings.Contains(err.Error(), "403") {
				return "", "", err
			}
		}
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("no Bright Data Web Unlocker zone configured")
	}
	return "", "", lastErr
}

func (t *TruePeopleSearch) unlockWithZone(ctx context.Context, target, zone string) (string, error) {
	if zone == "" {
		return "", fmt.Errorf("empty zone")
	}
	body, _ := json.Marshal(map[string]any{
		"zone":   zone,
		"url":    target,
		"format": "raw",
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.brightdata.com/request", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+t.Token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := t.client().Do(req)
	if err != nil {
		return "", fmt.Errorf("tps unlock: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("tps unlock http %d zone=%s: %s", resp.StatusCode, zone, truncate(string(raw), 200))
	}
	s := string(raw)
	// Bright Data sometimes returns JSON error with 200
	if strings.HasPrefix(strings.TrimSpace(s), "{") && strings.Contains(s, "not found") {
		return "", fmt.Errorf("zone %q not found: %s", zone, truncate(s, 120))
	}
	return s, nil
}

// --- parsing -----------------------------------------------------------------

var tpsSlugRe = regexp.MustCompile(`/find/person/([A-Za-z0-9_\-]+)`)

func tpsDetailSlugs(html string) []string {
	matches := tpsSlugRe.FindAllStringSubmatch(html, -1)
	seen := map[string]bool{}
	var out []string
	for _, m := range matches {
		slug := m[1]
		if seen[slug] {
			continue
		}
		seen[slug] = true
		out = append(out, slug)
	}
	return out
}

var (
	tpsStreetRe = regexp.MustCompile(`(?i)itemprop=["']streetAddress["'][^>]*>([^<]+)<`)
	tpsCityRe   = regexp.MustCompile(`(?i)itemprop=["']addressLocality["'][^>]*>([^<]+)<`)
	tpsStateRe  = regexp.MustCompile(`(?i)itemprop=["']addressRegion["'][^>]*>([^<]+)<`)
	tpsZipRe    = regexp.MustCompile(`(?i)itemprop=["']postalCode["'][^>]*>([^<]+)<`)
	tpsPhoneRe  = regexp.MustCompile(`(?i)itemprop=["']telephone["'][^>]*>([^<]+)<`)
	tpsNameRe   = regexp.MustCompile(`data-fn=["']([^"']*)["']\s+data-ln=["']([^"']*)["']`)
	// Results-card fallback: "123 Main St, City, ST 12345"
	tpsCardAddrRe = regexp.MustCompile(`(?i)(\d{1,6}\s+[A-Za-z0-9.'\-]+(?:\s+[A-Za-z0-9.'\-]+){0,5}\s+(?:Street|St|Avenue|Ave|Road|Rd|Drive|Dr|Lane|Ln|Court|Ct|Boulevard|Blvd|Way|Place|Pl|Circle|Cir)\.?)\s*,\s*([A-Za-z .]+?),\s*([A-Z]{2})\s+(\d{5})`)
)

func tpsParseDetail(html, fallbackName string) Candidate {
	all := tpsParseAllAddresses(html, fallbackName)
	if len(all) == 0 {
		return Candidate{FullName: fallbackName}
	}
	return all[0]
}

// tpsParseAllAddresses returns current + prior addresses from a detail page.
func tpsParseAllAddresses(html, fallbackName string) []Candidate {
	streets := tpsStreetRe.FindAllStringSubmatch(html, -1)
	cities := tpsCityRe.FindAllStringSubmatch(html, -1)
	states := tpsStateRe.FindAllStringSubmatch(html, -1)
	zips := tpsZipRe.FindAllStringSubmatch(html, -1)

	fullName := fallbackName
	if m := tpsNameRe.FindStringSubmatch(html); len(m) > 2 {
		fullName = strings.TrimSpace(m[1] + " " + m[2])
	}

	n := len(streets)
	if n == 0 {
		// Try card-style on detail HTML
		return tpsParseResultsCards(html, fullName, Query{})
	}
	var out []Candidate
	for i := 0; i < n && i < 5; i++ {
		c := Candidate{
			Line1:    strings.TrimSpace(streets[i][1]),
			FullName: fullName,
			Country:  "US",
			Kind:     KindStreet,
			Source:   "truepeoplesearch",
		}
		if i < len(cities) {
			c.City = strings.TrimSpace(cities[i][1])
		}
		if i < len(states) {
			c.Region = strings.TrimSpace(states[i][1])
		}
		if i < len(zips) {
			c.Postal = strings.TrimSpace(zips[i][1])
		}
		if i == 0 {
			c.Note = "TPS current address"
			c.Confidence = 0.72
		} else {
			c.Note = "TPS prior address (may be stale)"
			c.Confidence = 0.55
		}
		if IsRealStreet(c.Line1) {
			out = append(out, c)
		}
	}
	if m := tpsPhoneRe.FindStringSubmatch(html); len(m) > 1 && len(out) > 0 {
		out[0].Phone = strings.TrimSpace(m[1])
	}
	return out
}

func tpsParseResultsCards(html, name string, q Query) []Candidate {
	matches := tpsCardAddrRe.FindAllStringSubmatch(html, 8)
	var out []Candidate
	seen := map[string]bool{}
	for _, m := range matches {
		line1 := strings.TrimSpace(m[1])
		if !IsRealStreet(line1) || seen[strings.ToLower(line1)] {
			continue
		}
		seen[strings.ToLower(line1)] = true
		c := Candidate{
			Line1: line1, City: strings.TrimSpace(m[2]), Region: strings.TrimSpace(m[3]),
			Postal: strings.TrimSpace(m[4]), Country: "US", Kind: KindStreet,
			Source: "truepeoplesearch", FullName: name,
			Confidence: 0.65,
			Note:       "TPS results card — verify on detail before mail.",
		}
		if q.City != "" && !strings.EqualFold(c.City, q.City) {
			c.Confidence -= 0.08
		}
		out = append(out, c)
	}
	return out
}

// tpsFromEnv builds TruePeopleSearch from env.
func tpsFromEnv() *TruePeopleSearch {
	return &TruePeopleSearch{
		Token: strings.TrimSpace(os.Getenv("BRIGHTDATA_API_KEY")),
		Zone:  strings.TrimSpace(os.Getenv("BRIGHTDATA_UNLOCKER_ZONE")),
	}
}

// NewTruePeopleSearchFromEnv is exported for detective's dedicated hunter path.
func NewTruePeopleSearchFromEnv() *TruePeopleSearch {
	return tpsFromEnv()
}

// TPSStatus reports whether live TPS is likely to work (token + optional zone hint).
func TPSStatus() string {
	t := tpsFromEnv()
	if !t.Available() {
		return "unavailable: set BRIGHTDATA_API_KEY"
	}
	if z := strings.TrimSpace(os.Getenv("BRIGHTDATA_UNLOCKER_ZONE")); z != "" {
		return "configured zone=" + z
	}
	return "key set but BRIGHTDATA_UNLOCKER_ZONE unset — will try default zone names (often fails until you create a Web Unlocker zone)"
}
