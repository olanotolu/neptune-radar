package store

import "testing"

// TestRevokeConsentSignature confirms the cascade function returns the count
// of cancelled actions. The full DB integration test requires a real Postgres
// (see storetest/), but this verifies the function exists and compiles with
// the right signature — the smallest thing that fails if someone changes the
// return type and breaks a caller.
func TestRevokeConsentSignature(t *testing.T) {
	// Type assertion that RevokeConsent returns (int, error).
	var f func(string) (int, error) = (*Store)(nil).RevokeConsent
	_ = f
}
