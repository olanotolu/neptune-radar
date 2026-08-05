package outreach

import (
	"fmt"
	"strings"

	"neptune-social-radar/backend/internal/store"
)

// PrepResult is the Prep agent output: whether detective should spend budget.
// Models never invent streets here — only readiness + home-market inference.
type PrepResult struct {
	Score      float64  `json:"score"` // 0–1 detective_ready_score
	Ready      bool     `json:"ready"` // score >= ReadyThreshold
	Blockers   []string `json:"blockers,omitempty"`
	Warnings   []string `json:"warnings,omitempty"`
	HomeCity   string   `json:"home_city,omitempty"`
	HomeRegion string   `json:"home_region,omitempty"`
	HomeSource string   `json:"home_source,omitempty"` // bio_a | bio_b | bio_agree | kit | vendor_only | none
	// Role flags for UI
	VendorConfused bool `json:"vendor_confused,omitempty"`
	HasLastA       bool `json:"has_last_a"`
	HasLastB       bool `json:"has_last_b"`
	HasCity        bool `json:"has_city"`
	// FenrisCrossValidated is true when a Fenris Digital life event
	// independently confirms this couple — two independent signals = +0.15.
	FenrisCrossValidated bool `json:"fenris_cross_validated,omitempty"`
	Summary              string `json:"summary"`
}

// ReadyThreshold — auto-detective and aggressive hunters require this.
const ReadyThreshold = 0.70

// MinRunThreshold — manual Run detective allowed at this floor (with warnings).
// Below this, RunDetective refuses (missing both lasts or both identity broken).
const MinRunThreshold = 0.35

