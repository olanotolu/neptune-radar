-- Relationship Strength Score — LLM-predicted strength/seriousness of a couple's
-- relationship, augmenting the existing FAIR dispersion metric. Stored on the couple
-- (set once by the ingest worker), not recomputed per request.
-- relationship_strength_signals is a JSON array of short signal strings.

ALTER TABLE couples
  ADD COLUMN IF NOT EXISTS relationship_strength_score FLOAT,
  ADD COLUMN IF NOT EXISTS relationship_strength_category VARCHAR(32),
  ADD COLUMN IF NOT EXISTS relationship_strength_signals JSONB,
  ADD COLUMN IF NOT EXISTS relationship_strength_rationale TEXT;
