package api

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"neptune-social-radar/backend/internal/store"
)

// ingestFunnelWebhook closes the growth loop: Meet Neptune app (or a Zapier/
// Segment pipe) posts chat/booked/closed events. Auth: Bearer NEPTUNE_WEBHOOK_SECRET
// (preferred) or admin token. Idempotent via external_id.
//
// POST /api/webhooks/neptune
// {
//   "event": "chat_started" | "consult_booked" | "closed_won" | "closed_lost" | "handoff_clicked",
//   "couple_id": "optional",
//   "handoff_code": "optional ref from handoff link",
//   "utm_content": "optional (we set this to couple_id on handoff)",
//   "external_id": "idempotency key from product",
//   "occurred_at": "2026-07-31T12:00:00Z",
//   "metadata": { ... }
// }
func (s *Server) ingestFunnelWebhook(w http.ResponseWriter, r *http.Request) {
	if !webhookAuthorized(r) {
		writeError(w, http.StatusUnauthorized, errors.New("missing or invalid webhook credentials"))
		return
	}
	var in store.FunnelIngest
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if strings.TrimSpace(in.EventType) == "" {
		writeError(w, http.StatusBadRequest, errorString("event is required (chat_started|consult_booked|closed_won|closed_lost|handoff_clicked)"))
		return
	}
	if in.Source == "" {
		in.Source = "webhook"
	}
	ev, err := s.Store.IngestFunnelEvent(in)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	status := http.StatusCreated
	// If external_id hit existing, still 200 OK
	if in.ExternalID != "" && ev.ExternalID == in.ExternalID && !ev.CreatedAt.IsZero() {
		// CreatedAt equals occurred for new; for idempotent replay CreatedAt is older — use 200
		if time.Since(ev.CreatedAt) > 2*time.Second {
			status = http.StatusOK
		}
	}
	writeJSON(w, status, ev)
}

// listFunnelEvents is the operator view of closed-loop attribution.
func (s *Server) listFunnelEvents(w http.ResponseWriter, r *http.Request) {
	coupleID := r.URL.Query().Get("couple_id")
	events, err := s.Store.ListFunnelEvents(coupleID, 100)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if events == nil {
		events = []store.FunnelEvent{}
	}
	writeJSON(w, http.StatusOK, events)
}

// funnelStats returns 7d conversion rollup.
func (s *Server) funnelStats(w http.ResponseWriter, r *http.Request) {
	st, err := s.Store.GetFunnelStats()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, st)
}

// generateAutopsy runs a weekly (or custom range) false-positive autopsy.
// POST /api/trust/autopsy  { "days": 7 } or { "start": "...", "end": "..." }
func (s *Server) generateAutopsy(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Days  int    `json:"days"`
		Start string `json:"start"`
		End   string `json:"end"`
		By    string `json:"generated_by"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	end := time.Now().UTC()
	start := end.AddDate(0, 0, -7)
	if body.Days > 0 && body.Days <= 90 {
		start = end.AddDate(0, 0, -body.Days)
	}
	if body.Start != "" {
		if t, err := time.Parse(time.RFC3339, body.Start); err == nil {
			start = t.UTC()
		}
	}
	if body.End != "" {
		if t, err := time.Parse(time.RFC3339, body.End); err == nil {
			end = t.UTC()
		}
	}
	by := body.By
	if by == "" {
		by = "human:concierge"
	}
	rep, err := s.Store.GenerateAutopsy(start, end, by)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, rep)
}

func (s *Server) listAutopsies(w http.ResponseWriter, r *http.Request) {
	reps, err := s.Store.ListAutopsyReports(12)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if reps == nil {
		reps = []store.AutopsyReport{}
	}
	writeJSON(w, http.StatusOK, reps)
}

func (s *Server) getAutopsy(w http.ResponseWriter, r *http.Request) {
	rep, err := s.Store.GetAutopsyReport(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeJSON(w, http.StatusOK, rep)
}

// webhookAuthorized accepts NEPTUNE_WEBHOOK_SECRET or the admin bearer token.
func webhookAuthorized(r *http.Request) bool {
	secret := os.Getenv("NEPTUNE_WEBHOOK_SECRET")
	admin := os.Getenv("NEPTUNE_ADMIN_TOKEN")

	// Header form used by some providers
	if secret != "" {
		if h := r.Header.Get("X-Neptune-Webhook-Secret"); h != "" {
			if subtle.ConstantTimeCompare([]byte(h), []byte(secret)) == 1 {
				return true
			}
		}
	}

	const prefix = "Bearer "
	h := r.Header.Get("Authorization")
	if len(h) > len(prefix) && h[:len(prefix)] == prefix {
		tok := h[len(prefix):]
		if secret != "" && subtle.ConstantTimeCompare([]byte(tok), []byte(secret)) == 1 {
			return true
		}
		if admin != "" && subtle.ConstantTimeCompare([]byte(tok), []byte(admin)) == 1 {
			return true
		}
	}
	return false
}
