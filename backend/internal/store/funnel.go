package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// FunnelEvent is one product-side signal closing the growth loop.
type FunnelEvent struct {
	ID                 string    `json:"id"`
	CoupleID           string    `json:"couple_id,omitempty"`
	EventType          string    `json:"event_type"`
	ExternalID         string    `json:"external_id,omitempty"`
	HandoffCode        string    `json:"handoff_code,omitempty"`
	Source             string    `json:"source"`
	PayloadJSON        string    `json:"payload_json,omitempty"`
	JourneyStageBefore string    `json:"journey_stage_before,omitempty"`
	JourneyStageAfter  string    `json:"journey_stage_after,omitempty"`
	MatchedBy          string    `json:"matched_by,omitempty"`
	OccurredAt         time.Time `json:"occurred_at"`
	CreatedAt          time.Time `json:"created_at"`
}

// FunnelIngest is the webhook/operator payload to record a funnel event.
type FunnelIngest struct {
	EventType   string         `json:"event"`
	CoupleID    string         `json:"couple_id,omitempty"`
	HandoffCode string         `json:"handoff_code,omitempty"`
	// UTMContent is typically the couple_id we put in utm_content on handoff links.
	UTMContent  string         `json:"utm_content,omitempty"`
	ExternalID  string         `json:"external_id,omitempty"`
	Source      string         `json:"source,omitempty"`
	OccurredAt  *time.Time     `json:"occurred_at,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

// FunnelStats is conversion rollup for the ops dashboard.
type FunnelStats struct {
	ChatStarted7d   int `json:"chat_started_7d"`
	ConsultBooked7d int `json:"consult_booked_7d"`
	ClosedWon7d     int `json:"closed_won_7d"`
	ClosedLost7d    int `json:"closed_lost_7d"`
	HandoffsIssued  int `json:"handoffs_issued"`
	// Simple conversion: chats / handoffs (0 if no handoffs).
	ChatRate float64 `json:"chat_rate"`
	// Booked / chats
	BookRate float64 `json:"book_rate"`
}

var eventToJourney = map[string]string{
	"chat_started":    "in_chat",
	"consult_booked":  "booked",
	"closed_won":      "closed_won",
	"closed_lost":     "closed_lost",
	"handoff_clicked": "invited",
}

// journeyRank prevents regressions (booked should not go back to in_chat).
var journeyRank = map[string]int{
	"detected": 0, "approved": 1, "congratulated": 2, "invited": 3,
	"in_chat": 4, "booked": 5, "closed_won": 6, "closed_lost": 6, "do_not_contact": 99,
}

// ResolveCoupleForFunnel finds the couple from explicit id, handoff code, or utm_content.
func (s *Store) ResolveCoupleForFunnel(coupleID, handoffCode, utmContent string) (id string, matchedBy string, err error) {
	coupleID = strings.TrimSpace(coupleID)
	handoffCode = strings.TrimSpace(handoffCode)
	utmContent = strings.TrimSpace(utmContent)

	if coupleID != "" {
		var exists string
		err := s.DB.QueryRow(`SELECT id FROM couples WHERE id = $1`, coupleID).Scan(&exists)
		if err == nil {
			return exists, "couple_id", nil
		}
		if err != sql.ErrNoRows {
			return "", "", err
		}
	}
	if handoffCode != "" {
		var found string
		err := s.DB.QueryRow(`SELECT id FROM couples WHERE handoff_code = $1`, handoffCode).Scan(&found)
		if err == nil {
			return found, "handoff_code", nil
		}
		if err != sql.ErrNoRows {
			return "", "", err
		}
	}
	if utmContent != "" {
		// utm_content is couple id on our handoff links
		var found string
		err := s.DB.QueryRow(`SELECT id FROM couples WHERE id = $1 OR handoff_code = $1`, utmContent).Scan(&found)
		if err == nil {
			return found, "utm_content", nil
		}
		if err != sql.ErrNoRows {
			return "", "", err
		}
	}
	return "", "unresolved", nil
}

// IngestFunnelEvent records a product event, advances journey when matched, audits.
// Idempotent on (source, external_id) when external_id is set.
func (s *Store) IngestFunnelEvent(in FunnelIngest) (FunnelEvent, error) {
	evType := strings.TrimSpace(strings.ToLower(in.EventType))
	if _, ok := eventToJourney[evType]; !ok {
		return FunnelEvent{}, fmt.Errorf("unknown funnel event type %q", in.EventType)
	}
	source := in.Source
	if source == "" {
		source = "webhook"
	}

	// Idempotency
	if in.ExternalID != "" {
		var existing FunnelEvent
		err := s.DB.QueryRow(`
			SELECT id, COALESCE(couple_id,''), event_type, COALESCE(external_id,''), COALESCE(handoff_code,''),
			       source, COALESCE(payload_json,'{}'), COALESCE(journey_stage_before,''),
			       COALESCE(journey_stage_after,''), COALESCE(matched_by,''), occurred_at, created_at
			FROM funnel_events WHERE source = $1 AND external_id = $2`, source, in.ExternalID,
		).Scan(
			&existing.ID, &existing.CoupleID, &existing.EventType, &existing.ExternalID, &existing.HandoffCode,
			&existing.Source, &existing.PayloadJSON, &existing.JourneyStageBefore, &existing.JourneyStageAfter,
			&existing.MatchedBy, &existing.OccurredAt, &existing.CreatedAt,
		)
		if err == nil {
			return existing, nil
		}
		if err != sql.ErrNoRows {
			return FunnelEvent{}, err
		}
	}

	coupleID, matchedBy, err := s.ResolveCoupleForFunnel(in.CoupleID, in.HandoffCode, in.UTMContent)
	if err != nil {
		return FunnelEvent{}, err
	}

	meta := in.Metadata
	if meta == nil {
		meta = map[string]any{}
	}
	payload, _ := json.Marshal(meta)

	occurred := time.Now().UTC()
	if in.OccurredAt != nil && !in.OccurredAt.IsZero() {
		occurred = in.OccurredAt.UTC()
	}

	ev := FunnelEvent{
		ID:          NewID("funnel"),
		CoupleID:    coupleID,
		EventType:   evType,
		ExternalID:  in.ExternalID,
		HandoffCode: in.HandoffCode,
		Source:      source,
		PayloadJSON: string(payload),
		MatchedBy:   matchedBy,
		OccurredAt:  occurred,
		CreatedAt:   time.Now().UTC(),
	}

	targetStage := eventToJourney[evType]
	if coupleID != "" {
		var before string
		_ = s.DB.QueryRow(`SELECT COALESCE(journey_stage,'detected') FROM couples WHERE id = $1`, coupleID).Scan(&before)
		ev.JourneyStageBefore = before
		// Advance only if not regressing (closed_lost can set from any non-terminal)
		if shouldAdvanceJourney(before, targetStage) {
			if err := s.SetJourneyStage(coupleID, targetStage); err != nil {
				return FunnelEvent{}, err
			}
			ev.JourneyStageAfter = targetStage
		} else {
			ev.JourneyStageAfter = before
		}
	}

	_, err = s.DB.Exec(`
		INSERT INTO funnel_events (
		  id, couple_id, event_type, external_id, handoff_code, source, payload_json,
		  journey_stage_before, journey_stage_after, matched_by, occurred_at, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		ev.ID, nullIfEmpty(ev.CoupleID), ev.EventType, nullIfEmpty(ev.ExternalID), nullIfEmpty(ev.HandoffCode),
		ev.Source, ev.PayloadJSON, nullIfEmpty(ev.JourneyStageBefore), nullIfEmpty(ev.JourneyStageAfter),
		ev.MatchedBy, ev.OccurredAt, ev.CreatedAt,
	)
	if err != nil {
		return FunnelEvent{}, err
	}

	_, _ = s.Audit("couple", firstNonEmpty(coupleID, "unresolved"), "funnel_"+evType, map[string]any{
		"funnel_event_id": ev.ID,
		"matched_by":      matchedBy,
		"external_id":     in.ExternalID,
		"stage_after":     ev.JourneyStageAfter,
	}, "funnel", -1)

	return ev, nil
}

