package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// WeddingDateInput is the Instagram evidence the model reads to predict when a
// couple will marry — beyond what the marriage-license feed already tells us.
type WeddingDateInput struct {
	Caption        string
	BioA           string
	BioB           string
	RecentCaptions []string
	VenueTags      []string
	VendorTags     []string
	CurrentDate    time.Time
}

// WeddingDatePrediction is the model's proposal. PredictedDate is zero-valued
// when no clear signal is found (confidence 0) — the caller must not guess.
type WeddingDatePrediction struct {
	PredictedDate time.Time `json:"predicted_date"`
	Confidence    float64   `json:"confidence"`
	Reason        string    `json:"reason"`
	Source        string    `json:"source,omitempty"`
}

const weddingDateSystem = `You predict a couple's wedding date from Instagram signals. Use captions, bios, venue tags, vendor tags. Look for: "getting married in October", "wedding countdown", "save the date", venue booking patterns, photographer booking patterns. Return JSON: {"predicted_date":"","confidence":0.0,"reason":"","source":""}. The predicted_date must be ISO 8601 (YYYY-MM-DD). If no clear signal, return empty predicted_date with confidence 0. Never invent a date — only infer from explicit textual evidence.`

// PredictWeddingDateFromSocial uses the existing LLM fallback chain (OpenAI →
// Baseten) to infer a wedding date from Instagram signals. Safe to call from
// the ingest worker — failure is non-fatal and returns a zero-value prediction.
// ponytail: ceiling = no Claude tier in this chain (matches home_market.go);
// upgrade path = thread it through FallbackInterpreter if circuit-breaker
// protection across all three providers is needed.
func PredictWeddingDateFromSocial(ctx context.Context, in WeddingDateInput) (WeddingDatePrediction, error) {
	var bld strings.Builder
	now := in.CurrentDate
	if now.IsZero() {
		now = time.Now().UTC()
	}
	fmt.Fprintf(&bld, "Today is %s.\n", now.Format("2006-01-02"))
	if t := fence("caption", in.Caption); t != "" {
		bld.WriteString(t + "\n")
	}
	if t := fence("bio_a", in.BioA); t != "" {
		bld.WriteString(t + "\n")
	}
	if t := fence("bio_b", in.BioB); t != "" {
		bld.WriteString(t + "\n")
	}
	if len(in.RecentCaptions) > 0 {
		bld.WriteString("Recent captions:\n")
		for _, c := range in.RecentCaptions {
			if c = sanitizeLLMInput(c); c != "" {
				bld.WriteString("- " + c + "\n")
			}
		}
	}
	if len(in.VenueTags) > 0 {
		bld.WriteString("Venue tags: " + sanitizeLLMInput(strings.Join(in.VenueTags, ", ")) + "\n")
	}
	if len(in.VendorTags) > 0 {
		bld.WriteString("Vendor tags: " + sanitizeLLMInput(strings.Join(in.VendorTags, ", ")) + "\n")
	}
	bld.WriteString("\nReturn JSON only.")
	prompt := bld.String()

	var raw string
	var src string

	if o := NewOpenAIInterpreter(); o.Available() {
		r, _, err := o.complete(ctx, weddingDateSystem, prompt)
		if err == nil {
			raw, src = r, "openai:"+o.model
		}
	}
	if raw == "" {
		if b := NewBasetenInterpreter(); b.Available() {
			r, _, err := b.complete(ctx, weddingDateSystem, prompt)
			if err == nil {
				raw, src = r, "baseten:wedding_date"
			}
		}
	}
	if raw == "" {
		// No LLM available or all failed — zero-value prediction, not an error
		// that should block ingest. ponytail: returning nil error here is
		// intentional — the caller treats "no prediction" as a no-op.
		return WeddingDatePrediction{Source: ""}, nil
	}

	p, err := parseWeddingDateRaw(raw)
	if err != nil {
		return WeddingDatePrediction{}, err
	}
	p.Source = src
	// No date or zero confidence = no signal found. Don't guess.
	if p.Confidence == 0 {
		return WeddingDatePrediction{Source: p.Source, Reason: p.Reason}, nil
	}
	if !p.PredictedDate.IsZero() {
		p.PredictedDate = p.PredictedDate.UTC()
	}
	return p, nil
}

// ParseWeddingDateJSON decodes a raw LLM JSON blob into a WeddingDatePrediction.
// Exported so the test can exercise parsing without a live model call.
func ParseWeddingDateJSON(raw string) (WeddingDatePrediction, error) {
	p, err := parseWeddingDateRaw(raw)
	if err != nil {
		return WeddingDatePrediction{}, fmt.Errorf("parse wedding date: %w", err)
	}
	return p, nil
}

// parseWeddingDateRaw decodes the model JSON. The predicted_date is read as a
// string first because time.Time's UnmarshalJSON rejects "" (the no-signal
// case) — we only parse it into a time.Time when it's non-empty.
func parseWeddingDateRaw(raw string) (WeddingDatePrediction, error) {
	var dec struct {
		PredictedDate string  `json:"predicted_date"`
		Confidence    float64 `json:"confidence"`
		Reason        string  `json:"reason"`
		Source        string  `json:"source,omitempty"`
	}
	if err := json.Unmarshal([]byte(extractJSON(raw)), &dec); err != nil {
		return WeddingDatePrediction{}, fmt.Errorf("parse wedding date: %w", err)
	}
	var p WeddingDatePrediction
	p.Confidence = dec.Confidence
	p.Reason = dec.Reason
	p.Source = dec.Source
	if dec.PredictedDate != "" {
		t, ok := parseFlexibleDate(dec.PredictedDate)
		if !ok {
			return WeddingDatePrediction{}, fmt.Errorf("parse wedding date: bad predicted_date %q", dec.PredictedDate)
		}
		p.PredictedDate = t
	}
	if p.Confidence < 0 {
		p.Confidence = 0
	}
	if p.Confidence > 1 {
		p.Confidence = 1
	}
	return p, nil
}

// parseFlexibleDate accepts YYYY-MM-DD or full RFC3339. Returns false if the
// string can't be parsed as either.
func parseFlexibleDate(s string) (time.Time, bool) {
	for _, layout := range []string{time.RFC3339, "2006-01-02"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}
