package records

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// VoterRegistration looks up public voter-registration records by state.
// Most state portals are interactive (forms, CAPTCHAs), so for those we return
// the search URL as a low-confidence research-note candidate for an operator.
// North Carolina exposes a simple URL-based search we auto-fetch.
type VoterRegistration struct {
	Client *http.Client
}

func (v *VoterRegistration) Name() string    { return "voter_registration" }
func (v *VoterRegistration) Available() bool { return true } // free public records

func (v *VoterRegistration) client() *http.Client {
	if v.Client != nil {
		return v.Client
	}
	return &http.Client{Timeout: 20 * time.Second}
}

// Search builds a voter-lookup URL for the query's state. For NC it auto-fetches
// and parses results; for other supported states it returns the URL as a
// research note. Unsupported/empty states return an empty result.
func (v *VoterRegistration) Search(ctx context.Context, q Query) (Result, error) {
	region := strings.ToUpper(strings.TrimSpace(q.Region))
	if region == "" {
		return Result{Provider: "voter_registration", Status: "empty"}, nil
	}
	first := strings.TrimSpace(q.FirstName)
	last := strings.TrimSpace(q.LastName)
	if first == "" || last == "" {
		return Result{Provider: "voter_registration", Status: "empty"}, nil
	}
	fullName := strings.TrimSpace(first + " " + last)

	switch region {
	case "NC":
		return v.searchNC(ctx, q, first, last, fullName)
	case "OH", "TX", "FL", "GA", "PA", "NJ":
		return v.researchNote(region, q, first, last, fullName), nil
	default:
		return Result{Provider: "voter_registration", Status: "empty"}, nil
	}
}

// researchNote returns the state's voter-lookup URL as a research_link candidate
// for an operator to visit and search manually (URL never goes in Line1).
func (v *VoterRegistration) researchNote(region string, q Query, first, last, fullName string) Result {
	lookupURL := voterLookupURL(region, q, first, last)
	note := fmt.Sprintf("Voter registration (%s) — operator should visit and search", region)
	return Result{
		Provider: "voter_registration",
		Candidates: []Candidate{
			ResearchLink(lookupURL, q.City, region, fullName, "voter_registration", note),
		},
		Status:    "ok",
		CostCents: 0,
	}
}

// voterLookupURL returns the public voter-lookup portal URL for a state.
func voterLookupURL(region string, q Query, first, last string) string {
	county := CountyName(q.City, region)
	switch region {
	case "OH":
		base := "https://voterlookup.ohiosos.gov/voterlookup.aspx"
		if county != "" {
			return base + "?fn=" + url.QueryEscape(first) + "&ln=" + url.QueryEscape(last) + "&county=" + url.QueryEscape(county)
		}
		return base
	case "TX":
		return "https://www.sos.state.tx.us/elections/voter/votreglookup.shtml"
	case "FL":
		return "https://registration.elections.myflorida.com/CheckVoterStatus"
	case "GA":
		base := "https://www.mvp.sos.ga.gov/MVP/mvp.do"
		if county != "" {
			return base + "?firstName=" + url.QueryEscape(first) + "&lastName=" + url.QueryEscape(last) + "&county=" + url.QueryEscape(county)
		}
		return base
	case "PA":
		return "https://www.pavoterservices.pa.gov/Pages/VoterRegistrationStatus.aspx"
	case "NJ":
		return "https://voter.nj.gov/"
	default:
		return ""
	}
}

// searchNC fetches the NC voter search page and parses result rows.
// ponytail: regex parsing of HTML is fragile — if NCSBE changes markup this
// breaks silently (returns empty). Upgrade path: switch to x/net/html tokenizer.
func (v *VoterRegistration) searchNC(ctx context.Context, q Query, first, last, fullName string) (Result, error) {
	searchURL := fmt.Sprintf("https://vt.ncsbe.gov/voter_search_public/?firstName=%s&lastName=%s",
		url.QueryEscape(first), url.QueryEscape(last))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, searchURL, nil)
	if err != nil {
		return Result{Provider: "voter_registration", Status: "error", Error: err.Error()}, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	resp, err := v.client().Do(req)
	if err != nil {
		return Result{Provider: "voter_registration", Status: "error", Error: err.Error()}, err
	}
	defer resp.Body.Close()
	rawB, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	raw := string(rawB)

	if resp.StatusCode >= 300 {
		return Result{Provider: "voter_registration", Status: "error", RawJSON: truncate(raw, 4000),
			Error: fmt.Sprintf("nc voter http %d", resp.StatusCode)}, fmt.Errorf("nc voter http %d", resp.StatusCode)
	}

	cands := parseNCVoters(raw, q, fullName)
	if len(cands) == 0 {
		// Fall back to a research link so the URL isn't lost (never Line1).
		c := ResearchLink(searchURL, q.City, "NC", fullName, "voter_registration",
			"Voter registration (NC) — operator should visit and search")
		return Result{Provider: "voter_registration", Candidates: []Candidate{c}, Status: "ok",
			RawJSON: truncate(raw, 4000), CostCents: 0}, nil
	}
	return Result{Provider: "voter_registration", Candidates: cands, Status: "ok",
		RawJSON: truncate(raw, 4000), CostCents: 0}, nil
}

