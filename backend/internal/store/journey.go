package store

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// JourneyEvent is one node on the couple journey timeline — the "wow" demo
// view that tells the complete story from first signal to postcard mailed.
type JourneyEvent struct {
	Timestamp   string  `json:"timestamp"`
	EventType   string  `json:"event_type"`
	Description string  `json:"description"`
	Confidence  float64 `json:"confidence,omitempty"`
	Source      string  `json:"source,omitempty"`
}

// GetCoupleJourney builds a chronological timeline of every milestone for a
// couple by merging evidence, audit, kit, and funnel signals. Works with
// partial data — a couple with only a created_at row still renders one node.
// ponytail: ceiling = O(n) merge of 5 bounded queries (≤200 audit, ≤100 funnel);
// upgrade path is a materialized journey_events table if this view gets hot.
func (s *Store) GetCoupleJourney(coupleID string) ([]JourneyEvent, error) {
	var events []JourneyEvent

	// 1. Couple created = first signal detected.
	var createdAt time.Time
	if err := s.DB.QueryRow(`SELECT created_at FROM couples WHERE id = $1`, coupleID).Scan(&createdAt); err != nil {
		return nil, err
	}
	events = append(events, JourneyEvent{
		Timestamp:   createdAt.UTC().Format(time.RFC3339),
		EventType:   "signal_detected",
		Description: "First signal detected — couple created",
		Source:      "couples",
	})

	// 2. Evidence rows for the latest hypothesis (each piece with weight).
	var hypID string
	_ = s.DB.QueryRow(
		`SELECT id FROM life_event_hypotheses WHERE couple_id = $1 ORDER BY created_at DESC LIMIT 1`,
		coupleID,
	).Scan(&hypID)
	if hypID != "" {
		if ev, err := s.EvidenceForHypothesis(hypID); err == nil {
			for _, e := range ev {
				events = append(events, JourneyEvent{
					Timestamp:   e.CreatedAt.UTC().Format(time.RFC3339),
					EventType:   "evidence_added",
					Description: fmt.Sprintf("%s — %s", e.Kind, e.Description),
					Confidence:  e.Weight,
					Source:      "evidence",
				})
			}
		}
	}

	// 3. Audit events for the couple. Skip funnel_* rows — the funnel query
	// below covers those without duplicating.
	audits, _ := s.ListAudit(AuditFilter{EntityType: "couple", EntityID: coupleID, Limit: 200})
	for _, a := range audits {
		if strings.HasPrefix(a.Event, "funnel_") {
			continue
		}
		events = append(events, JourneyEvent{
			Timestamp:   a.CreatedAt.UTC().Format(time.RFC3339),
			EventType:   a.Event,
			Description: auditJourneyLabel(a.Event),
			Source:      "audit",
		})
	}

	// 4. Kit milestones (built, address found, mailed, follow-up).
	if kit, err := s.GetLatestKitForCouple(coupleID); err == nil {
		events = append(events, JourneyEvent{
			Timestamp:   kit.CreatedAt.UTC().Format(time.RFC3339),
			EventType:   "kit_built",
			Description: "Congratulate kit built",
			Source:      "kit",
		})
		if kit.VerifiedAt != nil {
			loc := strings.TrimSpace(kit.AddressCity + ", " + kit.AddressRegion)
			loc = strings.TrimSuffix(loc, ", ")
			events = append(events, JourneyEvent{
				Timestamp:   kit.VerifiedAt.UTC().Format(time.RFC3339),
				EventType:   "address_found",
				Description: "Address verified" + ifNotEmpty(loc, " — "+loc),
				Confidence:  kit.AddressConfidence,
				Source:      "kit",
			})
		}
		if kit.MailedAt != nil {
			events = append(events, JourneyEvent{
				Timestamp:   kit.MailedAt.UTC().Format(time.RFC3339),
				EventType:   "postcard_mailed",
				Description: "Postcard mailed",
				Source:      "kit",
			})
		}
		if kit.FollowUpSentAt != nil {
			events = append(events, JourneyEvent{
				Timestamp:   kit.FollowUpSentAt.UTC().Format(time.RFC3339),
				EventType:   "follow_up_sent",
				Description: "Follow-up sent",
				Source:      "kit",
			})
		}
	}

	// 5. Funnel events (chat_started, consult_booked, closed_won, ...).
	funnel, _ := s.ListFunnelEvents(coupleID, 100)
	for _, f := range funnel {
		events = append(events, JourneyEvent{
			Timestamp:   f.OccurredAt.UTC().Format(time.RFC3339),
			EventType:   f.EventType,
			Description: funnelJourneyLabel(f.EventType),
			Source:      "funnel",
		})
	}

	// Chronological, ascending.
	sort.SliceStable(events, func(i, j int) bool {
		return events[i].Timestamp < events[j].Timestamp
	})
	return events, nil
}

func ifNotEmpty(s, prefix string) string {
	if s == "" {
		return ""
	}
	return prefix + s
}

// auditJourneyLabel maps known audit event names to human descriptions;
// unknown events fall through to a prettified version of the event name.
func auditJourneyLabel(event string) string {
	switch event {
	case "automation_paused":
		return "Automation paused by operator"
	case "automation_resumed":
		return "Automation resumed"
	case "marked_mistaken":
		return "Marked as not a couple (mistaken)"
	case "handoff_issued":
		return "Tracked chat handoff issued"
	case "journey_stage_set":
		return "Journey stage set by operator"
	case "suppressed":
		return "Couple suppressed"
	default:
		return strings.ReplaceAll(event, "_", " ")
	}
}

// funnelJourneyLabel maps funnel event types to human descriptions.
func funnelJourneyLabel(t string) string {
	switch t {
	case "chat_started":
		return "Couple started a chat"
	case "consult_booked":
		return "Consultation booked"
	case "closed_won":
		return "Closed won — client signed"
	case "closed_lost":
		return "Closed lost"
	case "handoff_clicked":
		return "Handoff link clicked"
	default:
		return strings.ReplaceAll(t, "_", " ")
	}
}
