package store

import (
	"testing"

	"neptune-social-radar/backend/internal/ontology"
)

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

// TestCreateConsentForCoupleSignature confirms CreateConsentForCouple exists
// with the right signature and that RevokeConsent + GetConsentStatus compile.
// Full DB integration (both persons get consent rows, audit event recorded)
// requires a real Postgres — see storetest/. This is the smallest thing that
// fails if someone removes a function or changes its signature.
func TestCreateConsentForCoupleSignature(t *testing.T) {
	var grant func(string, []string) ([]ontology.ConsentPolicy, error) = (*Store)(nil).CreateConsentForCouple
	_ = grant
	var status func(string) (ConsentStatus, error) = (*Store)(nil).GetConsentStatus
	_ = status
	// ConsentStatus must carry the fields the celebrate API returns.
	st := ConsentStatus{Granted: true, Revoked: false, AllowedActions: []string{"postcard"}}
	if !st.Granted || len(st.AllowedActions) != 1 {
		t.Fatal("ConsentStatus should be usable with granted + actions fields")
	}
}
