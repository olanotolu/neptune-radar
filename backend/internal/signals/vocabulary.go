// Package signals is Neptune Radar's signal vocabulary: the deterministic,
// model-free taxonomy the Watchtower pipeline uses to read a post. In 2026
// Instagram caps posts at five hashtags and increasingly relies on captions,
// on-screen text, and content understanding — so this package watches a
// COMBINED vocabulary: hashtags (tiered by intent), caption phrases, account
// tags, source-account classes, visual/on-screen signals, locations, and
// linked evidence. Hashtags alone never decide anything; the vocabulary only
// finds possible moments — the ontology and scorer decide whether a moment is
// a real couple, a real engagement, and a real Neptune opportunity.
//
// This package imports nothing internal: analyst, scorer, and the llm
// template fallback all share it as the single source of truth, so "worth
// resolving identity for", "counts as an engagement candidate", and "counts
// as explicit engagement language" can never drift apart.
package signals

import (
	"sort"
	"strings"
)

// Points are the engagement-prospect scoring weights, taken directly from the
// product spec. Stored on Evidence rows as weight = points/100 so the
// existing 0–1 confidence plumbing stays intact; scorer.ProspectScore turns
// them back into points for the 90/70 policy tiers.
const (
	PtsExplicitLanguage    = 40  // explicit engagement language (caption phrase, high-intent hashtag, fiancé bio)
	PtsBothPartnersTagged  = 25  // both probable partners identified on the post (co-tagged, author+tag, or collab)
	PtsKnownVendorSource   = 15  // known photographer/planner/venue/jeweler/publication/registry/boutique source
	PtsVisualRing          = 10  // ring or proposal visually detected (supports the caption, never identifies people)
	PtsReciprocalEvidence  = 10  // reciprocal relationship evidence (mutual tag/follow)
	PtsRegistryMatch       = 15  // public registry with matching names
	PtsRecentPost          = 10  // fresh, original post
	PtsStyledShoot         = -50 // styled/editorial shoot or vendor inspiration content
	PtsAdvertisement       = -50 // advertisement, sponsorship, or giveaway
	PtsNoSecondPerson      = -30 // no identifiable second person
	PtsOldReposted         = -25 // old or reposted content
	PtsConflictingIdentity = -40 // conflicting identity evidence
	PtsRepeatedCooccurrence = 10 // pair keeps appearing together across independent source accounts
	PtsVendorInPair        = -30 // one account in the proposed pair is itself a vendor — wrong role
)

// --- Tier 1: high-intent engagement hashtags --------------------------------
// These can CREATE an engagement candidate on their own.
var HighIntentHashtags = map[string]bool{
	"justengaged": true, "newlyengaged": true, "weareengaged": true,
	"officiallyengaged": true, "engaged": true, "engagedcouple": true,
	"engagementannouncement": true, "isaidyes": true, "shesaidyes": true,
	"hesaidyes": true, "theysaidyes": true, "heproposed": true,
	"sheproposed": true, "saidyes": true, "justsaidyes": true,
	"heaskedisaidyes": true, "marriageproposal": true, "proposalstory": true,
	"howheasked": true, "howtheyasked": true, "proposalvideo": true,
	"surpriseproposal": true, "engagementday": true, "engagementstory": true,
}

// --- Tier 2: supporting engagement hashtags ---------------------------------
// These strengthen an existing candidate but must never create one alone —
// a jewelry company uses #EngagementRing without anyone getting engaged.
var SupportingHashtags = map[string]bool{
	"engagement": true, "proposal": true, "engagementring": true,
	"diamondring": true, "ringcheck": true, "putaringonit": true,
	"heputaringonit": true, "sayyestothering": true, "engagementphotos": true,
	"engagementphoto": true, "engagementsession": true, "engagementshoot": true,
	"engagementphotography": true, "engagementparty": true,
	"engagementseason": true, "engagementinspiration": true,
	"proposalphotography": true, "proposalphotos": true,
}

// --- Tier 3: relationship-status hashtags ------------------------------------
// The couple may be entering marriage planning; supporting weight only.
var RelationshipStatusHashtags = map[string]bool{
	"fiance": true, "fiancee": true, "futurehusband": true, "futurewife": true,
	"futurespouse": true, "husbandtobe": true, "wifetobe": true,
	"futuremrs": true, "futuremr": true, "soontobemrs": true,
	"misstomrs": true, "bridetobe": true, "groomtobe": true,
	"gettingmarried": true, "gettinghitched": true, "tietheknot": true,
	"planningstartsnow": true, "weddingplanning": true, "savethedate": true,
}

