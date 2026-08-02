package store

import (
	"fmt"
	"time"

	"neptune-social-radar/backend/internal/ontology"
)

// CoupleSearchHit is a couple row plus the confidence of its current
// relationship (0 when no relationship exists yet).
type CoupleSearchHit struct {
	ontology.Couple
	Confidence float64 `json:"confidence"`
}

// SearchResult holds grouped matches across the searchable entities.
type SearchResult struct {
	Couples []CoupleSearchHit   `json:"couples,omitempty"`
	Leads   []ontology.CRMLead  `json:"leads,omitempty"`
	Cases   []ontology.NeptuneCase `json:"cases,omitempty"`
}

// SearchParams carries the universal-search query filters.
type SearchParams struct {
	Query        string  // free-text substring
	Type         string  // "couples" | "leads" | "cases" | "all"
	State        string  // USPS state code (couples only)
	MinConfidence float64 // couples only; 0 = no floor
	Limit        int
}

// Search runs a universal ILIKE search across couples, leads, and cases.
// State and min_confidence only apply to couples (the only entity with geo +
// confidence); leads/cases have neither. ponytail: ceiling — ILIKE substring
// scan, no trigram/FTS index yet; fine at current row counts, upgrade to
// pg_trgm or tsvector when couples exceeds ~50k rows.
func (s *Store) Search(p SearchParams) (SearchResult, error) {
	if p.Limit <= 0 {
		p.Limit = 50
	}
	pattern := "%" + p.Query + "%"
	var res SearchResult
	var err error
	switch p.Type {
	case "leads":
		res.Leads, err = s.searchLeads(pattern, p.Limit)
	case "cases":
		res.Cases, err = s.searchCases(pattern, p.Limit)
	case "all", "":
		if res.Couples, err = s.searchCouples(pattern, p); err != nil {
			return res, err
		}
		if res.Leads, err = s.searchLeads(pattern, p.Limit); err != nil {
			return res, err
		}
		if res.Cases, err = s.searchCases(pattern, p.Limit); err != nil {
			return res, err
		}
	default: // "couples"
		res.Couples, err = s.searchCouples(pattern, p)
	}
	return res, err
}

// searchCouples matches on either partner's display name or handle, or the
// couple's inferred city. State filters inferred_region; min_confidence filters
// on the current (effective_to IS NULL) relationship's confidence.
func (s *Store) searchCouples(pattern string, p SearchParams) ([]CoupleSearchHit, error) {
	q := `SELECT c.id, c.person_a_id, c.person_b_id, c.created_at,
			COALESCE(c.inferred_city,''), COALESCE(c.inferred_region,''),
			COALESCE(c.location_source,''), COALESCE(r.confidence,0)
		FROM couples c
		LEFT JOIN LATERAL (
			SELECT confidence FROM relationships
			WHERE couple_id = c.id AND effective_to IS NULL
			ORDER BY effective_from DESC LIMIT 1
		) r ON TRUE
		WHERE (
			EXISTS (SELECT 1 FROM persons p WHERE p.id = c.person_a_id AND p.display_name ILIKE $1) OR
			EXISTS (SELECT 1 FROM persons p WHERE p.id = c.person_b_id AND p.display_name ILIKE $1) OR
			EXISTS (SELECT 1 FROM social_accounts a WHERE a.person_id = c.person_a_id AND a.handle ILIKE $1) OR
			EXISTS (SELECT 1 FROM social_accounts a WHERE a.person_id = c.person_b_id AND a.handle ILIKE $1) OR
			c.inferred_city ILIKE $1
		)`
	args := []any{pattern}
	if p.State != "" {
		args = append(args, p.State)
		q += fmt.Sprintf(" AND c.inferred_region = $%d", len(args))
	}
	if p.MinConfidence > 0 {
		args = append(args, p.MinConfidence)
		q += fmt.Sprintf(" AND COALESCE(r.confidence,0) >= $%d", len(args))
	}
	args = append(args, p.Limit)
	q += fmt.Sprintf(" ORDER BY c.created_at DESC LIMIT $%d", len(args))
	rows, err := s.DB.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []CoupleSearchHit
	for rows.Next() {
		var h CoupleSearchHit
		if err := rows.Scan(&h.ID, &h.PersonAID, &h.PersonBID, &h.CreatedAt,
			&h.InferredCity, &h.InferredRegion, &h.LocationSource, &h.Confidence); err != nil {
			return nil, err
		}
		out = append(out, h)
	}
	return out, rows.Err()
}

