package store

import (
	"database/sql"
	"errors"
	"strconv"
	"time"

	"neptune-social-radar/backend/internal/ontology"
)

// ErrDuplicateObservation is returned when the same (monitor,
// external_event_id) has already been ingested — the idempotency guard for
// duplicate event delivery.
var ErrDuplicateObservation = errors.New("duplicate observation")

func (s *Store) InsertObservation(o ontology.SocialObservation) (ontology.SocialObservation, error) {
	if o.ID == "" {
		o.ID = NewID("obs")
	}
	if o.IngestedAt.IsZero() {
		o.IngestedAt = time.Now().UTC()
	}
	var accountID any
	if o.AccountID != "" {
		accountID = o.AccountID
	}
	_, err := s.DB.Exec(
		`INSERT INTO social_observations
		 (id, monitor, external_event_id, account_id, observation_type, raw_payload, observed_at, ingested_at, source, freshness_seconds, consent_scope)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		o.ID, o.Monitor, o.ExternalEventID, accountID, o.ObservationType, o.RawPayload,
		o.ObservedAt.UTC(), o.IngestedAt.UTC(), o.Source, o.FreshnessSeconds, o.ConsentScope,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return ontology.SocialObservation{}, ErrDuplicateObservation
		}
		return ontology.SocialObservation{}, err
	}
	return o, nil
}

// empty monitor lists across all monitors (the live feed).
func (s *Store) ListObservations(monitor string, limit int) ([]ontology.SocialObservation, error) {
	q := `SELECT id, monitor, external_event_id, account_id, observation_type, raw_payload, observed_at, ingested_at, source, COALESCE(freshness_seconds,0), consent_scope
		 FROM social_observations`
	args := []any{}
	if monitor != "" {
		q += ` WHERE monitor = $1`
		args = append(args, monitor)
	}
	q += ` ORDER BY observed_at DESC, id DESC LIMIT ` + strconv.Itoa(limit)
	rows, err := s.DB.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ontology.SocialObservation
	for rows.Next() {
		var o ontology.SocialObservation
		var accountID sql.NullString
		if err := rows.Scan(&o.ID, &o.Monitor, &o.ExternalEventID, &accountID, &o.ObservationType, &o.RawPayload, &o.ObservedAt, &o.IngestedAt, &o.Source, &o.FreshnessSeconds, &o.ConsentScope); err != nil {
			return nil, err
		}
		o.AccountID = accountID.String
		out = append(out, o)
	}
	return out, rows.Err()
}

// ObservationExists reports whether a provider-native event ID has been
// ingested under ANY monitor — the same post can surface in a hashtag batch
// and a vendor feed, and it must only be processed once.
func (s *Store) ObservationExists(externalEventID string) (bool, error) {
	var n int
	err := s.DB.QueryRow(`SELECT COUNT(*) FROM social_observations WHERE external_event_id = $1`,
		externalEventID).Scan(&n)
	return n > 0, err
}
