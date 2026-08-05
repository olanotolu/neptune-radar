package records

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// PropertyAsset holds parsed financial data from county auditor HTML.
// All fields are zero-value when parsing fails — callers should check before use.
type PropertyAsset struct {
	AssessedValue int64   `json:"assessed_value,omitempty"` // county-assessed value (cents-free dollars)
	Sqft          int     `json:"sqft,omitempty"`
	YearBuilt     int     `json:"year_built,omitempty"`
	LotSize       float64 `json:"lot_size,omitempty"` // acres
	TaxAnnual     int64   `json:"tax_annual,omitempty"` // annual property tax (dollars)
}

// countyAvgPricePerSqft is a small table of avg $/sqft for supported counties.
// ponytail: static estimates — refresh annually; not a real estate API.
var countyAvgPricePerSqft = map[string]float64{
	// Ohio
	"franklin": 185, "hamilton": 175, "cuyahoga": 130, "delaware": 240, "summit": 150,
	"montgomery": 130, "warren": 210, "butler": 170, "lorain": 145, "lucas": 120,
	// Texas
	"harris": 165, "dallas": 210, "travis": 320, "bexar": 175, "collin": 250,
	"tarrant": 170, "fortbend": 200, "denton": 230, "williamson": 240,
	// Florida
	"miami-dade": 310, "broward": 260, "palmbeach": 350, "orange": 220,
	"hillsborough": 200, "pinellas": 230, "duval": 180, "lee": 210,
}

// CountyAvgPricePerSqft returns the avg $/sqft for a county, or 0 if unknown.
func CountyAvgPricePerSqft(county, region string) float64 {
	key := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(county), " ", ""))
	if v, ok := countyAvgPricePerSqft[key]; ok {
		return v
	}
	// State-level fallback
	switch strings.ToUpper(strings.TrimSpace(region)) {
	case "OH":
		return 160
	case "TX":
		return 200
	case "FL":
		return 250
	default:
		return 0
	}
}

// EstimateHomeValue estimates market value from property asset data.
// Model: sqft * countyAvgPricePerSqft. If assessed value is available,
// returns max(assessed, estimated) — assessed often lags market.
// Returns 0 when neither sqft nor assessed value is available.
func EstimateHomeValue(asset PropertyAsset, countyAvgPricePerSqft float64) int64 {
	var est int64
	if asset.Sqft > 0 && countyAvgPricePerSqft > 0 {
		est = int64(float64(asset.Sqft) * countyAvgPricePerSqft)
	}
	if asset.AssessedValue > est {
		return asset.AssessedValue
	}
	return est
}

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

// --- Asset parsing regexes (county auditor HTML) ---
// ponytail: regex over goquery — no new dep, matches existing codebase pattern.
// Each regex looks for label/value pairs common across OH/TX/FL auditor sites.
// [\s\S]{0,80}? allows HTML tags between label and value (e.g. </td><td>).
// If a county changes their HTML, these silently return zero — no crash.
var (
	assessedValRe = regexp.MustCompile(`(?i)(?:assessed|appraised|market)\s*(?:value|valuation)?[\s\S]{0,80}?\$?\s*([\d,]+(?:\.\d{2})?)`)
	sqftRe        = regexp.MustCompile(`(?i)(?:sq\.?\s*ft\.?|square\s*feet|living\s*area|gross\s*area)[\s\S]{0,80}?\$?\s*([\d,]+)`)
	yearBuiltRe   = regexp.MustCompile(`(?i)(?:year\s*built|yr\s*built|year\s*constructed)[\s\S]{0,80}?\b(\d{4})\b`)
	lotSizeRe     = regexp.MustCompile(`(?i)(?:lot\s*size|lot\s*area|acreage|land\s*area)[\s\S]{0,80}?\b([\d.]+)\b`)
	taxAnnualRe   = regexp.MustCompile(`(?i)(?:annual\s*tax|tax\s*amount|total\s*tax|property\s*tax)[\s\S]{0,80}?\$?\s*([\d,]+(?:\.\d{2})?)`)
)

// parseDollars strips commas/$ and parses a dollar string to int64.
// Handles both "285,000" and "285,000.00" (truncates cents).
func parseDollars(s string) int64 {
	s = strings.ReplaceAll(s, ",", "")
	s = strings.ReplaceAll(s, "$", "")
	s = strings.TrimSpace(s)
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return int64(f)
	}
	n, _ := strconv.ParseInt(s, 10, 64)
	return n
}

// parseInt strips commas and parses to int.
func parseInt(s string) int {
	s = strings.ReplaceAll(s, ",", "")
	s = strings.TrimSpace(s)
	n, _ := strconv.Atoi(s)
	return n
}

// parseFloat parses a float string.
func parseFloat(s string) float64 {
	s = strings.TrimSpace(s)
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

// parsePropertyAsset extracts financial fields from county auditor HTML.
// Returns zero-value PropertyAsset on any parse failure — never crashes.
func parsePropertyAsset(html string) PropertyAsset {
	var a PropertyAsset
	if m := assessedValRe.FindStringSubmatch(html); len(m) > 1 {
		a.AssessedValue = parseDollars(m[1])
	}
	if m := sqftRe.FindStringSubmatch(html); len(m) > 1 {
		a.Sqft = parseInt(m[1])
	}
	if m := yearBuiltRe.FindStringSubmatch(html); len(m) > 1 {
		a.YearBuilt = parseInt(m[1])
	}
	if m := lotSizeRe.FindStringSubmatch(html); len(m) > 1 {
		a.LotSize = parseFloat(m[1])
	}
	if m := taxAnnualRe.FindStringSubmatch(html); len(m) > 1 {
		a.TaxAnnual = parseDollars(m[1])
	}
	return a
}

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

	// Parse asset data once from the full page (county auditor detail pages
	// show one property; search result lists rarely include assessed values).
	asset := parsePropertyAsset(html)

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
		c := Candidate{
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
		}
		// Attach parsed asset data to the first real street candidate.
		if asset.AssessedValue > 0 || asset.Sqft > 0 {
			c.Asset = asset
		}
		cands = append(cands, c)
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