// --- Tier 4: vendor / photographer hashtags ----------------------------------
// Valuable because photographers and planners often tag BOTH members of the
// couple — the strongest non-biometric couple-resolution path there is.
var VendorHashtags = map[string]bool{
	"engagementphotographer": true, "proposalphotographer": true,
	"proposalplanner": true, "weddingplanner": true, "weddingphotographer": true,
	"couplesphotographer": true, "proposalsetup": true,
}

// WatchedSourceClasses are the public account types the radar classifies and
// weights: a photographer posting "Congratulations to Maya and Jordan" and
// tagging both people is substantially stronger than Maya using #Love. The
// Accounts get their class from the curated watched_sources table (managed
// via the dashboard); the ingest worker stamps it onto each post's payload.
var WatchedSourceClasses = map[string]bool{
	"engagement_photographer": true, "proposal_planner": true,
	"wedding_planner": true, "wedding_venue": true, "jeweler": true,
	"wedding_publication": true, "registry_provider": true,
	"bridal_boutique": true,
}

// --- Tier 5: location vocabulary ---------------------------------------------
// Location affects whether Neptune serves the couple and which legal workflow
// may eventually apply. Patterns are generated per supported market rather
// than enumerated: #[market]Proposal, #[market]Engagement, #[market]Engaged
// are high-intent; #[market]EngagementPhotographer is a vendor tag.
var Markets = []string{
	// Ohio — the actual product market. Columbus first, then the other
	// configured-county metros.
	"columbus", "cbus", "cleveland", "cincinnati", "dayton", "akron",
	"toledo", "delawareohio", "fairfieldcounty",
	// Legacy metro set (kept: the vocabulary is market-agnostic infra).
	"nyc", "newyork", "brooklyn", "manhattan", "centralpark",
	"la", "losangeles", "sf", "sanfrancisco", "chicago", "boston",
	"austin", "miami", "seattle", "dc", "denver", "nashville",
}

// --- Tier 6: inclusive and cultural vocabulary --------------------------------
// The detector must not assume every couple uses "bride" and "groom".
var InclusiveHashtags = map[string]bool{
	"gayengagement": true, "lesbianengagement": true, "queerengagement": true,
	"lgbtqengagement": true, "twobrides": true, "twogrooms": true,
	"engayged": true,
}

// CulturalHashtags are configurable cultural/language packs. They are
// SUPPORTING-only by default: #Roka, #Sagai, and #Mangni can describe a
// ceremony rather than a brand-new engagement, so they need culture-specific
// context before they create anything.
var CulturalHashtags = map[string]bool{
	"roka": true, "sagai": true, "mangni": true, "betrothal": true,
}

// --- Tier 7: weak hashtags that never create a prospect -----------------------
var WeakHashtags = map[string]bool{
	"love": true, "couple": true, "couplegoals": true,
	"relationshipgoals": true, "soulmate": true, "forever": true,
	"myperson": true, "loveofmylife": true, "romance": true, "datenight": true,
	"happy": true, "blessed": true, "wedding": true, "bride": true,
	"groom": true, "ring": true, "diamond": true, "happilyeverafter": true,
	"ido": true, "loveislove": true,
}

// --- Tier 8: negative / exclusion hashtags ------------------------------------
// Mapped to the specific scoring penalty each triggers. These prevent false
// leads: a jewelry ad using #SheSaidYes is suppressed by source type + #Ad,
// not amplified by the high-intent hashtag it borrowed.
var NegativeHashtagPenalties = map[string]string{
	// styled/editorial content → styled_shoot penalty
	"styledshoot": PenaltyStyledShoot, "editorialshoot": PenaltyStyledShoot,
	"bridaleditorial": PenaltyStyledShoot, "weddinginspiration": PenaltyStyledShoot,
	"engagementinspiration": PenaltyStyledShoot, "proposalideas": PenaltyStyledShoot,
	"model": PenaltyStyledShoot, "photoshoot": PenaltyStyledShoot,
	"weddingvendor": PenaltyStyledShoot,
	// ads and giveaways → advertisement penalty
	"giveaway": PenaltyAdvertisement, "sponsored": PenaltyAdvertisement,
	"ad": PenaltyAdvertisement, "jewelryad": PenaltyAdvertisement,
	"ringad": PenaltyAdvertisement,
	// old or repurposed content → old_reposted penalty
	"throwback": PenaltyOldReposted, "tbt": PenaltyOldReposted,
	"anniversary": PenaltyOldReposted, "vowrenewal": PenaltyOldReposted,
}

