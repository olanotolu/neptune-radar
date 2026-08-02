package connectors

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

// MarriageSearchQuery is the input to a marriage-record scrape. CountyFIPS
// identifies the county (matches geo.Counties); NamePrefix narrows results to
// parties whose names begin with the given string (optional — empty means all);
// DateFrom/DateTo bound the marriage date range (zero-value = unbounded on
// that side).
type MarriageSearchQuery struct {
	CountyFIPS string
	NamePrefix string
	DateFrom   time.Time
	DateTo     time.Time
	SearchURL  string // the portal base URL from the govSource pack
}

// MarriageRecord is one scraped marriage record. BookPage is set when the
// portal exposes the clerk's book/page reference; otherwise empty.
type MarriageRecord struct {
	Party1Name   string
	Party2Name   string
	MarriageDate time.Time
	CountyFIPS   string
	BookPage     string
}

// PortalType identifies which county-portal family a SearchURL belongs to.
type PortalType string

const (
	PortalPublicSearch  PortalType = "publicsearch"  // publicsearch.us (TX, FL, CO, NM, ...)
	PortalKofile        PortalType = "kofile"        // Kofile / CountyFusion / LandmarkWeb (CO, WA, ...)
	PortalOSCN          PortalType = "oscn"          // Oklahoma Supreme Court Network docket search
	PortalMOMS          PortalType = "moms"          // Minnesota Official Marriage System
	PortalTNCountyClerk PortalType = "tncountyclerk" // Tennessee statewide marriage lookup
	PortalUnknown       PortalType = "unknown"
)

// ClassifyPortal inspects a SearchURL and returns the portal family it
// belongs to. This drives which scraper implementation handles the query.
func ClassifyPortal(searchURL string) PortalType {
	u, err := url.Parse(searchURL)
	if err != nil {
		return PortalUnknown
	}
	host := strings.ToLower(u.Hostname())
	path := strings.ToLower(u.Path)
	switch {
	case strings.Contains(host, "publicsearch.us"):
		return PortalPublicSearch
	case strings.Contains(host, "kofiletech.us") || strings.Contains(host, "kofile.systems") || strings.Contains(host, "kofilequicklinks.com") || strings.Contains(host, "landmarkweb") || strings.Contains(path, "landmarkweb"):
		return PortalKofile
	case strings.Contains(host, "oscn.net"):
		return PortalOSCN
	case strings.Contains(host, "moms.mn.gov"):
		return PortalMOMS
	case strings.Contains(host, "tncountyclerk.com"):
		return PortalTNCountyClerk
	default:
		return PortalUnknown
	}
}

// MarriageRecordScraper scrapes marriage-record search results from county
// portals. Unlike HTTPHealthConnector (which only checks liveness), this
// connector actually parses result pages and returns structured records.
//
// Only the publicsearch.us portal family is implemented today; the other four
// portal types are stubbed and return ErrNotImplemented so callers can see
// which portals still need work without a silent empty-result bug.
type MarriageRecordScraper struct {
	Client *http.Client
	// RequestDelay is the polite delay between HTTP requests to the same
	// portal host. Default 2s — county servers are not built for scraping.
	RequestDelay time.Duration
}

func NewMarriageRecordScraper() *MarriageRecordScraper {
	return &MarriageRecordScraper{
		Client:       &http.Client{Timeout: 30 * time.Second},
		RequestDelay: 2 * time.Second,
	}
}

// ErrNotImplemented is returned by Scrape for portal types that are stubbed.
var ErrNotImplemented = fmt.Errorf("marriage scraper: portal type not yet implemented")

