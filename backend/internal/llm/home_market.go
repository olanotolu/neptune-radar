package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// HomeMarketInput is free-text evidence for guessing where a couple *lives*
// (not where they got engaged). The model must not invent street addresses.
type HomeMarketInput struct {
	PersonA       string
	PersonB       string
	HandleA       string
	HandleB       string
	BioA          string
	BioB          string
	Caption       string
	VendorCity    string
	VendorState   string
	PostLocation  string
	MarketHint    string // prep-suggested city
	EvidenceLines []string
}

// HomeMarketGuess is a city/region proposal only — never a street.
type HomeMarketGuess struct {
	City       string  `json:"city"`
	Region     string  `json:"region"` // 2-letter US state when possible
	Confidence float64 `json:"confidence"`
	Reason     string  `json:"reason"`
	Source     string  `json:"source,omitempty"`
}

const homeMarketSystem = `You infer where a couple most likely LIVES (home market) for US mailing.
Use bios, school, captions, geotags. Prefer home over wedding venue / photographer city.
NEVER invent a street address, apartment, or ZIP. Only city + US state abbrev when confident.
If evidence is non-US or unknown, return empty city.
Respond with JSON only: {"city":"","region":"","confidence":0.0,"reason":""}`

// InferHomeMarket uses OpenAI first (OPENAI_API_KEY from Vercel), then Baseten.
// Safe to call from detective — failure is non-fatal. Never invents streets.
func InferHomeMarket(ctx context.Context, in HomeMarketInput) (HomeMarketGuess, error) {
	var bld strings.Builder
	fmt.Fprintf(&bld, "Couple: %s (%s) & %s (%s)\n",
		sanitizeLLMInput(in.PersonA), sanitizeLLMInput(in.HandleA),
		sanitizeLLMInput(in.PersonB), sanitizeLLMInput(in.HandleB))
	if t := fence("bio_a", in.BioA); t != "" {
		bld.WriteString(t + "\n")
	}
	if t := fence("bio_b", in.BioB); t != "" {
		bld.WriteString(t + "\n")
	}
	if t := fence("caption", in.Caption); t != "" {
		bld.WriteString(t + "\n")
	}
	if in.PostLocation != "" {
		fmt.Fprintf(&bld, "Post venue tag: %s\n", sanitizeLLMInput(in.PostLocation))
	}
	if in.VendorCity != "" {
		fmt.Fprintf(&bld, "Photographer/vendor market: %s, %s (may be shoot city, not home)\n",
			sanitizeLLMInput(in.VendorCity), sanitizeLLMInput(in.VendorState))
	}
	if in.MarketHint != "" {
		fmt.Fprintf(&bld, "Heuristic hint: %s\n", sanitizeLLMInput(in.MarketHint))
	}
	if len(in.EvidenceLines) > 0 {
		bld.WriteString("Evidence:\n")
		for _, e := range in.EvidenceLines {
			if e = sanitizeLLMInput(e); e != "" {
				bld.WriteString("- " + e + "\n")
			}
		}
	}
	bld.WriteString("\nReturn JSON only.")
	prompt := bld.String()

	var raw string
	var src string
	var lastErr error

	if o := NewOpenAIInterpreter(); o.Available() {
		r, _, err := o.complete(ctx, homeMarketSystem, prompt)
		if err == nil {
			raw, src = r, "openai:"+o.model
		} else {
			lastErr = err
		}
	}
	if raw == "" {
		if b := NewBasetenInterpreter(); b.Available() {
			r, _, err := b.complete(ctx, homeMarketSystem, prompt)
			if err == nil {
				raw, src = r, "baseten:home_market"
			} else {
				lastErr = err
			}
		}
	}
	if raw == "" {
		if lastErr != nil {
			return HomeMarketGuess{}, lastErr
		}
		return HomeMarketGuess{}, fmt.Errorf("no LLM available for home market (set OPENAI_API_KEY)")
	}

	var g HomeMarketGuess
	if err := json.Unmarshal([]byte(extractJSON(raw)), &g); err != nil {
		return HomeMarketGuess{}, fmt.Errorf("parse home market: %w", err)
	}
	g.City = strings.TrimSpace(g.City)
	g.Region = strings.ToUpper(strings.TrimSpace(g.Region))
	if len(g.Region) > 2 {
		g.Region = stateAbbrev(g.Region)
	}
	if g.Confidence < 0 {
		g.Confidence = 0
	}
	if g.Confidence > 1 {
		g.Confidence = 1
	}
	g.Source = src
	if g.City == "" || strings.EqualFold(g.City, "unknown") {
		return HomeMarketGuess{Source: g.Source, Reason: g.Reason}, nil
	}
	return g, nil
}

func stateAbbrev(s string) string {
	m := map[string]string{
		"texas": "TX", "ohio": "OH", "new york": "NY", "california": "CA",
		"florida": "FL", "illinois": "IL", "georgia": "GA", "pennsylvania": "PA",
		"north carolina": "NC", "michigan": "MI", "tennessee": "TN", "colorado": "CO",
	}
	if a, ok := m[strings.ToLower(strings.TrimSpace(s))]; ok {
		return a
	}
	if len(s) == 2 {
		return strings.ToUpper(s)
	}
	return ""
}
