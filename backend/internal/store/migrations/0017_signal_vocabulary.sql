-- External signal vocabulary: lets ops add new engagement phrases and
-- hashtags without a code redeploy. The signals package loads these at
-- startup and merges them with the hardcoded defaults (which remain the
-- seed and the source of truth for the spec's points table).
-- category: 'explicit_phrase' (caption text), 'high_intent_hashtag',
-- 'bio_phrase', 'negative_phrase'. tier is the penalty/exclusion key for
-- negative phrases (e.g. 'styled_shoot', 'advertisement').

CREATE TABLE IF NOT EXISTS signal_vocabulary (
  id TEXT PRIMARY KEY,
  category TEXT NOT NULL CHECK (category IN ('explicit_phrase','high_intent_hashtag','bio_phrase','negative_phrase','supporting_hashtag')),
  phrase TEXT NOT NULL,
  tier TEXT,  -- for negative_phrase: the penalty key; null otherwise
  enabled BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS signal_vocabulary_unique_idx
  ON signal_vocabulary (category, phrase);
CREATE INDEX IF NOT EXISTS signal_vocabulary_enabled_idx
  ON signal_vocabulary (category) WHERE enabled = TRUE;
