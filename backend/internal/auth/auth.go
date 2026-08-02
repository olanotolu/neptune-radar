// Package auth resolves the bearer token to a real user identity and role.
// It replaces the shared admin token with per-user API keys while keeping
// backward compatibility: if no users exist in the database, the server falls
// back to the shared NEPTUNE_ADMIN_TOKEN (admin role) so existing deployments
// don't break on upgrade.
package auth

import (
	"context"
	"crypto/subtle"
	"errors"
	"net/http"

	"neptune-social-radar/backend/internal/store"
)

type contextKey string

const userKey contextKey = "user"

// Identity is the resolved caller, available to handlers via UserFromContext.
// In legacy mode (shared admin token, no users table), Email is "legacy-admin"
// and Role is admin — so decided_by attribution still works.
type Identity struct {
	User  store.User
	Email string // for decided_by attribution
	Role  store.Role
}

// UserFromContext returns the authenticated identity, or a zero value if
// the handler is on a public route (health, media).
func UserFromContext(ctx context.Context) Identity {
	u, _ := ctx.Value(userKey).(Identity)
	return u
}

// Middleware resolves the Authorization header to a user. Resolution order:
//  1. If users exist in the DB, the bearer must be a user API key.
//  2. If no users exist, fall back to the shared admin token (legacy mode).
//
// This means a deployment can upgrade without disruption: the shared token
// keeps working until the first user is created, after which only user API
// keys are accepted.
func Middleware(s *store.Store, adminToken string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if adminToken == "" {
			writeAuthError(w, "server misconfigured: NEPTUNE_ADMIN_TOKEN is not set")
			return
		}
		// Public routes skip auth.
		if r.URL.Path == "/api/health" || r.URL.Path == "/api/media" {
			next.ServeHTTP(w, r)
			return
		}
		// Webhook routes authenticate via webhook secret (handler re-checks).
		if r.URL.Path == "/api/webhooks/neptune" && r.Method == http.MethodPost {
			next.ServeHTTP(w, r)
			return
		}

		token := extractBearer(r)
		if token == "" {
			writeAuthError(w, "missing bearer token")
			return
		}

		// Try per-user API key first (the future state).
		userCount, err := s.UserCount()
		if err == nil && userCount > 0 {
			user, err := s.GetUserByAPIKey(token)
			if err != nil {
				writeAuthError(w, "invalid or disabled API key")
				return
			}
			_ = s.TouchUserLastSeen(user.ID)
			ctx := context.WithValue(r.Context(), userKey, Identity{
				User:  user,
				Email: user.Email,
				Role:  user.Role,
			})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// Legacy mode: shared admin token. No users in the DB yet.
		if subtle.ConstantTimeCompare([]byte(token), []byte(adminToken)) != 1 {
			writeAuthError(w, "invalid bearer token")
			return
		}
		ctx := context.WithValue(r.Context(), userKey, Identity{
			Email: "legacy-admin",
			Role:  store.RoleAdmin,
		})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func extractBearer(r *http.Request) string {
	const prefix = "Bearer "
	h := r.Header.Get("Authorization")
	if len(h) <= len(prefix) || h[:len(prefix)] != prefix {
		return ""
	}
	return h[len(prefix):]
}

func writeAuthError(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte(`{"error":"` + msg + `"}`))
}

// ErrUnauthorized is for handlers that need to double-check role permissions.
var ErrUnauthorized = errors.New("unauthorized: insufficient role")

// ScopeVisible reports whether the caller's role may see a row at the given
// visibility scope. attorney_only rows are blocked for concierge; everything
// else is visible to all authenticated roles. Handlers call this at the trust
// boundary before returning visibility_scope-labeled data to the dashboard.
func ScopeVisible(role store.Role, scope string) bool {
	return store.CanAccessScope(role, scope)
}

// RequireAdmin is a handler-level guard for system-level operations (user
// management, source configuration, ingest control, DLQ, janitor). Returns
// true if the caller is admin; if not, it writes a 403 and returns false so
// the handler can early-return: `if !auth.RequireAdmin(w, r) { return }`.
func RequireAdmin(w http.ResponseWriter, r *http.Request) bool {
	if UserFromContext(r.Context()).Role != store.RoleAdmin {
		writeAuthError(w, "admin role required")
		return false
	}
	return true
}
