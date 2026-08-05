// Package scorer is the Neptune Relevance Scorer. It collects corroborating
// (or contradicting) Evidence for a hypothesis from current DB state — not
// just the single triggering event — and combines it with the model's
// proposed confidence into one final score. This is still not a decision:
// internal/pipeline/policy applies the actual action thresholds.
//
// Engagement prospects are scored on the product spec's explicit points
// table (explicit language +40, both partners tagged +25, known vendor +15,
// visual +10, reciprocal +10, registry +15, recent +10; styled shoot/ad −50,
// no second person −30, old/reposted −25, identity conflict −40), stored as
// evidence weights of points/100 and recombined by ProspectScore.
// Relationship-state-change hypotheses keep the original 0.5·model +
// 0.5·evidence blend — the two triggers are different and stay different.
package scorer

import (
	"database/sql"
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"neptune-social-radar/backend/internal/ontology"
	"neptune-social-radar/backend/internal/pipeline/analyst"
	"neptune-social-radar/backend/internal/pipeline/identity"
	"neptune-social-radar/backend/internal/pipeline/watchtower"
	"neptune-social-radar/backend/internal/signals"
	"neptune-social-radar/backend/internal/store"
)

// Evidence kinds for the engagement-prospect path. Weights are the spec's
// points divided by 100.
const (
	EvExplicitLanguage   = "explicit_engagement_language"
	EvBothPartnersTagged = "both_partners_tagged"
	EvKnownVendorSource  = "known_vendor_source"
	EvVisualRing         = "visual_ring_detected"
	// EvRingDetected is the YOLOv8 ring-detection signal (confidence ≥ 0.5).
	// Stronger than EvVisualRing because it comes from a dedicated ring model,
	// not a general visual classifier.
	EvRingDetected = "ring_detected"
	// EvProposalPhoto is the CLIP zero-shot photo classification signal: the
	// image was classified as "marriage proposal" or "engagement photo shoot"
	// with high confidence.
	EvProposalPhoto = "proposal_photo"
	// EvDispersionScore is the FAIR dispersion metric: high dispersion of
	// mutual follows indicates a romantic relationship bridging social circles.
	EvDispersionScore = "dispersion_score"
	EvReciprocal         = "reciprocal_relationship"
	EvRegistryMatch      = "registry_match"
	EvRecentPost         = "recent_post"
	EvStyledShoot        = "styled_shoot"
	EvAdvertisement      = "advertisement"
	EvOldReposted        = "old_reposted"
	EvNoSecondPerson     = "no_identifiable_second_person"
	EvIdentityConflict   = "conflicting_identity"
	// EvRepeatedCooccurrence is the spec's "+10 repeated meaningful
	// co-occurrence": the pair has been referenced together before, by more
	// than one independent source account. Read from pair_cooccurrences —
	// never from a re-scan of raw posts.
	EvRepeatedCooccurrence = "repeated_cooccurrence"
	// EvVendorInPair is the spec's "−30 one profile is clearly a vendor":
	// role resolution found that an account in the proposed pair is itself
	// a vendor-shaped node (wrong role — vendor, not partner).
	EvVendorInPair = "vendor_in_pair"
)

func pts(p float64) float64 { return p / 100.0 }

// oldPostThreshold is how fresh a post must be to earn the +10 recency
// points. Thirty days: engagement announcements are acted on within weeks;
// anything older is context, not news. Tunable via NEPTUNE_OLD_POST_DAYS.
var oldPostThreshold = time.Duration(loadDurationEnv("NEPTUNE_OLD_POST_DAYS", 30)) * 24 * time.Hour

// evidenceHalfLife is the age at which evidence weight drops to 50%.
// 90 days: social signals for a life event go stale fast — a post from 6
// months ago should carry far less weight than one from last week. The decay
// is exponential (weight *= 0.5^(age/halfLife)), so at 180 days evidence is
// at 25%, at 270 days at 12.5%. This prevents a hypothesis from accumulating
// a high score purely from old evidence that was never corroborated.
// Tunable via NEPTUNE_EVIDENCE_HALF_LIFE_DAYS.
var evidenceHalfLife = time.Duration(loadDurationEnv("NEPTUNE_EVIDENCE_HALF_LIFE_DAYS", 90)) * 24 * time.Hour

// loadDurationEnv reads an integer number of days from an env var, falling
// back to the default if unset or invalid. Never crashes on bad input.
func loadDurationEnv(key string, defaultDays int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return defaultDays
}

