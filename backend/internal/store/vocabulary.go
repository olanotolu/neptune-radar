package store

import "database/sql"

// SignalVocabEntry is one row from the external signal_vocabulary table.
// Ops can add new phrases/hashtags without a redeploy; the signals package
// merges these with the hardcoded defaults at startup.
type SignalVocabEntry struct {
	ID       string
	Category string // "explicit_phrase", "high_intent_hashtag", "bio_phrase", "negative_phrase", "supporting_hashtag"
	Phrase   string
	Tier     string // for negative_phrase: the penalty key; empty otherwise
	Enabled  bool
}

// ListSignalVocabulary returns all enabled vocabulary entries for the given
// category. The signals package calls this at startup to augment the
// hardcoded defaults.
func (s *Store) ListSignalVocabulary(category string) ([]SignalVocabEntry, error) {
	rows, err := s.DB.Query(
		`SELECT id, category, phrase, COALESCE(tier,''), enabled FROM signal_vocabulary WHERE category = $1 AND enabled = TRUE`,
		category,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []SignalVocabEntry
	for rows.Next() {
		var e SignalVocabEntry
		if err := rows.Scan(&e.ID, &e.Category, &e.Phrase, &e.Tier, &e.Enabled); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// AddSignalVocabulary inserts a new vocabulary entry. Returns the entry with
// its generated ID. Caller is responsible for not duplicating existing
// hardcoded phrases (the unique index will catch exact dupes).
func (s *Store) AddSignalVocabulary(category, phrase, tier string) (SignalVocabEntry, error) {
	e := SignalVocabEntry{
		ID:       NewID("sig"),
		Category: category,
		Phrase:   phrase,
		Tier:     tier,
		Enabled:  true,
	}
	_, err := s.DB.Exec(
		`INSERT INTO signal_vocabulary (id, category, phrase, tier) VALUES ($1, $2, $3, $4) ON CONFLICT (category, phrase) DO UPDATE SET enabled = TRUE`,
		e.ID, e.Category, e.Phrase, sql.NullString{String: tier, Valid: tier != ""},
	)
	return e, err
}

// vocabAdapter adapts *Store to the signals.VocabularyReader interface
// without the signals package importing store. The adapter converts
// store.SignalVocabEntry to signals.VocabEntry.
type vocabAdapter struct{ s *Store }

func (a vocabAdapter) ListSignalVocabulary(category string) ([]any, error) {
	entries, err := a.s.ListSignalVocabulary(category)
	if err != nil {
		return nil, err
	}
	out := make([]any, len(entries))
	for i, e := range entries {
		out[i] = e
	}
	return out, nil
}
