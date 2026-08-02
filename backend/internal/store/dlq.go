package store

import (
	"database/sql"
	"strconv"
	"time"
)

// DLQItem is one dead-lettered provider item — either a fetch that failed
// retriable (network/5xx) or an item that couldn't be mapped to a RawEvent.
// Persisted instead of log-and-drop so nothing is silently lost.
type DLQItem struct {
	ID           string     `json:"id"`
	Source       string     `json:"source"`
	Monitor      string     `json:"monitor,omitempty"`
	ExternalID   string     `json:"external_id,omitempty"`
	RawPayload   string     `json:"raw_payload,omitempty"`
	ErrorMessage string     `json:"error_message"`
	Retries      int        `json:"retries"`
	LastRetryAt  *time.Time `json:"last_retry_at,omitempty"`
	Status       string     `json:"status"`
	CreatedAt    time.Time  `json:"created_at"`
}

// InsertDLQ writes one failed item to the dead-letter queue.
func (s *Store) InsertDLQ(source, monitor, externalID, rawPayload, errMsg string) error {
	_, err := s.DB.Exec(
		`INSERT INTO dlq_items (id, source, monitor, external_id, raw_payload, error_message) VALUES ($1, $2, $3, $4, $5, $6)`,
		NewID("dlq"), source, monitor, externalID, rawPayload, errMsg,
	)
	return err
}

// ListDLQ returns dead-lettered items, oldest first, bounded by limit. status
// selects "pending" (default when empty), "replayed", or "all".
func (s *Store) ListDLQ(status string, limit int) ([]DLQItem, error) {
	if limit <= 0 {
		limit = 100
	}
	q := `SELECT id, source, COALESCE(monitor,''), COALESCE(external_id,''), COALESCE(raw_payload,''), error_message, retries, last_retry_at, status, created_at
		FROM dlq_items`
	args := []any{}
	if status != "" && status != "all" {
		args = append(args, status)
		q += " WHERE status = $1"
	}
	args = append(args, limit)
	q += " ORDER BY created_at ASC LIMIT $" + strconv.Itoa(len(args))
	rows, err := s.DB.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []DLQItem
	for rows.Next() {
		var d DLQItem
		var lastRetry sql.NullTime
		if err := rows.Scan(&d.ID, &d.Source, &d.Monitor, &d.ExternalID, &d.RawPayload, &d.ErrorMessage, &d.Retries, &lastRetry, &d.Status, &d.CreatedAt); err != nil {
			return nil, err
		}
		d.LastRetryAt = timePtr(lastRetry)
		out = append(out, d)
	}
	return out, rows.Err()
}

// CountDLQPending returns the number of dead-lettered items still pending
// replay. Used by the health check.
func (s *Store) CountDLQPending() (int, error) {
	var n int
	err := s.DB.QueryRow(`SELECT COUNT(*) FROM dlq_items WHERE status = 'pending'`).Scan(&n)
	return n, err
}

// MarkDLQReplayed marks a dead-lettered item as successfully replayed.
func (s *Store) MarkDLQReplayed(id string) error {
	_, err := s.DB.Exec(`UPDATE dlq_items SET status = 'replayed' WHERE id = $1`, id)
	return err
}

// MarkDLQRetried increments the retry counter and updates last_retry_at.
func (s *Store) MarkDLQRetried(id string) error {
	_, err := s.DB.Exec(`UPDATE dlq_items SET retries = retries + 1, last_retry_at = now() WHERE id = $1`, id)
	return err
}
