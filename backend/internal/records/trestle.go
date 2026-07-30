package records

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Trestle IQ — identity data enrichment with a continuously maintained address graph.
// Docs: https://docs.trestleiq.com/
// Best for: name + city → current mailing address, demographics, relatives.
type Trestle struct {
	APIKey string
	Client *http.Client
}

func (t *Trestle) Name() string { return "trestle" }
func (t *Trestle) Available() bool {
	return strings.TrimSpace(t.APIKey) != ""
}

func (t *Trestle) client() *http.Client {
	if t.Client != nil {
		return t.Client
	}
	return &http.Client{Timeout: 25 * time.Second}
}

func (t *Trestle) Search(ctx context.Context, q Query) (Result, error) {
	if !t.Available() {
		return Result{Provider: "trestle", Status: "error", Error: "TRESTLE_API_KEY not set"}, fmt.Errorf("trestle unavailable")
	}

	// Trestle Person Search API: POST /v1/person/search
	body := map[string]any{
		"first_name": q.FirstName,
		"last_name":  q.LastName,
	}
	if q.City != "" {
		body["city"] = q.City
	}
	if q.Region != "" {
		body["state"] = q.Region
	}

	payload, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.trestleiq.com/v1/person/search", bytes.NewReader(payload))
	if err != nil {
		return Result{Provider: "trestle", Status: "error", Error: err.Error()}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+t.APIKey)

	resp, err := t.client().Do(req)
	if err != nil {
		return Result{Provider: "trestle", Status: "error", Error: err.Error()}, err
	}
	defer resp.Body.Close()
	rawB, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	raw := string(rawB)

	if resp.StatusCode == 404 || resp.StatusCode == 204 {
		return Result{Provider: "trestle", Status: "empty", RawJSON: raw, CostCents: 7}, nil
	}
	if resp.StatusCode >= 300 {
		return Result{Provider: "trestle", Status: "error", RawJSON: raw, Error: fmt.Sprintf("trestle http %d", resp.StatusCode), CostCents: 7},
			fmt.Errorf("trestle http %d: %s", resp.StatusCode, truncate(raw, 200))
	}

	var parsed map[string]any
	_ = json.Unmarshal(rawB, &parsed)
	cands := extractTrestleCandidates(parsed, q)
	st := "ok"
	if len(cands) == 0 {
		st = "empty"
	}
	return Result{Provider: "trestle", Status: st, Candidates: cands, RawJSON: raw, CostCents: 7}, nil
}

func extractTrestleCandidates(parsed map[string]any, q Query) []Candidate {
	var out []Candidate

	// Trestle returns a "results" array with person records
	results, _ := parsed["results"].([]any)
	if results == nil {
		// Try single record
		if person, ok := parsed["person"].(map[string]any); ok {
			results = []any{person}
		}
	}

	for _, item := range results {
		person, ok := item.(map[string]any)
		if !ok {
			continue
		}
		fullName := fmt.Sprintf("%s %s", str(person["first_name"]), str(person["last_name"]))

		// Addresses array — Trestle returns current and historical
		addresses, _ := person["addresses"].([]any)
		for _, addrItem := range addresses {
			addr, ok := addrItem.(map[string]any)
			if !ok {
				continue
			}
			c := Candidate{
				Line1:      firstStr(addr, "street_line_1", "line1", "street_address"),
				Line2:      firstStr(addr, "street_line_2", "line2", "unit"),
				City:       firstStr(addr, "city", "locality"),
				Region:     firstStr(addr, "state", "region", "region_code"),
				Postal:     firstStr(addr, "zip_code", "postal_code", "zip"),
				Country:    firstNonEmpty(str(addr["country"]), "US"),
				Confidence: 0.70,
				Source:     "trestle",
				FullName:   strings.TrimSpace(fullName),
				Note:       "Trestle IQ — continuously maintained address graph. Verify before mail.",
			}
			// Boost confidence if city matches query
			if q.City != "" && c.City != "" && strings.EqualFold(c.City, q.City) {
				c.Confidence = 0.80
			}
			if c.Line1 == "" && c.City == "" {
				continue
			}
			if c.Line1 == "" {
				c.Confidence = 0.40
				c.Note = "Trestle city-level only."
			}
			out = append(out, c)
			if len(out) >= 5 {
				break
			}
		}

		// Fallback: location if no addresses
		if len(out) == 0 {
			if loc, ok := person["location"].(map[string]any); ok {
				city := str(loc["city"])
				state := str(loc["state"])
				if city != "" {
					out = append(out, Candidate{
						City: city, Region: state, Country: "US",
						Confidence: 0.40, Source: "trestle_location",
						FullName:   strings.TrimSpace(fullName),
						Note:       "Trestle location only (no street on this match).",
					})
				}
			}
		}
		if len(out) >= 5 {
			break
		}
	}
	return out
}