// decayFactor returns the weight multiplier for evidence of the given age.
// Fresh evidence (age < halfLife) keeps most of its weight; old evidence
// decays exponentially. Zero-age evidence returns 1.0 (no decay).
func decayFactor(age time.Duration, now time.Time) float64 {
	if age <= 0 || evidenceHalfLife <= 0 {
		return 1.0
	}
	// 0.5^(age/halfLife): at halfLife, factor=0.5; at 2*halfLife, 0.25; etc.
	lambda := float64(age) / float64(evidenceHalfLife)
	return math.Pow(0.5, lambda)
}

// decayedWeight applies the time-decay factor to one evidence item's weight.
// Evidence with a zero CreatedAt (migration backfill, missing timestamp) is
// treated as full-weight — we don't penalize data we can't date.
func decayedWeight(e ontology.Evidence, now time.Time) float64 {
	if e.CreatedAt.IsZero() {
		return e.Weight
	}
	return e.Weight * decayFactor(now.Sub(e.CreatedAt), now)
}

// CollectEvidence re-derives supporting/contradicting facts from the
// database's current state (edges, account flags) each time a hypothesis is
// touched, so a later event (unfollow reversed, account re-enabled) can
// correct or retract evidence a prior event added. modelConfidence is the
// interpreter's proposal — used on the engagement path only to credit
// explicit language the deterministic vocabulary did NOT match (semantic
// variations); it can never invent points on its own.
func CollectEvidence(s *store.Store, hyp ontology.LifeEventHypothesis, res identity.Resolved, cand analyst.Candidate, raw watchtower.RawEvent, modelConfidence float64) ([]ontology.Evidence, error) {
	switch hyp.EventType {
	case analyst.CandidateEngagement:
		// Evidence splits into two families that ProspectScore reports
		// separately: "did an engagement-shaped event happen" (language,
		// vendor source, visuals, recency, exclusions) and "did we identify
		// the right two people" (both partners referenced, reciprocal edges,
		// registry match, identity conflicts).
		sig := signals.ExtractFromPayload(raw.Payload)
		switch raw.Type {
		case "post":
			if sig.HasExplicitLanguage() {
				desc := "explicit engagement language: " + sig.LanguageSummary()
				if sup := sig.SupportingSummary(); sup != "" {
					desc += " (corroborated by " + sup + ")"
				}
				if _, err := s.UpsertEvidenceKind(hyp.ID, EvExplicitLanguage, desc, pts(signals.PtsExplicitLanguage)); err != nil {
					return nil, err
				}
			} else if modelConfidence >= 0.7 {
				// The model's only scoring role: credit explicit language the
				// deterministic phrase list missed — "he asked and I said
				// forever". The template fallback never proposes this (it
				// caps itself at 0.5 without a deterministic match), so this
				// path requires a real model's judgment by construction.
				if _, err := s.UpsertEvidenceKind(hyp.ID, EvExplicitLanguage,
					"model judged the caption engagement-explicit beyond the deterministic vocabulary: \""+sig.Caption+"\"",
					pts(signals.PtsExplicitLanguage)); err != nil {
					return nil, err
				}
			}
			if sig.IsKnownVendor() {
				if _, err := s.UpsertEvidenceKind(hyp.ID, EvKnownVendorSource,
					"source is a classified "+strings.ReplaceAll(sig.SourceClass, "_", " ")+" account",
					pts(signals.PtsKnownVendorSource)); err != nil {
					return nil, err
				}
			}
			if sig.HasVisual() {
				if _, err := s.UpsertEvidenceKind(hyp.ID, EvVisualRing,
					"visual/on-screen signals detected: "+strings.Join(sig.Visual, ", ")+" (supports the caption; identifies no one)",
					pts(signals.PtsVisualRing)); err != nil {
					return nil, err
				}
			}
			// YOLOv8 ring detection — dedicated ring model, stronger signal
			// than the general visual classifier. Proposal photos with rings
			// are the highest-confidence engagement signal.
			if ringConf, ok := raw.Payload["ring_confidence"].(float64); ok && ringConf >= 0.5 {
				if _, err := s.UpsertEvidenceKind(hyp.ID, EvRingDetected,
					fmt.Sprintf("YOLOv8 ring detector confidence %.0f%%", ringConf*100),
					0.20); err != nil { // ponytail: +20 pts — stronger than general visual (+10)
					return nil, err
				}
			}
			// CLIP zero-shot photo classification — "marriage proposal" or
			// "engagement photo shoot" labels are direct engagement signals.
			if photoLabel, ok := raw.Payload["photo_label"].(string); ok && photoLabel != "" {
				if photoLabel == "marriage proposal" || photoLabel == "engagement photo shoot" {
					photoConf, _ := raw.Payload["photo_confidence"].(float64)
					if _, err := s.UpsertEvidenceKind(hyp.ID, EvProposalPhoto,
						fmt.Sprintf("CLIP classified photo as \"%s\" (%.0f%% confidence)", photoLabel, photoConf*100),
						0.15); err != nil { // ponytail: +15 pts — proposal-shaped photo
						return nil, err
					}
				}
			}
			// Recency is computed from the post's real timestamp, not assumed.
			// The vocabulary's old/reposted penalty (repost flag, throwback
			// hashtags) can also force "old" — but nothing fresh-looking is
			// credited +10 just for existing: a 3-year-old backfill post must
			// not score as "recent, original".
			old := sig.PenaltyKinds[signals.PenaltyOldReposted] || time.Since(raw.OccurredAt) > oldPostThreshold
			if old {
				if _, err := s.UpsertEvidenceKind(hyp.ID, EvOldReposted, "old or reposted content", pts(signals.PtsOldReposted)); err != nil {
					return nil, err
				}
				if err := s.DeleteEvidenceKind(hyp.ID, EvRecentPost); err != nil {
					return nil, err
				}
			} else {
				if _, err := s.UpsertEvidenceKind(hyp.ID, EvRecentPost, "recent, original post", pts(signals.PtsRecentPost)); err != nil {
					return nil, err
				}
			}
			if sig.PenaltyKinds[signals.PenaltyStyledShoot] {
				if _, err := s.UpsertEvidenceKind(hyp.ID, EvStyledShoot, "styled/editorial shoot or vendor inspiration content", pts(signals.PtsStyledShoot)); err != nil {
					return nil, err
				}
			} else {
				_ = s.DeleteEvidenceKind(hyp.ID, EvStyledShoot)
			}
			if sig.PenaltyKinds[signals.PenaltyAdvertisement] {
				if _, err := s.UpsertEvidenceKind(hyp.ID, EvAdvertisement, "advertisement, sponsorship, or giveaway", pts(signals.PtsAdvertisement)); err != nil {
					return nil, err
				}
			} else {
				_ = s.DeleteEvidenceKind(hyp.ID, EvAdvertisement)
			}
			if res.PartnerAccount != nil && signals.BothPartnersIdentified(sig, raw.Handle, res.Account.Handle, res.PartnerAccount.Handle) {
				if _, err := s.UpsertEvidenceKind(hyp.ID, EvBothPartnersTagged,
					"both probable partners referenced on the post ("+res.Account.Handle+" + "+res.PartnerAccount.Handle+")",
					pts(signals.PtsBothPartnersTagged)); err != nil {
					return nil, err
				}
			}
			// NOTE: the spec's cross-source corroboration points (registry
			// +15, identity-conflict −40) are intentionally NOT scored from
			// payload flags here — nothing produced those flags, so the
			// evidence was decorative. They land for real when the
			// government/registry/bulletin connectors feed the pipeline.

		case "bio_change":
			if signals.HasExplicitBioLanguage(res.Account.BioText) {
				if _, err := s.UpsertEvidenceKind(hyp.ID, EvExplicitLanguage, "bio references partner as fiancé(e)/future spouse: \""+res.Account.BioText+"\"", pts(signals.PtsExplicitLanguage)); err != nil {
					return nil, err
				}
			}
		}

		// Reciprocal relationship evidence is re-derived from CURRENT edges
		// on every touch, so a reversed unfollow retracts it.
		if res.PartnerAccount != nil {
			reciprocalTag, _ := hasReciprocalEdge(s, ontology.EdgeTaggedWith, res.Account.ID, res.PartnerAccount.ID)
			mutualFollow, _ := hasReciprocalEdge(s, ontology.EdgeFollows, res.Account.ID, res.PartnerAccount.ID)
			if reciprocalTag || mutualFollow {
				if _, err := s.UpsertEvidenceKind(hyp.ID, EvReciprocal, "reciprocal relationship evidence: the two accounts tag or follow each other", pts(signals.PtsReciprocalEvidence)); err != nil {
					return nil, err
				}
			} else if err := s.DeleteEvidenceKind(hyp.ID, EvReciprocal); err != nil {
				return nil, err
			}

			// Repeated meaningful co-occurrence: the pair has shown up
			// together before AND across more than one independent source
			// account (two posts by the same photographer is one source —
			// the same client relationship, not corroboration).
			cooc, err := s.GetPairCooccurrence(res.Account.ID, res.PartnerAccount.ID)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				return nil, err
			}
			if err == nil && cooc.SharedPosts >= 2 && cooc.DistinctSources >= 2 {
				desc := "the pair has been referenced together " + strconv.Itoa(cooc.SharedPosts) + " times by " + strconv.Itoa(cooc.DistinctSources) + " independent source accounts"
				if _, err := s.UpsertEvidenceKind(hyp.ID, EvRepeatedCooccurrence, desc, pts(signals.PtsRepeatedCooccurrence)); err != nil {
					return nil, err
				}
			} else if err := s.DeleteEvidenceKind(hyp.ID, EvRepeatedCooccurrence); err != nil {
				return nil, err
			}

			// Vendor-in-pair contradiction: the curated registry says one
			// account in the proposed pair is itself a vendor — the classic
			// "Jane + her florist" false couple. Re-derived every touch, so
			// a handle removed from the registry retracts the penalty.
			vendorFound := ""
			for _, acct := range []ontology.SocialAccount{res.Account, *res.PartnerAccount} {
				if class := s.SourceClassForHandle(acct.Handle); class != "" {
					vendorFound = acct.Handle + " is on the curated source registry as " + class + " — vendor, not partner"
					break
				}
			}
			if vendorFound != "" {
				if _, err := s.UpsertEvidenceKind(hyp.ID, EvVendorInPair, vendorFound, pts(signals.PtsVendorInPair)); err != nil {
					return nil, err
				}
			} else if err := s.DeleteEvidenceKind(hyp.ID, EvVendorInPair); err != nil {
				return nil, err
			}
		}

		// "No identifiable second person" fires only when ZERO positive
		// partner-identity evidence exists — a well-written caption pointing
		// at nobody identifiable is exactly what this penalty exists for.
		current, err := s.EvidenceForHypothesis(hyp.ID)
		if err != nil {
			return nil, err
		}
		partnerPositive := false
		for _, e := range current {
			if (e.Kind == EvBothPartnersTagged || e.Kind == EvReciprocal || e.Kind == EvRegistryMatch) && e.Weight > 0 {
				partnerPositive = true
			}
		}
		if !partnerPositive {
			if _, err := s.UpsertEvidenceKind(hyp.ID, EvNoSecondPerson, "no identifiable second person: no tags, reciprocal edges, or registry match corroborate a partner", pts(signals.PtsNoSecondPerson)); err != nil {
				return nil, err
			}
		} else if err := s.DeleteEvidenceKind(hyp.ID, EvNoSecondPerson); err != nil {
			return nil, err
		}

	case analyst.CandidateRelationshipChange:
		if res.PartnerAccount == nil {
			break
		}
		followEdge, err := s.GetEdge(ontology.EdgeFollows, res.Account.ID, res.PartnerAccount.ID)
		if err == nil {
			if !followEdge.Active {
				if _, err := s.UpsertEvidenceKind(hyp.ID, "unfollow_detected", res.Account.Handle+" no longer follows "+res.PartnerAccount.Handle, 0.35); err != nil {
					return nil, err
				}
				if err := s.DeleteEvidenceKind(hyp.ID, "follow_restored"); err != nil {
					return nil, err
				}
			} else {
				// The unfollow didn't stick — this is the "reversed shortly
				// after" adversarial case. Retract the prior signal instead
				// of merely offsetting it, so the hypothesis genuinely
				// reflects current reality.
				if err := s.DeleteEvidenceKind(hyp.ID, "unfollow_detected"); err != nil {
					return nil, err
				}
				if _, err := s.UpsertEvidenceKind(hyp.ID, "follow_restored", "follow was restored shortly after", -0.3); err != nil {
					return nil, err
				}
			}
		}

		bio := res.Account.BioText
		if !containsPartnerReference(bio) {
			if _, err := s.UpsertEvidenceKind(hyp.ID, "bio_reference_removed", "\"fiancé(e)\" bio reference was removed", 0.25); err != nil {
				return nil, err
			}
		} else {
			_ = s.DeleteEvidenceKind(hyp.ID, "bio_reference_removed")
		}

		if raw.Type == "post_archived" {
			if _, err := s.UpsertEvidenceKind(hyp.ID, "post_archived", "the engagement post was archived", 0.25); err != nil {
				return nil, err
			}
		}

		// Alternative-explanation checks: don't let an account outage or
		// rename masquerade as a meaningful relationship signal.
		if res.PartnerAccount.IsDisabled {
			if _, err := s.UpsertEvidenceKind(hyp.ID, "partner_account_uncertain", res.PartnerAccount.Handle+"'s account is currently disabled — cannot corroborate", -0.4); err != nil {
				return nil, err
			}
		} else {
			_ = s.DeleteEvidenceKind(hyp.ID, "partner_account_uncertain")
		}
	}

	return s.EvidenceForHypothesis(hyp.ID)
}

