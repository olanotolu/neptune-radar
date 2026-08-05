package store

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"neptune-social-radar/backend/internal/ontology"
	"neptune-social-radar/backend/internal/signals"
)

// EvidenceRow is one ledger line on the Couple Dossier.
type EvidenceRow struct {
	ID          string  `json:"id"`
	Kind        string  `json:"kind"`
	Description string  `json:"description"`
	Weight      float64 `json:"weight"`
	Points      int     `json:"points"` // weight*100 for engagement ledger
	Confirmed   bool    `json:"confirmed"`
	CreatedAt   string  `json:"created_at,omitempty"`
}

// JourneyStep is one step in the brand-safe celebrate → invite path.
type JourneyStep struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	Description string `json:"description"`
	Status      string `json:"status"` // done | current | upcoming | blocked
}

// BrandAction is a recommended operator action that respects Neptune brand.
type BrandAction struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Body        string `json:"body"`
	Tone        string `json:"tone"` // celebrate | soft_invite | internal | risk
	Allowed     bool   `json:"allowed"`
	BlockReason string `json:"block_reason,omitempty"`
}

// IdentityProof is reciprocal / non-biometric identity evidence.
type IdentityProof struct {
	Kind        string `json:"kind"`
	Description string `json:"description"`
	Strength    string `json:"strength"` // strong | medium | weak
}

// GodTierDossier is the full Couple Dossier for the operator console.
type GodTierDossier struct {
	CoupleID        string             `json:"couple_id"`
	HandleA         string             `json:"handle_a,omitempty"`
	HandleB         string             `json:"handle_b,omitempty"`
	PersonAName     string             `json:"person_a_name,omitempty"`
	PersonBName     string             `json:"person_b_name,omitempty"`
	BioA            string             `json:"bio_a,omitempty"`
	BioB            string             `json:"bio_b,omitempty"`
	ProfilePicA     string             `json:"profile_pic_a,omitempty"`
	ProfilePicB     string             `json:"profile_pic_b,omitempty"`
	City            string             `json:"city,omitempty"`
	Region          string             `json:"region,omitempty"`
	Stage           string             `json:"stage,omitempty"`
	AutomationPaused bool              `json:"automation_paused"`
	HasCase         bool               `json:"has_case"`
	// Scores
	EngagementScore float64            `json:"engagement_score"`
	PartnerScore    float64            `json:"partner_score,omitempty"`
	HypothesisScore float64            `json:"hypothesis_score"`
	ICP             signals.ICPFit     `json:"icp"`
	Runway          signals.WeddingRunway `json:"runway"`
	RunwayLabel     string             `json:"runway_label"`
	NeptuneRank     float64            `json:"neptune_rank"`
	Deliverability  float64            `json:"deliverability"`
	// Ledger + identity
	Evidence        []EvidenceRow      `json:"evidence"`
	Identity        []IdentityProof    `json:"identity"`
	// Pending action
	PendingActionID   string           `json:"pending_action_id,omitempty"`
	PendingActionType string           `json:"pending_action_type,omitempty"`
	ProposedPayload   string           `json:"proposed_payload,omitempty"`
	HypothesisID      string           `json:"hypothesis_id,omitempty"`
	// Journey + handoff
	JourneyStage    string             `json:"journey_stage"`
	JourneySteps    []JourneyStep      `json:"journey_steps"`
	HandoffCode     string             `json:"handoff_code,omitempty"`
	HandoffURL      string             `json:"handoff_url,omitempty"`
	HandoffUTM      string             `json:"handoff_utm,omitempty"`
	// Brand-safe recommended actions
	BrandActions    []BrandAction      `json:"brand_actions"`
	// Soft copy (never hard-sells prenup on first touch)
	CelebrateCopy   string             `json:"celebrate_copy"`
	SoftInviteCopy  string             `json:"soft_invite_copy"`
	// Observations + kit + audit
	Observations    []DossierPost      `json:"observations"`
	LatestKitID     string             `json:"latest_kit_id,omitempty"`
	LatestKitStatus string             `json:"latest_kit_status,omitempty"`
	AuditTrail      []ontology.AuditEvent `json:"audit_trail"`
	// Why this couple, now
	WhyNow          []string           `json:"why_now"`
	// Financial profile from county property records (internal operator use only)
	AssetProfile    *AssetProfile      `json:"asset_profile,omitempty"`
}

