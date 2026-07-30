package signals

import (
	"regexp"
	"sort"
	"strings"
)

// PostSignals is everything the radar can deterministically read from one
// post-shaped payload: the combined signal vocabulary, classified. It is
// pure mechanical fact — no interpretation, no model judgment. (Upstream
// content-understanding models — vendor classification, the vision
// classifier — supply source_account_type and visual_signals. Either way
// they arrive as plain payload fields, and this package classifies them.)
type PostSignals struct {
	Caption          string
	Hashtags         []string // every hashtag, normalized (lowercase, no '#')
	TaggedHandles    []string // accounts tagged in the image, in order
	MentionedHandles []string // accounts @mentioned in the caption, in order
	SourceClass      string   // classified source account type, "" for a personal/unknown account
	Visual           []string // recognized visual/on-screen signals only
	Location         string   // free-text location attached to the post
	Collab           string   // co-author handle on a collaboration post, if any
	RegistryMatch    bool     // a public registry contains the same two names
	Repost           bool     // connector/content classifier flagged old or reposted content
	IdentityConflict bool     // CRM cross-check found conflicting identity evidence

	PhraseHits     []string // explicit caption phrases matched
	HighIntentHits []string
	SupportingHits []string
	StatusHits     []string
	VendorHits     []string
	LocationHits   []string // hashtags that also locate the post in a supported market
	InclusiveHits  []string
	CulturalHits   []string
	WeakHits       []string
	PenaltyKinds   map[string]bool // PenaltyStyledShoot / PenaltyAdvertisement / PenaltyOldReposted
}

var (
	hashtagRe = regexp.MustCompile(`#([A-Za-z0-9_]+)`)
	mentionRe = regexp.MustCompile(`@([A-Za-z0-9_.]+)`)
	// phraseRe matches each explicit phrase on word boundaries — plain
	// substring matching would count "he said yes" inside "she said yes".
	phraseRe = buildPhraseRegexps()
)

func buildPhraseRegexps() map[string]*regexp.Regexp {
	out := map[string]*regexp.Regexp{}
	for _, p := range ExplicitPhrases {
		out[p] = regexp.MustCompile(`(^|[^a-z0-9])` + regexp.QuoteMeta(p) + `($|[^a-z0-9])`)
	}
	return out
}

// ExtractFromPayload reads a post payload (caption, hashtags, tags, source
// class, visual signals, location, linked-evidence flags) and classifies
// every signal against the vocabulary.
func ExtractFromPayload(payload map[string]any) PostSignals {
	sig := PostSignals{PenaltyKinds: map[string]bool{}}
	if payload == nil {
		return sig
	}
	sig.Caption, _ = payload["caption"].(string)
	sig.SourceClass, _ = payload["source_account_type"].(string)
	sig.Location, _ = payload["location"].(string)
	sig.Collab, _ = payload["collab_with"].(string)
	sig.RegistryMatch, _ = payload["registry_match"].(bool)
	sig.Repost, _ = payload["repost"].(bool)
	sig.IdentityConflict, _ = payload["identity_conflict"].(bool)
	sig.TaggedHandles = stringList(payload["tags"])
	for _, m := range mentionRe.FindAllStringSubmatch(sig.Caption, -1) {
		sig.MentionedHandles = append(sig.MentionedHandles, m[1])
	}
	// Some providers return @mentions as structured data rather than in the
	// caption string — merge those so the mention path never depends on
	// caption formatting.
	for _, h := range stringList(payload["provider_mentions"]) {
		if !containsHandle(sig.MentionedHandles, h) {
			sig.MentionedHandles = append(sig.MentionedHandles, h)
		}
	}

	seen := map[string]bool{}
	for _, h := range stringList(payload["hashtags"]) {
		sig.addHashtag(normalizeHashtag(h), seen)
	}
	for _, m := range hashtagRe.FindAllStringSubmatch(sig.Caption, -1) {
		sig.addHashtag(normalizeHashtag(m[1]), seen)
	}

	caption := strings.ToLower(sig.Caption)
	for _, p := range ExplicitPhrases {
		if phraseRe[p].MatchString(caption) {
			sig.PhraseHits = append(sig.PhraseHits, p)
		}
	}
	for phrase, penalty := range captionNegativePhrases {
		if strings.Contains(caption, phrase) {
			sig.PenaltyKinds[penalty] = true
		}
	}
	for _, v := range stringList(payload["visual_signals"]) {
		v = strings.ToLower(strings.TrimSpace(v))
		if VisualEngagementSignals[v] {
			sig.Visual = append(sig.Visual, v)
		}
	}
	if sig.Repost {
		sig.PenaltyKinds[PenaltyOldReposted] = true
	}
	return sig
}