// RunPrep evaluates kit identity + market for address hunting.
// Pure function of kit fields (and optional vendor class) — no I/O.
func RunPrep(k store.CongratulateKit) PrepResult {
	firstA := strings.TrimSpace(firstNonEmpty(k.FirstNameA, splitNameFirst(k.PersonAName)))
	lastA := strings.TrimSpace(firstNonEmpty(k.LastNameA, splitNameLast(k.PersonAName)))
	firstB := strings.TrimSpace(firstNonEmpty(k.FirstNameB, splitNameFirst(k.PersonBName)))
	lastB := strings.TrimSpace(firstNonEmpty(k.LastNameB, splitNameLast(k.PersonBName)))

	p := PrepResult{
		HasLastA: lastA != "",
		HasLastB: lastB != "",
	}
	// Fenris cross-validation: two independent signals (Instagram + Fenris
	// life event) = higher confidence. +0.15 boost to the prep score.
	if k.FenrisValidated {
		p.FenrisCrossValidated = true
	}

	// --- Blockers: identity ---
	if firstA == "" && firstB == "" {
		p.Blockers = append(p.Blockers, "missing_first_names")
	}
	if !p.HasLastA && !p.HasLastB {
		p.Blockers = append(p.Blockers, "missing_last_names")
	} else if !p.HasLastA {
		p.Warnings = append(p.Warnings, "missing_last_a — search weaker for person A")
	} else if !p.HasLastB {
		p.Warnings = append(p.Warnings, "missing_last_b — search weaker for person B")
	}

	// Vendor / brand confused with person
	if looksLikeBrand(firstA, k.HandleA, k.BioA) || looksLikeBrand(k.PersonAName, k.HandleA, k.BioA) {
		p.Blockers = append(p.Blockers, "person_a_looks_like_brand_or_venue")
		p.VendorConfused = true
	}
	if looksLikeBrand(firstB, k.HandleB, k.BioB) || looksLikeBrand(k.PersonBName, k.HandleB, k.BioB) {
		p.Warnings = append(p.Warnings, "person_b_may_be_brand_or_venue")
		p.VendorConfused = true
	}
	// Source class is vendor → market is shoot location not home
	if sc := strings.ToLower(k.SourceClass); sc != "" {
		switch {
		case strings.Contains(sc, "photo"), strings.Contains(sc, "venue"),
			strings.Contains(sc, "florist"), strings.Contains(sc, "planner"),
			strings.Contains(sc, "vendor"):
			p.Warnings = append(p.Warnings, "discovery_source_is_vendor — city may be shoot market not home")
		}
	}

	// Same display name both sides (Manolo Greco case)
	if firstA != "" && strings.EqualFold(strings.TrimSpace(firstA+" "+lastA), strings.TrimSpace(firstB+" "+lastB)) &&
		firstA != "" {
		p.Warnings = append(p.Warnings, "identical_names_both_sides — possible vendor self-pair")
		p.VendorConfused = true
	}

	// --- Home market inference (prefer bios over vendor) ---
	homeCity, homeRegion, homeSrc := inferHomeMarket(k)
	p.HomeCity, p.HomeRegion, p.HomeSource = homeCity, homeRegion, homeSrc
	kitCityRaw := strings.TrimSpace(firstNonEmpty(k.AddressCity, k.MarketCity))
	p.HasCity = homeCity != "" || (kitCityRaw != "" && !isGarbageCity(kitCityRaw))

	if !p.HasCity {
		p.Blockers = append(p.Blockers, "missing_city")
	} else if homeSrc == "vendor_only" {
		p.Warnings = append(p.Warnings, "city_from_vendor_only — confirm home market before mail")
	}

	// School/bio hints without structured city (e.g. Texas A&M)
	if !p.HasCity {
		if hint := schoolMarketHint(k.BioA + " " + k.BioB); hint != "" {
			p.Warnings = append(p.Warnings, "market_hint:"+hint+" — set city manually then re-run prep")
		}
	}

	// --- Score ---
	score := 0.0
	// Last names (0.40)
	if p.HasLastA && p.HasLastB {
		score += 0.40
	} else if p.HasLastA || p.HasLastB {
		score += 0.22
	}
	// First names present (0.10)
	if firstA != "" && firstB != "" {
		score += 0.10
	} else if firstA != "" || firstB != "" {
		score += 0.05
	}
	// City (0.30)
	switch homeSrc {
	case "bio_agree":
		score += 0.30
	case "bio_a", "bio_b", "kit":
		score += 0.22
	case "vendor_only":
		score += 0.10
	default:
		if p.HasCity {
			score += 0.15
		}
	}
	// Not vendor-confused (0.15)
	if !p.VendorConfused {
		score += 0.15
	} else {
		score += 0.03
	}
	// Handles present (0.05)
	if k.HandleA != "" && k.HandleB != "" {
		score += 0.05
	}
	// Fenris cross-validation (+0.15) — two independent signals per couple.
	if p.FenrisCrossValidated {
		score += 0.15
	}

	// Hard cap if critical blockers
	for _, b := range p.Blockers {
		if b == "missing_last_names" || b == "missing_first_names" {
			if score > 0.34 {
				score = 0.34
			}
		}
		if b == "person_a_looks_like_brand_or_venue" {
			if score > 0.40 {
				score = 0.40
			}
		}
	}

	if score > 1 {
		score = 1
	}
	p.Score = score
	p.Ready = score >= ReadyThreshold && len(hardBlockers(p.Blockers)) == 0
	p.Summary = prepSummary(p, firstA, lastA, firstB, lastB)
	return p
}

func hardBlockers(blockers []string) []string {
	var out []string
	for _, b := range blockers {
		switch b {
		case "missing_last_names", "missing_first_names", "person_a_looks_like_brand_or_venue":
			out = append(out, b)
		}
	}
	return out
}

func inferHomeMarket(k store.CongratulateKit) (city, region, source string) {
	// Prefer address/market already on kit if not garbage
	kitCity := strings.TrimSpace(firstNonEmpty(k.AddressCity, k.MarketCity))
	kitRegion := strings.TrimSpace(firstNonEmpty(k.AddressRegion, k.MarketRegion))
	if isGarbageCity(kitCity) {
		kitCity, kitRegion = "", ""
	}

	cityA, regionA, okA := locationFromBio(k.BioA)
	cityB, regionB, okB := locationFromBio(k.BioB)

	if okA && okB && strings.EqualFold(cityA, cityB) {
		return titleCaseCity(cityA), preferRegion(regionA, regionB, kitRegion), "bio_agree"
	}
	if okA {
		return titleCaseCity(cityA), preferRegion(regionA, kitRegion, ""), "bio_a"
	}
	if okB {
		return titleCaseCity(cityB), preferRegion(regionB, kitRegion, ""), "bio_b"
	}
	if kitCity != "" {
		src := firstNonEmpty(k.MarketSource, "kit")
		if strings.Contains(strings.ToLower(src), "vendor") || strings.Contains(strings.ToLower(src), "photo") {
			return kitCity, kitRegion, "vendor_only"
		}
		return kitCity, kitRegion, "kit"
	}
	// Vendor market only on kit
	if k.SourceHandle != "" && kitCity == "" {
		return "", "", "none"
	}
	return "", "", "none"
}

