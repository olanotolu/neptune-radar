// Package policy is the Policy Guard. This is the deterministic boundary of
// the whole system: consent, visibility, confidence thresholds, idempotency,
// and permitted actions live here in plain Go, never behind a model call.
//
// INVARIANT: this package must never import neptune-social-radar/backend/internal/llm.
// See policy_no_llm_import_test.go, which enforces this with `go list -deps`.
package policy

import (
	"encoding/json"
	"strings"

	"neptune-social-radar/backend/internal/ontology"
	"neptune-social-radar/backend/internal/store"
)

const (
	// Engagement-prospect tiers (the spec's points, normalized: 100 pts = 1.0):
	//   90+  → create Neptune prospect (ActionReview)
	//   70–89 → human investigation queue (ActionInvestigate)
	//   <70  → retained as an unconfirmed inference or discarded — no card
	ThresholdCreateProspect = 0.90
	ThresholdInvestigate    = 0.70
	// Relationship-state-change bar (unchanged): below this the system logs
	// and moves on — no concierge card, no customer visibility.
	ThresholdSurfaceReview = 0.60
	ThresholdActOnStage    = 0.60 // internal relationship-stage transition bar
)

type Decision struct {
	ShouldAct       bool
	ActionType      ontology.ActionType
	FinalConfidence float64
	Reasons         []string
	ConsentOK       bool
}

// Decide is pure: given a hypothesis, its finalConfidence (already computed
// by scorer — policy does not re-derive it, only judges it), and the
// person's consent, it returns what — if anything — Neptune is allowed to
// recommend. It never calls the model and never sends anything itself.
func Decide(s *store.Store, hyp ontology.LifeEventHypothesis, finalConfidence float64) (Decision, error) {
	d := Decision{FinalConfidence: finalConfidence}

	if hyp.PersonID != "" {
		switch {
		case hyp.EventType == ontology.EventTypeEngagement && isFreshDiscovery(s, hyp.PersonID):
			// Event-first discovery: this person was never a Neptune lead —
			// they were identified from a public signal moments ago, so
			// there is nothing to have consented to yet. Surfacing an
			// INTERNAL prospect card for human review is exactly what
			// discovery is for; it is not itself outreach. The approval
			// queue's human-in-the-loop gate (never automated sending) is
			// what keeps this safe, not a consent record that can't exist yet.
			d.ConsentOK = true
			d.Reasons = append(d.Reasons, "newly-discovered prospect, not yet a Neptune lead — no consent required to surface for human review; still required before any outreach")
		default:
			consent, err := s.ActiveConsentForPerson(hyp.PersonID)
			if err != nil {
				d.Reasons = append(d.Reasons, "no active consent policy on file — treating as no permitted actions")
				d.ConsentOK = false
			} else {
				d.ConsentOK = true
				allowed := allowedAction(hyp.EventType)
				if !actionAllowed(consent.AllowedActions, allowed) {
					d.ConsentOK = false
					d.Reasons = append(d.Reasons, "consent scope does not include "+allowed)
				}
			}
		}
	} else {
		d.Reasons = append(d.Reasons, "hypothesis has no resolved person yet")
	}

	// The confidence bar depends on the trigger: engagement prospects are
	// tiered (0.90 create / 0.70 investigate / below that nothing), state
	// changes keep the original 0.60 concierge bar. Below the bar the signal
	// is retained as an unconfirmed inference or discarded — never surfaced.
	bar := ThresholdSurfaceReview
	if hyp.EventType == ontology.EventTypeEngagement {
		bar = ThresholdInvestigate
	}
	if finalConfidence < bar {
		d.Reasons = append(d.Reasons, "confidence below the surfacing threshold — precision over recall, retained as unconfirmed inference, no card")
		d.ShouldAct = false
		return d, nil
	}

	if !d.ConsentOK {
		d.ShouldAct = false
		return d, nil
	}

	// Idempotency / duplicate suppression: never stack a second pending
	// recommendation for the same hypothesis.
	existing, err := existingPendingAction(s, hyp.ID)
	if err != nil {
		return d, err
	}
	if existing {
		d.Reasons = append(d.Reasons, "a recommended action for this hypothesis is already pending — suppressing duplicate")
		d.ShouldAct = false
		return d, nil
	}

	d.ShouldAct = true
	switch hyp.EventType {
	case ontology.EventTypeEngagement:
		if finalConfidence >= ThresholdCreateProspect {
			d.ActionType = ontology.ActionReview
			d.Reasons = append(d.Reasons, "confidence at or above the create-prospect tier (0.90) — surfacing as a Neptune prospect")
		} else {
			d.ActionType = ontology.ActionInvestigate
			d.Reasons = append(d.Reasons, "confidence in the investigation tier (0.70–0.89) — routing to the human investigation queue before any prospect is created")
		}
	case ontology.EventTypeRelationshipChange:
		d.ActionType = ontology.ActionConciergeReview
	default:
		d.ActionType = ontology.ActionNoAction
		d.ShouldAct = false
	}
	return d, nil
}

// isFreshDiscovery reports whether a person arrived via an inferred identity
// signal (a reciprocal tag, a co-tag) rather than real CRM intake (e.g. a
// prenup guide download). Those two cases get different consent treatment:
// a real lead who lacks/withdrew consent is off-limits; a stranger just
// discovered from a public post was never asked because they aren't a lead
// yet — that's the entire point of opportunity discovery.
func isFreshDiscovery(s *store.Store, personID string) bool {
	p, err := s.GetPerson(personID)
	if err != nil {
		return false
	}
	return strings.HasPrefix(p.CRMSource, "inferred_")
}

func allowedAction(eventType string) string {
	if eventType == ontology.EventTypeRelationshipChange {
		return "pause_automation"
	}
	return "create_lead"
}

func actionAllowed(allowed []string, action string) bool {
	for _, a := range allowed {
		if a == action {
			return true
		}
	}
	return false
}

func existingPendingAction(s *store.Store, hypothesisID string) (bool, error) {
	pending, err := s.ListActions("pending")
	if err != nil {
		return false, err
	}
	for _, a := range pending {
		if a.HypothesisID == hypothesisID {
			return true, nil
		}
	}
	return false, nil
}

// marshalReasons is a small helper the operator uses when persisting a
// decision's reasons onto the recommended_action payload for auditability.
func MarshalReasons(reasons []string) string {
	b, _ := json.Marshal(reasons)
	return string(b)
}
