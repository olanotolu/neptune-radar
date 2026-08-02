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
	VendorPairsSuppressed int                 `json:"vendor_pairs_suppressed"`
	ObservationFacts      int                 `json:"observation_facts_backfilled"`
	RetentionPurged       []store.PurgeResult `json:"retention_purged,omitempty"`
	SourceRepairActions   int                 `json:"source_repair_actions,omitempty"`
	Errors                []string            `json:"errors,omitempty"`
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
	// Purge expired rows per retention_classes config.
	purged, err := j.Store.PurgeExpired(ctx)
	if err != nil {
		r.Errors = append(r.Errors, "retention purge: "+err.Error())
	} else {
		r.RetentionPurged = purged
	}
	// Create repair tasks for stale sources so they appear in the work queue.
	repairs, err := j.Store.CreateSourceRepairActions()
	if err != nil {
		r.Errors = append(r.Errors, "source repair: "+err.Error())
	} else {
		r.SourceRepairActions = repairs
	}
	log.Printf("[janitor] vendor_pairs=%d obs_facts=%d purge=%v repairs=%d errs=%v",
		r.VendorPairsSuppressed, r.ObservationFacts, r.RetentionPurged, r.SourceRepairActions, r.Errors)
	return r
}
