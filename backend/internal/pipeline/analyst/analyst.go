// Package analyst is the Relationship Analyst. Deterministic code decides
// WHETHER a raw event is worth interpreting at all (cheap keyword/state
// checks below) — this keeps the system from burning an LLM call on every
// follow/unfollow/private-toggle. Once a candidate exists, the model judges
// how strongly the language actually supports it. The model's confidence is
// only ever a proposal: internal/pipeline/policy has final say.
package analyst

import (
	"context"
	"strings"

	"neptune-social-radar/backend/internal/llm"
	"neptune-social-radar/backend/internal/ontology"
	"neptune-social-radar/backend/internal/pipeline/identity"
	"neptune-social-radar/backend/internal/pipeline/watchtower"
	"neptune-social-radar/backend/internal/signals"
	"neptune-social-radar/backend/internal/store"
)

// Aliases onto the shared ontology constants — kept here so existing callers
// in this package's own files don't need renaming, but the canonical
// definitions live in ontology so policy can reference them without pulling
// in this package (and therefore internal/llm).
const (
	CandidateEngagement         = ontology.EventTypeEngagement
	CandidateRelationshipChange = ontology.EventTypeRelationshipChange
)

// Candidate holds the deterministic pre-filter's verdict.
type Candidate struct {
	EventType     string // CandidateEngagement or CandidateRelationshipChange
	ProposedStage ontology.RelationshipStage
	Text          string // caption/bio text relevant to the candidate, for the model to judge
	// SignalContext is a compact deterministic summary of the combined signal
	// vocabulary matched on the triggering post (hashtag tiers, source class,
	// visuals, references, location) — given to the model so it judges the
	// language in context instead of from the caption string alone.
	SignalContext string
}

// Detect is pure deterministic logic — no model call. It looks only at the
// current (post-identity-resolution) DB state plus the triggering event.
func Detect(s *store.Store, res identity.Resolved, raw watchtower.RawEvent, prior *ontology.Relationship) (*Candidate, error) {
	priorStage := ontology.StageUnknown
	if prior != nil {
		priorStage = prior.Stage
	}

	switch raw.Type {
	case "post":
		if res.Couple == nil {
			return nil, nil
		}
		confirmed, err := hasConfirmedEngagement(s, res.Couple.ID)
		if err != nil {
			return nil, err
		}
		// The combined signal vocabulary decides what is engagement-shaped:
		// explicit caption phrases, high-intent/inclusive/location hashtags,
		// a known vendor referencing exactly two people, or proposal-shaped
		// visuals with both people referenced. Weak/supporting tags never
		// create a candidate; exclusion signals are scored, not filtered
		// here, so the audit trail shows the suppression.
		sig := signals.ExtractFromPayload(raw.Payload)
		if sig.CreatesCandidate() && priorStage != ontology.StageMarried && !confirmed {
			// Not gated on priorStage == engaged: the internal ontology may
			// already have flipped to "engaged" (an unconfirmed belief
			// update) while a human still hasn't approved the case — a
			// later corroborating post must still count as reinforcing
			// evidence for that same open hypothesis, not be dropped.
			return &Candidate{
				EventType:     CandidateEngagement,
				ProposedStage: ontology.StageEngaged,
				Text:          candidateText(sig),
				SignalContext: signalContext(sig),
			}, nil
		}
		return nil, nil

	case "bio_change":
		if res.Couple == nil {
			return nil, nil
		}
		confirmed, err := hasConfirmedEngagement(s, res.Couple.ID)
		if err != nil {
			return nil, err
		}
		bio, _ := raw.Payload["bio"].(string)
		mentionsFiance := signals.HasExplicitBioLanguage(bio)
		switch {
		case mentionsFiance && priorStage != ontology.StageMarried && !confirmed:
			return &Candidate{EventType: CandidateEngagement, ProposedStage: ontology.StageEngaged, Text: bio}, nil
		case !mentionsFiance && (priorStage == ontology.StageEngaged || priorStage == ontology.StageMarried || priorStage == ontology.StageStatusUncertain):
			return &Candidate{EventType: CandidateRelationshipChange, ProposedStage: ontology.StageStatusUncertain, Text: "bio no longer references partner"}, nil
		}
		return nil, nil

	case "follow_change":
		if res.Couple == nil {
			return nil, nil
		}
		active := true
		if v, ok := raw.Payload["active"].(bool); ok {
			active = v
		}
		hasOpenStateChange, err := hasOpenStateChangeHypothesis(s, res.Couple.ID)
		if err != nil {
			return nil, err
		}
		if !active && (priorStage == ontology.StageEngaged || priorStage == ontology.StageMarried) {
			return &Candidate{EventType: CandidateRelationshipChange, ProposedStage: ontology.StageStatusUncertain, Text: "unfollow detected"}, nil
		}
		if hasOpenStateChange {
			// re-evaluation trigger: a reversed unfollow (or any further
			// follow-state churn) must feed back into the existing open
			// hypothesis rather than being ignored.
			return &Candidate{EventType: CandidateRelationshipChange, ProposedStage: ontology.StageStatusUncertain, Text: "follow-state changed again"}, nil
		}
		hasOpenEngagement, err := hasOpenEngagementHypothesis(s, res.Couple.ID)
		if err != nil {
			return nil, err
		}
		if active && hasOpenEngagement {
			// A follow forming (or a second direction of an existing one)
			// isn't engagement *language*, but for an event-first discovery
			// still waiting on partner-match evidence, it's exactly the
			// corroboration that raises confidence in who the couple is —
			// feed it back into the open hypothesis rather than dropping it.
			return &Candidate{EventType: CandidateEngagement, ProposedStage: ontology.StageEngaged, Text: "mutual follow corroborates the pairing"}, nil
		}
		return nil, nil

	case "post_archived":
		if res.Couple == nil {
			return nil, nil
		}
		hasOpenStateChange, err := hasOpenStateChangeHypothesis(s, res.Couple.ID)
		if err != nil {
			return nil, err
		}
		if hasOpenStateChange {
			return &Candidate{EventType: CandidateRelationshipChange, ProposedStage: ontology.StageStatusUncertain, Text: "referenced post was archived"}, nil
		}
		return nil, nil

	default:
		return nil, nil
	}
}

