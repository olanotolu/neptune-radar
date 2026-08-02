package store

import (
	"database/sql"
	"time"
)

// StageTiming is one recorded duration for a pipeline stage execution.
type StageTiming struct {
	Stage      string    `json:"stage"`
	DurationMs int64     `json:"duration_ms"`
	EventID    string    `json:"event_id,omitempty"`
	Timestamp  time.Time `json:"timestamp"`
}

// RecordStageTiming inserts a pipeline stage timing row.
func (s *Store) RecordStageTiming(t StageTiming) error {
	id := NewID("ptiming")
	if t.Timestamp.IsZero() {
		t.Timestamp = time.Now().UTC()
	}
	_, err := s.DB.Exec(
		`INSERT INTO pipeline_timings (id, stage, duration_ms, event_id, created_at) VALUES ($1, $2, $3, $4, $5)`,
		id, t.Stage, t.DurationMs, t.EventID, t.Timestamp,
	)
	return err
}

// GetStageTimings returns the most recent timings. If stage is non-empty,
// results are filtered to that stage. limit caps the result count.
func (s *Store) GetStageTimings(stage string, limit int) ([]StageTiming, error) {
	if limit <= 0 {
		limit = 100
	}
	var rows *sql.Rows
	var err error
	if stage != "" {
		rows, err = s.DB.Query(
			`SELECT stage, duration_ms, COALESCE(event_id,''), created_at
			 FROM pipeline_timings WHERE stage = $1
			 ORDER BY created_at DESC LIMIT $2`, stage, limit)
	} else {
		rows, err = s.DB.Query(
			`SELECT stage, duration_ms, COALESCE(event_id,''), created_at
			 FROM pipeline_timings
			 ORDER BY created_at DESC LIMIT $1`, limit)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []StageTiming
	for rows.Next() {
		var t StageTiming
		if err := rows.Scan(&t.Stage, &t.DurationMs, &t.EventID, &t.Timestamp); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// GetPipelineSummary returns the average duration (ms) per stage.
func (s *Store) GetPipelineSummary() (map[string]float64, error) {
	rows, err := s.DB.Query(
		`SELECT stage, AVG(duration_ms)::float8 FROM pipeline_timings GROUP BY stage`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make(map[string]float64)
	for rows.Next() {
		var stage string
		var avg float64
		if err := rows.Scan(&stage, &avg); err != nil {
			return nil, err
		}
		out[stage] = avg
	}
	return out, rows.Err()
}

// PipelineThroughput returns events processed per minute over the last hour.
func (s *Store) PipelineThroughput() (float64, error) {
	var count int
	err := s.DB.QueryRow(
		`SELECT COUNT(*) FROM pipeline_timings WHERE created_at > now() - interval '1 hour'`,
	).Scan(&count)
	if err != nil {
		return 0, err
	}
	return float64(count) / 60.0, nil
}
