package signals

import "testing"

func payload(kv ...any) map[string]any {
	m := map[string]any{}
	for i := 0; i < len(kv); i += 2 {
		m[kv[i].(string)] = kv[i+1]
	}
	return m
}

func TestExplicitCaptionPhrase_Alone_CreatesCandidate(t *testing.T) {
	// The 2026 premise: no hashtags at all, the caption carries the signal.
	sig := ExtractFromPayload(payload("caption", "I get to marry my best friend"))
	if !sig.HasExplicitLanguage() {
		t.Fatal("expected explicit language from caption phrase alone")
	}
	if !sig.CreatesCandidate() {
		t.Fatal("expected a caption-only explicit post to create a candidate")
	}
	if len(sig.Hashtags) != 0 {
		t.Errorf("expected no hashtags, got %v", sig.Hashtags)
	}
}

func TestHighIntentHashtag_CreatesCandidate(t *testing.T) {
	for _, h := range []string{"#JustEngaged", "#SheSaidYes", "#TheySaidYes", "#HeProposed", "shesaidyes"} {
		sig := ExtractFromPayload(payload("hashtags", []any{h}))
		if !sig.CreatesCandidate() {
			t.Errorf("expected %s to create a candidate", h)
		}
	}
}

func TestSupportingAndWeakTags_NeverCreateAlone(t *testing.T) {
	for _, h := range []string{"#EngagementRing", "#Proposal", "#EngagementPhotos"} {
		if sig := ExtractFromPayload(payload("hashtags", []any{h})); sig.CreatesCandidate() {
			t.Errorf("supporting tag %s must not create a candidate alone", h)
		}
	}
	for _, h := range []string{"#Love", "#CoupleGoals", "#Wedding", "#Bride", "#Ring"} {
		if sig := ExtractFromPayload(payload("hashtags", []any{h})); sig.CreatesCandidate() {
			t.Errorf("weak tag %s must never create a candidate", h)
		}
	}
	// Even a pile of weak/supporting tags together stays below the bar.
	sig := ExtractFromPayload(payload("hashtags", []any{"#Love", "#CoupleGoals", "#EngagementRing", "#Fiance"}))
	if sig.CreatesCandidate() {
		t.Error("supporting + status + weak tags combined must not create a candidate")
	}
	if len(sig.WeakHits) != 2 || len(sig.SupportingHits) != 1 || len(sig.StatusHits) != 1 {
		t.Errorf("classification drift: %+v", sig)
	}
}

func TestCulturalTags_AreSupportingOnly(t *testing.T) {
	sig := ExtractFromPayload(payload("hashtags", []any{"#Roka", "#Sagai"}))
	if sig.CreatesCandidate() {
		t.Error("cultural ceremony tags must not create a candidate without more context")
	}
	if len(sig.CulturalHits) != 2 {
		t.Errorf("expected cultural hits, got %+v", sig.CulturalHits)
	}
}

func TestInclusiveTags_CreateLikeAnyAnnouncement(t *testing.T) {
	for _, h := range []string{"#TwoGrooms", "#GayEngagement", "#Engayged", "#LGBTQEngagement"} {
		if sig := ExtractFromPayload(payload("hashtags", []any{h})); !sig.CreatesCandidate() {
			t.Errorf("inclusive announcement tag %s must create a candidate", h)
		}
	}
}

func TestLocationPatterns_GeneratedPerMarket(t *testing.T) {
	cases := map[string]string{
		"#NYCProposal":                   "high_intent",
		"#BrooklynEngagement":            "high_intent",
		"#CentralParkProposal":           "high_intent",
		"#MiamiEngaged":                  "high_intent",
		"#NYCEngagementPhotographer":     "vendor",
		"#ChicagoEngagementPhotographer": "vendor",
	}
	for h, want := range cases {
		sig := ExtractFromPayload(payload("hashtags", []any{h}))
		switch want {
		case "high_intent":
			if !sig.CreatesCandidate() {
				t.Errorf("location tag %s should be high-intent and create a candidate", h)
			}
		case "vendor":
			if len(sig.VendorHits) != 1 {
				t.Errorf("location tag %s should classify as vendor, got %+v", h, sig)
			}
		}
		if len(sig.LocationHits) != 1 {
			t.Errorf("location tag %s should be recorded as a location hit", h)
		}
	}
}

func TestVendorTaggingExactlyTwo_CreatesCandidate(t *testing.T) {
	sig := ExtractFromPayload(payload(
		"caption", "Congratulations to the happy couple!",
		"source_account_type", "engagement_photographer",
		"tags", []any{"maya", "jordan"},
	))
	if !sig.CreatesCandidate() {
		t.Fatal("a known photographer tagging exactly two clients must create a candidate")
	}
	// A vendor post referencing three people is ambiguous — no candidate.
	sig3 := ExtractFromPayload(payload(
		"source_account_type", "wedding_planner",
		"tags", []any{"a", "b", "c"},
	))
	if sig3.CreatesCandidate() {
		t.Error("vendor post with three referenced accounts is ambiguous — must not create a candidate")
	}
	// Unknown source class is not a vendor path.
	sigU := ExtractFromPayload(payload(
		"source_account_type", "random_account",
		"tags", []any{"maya", "jordan"},
	))
	if sigU.CreatesCandidate() {
		t.Error("unclassified source must not create a candidate via the vendor path")
	}
}

