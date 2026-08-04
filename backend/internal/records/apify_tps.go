package records

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

// ApifyTPS runs TruePeopleSearch via Apify actors (bypasses our missing Bright Data unlocker zone).
// Primary actor: jungle_synthesizer/truepeoplesearch-people-search-scraper (name + city + state).
//
// COST GUARD: PAUSED by default. Set APIFY_TPS_ENABLED=true to spend credits.
// When enabled: primary actor only (no dual fallback burn), maxItems small, optional call budget.
type ApifyTPS struct {
	Token  string
	Client *http.Client
	// Actor IDs (username~name). Override via env APIFY_TPS_ACTOR / APIFY_TPS_ACTOR_FALLBACK.
	Actor         string
	FallbackActor string
	Enabled       bool // APIFY_TPS_ENABLED — false = never call Apify API
	MaxItems      int  // APIFY_TPS_MAX_ITEMS (default 2)
	// MaxCalls is process-local remaining budget for this detective run; 0 = unlimited when enabled.
	// Detective sets this from APIFY_TPS_MAX_CALLS_PER_DETECTIVE (default 1).
	MaxCalls int
	calls    int
}

func (a *ApifyTPS) Name() string { return "apify_tps" }
func (a *ApifyTPS) Available() bool {
	// Available for discovery only when token set AND explicitly enabled (cost pause).
	return strings.TrimSpace(a.Token) != "" && a.Enabled
}

func (a *ApifyTPS) client() *http.Client {
	if a.Client != nil {
		return a.Client
	}
	return &http.Client{Timeout: 10 * time.Minute}
}

func apifyTPSEnabled() bool {
	// Master pause first — never spend if global Apify is off.
	if !ApifyGloballyEnabled() {
		return false
	}
	v := strings.ToLower(strings.TrimSpace(os.Getenv("APIFY_TPS_ENABLED")))
	return v == "1" || v == "true" || v == "yes" || v == "on"
}

func apifyTPSMaxItems() int {
	n := 2
	if s := strings.TrimSpace(os.Getenv("APIFY_TPS_MAX_ITEMS")); s != "" {
		var x int
		if _, err := fmt.Sscanf(s, "%d", &x); err == nil && x > 0 {
			n = x
		}
	}
	if n > 5 {
		n = 5 // hard ceiling
	}
	if n < 1 {
		n = 1
	}
	return n
}

// ApifyTPSMaxCallsPerDetective is how many Apify actor runs one detective may fire (default 1).
func ApifyTPSMaxCallsPerDetective() int {
	n := 1
	if s := strings.TrimSpace(os.Getenv("APIFY_TPS_MAX_CALLS_PER_DETECTIVE")); s != "" {
		var x int
		if _, err := fmt.Sscanf(s, "%d", &x); err == nil {
			n = x
		}
	}
	if n < 0 {
		n = 0
	}
	if n > 3 {
		n = 3 // never spray credits
	}
	return n
}

func apifyTPSFromEnv() *ApifyTPS {
	return &ApifyTPS{
		Token: strings.TrimSpace(os.Getenv("APIFY_TOKEN")),
		Actor: firstNonEmpty(
			strings.TrimSpace(os.Getenv("APIFY_TPS_ACTOR")),
			"jungle_synthesizer~truepeoplesearch-people-search-scraper",
		),
		FallbackActor: firstNonEmpty(
			strings.TrimSpace(os.Getenv("APIFY_TPS_ACTOR_FALLBACK")),
			"intelscrape~truepeoplesearch-scraper",
		),
		Enabled:  apifyTPSEnabled(),
		MaxItems: apifyTPSMaxItems(),
		MaxCalls: ApifyTPSMaxCallsPerDetective(),
	}
}

// NewApifyTPSFromEnv is the multi-hunter entry point (paused unless APIFY_TPS_ENABLED=true).
func NewApifyTPSFromEnv() *ApifyTPS {
	return apifyTPSFromEnv()
}

