// Command scrape-marriages runs the MarriageRecordScraper against a single
// county marriage-record portal identified by URL + county FIPS.
//
// Usage:
//
//	go run ./cmd/scrape-marriages -url=https://dallas.tx.publicsearch.us/ -fips=48113
//	go run ./cmd/scrape-marriages -url=https://dallas.tx.publicsearch.us/ -fips=48113 -name=SMITH
//	go run ./cmd/scrape-marriages -url=https://dallas.tx.publicsearch.us/ -fips=48113 -from=2024-01-01 -to=2024-12-31
//
// Only the publicsearch.us portal family is implemented; other portal types
// return "not yet implemented" so you can see which counties still need work.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"time"

	"neptune-social-radar/backend/internal/connectors"
	"neptune-social-radar/backend/internal/packs"
)

func main() {
	state := flag.String("state", "", "USPS state code (e.g. TX)")
	fips := flag.String("fips", "", "5-digit county FIPS code (e.g. 48113)")
	searchURL := flag.String("url", "", "override: search portal URL (skips pack lookup)")
	namePrefix := flag.String("name", "", "optional: name prefix to narrow results")
	fromStr := flag.String("from", "", "optional: start date YYYY-MM-DD")
	toStr := flag.String("to", "", "optional: end date YYYY-MM-DD")
	flag.Parse()

	if *fips == "" {
		log.Fatal("-fips is required (5-digit county FIPS)")
	}

	// Resolve the search URL: explicit -url wins, otherwise look up from the
	// verified govSource packs by state + FIPS.
	url := *searchURL
	var courtName, note string
	if url == "" {
		if *state == "" {
			log.Fatal("-state is required when -url is not given")
		}
		pack := packs.PackFor(*state)
		if pack == nil {
			log.Fatalf("no source pack found for state %s", *state)
		}
		found := false
		for _, g := range pack.Government {
			if g.CountyFIPS == *fips {
				url = g.SearchURL
				courtName = g.CourtName
				note = g.Note
				found = true
				break
			}
		}
		if !found {
			log.Fatalf("no government source found for FIPS %s in state %s", *fips, *state)
		}
	}

	// Parse optional date bounds.
	var dateFrom, dateTo time.Time
	if *fromStr != "" {
		t, err := time.Parse("2006-01-02", *fromStr)
		if err != nil {
			log.Fatalf("invalid -from date %q: %v (use YYYY-MM-DD)", *fromStr, err)
		}
		dateFrom = t
	}
	if *toStr != "" {
		t, err := time.Parse("2006-01-02", *toStr)
		if err != nil {
			log.Fatalf("invalid -to date %q: %v (use YYYY-MM-DD)", *toStr, err)
		}
		dateTo = t
	}

	pt := connectors.ClassifyPortal(url)
	fmt.Fprintf(log.Writer(), "portal: %s\n", pt)
	fmt.Fprintf(log.Writer(), "url:    %s\n", url)
	if courtName != "" {
		fmt.Fprintf(log.Writer(), "court:  %s\n", courtName)
	}
	if note != "" {
		fmt.Fprintf(log.Writer(), "note:   %s\n", note)
	}
	fmt.Fprintln(log.Writer())

	scraper := connectors.NewMarriageRecordScraper()
	query := connectors.MarriageSearchQuery{
		CountyFIPS: *fips,
		NamePrefix: *namePrefix,
		DateFrom:   dateFrom,
		DateTo:     dateTo,
		SearchURL:  url,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	records, err := scraper.Scrape(ctx, query)
	if err != nil {
		log.Fatalf("scrape failed: %v", err)
	}

	fmt.Printf("found %d marriage record(s)\n\n", len(records))
	for i, r := range records {
		fmt.Printf("[%d] %s & %s\n", i+1, r.Party1Name, r.Party2Name)
		if !r.MarriageDate.IsZero() {
			fmt.Printf("    date: %s\n", r.MarriageDate.Format("2006-01-02"))
		}
		if r.BookPage != "" {
			fmt.Printf("    book/page: %s\n", r.BookPage)
		}
		fmt.Printf("    fips: %s\n", r.CountyFIPS)
		fmt.Println()
	}

	// Also emit JSON for piping into other tools.
	if len(records) > 0 {
		out, _ := json.MarshalIndent(records, "", "  ")
		fmt.Println(string(out))
	}
}
