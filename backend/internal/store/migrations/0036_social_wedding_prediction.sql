-- Social Wedding Date Prediction — LLM-inferred wedding date from Instagram
-- signals (captions, bios, venue/vendor tags). Distinct from the marriage-
-- license predicted_wedding_date: the social prediction is lower-authority and
-- never overwrites a license prediction. These columns record the rationale +
-- confidence so the dashboard can show the source of a predicted wedding date.

ALTER TABLE couples
  ADD COLUMN IF NOT EXISTS social_wedding_prediction TEXT,
  ADD COLUMN IF NOT EXISTS social_wedding_confidence DOUBLE PRECISION NOT NULL DEFAULT 0;

-- ponytail: a partial index keeps the predictions view cheap — only couples
-- that actually have a predicted wedding date from any source are scanned.
CREATE INDEX IF NOT EXISTS idx_couples_predicted_wedding_date
  ON couples(predicted_wedding_date ASC) WHERE predicted_wedding_date IS NOT NULL;
