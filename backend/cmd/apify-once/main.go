package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"neptune-social-radar/backend/internal/records"
)

func main() {
	fmt.Println("Apify status:", records.ApifyTPSStatus())
	a := records.NewApifyTPSFromEnv()
	if !a.Available() {
		fmt.Println("NOT AVAILABLE — set APIFY_TPS_ENABLED=true for this process")
		os.Exit(1)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Minute)
	defer cancel()
	start := time.Now()
	res, err := a.Search(ctx, records.Query{
		FirstName: "Carly", LastName: "Jordan",
		City: "College Station", Region: "TX",
		Handle: "carlyyjordan",
	})
	fmt.Printf("elapsed=%s err=%v status=%s provider=%s error_field=%q cands=%d\n",
		time.Since(start).Round(time.Second), err, res.Status, res.Provider, res.Error, len(res.Candidates))
	streetN := 0
	for i, c := range res.Candidates {
		st := records.IsRealStreet(c.Line1)
		if st {
			streetN++
		}
		fmt.Printf("  [%d] kind=%s street=%q city=%s %s %s conf=%.0f%% src=%s\n",
			i, c.Kind, c.Line1, c.City, c.Region, c.Postal, c.Confidence*100, c.Source)
		if c.Note != "" {
			n := c.Note
			if len(n) > 180 {
				n = n[:180] + "…"
			}
			fmt.Printf("       note=%s\n", n)
		}
		if c.URL != "" {
			fmt.Printf("       url=%s\n", c.URL)
		}
	}
	fmt.Printf("RESULT streets=%d\n", streetN)
	if streetN == 0 {
		os.Exit(2)
	}
}
