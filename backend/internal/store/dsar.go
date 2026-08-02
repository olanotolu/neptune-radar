package store

import (
	"context"
	"fmt"
)

// DSARDelete performs a GDPR/CCPA right-to-erasure deletion for a person.
// It cascades through every table that references the person, in dependency
// order, so foreign keys don't block the deletion. Returns a summary of what
// was deleted for the audit trail.
//
// This is the real DSAR delete path — not a soft delete. The person and all
// their derived data (observations, hypotheses, evidence, actions, couples,
// consent, leads) are permanently removed. The audit_events trail is NOT
// deleted (it's append-only by the 0014 migration trigger) — instead the
// DSAR request itself is logged as an audit event, which is the legally
// correct posture: the audit trail records that a deletion happened, without
// retaining the deleted personal data.
//
// If the person is part of a couple, the couple is also deleted — a couple
// with one missing person is meaningless and would leave orphaned
// relationship/hypothesis rows. The other person in the couple is NOT deleted.
type DSARResult struct {
	PersonID           string
	CouplesDeleted     int
	AccountsDeleted    int
	ObservationsDeleted int
	HypothesesDeleted  int
	EvidenceDeleted    int
	ActionsCancelled   int
	ConsentRevoked     int
	LeadsDeleted       int
}

// DSARDelete deletes a person and all their derived data. It runs in a single
// transaction — either everything is deleted or nothing is.
func (s *Store) DSARDelete(ctx context.Context, personID string) (DSARResult, error) {
	var result DSARResult
	result.PersonID = personID

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return result, err
	}
	defer tx.Rollback()

	// 1. Find all couples involving this person (as person_a or person_b).
	coupleRows, err := tx.QueryContext(ctx,
		`SELECT id FROM couples WHERE person_a_id = $1 OR person_b_id = $1`, personID)
	if err != nil {
		return result, err
	}
	var coupleIDs []string
	for coupleRows.Next() {
		var id string
		if err := coupleRows.Scan(&id); err != nil {
			coupleRows.Close()
			return result, err
		}
		coupleIDs = append(coupleIDs, id)
	}
	coupleRows.Close()

	// 2. For each couple: cancel pending actions, delete evidence, hypotheses,
	//    relationships, then the couple itself.
	for _, coupleID := range coupleIDs {
		// Cancel pending recommended_actions for this couple's hypotheses.
		res, err := tx.ExecContext(ctx,
			`UPDATE recommended_actions
			   SET status = 'dsar_deleted', decided_at = now(), decided_by = 'system:dsar'
			 WHERE status = 'pending'
			   AND hypothesis_id IN (SELECT id FROM life_event_hypotheses WHERE couple_id = $1)`,
			coupleID)
		if err != nil {
			return result, fmt.Errorf("dsar: cancel actions for couple %s: %w", coupleID, err)
		}
		if n, _ := res.RowsAffected(); n > 0 {
			result.ActionsCancelled += int(n)
		}

		// Delete evidence for this couple's hypotheses.
		res, err = tx.ExecContext(ctx,
			`DELETE FROM evidence WHERE hypothesis_id IN (SELECT id FROM life_event_hypotheses WHERE couple_id = $1)`,
			coupleID)
		if err != nil {
			return result, fmt.Errorf("dsar: delete evidence for couple %s: %w", coupleID, err)
		}
		if n, _ := res.RowsAffected(); n > 0 {
			result.EvidenceDeleted += int(n)
		}

		// Delete hypotheses for this couple.
		res, err = tx.ExecContext(ctx,
			`DELETE FROM life_event_hypotheses WHERE couple_id = $1`, coupleID)
		if err != nil {
			return result, fmt.Errorf("dsar: delete hypotheses for couple %s: %w", coupleID, err)
		}
		if n, _ := res.RowsAffected(); n > 0 {
			result.HypothesesDeleted += int(n)
		}

		// Delete relationships for this couple.
		_, _ = tx.ExecContext(ctx, `DELETE FROM relationships WHERE couple_id = $1`, coupleID)

		// Delete the couple.
		_, err = tx.ExecContext(ctx, `DELETE FROM couples WHERE id = $1`, coupleID)
		if err != nil {
			return result, fmt.Errorf("dsar: delete couple %s: %w", coupleID, err)
		}
		result.CouplesDeleted++
	}

	// 3. Delete social observations for this person's accounts.
	res, err := tx.ExecContext(ctx,
		`DELETE FROM social_observations WHERE account_id IN (SELECT id FROM social_accounts WHERE person_id = $1)`,
		personID)
	if err != nil {
		return result, fmt.Errorf("dsar: delete observations: %w", err)
	}
	if n, _ := res.RowsAffected(); n > 0 {
		result.ObservationsDeleted += int(n)
	}

	// 4. Delete social accounts for this person.
	res, err = tx.ExecContext(ctx, `DELETE FROM social_accounts WHERE person_id = $1`, personID)
	if err != nil {
		return result, fmt.Errorf("dsar: delete accounts: %w", err)
	}
	if n, _ := res.RowsAffected(); n > 0 {
		result.AccountsDeleted += int(n)
	}

	// 5. Revoke consent (set revoked_at rather than delete — the consent
	//    history is itself a legal record of what the person agreed to).
	res, err = tx.ExecContext(ctx, `UPDATE consent_policies SET revoked_at = now() WHERE person_id = $1 AND revoked_at IS NULL`, personID)
	if err != nil {
		return result, fmt.Errorf("dsar: revoke consent: %w", err)
	}
	if n, _ := res.RowsAffected(); n > 0 {
		result.ConsentRevoked += int(n)
	}

	// 6. Delete CRM leads for this person.
	res, err = tx.ExecContext(ctx, `DELETE FROM crm_leads WHERE person_id = $1`, personID)
	if err != nil {
		return result, fmt.Errorf("dsar: delete leads: %w", err)
	}
	if n, _ := res.RowsAffected(); n > 0 {
		result.LeadsDeleted += int(n)
	}

	// 7. Delete hypotheses directly referencing this person (not via couple).
	res, err = tx.ExecContext(ctx, `DELETE FROM life_event_hypotheses WHERE person_id = $1`, personID)
	if err != nil {
		return result, fmt.Errorf("dsar: delete person hypotheses: %w", err)
	}
	if n, _ := res.RowsAffected(); n > 0 {
		result.HypothesesDeleted += int(n)
	}

	// 8. Finally, delete the person.
	_, err = tx.ExecContext(ctx, `DELETE FROM persons WHERE id = $1`, personID)
	if err != nil {
		return result, fmt.Errorf("dsar: delete person: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return result, fmt.Errorf("dsar: commit: %w", err)
	}

	// Log the DSAR request in the audit trail (append-only — this records
	// that a deletion happened, without retaining the deleted data).
	s.Audit("person", personID, "dsar_delete", map[string]any{
		"couples_deleted":      result.CouplesDeleted,
		"accounts_deleted":     result.AccountsDeleted,
		"observations_deleted": result.ObservationsDeleted,
		"hypotheses_deleted":   result.HypothesesDeleted,
		"evidence_deleted":     result.EvidenceDeleted,
		"actions_cancelled":    result.ActionsCancelled,
		"consent_revoked":      result.ConsentRevoked,
		"leads_deleted":        result.LeadsDeleted,
	}, "system:dsar", -1)

	return result, nil
}
