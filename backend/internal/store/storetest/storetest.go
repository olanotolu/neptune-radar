// Package storetest provides a real-Postgres store for tests. There is no
// in-memory substitute: the production driver is Postgres, and tests that
// ran against a different database proved nothing about the one that ships.
// Tests skip cleanly when TEST_DATABASE_URL is unset; CI provides it via a
// postgres service container.
package storetest

import (
	"os"
	"testing"

	"neptune-social-radar/backend/internal/store"
)

// Open returns a migrated store on a throwaway schema, or skips the test when
// no test database is configured. Each call resets all tables, so tests never
// share state.
func Open(t *testing.T) *store.Store {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set — skipping Postgres-backed test")
	}
	s, err := store.Open(dsn)
	if err != nil {
		t.Fatalf("open test store: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	reset(t, s)
	return s
}

func reset(t *testing.T, s *store.Store) {
	t.Helper()
	tables := []string{
		"audit_events", "executed_actions", "recommended_actions", "neptune_cases",
		"crm_leads", "evidence", "life_event_hypotheses", "consent_policies",
		"relationships", "couples", "edges", "social_observations",
		"pair_cooccurrence_sources", "pair_cooccurrences",
		"social_accounts", "persons", "api_usage", "ingest_cursors",
		"social_sources", "parishes", "church_jurisdictions", "connector_runs",
		"connectors", "source_endpoints", "source_organizations",
		"watched_sources", "cities", "counties", "states",
	}
	for _, table := range tables {
		if _, err := s.DB.Exec(`DELETE FROM ` + table); err != nil {
			t.Fatalf("reset %s: %v", table, err)
		}
	}
}
