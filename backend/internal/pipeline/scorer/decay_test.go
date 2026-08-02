package scorer

import (
	"math"
	"testing"
	"time"

	"neptune-social-radar/backend/internal/ontology"
)

func TestDecayFactor(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name string
		age  time.Duration
		want float64
	}{
		{"zero age", 0, 1.0},
		{"negative age (future)", -time.Hour, 1.0},
		{"half life", evidenceHalfLife, 0.5},
		{"2x half life", 2 * evidenceHalfLife, 0.25},
		{"3x half life", 3 * evidenceHalfLife, 0.125},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := decayFactor(tt.age, now)
			if math.Abs(got-tt.want) > 0.001 {
				t.Errorf("decayFactor(%v) = %.4f, want %.4f", tt.age, got, tt.want)
			}
		})
	}
}

func TestDecayedWeight(t *testing.T) {
	now := time.Now()
	// Evidence at half-life age should have half its weight.
	e := ontology.Evidence{Weight: 1.0, CreatedAt: now.Add(-evidenceHalfLife)}
	got := decayedWeight(e, now)
	if math.Abs(got-0.5) > 0.001 {
		t.Errorf("decayedWeight at halfLife = %.4f, want 0.5", got)
	}

	// Fresh evidence keeps full weight.
	fresh := ontology.Evidence{Weight: 1.0, CreatedAt: now.Add(-time.Hour)}
	got = decayedWeight(fresh, now)
	if math.Abs(got-1.0) > 0.01 {
		t.Errorf("decayedWeight fresh = %.4f, want ~1.0", got)
	}

	// Zero CreatedAt (missing timestamp) keeps full weight — don't penalize
	// data we can't date.
	undated := ontology.Evidence{Weight: 0.8, CreatedAt: time.Time{}}
	got = decayedWeight(undated, now)
	if got != 0.8 {
		t.Errorf("decayedWeight undated = %.4f, want 0.8", got)
	}
}

func TestScoreAppliesDecay(t *testing.T) {
	now := time.Now()
	// Two evidence items: one fresh (weight 0.4), one at 2x half-life (weight 0.4 → 0.1 after decay).
	// Without decay: sum = 0.8, score = 0.5*0.5 + 0.5*0.8 = 0.65
	// With decay:    sum = 0.4 + 0.1 = 0.5, score = 0.5*0.5 + 0.5*0.5 = 0.5
	fresh := ontology.Evidence{Weight: 0.4, CreatedAt: now.Add(-time.Hour)}
	old := ontology.Evidence{Weight: 0.4, CreatedAt: now.Add(-2 * evidenceHalfLife)}
	score := Score(0.5, []ontology.Evidence{fresh, old})
	if math.Abs(score-0.5) > 0.01 {
		t.Errorf("Score with decay = %.4f, want ~0.5", score)
	}
	// Without decay it would be 0.65 — confirm decay actually changed the score.
	noDecayScore := 0.5*0.5 + 0.5*(0.4+0.4)
	if math.Abs(score-noDecayScore) < 0.05 {
		t.Error("decay should have meaningfully changed the score")
	}
}

func TestProspectScoreAppliesDecay(t *testing.T) {
	now := time.Now()
	// Fresh explicit language (40 pts) + old vendor source (15 pts at 2x half-life → 3.75 pts)
	fresh := ontology.Evidence{Kind: EvExplicitLanguage, Weight: 0.4, CreatedAt: now.Add(-time.Hour)}
	old := ontology.Evidence{Kind: EvKnownVendorSource, Weight: 0.15, CreatedAt: now.Add(-2 * evidenceHalfLife)}
	final, eng, _ := ProspectScore([]ontology.Evidence{fresh, old})
	// Without decay: final = (40 + 15)/100 = 0.55
	// With decay:    final = (40 + 3.75)/100 = 0.4375
	if final >= 0.5 {
		t.Errorf("ProspectScore with decay = %.4f, expected < 0.5 (decay should reduce old evidence)", final)
	}
	// Engagement sub-score: (40 + 3.75) / maxEngagementPoints(75) = 0.583
	if eng < 0.55 || eng > 0.62 {
		t.Errorf("engagement sub-score = %.4f, expected ~0.58", eng)
	}
}
