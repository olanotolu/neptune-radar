CREATE TABLE IF NOT EXISTS interview_sessions (
    id TEXT PRIMARY KEY,
    couple_a_label TEXT NOT NULL DEFAULT 'Couple A',
    couple_b_label TEXT NOT NULL DEFAULT 'Couple B',
    status TEXT NOT NULL DEFAULT 'active', -- active | completed
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS interview_messages (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL REFERENCES interview_sessions(id) ON DELETE CASCADE,
    speaker TEXT NOT NULL, -- "a1", "a2", "b1", "b2" (person 1 or 2 of couple A or B)
    couple TEXT NOT NULL, -- "A" or "B"
    text TEXT NOT NULL,
    audio_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_interview_messages_session ON interview_messages(session_id, created_at);

CREATE TABLE IF NOT EXISTS interview_extractions (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL REFERENCES interview_sessions(id) ON DELETE CASCADE,
    agent_type TEXT NOT NULL, -- "relationship_stage", "wedding_timeline", "vendor_interest", "location", "budget"
    findings JSONB NOT NULL DEFAULT '{}',
    confidence FLOAT NOT NULL DEFAULT 0,
    summary TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_inttractions_session ON interview_extractions(session_id, created_at);
