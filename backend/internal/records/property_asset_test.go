package records

import "testing"

// TestParsePropertyAsset_FranklinCountyOH verifies the regex parser extracts
// financial fields from a sample county auditor HTML snippet.
func TestParsePropertyAsset_FranklinCountyOH(t *testing.T) {
	html := `
<html><body>
<div class="property-detail">
  <h2>Property Search Results</h2>
  <table class="property-info">
    <tr><td>Owner:</td><td>Smith John D</td></tr>
    <tr><td>Address:</td><td>123 Main St Columbus OH 43215</td></tr>
    <tr><td>Assessed Value:</td><td>$285,000.00</td></tr>
    <tr><td>Sq Ft:</td><td>2,150</td></tr>
    <tr><td>Year Built:</td><td>1998</td></tr>
    <tr><td>Lot Size:</td><td>0.25 acres</td></tr>
    <tr><td>Annual Tax:</td><td>$4,200.00</td></tr>
  </table>
</div>
</body></html>`

	a := parsePropertyAsset(html)
	if a.AssessedValue != 285000 {
		t.Errorf("AssessedValue: want 285000, got %d", a.AssessedValue)
	}
	if a.Sqft != 2150 {
		t.Errorf("Sqft: want 2150, got %d", a.Sqft)
	}
	if a.YearBuilt != 1998 {
		t.Errorf("YearBuilt: want 1998, got %d", a.YearBuilt)
	}
	if a.LotSize != 0.25 {
		t.Errorf("LotSize: want 0.25, got %f", a.LotSize)
	}
	if a.TaxAnnual != 4200 {
		t.Errorf("TaxAnnual: want 4200, got %d", a.TaxAnnual)
	}
}

// TestParsePropertyAsset_GarbageHTML verifies graceful failure — no crash, all zero.
func TestParsePropertyAsset_GarbageHTML(t *testing.T) {
	a := parsePropertyAsset("<html><body>JS-rendered portal, no data</body></html>")
	if a.AssessedValue != 0 || a.Sqft != 0 || a.YearBuilt != 0 {
		t.Errorf("expected all zero on garbage HTML, got %+v", a)
	}
}

// TestEstimateHomeValue verifies the estimation model.
func TestEstimateHomeValue(t *testing.T) {
	// sqft-based estimate
	a := PropertyAsset{Sqft: 2000, AssessedValue: 250000}
	est := EstimateHomeValue(a, 185) // Franklin County OH
	want := int64(2000 * 185) // 370000 > 250000
	if est != want {
		t.Errorf("EstimateHomeValue: want %d, got %d", want, est)
	}

	// assessed value wins when higher
	a2 := PropertyAsset{Sqft: 1000, AssessedValue: 500000}
	est2 := EstimateHomeValue(a2, 185)
	if est2 != 500000 {
		t.Errorf("assessed should win: want 500000, got %d", est2)
	}

	// no data → 0
	est3 := EstimateHomeValue(PropertyAsset{}, 185)
	if est3 != 0 {
		t.Errorf("empty asset should be 0, got %d", est3)
	}
}

// TestCountyAvgPricePerSqft verifies the lookup table.
func TestCountyAvgPricePerSqft(t *testing.T) {
	if v := CountyAvgPricePerSqft("Franklin", "OH"); v != 185 {
		t.Errorf("Franklin OH: want 185, got %f", v)
	}
	if v := CountyAvgPricePerSqft("Unknown", "OH"); v != 160 {
		t.Errorf("Unknown OH fallback: want 160, got %f", v)
	}
	if v := CountyAvgPricePerSqft("Unknown", "XX"); v != 0 {
		t.Errorf("Unknown state: want 0, got %f", v)
	}
}