func hasOpenStateChangeHypothesis(s *store.Store, coupleID string) (bool, error) {
	h, err := s.LatestHypothesisForCouple(coupleID)
	if err != nil {
		return false, nil //nolint:nilerr // sql.ErrNoRows just means "no open hypothesis"
	}
	return h.EventType == CandidateRelationshipChange && (h.Status == ontology.HypothesisUnconfirmed || h.Status == ontology.HypothesisCorroborating), nil
}

func hasOpenEngagementHypothesis(s *store.Store, coupleID string) (bool, error) {
	h, err := s.LatestHypothesisForCouple(coupleID)
	if err != nil {
		return false, nil //nolint:nilerr // sql.ErrNoRows just means "no open hypothesis"
	}
	return h.EventType == CandidateEngagement && (h.Status == ontology.HypothesisUnconfirmed || h.Status == ontology.HypothesisCorroborating), nil
}

// WorthResolvingIdentity is the cheap, exported pre-check the orchestrator
// uses to decide whether a post is even worth resolving event-first identity
// for (i.e. calling identity.ResolveCoupleFromPair). It delegates to the
// shared signals vocabulary — the same classification Detect uses — so
// "worth resolving identity for" and "counts as an engagement candidate"
// never drift apart. Unlike Detect's candidate check it hard-excludes ads,
// styled shoots, and reposts: exclusion-tagged posts must not mint brand-new
// couple records for models.
func WorthResolvingIdentity(raw watchtower.RawEvent) bool {
	if raw.Type != "post" {
		return false
	}
	return signals.ExtractFromPayload(raw.Payload).WorthResolvingIdentity()
}

// candidateText is what the model judges. Usually the caption; for a
// caption-less vendor/visual post, a plain description of the deterministic
// signal so the model still has something concrete to evaluate.
func candidateText(sig signals.PostSignals) string {
	if sig.Caption != "" {
		return sig.Caption
	}
	parts := []string{}
	if sig.IsKnownVendor() {
		parts = append(parts, "post by a known "+strings.ReplaceAll(sig.SourceClass, "_", " "))
	} else {
		parts = append(parts, "post")
	}
	if refs := sig.ReferencedHandles(); len(refs) > 0 {
		parts = append(parts, "referencing @"+strings.Join(refs, " and @"))
	}
	if sig.HasVisual() {
		parts = append(parts, "with visual signals: "+strings.Join(sig.Visual, ", "))
	}
	return strings.Join(parts, " ")
}

// signalContext is the compact deterministic signal summary handed to the
// model alongside the text — the vocabulary matches, source class, visuals,
// referenced accounts, and location that led to this candidate.
func signalContext(sig signals.PostSignals) string {
	var parts []string
	if lang := sig.LanguageSummary(); lang != "" {
		parts = append(parts, "explicit language: "+lang)
	}
	if sup := sig.SupportingSummary(); sup != "" {
		parts = append(parts, "supporting tags: "+sup)
	}
	if sig.IsKnownVendor() {
		parts = append(parts, "source account type: "+sig.SourceClass)
	}
	if sig.HasVisual() {
		parts = append(parts, "visual signals: "+strings.Join(sig.Visual, ", "))
	}
	if refs := sig.ReferencedHandles(); len(refs) > 0 {
		parts = append(parts, "accounts referenced: "+strings.Join(refs, ", "))
	}
	if sig.Location != "" {
		parts = append(parts, "location: "+sig.Location)
	}
	if len(sig.PenaltyKinds) > 0 {
		var penalties []string
		for p := range sig.PenaltyKinds {
			penalties = append(penalties, p)
		}
		parts = append(parts, "exclusion signals present: "+strings.Join(penalties, ", "))
	}
	return strings.Join(parts, "; ")
}

// hasConfirmedEngagement reports whether a human has already approved an
// engagement hypothesis for this couple — once that's true, further
// engagement-shaped posts are just noise, not a new signal to act on.
func hasConfirmedEngagement(s *store.Store, coupleID string) (bool, error) {
	var count int
	err := s.DB.QueryRow(
		`SELECT COUNT(*) FROM life_event_hypotheses WHERE couple_id = $1 AND event_type = $2 AND status = $3`,
		coupleID, CandidateEngagement, ontology.HypothesisConfirmed,
	).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Interpret asks the model (or its deterministic fallback) to judge a
// detected candidate and returns the raw suggestion — callers must still run
// it through the scorer and policy before it can do anything.
func Interpret(ctx context.Context, interp llm.Interpreter, cand Candidate, raw watchtower.RawEvent, partnerHandle, priorStage string, existingEvidence []string) (llm.Interpretation, error) {
	return interp.InterpretSignal(ctx, llm.SignalRequest{
		CandidateEventType: cand.EventType,
		ObservationType:    raw.Type,
		Text:               cand.Text,
		SignalContext:      cand.SignalContext,
		Handle:             raw.Handle,
		PartnerHandle:      partnerHandle,
		PriorStage:         priorStage,
		ExistingEvidence:   existingEvidence,
	})
}
