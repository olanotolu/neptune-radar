package ingest

import (
	"encoding/json"
	"testing"
	"time"

	"neptune-social-radar/backend/internal/ontology"
)

func TestParseFenrisResponse(t *testing.T) {
	sample := `{
		"events": [
			{
				"event_type": "Newly Engaged",
				"person_name": "Jane Smith",
				"household_id": "HH-12345",
				"address": "123 Main St",
				"city": "Columbus",
				"state": "OH",
				"zip": "43215",
				"event_date": "2025-07-15",
				"confidence": 0.92
			},
			{
				"event_type": "Newly Married",
				"person_name": "John Doe",
				"household_id": "HH-67890",
				"city": "Austin",
				"state": "TX",
				"event_date": "2025-07-20",
				"confidence": 0.88
			}
		]
	}`

	var fr fenrisResponse
	if err := json.Unmarshal([]byte(sample), &fr); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(fr.Events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(fr.Events))
	}

	// Parse into LifeEvent structs (same logic as FetchLifeEvents).
	events := make([]ontology.LifeEvent, 0, len(fr.Events))
	for _, fe := range fr.Events {
		date, err := time.Parse("2006-01-02", fe.EventDate)
		if err != nil {
			t.Fatalf("parse date %q: %v", fe.EventDate, err)
		}
		events = append(events, ontology.LifeEvent{
			EventType:   fe.EventType,
			PersonName:  fe.PersonName,
			HouseholdID: fe.HouseholdID,
			Address:     fe.Address,
			City:        fe.City,
			State:       fe.State,
			Zip:         fe.Zip,
			EventDate:   date,
			Confidence:  fe.Confidence,
		})
	}

	// First event: Newly Engaged in Columbus, OH
	if events[0].EventType != "Newly Engaged" {
		t.Errorf("event[0] type: got %q, want %q", events[0].EventType, "Newly Engaged")
	}
	if events[0].PersonName != "Jane Smith" {
		t.Errorf("event[0] name: got %q, want %q", events[0].PersonName, "Jane Smith")
	}
	if events[0].State != "OH" {
		t.Errorf("event[0] state: got %q, want %q", events[0].State, "OH")
	}
	if events[0].Confidence != 0.92 {
		t.Errorf("event[0] confidence: got %f, want 0.92", events[0].Confidence)
	}
	if !events[0].EventDate.Equal(time.Date(2025, 7, 15, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("event[0] date: got %v, want 2025-07-15", events[0].EventDate)
	}

	// Second event: Newly Married in Austin, TX (no address field)
	if events[1].EventType != "Newly Married" {
		t.Errorf("event[1] type: got %q, want %q", events[1].EventType, "Newly Married")
	}
	if events[1].Address != "" {
		t.Errorf("event[1] address: got %q, want empty", events[1].Address)
	}
	if events[1].State != "TX" {
		t.Errorf("event[1] state: got %q, want %q", events[1].State, "TX")
	}
}