// Scrape dispatches to the portal-specific scraper based on the SearchURL in
// the query. Returns marriage records matching the query, or
// ErrNotImplemented for portal types that are still stubs.
func (s *MarriageRecordScraper) Scrape(ctx context.Context, query MarriageSearchQuery) ([]MarriageRecord, error) {
	pt := ClassifyPortal(query.SearchURL)
	switch pt {
	case PortalPublicSearch:
		return s.scrapePublicSearch(ctx, query)
	case PortalKofile:
		return nil, fmt.Errorf("%w: kofile/landmarkweb", ErrNotImplemented)
	case PortalOSCN:
		return nil, fmt.Errorf("%w: oscn", ErrNotImplemented)
	case PortalMOMS:
		return nil, fmt.Errorf("%w: moms", ErrNotImplemented)
	case PortalTNCountyClerk:
		return nil, fmt.Errorf("%w: tncountyclerk", ErrNotImplemented)
	default:
		return nil, fmt.Errorf("marriage scraper: unrecognized portal for %s", query.SearchURL)
	}
}

// CheckHealth lets MarriageRecordScraper satisfy SourceConnector — a simple
// reachability GET delegated to HTTPHealthConnector. The real work is Scrape.
func (s *MarriageRecordScraper) CheckHealth(ctx context.Context, endpointURL string) CheckResult {
	return NewHTTPHealthConnector().CheckHealth(ctx, endpointURL)
}

// --- publicsearch.us scraper ------------------------------------------------
//
// publicsearch.us portals (e.g. dallas.tx.publicsearch.us,
// collin.tx.publicsearch.us, arapahoe.co.publicsearch.us) host an HTML search
// form. The search results page is an HTML table where each row is one
// recorded document. Marriage records appear with a document type like
// "Marriage License" and two party names in the grantor/grantee columns.
//
// The portal accepts GET query parameters for name and date filtering. The
// exact parameter names vary slightly by deployment, but the common ones are:
//
//	?searchType=marriage&lastName=SMITH&fromDate=01/01/2020&toDate=12/31/2020
//
// ponytail: The result-table parsing below uses regex on the HTML rather than
// a full DOM parser (golang.org/x/net/html is not in go.mod and the rules say
// stdlib + existing deps only). This is fragile to markup changes — the
// upgrade path is to add x/net/html or a CSS-selector lib when the portal
// changes its table structure.

// publicSearchRow captures one <tr> from the results table.
var publicSearchRowRe = regexp.MustCompile(`(?is)<tr[^>]*>(.*?)</tr>`)
var publicSearchCellRe = regexp.MustCompile(`(?is)<td[^>]*>(.*?)</td>`)
var publicSearchDocTypeRe = regexp.MustCompile(`(?i)marriage`)
var publicSearchBookPageRe = regexp.MustCompile(`(?i)(?:book|bk)\s*[:#]?\s*(\S+)\s*(?:page|pg)\s*[:#]?\s*(\S+)`)

func (s *MarriageRecordScraper) scrapePublicSearch(ctx context.Context, query MarriageSearchQuery) ([]MarriageRecord, error) {
	searchURL := buildPublicSearchURL(query)

	body, err := s.fetchPage(ctx, searchURL)
	if err != nil {
		return nil, fmt.Errorf("publicsearch fetch: %w", err)
	}

	return parsePublicSearchResults(body, query.CountyFIPS), nil
}

// buildPublicSearchURL constructs the search query URL for a publicsearch.us
// portal. The portal accepts a marriage document-type filter and optional
// name/date parameters.
func buildPublicSearchURL(query MarriageSearchQuery) string {
	u, err := url.Parse(query.SearchURL)
	if err != nil {
		return query.SearchURL
	}
	q := u.Query()
	q.Set("searchType", "marriage")
	if query.NamePrefix != "" {
		q.Set("lastName", query.NamePrefix)
	}
	if !query.DateFrom.IsZero() {
		q.Set("fromDate", query.DateFrom.Format("01/02/2006"))
	}
	if !query.DateTo.IsZero() {
		q.Set("toDate", query.DateTo.Format("01/02/2006"))
	}
	u.RawQuery = q.Encode()
	return u.String()
}

