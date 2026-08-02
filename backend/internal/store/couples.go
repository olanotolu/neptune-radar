package store

import (
	"database/sql"
	"fmt"
	"time"

	"neptune-social-radar/backend/internal/ontology"
)

// EnsureCouple returns the couple linking two persons, creating it (in a
// canonical person-id order) if it doesn't exist yet. The INSERT is
// concurrency-safe: the couples_pair_unique constraint (migration 0006) is
// the real guard, and ON CONFLICT returns the row the racing writer won.
func (s *Store) EnsureCouple(personAID, personBID string) (ontology.Couple, error) {
	a, b := personAID, personBID
	if a > b {
		a, b = b, a
	}
	var c ontology.Couple
	err := s.DB.QueryRow(
		`SELECT id, person_a_id, person_b_id, created_at, mistaken FROM couples WHERE person_a_id = $1 AND person_b_id = $2`, a, b,
	).Scan(&c.ID, &c.PersonAID, &c.PersonBID, &c.CreatedAt, &c.Mistaken)
	if err == nil {
		return c, nil
	}
	if err != sql.ErrNoRows {
		return c, err
	}
	c.ID = NewID("couple")
	c.PersonAID, c.PersonBID = a, b
	c.CreatedAt = time.Now().UTC()
	err = s.DB.QueryRow(
		`INSERT INTO couples (id, person_a_id, person_b_id, created_at) VALUES ($1, $2, $3, $4)
		 ON CONFLICT (person_a_id, person_b_id) DO NOTHING
		 RETURNING id`, c.ID, a, b, c.CreatedAt,
	).Scan(&c.ID)
	if err == sql.ErrNoRows {
		// We lost the race — fetch the winner's row (and its real id).
		err = s.DB.QueryRow(
			`SELECT id, person_a_id, person_b_id, created_at, mistaken FROM couples WHERE person_a_id = $1 AND person_b_id = $2`, a, b,
		).Scan(&c.ID, &c.PersonAID, &c.PersonBID, &c.CreatedAt, &c.Mistaken)
	}
	if err != nil {
		return c, err
	}
	return c, nil
}

