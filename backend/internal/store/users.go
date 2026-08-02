package store

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"
)

// Role controls what a user can see and do. The visibility_scope gating
// (attorney_only data) keys off this: only 'attorney' and 'admin' roles can
// read attorney_only rows; 'concierge' cannot.
type Role string

const (
	RoleAdmin     Role = "admin"
	RoleConcierge Role = "concierge"
	RoleAttorney  Role = "attorney"
)

// User is one authenticated dashboard operator. The API key is never stored —
// only its SHA-256 hash. The prefix ("npt_xxxx") is stored for display so the
// UI can show which key is active without revealing it.
type User struct {
	ID           string     `json:"id"`
	Email        string     `json:"email"`
	DisplayName  string     `json:"display_name"`
	Role         Role       `json:"role"`
	APIKeyPrefix string     `json:"api_key_prefix"`
	DisabledAt   *time.Time `json:"disabled_at,omitempty"`
	LastSeenAt   *time.Time `json:"last_seen_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

// HashAPIKey returns the SHA-256 hex digest of a plaintext API key.
func HashAPIKey(key string) string {
	h := sha256.Sum256([]byte(key))
	return hex.EncodeToString(h[:])
}

// GenerateAPIKey creates a new plaintext API key ("npt_" + 32 hex chars).
// The caller stores the hash; the plaintext is returned once and never again.
func GenerateAPIKey() string {
	return "npt_" + NewID("") // reuse uuid generator for the random portion
}

// ErrUserNotFound is returned when no active user matches the given API key.
var ErrUserNotFound = errors.New("user not found")

// GetUserByAPIKey looks up a user by their API key hash. Returns
// ErrUserNotFound if the key is invalid or the user is disabled.
func (s *Store) GetUserByAPIKey(apiKey string) (User, error) {
	hash := HashAPIKey(apiKey)
	var u User
	var disabledAt, lastSeenAt sql.NullTime
	err := s.DB.QueryRow(
		`SELECT id, email, display_name, role, api_key_prefix, disabled_at, last_seen_at, created_at
		 FROM users WHERE api_key_hash = $1 AND disabled_at IS NULL`,
		hash,
	).Scan(&u.ID, &u.Email, &u.DisplayName, &u.Role, &u.APIKeyPrefix, &disabledAt, &lastSeenAt, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return User{}, ErrUserNotFound
	}
	if err != nil {
		return User{}, err
	}
	u.DisabledAt = timePtr(disabledAt)
	u.LastSeenAt = timePtr(lastSeenAt)
	return u, nil
}

// UserCount returns the total number of users (active or disabled). Used to
// decide whether to fall back to the shared admin token (no users = legacy mode).
func (s *Store) UserCount() (int, error) {
	var n int
	err := s.DB.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&n)
	return n, err
}

// CreateUser inserts a new user. The plaintext API key is returned once; only
// the hash is stored.
func (s *Store) CreateUser(email, displayName string, role Role) (User, string, error) {
	plaintext := GenerateAPIKey()
	u := User{
		ID:           NewID("user"),
		Email:        email,
		DisplayName:  displayName,
		Role:         role,
		APIKeyPrefix: plaintext[:12],
	}
	_, err := s.DB.Exec(
		`INSERT INTO users (id, email, display_name, role, api_key_hash, api_key_prefix) VALUES ($1, $2, $3, $4, $5, $6)`,
		u.ID, u.Email, u.DisplayName, u.Role, HashAPIKey(plaintext), u.APIKeyPrefix,
	)
	if err != nil {
		return User{}, "", err
	}
	return u, plaintext, nil
}

// ListUsers returns all users for the admin management UI.
func (s *Store) ListUsers() ([]User, error) {
	rows, err := s.DB.Query(
		`SELECT id, email, display_name, role, api_key_prefix, disabled_at, last_seen_at, created_at FROM users ORDER BY created_at`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []User
	for rows.Next() {
		var u User
		var disabledAt, lastSeenAt sql.NullTime
		if err := rows.Scan(&u.ID, &u.Email, &u.DisplayName, &u.Role, &u.APIKeyPrefix, &disabledAt, &lastSeenAt, &u.CreatedAt); err != nil {
			return nil, err
		}
		u.DisabledAt = timePtr(disabledAt)
		u.LastSeenAt = timePtr(lastSeenAt)
		out = append(out, u)
	}
	return out, rows.Err()
}

// TouchUserLastSeen updates last_seen_at for the given user ID.
func (s *Store) TouchUserLastSeen(userID string) error {
	_, err := s.DB.Exec(`UPDATE users SET last_seen_at = now() WHERE id = $1`, userID)
	return err
}

// RotateAPIKey generates a new API key for the user, stores its hash, and
// returns the plaintext key (shown once). The old key is immediately invalid.
func (s *Store) RotateAPIKey(userID string) (string, error) {
	plaintext := GenerateAPIKey()
	prefix := plaintext[:12]
	res, err := s.DB.Exec(
		`UPDATE users SET api_key_hash = $1, api_key_prefix = $2 WHERE id = $3`,
		HashAPIKey(plaintext), prefix, userID,
	)
	if err != nil {
		return "", err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return "", err
	}
	if n == 0 {
		return "", ErrUserNotFound
	}
	return plaintext, nil
}

// DisableUser sets disabled_at to now, preventing API key auth for this user.
func (s *Store) DisableUser(userID string) error {
	res, err := s.DB.Exec(`UPDATE users SET disabled_at = now() WHERE id = $1`, userID)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrUserNotFound
	}
	return nil
}

// EnableUser clears disabled_at, re-enabling API key auth for this user.
func (s *Store) EnableUser(userID string) error {
	res, err := s.DB.Exec(`UPDATE users SET disabled_at = NULL WHERE id = $1`, userID)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrUserNotFound
	}
	return nil
}

// CanAccessScope reports whether a role is allowed to read rows at the given
// visibility scope. attorney_only is restricted to attorney + admin; everything
// else is visible to all authenticated roles.
func CanAccessScope(role Role, scope string) bool {
	if scope == "attorney_only" {
		return role == RoleAttorney || role == RoleAdmin
	}
	return true
}

func timePtr(n sql.NullTime) *time.Time {
	if !n.Valid {
		return nil
	}
	t := n.Time
	return &t
}
