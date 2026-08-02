-- Growth OS: brand-safe journey states + tracked chat handoff for Meet Neptune funnel.
-- Runway/ICP are primarily computed live from bios/captions; journey is durable.

ALTER TABLE couples
  ADD COLUMN IF NOT EXISTS journey_stage TEXT NOT NULL DEFAULT 'detected'
    CHECK (journey_stage IN (
      'detected',        -- saw engagement signal
      'approved',        -- human approved prospect
      'congratulated',   -- postcard/kit path started or mailed
      'invited',         -- soft invite / handoff link issued
      'in_chat',         -- operator marked chat started
      'booked',          -- consult booked
      'closed_won',      -- $5k path closed
      'closed_lost',     -- declined / not a fit
      'do_not_contact'   -- hard stop
    ));

ALTER TABLE couples
  ADD COLUMN IF NOT EXISTS journey_updated_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS handoff_code TEXT,
  ADD COLUMN IF NOT EXISTS handoff_url TEXT,
  ADD COLUMN IF NOT EXISTS handoff_utm TEXT,
  ADD COLUMN IF NOT EXISTS handoff_created_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS couples_journey_stage_idx
  ON couples (journey_stage)
  WHERE suppressed_at IS NULL;

COMMENT ON COLUMN couples.journey_stage IS 'Brand-safe Neptune funnel stage; celebrate before pitch';
COMMENT ON COLUMN couples.handoff_url IS 'Tracked deep link into app.meetneptune.com/chat';