// AssetProfile is the financial summary from county property records.
// Internal operator use only — never appears on postcards.
type AssetProfile struct {
	EstimatedHomeValue int64         `json:"estimated_home_value,omitempty"`
	PropertyAsset      PropertyAsset `json:"property_asset,omitempty"`
	Confidence         float64       `json:"confidence"` // 0-1, based on data completeness
	Source             string        `json:"source,omitempty"`
}

// GetGodTierDossier builds the operator dossier for one couple.
func (s *Store) GetGodTierDossier(coupleID string) (GodTierDossier, error) {
	var d GodTierDossier
	d.CoupleID = coupleID
	d.Evidence = []EvidenceRow{}
	d.Identity = []IdentityProof{}
	d.Observations = []DossierPost{}
	d.AuditTrail = []ontology.AuditEvent{}
	d.WhyNow = []string{}
	d.JourneyStage = "detected"

	// Base board projection
	board, err := s.ListProspectBoard(400)
	if err != nil {
		return d, err
	}
	var card *ProspectCard
	for i := range board {
		if board[i].CoupleID == coupleID {
			card = &board[i]
			break
		}
	}
	if card != nil {
		d.HandleA, d.HandleB = card.HandleA, card.HandleB
		d.PersonAName, d.PersonBName = card.PersonALabel, card.PersonBLabel
		d.BioA, d.BioB = card.BioA, card.BioB
		d.ProfilePicA, d.ProfilePicB = card.ProfilePicA, card.ProfilePicB
		d.City, d.Region = card.City, card.Region
		d.Stage = card.Stage
		d.AutomationPaused = card.AutomationPaused
		d.HasCase = card.HasCase
		d.HypothesisScore = card.HypothesisScore
		if d.HypothesisScore == 0 {
			d.HypothesisScore = card.Confidence
		}
		d.EngagementScore = card.HypothesisScore
		if card.Confidence > d.EngagementScore {
			d.EngagementScore = card.Confidence
		}
		d.PendingActionID = card.PendingActionID
		d.PendingActionType = card.PendingActionType
		d.ProposedPayload = card.ProposedPayload
		// Prefer enriched rank fields if already computed on card
		d.ICP = card.ICP
		d.Runway = card.Runway
		d.NeptuneRank = card.NeptuneRank
		d.RunwayLabel = card.RunwayLabel
		d.JourneyStage = card.JourneyStage
		if d.JourneyStage == "" {
			d.JourneyStage = "detected"
		}
	}

	// Journey + handoff from DB
	var journey, handoffCode, handoffURL, handoffUTM string
	_ = s.DB.QueryRow(`
		SELECT COALESCE(journey_stage,'detected'), COALESCE(handoff_code,''),
		       COALESCE(handoff_url,''), COALESCE(handoff_utm,'')
		FROM couples WHERE id = $1`, coupleID,
	).Scan(&journey, &handoffCode, &handoffURL, &handoffUTM)
	if journey != "" {
		d.JourneyStage = journey
	}
	d.HandoffCode, d.HandoffURL, d.HandoffUTM = handoffCode, handoffURL, handoffUTM

	// Latest hypothesis + evidence ledger
	var hypID string
	var engConf, partnerConf float64
	err = s.DB.QueryRow(`
		SELECT id, confidence,
		       COALESCE(engagement_confidence, confidence),
		       COALESCE(partner_confidence, 0)
		FROM life_event_hypotheses
		WHERE couple_id = $1
		ORDER BY created_at DESC LIMIT 1`, coupleID,
	).Scan(&hypID, &d.HypothesisScore, &engConf, &partnerConf)
	if err == nil && hypID != "" {
		d.HypothesisID = hypID
		d.EngagementScore = engConf
		d.PartnerScore = partnerConf
		ev, err := s.EvidenceForHypothesis(hypID)
		if err == nil {
			for _, e := range ev {
				d.Evidence = append(d.Evidence, EvidenceRow{
					ID: e.ID, Kind: e.Kind, Description: e.Description,
					Weight: e.Weight, Points: int(e.Weight * 100),
					Confirmed: e.Confirmed,
					CreatedAt: e.CreatedAt.UTC().Format(time.RFC3339),
				})
			}
		}
	}

	// Captions for runway extraction
	base, _ := s.GetCoupleDossier(coupleID)
	if d.HandleA == "" {
		d.HandleA, d.HandleB = base.HandleA, base.HandleB
		d.PersonAName, d.PersonBName = base.PersonAName, base.PersonBName
		d.BioA, d.BioB = base.BioA, base.BioB
		d.ProfilePicA, d.ProfilePicB = base.ProfilePicA, base.ProfilePicB
		d.City, d.Region = base.City, base.Region
	}
	d.Observations = base.Observations
	d.LatestKitID = base.LatestKitID
	if kit, err := s.GetLatestKitForCouple(coupleID); err == nil {
		d.LatestKitID = kit.ID
		d.LatestKitStatus = kit.Status
		// Financial profile from county property records (internal only)
		if kit.EstimatedHomeValue > 0 || kit.PropertyAsset.AssessedValue > 0 || kit.PropertyAsset.Sqft > 0 {
			conf := 0.0
			if kit.PropertyAsset.AssessedValue > 0 {
				conf += 0.4
			}
			if kit.PropertyAsset.Sqft > 0 {
				conf += 0.3
			}
			if kit.PropertyAsset.YearBuilt > 0 {
				conf += 0.15
			}
			if kit.PropertyAsset.TaxAnnual > 0 {
				conf += 0.15
			}
			d.AssetProfile = &AssetProfile{
				EstimatedHomeValue: kit.EstimatedHomeValue,
				PropertyAsset:      kit.PropertyAsset,
				Confidence:         conf,
				Source:             "county_property",
			}
		}
	}

	// Aggregate caption + tags for runway
	var captions []string
	var tags []string
	for _, o := range d.Observations {
		if o.Caption != "" {
			captions = append(captions, o.Caption)
		}
		tags = append(tags, o.Tags...)
	}
	captionBlob := strings.Join(captions, "\n")

	// Recompute ICP + runway live (always freshest)
	d.ICP = signals.ExtractICPFit(d.BioA, d.BioB, d.City, d.Region)
	d.Runway = signals.ExtractWeddingRunway(captionBlob, d.BioA, d.BioB, tags, time.Time{})
	d.RunwayLabel = signals.FormatRunwayLabel(d.Runway)

	// Deliverability: pics + location + address kit progress
	deliv := 0.4
	if d.ProfilePicA != "" && d.ProfilePicB != "" {
		deliv += 0.25
	} else if d.ProfilePicA != "" || d.ProfilePicB != "" {
		deliv += 0.1
	}
	if d.City != "" {
		deliv += 0.2
	}
	switch d.LatestKitStatus {
	case "mailed", "ready_to_mail", "address_verified":
		deliv += 0.15
	case "ready_review", "draft":
		deliv += 0.05
	}
	if deliv > 1 {
		deliv = 1
	}
	d.Deliverability = deliv
	d.NeptuneRank = signals.NeptuneRank(d.EngagementScore, d.ICP.Score, d.Runway.Factor, deliv)

	// Identity proofs from edges between the two accounts
	d.Identity = s.identityProofs(coupleID, d.HandleA, d.HandleB)

	// Audit trail for this couple + related hypothesis/actions
	aud, _ := s.ListAudit(AuditFilter{EntityID: coupleID, Limit: 40})
	d.AuditTrail = aud
	if hypID != "" {
		if a2, err := s.ListAudit(AuditFilter{EntityID: hypID, Limit: 20}); err == nil {
			d.AuditTrail = append(d.AuditTrail, a2...)
		}
	}

	// Why now
	if d.EngagementScore >= 0.9 {
		d.WhyNow = append(d.WhyNow, "Engagement score ≥ 90% — create-prospect bar")
	} else if d.EngagementScore >= 0.7 {
		d.WhyNow = append(d.WhyNow, "Engagement score 70–89% — investigation queue")
	}
	if d.Runway.Band == "green" {
		d.WhyNow = append(d.WhyNow, "Wedding runway supports full prenup process ("+d.RunwayLabel+")")
	} else if d.Runway.Band == "amber" {
		d.WhyNow = append(d.WhyNow, "Wedding runway is tight but workable — prioritize if ICP fits")
	} else if d.Runway.Band == "red" {
		d.WhyNow = append(d.WhyNow, "⚠ Runway too short for standard prenup path — do not rush outreach")
	} else {
		d.WhyNow = append(d.WhyNow, "Wedding date unknown — extract or ask before hard prioritization")
	}
	if d.ICP.Score >= 0.55 {
		d.WhyNow = append(d.WhyNow, "ICP fit strong: "+strings.Join(d.ICP.Labels, ", "))
	}
	if d.PendingActionType == "concierge_review" || d.PendingActionType == "pause_automation" {
		d.WhyNow = append(d.WhyNow, "Relationship-risk path — NEVER pitch; concierge only")
	}

	// Brand-safe copy (dual-partner, celebrate-first)
	nameA := firstNonEmptyStr(d.PersonAName, d.HandleA, "there")
	nameB := firstNonEmptyStr(d.PersonBName, d.HandleB, "partner")
	d.CelebrateCopy = fmt.Sprintf(
		"Dear %s & %s — congratulations on your engagement. Wishing you a beautiful chapter ahead. With warmth, the Neptune team.",
		nameA, nameB,
	)
	d.SoftInviteCopy = fmt.Sprintf(
		"Hi %s & %s — when you're ready to plan the admin of partnership together (no pressure, both of you in the room), Neptune's free chat helps couples get aligned before talking to attorneys. %s",
		nameA, nameB,
		func() string {
			if d.HandoffURL != "" {
				return d.HandoffURL
			}
			return "https://app.meetneptune.com/chat"
		}(),
	)

	d.JourneySteps = buildJourneySteps(d.JourneyStage, d.LatestKitStatus, d.HandoffURL != "")
	d.BrandActions = buildBrandActions(d)

	return d, nil
}