// --- NC HTML parsing ---------------------------------------------------------

// NC voter search results list voters in a table. Each row links to a detail
// page and shows name + address. We extract address lines and city/state/zip.
var (
	ncRowRe   = regexp.MustCompile(`(?s)<tr[^>]*class="[^"]*(?:voter|result|data)[^"]*"[^>]*>(.*?)</tr>`)
	ncAddrRe  = regexp.MustCompile(`(?s)<td[^>]*>(.*?)</td>`)
	ncStreetRe = regexp.MustCompile(`(\d{1,6}\s+[A-Za-z0-9\s.]+(?:St|Ave|Blvd|Rd|Dr|Ln|Ct|Way|Pl|Cir|Hwy)\.?)`)
	ncZipRe   = regexp.MustCompile(`\b(\d{5})(?:-\d{4})?\b`)
)

// parseNCVoters extracts up to 5 voter address candidates from the NC results HTML.
// ponytail: assumes address appears in a table cell near the name — NC's actual
// markup may differ; if no rows match we return empty (caller falls back to URL).
func parseNCVoters(html string, q Query, fullName string) []Candidate {
	var cands []Candidate
	seen := map[string]bool{}

	rows := ncRowRe.FindAllStringSubmatch(html, -1)
	if len(rows) == 0 {
		// Fallback: scan whole page for street+zip pairs near the name.
		return parseNCFallback(html, q, fullName)
	}

	for _, r := range rows {
		if len(cands) >= 5 {
			break
		}
		cellHTML := r[1]
		cells := ncAddrRe.FindAllStringSubmatch(cellHTML, -1)
		var street, city, zip string
		for _, c := range cells {
			text := stripTags(c[1])
			if m := ncStreetRe.FindString(text); m != "" && street == "" {
				street = strings.TrimSpace(m)
			}
			if m := ncZipRe.FindStringSubmatch(text); len(m) > 1 && zip == "" {
				zip = m[1]
			}
		}
		if street == "" {
			continue
		}
		key := street + zip
		if seen[key] {
			continue
		}
		seen[key] = true
		city = firstNonEmpty(q.City, city)
		cands = append(cands, Candidate{
			Line1:      street,
			City:       city,
			Region:     "NC",
			Postal:     zip,
			Country:    "US",
			Confidence: 0.55,
			Source:     "voter_registration",
			FullName:   fullName,
			Note:       "NC voter registration record — verify before mail.",
		})
	}
	return cands
}

// parseNCFallback does a whole-page scan for street+zip when no table rows match.
func parseNCFallback(html string, q Query, fullName string) []Candidate {
	var cands []Candidate
	seen := map[string]bool{}
	streets := ncStreetRe.FindAllString(html, -1)
	zips := ncZipRe.FindAllStringSubmatch(html, -1)
	for i, s := range streets {
		if i >= 5 {
			break
		}
		s = strings.TrimSpace(s)
		if seen[s] {
			continue
		}
		seen[s] = true
		zip := ""
		if i < len(zips) {
			zip = zips[i][1]
		}
		cands = append(cands, Candidate{
			Line1:      s,
			City:       q.City,
			Region:     "NC",
			Postal:     zip,
			Country:    "US",
			Confidence: 0.45,
			Source:     "voter_registration",
			FullName:   fullName,
			Note:       "NC voter registration record (fallback parse) — verify before mail.",
		})
	}
	return cands
}

// stripTags removes HTML tags and collapses whitespace.
var tagRe = regexp.MustCompile(`<[^>]+>`)

func stripTags(s string) string {
	s = tagRe.ReplaceAllString(s, " ")
	s = strings.ReplaceAll(s, "&nbsp;", " ")
	return strings.Join(strings.Fields(s), " ")
}
