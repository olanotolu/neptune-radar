package store

import (
	"database/sql"
	"strings"
	"time"

	"neptune-social-radar/backend/internal/ontology"
)

// FenrisEventRow is the stored representation of a Fenris life event, with
// couple linkage and cross-validation status for the dashboard.
type FenrisEventRow struct {
	ID             string    `json:"id"`
	EventType      string    `json:"event_type"`
	PersonName     string    `json:"person_name"`
	State          string    `json:"state"`
	City           string    `json:"city"`
	EventDate      time.Time `json:"event_date"`
	Confidence     float64   `json:"confidence"`
	CoupleID       string    `json:"couple_id,omitempty"`
	CrossValidated bool      `json:"cross_validated"`
	IngestedAt     time.Time `json:"ingested_at"`
}

// InsertFenrisEvent stores a Fenris life event, deduped by external_id.
// Returns ErrDuplicateFenrisEvent if already ingested.
var ErrDuplicateFenrisEvent = ErrDuplicateObservation

func (s *Store) InsertFenrisEvent(ev ontology.LifeEvent, externalID, coupleID string, crossValidated bool) (FenrisEventRow, error) {
	row := FenrisEventRow{
		ID: NewID("fenris"), EventType: ev.EventType, PersonName: ev.PersonName,
		State: ev.State, City: ev.City, EventDate: ev.EventDate, Confidence: ev.Confidence,
		CoupleID: coupleID, CrossValidated: crossValidated, IngestedAt: time.Now().UTC(),
	}
	var cid any
	if coupleID != "" {
		cid = coupleID
	}
	_, err := s.DB.Exec(
		`INSERT INTO fenris_life_events
		 (id, external_id, event_type, person_name, household_id, address, city, state, zip, event_date, confidence, couple_id, cross_validated, ingested_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		row.ID, externalID, ev.EventType, ev.PersonName, nullIfEmpty(ev.HouseholdID),
		nullIfEmpty(ev.Address), nullIfEmpty(ev.City), nullIfEmpty(ev.State), nullIfEmpty(ev.Zip),
		ev.EventDate, ev.Confidence, cid, crossValidated, row.IngestedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return FenrisEventRow{}, ErrDuplicateFenrisEvent
		}
		return FenrisEventRow{}, err
	}
	return row, nil
}

// ListFenrisEvents returns recent Fenris life events for the dashboard.
func (s *Store) ListFenrisEvents(limit int) ([]FenrisEventRow, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := s.DB.Query(
		`SELECT id, event_type, person_name, COALESCE(state,''), COALESCE(city,''),
		        event_date, confidence, COALESCE(couple_id,''), cross_validated, ingested_at
		 FROM fenris_life_events ORDER BY event_date DESC, id DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []FenrisEventRow
	for rows.Next() {
		var r FenrisEventRow
		if err := rows.Scan(&r.ID, &r.EventType, &r.PersonName, &r.State, &r.City,
			&r.EventDate, &r.Confidence, &r.CoupleID, &r.CrossValidated, &r.IngestedAt); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// FindCoupleByNameState matches a Fenris event to an existing couple by
// person name + state. Checks both person_a and person_b display names.
// ponytail: ILIKE on display_name + inferred_region — no fuzzy matching.
// Ceiling: common names in large states will false-positive; upgrade path =
// add household_id or address matching when Fenris provides them consistently.
func (s *Store) FindCoupleByNameState(name, state string) (ontology.Couple, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return ontology.Couple{}, sql.ErrNoRows
	}
	var c ontology.Couple
	err := s.DB.QueryRow(
		`SELECT c.id, c.person_a_id, c.person_b_id, c.created_at, c.fenris_validated
		 FROM couples c
		 JOIN persons pa ON pa.id = c.person_a_id
		 LEFT JOIN persons pb ON pb.id = c.person_b_id
		 WHERE (pa.display_name ILIKE $1 OR pb.display_name ILIKE $1)
		   AND COALESCE(c.inferred_region, '') = $2
		 ORDER BY c.created_at DESC LIMIT 1`,
		"%"+name+"%", strings.ToUpper(state),
	).Scan(&c.ID, &c.PersonAID, &c.PersonBID, &c.CreatedAt, &c.FenrisValidated)
	return c, err
}

// SetFenrisValidated marks a couple as independently confirmed by a Fenris
// life event — the cross-validation flag that boosts prep scoring.
func (s *Store) SetFenrisValidated(coupleID string) error {
	_, err := s.DB.Exec(`UPDATE couples SET fenris_validated = TRUE WHERE id = $1`, coupleID)
	return err
}

// SetCoupleSource sets the discovery source on a couple row.
func (s *Store) SetCoupleSource(coupleID, source string) error {
	_, err := s.DB.Exec(`UPDATE couples SET source = $2 WHERE id = $1`, coupleID, source)
	return err
}

