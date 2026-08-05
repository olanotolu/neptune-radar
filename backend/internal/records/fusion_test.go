package records

import "testing"

// TestFuseCandidates verifies that Bayesian fusion re-ranks candidates when
// providers have different historical accuracy. Three street candidates from
// three providers, all same base confidence. Trestle (85% in NY) should
// outrank TPS (60% in OH) which should outrank heuristic (no history → 0.5).
func TestFuseCandidates(t *testing.T) {
	cands := []Candidate{
		{Line1: "123 Main St", City: "Columbus", Region: "OH", Country: "US",
			Confidence: 0.80, Source: "truepeoplesearch", Kind: KindStreet},
		{Line1: "456 Oak Ave", City: "New York", Region: "NY", Country: "US",
			Confidence: 0.80, Source: "trestle", Kind: KindStreet},
		{Line1: "789 Elm Rd", City: "Anywhere", Region: "TX", Country: "US",
			Confidence: 0.80, Source: "heuristic", Kind: KindStreet},
	}
	accuracy := map[string]map[string]float64{
		"trestle":          {"NY": 0.85},
		"truepeoplesearch": {"OH": 0.60},
		// heuristic: no entry → 0.5 prior
	}
	fused := FuseCandidates(cands, accuracy)
	if len(fused) != 3 {
		t.Fatalf("expected 3 candidates, got %d", len(fused))
	}
	// Trestle (0.80 × 0.85 = 0.68) should rank above TPS (0.80 × 0.60 = 0.48)
	// which should rank above heuristic (0.80 × 0.50 = 0.40).
	if fused[0].Source != "trestle" {
		t.Errorf("expected trestle first, got %s (conf %.3f)", fused[0].Source, fused[0].Confidence)
	}
	if fused[1].Source != "truepeoplesearch" {
		t.Errorf("expected truepeoplesearch second, got %s (conf %.3f)", fused[1].Source, fused[1].Confidence)
	}
	if fused[2].Source != "heuristic" {
		t.Errorf("expected heuristic third, got %s (conf %.3f)", fused[2].Source, fused[2].Confidence)
	}
	// Verify the fused confidence is base × accuracy
	if fused[0].Confidence != 0.80*0.85 {
		t.Errorf("trestle fused confidence: got %.4f, want %.4f", fused[0].Confidence, 0.80*0.85)
	}
	// Verify input is not mutated
	if cands[0].Confidence != 0.80 {
		t.Errorf("FuseCandidates mutated input: original conf = %.4f", cands[0].Confidence)
	}
}

// TestFuseCandidatesColdStart verifies that with no accuracy data, the 0.5
// prior is applied uniformly and ranking is preserved (all equal).
func TestFuseCandidatesColdStart(t *testing.T) {
	cands := []Candidate{
		{Line1: "123 Main St", City: "Columbus", Region: "OH",
			Confidence: 0.90, Source: "trestle", Kind: KindStreet},
		{Line1: "456 Oak Ave", City: "New York", Region: "NY",
			Confidence: 0.70, Source: "pdl", Kind: KindStreet},
	}
	fused := FuseCandidates(cands, nil)
	// No accuracy map → 0.5 prior for all → order preserved by confidence
	if fused[0].Source != "trestle" {
		t.Errorf("cold start should preserve confidence order, got %s first", fused[0].Source)
	}
	if fused[0].Confidence != 0.90*DefaultAccuracyPrior {
		t.Errorf("cold start fused confidence: got %.4f, want %.4f",
			fused[0].Confidence, 0.90*DefaultAccuracyPrior)
	}
}
