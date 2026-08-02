package api

import (
	"encoding/json"
	"net/http"

	"neptune-social-radar/backend/internal/auth"
	"neptune-social-radar/backend/internal/store"
)

// User management endpoints (admin-only). These let the first admin bootstrap
// more users without CLI access — the first user must be created via a
// bootstrap command (cmd/server -bootstrap-user email@x.com) or direct SQL,
// since the API requires an existing admin to create users.

type createUserRequest struct {
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	Role        string `json:"role"`
}

type userResponse struct {
	store.User
	APIKey string `json:"api_key,omitempty"` // only returned on creation
}

func (s *Server) listUsers(w http.ResponseWriter, r *http.Request) {
	users, err := s.Store.ListUsers()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, users)
}

func (s *Server) createUser(w http.ResponseWriter, r *http.Request) {
	identity := auth.UserFromContext(r.Context())
	if identity.Role != store.RoleAdmin {
		writeError(w, http.StatusForbidden, auth.ErrUnauthorized)
		return
	}
	var req createUserRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if req.Email == "" || req.DisplayName == "" {
		writeError(w, http.StatusBadRequest, errBadRequest("email and display_name are required"))
		return
	}
	role := store.Role(req.Role)
	if role == "" {
		role = store.RoleConcierge
	}
	if role != store.RoleAdmin && role != store.RoleConcierge && role != store.RoleAttorney {
		writeError(w, http.StatusBadRequest, errBadRequest("role must be admin, concierge, or attorney"))
		return
	}
	user, plaintext, err := s.Store.CreateUser(req.Email, req.DisplayName, role)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, userResponse{User: user, APIKey: plaintext})
}

// rotateAPIKey generates a new API key for the given user. Admin-only.
// The plaintext key is returned once; the old key is immediately invalid.
func (s *Server) rotateAPIKey(w http.ResponseWriter, r *http.Request) {
	identity := auth.UserFromContext(r.Context())
	if identity.Role != store.RoleAdmin {
		writeError(w, http.StatusForbidden, auth.ErrUnauthorized)
		return
	}
	userID := r.PathValue("id")
	if userID == "" {
		writeError(w, http.StatusBadRequest, errBadRequest("user id required"))
		return
	}
	plaintext, err := s.Store.RotateAPIKey(userID)
	if err != nil {
		if err == store.ErrUserNotFound {
			writeError(w, http.StatusNotFound, err)
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"api_key": plaintext})
}

// disableUser sets disabled_at on the user, blocking API key auth. Admin-only.
func (s *Server) disableUser(w http.ResponseWriter, r *http.Request) {
	identity := auth.UserFromContext(r.Context())
	if identity.Role != store.RoleAdmin {
		writeError(w, http.StatusForbidden, auth.ErrUnauthorized)
		return
	}
	userID := r.PathValue("id")
	if userID == "" {
		writeError(w, http.StatusBadRequest, errBadRequest("user id required"))
		return
	}
	if err := s.Store.DisableUser(userID); err != nil {
		if err == store.ErrUserNotFound {
			writeError(w, http.StatusNotFound, err)
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "disabled"})
}

// enableUser clears disabled_at on the user, re-enabling API key auth. Admin-only.
func (s *Server) enableUser(w http.ResponseWriter, r *http.Request) {
	identity := auth.UserFromContext(r.Context())
	if identity.Role != store.RoleAdmin {
		writeError(w, http.StatusForbidden, auth.ErrUnauthorized)
		return
	}
	userID := r.PathValue("id")
	if userID == "" {
		writeError(w, http.StatusBadRequest, errBadRequest("user id required"))
		return
	}
	if err := s.Store.EnableUser(userID); err != nil {
		if err == store.ErrUserNotFound {
			writeError(w, http.StatusNotFound, err)
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "enabled"})
}

func decodeJSON(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}

func errBadRequest(msg string) error {
	return &badRequestError{msg: msg}
}

type badRequestError struct{ msg string }

func (e *badRequestError) Error() string { return e.msg }
