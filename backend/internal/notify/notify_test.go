package notify

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSend_NoOpWhenEmpty(t *testing.T) {
	n := NewNotifier("")
	if err := n.Send(context.Background(), Alert{Type: "high_confidence_couple"}); err != nil {
		t.Fatalf("expected no-op nil, got %v", err)
	}
}

func TestSend_SlackJSONFormat(t *testing.T) {
	var received map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(body, &received); err != nil {
			t.Errorf("invalid JSON: %v", err)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", ct)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	n := NewNotifier(srv.URL)
	alert := Alert{
		Type:      "high_confidence_couple",
		CoupleID:  "cpl_123",
		Handles:   []string{"@alice", "@bob"},
		Score:     0.95,
		Stage:     "engaged",
		City:      "Columbus",
		State:     "OH",
		Timestamp: time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC),
	}
	if err := n.Send(context.Background(), alert); err != nil {
		t.Fatalf("Send: %v", err)
	}

	if received == nil {
		t.Fatal("no request received")
	}
	text, _ := received["text"].(string)
	if text != "💍 New high-confidence couple detected" {
		t.Errorf("text = %q, want high-confidence title", text)
	}
	attachments, ok := received["attachments"].([]any)
	if !ok || len(attachments) != 1 {
		t.Fatalf("expected 1 attachment, got %v", received["attachments"])
	}
	att := attachments[0].(map[string]any)
	if att["color"] != "good" {
		t.Errorf("color = %v, want good", att["color"])
	}
	fields, ok := att["fields"].([]any)
	if !ok {
		t.Fatalf("expected fields array, got %v", att["fields"])
	}
	// Verify couple ID and score fields are present.
	var foundCouple, foundScore bool
	for _, f := range fields {
		m := f.(map[string]any)
		if m["title"] == "Couple ID" && m["value"] == "cpl_123" {
			foundCouple = true
		}
		if m["title"] == "Score" && m["value"] == "0.95" {
			foundScore = true
		}
	}
	if !foundCouple {
		t.Error("missing Couple ID field")
	}
	if !foundScore {
		t.Error("missing Score field")
	}
}

func TestSend_StageTransitionTitle(t *testing.T) {
	var received map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &received)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	n := NewNotifier(srv.URL)
	alert := Alert{Type: "stage_transition", Stage: "married", Score: 0.92, CoupleID: "cpl_456"}
	if err := n.Send(context.Background(), alert); err != nil {
		t.Fatalf("Send: %v", err)
	}
	text, _ := received["text"].(string)
	if text != "💍 Couple advanced to married" {
		t.Errorf("text = %q, want stage transition title", text)
	}
}
