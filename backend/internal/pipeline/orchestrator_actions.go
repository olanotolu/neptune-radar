package pipeline

import (
	"context"

	"neptune-social-radar/backend/internal/ontology"
	"neptune-social-radar/backend/internal/pipeline/conversation"
	"neptune-social-radar/backend/internal/pipeline/identity"
	"neptune-social-radar/backend/internal/pipeline/operator"
	"neptune-social-radar/backend/internal/pipeline/policy"
	"neptune-social-radar/backend/internal/pipeline/watchtower"
)

func (o *Orchestrator) proposeAction(ctx context.Context, decision policy.Decision, hyp ontology.LifeEventHypothesis, res identity.Resolved, evidence []ontology.Evidence, raw watchtower.RawEvent) (ontology.RecommendedAction, error) {
	personName := res.Account.Handle
	if hyp.PersonID != "" {
		if p, err := o.Store.GetPerson(hyp.PersonID); err == nil {
			personName = p.DisplayName
		}
	}
	partnerName := ""
	if res.PartnerAccount != nil {
		partnerName = res.PartnerAccount.Handle
		if res.PartnerAccount.PersonID != "" {
			if p, err := o.Store.GetPerson(res.PartnerAccount.PersonID); err == nil {
				partnerName = p.DisplayName
			}
		}
	}

	var evidenceSummary []string
	for _, e := range evidence {
		evidenceSummary = append(evidenceSummary, e.Description)
	}

	// Location is routing context, not points: it affects which market and
	// legal workflow apply, so it travels onto the prospect card.
	location, _ := raw.Payload["location"].(string)
	draftedCopy, err := conversation.Draft(ctx, o.Interp, decision.ActionType, hyp, personName, partnerName, evidenceSummary, location)
	if err != nil {
		return ontology.RecommendedAction{}, err
	}

	caseID := ""
	if res.Couple != nil {
		if c, err := o.Store.GetActiveCaseForCouple(res.Couple.ID); err == nil {
			caseID = c.ID
		}
	}

	return operator.ProposeAction(o.Store, decision, hyp, caseID, operator.ActionCopy{
		InternalNote:   draftedCopy.InternalNote,
		CustomerFacing: draftedCopy.CustomerFacing,
	})
}
