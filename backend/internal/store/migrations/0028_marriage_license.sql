-- Marriage License Monitoring — couples discovered via public marriage-license
-- filings (MarriageSignals feed). The license window (filed 30-90 days before
-- the wedding) is the exact prenup-signing window, so we track the filing date,
-- the county, and a predicted wedding date that drives outreach priority.

ALTER TABLE couples
  ADD COLUMN IF NOT EXISTS source TEXT NOT NULL DEFAULT 'social',
  ADD COLUMN IF NOT EXISTS license_county TEXT,
  ADD COLUMN IF NOT EXISTS license_filing_date TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS predicted_wedding_date TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS wedding_date TIMESTAMPTZ;

-- ponytail: source is a free-form text tag, not a FK to a sources table.
-- Ceiling: a normalized source registry would make joins exact; upgrade path =
-- add a source_registry FK + backfill. Until then the CHECK below keeps the
-- vocabulary closed so dashboards can filter without LIKE.
ALTER TABLE couples
  ADD CONSTRAINT couples_source_check
    CHECK (source IN ('social','marriage_license'));

CREATE INDEX IF NOT EXISTS idx_couples_source ON couples(source);
CREATE INDEX IF NOT EXISTS idx_couples_license_filing_date
  ON couples(license_filing_date DESC) WHERE source = 'marriage_license';
