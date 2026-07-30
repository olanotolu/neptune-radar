package signals

import (
	"regexp"
	"strings"
)

// LocationGuess is a lightweight geo inference from free-text bios / post
// location strings. Deterministic only — no geocoder API in v1.
type LocationGuess struct {
	City   string
	Region string // state abbreviation or full region when known
	Source string // "bio" | "post" | "caption"
}

// Common US city/region patterns that show up in engagement-couple bios.
// Order matters: longer / more specific patterns first.
var cityPatterns = []struct {
	re     *regexp.Regexp
	city   string
	region string
}{
	{regexp.MustCompile(`(?i)\bcolumbus,?\s*oh(?:io)?\b`), "Columbus", "OH"},
	{regexp.MustCompile(`(?i)\bcleveland,?\s*oh(?:io)?\b`), "Cleveland", "OH"},
	{regexp.MustCompile(`(?i)\bcincinnati,?\s*oh(?:io)?\b`), "Cincinnati", "OH"},
	{regexp.MustCompile(`(?i)\bdublin,?\s*oh(?:io)?\b`), "Dublin", "OH"},
	{regexp.MustCompile(`(?i)\bworthington,?\s*oh(?:io)?\b`), "Worthington", "OH"},
	{regexp.MustCompile(`(?i)\bwesterville,?\s*oh(?:io)?\b`), "Westerville", "OH"},
	{regexp.MustCompile(`(?i)\bbrooklyn,?\s*(?:ny|new\s*york)?\b`), "Brooklyn", "NY"},
	{regexp.MustCompile(`(?i)\bmanhattan,?\s*(?:ny|new\s*york)?\b`), "Manhattan", "NY"},
	{regexp.MustCompile(`(?i)\bnew\s*york\s*city\b|\bnyc\b`), "New York", "NY"},
	{regexp.MustCompile(`(?i)\blos\s*angeles\b`), "Los Angeles", "CA"},
	{regexp.MustCompile(`(?i)(?:^|[\s,|•·])la(?:[\s,|•·]|$)`), "Los Angeles", "CA"},
	{regexp.MustCompile(`(?i)\bchicago,?\s*il(?:linois)?\b`), "Chicago", "IL"},
	{regexp.MustCompile(`(?i)\bmiami,?\s*fl(?:orida)?\b`), "Miami", "FL"},
	{regexp.MustCompile(`(?i)\baustin,?\s*tx(?:as)?\b`), "Austin", "TX"},
	{regexp.MustCompile(`(?i)\bdallas,?\s*tx(?:as)?\b`), "Dallas", "TX"},
	{regexp.MustCompile(`(?i)\bhouston,?\s*tx(?:as)?\b`), "Houston", "TX"},
	{regexp.MustCompile(`(?i)\bseattle,?\s*wa(?:shington)?\b`), "Seattle", "WA"},
	{regexp.MustCompile(`(?i)\bboston,?\s*ma(?:ssachusetts)?\b`), "Boston", "MA"},
	{regexp.MustCompile(`(?i)\bphiladelphia,?\s*pa(?:ennsylvania)?\b`), "Philadelphia", "PA"},
	{regexp.MustCompile(`(?i)\bdenver,?\s*co(?:lorado)?\b`), "Denver", "CO"},
	{regexp.MustCompile(`(?i)\batlanta,?\s*ga(?:orgia)?\b`), "Atlanta", "GA"},
	{regexp.MustCompile(`(?i)\bcentral\s*park\b`), "New York", "NY"},
	// Generic "City, ST" fallback
	{regexp.MustCompile(`(?i)\b([A-Z][a-z]+(?:\s[A-Z][a-z]+)?),\s*([A-Z]{2})\b`), "", ""},
}

// InferLocationFromText pulls a city/region guess from bio or location text.
func InferLocationFromText(text, source string) (LocationGuess, bool) {
	text = strings.TrimSpace(text)
	if text == "" {
		return LocationGuess{}, false
	}
	for _, p := range cityPatterns {
		m := p.re.FindStringSubmatch(text)
		if m == nil {
			continue
		}
		if p.city != "" {
			return LocationGuess{City: p.city, Region: p.region, Source: source}, true
		}
		// Generic City, ST capture
		if len(m) >= 3 {
			return LocationGuess{City: m[1], Region: strings.ToUpper(m[2]), Source: source}, true
		}
	}
	return LocationGuess{}, false
}

// BestLocation prefers post geotag over bio, then caption.
func BestLocation(postLocation, bioA, bioB, caption string) (LocationGuess, bool) {
	if g, ok := InferLocationFromText(postLocation, "post"); ok {
		return g, true
	}
	if g, ok := InferLocationFromText(bioA, "bio"); ok {
		return g, true
	}
	if g, ok := InferLocationFromText(bioB, "bio"); ok {
		return g, true
	}
	if g, ok := InferLocationFromText(caption, "caption"); ok {
		return g, true
	}
	return LocationGuess{}, false
}
