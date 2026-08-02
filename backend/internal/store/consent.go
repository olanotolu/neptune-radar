package store

import (
	"encoding/json"

	"neptune-social-radar/backend/internal/ontology"
)

func (s *Store) CreateConsentPolicy(c ontology.ConsentPolicy) (ontology.ConsentPolicy, error) {
	if c.ID == "" {
		c.ID = NewID("consent")
	}
	actionsJSON, err := json.Marshal(c.AllowedActions)
	if err != nil {
		return c, err
	}
	_, err = s.DB.Exec(`INSERT INTO consent_policies (id, person_id, scope, allowed_actions) VALUES ($1, $2, $3, $4)`,
		c.ID, c.PersonID, c.Scope, string(actionsJSON))
	return c, err
}

// ActiveConsentForPerson returns the person's current (non-revoked) consent
// policy. sql.ErrNoRows means no consent on file — policy code must treat
// that as "no permitted actions".
func (s *Store) ActiveConsentForPerson(personID string) (ontology.ConsentPolicy, error) {
	var c ontology.ConsentPolicy
	var actionsJSON string
	err := s.DB.QueryRow(
		`SELECT id, person_id, scope, allowed_actions, revoked_at FROM consent_policies
		 WHERE person_id = $1 AND revoked_at IS NULL ORDER BY granted_at DESC LIMIT 1`, personID,
	).Scan(&c.ID, &c.PersonID, &c.Scope, &actionsJSON, &c.RevokedAt)
	if err != nil {
		return c, err
	}
	_ = json.Unmarshal([]byte(actionsJSON), &c.AllowedActions)
	return c, nil
}

// RevokeConsent revokes all active consent for a person AND cascades: any
// pending recommended_actions for couples involving that person are cancelled
// (marked "consent_revoked") so they can never be approved after the person
// withdrew consent. This is the consent-revocation cascade — without it,
// revoking consent only blocks NEW actions while already-queued ones remain
// approvable, which defeats the purpose of revocation.
func (s *Store) RevokeConsent(personID string) (int, error) {
	tx, err := s.DB.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	// 1. Revoke the consent policy rows.
	if _, err := tx.Exec(`UPDATE consent_policies SET revoked_at = now() WHERE person_id = $1 AND revoked_at IS NULL`, personID); err != nil {
		return 0, err
	}

	// 2. Cancel pending recommended_actions for any couple involving this person.
	// A person can be person_a or person_b in a couple; the hypothesis links
	// the action to the couple, so we join through both sides.
	res, err := tx.Exec(
		`UPDATE recommended_actions
		   SET status = 'consent_revoked', decided_at = now(), decided_by = 'system:consent_revocation'
		 WHERE status = 'pending'
		   AND hypothesis_id IN (
		     SELECT h.id FROM life_event_hypotheses h
		     JOIN couples c ON h.couple_id = c.id
		     WHERE c.person_a_id = $1 OR c.person_b_id = $1
		   )`,
		personID,
	)
	if err != nil {
		return 0, err
	}
	cancelled, _ := res.RowsAffected()

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return int(cancelled), nil
}