func (s *Store) ListCouples() ([]ontology.Couple, error) {
	rows, err := s.DB.Query(
		`SELECT id, person_a_id, person_b_id, created_at,
		        COALESCE(inferred_city,''), COALESCE(inferred_region,''),
		        inferred_lat, inferred_lng, COALESCE(location_source,''),
		        mistaken, COALESCE(mistaken_reason,''), COALESCE(mistaken_by,''), mistaken_at
		 FROM couples ORDER BY created_at ASC, id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ontology.Couple
	for rows.Next() {
		var c ontology.Couple
		var lat, lng sql.NullFloat64
		var mistakenAt sql.NullTime
		if err := rows.Scan(&c.ID, &c.PersonAID, &c.PersonBID, &c.CreatedAt,
			&c.InferredCity, &c.InferredRegion, &lat, &lng, &c.LocationSource,
			&c.Mistaken, &c.MistakenReason, &c.MistakenBy, &mistakenAt); err != nil {
			return nil, err
		}
		if lat.Valid {
			v := lat.Float64
			c.InferredLat = &v
		}
		if lng.Valid {
			v := lng.Float64
			c.InferredLng = &v
		}
		if mistakenAt.Valid {
			t := mistakenAt.Time
			c.MistakenAt = &t
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *Store) GetCouple(id string) (ontology.Couple, error) {
	var c ontology.Couple
	var lat, lng sql.NullFloat64
	var mistakenAt sql.NullTime
	err := s.DB.QueryRow(
		`SELECT id, person_a_id, person_b_id, created_at,
		        COALESCE(inferred_city,''), COALESCE(inferred_region,''),
		        inferred_lat, inferred_lng, COALESCE(location_source,''),
		        mistaken, COALESCE(mistaken_reason,''), COALESCE(mistaken_by,''), mistaken_at
		 FROM couples WHERE id = $1`, id,
	).Scan(&c.ID, &c.PersonAID, &c.PersonBID, &c.CreatedAt,
		&c.InferredCity, &c.InferredRegion, &lat, &lng, &c.LocationSource,
		&c.Mistaken, &c.MistakenReason, &c.MistakenBy, &mistakenAt)
	if err != nil {
		return c, err
	}
	if lat.Valid {
		v := lat.Float64
		c.InferredLat = &v
	}
	if lng.Valid {
		v := lng.Float64
		c.InferredLng = &v
	}
	if mistakenAt.Valid {
		t := mistakenAt.Time
		c.MistakenAt = &t
	}
	return c, nil
}

// MarkCoupleMistaken is the human override path: a concierge marks a couple
// as NOT actually a couple (identity resolution was wrong). The scorer
// checks this before creating new hypotheses for the couple, so the
// override is respected permanently — not just for the current hypothesis.
func (s *Store) MarkCoupleMistaken(coupleID, reason, decidedBy string) error {
	_, err := s.DB.Exec(
		`UPDATE couples SET mistaken = TRUE, mistaken_reason = $2, mistaken_by = $3, mistaken_at = now() WHERE id = $1`,
		coupleID, reason, decidedBy)
	if err != nil {
		return err
	}
	// Also reject all pending hypotheses for this couple — a mistaken couple
	// has no valid hypotheses.
	_, err = s.DB.Exec(
		`UPDATE life_event_hypotheses SET status = 'rejected' WHERE couple_id = $1 AND status NOT IN ('rejected','expired')`,
		coupleID)
	if err != nil {
		return fmt.Errorf("mark couple mistaken: reject hypotheses: %w", err)
	}
	s.Audit("couple", coupleID, "marked_mistaken",
		map[string]any{"reason": reason, "decided_by": decidedBy}, decidedBy, -1)
	return nil
}

// GetCoupleForAccountPair returns the couple linking the persons behind two
// accounts, or sql.ErrNoRows if either account is person-less or no couple
// links them. It creates nothing — used to attach a third party's post (a
// photographer's announcement, a jeweler's ad) to a couple the ontology
// already knows.
func (s *Store) GetCoupleForAccountPair(accountAID, accountBID string) (ontology.Couple, error) {
	a, err := s.GetAccount(accountAID)
	if err != nil {
		return ontology.Couple{}, err
	}
	b, err := s.GetAccount(accountBID)
	if err != nil {
		return ontology.Couple{}, err
	}
	if a.PersonID == "" || b.PersonID == "" || a.PersonID == b.PersonID {
		return ontology.Couple{}, sql.ErrNoRows
	}
	x, y := a.PersonID, b.PersonID
	if x > y {
		x, y = y, x
	}
	var c ontology.Couple
	err = s.DB.QueryRow(
		`SELECT id, person_a_id, person_b_id, created_at, mistaken FROM couples WHERE person_a_id = $1 AND person_b_id = $2`, x, y,
	).Scan(&c.ID, &c.PersonAID, &c.PersonBID, &c.CreatedAt, &c.Mistaken)
	if err != nil {
		return c, err
	}
	return c, nil
}

// CurrentRelationship returns the row with effective_to IS NULL, or
// sql.ErrNoRows if the couple has no relationship state yet.
func (s *Store) CurrentRelationship(coupleID string) (ontology.Relationship, error) {
	var r ontology.Relationship
	var effTo sql.NullTime
	var supersededBy sql.NullString
	err := s.DB.QueryRow(
		`SELECT id, couple_id, stage, confidence, effective_from, effective_to, COALESCE(superseded_by,''), automation_paused, visibility_scope
		 FROM relationships WHERE couple_id = $1 AND effective_to IS NULL ORDER BY effective_from DESC LIMIT 1`, coupleID,
	).Scan(&r.ID, &r.CoupleID, &r.Stage, &r.Confidence, &r.EffectiveFrom, &effTo, &supersededBy, &r.AutomationPaused, &r.VisibilityScope)
	if err != nil {
		return r, err
	}
	if effTo.Valid {
		t := effTo.Time
		r.EffectiveTo = &t
	}
	r.SupersededBy = supersededBy.String
	return r, nil
}

// TransitionRelationship closes out the current relationship row (if any)
// and inserts a new one superseding it. This is the only place relationship
// stage changes; callers are policy code, never the LLM directly.
func (s *Store) TransitionRelationship(coupleID string, stage ontology.RelationshipStage, confidence float64, scope ontology.VisibilityScope, automationPaused bool) (ontology.Relationship, error) {
	newRow := ontology.Relationship{
		ID:               NewID("rel"),
		CoupleID:         coupleID,
		Stage:            stage,
		Confidence:       confidence,
		EffectiveFrom:    time.Now().UTC(),
		VisibilityScope:  scope,
		AutomationPaused: automationPaused,
	}
	tx, err := s.DB.Begin()
	if err != nil {
		return newRow, err
	}
	defer tx.Rollback()

	var priorID string
	err = tx.QueryRow(
		`SELECT id FROM relationships WHERE couple_id = $1 AND effective_to IS NULL ORDER BY effective_from DESC LIMIT 1`, coupleID,
	).Scan(&priorID)
	if err != nil && err != sql.ErrNoRows {
		return newRow, err
	}
	hasPrior := err == nil

	// Insert the new row before pointing the old one at it — superseded_by
	// has a foreign key to relationships(id), checked per-statement.
	_, err = tx.Exec(
		`INSERT INTO relationships (id, couple_id, stage, confidence, effective_from, visibility_scope, automation_paused)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		newRow.ID, newRow.CoupleID, newRow.Stage, newRow.Confidence, newRow.EffectiveFrom, newRow.VisibilityScope, newRow.AutomationPaused,
	)
	if err != nil {
		return newRow, err
	}

	if hasPrior {
		if _, err := tx.Exec(`UPDATE relationships SET effective_to = $1, superseded_by = $2 WHERE id = $3`,
			newRow.EffectiveFrom, newRow.ID, priorID); err != nil {
			return newRow, err
		}
	}
	return newRow, tx.Commit()
}

