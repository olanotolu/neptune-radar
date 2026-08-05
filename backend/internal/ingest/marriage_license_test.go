package ingest

import (
	"encoding/json"
	"testing"
	"time"
)

// ponytail: one check, no framework bloat. Verifies the parser maps the
// MarriageSignals JSON shape into LicenseFiling records and that
// PredictWeddingDate falls back to filing+60d when no wedding date is given.
func TestParseAndPredictWeddingDate(t *testing.T) {
	raw := []byte(`{"filings":[
	  {"party_a_name":"Jane Roe","party_b_name":"John Doe","county":"Franklin","state":"OH","filing_date":"2026-09-01T10:00:00Z","external_id":"f1"},
	  {"party_a_name":"A B","party_b_name":"C D","county":"King","state":"WA","filing_date":"2026-09-10T10:00:00Z","wedding_date":"2026-10-15T10:00:00Z","external_id":"f2"}
	]}`)
	var resp struct {
		Filings []struct {
			PartyAName  string  `json:"party_a_name"`
			PartyBName  string  `json:"party_b_name"`
			County      string  `json:"county"`
			FilingDate  string  `json:"filing_date"`
			WeddingDate *string `json:"wedding_date"`
			ExternalID  string  `json:"external_id"`
		} `json:"filings"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(resp.Filings) != 2 {
		t.Fatalf("want 2 filings, got %d", len(resp.Filings))
	}
	// Filing 1: no wedding date → predicted = filed + 60d.
	filed1, _ := time.Parse(time.RFC3339, resp.Filings[0].FilingDate)
	f1 := LicenseFiling{PartyAName: resp.Filings[0].PartyAName, County: resp.Filings[0].County, FilingDate: filed1}
	pred := PredictWeddingDate(f1)
	want := f1.FilingDate.AddDate(0, 0, defaultWeddingWindowDays)
	if !pred.Equal(want) {
		t.Errorf("predicted = %v, want %v (filed+60d)", pred, want)
	}
	// Filing 2: explicit wedding date → predicted returns it verbatim.
	filed2, _ := time.Parse(time.RFC3339, resp.Filings[1].FilingDate)
	wd, _ := time.Parse(time.RFC3339, *resp.Filings[1].WeddingDate)
	f2 := LicenseFiling{FilingDate: filed2, WeddingDate: &wd}
	pred2 := PredictWeddingDate(f2)
	if !pred2.Equal(wd) {
		t.Errorf("predicted = %v, want explicit %v", pred2, wd)
	}
}
