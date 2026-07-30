package scorer

import (
	"testing"

	"neptune-social-radar/backend/internal/ontology"
)

func ev(kind string, points float64) ontology.Evidence {
	return ontology.Evidence{Kind: kind, Weight: points / 100}
}

func TestProspectScore_SpecExamples(t *testing.T) {
	cases := []struct {
		name      string
		evidence  []ontology.Evidence
		wantFinal float64
	}{
		{
			// The spec's canonical case: a photographer posts "she said yes"
			// and tags both clients — 40+25+15+10 = exactly 90, the
			// create-prospect floor.
			name: "vendor + language + both tagged = create-prospect floor",
			evidence: []ontology.Evidence{
				ev(EvExplicitLanguage, 40), ev(EvBothPartnersTagged, 25),
				ev(EvKnownVendorSource, 15), ev(EvRecentPost, 10),
			},
			wantFinal: 0.90,
		},
		{
			// A self-posted "I said yes" with a tagged partner and mutual
			// follows: 85 — investigation queue, not a prospect.
			name: "self post, no vendor/visual/registry = investigation tier",
			evidence: []ontology.Evidence{
				ev(EvExplicitLanguage, 40), ev(EvBothPartnersTagged, 25),
				ev(EvReciprocal, 10), ev(EvRecentPost, 10),
			},
			wantFinal: 0.85,
		},
		{
			// The jewelry-ad suppression: everything tempting plus the −50.
			name: "ad with every positive signal stacked = below the bar",
			evidence: []ontology.Evidence{
				ev(EvExplicitLanguage, 40), ev(EvBothPartnersTagged, 25),
				ev(EvKnownVendorSource, 15), ev(EvVisualRing, 10),
				ev(EvReciprocal, 10), ev(EvRecentPost, 10),
				ev(EvAdvertisement, -50),
			},
			wantFinal: 0.60,
		},
		{
			// A confident caption pointing at nobody identifiable can never
			// reach even the investigation tier (spec: −30).
			name: "explicit language with no second person is capped at discard",
			evidence: []ontology.Evidence{
				ev(EvExplicitLanguage, 40), ev(EvKnownVendorSource, 15),
				ev(EvVisualRing, 10), ev(EvRegistryMatch, 15),
				ev(EvRecentPost, 10), ev(EvNoSecondPerson, -30),
			},
			wantFinal: 0.60,
		},
		{
			name:      "a single weak signal is nothing",
			evidence:  []ontology.Evidence{ev(EvRecentPost, 10)},
			wantFinal: 0.10,
		},
		{
			name: "negative totals clamp to zero",
			evidence: []ontology.Evidence{
				ev(EvExplicitLanguage, 40), ev(EvAdvertisement, -50), ev(EvStyledShoot, -50),
			},
			wantFinal: 0,
		},
	}
	for _, tc := range cases {
		final, _, _ := ProspectScore(tc.evidence)
		if final != tc.wantFinal {
			t.Errorf("%s: final = %.2f, want %.2f", tc.name, final, tc.wantFinal)
		}
	}
}

func TestProspectScore_SubScoresAreSeparate(t *testing.T) {
	// 40 language + 10 recent + 10 visual = 60/75 engagement points, and
	// only the −30 no-second-person on the partner side — the two numbers
	// must stay distinguishable, never averaged.
	final, engagement, partner := ProspectScore([]ontology.Evidence{
		ev(EvExplicitLanguage, 40), ev(EvVisualRing, 10), ev(EvRecentPost, 10),
		ev(EvNoSecondPerson, -30),
	})
	if final != 0.30 {
		t.Errorf("final = %.2f, want 0.30", final)
	}
	if engagement != 0.80 {
		t.Errorf("engagement sub-score = %.2f, want 0.80 (60/75)", engagement)
	}
	if partner != 0 {
		t.Errorf("partner sub-score = %.2f, want 0 (negative clamps)", partner)
	}
}
