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

func (s *Store) RevokeConsent(personID string) error {
	_, err := s.DB.Exec(`UPDATE consent_policies SET revoked_at = now() WHERE person_id = $1 AND revoked_at IS NULL`, personID)
	return err
}
