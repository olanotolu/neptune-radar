// Package llm is the only place model calls happen. Its output is always a
// *proposal* — internal/pipeline/policy has final say over what the system
// is permitted to do with it, and never imports this package.
package llm

import (
	"context"
	"fmt"
)

// SignalRequest describes one candidate life-event signal for the model to
// interpret. Detecting that a candidate exists at all is deterministic
// keyword/rule matching upstream (see pipeline/analyst); the model's job is
// narrower — judge how strongly the language actually supports the
// hypothesis and explain why, in plain terms.
type SignalRequest struct {
	CandidateEventType string // "engagement" or "relationship_state_change"
	ObservationType    string // "post", "bio_change", "follow_change", "post_archived", ...
	Text               string // the caption/bio text under consideration, if any
	// SignalContext is the deterministic signal-vocabulary summary for the
	// post (hashtag tiers matched, source account class, visual signals,
	// referenced accounts, location). Empty for non-post observations.
	SignalContext    string
	Handle           string
	PartnerHandle    string
	PriorStage       string
	ExistingEvidence []string
}

type Interpretation struct {
	Confidence    float64 // model's suggestion only — policy re-validates/clamps
	Rationale     string
	ProposedStage string
	Source        string `json:"-"` // e.g. "claude:claude-sonnet-5" or "template:bio_regex_v1", set by the interpreter that answered
}

// CopyRequest asks for both the internal (funny, safe to be blunt) and
// customer-facing (calm, neutral, never claims a breakup) copy for one
// recommended action.
type CopyRequest struct {
	ActionType      string // "review" (engagement) or "concierge_review" (state change)
	EventType       string
	PersonName      string
	PartnerName     string
	Confidence      float64
	EvidenceSummary []string
	// EngagementConfidence/PartnerConfidence are set (non-nil) only for
	// event-first engagement prospects, where "did this happen" and "did we
	// get the right two people" are reported as two separate numbers rather
	// than one blended score.
	EngagementConfidence *float64
	PartnerConfidence    *float64
	DetectedAt           string // human-readable "how long ago", for the prospect card
	Location             string
}

type Copy struct {
	InternalNote   string
	CustomerFacing string
}

type Interpreter interface {
	InterpretSignal(ctx context.Context, req SignalRequest) (Interpretation, error)
	DraftCopy(ctx context.Context, req CopyRequest) (Copy, error)
}

// formatCopyPrompt is shared by ClaudeInterpreter and BasetenInterpreter so
// the two independently-gated scores reach a real model the same way they
// reach the template fallback.
func formatCopyPrompt(req CopyRequest) string {
	base := fmt.Sprintf(
		"Action type: %s\nEvent type: %s\nPerson: %s\nPartner: %s\nConfidence: %.2f\nEvidence: %v",
		req.ActionType, req.EventType, req.PersonName, req.PartnerName, req.Confidence, req.EvidenceSummary,
	)
	if req.EngagementConfidence != nil && req.PartnerConfidence != nil {
		base += fmt.Sprintf("\nEngagement confidence: %.2f\nPartner-match confidence: %.2f", *req.EngagementConfidence, *req.PartnerConfidence)
	}
	if req.DetectedAt != "" {
		base += "\nDetected: " + req.DetectedAt
	}
	if req.Location != "" {
		base += "\nLocation: " + req.Location
	}
	return base
}
