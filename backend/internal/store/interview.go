package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
)

// InterviewSession is a two-couple conversation under multi-agent extraction.
type InterviewSession struct {
	ID           string `json:"id"`
	CoupleALabel string `json:"couple_a_label"`
	CoupleBLabel string `json:"couple_b_label"`
	Status       string `json:"status"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

// InterviewMessage is one utterance in an interview session.
type InterviewMessage struct {
	ID        string `json:"id"`
	SessionID string `json:"session_id"`
	Speaker   string `json:"speaker"`
	Couple    string `json:"couple"`
	Text      string `json:"text"`
	AudioURL  string `json:"audio_url,omitempty"`
	CreatedAt string `json:"created_at"`
}

// InterviewExtraction is one agent's findings for a session.
type InterviewExtraction struct {
	ID         string          `json:"id"`
	SessionID  string          `json:"session_id"`
	AgentType  string          `json:"agent_type"`
	Findings   json.RawMessage `json:"findings"`
	Confidence float64         `json:"confidence"`
	Summary    string          `json:"summary"`
	CreatedAt  string          `json:"created_at"`
}

// CreateInterviewSession starts a new two-couple interview session.
func (s *Store) CreateInterviewSession(ctx context.Context, coupleA, coupleB string) (InterviewSession, error) {
	if coupleA == "" {
		coupleA = "Couple A"
	}
	if coupleB == "" {
		coupleB = "Couple B"
	}
	id := NewID("interview")
	var sess InterviewSession
	err := s.DB.QueryRowContext(ctx,
		`INSERT INTO interview_sessions (id, couple_a_label, couple_b_label)
		 VALUES ($1, $2, $3)
		 RETURNING id, couple_a_label, couple_b_label, status,
		           to_char(created_at, 'YYYY-MM-DD"T"HH24:MI:SSOF'),
		           to_char(updated_at, 'YYYY-MM-DD"T"HH24:MI:SSOF')`,
		id, coupleA, coupleB,
	).Scan(&sess.ID, &sess.CoupleALabel, &sess.CoupleBLabel, &sess.Status, &sess.CreatedAt, &sess.UpdatedAt)
	if err != nil {
		return InterviewSession{}, fmt.Errorf("create interview session: %w", err)
	}
	return sess, nil
}

// GetInterviewSession fetches a single session by id.
func (s *Store) GetInterviewSession(ctx context.Context, id string) (InterviewSession, error) {
	var sess InterviewSession
	err := s.DB.QueryRowContext(ctx,
		`SELECT id, couple_a_label, couple_b_label, status,
		        to_char(created_at, 'YYYY-MM-DD"T"HH24:MI:SSOF'),
		        to_char(updated_at, 'YYYY-MM-DD"T"HH24:MI:SSOF')
		 FROM interview_sessions WHERE id = $1`, id,
	).Scan(&sess.ID, &sess.CoupleALabel, &sess.CoupleBLabel, &sess.Status, &sess.CreatedAt, &sess.UpdatedAt)
	if err != nil {
		return InterviewSession{}, fmt.Errorf("get interview session: %w", err)
	}
	return sess, nil
}

// ListInterviewSessions returns the most recently created sessions.
func (s *Store) ListInterviewSessions(ctx context.Context, limit int) ([]InterviewSession, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.DB.QueryContext(ctx,
		`SELECT id, couple_a_label, couple_b_label, status,
		        to_char(created_at, 'YYYY-MM-DD"T"HH24:MI:SSOF'),
		        to_char(updated_at, 'YYYY-MM-DD"T"HH24:MI:SSOF')
		 FROM interview_sessions ORDER BY created_at DESC LIMIT $1`, limit)
	if err != nil {
		return nil, fmt.Errorf("list interview sessions: %w", err)
	}
	defer rows.Close()
	var out []InterviewSession
	for rows.Next() {
		var sess InterviewSession
		if err := rows.Scan(&sess.ID, &sess.CoupleALabel, &sess.CoupleBLabel, &sess.Status, &sess.CreatedAt, &sess.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, sess)
	}
	return out, rows.Err()
}

// AddInterviewMessage appends one utterance to a session.
func (s *Store) AddInterviewMessage(ctx context.Context, sessionID, speaker, couple, text, audioURL string) (InterviewMessage, error) {
	id := NewID("imsg")
	var m InterviewMessage
	err := s.DB.QueryRowContext(ctx,
		`INSERT INTO interview_messages (id, session_id, speaker, couple, text, audio_url)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, session_id, speaker, couple, text, COALESCE(audio_url,''),
		           to_char(created_at, 'YYYY-MM-DD"T"HH24:MI:SSOF')`,
		id, sessionID, speaker, couple, text, nullableStr(audioURL),
	).Scan(&m.ID, &m.SessionID, &m.Speaker, &m.Couple, &m.Text, &m.AudioURL, &m.CreatedAt)
	if err != nil {
		return InterviewMessage{}, fmt.Errorf("add interview message: %w", err)
	}
	return m, nil
}

// ListInterviewMessages returns all messages for a session in chronological order.
func (s *Store) ListInterviewMessages(ctx context.Context, sessionID string) ([]InterviewMessage, error) {
	rows, err := s.DB.QueryContext(ctx,
		`SELECT id, session_id, speaker, couple, text, COALESCE(audio_url,''),
		        to_char(created_at, 'YYYY-MM-DD"T"HH24:MI:SSOF')
		 FROM interview_messages WHERE session_id = $1 ORDER BY created_at ASC`, sessionID)
	if err != nil {
		return nil, fmt.Errorf("list interview messages: %w", err)
	}
	defer rows.Close()
	var out []InterviewMessage
	for rows.Next() {
		var m InterviewMessage
		if err := rows.Scan(&m.ID, &m.SessionID, &m.Speaker, &m.Couple, &m.Text, &m.AudioURL, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// SaveExtraction persists one agent's findings for a session.
func (s *Store) SaveExtraction(ctx context.Context, sessionID, agentType string, findings json.RawMessage, confidence float64, summary string) (InterviewExtraction, error) {
	id := NewID("iext")
	var e InterviewExtraction
	err := s.DB.QueryRowContext(ctx,
		`INSERT INTO interview_extractions (id, session_id, agent_type, findings, confidence, summary)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, session_id, agent_type, findings, confidence, summary,
		           to_char(created_at, 'YYYY-MM-DD"T"HH24:MI:SSOF')`,
		id, sessionID, agentType, []byte(findings), confidence, summary,
	).Scan(&e.ID, &e.SessionID, &e.AgentType, &e.Findings, &e.Confidence, &e.Summary, &e.CreatedAt)
	if err != nil {
		return InterviewExtraction{}, fmt.Errorf("save extraction: %w", err)
	}
	return e, nil
}

// ListExtractions returns all extractions for a session, newest first.
func (s *Store) ListExtractions(ctx context.Context, sessionID string) ([]InterviewExtraction, error) {
	rows, err := s.DB.QueryContext(ctx,
		`SELECT id, session_id, agent_type, findings, confidence, summary,
		        to_char(created_at, 'YYYY-MM-DD"T"HH24:MI:SSOF')
		 FROM interview_extractions WHERE session_id = $1 ORDER BY created_at DESC`, sessionID)
	if err != nil {
		return nil, fmt.Errorf("list extractions: %w", err)
	}
	defer rows.Close()
	var out []InterviewExtraction
	for rows.Next() {
		var e InterviewExtraction
		if err := rows.Scan(&e.ID, &e.SessionID, &e.AgentType, &e.Findings, &e.Confidence, &e.Summary, &e.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// UpdateSessionStatus sets the session status (e.g. "completed") and bumps updated_at.
func (s *Store) UpdateSessionStatus(ctx context.Context, id, status string) error {
	res, err := s.DB.ExecContext(ctx,
		`UPDATE interview_sessions SET status = $2, updated_at = now() WHERE id = $1`, id, status)
	if err != nil {
		return fmt.Errorf("update session status: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// nullableStr returns a nil pointer for empty strings so the column stays NULL.
func nullableStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}
