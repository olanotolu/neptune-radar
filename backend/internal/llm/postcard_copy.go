package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// PostcardCopyInput is the evidence the model uses to write a personalized
// postcard message. All fields are optional — the model works with whatever
// signals are available.
type PostcardCopyInput struct {
	NameA           string
	NameB           string
	City            string
	Venue           string   // detected venue from discovery post / geotag
	BioA            string
	BioB            string
	DiscoveryCaption string
	PrenupSignals   []string // prenup intent signals (e.g. "founder", "second marriage")
}

// PostcardCopyOutput is the personalized message + tone label.
type PostcardCopyOutput struct {
	Message string `json:"message"`
	Tone    string `json:"tone"`
	Source  string `json:"source,omitempty"`
}

// maxPostcardChars is the postcard print limit.
const maxPostcardChars = 280

const postcardCopySystem = `Write a warm, concise postcard message (max 280 chars) congratulating this couple
on their engagement. Reference specific details from their signals. Be genuine, not salesy.
End with a soft CTA to visit meetneptune.com.
Respond with JSON only: {"message":"...","tone":"warm|formal|direct"}`

// PersonalizePostcardCopy uses the existing LLM fallback chain (OpenAI → Baseten
// → Claude), mirroring PredictPrenupIntent. Falls back to templateCopy on any
// failure — never blocks the pipeline.
func PersonalizePostcardCopy(ctx context.Context, in PostcardCopyInput, templateCopy string) (PostcardCopyOutput, error) {
	var bld strings.Builder
	fmt.Fprintf(&bld, "Couple: %s & %s\n", sanitizeLLMInput(in.NameA), sanitizeLLMInput(in.NameB))
	if in.City != "" {
		fmt.Fprintf(&bld, "City: %s\n", sanitizeLLMInput(in.City))
	}
	if in.Venue != "" {
		fmt.Fprintf(&bld, "Venue: %s\n", sanitizeLLMInput(in.Venue))
	}
	if t := fence("bio_a", in.BioA); t != "" {
		bld.WriteString(t + "\n")
	}
	if t := fence("bio_b", in.BioB); t != "" {
		bld.WriteString(t + "\n")
	}
	if t := fence("discovery_caption", in.DiscoveryCaption); t != "" {
		bld.WriteString(t + "\n")
	}
	if len(in.PrenupSignals) > 0 {
		bld.WriteString("Prenup intent signals:\n")
		for _, s := range in.PrenupSignals {
			if s = sanitizeLLMInput(s); s != "" {
				bld.WriteString("- " + s + "\n")
			}
		}
	}
	bld.WriteString("\nReturn JSON only.")
	prompt := bld.String()

	var raw string
	var src string
	var lastErr error

	// ponytail: same fallback order as PredictPrenupIntent — no new chain.
	if o := NewOpenAIInterpreter(); o.Available() {
		r, _, err := o.complete(ctx, postcardCopySystem, prompt)
		if err == nil {
			raw, src = r, "openai:"+o.model
		} else {
			lastErr = err
		}
	}
	if raw == "" {
		if b := NewBasetenInterpreter(); b.Available() {
			r, _, err := b.complete(ctx, postcardCopySystem, prompt)
			if err == nil {
				raw, src = r, "baseten:postcard_copy"
			} else {
				lastErr = err
			}
		}
	}
	if raw == "" {
		if c := NewClaudeInterpreter(); c.Available() {
			r, _, err := c.complete(ctx, postcardCopySystem, prompt)
			if err == nil {
				raw, src = r, "claude:"+c.model
			} else {
				lastErr = err
			}
		}
	}

	// Fallback to template copy — never block the pipeline on an LLM outage.
	if raw == "" {
		return PostcardCopyOutput{
			Message: truncatePostcard(templateCopy),
			Tone:    "template",
			Source:  "fallback",
		}, lastErr
	}

	var out PostcardCopyOutput
	if err := json.Unmarshal([]byte(extractJSON(raw)), &out); err != nil {
		return PostcardCopyOutput{
			Message: truncatePostcard(templateCopy),
			Tone:    "template",
			Source:  "fallback",
		}, fmt.Errorf("parse postcard copy: %w", err)
	}
	out.Message = truncatePostcard(strings.TrimSpace(out.Message))
	if out.Message == "" {
		out.Message = truncatePostcard(templateCopy)
	}
	if out.Tone == "" {
		out.Tone = "warm"
	}
	out.Source = src
	return out, nil
}

func truncatePostcard(s string) string {
	if len(s) <= maxPostcardChars {
		return s
	}
	return s[:maxPostcardChars-1] + "…"
}
