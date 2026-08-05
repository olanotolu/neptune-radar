package records

import (
	"strings"
)

// DefaultAccuracyPrior is the cold-start Bayesian prior when no history exists
// for a provider×state. 0.5 = no information either way.
const DefaultAccuracyPrior = 0.5

// FuseCandidates re-weights and re-ranks candidates by their provider's
// historical accuracy for the candidate's state. Pure function: does not
// mutate the input slice.
//
// Bayesian fusion: fused_confidence = confidence × accuracy(provider, state).
// No history → 0.5 prior (neutral). rankCandidates is reused so kind-first
// ordering (street > locality > research_link) is preserved and only the
// within-tier order changes.
func FuseCandidates(cands []Candidate, accuracy map[string]map[string]float64) []Candidate {
	out := make([]Candidate, len(cands))
	copy(out, cands)
	for i := range out {
		out[i].Confidence *= providerAccuracy(out[i].Source, out[i].Region, accuracy)
	}
	rankCandidates(out)
	return out
}

// providerAccuracy looks up the historical hit rate for a provider in a state.
// Returns the 0.5 prior when no data exists (cold start).
func providerAccuracy(provider, state string, accuracy map[string]map[string]float64) float64 {
	if states, ok := accuracy[provider]; ok {
		if acc, ok := states[strings.ToUpper(strings.TrimSpace(state))]; ok && acc > 0 {
			return acc
		}
	}
	return DefaultAccuracyPrior
}
