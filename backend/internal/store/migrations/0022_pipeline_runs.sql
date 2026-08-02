-- pipeline_runs is the index row for one ProcessEvent execution: it stores
-- the cross-cutting summary (agent, model, tokens, cost, confidence, stop
-- reason, outcome) that doesn't belong in any single stage table. The
-- per-stage detail still lives in audit_events and pipeline_timings (both
-- keyed by observation_id); this row makes "show me the run" a single query
-- instead of a four-table join.
CREATE TABLE IF NOT EXISTS pipeline_runs (
    id              TEXT PRIMARY KEY,
    observation_id  TEXT NOT NULL,
    agent_name      TEXT NOT NULL DEFAULT 'orchestrator',
    model           TEXT,
    prompt_tokens   INTEGER NOT NULL DEFAULT 0,
    completion_tokens INTEGER NOT NULL DEFAULT 0,
    cost_usd        NUMERIC(12,6) NOT NULL DEFAULT 0,
    confidence      REAL,
    stop_reason     TEXT NOT NULL DEFAULT 'completed',
    hypothesis_id   TEXT,
    action_id       TEXT,
    couple_id       TEXT,
    monitor         TEXT,
    started_at      TIMESTAMPTZ NOT NULL,
    ended_at        TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS pipeline_runs_obs_idx ON pipeline_runs (observation_id);
CREATE INDEX IF NOT EXISTS pipeline_runs_created_idx ON pipeline_runs (created_at DESC);
CREATE INDEX IF NOT EXISTS pipeline_runs_couple_idx ON pipeline_runs (couple_id);