func TestVisualWithTwoTagged_CreatesCandidate(t *testing.T) {
	sig := ExtractFromPayload(payload(
		"caption", "POV: you just got engaged",
		"visual_signals", []any{"on_screen_text_engaged", "ring", "not_a_real_signal"},
		"tags", []any{"maya", "jordan"},
	))
	if !sig.CreatesCandidate() {
		t.Fatal("proposal-shaped visuals with both people tagged must create a candidate")
	}
	if len(sig.Visual) != 2 {
		t.Errorf("unrecognized visual signals must be dropped, got %v", sig.Visual)
	}
	// Visuals alone with no second person referenced: nothing.
	sigNoTags := ExtractFromPayload(payload("visual_signals", []any{"ring"}))
	if sigNoTags.CreatesCandidate() {
		t.Error("a ring photo with nobody referenced must not create a candidate")
	}
}

func TestNegatives_MapToPenalties(t *testing.T) {
	sig := ExtractFromPayload(payload("hashtags", []any{"#SheSaidYes", "#StyledShoot", "#Ad", "#TBT"}))
	if !sig.PenaltyKinds[PenaltyStyledShoot] || !sig.PenaltyKinds[PenaltyAdvertisement] || !sig.PenaltyKinds[PenaltyOldReposted] {
		t.Errorf("expected all three penalty kinds, got %v", sig.PenaltyKinds)
	}
	// The high-intent tag still classifies (so scoring can show its +40 was
	// overwhelmed) — but identity resolution must be refused.
	if !sig.CreatesCandidate() {
		t.Error("penalty-carrying posts still create candidates (scoring applies the −50s)")
	}
	if sig.WorthResolvingIdentity() {
		t.Error("ads/styled shoots must never mint a new couple at the identity gate")
	}
}

func TestCaptionNegativePhrases(t *testing.T) {
	sig := ExtractFromPayload(payload("caption", "Throwback to our wedding day 💍"))
	if !sig.PenaltyKinds[PenaltyOldReposted] {
		t.Error("caption-level 'throwback' should flag old content")
	}
	if sig.CreatesCandidate() {
		t.Error("a throwback with no explicit language must not create a candidate")
	}
}

func TestCaptionMentions_AreReferencesNotTags(t *testing.T) {
	sig := ExtractFromPayload(payload(
		"caption", "She said yes @jordan! #JustEngaged",
		"source_account_type", "proposal_planner",
	))
	if len(sig.TaggedHandles) != 0 {
		t.Errorf("caption mentions are not image tags: %v", sig.TaggedHandles)
	}
	if len(sig.MentionedHandles) != 1 || sig.MentionedHandles[0] != "jordan" {
		t.Errorf("expected @jordan mentioned, got %v", sig.MentionedHandles)
	}
	// Explicit language creates the candidate regardless; the mention feeds
	// partner identification.
	if !sig.CreatesCandidate() {
		t.Fatal("explicit phrase + hashtag must create a candidate")
	}
}

func TestBothPartnersIdentified(t *testing.T) {
	cases := []struct {
		name    string
		payload map[string]any
		author  string
		want    bool
	}{
		{"third party references both", payload("tags", []any{"maya", "jordan"}), "photog", true},
		{"third party mentions both in caption", payload("caption", "congrats @maya and @jordan"), "photog", true},
		{"author tags partner", payload("tags", []any{"jordan"}), "maya", true},
		{"collab between the two", payload("collab_with", "jordan"), "maya", true},
		{"only one referenced", payload("tags", []any{"jordan"}), "photog", false},
		{"two strangers referenced", payload("tags", []any{"alex", "sam"}), "photog", false},
	}
	for _, tc := range cases {
		sig := ExtractFromPayload(tc.payload)
		if got := BothPartnersIdentified(sig, tc.author, "maya", "jordan"); got != tc.want {
			t.Errorf("%s: BothPartnersIdentified = %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestPhraseMatching_UsesWordBoundaries(t *testing.T) {
	// "she said yes" must not also count "he said yes" (the "he" in "she").
	sig := ExtractFromPayload(payload("caption", "She said yes!"))
	if len(sig.PhraseHits) != 1 || sig.PhraseHits[0] != "she said yes" {
		t.Errorf("expected exactly [she said yes], got %v", sig.PhraseHits)
	}
	// Both can co-occur when genuinely both are written.
	sig2 := ExtractFromPayload(payload("caption", "He said yes... and she said yes!"))
	if len(sig2.PhraseHits) != 2 {
		t.Errorf("expected both phrases, got %v", sig2.PhraseHits)
	}
}

func TestRingEmojiAlone_IsNotExplicit(t *testing.T) {
	// "Team Jordan forever 💍" — a fan account's caption — is not an
	// engagement announcement.
	sig := ExtractFromPayload(payload("caption", "Team Jordan forever 💍"))
	if sig.HasExplicitLanguage() {
		t.Error("a 💍 emoji with no explicit phrase must not count as explicit language")
	}
}

func TestHasExplicitBioLanguage(t *testing.T) {
	if !HasExplicitBioLanguage("future Mrs. Lee 💍") {
		t.Error("expected 'future Mrs. Lee' to be explicit bio language")
	}
	if !HasExplicitBioLanguage("fiancée of jordan 💍") {
		t.Error("expected fiancée bio to be explicit")
	}
	if HasExplicitBioLanguage("dog mom. coffee first.") {
		t.Error("unrelated bio must not be explicit")
	}
}
