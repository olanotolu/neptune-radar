// Package ops holds maintenance / janitor jobs for board health.
package ops

import (
	"context"
	"log"
	"strings"

	"neptune-social-radar/backend/internal/store"
)

// JanitorResult summarizes a cleanup pass.
type JanitorResult struct {
	VendorPairsSuppressed int `json:"vendor_pairs_suppressed"`
	ObservationFacts      int `json:"observation_facts_backfilled"`
	Errors                []string `json:"errors,omitempty"`
}

// Janitor keeps Supabase tidy so detective agents see clean evidence.
type Janitor struct {
	Store *store.Store
}

// Run executes safe maintenance jobs (no external spend unless you add enrich later).
func (j *Janitor) Run(ctx context.Context) JanitorResult {
	_ = ctx
	var r JanitorResult
	if j.Store == nil {
		r.Errors = append(r.Errors, "store nil")
		return r
	}
	// Build the registered-vendor set from the watched_sources table.
	registered := map[string]bool{}
	if all, err := j.Store.ListWatchedSources(false); err == nil {
		for _, s := range all {
			registered[strings.ToLower(s.Handle)] = true
		}
	}
	n, err := j.Store.SuppressVendorVendorCouples(registered)
	if err != nil {
		r.Errors = append(r.Errors, "suppress vendors: "+err.Error())
	} else {
		r.VendorPairsSuppressed = n
	}
	facts, err := j.Store.ExtractObservationFacts(500)
	if err != nil {
		r.Errors = append(r.Errors, "observation facts: "+err.Error())
	} else {
		r.ObservationFacts = facts
	}
	log.Printf("[janitor] vendor_pairs=%d obs_facts=%d errs=%v", r.VendorPairsSuppressed, r.ObservationFacts, r.Errors)
	return r
}