func firstNonEmptyStr(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func (s *Store) identityProofs(coupleID, handleA, handleB string) []IdentityProof {
	var out []IdentityProof
	// Reciprocal tags / follows between person accounts
	rows, err := s.DB.Query(`
		SELECT e.kind, e.active, fa.handle, ta.handle
		FROM edges e
		JOIN social_accounts fa ON fa.id = e.from_account_id
		JOIN social_accounts ta ON ta.id = e.to_account_id
		JOIN couples c ON c.id = $1
		WHERE (
		  (fa.person_id = c.person_a_id AND ta.person_id = c.person_b_id)
		  OR (fa.person_id = c.person_b_id AND ta.person_id = c.person_a_id)
		)
		ORDER BY e.last_observed_at DESC
		LIMIT 20`, coupleID)
	if err != nil {
		// Fall back to handle-based narrative
		if handleA != "" && handleB != "" {
			out = append(out, IdentityProof{
				Kind: "handles_present",
				Description: fmt.Sprintf("@%s and @%s linked as a couple candidate", handleA, handleB),
				Strength: "weak",
			})
		}
		return out
	}
	defer rows.Close()
	tagAB, tagBA, followAB, followBA := false, false, false, false
	for rows.Next() {
		var kind string
		var active bool
		var fromH, toH string
		if err := rows.Scan(&kind, &active, &fromH, &toH); err != nil {
			continue
		}
		if !active {
			continue
		}
		switch kind {
		case "tagged_with":
			if strings.EqualFold(fromH, handleA) {
				tagAB = true
			} else {
				tagBA = true
			}
		case "follows":
			if strings.EqualFold(fromH, handleA) {
				followAB = true
			} else {
				followBA = true
			}
		}
	}
	if tagAB && tagBA {
		out = append(out, IdentityProof{
			Kind: "reciprocal_tag", Description: "Mutual image tags between accounts", Strength: "strong",
		})
	} else if tagAB || tagBA {
		out = append(out, IdentityProof{
			Kind: "one_way_tag", Description: "One-way tag (not yet reciprocal)", Strength: "medium",
		})
	}
	if followAB && followBA {
		out = append(out, IdentityProof{
			Kind: "reciprocal_follow", Description: "Mutual follows", Strength: "strong",
		})
	} else if followAB || followBA {
		out = append(out, IdentityProof{
			Kind: "one_way_follow", Description: "One-way follow", Strength: "weak",
		})
	}
	if len(out) == 0 {
		out = append(out, IdentityProof{
			Kind: "event_first_pair",
			Description: "Event-first pair from engagement-shaped post (no reciprocal edge yet)",
			Strength: "medium",
		})
	}
	return out
}

func buildJourneySteps(stage, kitStatus string, hasHandoff bool) []JourneyStep {
	order := []struct {
		id, label, desc string
	}{
		{"detected", "Detected", "Engagement signal scored by the radar"},
		{"approved", "Approved", "Human approved the prospect"},
		{"congratulated", "Congratulated", "Celebrate-first touch (postcard / soft note)"},
		{"invited", "Invited", "Soft invite + tracked chat handoff"},
		{"in_chat", "In chat", "Couple started Neptune AI chat"},
		{"booked", "Booked", "Attorney consult booked"},
		{"closed_won", "Closed", "Prenup path completed"},
	}
	rank := map[string]int{
		"detected": 0, "approved": 1, "congratulated": 2, "invited": 3,
		"in_chat": 4, "booked": 5, "closed_won": 6, "closed_lost": -1, "do_not_contact": -1,
	}
	// Kit status can advance congratulate without explicit journey write
	cur := rank[stage]
	if cur < 2 && (kitStatus == "mailed" || kitStatus == "ready_to_mail") {
		cur = 2
	}
	if cur < 3 && hasHandoff {
		cur = 3
	}
	var steps []JourneyStep
	for i, s := range order {
		st := "upcoming"
		if cur < 0 {
			st = "blocked"
		} else if i < cur {
			st = "done"
		} else if i == cur {
			st = "current"
		}
		steps = append(steps, JourneyStep{ID: s.id, Label: s.label, Description: s.desc, Status: st})
	}
	return steps
}

func buildBrandActions(d GodTierDossier) []BrandAction {
	risk := d.PendingActionType == "concierge_review" || d.PendingActionType == "pause_automation" || d.AutomationPaused
	var out []BrandAction

	// Celebrate
	celebrateOK := !risk && d.JourneyStage != "do_not_contact" && d.JourneyStage != "closed_lost"
	block := ""
	if risk {
		block = "Relationship-risk path — celebration outreach is blocked"
	}
	if d.Runway.SuppressOutreach && d.PendingActionType == "review" {
		// Still allow celebrate; block only hard pitch later
	}
	out = append(out, BrandAction{
		ID: "congratulate", Title: "Congratulate (postcard / kit)",
		Body: "Celebration only — no prenup language. Build kit → verify address → mail.",
		Tone: "celebrate", Allowed: celebrateOK, BlockReason: block,
	})

	// Soft invite
	inviteOK := celebrateOK && (d.JourneyStage == "congratulated" || d.JourneyStage == "approved" || d.LatestKitStatus == "mailed")
	inviteBlock := ""
	if risk {
		inviteBlock = "Risk path — no invite"
	} else if d.Runway.SuppressOutreach {
		inviteBlock = "Wedding runway too short for standard prenup process"
	} else if !inviteOK && celebrateOK {
		inviteBlock = "Celebrate first — issue invite after congratulate step"
	}
	out = append(out, BrandAction{
		ID: "soft_invite", Title: "Soft invite to Neptune chat",
		Body: "Both partners, no hard sell. Uses tracked handoff link into app.meetneptune.com/chat.",
		Tone: "soft_invite", Allowed: inviteOK && !d.Runway.SuppressOutreach,
		BlockReason: inviteBlock,
	})

	// Approve pending
	if d.PendingActionID != "" {
		out = append(out, BrandAction{
			ID: "approve_pending", Title: "Approve pending action (" + d.PendingActionType + ")",
			Body: "Human-in-the-loop gate. Policy already decided; this writes case/lead state.",
			Tone: "internal", Allowed: true,
		})
	}

	// Risk
	if risk {
		out = append(out, BrandAction{
			ID: "concierge_only", Title: "Concierge review only",
			Body: "Pause automation. Do not pitch. Review relationship context with care.",
			Tone: "risk", Allowed: true,
		})
	}

	// Suppress
	out = append(out, BrandAction{
		ID: "not_a_couple", Title: "Not a couple / wrong identity",
		Body: "Permanent suppress — scorer will not resurface this pair.",
		Tone: "internal", Allowed: true,
	})

	return out
}

// EnsureHandoff creates a tracked chat handoff URL if missing.
func (s *Store) EnsureHandoff(coupleID string) (code, url, utm string, err error) {
	var existingCode, existingURL, existingUTM string
	err = s.DB.QueryRow(`
		SELECT COALESCE(handoff_code,''), COALESCE(handoff_url,''), COALESCE(handoff_utm,'')
		FROM couples WHERE id = $1`, coupleID,
	).Scan(&existingCode, &existingURL, &existingUTM)
	if err != nil {
		return "", "", "", err
	}
	if existingURL != "" {
		return existingCode, existingURL, existingUTM, nil
	}
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		return "", "", "", err
	}
	code = hex.EncodeToString(b)
	base := os.Getenv("NEPTUNE_CHAT_BASE_URL")
	if base == "" {
		base = "https://app.meetneptune.com/chat"
	}
	utm = fmt.Sprintf("utm_source=neptune_radar&utm_medium=handoff&utm_campaign=engagement_detect&utm_content=%s", coupleID)
	url = fmt.Sprintf("%s?%s&ref=%s", base, utm, code)
	_, err = s.DB.Exec(`
		UPDATE couples SET handoff_code = $2, handoff_url = $3, handoff_utm = $4, handoff_created_at = now()
		WHERE id = $1`, coupleID, code, url, utm)
	if err != nil {
		return "", "", "", err
	}
	_, _ = s.Audit("couple", coupleID, "handoff_created", map[string]string{
		"code": code, "url": url,
	}, "growth_os", -1)
	return code, url, utm, nil
}

