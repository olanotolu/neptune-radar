package records

import (
	"testing"

	"neptune-social-radar/backend/internal/vision"
)

// TestEstimateNetWorth verifies the heuristic model + confidence tiers.
func TestEstimateNetWorth(t *testing.T) {
	// High tier: property + 2 signal categories (luxury venue + jewelry).
	s := vision.WealthSignals{
		PropertyValue:      400_000,
		LuxuryVenues:       []string{"four seasons"},
		JewelryIndicators:  []string{"tiffany", "diamond"},
		InternationalTravel: true,
	}
	est, tier, bd := EstimateNetWorth(s)
	// property 400k×1.5=600k + 1 venue 100k + 2 jewelry 200k + 75k travel = 975k
	want := int64(600_000 + 100_000 + 200_000 + 75_000)
	if est != want {
		t.Errorf("estimated: want %d, got %d", want, est)
	}
	if tier != "high" {
		t.Errorf("tier: want high, got %s", tier)
	}
	if bd["property"] != 600_000 {
		t.Errorf("breakdown property: want 600000, got %d", bd["property"])
	}
	if bd["jewelry"] != 200_000 {
		t.Errorf("breakdown jewelry: want 200000, got %d", bd["jewelry"])
	}

	// Medium tier: property + 1 signal category.
	est2, tier2, _ := EstimateNetWorth(vision.WealthSignals{
		PropertyValue:  300_000,
		DesignerBrands: []string{"vera wang"},
	})
	if tier2 != "medium" {
		t.Errorf("tier2: want medium, got %s", tier2)
	}
	// 300k×1.5=450k + 50k = 500k
	if est2 != 500_000 {
		t.Errorf("estimated2: want 500000, got %d", est2)
	}

	// Low tier: property only, no signals.
	_, tier3, _ := EstimateNetWorth(vision.WealthSignals{PropertyValue: 250_000})
	if tier3 != "low" {
		t.Errorf("tier3: want low, got %s", tier3)
	}

	// Low tier: signals only, no property.
	_, tier4, _ := EstimateNetWorth(vision.WealthSignals{
		LuxuryVenues: []string{"ritz carlton"}, JewelryIndicators: []string{"cartier"},
	})
	if tier4 != "low" {
		t.Errorf("tier4 (signals only): want low, got %s", tier4)
	}

	// Empty: 0 estimate, low tier.
	est0, tier0, bd0 := EstimateNetWorth(vision.WealthSignals{})
	if est0 != 0 || tier0 != "low" || len(bd0) != 0 {
		t.Errorf("empty: want 0/low/empty, got %d/%s/%v", est0, tier0, bd0)
	}
}