// Penalty kinds, shared by hashtag penalties and caption-phrase penalties.
const (
	PenaltyStyledShoot   = "styled_shoot"
	PenaltyAdvertisement = "advertisement"
	PenaltyOldReposted   = "old_reposted"
)

// --- Explicit caption phrases -------------------------------------------------
// Captions matter more than hashtags now. Deterministic matching covers the
// canonical phrasings; the model (or its template fallback) judges semantic
// variations — "he asked and I said forever" — so the system never relies on
// exact string matching alone.
var ExplicitPhrases = []string{
	"we're engaged", "we are engaged", "i said yes", "she said yes",
	"he said yes", "they said yes", "will you marry me",
	"the easiest yes", "officially fiancé and fiancée", "officially fiances",
	"officially fiancés", "i get to marry my best friend",
	"can't wait to marry you", "cant wait to marry you", "forever starts now",
	"from boyfriend to fiancé", "from girlfriend to fiancée",
	"meet my future husband", "meet my future wife", "meet my future spouse",
	"went on a walk and came back engaged", "the answer was yes",
	"save the date",
}

// bioPhrases are the account-bio counterpart to ExplicitPhrases (spec §8:
// "does one bio mention the other?" — e.g. @maya's bio says "future Mrs.
// Lee"). A fiancé-reference bio is explicit engagement language.
var bioPhrases = []string{
	"fiance", "fiancé", "fiancée", "engaged to", "future mrs", "future mr",
	"future husband", "future wife", "future spouse", "soon to be mrs",
}

// captionNegativePhrases catch exclusion language written as prose rather
// than as a hashtag ("Throwback to our wedding day", "styled shoot for...").
var captionNegativePhrases = map[string]string{
	"throwback": PenaltyOldReposted, "tbt": PenaltyOldReposted,
	"anniversary": PenaltyOldReposted, "vow renewal": PenaltyOldReposted,
	"styled shoot": PenaltyStyledShoot, "editorial shoot": PenaltyStyledShoot,
	"giveaway": PenaltyAdvertisement, "sponsored": PenaltyAdvertisement,
}

// VisualEngagementSignals are what the vision classifier (a Baseten-hosted
// multimodal model — see internal/llm/vision.go) is constrained to emit.
// Visual evidence supports the caption — it never independently identifies
// the people.
var VisualEngagementSignals = map[string]bool{
	"ring": true, "ring_box": true, "proposal_scene": true,
	"marry_me_sign": true, "engagement_party_signage": true,
	"champagne_celebration": true, "on_screen_text_engaged": true,
	"countdown_screenshot": true, "save_the_date_card": true,
}

// HasExplicitBioLanguage reports whether a bio affirms a fiancé/e engagement.
func HasExplicitBioLanguage(bio string) bool {
	bio = strings.ToLower(bio)
	for _, p := range bioPhrases {
		if strings.Contains(bio, p) {
			return true
		}
	}
	return false
}

// MonitoredHashtags is the watch list for the hashtag monitor: every
// high-intent and inclusive announcement tag, plus the generated high-intent
// location patterns for the active markets. Supporting/weak/negative tags
// are deliberately NOT watched — they classify posts, they don't find them.
// Sorted for deterministic config/diffs.
func MonitoredHashtags(activeMarkets []string) []string {
	set := map[string]bool{}
	for h := range HighIntentHashtags {
		set[h] = true
	}
	for h := range InclusiveHashtags {
		set[h] = true
	}
	for _, m := range activeMarkets {
		m = strings.ToLower(strings.TrimSpace(m))
		if m == "" {
			continue
		}
		set[m+"proposal"] = true
		set[m+"engagement"] = true
		set[m+"engaged"] = true
	}
	out := make([]string, 0, len(set))
	for h := range set {
		out = append(out, h)
	}
	sort.Strings(out)
	return out
}

// VisualSignalLabels returns the recognized visual-signal vocabulary, sorted —
// the label list the vision classifier is constrained to emit.
func VisualSignalLabels() []string {
	out := make([]string, 0, len(VisualEngagementSignals))
	for v := range VisualEngagementSignals {
		out = append(out, v)
	}
	sort.Strings(out)
	return out
}
