CREATE TABLE IF NOT EXISTS pipeline_timings (
    id TEXT PRIMARY KEY,
    stage TEXT NOT NULL,
    duration_ms BIGINT NOT NULL,
    event_id TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS pipeline_timings_stage_idx ON pipeline_timings (stage, created_at DESC);
