package store

import (
	"database/sql"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"neptune-social-radar/backend/internal/ontology"
)

// SourcePost is a vendor observation shaped for the Sources detail UI.
type SourcePost struct {
	ID         string   `json:"id"`
	Monitor    string   `json:"monitor"`
	Handle     string   `json:"handle"`
	Caption    string   `json:"caption,omitempty"`
	URL        string   `json:"url,omitempty"`
	ImageURL   string   `json:"image_url,omitempty"`
	Tags       []string `json:"tags,omitempty"`
	Mentions   []string `json:"mentions,omitempty"`
	Location   string   `json:"location,omitempty"`
	ObservedAt string   `json:"observed_at"`
}

// ListSourcePosts returns recent posts attributed to a watched vendor handle.
func (s *Store) ListSourcePosts(handle string, limit int) ([]SourcePost, error) {
	if limit <= 0 || limit > 100 {
		limit = 40
	}
	monitor := "vendor:" + handle
	rows, err := s.DB.Query(
		`SELECT id, monitor, raw_payload, observed_at
		 FROM social_observations
		 WHERE observation_type = 'post'
		   AND (monitor = $1 OR lower(raw_payload::jsonb->>'handle') = lower($2))
		 ORDER BY observed_at DESC, id DESC
		 LIMIT `+strconv.Itoa(limit),
		monitor, handle,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []SourcePost
	for rows.Next() {
		var id, mon, raw string
		var at time.Time
		if err := rows.Scan(&id, &mon, &raw, &at); err != nil {
			return nil, err
		}
		p := SourcePost{ID: id, Monitor: mon, ObservedAt: at.UTC().Format(time.RFC3339)}
		var payload map[string]any
		_ = json.Unmarshal([]byte(raw), &payload)
		p.Handle, _ = payload["handle"].(string)
		if p.Handle == "" {
			p.Handle = handle
		}
		p.Caption, _ = payload["caption"].(string)
		p.URL, _ = payload["url"].(string)
		p.ImageURL, _ = payload["image_url"].(string)
		if p.ImageURL == "" {
			p.ImageURL, _ = payload["display_url"].(string)
		}
		p.Location, _ = payload["location"].(string)
		p.Tags = stringSlice(payload["tags"])
		p.Mentions = stringSlice(payload["provider_mentions"])
		if len(p.Mentions) == 0 {
			p.Mentions = stringSlice(payload["mentions"])
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// CountSourcePosts returns how many posts we have stored for a vendor handle.
func (s *Store) CountSourcePosts(handle string) (int, error) {
	var n int
	err := s.DB.QueryRow(
		`SELECT COUNT(*) FROM social_observations
		 WHERE observation_type = 'post'
		   AND (monitor = $1 OR lower(raw_payload::jsonb->>'handle') = lower($2))`,
		"vendor:"+handle, handle,
	).Scan(&n)
	return n, err
}

// SourceLastPostAt is the newest observation time for a vendor (or zero).
func (s *Store) SourceLastPostAt(handle string) (*time.Time, error) {
	var t sql.NullTime
	err := s.DB.QueryRow(
		`SELECT MAX(observed_at) FROM social_observations
		 WHERE observation_type = 'post'
		   AND (monitor = $1 OR lower(raw_payload::jsonb->>'handle') = lower($2))`,
		"vendor:"+handle, handle,
	).Scan(&t)
	if err != nil || !t.Valid {
		return nil, err
	}
	tt := t.Time.UTC()
	return &tt, nil
}

func stringSlice(v any) []string {
	arr, ok := v.([]any)
	if !ok {
		if ss, ok := v.([]string); ok {
			return ss
		}
		return nil
	}
	var out []string
	for _, x := range arr {
		if s, ok := x.(string); ok && s != "" {
			out = append(out, s)
		}
	}
	return out
}

// ProspectCard is one Kanban card on the prospect board / workbench.
type ProspectCard struct {
	CoupleID          string  `json:"couple_id"`
	Column            string  `json:"column"`
	PersonALabel      string  `json:"person_a_label"`
	PersonBLabel      string  `json:"person_b_label"`
	HandleA           string  `json:"handle_a,omitempty"`
	HandleB           string  `json:"handle_b,omitempty"`
	ProfilePicA       string  `json:"profile_pic_a,omitempty"`
	ProfilePicB       string  `json:"profile_pic_b,omitempty"`
	BioA              string  `json:"bio_a,omitempty"`
	BioB              string  `json:"bio_b,omitempty"`
	Stage             string  `json:"stage,omitempty"`
	Confidence        float64 `json:"confidence,omitempty"`
	HypothesisScore   float64 `json:"hypothesis_score,omitempty"`
	PendingActionID   string  `json:"pending_action_id,omitempty"`
	PendingActionType string  `json:"pending_action_type,omitempty"`
	ProposedPayload   string  `json:"proposed_payload,omitempty"`
	City              string  `json:"city,omitempty"`
	Region            string  `json:"region,omitempty"`
	AutomationPaused  bool    `json:"automation_paused"`
	HasCase           bool    `json:"has_case"`
	NeedsPics         bool    `json:"needs_pics"`
	NeedsLocation     bool    `json:"needs_location"`
	NeedsAction       bool    `json:"needs_action"`
	CreatedAt         string  `json:"created_at"`
}

// Prospect columns used by the board UI.
const (
	ColTaggedPair     = "tagged_pair"
	ColInvestigating  = "investigating"
	ColEngagedSignal  = "engaged_signal"
	ColReadyOutreach  = "ready_outreach"
	ColApprovedPaused = "approved_paused"
)

// ListProspectBoard projects couples into stage columns for the Kanban UI.
// Suppressed couples are excluded.
func (s *Store) ListProspectBoard(limit int) ([]ProspectCard, error) {
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	rows, err := s.DB.Query(`
		SELECT
		  c.id, c.person_a_id, c.person_b_id, c.created_at,
		  COALESCE(c.inferred_city,''), COALESCE(c.inferred_region,''),
		  COALESCE(pa.display_name,''), COALESCE(pb.display_name,''),
		  COALESCE(aa.handle,''), COALESCE(ab.handle,''),
		  COALESCE(aa.profile_pic_url,''), COALESCE(ab.profile_pic_url,''),
		  COALESCE(aa.bio_text,''), COALESCE(ab.bio_text,''),
		  COALESCE(r.stage,'unknown'), COALESCE(r.confidence,0), COALESCE(r.automation_paused,false),
		  COALESCE(h.confidence,0),
		  COALESCE(act.id,''), COALESCE(act.action_type,''), COALESCE(act.proposed_payload,''),
		  EXISTS(SELECT 1 FROM neptune_cases nc WHERE nc.couple_id = c.id) AS has_case
		FROM couples c
		LEFT JOIN persons pa ON pa.id = c.person_a_id
		LEFT JOIN persons pb ON pb.id = c.person_b_id
		LEFT JOIN LATERAL (
		  SELECT handle, profile_pic_url, bio_text FROM social_accounts WHERE person_id = c.person_a_id LIMIT 1
		) aa ON true
		LEFT JOIN LATERAL (
		  SELECT handle, profile_pic_url, bio_text FROM social_accounts WHERE person_id = c.person_b_id LIMIT 1
		) ab ON true
		LEFT JOIN LATERAL (
		  SELECT stage, confidence, automation_paused FROM relationships
		  WHERE couple_id = c.id AND effective_to IS NULL
		  ORDER BY effective_from DESC LIMIT 1
		) r ON true
		LEFT JOIN LATERAL (
		  SELECT confidence FROM life_event_hypotheses
		  WHERE couple_id = c.id
		  ORDER BY created_at DESC LIMIT 1
		) h ON true
		LEFT JOIN LATERAL (
		  SELECT ra.id, ra.action_type, ra.proposed_payload
		  FROM recommended_actions ra
		  JOIN life_event_hypotheses lh ON lh.id = ra.hypothesis_id
		  WHERE lh.couple_id = c.id AND ra.status = 'pending'
		  ORDER BY ra.created_at DESC LIMIT 1
		) act ON true
		WHERE c.suppressed_at IS NULL
		ORDER BY
		  CASE WHEN act.id IS NOT NULL AND act.id <> '' THEN 0 ELSE 1 END,
		  COALESCE(h.confidence, r.confidence, 0) DESC,
		  c.created_at DESC
		LIMIT `+strconv.Itoa(limit))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []ProspectCard
	for rows.Next() {
		var card ProspectCard
		var personAID, personBID string
		var created time.Time
		var stage string
		var conf, hypConf float64
		var paused, hasCase bool
		if err := rows.Scan(
			&card.CoupleID, &personAID, &personBID, &created,
			&card.City, &card.Region,
			&card.PersonALabel, &card.PersonBLabel,
			&card.HandleA, &card.HandleB,
			&card.ProfilePicA, &card.ProfilePicB,
			&card.BioA, &card.BioB,
			&stage, &conf, &paused,
			&hypConf,
			&card.PendingActionID, &card.PendingActionType, &card.ProposedPayload,
			&hasCase,
		); err != nil {
			return nil, err
		}
		card.Stage = stage
		card.Confidence = conf
		card.HypothesisScore = hypConf
		card.AutomationPaused = paused
		card.HasCase = hasCase
		card.CreatedAt = created.UTC().Format(time.RFC3339)
		card.NeedsPics = card.ProfilePicA == "" || card.ProfilePicB == ""
		card.NeedsLocation = card.City == ""
		card.NeedsAction = card.PendingActionID != ""
		if card.PersonALabel == "" || strings.EqualFold(card.PersonALabel, card.HandleA) {
			if card.HandleA != "" {
				card.PersonALabel = card.HandleA
			}
		}
		if card.PersonBLabel == "" || strings.EqualFold(card.PersonBLabel, card.HandleB) {
			if card.HandleB != "" {
				card.PersonBLabel = card.HandleB
			}
		}
		// Truncate bios for payload size
		card.BioA = truncateStr(card.BioA, 220)
		card.BioB = truncateStr(card.BioB, 220)
		card.Column = classifyProspectColumn(card)
		out = append(out, card)
	}
	return out, rows.Err()
}

func truncateStr(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}

// classifyProspectColumn balances the board so "engaged" stage alone does not
// dump everyone into one column — pending actions and score drive placement.
func classifyProspectColumn(c ProspectCard) string {
	if c.HasCase || c.AutomationPaused || c.PendingActionType == "pause_automation" || c.PendingActionType == "concierge_review" {
		return ColApprovedPaused
	}
	switch c.PendingActionType {
	case "review", "create_case", "draft_outreach":
		return ColReadyOutreach
	case "investigate":
		return ColInvestigating
	}

	score := c.HypothesisScore
	if c.Confidence > score {
		score = c.Confidence
	}

	switch ontology.RelationshipStage(c.Stage) {
	case ontology.StageEngaged, ontology.StageMarried:
		// High score engaged → ready or engaged column; weak score → investigate
		if score >= 0.9 {
			return ColReadyOutreach
		}
		if score >= 0.7 {
			return ColEngagedSignal
		}
		// Noisy auto-engaged without strong confidence
		return ColInvestigating
	}

	if score >= 0.9 {
		return ColReadyOutreach
	}
	if score >= 0.55 || c.Stage == string(ontology.StageDatingSuspected) {
		return ColInvestigating
	}
	return ColTaggedPair
}

// SuppressCouple marks a false-positive pair so it never reappears on the board.
func (s *Store) SuppressCouple(coupleID, reason string) error {
	if reason == "" {
		reason = "not_a_couple"
	}
	_, err := s.DB.Exec(
		`UPDATE couples SET suppressed_at = now(), suppressed_reason = $2 WHERE id = $1`,
		coupleID, reason,
	)
	return err
}

// SuppressVendorVendorCouples hides board couples where both handles are registered sources.
// Returns how many were suppressed. registeredHandles is a lowercase set of watched-source handles.
func (s *Store) SuppressVendorVendorCouples(registeredHandles map[string]bool) (int, error) {
	cards, err := s.ListProspectBoard(500)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, c := range cards {
		if c.HandleA == "" || c.HandleB == "" {
			continue
		}
		a := strings.ToLower(c.HandleA)
		b := strings.ToLower(c.HandleB)
		if registeredHandles[a] && registeredHandles[b] {
			if err := s.SuppressCouple(c.CoupleID, "vendor_vendor_pair"); err == nil {
				n++
			}
		}
	}
	return n, nil
}

// UnsuppressCouple restores a couple to the board.
func (s *Store) UnsuppressCouple(coupleID string) error {
	_, err := s.DB.Exec(
		`UPDATE couples SET suppressed_at = NULL, suppressed_reason = NULL WHERE id = $1`,
		coupleID,
	)
	return err
}

// ListAccountsNeedingProfile returns social accounts missing pics or stale bios,
// limited for budget-aware enrich jobs.
func (s *Store) ListAccountsNeedingProfile(limit int) ([]ontology.SocialAccount, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	rows, err := s.DB.Query(
		`SELECT id, COALESCE(person_id,''), platform, handle, COALESCE(display_name,''), COALESCE(bio_text,''),
		        is_private, is_disabled, last_seen_at, COALESCE(profile_pic_url,''), follower_count, following_count,
		        profile_checked_at, COALESCE(inferred_city,''), COALESCE(inferred_region,''), COALESCE(location_source,'')
		 FROM social_accounts
		 WHERE handle IS NOT NULL AND handle <> ''
		   AND (
		     profile_pic_url IS NULL OR profile_pic_url = ''
		     OR profile_checked_at IS NULL
		     OR profile_checked_at < now() - interval '7 days'
		   )
		   AND person_id IS NOT NULL
		   AND person_id IN (SELECT person_a_id FROM couples WHERE suppressed_at IS NULL
		                     UNION SELECT person_b_id FROM couples WHERE suppressed_at IS NULL)
		 ORDER BY profile_checked_at NULLS FIRST
		 LIMIT `+strconv.Itoa(limit))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ontology.SocialAccount
	for rows.Next() {
		a, err := scanAccountRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func scanAccountRow(row interface{ Scan(dest ...any) error }) (ontology.SocialAccount, error) {
	var a ontology.SocialAccount
	var lastSeen, checkedAt sql.NullTime
	var followers, following sql.NullInt64
	err := row.Scan(
		&a.ID, &a.PersonID, &a.Platform, &a.Handle, &a.DisplayName, &a.BioText,
		&a.IsPrivate, &a.IsDisabled, &lastSeen, &a.ProfilePicURL, &followers, &following,
		&checkedAt, &a.InferredCity, &a.InferredRegion, &a.LocationSource,
	)
	if err != nil {
		return a, err
	}
	if lastSeen.Valid {
		a.LastSeenAt = lastSeen.Time
	}
	if checkedAt.Valid {
		t := checkedAt.Time
		a.ProfileCheckedAt = &t
	}
	if followers.Valid {
		n := int(followers.Int64)
		a.FollowerCount = &n
	}
	if following.Valid {
		n := int(following.Int64)
		a.FollowingCount = &n
	}
	return a, nil
}

// ListCouplesMissingLocation for geo backfill.
func (s *Store) ListCouplesMissingLocation(limit int) ([]ontology.Couple, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := s.DB.Query(
		`SELECT id, person_a_id, person_b_id, created_at,
		        COALESCE(inferred_city,''), COALESCE(inferred_region,''),
		        inferred_lat, inferred_lng, COALESCE(location_source,'')
		 FROM couples
		 WHERE suppressed_at IS NULL
		   AND (inferred_city IS NULL OR inferred_city = '')
		 ORDER BY created_at DESC
		 LIMIT `+strconv.Itoa(limit))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ontology.Couple
	for rows.Next() {
		var c ontology.Couple
		var lat, lng sql.NullFloat64
		if err := rows.Scan(&c.ID, &c.PersonAID, &c.PersonBID, &c.CreatedAt,
			&c.InferredCity, &c.InferredRegion, &lat, &lng, &c.LocationSource); err != nil {
			return nil, err
		}
		if lat.Valid {
			v := lat.Float64
			c.InferredLat = &v
		}
		if lng.Valid {
			v := lng.Float64
			c.InferredLng = &v
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// ProspectPin is a map marker for a couple with inferred location.
type ProspectPin struct {
	CoupleID     string   `json:"couple_id"`
	PersonALabel string   `json:"person_a_label"`
	PersonBLabel string   `json:"person_b_label"`
	HandleA      string   `json:"handle_a,omitempty"`
	HandleB      string   `json:"handle_b,omitempty"`
	ProfilePicA  string   `json:"profile_pic_a,omitempty"`
	ProfilePicB  string   `json:"profile_pic_b,omitempty"`
	City         string   `json:"city"`
	Region       string   `json:"region,omitempty"`
	Lat          *float64 `json:"lat,omitempty"`
	Lng          *float64 `json:"lng,omitempty"`
	Stage        string   `json:"stage,omitempty"`
	Column       string   `json:"column,omitempty"`
}

// ListProspectPins returns couples that have an inferred city (and optional lat/lng).
func (s *Store) ListProspectPins() ([]ProspectPin, error) {
	rows, err := s.DB.Query(`
		SELECT
		  c.id, COALESCE(c.inferred_city,''), COALESCE(c.inferred_region,''),
		  c.inferred_lat, c.inferred_lng,
		  COALESCE(pa.display_name,''), COALESCE(pb.display_name,''),
		  COALESCE(aa.handle,''), COALESCE(ab.handle,''),
		  COALESCE(aa.profile_pic_url,''), COALESCE(ab.profile_pic_url,''),
		  COALESCE(r.stage,'unknown')
		FROM couples c
		LEFT JOIN persons pa ON pa.id = c.person_a_id
		LEFT JOIN persons pb ON pb.id = c.person_b_id
		LEFT JOIN LATERAL (
		  SELECT handle, profile_pic_url FROM social_accounts WHERE person_id = c.person_a_id LIMIT 1
		) aa ON true
		LEFT JOIN LATERAL (
		  SELECT handle, profile_pic_url FROM social_accounts WHERE person_id = c.person_b_id LIMIT 1
		) ab ON true
		LEFT JOIN LATERAL (
		  SELECT stage FROM relationships WHERE couple_id = c.id AND effective_to IS NULL
		  ORDER BY effective_from DESC LIMIT 1
		) r ON true
		WHERE c.suppressed_at IS NULL
		  AND c.inferred_city IS NOT NULL AND c.inferred_city <> ''
		ORDER BY c.created_at DESC
		LIMIT 500`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ProspectPin
	for rows.Next() {
		var p ProspectPin
		var lat, lng sql.NullFloat64
		if err := rows.Scan(
			&p.CoupleID, &p.City, &p.Region, &lat, &lng,
			&p.PersonALabel, &p.PersonBLabel,
			&p.HandleA, &p.HandleB, &p.ProfilePicA, &p.ProfilePicB,
			&p.Stage,
		); err != nil {
			return nil, err
		}
		if lat.Valid {
			v := lat.Float64
			p.Lat = &v
		}
		if lng.Valid {
			v := lng.Float64
			p.Lng = &v
		}
		if p.PersonALabel == "" {
			p.PersonALabel = p.HandleA
		}
		if p.PersonBLabel == "" {
			p.PersonBLabel = p.HandleB
		}
		card := ProspectCard{Stage: p.Stage, City: p.City}
		p.Column = classifyProspectColumn(card)
		out = append(out, p)
	}
	return out, rows.Err()
}

// UpdateCoupleLocation writes inferred geo onto a couple row.
func (s *Store) UpdateCoupleLocation(coupleID, city, region, source string, lat, lng *float64) error {
	_, err := s.DB.Exec(
		`UPDATE couples SET inferred_city = $1, inferred_region = $2, location_source = $3,
		  inferred_lat = $4, inferred_lng = $5 WHERE id = $6`,
		nullIfEmpty(city), nullIfEmpty(region), nullIfEmpty(source), lat, lng, coupleID,
	)
	return err
}

// OpsSummary is the Today workbench KPI strip.
type OpsSummary struct {
	CouplesTotal       int `json:"couples_total"`
	Couples24h         int `json:"couples_24h"`
	PendingActions     int `json:"pending_actions"`
	NeedsPics          int `json:"needs_pics"`
	NeedsLocation      int `json:"needs_location"`
	SourcesTotal       int `json:"sources_total"`
	SourcesWithLoc     int `json:"sources_with_loc"`
	SourcesStale       int `json:"sources_stale"` // no posts in 7d or never
	MapPins            int `json:"map_pins"`
	ResultsUsedToday   int `json:"results_used_today"`
}

// GetOpsSummary aggregates operator KPIs.
func (s *Store) GetOpsSummary() (OpsSummary, error) {
	var o OpsSummary
	_ = s.DB.QueryRow(`SELECT COUNT(*) FROM couples WHERE suppressed_at IS NULL`).Scan(&o.CouplesTotal)
	_ = s.DB.QueryRow(`SELECT COUNT(*) FROM couples WHERE suppressed_at IS NULL AND created_at > now() - interval '24 hours'`).Scan(&o.Couples24h)
	_ = s.DB.QueryRow(`SELECT COUNT(*) FROM recommended_actions WHERE status = 'pending'`).Scan(&o.PendingActions)
	_ = s.DB.QueryRow(`
		SELECT COUNT(*) FROM couples c
		WHERE c.suppressed_at IS NULL AND (
		  NOT EXISTS (SELECT 1 FROM social_accounts a WHERE a.person_id = c.person_a_id AND a.profile_pic_url IS NOT NULL AND a.profile_pic_url <> '')
		  OR NOT EXISTS (SELECT 1 FROM social_accounts b WHERE b.person_id = c.person_b_id AND b.profile_pic_url IS NOT NULL AND b.profile_pic_url <> '')
		)`).Scan(&o.NeedsPics)
	_ = s.DB.QueryRow(`SELECT COUNT(*) FROM couples WHERE suppressed_at IS NULL AND (inferred_city IS NULL OR inferred_city = '')`).Scan(&o.NeedsLocation)
	_ = s.DB.QueryRow(`SELECT COUNT(*) FROM watched_sources WHERE active`).Scan(&o.SourcesTotal)
	_ = s.DB.QueryRow(`SELECT COUNT(*) FROM watched_sources WHERE active AND city IS NOT NULL AND city <> ''`).Scan(&o.SourcesWithLoc)
	_ = s.DB.QueryRow(`SELECT COUNT(*) FROM couples WHERE suppressed_at IS NULL AND inferred_city IS NOT NULL AND inferred_city <> ''`).Scan(&o.MapPins)
	used, _ := s.UsageToday("apify")
	o.ResultsUsedToday = used
	// Stale sources: active, no post in last 7 days under vendor:handle or as author
	_ = s.DB.QueryRow(`
		SELECT COUNT(*) FROM watched_sources w
		WHERE w.active AND NOT EXISTS (
		  SELECT 1 FROM social_observations o
		  WHERE o.observation_type = 'post'
		    AND o.observed_at > now() - interval '7 days'
		    AND (o.monitor = 'vendor:' || w.handle OR lower(o.raw_payload::jsonb->>'handle') = lower(w.handle))
		)`).Scan(&o.SourcesStale)
	return o, nil
}
