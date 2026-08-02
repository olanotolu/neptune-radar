package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"neptune-social-radar/backend/internal/auth"
	"neptune-social-radar/backend/internal/store"
)

// markCoupleMistaken is the human override path: a concierge marks a couple
// as NOT actually a couple (identity resolution was wrong). The scorer
// respects this permanently — no new hypotheses fire for a mistaken couple.
func (s *Server) markCoupleMistaken(w http.ResponseWriter, r *http.Request) {
	identity := auth.UserFromContext(r.Context())
	if identity.Role != store.RoleAdmin && identity.Role != store.RoleConcierge {
		writeError(w, http.StatusForbidden, auth.ErrUnauthorized)
		return
	}
	var req struct {
		CoupleID string `json:"couple_id"`
		Reason   string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if req.CoupleID == "" {
		writeError(w, http.StatusBadRequest, errors.New("couple_id is required"))
		return
	}
	decidedBy := identity.Email
	if err := s.Store.MarkCoupleMistaken(req.CoupleID, req.Reason, decidedBy); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"status": "mistaken"})
}

// rejectHypothesis is the human override path: a concierge marks a
// hypothesis as rejected (the event didn't happen, or was misidentified).
// Pending recommended_actions for the hypothesis are cancelled.
func (s *Server) rejectHypothesis(w http.ResponseWriter, r *http.Request) {
	identity := auth.UserFromContext(r.Context())
	if identity.Role != store.RoleAdmin && identity.Role != store.RoleConcierge {
		writeError(w, http.StatusForbidden, auth.ErrUnauthorized)
		return
	}
	var req struct {
		HypothesisID string `json:"hypothesis_id"`
		Reason       string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if req.HypothesisID == "" {
		writeError(w, http.StatusBadRequest, errors.New("hypothesis_id is required"))
		return
	}
	decidedBy := identity.Email
	if err := s.Store.RejectHypothesis(req.HypothesisID, req.Reason, decidedBy); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"status": "rejected"})
}
