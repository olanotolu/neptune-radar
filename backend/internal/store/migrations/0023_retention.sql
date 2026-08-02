-- retention_classes defines how long each entity type is kept before purge.
-- The janitor reads this table and deletes rows older than max_age_days.
-- Default values are conservative; admins can tighten them via the API.
CREATE TABLE IF NOT EXISTS retention_classes (
    entity_type  TEXT PRIMARY KEY,
    max_age_days INTEGER NOT NULL,
    description  TEXT,
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO retention_classes (entity_type, max_age_days, description) VALUES
    ('social_observations', 365, 'Raw social posts/observations — keep 1 year for evidence trail'),
    ('pipeline_timings',   90,  'Per-stage duration metrics — keep 90 days for trend analysis'),
    ('pipeline_runs',      365, 'Run ledger — keep 1 year for audit/explainability'),
    ('dlq_items',           30,  'Dead-letter queue — keep 30 days for retry analysis')
ON CONFLICT (entity_type) DO NOTHING;
