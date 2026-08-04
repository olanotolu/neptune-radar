package ingest

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestApifyRetryOn5xx(t *testing.T) {

	t.Setenv("APIFY_ENABLED", "true")
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("server error"))
			return
		}
		w.Header().Set("content-type", "application/json")
		w.Write([]byte(`[{"id":"ok"}]`))
	}))
	defer srv.Close()

	c := NewApifyClient("test-token")
	c.client = &http.Client{Timeout: 5 * time.Second}
	c.baseURL = srv.URL + "/v2/acts"

	items, err := c.runSync(context.Background(), "test-actor", map[string]any{"test": true})
	if err != nil {
		t.Fatalf("expected success after retries, got: %v", err)
	}
	if calls != 3 {
		t.Errorf("expected 3 calls (2 retries + success), got %d", calls)
	}
	if len(items) != 1 {
		t.Errorf("expected 1 item, got %d", len(items))
	}
}

func TestApifyNoRetryOn4xx(t *testing.T) {

	t.Setenv("APIFY_ENABLED", "true")
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("bad request"))
	}))
	defer srv.Close()

	c := NewApifyClient("test-token")
	c.client = &http.Client{Timeout: 5 * time.Second}
	c.baseURL = srv.URL + "/v2/acts"

	_, err := c.runSync(context.Background(), "test-actor", map[string]any{"test": true})
	if err == nil {
		t.Error("expected error on 400")
	}
	if !strings.Contains(err.Error(), "http 400") {
		t.Errorf("expected http 400 error, got: %v", err)
	}
	if calls > 1 {
		t.Errorf("4xx should not retry, got %d calls", calls)
	}
}

func TestApifyRetryOn429(t *testing.T) {

	t.Setenv("APIFY_ENABLED", "true")
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls < 2 {
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte("rate limited"))
			return
		}
		w.Header().Set("content-type", "application/json")
		w.Write([]byte(`[{"id":"ok"}]`))
	}))
	defer srv.Close()

	c := NewApifyClient("test-token")
	c.client = &http.Client{Timeout: 5 * time.Second}
	c.baseURL = srv.URL + "/v2/acts"

	_, err := c.runSync(context.Background(), "test-actor", map[string]any{"test": true})
	if err != nil {
		t.Fatalf("expected success after 429 retry, got: %v", err)
	}
	if calls != 2 {
		t.Errorf("expected 2 calls (1 retry + success), got %d", calls)
	}
}

func TestApifyRetryOnNetworkError(t *testing.T) {

	t.Setenv("APIFY_ENABLED", "true")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if hj, ok := w.(http.Hijacker); ok {
			conn, _, _ := hj.Hijack()
			conn.Close()
		}
	}))
	defer srv.Close()

	c := NewApifyClient("test-token")
	c.client = &http.Client{Timeout: 2 * time.Second}
	c.baseURL = srv.URL + "/v2/acts"

	_, err := c.runSync(context.Background(), "test-actor", map[string]any{"test": true})
	if err == nil {
		t.Error("expected error on network failure")
	}
	if !strings.Contains(err.Error(), "apify") {
		t.Errorf("error should mention apify, got: %v", err)
	}
}
