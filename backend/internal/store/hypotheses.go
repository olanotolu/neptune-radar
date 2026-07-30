package store

import (
	"database/sql"
	"time"

	"neptune-social-radar/backend/internal/ontology"
)

func (s *Store) CreateHypothesis(h ontology.LifeEventHypothesis) (ontology.LifeEventHypothesis, error) {
	if h.ID == "" {
		h.ID = NewID("hyp")
	}
	if h.Status == "" {
		h.Status = ontology.HypothesisUnconfirmed
	}
	if h.VisibilityScope == "" {
		h.VisibilityScope = ontology.ScopeUnconfirmedInfer
	}
	var coupleID, personID any
	var expiresAt, engagementConf, partnerConf any
	if h.CoupleID != "" {
		coupleID = h.CoupleID
	}
	if h.PersonID != "" {
		personID = h.PersonID
	}
	if h.ExpiresAt != nil {
		expiresAt = h.ExpiresAt.UTC()
	}
	if h.EngagementConfidence != nil {
		engagementConf = *h.EngagementConfidence
	}
	if h.PartnerConfidence != nil {
		partnerConf = *h.PartnerConfidence
	}
	_, err := s.DB.Exec(
		`INSERT INTO life_event_hypotheses
		 (id, couple_id, person_id, event_type, proposed_stage, confidence, engagement_confidence, partner_confidence, model_or_rule, status, visibility_scope, consent_scope, expires_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
		h.ID, coupleID, personID, h.EventType, h.ProposedStage, h.Confidence, engagementConf, partnerConf, h.ModelOrRule, h.Status, h.VisibilityScope, h.ConsentScope, expiresAt,
	)
	if err != nil {
		return h, err
	}
	return s.GetHypothesis(h.ID)
}

func (s *Store) GetHypothesis(id string) (ontology.LifeEventHypothesis, error) {
	var h ontology.LifeEventHypothesis
	var coupleID, personID, proposedStage sql.NullString
	var expiresAt sql.NullTime
	var engagementConf, partnerConf sql.NullFloat64
	err := s.DB.QueryRow(
		`SELECT id, COALESCE(couple_id,''), COALESCE(person_id,''), event_type, COALESCE(proposed_stage,''), confidence, engagement_confidence, partner_confidence, model_or_rule, status, visibility_scope, consent_scope, expires_at, created_at, updated_at
		 FROM life_event_hypotheses WHERE id = $1`, id,
	).Scan(&h.ID, &coupleID, &personID, &h.EventType, &proposedStage, &h.Confidence, &engagementConf, &partnerConf, &h.ModelOrRule, &h.Status, &h.VisibilityScope, &h.ConsentScope, &expiresAt, &h.CreatedAt, &h.UpdatedAt)
	if err != nil {
		return h, err
	}
	h.CoupleID, h.PersonID = coupleID.String, personID.String
	h.ProposedStage = ontology.RelationshipStage(proposedStage.String)
	if expiresAt.Valid {
		t := expiresAt.Time
		h.ExpiresAt = &t
	}
	if engagementConf.Valid {
		h.EngagementConfidence = &engagementConf.Float64
	}
	if partnerConf.Valid {
		h.PartnerConfidence = &partnerConf.Float64
	}
	return h, nil
}

func (s *Store) UpdateHypothesisStatus(id string, status ontology.HypothesisStatus) error {
	_, err := s.DB.Exec(`UPDATE life_event_hypotheses SET status = $1, updated_at = $2 WHERE id = $3`,
		status, time.Now().UTC(), id)
	return err
}

func (s *Store) UpdateHypothesisModelRule(id, modelOrRule string) error {
	_, err := s.DB.Exec(`UPDATE life_event_hypotheses SET model_or_rule = $1, updated_at = $2 WHERE id = $3`,
		modelOrRule, time.Now().UTC(), id)
	return err
}

func (s *Store) UpdateHypothesisConfidence(id string, confidence float64) error {
	_, err := s.DB.Exec(`UPDATE life_event_hypotheses SET confidence = $1, updated_at = $2 WHERE id = $3`,
		confidence, time.Now().UTC(), id)
	return err
}

// UpdateHypothesisSubScores stores the two separately-reported event-first
// scores. Only meaningful for engagement hypotheses.
func (s *Store) UpdateHypothesisSubScores(id string, engagementConfidence, partnerConfidence float64) error {
	_, err := s.DB.Exec(`UPDATE life_event_hypotheses SET engagement_confidence = $1, partner_confidence = $2, updated_at = $3 WHERE id = $4`,
		engagementConfidence, partnerConfidence, time.Now().UTC(), id)
	return err
}

// LatestHypothesisForCouple returns the most recently created, still-relevant
// (not rejected/expired) hypothesis for a couple, if any.
func (s *Store) LatestHypothesisForCouple(coupleID string) (ontology.LifeEventHypothesis, error) {
	var id string
	err := s.DB.QueryRow(
		`SELECT id FROM life_event_hypotheses WHERE couple_id = $1 AND status NOT IN ('rejected','expired') ORDER BY created_at DESC, id DESC LIMIT 1`, coupleID,
	).Scan(&id)
	if err != nil {
		return ontology.LifeEventHypothesis{}, err
	}
	return s.GetHypothesis(id)
}

func (s *Store) ExpireStaleHypotheses(before time.Time) (int, error) {
	res, err := s.DB.Exec(
		`UPDATE life_event_hypotheses SET status = 'expired', updated_at = $1
		 WHERE status = 'unconfirmed' AND expires_at IS NOT NULL AND expires_at < $2`,
		time.Now().UTC(), before.UTC(),
	)
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return int(n), nil
}
