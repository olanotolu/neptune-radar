-- Track last agent scan per watched source for SLA / UI.
ALTER TABLE watched_sources
  ADD COLUMN IF NOT EXISTS last_scanned_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS last_scan_couples INTEGER,
  ADD COLUMN IF NOT EXISTS last_scan_actions INTEGER;
