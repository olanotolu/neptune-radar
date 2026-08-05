-- Address reasoning: LLM secondary confidence boost + rationale.
-- After Bayesian fusion ranks candidates, an LLM agent reasons about the top
-- candidates using couple context (bios, geotags, vendor city) and provides a
-- human-readable rationale. The boost is capped at ±0.05 — a tiebreaker, not
-- a replacement for Bayesian fusion.

ALTER TABLE congratulate_kits
  ADD COLUMN IF NOT EXISTS address_reasoning TEXT;
ALTER TABLE congratulate_kits
  ADD COLUMN IF NOT EXISTS address_reasoning_agreement BOOLEAN DEFAULT FALSE;
