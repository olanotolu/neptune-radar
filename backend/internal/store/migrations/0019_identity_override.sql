-- Identity override: a human can mark a couple as mistaken (the two people
-- are NOT actually a couple) or a hypothesis as rejected (the event didn't
-- happen / was misidentified). This is the documented appeals/override path
-- the production gaps list calls out: reciprocal tag/follow + co-tag is
-- deliberately conservative, but it will be wrong sometimes, and a human
-- needs a way to mark a couple/hypothesis as mistaken that the scorer
-- respects permanently.

ALTER TABLE couples ADD COLUMN IF NOT EXISTS mistaken BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE couples ADD COLUMN IF NOT EXISTS mistaken_reason TEXT;
ALTER TABLE couples ADD COLUMN IF NOT EXISTS mistaken_by TEXT;
ALTER TABLE couples ADD COLUMN IF NOT EXISTS mistaken_at TIMESTAMPTZ;

-- Index for finding mistaken couples (the scorer checks this before
-- creating new hypotheses for a couple).
CREATE INDEX IF NOT EXISTS couples_mistaken_idx ON couples (mistaken) WHERE mistaken = TRUE;
