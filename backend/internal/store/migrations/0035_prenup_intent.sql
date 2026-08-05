-- Prenup Intent Score — LLM-predicted likelihood a couple needs/wants a prenup.
-- Stored on the couple (set once by the prep gate), not recomputed per request.
-- prenup_intent_signals is a JSON array of short signal strings.

ALTER TABLE couples
  ADD COLUMN IF NOT EXISTS prenup_intent_score DOUBLE PRECISION NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS prenup_intent_reason TEXT,
  ADD COLUMN IF NOT EXISTS prenup_intent_signals JSONB;

-- High-intent couples (>0.7) get outreach priority — index for fast filtering.
CREATE INDEX IF NOT EXISTS idx_couples_prenup_intent_score
  ON couples(prenup_intent_score DESC) WHERE prenup_intent_score > 0.7;
