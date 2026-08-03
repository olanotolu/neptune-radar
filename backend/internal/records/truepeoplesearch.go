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
	"time"
)

// TruePeopleSearch scrapes truepeoplesearch.com (free public records) via
// Bright Data's Web Unlocker zone to bypass Cloudflare. No official API.
// ponytail: regex parsing of HTML is fragile — if TPS changes markup this breaks
// silently (returns empty). Upgrade path: switch to x/net/html tokenizer.
type TruePeopleSearch struct {
	Token  string
	Client *http.Client
}

func (t *TruePeopleSearch) Name() string { return "truepeoplesearch" }
func (t *TruePeopleSearch) Available() bool {
	return strings.TrimSpace(t.Token) != ""
}

func (t *TruePeopleSearch) client() *http.Client {
	if t.Client != nil {
		return t.Client
	}
	return &http.Client{Timeout: 60 * time.Second}
}

// Search builds a name+city search URL, fetches the results page via Bright Data
// Web Unlocker, then fetches each detail page for the current address.
func (t *TruePeopleSearch) Search(ctx context.Context, q Query) (Result, error) {
	if !t.Available() {
		return Result{Provider: "truepeoplesearch", Status: "error", Error: "BRIGHTDATA_API_KEY not set"}, fmt.Errorf("truepeoplesearch unavailable")
	}

	name := strings.TrimSpace(q.FirstName + " " + q.LastName)
	if name == "" {
		return Result{Provider: "truepeoplesearch", Status: "empty"}, nil
	}
	loc := strings.TrimSpace(strings.Trim(q.City+", "+q.Region, ", "))

	searchURL := "https://www.truepeoplesearch.com/results?name=" + url.QueryEscape(name)
	if loc != "" {
		searchURL += "&citystatezip=" + url.QueryEscape(loc)
	}

	html, err := t.unlock(ctx, searchURL)
	if err != nil {
		return Result{Provider: "truepeoplesearch", Status: "error", Error: err.Error()}, err
	}

	slugs := tpsDetailSlugs(html)
	if len(slugs) == 0 {
		return Result{Provider: "truepeoplesearch", Status: "empty", RawJSON: truncate(html, 4000)}, nil
	}

	var cands []Candidate
	for _, slug := range slugs {
		if len(cands) >= 5 {
			break
		}
		// Be polite: 2s delay between detail-page fetches.
		select {
		case <-ctx.Done():
			return Result{Provider: "truepeoplesearch", Status: "error", Error: ctx.Err().Error()}, ctx.Err()
		case <-time.After(2 * time.Second):
		}

		detailURL := "https://www.truepeoplesearch.com/find/person/" + slug
		detail, dErr := t.unlock(ctx, detailURL)
		if dErr != nil {
			continue
		}
		c := tpsParseDetail(detail, name)
		if c.Line1 != "" || c.City != "" {
			c.Source = "truepeoplesearch"
			c.Confidence = 0.70
			c.Country = firstNonEmpty(c.Country, "US")
			if c.FullName == "" {
				c.FullName = name
			}
			cands = append(cands, c)
		}
	}

	st := "ok"
	if len(cands) == 0 {
		st = "empty"
	}
	return Result{Provider: "truepeoplesearch", Status: st, Candidates: cands, RawJSON: truncate(html, 4000), CostCents: 0}, nil
}

// unlock fetches a URL through Bright Data Web Unlocker (raw HTML).
func (t *TruePeopleSearch) unlock(ctx context.Context, target string) (string, error) {
	body, _ := json.Marshal(map[string]any{
		"zone":   "unlocker",
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
		return "", fmt.Errorf("tps unlock http %d: %s", resp.StatusCode, truncate(string(raw), 200))
	}
	return string(raw), nil
}

// --- parsing -----------------------------------------------------------------

// tpsDetailSlugs extracts /find/person/{slug} links from the results page.
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

// itemprop extractors for the detail page address block.
var (
	tpsStreetRe = regexp.MustCompile(`itemprop="streetAddress"[^>]*>([^<]+)<`)
	tpsCityRe   = regexp.MustCompile(`itemprop="addressLocality"[^>]*>([^<]+)<`)
	tpsStateRe  = regexp.MustCompile(`itemprop="addressRegion"[^>]*>([^<]+)<`)
	tpsZipRe    = regexp.MustCompile(`itemprop="postalCode"[^>]*>([^<]+)<`)
	tpsPhoneRe  = regexp.MustCompile(`itemprop="telephone"[^>]*>([^<]+)<`)
	tpsNameRe   = regexp.MustCompile(`data-fn="([^"]*)"\s+data-ln="([^"]*)"`)
)

// tpsParseDetail pulls the first (current) address from a person detail page.
// ponytail: only the first address block is used — previous addresses ignored
// (lower confidence, often stale). Upgrade: parse all and let heuristic rank.
func tpsParseDetail(html, fallbackName string) Candidate {
	c := Candidate{}
	if m := tpsStreetRe.FindStringSubmatch(html); len(m) > 1 {
		c.Line1 = strings.TrimSpace(m[1])
	}
	if m := tpsCityRe.FindStringSubmatch(html); len(m) > 1 {
		c.City = strings.TrimSpace(m[1])
	}
	if m := tpsStateRe.FindStringSubmatch(html); len(m) > 1 {
		c.Region = strings.TrimSpace(m[1])
	}
	if m := tpsZipRe.FindStringSubmatch(html); len(m) > 1 {
		c.Postal = strings.TrimSpace(m[1])
	}
	if m := tpsPhoneRe.FindStringSubmatch(html); len(m) > 1 {
		c.Phone = strings.TrimSpace(m[1])
	}
	if m := tpsNameRe.FindStringSubmatch(html); len(m) > 2 {
		c.FullName = strings.TrimSpace(m[1] + " " + m[2])
	}
	if c.FullName == "" {
		c.FullName = fallbackName
	}
	if c.Line1 != "" {
		c.Note = "TruePeopleSearch public records — verify before mail."
	}
	return c
}

// tpsFromEnv builds a TruePeopleSearch provider from BRIGHTDATA_API_KEY.
func tpsFromEnv() *TruePeopleSearch {
	return &TruePeopleSearch{Token: strings.TrimSpace(os.Getenv("BRIGHTDATA_API_KEY"))}
}
