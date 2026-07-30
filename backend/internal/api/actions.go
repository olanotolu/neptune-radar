package api

import (
	"net/http"
	"strconv"

	"neptune-social-radar/backend/internal/pipeline/operator"
	"neptune-social-radar/backend/internal/pipeline/verifier"
	"neptune-social-radar/backend/internal/store"
)

func (s *Server) listActions(w http.ResponseWriter, r *http.Request) {
	actions, err := s.Store.ListActions(r.URL.Query().Get("status"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, actions)
}

type approveResponse struct {
	ID       string `json:"id"`
	Status   string `json:"status"`
	Verified bool   `json:"verified"`
}

// approveAction is the "human-reviewed action" step made concrete: it
// triggers the operator's side effect, then immediately asks the Verifier to
// re-read the database and confirm the intended state actually landed.
func (s *Server) approveAction(w http.ResponseWriter, r *http.Request) {
	actionID := r.PathValue("id")
	exec, err := operator.Approve(s.Store, actionID, "human:concierge")
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	action, err := s.Store.GetAction(actionID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	verified, err := verifier.Confirm(s.Store, exec, action)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, approveResponse{ID: actionID, Status: string(action.Status), Verified: verified})
}

func (s *Server) ignoreAction(w http.ResponseWriter, r *http.Request) {
	actionID := r.PathValue("id")
	if err := operator.Ignore(s.Store, actionID, "human:concierge"); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"id": actionID, "status": "ignored"})
}

func ontologyAuditFilterFromQuery(r *http.Request) store.AuditFilter {
	f := store.AuditFilter{
		EntityType: r.URL.Query().Get("entity_type"),
		EntityID:   r.URL.Query().Get("entity_id"),
		Monitor:    r.URL.Query().Get("monitor"),
	}
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			f.Limit = n
		}
	}
	return f
}
