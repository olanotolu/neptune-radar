package records

import "neptune-social-radar/backend/internal/vision"

// EstimateNetWorth produces a conservative couple net-worth estimate from
// property value + Instagram luxury signals. This is an ESTIMATE, not a fact —
// callers must label it "estimated" in any operator-facing UI. Never expose on
// postcards (internal operator use only).
//
// Model (ponytail: deliberately simple heuristic — not a financial model):
//   - Property value × 1.5  (primary residence is ~60-70% of net worth, so the
//     total estate is roughly 1.5× the home value)
//   - +$100k per luxury venue signal
//   - +$50k per designer brand signal
//   - +$75k for international travel
//   - +$100k per jewelry indicator
//
// Confidence tier:
//   - "high":   property value present AND 2+ signal categories
//   - "medium": property value present AND 1 signal category
//   - "low":    property only, or signals only (no property)
//
// Ceiling: no income/liquid-asset data; a couple renting in Manhattan with a
// 7-figure income scores low here. Upgrade path: add income proxies (job title
// inference, business ownership) when available.
func EstimateNetWorth(signals vision.WealthSignals) (estimated int64, confidenceTier string, breakdown map[string]int64) {
	breakdown = map[string]int64{}

	var total int64
	if signals.PropertyValue > 0 {
		prop := signals.PropertyValue * 3 / 2 // × 1.5
		breakdown["property"] = prop
		total += prop
	}
	if n := len(signals.LuxuryVenues); n > 0 {
		v := int64(n) * 100_000
		breakdown["luxury_venues"] = v
		total += v
	}
	if n := len(signals.DesignerBrands); n > 0 {
		v := int64(n) * 50_000
		breakdown["designer_brands"] = v
		total += v
	}
	if signals.InternationalTravel {
		breakdown["international_travel"] = 75_000
		total += 75_000
	}
	if n := len(signals.JewelryIndicators); n > 0 {
		v := int64(n) * 100_000
		breakdown["jewelry"] = v
		total += v
	}

	// Count non-empty signal categories for tiering.
	categories := 0
	if len(signals.LuxuryVenues) > 0 {
		categories++
	}
	if len(signals.DesignerBrands) > 0 {
		categories++
	}
	if signals.InternationalTravel {
		categories++
	}
	if len(signals.JewelryIndicators) > 0 {
		categories++
	}

	hasProperty := signals.PropertyValue > 0
	switch {
	case hasProperty && categories >= 2:
		confidenceTier = "high"
	case hasProperty && categories == 1:
		confidenceTier = "medium"
	default:
		confidenceTier = "low"
	}
	return total, confidenceTier, breakdown
}
