package store

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
)

// AutopsyCase is one false-positive or human-reject lesson.
type AutopsyCase struct {
	CoupleID     string   `json:"couple_id,omitempty"`
	Kind         string   `json:"kind"` // suppressed | ignored | rejected
	Reason       string   `json:"reason"`
	Handles      []string `json:"handles,omitempty"`
	Labels       []string `json:"labels,omitempty"`
	Score        float64  `json:"score,omitempty"`
	ActionType   string   `json:"action_type,omitempty"`
	Lesson       string   `json:"lesson"`
	OccurredAt   string   `json:"occurred_at,omitempty"`
	HypothesisID string   `json:"hypothesis_id,omitempty"`
}

// AutopsySummary is the executive rollup for legal / ops trust.
type AutopsySummary struct {
	PeriodStart           string         `json:"period_start"`
	PeriodEnd             string         `json:"period_end"`
	SuppressedCouples     int            `json:"suppressed_couples"`
	IgnoredActions        int            `json:"ignored_actions"`
	RejectedHypotheses    int            `json:"rejected_hypotheses"`
	ApprovedActions       int            `json:"approved_actions"`
	PendingActions        int            `json:"pending_actions"`
	// HumanRejectRate = (ignored + suppressed) / (approved + ignored + suppressed)
	// when denominator > 0. Proxy for "how often operators disagree with the model."
	HumanRejectRate float64          `json:"human_reject_rate"`
	ByReason        map[string]int   `json:"by_reason"`
	ByActionType    map[string]int   `json:"by_action_type_ignored"`
	TopFailureModes []string         `json:"top_failure_modes"`
	Funnel          FunnelStats      `json:"funnel"`
	Notes           []string         `json:"notes"`
}

// AutopsyReport is a persisted weekly (or ad-hoc) trust report.
type AutopsyReport struct {
	ID          string         `json:"id"`
	PeriodStart time.Time      `json:"period_start"`
	PeriodEnd   time.Time      `json:"period_end"`
	Summary     AutopsySummary `json:"summary"`
	Cases       []AutopsyCase  `json:"cases"`
	GeneratedBy string         `json:"generated_by"`
	CreatedAt   time.Time      `json:"created_at"`
}

