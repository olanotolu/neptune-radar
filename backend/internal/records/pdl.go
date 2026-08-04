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

	// Top-level street (some plans)
	if line := pdlString(data["location_street_address"]); line != "" && IsRealStreet(line) {
		out = append(out, Candidate{
			Line1: line,
			City:  firstNonEmpty(pdlString(data["location_locality"]), q.City),
			Region: regionAbbrev(firstNonEmpty(pdlString(data["location_region"]), q.Region)),
			Postal: pdlString(data["location_postal_code"]),
			Country: "US", Kind: KindStreet, Confidence: 0.70, Source: "pdl",
			FullName: fullName, Note: "PDL location_street_address — verify before mail.",
		})
	}

	// street_addresses array (PDL schema varies by plan; free/basic often returns false not array)
	if arr, ok := data["street_addresses"].([]any); ok {
		for _, item := range arr {
			m, ok := item.(map[string]any)
			if !ok {
				continue
			}
			c := Candidate{
				Line1:      firstStr(m, "street_address", "address_line_1", "line1"),
				City:       firstStr(m, "locality", "city"),
				Region:     regionAbbrev(firstStr(m, "region", "region_code", "state")),
				Postal:     firstStr(m, "postal_code", "postal"),
				Country:    firstNonEmpty(str(m["country"]), "US"),
				Kind:       KindStreet,
				Confidence: 0.72,
				Source:     "pdl",
				FullName:   fullName,
				Note:       "People Data Labs street address — verify before mail.",
			}
			if !IsRealStreet(c.Line1) {
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

	// Locality fallbacks when plan does not include streets (boolean false on street fields)
	if len(out) == 0 {
		city, region := "", ""
		if loc := pdlString(data["location_name"]); loc != "" {
			parts := strings.Split(loc, ",")
			if len(parts) >= 1 {
				city = strings.TrimSpace(parts[0])
			}
			if len(parts) >= 2 {
				region = regionAbbrev(strings.TrimSpace(parts[1]))
			}
		}
		if city == "" {
			city = pdlString(data["location_locality"])
		}
		if region == "" {
			region = regionAbbrev(pdlString(data["location_region"]))
		}
		// Education school location (e.g. Texas A&M → College Station) — strong market signal
		if city == "" {
			if edu, ok := data["education"].([]any); ok && len(edu) > 0 {
				if em, ok := edu[0].(map[string]any); ok {
					if school, ok := em["school"].(map[string]any); ok {
						if loc, ok := school["location"].(map[string]any); ok {
							city = firstStr(loc, "locality", "name")
							if strings.Contains(city, ",") {
								parts := strings.Split(city, ",")
								city = strings.TrimSpace(parts[0])
								if len(parts) > 1 && region == "" {
									region = regionAbbrev(strings.TrimSpace(parts[1]))
								}
							}
							if region == "" {
								region = regionAbbrev(firstStr(loc, "region"))
							}
						}
					}
				}
			}
		}
		if city == "" {
			city = q.City
			region = q.Region
		}
		if city != "" {
			note := "PDL locality only — street fields not on this PDL plan/match (street_addresses unavailable)."
			out = append(out, Candidate{
				City: city, Region: region, Country: "US",
				Kind: KindLocality, Confidence: 0.42, Source: "pdl_location",
				FullName: fullName, Note: note,
			})
		}
	}
	return out
}

// pdlString coerces PDL fields; plan-limited fields often arrive as bool false.
func pdlString(v any) string {
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t)
	case float64:
		return strings.TrimSpace(fmt.Sprintf("%.0f", t))
	default:
		return ""
	}
}

func regionAbbrev(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if len(s) == 2 {
		return strings.ToUpper(s)
	}
	m := map[string]string{
		"texas": "TX", "ohio": "OH", "california": "CA", "new york": "NY",
		"florida": "FL", "illinois": "IL", "georgia": "GA", "pennsylvania": "PA",
		"north carolina": "NC", "michigan": "MI", "tennessee": "TN", "colorado": "CO",
	}
	if a, ok := m[strings.ToLower(s)]; ok {
		return a
	}
	return s
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
