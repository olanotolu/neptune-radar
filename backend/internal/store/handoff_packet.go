package store

import (
	"fmt"
	"strings"
	"time"
)

// HandoffPacket is the structured export for attorney handoff — a
// evidence-cited summary of what we know about a couple, formatted for
// a professional referral. This is the "alignment narrative" made concrete.
type HandoffPacket struct {
	CoupleID       string           `json:"couple_id"`
	GeneratedAt    time.Time        `json:"generated_at"`
	Names          [2]string        `json:"names"`
	Handles        [2]string        `json:"handles"`
	City           string           `json:"city,omitempty"`
	Region         string           `json:"region,omitempty"`
	Stage          string           `json:"stage"`
	Confidence     float64          `json:"confidence"`
	RunwayLabel    string           `json:"runway_label,omitempty"`
	ICPLabels      []string         `json:"icp_labels,omitempty"`
	WhyNow         []string         `json:"why_now"`
	Evidence       []PacketEvidence `json:"evidence"`
	AlignmentNotes string           `json:"alignment_notes"`
	JourneyStage   string           `json:"journey_stage"`
}

type PacketEvidence struct {
	Date    string `json:"date"`
	Source  string `json:"source"`
	Type    string `json:"type"`
	Summary string `json:"summary"`
	URL     string `json:"url,omitempty"`
}

// GenerateHandoffPacket assembles the evidence-cited narrative for a couple.
// It pulls the god-tier dossier, extracts evidence with citations, and formats
// it as a structured packet suitable for attorney referral.
func (s *Store) GenerateHandoffPacket(coupleID string) (HandoffPacket, error) {
	d, err := s.GetGodTierDossier(coupleID)
	if err != nil {
		return HandoffPacket{}, err
	}
	pkt := HandoffPacket{
		CoupleID:     coupleID,
		GeneratedAt:  time.Now().UTC(),
		Names:        [2]string{d.PersonAName, d.PersonBName},
		Handles:      [2]string{d.HandleA, d.HandleB},
		City:         d.City,
		Region:       d.Region,
		Stage:        d.Stage,
		Confidence:   d.HypothesisScore,
		RunwayLabel:  d.RunwayLabel,
		ICPLabels:    d.ICP.Labels,
		WhyNow:       d.WhyNow,
		JourneyStage: d.JourneyStage,
	}
	// Convert dossier evidence to packet evidence with citations.
	for _, e := range d.Evidence {
		pe := PacketEvidence{
			Date:    e.CreatedAt,
			Type:    e.Kind,
			Summary: e.Description,
		}
		if e.Confirmed {
			pe.Source = "confirmed"
		}
		pkt.Evidence = append(pkt.Evidence, pe)
	}
	// Build alignment notes — the narrative connecting evidence to the
	// engagement/marriage conclusion.
	pkt.AlignmentNotes = buildAlignmentNotes(d)
	return pkt, nil
}

func buildAlignmentNotes(d GodTierDossier) string {
	var notes []string
	if d.Stage != "" {
		notes = append(notes, fmt.Sprintf("Stage: %s (confidence: %.0f%%)", d.Stage, d.HypothesisScore*100))
	}
	if d.RunwayLabel != "" {
		notes = append(notes, "Runway: "+d.RunwayLabel)
	}
	if len(d.ICP.Labels) > 0 {
		notes = append(notes, "ICP fit: "+strings.Join(d.ICP.Labels, ", "))
	}
	if len(d.WhyNow) > 0 {
		notes = append(notes, "Why now: "+strings.Join(d.WhyNow, "; "))
	}
	if len(d.Evidence) > 0 {
		notes = append(notes, fmt.Sprintf("Evidence: %d items cited", len(d.Evidence)))
	}
	return strings.Join(notes, "\n")
}
