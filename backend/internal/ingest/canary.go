package ingest

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// SchemaDriftCanary checks whether a batch of provider items still contains
// the expected fields. Apify actor upgrades can silently change the JSON
// schema (rename fields, drop fields), which degrades ingestion to zero
// signal without any error — the mapper just sees empty fields. This canary
// catches that: if fewer than minPresentPct of items have the expected fields,
// it returns a drift report.
//
// The expected fields are the ones the mapper depends on:
//   - "id" or "shortCode" (external event ID)
//   - "caption" (the text the whole pipeline reads)
//   - "ownerUsername" (who posted)
//   - "timestamp" or "takenAtTimestamp" (when it happened)
//
// ponytail: ceiling — this is a per-batch check, not continuous. It runs
// after each batch is fetched; a slow drift that takes hours to cross the
// threshold is still caught within one poll interval.
type SchemaDriftReport struct {
	Actor       string    `json:"actor"`
	BatchSize   int       `json:"batch_size"`
	CheckedAt   time.Time `json:"checked_at"`
	FieldStats  map[string]float64 `json:"field_stats"` // field → % present
	Drifted     bool      `json:"drifted"`
	MissingFields []string `json:"missing_fields,omitempty"`
}

// minFieldPresence is the percentage of items that must have a field for it
// to count as "present in the schema." 80% is generous — a few items missing
// a field is normal (e.g. a post with no caption), but if 80%+ are missing it,
// the field was renamed or dropped.
const minFieldPresence = 0.80

// expectedPostFields are the fields the mapper depends on, with their
// acceptable alternatives (either field is fine).
var expectedPostFields = []struct {
	primary string
	alt     string
}{
	{"id", "shortCode"},
	{"caption", ""},
	{"ownerUsername", ""},
	{"timestamp", "takenAtTimestamp"},
}

// CheckSchemaDrift inspects a batch of raw items and returns a drift report.
// If drifted, the report lists which fields are missing. The caller should
// log/alert on drift — the batch is still processed (degraded but not lost).
func CheckSchemaDrift(actor string, items []json.RawMessage) SchemaDriftReport {
	report := SchemaDriftReport{
		Actor:      actor,
		BatchSize:  len(items),
		CheckedAt:  time.Now(),
		FieldStats: make(map[string]float64),
	}
	if len(items) == 0 {
		return report
	}

	for _, ef := range expectedPostFields {
		present := 0
		for _, item := range items {
			var m map[string]json.RawMessage
			if json.Unmarshal(item, &m) != nil {
				continue
			}
			if hasField(m, ef.primary) || (ef.alt != "" && hasField(m, ef.alt)) {
				present++
			}
		}
		pct := float64(present) / float64(len(items))
		fieldName := ef.primary
		if ef.alt != "" {
			fieldName = ef.primary + "/" + ef.alt
		}
		report.FieldStats[fieldName] = pct
		if pct < minFieldPresence {
			report.Drifted = true
			report.MissingFields = append(report.MissingFields,
				fmt.Sprintf("%s (%.0f%% present, need %.0f%%)", fieldName, pct*100, minFieldPresence*100))
		}
	}

	return report
}

// hasField reports whether a field exists and is non-empty in the raw JSON.
// A field that exists but is null or "" counts as missing — the mapper
// can't use an empty value.
func hasField(m map[string]json.RawMessage, field string) bool {
	v, ok := m[field]
	if !ok {
		return false
	}
	s := strings.TrimSpace(string(v))
	return s != "" && s != "null" && s != "\"\""
}
