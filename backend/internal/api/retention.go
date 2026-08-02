package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"neptune-social-radar/backend/internal/auth"
)

// listRetention returns all configured retention classes.
func (s *Server) listRetention(w http.ResponseWriter, r *http.Request) {
	classes, err := s.Store.ListRetentionClasses()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, classes)
}

// setRetention upserts a retention class. Admin-only.
func (s *Server) setRetention(w http.ResponseWriter, r *http.Request) {
	if !auth.RequireAdmin(w, r) {
		return
	}
	var req struct {
		EntityType  string `json:"entity_type"`
		MaxAgeDays  int    `json:"max_age_days"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if req.EntityType == "" || req.MaxAgeDays <= 0 {
		writeError(w, http.StatusBadRequest, errors.New("entity_type and positive max_age_days required"))
		return
	}
	if err := s.Store.SetRetentionClass(req.EntityType, req.MaxAgeDays, req.Description); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	s.Store.Audit("retention", req.EntityType, "set",
		map[string]any{"max_age_days": req.MaxAgeDays, "by": "human:" + auth.UserFromContext(r.Context()).Email}, "", 0)
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// purgePreview returns what WOULD be purged based on retention classes.
func (s *Server) purgePreview(w http.ResponseWriter, r *http.Request) {
	preview, err := s.Store.PurgePreview(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, preview)
}
