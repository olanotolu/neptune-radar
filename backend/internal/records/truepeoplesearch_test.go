package records

import "testing"

func TestTPSParseAllAddresses(t *testing.T) {
	html := `
	<span itemprop="streetAddress">123 Main Street</span>
	<span itemprop="addressLocality">College Station</span>
	<span itemprop="addressRegion">TX</span>
	<span itemprop="postalCode">77840</span>
	<span data-fn="Carly" data-ln="Jordan"></span>
	`
	cands := tpsParseAllAddresses(html, "Carly Jordan")
	if len(cands) != 1 {
		t.Fatalf("want 1, got %d", len(cands))
	}
	if !IsRealStreet(cands[0].Line1) {
		t.Errorf("line1: %q", cands[0].Line1)
	}
	if cands[0].City != "College Station" || cands[0].Region != "TX" || cands[0].Postal != "77840" {
		t.Errorf("addr: %+v", cands[0])
	}
}

func TestTPSParseResultsCards(t *testing.T) {
	html := `Something 456 Oak Avenue, Houston, TX 77001 more`
	cands := tpsParseResultsCards(html, "Jane Doe", Query{City: "Houston", Region: "TX"})
	if len(cands) == 0 {
		t.Fatal("expected card parse")
	}
	if !IsRealStreet(cands[0].Line1) {
		t.Errorf("line1 %q", cands[0].Line1)
	}
}

func TestTPSRequiresLastName(t *testing.T) {
	tps := &TruePeopleSearch{Token: "x"}
	res, _ := tps.Search(nil, Query{FirstName: "Carly", City: "Houston", Region: "TX"})
	if res.Status != "empty" {
		t.Errorf("want empty without last, got %s", res.Status)
	}
}
