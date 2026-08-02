package main

import (
	"strings"
	"testing"

	"neptune-social-radar/backend/internal/packs"
)

var allStates = []string{
	"AK", "AL", "AR", "AZ", "CA", "CO", "CT", "DC", "DE", "FL",
	"GA", "HI", "IA", "ID", "IL", "IN", "KS", "KY", "LA", "MA",
	"MD", "ME", "MI", "MN", "MO", "MS", "MT", "NC", "ND", "NE",
	"NH", "NJ", "NM", "NV", "NY", "OH", "OK", "OR", "PA", "RI",
	"SC", "SD", "TN", "TX", "UT", "VA", "VT", "WA", "WI", "WV", "WY",
}

func TestPackForAllJurisdictions(t *testing.T) {
	for _, st := range allStates {
		p := packs.PackFor(st)
		if p == nil {
			t.Errorf("packs.PackFor(%q) returned nil", st)
			continue
		}
		if p.State != st {
			t.Errorf("packs.PackFor(%q) returned pack with State %q", st, p.State)
		}
	}
}

func TestEveryPackHasGovSource(t *testing.T) {
	for _, st := range allStates {
		p := packs.PackFor(st)
		if p == nil {
			continue
		}
		if len(p.Government) == 0 {
			t.Errorf("%s: no government sources", st)
		}
	}
}

func TestEveryPackHasDiocese(t *testing.T) {
	for _, st := range allStates {
		p := packs.PackFor(st)
		if p == nil {
			continue
		}
		if len(p.Dioceses) == 0 {
			t.Errorf("%s: no dioceses", st)
		}
	}
}

func TestEveryPackHasParish(t *testing.T) {
	for _, st := range allStates {
		p := packs.PackFor(st)
		if p == nil {
			continue
		}
		if len(p.Parishes) == 0 {
			t.Errorf("%s: no parishes", st)
		}
	}
}

func TestEveryPackHasVendor(t *testing.T) {
	for _, st := range allStates {
		p := packs.PackFor(st)
		if p == nil {
			continue
		}
		if len(p.Vendors) == 0 {
			t.Errorf("%s: no vendors", st)
		}
	}
}

func TestCountyFIPSFormat(t *testing.T) {
	for _, st := range allStates {
		p := packs.PackFor(st)
		if p == nil {
			continue
		}
		for _, g := range p.Government {
			if len(g.CountyFIPS) != 5 {
				t.Errorf("%s: CountyFIPS %q is not 5 digits", st, g.CountyFIPS)
			}
			for _, c := range g.CountyFIPS {
				if c < '0' || c > '9' {
					t.Errorf("%s: CountyFIPS %q contains non-digit", st, g.CountyFIPS)
					break
				}
			}
		}
	}
}

func TestVendorHandles(t *testing.T) {
	for _, st := range allStates {
		p := packs.PackFor(st)
		if p == nil {
			continue
		}
		for _, v := range p.Vendors {
			if v.Handle == "" {
				t.Errorf("%s: vendor %q has empty handle", st, v.Name)
			}
			if strings.HasPrefix(v.Handle, "@") {
				t.Errorf("%s: vendor %q handle %q starts with @", st, v.Name, v.Handle)
			}
		}
	}
}

var validSourceClasses = map[string]bool{
	"engagement_photographer": true,
	"wedding_venue":           true,
	"jeweler":                 true,
	"wedding_planner":         true,
	"florist":                 true,
	"videographer":            true,
	"wedding_cake":            true,
	"bridal_shop":             true,
	"officiant":               true,
}

func TestVendorSourceClass(t *testing.T) {
	for _, st := range allStates {
		p := packs.PackFor(st)
		if p == nil {
			continue
		}
		for _, v := range p.Vendors {
			if !validSourceClasses[v.SourceClass] {
				t.Errorf("%s: vendor %q has invalid SourceClass %q", st, v.Name, v.SourceClass)
			}
		}
	}
}

func TestEpiscopalDiocesesForAllStates(t *testing.T) {
	for _, st := range allStates {
		d := packs.EpiscopalDiocesesFor(st)
		if len(d) == 0 {
			t.Errorf("%s: no Episcopal dioceses", st)
		}
		for _, dioc := range d {
			if dioc.Denomination != "episcopal" {
				t.Errorf("%s: Episcopal diocese %q has denomination %q", st, dioc.Name, dioc.Denomination)
			}
		}
	}
}

func TestMethodistConferencesForAllStates(t *testing.T) {
	for _, st := range allStates {
		d := packs.MethodistConferencesFor(st)
		if len(d) == 0 {
			t.Errorf("%s: no Methodist conferences", st)
		}
		for _, dioc := range d {
			if dioc.Denomination != "methodist" {
				t.Errorf("%s: Methodist conference %q has denomination %q", st, dioc.Name, dioc.Denomination)
			}
		}
	}
}

func TestJewishFederationsForAllStates(t *testing.T) {
	for _, st := range allStates {
		d := packs.JewishFederationsFor(st)
		if len(d) == 0 {
			t.Errorf("%s: no Jewish federations", st)
		}
		for _, dioc := range d {
			if dioc.Denomination != "jewish" {
				t.Errorf("%s: Jewish federation %q has denomination %q", st, dioc.Name, dioc.Denomination)
			}
		}
	}
}
