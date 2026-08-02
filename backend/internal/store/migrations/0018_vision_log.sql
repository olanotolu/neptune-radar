-- Vision classifier calibration log: records every vision classification
-- call (image URL, labels returned, model used) so we can track the
-- classifier's precision/recall over time. Without this, a mis-calibrated
-- vision model silently degrades engagement scoring and there's no way to
-- audit when it started going wrong.

CREATE TABLE IF NOT EXISTS vision_classifications (
  id TEXT PRIMARY KEY,
  observation_id TEXT,                -- the social observation that triggered the call
  external_event_id TEXT,            -- the provider's post ID
  image_url TEXT NOT NULL,
  labels TEXT NOT NULL,               -- JSON array of returned labels
  model TEXT NOT NULL,               -- which model answered (e.g. "baseten:claude-3.5-sonnet")
  error TEXT,                         -- empty if success, error message if failed
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS vision_classifications_created_idx ON vision_classifications (created_at);
CREATE INDEX IF NOT EXISTS vision_classifications_model_idx ON vision_classifications (model, created_at);
