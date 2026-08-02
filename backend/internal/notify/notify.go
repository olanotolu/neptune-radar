// Package notify sends Slack-formatted webhook alerts for high-signal
// pipeline events (high-confidence couples, stage transitions). It uses
// only the standard library — no new dependencies.
package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Alert is one pipeline event worth telling an operator about.
type Alert struct {
	Type      string // "high_confidence_couple", "stage_transition"
	CoupleID  string
	Handles   []string
	Score     float64
	Stage     string
	City      string
	State     string
	Timestamp time.Time
}

// Notifier posts alerts to a Slack incoming-webhook URL. A nil or empty
// webhookURL makes Send a no-op so callers don't need to branch.
type Notifier struct {
	webhookURL string
	client     *http.Client
}

// NewNotifier returns a Notifier for the given Slack webhook URL. Pass an
// empty string to disable notifications (Send becomes a no-op).
func NewNotifier(webhookURL string) *Notifier {
	return &Notifier{
		webhookURL: webhookURL,
		client:     &http.Client{Timeout: 10 * time.Second},
	}
}

// Enabled reports whether a webhook URL is configured.
func (n *Notifier) Enabled() bool {
	return n != nil && n.webhookURL != ""
}

// Send posts a Slack-formatted JSON message to the webhook URL. If the
// webhook URL is empty it returns nil immediately (no-op, not an error).
func (n *Notifier) Send(ctx context.Context, alert Alert) error {
	if !n.Enabled() {
		return nil
	}
	payload := buildSlackPayload(alert)
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("notify: marshal payload: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, n.webhookURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("notify: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("notify: post webhook: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("notify: webhook returned %s", resp.Status)
	}
	return nil
}

// buildSlackPayload constructs the Slack incoming-webhook message body.
func buildSlackPayload(a Alert) map[string]any {
	title := "💍 New high-confidence couple detected"
	if a.Type == "stage_transition" {
		title = fmt.Sprintf("💍 Couple advanced to %s", a.Stage)
	}
	fields := []map[string]string{
		{"title": "Couple ID", "value": a.CoupleID, "short": "true"},
		{"title": "Score", "value": fmt.Sprintf("%.2f", a.Score), "short": "true"},
	}
	if len(a.Handles) > 0 {
		handles := a.Handles[0]
		for _, h := range a.Handles[1:] {
			handles += " + " + h
		}
		fields = append(fields, map[string]string{"title": "Handles", "value": handles, "short": "false"})
	}
	if a.City != "" || a.State != "" {
		loc := a.City
		if a.State != "" {
			if loc != "" {
				loc += ", "
			}
			loc += a.State
		}
		fields = append(fields, map[string]string{"title": "Location", "value": loc, "short": "true"})
	}
	if a.Stage != "" {
		fields = append(fields, map[string]string{"title": "Stage", "value": a.Stage, "short": "true"})
	}
	ts := a.Timestamp
	if ts.IsZero() {
		ts = time.Now().UTC()
	}
	fields = append(fields, map[string]string{"title": "Detected", "value": ts.UTC().Format(time.RFC3339), "short": "false"})

	return map[string]any{
		"text": title,
		"attachments": []map[string]any{
			{
				"color":  "good",
				"fields": fields,
			},
		},
	}
}
