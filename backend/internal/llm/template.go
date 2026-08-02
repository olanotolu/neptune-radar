package llm

import (
	"context"
	"fmt"
	"strings"

	"neptune-social-radar/backend/internal/signals"
)

// TemplateInterpreter is the deterministic fallback: no network calls, fully
// reproducible, used automatically when ANTHROPIC_API_KEY is unset or a
// Claude call fails, and used directly by tests so they never depend on
// network access.
type TemplateInterpreter struct{}

func NewTemplateInterpreter() *TemplateInterpreter { return &TemplateInterpreter{} }

var breakupWords = []string{"breakup", "broke up", "single again", "ex-fiance", "ex fiancé"}

func (t *TemplateInterpreter) InterpretSignal(ctx context.Context, req SignalRequest) (Interpretation, error) {
	text := strings.ToLower(req.Text)
	switch req.CandidateEventType {
	case "engagement":
		// The confidence means "this text is explicit engagement language."
		// It is judged against the shared signal vocabulary — the same
		// deterministic phrases/hashtags the analyst used, plus bio-style
		// fiancé references — never a private keyword list that could drift.
		sig := signals.ExtractFromPayload(map[string]any{"caption": req.Text})
		switch {
		case sig.HasExplicitLanguage() || signals.HasExplicitBioLanguage(req.Text):
			return Interpretation{
				Confidence:    0.9,
				ProposedStage: "engaged",
				Rationale:     "template: explicit engagement language confirmed" + templateDetail(sig),
				Source:        "template:signal_vocabulary_v2",
			}, nil
		case len(sig.SupportingHits)+len(sig.StatusHits)+len(sig.VendorHits)+len(sig.CulturalHits) > 0 || strings.Contains(text, "💍"):
			// Engagement-adjacent but not explicit: a ring emoji, #Fiance,
			// vendor tags. Not enough to assert an announcement.
			return Interpretation{
				Confidence:    0.5,
				ProposedStage: "engaged",
				Rationale:     "template: engagement-adjacent signals only (supporting/vendor/status tags or ring emoji), no explicit announcement language",
				Source:        "template:signal_vocabulary_v2",
			}, nil
		default:
			return Interpretation{
				Confidence:    0.2,
				ProposedStage: "engaged",
				Rationale:     "template: no explicit engagement language found",
				Source:        "template:signal_vocabulary_v2",
			}, nil
		}
	case "relationship_state_change":
		// Base confidence comes from the deterministic evidence collected
		// upstream (unfollow persistence, bio/post reversal); the model's
		// role here is limited to language plausibility, so we start
		// moderate and let corroboration in the scorer do the real work.
		conf := 0.55
		if countHits(text, breakupWords) > 0 {
			conf = 0.3 // explicit breakup language is exactly what we must NOT amplify into a claim
		}
		return Interpretation{
			Confidence:    conf,
			ProposedStage: "status_uncertain",
			Rationale:     "template: unfollow/bio-reversal pattern observed; language alone does not confirm a breakup",
			Source:        "template:state_change_v1",
		}, nil
	default:
		return Interpretation{Confidence: 0, ProposedStage: req.PriorStage, Rationale: "template: no candidate signal", Source: "template:none"}, nil
	}
}

func countHits(text string, words []string) int {
	n := 0
	for _, w := range words {
		if strings.Contains(text, w) {
			n++
		}
	}
	return n
}

func templateDetail(sig signals.PostSignals) string {
	if lang := sig.LanguageSummary(); lang != "" {
		return " (" + lang + ")"
	}
	return ""
}

