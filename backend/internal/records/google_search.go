package records

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

// GoogleSearch uses Google Custom Search API (free 100 queries/day) to find
// people-search results for name + city. Parses Whitepages/FastPeopleSearch
// snippets from search results for address hints.
type GoogleSearch struct {
	APIKey string
	Cx     string // Custom Search Engine ID
	Client *http.Client
}

func (g *GoogleSearch) Name() string { return "google_search" }
func (g *GoogleSearch) Available() bool {
	return strings.TrimSpace(g.APIKey) != "" && strings.TrimSpace(g.Cx) != ""
}

func (g *GoogleSearch) client() *http.Client {
	if g.Client != nil {
		return g.Client
	}
	return &http.Client{Timeout: 15 * time.Second}
}

func (g *GoogleSearch) Search(ctx context.Context, q Query) (Result, error) {
	if !g.Available() {
		return Result{Provider: "google_search", Status: "error", Error: "GOOGLE_SEARCH_API_KEY or GOOGLE_SEARCH_CX not set"}, fmt.Errorf("google search unavailable")
	}

	name := strings.TrimSpace(q.FirstName + " " + q.LastName)
	if name == " " {
		name = q.FirstName
	}
	city := q.City
	region := q.Region

	// Build search query targeting people-search sites
	query := fmt.Sprintf(`"%s" %s %s address`, name, city, region)
_searchURL := fmt.Sprintf("https://www.googleapis.com/customsearch/v1?key=%s&cx=%s&q=%s&num=5",
		url.QueryEscape(g.APIKey),
		url.QueryEscape(g.Cx),
		url.QueryEscape(query))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, _searchURL, nil)
	if err != nil {
		return Result{Provider: "google_search", Status: "error", Error: err.Error()}, err
	}

	resp, err := g.client().Do(req)
	if err != nil {
		return Result{Provider: "google_search", Status: "error", Error: err.Error()}, err
	}
	defer resp.Body.Close()
	rawB, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	raw := string(rawB)

	if resp.StatusCode >= 300 {
		return Result{Provider: "google_search", Status: "error", RawJSON: raw, Error: fmt.Sprintf("google http %d", resp.StatusCode), CostCents: 0},
			fmt.Errorf("google http %d", resp.StatusCode)
	}

	var parsed map[string]any
	_ = json.Unmarshal(rawB, &parsed)

	items, _ := parsed["items"].([]any)
	cands := extractGoogleCandidates(items, q)
	st := "ok"
	if len(cands) == 0 {
		st = "empty"
	}
	return Result{Provider: "google_search", Status: st, Candidates: cands, RawJSON: raw, CostCents: 0}, nil
}

// Street address pattern: "123 Main St" etc.
var streetRe = regexp.MustCompile(`(\d{1,5}\s+[A-Za-z0-9\s.]+(?:St|Ave|Blvd|Rd|Dr|Ln|Ct|Way|Pl|Cir|Hwy|Road|Street|Avenue|Boulevard|Drive|Lane|Court|Place|Circle)\.?(?:\s*(?:#|Apt|Suite|Ste|Unit)\s*\d+)?)`)
var zipRe = regexp.MustCompile(`\b(\d{5})(?:-\d{4})?\b`)

func extractGoogleCandidates(items []any, q Query) []Candidate {
	var cands []Candidate
	seen := map[string]bool{}

	for _, item := range items {
		it, ok := item.(map[string]any)
		if !ok {
			continue
		}
		title, _ := it["title"].(string)
		snippet, _ := it["snippet"].(string)
		link, _ := it["link"].(string)
		text := title + " " + snippet

		// Extract street addresses from snippet
		streets := streetRe.FindAllString(text, -1)
		zips := zipRe.FindAllString(text, -1)

		for _, street := range streets {
			street = strings.TrimSpace(street)
			if seen[street] {
				continue
			}
			seen[street] = true

			zip := ""
			if len(zips) > 0 {
				zip = zips[0]
			}
			cands = append(cands, Candidate{
				Line1:      street,
				City:       q.City,
				Region:     q.Region,
				Postal:     zip,
				Country:    "US",
				Confidence: 0.50,
				Source:     "google_search",
				FullName:   strings.TrimSpace(q.FirstName + " " + q.LastName),
				Note:       fmt.Sprintf("Found via Google search on %s. Verify with Lob.", extractDomain(link)),
			})
		}

		// If no street found but zip is present, add city+zip candidate
		if len(streets) == 0 && len(zips) > 0 {
			zip := zips[0]
			if !seen[zip] {
				seen[zip] = true
				cands = append(cands, Candidate{
					City:       q.City,
					Region:     q.Region,
					Postal:     zip,
					Country:    "US",
					Confidence: 0.40,
					Source:     "google_search_zip",
					FullName:   strings.TrimSpace(q.FirstName + " " + q.LastName),
					Note:       fmt.Sprintf("Zip code %s found via Google. Verify with Lob.", zip),
				})
			}
		}
	}

	if len(cands) > 5 {
		cands = cands[:5]
	}
	return cands
}

func extractDomain(link string) string {
	u, err := url.Parse(link)
	if err != nil {
		return "web"
	}
	host := u.Hostname()
	host = strings.TrimPrefix(host, "www.")
	return host
}
