package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// RelationshipStrengthInput is the social-media evidence the model reads to
// assess how strong/serious a couple's relationship is. It augments — never
// replaces — the existing FAIR dispersion metric (passed in as DispersionScore).
type RelationshipStrengthInput struct {
	Partner1Name    string
	Partner2Name    string
	Partner1Bio     string
	Partner2Bio     string
	RecentCaptions  []string // last 10-20 posts mentioning both
	TagFrequency    int      // how often they tag each other
	PostFrequency   float64  // posts per week
	MutualFollows   bool
	DispersionScore float64 // from existing FAIR dispersion metric
}

// RelationshipStrengthResult is the model's assessment.
type RelationshipStrengthResult struct {
	Score      float64  `json:"score"`       // 0-1
	Category   string   `json:"category"`    // casual_dating | serious | engaged | married | uncertain
	KeySignals []string `json:"key_signals"` // e.g. "consistent romantic language"
	Rationale  string   `json:"rationale"`
	Source     string   `json:"source,omitempty"`
}

const relationshipStrengthSystem = `You are a relationship analyst. Given these social media signals for a couple,
assess the strength and seriousness of their relationship. Consider: post frequency, tag patterns,
caption sentiment, mutual engagement, and the provided dispersion score (a graph-based metric where
high dispersion indicates the relationship bridges distinct social circles — a strong romantic signal).
Return JSON only: {"score":0.85,"category":"engaged","key_signals":["...","..."],"rationale":"..."}`

// ScoreRelationshipStrength uses the existing LLM fallback chain (OpenAI →
// Baseten), mirroring home_market.go. Failure is non-fatal: falls back to the
// dispersion_score only (no LLM) so the pipeline is never blocked.
// ponytail: ceiling = no Claude tier in this chain (matches home_market.go);
// upgrade path = thread through FallbackInterpreter if circuit-breaker
// protection across all three providers is needed.
func ScoreRelationshipStrength(ctx context.Context, in RelationshipStrengthInput) (RelationshipStrengthResult, error) {
	var bld strings.Builder
	fmt.Fprintf(&bld, "Couple: %s & %s\n",
		sanitizeLLMInput(in.Partner1Name), sanitizeLLMInput(in.Partner2Name))
	if t := fence("bio_1", in.Partner1Bio); t != "" {
		bld.WriteString(t + "\n")
	}
	if t := fence("bio_2", in.Partner2Bio); t != "" {
		bld.WriteString(t + "\n")
	}
	if len(in.RecentCaptions) > 0 {
		bld.WriteString("Recent captions (posts mentioning both):\n")
		for _, c := range in.RecentCaptions {
			if c = sanitizeLLMInput(c); c != "" {
				bld.WriteString("- " + c + "\n")
			}
		}
	}
	fmt.Fprintf(&bld, "Tag frequency (mutual tags): %d\n", in.TagFrequency)
	fmt.Fprintf(&bld, "Post frequency: %.1f posts/week\n", in.PostFrequency)
	fmt.Fprintf(&bld, "Mutual follows: %v\n", in.MutualFollows)
	fmt.Fprintf(&bld, "Dispersion score (FAIR metric, 0-1): %.2f\n", in.DispersionScore)
	bld.WriteString("\nReturn JSON only.")
	prompt := bld.String()

	var raw string
	var src string
	var lastErr error

	if o := NewOpenAIInterpreter(); o.Available() {
		r, _, err := o.complete(ctx, relationshipStrengthSystem, prompt)
		if err == nil {
			raw, src = r, "openai:"+o.model
		} else {
			lastErr = err
		}
	}
	if raw == "" {
		if b := NewBasetenInterpreter(); b.Available() {
			r, _, err := b.complete(ctx, relationshipStrengthSystem, prompt)
			if err == nil {
				raw, src = r, "baseten:relationship_strength"
			} else {
				lastErr = err
			}
		}
	}

	// Fallback to dispersion_score only — never block the pipeline on an LLM outage.
	if raw == "" {
		return RelationshipStrengthResult{
			Score:    clampScore(in.DispersionScore),
			Category: categoryFromDispersion(in.DispersionScore),
			KeySignals: []string{"dispersion_fallback"},
			Rationale: "LLM unavailable — score derived from FAIR dispersion metric only",
			Source:    "fallback",
		}, lastErr
	}

	var r RelationshipStrengthResult
	if err := json.Unmarshal([]byte(extractJSON(raw)), &r); err != nil {
		return RelationshipStrengthResult{
			Score:    clampScore(in.DispersionScore),
			Category: categoryFromDispersion(in.DispersionScore),
			KeySignals: []string{"parse_fallback"},
			Rationale: "LLM parse error — score derived from FAIR dispersion metric only",
			Source:    "fallback",
		}, fmt.Errorf("parse relationship strength: %w", err)
	}
	r.Score = clampScore(r.Score)
	r.Category = strings.TrimSpace(strings.ToLower(r.Category))
	if r.Category == "" {
		r.Category = categoryFromDispersion(in.DispersionScore)
	}
	if len(r.KeySignals) == 0 {
		r.KeySignals = []string{"no_signals_returned"}
	}
	r.Source = src
	return r, nil
}

// ParseRelationshipStrengthJSON decodes a raw LLM JSON blob. Exported so the
// test can exercise parsing without a live model call.
func ParseRelationshipStrengthJSON(raw string) (RelationshipStrengthResult, error) {
	var r RelationshipStrengthResult
	if err := json.Unmarshal([]byte(extractJSON(raw)), &r); err != nil {
		return RelationshipStrengthResult{}, fmt.Errorf("parse relationship strength: %w", err)
	}
	r.Score = clampScore(r.Score)
	r.Category = strings.TrimSpace(strings.ToLower(r.Category))
	return r, nil
}

func clampScore(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// categoryFromDispersion maps the FAIR dispersion score to a rough category
// when the LLM is unavailable. High dispersion (>0.7) = strong signal.
func categoryFromDispersion(d float64) string {
	if d > 0.7 {
		return "serious"
	}
	if d > 0.4 {
		return "uncertain"
	}
	return "casual_dating"
}
