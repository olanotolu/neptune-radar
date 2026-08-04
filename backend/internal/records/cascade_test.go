package records

import (
	"context"
	"strings"
	"testing"
)

func TestMultiCascade_FreeProvidersIncluded(t *testing.T) {
	m := NewMulti()
	// Free providers should always include property + voter; Google/TPS optional by env
	if len(m.Free) < 2 {
		t.Fatalf("expected at least 2 free providers, got %d", len(m.Free))
	}
	foundProp, foundVoter := false, false
	for _, p := range m.Free {
		switch p.Name() {
		case "county_property":
			foundProp = true
		case "voter_registration":
			foundVoter = true
		}
	}
	if !foundProp {
		t.Error("county_property provider not in free cascade")
	}
	if !foundVoter {
		t.Error("voter_registration provider not in free cascade")
	}
}

func TestMultiCascade_FallbackToHeuristic(t *testing.T) {
	m := NewMulti()
	// With no API keys and no Bright Data, should still return heuristic candidates
	q := Query{
		FirstName: "Jane",
		LastName:  "Doe",
		City:      "Columbus",
		Region:    "OH",
	}
	res, err := m.Search(context.Background(), q)
	if err != nil {
		t.Fatalf("cascade should not error: %v", err)
	}
	if len(res.Candidates) == 0 {
		t.Error("cascade should return at least heuristic candidates")
	}
}

func TestHasStreetCandidates(t *testing.T) {
	if hasStreetCandidates(nil) {
		t.Error("nil should not have street candidates")
	}
	if hasStreetCandidates([]Candidate{{City: "Columbus"}}) {
		t.Error("candidate without Line1 should not count as street-level")
	}
	if !hasStreetCandidates([]Candidate{{Line1: "123 Main St", City: "Columbus"}}) {
		t.Error("candidate with Line1 should count as street-level")
	}
}

func TestIsRealStreet_RejectsURLs(t *testing.T) {
	urls := []string{
		"https://www.truepeoplesearch.com/results?name=Jane",
		"http://voterlookup.ohiosos.gov/voterlookup.aspx",
		"https://example.com/path",
	}
	for _, u := range urls {
		if IsRealStreet(u) {
			t.Errorf("URL must not count as street: %s", u)
		}
		if hasStreetCandidates([]Candidate{{Line1: u, City: "Columbus"}}) {
			t.Errorf("hasStreetCandidates must reject URL Line1: %s", u)
		}
	}
	// Research link kind never counts even with fake line1
	if hasStreetCandidates([]Candidate{{
		Kind: KindResearchLink, URL: "https://example.com", Line1: "123 Main St",
	}}) {
		t.Error("research_link kind must not count as street")
	}
	if !IsRealStreet("123 Main St") {
		t.Error("123 Main St should be a real street")
	}
	if !IsRealStreet("456 N High Street") {
		t.Error("456 N High Street should be a real street")
	}
	if IsRealStreet("") {
		t.Error("empty should not be street")
	}
	if IsRealStreet("Columbus") {
		t.Error("city name alone should not be street")
	}
}

func TestNormalizeCandidates_PromotesURLOutOfLine1(t *testing.T) {
	in := []Candidate{{
		Line1: "https://www.truepeoplesearch.com/results?name=x",
		City:  "Columbus", Region: "OH", Source: "truepeoplesearch", Confidence: 0.3,
	}}
	out := normalizeCandidates(in)
	if len(out) != 1 {
		t.Fatalf("expected 1, got %d", len(out))
	}
	if out[0].Line1 != "" {
		t.Errorf("Line1 should be cleared, got %q", out[0].Line1)
	}
	if out[0].Kind != KindResearchLink {
		t.Errorf("kind want research_link, got %s", out[0].Kind)
	}
	if !strings.HasPrefix(out[0].URL, "https://") {
		t.Errorf("URL should be set, got %q", out[0].URL)
	}
	if hasStreetCandidates(out) {
		t.Error("normalized research link must not count as street")
	}
}

func TestResearchLink_NoLine1(t *testing.T) {
	c := ResearchLink("https://example.com/search", "Columbus", "OH", "Jane Doe", "voter_registration", "visit me")
	if c.Line1 != "" {
		t.Errorf("ResearchLink Line1 must be empty, got %q", c.Line1)
	}
	if c.URL != "https://example.com/search" {
		t.Errorf("URL: %s", c.URL)
	}
	if c.Kind != KindResearchLink {
		t.Errorf("kind: %s", c.Kind)
	}
	if hasStreetCandidates([]Candidate{c}) {
		t.Error("ResearchLink must not count as street")
	}
}

