package outreach

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"strings"

	"neptune-social-radar/backend/internal/store"
)

// applyVariantTone adjusts the drafted copy's tone based on the variant's
// CopyModifier. ponytail: simple string substitution, not a full rewrite —
// the LLM already drafted good copy; we just nudge the register. Ceiling:
// won't handle edge cases where the exact phrase is absent (no-op then).
func applyVariantTone(copy, modifier, nameA, nameB, locLabel string) string {
	switch modifier {
	case "warm_extra":
		// B: warmer — soften the opening, add an exclamation.
		copy = strings.Replace(copy, "Congratulations", "We are so thrilled for you both", 1)
		if !strings.Contains(copy, "so happy") {
			copy = strings.Replace(copy, "\n\n", "\n\nSending so much happiness your way.\n\n", 1)
		}
	case "direct":
		// C: action-oriented — replace soft closing with a clear next step.
		copy = strings.Replace(copy, "With warm regards,", "Ready to start planning?", 1)
		copy = strings.Replace(copy, "With care,", "Let's make it happen —", 1)
		copy = strings.Replace(copy, "Warmly,", "Your next step starts here —", 1)
		if !strings.Contains(copy, "meetneptune.com") {
			copy += "\n\nVisit meetneptune.com to get started."
		}
	}
	return copy
}

// Variant is one arm of an A/B experiment. CopyModifier controls tone ("formal",
// "warm", "direct"); Personalized flags whether LLM personalization runs.
type Variant struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	TemplateID   string `json:"template_id"`
	CopyModifier string `json:"copy_modifier"` // "formal", "warm", "direct"
	Personalized bool   `json:"personalized"`
}

// Experiment is a named set of variants assigned deterministically by couple.
type Experiment struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	Variants []Variant `json:"variants"`
}

// DefaultExperiment is the postcard copy A/B test. A = control (existing
// template), B = warmer tone, C = direct/action-oriented.
func DefaultExperiment() Experiment {
	return Experiment{
		ID:   "postcard_copy_v1",
		Name: "Postcard Copy v1",
		Variants: []Variant{
			{ID: "A", Name: "Control", TemplateID: "warm_ohio", CopyModifier: "warm"},
			{ID: "B", Name: "Warmer tone", TemplateID: "warm_ohio", CopyModifier: "warm_extra"},
			{ID: "C", Name: "Direct / action", TemplateID: "bright_casual", CopyModifier: "direct"},
		},
	}
}

// AssignVariant deterministically assigns a variant by hashing the coupleID.
// Same couple always gets the same variant — no sticky storage needed.
func AssignVariant(coupleID string, exp Experiment) Variant {
	if len(exp.Variants) == 0 {
		return Variant{}
	}
	h := sha256.Sum256([]byte(coupleID + ":" + exp.ID))
	n := binary.BigEndian.Uint64(h[:8]) % uint64(len(exp.Variants))
	return exp.Variants[n]
}

// VariantResult is per-variant conversion performance for one experiment.
type VariantResult struct {
	VariantID      string  `json:"variant_id"`
	VariantName    string  `json:"variant_name"`
	Mailed         int     `json:"mailed"`
	Scans          int     `json:"scans"`
	ScanRate       float64 `json:"scan_rate"`
	Chats          int     `json:"chats"`
	ConversionRate float64 `json:"conversion_rate"`
	SampleSize     int     `json:"sample_size"`
}

// ExperimentResults is the full rollup returned by GET /api/experiments/{id}/results.
type ExperimentResults struct {
	ExperimentID string         `json:"experiment_id"`
	ExperimentName string       `json:"experiment_name"`
	Variants     []VariantResult `json:"variants"`
	WinnerID     string         `json:"winner_id,omitempty"`
}

// minSampleForWinner is the minimum mailed count before we declare a winner.
// ponytail: 10 is a soft floor — real significance needs Fisher's exact, but
// this stops a 1/1 fluke from crowning a winner. Upgrade: compute p-value.
const minSampleForWinner = 10

// GetExperimentResults rolls up per-variant conversion from kits + funnel events.
// Mailed = kits with status 'mailed'; Scans = kits with qr_scan_count > 0;
// Chats = couples with a 'chat_started' funnel event.
func GetExperimentResults(s *store.Store, experimentID string) (ExperimentResults, error) {
	exp := DefaultExperiment()
	if experimentID != "" && experimentID != exp.ID {
		// ponytail: only one experiment exists today; unknown IDs return empty.
		return ExperimentResults{ExperimentID: experimentID}, nil
	}

	rows, err := s.DB.Query(`
		SELECT COALESCE(k.variant_id,'A') AS variant,
		       COUNT(*) AS mailed,
		       COUNT(*) FILTER (WHERE k.qr_scan_count > 0) AS scans,
		       COUNT(*) FILTER (WHERE EXISTS (
		         SELECT 1 FROM funnel_events fe
		         WHERE fe.couple_id = k.couple_id AND fe.event_type = 'chat_started'
		       )) AS chats
		FROM congratulate_kits k
		WHERE k.experiment_id = $1 AND k.status = 'mailed'
		GROUP BY COALESCE(k.variant_id,'A')
		ORDER BY variant`, experimentID)
	if err != nil {
		return ExperimentResults{}, fmt.Errorf("experiment results: %w", err)
	}
	defer rows.Close()

	byID := map[string]VariantResult{}
	for rows.Next() {
		var vr VariantResult
		if err := rows.Scan(&vr.VariantID, &vr.Mailed, &vr.Scans, &vr.Chats); err != nil {
			return ExperimentResults{}, err
		}
		vr.SampleSize = vr.Mailed
		if vr.Mailed > 0 {
			vr.ScanRate = float64(vr.Scans) / float64(vr.Mailed)
			vr.ConversionRate = float64(vr.Chats) / float64(vr.Mailed)
		}
		byID[vr.VariantID] = vr
	}

	// Fill in all variants from the experiment definition (even if zero mailed).
	var variants []VariantResult
	var winnerID string
	bestConv := -1.0
	for _, v := range exp.Variants {
		vr := byID[v.ID]
		vr.VariantID = v.ID
		vr.VariantName = v.Name
		if vr.Mailed == 0 {
			vr.SampleSize = 0
		}
		variants = append(variants, vr)
		if vr.Mailed >= minSampleForWinner && vr.ConversionRate > bestConv {
			bestConv = vr.ConversionRate
			winnerID = v.ID
		}
	}

	return ExperimentResults{
		ExperimentID:   exp.ID,
		ExperimentName: exp.Name,
		Variants:       variants,
		WinnerID:       winnerID,
	}, rows.Err()
}
