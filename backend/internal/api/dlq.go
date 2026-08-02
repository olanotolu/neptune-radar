package api

import (
	"net/http"
	"strconv"

	"neptune-social-radar/backend/internal/auth"
)

// listDLQ returns dead-lettered items. ?status=pending|replayed|all&limit=50.
func (s *Server) listDLQ(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	if status == "" {
		status = "pending"
	}
	limit := 50
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			limit = n
		}
	}
	items, err := s.Store.ListDLQ(status, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

// replayDLQ marks a dead-lettered item as successfully replayed.
func (s *Server) replayDLQ(w http.ResponseWriter, r *http.Request) {
	if !auth.RequireAdmin(w, r) {
		return
	}
	id := r.PathValue("id")
	if err := s.Store.MarkDLQReplayed(id); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	s.Store.Audit("dlq", id, "replayed",
		map[string]any{"by": "human:" + auth.UserFromContext(r.Context()).Email}, "", 0)
	writeJSON(w, http.StatusOK, map[string]string{"id": id, "status": "replayed"})
}

// retryDLQ increments a dead-lettered item's retry counter.
func (s *Server) retryDLQ(w http.ResponseWriter, r *http.Request) {
	if !auth.RequireAdmin(w, r) {
		return
	}
	id := r.PathValue("id")
	if err := s.Store.MarkDLQRetried(id); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	s.Store.Audit("dlq", id, "retried",
		map[string]any{"by": "human:" + auth.UserFromContext(r.Context()).Email}, "", 0)
	writeJSON(w, http.StatusOK, map[string]string{"id": id, "status": "retried"})
}
