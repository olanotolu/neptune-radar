package records

import (
	"context"
	"testing"
)

func TestMultiCascade_FreeProvidersIncluded(t *testing.T) {
	m := NewMulti()
	// Free providers should always be present (property + voter are always available)
	if len(m.Free) < 2 {
		t.Errorf("expected at least 2 free providers, got %d", len(m.Free))
	}
	// PropertyRecords and VoterRegistration should be in the free list
	foundProp, foundVoter := false, false
	for _, p := range m.Free {
		switch p.Name() {
		case "county_property":
			foundProp = true
		case "voter_registration":
			foundVoter = true
		}
	}
	if !foundProp {
		t.Error("county_property provider not in free cascade")
	}
	if !foundVoter {
		t.Error("voter_registration provider not in free cascade")
	}
}

func TestMultiCascade_FallbackToHeuristic(t *testing.T) {
	m := NewMulti()
	// With no API keys and no Bright Data, should still return heuristic candidates
	q := Query{
		FirstName: "Jane",
		LastName:  "Doe",
		City:      "Columbus",
		Region:    "OH",
	}
	res, err := m.Search(context.Background(), q)
	if err != nil {
		t.Fatalf("cascade should not error: %v", err)
	}
	if len(res.Candidates) == 0 {
		t.Error("cascade should return at least heuristic candidates")
	}
}

func TestHasStreetCandidates(t *testing.T) {
	if hasStreetCandidates(nil) {
		t.Error("nil should not have street candidates")
	}
	if hasStreetCandidates([]Candidate{{City: "Columbus"}}) {
		t.Error("candidate without Line1 should not count as street-level")
	}
	if !hasStreetCandidates([]Candidate{{Line1: "123 Main St", City: "Columbus"}}) {
		t.Error("candidate with Line1 should count as street-level")
	}
}

func TestPropertyRecords_Available(t *testing.T) {
	p := &PropertyRecords{}
	if !p.Available() {
		t.Error("property records should always be available")
	}
	if p.Name() != "county_property" {
		t.Errorf("expected county_property, got %s", p.Name())
	}
}

func TestVoterRegistration_Available(t *testing.T) {
	v := &VoterRegistration{}
	if !v.Available() {
		t.Error("voter registration should always be available")
	}
	if v.Name() != "voter_registration" {
		t.Errorf("expected voter_registration, got %s", v.Name())
	}
}

func TestVoterRegistration_EmptyRegion(t *testing.T) {
	v := &VoterRegistration{}
	res, err := v.Search(context.Background(), Query{FirstName: "Jane", LastName: "Doe"})
	if err != nil {
		t.Fatalf("should not error on empty region: %v", err)
	}
	if res.Status != "empty" {
		t.Errorf("expected empty status for no region, got %s", res.Status)
	}
}
