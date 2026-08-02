-- Per-user identity + RBAC: replaces the shared admin token with per-user API keys.
-- The frontend already stores a bearer token in localStorage; each user enters
-- their own API key instead of the shared admin token. Backward compat: if no
-- users exist, the server falls back to NEPTUNE_ADMIN_TOKEN (admin role).

CREATE TABLE IF NOT EXISTS users (
  id TEXT PRIMARY KEY,
  email TEXT NOT NULL UNIQUE,
  display_name TEXT NOT NULL,
  role TEXT NOT NULL DEFAULT 'concierge' CHECK (role IN ('admin', 'concierge', 'attorney')),
  api_key_hash TEXT NOT NULL UNIQUE,  -- SHA-256 of the plaintext key
  api_key_prefix TEXT NOT NULL,       -- first 8 chars for display ("npt_xxxx…")
  disabled_at TIMESTAMPTZ,
  last_seen_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS users_api_key_hash_idx ON users (api_key_hash) WHERE disabled_at IS NULL;
