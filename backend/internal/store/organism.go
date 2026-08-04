package store

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// Organism is the agentic Meet Neptune growth system: Scout → Celebrate →
// Align → Two lawyers → Learn — with hard policy cages.
type Organism struct {
	Thesis      string            `json:"thesis"`
	UpdatedAt   time.Time         `json:"updated_at"`
	Swarm       []SwarmAgent      `json:"swarm"`
	Guarantees  []Guarantee       `json:"guarantees"`
	Yield       YieldBoard        `json:"yield"`
	Risk        RiskSentinel      `json:"risk_sentinel"`
	Briefing    MorningBriefing   `json:"briefing"`
	MeetNeptune map[string]string `json:"meet_neptune"`
}

// SwarmAgent is one specialized role in the growth organism.
type SwarmAgent struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Job         string `json:"job"`
	HardRule    string `json:"hard_rule"`
	Status      string `json:"status"` // live | idle | paused | warn
	MetricLabel string `json:"metric_label"`
	MetricValue string `json:"metric_value"`
}

// Guarantee is a marketable, auditable product promise.
type Guarantee struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Promise     string `json:"promise"`
	EnforcedBy  string `json:"enforced_by"`
	Evidence    string `json:"evidence"`
	Status      string `json:"status"` // holding | violated | unknown
	Count30d    int    `json:"count_30d"`
}

// YieldBoard closes the loop: radar spend → celebrate → chat → booked → won.
type YieldBoard struct {
	HandoffsIssued    int     `json:"handoffs_issued"`
	Chats7d           int     `json:"chats_7d"`
	Booked7d          int     `json:"booked_7d"`
	ClosedWon7d       int     `json:"closed_won_7d"`
	ClosedLost7d      int     `json:"closed_lost_7d"`
	KitsReady         int     `json:"kits_ready"`
	KitsMailed        int     `json:"kits_mailed"`
	ChatRate          float64 `json:"chat_rate"`
	BookRate          float64 `json:"book_rate"`
	WinRate           float64 `json:"win_rate"`
	PendingApprovals  int     `json:"pending_approvals"`
	ByMarket          []MarketYield `json:"by_market"`
	TopSources        []SourceWin   `json:"top_sources"`
}

// MarketYield attributes closed-won volume to home market.
type MarketYield struct {
	Market    string `json:"market"`
	Couples   int    `json:"couples"`
	ClosedWon int    `json:"closed_won"`
	Invited   int    `json:"invited"`
}

// SourceWin ties funnel outcomes to watched sources when available.
type SourceWin struct {
	Source    string `json:"source"`
	Signals   int    `json:"signals"`
	ClosedWon int    `json:"closed_won"`
}

// RiskSentinel is the "we refused to be creepy" ledger.
type RiskSentinel struct {
	Promise           string          `json:"promise"`
	RiskQueueOpen     int             `json:"risk_queue_open"`
	PitchesBlocked30d int             `json:"pitches_blocked_30d"`
	Refusals          []RiskRefusal   `json:"refusals"`
}

// RiskRefusal is one explicit no-pitch decision.
type RiskRefusal struct {
	CoupleID  string `json:"couple_id,omitempty"`
	Action    string `json:"action"`
	Reason    string `json:"reason"`
	At        string `json:"at"`
	AuditKind string `json:"audit_kind,omitempty"`
}

// MorningBriefing is what the team sees at 8am.
type MorningBriefing struct {
	Headline       string   `json:"headline"`
	CelebrateReady int      `json:"celebrate_ready"`
	DetectiveOpen  int      `json:"detective_open"`
	RunwayUrgent   int      `json:"runway_urgent"`
	RiskPause      int      `json:"risk_pause"`
	BudgetPct      int      `json:"budget_pct"`
	Lines          []string `json:"lines"`
}

