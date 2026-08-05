package llm

import "testing"

// TestParseRelationshipStrengthJSON exercises the LLM response parser with a
// mocked model output — the one check that fails if the JSON contract breaks.
// No live model call; this is the ponytail "smallest thing that fails if logic
// breaks".
func TestParseRelationshipStrengthJSON(t *testing.T) {
	raw := `{"score":0.85,"category":"engaged","key_signals":["consistent romantic language","high mutual tag frequency","proposal language detected"],"rationale":"captions show consistent romantic language and proposal references; high mutual tag frequency"}`
	r, err := ParseRelationshipStrengthJSON(raw)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if r.Score != 0.85 {
		t.Errorf("score = %.2f, want 0.85", r.Score)
	}
	if r.Category != "engaged" {
		t.Errorf("category = %q, want \"engaged\"", r.Category)
	}
	if len(r.KeySignals) != 3 {
		t.Errorf("key_signals len = %d, want 3", len(r.KeySignals))
	}
	if r.Rationale == "" {
		t.Error("rationale is empty, want the model's rationale")
	}

	// Score clamp: out-of-range values snap to [0,1].
	hi, _ := ParseRelationshipStrengthJSON(`{"score":1.5,"category":"married","key_signals":[],"rationale":"x"}`)
	if hi.Score != 1 {
		t.Errorf("clamp high = %.2f, want 1", hi.Score)
	}
	lo, _ := ParseRelationshipStrengthJSON(`{"score":-0.3,"category":"casual_dating","key_signals":[],"rationale":"x"}`)
	if lo.Score != 0 {
		t.Errorf("clamp low = %.2f, want 0", lo.Score)
	}

	// Category normalization: uppercase/whitespace trimmed to lowercase.
	r2, _ := ParseRelationshipStrengthJSON(`{"score":0.5,"category":"  Serious  ","key_signals":[],"rationale":"x"}`)
	if r2.Category != "serious" {
		t.Errorf("normalized category = %q, want \"serious\"", r2.Category)
	}
}
