-- God-tier upgrades: business address, priority score, follow-up sequence.

-- Store Instagram business profile street address (free data we already parse)
ALTER TABLE social_accounts
  ADD COLUMN IF NOT EXISTS business_street TEXT,
  ADD COLUMN IF NOT EXISTS business_city TEXT,
  ADD COLUMN IF NOT EXISTS business_state TEXT,
  ADD COLUMN IF NOT EXISTS business_postal TEXT;

-- Kit priority score (combines confidence, name completeness, recency, vendor quality)
ALTER TABLE congratulate_kits
  ADD COLUMN IF NOT EXISTS priority_score DOUBLE PRECISION NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS follow_up_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS follow_up_template TEXT,
  ADD COLUMN IF NOT EXISTS follow_up_sent_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS follow_up_count INTEGER NOT NULL DEFAULT 0;

-- Index for operator queue sorting by priority
CREATE INDEX IF NOT EXISTS kits_priority_idx ON congratulate_kits (priority_score DESC)
  WHERE status = 'ready_review';