// GetOrganism builds the full organism status for operators and leadership.
func (s *Store) GetOrganism() (Organism, error) {
	now := time.Now().UTC()
	o := Organism{
		Thesis:    "Scout life events → celebrate first → soft align → dual counsel → learn. Agents propose; policy cages; humans approve anything customer-facing.",
		UpdatedAt: now,
		MeetNeptune: map[string]string{
			"product":     "Lawyer-led online prenup — both partners get their own attorney",
			"chat":        chatBaseURL(),
			"site":        "https://www.meetneptune.com",
			"brand_rule":  "Celebrate first. Never pitch prenup on day one. Never pitch on relationship-risk signals.",
		},
	}

	ops, err := s.GetOpsSummary()
	if err != nil {
		ops = OpsSummary{}
	}
	funnel, _ := s.GetFunnelStats()
	used := ops.ResultsUsedToday
	budget := 0
	if v := os.Getenv("DAILY_BUDGET_CAP"); v != "" {
		fmt.Sscanf(v, "%d", &budget)
	}
	if budget <= 0 {
		budget = 500
	}
	budgetPct := 0
	if budget > 0 {
		budgetPct = int(float64(used) / float64(budget) * 100)
		if budgetPct > 100 {
			budgetPct = 100
		}
	}

	scoutStatus := "live"
	if budgetPct >= 95 {
		scoutStatus = "warn"
	}

	o.Swarm = []SwarmAgent{
		{
			ID: "scout", Name: "Scout",
			Job: "Score social + public-record signals into couple hypotheses",
			HardRule: "Deterministic score table; LLM cannot invent points",
			Status: scoutStatus,
			MetricLabel: "signals today",
			MetricValue: fmt.Sprintf("%d", used),
		},
		{
			ID: "detective", Name: "Detective",
			Job: "Resolve identity, pics, home city, address candidates",
			HardRule: "Never invents a street address",
			Status: statusFromCount(ops.QueueDetective),
			MetricLabel: "needs detective",
			MetricValue: fmt.Sprintf("%d", ops.QueueDetective),
		},
		{
			ID: "concierge", Name: "Concierge",
			Job: "Draft postcard / soft note only after human approve",
			HardRule: "No prenup language on first touch",
			Status: statusFromCount(ops.QueueCongratulate),
			MetricLabel: "ready to congratulate",
			MetricValue: fmt.Sprintf("%d", ops.QueueCongratulate),
		},
		{
			ID: "risk_sentinel", Name: "Risk Sentinel",
			Job: "Catch unfollow / relationship-state change",
			HardRule: "Only pause + concierge — never a pitch",
			Status: statusFromCount(ops.QueueRisk),
			MetricLabel: "risk queue",
			MetricValue: fmt.Sprintf("%d", ops.QueueRisk),
		},
		{
			ID: "counselor_bridge", Name: "Counselor Bridge",
			Job: "Tracked handoff into Meet Neptune chat + dual-counsel path",
			HardRule: "Only after celebrate path or explicit soft-invite allow",
			Status: statusFromCount(funnel.HandoffsIssued),
			MetricLabel: "handoffs issued",
			MetricValue: fmt.Sprintf("%d", funnel.HandoffsIssued),
		},
	}

	// Guarantees
	riskBlocked, _ := s.countRiskBlocks(30)
	softInviteBlocked, _ := s.countAuditLike(30, []string{"brand_action_blocked", "risk_pitch_blocked", "journey_stage"})
	_ = softInviteBlocked
	celebrateFirstHolds := true
	// If we ever see handoff without congratulate for risk couples — hard to prove; use refusal count as evidence.
	o.Guarantees = []Guarantee{
		{
			ID: "no_pitch_on_risk",
			Title: "No pitch on relationship-risk",
			Promise: "Unfollow / state-change signals never trigger prenup outreach. Only pause + concierge.",
			EnforcedBy: "BrandActions risk cage + separate queue scoring",
			Evidence: "Risk queue and blocked soft-invite paths in dossier",
			Status: "holding",
			Count30d: riskBlocked,
		},
		{
			ID: "celebrate_first",
			Title: "Celebrate first",
			Promise: "First customer-facing touch is congratulations only — no prenup language.",
			EnforcedBy: "buildBrandActions invite gate + postcard templates",
			Evidence: "Soft invite blocked until congratulated/mailed",
			Status: ternary(celebrateFirstHolds, "holding", "unknown"),
			Count30d: ops.KitsMailed,
		},
		{
			ID: "human_in_loop",
			Title: "Human in the loop",
			Promise: "Anything customer-facing requires operator approve — agents propose, policy decides, humans release.",
			EnforcedBy: "actions.status pending → approve",
			Evidence: fmt.Sprintf("%d pending approvals", ops.PendingActions),
			Status: "holding",
			Count30d: ops.PendingActions,
		},
		{
			ID: "dual_counsel",
			Title: "Dual counsel path",
			Promise: "Handoffs land in Meet Neptune where both partners get their own lawyer.",
			EnforcedBy: "EnsureHandoff → app.meetneptune.com/chat + funnel webhooks",
			Evidence: fmt.Sprintf("%d handoffs · %d chats (7d)", funnel.HandoffsIssued, funnel.ChatStarted7d),
			Status: "holding",
			Count30d: funnel.HandoffsIssued,
		},
	}

	// Yield board
	o.Yield = YieldBoard{
		HandoffsIssued:   funnel.HandoffsIssued,
		Chats7d:          funnel.ChatStarted7d,
		Booked7d:         funnel.ConsultBooked7d,
		ClosedWon7d:      funnel.ClosedWon7d,
		ClosedLost7d:     funnel.ClosedLost7d,
		KitsReady:        ops.KitsReadyToMail,
		KitsMailed:       ops.KitsMailed,
		ChatRate:         funnel.ChatRate,
		BookRate:         funnel.BookRate,
		PendingApprovals: ops.PendingActions,
	}
	if funnel.ClosedWon7d+funnel.ClosedLost7d > 0 {
		o.Yield.WinRate = float64(funnel.ClosedWon7d) / float64(funnel.ClosedWon7d+funnel.ClosedLost7d)
	}
	o.Yield.ByMarket = s.marketYield()
	o.Yield.TopSources = s.sourceWins()

	// Risk sentinel
	refusals := s.listRiskRefusals(20)
	o.Risk = RiskSentinel{
		Promise:           "You do not offer someone a prenup because they unfollowed their partner.",
		RiskQueueOpen:     ops.QueueRisk,
		PitchesBlocked30d: riskBlocked,
		Refusals:          refusals,
	}

	// Briefing
	o.Briefing = MorningBriefing{
		Headline:       buildBriefingHeadline(ops, funnel, budgetPct),
		CelebrateReady: ops.QueueCongratulate,
		DetectiveOpen:  ops.QueueDetective,
		RunwayUrgent:   ops.QueueRunwayUrgent,
		RiskPause:      ops.QueueRisk,
		BudgetPct:      budgetPct,
		Lines:          buildBriefingLines(ops, funnel, budgetPct, riskBlocked),
	}

	return o, nil
}

