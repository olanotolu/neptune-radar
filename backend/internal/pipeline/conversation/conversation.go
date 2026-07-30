// Package conversation is the Conversation Agent: it drafts the funny
// internal copy and the careful, calm customer-facing copy for one
// recommended action. It is one of only two packages allowed to call
// internal/llm (the other is analyst) — policy and operator never do.
package conversation

import (
	"context"

	"neptune-social-radar/backend/internal/llm"
	"neptune-social-radar/backend/internal/ontology"
)

func Draft(ctx context.Context, interp llm.Interpreter, actionType ontology.ActionType, hyp ontology.LifeEventHypothesis, personName, partnerName string, evidenceSummary []string, location string) (llm.Copy, error) {
	return interp.DraftCopy(ctx, llm.CopyRequest{
		ActionType:           string(actionType),
		EventType:            hyp.EventType,
		PersonName:           personName,
		PartnerName:          partnerName,
		Confidence:           hyp.Confidence,
		EvidenceSummary:      evidenceSummary,
		EngagementConfidence: hyp.EngagementConfidence,
		PartnerConfidence:    hyp.PartnerConfidence,
		Location:             location,
	})
}
