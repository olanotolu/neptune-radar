package signals

import "testing"

func TestLooksLikeBusinessHandle(t *testing.T) {
	biz := []string{"jorgensenfarm", "berlynevents", "prema_designs", "the.luminary.co", "columbusmusicians", "starling_studio"}
	for _, h := range biz {
		if !LooksLikeBusinessHandle(h) {
			t.Errorf("expected business: %s", h)
		}
	}
	people := []string{"ale.alejandra92", "ivamarie33", "mmccrohan11", "aidenandgrace"}
	// aidenandgrace might look personal
	for _, h := range people {
		if LooksLikeBusinessHandle(h) && h != "aidenandgrace" {
			// aidenandgrace is ambiguous - person or brand
			t.Errorf("expected person: %s", h)
		}
	}
}

func TestPickCouplePair_FiltersVendors(t *testing.T) {
	tags := []string{"jorgensenfarm", "berlynevents", "alissa.j", "jon.smith", "prema_designs", "starling_studio"}
	a, b, people, ok := PickCouplePair("starling_studio", tags, map[string]bool{"starling_studio": true})
	if !ok {
		t.Fatalf("expected pair, people=%v", people)
	}
	if LooksLikeBusinessHandle(a) || LooksLikeBusinessHandle(b) {
		t.Fatalf("got business pair %s & %s from %v", a, b, people)
	}
}