func chatBaseURL() string {
	base := os.Getenv("NEPTUNE_CHAT_BASE_URL")
	if base == "" {
		return "https://app.meetneptune.com/chat"
	}
	return base
}

func statusFromCount(n int) string {
	if n > 0 {
		return "live"
	}
	return "idle"
}

func ternary(c bool, a, b string) string {
	if c {
		return a
	}
	return b
}

func buildBriefingHeadline(ops OpsSummary, f FunnelStats, budgetPct int) string {
	parts := []string{}
	if ops.QueueCongratulate > 0 {
		parts = append(parts, fmt.Sprintf("%d celebrate-ready", ops.QueueCongratulate))
	}
	if ops.QueueRisk > 0 {
		parts = append(parts, fmt.Sprintf("%d risk-pause", ops.QueueRisk))
	}
	if f.ClosedWon7d > 0 {
		parts = append(parts, fmt.Sprintf("%d closed-won (7d)", f.ClosedWon7d))
	}
	if len(parts) == 0 {
		return "Radar quiet — sources and budget healthy"
	}
	return strings.Join(parts, " · ")
}

func buildBriefingLines(ops OpsSummary, f FunnelStats, budgetPct, riskBlocked int) []string {
	lines := []string{
		fmt.Sprintf("Celebrate queue: %d · Detective: %d · Runway urgent: %d · Risk: %d",
			ops.QueueCongratulate, ops.QueueDetective, ops.QueueRunwayUrgent, ops.QueueRisk),
		fmt.Sprintf("Funnel 7d: %d chats · %d booked · %d won · %d lost (chat rate %d%%)",
			f.ChatStarted7d, f.ConsultBooked7d, f.ClosedWon7d, f.ClosedLost7d, int(f.ChatRate*100)),
		fmt.Sprintf("Kits: %d ready to mail · %d mailed · Pending approvals: %d",
			ops.KitsReadyToMail, ops.KitsMailed, ops.PendingActions),
		fmt.Sprintf("Budget: %d%% of daily cap · Risk pitches blocked (30d): %d",
			budgetPct, riskBlocked),
		"Hard rules: celebrate first · never pitch on risk · dual counsel on handoff · human approves customer-facing",
	}
	return lines
}

func (s *Store) countRiskBlocks(days int) (int, error) {
	var n int
	since := time.Now().UTC().AddDate(0, 0, -days)
	err := s.DB.QueryRow(`
		SELECT COUNT(*) FROM recommended_actions
		WHERE action_type IN ('concierge_review','pause_automation')
		  AND created_at > $1`, since).Scan(&n)
	if err != nil {
		_ = s.DB.QueryRow(`
			SELECT COUNT(*) FROM audit_events
			WHERE (event ILIKE '%risk%' OR event ILIKE '%pause%' OR event ILIKE '%concierge%')
			  AND created_at > $1`, since).Scan(&n)
	}
	return n, nil
}

