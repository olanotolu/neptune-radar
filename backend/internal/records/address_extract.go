package records

import (
	"regexp"
	"strings"
)

// ExtractedAddress is a street-level address parsed from free text (bio, caption, post).
type ExtractedAddress struct {
	Line1      string  `json:"line1"`
	Line2      string  `json:"line2,omitempty"`
	City       string  `json:"city,omitempty"`
	Region     string  `json:"region,omitempty"`
	Postal     string  `json:"postal,omitempty"`
	Confidence float64 `json:"confidence"`
	Source     string  `json:"source"`
}

// StreetTypePattern is shared by extractors and IsRealStreet — keep in sync.
// Longer tokens first so "Street" is not matched as "St" + leftover "reet".
const StreetTypePattern = `(?:Street|St|Avenue|Ave|Road|Rd|Drive|Dr|Lane|Ln|Court|Ct|Boulevard|Blvd|Parkway|Pkwy|Highway|Hwy|Terrace|Ter|Trail|Trl|Circle|Cir|Place|Pl|Loop|Cove|Way|Park|Pass|Run|Path|Row|Square|Sq)`

// bioStreetRe matches US street addresses: "123 Main St", "456 N High Street Apt 4"
var bioStreetRe = regexp.MustCompile(`(?i)\b(\d{1,6})\s+` +
	`(?:(?:N|S|E|W|NE|NW|SE|SW)\.?\s+)?` +
	`([A-Za-z][A-Za-z0-9.'\-]*(?:\s+[A-Za-z][A-Za-z0-9.'\-]*){0,4}\s+` + StreetTypePattern + `\.?)` +
	`(?:\s*,?\s*(?:Apt|Unit|Ste|Suite|#)\s*\.?\s*[A-Za-z0-9\-]+)?`,
)

var bioAptRe = regexp.MustCompile(`(?i)\b(?:Apt|Apartment|Unit|Ste|Suite|#)\s*\.?\s*([A-Za-z0-9\-]+)\b`)
var bioZipRe = regexp.MustCompile(`\b(\d{5})(?:-\d{4})?\b`)
var bioStateRe = regexp.MustCompile(`\b([A-Z]{2})\b`)
var streetTypeRe = regexp.MustCompile(`(?i)\b` + StreetTypePattern + `\b\.?`)

// ExtractAddressFromBio attempts to parse a street address from Instagram bio text.
// Returns nil if no street address found — never invents.
func ExtractAddressFromBio(bio string) *ExtractedAddress {
	if bio == "" {
		return nil
	}

	idx := bioStreetRe.FindStringIndex(bio)
	if idx == nil {
		return nil
	}

	addr := &ExtractedAddress{Source: "bio_regex"}
	fullMatch := strings.TrimSpace(bio[idx[0]:idx[1]])
	if fullMatch == "" {
		return nil
	}

	// Split unit from street — keep full street name (not first two tokens).
	parts := strings.Fields(fullMatch)
	unitAt := -1
	for i, p := range parts {
		up := strings.ToUpper(strings.Trim(p, ".,"))
		if up == "APT" || up == "APARTMENT" || up == "UNIT" || up == "STE" || up == "SUITE" || up == "#" {
			unitAt = i
			break
		}
		if strings.HasPrefix(up, "#") && len(up) > 1 {
			unitAt = i
			break
		}
	}
	if unitAt > 0 {
		addr.Line1 = strings.Join(parts[:unitAt], " ")
		addr.Line2 = strings.Join(parts[unitAt:], " ")
	} else {
		addr.Line1 = fullMatch
	}
	// Clean trailing comma on Line1
	addr.Line1 = strings.Trim(addr.Line1, " ,")

	// Window for city/state/zip
	windowStart := idx[0] - 40
	if windowStart < 0 {
		windowStart = 0
	}
	windowEnd := idx[1] + 80
	if windowEnd > len(bio) {
		windowEnd = len(bio)
	}
	window := bio[windowStart:windowEnd]

	if zm := bioZipRe.FindString(window); zm != "" {
		addr.Postal = zm
		addr.Confidence = 0.75
	} else {
		addr.Confidence = 0.60
	}

	// Prefer "City, ST" over bare ST
	cityRe := regexp.MustCompile(`([A-Z][a-z]+(?:\s[A-Z][a-z]+)?),\s*([A-Z]{2})\b`)
	if m := cityRe.FindStringSubmatch(window); len(m) >= 3 {
		addr.City = strings.TrimSpace(m[1])
		addr.Region = strings.TrimSpace(m[2])
		addr.Confidence = 0.80
	} else if sm := bioStateRe.FindString(window); sm != "" {
		// Avoid matching street directionals as state when alone — only if 2-letter US state list
		if isUSState(sm) {
			addr.Region = sm
		}
	}

	// Apt outside the street match
	if addr.Line2 == "" {
		if m := bioAptRe.FindStringSubmatch(window); len(m) > 0 {
			addr.Line2 = strings.TrimSpace(m[0])
		}
	}

	if !IsRealStreet(addr.Line1) {
		return nil
	}
	return addr
}

