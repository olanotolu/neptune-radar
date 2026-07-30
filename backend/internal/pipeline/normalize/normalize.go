// Package normalize turns an ingestion RawEvent into a SocialObservation
// row shape. It does no interpretation — that's the Analyst's job — only
// shape and idempotency-key plumbing.
package normalize

import (
	"encoding/json"

	"neptune-social-radar/backend/internal/ontology"
	"neptune-social-radar/backend/internal/pipeline/watchtower"
)

func Normalize(raw watchtower.RawEvent) (ontology.SocialObservation, error) {
	// The acting handle isn't its own column on social_observations (see
	// schema) — fold it into the stored payload so the signal feed API can
	// display "who" without needing a join through account resolution,
	// which may not have happened yet for a brand new handle.
	merged := map[string]any{"handle": raw.Handle}
	for k, v := range raw.Payload {
		merged[k] = v
	}
	payload, err := json.Marshal(merged)
	if err != nil {
		return ontology.SocialObservation{}, err
	}
	source := raw.Source
	if source == "" {
		source = "unknown"
	}
	return ontology.SocialObservation{
		Monitor:         raw.Monitor,
		ExternalEventID: raw.ExternalEventID,
		ObservationType: raw.Type,
		RawPayload:      string(payload),
		ObservedAt:      raw.OccurredAt,
		Source:          source,
		// Raw social observations are unconfirmed by construction — actual
		// permission to act on them is checked against ConsentPolicy at
		// decision time, not by inflating this field.
		ConsentScope: ontology.ScopeUnconfirmedInfer,
	}, nil
}
