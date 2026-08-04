package records

import (
	"os"
	"strconv"
	"strings"
)

// DefaultDetectivePaidCap is the max paid people-data API calls per detective run
// when DETECTIVE_PAID_CAP is unset.
const DefaultDetectivePaidCap = 8

// DetectivePaidCap reads DETECTIVE_PAID_CAP (default 8). Values < 1 become default.
func DetectivePaidCap() int {
	raw := strings.TrimSpace(os.Getenv("DETECTIVE_PAID_CAP"))
	if raw == "" {
		return DefaultDetectivePaidCap
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 1 {
		return DefaultDetectivePaidCap
	}
	if n > 40 {
		return 40 // hard ceiling so married×location can't explode
	}
	return n
}

// LocVariant is one city/region to try in detective fan-out.
type LocVariant struct {
	City   string
	Region string
	Source string // kit | couple | account_a | account_b | vendor | post
}

// LocationVariants returns ordered unique locations from a query's signals.
// Order: primary kit/market city → account A → account B → vendor → post venue.
func LocationVariants(q Query) []LocVariant {
	var out []LocVariant
	seen := map[string]bool{}
	add := func(city, region, source string) {
		city = strings.TrimSpace(city)
		region = strings.ToUpper(strings.TrimSpace(region))
		if city == "" {
			return
		}
		key := strings.ToLower(city) + "|" + region
		if seen[key] {
			return
		}
		seen[key] = true
		out = append(out, LocVariant{City: city, Region: region, Source: source})
	}

	add(q.City, q.Region, "kit")
	add(q.AccountCityA, q.AccountRegionA, "account_a")
	add(q.AccountCityB, q.AccountRegionB, "account_b")
	add(q.VendorCity, q.VendorState, "vendor")
	if q.PostLocation != "" {
		if city, region, ok := parsePostLocationCity(q.PostLocation); ok {
			add(city, region, "post")
		}
	}
	// Title-case Ohio city keys that parsePostLocation returns lowercased
	for i := range out {
		out[i].City = titleCity(out[i].City)
	}
	return out
}

func titleCity(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}
	parts := strings.Fields(strings.ToLower(s))
	for i, p := range parts {
		if p == "" {
			continue
		}
		parts[i] = strings.ToUpper(p[:1]) + p[1:]
	}
	return strings.Join(parts, " ")
}

// NameVariant is one person identity to search.
type NameVariant struct {
	First  string
	Last   string
	Handle string
	Note   string // e.g. married-name explanation
}

// NameVariants builds primary A, primary B, then married-name variants.
// Order preserves paid budget for both partners before married swaps.
// Skips empty first names and exact duplicates.
func NameVariants(firstA, lastA, handleA, firstB, lastB, handleB string) []NameVariant {
	var out []NameVariant
	seen := map[string]bool{}
	add := func(first, last, handle, note string) {
		first = strings.TrimSpace(first)
		last = strings.TrimSpace(last)
		if first == "" {
			return
		}
		key := strings.ToLower(first) + "|" + strings.ToLower(last)
		if seen[key] {
			return
		}
		seen[key] = true
		out = append(out, NameVariant{First: first, Last: last, Handle: handle, Note: note})
	}

	// Primary partners first (budget-friendly for person B)
	add(firstA, lastA, handleA, "")
	add(firstB, lastB, handleB, "")
	// Married-name swaps last
	if lastB != "" && !strings.EqualFold(lastB, lastA) {
		add(firstA, lastB, handleA, "Married-name variant: searched as "+firstA+" "+lastB)
	}
	if lastA != "" && !strings.EqualFold(lastA, lastB) {
		add(firstB, lastA, handleB, "Married-name variant: searched as "+firstB+" "+lastA)
	}
	return out
}

// TextSource is free text that may contain a street address.
type TextSource struct {
	Text   string
	Source string // bio_a | bio_b | discovery_caption | recent_caption
}

// ExtractStreetsFromTexts parses street addresses from bios/captions (no API).
func ExtractStreetsFromTexts(texts []TextSource, fallbackCity, fallbackRegion string) []Candidate {
	var out []Candidate
	seen := map[string]bool{}
	for _, t := range texts {
		if strings.TrimSpace(t.Text) == "" {
			continue
		}
		var addr *ExtractedAddress
		switch {
		case strings.HasPrefix(t.Source, "bio"):
			addr = ExtractAddressFromBio(t.Text)
		default:
			addr = ExtractAddressFromCaption(t.Text)
		}
		if addr == nil || !IsRealStreet(addr.Line1) {
			continue
		}
		key := strings.ToLower(addr.Line1) + "|" + strings.ToLower(addr.City)
		if seen[key] {
			continue
		}
		seen[key] = true
		city := firstNonEmpty(addr.City, fallbackCity)
		region := firstNonEmpty(addr.Region, fallbackRegion)
		note := "Street parsed from " + t.Source + " — verify before mail."
		out = append(out, Candidate{
			Line1:      addr.Line1,
			Line2:      addr.Line2,
			City:       city,
			Region:     region,
			Postal:     addr.Postal,
			Country:    "US",
			Kind:       KindStreet,
			Confidence: addr.Confidence,
			Source:     t.Source + "_regex",
			Note:       note,
		})
	}
	return out
}

// MergeResults unions candidates from multiple provider Results and ranks them.
func MergeResults(parts ...Result) Result {
	var all []Candidate
	var providers []string
	var cost, paid int
	var rawParts []string
	var lastErr string
	st := "empty"
	for _, r := range parts {
		if r.Provider != "" {
			providers = append(providers, r.Provider)
		}
		all = append(all, normalizeCandidates(r.Candidates)...)
		cost += r.CostCents
		paid += r.PaidCalls
		if r.RawJSON != "" {
			rawParts = append(rawParts, r.RawJSON)
		}
		if r.Error != "" {
			lastErr = r.Error
		}
		if r.Status == "ok" || len(r.Candidates) > 0 {
			st = "ok"
		} else if r.Status == "error" && st == "empty" {
			st = "error"
		}
	}
	all = dedupeCandidates(all)
	rankCandidates(all)
	if len(all) > 0 {
		st = "ok"
	}
	name := strings.Join(providers, "+")
	if name == "" {
		name = "merged"
	}
	return Result{
		Provider:   name,
		Candidates: all,
		Status:     st,
		Error:      lastErr,
		CostCents:  cost,
		PaidCalls:  paid,
		RawJSON:    strings.Join(rawParts, "\n---\n"),
	}
}

// MaxStreetConf is exported for detective orchestration.
func MaxStreetConf(cands []Candidate) float64 {
	return maxStreetConf(cands)
}

// HasStreetCandidates is exported for detective orchestration.
func HasStreetCandidates(cands []Candidate) bool {
	return hasStreetCandidates(cands)
}
