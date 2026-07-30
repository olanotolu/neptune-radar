-- Integrity hardening (2026-07 audit): uniqueness the code assumed but the
-- schema never enforced. Every constraint below backs an existing SELECT-
-- then-INSERT path that raced under concurrent workers (watch tick + scan
-- jobs + ?sync=1 API scans all call ProcessEvent at once).

-- couples: one row per unordered pair. EnsureCouple canonicalizes order in
-- Go (person_a_id < person_b_id); the schema now makes that a guarantee and
-- forbids self-couples.
DELETE FROM couples a USING couples b
  WHERE a.person_a_id = b.person_a_id AND a.person_b_id = b.person_b_id
    AND a.id > b.id;

ALTER TABLE couples
  ADD CONSTRAINT couples_pair_unique UNIQUE (person_a_id, person_b_id),
  ADD CONSTRAINT couples_no_self_pair CHECK (person_a_id <> person_b_id);

-- edges: one row per (kind, from, to). Duplicates previously diverged —
-- UpsertEdge updated whichever row the planner returned first, so one row
-- could read active=false while its twin still read active=true.
DELETE FROM edges a USING edges b
  WHERE a.kind = b.kind AND a.from_account_id = b.from_account_id
    AND a.to_account_id = b.to_account_id AND a.id > b.id;

ALTER TABLE edges
  ADD CONSTRAINT edges_unique_triple UNIQUE (kind, from_account_id, to_account_id),
  ADD CONSTRAINT edges_no_self_loop CHECK (from_account_id <> to_account_id);

-- evidence: UpsertEvidenceKind assumes one row per (hypothesis, kind).
DELETE FROM evidence a USING evidence b
  WHERE a.hypothesis_id = b.hypothesis_id AND a.kind = b.kind AND a.id > b.id;

ALTER TABLE evidence
  ADD CONSTRAINT evidence_hypothesis_kind_unique UNIQUE (hypothesis_id, kind);

-- One open relationship row per couple: two concurrent transitions could
-- both insert effective_to IS NULL and CurrentRelationship picked one
-- arbitrarily. Partial unique index makes "current" singular.
DELETE FROM relationships a USING relationships b
  WHERE a.couple_id = b.couple_id AND a.effective_to IS NULL
    AND b.effective_to IS NULL AND a.id > b.id;

CREATE UNIQUE INDEX IF NOT EXISTS relationships_one_open_per_couple
  ON relationships(couple_id) WHERE effective_to IS NULL;

-- hypotheses: a hypothesis must be about someone, and the enums-as-text get
-- the same CHECK discipline relationships.stage already had.
ALTER TABLE life_event_hypotheses
  ADD CONSTRAINT hypotheses_subject_required
    CHECK (couple_id IS NOT NULL OR person_id IS NOT NULL),
  ADD CONSTRAINT hypotheses_event_type_check
    CHECK (event_type IN ('engagement', 'relationship_state_change')),
  ADD CONSTRAINT hypotheses_proposed_stage_check
    CHECK (proposed_stage IN (
      'unknown','dating_suspected','engaged','married','status_uncertain','ended_suspected'
    ));

-- Registry link tables: the comments promised one link per org/source; the
-- schema now enforces it.
DELETE FROM church_jurisdictions a USING church_jurisdictions b
  WHERE a.source_organization_id = b.source_organization_id AND a.id > b.id;
ALTER TABLE church_jurisdictions
  ADD CONSTRAINT church_jurisdictions_org_unique UNIQUE (source_organization_id);

DELETE FROM parishes a USING parishes b
  WHERE a.source_organization_id = b.source_organization_id AND a.id > b.id;
ALTER TABLE parishes
  ADD CONSTRAINT parishes_org_unique UNIQUE (source_organization_id);

DELETE FROM social_sources a USING social_sources b
  WHERE a.watched_source_id = b.watched_source_id AND a.id > b.id;
ALTER TABLE social_sources
  ADD CONSTRAINT social_sources_watched_unique UNIQUE (watched_source_id);

-- One active consent policy per (person, scope): "most recent wins" is not
-- an integrity model for a legal-permission table.
DELETE FROM consent_policies a USING consent_policies b
  WHERE a.person_id = b.person_id AND a.scope = b.scope
    AND a.revoked_at IS NULL AND b.revoked_at IS NULL AND a.id > b.id;

CREATE UNIQUE INDEX IF NOT EXISTS consent_one_active_per_scope
  ON consent_policies(person_id, scope) WHERE revoked_at IS NULL;
