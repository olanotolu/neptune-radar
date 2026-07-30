-- Pair co-occurrence: the spec's PairCooccurrence object. Every time a post
-- references two accounts together (tags or caption mentions), one row tracks
-- the running total — how many shared posts, from how many DISTINCT source
-- accounts, over what date range, and whether the references are mutual.
-- This is what separates "two people in one photographer post" from "two
-- people who keep appearing together across independent sources" — the
-- +10 repeated-co-occurrence evidence reads from here, never recomputed
-- ad hoc from raw observations.
--
-- account_a_id < account_b_id always (canonical order), so the pair
-- (X,Y) and (Y,X) share one row.
CREATE TABLE IF NOT EXISTS pair_cooccurrences (
  account_a_id TEXT NOT NULL REFERENCES social_accounts(id),
  account_b_id TEXT NOT NULL REFERENCES social_accounts(id),
  shared_posts INTEGER NOT NULL DEFAULT 0,
  distinct_sources INTEGER NOT NULL DEFAULT 0,
  first_seen_at TIMESTAMPTZ NOT NULL,
  last_seen_at TIMESTAMPTZ NOT NULL,
  PRIMARY KEY (account_a_id, account_b_id)
);

-- Which source accounts have referenced this pair — distinct_sources on the
-- parent row is maintained from this table (never trust the counter alone).
CREATE TABLE IF NOT EXISTS pair_cooccurrence_sources (
  account_a_id TEXT NOT NULL,
  account_b_id TEXT NOT NULL,
  source_account_id TEXT NOT NULL REFERENCES social_accounts(id),
  first_seen_at TIMESTAMPTZ NOT NULL,
  PRIMARY KEY (account_a_id, account_b_id, source_account_id),
  FOREIGN KEY (account_a_id, account_b_id)
    REFERENCES pair_cooccurrences(account_a_id, account_b_id)
);
