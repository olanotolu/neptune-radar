package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// PrenupIntentInput is the evidence the model uses to predict how likely a
// couple is to need/want a prenup. Careers and age indicators are extracted
// from bios by the model — callers just pass the raw text + asset signals.
type PrenupIntentInput struct {
	BioA             string
	BioB             string
	PropertyValue    int64  // estimated home value from county records (0 if unknown)
	BusinessSignals  string // free-text: "Person A bio mentions 'founder/CEO'"
	AgeIndicators    string // free-text: "second marriage signals, 'kids from prior'"
}

// PrenupIntentPrediction is the model's prenup-intent estimate.
type PrenupIntentPrediction struct {
	IntentScore float64  `json:"intent_score"` // 0-1
	Reason      string   `json:"reason"`
	Signals     []string `json:"signals"`
	Source      string   `json:"source,omitempty"`
}

const prenupIntentSystem = `You predict prenup intent from couple signals. Consider: career fields
(law, medicine, business, tech), asset levels, age gap, education, prior marriages.
Be conservative. Return JSON only: {"intent_score":0.0,"reason":"one sentence","signals":["short","short"]}`

// PredictPrenupIntent uses the existing LLM fallback chain (OpenAI → Baseten →
// Claude), mirroring InferHomeMarket. Failure is non-fatal: returns a neutral
// 0.3 score so the feature degrades gracefully without blocking the pipeline.
func PredictPrenupIntent(ctx context.Context, in PrenupIntentInput) (PrenupIntentPrediction, error) {
	var bld strings.Builder
	if t := fence("bio_a", in.BioA); t != "" {
		bld.WriteString(t + "\n")
	}
	if t := fence("bio_b", in.BioB); t != "" {
		bld.WriteString(t + "\n")
	}
	if in.PropertyValue > 0 {
		fmt.Fprintf(&bld, "Estimated property value: $%d\n", in.PropertyValue)
	}
	if in.BusinessSignals != "" {
		fmt.Fprintf(&bld, "Business ownership signals: %s\n", sanitizeLLMInput(in.BusinessSignals))
	}
	if in.AgeIndicators != "" {
		fmt.Fprintf(&bld, "Age / prior marriage indicators: %s\n", sanitizeLLMInput(in.AgeIndicators))
	}
	bld.WriteString("\nReturn JSON only.")
	prompt := bld.String()

	var raw string
	var src string
	var lastErr error

	// ponytail: same fallback order as InferHomeMarket — OpenAI, then Baseten,
	// then Claude. No new chain; reuse the per-provider complete() methods.
	if o := NewOpenAIInterpreter(); o.Available() {
		r, _, err := o.complete(ctx, prenupIntentSystem, prompt)
		if err == nil {
			raw, src = r, "openai:"+o.model
		} else {
			lastErr = err
		}
	}
	if raw == "" {
		if b := NewBasetenInterpreter(); b.Available() {
			r, _, err := b.complete(ctx, prenupIntentSystem, prompt)
			if err == nil {
				raw, src = r, "baseten:prenup_intent"
			} else {
				lastErr = err
			}
		}
	}
	if raw == "" {
		if c := NewClaudeInterpreter(); c.Available() {
			r, _, err := c.complete(ctx, prenupIntentSystem, prompt)
			if err == nil {
				raw, src = r, "claude:"+c.model
			} else {
				lastErr = err
			}
		}
	}

	// Neutral fallback — never block the pipeline on an LLM outage.
	if raw == "" {
		return PrenupIntentPrediction{
			IntentScore: 0.3,
			Reason:      "neutral default — LLM unavailable",
			Signals:     []string{"llm_fallback"},
			Source:      "fallback",
		}, lastErr
	}

	var p PrenupIntentPrediction
	if err := json.Unmarshal([]byte(extractJSON(raw)), &p); err != nil {
		return PrenupIntentPrediction{
			IntentScore: 0.3,
			Reason:      "neutral default — parse error",
			Signals:     []string{"parse_fallback"},
			Source:      "fallback",
		}, fmt.Errorf("parse prenup intent: %w", err)
	}
	if p.IntentScore < 0 {
		p.IntentScore = 0
	}
	if p.IntentScore > 1 {
		p.IntentScore = 1
	}
	if len(p.Signals) == 0 {
		p.Signals = []string{"no_signals_returned"}
	}
	p.Source = src
	return p, nil
}
