// Package connectors performs the actual, real checks against registered
// source endpoints — government record-search pages, church bulletin
// archives, and (via a thin wrapper around the real Apify pipeline)
// Instagram accounts. Every CheckResult here comes from a real network call;
// nothing in this package fabricates a status or a timestamp.
package connectors

import "context"

// CheckResult is one real measurement. Status is "success" or "failure" —
// callers persist it via store.RecordConnectorRun, which is the only place
// a connector's aggregate status is derived from these results.
type CheckResult struct {
	Status             string // "success" | "failure"
	HTTPStatus         int
	ResponseTimeMs     int
	StructureSignature string
	ErrorMessage       string
}

// SourceConnector performs a real check against an endpoint URL. Discover/
// Fetch are intentionally not part of this interface yet — this phase only
// needs reachability and structure checks; record discovery and document
// fetching are a later, separately-scoped extension (see PRODUCTION_GAPS.md).
type SourceConnector interface {
	CheckHealth(ctx context.Context, endpointURL string) CheckResult
}
