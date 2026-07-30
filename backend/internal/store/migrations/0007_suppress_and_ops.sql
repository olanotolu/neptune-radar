-- Operator suppress ("not a couple") + indexes for workbench queries.

ALTER TABLE couples
  ADD COLUMN IF NOT EXISTS suppressed_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS suppressed_reason TEXT;

CREATE INDEX IF NOT EXISTS idx_couples_active_board
  ON couples (created_at DESC)
  WHERE suppressed_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_social_accounts_person_pic
  ON social_accounts (person_id)
  WHERE profile_pic_url IS NOT NULL AND profile_pic_url <> '';
