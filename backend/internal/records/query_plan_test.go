package records

import (
	"context"
	"strings"
	"testing"
)

func TestLocationVariants_OrderedUnique(t *testing.T) {
	q := Query{
		City: "Columbus", Region: "OH",
		AccountCityA: "Cleveland", AccountRegionA: "OH",
		AccountCityB: "Columbus", AccountRegionB: "OH", // dup of kit
		VendorCity: "Dayton", VendorState: "OH",
		PostLocation: "The Joseph Hotel, Columbus OH",
	}
	locs := LocationVariants(q)
	if len(locs) < 3 {
		t.Fatalf("expected at least kit+accountA+vendor, got %+v", locs)
	}
	if !strings.EqualFold(locs[0].City, "Columbus") || locs[0].Source != "kit" {
		t.Errorf("first should be kit Columbus, got %+v", locs[0])
	}
	// Cleveland from account A
	foundCle, foundDay := false, false
	for _, l := range locs {
		if strings.EqualFold(l.City, "Cleveland") {
			foundCle = true
		}
		if strings.EqualFold(l.City, "Dayton") {
			foundDay = true
		}
	}
	if !foundCle {
		t.Error("expected Cleveland from account_a")
	}
	if !foundDay {
		t.Error("expected Dayton from vendor")
	}
	// No duplicate Columbus
	nCol := 0
	for _, l := range locs {
		if strings.EqualFold(l.City, "Columbus") {
			nCol++
		}
	}
	if nCol != 1 {
		t.Errorf("Columbus should appear once, got %d in %+v", nCol, locs)
	}
}

func TestNameVariants_MarriedAndDedupe(t *testing.T) {
	ns := NameVariants("Jane", "Doe", "jane", "John", "Smith", "john")
	// Primary A, primary B, then married: Jane Doe, John Smith, Jane Smith, John Doe
	if len(ns) != 4 {
		t.Fatalf("want 4 variants, got %d: %+v", len(ns), ns)
	}
	if ns[0].Last != "Doe" || ns[1].Last != "Smith" {
		t.Fatalf("primaries first: got %+v", ns)
	}
	// Same last names — no married variants
	ns2 := NameVariants("Jane", "Doe", "j", "John", "Doe", "jo")
	if len(ns2) != 2 {
		t.Fatalf("same last: want 2, got %d %+v", len(ns2), ns2)
	}
	// Empty first B: still allow married-name Jane+Smith if lastB set
	ns3 := NameVariants("Jane", "Doe", "j", "", "Smith", "")
	if len(ns3) != 2 {
		t.Fatalf("empty firstB with lastB: want Jane Doe + Jane Smith, got %+v", ns3)
	}
}

func TestExtractStreetsFromTexts(t *testing.T) {
	texts := []TextSource{
		{Text: "📸 123 Main St Columbus OH 43215 💍", Source: "bio_a"},
		{Text: "She said yes at the park!", Source: "discovery_caption"},
		{Text: "Venue: 456 High Street, Columbus, OH", Source: "recent_caption"},
	}
	cands := ExtractStreetsFromTexts(texts, "Columbus", "OH")
	if len(cands) < 1 {
		t.Fatalf("expected street from bio, got %+v", cands)
	}
	for _, c := range cands {
		if !IsRealStreet(c.Line1) {
			t.Errorf("not real street: %+v", c)
		}
		if c.Kind != KindStreet {
			t.Errorf("kind want street: %+v", c)
		}
	}
}

func TestDetectivePaidCap(t *testing.T) {
	t.Setenv("DETECTIVE_PAID_CAP", "")
	if DetectivePaidCap() != DefaultDetectivePaidCap {
		t.Errorf("default want %d", DefaultDetectivePaidCap)
	}
	t.Setenv("DETECTIVE_PAID_CAP", "5")
	if DetectivePaidCap() != 5 {
		t.Errorf("want 5, got %d", DetectivePaidCap())
	}
	t.Setenv("DETECTIVE_PAID_CAP", "0")
	if DetectivePaidCap() != DefaultDetectivePaidCap {
		t.Error("0 should fall back to default")
	}
	t.Setenv("DETECTIVE_PAID_CAP", "999")
	if DetectivePaidCap() != 40 {
		t.Errorf("ceiling 40, got %d", DetectivePaidCap())
	}
}

func TestMulti_MaxPaidCalls(t *testing.T) {
	p1 := &countProvider{name: "trestle"}
	p2 := &countProvider{name: "pdl"}
	// countProvider.Name returns name — isPaidProvider checks trestle/pdl
	// but countProvider is not named via type - we use Name() string
	// isPaidProvider uses pr.Name() so "trestle" and "pdl" count as paid.
	// Wait - countProvider in cascade_test has name field - good.
	// But we need Available() true and Search - countProvider exists in cascade_test.go same package.

	m := &Multi{
		Primary:      p1,
		Paid:         []Provider{p2},
		MaxPaidCalls: 1,
		SkipFree:     true,
		SkipFallback: true,
	}
	// Use a different stub that returns high conf to stop? With MaxPaidCalls=1,
	// primary runs (1 paid), then p2 should be skipped.
	_, _ = m.Search(context.Background(), Query{FirstName: "A", LastName: "B", City: "Columbus", Region: "OH"})
	if p1.calls != 1 {
		t.Errorf("primary calls want 1, got %d", p1.calls)
	}
	if p2.calls != 0 {
		t.Errorf("second paid should be skipped by budget, got %d", p2.calls)
	}
}

func TestMulti_SkipFree(t *testing.T) {
	free := &countProvider{name: "county_property"}
	m := &Multi{
		Primary: &stubProvider{name: "heuristic", cands: []Candidate{
			{City: "Columbus", Region: "OH", Confidence: 0.3, Kind: KindLocality},
		}},
		Free:         []Provider{free},
		SkipFree:     true,
		SkipFallback: true,
	}
	res, err := m.Search(context.Background(), Query{FirstName: "J", LastName: "D", City: "Columbus", Region: "OH"})
	if err != nil {
		t.Fatal(err)
	}
	if free.calls != 0 {
		t.Errorf("free should be skipped, calls=%d", free.calls)
	}
	if len(res.Candidates) == 0 {
		t.Error("expected primary candidates")
	}
}

func TestMergeResults_RanksStreets(t *testing.T) {
	a := Result{Provider: "a", Candidates: []Candidate{
		{City: "Columbus", Region: "OH", Confidence: 0.4, Kind: KindLocality, Source: "a"},
	}}
	b := Result{Provider: "b", Candidates: []Candidate{
		{Line1: "10 Oak St", City: "Columbus", Region: "OH", Confidence: 0.7, Kind: KindStreet, Source: "b"},
	}}
	m := MergeResults(a, b)
	if !HasStreetCandidates(m.Candidates) {
		t.Fatal("expected street")
	}
	if !IsRealStreet(m.Candidates[0].Line1) {
		t.Errorf("top should be street: %+v", m.Candidates[0])
	}
	if m.PaidCalls != 0 {
		// neither had paid calls set
	}
}
