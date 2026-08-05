package llm

import "testing"

// TestParseAddressReasoningJSON exercises the LLM response parser with a mocked
// model output — the one check that fails if the JSON contract breaks. No live
// model call; this is the ponytail "smallest thing that fails if logic breaks".
func TestParseAddressReasoningJSON(t *testing.T) {
	raw := `{"ranked_indices":[0,2,1,3,4],"rationale":"Candidate 0 matches the city in bio_a and is a street-level hit from a reliable provider.","confidence":0.85}`
	r, err := ParseAddressReasoningJSON(raw)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(r.RankedIndices) != 5 {
		t.Fatalf("ranked_indices len = %d, want 5", len(r.RankedIndices))
	}
	if r.RankedIndices[0] != 0 {
		t.Errorf("top index = %d, want 0", r.RankedIndices[0])
	}
	if r.Confidence != 0.85 {
		t.Errorf("confidence = %.2f, want 0.85", r.Confidence)
	}
	if r.Rationale == "" {
		t.Error("rationale is empty, want the model's explanation")
	}

	// Disagreement case: model picks index 2 as top.
	disagree, _ := ParseAddressReasoningJSON(`{"ranked_indices":[2,0,1],"rationale":"Candidate 2 better matches geotag.","confidence":0.7}`)
	if disagree.RankedIndices[0] != 2 {
		t.Errorf("disagree top = %d, want 2", disagree.RankedIndices[0])
	}

	// Confidence clamp: out-of-range values snap to [0,1].
	hi, _ := ParseAddressReasoningJSON(`{"ranked_indices":[0],"rationale":"x","confidence":1.5}`)
	if hi.Confidence != 1 {
		t.Errorf("clamp high = %.2f, want 1", hi.Confidence)
	}
	lo, _ := ParseAddressReasoningJSON(`{"ranked_indices":[0],"rationale":"x","confidence":-0.3}`)
	if lo.Confidence != 0 {
		t.Errorf("clamp low = %.2f, want 0", lo.Confidence)
	}
}
