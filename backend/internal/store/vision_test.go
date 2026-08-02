package store

import "testing"

// TestVisionClassificationStatsType confirms the stats struct exists with
// the right shape. Full DB integration requires a real Postgres.
func TestVisionClassificationStatsType(t *testing.T) {
	var s VisionClassificationStats
	s.LabelCounts = make(map[string]int)
	s.TotalCalls = 10
	s.ErrorCount = 2
	s.LabelCounts["ring"] = 5
	if s.TotalCalls != 10 || s.LabelCounts["ring"] != 5 {
		t.Error("VisionClassificationStats should be usable")
	}
}