func hasReciprocalEdge(s *store.Store, kind ontology.EdgeKind, a, b string) (bool, error) {
	ab, err := s.GetEdge(kind, a, b)
	if err != nil {
		return false, nil //nolint:nilerr
	}
	ba, err := s.GetEdge(kind, b, a)
	if err != nil {
		return false, nil //nolint:nilerr
	}
	return ab.Active && ba.Active, nil
}

func containsPartnerReference(bio string) bool {
	bio = strings.ToLower(bio)
	for _, w := range []string{"fiance", "fiancé", "fiancée", "engaged"} {
		if strings.Contains(bio, w) {
			return true
		}
	}
	return false
}

// Score combines the model's proposed confidence with the sum of current
// evidence weights. Policy — not this function — decides what confidence
// level is enough to act.
func Score(modelConfidence float64, evidence []ontology.Evidence) float64 {
	now := time.Now()
	sum := 0.0
	for _, e := range evidence {
		sum += decayedWeight(e, now)
	}
	// Deliberately no neutral baseline: a single weak signal (e.g. one
	// unfollow with no corroboration) must not be enough to cross an action
	// threshold on its own — precision over recall.
	score := 0.5*modelConfidence + 0.5*sum
	if score < 0 {
		score = 0
	}
	if score > 1 {
		score = 1
	}
	return score
}