func locationFromBio(bio string) (city, region string, ok bool) {
	bio = strings.TrimSpace(bio)
	if bio == "" {
		return "", "", false
	}
	// Reuse records-style patterns via light local parse
	lower := strings.ToLower(bio)
	// "Based in X" / "📍 City"
	for _, needle := range []struct{ n, city, region string }{
		{"college station", "College Station", "TX"},
		{"texas a&m", "College Station", "TX"},
		{"texas am", "College Station", "TX"},
		{"houston", "Houston", "TX"},
		{"dallas", "Dallas", "TX"},
		{"austin", "Austin", "TX"},
		{"san antonio", "San Antonio", "TX"},
		{"columbus", "Columbus", "OH"},
		{"cleveland", "Cleveland", "OH"},
		{"cincinnati", "Cincinnati", "OH"},
		{"new york", "New York", "NY"},
		{"nyc", "New York", "NY"},
		{"brooklyn", "Brooklyn", "NY"},
		{"chicago", "Chicago", "IL"},
		{"los angeles", "Los Angeles", "CA"},
		{"nashville", "Nashville", "TN"},
		{"atlanta", "Atlanta", "GA"},
		{"miami", "Miami", "FL"},
		{"denver", "Denver", "CO"},
		{"seattle", "Seattle", "WA"},
		{"based in italy", "", ""}, // not US home for mail
		{"italy", "", ""},
	} {
		if strings.Contains(lower, needle.n) {
			if needle.city == "" {
				return "", "", false
			}
			return needle.city, needle.region, true
		}
	}
	return "", "", false
}

func schoolMarketHint(blob string) string {
	lower := strings.ToLower(blob)
	if strings.Contains(lower, "texas a&m") || strings.Contains(lower, "a&m") {
		return "College Station, TX (Texas A&M)"
	}
	if strings.Contains(lower, "ohio state") {
		return "Columbus, OH (Ohio State)"
	}
	return ""
}

func isGarbageCity(city string) bool {
	c := strings.ToLower(strings.TrimSpace(city))
	if c == "" {
		return true
	}
	// Caption fragments that leaked into market_city
	garbage := []string{"the moment", "forever story", "the kids", "she said", "just engaged",
		"glowing", "illuminating", "so ", "we "}
	for _, g := range garbage {
		if c == strings.TrimSpace(g) || strings.HasPrefix(c, g) {
			return true
		}
	}
	// Too short or all lowercase single dictionary-ish words without state context
	if len(c) < 3 {
		return true
	}
	return false
}

func looksLikeBrand(name, handle, bio string) bool {
	n := strings.ToLower(strings.TrimSpace(name))
	h := strings.ToLower(strings.TrimSpace(handle))
	b := strings.ToLower(bio)
	if n == "" && h == "" {
		return false
	}
	brandBits := []string{"photo", "films", "studio", "events", "library", "museum", "hotel",
		"venue", "picnic", "restaurant", "catering", "florist", "bridal", "co.", "llc",
		"productions", "production", "official", "the wooly", "morgan"}
	for _, bit := range brandBits {
		if strings.Contains(n, bit) || strings.Contains(h, bit) {
			return true
		}
	}
	// Bio is a venue/restaurant
	for _, bit := range []string{"reservations", "event requests", "catering", "book now", "dm for rates"} {
		if strings.Contains(b, bit) {
			return true
		}
	}
	return false
}

func preferRegion(a, b, c string) string {
	for _, r := range []string{a, b, c} {
		r = strings.ToUpper(strings.TrimSpace(r))
		if len(r) == 2 {
			return r
		}
	}
	return strings.TrimSpace(firstNonEmpty(a, b, c))
}

func titleCaseCity(s string) string {
	parts := strings.Fields(strings.ToLower(strings.TrimSpace(s)))
	for i, p := range parts {
		if p == "" {
			continue
		}
		parts[i] = strings.ToUpper(p[:1]) + p[1:]
	}
	return strings.Join(parts, " ")
}