func TestMultiSearch_MergesProviders(t *testing.T) {
	// Primary returns only locality; free returns a real street — must surface street.
	m := &Multi{
		Primary: &stubProvider{name: "paid_empty", cands: []Candidate{
			{City: "Columbus", Region: "OH", Confidence: 0.35, Source: "paid_empty", Kind: KindLocality},
		}},
		Free: []Provider{
			&stubProvider{name: "free_street", cands: []Candidate{
				{Line1: "123 Main St", City: "Columbus", Region: "OH", Postal: "43215",
					Confidence: 0.70, Source: "free_street", Kind: KindStreet},
			}},
			&stubProvider{name: "free_link", cands: []Candidate{
				ResearchLink("https://example.com/voter", "Columbus", "OH", "Jane Doe", "free_link", "check voter"),
			}},
		},
		Fallback: &Heuristic{},
	}
	res, err := m.Search(context.Background(), Query{FirstName: "Jane", LastName: "Doe", City: "Columbus", Region: "OH"})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !hasStreetCandidates(res.Candidates) {
		t.Fatalf("expected merged street candidate, got %+v", res.Candidates)
	}
	// Street should rank first
	if !IsRealStreet(res.Candidates[0].Line1) {
		t.Errorf("top candidate should be street, got %+v", res.Candidates[0])
	}
	// Research link should still be present
	foundLink := false
	for _, c := range res.Candidates {
		if c.Kind == KindResearchLink || c.URL != "" {
			foundLink = true
			if c.Line1 != "" {
				t.Errorf("research link must not have Line1: %+v", c)
			}
		}
	}
	if !foundLink {
		t.Error("expected research link in merged results")
	}
	if !strings.Contains(res.Provider, "paid_empty") || !strings.Contains(res.Provider, "free_street") {
		t.Errorf("provider should list both: %s", res.Provider)
	}
}

func TestMultiSearch_DoesNotStopOnResearchURL(t *testing.T) {
	// Old bug: URL in Line1 stopped cascade. Free1 returns bad URL-as-Line1;
	// free2 returns real street — after normalize, free2 must appear.
	m := &Multi{
		Primary: &stubProvider{name: "none", cands: nil},
		Free: []Provider{
			&stubProvider{name: "url_bug", cands: []Candidate{
				{Line1: "https://www.truepeoplesearch.com/results?name=Jane", City: "Columbus", Region: "OH",
					Confidence: 0.30, Source: "url_bug"},
			}},
			&stubProvider{name: "real", cands: []Candidate{
				{Line1: "99 Broad St", City: "Columbus", Region: "OH", Confidence: 0.72, Source: "real"},
			}},
		},
		Fallback: &Heuristic{},
	}
	res, err := m.Search(context.Background(), Query{FirstName: "Jane", LastName: "Doe", City: "Columbus", Region: "OH"})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	found := false
	for _, c := range res.Candidates {
		if strings.Contains(c.Line1, "99 Broad") {
			found = true
		}
		if strings.HasPrefix(strings.ToLower(c.Line1), "http") {
			t.Errorf("Line1 must never be URL after merge: %q", c.Line1)
		}
	}
	if !found {
		t.Fatalf("expected real street after research URL, got %+v", res.Candidates)
	}
}

func TestMultiSearch_EarlyStopHighConf(t *testing.T) {
	// High-confidence street from primary should skip later free providers
	// (cost control). Free stub would panic if called — use counting stub.
	free := &countProvider{name: "free_skip"}
	m := &Multi{
		Primary: &stubProvider{name: "hi", cands: []Candidate{
			{Line1: "1 High St", City: "Columbus", Region: "OH", Confidence: 0.90, Source: "hi", Kind: KindStreet},
		}},
		Free:     []Provider{free},
		Fallback: &Heuristic{},
	}
	res, err := m.Search(context.Background(), Query{FirstName: "Jane", LastName: "Doe", City: "Columbus", Region: "OH"})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if free.calls != 0 {
		t.Errorf("expected free tier skipped on high-conf street, calls=%d", free.calls)
	}
	if len(res.Candidates) == 0 || !IsRealStreet(res.Candidates[0].Line1) {
		t.Fatalf("expected street: %+v", res.Candidates)
	}
}

