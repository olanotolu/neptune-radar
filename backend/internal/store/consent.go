package store

import (
	"encoding/json"
	"time"

	"neptune-social-radar/backend/internal/ontology"
)

// CreateConsentForCouple creates an active consent policy for both persons in
// a couple with the given allowed actions. Returns the two policies. Used by
// the celebrate-page consent capture flow (GDPR/CCPA layer).
func (s *Store) CreateConsentForCouple(coupleID string, actions []string) ([]ontology.ConsentPolicy, error) {
	c, err := s.GetCouple(coupleID)
	if err != nil {
		return nil, err
	}
	out := make([]ontology.ConsentPolicy, 0, 2)
	for _, pid := range []string{c.PersonAID, c.PersonBID} {
		p, err := s.CreateConsentPolicy(ontology.ConsentPolicy{
			PersonID:       pid,
			Scope:          ontology.ScopeSharedCouple,
			AllowedActions: actions,
		})
		if err != nil {
			return out, err
		}
		out = append(out, p)
	}
	return out, nil
}

// ConsentStatus is the per-couple consent snapshot returned to the celebrate page.
type ConsentStatus struct {
	Granted        bool       `json:"granted"`
	Revoked        bool       `json:"revoked"`
	AllowedActions []string   `json:"allowed_actions,omitempty"`
	GrantedAt      *time.Time `json:"granted_at,omitempty"`
}

// GetConsentStatus returns the current consent state for a couple. Consent is
// considered granted when at least one partner has an active (non-revoked)
// policy; revoked when the most recent policy for either partner is revoked.
// ponytail: ceiling — checks person_a only; for a symmetric couple flow both
// partners consent together, so one active policy is sufficient signal. Upgrade
// path: AND both partners if independent per-person consent is ever needed.
func (s *Store) GetConsentStatus(coupleID string) (ConsentStatus, error) {
	c, err := s.GetCouple(coupleID)
	if err != nil {
		return ConsentStatus{}, err
	}
	// Most recent policy for either partner (active or revoked).
	var pol ontology.ConsentPolicy
	var actionsJSON string
	err = s.DB.QueryRow(
		`SELECT id, person_id, scope, allowed_actions, revoked_at, granted_at FROM consent_policies
		 WHERE person_id IN ($1, $2) ORDER BY granted_at DESC LIMIT 1`,
		c.PersonAID, c.PersonBID,
	).Scan(&pol.ID, &pol.PersonID, &pol.Scope, &actionsJSON, &pol.RevokedAt, &pol.GrantedAt)
	if err != nil {
		// ponytail: sql.ErrNoRows → no consent on file. Treat as not-granted.
		return ConsentStatus{Granted: false, Revoked: false}, nil
	}
	_ = json.Unmarshal([]byte(actionsJSON), &pol.AllowedActions)
	st := ConsentStatus{AllowedActions: pol.AllowedActions, GrantedAt: &pol.GrantedAt}
	if pol.RevokedAt != nil {
		st.Revoked = true
		st.Granted = false
	} else {
		st.Granted = true
	}
	return st, nil
}

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

	// 3. Suppress couples involving this person so no further postcards or
	// follow-ups are generated. Opt-out must take immediate effect.
	if _, err := tx.Exec(
		`UPDATE couples SET suppressed_at = now(), suppressed_reason = 'consent_revoked'
		 WHERE (person_a_id = $1 OR person_b_id = $1) AND suppressed_at IS NULL`, personID,
	); err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return int(cancelled), nil
}