// SetJourneyStage updates the brand-safe funnel stage.
func (s *Store) SetJourneyStage(coupleID, stage string) error {
	valid := map[string]bool{
		"detected": true, "approved": true, "congratulated": true, "invited": true,
		"in_chat": true, "booked": true, "closed_won": true, "closed_lost": true, "do_not_contact": true,
	}
	if !valid[stage] {
		return fmt.Errorf("invalid journey stage %q", stage)
	}
	_, err := s.DB.Exec(`
		UPDATE couples SET journey_stage = $2, journey_updated_at = now() WHERE id = $1`,
		coupleID, stage)
	if err != nil {
		return err
	}
	_, _ = s.Audit("couple", coupleID, "journey_stage", map[string]string{"stage": stage}, "growth_os", -1)
	return nil
}

// EnrichProspectCard fills runway, ICP, rank, journey on a card (mutates).
func EnrichProspectCard(c *ProspectCard, captionBlob string, tags []string) {
	if c == nil {
		return
	}
	c.ICP = signals.ExtractICPFit(c.BioA, c.BioB, c.City, c.Region)
	c.Runway = signals.ExtractWeddingRunway(captionBlob, c.BioA, c.BioB, tags, time.Time{})
	c.RunwayLabel = signals.FormatRunwayLabel(c.Runway)
	deliv := 0.5
	if c.ProfilePicA != "" && c.ProfilePicB != "" {
		deliv += 0.25
	}
	if c.City != "" {
		deliv += 0.25
	}
	eng := c.HypothesisScore
	if c.Confidence > eng {
		eng = c.Confidence
	}
	c.NeptuneRank = signals.NeptuneRank(eng, c.ICP.Score, c.Runway.Factor, deliv)
}

