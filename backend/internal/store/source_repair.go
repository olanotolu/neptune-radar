package store

import (
	"fmt"
	"time"

	"neptune-social-radar/backend/internal/ontology"
)

// StaleSourceThreshold is how long since last scan before a source is
// considered stale and worth a repair task. 72h = a source that hasn't
// yielded in 3 days deserves a look.
const StaleSourceThreshold = 72 * time.Hour

// CreateSourceRepairActions finds stale active sources and creates
// source_repair recommended_actions for them. Idempotent: if a pending
// source_repair action already exists for a source handle, it's skipped.
// Returns the number of new repair tasks created.
func (s *Store) CreateSourceRepairActions() (int, error) {
	cutoff := time.Now().UTC().Add(-StaleSourceThreshold)
	rows, err := s.DB.Query(
		`SELECT handle FROM watched_sources
		 WHERE active = true
		   AND (last_scanned_at IS NULL OR last_scanned_at < $1)
		   AND handle NOT IN (
		     SELECT (proposed_payload::json->>'handle')::text
		     FROM recommended_actions
		     WHERE action_type = 'source_repair' AND status = 'pending'
		   )`,
		cutoff)
	if err != nil {
		return 0, err
	}
	var handles []string
	for rows.Next() {
		var h string
		if err := rows.Scan(&h); err != nil {
			rows.Close()
			return 0, err
		}
		handles = append(handles, h)
	}
	rows.Close()

	created := 0
	for _, h := range handles {
		payload := fmt.Sprintf(`{"handle":"%s","reason":"stale_source"}`, h)
		_, err := s.DB.Exec(
			`INSERT INTO recommended_actions (id, action_type, proposed_payload, status, priority, reason)
			 VALUES ($1, 'source_repair', $2, 'pending', 5, 'Source not scanned in 72h')`,
			NewID("action"), payload)
		if err != nil {
			continue // don't fail the whole batch on one
		}
		created++
	}
	if created > 0 {
		s.Audit("system", "source_repair", "stale_sources_found",
			map[string]any{"count": created, "handles": handles}, "system:janitor", -1)
	}
	return created, nil
}

// SourceHealthSummary is the per-source metrics row for the repair center.
type SourceHealthSummary struct {
	Handle         string  `json:"handle"`
	SourceClass    string  `json:"source_class"`
	Active         bool    `json:"active"`
	LastScannedAt  *string `json:"last_scanned_at,omitempty"`
	ScanCouples    int     `json:"scan_couples"`
	ScanActions    int     `json:"scan_actions"`
	Stale          bool    `json:"stale"`
	State          string  `json:"state,omitempty"`
	City           string  `json:"city,omitempty"`
}

// ListSourceHealth returns per-source health metrics for the repair center.
func (s *Store) ListSourceHealth() ([]SourceHealthSummary, error) {
	cutoff := time.Now().UTC().Add(-StaleSourceThreshold)
	rows, err := s.DB.Query(
		`SELECT handle, source_class, active,
		   last_scanned_at, COALESCE(last_scan_couples,0), COALESCE(last_scan_actions,0),
		   COALESCE(state,''), COALESCE(city,''),
		   CASE WHEN last_scanned_at IS NULL OR last_scanned_at < $1 THEN true ELSE false END as stale
		 FROM watched_sources
		 ORDER BY stale DESC, last_scanned_at ASC NULLS FIRST`,
		cutoff)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []SourceHealthSummary
	for rows.Next() {
		var h SourceHealthSummary
		var lastScanStr *string
		if err := rows.Scan(&h.Handle, &h.SourceClass, &h.Active,
			&lastScanStr, &h.ScanCouples, &h.ScanActions,
			&h.State, &h.City, &h.Stale); err != nil {
			return nil, err
		}
		if lastScanStr != nil {
			h.LastScannedAt = lastScanStr
		}
		out = append(out, h)
	}
	return out, rows.Err()
}

// MarkSourceRepaired completes a source_repair action after the source has
// been re-scanned or re-enriched.
func (s *Store) MarkSourceRepaired(actionID, decidedBy string) error {
	return s.DecideAction(actionID, ontology.ActionApproved, decidedBy)
}
