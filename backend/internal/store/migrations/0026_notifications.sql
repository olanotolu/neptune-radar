-- Durable in-app notification inbox. SSE events are ephemeral — if you're
-- not watching the dashboard, you miss them. This table gives the team a
-- persistent record of what happened, with read/ack tracking.
CREATE TABLE IF NOT EXISTS notifications (
    id          TEXT PRIMARY KEY,
    type        TEXT NOT NULL,           -- 'action_created', 'stage_transition', 'dlq_item', 'source_stale', 'alert'
    title       TEXT NOT NULL,
    body        TEXT,
    entity_type TEXT,
    entity_id   TEXT,
    severity    TEXT NOT NULL DEFAULT 'info' CHECK (severity IN ('info','warn','critical')),
    read_at     TIMESTAMPTZ,
    acked_at    TIMESTAMPTZ,
    acked_by    TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS notifications_unread_idx
  ON notifications (created_at DESC)
  WHERE read_at IS NULL;
CREATE INDEX IF NOT EXISTS notifications_created_idx
  ON notifications (created_at DESC);
