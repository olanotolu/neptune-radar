package records

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// Heuristic is the offline/no-key provider: Ohio zip-aware candidates + research notes.
// Never invents a street number — returns verified zip+city combos.
type Heuristic struct{}

func (h *Heuristic) Name() string     { return "heuristic" }
func (h *Heuristic) Available() bool  { return true }

func (h *Heuristic) Search(ctx context.Context, q Query) (Result, error) {
	_ = ctx
	var cands []Candidate
	city, region := strings.TrimSpace(q.City), strings.ToUpper(strings.TrimSpace(q.Region))
	if city == "" {
		raw, _ := json.Marshal(map[string]any{"query": q, "reason": "no_city"})
		return Result{
			Provider: "heuristic", Status: "empty", RawJSON: string(raw),
			Candidates: nil,
		}, nil
	}

	name := strings.TrimSpace(q.FirstName + " " + q.LastName)
	if name == "" {
		name = q.FirstName
	}

	// Ohio-aware candidates: zip codes + neighborhoods
	if region == "" || region == "OH" || region == "OHIO" {
		// Level 1: Most common zip code for the city (55% confidence)
		zipCands := getOhioZipCandidates(city, region)
		cands = append(cands, zipCands...)

		// Level 2: Neighborhood-level candidates (45-55% confidence)
		hoodCands := getOhioNeighborhoodCandidates(city, region)
		cands = append(cands, hoodCands...)
	}

	// Level 3: Generic city/state candidate (lowest confidence)
	if len(cands) == 0 {
		conf := 0.30
		note := "City/region only — configure TRESTLE_API_KEY or PDL_API_KEY for street candidates."
		if q.LastName != "" {
			conf = 0.40
			note = fmt.Sprintf("Market hit for %s in %s, %s. Street requires people-data provider.", name, city, region)
		}
		cands = append(cands, Candidate{
			City: city, Region: region, Country: "US",
			Confidence: conf, Source: "market_inference",
			FullName: name, Note: note,
		})
	}

	// Partner household hint (still no street, but boosts top candidate)
	if q.PartnerFirst != "" && len(cands) > 0 {
		// Add a partner-associated candidate at the best zip
		top := cands[0]
		cands = append([]Candidate{{
			City:       top.City,
			Region:     top.Region,
			Postal:     top.Postal,
			Country:    "US",
			Confidence: top.Confidence + 0.05,
			Source:     "household_market",
			FullName:   strings.TrimSpace(q.PartnerFirst + " " + q.PartnerLast),
			Note:       "Partner also associated with " + city + " — likely shared household at zip " + top.Postal + ".",
		}}, cands...)
	}

	// Cap at 8 candidates
	if len(cands) > 8 {
		cands = cands[:8]
	}

	raw, _ := json.Marshal(map[string]any{"query": q, "mode": "ohio_heuristic", "zip_sources": len(ohioCityZips)})
	return Result{
		Provider: "heuristic", Status: "ok", Candidates: cands, RawJSON: string(raw), CostCents: 0,
	}, nil
}
