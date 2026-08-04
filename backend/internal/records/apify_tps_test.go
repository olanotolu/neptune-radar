package records

import (
	"context"
	"os"
	"strings"
	"testing"
)

func TestApifyTPS_PausedByDefault(t *testing.T) {
	t.Setenv("APIFY_TOKEN", "apify_api_test_token")
	t.Setenv("APIFY_ENABLED", "false")
	t.Setenv("APIFY_TPS_ENABLED", "false")
	a := NewApifyTPSFromEnv()
	if a.Available() {
		t.Fatal("Available must be false when paused")
	}
	st := ApifyTPSStatus()
	if !strings.Contains(st, "PAUSED") {
		t.Fatalf("status want PAUSED, got %q", st)
	}
	res, err := a.Search(context.Background(), Query{FirstName: "Carly", LastName: "Jordan", City: "College Station", Region: "TX"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Error == "" || !strings.Contains(res.Error, "PAUSED") {
		t.Fatalf("search must refuse without spending: %+v", res)
	}
}

func TestApifyTPS_GlobalPauseBlocksEvenIfTPSEnabled(t *testing.T) {
	t.Setenv("APIFY_TOKEN", "apify_api_test_token")
	t.Setenv("APIFY_ENABLED", "false")
	t.Setenv("APIFY_TPS_ENABLED", "true")
	a := NewApifyTPSFromEnv()
	if a.Available() {
		t.Fatal("global pause must block TPS even if APIFY_TPS_ENABLED=true")
	}
}

func TestApifyTPS_EnabledFlag(t *testing.T) {
	t.Setenv("APIFY_TOKEN", "apify_api_test_token")
	t.Setenv("APIFY_ENABLED", "true")
	t.Setenv("APIFY_TPS_ENABLED", "true")
	t.Setenv("APIFY_TPS_MAX_CALLS_PER_DETECTIVE", "1")
	a := NewApifyTPSFromEnv()
	if !a.Available() {
		t.Fatal("Available must be true when both flags enabled")
	}
	if a.MaxCalls != 1 {
		t.Fatalf("MaxCalls want 1 got %d", a.MaxCalls)
	}
	_ = os.Getenv
}

func TestParseUSAddress_SmashedCity(t *testing.T) {
	q := Query{City: "College Station", Region: "TX"}
	c, ok := parseUSAddressLine("304 Circleview Dr NHurst, TX 76054", "Carly Jordan", q)
	if !ok {
		t.Fatal("expected parse")
	}
	if !IsRealStreet(c.Line1) {
		t.Fatalf("line1 not street: %q", c.Line1)
	}
	if strings.Contains(c.Line1, "Hurst") {
		t.Fatalf("city leaked into line1: %q", c.Line1)
	}
	if !strings.EqualFold(c.City, "Hurst") {
		t.Fatalf("city want Hurst got %q", c.City)
	}
	if c.Postal != "76054" || c.Region != "TX" {
		t.Fatalf("region/zip: %s %s", c.Region, c.Postal)
	}
}

func TestParseUSAddress_SmashedMultiWordCity(t *testing.T) {
	q := Query{City: "College Station", Region: "TX"}
	c, ok := parseUSAddressLine("6721 Driffield Cir WNorth Richland Hills, TX 76182", "Carly Jordan", q)
	if !ok {
		t.Fatal("expected parse")
	}
	if strings.Contains(strings.ToLower(c.Line1), "richland") {
		t.Fatalf("city in line1: %q", c.Line1)
	}
	if !strings.Contains(strings.ToLower(c.City), "richland") {
		t.Fatalf("city want North Richland Hills got %q", c.City)
	}
	if c.Postal != "76182" {
		t.Fatalf("zip %s", c.Postal)
	}
}

func TestParseUSAddress_Standard(t *testing.T) {
	q := Query{City: "Houston", Region: "TX"}
	c, ok := parseUSAddressLine("123 Main Street, Houston, TX 77001", "Jane Doe", q)
	if !ok || c.Line1 != "123 Main Street" || c.City != "Houston" || c.Postal != "77001" {
		t.Fatalf("got %+v ok=%v", c, ok)
	}
}

func TestScoreApifyCityMatch(t *testing.T) {
	q := Query{City: "College Station", Region: "TX"}
	home := scoreApifyCandidate(Candidate{
		Line1: "100 University Dr", City: "College Station", Region: "TX", Postal: "77840",
		Kind: KindStreet, Confidence: 0.7,
	}, q, true)
	away := scoreApifyCandidate(Candidate{
		Line1: "304 Circleview Dr N", City: "Hurst", Region: "TX", Postal: "76054",
		Kind: KindStreet, Confidence: 0.7,
	}, q, true)
	if home.Confidence <= away.Confidence {
		t.Fatalf("home city should rank higher: home=%.2f away=%.2f", home.Confidence, away.Confidence)
	}
	if !strings.Contains(away.Note, "namesake") && !strings.Contains(away.Note, "differs") {
		t.Fatalf("away note should flag mismatch: %q", away.Note)
	}
}
