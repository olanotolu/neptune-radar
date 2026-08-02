-- Action ownership: priority, assignee, snooze, and escalation metadata.
-- These columns let the concierge team own, prioritize, and defer work
-- items instead of treating the action queue as a flat unsorted list.
ALTER TABLE recommended_actions
  ADD COLUMN IF NOT EXISTS priority INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS owner TEXT,
  ADD COLUMN IF NOT EXISTS snooze_until TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS reason TEXT;

-- Index for "my queue" — pending actions assigned to a specific user.
CREATE INDEX IF NOT EXISTS idx_actions_owner
  ON recommended_actions (owner, status)
  WHERE status = 'pending';

-- Index for priority-sorted queue — highest priority first.
CREATE INDEX IF NOT EXISTS idx_actions_priority
  ON recommended_actions (priority DESC, created_at)
  WHERE status = 'pending';
