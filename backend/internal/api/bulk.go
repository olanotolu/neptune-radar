package api

import (
	"encoding/json"
	"net/http"
)

// bulkCoupleRequest is the payload for POST /api/couples/bulk.
type bulkCoupleRequest struct {
	Action string   `json:"action"` // "pause" | "resume" | "suppress"
	IDs    []string `json:"ids"`
	Reason string   `json:"reason"`
}

// bulkCouple applies the same action to many couples at once. Returns the
// number of couples affected.
func (s *Server) bulkCouple(w http.ResponseWriter, r *http.Request) {
	var body bulkCoupleRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if len(body.IDs) == 0 {
		writeError(w, http.StatusBadRequest, errorString("ids must not be empty"))
		return
	}
	switch body.Action {
	case "pause", "resume", "suppress":
	default:
		writeError(w, http.StatusBadRequest, errorString("action must be pause, resume, or suppress"))
		return
	}
	n, err := s.Store.BulkUpdateCouples(body.IDs, body.Action, body.Reason)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"action":   body.Action,
		"affected": n,
		"requested": len(body.IDs),
	})
}
