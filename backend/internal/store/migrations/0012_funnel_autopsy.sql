-- Closed-loop funnel events (Meet Neptune app → Radar) + weekly false-positive autopsy.

CREATE TABLE IF NOT EXISTS funnel_events (
  id TEXT PRIMARY KEY,
  couple_id TEXT REFERENCES couples(id),
  event_type TEXT NOT NULL CHECK (event_type IN (
    'chat_started',
    'consult_booked',
    'closed_won',
    'closed_lost',
    'handoff_clicked'  -- optional intermediate from marketing
  )),
  external_id TEXT,                 -- idempotency from source system
  handoff_code TEXT,
  source TEXT NOT NULL DEFAULT 'webhook',
  payload_json TEXT NOT NULL DEFAULT '{}',
  journey_stage_before TEXT,
  journey_stage_after TEXT,
  matched_by TEXT,                  -- couple_id | handoff_code | utm_content | unresolved
  occurred_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS funnel_events_external_unique
  ON funnel_events (source, external_id)
  WHERE external_id IS NOT NULL AND external_id <> '';

CREATE INDEX IF NOT EXISTS funnel_events_couple_idx
  ON funnel_events (couple_id, occurred_at DESC);

CREATE INDEX IF NOT EXISTS funnel_events_type_idx
  ON funnel_events (event_type, occurred_at DESC);

CREATE TABLE IF NOT EXISTS autopsy_reports (
  id TEXT PRIMARY KEY,
  period_start TIMESTAMPTZ NOT NULL,
  period_end TIMESTAMPTZ NOT NULL,
  summary_json TEXT NOT NULL DEFAULT '{}',
  cases_json TEXT NOT NULL DEFAULT '[]',
  generated_by TEXT NOT NULL DEFAULT 'system',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS autopsy_reports_period_idx
  ON autopsy_reports (period_end DESC);

COMMENT ON TABLE funnel_events IS 'Meet Neptune product events closing the radar → chat → booked loop';
COMMENT ON TABLE autopsy_reports IS 'Weekly false-positive / ignore autopsy for legal + model trust';
