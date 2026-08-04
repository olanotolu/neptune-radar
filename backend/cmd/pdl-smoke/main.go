package main

import (
	"context"
	"fmt"
	"os"

	"neptune-social-radar/backend/internal/records"
)

func main() {
	p := &records.PDL{APIKey: os.Getenv("PDL_API_KEY")}
	for _, q := range []records.Query{
		{FirstName: "Carly", LastName: "Jordan", City: "College Station", Region: "TX", Handle: "carlyyjordan"},
		{FirstName: "Baylor", LastName: "Dawes", City: "College Station", Region: "TX", Handle: "dawes.baylor"},
	} {
		res, err := p.Search(context.Background(), q)
		fmt.Printf("\n%s %s status=%s err=%v cands=%d\n", q.FirstName, q.LastName, res.Status, err, len(res.Candidates))
		for i, c := range res.Candidates {
			fmt.Printf("  [%d] kind=%s line1=%q city=%s region=%s zip=%s conf=%.2f\n      note=%s\n",
				i, c.Kind, c.Line1, c.City, c.Region, c.Postal, c.Confidence, c.Note)
		}
	}
}
