package store

import (
	"database/sql"
	"time"

	"neptune-social-radar/backend/internal/ontology"
)

func (s *Store) CreateLead(l ontology.CRMLead) (ontology.CRMLead, error) {
	if l.ID == "" {
		l.ID = NewID("lead")
	}
	if l.Status == "" {
		l.Status = "new"
	}
	var hypID any
	if l.HypothesisID != "" {
		hypID = l.HypothesisID
	}
	_, err := s.DB.Exec(`INSERT INTO crm_leads (id, person_id, hypothesis_id, lead_type, status) VALUES ($1, $2, $3, $4, $5)`,
		l.ID, l.PersonID, hypID, l.LeadType, l.Status)
	if err != nil {
		return l, err
	}
	return s.GetLead(l.ID)
}

func (s *Store) GetLead(id string) (ontology.CRMLead, error) {
	var l ontology.CRMLead
	var hypID sql.NullString
	err := s.DB.QueryRow(
		`SELECT id, person_id, COALESCE(hypothesis_id,''), lead_type, status, created_at FROM crm_leads WHERE id = $1`, id,
	).Scan(&l.ID, &l.PersonID, &hypID, &l.LeadType, &l.Status, &l.CreatedAt)
	if err != nil {
		return l, err
	}
	l.HypothesisID = hypID.String
	return l, nil
}

func (s *Store) ListLeads(status string) ([]ontology.CRMLead, error) {
	q := `SELECT id, person_id, COALESCE(hypothesis_id,''), lead_type, status, created_at FROM crm_leads`
	args := []any{}
	if status != "" {
		q += ` WHERE status = $1`
		args = append(args, status)
	}
	q += ` ORDER BY created_at DESC, id DESC`
	rows, err := s.DB.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ontology.CRMLead
	for rows.Next() {
		var l ontology.CRMLead
		var hypID sql.NullString
		if err := rows.Scan(&l.ID, &l.PersonID, &hypID, &l.LeadType, &l.Status, &l.CreatedAt); err != nil {
			return nil, err
		}
		l.HypothesisID = hypID.String
		out = append(out, l)
	}
	return out, rows.Err()
}

func (s *Store) CreateCase(c ontology.NeptuneCase) (ontology.NeptuneCase, error) {
	if c.ID == "" {
		c.ID = NewID("case")
	}
	if c.Status == "" {
		c.Status = "intake"
	}
	var coupleID, leadID any
	if c.CoupleID != "" {
		coupleID = c.CoupleID
	}
	if c.LeadID != "" {
		leadID = c.LeadID
	}
	_, err := s.DB.Exec(`INSERT INTO neptune_cases (id, couple_id, lead_id, case_type, status) VALUES ($1, $2, $3, $4, $5)`,
		c.ID, coupleID, leadID, c.CaseType, c.Status)
	if err != nil {
		return c, err
	}
	return s.GetCase(c.ID)
}

func (s *Store) GetCase(id string) (ontology.NeptuneCase, error) {
	var c ontology.NeptuneCase
	var coupleID, leadID sql.NullString
	err := s.DB.QueryRow(
		`SELECT id, COALESCE(couple_id,''), COALESCE(lead_id,''), case_type, status, created_at, updated_at FROM neptune_cases WHERE id = $1`, id,
	).Scan(&c.ID, &coupleID, &leadID, &c.CaseType, &c.Status, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return c, err
	}
	c.CoupleID, c.LeadID = coupleID.String, leadID.String
	return c, nil
}

func (s *Store) GetActiveCaseForCouple(coupleID string) (ontology.NeptuneCase, error) {
	var id string
	err := s.DB.QueryRow(
		`SELECT id FROM neptune_cases WHERE couple_id = $1 AND status != 'closed' ORDER BY created_at DESC, id DESC LIMIT 1`, coupleID,
	).Scan(&id)
	if err != nil {
		return ontology.NeptuneCase{}, err
	}
	return s.GetCase(id)
}

func (s *Store) UpdateCaseStatus(id, status string) error {
	_, err := s.DB.Exec(`UPDATE neptune_cases SET status = $1, updated_at = $2 WHERE id = $3`, status, time.Now().UTC(), id)
	return err
}

func (s *Store) ListCases(status string) ([]ontology.NeptuneCase, error) {
	q := `SELECT id, COALESCE(couple_id,''), COALESCE(lead_id,''), case_type, status, created_at, updated_at FROM neptune_cases`
	args := []any{}
	if status != "" {
		q += ` WHERE status = $1`
		args = append(args, status)
	}
	q += ` ORDER BY updated_at DESC`
	rows, err := s.DB.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ontology.NeptuneCase
	for rows.Next() {
		var c ontology.NeptuneCase
		var coupleID, leadID sql.NullString
		if err := rows.Scan(&c.ID, &coupleID, &leadID, &c.CaseType, &c.Status, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		c.CoupleID, c.LeadID = coupleID.String, leadID.String
		out = append(out, c)
	}
	return out, rows.Err()
}