func shouldAdvanceJourney(current, target string) bool {
	if target == "" {
		return false
	}
	// Always allow closed_lost / do_not_contact from non-closed
	if target == "closed_lost" && current != "closed_won" && current != "do_not_contact" {
		return true
	}
	cr, okC := journeyRank[current]
	tr, okT := journeyRank[target]
	if !okT {
		return false
	}
	if !okC {
		return true
	}
	return tr > cr
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

// ListFunnelEvents returns recent funnel events, optionally for one couple.
func (s *Store) ListFunnelEvents(coupleID string, limit int) ([]FunnelEvent, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	q := `
		SELECT id, COALESCE(couple_id,''), event_type, COALESCE(external_id,''), COALESCE(handoff_code,''),
		       source, COALESCE(payload_json,'{}'), COALESCE(journey_stage_before,''),
		       COALESCE(journey_stage_after,''), COALESCE(matched_by,''), occurred_at, created_at
		FROM funnel_events`
	args := []any{}
	if coupleID != "" {
		q += ` WHERE couple_id = $1`
		args = append(args, coupleID)
	}
	q += ` ORDER BY occurred_at DESC LIMIT ` + fmt.Sprintf("%d", limit)

	rows, err := s.DB.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []FunnelEvent
	for rows.Next() {
		var e FunnelEvent
		if err := rows.Scan(
			&e.ID, &e.CoupleID, &e.EventType, &e.ExternalID, &e.HandoffCode,
			&e.Source, &e.PayloadJSON, &e.JourneyStageBefore, &e.JourneyStageAfter,
			&e.MatchedBy, &e.OccurredAt, &e.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// GetFunnelStats rolls up last-7-day conversion signals.
func (s *Store) GetFunnelStats() (FunnelStats, error) {
	var st FunnelStats
	countType := func(t string) int {
		var n int
		_ = s.DB.QueryRow(`
			SELECT COUNT(*) FROM funnel_events
			WHERE event_type = $1 AND occurred_at > now() - interval '7 days'`, t).Scan(&n)
		return n
	}
	st.ChatStarted7d = countType("chat_started")
	st.ConsultBooked7d = countType("consult_booked")
	st.ClosedWon7d = countType("closed_won")
	st.ClosedLost7d = countType("closed_lost")
	_ = s.DB.QueryRow(`SELECT COUNT(*) FROM couples WHERE handoff_url IS NOT NULL AND handoff_url <> ''`).Scan(&st.HandoffsIssued)

	if st.HandoffsIssued > 0 {
		st.ChatRate = float64(st.ChatStarted7d) / float64(st.HandoffsIssued)
		if st.ChatRate > 1 {
			st.ChatRate = 1
		}
	}
	if st.ChatStarted7d > 0 {
		st.BookRate = float64(st.ConsultBooked7d) / float64(st.ChatStarted7d)
		if st.BookRate > 1 {
			st.BookRate = 1
		}
	}
	return st, nil
}
