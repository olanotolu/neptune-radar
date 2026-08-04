package records

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// PropertyRecords searches free county auditor property portals by owner name.
// After marriage, couples often buy property together; property records are
// public and searchable by owner name, giving a high-signal mailing address.
type PropertyRecords struct {
	Client *http.Client
}

func (p *PropertyRecords) Name() string    { return "county_property" }
func (p *PropertyRecords) Available() bool { return true } // free public records, no API key

func (p *PropertyRecords) client() *http.Client {
	if p.Client != nil {
		return p.Client
	}
	return &http.Client{Timeout: 15 * time.Second}
}

// countyAuditorURL returns the property search URL for a city/state.
// Returns "" for unsupported states. ponytail: only OH/TX/FL mapped — other
// monitored states (GA, NC, NJ, PA, CT) fall back to the research-note path;
// add their county URL schemes here as they're needed.
func countyAuditorURL(city, region string) string {
	county := CountyName(city, region)
	region = strings.ToUpper(strings.TrimSpace(region))
	if county == "" {
		return "" // no county mapping for this city/state
	}
	switch region {
	case "OH":
		c := strings.ToLower(strings.ReplaceAll(county, " ", ""))
		return fmt.Sprintf("https://%scountyauditor.org/property-search", c)
	case "TX":
		c := strings.ToLower(strings.ReplaceAll(county, " ", ""))
		return fmt.Sprintf("https://%s.taxrecords.com/Search/Owner", c)
	case "FL":
		c := strings.ToLower(strings.ReplaceAll(county, " ", ""))
		return fmt.Sprintf("https://%s.pa.dor.state.fl.us/", c)
	default:
		return ""
	}
}

// streetAddrRe matches US street addresses in county auditor HTML output.
// Copied from bioStreetRe (address_extract.go) per spec — not imported to
// avoid coupling to bio-specific parsing.
var streetAddrRe = regexp.MustCompile(`(?i)\b(\d{1,5})\s+` +
	`(?:(?:N|S|E|W|NE|NW|SE|SW)\.?\s+)?` +
	`([A-Za-z][a-zA-Z]+(?:\s(?:St|Street|Ave|Avenue|Rd|Road|Dr|Drive|Ln|Lane|Ct|Court|Blvd|Boulevard|Way|Pl|Place|Cir|Circle|Pkwy|Parkway|Hwy|Highway|Ter|Terrace|Trail|Trl|Loop|Cove))\.?)` +
	`(?:\s+(?:Apt|Unit|Ste|#)\s*\d+[A-Za-z]?)?`,
)

var propZipRe = regexp.MustCompile(`\b(\d{5})(?:-\d{4})?\b`)

func (p *PropertyRecords) Search(ctx context.Context, q Query) (Result, error) {
	city := strings.TrimSpace(q.City)
	region := strings.ToUpper(strings.TrimSpace(q.Region))
	lastName := strings.TrimSpace(q.LastName)
	if lastName == "" || city == "" || region == "" {
		return Result{Provider: "county_property", Status: "empty"}, nil
	}

	searchURL := countyAuditorURL(city, region)
	if searchURL == "" {
		// Unsupported state/county — not an error, just nothing to search.
		return Result{Provider: "county_property", Status: "empty"}, nil
	}

	// Fetch the county auditor page. Many are JS-rendered; we attempt HTML
	// parse and fall back to a research-note candidate with the URL.
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, searchURL, nil)
	if err != nil {
		return Result{Provider: "county_property", Status: "error", Error: err.Error()}, nil
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	resp, err := p.client().Do(req)
	if err != nil {
		return Result{Provider: "county_property", Status: "error", Error: err.Error()}, nil
	}
	defer resp.Body.Close()
	rawB, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	raw := string(rawB)

	// Be polite to county servers: 2s delay between requests in a session.
	// ponytail: process-wide sleep — fine for the low-volume detective pipeline.
	time.Sleep(2 * time.Second)

	if resp.StatusCode >= 300 {
		return Result{Provider: "county_property", Status: "error", RawJSON: raw, Error: fmt.Sprintf("county auditor http %d", resp.StatusCode)}, nil
	}

	cands := parsePropertyHTML(raw, q, searchURL)
	st := "ok"
	if len(cands) == 0 {
		st = "empty"
	}
	return Result{
		Provider:   "county_property",
		Candidates: cands,
		RawJSON:    raw,
		Status:     st,
		CostCents:  0,
	}, nil
}

// parsePropertyHTML extracts owner-name + address matches from county auditor
// HTML. If no addresses parse (common with JS-rendered portals), returns a
// single research-note candidate pointing the operator at the search URL.
func parsePropertyHTML(html string, q Query, searchURL string) []Candidate {
	lname := strings.ToLower(strings.TrimSpace(q.LastName))
	var cands []Candidate
	seen := map[string]bool{}

	matches := streetAddrRe.FindAllString(html, -1)
	for _, m := range matches {
		m = strings.TrimSpace(m)
		if seen[m] {
			continue
		}
		// Only keep addresses near the owner's last name (cheap proximity check).
		idx := strings.Index(strings.ToLower(html), strings.ToLower(m))
		lo := idx - 200
		if lo < 0 {
			lo = 0
		}
		hi := idx + len(m) + 200
		if hi > len(html) {
			hi = len(html)
		}
		window := html[lo:hi]
		if !strings.Contains(strings.ToLower(window), lname) {
			continue
		}
		seen[m] = true
		zip := ""
		if z := propZipRe.FindString(window); z != "" {
			zip = z
		}
		cands = append(cands, Candidate{
			Line1:      m,
			City:       q.City,
			Region:     q.Region,
			Postal:     zip,
			Country:    "US",
			Kind:       KindStreet,
			Confidence: 0.65,
			Source:     "county_property",
			FullName:   strings.TrimSpace(q.FirstName + " " + q.LastName),
			Note:       "County auditor property record — owner-occupied mailing address. Verify before mailing.",
		})
		if len(cands) >= 5 {
			break
		}
	}

	// JS-rendered portal or no parseable rows: research link (URL not in Line1).
	if len(cands) == 0 {
		cands = append(cands, ResearchLink(
			searchURL, q.City, q.Region,
			strings.TrimSpace(q.FirstName+" "+q.LastName),
			"county_property",
			"County auditor property search — operator should visit and verify",
		))
	}
	return cands
}
