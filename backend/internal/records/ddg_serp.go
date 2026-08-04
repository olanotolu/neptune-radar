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

// DDGSerp is a free hunter: DuckDuckGo HTML SERP for "Name City ST address".
// Often only yields research links (Spokeo/Whitepages/TPS), but occasionally
// snippets contain street-level addresses. Never invents streets.
type DDGSerp struct {
	Client *http.Client
}

func (d *DDGSerp) Name() string    { return "ddg_serp" }
func (d *DDGSerp) Available() bool { return true }

func (d *DDGSerp) client() *http.Client {
	if d.Client != nil {
		return d.Client
	}
	return &http.Client{Timeout: 25 * time.Second}
}

func (d *DDGSerp) Search(ctx context.Context, q Query) (Result, error) {
	name := strings.TrimSpace(q.FirstName + " " + q.LastName)
	if name == "" || name == " " {
		return Result{Provider: "ddg_serp", Status: "empty"}, nil
	}
	loc := strings.TrimSpace(strings.Trim(q.City+" "+q.Region, " "))
	queries := []string{
		fmt.Sprintf(`"%s" %s address`, name, loc),
		fmt.Sprintf(`%s %s phone address`, name, loc),
		fmt.Sprintf(`"%s" "%s"`, name, q.City),
	}
	var cands []Candidate
	var rawParts []string
	seenURL := map[string]bool{}
	for _, qq := range queries {
		if err := ctx.Err(); err != nil {
			break
		}
		html, err := d.fetch(ctx, qq)
		if err != nil {
			rawParts = append(rawParts, "err:"+err.Error())
			continue
		}
		rawParts = append(rawParts, truncate(html, 1500))
		// Streets only if last name (or full name) appears near the match — cuts ad noise.
		plain := ddgStripTags(html)
		last := strings.ToLower(strings.TrimSpace(q.LastName))
		for _, m := range fullAddrRe.FindAllStringIndex(plain, 8) {
			snippet := windowAround(plain, m[0], m[1], 120)
			if last != "" && !strings.Contains(strings.ToLower(snippet), last) {
				continue
			}
			addr := plain[m[0]:m[1]]
			if c, ok := parseUSAddressLine(addr, name, q); ok && IsRealStreet(c.Line1) {
				c.Source = "ddg_serp"
				c.Note = "Parsed from DuckDuckGo SERP near name — still verify before mail."
				c.Confidence = 0.62
				cands = append(cands, c)
			}
		}
		for _, m := range simpleStreetRe.FindAllStringIndex(plain, 10) {
			snippet := windowAround(plain, m[0], m[1], 100)
			if last != "" && !strings.Contains(strings.ToLower(snippet), last) {
				continue
			}
			line := strings.TrimSpace(plain[m[0]:m[1]])
			if !IsRealStreet(line) {
				continue
			}
			cands = append(cands, Candidate{
				Line1: line, City: q.City, Region: q.Region, Country: "US",
				Kind: KindStreet, Source: "ddg_serp", FullName: name,
				Confidence: 0.52,
				Note:       "Street near name in SERP — city assumed from query; verify.",
			})
		}
		// Research links from results
		for _, u := range extractResultURLs(html) {
			if seenURL[u] {
				continue
			}
			seenURL[u] = true
			host := hostOf(u)
			if !isPeopleSearchHost(host) {
				continue
			}
			cands = append(cands, ResearchLink(u, q.City, q.Region, name, "ddg_serp",
				fmt.Sprintf("SERP hit on %s — open and paste street", host)))
		}
	}
	// Always add curated research URLs aimed at this name+city
	if name != "" {
		for _, u := range curatedResearchURLs(name, q.City, q.Region) {
			if seenURL[u] {
				continue
			}
			seenURL[u] = true
			cands = append(cands, ResearchLink(u, q.City, q.Region, name, "ddg_serp",
				"Curated people-search URL for operator paste"))
		}
	}
	cands = dedupeCandidates(normalizeCandidates(cands))
	rankCandidates(cands)
	st := "ok"
	if len(cands) == 0 {
		st = "empty"
	}
	return Result{
		Provider: "ddg_serp", Status: st, Candidates: cands,
		RawJSON: strings.Join(rawParts, "\n---\n"), CostCents: 0,
	}, nil
}

