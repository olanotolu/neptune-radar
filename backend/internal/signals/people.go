package signals

import (
	"strings"
)

// Business-like handle fragments — vendors/venues/planners that photographers
// tag alongside the couple. Never treat these as partner candidates.
var businessHandleFragments = []string{
	"photo", "photos", "photography", "photog", "studio", "studios",
	"wedding", "weddings", "bridal", "boutique", "florist", "floral",
	"venue", "venues", "farm", "farms", "estate", "hall", "manor",
	"event", "events", "planner", "planning", "design", "designs",
	"decor", "décor", "rentals", "catering", "dj", "band", "music",
	"musician", "musicians", "lighting", "makeup", "beauty", "hair",
	"jeweler", "jewelry", "jewellery", "ring", "rings",
	"magazine", "publication", "registry", "hotel", "resort",
	"hilton", "marriott", "hyatt", "ritz", "westin", "sheraton",
	"co.", "llc", "inc", "official", "the.", "and.",
	"luminary", "conservatory", "athenaeum", "countryclub",
	"florist", "cake", "cakes", "bakery", "pastry", "officiant", "celebrant",
	"salon", "spa", "linens", "stationery", "paper", "tuxedo", "tux",
	"lapel", "menswear", "gown", "atelier", "couture", "shoots",
}

// LooksLikeBusinessHandle scores whether an Instagram handle is probably a
// vendor/venue/business rather than a person (couple partner).
func LooksLikeBusinessHandle(handle string) bool {
	h := strings.ToLower(strings.TrimPrefix(strings.TrimSpace(handle), "@"))
	if h == "" {
		return true
	}
	// Pure business keywords
	for _, frag := range businessHandleFragments {
		if strings.Contains(h, frag) {
			return true
		}
	}
	// Domain-like handles
	if strings.Contains(h, ".com") || strings.Contains(h, ".co") || strings.Contains(h, ".net") {
		return true
	}
	// Very long snake handles with multiple separators often brands
	if strings.Count(h, "_") >= 3 || strings.Count(h, ".") >= 2 {
		return true
	}
	return false
}

// LooksLikePersonHandle is the inverse heuristic used when ranking tags.
func LooksLikePersonHandle(handle string) bool {
	return !LooksLikeBusinessHandle(handle)
}

// ExtractCoupleNamesFromCaption is implemented in names.go (richer patterns).

// FilterPersonTags drops the source vendor + any business/vendor handles,
// returning only person-like tags (order preserved).
func FilterPersonTags(sourceHandle string, tags []string, knownVendors map[string]bool) []string {
	src := strings.ToLower(strings.TrimPrefix(sourceHandle, "@"))
	var out []string
	seen := map[string]bool{}
	for _, t := range tags {
		t = strings.ToLower(strings.TrimPrefix(strings.TrimSpace(t), "@"))
		if t == "" || t == src || seen[t] {
			continue
		}
		if knownVendors[t] {
			continue
		}
		if LooksLikeBusinessHandle(t) {
			continue
		}
		seen[t] = true
		out = append(out, t)
	}
	return out
}

// PickCouplePair chooses the best two person tags from a tag list.
// Prefers person-like handles; returns ok=false if fewer than 2 people remain.
func PickCouplePair(sourceHandle string, tags []string, knownVendors map[string]bool) (a, b string, people []string, ok bool) {
	people = FilterPersonTags(sourceHandle, tags, knownVendors)
	if len(people) < 2 {
		return "", "", people, false
	}
	a, b = people[0], people[1]
	if a > b {
		a, b = b, a
	}
	return a, b, people, true
}

// IsStyledOrAdContent detects editorial/styled shoots we should never mint couples from.
func IsStyledOrAdContent(caption string, hashtags []string) bool {
	cap := strings.ToLower(caption)
	for phrase, pen := range captionNegativePhrases {
		if strings.Contains(cap, phrase) {
			if pen == PenaltyStyledShoot || pen == PenaltyAdvertisement {
				return true
			}
		}
	}
	// Common styled-shoot phrasing
	for _, bad := range []string{
		"styled shoot", "styledshoot", "editorial", "model couple",
		"shoot across america", "inspiration shoot", "workshop",
		"#ad", "#sponsored", "sponsored by", "giveaway",
	} {
		if strings.Contains(cap, bad) {
			return true
		}
	}
	for _, h := range hashtags {
		h = strings.ToLower(strings.TrimPrefix(h, "#"))
		if pen, ok := NegativeHashtagPenalties[h]; ok {
			if pen == PenaltyStyledShoot || pen == PenaltyAdvertisement {
				return true
			}
		}
	}
	return false
}

// CoupleQualityScore ranks how likely a tag pair is a real couple (0–100).
// Used to sort scan results and hide chandelier/vendor noise.
func CoupleQualityScore(caption string, handleA, handleB string, allTags []string, hasImage bool, visualPeople bool) int {
	score := 40
	if LooksLikePersonHandle(handleA) {
		score += 15
	}
	if LooksLikePersonHandle(handleB) {
		score += 15
	}
	if LooksLikeBusinessHandle(handleA) || LooksLikeBusinessHandle(handleB) {
		score -= 40
	}
	// Caption engagement language
	cap := strings.ToLower(caption)
	for _, phrase := range ExplicitPhrases {
		if strings.Contains(cap, phrase) {
			score += 20
			break
		}
	}
	// "Alissa and Jon" style names in caption — strong couple signal
	if na, nb, ok := ExtractCoupleNamesFromCaption(caption); ok {
		score += 20
		// Boost more if names loosely match handles
		ha := strings.ToLower(handleA)
		hb := strings.ToLower(handleB)
		na, nb = strings.ToLower(na), strings.ToLower(nb)
		if strings.Contains(ha, na) || strings.Contains(hb, na) || strings.Contains(ha, nb) || strings.Contains(hb, nb) {
			score += 15
		}
	}
	// High-intent hashtags in caption
	for h := range HighIntentHashtags {
		if strings.Contains(cap, h) || strings.Contains(cap, "#"+h) {
			score += 10
			break
		}
	}
	// Vendor-heavy tag soup is a bad sign for the *pair* even if we filtered
	biz := 0
	for _, t := range allTags {
		if LooksLikeBusinessHandle(t) {
			biz++
		}
	}
	if biz >= 4 {
		score -= 10
	}
	if hasImage {
		score += 5
	}
	if visualPeople {
		score += 20
	}
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}
	return score
}
