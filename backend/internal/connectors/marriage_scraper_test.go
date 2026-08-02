package connectors

import (
	"testing"
	"time"
)

// Self-check: parsePublicSearchResults must extract marriage records from a
// representative publicsearch.us HTML results table, and ClassifyPortal must
// route each known portal family correctly. This is the smallest thing that
// fails if the regex parsing or portal classification breaks.
func TestParsePublicSearchResults(t *testing.T) {
	// Minimal HTML mimicking a publicsearch.us results table: one marriage
	// row and one non-marriage row (deed) that must be skipped.
	html := []byte(`<table><tbody>
<tr><td>01/15/2024</td><td>Marriage License</td><td>SMITH JOHN</td><td>DOE JANE</td><td>Book: 42 Page: 101</td></tr>
<tr><td>02/01/2024</td><td>Deed</td><td>JONES BOB</td><td>LEE MARY</td><td>Book: 5 Page: 9</td></tr>
</tbody></table>`)

	records := parsePublicSearchResults(html, "48113")
	if len(records) != 1 {
		t.Fatalf("expected 1 marriage record, got %d", len(records))
	}
	r := records[0]
	if r.CountyFIPS != "48113" {
		t.Errorf("CountyFIPS = %q, want 48113", r.CountyFIPS)
	}
	if r.Party1Name == "" || r.Party2Name == "" {
		t.Errorf("expected non-empty party names, got %q & %q", r.Party1Name, r.Party2Name)
	}
	if r.BookPage != "42/101" {
		t.Errorf("BookPage = %q, want 42/101", r.BookPage)
	}
	want := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	if !r.MarriageDate.Equal(want) {
		t.Errorf("MarriageDate = %v, want %v", r.MarriageDate, want)
	}
}

func TestClassifyPortal(t *testing.T) {
	cases := []struct {
		url  string
		want PortalType
	}{
		{"https://dallas.tx.publicsearch.us/", PortalPublicSearch},
		{"https://collin.tx.publicsearch.us/", PortalPublicSearch},
		{"https://arapahoe.co.publicsearch.us/", PortalPublicSearch},
		{"https://countyfusion3.kofiletech.us/countyweb/loginDisplay.action?countyname=Denver", PortalKofile},
		{"https://boulder.co.ds.kofile.systems/", PortalKofile},
		{"https://recordsearch.kingcounty.gov/LandmarkWeb/search/index", PortalKofile},
		{"https://www.oscn.net/dockets/Search.aspx", PortalOSCN},
		{"https://moms.mn.gov/Search", PortalMOMS},
		{"https://secure.tncountyclerk.com/marriagelookup/index.php?countylist=19", PortalTNCountyClerk},
		{"https://example.com/unknown", PortalUnknown},
	}
	for _, c := range cases {
		got := ClassifyPortal(c.url)
		if got != c.want {
			t.Errorf("ClassifyPortal(%q) = %q, want %q", c.url, got, c.want)
		}
	}
}

func TestBuildPublicSearchURL(t *testing.T) {
	q := MarriageSearchQuery{
		SearchURL:  "https://dallas.tx.publicsearch.us/",
		NamePrefix: "SMITH",
		DateFrom:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		DateTo:     time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC),
	}
	got := buildPublicSearchURL(q)
	// Should contain the marriage filter and formatted date params.
	for _, want := range []string{"searchType=marriage", "lastName=SMITH", "fromDate=01%2F01%2F2024", "toDate=12%2F31%2F2024"} {
		if !contains(got, want) {
			t.Errorf("buildPublicSearchURL: %q missing %q", got, want)
		}
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