// ExtractAddressFromCaption parses an address from a post caption (noisier, lower confidence).
func ExtractAddressFromCaption(caption string) *ExtractedAddress {
	addr := ExtractAddressFromBio(caption)
	if addr != nil {
		addr.Source = "caption_regex"
		addr.Confidence *= 0.85
		lower := strings.ToLower(caption)
		// Venue cues → not necessarily residence
		for _, cue := range []string{"hotel", "venue", "said yes at", "at the", "reception at", "ceremony at"} {
			if strings.Contains(lower, cue) {
				addr.Confidence *= 0.85
				break
			}
		}
	}
	return addr
}

// ExtractLocationFromText parses city/state from free text.
func ExtractLocationFromText(text string) (city, region string, ok bool) {
	if text == "" {
		return "", "", false
	}
	cityRe := regexp.MustCompile(`([A-Z][a-z]+(?:\s[A-Z][a-z]+)?),\s*([A-Z]{2})\b`)
	if m := cityRe.FindStringSubmatch(text); len(m) >= 3 {
		return m[1], m[2], true
	}
	lower := strings.ToLower(text)
	for _, e := range []struct{ needle, city, region string }{
		{"columbus", "Columbus", "OH"},
		{"cleveland", "Cleveland", "OH"},
		{"cincinnati", "Cincinnati", "OH"},
		{"dayton", "Dayton", "OH"},
		{"akron", "Akron", "OH"},
		{"toledo", "Toledo", "OH"},
	} {
		if strings.Contains(lower, e.needle) {
			return e.city, e.region, true
		}
	}
	return "", "", false
}

var usStates = map[string]bool{
	"AL": true, "AK": true, "AZ": true, "AR": true, "CA": true, "CO": true, "CT": true,
	"DE": true, "FL": true, "GA": true, "HI": true, "ID": true, "IL": true, "IN": true,
	"IA": true, "KS": true, "KY": true, "LA": true, "ME": true, "MD": true, "MA": true,
	"MI": true, "MN": true, "MS": true, "MO": true, "MT": true, "NE": true, "NV": true,
	"NH": true, "NJ": true, "NM": true, "NY": true, "NC": true, "ND": true, "OH": true,
	"OK": true, "OR": true, "PA": true, "RI": true, "SC": true, "SD": true, "TN": true,
	"TX": true, "UT": true, "VT": true, "VA": true, "WA": true, "WV": true, "WI": true,
	"WY": true, "DC": true,
}

func isUSState(s string) bool {
	return usStates[strings.ToUpper(strings.TrimSpace(s))]
}

// HasStreetType reports whether s contains a recognized US street suffix.
func HasStreetType(s string) bool {
	return streetTypeRe.MatchString(s)
}
