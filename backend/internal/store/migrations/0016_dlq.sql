-- Dead-letter queue for items that failed provider fetch (retryable) or
-- mapping (unmappable). Replaces the worker's log-and-drop behavior: failed
-- items are now persisted for later inspection/replay instead of being lost.

CREATE TABLE IF NOT EXISTS dlq_items (
  id TEXT PRIMARY KEY,
  source TEXT NOT NULL,           -- "apify:hashtag", "apify:vendor", "apify:profile", "records:trestle", ...
  monitor TEXT,                    -- the monitor that produced the item, if any
  external_id TEXT,                -- provider's item ID if available
  raw_payload TEXT,                -- the raw item JSON that failed to map
  error_message TEXT NOT NULL,
  retries INTEGER NOT NULL DEFAULT 0,
  last_retry_at TIMESTAMPTZ,
  status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'replayed', 'discarded')),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS dlq_pending_idx ON dlq_items (created_at) WHERE status = 'pending';
CREATE INDEX IF NOT EXISTS dlq_source_idx ON dlq_items (source, status);