// Engagement-family evidence answers "did an engagement-shaped event
// happen"; partner-family evidence answers "did we identify the right two
// people." ProspectScore reports the two sub-scores separately so a confident
// caption tagging the wrong second person is visible as exactly that.
var engagementEvidenceKinds = map[string]bool{
	EvExplicitLanguage: true, EvKnownVendorSource: true, EvVisualRing: true,
	EvRecentPost: true, EvStyledShoot: true, EvAdvertisement: true, EvOldReposted: true,
}
var partnerEvidenceKinds = map[string]bool{
	EvBothPartnersTagged: true, EvReciprocal: true, EvRegistryMatch: true,
	EvNoSecondPerson: true, EvIdentityConflict: true,
	EvRepeatedCooccurrence: true, EvVendorInPair: true,
}

// Maximum positive points achievable in each family — used to normalize the
// two display sub-scores. The action gate itself uses the raw combined total.
const (
	maxEngagementPoints = 75.0 // 40 language + 15 vendor + 10 visual + 10 recent
	maxPartnerPoints    = 60.0 // 25 both-tagged + 10 reciprocal + 15 registry + 10 repeated co-occurrence
)

// ProspectScore recombines engagement-prospect evidence into the spec's
// points: final = combined points normalized to 0–1 (100 points == 1.0),
// which policy then tiers at 0.90 (create prospect) and 0.70 (investigation
// queue). There is deliberately no model-confidence term and no neutral
// baseline: the model's only influence was deciding whether borderline
// language counted as explicit, and every point on the board is an auditable
// deterministic signal.
func ProspectScore(evidence []ontology.Evidence) (final, engagementSub, partnerSub float64) {
	now := time.Now()
	total, engagement, partner := 0.0, 0.0, 0.0
	for _, e := range evidence {
		points := decayedWeight(e, now) * 100
		total += points
		switch {
		case engagementEvidenceKinds[e.Kind]:
			engagement += points
		case partnerEvidenceKinds[e.Kind]:
			partner += points
		}
	}
	return clamp01(total / 100), clamp01(engagement / maxEngagementPoints), clamp01(partner / maxPartnerPoints)
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
