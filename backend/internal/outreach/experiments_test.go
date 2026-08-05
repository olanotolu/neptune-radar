package outreach

import "testing"

// TestAssignVariantDeterministic verifies that the same coupleID always gets
// the same variant assignment across calls, and that different couples spread
// across all variants.
func TestAssignVariantDeterministic(t *testing.T) {
	exp := DefaultExperiment()
	if len(exp.Variants) != 3 {
		t.Fatalf("expected 3 variants, got %d", len(exp.Variants))
	}

	// Same couple → same variant, every time.
	coupleID := "cpl_abc123"
	first := AssignVariant(coupleID, exp)
	for i := 0; i < 100; i++ {
		got := AssignVariant(coupleID, exp)
		if got.ID != first.ID {
			t.Fatalf("non-deterministic: couple %s got %s then %s on iter %d", coupleID, first.ID, got.ID, i)
		}
	}

	// Different couples should cover all 3 variants (deterministic hash spread).
	seen := map[string]bool{}
	for i := 0; i < 300; i++ {
		cid := "cpl_" + string(rune('a'+i%26)) + string(rune('a'+(i/26)%26)) + string(rune('a'+i%7))
		v := AssignVariant(cid, exp)
		seen[v.ID] = true
	}
	for _, v := range exp.Variants {
		if !seen[v.ID] {
			t.Errorf("variant %s never assigned across 300 couples", v.ID)
		}
	}
}
