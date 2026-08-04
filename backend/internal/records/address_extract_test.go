package records

import "testing"

func TestExtractAddressFromBio_FullStreet(t *testing.T) {
	addr := ExtractAddressFromBio("📸 123 Main St Apt 4B Columbus, OH 43215 💍")
	if addr == nil {
		t.Fatal("expected address")
	}
	if addr.Line1 != "123 Main St" {
		t.Errorf("Line1 want '123 Main St', got %q", addr.Line1)
	}
	if addr.Line2 == "" {
		t.Error("expected Line2 apt")
	}
	if addr.Postal != "43215" {
		t.Errorf("postal: %q", addr.Postal)
	}
}

func TestExtractAddressFromBio_MultiWordStreet(t *testing.T) {
	addr := ExtractAddressFromBio("456 N High Street, Columbus, OH")
	if addr == nil {
		t.Fatal("expected address")
	}
	if addr.Line1 != "456 N High Street" {
		t.Errorf("Line1 want full street, got %q (truncation bug)", addr.Line1)
	}
}

func TestIsRealStreet_RequiresSuffix(t *testing.T) {
	if IsRealStreet("123 Main") {
		t.Error("123 Main without street type should fail")
	}
	if !IsRealStreet("123 Main St") {
		t.Error("123 Main St should pass")
	}
	if IsRealStreet("https://example.com/123") {
		t.Error("URL should fail")
	}
}

func TestHasConfirmableStreet(t *testing.T) {
	weak := []Candidate{{Line1: "123 Main St", City: "", Confidence: 0.9, Kind: KindStreet}}
	if HasConfirmableStreet(weak) {
		t.Error("no city should not be confirmable")
	}
	ok := []Candidate{{Line1: "123 Main St", City: "Columbus", Postal: "43215", Confidence: 0.6, Kind: KindStreet}}
	if !HasConfirmableStreet(ok) {
		t.Error("street+city+zip should be confirmable")
	}
}