// searchLeads matches on the lead person's display name or social handle.
func (s *Store) searchLeads(pattern string, limit int) ([]ontology.CRMLead, error) {
	rows, err := s.DB.Query(
		`SELECT l.id, l.person_id, COALESCE(l.hypothesis_id,''), l.lead_type, l.status, l.created_at
		 FROM crm_leads l
		 WHERE EXISTS (SELECT 1 FROM persons p WHERE p.id = l.person_id AND p.display_name ILIKE $1)
		    OR EXISTS (SELECT 1 FROM social_accounts a WHERE a.person_id = l.person_id AND a.handle ILIKE $1)
		 ORDER BY l.created_at DESC LIMIT $2`,
		pattern, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ontology.CRMLead
	for rows.Next() {
		var l ontology.CRMLead
		if err := rows.Scan(&l.ID, &l.PersonID, &l.HypothesisID, &l.LeadType, &l.Status, &l.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

// searchCases matches on the case's couple_id or status.
func (s *Store) searchCases(pattern string, limit int) ([]ontology.NeptuneCase, error) {
	rows, err := s.DB.Query(
		`SELECT id, COALESCE(couple_id,''), COALESCE(lead_id,''), case_type, status, created_at, updated_at
		 FROM neptune_cases
		 WHERE couple_id ILIKE $1 OR status ILIKE $1
		 ORDER BY updated_at DESC LIMIT $2`,
		pattern, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ontology.NeptuneCase
	for rows.Next() {
		var c ontology.NeptuneCase
		if err := rows.Scan(&c.ID, &c.CoupleID, &c.LeadID, &c.CaseType, &c.Status, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// coupleExportRow is a flattened couple for CSV export.
type coupleExportRow struct {
	ID             string
	PartnerAHandle string
	PartnerBHandle string
	Stage          string
	Confidence     float64
	City           string
	State          string
	CreatedAt      time.Time
}

// ExportCouples returns couples flattened with partner handles + current
// relationship stage/confidence, optionally filtered by state and stage.
func (s *Store) ExportCouples(state, stage string) ([]coupleExportRow, error) {
	q := `SELECT c.id,
			COALESCE((SELECT a.handle FROM social_accounts a WHERE a.person_id = c.person_a_id LIMIT 1),''),
			COALESCE((SELECT a.handle FROM social_accounts a WHERE a.person_id = c.person_b_id LIMIT 1),''),
			COALESCE(r.stage,''), COALESCE(r.confidence,0),
			COALESCE(c.inferred_city,''), COALESCE(c.inferred_region,''), c.created_at
		FROM couples c
		LEFT JOIN LATERAL (
			SELECT stage, confidence FROM relationships
			WHERE couple_id = c.id AND effective_to IS NULL
			ORDER BY effective_from DESC LIMIT 1
		) r ON TRUE`
	args := []any{}
	if state != "" {
		args = append(args, state)
		q += fmt.Sprintf(" WHERE c.inferred_region = $%d", len(args))
		if stage != "" {
			args = append(args, stage)
			q += fmt.Sprintf(" AND r.stage = $%d", len(args))
		}
	} else if stage != "" {
		args = append(args, stage)
		q += fmt.Sprintf(" WHERE r.stage = $%d", len(args))
	}
	q += " ORDER BY c.created_at ASC"
	rows, err := s.DB.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []coupleExportRow
	for rows.Next() {
		var r coupleExportRow
		if err := rows.Scan(&r.ID, &r.PartnerAHandle, &r.PartnerBHandle, &r.Stage, &r.Confidence, &r.City, &r.State, &r.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// leadExportRow is a flattened lead for CSV export.
type leadExportRow struct {
	ID        string
	CoupleID  string
	Name      string
	Handle    string
	Stage     string
	CreatedAt time.Time
}

// ExportLeads returns leads flattened with person name, handle, and the
// couple_id + proposed_stage from the lead's hypothesis (if any).
func (s *Store) ExportLeads() ([]leadExportRow, error) {
	rows, err := s.DB.Query(
		`SELECT l.id, COALESCE(h.couple_id,''),
			COALESCE(p.display_name,''), COALESCE(a.handle,''),
			COALESCE(l.status,''), l.created_at
		 FROM crm_leads l
		 LEFT JOIN persons p ON p.id = l.person_id
		 LEFT JOIN social_accounts a ON a.person_id = l.person_id
		 LEFT JOIN life_event_hypotheses h ON h.id = l.hypothesis_id
		 ORDER BY l.created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []leadExportRow
	for rows.Next() {
		var r leadExportRow
		if err := rows.Scan(&r.ID, &r.CoupleID, &r.Name, &r.Handle, &r.Stage, &r.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// auditExportRow is a flattened audit event for CSV export.
type auditExportRow struct {
	ID         string
	EntityType string
	EntityID   string
	Event      string
	Monitor    string
	CreatedAt  time.Time
}

// ExportAudit returns audit events newest-first, bounded by limit.
func (s *Store) ExportAudit(limit int) ([]auditExportRow, error) {
	if limit <= 0 {
		limit = 1000
	}
	rows, err := s.DB.Query(
		`SELECT id, entity_type, entity_id, event, COALESCE(monitor,''), created_at
		 FROM audit_events ORDER BY created_at DESC LIMIT $1`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []auditExportRow
	for rows.Next() {
		var r auditExportRow
		if err := rows.Scan(&r.ID, &r.EntityType, &r.EntityID, &r.Event, &r.Monitor, &r.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
