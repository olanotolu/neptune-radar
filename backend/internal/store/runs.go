package store

import (
	"database/sql"
	"fmt"
	"time"

	"neptune-social-radar/backend/internal/ontology"
)

// RecordPipelineRun inserts one summary row for a completed ProcessEvent
// execution. The per-stage detail (audit_events, pipeline_timings) is
// written by the orchestrator as it runs; this row is the cross-cutting
// index that ties them together.
func (s *Store) RecordPipelineRun(r ontology.PipelineRun) error {
	if r.ID == "" {
		r.ID = NewID("run")
	}
	if r.StartedAt.IsZero() {
		r.StartedAt = time.Now().UTC()
	}
	_, err := s.DB.Exec(
		`INSERT INTO pipeline_runs
		   (id, observation_id, agent_name, model, prompt_tokens, completion_tokens,
		    cost_usd, confidence, stop_reason, hypothesis_id, action_id, couple_id,
		    monitor, started_at, ended_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
		r.ID, r.ObservationID, r.AgentName, r.Model, r.PromptTokens, r.CompletionTokens,
		r.CostUSD, r.Confidence, r.StopReason, r.HypothesisID, r.ActionID, r.CoupleID,
		r.Monitor, r.StartedAt, r.EndedAt,
	)
	return err
}

type RunFilter struct {
	CoupleID   string
	Monitor    string
	StopReason string
	Limit      int
}

func (s *Store) ListPipelineRuns(f RunFilter) ([]ontology.PipelineRun, error) {
	if f.Limit <= 0 {
		f.Limit = 100
	}
	q := `SELECT id, observation_id, agent_name, COALESCE(model,''),
		       prompt_tokens, completion_tokens, cost_usd::float8,
		       COALESCE(confidence,0), stop_reason, COALESCE(hypothesis_id,''),
		       COALESCE(action_id,''), COALESCE(couple_id,''), COALESCE(monitor,''),
		       started_at, COALESCE(ended_at, started_at), created_at
		FROM pipeline_runs WHERE 1=1`
	var args []any
	add := func(clause string, v any) {
		args = append(args, v)
		q += ` AND ` + clause + fmt.Sprintf("$%d", len(args))
	}
	if f.CoupleID != "" {
		add("couple_id = ", f.CoupleID)
	}
	if f.Monitor != "" {
		add("monitor = ", f.Monitor)
	}
	if f.StopReason != "" {
		add("stop_reason = ", f.StopReason)
	}
	args = append(args, f.Limit)
	q += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d", len(args))
	rows, err := s.DB.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ontology.PipelineRun
	for rows.Next() {
		var r ontology.PipelineRun
		var conf float64
		var ended time.Time
		if err := rows.Scan(&r.ID, &r.ObservationID, &r.AgentName, &r.Model,
			&r.PromptTokens, &r.CompletionTokens, &r.CostUSD, &conf, &r.StopReason,
			&r.HypothesisID, &r.ActionID, &r.CoupleID, &r.Monitor,
			&r.StartedAt, &ended, &r.CreatedAt); err != nil {
			return nil, err
		}
		if conf > 0 {
			r.Confidence = &conf
		}
		r.EndedAt = &ended
		out = append(out, r)
	}
	return out, rows.Err()
}

// GetPipelineRun returns one run by id, with the cross-cutting summary.
func (s *Store) GetPipelineRun(id string) (ontology.PipelineRun, error) {
	var r ontology.PipelineRun
	var conf sql.NullFloat64
	var ended sql.NullTime
	err := s.DB.QueryRow(
		`SELECT id, observation_id, agent_name, COALESCE(model,''),
		   prompt_tokens, completion_tokens, cost_usd::float8,
		   confidence, stop_reason, COALESCE(hypothesis_id,''),
		   COALESCE(action_id,''), COALESCE(couple_id,''), COALESCE(monitor,''),
		   started_at, ended_at, created_at
		 FROM pipeline_runs WHERE id = $1`, id,
	).Scan(&r.ID, &r.ObservationID, &r.AgentName, &r.Model,
		&r.PromptTokens, &r.CompletionTokens, &r.CostUSD, &conf, &r.StopReason,
		&r.HypothesisID, &r.ActionID, &r.CoupleID, &r.Monitor,
		&r.StartedAt, &ended, &r.CreatedAt)
	if conf.Valid {
		v := conf.Float64
		r.Confidence = &v
	}
	if ended.Valid {
		r.EndedAt = &ended.Time
	}
	return r, err
}

// PipelineRunDetail is the run summary plus the per-stage audit events and
// timings that belong to it, so the viewer can render the full trace in one
// round-trip.
type PipelineRunDetail struct {
	ontology.PipelineRun
	Timings []StageTiming         `json:"timings"`
	Events  []ontology.AuditEvent `json:"events"`
}

func (s *Store) GetPipelineRunDetail(id string) (PipelineRunDetail, error) {
	r, err := s.GetPipelineRun(id)
	if err != nil {
		return PipelineRunDetail{}, err
	}
	d := PipelineRunDetail{PipelineRun: r}
	// pipeline_timings.event_id is the observation_id; the run id is the
	// observation_id (one run per ProcessEvent call).
	timings, err := s.DB.Query(
		`SELECT stage, duration_ms, COALESCE(event_id,''), created_at
		 FROM pipeline_timings WHERE event_id = $1 ORDER BY created_at`, r.ObservationID)
	if err != nil {
		return d, err
	}
	defer timings.Close()
	for timings.Next() {
		var t StageTiming
		if err := timings.Scan(&t.Stage, &t.DurationMs, &t.EventID, &t.Timestamp); err != nil {
			return d, err
		}
		d.Timings = append(d.Timings, t)
	}
	// Audit events for this run: observation-level events key on entity_id =
	// observation_id; hypothesis/action events key on their own ids. We pull
	// the observation-level trail here (the spine of the run); the viewer
	// can fetch hypothesis/action detail separately if needed.
	events, err := s.ListAudit(AuditFilter{EntityID: r.ObservationID, Limit: 200})
	if err != nil {
		return d, err
	}
	d.Events = events
	return d, nil
}