// parsePublicSearchResults extracts marriage records from a publicsearch.us
// HTML results page. Each result-table row is inspected; rows whose document
// type column mentions "marriage" are parsed into MarriageRecords.
func parsePublicSearchResults(html []byte, countyFIPS string) []MarriageRecord {
	var records []MarriageRecord
	rows := publicSearchRowRe.FindAllSubmatch(html, -1)
	for _, rowMatch := range rows {
		rowHTML := rowMatch[1]
		cells := publicSearchCellRe.FindAllSubmatch(rowHTML, -1)
		if len(cells) < 4 {
			continue
		}
		// Strip HTML tags from each cell for plain-text comparison.
		cellTexts := make([]string, len(cells))
		hasMarriage := false
		for i, c := range cells {
			cellTexts[i] = stripTags(string(c[1]))
			if publicSearchDocTypeRe.MatchString(cellTexts[i]) {
				hasMarriage = true
			}
		}
		if !hasMarriage {
			continue
		}

		rec := MarriageRecord{CountyFIPS: countyFIPS}

		// publicsearch.us result tables typically have columns in this order:
		// [date, doc_type, grantor, grantee, book/page, ...]. The exact
		// column order varies by deployment, so we scan for recognizable
		// patterns rather than assuming fixed positions.
		for _, text := range cellTexts {
			trimmed := strings.TrimSpace(text)
			if rec.MarriageDate.IsZero() {
				if d, ok := parseFlexibleDate(trimmed); ok {
					rec.MarriageDate = d
					continue
				}
			}
			if rec.BookPage == "" {
				if bp := publicSearchBookPageRe.FindStringSubmatch(trimmed); bp != nil {
					rec.BookPage = bp[1] + "/" + bp[2]
					continue
				}
			}
			// The first two non-date, non-bookpage, non-doctype cells that
			// look like person names become Party1/Party2. Don't break after
			// Party2 — later cells may still hold the book/page reference.
			if looksLikeName(trimmed) && !publicSearchDocTypeRe.MatchString(trimmed) {
				if rec.Party1Name == "" {
					rec.Party1Name = trimmed
				} else if rec.Party2Name == "" {
					rec.Party2Name = trimmed
				}
			}
		}

		// Only keep rows where we found at least one party name — a marriage
		// row with no names is a header or a malformed row.
		if rec.Party1Name != "" {
			records = append(records, rec)
		}
	}
	return records
}

// fetchPage performs a polite GET (with the configured delay) and returns the
// response body. The delay is applied before the request to avoid hammering
// county servers with back-to-back calls.
func (s *MarriageRecordScraper) fetchPage(ctx context.Context, pageURL string) ([]byte, error) {
	if s.RequestDelay > 0 {
		select {
		case <-time.After(s.RequestDelay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pageURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "NeptuneRadar-MarriageScraper/1.0 (+https://meetneptune.com; marriage record enumeration)")
	resp, err := s.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d from %s", resp.StatusCode, pageURL)
	}
	return io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024))
}

// --- helpers ----------------------------------------------------------------

var tagRe = regexp.MustCompile(`<[^>]*>`)
var wsRe = regexp.MustCompile(`\s+`)

func stripTags(s string) string {
	s = tagRe.ReplaceAllString(s, " ")
	s = strings.ReplaceAll(s, "&nbsp;", " ")
	s = strings.ReplaceAll(s, "&amp;", "&")
	s = wsRe.ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}

// parseFlexibleDate tries common US date formats found in county portals.
func parseFlexibleDate(s string) (time.Time, bool) {
	s = strings.TrimSpace(s)
	for _, layout := range []string{
		"01/02/2006",
		"1/2/2006",
		"01/02/06",
		"1/2/06",
		"2006-01-02",
		"Jan 2, 2006",
		"January 2, 2006",
	} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

// looksLikeName is a naive heuristic: a non-empty string with letters and no
// digits, <= 80 chars. Good enough to skip date/doc-type/book-page cells.
// ponytail: ceiling — will misclassify unusual single-word grantor names like
// "SMITH" vs a doc-type abbreviation; upgrade to a name DB lookup if precision
// matters.
func looksLikeName(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" || len(s) > 80 {
		return false
	}
	hasLetter := false
	for _, r := range s {
		if r >= '0' && r <= '9' {
			return false
		}
		if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
			hasLetter = true
		}
	}
	return hasLetter
}