func prepSummary(p PrepResult, firstA, lastA, firstB, lastB string) string {
	status := "BLOCKED"
	if p.Ready {
		status = "READY"
	} else if p.Score >= MinRunThreshold {
		status = "WEAK"
	}
	return fmt.Sprintf("Prep %s (%.0f%%) · %s %s & %s %s · home=%s, %s (%s) · blockers=%v",
		status, p.Score*100,
		firstA, lastA, firstB, lastB,
		p.HomeCity, p.HomeRegion, p.HomeSource, p.Blockers)
}

// ApplyPrepToKit writes prep into mail_payload and soft-fills city when empty/garbage.
func ApplyPrepToKit(k *store.CongratulateKit, p PrepResult) {
	if k.MailPayload == nil {
		k.MailPayload = map[string]any{}
	}
	k.MailPayload["detective_prep"] = map[string]any{
		"score":           p.Score,
		"ready":           p.Ready,
		"blockers":        p.Blockers,
		"warnings":        p.Warnings,
		"home_city":       p.HomeCity,
		"home_region":     p.HomeRegion,
		"home_source":     p.HomeSource,
		"vendor_confused": p.VendorConfused,
		"has_last_a":      p.HasLastA,
		"has_last_b":      p.HasLastB,
		"has_city":        p.HasCity,
		"summary":         p.Summary,
	}
	// Soft-fill market/address city when missing or garbage
	if p.HomeCity != "" {
		if isGarbageCity(k.MarketCity) || k.MarketCity == "" {
			k.MarketCity = p.HomeCity
			k.MarketRegion = p.HomeRegion
			k.MarketSource = "prep:" + p.HomeSource
		}
		if isGarbageCity(k.AddressCity) || k.AddressCity == "" {
			k.AddressCity = p.HomeCity
			k.AddressRegion = p.HomeRegion
		}
	}
	// Research note line
	line := "\n\n--- Prep agent ---\n" + p.Summary
	if len(p.Warnings) > 0 {
		line += "\nWarnings: " + strings.Join(p.Warnings, "; ")
	}
	if !strings.Contains(k.ResearchNotes, "--- Prep agent ---") {
		k.ResearchNotes = strings.TrimSpace(k.ResearchNotes + line)
	} else {
		// Replace last prep block simply by appending fresh run
		k.ResearchNotes = strings.TrimSpace(k.ResearchNotes + line)
	}
}

// PrepFromKit reads stored prep if present.
func PrepFromKit(k store.CongratulateKit) (PrepResult, bool) {
	if k.MailPayload == nil {
		return PrepResult{}, false
	}
	raw, ok := k.MailPayload["detective_prep"].(map[string]any)
	if !ok {
		return PrepResult{}, false
	}
	p := PrepResult{
		Score:          floatFrom(raw["score"]),
		Ready:          boolFrom(raw["ready"]),
		HomeCity:       strFrom(raw["home_city"]),
		HomeRegion:     strFrom(raw["home_region"]),
		HomeSource:     strFrom(raw["home_source"]),
		VendorConfused: boolFrom(raw["vendor_confused"]),
		HasLastA:       boolFrom(raw["has_last_a"]),
		HasLastB:       boolFrom(raw["has_last_b"]),
		HasCity:        boolFrom(raw["has_city"]),
		Summary:        strFrom(raw["summary"]),
	}
	p.Blockers = stringSliceFrom(raw["blockers"])
	p.Warnings = stringSliceFrom(raw["warnings"])
	return p, true
}

func floatFrom(v any) float64 {
	switch t := v.(type) {
	case float64:
		return t
	case float32:
		return float64(t)
	case int:
		return float64(t)
	default:
		return 0
	}
}

func boolFrom(v any) bool {
	b, _ := v.(bool)
	return b
}

func strFrom(v any) string {
	s, _ := v.(string)
	return s
}

func stringSliceFrom(v any) []string {
	arr, ok := v.([]any)
	if !ok {
		// JSON round-trip may use []string
		if ss, ok := v.([]string); ok {
			return ss
		}
		return nil
	}
	var out []string
	for _, x := range arr {
		if s, ok := x.(string); ok {
			out = append(out, s)
		}
	}
	return out
}
