package ingest

import (
	"encoding/json"
	"testing"
)

func TestCheckSchemaDriftNoDrift(t *testing.T) {
	items := []json.RawMessage{
		json.RawMessage(`{"id":"abc","caption":"yes!","ownerUsername":"user1","timestamp":"2026-01-01"}`),
		json.RawMessage(`{"id":"def","caption":"engaged!","ownerUsername":"user2","timestamp":"2026-01-02"}`),
		json.RawMessage(`{"shortCode":"xyz","caption":"she said yes","ownerUsername":"user3","takenAtTimestamp":1234567890}`),
	}
	report := CheckSchemaDrift("test", items)
	if report.Drifted {
		t.Errorf("expected no drift with all fields present, got: %v", report.MissingFields)
	}
	if report.FieldStats["id/shortCode"] != 1.0 {
		t.Errorf("id/shortCode should be 100%% present, got %.2f", report.FieldStats["id/shortCode"])
	}
}

func TestCheckSchemaDriftDetectsMissingCaption(t *testing.T) {
	// 90% of items missing caption = drift
	items := []json.RawMessage{}
	for i := 0; i < 10; i++ {
		items = append(items, json.RawMessage(`{"id":"x","ownerUsername":"u","timestamp":"2026-01-01"}`))
	}
	items = append(items, json.RawMessage(`{"id":"y","caption":"rare","ownerUsername":"u","timestamp":"2026-01-01"}`))
	report := CheckSchemaDrift("test", items)
	if !report.Drifted {
		t.Error("expected drift when caption is missing in 90% of items")
	}
}

func TestCheckSchemaDriftEmptyBatch(t *testing.T) {
	report := CheckSchemaDrift("test", nil)
	if report.Drifted {
		t.Error("empty batch should not report drift")
	}
}

func TestCheckSchemaDriftEmptyFieldsCount(t *testing.T) {
	// Fields that exist but are empty/null should count as missing.
	items := []json.RawMessage{
		json.RawMessage(`{"id":"abc","caption":"","ownerUsername":"user1","timestamp":"2026-01-01"}`),
		json.RawMessage(`{"id":"def","caption":"","ownerUsername":"user2","timestamp":"2026-01-02"}`),
	}
	report := CheckSchemaDrift("test", items)
	if !report.Drifted {
		t.Error("empty captions should count as missing → drift")
	}
}

func TestHasField(t *testing.T) {
	m := map[string]json.RawMessage{
		"present":  json.RawMessage(`"value"`),
		"empty":    json.RawMessage(`""`),
		"null":     json.RawMessage(`null`),
		"whitespace": json.RawMessage(`  `),
	}
	if !hasField(m, "present") {
		t.Error("present field should return true")
	}
	if hasField(m, "empty") {
		t.Error("empty string should return false")
	}
	if hasField(m, "null") {
		t.Error("null should return false")
	}
	if hasField(m, "whitespace") {
		t.Error("whitespace-only should return false")
	}
	if hasField(m, "missing") {
		t.Error("missing field should return false")
	}
}
