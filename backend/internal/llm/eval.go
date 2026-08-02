package llm

import (
	"context"
	"fmt"
	"strings"
)

// EvalCase is one golden test case: a signal input and the expected
// interpretation. The harness runs the interpreter against the input and
// compares the output to the expected fields.
type EvalCase struct {
	ID          string        `json:"id"`
	Description string        `json:"description"`
	Input       SignalRequest `json:"input"`
	// Expected fields — partial match: only set fields are checked.
	ExpectedStage      string  `json:"expected_stage,omitempty"`
	ExpectedConfidence float64 `json:"expected_confidence,omitempty"` // ±0.15 tolerance
}

// EvalResult is the outcome of running one eval case.
type EvalResult struct {
	CaseID        string  `json:"case_id"`
	Passed        bool    `json:"passed"`
	GotStage      string  `json:"got_stage"`
	GotConfidence float64 `json:"got_confidence"`
	GotSource     string  `json:"got_source,omitempty"`
	Diff          string  `json:"diff,omitempty"`
}

// GoldenCases is the built-in eval set. These are hand-verified cases that
// any interpreter (Baseten, Claude, template) should get right. Add cases
// here as bugs are found in production — each case is a regression test.
// ponytail: ceiling — in-file, hand-curated. Upgrade to a DB-backed eval
// set if the team needs to add cases via UI without redeploying.
var GoldenCases = []EvalCase{
	{
		ID:          "engaged-ring-caption",
		Description: "Engagement ring photo with 'she said yes' caption",
		Input: SignalRequest{
			CandidateEventType: "engagement",
			ObservationType:    "post",
			Text:               "She said yes! 💍",
			SignalContext:      "hashtag_tier1:engaged,shesaidyes,wedding; visual:ring",
		},
		ExpectedStage:      "engaged",
		ExpectedConfidence: 0.85,
	},
	{
		ID:          "married-caption",
		Description: "Wedding photo with 'just married' caption",
		Input: SignalRequest{
			CandidateEventType: "engagement",
			ObservationType:    "post",
			Text:               "Just married! Best day ever ❤️",
			SignalContext:      "hashtag_tier1:married,weddingday,justmarried; visual:ceremony",
		},
		ExpectedStage:      "married",
		ExpectedConfidence: 0.90,
	},
	{
		ID:          "vendor-portfolio",
		Description: "Vendor portfolio post — should not trigger high confidence",
		Input: SignalRequest{
			CandidateEventType: "engagement",
			ObservationType:    "post",
			Text:               "Beautiful wedding venue tour today, DM for bookings",
			SignalContext:      "hashtag_tier2:weddingvenue,weddingplanner; source_class:vendor",
		},
		ExpectedStage:      "",
		ExpectedConfidence: 0.0,
	},
}

// RunEval runs the golden cases against an interpreter and returns the
// results. A case passes if the got-stage matches expected-stage (or
// expected is empty) and the confidence is within ±0.15 of expected.
func RunEval(ctx context.Context, interp Interpreter) []EvalResult {
	var results []EvalResult
	for _, c := range GoldenCases {
		out, err := interp.InterpretSignal(ctx, c.Input)
		r := EvalResult{
			CaseID:        c.ID,
			GotStage:      out.ProposedStage,
			GotConfidence: out.Confidence,
			GotSource:     out.Source,
		}
		if err != nil {
			r.Diff = fmt.Sprintf("error: %v", err)
			results = append(results, r)
			continue
		}
		var diffs []string
		if c.ExpectedStage != "" && c.ExpectedStage != out.ProposedStage {
			diffs = append(diffs, fmt.Sprintf("stage: want %q got %q", c.ExpectedStage, out.ProposedStage))
		}
		if c.ExpectedConfidence > 0 {
			lo := c.ExpectedConfidence - 0.15
			hi := c.ExpectedConfidence + 0.15
			if out.Confidence < lo || out.Confidence > hi {
				diffs = append(diffs, fmt.Sprintf("confidence: want %.2f±0.15 got %.2f", c.ExpectedConfidence, out.Confidence))
			}
		}
		r.Passed = len(diffs) == 0
		r.Diff = strings.Join(diffs, "; ")
		results = append(results, r)
	}
	return results
}
