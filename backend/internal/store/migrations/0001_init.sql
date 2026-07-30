-- Neptune Radar — core schema (PostgreSQL)
-- Design: explicit tables for ontology entities with real lifecycle/business logic.
-- Low-attribute social edges (follows/tagged_with/mentioned_by) share one generic `edges`
-- table; edges with rich behavior (partner_of, supersedes, supported_by_evidence,
-- associated_with_lead, enrolled_in_case) are realized as foreign keys on the owning table.

CREATE TABLE IF NOT EXISTS persons (
  id TEXT PRIMARY KEY,
  display_name TEXT NOT NULL,
  email TEXT,
  crm_source TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS social_accounts (
  id TEXT PRIMARY KEY,
  person_id TEXT REFERENCES persons(id),
  platform TEXT NOT NULL DEFAULT 'instagram',
  handle TEXT NOT NULL,
  display_name TEXT,
  bio_text TEXT,
  is_private BOOLEAN NOT NULL DEFAULT FALSE,
  is_disabled BOOLEAN NOT NULL DEFAULT FALSE,
  last_seen_at TIMESTAMPTZ,
  UNIQUE(platform, handle)
);

CREATE TABLE IF NOT EXISTS couples (
  id TEXT PRIMARY KEY,
  person_a_id TEXT NOT NULL REFERENCES persons(id),
  person_b_id TEXT NOT NULL REFERENCES persons(id),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- partner_of edge realized as couple membership above; relationship "stage" lives here.
CREATE TABLE IF NOT EXISTS relationships (
  id TEXT PRIMARY KEY,
  couple_id TEXT NOT NULL REFERENCES couples(id),
  stage TEXT NOT NULL CHECK (stage IN (
    'unknown','dating_suspected','engaged','married','status_uncertain','ended_suspected'
  )),
  confidence DOUBLE PRECISION NOT NULL,
  effective_from TIMESTAMPTZ NOT NULL DEFAULT now(),
  effective_to TIMESTAMPTZ,
  superseded_by TEXT REFERENCES relationships(id),
  automation_paused BOOLEAN NOT NULL DEFAULT FALSE,
  visibility_scope TEXT NOT NULL DEFAULT 'neptune_internal' CHECK (visibility_scope IN (
    'private_person_a','private_person_b','shared_couple','neptune_internal','attorney_only','unconfirmed_inference'
  ))
);

CREATE TABLE IF NOT EXISTS social_observations (
  id TEXT PRIMARY KEY,
  monitor TEXT NOT NULL,               -- watch source, e.g. "hashtag:justengaged", "vendor:weddingsbynoor", "profile:maya"
  external_event_id TEXT NOT NULL,     -- provider-native idempotency key
  account_id TEXT REFERENCES social_accounts(id),
  observation_type TEXT NOT NULL,
  raw_payload TEXT NOT NULL,
  observed_at TIMESTAMPTZ NOT NULL,
  ingested_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  source TEXT NOT NULL DEFAULT 'apify',
  freshness_seconds INTEGER,
  consent_scope TEXT NOT NULL,
  UNIQUE(monitor, external_event_id)
);

-- generic low-attribute edges: follows / tagged_with / mentioned_by
CREATE TABLE IF NOT EXISTS edges (
  id TEXT PRIMARY KEY,
  kind TEXT NOT NULL CHECK (kind IN ('follows','tagged_with','mentioned_by')),
  from_account_id TEXT NOT NULL REFERENCES social_accounts(id),
  to_account_id TEXT NOT NULL REFERENCES social_accounts(id),
  active BOOLEAN NOT NULL DEFAULT TRUE,
  first_observed_at TIMESTAMPTZ NOT NULL,
  last_observed_at TIMESTAMPTZ NOT NULL,
  source_observation_id TEXT REFERENCES social_observations(id)
);

CREATE TABLE IF NOT EXISTS life_event_hypotheses (
  id TEXT PRIMARY KEY,
  couple_id TEXT REFERENCES couples(id),
  person_id TEXT REFERENCES persons(id),
  event_type TEXT NOT NULL,
  proposed_stage TEXT,
  confidence DOUBLE PRECISION NOT NULL,
  -- Event-first discovery splits "did something engagement-shaped happen" from
  -- "did we identify the right two people" as two separately-reported scores —
  -- a strong caption tagging the wrong person must not create a prospect just
  -- because the language was convincing. NULL for non-engagement hypotheses.
  engagement_confidence DOUBLE PRECISION,
  partner_confidence DOUBLE PRECISION,
  model_or_rule TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'unconfirmed' CHECK (status IN ('unconfirmed','corroborating','confirmed','rejected','expired')),
  visibility_scope TEXT NOT NULL DEFAULT 'unconfirmed_inference' CHECK (visibility_scope IN (
    'private_person_a','private_person_b','shared_couple','neptune_internal','attorney_only','unconfirmed_inference'
  )),
  consent_scope TEXT NOT NULL,
  expires_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- supported_by_evidence edge realized as evidence.hypothesis_id FK
CREATE TABLE IF NOT EXISTS evidence (
  id TEXT PRIMARY KEY,
  hypothesis_id TEXT REFERENCES life_event_hypotheses(id),
  observation_id TEXT REFERENCES social_observations(id),
  kind TEXT NOT NULL,
  description TEXT NOT NULL,
  weight DOUBLE PRECISION NOT NULL,
  confirmed BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS consent_policies (
  id TEXT PRIMARY KEY,
  person_id TEXT NOT NULL REFERENCES persons(id),
  scope TEXT NOT NULL CHECK (scope IN (
    'private_person_a','private_person_b','shared_couple','neptune_internal','attorney_only','unconfirmed_inference'
  )),
  allowed_actions TEXT NOT NULL,
  granted_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  revoked_at TIMESTAMPTZ
);

-- associated_with_lead edge realized as crm_leads.hypothesis_id FK
CREATE TABLE IF NOT EXISTS crm_leads (
  id TEXT PRIMARY KEY,
  person_id TEXT NOT NULL REFERENCES persons(id),
  hypothesis_id TEXT REFERENCES life_event_hypotheses(id),
  lead_type TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'new' CHECK (status IN ('new','reviewed','investigated','converted','ignored')),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- enrolled_in_case edge realized as neptune_cases.lead_id / couple_id FKs
CREATE TABLE IF NOT EXISTS neptune_cases (
  id TEXT PRIMARY KEY,
  couple_id TEXT REFERENCES couples(id),
  lead_id TEXT REFERENCES crm_leads(id),
  case_type TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'intake' CHECK (status IN ('intake','active','paused','review','closed')),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS recommended_actions (
  id TEXT PRIMARY KEY,
  hypothesis_id TEXT REFERENCES life_event_hypotheses(id),
  case_id TEXT REFERENCES neptune_cases(id),
  action_type TEXT NOT NULL CHECK (action_type IN (
    'review','ignore','draft_outreach','pause_automation','create_case','concierge_review','investigate','no_action'
  )),
  proposed_payload TEXT,
  status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','approved','ignored','executed','failed')),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  decided_at TIMESTAMPTZ,
  decided_by TEXT
);

CREATE TABLE IF NOT EXISTS executed_actions (
  id TEXT PRIMARY KEY,
  recommended_action_id TEXT NOT NULL REFERENCES recommended_actions(id),
  result TEXT NOT NULL CHECK (result IN ('success','failure')),
  detail TEXT,
  verified BOOLEAN NOT NULL DEFAULT FALSE,
  executed_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS audit_events (
  id TEXT PRIMARY KEY,
  entity_type TEXT NOT NULL,
  entity_id TEXT NOT NULL,
  event TEXT NOT NULL,
  detail TEXT,
  monitor TEXT,
  step_index INTEGER,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_obs_monitor ON social_observations(monitor, observed_at);
CREATE INDEX IF NOT EXISTS idx_evidence_hyp ON evidence(hypothesis_id);
CREATE INDEX IF NOT EXISTS idx_hyp_couple ON life_event_hypotheses(couple_id);
CREATE INDEX IF NOT EXISTS idx_audit_entity ON audit_events(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_audit_monitor ON audit_events(monitor, step_index);
CREATE INDEX IF NOT EXISTS idx_edges_pair ON edges(from_account_id, to_account_id, kind);
CREATE INDEX IF NOT EXISTS idx_accounts_handle ON social_accounts(handle);
