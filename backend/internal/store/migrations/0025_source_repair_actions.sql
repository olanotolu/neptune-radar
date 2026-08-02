-- Add 'source_repair' to the recommended_actions action_type CHECK constraint
-- so the janitor can create repair tasks for stale/degraded sources. These
-- appear in the work queue alongside prospect actions, so a stale source
-- can't be missed just because nobody opened the Sources tab.
ALTER TABLE recommended_actions
  DROP CONSTRAINT IF EXISTS recommended_actions_action_type_check;
ALTER TABLE recommended_actions
  ADD CONSTRAINT recommended_actions_action_type_check CHECK (action_type IN (
    'review','ignore','draft_outreach','pause_automation','create_case',
    'concierge_review','investigate','no_action','source_repair'
  ));
