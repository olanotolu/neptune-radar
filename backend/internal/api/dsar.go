package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"neptune-social-radar/backend/internal/auth"
	"neptune-social-radar/backend/internal/store"
)

// dsarDelete handles a GDPR/CCPA right-to-erasure request. Admin-only —
// this is a destructive, irreversible operation that deletes a person and
// all their derived data. The request itself is logged in the audit trail
// (which is append-only) so there's a record that the deletion happened.
func (s *Server) dsarDelete(w http.ResponseWriter, r *http.Request) {
	identity := auth.UserFromContext(r.Context())
	if identity.Role != store.RoleAdmin {
		writeError(w, http.StatusForbidden, auth.ErrUnauthorized)
		return
	}
	var req struct {
		PersonID string `json:"person_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if req.PersonID == "" {
		writeError(w, http.StatusBadRequest, errors.New("person_id is required"))
		return
	}
	result, err := s.Store.DSARDelete(r.Context(), req.PersonID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}