func (t *TemplateInterpreter) DraftCopy(ctx context.Context, req CopyRequest) (Copy, error) {
	switch req.ActionType {
	case "review":
		engagementConf, partnerConf := req.Confidence, req.Confidence
		if req.EngagementConfidence != nil {
			engagementConf = *req.EngagementConfidence
		}
		if req.PartnerConfidence != nil {
			partnerConf = *req.PartnerConfidence
		}
		var lines []string
		lines = append(lines, "💍 NEW NEPTUNE PROSPECT", "")
		lines = append(lines, fmt.Sprintf("People: %s + %s", req.PersonName, req.PartnerName))
		lines = append(lines, "Event: Engagement")
		if req.DetectedAt != "" {
			lines = append(lines, "Detected: "+req.DetectedAt)
		}
		if req.Location != "" {
			lines = append(lines, "Location: "+req.Location)
		}
		lines = append(lines, "Evidence:")
		for _, e := range req.EvidenceSummary {
			lines = append(lines, "- "+e)
		}
		lines = append(lines, "",
			fmt.Sprintf("Engagement confidence: %.0f%%", engagementConf*100),
			fmt.Sprintf("Partner-match confidence: %.0f%%", partnerConf*100),
			"Recommended action: Human review",
		)
		// Brand-safe: dual-partner, celebrate-first, never hard-sell prenup.
		partner := req.PartnerName
		if partner == "" {
			partner = "your partner"
		}
		return Copy{
			InternalNote: strings.Join(lines, "\n") + "\n\nBrand rule: celebrate first. Soft invite to chat only after congratulate. Never pitch prenup on day-one touch.",
			CustomerFacing: fmt.Sprintf(
				"Dear %s & %s — congratulations on your engagement. Wishing you both a beautiful chapter ahead. When you're ready to plan the admin of partnership together (no pressure), Neptune is here with a free guided chat and flat-fee attorneys — both of you, side by side.",
				req.PersonName, partner,
			),
		}, nil
	case "investigate":
		engagementConf, partnerConf := req.Confidence, req.Confidence
		if req.EngagementConfidence != nil {
			engagementConf = *req.EngagementConfidence
		}
		if req.PartnerConfidence != nil {
			partnerConf = *req.PartnerConfidence
		}
		var lines []string
		lines = append(lines, "🔍 NEPTUNE PROSPECT — INVESTIGATION QUEUE", "")
		lines = append(lines, fmt.Sprintf("People: %s + %s", req.PersonName, req.PartnerName))
		lines = append(lines, "Event: Possible engagement (below the create-prospect bar)")
		if req.DetectedAt != "" {
			lines = append(lines, "Detected: "+req.DetectedAt)
		}
		if req.Location != "" {
			lines = append(lines, "Location: "+req.Location)
		}
		lines = append(lines, "Evidence:")
		for _, e := range req.EvidenceSummary {
			lines = append(lines, "- "+e)
		}
		lines = append(lines, "",
			fmt.Sprintf("Engagement confidence: %.0f%%", engagementConf*100),
			fmt.Sprintf("Partner-match confidence: %.0f%%", partnerConf*100),
			"Recommended action: Human investigation — verify the couple and the event are real before any outreach is considered.",
		)
		partner := req.PartnerName
		if partner == "" {
			partner = "your partner"
		}
		return Copy{
			InternalNote: strings.Join(lines, "\n") + "\n\nBrand rule: no customer-facing send until identity is confirmed.",
			CustomerFacing: fmt.Sprintf(
				"Dear %s & %s — congratulations on this exciting chapter. Whenever you both are ready, Neptune has a calm, no-pressure way to plan ahead together.",
				req.PersonName, partner,
			),
		}, nil
	case "concierge_review":
		return Copy{
			InternalNote: fmt.Sprintf(
				"🚨 Relationship context changed\n\n%s no longer follows %s. A prior engagement reference was also removed.\nMeaningful-change confidence: %.0f%%.\n\nTheir prenup intake is currently active.\n\nRecommended action: pause automated reminders and ask the concierge to review.\n\n\"Mutual follow agreement may have been unilaterally amended.\"",
				req.PersonName, req.PartnerName, req.Confidence*100,
			),
			// Calm, neutral, never claims a breakup — matches the spec's
			// exact requirement even in the deterministic fallback path.
			CustomerFacing: fmt.Sprintf(
				"Hey %s, just checking in. Would you like to continue, pause, or close your Neptune process? No explanation needed.",
				req.PersonName,
			),
		}, nil
	case "postcard":
		where := ""
		if req.Location != "" {
			where = " in " + req.Location
		}
		a, b := req.PersonName, req.PartnerName
		var lines []string
		lines = append(lines, "✉ CONGRATULATE POSTCARD DRAFT", "")
		lines = append(lines, fmt.Sprintf("First names: %s & %s", a, b))
		if req.Location != "" {
			lines = append(lines, "Market: "+req.Location)
		}
		lines = append(lines, "Evidence:")
		for _, e := range req.EvidenceSummary {
			lines = append(lines, "- "+e)
		}
		lines = append(lines, "",
			"Address: research only until a human verifies street + ZIP.",
			"Policy: never auto-mail. Kit must reach ready_to_mail after verification.",
			"Copy rule: greet with first names only (never @handles on the card).",
		)
		return Copy{
			InternalNote: strings.Join(lines, "\n"),
			CustomerFacing: fmt.Sprintf(
				"Dear %s & %s,\n\nCongratulations on your engagement%s! May this season of planning be full of joy, good light, and the people you love most.\n\nWith warm regards,\nNeptune",
				a, b, where,
			),
		}, nil
	default:
		return Copy{InternalNote: "no action", CustomerFacing: ""}, nil
	}
}