// ApifyTPSStatus is human-readable config state for research notes.
func ApifyTPSStatus() string {
	a := apifyTPSFromEnv()
	if strings.TrimSpace(a.Token) == "" {
		return "unavailable: set APIFY_TOKEN"
	}
	if !ApifyGloballyEnabled() {
		return "PAUSED globally (APIFY_ENABLED=false) — zero Apify spend (TPS + ingest)"
	}
	if !a.Enabled {
		return "PAUSED TPS (APIFY_TPS_ENABLED=false) — global on but TPS hunters off"
	}
	return fmt.Sprintf("ENABLED actor=%s maxItems=%d maxCalls/run=%d (primary only, no dual fallback)",
		a.Actor, a.MaxItems, a.MaxCalls)
}

// Search runs Apify actors for first+last (+ optional city/state).
func (a *ApifyTPS) Search(ctx context.Context, q Query) (Result, error) {
	if strings.TrimSpace(a.Token) == "" {
		return Result{Provider: "apify_tps", Status: "empty", Error: "APIFY_TOKEN not set"}, nil
	}
	if !ApifyGloballyEnabled() {
		return Result{
			Provider: "apify_tps", Status: "empty",
			Error: "PAUSED: APIFY_ENABLED=false — no Apify API call made",
		}, nil
	}
	if !a.Enabled {
		return Result{
			Provider: "apify_tps", Status: "empty",
			Error: "PAUSED: APIFY_TPS_ENABLED is not true — no Apify API call made",
		}, nil
	}
	if a.MaxCalls > 0 && a.calls >= a.MaxCalls {
		return Result{
			Provider: "apify_tps", Status: "empty",
			Error: fmt.Sprintf("budget exhausted: max %d Apify call(s) per detective", a.MaxCalls),
		}, nil
	}
	if strings.TrimSpace(q.FirstName) == "" || strings.TrimSpace(q.LastName) == "" {
		return Result{Provider: "apify_tps", Status: "empty", Error: "first+last required"}, nil
	}

	a.calls++

	// COST: primary actor only. Do NOT auto-run fallback actor (doubles spend).
	cands, raw, err := a.runJungle(ctx, q)
	if err == nil && hasStreetCandidates(cands) {
		return Result{
			Provider: "apify_tps", Status: "ok", Candidates: cands,
			RawJSON: truncate(raw, 6000), CostCents: 5,
		}, nil
	}
	// Optional single fallback only if explicitly allowed
	allowFallback := strings.EqualFold(strings.TrimSpace(os.Getenv("APIFY_TPS_ALLOW_FALLBACK")), "true")
	var cands2 []Candidate
	var raw2 string
	var err2 error
	if allowFallback && a.MaxCalls > 0 && a.calls < a.MaxCalls {
		a.calls++
		cands2, raw2, err2 = a.runIntel(ctx, q)
		if err2 == nil && len(cands2) > 0 {
			merged := append(cands, cands2...)
			merged = dedupeCandidates(normalizeCandidates(merged))
			rankCandidates(merged)
			st := "ok"
			if !hasStreetCandidates(merged) {
				st = "empty"
			}
			return Result{
				Provider: "apify_tps", Status: st, Candidates: merged,
				RawJSON: truncate(raw+"\n---\n"+raw2, 8000), CostCents: 8,
			}, nil
		}
	}
	// Return whatever we have + research links (free)
	all := append(cands, cands2...)
	if len(all) == 0 {
		name := strings.TrimSpace(q.FirstName + " " + q.LastName)
		loc := strings.Trim(q.City+", "+q.Region, ", ")
		u := "https://www.truepeoplesearch.com/results?name=" + strings.ReplaceAll(name, " ", "+")
		if loc != "" {
			u += "&citystatezip=" + strings.ReplaceAll(loc, " ", "+")
		}
		msg := "Apify TPS: no street yet"
		if err != nil {
			msg += " · primary: " + err.Error()
		}
		if err2 != nil {
			msg += " · fallback: " + err2.Error()
		}
		all = append(all, ResearchLink(u, q.City, q.Region, name, "apify_tps", msg))
	}
	return Result{
		Provider: "apify_tps", Status: "ok", Candidates: all,
		RawJSON: truncate(raw+"\n"+raw2, 4000), CostCents: 3,
		Error: firstNonEmpty(errString(err), errString(err2)),
	}, nil
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func (a *ApifyTPS) runJungle(ctx context.Context, q Query) ([]Candidate, string, error) {
	maxItems := a.MaxItems
	if maxItems < 1 {
		maxItems = 2
	}
	input := map[string]any{
		"searchMode":        "name",
		"firstName":         q.FirstName,
		"lastName":          q.LastName,
		"maxItems":          maxItems,
		"sp_intended_usage": "Neptune Radar congratulate kit mailing address research — human verify before mail",
	}
	if q.City != "" {
		input["city"] = q.City
	}
	if q.Region != "" {
		input["state"] = strings.ToUpper(q.Region)
	}
	return a.runActor(ctx, a.Actor, input, q)
}

func (a *ApifyTPS) runIntel(ctx context.Context, q Query) ([]Candidate, string, error) {
	name := strings.TrimSpace(q.FirstName + " " + q.LastName)
	if q.City != "" {
		name = name + ", " + q.City
		if q.Region != "" {
			name += " " + strings.ToUpper(q.Region)
		}
	}
	input := map[string]any{
		"names":              []string{name},
		"maxResultsPerQuery": 5,
		"searchState":        strings.ToUpper(q.Region),
		"demoMode":           false,
	}
	return a.runActor(ctx, a.FallbackActor, input, q)
}

func (a *ApifyTPS) runActor(ctx context.Context, actorID string, input map[string]any, q Query) ([]Candidate, string, error) {
	if actorID == "" {
		return nil, "", fmt.Errorf("empty actor id")
	}
	// Start run
	payload, _ := json.Marshal(input)
	startURL := fmt.Sprintf("https://api.apify.com/v2/acts/%s/runs?waitForFinish=300", actorID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, startURL, bytes.NewReader(payload))
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("Authorization", "Bearer "+a.Token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := a.client().Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("apify start: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if resp.StatusCode >= 300 {
		return nil, string(body), fmt.Errorf("apify start http %d: %s", resp.StatusCode, truncate(string(body), 300))
	}
	var start struct {
		Data struct {
			ID               string `json:"id"`
			Status           string `json:"status"`
			DefaultDatasetID string `json:"defaultDatasetId"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &start); err != nil {
		return nil, string(body), fmt.Errorf("apify decode start: %w", err)
	}
	datasetID := start.Data.DefaultDatasetID
	status := start.Data.Status
	runID := start.Data.ID

	// Poll if not finished (waitForFinish may return early; TPS actors often need 5–10m)
	deadline := time.Now().Add(10 * time.Minute)
	for status != "SUCCEEDED" && status != "FAILED" && status != "ABORTED" && status != "TIMED-OUT" {
		if time.Now().After(deadline) {
			return nil, string(body), fmt.Errorf("apify run timeout status=%s run=%s", status, runID)
		}
		select {
		case <-ctx.Done():
			return nil, "", ctx.Err()
		case <-time.After(5 * time.Second):
		}
		pr, err := http.NewRequestWithContext(ctx, http.MethodGet,
			"https://api.apify.com/v2/actor-runs/"+runID, nil)
		if err != nil {
			return nil, "", err
		}
		pr.Header.Set("Authorization", "Bearer "+a.Token)
		presp, err := a.client().Do(pr)
		if err != nil {
			return nil, "", err
		}
		pb, _ := io.ReadAll(io.LimitReader(presp.Body, 1<<20))
		presp.Body.Close()
		var prun struct {
			Data struct {
				Status           string `json:"status"`
				DefaultDatasetID string `json:"defaultDatasetId"`
			} `json:"data"`
		}
		_ = json.Unmarshal(pb, &prun)
		status = prun.Data.Status
		if prun.Data.DefaultDatasetID != "" {
			datasetID = prun.Data.DefaultDatasetID
		}
		body = pb
	}
	if status != "SUCCEEDED" {
		return nil, string(body), fmt.Errorf("apify run %s status=%s", runID, status)
	}
	if datasetID == "" {
		return nil, string(body), fmt.Errorf("apify run %s: no dataset", runID)
	}

	// Fetch dataset items
	ir, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("https://api.apify.com/v2/datasets/%s/items?format=json&clean=true", datasetID), nil)
	if err != nil {
		return nil, "", err
	}
	ir.Header.Set("Authorization", "Bearer "+a.Token)
	iresp, err := a.client().Do(ir)
	if err != nil {
		return nil, "", err
	}
	defer iresp.Body.Close()
	itemsRaw, _ := io.ReadAll(io.LimitReader(iresp.Body, 8<<20))
	if iresp.StatusCode >= 300 {
		return nil, string(itemsRaw), fmt.Errorf("apify dataset http %d", iresp.StatusCode)
	}
	var items []map[string]any
	if err := json.Unmarshal(itemsRaw, &items); err != nil {
		// sometimes wrapped
		var wrap struct {
			Data []map[string]any `json:"data"`
		}
		if json.Unmarshal(itemsRaw, &wrap) == nil {
			items = wrap.Data
		} else {
			return nil, string(itemsRaw), fmt.Errorf("apify items decode: %w", err)
		}
	}
	cands := extractApifyPeopleItems(items, q)
	return cands, string(itemsRaw), nil
}

// Standard: "123 Main St, Houston, TX 77001"
var apifyStreetZipRe = regexp.MustCompile(`(?i)^(\d{1,6}\s+.+?)\s*,\s*([A-Za-z][A-Za-z .'-]+?)\s*,\s*([A-Z]{2})\s+(\d{5})(?:-\d{4})?$`)

// Smashed city (no comma after street): "304 Circleview Dr NHurst, TX 76054"
// or "6721 Driffield Cir WNorth Richland Hills, TX 76182"
// Street ends at street-type + optional directional; city may be multi-word.
var apifySmashedRe = regexp.MustCompile(`(?i)^(\d{1,6}\s+(?:[A-Za-z0-9.'\-]+\s+){0,6}(?:Street|St|Avenue|Ave|Road|Rd|Drive|Dr|Lane|Ln|Court|Ct|Boulevard|Blvd|Way|Place|Pl|Circle|Cir|Parkway|Pkwy)(?:\s+[NSEW])?)\s*([A-Z][a-zA-Z]+(?:\s+[A-Z][a-zA-Z]+){0,4})\s*,\s*([A-Z]{2})\s+(\d{5})(?:-\d{4})?$`)

// No-comma full string ending ST ZIP: split street-type boundary from multi-word city.
var apifySmashedNoCommaRe = regexp.MustCompile(`(?i)^(\d{1,6}\s+(?:[A-Za-z0-9.'\-]+\s+){0,6}(?:Street|St|Avenue|Ave|Road|Rd|Drive|Dr|Lane|Ln|Court|Ct|Boulevard|Blvd|Way|Place|Pl|Circle|Cir)(?:\s+[NSEW])?)([A-Z][a-zA-Z]+(?:\s+[A-Z][a-zA-Z]+){0,4})\s*,?\s*([A-Z]{2})\s+(\d{5})(?:-\d{4})?$`)

// "City, ST 12345" or "City ST 12345"
var apifyCityStateZipRe = regexp.MustCompile(`(?i)^([A-Za-z][A-Za-z .'-]+?),\s*([A-Z]{2})\s+(\d{5})(?:-\d{4})?$`)
var apifyCityStateZipRe2 = regexp.MustCompile(`(?i)^([A-Za-z][A-Za-z .'-]+?)\s+([A-Z]{2})\s+(\d{5})(?:-\d{4})?$`)

func extractApifyPeopleItems(items []map[string]any, q Query) []Candidate {
	var out []Candidate
	for _, it := range items {
		full := firstStr(it, "fullName", "name", "full_name")
		phone := firstPhone(it)

		// Structured nested current residence
		if m, ok := it["currentResidence"].(map[string]any); ok {
			if line := firstStr(m, "street", "line1", "address1", "streetAddress"); line != "" {
				c := Candidate{
					Line1: line, City: firstStr(m, "city", "locality"),
					Region: regionAbbrev(firstStr(m, "state", "region")),
					Postal: firstStr(m, "zip", "postal", "postalCode"),
					Country: "US", Kind: KindStreet, Source: "apify_tps",
					FullName: full, Phone: phone,
					Note: "Apify TPS current (structured) — verify before mail.",
				}
				if IsRealStreet(c.Line1) {
					out = append(out, scoreApifyCandidate(c, q, true))
				}
			}
		}

		// String current fields
		for _, key := range []string{"currentResidence", "currentAddress", "current_address", "address"} {
			curr := firstStr(it, key)
			if curr == "" {
				continue
			}
			if c, ok := parseUSAddressLine(curr, full, q); ok {
				c.Source = "apify_tps"
				c.Phone = phone
				if c.Kind == KindStreet {
					c.Note = "Apify TPS current — verify before mail."
					c = scoreApifyCandidate(c, q, true)
				}
				out = append(out, c)
			}
		}

		// Prior addresses
		for _, key := range []string{"previousResidences", "pastAddresses", "previous_addresses", "addresses"} {
			arr, ok := it[key].([]any)
			if !ok {
				continue
			}
			for i, v := range arr {
				s, _ := v.(string)
				if s == "" {
					if m, ok := v.(map[string]any); ok {
						if line := firstStr(m, "street", "line1", "address1"); line != "" {
							c := Candidate{
								Line1: line, City: firstStr(m, "city", "locality"),
								Region: regionAbbrev(firstStr(m, "state", "region")),
								Postal: firstStr(m, "zip", "postal", "postalCode"),
								Country: "US", Kind: KindStreet, Source: "apify_tps",
								FullName: full,
								Note:     fmt.Sprintf("Apify TPS prior #%d — may be stale.", i+1),
							}
							if IsRealStreet(c.Line1) {
								out = append(out, scoreApifyCandidate(c, q, false))
							}
							continue
						}
						s = firstStr(m, "full", "formatted", "address")
					}
				}
				if s == "" {
					continue
				}
				if c, ok := parseUSAddressLine(s, full, q); ok {
					c.Source = "apify_tps"
					c.Note = fmt.Sprintf("Apify TPS prior #%d — may be stale.", i+1)
					if c.Kind == KindStreet {
						c = scoreApifyCandidate(c, q, false)
					} else {
						c.Confidence = 0.40
					}
					out = append(out, c)
				}
			}
		}
		if len(out) >= 12 {
			break
		}
	}
	out = dedupeCandidates(normalizeCandidates(out))
	rankCandidates(out)
	return out
}

// scoreApifyCandidate boosts home-city hits and demotes wrong-metro streets.
func scoreApifyCandidate(c Candidate, q Query, current bool) Candidate {
	base := 0.62
	if current {
		base = 0.72
	}
	c.Confidence = base
	qc := strings.ToLower(strings.TrimSpace(q.City))
	cc := strings.ToLower(strings.TrimSpace(c.City))
	qr := strings.ToUpper(strings.TrimSpace(q.Region))
	cr := strings.ToUpper(strings.TrimSpace(c.Region))

	switch {
	case qc != "" && cc != "" && (cc == qc || strings.Contains(cc, qc) || strings.Contains(qc, cc)):
		c.Confidence += 0.14
		c.Note = strings.TrimSpace(c.Note + " · city matches query")
	case qc != "" && cc != "" && cc != qc:
		c.Confidence -= 0.18
		c.Note = strings.TrimSpace(c.Note + " · city differs from query (" + c.City + " vs " + q.City + ") — possible namesake")
	}
	if qr != "" && cr != "" && qr == cr {
		c.Confidence += 0.04
	} else if qr != "" && cr != "" && qr != cr {
		c.Confidence -= 0.10
		c.Note = strings.TrimSpace(c.Note + " · state mismatch")
	}
	if c.Postal != "" {
		c.Confidence += 0.03
	}
	if c.Confidence < 0.25 {
		c.Confidence = 0.25
	}
	if c.Confidence > 0.90 {
		c.Confidence = 0.90
	}
	if c.Kind == "" && IsRealStreet(c.Line1) {
		c.Kind = KindStreet
	}
	return c
}

func firstPhone(it map[string]any) string {
	if s := firstStr(it, "homePhone", "phone", "mobile"); s != "" {
		return s
	}
	for _, key := range []string{"mobileNumbers", "phones", "landlineNumbers"} {
		if arr, ok := it[key].([]any); ok && len(arr) > 0 {
			if s, ok := arr[0].(string); ok {
				return s
			}
		}
	}
	return ""
}

func parseUSAddressLine(s, fullName string, q Query) (Candidate, bool) {
	s = strings.TrimSpace(s)
	s = strings.Join(strings.Fields(s), " ") // collapse whitespace
	if s == "" {
		return Candidate{}, false
	}

	// 1) "123 Main St, City, ST ZIP"
	if m := apifyStreetZipRe.FindStringSubmatch(s); len(m) >= 5 {
		line1 := strings.TrimSpace(m[1])
		if IsRealStreet(line1) {
			return Candidate{
				Line1: line1, City: strings.TrimSpace(m[2]),
				Region: strings.ToUpper(strings.TrimSpace(m[3])), Postal: strings.TrimSpace(m[4]),
				Country: "US", Kind: KindStreet, FullName: fullName, Confidence: 0.74,
			}, true
		}
	}

	// 2) Smashed: "304 Circleview Dr NHurst, TX 76054" / "6721 Driffield Cir WNorth Richland Hills, TX 76182"
	for _, re := range []*regexp.Regexp{apifySmashedRe, apifySmashedNoCommaRe} {
		if m := re.FindStringSubmatch(s); len(m) >= 5 {
			line1 := strings.TrimSpace(m[1])
			city := strings.TrimSpace(m[2])
			// Fix glued directional+city: "WNorth" → city "North …" keep W on street
			if len(city) > 1 && (city[0] == 'N' || city[0] == 'S' || city[0] == 'E' || city[0] == 'W') {
				rest := city[1:]
				if len(rest) > 0 && rest[0] >= 'A' && rest[0] <= 'Z' {
					// "WNorth Richland Hills" → street gets " W", city "North Richland Hills"
					if !strings.HasSuffix(line1, " "+string(city[0])) && !strings.HasSuffix(line1, string(city[0])) {
						line1 = line1 + " " + string(city[0])
					}
					city = rest
				}
			}
			if IsRealStreet(line1) && city != "" {
				return Candidate{
					Line1: line1, City: city,
					Region: strings.ToUpper(strings.TrimSpace(m[3])), Postal: strings.TrimSpace(m[4]),
					Country: "US", Kind: KindStreet, FullName: fullName, Confidence: 0.70,
				}, true
			}
		}
	}

	// 3) Loose reverse parse: … ST ZIP at end
	if m := regexp.MustCompile(`(?i)^(.+?)\s+([A-Z]{2})\s+(\d{5})(?:-\d{4})?$`).FindStringSubmatch(s); len(m) >= 4 {
		left := strings.TrimSpace(m[1])
		region := strings.ToUpper(strings.TrimSpace(m[2]))
		postal := strings.TrimSpace(m[3])
		// left may be "street, city" or "street city"
		if i := strings.LastIndex(left, ","); i > 0 {
			line1 := strings.TrimSpace(left[:i])
			city := strings.TrimSpace(left[i+1:])
			if IsRealStreet(line1) {
				return Candidate{
					Line1: line1, City: city, Region: region, Postal: postal,
					Country: "US", Kind: KindStreet, FullName: fullName, Confidence: 0.72,
				}, true
			}
		}
		// Try split street-type from city without comma
		if sm := apifySmashedRe.FindStringSubmatch(s); len(sm) >= 5 {
			// already handled above
		} else if IsRealStreet(left) {
			// whole left is street; city unknown
			return Candidate{
				Line1: left, City: q.City, Region: region, Postal: postal,
				Country: "US", Kind: KindStreet, FullName: fullName, Confidence: 0.58,
				Note: "City assumed from query — verify.",
			}, true
		}
	}

	// 4) City-only
	if m := apifyCityStateZipRe.FindStringSubmatch(s); len(m) >= 4 {
		return Candidate{
			City: strings.TrimSpace(m[1]), Region: strings.ToUpper(strings.TrimSpace(m[2])),
			Postal: strings.TrimSpace(m[3]), Country: "US", Kind: KindLocality,
			FullName: fullName, Confidence: 0.42,
			Note: "Apify TPS locality only (no street on this record).",
		}, true
	}
	if m := apifyCityStateZipRe2.FindStringSubmatch(s); len(m) >= 4 {
		return Candidate{
			City: strings.TrimSpace(m[1]), Region: strings.ToUpper(strings.TrimSpace(m[2])),
			Postal: strings.TrimSpace(m[3]), Country: "US", Kind: KindLocality,
			FullName: fullName, Confidence: 0.40,
			Note: "Apify TPS locality only (no street on this record).",
		}, true
	}

	if IsRealStreet(s) {
		return Candidate{
			Line1: s, City: q.City, Region: q.Region, Country: "US",
			Kind: KindStreet, FullName: fullName, Confidence: 0.55,
			Note: "Street only — city from query.",
		}, true
	}
	return Candidate{}, false
}
