-- Neptune Radar — ingestion sources, cursors, and provider usage accounting.

-- The curated list of public accounts the radar watches (engagement
-- photographers, proposal planners, venues, jewelers, publications, registry
-- providers, boutiques — the classes in signals.WatchedSourceClasses). Ops
-- manages this list via the API; no redeploy needed to add a vendor.
CREATE TABLE IF NOT EXISTS watched_sources (
  id TEXT PRIMARY KEY,
  handle TEXT NOT NULL UNIQUE,
  source_class TEXT NOT NULL,
  active BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- One cursor per monitor ("hashtag:justengaged", "vendor:weddingsbynoor",
-- "profile:maya"): how far each source has been consumed, so restarts and
-- overlapping runs never re-ingest or miss.
CREATE TABLE IF NOT EXISTS ingest_cursors (
  monitor TEXT PRIMARY KEY,
  last_seen_at TIMESTAMPTZ,
  last_run_at TIMESTAMPTZ,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Provider usage accounting — scraping APIs bill per result, so every fetch
-- is logged and the worker enforces a daily cap before spending more.
CREATE TABLE IF NOT EXISTS api_usage (
  id TEXT PRIMARY KEY,
  provider TEXT NOT NULL,
  monitor TEXT NOT NULL,
  results_count INTEGER NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_api_usage_day ON api_usage(created_at);
