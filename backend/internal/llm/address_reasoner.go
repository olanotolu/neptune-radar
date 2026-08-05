package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// AddressCandidateInput is one address candidate sent to the model for ranking.
type AddressCandidateInput struct {
	Index      int     `json:"index"`
	Line1      string  `json:"line1,omitempty"`
	City       string  `json:"city,omitempty"`
	Region     string  `json:"region,omitempty"`
	Postal     string  `json:"postal,omitempty"`
	Confidence float64 `json:"confidence"`
	Source     string  `json:"source"`
}

// AddressReasoningInput is the context the model uses to rank address candidates.
type AddressReasoningInput struct {
	PersonA     string
	PersonB     string
	BioA        string
	BioB        string
	VendorCity  string
	VendorState string
	PostGeotags []string
	Candidates  []AddressCandidateInput
}

// AddressReasoningResult is the model's ranking + rationale.
type AddressReasoningResult struct {
	RankedIndices []int   `json:"ranked_indices"`
	Rationale     string  `json:"rationale"`
	Confidence    float64 `json:"confidence"`
	Agreement     bool    `json:"agreement"`
	Source        string  `json:"source,omitempty"`
}

const addressReasonerSystem = `You are an address verification expert. Given address candidates for a couple,
along with context clues from their social media, rank the candidates by likelihood
of being their correct home address. Consider: city match with bios/geotags, provider
reliability, and street-level vs locality-level precision. Return JSON only:
{"ranked_indices":[0,2,1,3,4],"rationale":"one to three sentences","confidence":0.8}`

// ReasonAboutAddresses uses the existing LLM fallback chain (OpenAI → Baseten),
// mirroring InferHomeMarket. Failure is non-fatal: returns an empty no-op result
// so the detective pipeline is never blocked.
func ReasonAboutAddresses(ctx context.Context, in AddressReasoningInput) (AddressReasoningResult, error) {
	if len(in.Candidates) == 0 {
		return AddressReasoningResult{}, nil
	}

	var bld strings.Builder
	fmt.Fprintf(&bld, "Couple: %s & %s\n",
		sanitizeLLMInput(in.PersonA), sanitizeLLMInput(in.PersonB))
	if t := fence("bio_a", in.BioA); t != "" {
		bld.WriteString(t + "\n")
	}
	if t := fence("bio_b", in.BioB); t != "" {
		bld.WriteString(t + "\n")
	}
	if in.VendorCity != "" {
		fmt.Fprintf(&bld, "Vendor market: %s, %s (may be shoot city, not home)\n",
			sanitizeLLMInput(in.VendorCity), sanitizeLLMInput(in.VendorState))
	}
	if len(in.PostGeotags) > 0 {
		bld.WriteString("Post geotags:\n")
		for _, g := range in.PostGeotags {
			if g = sanitizeLLMInput(g); g != "" {
				bld.WriteString("- " + g + "\n")
			}
		}
	}
	bld.WriteString("\nAddress candidates (index → address):\n")
	for _, c := range in.Candidates {
		fmt.Fprintf(&bld, "[%d] %s, %s, %s (conf %.2f, source: %s)\n",
			c.Index, sanitizeLLMInput(c.Line1), sanitizeLLMInput(c.City),
			sanitizeLLMInput(c.Region), c.Confidence, sanitizeLLMInput(c.Source))
	}
	bld.WriteString("\nReturn JSON only.")
	prompt := bld.String()

	var raw string
	var src string
	var lastErr error

	// ponytail: same fallback order as InferHomeMarket — OpenAI, then Baseten.
	if o := NewOpenAIInterpreter(); o.Available() {
		r, _, err := o.complete(ctx, addressReasonerSystem, prompt)
		if err == nil {
			raw, src = r, "openai:"+o.model
		} else {
			lastErr = err
		}
	}
	if raw == "" {
		if b := NewBasetenInterpreter(); b.Available() {
			r, _, err := b.complete(ctx, addressReasonerSystem, prompt)
			if err == nil {
				raw, src = r, "baseten:address_reasoner"
			} else {
				lastErr = err
			}
		}
	}
	if raw == "" {
		// No-op fallback — never block the pipeline on an LLM outage.
		return AddressReasoningResult{}, lastErr
	}

	res, err := ParseAddressReasoningJSON(raw)
	if err != nil {
		return AddressReasoningResult{}, err
	}
	res.Source = src
	// Agreement: does the model's top pick match index 0 (Bayesian top)?
	res.Agreement = len(res.RankedIndices) > 0 && res.RankedIndices[0] == 0
	return res, nil
}

// ParseAddressReasoningJSON parses the model's JSON response. Separated so the
// test can exercise it without a live model call.
func ParseAddressReasoningJSON(raw string) (AddressReasoningResult, error) {
	var r AddressReasoningResult
	if err := json.Unmarshal([]byte(extractJSON(raw)), &r); err != nil {
		return AddressReasoningResult{}, fmt.Errorf("parse address reasoning: %w", err)
	}
	if r.Confidence < 0 {
		r.Confidence = 0
	}
	if r.Confidence > 1 {
		r.Confidence = 1
	}
	// ponytail: clamp indices to valid range — model may hallucinate out-of-bounds.
	for i, idx := range r.RankedIndices {
		if idx < 0 {
			r.RankedIndices[i] = 0
		}
	}
	return r, nil
}
