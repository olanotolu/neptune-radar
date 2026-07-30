-- Neptune Radar — real-world source registry: the geographic sensor network
-- (government, church, and social/wedding-industry sources) that will feed
-- the ontology above. Read-only visualization phase: connectors here check
-- reachability/structure, they do not yet bulk-extract records. `status` on
-- `connectors` must never be set to anything but 'setup' except by a real
-- check recording its real result in `connector_runs` — there is no code
-- path that writes 'healthy' without a corresponding successful run.

-- Geography backbone. Map *rendering* geometry (state/county polygons) stays
-- client-side via us-atlas, unchanged — these tables only relate real
-- institutions/businesses to a real place.
CREATE TABLE IF NOT EXISTS states (
  id TEXT PRIMARY KEY,              -- USPS abbreviation, e.g. 'OH'
  name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS counties (
  id TEXT PRIMARY KEY,              -- 5-digit FIPS, e.g. '39049' = Franklin County, OH
  state_id TEXT NOT NULL REFERENCES states(id),
  name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS cities (
  id TEXT PRIMARY KEY,
  state_id TEXT NOT NULL REFERENCES states(id),
  primary_county_id TEXT REFERENCES counties(id),
  name TEXT NOT NULL,
  lat DOUBLE PRECISION,
  lng DOUBLE PRECISION
);

-- data_mode keeps fixtures out of the live app by construction: every store
-- query for these tables defaults to data_mode != 'fixture' unless the
-- server is started with ENABLE_DEMO_FIXTURES=true. 'manual' means a human
-- (or the bootstrap script, on a human's behalf) confirmed the real-world
-- fact; 'verified_import' means it was pulled live from an authoritative
-- source (e.g. the diocese's own directory) at bootstrap time.
CREATE TABLE IF NOT EXISTS source_organizations (
  id TEXT PRIMARY KEY,
  org_type TEXT NOT NULL CHECK (org_type IN ('government_office','diocese','parish','business')),
  name TEXT NOT NULL,
  city_id TEXT REFERENCES cities(id),
  county_id TEXT REFERENCES counties(id),
  official_url TEXT,
  provenance TEXT NOT NULL CHECK (provenance IN (
    'manually_curated','official_directory','official_government_website',
    'official_parish_website','public_business_website'
  )),
  data_mode TEXT NOT NULL DEFAULT 'manual' CHECK (data_mode IN ('live','verified_import','manual','fixture')),
  -- Free-form JSON for facts that vary by org_type and don't warrant a
  -- dedicated column at this scale (e.g. government: phone, coverage dates;
  -- church: deanery). Same pattern as social_observations.raw_payload.
  metadata TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS source_endpoints (
  id TEXT PRIMARY KEY,
  organization_id TEXT NOT NULL REFERENCES source_organizations(id),
  endpoint_type TEXT NOT NULL CHECK (endpoint_type IN ('marriage_record_search','bulletin_archive','parish_directory','social_profile')),
  url TEXT NOT NULL,
  access_method TEXT NOT NULL CHECK (access_method IN (
    'html_search_form','public_index','downloadable_pdf','structured_api','browser_required'
  )),
  is_official BOOLEAN NOT NULL DEFAULT TRUE,
  data_mode TEXT NOT NULL DEFAULT 'manual' CHECK (data_mode IN ('live','verified_import','manual','fixture')),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- status starts 'setup' on every insert and only ever changes as the direct
-- result of a connector_runs row being written — see internal/connectors.
CREATE TABLE IF NOT EXISTS connectors (
  id TEXT PRIMARY KEY,
  source_endpoint_id TEXT NOT NULL REFERENCES source_endpoints(id),
  connector_type TEXT NOT NULL CHECK (connector_type IN ('http_health','bulletin_discovery','apify_instagram')),
  provider TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'setup' CHECK (status IN ('setup','healthy','degraded','offline')),
  schedule TEXT,
  last_checked_at TIMESTAMPTZ,
  last_success_at TIMESTAMPTZ,
  last_failure_at TIMESTAMPTZ,
  error_message TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- One row per actual check execution — the real measurement. Never inserted
-- speculatively; a connector with zero rows here has never been checked.
CREATE TABLE IF NOT EXISTS connector_runs (
  id TEXT PRIMARY KEY,
  connector_id TEXT NOT NULL REFERENCES connectors(id),
  started_at TIMESTAMPTZ NOT NULL,
  completed_at TIMESTAMPTZ,
  status TEXT NOT NULL CHECK (status IN ('success','failure')),
  http_status INTEGER,
  response_time_ms INTEGER,
  structure_signature TEXT,   -- hash of stable page markers, to detect layout drift over time
  error_message TEXT
);
CREATE INDEX IF NOT EXISTS idx_connector_runs_connector ON connector_runs(connector_id, started_at DESC);

CREATE TABLE IF NOT EXISTS church_jurisdictions (
  id TEXT PRIMARY KEY,
  source_organization_id TEXT NOT NULL REFERENCES source_organizations(id),
  jurisdiction_type TEXT NOT NULL CHECK (jurisdiction_type IN ('diocese','archdiocese')),
  hub_city_id TEXT REFERENCES cities(id)
);

CREATE TABLE IF NOT EXISTS parishes (
  id TEXT PRIMARY KEY,
  source_organization_id TEXT NOT NULL REFERENCES source_organizations(id),
  jurisdiction_id TEXT NOT NULL REFERENCES church_jurisdictions(id),
  bulletin_endpoint_id TEXT REFERENCES source_endpoints(id)
);
CREATE INDEX IF NOT EXISTS idx_parishes_jurisdiction ON parishes(jurisdiction_id);

-- Bridges the registry to the REAL, already-working Instagram pipeline —
-- watched_source_id links to the actual row internal/ingest's worker polls,
-- so this is never a second, disconnected "pretend" source of truth.
CREATE TABLE IF NOT EXISTS social_sources (
  id TEXT PRIMARY KEY,
  source_organization_id TEXT NOT NULL REFERENCES source_organizations(id),
  platform TEXT NOT NULL DEFAULT 'instagram',
  category TEXT NOT NULL,     -- one of signals.WatchedSourceClasses
  city_market_id TEXT REFERENCES cities(id),
  manually_verified BOOLEAN NOT NULL DEFAULT FALSE,
  watched_source_id TEXT REFERENCES watched_sources(id)
);

-- Extend the existing (real, already-live) Instagram ingestion table so the
-- map's Instagram layer can filter to a state/city without a parallel table.
ALTER TABLE watched_sources ADD COLUMN IF NOT EXISTS state TEXT;
ALTER TABLE watched_sources ADD COLUMN IF NOT EXISTS city TEXT;
CREATE INDEX IF NOT EXISTS idx_watched_sources_state ON watched_sources(state) WHERE state IS NOT NULL;
