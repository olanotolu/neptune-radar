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

// Cleanlist API v2 — search + verified enrichment via a 15-provider waterfall.
// Docs: https://docs.cleanlist.ai
// Best for: batch agent-driven resolution, free search reads, 98% email / 85% phone verification.
type Cleanlist struct {
	APIKey string
	Client *http.Client
}

func (c *Cleanlist) Name() string { return "cleanlist" }
func (c *Cleanlist) Available() bool {
	return strings.TrimSpace(c.APIKey) != ""
}

func (c *Cleanlist) client() *http.Client {
	if c.Client != nil {
		return c.Client
	}
	return &http.Client{Timeout: 30 * time.Second}
}

func (c *Cleanlist) Search(ctx context.Context, q Query) (Result, error) {
	if !c.Available() {
		return Result{Provider: "cleanlist", Status: "error", Error: "CLEANLIST_API_KEY not set"}, fmt.Errorf("cleanlist unavailable")
	}

	// Cleanlist person enrichment: POST /api/v2/enrichment/person
	// Accepts name + location and returns verified addresses
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
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.cleanlist.ai/api/v2/enrichment/person", bytes.NewReader(payload))
	if err != nil {
		return Result{Provider: "cleanlist", Status: "error", Error: err.Error()}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.client().Do(req)
	if err != nil {
		return Result{Provider: "cleanlist", Status: "error", Error: err.Error()}, err
	}
	defer resp.Body.Close()
	rawB, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	raw := string(rawB)

	if resp.StatusCode == 404 || resp.StatusCode == 204 {
		return Result{Provider: "cleanlist", Status: "empty", RawJSON: raw, CostCents: 1}, nil
	}
	if resp.StatusCode >= 300 {
		return Result{Provider: "cleanlist", Status: "error", RawJSON: raw, Error: fmt.Sprintf("cleanlist http %d", resp.StatusCode), CostCents: 1},
			fmt.Errorf("cleanlist http %d: %s", resp.StatusCode, truncate(raw, 200))
	}

	var parsed map[string]any
	_ = json.Unmarshal(rawB, &parsed)
	cands := extractCleanlistCandidates(parsed, q)
	st := "ok"
	if len(cands) == 0 {
		st = "empty"
	}
	return Result{Provider: "cleanlist", Status: st, Candidates: cands, RawJSON: raw, CostCents: 1}, nil
}

func extractCleanlistCandidates(parsed map[string]any, q Query) []Candidate {
	var out []Candidate
	fullName := fmt.Sprintf("%s %s", str(parsed["first_name"]), str(parsed["last_name"]))

	// Cleanlist returns address info in the enrichment response
	addresses, _ := parsed["addresses"].([]any)
	if addresses == nil {
		// Try single address object
		if addr, ok := parsed["address"].(map[string]any); ok {
			addresses = []any{addr}
		}
	}

	for _, addrItem := range addresses {
		addr, ok := addrItem.(map[string]any)
		if !ok {
			continue
		}
		c := Candidate{
			Line1:      firstStr(addr, "street", "line1", "address_line_1"),
			Line2:      firstStr(addr, "unit", "line2", "address_line_2"),
			City:       firstStr(addr, "city"),
			Region:     firstStr(addr, "state", "region"),
			Postal:     firstStr(addr, "zip", "postal_code"),
			Country:    firstNonEmpty(str(addr["country"]), "US"),
			Confidence: 0.75,
			Source:     "cleanlist",
			FullName:   strings.TrimSpace(fullName),
			Note:       "Cleanlist 15-provider waterfall — verified address. Confirm before mail.",
		}
		if q.City != "" && c.City != "" && strings.EqualFold(c.City, q.City) {
			c.Confidence = 0.85
		}
		if c.Line1 == "" && c.City == "" {
			continue
		}
		if c.Line1 == "" {
			c.Confidence = 0.45
			c.Note = "Cleanlist city-level only."
		}
		out = append(out, c)
		if len(out) >= 5 {
			break
		}
	}

	// Fallback: location_name
	if len(out) == 0 {
		if loc, ok := parsed["location"].(map[string]any); ok {
			city := str(loc["city"])
			state := str(loc["state"])
			if city != "" {
				out = append(out, Candidate{
					City: city, Region: state, Country: "US",
					Confidence: 0.45, Source: "cleanlist_location",
					FullName:   strings.TrimSpace(fullName),
					Note:       "Cleanlist location only (no street on this match).",
				})
			}
		}
	}
	return out
}
