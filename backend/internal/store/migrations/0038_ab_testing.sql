-- A/B testing infrastructure: variant assignment + LLM personalization tracking.
-- Each kit records which experiment variant it was assigned to, so conversion
-- (scan → chat) can be attributed back to the postcard copy variant.

ALTER TABLE congratulate_kits
  ADD COLUMN IF NOT EXISTS variant_id VARCHAR(32);
ALTER TABLE congratulate_kits
  ADD COLUMN IF NOT EXISTS experiment_id VARCHAR(64);
ALTER TABLE congratulate_kits
  ADD COLUMN IF NOT EXISTS is_personalized BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE congratulate_kits
  ADD COLUMN IF NOT EXISTS personalized_copy TEXT;
