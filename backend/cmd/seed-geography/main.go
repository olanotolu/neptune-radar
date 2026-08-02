// Command seed-geography upserts all 50 US states + DC and every county
// from the us-atlas/Census TIGER-derived FIPS list into the source registry.
//
// No organizations, connectors, or vendors are invented — only geography.
// Safe to re-run (idempotent upserts).
//
// Usage:
//
//	DATABASE_URL=… go run ./cmd/seed-geography
package main

import (
	"fmt"
	"log"
	"os"

	"neptune-social-radar/backend/internal/geo"
	"neptune-social-radar/backend/internal/store"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	s, err := store.Open(dsn)
	if err != nil {
		log.Fatalf("open store: %v", err)
	}
	defer s.Close()

	log.Printf("[seed-geography] upserting %d states…", len(geo.States))
	for _, st := range geo.States {
		if err := s.UpsertState(st.ID, st.Name); err != nil {
			log.Fatalf("state %s: %v", st.ID, err)
		}
	}

	log.Printf("[seed-geography] upserting %d counties…", len(geo.Counties))
	n := 0
	for _, c := range geo.Counties {
		if err := s.UpsertCounty(c.FIPS, c.StateID, c.Name); err != nil {
			log.Fatalf("county %s: %v", c.FIPS, err)
		}
		n++
		if n%500 == 0 {
			log.Printf("[seed-geography] …%d/%d", n, len(geo.Counties))
		}
	}

	if _, err := s.Audit("bootstrap", "geography", "seed_geography_completed", map[string]any{
		"states":   len(geo.States),
		"counties": len(geo.Counties),
	}, "", -1); err != nil {
		log.Printf("[seed-geography] audit write failed: %v", err)
	}

	fmt.Printf("seed-geography ok: %d states, %d counties\n", len(geo.States), len(geo.Counties))
}
