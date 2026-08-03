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

// PDL is People Data Labs Person Enrichment / Identify-style lookup.
// Docs: https://docs.peopledatalabs.com/
type PDL struct {
	APIKey string
	Client *http.Client
}

func (p *PDL) Name() string { return "pdl" }
func (p *PDL) Available() bool {
	return strings.TrimSpace(p.APIKey) != ""
}

func (p *PDL) client() *http.Client {
	if p.Client != nil {
		return p.Client
	}
	return &http.Client{Timeout: 25 * time.Second}
}

func (p *PDL) Search(ctx context.Context, q Query) (Result, error) {
	if !p.Available() {
		return Result{Provider: "pdl", Status: "error", Error: "PDL_API_KEY not set"}, fmt.Errorf("pdl unavailable")
	}

	// Person Enrichment accepts name + location (and optional profile).
	body := map[string]any{
		"pretty": true,
	}
	if q.FirstName != "" {
		body["first_name"] = q.FirstName
	}
	if q.LastName != "" {
		body["last_name"] = q.LastName
	}
	if q.FirstName != "" && q.LastName == "" {
		body["name"] = q.FirstName
	}
	if q.FirstName != "" && q.LastName != "" {
		body["name"] = strings.TrimSpace(q.FirstName + " " + q.LastName)
	}
	if q.City != "" || q.Region != "" {
		loc := strings.Trim(q.City+", "+q.Region, ", ")
		body["location"] = loc
		body["locality"] = q.City
		body["region"] = q.Region
	}
	if q.Handle != "" {
		h := strings.TrimSpace(q.Handle)
		h = strings.TrimPrefix(h, "@")
		body["profile"] = "instagram.com/" + h
	}

	payload, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.peopledatalabs.com/v5/person/enrich", bytes.NewReader(payload))
	if err != nil {
		return Result{Provider: "pdl", Status: "error", Error: err.Error()}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", p.APIKey)

	resp, err := p.client().Do(req)
	if err != nil {
		return Result{Provider: "pdl", Status: "error", Error: err.Error()}, err
	}
	defer resp.Body.Close()
	rawB, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	raw := string(rawB)

	if resp.StatusCode == 404 {
		// No match — try partner or return empty
		return Result{Provider: "pdl", Status: "empty", RawJSON: raw, CostCents: 1}, nil
	}
	if resp.StatusCode >= 300 {
		return Result{Provider: "pdl", Status: "error", RawJSON: raw, Error: fmt.Sprintf("pdl http %d", resp.StatusCode), CostCents: 1},
			fmt.Errorf("pdl http %d: %s", resp.StatusCode, truncate(raw, 200))
	}

	var parsed map[string]any
	_ = json.Unmarshal(rawB, &parsed)
	data, _ := parsed["data"].(map[string]any)
	if data == nil {
		data = parsed
	}
	cands := extractPDLCandidates(data, q)
	st := "ok"
	if len(cands) == 0 {
		st = "empty"
	}
	return Result{Provider: "pdl", Status: st, Candidates: cands, RawJSON: raw, CostCents: 5}, nil
}

func extractPDLCandidates(data map[string]any, q Query) []Candidate {
	var out []Candidate
	fullName, _ := data["full_name"].(string)
	// street_addresses array (PDL schema varies by plan)
	if arr, ok := data["street_addresses"].([]any); ok {
		for _, item := range arr {
			m, ok := item.(map[string]any)
			if !ok {
				continue
			}
			c := Candidate{
				Line1:      str(m["street_address"]),
				City:       firstStr(m, "locality", "city"),
				Region:     firstStr(m, "region", "region_code", "state"),
				Postal:     firstStr(m, "postal_code", "postal"),
				Country:    firstNonEmpty(str(m["country"]), "US"),
				Confidence: 0.55,
				Source:     "pdl",
				FullName:   fullName,
				Note:       "People Data Labs street address — verify before mail.",
			}
			if c.Line1 == "" {
				continue
			}
			if q.City != "" && c.City != "" && !strings.EqualFold(c.City, q.City) {
				c.Confidence -= 0.1
			}
			out = append(out, c)
			if len(out) >= 5 {
				break
			}
		}
	}
	// location_names fallback
	if len(out) == 0 {
		if loc, ok := data["location_name"].(string); ok && loc != "" {
			// "Columbus, Ohio, United States"
			parts := strings.Split(loc, ",")
			city, region := q.City, q.Region
			if len(parts) >= 1 {
				city = strings.TrimSpace(parts[0])
			}
			if len(parts) >= 2 {
				region = strings.TrimSpace(parts[1])
			}
			out = append(out, Candidate{
				City: city, Region: region, Country: "US",
				Confidence: 0.35, Source: "pdl_location",
				FullName: fullName,
				Note:     "PDL locality only (no street on this plan/match).",
			})
		}
	}
	return out
}

func str(v any) string {
	s, _ := v.(string)
	return strings.TrimSpace(s)
}

func firstStr(m map[string]any, keys ...string) string {
	for _, k := range keys {
		if s := str(m[k]); s != "" {
			return s
		}
	}
	return ""
}

func firstNonEmpty(a, b string) string {
	if strings.TrimSpace(a) != "" {
		return a
	}
	return b
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
