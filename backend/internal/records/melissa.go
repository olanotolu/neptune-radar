package records

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Melissa People Business Search — name + city/state → consumer addresses.
// Docs: https://docs.melissa.com/cloud-api/people-business-search/
type Melissa struct {
	LicenseKey string
	Client     *http.Client
}

func (m *Melissa) Name() string { return "melissa" }
func (m *Melissa) Available() bool {
	return strings.TrimSpace(m.LicenseKey) != ""
}

func (m *Melissa) client() *http.Client {
	if m.Client != nil {
		return m.Client
	}
	return &http.Client{Timeout: 25 * time.Second}
}

func (m *Melissa) Search(ctx context.Context, q Query) (Result, error) {
	if !m.Available() {
		return Result{Provider: "melissa", Status: "error", Error: "MELISSA_LICENSE_KEY not set"}, fmt.Errorf("melissa unavailable")
	}
	u, _ := url.Parse("https://search.melissadata.net/v5/web/contactsearch/docontactSearch")
	qs := u.Query()
	qs.Set("id", m.LicenseKey)
	qs.Set("format", "json")
	qs.Set("ff", q.FirstName)
	if q.LastName != "" {
		qs.Set("lf", q.LastName)
	}
	if q.City != "" {
		qs.Set("city", q.City)
	}
	if q.Region != "" {
		qs.Set("state", q.Region)
	}
	qs.Set("maxrecords", "5")
	u.RawQuery = qs.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return Result{Provider: "melissa", Status: "error", Error: err.Error()}, err
	}
	resp, err := m.client().Do(req)
	if err != nil {
		return Result{Provider: "melissa", Status: "error", Error: err.Error()}, err
	}
	defer resp.Body.Close()
	rawB, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	raw := string(rawB)
	if resp.StatusCode >= 300 {
		return Result{Provider: "melissa", Status: "error", RawJSON: raw, Error: fmt.Sprintf("http %d", resp.StatusCode)},
			fmt.Errorf("melissa http %d", resp.StatusCode)
	}

	var parsed map[string]any
	_ = json.Unmarshal(rawB, &parsed)
	cands := extractMelissa(parsed)
	st := "ok"
	if len(cands) == 0 {
		st = "empty"
	}
	return Result{Provider: "melissa", Status: st, Candidates: cands, RawJSON: raw, CostCents: 3}, nil
}

func extractMelissa(parsed map[string]any) []Candidate {
	var out []Candidate
	// Response shapes vary; try Records array
	records, _ := parsed["Records"].([]any)
	if records == nil {
		if r, ok := parsed["records"].([]any); ok {
			records = r
		}
	}
	for _, item := range records {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		// Nested Address or flat fields
		addr, _ := m["Address"].(map[string]any)
		if addr == nil {
			addr = m
		}
		c := Candidate{
			Line1:      firstStr(addr, "AddressLine1", "address_line1", "Address1", "Street"),
			Line2:      firstStr(addr, "AddressLine2", "address_line2", "Address2"),
			City:       firstStr(addr, "City", "city"),
			Region:     firstStr(addr, "State", "state", "Region"),
			Postal:     firstStr(addr, "PostalCode", "postal", "Zip", "ZIP"),
			Country:    "US",
			Confidence: 0.6,
			Source:     "melissa",
			FullName:   firstStr(m, "FullName", "Name", "full_name"),
			Note:       "Melissa People Search — confirm before mail.",
		}
		if c.Line1 == "" && c.City == "" {
			continue
		}
		if c.Line1 == "" {
			c.Confidence = 0.35
			c.Note = "Melissa city-level only."
		}
		out = append(out, c)
		if len(out) >= 5 {
			break
		}
	}
	return out
}
