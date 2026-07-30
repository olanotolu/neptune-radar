-- Prospect enrichment: profile pics on social accounts, geo on couples.

ALTER TABLE social_accounts
  ADD COLUMN IF NOT EXISTS profile_pic_url TEXT,
  ADD COLUMN IF NOT EXISTS follower_count INTEGER,
  ADD COLUMN IF NOT EXISTS following_count INTEGER,
  ADD COLUMN IF NOT EXISTS profile_checked_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS inferred_city TEXT,
  ADD COLUMN IF NOT EXISTS inferred_region TEXT,
  ADD COLUMN IF NOT EXISTS location_source TEXT;

ALTER TABLE couples
  ADD COLUMN IF NOT EXISTS inferred_city TEXT,
  ADD COLUMN IF NOT EXISTS inferred_region TEXT,
  ADD COLUMN IF NOT EXISTS inferred_lat DOUBLE PRECISION,
  ADD COLUMN IF NOT EXISTS inferred_lng DOUBLE PRECISION,
  ADD COLUMN IF NOT EXISTS location_source TEXT;

CREATE INDEX IF NOT EXISTS idx_couples_inferred_city ON couples (inferred_city)
  WHERE inferred_city IS NOT NULL AND inferred_city <> '';
