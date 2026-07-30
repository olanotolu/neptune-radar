package store

import (
	"database/sql"
	"time"

	"neptune-social-radar/backend/internal/ontology"
)

// UpsertEdge records or updates a follows/tagged_with/mentioned_by edge. For
// `follows` this is how we detect an unfollow: the same (from,to,kind) row
// flips `active` rather than being deleted, preserving history. Backed by
// the edges_unique_triple constraint (migration 0006): one row per triple,
// upserted atomically — the old SELECT-then-INSERT raced concurrent workers
// into duplicate rows that then diverged on opposite active values.
func (s *Store) UpsertEdge(kind ontology.EdgeKind, fromAccountID, toAccountID string, active bool, observedAt time.Time, sourceObservationID string) (ontology.Edge, error) {
	e := ontology.Edge{
		ID: NewID("edge"), Kind: kind, FromAccountID: fromAccountID, ToAccountID: toAccountID,
		Active: active, FirstObservedAt: observedAt, LastObservedAt: observedAt, SourceObservationID: sourceObservationID,
	}
	err := s.DB.QueryRow(
		`INSERT INTO edges (id, kind, from_account_id, to_account_id, active, first_observed_at, last_observed_at, source_observation_id)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 ON CONFLICT (kind, from_account_id, to_account_id) DO UPDATE SET
		   active = EXCLUDED.active,
		   last_observed_at = EXCLUDED.last_observed_at,
		   source_observation_id = EXCLUDED.source_observation_id
		 RETURNING id`,
		e.ID, e.Kind, e.FromAccountID, e.ToAccountID, e.Active,
		e.FirstObservedAt.UTC(), e.LastObservedAt.UTC(), e.SourceObservationID,
	).Scan(&e.ID)
	return e, err
}

func (s *Store) GetEdge(kind ontology.EdgeKind, fromAccountID, toAccountID string) (ontology.Edge, error) {
	var e ontology.Edge
	var src sql.NullString
	err := s.DB.QueryRow(
		`SELECT id, kind, from_account_id, to_account_id, active, first_observed_at, last_observed_at, COALESCE(source_observation_id,'')
		 FROM edges WHERE kind = $1 AND from_account_id = $2 AND to_account_id = $3`, kind, fromAccountID, toAccountID,
	).Scan(&e.ID, &e.Kind, &e.FromAccountID, &e.ToAccountID, &e.Active, &e.FirstObservedAt, &e.LastObservedAt, &src)
	if err != nil {
		return e, err
	}
	e.SourceObservationID = src.String
	return e, nil
}

// FollowState reports the current stored follows-edge state between two
// handles, or nil if no edge exists yet (first check counts as a change).
// The follow-check worker reads this before emitting an event so unchanged
// state never mints observation rows and same-day flaps are never deduped
// away.
func (s *Store) FollowState(fromHandle, toHandle string) (*bool, error) {
	var active bool
	err := s.DB.QueryRow(
		`SELECT e.active FROM edges e
		 JOIN social_accounts fa ON fa.id = e.from_account_id
		 JOIN social_accounts ta ON ta.id = e.to_account_id
		 WHERE e.kind = 'follows' AND fa.handle = $1 AND ta.handle = $2`,
		fromHandle, toHandle,
	).Scan(&active)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &active, nil
}

func (s *Store) EdgesForAccount(accountID string) ([]ontology.Edge, error) {
	rows, err := s.DB.Query(
		`SELECT id, kind, from_account_id, to_account_id, active, first_observed_at, last_observed_at, COALESCE(source_observation_id,'')
		 FROM edges WHERE from_account_id = $1 OR to_account_id = $2`, accountID, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ontology.Edge
	for rows.Next() {
		var e ontology.Edge
		var src sql.NullString
		if err := rows.Scan(&e.ID, &e.Kind, &e.FromAccountID, &e.ToAccountID, &e.Active, &e.FirstObservedAt, &e.LastObservedAt, &src); err != nil {
			return nil, err
		}
		e.SourceObservationID = src.String
		out = append(out, e)
	}
	return out, rows.Err()
}
