package store

import (
	"database/sql"
	"errors"
	"time"

	"neptune-social-radar/backend/internal/ontology"
)

func (s *Store) CreateRecommendedAction(a ontology.RecommendedAction) (ontology.RecommendedAction, error) {
	if a.ID == "" {
		a.ID = NewID("action")
	}
	if a.Status == "" {
		a.Status = ontology.ActionPending
	}
	var hypID, caseID any
	if a.HypothesisID != "" {
		hypID = a.HypothesisID
	}
	if a.CaseID != "" {
		caseID = a.CaseID
	}
	_, err := s.DB.Exec(
		`INSERT INTO recommended_actions (id, hypothesis_id, case_id, action_type, proposed_payload, status)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		a.ID, hypID, caseID, a.ActionType, a.ProposedPayload, a.Status,
	)
	if err != nil {
		return a, err
	}
	return s.GetAction(a.ID)
}

// LatestPendingActionForCouple returns the newest pending recommended_action
// linked to a couple via its hypothesis, if any.
func (s *Store) LatestPendingActionForCouple(coupleID string) (ontology.RecommendedAction, error) {
	var a ontology.RecommendedAction
	var hypID, caseID, decidedBy sql.NullString
	var decidedAt sql.NullTime
	err := s.DB.QueryRow(
		`SELECT ra.id, COALESCE(ra.hypothesis_id,''), COALESCE(ra.case_id,''), ra.action_type,
		        COALESCE(ra.proposed_payload,''), ra.status, ra.created_at, ra.decided_at, COALESCE(ra.decided_by,'')
		 FROM recommended_actions ra
		 JOIN life_event_hypotheses h ON h.id = ra.hypothesis_id
		 WHERE h.couple_id = $1 AND ra.status = 'pending'
		 ORDER BY ra.created_at DESC LIMIT 1`, coupleID,
	).Scan(&a.ID, &hypID, &caseID, &a.ActionType, &a.ProposedPayload, &a.Status, &a.CreatedAt, &decidedAt, &decidedBy)
	if err != nil {
		return a, err
	}
	a.HypothesisID, a.CaseID, a.DecidedBy = hypID.String, caseID.String, decidedBy.String
	if decidedAt.Valid {
		t := decidedAt.Time
		a.DecidedAt = &t
	}
	return a, nil
}

func (s *Store) GetAction(id string) (ontology.RecommendedAction, error) {
	var a ontology.RecommendedAction
	var hypID, caseID, decidedBy, owner, reason sql.NullString
	var decidedAt, snoozeUntil sql.NullTime
	err := s.DB.QueryRow(
		`SELECT id, COALESCE(hypothesis_id,''), COALESCE(case_id,''), action_type, COALESCE(proposed_payload,''),
		   status, created_at, decided_at, COALESCE(decided_by,''),
		   priority, COALESCE(owner,''), snooze_until, COALESCE(reason,'')
		 FROM recommended_actions WHERE id = $1`, id,
	).Scan(&a.ID, &hypID, &caseID, &a.ActionType, &a.ProposedPayload, &a.Status, &a.CreatedAt, &decidedAt, &decidedBy,
		&a.Priority, &owner, &snoozeUntil, &reason)
	if err != nil {
		return a, err
	}
	a.HypothesisID, a.CaseID, a.DecidedBy, a.Owner, a.Reason = hypID.String, caseID.String, decidedBy.String, owner.String, reason.String
	if decidedAt.Valid {
		t := decidedAt.Time
		a.DecidedAt = &t
	}
	if snoozeUntil.Valid {
		t := snoozeUntil.Time
		a.SnoozeUntil = &t
	}
	return a, nil
}

// CountPendingActions returns the number of recommended_actions still
// awaiting a human decision. Used by the health check.
func (s *Store) CountPendingActions() (int, error) {
	var n int
	err := s.DB.QueryRow(`SELECT COUNT(*) FROM recommended_actions WHERE status = 'pending'`).Scan(&n)
	return n, err
}

func (s *Store) ListActions(status string) ([]ontology.RecommendedAction, error) {
	q := `SELECT id, COALESCE(hypothesis_id,''), COALESCE(case_id,''), action_type, COALESCE(proposed_payload,''),
	   status, created_at, decided_at, COALESCE(decided_by,''),
	   priority, COALESCE(owner,''), snooze_until, COALESCE(reason,'')
	 FROM recommended_actions`
	args := []any{}
	if status != "" {
		q += ` WHERE status = $1`
		args = append(args, status)
	}
	q += ` ORDER BY priority DESC, created_at DESC, id DESC`
	rows, err := s.DB.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ontology.RecommendedAction
	for rows.Next() {
		var a ontology.RecommendedAction
		var hypID, caseID, decidedBy, owner, reason sql.NullString
		var decidedAt, snoozeUntil sql.NullTime
		if err := rows.Scan(&a.ID, &hypID, &caseID, &a.ActionType, &a.ProposedPayload, &a.Status, &a.CreatedAt, &decidedAt, &decidedBy,
			&a.Priority, &owner, &snoozeUntil, &reason); err != nil {
			return nil, err
		}
		a.HypothesisID, a.CaseID, a.DecidedBy, a.Owner, a.Reason = hypID.String, caseID.String, decidedBy.String, owner.String, reason.String
		if decidedAt.Valid {
			t := decidedAt.Time
			a.DecidedAt = &t
		}
		if snoozeUntil.Valid {
			t := snoozeUntil.Time
			a.SnoozeUntil = &t
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// AssignAction sets the owner and optional priority on a recommended action.
func (s *Store) AssignAction(id, owner string, priority int) error {
	_, err := s.DB.Exec(`UPDATE recommended_actions SET owner = $1, priority = $2 WHERE id = $3`, owner, priority, id)
	return err
}

// SnoozeAction defers an action until a future time with a reason.
func (s *Store) SnoozeAction(id string, until time.Time, reason string) error {
	_, err := s.DB.Exec(`UPDATE recommended_actions SET snooze_until = $1, reason = $2 WHERE id = $3`, until, reason, id)
	return err
}

func (s *Store) DecideAction(id string, status ontology.ActionStatus, decidedBy string) error {
	_, err := s.DB.Exec(`UPDATE recommended_actions SET status = $1, decided_at = $2, decided_by = $3 WHERE id = $4`,
		status, time.Now().UTC(), decidedBy, id)
	return err
}

// ErrActionNotPending is returned when a decision is attempted on an action
// that was already decided — the guard against double-approve creating two
// leads/cases, and against ignore-after-approve rewriting history.
var ErrActionNotPending = errors.New("action is not pending")

// DecideActionIfPending is DecideAction with a state guard: the UPDATE only
// lands when the action is still pending, and the row-count check makes a
// double decision an error instead of silent history corruption.
func (s *Store) DecideActionIfPending(id string, status ontology.ActionStatus, decidedBy string) error {
	res, err := s.DB.Exec(
		`UPDATE recommended_actions SET status = $1, decided_at = $2, decided_by = $3 WHERE id = $4 AND status = 'pending'`,
		status, time.Now().UTC(), decidedBy, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrActionNotPending
	}
	return nil
}

func (s *Store) CreateExecutedAction(e ontology.ExecutedAction) (ontology.ExecutedAction, error) {
	if e.ID == "" {
		e.ID = NewID("exec")
	}
	_, err := s.DB.Exec(
		`INSERT INTO executed_actions (id, recommended_action_id, result, detail, verified) VALUES ($1, $2, $3, $4, $5)`,
		e.ID, e.RecommendedActionID, e.Result, e.Detail, e.Verified,
	)
	return e, err
}

func (s *Store) SetExecutedVerified(id string, verified bool) error {
	_, err := s.DB.Exec(`UPDATE executed_actions SET verified = $1 WHERE id = $2`, verified, id)
	return err
}