// ReferencedHandles is the union of image tags and caption @mentions — every
// account this post points at, in order, deduplicated. Per spec §8 the most
// important "tags" may be @mentions, so candidate creation and partner
// identification look at the union; the ontology still records tags and
// mentions as different edge kinds.
func (sig PostSignals) ReferencedHandles() []string {
	out := append([]string{}, sig.TaggedHandles...)
	seen := map[string]bool{}
	for _, h := range out {
		seen[h] = true
	}
	for _, h := range sig.MentionedHandles {
		if !seen[h] {
			out = append(out, h)
			seen[h] = true
		}
	}
	return out
}

// addHashtag classifies one normalized hashtag into exactly one tier.
// Precedence: negative > high-intent (incl. generated location patterns) >
// inclusive > location-vendor > vendor > supporting > relationship-status >
// cultural > weak. A hashtag is never counted in two tiers —
// #EngagementInspiration is a negative (styled content), not supporting,
// despite sitting near both lists in the source taxonomy.
func (sig *PostSignals) addHashtag(h string, seen map[string]bool) {
	if h == "" || seen[h] {
		return
	}
	seen[h] = true
	sig.Hashtags = append(sig.Hashtags, h)
	if penalty, ok := NegativeHashtagPenalties[h]; ok {
		sig.PenaltyKinds[penalty] = true
		return
	}
	if HighIntentHashtags[h] {
		sig.HighIntentHits = append(sig.HighIntentHits, h)
		return
	}
	if InclusiveHashtags[h] {
		// #GayEngagement, #TwoGrooms, #Engayged etc. announce an engagement —
		// they create candidates exactly like any other announcement tag.
		sig.InclusiveHits = append(sig.InclusiveHits, h)
		sig.HighIntentHits = append(sig.HighIntentHits, h)
		return
	}
	if tier, _ := classifyLocationHashtag(h); tier != "" {
		sig.LocationHits = append(sig.LocationHits, h)
		if tier == "vendor" {
			sig.VendorHits = append(sig.VendorHits, h)
		} else {
			sig.HighIntentHits = append(sig.HighIntentHits, h)
		}
		return
	}
	if VendorHashtags[h] {
		sig.VendorHits = append(sig.VendorHits, h)
		return
	}
	if SupportingHashtags[h] {
		sig.SupportingHits = append(sig.SupportingHits, h)
		return
	}
	if RelationshipStatusHashtags[h] {
		sig.StatusHits = append(sig.StatusHits, h)
		return
	}
	if CulturalHashtags[h] {
		// Supporting-only: #Roka/#Sagai/#Mangni can describe a ceremony
		// rather than a brand-new engagement — they strengthen an existing
		// candidate but never create one without culture-specific context.
		sig.CulturalHits = append(sig.CulturalHits, h)
		return
	}
	if WeakHashtags[h] {
		sig.WeakHits = append(sig.WeakHits, h)
		return
	}
}

// classifyLocationHashtag matches generated per-market patterns:
// #[market]Proposal / #[market]Engagement / #[market]Engaged (high-intent)
// and #[market]EngagementPhotographer (vendor).
func classifyLocationHashtag(h string) (tier string, market string) {
	for _, m := range Markets {
		if !strings.HasPrefix(h, m) {
			continue
		}
		switch strings.TrimPrefix(h, m) {
		case "proposal", "engagement", "engaged":
			return "high_intent", m
		case "engagementphotographer":
			return "vendor", m
		}
	}
	return "", ""
}

// HasExplicitLanguage is tier-1 detection: an explicit caption phrase or a
// high-intent/inclusive/location announcement hashtag. This — not "#Love" —
// is what can start an engagement candidate.
func (sig PostSignals) HasExplicitLanguage() bool {
	return len(sig.PhraseHits) > 0 || len(sig.HighIntentHits) > 0
}

// IsKnownVendor reports whether the source account is a classified
// engagement/wedding professional, venue, jeweler, publication, registry
// provider, or boutique.
func (sig PostSignals) IsKnownVendor() bool { return WatchedSourceClasses[sig.SourceClass] }

