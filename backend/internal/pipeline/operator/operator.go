// Package operator is the Workflow Operator. It writes state: it creates the
// pending recommended_action a human must approve, and — only once approved
// — performs the actual side effect (CRM lead/case creation, pausing
// automation). It never calls a model; copy is handed to it as plain strings
// already drafted by internal/pipeline/conversation.
package operator

import (
	"encoding/json"

	"neptune-social-radar/backend/internal/ontology"
	"neptune-social-radar/backend/internal/pipeline/policy"
	"neptune-social-radar/backend/internal/store"
)

// ActionCopy mirrors llm.Copy in shape only — operator never imports
// internal/llm, so the orchestrator converts before calling in here.
type ActionCopy struct {
	InternalNote   string
	CustomerFacing string
}

type proposedPayload struct {
	InternalNote   string   `json:"internal_note"`
	CustomerFacing string   `json:"customer_facing"`
	Reasons        []string `json:"reasons"`
}

// ProposeAction persists the pending human-reviewed card. Nothing customer
// visible happens until a human calls Approve.
func ProposeAction(s *store.Store, decision policy.Decision, hyp ontology.LifeEventHypothesis, caseID string, copy ActionCopy) (ontology.RecommendedAction, error) {
	payload, err := json.Marshal(proposedPayload{
		InternalNote:   copy.InternalNote,
		CustomerFacing: copy.CustomerFacing,
		Reasons:        decision.Reasons,
	})
	if err != nil {
		return ontology.RecommendedAction{}, err
	}
	return s.CreateRecommendedAction(ontology.RecommendedAction{
		HypothesisID:    hyp.ID,
		CaseID:          caseID,
		ActionType:      decision.ActionType,
		ProposedPayload: string(payload),
	})
}

// Approve executes the side effect a human just signed off on: for an
// engagement lead this opens the Neptune case; for a relationship-state
// change it pauses automation and moves the case into concierge review.
// Nothing here contacts the customer directly — in this prototype "sending"
// the customer-facing message just means marking the action executed.
func Approve(s *store.Store, actionID, decidedBy string) (ontology.ExecutedAction, error) {
	action, err := s.GetAction(actionID)
	if err != nil {
		return ontology.ExecutedAction{}, err
	}
	hyp, err := s.GetHypothesis(action.HypothesisID)
	if err != nil {
		return ontology.ExecutedAction{}, err
	}

	// Claim the decision BEFORE any side effect. Without the pending guard,
	// a double-click (or a retried request) ran this whole body twice —
	// minting duplicate CRM leads and cases from one human decision.
	if err := s.DecideActionIfPending(actionID, ontology.ActionExecuted, decidedBy); err != nil {
		return ontology.ExecutedAction{}, err
	}

	detail := map[string]any{"action_type": string(action.ActionType)}

	switch action.ActionType {
	case ontology.ActionReview, ontology.ActionInvestigate:
		// Both engagement tiers open the prenup case on approval — the only
		// difference is how the prospect got here (90+ direct vs. cleared by
		// a human investigation), which the lead status records.
		leadStatus := "reviewed"
		if action.ActionType == ontology.ActionInvestigate {
			leadStatus = "investigated"
		}
		lead, err := s.CreateLead(ontology.CRMLead{PersonID: hyp.PersonID, HypothesisID: hyp.ID, LeadType: "prenup", Status: leadStatus})
		if err != nil {
			return ontology.ExecutedAction{}, err
		}
		c, err := s.CreateCase(ontology.NeptuneCase{CoupleID: hyp.CoupleID, LeadID: lead.ID, CaseType: "prenup", Status: "intake"})
		if err != nil {
			return ontology.ExecutedAction{}, err
		}
		detail["lead_id"] = lead.ID
		detail["case_id"] = c.ID

	case ontology.ActionConciergeReview:
		rel, err := s.CurrentRelationship(hyp.CoupleID)
		if err == nil {
			if _, err := s.TransitionRelationship(hyp.CoupleID, rel.Stage, rel.Confidence, rel.VisibilityScope, true); err != nil {
				return ontology.ExecutedAction{}, err
			}
		}
		if activeCase, err := s.GetActiveCaseForCouple(hyp.CoupleID); err == nil {
			if err := s.UpdateCaseStatus(activeCase.ID, "review"); err != nil {
				return ontology.ExecutedAction{}, err
			}
			detail["case_id"] = activeCase.ID
		}
	}

	if err := s.UpdateHypothesisStatus(hyp.ID, ontology.HypothesisConfirmed); err != nil {
		return ontology.ExecutedAction{}, err
	}

	detailJSON, _ := json.Marshal(detail)
	exec, err := s.CreateExecutedAction(ontology.ExecutedAction{
		RecommendedActionID: actionID,
		Result:              "success",
		Detail:              string(detailJSON),
	})
	return exec, err
}

func Ignore(s *store.Store, actionID, decidedBy string) error {
	action, err := s.GetAction(actionID)
	if err != nil {
		return err
	}
	// Same guard as Approve: ignoring an already-executed action used to
	// flip its hypothesis to "rejected" AFTER it was confirmed and executed
	// — retroactive history corruption.
	if err := s.DecideActionIfPending(actionID, ontology.ActionIgnored, decidedBy); err != nil {
		return err
	}
	return s.UpdateHypothesisStatus(action.HypothesisID, ontology.HypothesisRejected)
}
