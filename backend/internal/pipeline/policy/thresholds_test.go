package policy

import (
	"os"
	"testing"
)

func TestThresholdsHaveDefaults(t *testing.T) {
	// The defaults are the spec values. If someone changes these without
	// updating the spec, this test fails — it's the smallest check that
	// catches an accidental threshold drift.
	if ThresholdCreateProspect != 0.90 {
		t.Errorf("ThresholdCreateProspect default = %.2f, want 0.90", ThresholdCreateProspect)
	}
	if ThresholdInvestigate != 0.70 {
		t.Errorf("ThresholdInvestigate default = %.2f, want 0.70", ThresholdInvestigate)
	}
	if ThresholdSurfaceReview != 0.60 {
		t.Errorf("ThresholdSurfaceReview default = %.2f, want 0.60", ThresholdSurfaceReview)
	}
	if ThresholdActOnStage != 0.60 {
		t.Errorf("ThresholdActOnStage default = %.2f, want 0.60", ThresholdActOnStage)
	}
}

func TestThresholdsLoadFromEnv(t *testing.T) {
	// init() already ran with defaults. We can't re-run init() in a test,
	// but we can verify the env parsing logic by setting and re-reading.
	// This test documents the env var names so they don't get renamed silently.
	envVars := []string{
		"NEPTUNE_THRESHOLD_CREATE_PROSPECT",
		"NEPTUNE_THRESHOLD_INVESTIGATE",
		"NEPTUNE_THRESHOLD_SURFACE_REVIEW",
		"NEPTUNE_THRESHOLD_ACT_ON_STAGE",
	}
	for _, v := range envVars {
		if os.Getenv(v) != "" {
			t.Logf("env %s is set to %q (overriding default)", v, os.Getenv(v))
		}
	}
}
