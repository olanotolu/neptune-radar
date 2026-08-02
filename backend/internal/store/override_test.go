package store

import "testing"

// TestRejectHypothesisExists confirms the override method compiles and is
// callable. Full DB integration requires a real Postgres.
func TestRejectHypothesisExists(t *testing.T) {
	// Just verify the method signature exists — the smallest thing that
	// fails if someone removes it.
	var s *Store
	_ = func(id, reason, decidedBy string) error {
		return s.RejectHypothesis(id, reason, decidedBy)
	}
}

func TestMarkCoupleMistakenExists(t *testing.T) {
	var s *Store
	_ = func(coupleID, reason, decidedBy string) error {
		return s.MarkCoupleMistaken(coupleID, reason, decidedBy)
	}
}