func (d *DDGSerp) fetch(ctx context.Context, q string) (string, error) {
	u := "https://html.duckduckgo.com/html/?q=" + url.QueryEscape(q)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36")
	resp, err := d.client().Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("ddg http %d", resp.StatusCode)
	}
	return string(b), nil
}

var (
	fullAddrRe      = regexp.MustCompile(`(?i)\b\d{1,6}\s+[A-Za-z0-9.'\-]+(?:\s+[A-Za-z0-9.'\-]+){0,4}\s+(?:Street|St|Avenue|Ave|Road|Rd|Drive|Dr|Lane|Ln|Court|Ct|Boulevard|Blvd|Way|Place|Pl)\.?\s*,\s*[A-Za-z .]+,\s*[A-Z]{2}\s+\d{5}`)
	simpleStreetRe  = regexp.MustCompile(`(?i)\b\d{1,6}\s+[A-Za-z][A-Za-z0-9.'\-]*(?:\s+[A-Za-z0-9.'\-]+){0,3}\s+(?:Street|St|Avenue|Ave|Road|Rd|Drive|Dr|Lane|Ln|Court|Ct|Boulevard|Blvd|Way)\b`)
	resultHrefRe    = regexp.MustCompile(`uddg=([^&"]+)`)
	resultHrefRe2   = regexp.MustCompile(`class="result__a"[^>]+href="(https?://[^"]+)"`)
)

func ddgStripTags(s string) string {
	re := regexp.MustCompile(`<[^>]+>`)
	return re.ReplaceAllString(s, " ")
}

func windowAround(s string, start, end, pad int) string {
	lo := start - pad
	if lo < 0 {
		lo = 0
	}
	hi := end + pad
	if hi > len(s) {
		hi = len(s)
	}
	return s[lo:hi]
}

func extractResultURLs(html string) []string {
	var out []string
	seen := map[string]bool{}
	for _, m := range resultHrefRe.FindAllStringSubmatch(html, -1) {
		u, err := url.QueryUnescape(m[1])
		if err != nil || !strings.HasPrefix(u, "http") {
			continue
		}
		if seen[u] {
			continue
		}
		seen[u] = true
		out = append(out, u)
	}
	for _, m := range resultHrefRe2.FindAllStringSubmatch(html, -1) {
		u := m[1]
		if seen[u] {
			continue
		}
		seen[u] = true
		out = append(out, u)
	}
	return out
}

func hostOf(u string) string {
	p, err := url.Parse(u)
	if err != nil {
		return ""
	}
	return strings.ToLower(p.Host)
}

func isPeopleSearchHost(h string) bool {
	h = strings.TrimPrefix(h, "www.")
	for _, ok := range []string{
		"truepeoplesearch.com", "fastpeoplesearch.com", "whitepages.com",
		"spokeo.com", "beenverified.com", "anywho.com", "intelius.com",
		"veripages.com", "radaris.com", "thatsthem.com", "cyberbackgroundchecks.com",
		"familytreenow.com", "peoplefinders.com", "nuwber.com",
	} {
		if h == ok || strings.HasSuffix(h, "."+ok) {
			return true
		}
	}
	return false
}

func curatedResearchURLs(name, city, region string) []string {
	n := url.QueryEscape(name)
	loc := url.QueryEscape(strings.Trim(city+", "+region, ", "))
	return []string{
		"https://www.truepeoplesearch.com/results?name=" + n + "&citystatezip=" + loc,
		"https://www.fastpeoplesearch.com/name/" + strings.ReplaceAll(strings.ToLower(name), " ", "-"),
		"https://www.whitepages.com/name/" + strings.ReplaceAll(name, " ", "-"),
		"https://www.spokeo.com/" + strings.ReplaceAll(name, " ", "-") + "?q=" + n,
		"https://www.google.com/search?q=" + url.QueryEscape(fmt.Sprintf(`"%s" %s %s address`, name, city, region)),
	}
}
