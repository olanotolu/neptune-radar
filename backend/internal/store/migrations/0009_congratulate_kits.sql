-- Congratulate kits: human-reviewed outreach packages (postcard + address research).
-- Nothing is mailed without status ready_to_mail after a human verifies the address.

CREATE TABLE IF NOT EXISTS congratulate_kits (
  id TEXT PRIMARY KEY,
  couple_id TEXT NOT NULL REFERENCES couples(id),
  status TEXT NOT NULL DEFAULT 'draft' CHECK (status IN (
    'draft', 'ready_review', 'address_verified', 'ready_to_mail', 'mailed', 'cancelled'
  )),
  handle_a TEXT,
  handle_b TEXT,
  person_a_name TEXT,
  person_b_name TEXT,
  bio_a TEXT,
  bio_b TEXT,
  profile_pic_a TEXT,
  profile_pic_b TEXT,
  market_city TEXT,
  market_region TEXT,
  market_source TEXT,
  source_handle TEXT,
  source_class TEXT,
  discovery_caption TEXT,
  discovery_image_url TEXT,
  discovery_post_url TEXT,
  evidence_json TEXT NOT NULL DEFAULT '[]',
  research_notes TEXT,
  research_steps_json TEXT NOT NULL DEFAULT '[]',
  address_line1 TEXT,
  address_line2 TEXT,
  address_city TEXT,
  address_region TEXT,
  address_postal TEXT,
  address_country TEXT NOT NULL DEFAULT 'US',
  address_confidence DOUBLE PRECISION NOT NULL DEFAULT 0,
  address_source TEXT,
  address_candidates_json TEXT NOT NULL DEFAULT '[]',
  headline TEXT,
  body_message TEXT,
  internal_note TEXT,
  postcard_html TEXT,
  mail_payload_json TEXT,
  verified_by TEXT,
  verified_at TIMESTAMPTZ,
  mailed_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS congratulate_kits_couple_idx ON congratulate_kits(couple_id);
CREATE INDEX IF NOT EXISTS congratulate_kits_status_idx ON congratulate_kits(status);