// GenerateAutopsy builds a false-positive / ignore report for [start, end).
func (s *Store) GenerateAutopsy(start, end time.Time, generatedBy string) (AutopsyReport, error) {
	if end.IsZero() {
		end = time.Now().UTC()
	}
	if start.IsZero() {
		start = end.AddDate(0, 0, -7)
	}
	if generatedBy == "" {
		generatedBy = "operator"
	}

	rep := AutopsyReport{
		ID:          NewID("autopsy"),
		PeriodStart: start.UTC(),
		PeriodEnd:   end.UTC(),
		GeneratedBy: generatedBy,
		CreatedAt:   time.Now().UTC(),
		Cases:       []AutopsyCase{},
	}
	sum := AutopsySummary{
		PeriodStart: start.UTC().Format(time.RFC3339),
		PeriodEnd:   end.UTC().Format(time.RFC3339),
		ByReason:    map[string]int{},
		ByActionType: map[string]int{},
		Notes:       []string{},
	}

	// --- Suppressed couples ---
	rows, err := s.DB.Query(`
		SELECT c.id, COALESCE(c.suppressed_reason,'unspecified'), c.suppressed_at,
		       COALESCE(pa.display_name,''), COALESCE(pb.display_name,''),
		       COALESCE(aa.handle,''), COALESCE(ab.handle,''),
		       COALESCE(h.confidence, 0)
		FROM couples c
		LEFT JOIN persons pa ON pa.id = c.person_a_id
		LEFT JOIN persons pb ON pb.id = c.person_b_id
		LEFT JOIN LATERAL (
		  SELECT handle FROM social_accounts WHERE person_id = c.person_a_id LIMIT 1
		) aa ON true
		LEFT JOIN LATERAL (
		  SELECT handle FROM social_accounts WHERE person_id = c.person_b_id LIMIT 1
		) ab ON true
		LEFT JOIN LATERAL (
		  SELECT confidence FROM life_event_hypotheses WHERE couple_id = c.id
		  ORDER BY created_at DESC LIMIT 1
		) h ON true
		WHERE c.suppressed_at IS NOT NULL
		  AND c.suppressed_at >= $1 AND c.suppressed_at < $2
		ORDER BY c.suppressed_at DESC
		LIMIT 200`, start, end)
	if err != nil {
		return rep, err
	}
	for rows.Next() {
		var coupleID, reason string
		var at time.Time
		var la, lb, ha, hb string
		var conf float64
		if err := rows.Scan(&coupleID, &reason, &at, &la, &lb, &ha, &hb, &conf); err != nil {
			rows.Close()
			return rep, err
		}
		sum.SuppressedCouples++
		sum.ByReason[reason]++
		handles := nonEmptyHandles(ha, hb)
		rep.Cases = append(rep.Cases, AutopsyCase{
			CoupleID:   coupleID,
			Kind:       "suppressed",
			Reason:     reason,
			Handles:    handles,
			Labels:     nonEmptyHandles(la, lb),
			Score:      conf,
			Lesson:     lessonForSuppress(reason, conf),
			OccurredAt: at.UTC().Format(time.RFC3339),
		})
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return rep, err
	}

	// --- Ignored recommended actions ---
	rows, err = s.DB.Query(`
		SELECT ra.id, ra.action_type, ra.decided_at, COALESCE(ra.decided_by,''),
		       COALESCE(h.id,''), COALESCE(h.couple_id,''), COALESCE(h.confidence,0),
		       COALESCE(aa.handle,''), COALESCE(ab.handle,'')
		FROM recommended_actions ra
		LEFT JOIN life_event_hypotheses h ON h.id = ra.hypothesis_id
		LEFT JOIN couples c ON c.id = h.couple_id
		LEFT JOIN LATERAL (
		  SELECT handle FROM social_accounts WHERE person_id = c.person_a_id LIMIT 1
		) aa ON true
		LEFT JOIN LATERAL (
		  SELECT handle FROM social_accounts WHERE person_id = c.person_b_id LIMIT 1
		) ab ON true
		WHERE ra.status = 'ignored'
		  AND ra.decided_at IS NOT NULL
		  AND ra.decided_at >= $1 AND ra.decided_at < $2
		ORDER BY ra.decided_at DESC
		LIMIT 200`, start, end)
	if err != nil {
		return rep, err
	}
	for rows.Next() {
		var actionID, actionType, decidedBy, hypID, coupleID string
		var decidedAt time.Time
		var conf float64
		var ha, hb string
		if err := rows.Scan(&actionID, &actionType, &decidedAt, &decidedBy, &hypID, &coupleID, &conf, &ha, &hb); err != nil {
			rows.Close()
			return rep, err
		}
		sum.IgnoredActions++
		sum.ByActionType[actionType]++
		reason := "ignored:" + actionType
		sum.ByReason[reason]++
		rep.Cases = append(rep.Cases, AutopsyCase{
			CoupleID:     coupleID,
			Kind:         "ignored",
			Reason:       reason,
			Handles:      nonEmptyHandles(ha, hb),
			Score:        conf,
			ActionType:   actionType,
			HypothesisID: hypID,
			Lesson:       lessonForIgnore(actionType, conf),
			OccurredAt:   decidedAt.UTC().Format(time.RFC3339),
		})
	}
	rows.Close()

	// --- Rejected hypotheses ---
	rows, err = s.DB.Query(`
		SELECT h.id, COALESCE(h.couple_id,''), h.confidence, h.updated_at, h.event_type
		FROM life_event_hypotheses h
		WHERE h.status = 'rejected'
		  AND h.updated_at >= $1 AND h.updated_at < $2
		ORDER BY h.updated_at DESC
		LIMIT 100`, start, end)
	if err != nil {
		return rep, err
	}
	for rows.Next() {
		var hypID, coupleID, eventType string
		var conf float64
		var at time.Time
		if err := rows.Scan(&hypID, &coupleID, &conf, &at, &eventType); err != nil {
			rows.Close()
			return rep, err
		}
		sum.RejectedHypotheses++
		sum.ByReason["hypothesis_rejected"]++
		rep.Cases = append(rep.Cases, AutopsyCase{
			CoupleID:     coupleID,
			Kind:         "rejected",
			Reason:       "hypothesis_rejected:" + eventType,
			Score:        conf,
			HypothesisID: hypID,
			Lesson:       "Hypothesis rejected after human review — keep on audit trail; do not re-surface without new evidence.",
			OccurredAt:   at.UTC().Format(time.RFC3339),
		})
	}
	rows.Close()

	// Approvals + pending for rate
	_ = s.DB.QueryRow(`
		SELECT COUNT(*) FROM recommended_actions
		WHERE status = 'approved' AND decided_at IS NOT NULL
		  AND decided_at >= $1 AND decided_at < $2`, start, end).Scan(&sum.ApprovedActions)
	_ = s.DB.QueryRow(`SELECT COUNT(*) FROM recommended_actions WHERE status = 'pending'`).Scan(&sum.PendingActions)

	denom := sum.ApprovedActions + sum.IgnoredActions + sum.SuppressedCouples
	if denom > 0 {
		sum.HumanRejectRate = float64(sum.IgnoredActions+sum.SuppressedCouples) / float64(denom)
	}

	// Top failure modes by reason count
	type kv struct {
		k string
		v int
	}
	var ranked []kv
	for k, v := range sum.ByReason {
		ranked = append(ranked, kv{k, v})
	}
	sort.Slice(ranked, func(i, j int) bool { return ranked[i].v > ranked[j].v })
	for i, x := range ranked {
		if i >= 8 {
			break
		}
		sum.TopFailureModes = append(sum.TopFailureModes, fmt.Sprintf("%s (%d)", x.k, x.v))
	}

	if st, err := s.GetFunnelStats(); err == nil {
		sum.Funnel = st
	}

	// Narrative notes for legal/ops
	sum.Notes = append(sum.Notes,
		"Suppressions are permanent 'not a couple / wrong identity' judgments — the strongest FP signal.",
		"Ignores mean the model proposed an action a human rejected; review high-score ignores first.",
		"Human reject rate is a proxy, not a legal metric. Track trend week-over-week.",
	)
	if sum.HumanRejectRate >= 0.4 && denom >= 5 {
		sum.Notes = append(sum.Notes, "⚠ Human reject rate ≥ 40% this period — tighten scoring or identity gates before scaling outreach.")
	}
	if sum.SuppressedCouples == 0 && sum.IgnoredActions == 0 {
		sum.Notes = append(sum.Notes, "No suppressions or ignores in window — either quiet period or operators are not using reject paths (check audit).")
	}

	rep.Summary = sum

	// Persist
	sumJSON, _ := json.Marshal(sum)
	casesJSON, _ := json.Marshal(rep.Cases)
	_, err = s.DB.Exec(`
		INSERT INTO autopsy_reports (id, period_start, period_end, summary_json, cases_json, generated_by, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		rep.ID, rep.PeriodStart, rep.PeriodEnd, string(sumJSON), string(casesJSON), rep.GeneratedBy, rep.CreatedAt,
	)
	if err != nil {
		return rep, err
	}
	_, _ = s.Audit("system", rep.ID, "autopsy_generated", map[string]any{
		"period_start": sum.PeriodStart,
		"period_end":   sum.PeriodEnd,
		"suppressions": sum.SuppressedCouples,
		"ignores":      sum.IgnoredActions,
		"reject_rate":  sum.HumanRejectRate,
	}, "trust", -1)
	return rep, nil
}

func nonEmptyHandles(a, b string) []string {
	var out []string
	if strings.TrimSpace(a) != "" {
		out = append(out, a)
	}
	if strings.TrimSpace(b) != "" {
		out = append(out, b)
	}
	return out
}

func lessonForSuppress(reason string, conf float64) string {
	switch reason {
	case "not_a_couple":
		return "Operator marked not a couple — reciprocal identity may be weak; check vendor/model pairs and co-tag noise."
	case "vendor_vendor_pair":
		return "Both handles are vendors — keep vendor-vendor auto-suppress; do not mint couple records for businesses."
	default:
		if conf >= 0.9 {
			return fmt.Sprintf("High-score (%.0f%%) couple was suppressed — high-priority FP; sample for vision/identity calibration.", conf*100)
		}
		return "Suppressed prospect — retain reason on board; autopsy aggregates these for weekly review."
	}
}

func lessonForIgnore(actionType string, conf float64) string {
	switch actionType {
	case "review", "create_case", "draft_outreach":
		if conf >= 0.9 {
			return "Create-prospect bar fired but human ignored — verify ad/styled-shoot exclusions and partner-match score."
		}
		return "Outreach-class action ignored — confirm brand-safe path still celebrate-first."
	case "investigate":
		return "Investigate action ignored — may be noise or already known; lower investigate volume if pattern continues."
	case "concierge_review", "pause_automation":
		return "Risk-path action ignored — ensure relationship-state changes stay off the sales path."
	default:
		return "Action ignored by human — logged for weekly false-positive autopsy."
	}
}

// ListAutopsyReports returns newest reports first.
func (s *Store) ListAutopsyReports(limit int) ([]AutopsyReport, error) {
	if limit <= 0 || limit > 50 {
		limit = 12
	}
	rows, err := s.DB.Query(`
		SELECT id, period_start, period_end, summary_json, cases_json, generated_by, created_at
		FROM autopsy_reports
		ORDER BY created_at DESC
		LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []AutopsyReport
	for rows.Next() {
		r, err := scanAutopsy(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// GetAutopsyReport loads one report by id.
func (s *Store) GetAutopsyReport(id string) (AutopsyReport, error) {
	row := s.DB.QueryRow(`
		SELECT id, period_start, period_end, summary_json, cases_json, generated_by, created_at
		FROM autopsy_reports WHERE id = $1`, id)
	return scanAutopsy(row)
}

func scanAutopsy(row interface{ Scan(dest ...any) error }) (AutopsyReport, error) {
	var r AutopsyReport
	var sumRaw, casesRaw string
	err := row.Scan(&r.ID, &r.PeriodStart, &r.PeriodEnd, &sumRaw, &casesRaw, &r.GeneratedBy, &r.CreatedAt)
	if err != nil {
		return r, err
	}
	_ = json.Unmarshal([]byte(sumRaw), &r.Summary)
	_ = json.Unmarshal([]byte(casesRaw), &r.Cases)
	if r.Cases == nil {
		r.Cases = []AutopsyCase{}
	}
	return r, nil
}