// HasVisual reports whether a ring/proposal-shaped visual signal was detected.
func (sig PostSignals) HasVisual() bool { return len(sig.Visual) > 0 }

// HasPenalties reports whether any exclusion signal fired.
func (sig PostSignals) HasPenalties() bool { return len(sig.PenaltyKinds) > 0 }

// CreatesCandidate is the analyst's deterministic post filter: explicit
// language, OR a known vendor referencing exactly two clients, OR
// proposal-shaped visuals with at least two people referenced. Supporting,
// status, cultural, and weak tags NEVER create a candidate alone. Exclusion
// penalties do not block creation here — they are scored (−50/−25) so the
// audit trail shows exactly why a tempting-looking post produced nothing.
func (sig PostSignals) CreatesCandidate() bool {
	switch {
	case sig.HasExplicitLanguage():
		return true
	case sig.IsKnownVendor() && len(sig.ReferencedHandles()) == 2:
		return true
	case sig.HasVisual() && len(sig.ReferencedHandles()) >= 2:
		return true
	}
	return false
}

// WorthResolvingIdentity is the event-first gate the orchestrator uses before
// naming a never-before-seen referenced pair as a candidate couple. It is
// CreatesCandidate PLUS a hard exclusion check: ads, styled shoots, and
// reposts must not mint brand-new couple records for models — for identity
// creation, suppression belongs at the gate. (For couples the ontology
// already knows, the same signals flow through scoring instead, where the
// −50/−25 penalties land on the audit trail.)
func (sig PostSignals) WorthResolvingIdentity() bool {
	return sig.CreatesCandidate() && !sig.HasPenalties()
}

// BothPartnersIdentified implements the spec's "+25 both probable partners
// tagged": a third party referenced both accounts, OR one partner authored
// the post and referenced the other, OR the post is a collaboration between
// the two accounts.
func BothPartnersIdentified(sig PostSignals, author, handleA, handleB string) bool {
	refs := sig.ReferencedHandles()
	refA, refB := containsHandle(refs, handleA), containsHandle(refs, handleB)
	switch {
	case refA && refB:
		return true // e.g. a photographer referenced exactly the two clients
	case author == handleA && refB, author == handleB && refA:
		return true // one partner posted, referencing the other
	case sig.Collab != "" && (sig.Collab == handleA || sig.Collab == handleB):
		return true // collaboration post: both accounts co-authored it
	}
	return false
}

// LanguageSummary is a compact human-readable description of the explicit
// language matched, for evidence descriptions and audit detail.
func (sig PostSignals) LanguageSummary() string {
	var parts []string
	for _, p := range sig.PhraseHits {
		parts = append(parts, "phrase \""+p+"\"")
	}
	for _, h := range sig.HighIntentHits {
		parts = append(parts, "#"+h)
	}
	return strings.Join(parts, ", ")
}

// SupportingSummary lists the corroborating-but-never-creating tags, for
// evidence descriptions that show the full picture without scoring it.
func (sig PostSignals) SupportingSummary() string {
	var parts []string
	for _, hits := range [][]string{sig.SupportingHits, sig.StatusHits, sig.VendorHits, sig.CulturalHits, sig.WeakHits} {
		for _, h := range hits {
			parts = append(parts, "#"+h)
		}
	}
	sort.Strings(parts)
	return strings.Join(parts, ", ")
}

func normalizeHashtag(h string) string {
	return strings.ToLower(strings.TrimPrefix(strings.TrimSpace(h), "#"))
}

// stringList accepts BOTH []string (what the mapper and worker actually put
// in payloads) and []any-of-string (what hand-built test payloads and JSON
// round-trips produce). A plain []any assertion silently no-ops on []string —
// that mismatch once killed the entire tag/visual path in production.
func stringList(v any) []string {
	switch list := v.(type) {
	case []string:
		var out []string
		for _, s := range list {
			if s != "" {
				out = append(out, s)
			}
		}
		return out
	case []any:
		var out []string
		for _, item := range list {
			if s, ok := item.(string); ok && s != "" {
				out = append(out, s)
			}
		}
		return out
	}
	return nil
}

func containsHandle(handles []string, handle string) bool {
	for _, h := range handles {
		if h == handle {
			return true
		}
	}
	return false
}
