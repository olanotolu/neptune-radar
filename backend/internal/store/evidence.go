package store

import (
	"database/sql"

	"neptune-social-radar/backend/internal/ontology"
)

func (s *Store) InsertEvidence(e ontology.Evidence) (ontology.Evidence, error) {
	if e.ID == "" {
		e.ID = NewID("ev")
	}
	var obsID any
	if e.ObservationID != "" {
		obsID = e.ObservationID
	}
	_, err := s.DB.Exec(
		`INSERT INTO evidence (id, hypothesis_id, observation_id, kind, description, weight, confirmed)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		e.ID, e.HypothesisID, obsID, e.Kind, e.Description, e.Weight, e.Confirmed,
	)
	return e, err
}

// UpsertEvidenceKind keeps at most one row per (hypothesis_id, kind) —
// evidence represents the system's *current* understanding of a supporting
// fact, not a running log of every fluctuation (that's what audit_events is
// for). Atomic on the evidence_hypothesis_kind_unique constraint (migration
// 0006) — the old SELECT-then-INSERT raced into duplicate kind rows that
// then diverged, updating whichever row the planner found first.
func (s *Store) UpsertEvidenceKind(hypothesisID, kind, description string, weight float64) (ontology.Evidence, error) {
	e := ontology.Evidence{
		ID: NewID("ev"), HypothesisID: hypothesisID, Kind: kind,
		Description: description, Weight: weight,
	}
	err := s.DB.QueryRow(
		`INSERT INTO evidence (id, hypothesis_id, kind, description, weight, confirmed)
		 VALUES ($1, $2, $3, $4, $5, FALSE)
		 ON CONFLICT (hypothesis_id, kind) DO UPDATE SET
		   description = EXCLUDED.description, weight = EXCLUDED.weight
		 RETURNING id`,
		e.ID, e.HypothesisID, e.Kind, e.Description, e.Weight,
	).Scan(&e.ID)
	return e, err
}

// DeleteEvidenceKind removes a no-longer-supported signal, e.g. an
// "unfollow_detected" row once the follow is restored.
func (s *Store) DeleteEvidenceKind(hypothesisID, kind string) error {
	_, err := s.DB.Exec(`DELETE FROM evidence WHERE hypothesis_id = $1 AND kind = $2`, hypothesisID, kind)
	return err
}

func (s *Store) EvidenceForHypothesis(hypothesisID string) ([]ontology.Evidence, error) {
	rows, err := s.DB.Query(
		`SELECT id, hypothesis_id, COALESCE(observation_id,''), kind, description, weight, confirmed, created_at
		 FROM evidence WHERE hypothesis_id = $1 ORDER BY created_at ASC, id ASC`, hypothesisID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ontology.Evidence
	for rows.Next() {
		var e ontology.Evidence
		var obsID sql.NullString
		if err := rows.Scan(&e.ID, &e.HypothesisID, &obsID, &e.Kind, &e.Description, &e.Weight, &e.Confirmed, &e.CreatedAt); err != nil {
			return nil, err
		}
		e.ObservationID = obsID.String
		out = append(out, e)
	}
	return out, rows.Err()
}