// ListRecentCaptionsForCouple returns a blob of recent captions for runway scoring.
func (s *Store) ListRecentCaptionsForCouple(handleA, handleB string, limit int) (caption string, tags []string) {
	if limit <= 0 {
		limit = 10
	}
	if handleA == "" && handleB == "" {
		return "", nil
	}
	ha, hb := handleA, handleB
	if ha == "" {
		ha = hb
	}
	if hb == "" {
		hb = ha
	}
	rows, err := s.DB.Query(`
		SELECT COALESCE(caption,''), COALESCE(tags_json,'[]'), raw_payload
		FROM social_observations
		WHERE raw_payload ILIKE '%' || $1 || '%' OR raw_payload ILIKE '%' || $2 || '%'
		ORDER BY observed_at DESC LIMIT $3`, ha, hb, limit)
	if err != nil {
		return "", nil
	}
	defer rows.Close()
	var parts []string
	tagSet := map[string]bool{}
	for rows.Next() {
		var cap, tagsRaw, raw string
		if err := rows.Scan(&cap, &tagsRaw, &raw); err != nil {
			continue
		}
		if cap == "" {
			var payload map[string]any
			if json.Unmarshal([]byte(raw), &payload) == nil {
				cap, _ = payload["caption"].(string)
				for _, t := range stringSlice(payload["tags"]) {
					if !tagSet[t] {
						tagSet[t] = true
						tags = append(tags, t)
					}
				}
			}
		} else {
			var t []string
			_ = json.Unmarshal([]byte(tagsRaw), &t)
			for _, x := range t {
				if !tagSet[x] {
					tagSet[x] = true
					tags = append(tags, x)
				}
			}
		}
		if cap != "" {
			parts = append(parts, cap)
		}
	}
	return strings.Join(parts, "\n"), tags
}
