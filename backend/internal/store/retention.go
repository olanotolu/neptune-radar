package store

import (
	"context"
	"fmt"
	"time"
)

// RetentionClass maps an entity type to its max age before purge.
type RetentionClass struct {
	EntityType  string `json:"entity_type"`
	MaxAgeDays  int    `json:"max_age_days"`
	Description string `json:"description,omitempty"`
}

// ListRetentionClasses returns all configured retention classes.
func (s *Store) ListRetentionClasses() ([]RetentionClass, error) {
	rows, err := s.DB.Query(
		`SELECT entity_type, max_age_days, COALESCE(description,'') FROM retention_classes ORDER BY entity_type`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []RetentionClass
	for rows.Next() {
		var r RetentionClass
		if err := rows.Scan(&r.EntityType, &r.MaxAgeDays, &r.Description); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// SetRetentionClass upserts a retention class. Admin-only via the API.
func (s *Store) SetRetentionClass(entityType string, maxAgeDays int, description string) error {
	_, err := s.DB.Exec(
		`INSERT INTO retention_classes (entity_type, max_age_days, description, updated_at)
		 VALUES ($1, $2, $3, now())
		 ON CONFLICT (entity_type) DO UPDATE SET max_age_days = $2, description = $3, updated_at = now()`,
		entityType, maxAgeDays, description)
	return err
}

// PurgePreview returns what WOULD be purged for each entity type based on
// the configured retention classes, without actually deleting anything.
type PurgePreview struct {
	EntityType string `json:"entity_type"`
	MaxAgeDays int    `json:"max_age_days"`
	RowCount   int    `json:"row_count"`
}

func (s *Store) PurgePreview(ctx context.Context) ([]PurgePreview, error) {
	classes, err := s.ListRetentionClasses()
	if err != nil {
		return nil, err
	}
	var out []PurgePreview
	for _, c := range classes {
		cutoff := time.Now().UTC().AddDate(0, 0, -c.MaxAgeDays)
		var count int
		switch c.EntityType {
		case "social_observations":
			err = s.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM social_observations WHERE created_at < $1`, cutoff).Scan(&count)
		case "pipeline_timings":
			err = s.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM pipeline_timings WHERE created_at < $1`, cutoff).Scan(&count)
		case "pipeline_runs":
			err = s.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM pipeline_runs WHERE created_at < $1`, cutoff).Scan(&count)
		case "dlq_items":
			err = s.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM dlq_items WHERE created_at < $1`, cutoff).Scan(&count)
		default:
			continue // unknown entity type — skip
		}
		if err != nil {
			return nil, fmt.Errorf("purge preview %s: %w", c.EntityType, err)
		}
		out = append(out, PurgePreview{EntityType: c.EntityType, MaxAgeDays: c.MaxAgeDays, RowCount: count})
	}
	return out, nil
}

// PurgeExpired deletes rows older than the retention class max age for each
// configured entity type. Returns a summary of what was purged. Called by
// the janitor on a schedule.
type PurgeResult struct {
	EntityType string `json:"entity_type"`
	Deleted    int    `json:"deleted"`
}

func (s *Store) PurgeExpired(ctx context.Context) ([]PurgeResult, error) {
	classes, err := s.ListRetentionClasses()
	if err != nil {
		return nil, err
	}
	var out []PurgeResult
	for _, c := range classes {
		cutoff := time.Now().UTC().AddDate(0, 0, -c.MaxAgeDays)
		var res struct{ n int64 }
		switch c.EntityType {
		case "social_observations":
			r, err := s.DB.ExecContext(ctx, `DELETE FROM social_observations WHERE created_at < $1`, cutoff)
			if err != nil {
				return out, fmt.Errorf("purge %s: %w", c.EntityType, err)
			}
			res.n, _ = r.RowsAffected()
		case "pipeline_timings":
			r, err := s.DB.ExecContext(ctx, `DELETE FROM pipeline_timings WHERE created_at < $1`, cutoff)
			if err != nil {
				return out, fmt.Errorf("purge %s: %w", c.EntityType, err)
			}
			res.n, _ = r.RowsAffected()
		case "pipeline_runs":
			r, err := s.DB.ExecContext(ctx, `DELETE FROM pipeline_runs WHERE created_at < $1`, cutoff)
			if err != nil {
				return out, fmt.Errorf("purge %s: %w", c.EntityType, err)
			}
			res.n, _ = r.RowsAffected()
		case "dlq_items":
			r, err := s.DB.ExecContext(ctx, `DELETE FROM dlq_items WHERE created_at < $1`, cutoff)
			if err != nil {
				return out, fmt.Errorf("purge %s: %w", c.EntityType, err)
			}
			res.n, _ = r.RowsAffected()
		default:
			continue
		}
		out = append(out, PurgeResult{EntityType: c.EntityType, Deleted: int(res.n)})
	}
	return out, nil
}
