package records

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// Heuristic is the offline/no-key provider: city-level candidates + research notes.
// Never invents a street number.
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

	// Partner household hint (still no street)
	if q.PartnerFirst != "" {
		cands = append(cands, Candidate{
			City: city, Region: region, Country: "US",
			Confidence: conf + 0.05,
			Source:     "household_market",
			FullName:   strings.TrimSpace(q.PartnerFirst + " " + q.PartnerLast),
			Note:       "Partner also associated with same market — likely shared household once street is known.",
		})
	}

	raw, _ := json.Marshal(map[string]any{"query": q, "mode": "heuristic"})
	return Result{
		Provider: "heuristic", Status: "ok", Candidates: cands, RawJSON: string(raw), CostCents: 0,
	}, nil
}