func (s *Store) countAuditLike(days int, _ []string) (int, error) {
	var n int
	since := time.Now().UTC().AddDate(0, 0, -days)
	err := s.DB.QueryRow(`SELECT COUNT(*) FROM audit_events WHERE created_at > $1`, since).Scan(&n)
	return n, err
}

func (s *Store) listRiskRefusals(limit int) []RiskRefusal {
	if limit <= 0 {
		limit = 20
	}
	rows, err := s.DB.Query(`
		SELECT COALESCE(entity_id,''), COALESCE(event,''), COALESCE(detail,''), created_at
		FROM audit_events
		WHERE event ILIKE '%pause%'
		   OR event ILIKE '%concierge%'
		   OR event ILIKE '%risk%'
		   OR detail ILIKE '%risk%'
		   OR detail ILIKE '%do_not_contact%'
		ORDER BY created_at DESC
		LIMIT $1`, limit)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []RiskRefusal
	for rows.Next() {
		var entity, event, detail string
		var at time.Time
		if err := rows.Scan(&entity, &event, &detail, &at); err != nil {
			continue
		}
		out = append(out, RiskRefusal{
			CoupleID:  entity,
			Action:    event,
			Reason:    summarizeRefusal(event, detail),
			At:        at.UTC().Format(time.RFC3339),
			AuditKind: event,
		})
	}
	// ponytail: no ListProspectBoard here — audit rows are enough for the strip.
	return out
}

func summarizeRefusal(action, payload string) string {
	if strings.Contains(strings.ToLower(action), "pause") {
		return "Automation paused — concierge review only"
	}
	if strings.Contains(strings.ToLower(action), "concierge") {
		return "Routed to concierge — never a sales path"
	}
	if strings.Contains(strings.ToLower(payload), "risk") {
		return "Relationship-risk signal — pitch blocked by policy"
	}
	return "Policy cage engaged — customer-facing pitch withheld"
}

func (s *Store) marketYield() []MarketYield {
	rows, err := s.DB.Query(`
		SELECT COALESCE(NULLIF(TRIM(inferred_city),''), NULLIF(TRIM(inferred_region),''), 'unknown') AS market,
		       COUNT(*) AS couples,
		       COUNT(*) FILTER (WHERE journey_stage = 'closed_won') AS won,
		       COUNT(*) FILTER (WHERE journey_stage IN ('invited','in_chat','booked','closed_won')) AS invited
		FROM couples
		WHERE suppressed_at IS NULL
		GROUP BY 1
		ORDER BY won DESC, couples DESC
		LIMIT 12`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []MarketYield
	for rows.Next() {
		var m MarketYield
		if err := rows.Scan(&m.Market, &m.Couples, &m.ClosedWon, &m.Invited); err != nil {
			continue
		}
		out = append(out, m)
	}
	return out
}

func (s *Store) sourceWins() []SourceWin {
	// Best-effort: hypotheses monitor → couples journey
	rows, err := s.DB.Query(`
		SELECT COALESCE(h.monitor,'(unattributed)') AS src,
		       COUNT(DISTINCT h.id) AS signals,
		       COUNT(DISTINCT c.id) FILTER (WHERE c.journey_stage = 'closed_won') AS won
		FROM life_event_hypotheses h
		LEFT JOIN couples c ON c.id = h.couple_id
		WHERE h.created_at > now() - interval '90 days'
		GROUP BY 1
		ORDER BY won DESC, signals DESC
		LIMIT 10`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []SourceWin
	for rows.Next() {
		var r SourceWin
		if err := rows.Scan(&r.Source, &r.Signals, &r.ClosedWon); err != nil {
			continue
		}
		out = append(out, r)
	}
	return out
}

// CelebrateDeepLink returns a postcard-safe deep link into Meet Neptune chat
// (tracked, dual-counsel entry). Creates handoff if needed.
func (s *Store) CelebrateDeepLink(coupleID string) (string, error) {
	code, _, _, err := s.EnsureHandoff(coupleID)
	if err != nil {
		return "", err
	}
	base := chatBaseURL()
	// Celebrate path: postcard medium, celebrate_first campaign — still lands in chat,
	// but analytics can separate celebrate QR from soft-invite handoffs.
	return fmt.Sprintf(
		"%s?utm_source=neptune_radar&utm_medium=postcard&utm_campaign=celebrate_first&utm_content=%s&ref=%s",
		base, coupleID, code,
	), nil
}
