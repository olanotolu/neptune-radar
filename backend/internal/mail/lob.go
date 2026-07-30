// Package mail sends physical postcards and verifies US addresses (Lob primary).
package mail

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// Address is a US mailing address.
type Address struct {
	Name           string `json:"name"`
	AddressLine1   string `json:"address_line1"`
	AddressLine2   string `json:"address_line2,omitempty"`
	AddressCity    string `json:"address_city"`
	AddressState   string `json:"address_state"`
	AddressZip     string `json:"address_zip"`
	AddressCountry string `json:"address_country"`
}

// VerifyResult is USPS-style deliverability from Lob.
type VerifyResult struct {
	Deliverable bool
	Address     Address
	RawJSON     string
	Error       string
}

// SendResult is a created postcard.
type SendResult struct {
	ExternalID           string
	Status               string
	ExpectedDeliveryDate string
	RawJSON              string
	CostCents            int
	Error                string
}

// Client talks to Lob Print & Mail.
type Client struct {
	APIKey string
	HTTP   *http.Client
	From   Address // return address (operator / Neptune)
}

// NewFromEnv builds a Lob client when LOB_API_KEY is set.
func NewFromEnv() *Client {
	key := strings.TrimSpace(os.Getenv("LOB_API_KEY"))
	if key == "" {
		return nil
	}
	from := Address{
		Name:           envOr("LOB_FROM_NAME", "Neptune"),
		AddressLine1:   os.Getenv("LOB_FROM_LINE1"),
		AddressCity:    os.Getenv("LOB_FROM_CITY"),
		AddressState:   os.Getenv("LOB_FROM_STATE"),
		AddressZip:     os.Getenv("LOB_FROM_ZIP"),
		AddressCountry: envOr("LOB_FROM_COUNTRY", "US"),
	}
	return &Client{
		APIKey: key,
		HTTP:   &http.Client{Timeout: 45 * time.Second},
		From:   from,
	}
}

func (c *Client) Available() bool {
	return c != nil && strings.TrimSpace(c.APIKey) != ""
}

func (c *Client) IsTest() bool {
	return strings.HasPrefix(c.APIKey, "test_")
}

func (c *Client) VerifyAddress(ctx context.Context, a Address) (VerifyResult, error) {
	if !c.Available() {
		return VerifyResult{Error: "LOB_API_KEY not set"}, fmt.Errorf("lob unavailable")
	}
	body := map[string]any{
		"primary_line":   a.AddressLine1,
		"secondary_line": a.AddressLine2,
		"city":           a.AddressCity,
		"state":          a.AddressState,
		"zip_code":       a.AddressZip,
	}
	raw, status, err := c.post(ctx, "https://api.lob.com/v1/us_verifications", body)
	if err != nil {
		return VerifyResult{Error: err.Error(), RawJSON: raw}, err
	}
	if status >= 300 {
		return VerifyResult{Error: fmt.Sprintf("http %d", status), RawJSON: raw}, fmt.Errorf("lob verify http %d", status)
	}
	var parsed map[string]any
	_ = json.Unmarshal([]byte(raw), &parsed)
	deliv := strings.EqualFold(str(parsed["deliverability"]), "deliverable") ||
		strings.EqualFold(str(parsed["deliverability"]), "deliverable_unnecessary_unit") ||
		strings.EqualFold(str(parsed["deliverability"]), "deliverable_incorrect_unit") ||
		strings.EqualFold(str(parsed["deliverability"]), "deliverable_missing_unit")
	// Also accept components when deliverability field present
	if d, ok := parsed["deliverability"].(string); ok && d != "" {
		deliv = strings.HasPrefix(strings.ToLower(d), "deliverable")
	}
	comps, _ := parsed["components"].(map[string]any)
	out := a
	if comps != nil {
		if s := str(comps["primary_line"]); s != "" {
			out.AddressLine1 = s
		}
		if s := str(comps["city"]); s != "" {
			out.AddressCity = s
		}
		if s := str(comps["state"]); s != "" {
			out.AddressState = s
		}
		if s := str(comps["zip_code"]); s != "" {
			out.AddressZip = s
		}
	}
	// Lob sometimes returns top-level rewritten fields
	if s := str(parsed["primary_line"]); s != "" {
		out.AddressLine1 = s
	}
	if s := str(parsed["city"]); s != "" {
		out.AddressCity = s
	}
	if s := str(parsed["state"]); s != "" {
		out.AddressState = s
	}
	if s := str(parsed["zip_code"]); s != "" {
		out.AddressZip = s
	}
	if out.AddressCountry == "" {
		out.AddressCountry = "US"
	}
	return VerifyResult{Deliverable: deliv || out.AddressLine1 != "", Address: out, RawJSON: raw}, nil
}

// SendPostcard creates a 4x6 postcard. front/back can be HTML URLs or raw HTML
// depending on Lob account settings — we send HTML strings for both sides.
func (c *Client) SendPostcard(ctx context.Context, to Address, frontHTML, backHTML, description string) (SendResult, error) {
	if !c.Available() {
		return SendResult{Error: "LOB_API_KEY not set"}, fmt.Errorf("lob unavailable")
	}
	if c.From.AddressLine1 == "" {
		return SendResult{Error: "LOB_FROM_LINE1 required for send"}, fmt.Errorf("missing return address")
	}
	body := map[string]any{
		"description": description,
		"to": map[string]any{
			"name":            to.Name,
			"address_line1":   to.AddressLine1,
			"address_line2":   to.AddressLine2,
			"address_city":    to.AddressCity,
			"address_state":   to.AddressState,
			"address_zip":     to.AddressZip,
			"address_country": firstNonEmpty(to.AddressCountry, "US"),
		},
		"from": map[string]any{
			"name":            c.From.Name,
			"address_line1":   c.From.AddressLine1,
			"address_line2":   c.From.AddressLine2,
			"address_city":    c.From.AddressCity,
			"address_state":   c.From.AddressState,
			"address_zip":     c.From.AddressZip,
			"address_country": firstNonEmpty(c.From.AddressCountry, "US"),
		},
		"front": frontHTML,
		"back":  backHTML,
		"size":  "4x6",
	}
	raw, status, err := c.post(ctx, "https://api.lob.com/v1/postcards", body)
	if err != nil {
		return SendResult{Error: err.Error(), RawJSON: raw}, err
	}
	if status >= 300 {
		return SendResult{Error: fmt.Sprintf("http %d: %s", status, truncate(raw, 300)), RawJSON: raw},
			fmt.Errorf("lob postcard http %d", status)
	}
	var parsed map[string]any
	_ = json.Unmarshal([]byte(raw), &parsed)
	return SendResult{
		ExternalID:           str(parsed["id"]),
		Status:               str(parsed["status"]),
		ExpectedDeliveryDate: str(parsed["expected_delivery_date"]),
		RawJSON:              raw,
		CostCents:            90, // approx; actual in Lob dashboard
	}, nil
}

func (c *Client) post(ctx context.Context, endpoint string, body any) (raw string, status int, err error) {
	payload, err := json.Marshal(body)
	if err != nil {
		return "", 0, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return "", 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(c.APIKey, "")
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	return string(b), resp.StatusCode, nil
}

func envOr(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func str(v any) string {
	s, _ := v.(string)
	return strings.TrimSpace(s)
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
