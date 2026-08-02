package store

import (
	"database/sql"
	"time"
)

// Notification is one durable in-app inbox item. Created by the pipeline
// for key events (action_created, stage_transition, dlq_item, source_stale).
type Notification struct {
	ID         string     `json:"id"`
	Type       string     `json:"type"`
	Title      string     `json:"title"`
	Body       string     `json:"body,omitempty"`
	EntityType string     `json:"entity_type,omitempty"`
	EntityID   string     `json:"entity_id,omitempty"`
	Severity   string     `json:"severity"`
	ReadAt     *time.Time `json:"read_at,omitempty"`
	AckedAt    *time.Time `json:"acked_at,omitempty"`
	AckedBy    string     `json:"acked_by,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

// CreateNotification inserts one notification row.
func (s *Store) CreateNotification(n Notification) error {
	if n.ID == "" {
		n.ID = NewID("notif")
	}
	if n.Severity == "" {
		n.Severity = "info"
	}
	_, err := s.DB.Exec(
		`INSERT INTO notifications (id, type, title, body, entity_type, entity_id, severity)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		n.ID, n.Type, n.Title, n.Body, n.EntityType, n.EntityID, n.Severity)
	return err
}

// ListNotifications returns notifications newest first. If unreadOnly is
// true, only unread notifications are returned.
func (s *Store) ListNotifications(unreadOnly bool, limit int) ([]Notification, error) {
	if limit <= 0 {
		limit = 50
	}
	q := `SELECT id, type, title, COALESCE(body,''), COALESCE(entity_type,''),
	       COALESCE(entity_id,''), severity, read_at, acked_at, COALESCE(acked_by,''),
	       created_at
	    FROM notifications`
	if unreadOnly {
		q += ` WHERE read_at IS NULL`
	}
	q += ` ORDER BY created_at DESC LIMIT $1`
	rows, err := s.DB.Query(q, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Notification
	for rows.Next() {
		var n Notification
		if err := rows.Scan(&n.ID, &n.Type, &n.Title, &n.Body, &n.EntityType,
			&n.EntityID, &n.Severity, &n.ReadAt, &n.AckedAt, &n.AckedBy,
			&n.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

// MarkNotificationRead marks a notification as read.
func (s *Store) MarkNotificationRead(id string) error {
	_, err := s.DB.Exec(`UPDATE notifications SET read_at = now() WHERE id = $1 AND read_at IS NULL`, id)
	return err
}

// MarkNotificationAcked marks a notification as acknowledged by a user.
func (s *Store) MarkNotificationAcked(id, ackedBy string) error {
	_, err := s.DB.Exec(`UPDATE notifications SET acked_at = now(), acked_by = $2, read_at = COALESCE(read_at, now()) WHERE id = $1`, id, ackedBy)
	return err
}

// CountUnreadNotifications returns the count of unread notifications.
func (s *Store) CountUnreadNotifications() (int, error) {
	var n int
	err := s.DB.QueryRow(`SELECT COUNT(*) FROM notifications WHERE read_at IS NULL`).Scan(&n)
	return n, err
}

// MarkAllNotificationsRead marks all unread notifications as read.
func (s *Store) MarkAllNotificationsRead() (int64, error) {
	res, err := s.DB.Exec(`UPDATE notifications SET read_at = now() WHERE read_at IS NULL`)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

var _ = sql.ErrNoRows // keep import for future use
