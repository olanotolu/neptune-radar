package llm

import "testing"

// TestParseWeddingDateJSON exercises the LLM response parser with a mocked
// model output — the one check that fails if the JSON contract breaks. No live
// model call; this is the ponytail "smallest thing that fails if logic breaks".
func TestParseWeddingDateJSON(t *testing.T) {
	raw := `{"predicted_date":"2025-10-18T00:00:00Z","confidence":0.82,"reason":"caption says 'getting married in October' + venue tag for The Barn","source":"openai:gpt-4.1-mini"}`
	p, err := ParseWeddingDateJSON(raw)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if p.Confidence != 0.82 {
		t.Errorf("confidence = %.2f, want 0.82", p.Confidence)
	}
	if p.PredictedDate.IsZero() {
		t.Fatal("predicted_date is zero, want a parsed date")
	}
	if p.PredictedDate.Year() != 2025 || p.PredictedDate.Month() != 10 || p.PredictedDate.Day() != 18 {
		t.Errorf("predicted_date = %v, want 2025-10-18", p.PredictedDate)
	}
	if p.Reason == "" {
		t.Error("reason is empty, want the model's rationale")
	}

	// No-signal case: empty date + confidence 0 must not produce a date.
	none, err := ParseWeddingDateJSON(`{"predicted_date":"","confidence":0,"reason":"no clear signal","source":""}`)
	if err != nil {
		t.Fatalf("parse no-signal: %v", err)
	}
	if none.Confidence != 0 {
		t.Errorf("no-signal confidence = %.2f, want 0", none.Confidence)
	}

	// Confidence clamp: out-of-range values snap to [0,1].
	hi, _ := ParseWeddingDateJSON(`{"predicted_date":"2025-10-18","confidence":1.5,"reason":"x"}`)
	if hi.Confidence != 1 {
		t.Errorf("clamp high = %.2f, want 1", hi.Confidence)
	}
	lo, _ := ParseWeddingDateJSON(`{"predicted_date":"2025-10-18","confidence":-0.3,"reason":"x"}`)
	if lo.Confidence != 0 {
		t.Errorf("clamp low = %.2f, want 0", lo.Confidence)
	}
}
