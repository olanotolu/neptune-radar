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

	// --- Location signal fusion: count how many independent sources agree on this city ---
	signalCount := 0
	signalNotes := []string{}
	if q.VendorCity != "" && strings.EqualFold(q.VendorCity, city) {
		signalCount++
		signalNotes = append(signalNotes, "vendor/photographer city matches")
	}
	if q.AccountCityA != "" && strings.EqualFold(q.AccountCityA, city) {
		signalCount++
		signalNotes = append(signalNotes, "person A bio city matches")
	}
	if q.AccountCityB != "" && strings.EqualFold(q.AccountCityB, city) {
		signalCount++
		signalNotes = append(signalNotes, "person B bio city matches")
	}
	if q.PostLocation != "" {
		// Parse city from venue tag like "The Joseph Hotel, Columbus OH"
		if postCity, _, ok := parsePostLocationCity(q.PostLocation); ok && strings.EqualFold(postCity, city) {
			signalCount++
			signalNotes = append(signalNotes, "post venue location matches")
		}
	}

	// Base confidence boost from signal agreement: +5% per agreeing signal (max +20%)
	signalBoost := float64(signalCount) * 0.05
	if signalBoost > 0.20 {
		signalBoost = 0.20
	}

	// Ohio-aware candidates: zip codes + neighborhoods
	if region == "" || region == "OH" || region == "OHIO" {
		// Level 1: Most common zip code for the city (55% confidence)
		zipCands := getOhioZipCandidates(city, region)
		for i := range zipCands {
			zipCands[i].Confidence += signalBoost
			if signalCount > 1 {
				zipCands[i].Note = fmt.Sprintf("%s — %d location signals agree: %s",
					zipCands[i].Note, signalCount, strings.Join(signalNotes, "; "))
			}
		}
		cands = append(cands, zipCands...)

		// Level 2: Neighborhood-level candidates (45-55% confidence)
		hoodCands := getOhioNeighborhoodCandidates(city, region)
		for i := range hoodCands {
			hoodCands[i].Confidence += signalBoost
			if signalCount > 1 {
				hoodCands[i].Note = fmt.Sprintf("%s — %d location signals agree: %s",
					hoodCands[i].Note, signalCount, strings.Join(signalNotes, "; "))
			}
		}
		cands = append(cands, hoodCands...)
	}

	// Level 3: Generic city/state candidate (lowest confidence)
	if len(cands) == 0 {
		conf := 0.30 + signalBoost
		note := "City/region only — configure TRESTLE_API_KEY or PDL_API_KEY for street candidates."
		if q.LastName != "" {
			conf = 0.40 + signalBoost
			note = fmt.Sprintf("Market hit for %s in %s, %s. Street requires people-data provider.", name, city, region)
		}
		if signalCount > 1 {
			note += fmt.Sprintf(" — %d location signals agree: %s", signalCount, strings.Join(signalNotes, "; "))
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

	raw, _ := json.Marshal(map[string]any{
		"query": q, "mode": "ohio_heuristic",
		"zip_sources":    len(ohioCityZips),
		"signal_count":   signalCount,
		"signal_notes":   signalNotes,
		"signal_boost":   signalBoost,
	})
	return Result{
		Provider: "heuristic", Status: "ok", Candidates: cands, RawJSON: string(raw), CostCents: 0,
	}, nil
}
