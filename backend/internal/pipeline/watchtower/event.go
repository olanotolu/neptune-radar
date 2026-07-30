// Package watchtower is the ingestion boundary — the pipeline's eyes. It
// knows nothing about relationships or Neptune: its only job is producing
// RawEvents from monitored sources (hashtag monitors, watched vendor
// accounts, profile enrichment) for the pipeline to interpret. The Apify
// connector in this package is the production implementation.
package watchtower

import "time"

// RawEvent is the ingestion stage's output shape — deliberately close to what
// the upstream data provider delivers.
type RawEvent struct {
	// Monitor identifies which watch source produced the event, e.g.
	// "hashtag:justengaged", "vendor:weddingsbynoor", or "profile:maya".
	// It doubles as the audit trail's grouping key (the pipeline itself is
	// monitor-agnostic).
	Monitor string
	// Source is the data provider that delivered the event ("apify", ...).
	Source string
	// ExternalEventID is the provider-native idempotency key (post ID,
	// profile snapshot hash, ...). The store enforces uniqueness per
	// (monitor, external_event_id).
	ExternalEventID string
	Handle          string
	Type            string // "post","bio_change","follow_change","post_archived","account_disabled","account_private","username_change"
	Payload         map[string]any
	OccurredAt      time.Time
}