// SetAutomationPaused flips the automation_paused flag on a couple's current
// relationship without changing stage/confidence/scope — the same
// TransitionRelationship path the operator uses, just with the paused bit
// toggled. Returns the new relationship row.
func (s *Store) SetAutomationPaused(coupleID string, paused bool) (ontology.Relationship, error) {
	current, err := s.CurrentRelationship(coupleID)
	if err != nil {
		return ontology.Relationship{}, err
	}
	return s.TransitionRelationship(coupleID, current.Stage, current.Confidence, current.VisibilityScope, paused)
}

// BulkUpdateCouples applies the same action to many couples at once.
// action is "pause", "resume", or "suppress". Returns the number of couples
// successfully updated. Failures on individual couples are counted as skipped
// (not an error) so a single bad id doesn't abort the batch.
func (s *Store) BulkUpdateCouples(ids []string, action string, reason string) (int, error) {
	n := 0
	for _, id := range ids {
		var err error
		switch action {
		case "pause":
			_, err = s.SetAutomationPaused(id, true)
		case "resume":
			_, err = s.SetAutomationPaused(id, false)
		case "suppress":
			err = s.SuppressCouple(id, reason)
		}
		if err != nil {
			continue
		}
		n++
	}
	return n, nil
}

func (s *Store) RelationshipHistory(coupleID string) ([]ontology.Relationship, error) {
	rows, err := s.DB.Query(
		`SELECT id, couple_id, stage, confidence, effective_from, effective_to, automation_paused, visibility_scope
		 FROM relationships WHERE couple_id = $1 ORDER BY effective_from ASC, id ASC`, coupleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ontology.Relationship
	for rows.Next() {
		var r ontology.Relationship
		var effTo sql.NullTime
		if err := rows.Scan(&r.ID, &r.CoupleID, &r.Stage, &r.Confidence, &r.EffectiveFrom, &effTo, &r.AutomationPaused, &r.VisibilityScope); err != nil {
			return nil, err
		}
		if effTo.Valid {
			t := effTo.Time
			r.EffectiveTo = &t
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