func TestDedupeAndRank(t *testing.T) {
	cands := []Candidate{
		ResearchLink("https://a.example", "Columbus", "OH", "", "a", "link"),
		{Line1: "10 Oak Ave", City: "Columbus", Region: "OH", Confidence: 0.55, Source: "pdl", Kind: KindStreet},
		{Line1: "10 Oak Ave", City: "Columbus", Region: "OH", Confidence: 0.70, Source: "tps", Kind: KindStreet},
		{City: "Columbus", Region: "OH", Confidence: 0.40, Source: "heuristic", Kind: KindLocality},
	}
	cands = dedupeCandidates(cands)
	rankCandidates(cands)
	if len(cands) != 3 {
		t.Fatalf("dedupe expected 3, got %d: %+v", len(cands), cands)
	}
	if cands[0].Confidence != 0.70 {
		t.Errorf("top should be higher conf street, got conf=%v source=%s", cands[0].Confidence, cands[0].Source)
	}
	if !IsRealStreet(cands[0].Line1) {
		t.Error("first should be street")
	}
}

func TestPropertyRecords_Available(t *testing.T) {
	p := &PropertyRecords{}
	if !p.Available() {
		t.Error("property records should always be available")
	}
	if p.Name() != "county_property" {
		t.Errorf("expected county_property, got %s", p.Name())
	}
}

func TestVoterRegistration_Available(t *testing.T) {
	v := &VoterRegistration{}
	if !v.Available() {
		t.Error("voter registration should always be available")
	}
	if v.Name() != "voter_registration" {
		t.Fatalf("expected voter_registration, got %s", v.Name())
	}
}

func TestVoterRegistration_EmptyRegion(t *testing.T) {
	v := &VoterRegistration{}
	res, err := v.Search(context.Background(), Query{FirstName: "Jane", LastName: "Doe"})
	if err != nil {
		t.Fatalf("should not error on empty region: %v", err)
	}
	if res.Status != "empty" {
		t.Errorf("expected empty status for no region, got %s", res.Status)
	}
}

func TestVoterRegistration_ResearchNoteNoStreet(t *testing.T) {
	v := &VoterRegistration{}
	res, err := v.Search(context.Background(), Query{
		FirstName: "Jane", LastName: "Doe", City: "Columbus", Region: "OH",
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(res.Candidates) == 0 {
		t.Fatal("expected research link candidate")
	}
	if hasStreetCandidates(res.Candidates) {
		t.Fatalf("OH voter research must not count as street: %+v", res.Candidates)
	}
	if res.Candidates[0].URL == "" {
		t.Error("expected URL on research link")
	}
	if res.Candidates[0].Line1 != "" {
		t.Errorf("Line1 must be empty, got %q", res.Candidates[0].Line1)
	}
}

func TestNewMulti_WiresMelissaWhenEnv(t *testing.T) {
	t.Setenv("MELISSA_LICENSE_KEY", "test-key")
	t.Setenv("TRESTLE_API_KEY", "")
	t.Setenv("PDL_API_KEY", "")
	t.Setenv("CLEANLIST_API_KEY", "")
	t.Setenv("GOOGLE_SEARCH_API_KEY", "")
	m := NewMulti()
	// Melissa should be primary when only MELISSA is set
	if m.Primary == nil || m.Primary.Name() != "melissa" {
		t.Fatalf("expected melissa primary, got %v", m.Primary)
	}
}

func TestNewMulti_GoogleInFreeWhenPaidPrimary(t *testing.T) {
	t.Setenv("TRESTLE_API_KEY", "t-key")
	t.Setenv("GOOGLE_SEARCH_API_KEY", "g-key")
	t.Setenv("GOOGLE_SEARCH_CX", "cx")
	t.Setenv("PDL_API_KEY", "")
	t.Setenv("MELISSA_LICENSE_KEY", "")
	t.Setenv("CLEANLIST_API_KEY", "")
	t.Setenv("BRIGHTDATA_API_KEY", "")
	m := NewMulti()
	if m.Primary == nil || m.Primary.Name() != "trestle" {
		t.Fatalf("primary trestle, got %v", m.Primary)
	}
	found := false
	for _, p := range m.Free {
		if p.Name() == "google_search" {
			found = true
		}
	}
	if !found {
		t.Error("google_search should be in free tier when paid primary is set")
	}
}

// --- stubs -----------------------------------------------------------------

type stubProvider struct {
	name  string
	cands []Candidate
	err   error
}

func (s *stubProvider) Name() string     { return s.name }
func (s *stubProvider) Available() bool  { return true }
func (s *stubProvider) Search(ctx context.Context, q Query) (Result, error) {
	_ = ctx
	_ = q
	st := "ok"
	if len(s.cands) == 0 {
		st = "empty"
	}
	return Result{Provider: s.name, Candidates: s.cands, Status: st}, s.err
}

type countProvider struct {
	name  string
	calls int
}

func (c *countProvider) Name() string    { return c.name }
func (c *countProvider) Available() bool { return true }
func (c *countProvider) Search(ctx context.Context, q Query) (Result, error) {
	c.calls++
	return Result{Provider: c.name, Status: "empty"}, nil
}
