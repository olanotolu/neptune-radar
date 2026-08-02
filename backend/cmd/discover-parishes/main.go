// Command discover-parishes scrapes the parish-finder directory of every
// diocese in a state pack and reports the parishes it finds — including, where
// reachable, a link to each parish's own bulletin archive.
//
// It does NOT write to the database. It only discovers and reports, so a bad
// scrape can be reviewed before anything is registered. Wire the output into
// bootstrap-state's curated packs once a human has eyeballed it.
//
// Usage:
//
//	go run ./cmd/discover-parishes NY
//	go run ./cmd/discover-parishes TX
//
// A failed diocese is logged and skipped; it never stops the whole run.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"neptune-social-radar/backend/internal/connectors"
	"neptune-social-radar/backend/internal/packs"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("usage: discover-parishes <STATE> (e.g. NY)")
	}
	st := strings.ToUpper(strings.TrimSpace(os.Args[1]))
	pack := packs.PackFor(st)
	if pack == nil {
		log.Fatalf("no state pack defined for %s", st)
	}
	if len(pack.Dioceses) == 0 {
		log.Fatalf("state %s has no dioceses in its pack", st)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	discoverer := connectors.NewParishDiscoveryConnector()
	bulletins := connectors.NewBulletinDiscoveryConnector()

	type parishHit struct {
		Diocese     string
		Name        string
		Address     string
		WebsiteURL  string
		BulletinURL string
	}
	var all []parishHit
	totalBulletins := 0

	for _, d := range pack.Dioceses {
		if d.Directory == "" {
			log.Printf("[discover-parishes] %s: %s has no directory URL — skipping", st, d.Name)
			continue
		}
		log.Printf("[discover-parishes] %s: scanning %s — %s", st, d.Name, d.Directory)

		parishes, err := discoverer.DiscoverParishes(ctx, d.Directory)
		if err != nil {
			// A failed diocese is logged and skipped, never fatal.
			log.Printf("[discover-parishes]   FAIL %s: %v", d.Name, err)
			continue
		}
		log.Printf("[discover-parishes]   found %d parishes for %s", len(parishes), d.Name)

		for _, p := range parishes {
			hit := parishHit{Diocese: d.Name, Name: p.Name, Address: p.Address, WebsiteURL: p.WebsiteURL}
			// Only probe the parish's own site for a bulletin link when we
			// actually have one — never guess a URL.
			if p.WebsiteURL != "" {
				if bu, err := bulletins.DiscoverBulletinURL(ctx, p.WebsiteURL); err != nil {
					log.Printf("[discover-parishes]     bulletin probe failed for %s: %v", p.Name, err)
				} else if bu != "" {
					hit.BulletinURL = bu
					totalBulletins++
				}
			}
			all = append(all, hit)
		}
	}

	// --- Summary --------------------------------------------------------------
	fmt.Printf("\n=== discover-parishes %s: %d parishes across %d dioceses (%d bulletin archives) ===\n",
		st, len(all), len(pack.Dioceses), totalBulletins)
	for _, h := range all {
		fmt.Printf("- [%s] %s", h.Diocese, h.Name)
		if h.Address != "" {
			fmt.Printf(" | %s", h.Address)
		}
		if h.WebsiteURL != "" {
			fmt.Printf(" | site=%s", h.WebsiteURL)
		}
		if h.BulletinURL != "" {
			fmt.Printf(" | bulletin=%s", h.BulletinURL)
		}
		fmt.Println()
	}
	fmt.Printf("\n(dry run — nothing was written to the database)\n")
}
