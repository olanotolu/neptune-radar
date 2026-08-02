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

// bioStreetRe matches US street addresses in Instagram bios: "123 Main St", "123 Main St Apt 4"
var bioStreetRe = regexp.MustCompile(`(?i)\b(\d{1,5})\s+` +
	`(?:(?:N|S|E|W|NE|NW|SE|SW)\.?\s+)?` + // directional prefix
	`([A-Za-z][a-zA-Z]+(?:\s(?:St|Street|Ave|Avenue|Rd|Road|Dr|Drive|Ln|Lane|Ct|Court|Blvd|Boulevard|Way|Pl|Place|Cir|Circle|Pkwy|Parkway|Hwy|Highway|Ter|Terrace|Trail|Trl|Loop|Cove))\.?)` +
	`(?:\s+(?:Apt|Unit|Ste|#)\s*\d+[A-Za-z]?)?`,
)

var bioAptRe = regexp.MustCompile(`(?i)\b(?:Apt|Unit|Ste|#)\s*(\d+[A-Za-z]?)\b`)
var bioZipRe = regexp.MustCompile(`\b(\d{5})(?:-\d{4})?\b`)
var bioStateRe = regexp.MustCompile(`\b([A-Z]{2})\b`)

// ExtractAddressFromBio attempts to parse a street address from Instagram bio text.
// Returns nil if no street address found — never invents.
func ExtractAddressFromBio(bio string) *ExtractedAddress {
	if bio == "" {
		return nil
	}

	// Try to find a street address pattern
	idx := bioStreetRe.FindStringIndex(bio)
	if idx == nil {
		return nil
	}

	addr := &ExtractedAddress{Source: "bio_regex"}
	fullMatch := bio[idx[0]:idx[1]]

	// Parse street number + name from the match
	parts := strings.Fields(fullMatch)
	if len(parts) < 2 {
		return nil
	}
	addr.Line1 = strings.Join(parts[:2], " ")

	// If there's an apartment/unit in the match, split it out
	for i := 2; i < len(parts); i++ {
		p := strings.ToUpper(parts[i])
		if p == "APT" || p == "UNIT" || p == "STE" || p == "#" {
			addr.Line2 = strings.Join(parts[i:], " ")
			break
		}
	}

	// Search for zip + state in a window around the match
	windowStart := idx[0] - 40
	if windowStart < 0 {
		windowStart = 0
	}
	windowEnd := idx[1] + 60
	if windowEnd > len(bio) {
		windowEnd = len(bio)
	}
	window := bio[windowStart:windowEnd]

	// Find zip code
	if zm := bioZipRe.FindString(window); zm != "" {
		addr.Postal = zm
		addr.Confidence = 0.75
	} else {
		addr.Confidence = 0.60
	}

	// Find state
	if sm := bioStateRe.FindString(window); sm != "" {
		addr.Region = sm
	}

	// Find city — "City, ST" pattern
	cityRe := regexp.MustCompile(`([A-Z][a-z]+(?:\s[A-Z][a-z]+)?),\s*([A-Z]{2})`)
	if cm := cityRe.FindString(window); cm != "" {
		parts := strings.SplitN(cm, ",", 2)
		addr.City = strings.TrimSpace(parts[0])
		addr.Region = strings.TrimSpace(parts[1])
		addr.Confidence = 0.80
	}

	if addr.Line1 == "" {
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
	}
	return addr
}

// ExtractLocationFromText parses city/state from free text.
func ExtractLocationFromText(text string) (city, region string, ok bool) {
	if text == "" {
		return "", "", false
	}
	cityRe := regexp.MustCompile(`([A-Z][a-z]+(?:\s[A-Z][a-z]+)?),\s*([A-Z]{2})`)
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
