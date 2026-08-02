// Package vocab provides the adapter between the store (which owns the
// signal_vocabulary table) and the signals package (which owns the
// in-memory vocabulary). This package exists so neither store nor signals
// imports the other — signals stays dependency-free, store stays
// entity-scoped.
package vocab

import (
	"neptune-social-radar/backend/internal/signals"
	"neptune-social-radar/backend/internal/store"
)

// StoreAdapter wraps *store.Store to implement signals.VocabularyReader.
type StoreAdapter struct{ S *store.Store }

func (a StoreAdapter) ListSignalVocabulary(category string) ([]signals.VocabEntry, error) {
	entries, err := a.S.ListSignalVocabulary(category)
	if err != nil {
		return nil, err
	}
	out := make([]signals.VocabEntry, len(entries))
	for i, e := range entries {
		out[i] = signals.VocabEntry{
			Category: e.Category,
			Phrase:   e.Phrase,
			Tier:     e.Tier,
		}
	}
	return out, nil
}

// LoadFromStore is the one-line call main.go makes at startup to merge
// DB-driven vocabulary into the in-memory defaults.
func LoadFromStore(s *store.Store) error {
	return signals.LoadExternalVocabulary(StoreAdapter{S: s})
}
