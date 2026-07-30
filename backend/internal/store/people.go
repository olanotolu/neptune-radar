package store

import (

	"neptune-social-radar/backend/internal/ontology"
)

func (s *Store) CreatePerson(p ontology.Person) (ontology.Person, error) {
	if p.ID == "" {
		p.ID = NewID("person")
	}
	_, err := s.DB.Exec(
		`INSERT INTO persons (id, display_name, email, crm_source) VALUES ($1, $2, $3, $4)`,
		p.ID, p.DisplayName, p.Email, p.CRMSource,
	)
	if err != nil {
		return ontology.Person{}, err
	}
	return s.GetPerson(p.ID)
}

func (s *Store) GetPerson(id string) (ontology.Person, error) {
	var p ontology.Person
	err := s.DB.QueryRow(
		`SELECT id, display_name, COALESCE(email,''), COALESCE(crm_source,''), created_at FROM persons WHERE id = $1`, id,
	).Scan(&p.ID, &p.DisplayName, &p.Email, &p.CRMSource, &p.CreatedAt)
	if err != nil {
		return p, err
	}
	return p, nil
}

