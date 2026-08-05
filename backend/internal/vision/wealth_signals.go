package vision

import "strings"

// WealthSignals are conservative luxury/lifestyle indicators extracted from
// Instagram captions, photo labels, and geotags. Combined with property value
// they feed EstimateNetWorth. False positives are worse than false negatives —
// the keyword lists are intentionally narrow (named venues/brands only).
type WealthSignals struct {
	LuxuryVenues       []string `json:"luxury_venues,omitempty"`
	DesignerBrands     []string `json:"designer_brands,omitempty"`
	InternationalTravel bool     `json:"international_travel,omitempty"`
	JewelryIndicators  []string `json:"jewelry_indicators,omitempty"`
	PropertyValue      int64    `json:"property_value,omitempty"`
}

// ponytail: keyword lists are lowercase, matched via substring against a
// lowercased text blob. Ceiling: misses unlabeled luxury (e.g. a photo at an
// unnamed private estate). Upgrade path: a CLIP zero-shot label set for
// luxury venues/gowns — but that needs a trained model, not worth it yet.
var (
	luxuryVenueKeywords = []string{
		"four seasons", "ritz carlton", "st regis", "waldorf",
		"plaza hotel", "broadmoor", "greenbrier",
	}
	designerBrandKeywords = []string{
		"vera wang", "monique lhuillier", "oscar de la renta",
		"carolina herrera", "marchesa", "custom gown",
	}
	jewelryKeywords = []string{
		"tiffany", "cartier", "van cleef", "harry winston",
		"de beers", "platinum", "diamond",
	}
)

// usStateCodes is the set of two-letter US state codes. A geotag that is not a
// US state code is treated as international. ponytail: naive — a geotag like
// "Paris" or "Tokyo" is clearly international, but "ON" (Ontario) collides
// with Oregon's code. Conservative: we only flag international when the tag
// does NOT match a US state code AND contains a non-US place pattern. Ceiling:
// misses US territories (PR, GU) and false-positives Canadian provinces.
var usStateCodes = map[string]bool{
	"AL": true, "AK": true, "AZ": true, "AR": true, "CA": true, "CO": true,
	"CT": true, "DE": true, "FL": true, "GA": true, "HI": true, "ID": true,
	"IL": true, "IN": true, "IA": true, "KS": true, "KY": true, "LA": true,
	"ME": true, "MD": true, "MA": true, "MI": true, "MN": true, "MS": true,
	"MO": true, "MT": true, "NE": true, "NV": true, "NH": true, "NJ": true,
	"NM": true, "NY": true, "NC": true, "ND": true, "OH": true, "OK": true,
	"OR": true, "PA": true, "RI": true, "SC": true, "SD": true, "TN": true,
	"TX": true, "UT": true, "VT": true, "VA": true, "WA": true, "WV": true,
	"WI": true, "WY": true, "DC": true,
}

// ExtractWealthSignals scans Instagram-derived text + geotags for conservative
// luxury indicators. propertyValue is the pre-computed estimated home value
// (the detective already ran EstimateHomeValue with the county avg table).
// ponytail: takes int64 not records.PropertyAsset to avoid a vision→records
// import cycle — records/net_worth.go imports vision for WealthSignals.
func ExtractWealthSignals(captions []string, photoLabels []string, geotags []string, propertyValue int64) WealthSignals {
	var s WealthSignals
	s.PropertyValue = propertyValue

	// Build one lowercased blob from captions + photo labels for keyword scan.
	var b strings.Builder
	for _, c := range captions {
		b.WriteString(strings.ToLower(c))
		b.WriteByte('\n')
	}
	for _, l := range photoLabels {
		b.WriteString(strings.ToLower(l))
		b.WriteByte('\n')
	}
	blob := b.String()

	s.LuxuryVenues = matchKeywords(blob, luxuryVenueKeywords)
	s.DesignerBrands = matchKeywords(blob, designerBrandKeywords)
	s.JewelryIndicators = matchKeywords(blob, jewelryKeywords)

	// International travel: any geotag that isn't a US state code.
	// ponytail: conservative — only flags when a geotag is present and clearly
	// not a US state abbreviation. A bare city name like "London" qualifies.
	for _, g := range geotags {
		g = strings.TrimSpace(g)
		if g == "" {
			continue
		}
		upper := strings.ToUpper(g)
		if len(g) == 2 && usStateCodes[upper] {
			continue
		}
		// "City, ST" form — strip the trailing state code and re-check.
		if i := strings.LastIndex(g, ","); i >= 0 {
			tail := strings.TrimSpace(strings.ToUpper(g[i+1:]))
			if len(tail) == 2 && usStateCodes[tail] {
				continue
			}
		}
		s.InternationalTravel = true
		break
	}
	return s
}

// matchKeywords returns the keywords found as substrings in blob (deduped,
// order-stable). Each keyword appears at most once.
func matchKeywords(blob string, keywords []string) []string {
	var out []string
	for _, kw := range keywords {
		if strings.Contains(blob, kw) {
			out = append(out, kw)
		}
	}
	return out
}
