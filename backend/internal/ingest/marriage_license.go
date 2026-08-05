package ingest

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"neptune-social-radar/backend/internal/ontology"
	"neptune-social-radar/backend/internal/store"
)

// LicenseFiling is one record from a marriage-license feed (MarriageSignals or
// any provider that emits the same shape). Names + county + filing date are the
// minimum; wedding date is optional (most filings don't include it yet).
type LicenseFiling struct {
	PartyAName    string     `json:"party_a_name"`
	PartyBName    string     `json:"party_b_name"`
	County        string     `json:"county"`
	State         string     `json:"state"`
	FilingDate    time.Time  `json:"filing_date"`
	WeddingDate   *time.Time `json:"wedding_date,omitempty"`
	ExternalID    string     `json:"external_id,omitempty"`
}

// LicenseFetcher is the pluggable source for marriage-license filings. The
// real provider (MarriageSignals) implements this; tests inject a stub.
type LicenseFetcher interface {
	FetchFilings(ctx context.Context, since time.Time) ([]LicenseFiling, error)
}

// MarriageSignalsFetcher hits the MarriageSignals API. The endpoint + key come
// from env vars (MARRIAGE_SIGNALS_URL, MARRIAGE_SIGNALS_KEY) so no new config
// system is needed. ponytail: ceiling = pagination/rate-limit handling; upgrade
// path = add a cursor + backoff when the feed volume exceeds one page.
type MarriageSignalsFetcher struct {
	URL string
	Key string
}

func NewMarriageSignalsFetcher() *MarriageSignalsFetcher {
	return &MarriageSignalsFetcher{
		URL: os.Getenv("MARRIAGE_SIGNALS_URL"),
		Key: os.Getenv("MARRIAGE_SIGNALS_KEY"),
	}
}

func (f *MarriageSignalsFetcher) Available() bool { return f.URL != "" && f.Key != "" }

func (f *MarriageSignalsFetcher) FetchFilings(ctx context.Context, since time.Time) ([]LicenseFiling, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", f.URL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+f.Key)
	q := req.URL.Query()
	q.Set("since", since.Format(time.RFC3339))
	req.URL.RawQuery = q.Encode()
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("marriage signals: %s", resp.Status)
	}
	var raw struct {
		Filings []struct {
			PartyAName  string  `json:"party_a_name"`
			PartyBName  string  `json:"party_b_name"`
			County      string  `json:"county"`
			State       string  `json:"state"`
			FilingDate  string  `json:"filing_date"`
			WeddingDate *string `json:"wedding_date"`
			ExternalID  string  `json:"external_id"`
		} `json:"filings"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decode marriage signals response: %w", err)
	}
	out := make([]LicenseFiling, 0, len(raw.Filings))
	for _, r := range raw.Filings {
		filed, err := time.Parse(time.RFC3339, r.FilingDate)
		if err != nil {
			log.Printf("[marriage-license] skip filing %s: bad filing_date %q", r.ExternalID, r.FilingDate)
			continue
		}
		f := LicenseFiling{
			PartyAName: r.PartyAName, PartyBName: r.PartyBName,
			County: r.County, State: r.State, FilingDate: filed, ExternalID: r.ExternalID,
		}
		if r.WeddingDate != nil {
			if wd, err := time.Parse(time.RFC3339, *r.WeddingDate); err == nil {
				f.WeddingDate = &wd
			}
		}
		out = append(out, f)
	}
	return out, nil
}

// PredictWeddingDate estimates the wedding date from a filing. Marriage licenses
// are valid 30-90 days depending on state; we use the midpoint (60 days) when no
// explicit wedding date is supplied. ponytail: ceiling = a per-state validity
// table would tighten this; upgrade path = replace defaultDays with a state→days
// map when the county-level rules matter for outreach timing.
var defaultWeddingWindowDays = 60

func PredictWeddingDate(f LicenseFiling) time.Time {
	if f.WeddingDate != nil {
		return *f.WeddingDate
	}
	return f.FilingDate.AddDate(0, 0, defaultWeddingWindowDays)
}

// IngestMarriageLicenses pulls filings from the fetcher and writes them as
// couples. Returns the number of couples created/updated. Idempotent: re-ingest
// of the same person pair updates license fields rather than duplicating.
func IngestMarriageLicenses(ctx context.Context, s *store.Store, fetcher LicenseFetcher, since time.Time) (int, error) {
	filings, err := fetcher.FetchFilings(ctx, since)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, f := range filings {
		if f.PartyAName == "" || f.PartyBName == "" {
			continue
		}
		predicted := PredictWeddingDate(f)
		_, err := s.IngestMarriageLicenseFiling(f.PartyAName, f.PartyBName, f.County, f.FilingDate, &predicted, f.WeddingDate)
		if err != nil {
			log.Printf("[marriage-license] ingest filing %s (%s + %s, %s): %v", f.ExternalID, f.PartyAName, f.PartyBName, f.County, err)
			continue
		}
		n++
	}
	return n, nil
}

// PriorityBucket classifies a couple by days-until-wedding for outreach
// prioritization. Couples in the 30-60 day window are the prime prenup-signing
// moment — that's "priority". <30 days is "urgent" (window closing), 60-90 is
// "early" (nurture), >90 or unknown is "monitor".
func PriorityBucket(c ontology.Couple) string {
	ref := c.PredictedWeddingDate
	if c.WeddingDate != nil {
		ref = c.WeddingDate
	}
	if ref == nil {
		return "monitor"
	}
	days := int(time.Until(*ref).Hours() / 24)
	switch {
	case days < 30:
		return "urgent"
	case days <= 60:
		return "priority"
	case days <= 90:
		return "early"
	default:
		return "monitor"
	}
}
