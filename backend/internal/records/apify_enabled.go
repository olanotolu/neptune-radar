package records

import (
	"os"
	"strings"
)

// ApifyGloballyEnabled is the master kill-switch for ALL Apify spend
// (people-search TPS actors AND Instagram ingest fallback).
//
// Default: OFF. Requires APIFY_ENABLED=true (or 1/yes/on).
// Even when global is on, TPS still requires APIFY_TPS_ENABLED=true.
func ApifyGloballyEnabled() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("APIFY_ENABLED")))
	return v == "1" || v == "true" || v == "yes" || v == "on"
}

// ApifyGlobalStatus for research notes / logs.
func ApifyGlobalStatus() string {
	if strings.TrimSpace(os.Getenv("APIFY_TOKEN")) == "" {
		return "no APIFY_TOKEN"
	}
	if !ApifyGloballyEnabled() {
		return "PAUSED globally (APIFY_ENABLED=false) — zero Apify API spend"
	}
	return "APIFY_ENABLED=true (TPS still needs APIFY_TPS_ENABLED=true)"
}
