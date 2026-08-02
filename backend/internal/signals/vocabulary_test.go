package signals

import (
	"strings"
	"testing"
)

// mockVocabReader implements VocabularyReader for testing.
type mockVocabReader struct {
	entries map[string][]VocabEntry
}

func (m mockVocabReader) ListSignalVocabulary(category string) ([]VocabEntry, error) {
	return m.entries[category], nil
}

func TestLoadExternalVocabularyMerges(t *testing.T) {
	// Save originals to restore after test.
	origExplicit := append([]string{}, ExplicitPhrases...)
	origHighIntent := make(map[string]bool, len(HighIntentHashtags))
	for k, v := range HighIntentHashtags {
		origHighIntent[k] = v
	}
	defer func() {
		ExplicitPhrases = origExplicit
		HighIntentHashtags = origHighIntent
		phraseRe = buildPhraseRegexps()
	}()

	mock := mockVocabReader{
		entries: map[string][]VocabEntry{
			"explicit_phrase": {
				{Category: "explicit_phrase", Phrase: "she said yes to forever"},
				{Category: "explicit_phrase", Phrase: "we're engaged"}, // duplicate of default — should not double-add
			},
			"high_intent_hashtag": {
				{Category: "high_intent_hashtag", Phrase: "newphrasehashtag"},
			},
		},
	}

	if err := LoadExternalVocabulary(mock); err != nil {
		t.Fatalf("LoadExternalVocabulary: %v", err)
	}

	// New phrase should be added.
	found := false
	for _, p := range ExplicitPhrases {
		if p == "she said yes to forever" {
			found = true
		}
	}
	if !found {
		t.Error("new explicit phrase should be merged into ExplicitPhrases")
	}

	// Duplicate should not be double-added.
	count := 0
	for _, p := range ExplicitPhrases {
		if p == "we're engaged" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("duplicate phrase should not be double-added, got count=%d", count)
	}

	// New hashtag should be added.
	if !HighIntentHashtags["newphrasehashtag"] {
		t.Error("new high-intent hashtag should be merged")
	}

	// The regexp cache should be rebuilt with the new phrase.
	if _, ok := phraseRe["she said yes to forever"]; !ok {
		t.Error("phraseRe should be rebuilt with new phrases")
	}
}

func TestLoadExternalVocabularyNilReader(t *testing.T) {
	// nil reader should be a no-op, not a panic.
	if err := LoadExternalVocabulary(nil); err != nil {
		t.Errorf("nil reader should not error, got: %v", err)
	}
}

func TestLoadExternalVocabularyNegativePhrase(t *testing.T) {
	origNegative := make(map[string]string, len(captionNegativePhrases))
	for k, v := range captionNegativePhrases {
		origNegative[k] = v
	}
	defer func() {
		captionNegativePhrases = origNegative
	}()

	mock := mockVocabReader{
		entries: map[string][]VocabEntry{
			"negative_phrase": {
				{Category: "negative_phrase", Phrase: "flashback friday", Tier: "old_reposted"},
			},
		},
	}

	if err := LoadExternalVocabulary(mock); err != nil {
		t.Fatalf("LoadExternalVocabulary: %v", err)
	}

	if captionNegativePhrases["flashback friday"] != "old_reposted" {
		t.Error("negative phrase with tier should be merged")
	}
}

func TestExplicitPhrasesAreLowercase(t *testing.T) {
	// All hardcoded phrases must be lowercase — the matcher lowercases
	// the caption before comparing. An uppercase phrase would never match.
	for _, p := range ExplicitPhrases {
		if p != strings.ToLower(p) {
			t.Errorf("ExplicitPhrase %q is not lowercase — it would never match", p)
		}
	}
}
