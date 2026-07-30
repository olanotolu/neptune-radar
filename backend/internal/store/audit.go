package store

import (
	"encoding/json"
	"fmt"

	"neptune-social-radar/backend/internal/ontology"
)

// Audit records one row per pipeline-stage decision, not just terminal
// outcomes — this is what makes the "why did the system do X" trail complete.
// monitor groups the trail by watch source ("hashtag:justengaged", ...);
// stepIndex is -1 for live ingestion (no replay step semantics).
func (s *Store) Audit(entityType, entityID, event string, detail any, monitor string, stepIndex int) (ontology.AuditEvent, error) {
	detailJSON, err := json.Marshal(detail)
	if err != nil {
		return ontology.AuditEvent{}, err
	}
	a := ontology.AuditEvent{
		ID: NewID("audit"), EntityType: entityType, EntityID: entityID, Event: event,
		Detail: string(detailJSON), Monitor: monitor, StepIndex: stepIndex,
	}
	_, err = s.DB.Exec(
		`INSERT INTO audit_events (id, entity_type, entity_id, event, detail, monitor, step_index) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		a.ID, a.EntityType, a.EntityID, a.Event, a.Detail, a.Monitor, a.StepIndex,
	)
	return a, err
}

type AuditFilter struct {
	EntityType string
	EntityID   string
	Monitor    string
	Limit      int
}

func (s *Store) ListAudit(f AuditFilter) ([]ontology.AuditEvent, error) {
	q := `SELECT id, entity_type, entity_id, event, COALESCE(detail,''), COALESCE(monitor,''), COALESCE(step_index,0), created_at FROM audit_events WHERE 1=1`
	var args []any
	add := func(clause string, v any) {
		args = append(args, v)
		q += fmt.Sprintf(" AND %s$%d", clause, len(args))
	}
	if f.EntityType != "" {
		add("entity_type = ", f.EntityType)
	}
	if f.EntityID != "" {
		add("entity_id = ", f.EntityID)
	}
	if f.Monitor != "" {
		add("monitor = ", f.Monitor)
	}
	// Newest first — this is a feed. And the read is always bounded: without
	// a default limit the audit UI fetched every event since the beginning
	// of time, oldest first, on every page load.
	q += ` ORDER BY created_at DESC, step_index DESC`
	if f.Limit <= 0 {
		f.Limit = 500
	}
	args = append(args, f.Limit)
	q += fmt.Sprintf(" LIMIT $%d", len(args))
	rows, err := s.DB.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ontology.AuditEvent
	for rows.Next() {
		var a ontology.AuditEvent
		if err := rows.Scan(&a.ID, &a.EntityType, &a.EntityID, &a.Event, &a.Detail, &a.Monitor, &a.StepIndex, &a.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}
