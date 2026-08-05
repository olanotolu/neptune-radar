-- Wedding Website Discovery — self-reported wedding website data scraped from
-- The Knot / Zola / WeddingWire. These columns are the couple's own published
-- details: wedding_website_date is AUTHORITATIVE over predicted_wedding_date
-- (self-reported vs inferred). registry_urls is a JSONB array of registry links.

ALTER TABLE couples
  ADD COLUMN IF NOT EXISTS wedding_website_url TEXT,
  ADD COLUMN IF NOT EXISTS wedding_website_platform VARCHAR(32),
  ADD COLUMN IF NOT EXISTS wedding_website_date DATE,
  ADD COLUMN IF NOT EXISTS wedding_venue_name TEXT,
  ADD COLUMN IF NOT EXISTS wedding_venue_city VARCHAR(128),
  ADD COLUMN IF NOT EXISTS wedding_venue_state VARCHAR(32),
  ADD COLUMN IF NOT EXISTS registry_urls JSONB;
