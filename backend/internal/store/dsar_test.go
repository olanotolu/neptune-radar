package store

import "testing"

// TestDSARResultShape confirms the DSAR result struct has the right fields.
// Full DB integration requires a real Postgres; this verifies the struct
// compiles and is usable — the smallest thing that fails if someone removes
// a field a caller depends on.
func TestDSARResultShape(t *testing.T) {
	r := DSARResult{
		PersonID:            "person_test",
		CouplesDeleted:      1,
		AccountsDeleted:     2,
		ObservationsDeleted: 10,
		HypothesesDeleted:   3,
		EvidenceDeleted:     5,
		ActionsCancelled:    1,
		ConsentRevoked:      1,
		LeadsDeleted:        1,
	}
	if r.PersonID != "person_test" || r.CouplesDeleted != 1 || r.ObservationsDeleted != 10 {
		t.Error("DSARResult should be usable with all fields")
	}
}
