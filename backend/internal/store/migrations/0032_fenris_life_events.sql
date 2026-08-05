-- Fenris Digital Life Events API — licensed data-broker signal that
-- cross-validates Instagram discovery. Two independent signals per couple.

-- Extend the couples source vocabulary to include fenris_life_event.
ALTER TABLE couples DROP CONSTRAINT IF EXISTS couples_source_check;
ALTER TABLE couples
  ADD CONSTRAINT couples_source_check
    CHECK (source IN ('social','marriage_license','fenris_life_event'));

-- Track whether a Fenris life event independently validates this couple.
ALTER TABLE couples ADD COLUMN IF NOT EXISTS fenris_validated BOOLEAN NOT NULL DEFAULT FALSE;
CREATE INDEX IF NOT EXISTS idx_couples_fenris_validated
  ON couples(fenris_validated) WHERE fenris_validated = TRUE;

-- Fenris life events log: one row per ingested event, deduped by external_id.
CREATE TABLE IF NOT EXISTS fenris_life_events (
  id            TEXT PRIMARY KEY,
  external_id   TEXT NOT NULL UNIQUE,
  event_type    TEXT NOT NULL,
  person_name   TEXT NOT NULL,
  household_id  TEXT,
  address       TEXT,
  city          TEXT,
  state         TEXT,
  zip           TEXT,
  event_date    TIMESTAMPTZ NOT NULL,
  confidence    DOUBLE PRECISION NOT NULL DEFAULT 0,
  couple_id     TEXT REFERENCES couples(id) ON DELETE SET NULL,
  cross_validated BOOLEAN NOT NULL DEFAULT FALSE,
  ingested_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_fenris_events_event_date
  ON fenris_life_events(event_date DESC);
CREATE INDEX IF NOT EXISTS idx_fenris_events_state
  ON fenris_life_events(state);
