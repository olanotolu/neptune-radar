package outreach

import (
	"strings"
	"testing"

	"neptune-social-radar/backend/internal/store"
)

func TestRunPrep_CarlyBaylorReady(t *testing.T) {
	k := store.CongratulateKit{
		FirstNameA: "Carly", LastNameA: "Jordan",
		FirstNameB: "Baylor", LastNameB: "Dawes",
		HandleA: "carlyyjordan", HandleB: "dawes.baylor",
		BioA: "Texas A&M\nPsalm 118:6",
		BioB: "#lovebigdreambig",
	}
	p := RunPrep(k)
	if p.HomeCity != "College Station" {
		t.Errorf("home city from A&M: got %q", p.HomeCity)
	}
	if p.HomeRegion != "TX" {
		t.Errorf("region: %q", p.HomeRegion)
	}
	if !p.HasLastA || !p.HasLastB {
		t.Error("expected both lasts")
	}
	if !p.Ready {
		t.Errorf("expected READY, score=%.2f blockers=%v warnings=%v", p.Score, p.Blockers, p.Warnings)
	}
	if p.Score < ReadyThreshold {
		t.Errorf("score %.2f < threshold", p.Score)
	}
}

func TestRunPrep_MissingLastsBlocked(t *testing.T) {
	k := store.CongratulateKit{
		FirstNameA: "Alida", FirstNameB: "Andrew",
		MarketCity: "New York", MarketRegion: "NY",
	}
	p := RunPrep(k)
	if p.Ready {
		t.Error("should not be ready without lasts")
	}
	if p.Score >= MinRunThreshold {
		// may still be weak with city — missing lasts should hard-cap
	}
	found := false
	for _, b := range p.Blockers {
		if b == "missing_last_names" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected missing_last_names, got %v", p.Blockers)
	}
	if p.Score >= MinRunThreshold {
		t.Errorf("missing lasts should keep score under run threshold, got %.2f", p.Score)
	}
}

func TestRunPrep_GarbageCity(t *testing.T) {
	k := store.CongratulateKit{
		FirstNameA: "Jane", LastNameA: "Doe",
		FirstNameB: "John", LastNameB: "Smith",
		MarketCity: "the moment", MarketRegion: "WE",
	}
	p := RunPrep(k)
	if p.HomeCity == "the moment" {
		t.Error("garbage city should not be home")
	}
	// Without bio city, missing_city blocker
	hasMissing := false
	for _, b := range p.Blockers {
		if b == "missing_city" {
			hasMissing = true
		}
	}
	if !hasMissing && p.HasCity {
		t.Errorf("unexpected has city from garbage: %+v", p)
	}
}

func TestRunPrep_BrandBlocked(t *testing.T) {
	k := store.CongratulateKit{
		FirstNameA: "The Wooly", LastNameA: "NewYork",
		FirstNameB: "Taylor", LastNameB: "Feinman",
		HandleA: "thewoolynewyork",
		BioA:    "Reservations via resy. Event requests: events@...",
		MarketCity: "New York", MarketRegion: "NY",
	}
	p := RunPrep(k)
	if !p.VendorConfused {
		t.Error("expected vendor confused")
	}
	found := false
	for _, b := range p.Blockers {
		if strings.Contains(b, "brand") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected brand blocker, got %v", p.Blockers)
	}
}

func TestApplyPrepToKit_FillsCity(t *testing.T) {
	k := store.CongratulateKit{
		FirstNameA: "Carly", LastNameA: "Jordan",
		FirstNameB: "Baylor", LastNameB: "Dawes",
		BioA: "Texas A&M",
	}
	p := RunPrep(k)
	ApplyPrepToKit(&k, p)
	if k.AddressCity != "College Station" {
		t.Errorf("expected fill AddressCity, got %q", k.AddressCity)
	}
	if k.MailPayload["detective_prep"] == nil {
		t.Error("expected detective_prep in mail_payload")
	}
	got, ok := PrepFromKit(k)
	if !ok || !got.Ready {
		t.Errorf("round-trip prep: ok=%v ready=%v", ok, got.Ready)
	}
}
