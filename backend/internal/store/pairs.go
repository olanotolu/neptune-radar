package store

import (
	"time"
)

// PairCooccurrence is the read model for the spec's PairCooccurrence object:
// how often two accounts have been referenced together, by how many distinct
// source accounts, across what span of time.
type PairCooccurrence struct {
	AccountAID      string
	AccountBID      string
	SharedPosts     int
	DistinctSources int
	FirstSeenAt     time.Time
	LastSeenAt      time.Time
}

// canonicalPair orders the two account ids so (X,Y) and (Y,X) share one row.
func canonicalPair(a, b string) (string, string) {
	if a < b {
		return a, b
	}
	return b, a
}

// RecordPairCooccurrence increments the shared-post counter for a pair of
// accounts referenced together on one post. sourceAccountID is the post's
// author; the first time a given source references the pair, distinct_sources
// goes up. Idempotent per (pair, source) for repeat sightings of the same
// post? No — each call is one post; duplicate observations are suppressed
// upstream (social_observations external_event_id), so every call here is a
// genuinely new shared post.
func (s *Store) RecordPairCooccurrence(accountAID, accountBID, sourceAccountID string, seenAt time.Time) error {
	a, b := canonicalPair(accountAID, accountBID)
	_, err := s.DB.Exec(
		`INSERT INTO pair_cooccurrences (account_a_id, account_b_id, shared_posts, distinct_sources, first_seen_at, last_seen_at)
		 VALUES ($1, $2, 1, 0, $3, $3)
		 ON CONFLICT (account_a_id, account_b_id) DO UPDATE SET
		   shared_posts = pair_cooccurrences.shared_posts + 1,
		   last_seen_at = GREATEST(pair_cooccurrences.last_seen_at, EXCLUDED.last_seen_at),
		   first_seen_at = LEAST(pair_cooccurrences.first_seen_at, EXCLUDED.first_seen_at)`,
		a, b, seenAt.UTC(),
	)
	if err != nil {
		return err
	}
	// distinct_sources is derived from the sources table, then written back,
	// so the counter can never drift from the rows backing it.
	_, err = s.DB.Exec(
		`INSERT INTO pair_cooccurrence_sources (account_a_id, account_b_id, source_account_id, first_seen_at)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (account_a_id, account_b_id, source_account_id) DO NOTHING`,
		a, b, sourceAccountID, seenAt.UTC(),
	)
	if err != nil {
		return err
	}
	_, err = s.DB.Exec(
		`UPDATE pair_cooccurrences p SET distinct_sources = (
		   SELECT COUNT(*) FROM pair_cooccurrence_sources ps
		   WHERE ps.account_a_id = p.account_a_id AND ps.account_b_id = p.account_b_id
		 ) WHERE p.account_a_id = $1 AND p.account_b_id = $2`,
		a, b,
	)
	return err
}

// GetPairCooccurrence returns the current co-occurrence stats for a pair,
// or sql.ErrNoRows if the two accounts have never been referenced together.
func (s *Store) GetPairCooccurrence(accountAID, accountBID string) (PairCooccurrence, error) {
	a, b := canonicalPair(accountAID, accountBID)
	var pc PairCooccurrence
	err := s.DB.QueryRow(
		`SELECT account_a_id, account_b_id, shared_posts, distinct_sources, first_seen_at, last_seen_at
		 FROM pair_cooccurrences WHERE account_a_id = $1 AND account_b_id = $2`,
		a, b,
	).Scan(&pc.AccountAID, &pc.AccountBID, &pc.SharedPosts, &pc.DistinctSources, &pc.FirstSeenAt, &pc.LastSeenAt)
	if err != nil {
		return PairCooccurrence{}, err
	}
	return pc, nil
}
