-- Detective core: structured names, observation facts, address lookups, mail sends.
-- Makes Instagram evidence agent-queryable and auditable for address/mail.

-- Structured names on persons + social accounts
ALTER TABLE persons
  ADD COLUMN IF NOT EXISTS first_name TEXT,
  ADD COLUMN IF NOT EXISTS last_name TEXT,
  ADD COLUMN IF NOT EXISTS name_source TEXT,
  ADD COLUMN IF NOT EXISTS name_confidence DOUBLE PRECISION;

ALTER TABLE social_accounts
  ADD COLUMN IF NOT EXISTS first_name TEXT,
  ADD COLUMN IF NOT EXISTS last_name TEXT,
  ADD COLUMN IF NOT EXISTS name_source TEXT,
  ADD COLUMN IF NOT EXISTS name_confidence DOUBLE PRECISION;

-- Promote key fields out of raw_payload for detective queries
ALTER TABLE social_observations
  ADD COLUMN IF NOT EXISTS caption TEXT,
  ADD COLUMN IF NOT EXISTS image_url TEXT,
  ADD COLUMN IF NOT EXISTS post_url TEXT,
  ADD COLUMN IF NOT EXISTS location_name TEXT,
  ADD COLUMN IF NOT EXISTS source_handle TEXT,
  ADD COLUMN IF NOT EXISTS tags_json TEXT DEFAULT '[]',
  ADD COLUMN IF NOT EXISTS mentions_json TEXT DEFAULT '[]',
  ADD COLUMN IF NOT EXISTS facts_extracted_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_obs_source_handle ON social_observations (source_handle)
  WHERE source_handle IS NOT NULL AND source_handle <> '';
CREATE INDEX IF NOT EXISTS idx_obs_caption_trgm ON social_observations (observed_at DESC);

-- Kit first/last names (postcard + people-search inputs)
ALTER TABLE congratulate_kits
  ADD COLUMN IF NOT EXISTS first_name_a TEXT,
  ADD COLUMN IF NOT EXISTS last_name_a TEXT,
  ADD COLUMN IF NOT EXISTS first_name_b TEXT,
  ADD COLUMN IF NOT EXISTS last_name_b TEXT,
  ADD COLUMN IF NOT EXISTS name_source_a TEXT,
  ADD COLUMN IF NOT EXISTS name_source_b TEXT,
  ADD COLUMN IF NOT EXISTS last_detective_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS mail_external_id TEXT,
  ADD COLUMN IF NOT EXISTS mail_provider TEXT;

-- Every paid/free address research call (audit + cost)
CREATE TABLE IF NOT EXISTS address_lookups (
  id TEXT PRIMARY KEY,
  kit_id TEXT REFERENCES congratulate_kits(id),
  couple_id TEXT REFERENCES couples(id),
  provider TEXT NOT NULL,
  query_json TEXT NOT NULL DEFAULT '{}',
  response_json TEXT,
  candidates_json TEXT NOT NULL DEFAULT '[]',
  status TEXT NOT NULL DEFAULT 'ok',
  error_message TEXT,
  cost_cents INTEGER NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS address_lookups_kit_idx ON address_lookups(kit_id);
CREATE INDEX IF NOT EXISTS address_lookups_couple_idx ON address_lookups(couple_id);

-- Physical mail send log (Lob / PostGrid / Thanks.io)
CREATE TABLE IF NOT EXISTS mail_sends (
  id TEXT PRIMARY KEY,
  kit_id TEXT NOT NULL REFERENCES congratulate_kits(id),
  couple_id TEXT REFERENCES couples(id),
  provider TEXT NOT NULL,
  external_id TEXT,
  status TEXT NOT NULL DEFAULT 'queued',
  to_address_json TEXT,
  from_address_json TEXT,
  raw_response TEXT,
  error_message TEXT,
  cost_cents INTEGER NOT NULL DEFAULT 0,
  expected_delivery_date TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS mail_sends_kit_idx ON mail_sends(kit_id);
