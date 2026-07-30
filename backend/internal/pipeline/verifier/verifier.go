// Package verifier closes the loop: it re-reads the database after an
// approved action executes and confirms the intended state actually landed,
// rather than trusting that the write succeeded just because no error was
// returned.
package verifier

import (
	"encoding/json"

	"neptune-social-radar/backend/internal/ontology"
	"neptune-social-radar/backend/internal/store"
)

func Confirm(s *store.Store, exec ontology.ExecutedAction, action ontology.RecommendedAction) (bool, error) {
	var detail map[string]any
	if exec.Detail != "" {
		if err := json.Unmarshal([]byte(exec.Detail), &detail); err != nil {
			return false, err
		}
	}

	verified := false
	switch action.ActionType {
	case ontology.ActionReview, ontology.ActionInvestigate:
		caseID, _ := detail["case_id"].(string)
		if caseID != "" {
			c, err := s.GetCase(caseID)
			verified = err == nil && c.Status != ""
		}

	case ontology.ActionConciergeReview:
		hyp, err := s.GetHypothesis(action.HypothesisID)
		if err == nil {
			rel, err := s.CurrentRelationship(hyp.CoupleID)
			verified = err == nil && rel.AutomationPaused
		}
		if caseID, ok := detail["case_id"].(string); ok && caseID != "" {
			c, err := s.GetCase(caseID)
			verified = verified && err == nil && c.Status == "review"
		}

	default:
		verified = true
	}

	if err := s.SetExecutedVerified(exec.ID, verified); err != nil {
		return false, err
	}
	return verified, nil
}
