package store

import (
	"database/sql"
	"encoding/json"
	"time"
)

// CongratulateKit is a human-reviewed outreach package: dossier, address research,
// and a postcard draft. Nothing is mailed until status is ready_to_mail.
type CongratulateKit struct {
	ID                   string             `json:"id"`
	CoupleID             string             `json:"couple_id"`
	Status               string             `json:"status"`
	HandleA              string             `json:"handle_a,omitempty"`
	HandleB              string             `json:"handle_b,omitempty"`
	PersonAName          string             `json:"person_a_name,omitempty"`
	PersonBName          string             `json:"person_b_name,omitempty"`
	FirstNameA           string             `json:"first_name_a,omitempty"`
	LastNameA            string             `json:"last_name_a,omitempty"`
	FirstNameB           string             `json:"first_name_b,omitempty"`
	LastNameB            string             `json:"last_name_b,omitempty"`
	NameSourceA          string             `json:"name_source_a,omitempty"`
	NameSourceB          string             `json:"name_source_b,omitempty"`
	BioA                 string             `json:"bio_a,omitempty"`
	BioB                 string             `json:"bio_b,omitempty"`
	ProfilePicA          string             `json:"profile_pic_a,omitempty"`
	ProfilePicB          string             `json:"profile_pic_b,omitempty"`
	MarketCity           string             `json:"market_city,omitempty"`
	MarketRegion         string             `json:"market_region,omitempty"`
	MarketSource         string             `json:"market_source,omitempty"`
	SourceHandle         string             `json:"source_handle,omitempty"`
	SourceClass          string             `json:"source_class,omitempty"`
	DiscoveryCaption     string             `json:"discovery_caption,omitempty"`
	DiscoveryImageURL    string             `json:"discovery_image_url,omitempty"`
	DiscoveryPostURL     string             `json:"discovery_post_url,omitempty"`
	Evidence             []string           `json:"evidence,omitempty"`
	ResearchNotes        string             `json:"research_notes,omitempty"`
	ResearchSteps        []ResearchStep     `json:"research_steps,omitempty"`
	AddressLine1         string             `json:"address_line1,omitempty"`
	AddressLine2         string             `json:"address_line2,omitempty"`
	AddressCity          string             `json:"address_city,omitempty"`
	AddressRegion        string             `json:"address_region,omitempty"`
	AddressPostal        string             `json:"address_postal,omitempty"`
	AddressCountry       string             `json:"address_country,omitempty"`
	AddressConfidence    float64            `json:"address_confidence"`
	AddressSource        string             `json:"address_source,omitempty"`
	AddressCandidates    []AddressCandidate `json:"address_candidates,omitempty"`
	Headline             string             `json:"headline,omitempty"`
	BodyMessage          string             `json:"body_message,omitempty"`
	InternalNote         string             `json:"internal_note,omitempty"`
	PostcardHTML         string             `json:"postcard_html,omitempty"`
	MailPayload          map[string]any     `json:"mail_payload,omitempty"`
	VerifiedBy           string             `json:"verified_by,omitempty"`
	VerifiedAt           *time.Time         `json:"verified_at,omitempty"`
	MailedAt             *time.Time         `json:"mailed_at,omitempty"`
	PriorityScore        float64            `json:"priority_score"`
	FollowUpAt           *time.Time         `json:"follow_up_at,omitempty"`
	FollowUpTemplate     string             `json:"follow_up_template,omitempty"`
	FollowUpSentAt       *time.Time         `json:"follow_up_sent_at,omitempty"`
	FollowUpCount        int                `json:"follow_up_count"`
	CreatedAt            time.Time          `json:"created_at"`
	UpdatedAt            time.Time          `json:"updated_at"`
}

type ResearchStep struct {
	ID      string `json:"id"`
	Label   string `json:"label"`
	Detail  string `json:"detail"`
	Status  string `json:"status"` // done | suggested | blocked
	URL     string `json:"url,omitempty"`
}

type AddressCandidate struct {
	Line1      string  `json:"line1,omitempty"`
	Line2      string  `json:"line2,omitempty"`
	City       string  `json:"city,omitempty"`
	Region     string  `json:"region,omitempty"`
	Postal     string  `json:"postal,omitempty"`
	Country    string  `json:"country,omitempty"`
	Confidence float64 `json:"confidence"`
	Source     string  `json:"source"`
	Note       string  `json:"note,omitempty"`
}

func (s *Store) UpsertCongratulateKit(k CongratulateKit) (CongratulateKit, error) {
	if k.ID == "" {
		k.ID = NewID("kit")
	}
	if k.Status == "" {
		k.Status = "draft"
	}
	if k.AddressCountry == "" {
		k.AddressCountry = "US"
	}
	now := time.Now().UTC()
	k.UpdatedAt = now
	if k.CreatedAt.IsZero() {
		k.CreatedAt = now
	}
	ev, _ := json.Marshal(k.Evidence)
	if k.Evidence == nil {
		ev = []byte("[]")
	}
	steps, _ := json.Marshal(k.ResearchSteps)
	if k.ResearchSteps == nil {
		steps = []byte("[]")
	}
	cands, _ := json.Marshal(k.AddressCandidates)
	if k.AddressCandidates == nil {
		cands = []byte("[]")
	}
	mail, _ := json.Marshal(k.MailPayload)
	if k.MailPayload == nil {
		mail = []byte("null")
	}

	_, err := s.DB.Exec(`
		INSERT INTO congratulate_kits (
			id, couple_id, status, handle_a, handle_b, person_a_name, person_b_name,
			bio_a, bio_b, profile_pic_a, profile_pic_b,
			market_city, market_region, market_source, source_handle, source_class,
			discovery_caption, discovery_image_url, discovery_post_url,
			evidence_json, research_notes, research_steps_json,
			address_line1, address_line2, address_city, address_region, address_postal, address_country,
			address_confidence, address_source, address_candidates_json,
			headline, body_message, internal_note, postcard_html, mail_payload_json,
			verified_by, verified_at, mailed_at,
			priority_score, follow_up_at, follow_up_template, follow_up_sent_at, follow_up_count,
			created_at, updated_at
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,
			$23,$24,$25,$26,$27,$28,$29,$30,$31,$32,$33,$34,$35,$36,$37,$38,$39,$40,$41,
			$42,$43,$44,$45,$46,$47
		)
		ON CONFLICT (id) DO UPDATE SET
			status = EXCLUDED.status,
			handle_a = EXCLUDED.handle_a, handle_b = EXCLUDED.handle_b,
			person_a_name = EXCLUDED.person_a_name, person_b_name = EXCLUDED.person_b_name,
			bio_a = EXCLUDED.bio_a, bio_b = EXCLUDED.bio_b,
			profile_pic_a = EXCLUDED.profile_pic_a, profile_pic_b = EXCLUDED.profile_pic_b,
			market_city = EXCLUDED.market_city, market_region = EXCLUDED.market_region,
			market_source = EXCLUDED.market_source, source_handle = EXCLUDED.source_handle,
			source_class = EXCLUDED.source_class,
			discovery_caption = EXCLUDED.discovery_caption,
			discovery_image_url = EXCLUDED.discovery_image_url,
			discovery_post_url = EXCLUDED.discovery_post_url,
			evidence_json = EXCLUDED.evidence_json,
			research_notes = EXCLUDED.research_notes,
			research_steps_json = EXCLUDED.research_steps_json,
			address_line1 = EXCLUDED.address_line1, address_line2 = EXCLUDED.address_line2,
			address_city = EXCLUDED.address_city, address_region = EXCLUDED.address_region,
			address_postal = EXCLUDED.address_postal, address_country = EXCLUDED.address_country,
			address_confidence = EXCLUDED.address_confidence, address_source = EXCLUDED.address_source,
			address_candidates_json = EXCLUDED.address_candidates_json,
			headline = EXCLUDED.headline, body_message = EXCLUDED.body_message,
			internal_note = EXCLUDED.internal_note, postcard_html = EXCLUDED.postcard_html,
			mail_payload_json = EXCLUDED.mail_payload_json,
			verified_by = EXCLUDED.verified_by, verified_at = EXCLUDED.verified_at,
			mailed_at = EXCLUDED.mailed_at,
			priority_score = EXCLUDED.priority_score,
			follow_up_at = EXCLUDED.follow_up_at, follow_up_template = EXCLUDED.follow_up_template,
			follow_up_sent_at = EXCLUDED.follow_up_sent_at, follow_up_count = EXCLUDED.follow_up_count,
			updated_at = EXCLUDED.updated_at
	`,
		k.ID, k.CoupleID, k.Status, nullIfEmpty(k.HandleA), nullIfEmpty(k.HandleB),
		nullIfEmpty(k.PersonAName), nullIfEmpty(k.PersonBName),
		nullIfEmpty(k.BioA), nullIfEmpty(k.BioB), nullIfEmpty(k.ProfilePicA), nullIfEmpty(k.ProfilePicB),
		nullIfEmpty(k.MarketCity), nullIfEmpty(k.MarketRegion), nullIfEmpty(k.MarketSource),
		nullIfEmpty(k.SourceHandle), nullIfEmpty(k.SourceClass),
		nullIfEmpty(k.DiscoveryCaption), nullIfEmpty(k.DiscoveryImageURL), nullIfEmpty(k.DiscoveryPostURL),
		string(ev), nullIfEmpty(k.ResearchNotes), string(steps),
		nullIfEmpty(k.AddressLine1), nullIfEmpty(k.AddressLine2),
		nullIfEmpty(k.AddressCity), nullIfEmpty(k.AddressRegion), nullIfEmpty(k.AddressPostal),
		k.AddressCountry, k.AddressConfidence, nullIfEmpty(k.AddressSource), string(cands),
		nullIfEmpty(k.Headline), nullIfEmpty(k.BodyMessage), nullIfEmpty(k.InternalNote),
		nullIfEmpty(k.PostcardHTML), string(mail),
		nullIfEmpty(k.VerifiedBy), k.VerifiedAt, k.MailedAt,
		k.PriorityScore, k.FollowUpAt, nullIfEmpty(k.FollowUpTemplate), k.FollowUpSentAt, k.FollowUpCount,
		k.CreatedAt, k.UpdatedAt,
	)
	if err != nil {
		return k, err
	}
	// Name columns from 0010 — best-effort (ignore if migration lag in tests).
	_, _ = s.DB.Exec(`
		UPDATE congratulate_kits SET
		  first_name_a = $2, last_name_a = $3, first_name_b = $4, last_name_b = $5,
		  name_source_a = $6, name_source_b = $7, updated_at = now()
		WHERE id = $1`,
		k.ID,
		nullIfEmpty(k.FirstNameA), nullIfEmpty(k.LastNameA),
		nullIfEmpty(k.FirstNameB), nullIfEmpty(k.LastNameB),
		nullIfEmpty(k.NameSourceA), nullIfEmpty(k.NameSourceB),
	)
	return s.GetCongratulateKit(k.ID)
}

func (s *Store) GetCongratulateKit(id string) (CongratulateKit, error) {
	return s.scanKit(s.DB.QueryRow(kitSelect+` WHERE id = $1`, id))
}

func (s *Store) GetLatestKitForCouple(coupleID string) (CongratulateKit, error) {
	return s.scanKit(s.DB.QueryRow(kitSelect+` WHERE couple_id = $1 ORDER BY created_at DESC LIMIT 1`, coupleID))
}

func (s *Store) ListCongratulateKits(status string, limit int) ([]CongratulateKit, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	q := kitSelect
	args := []any{}
	if status != "" {
		q += ` WHERE status = $1 ORDER BY updated_at DESC LIMIT $2`
		args = append(args, status, limit)
	} else {
		q += ` ORDER BY updated_at DESC LIMIT $1`
		args = append(args, limit)
	}
	rows, err := s.DB.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []CongratulateKit
	for rows.Next() {
		k, err := s.scanKitRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, k)
	}
	return out, rows.Err()
}

const kitSelect = `SELECT id, couple_id, status,
	COALESCE(handle_a,''), COALESCE(handle_b,''), COALESCE(person_a_name,''), COALESCE(person_b_name,''),
	COALESCE(first_name_a,''), COALESCE(last_name_a,''), COALESCE(first_name_b,''), COALESCE(last_name_b,''),
	COALESCE(name_source_a,''), COALESCE(name_source_b,''),
	COALESCE(bio_a,''), COALESCE(bio_b,''), COALESCE(profile_pic_a,''), COALESCE(profile_pic_b,''),
	COALESCE(market_city,''), COALESCE(market_region,''), COALESCE(market_source,''),
	COALESCE(source_handle,''), COALESCE(source_class,''),
	COALESCE(discovery_caption,''), COALESCE(discovery_image_url,''), COALESCE(discovery_post_url,''),
	COALESCE(evidence_json,'[]'), COALESCE(research_notes,''), COALESCE(research_steps_json,'[]'),
	COALESCE(address_line1,''), COALESCE(address_line2,''), COALESCE(address_city,''),
	COALESCE(address_region,''), COALESCE(address_postal,''), COALESCE(address_country,'US'),
	address_confidence, COALESCE(address_source,''), COALESCE(address_candidates_json,'[]'),
	COALESCE(headline,''), COALESCE(body_message,''), COALESCE(internal_note,''),
	COALESCE(postcard_html,''), COALESCE(mail_payload_json,''),
	COALESCE(verified_by,''), verified_at, mailed_at,
	priority_score, follow_up_at, COALESCE(follow_up_template,''), follow_up_sent_at, follow_up_count,
	created_at, updated_at
	FROM congratulate_kits`

func (s *Store) scanKit(row *sql.Row) (CongratulateKit, error) {
	var k CongratulateKit
	var evidence, steps, cands, mail string
	var verifiedAt, mailedAt, followUpAt, followUpSentAt sql.NullTime
	err := row.Scan(
		&k.ID, &k.CoupleID, &k.Status,
		&k.HandleA, &k.HandleB, &k.PersonAName, &k.PersonBName,
		&k.FirstNameA, &k.LastNameA, &k.FirstNameB, &k.LastNameB,
		&k.NameSourceA, &k.NameSourceB,
		&k.BioA, &k.BioB, &k.ProfilePicA, &k.ProfilePicB,
		&k.MarketCity, &k.MarketRegion, &k.MarketSource,
		&k.SourceHandle, &k.SourceClass,
		&k.DiscoveryCaption, &k.DiscoveryImageURL, &k.DiscoveryPostURL,
		&evidence, &k.ResearchNotes, &steps,
		&k.AddressLine1, &k.AddressLine2, &k.AddressCity, &k.AddressRegion, &k.AddressPostal, &k.AddressCountry,
		&k.AddressConfidence, &k.AddressSource, &cands,
		&k.Headline, &k.BodyMessage, &k.InternalNote,
		&k.PostcardHTML, &mail,
		&k.VerifiedBy, &verifiedAt, &mailedAt,
		&k.PriorityScore, &followUpAt, &k.FollowUpTemplate, &followUpSentAt, &k.FollowUpCount,
		&k.CreatedAt, &k.UpdatedAt,
	)
	if err != nil {
		return k, err
	}
	_ = json.Unmarshal([]byte(evidence), &k.Evidence)
	_ = json.Unmarshal([]byte(steps), &k.ResearchSteps)
	_ = json.Unmarshal([]byte(cands), &k.AddressCandidates)
	if mail != "" && mail != "null" {
		_ = json.Unmarshal([]byte(mail), &k.MailPayload)
	}
	if verifiedAt.Valid {
		t := verifiedAt.Time
		k.VerifiedAt = &t
	}
	if mailedAt.Valid {
		t := mailedAt.Time
		k.MailedAt = &t
	}
	if followUpAt.Valid {
		t := followUpAt.Time
		k.FollowUpAt = &t
	}
	if followUpSentAt.Valid {
		t := followUpSentAt.Time
		k.FollowUpSentAt = &t
	}
	return k, nil
}

func (s *Store) scanKitRow(rows *sql.Rows) (CongratulateKit, error) {
	var k CongratulateKit
	var evidence, steps, cands, mail string
	var verifiedAt, mailedAt, followUpAt, followUpSentAt sql.NullTime
	err := rows.Scan(
		&k.ID, &k.CoupleID, &k.Status,
		&k.HandleA, &k.HandleB, &k.PersonAName, &k.PersonBName,
		&k.FirstNameA, &k.LastNameA, &k.FirstNameB, &k.LastNameB,
		&k.NameSourceA, &k.NameSourceB,
		&k.BioA, &k.BioB, &k.ProfilePicA, &k.ProfilePicB,
		&k.MarketCity, &k.MarketRegion, &k.MarketSource,
		&k.SourceHandle, &k.SourceClass,
		&k.DiscoveryCaption, &k.DiscoveryImageURL, &k.DiscoveryPostURL,
		&evidence, &k.ResearchNotes, &steps,
		&k.AddressLine1, &k.AddressLine2, &k.AddressCity, &k.AddressRegion, &k.AddressPostal, &k.AddressCountry,
		&k.AddressConfidence, &k.AddressSource, &cands,
		&k.Headline, &k.BodyMessage, &k.InternalNote,
		&k.PostcardHTML, &mail,
		&k.VerifiedBy, &verifiedAt, &mailedAt,
		&k.PriorityScore, &followUpAt, &k.FollowUpTemplate, &followUpSentAt, &k.FollowUpCount,
		&k.CreatedAt, &k.UpdatedAt,
	)
	if err != nil {
		return k, err
	}
	_ = json.Unmarshal([]byte(evidence), &k.Evidence)
	_ = json.Unmarshal([]byte(steps), &k.ResearchSteps)
	_ = json.Unmarshal([]byte(cands), &k.AddressCandidates)
	if mail != "" && mail != "null" {
		_ = json.Unmarshal([]byte(mail), &k.MailPayload)
	}
	if verifiedAt.Valid {
		t := verifiedAt.Time
		k.VerifiedAt = &t
	}
	if mailedAt.Valid {
		t := mailedAt.Time
		k.MailedAt = &t
	}
	if followUpAt.Valid {
		t := followUpAt.Time
		k.FollowUpAt = &t
	}
	if followUpSentAt.Valid {
		t := followUpSentAt.Time
		k.FollowUpSentAt = &t
	}
	return k, nil
}

// FindDiscoveryPost looks for a stored post that tags both couple handles (or either).
func (s *Store) FindDiscoveryPost(handleA, handleB string) (caption, imageURL, postURL, location, sourceHandle string, ok bool) {
	rows, err := s.DB.Query(`
		SELECT monitor, raw_payload FROM social_observations
		WHERE observation_type IN ('post','vendor_post')
		  AND (
		    raw_payload ILIKE '%' || $1 || '%'
		    OR raw_payload ILIKE '%' || $2 || '%'
		  )
		ORDER BY observed_at DESC LIMIT 40`, handleA, handleB)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var mon, raw string
		if err := rows.Scan(&mon, &raw); err != nil {
			continue
		}
		var payload map[string]any
		if json.Unmarshal([]byte(raw), &payload) != nil {
			continue
		}
		tags := stringSlice(payload["tags"])
		mentions := stringSlice(payload["provider_mentions"])
		if len(mentions) == 0 {
			mentions = stringSlice(payload["mentions"])
		}
		all := append(append([]string{}, tags...), mentions...)
		ha, hb := lower(handleA), lower(handleB)
		hasA, hasB := false, false
		for _, t := range all {
			t = lower(t)
			if t == ha {
				hasA = true
			}
			if t == hb {
				hasB = true
			}
		}
		// Prefer posts that tag both; accept either on later pass
		if !(hasA && hasB) && !(hasA || hasB) {
			continue
		}
		caption, _ = payload["caption"].(string)
		imageURL, _ = payload["image_url"].(string)
		if imageURL == "" {
			imageURL, _ = payload["display_url"].(string)
		}
		postURL, _ = payload["url"].(string)
		location, _ = payload["location"].(string)
		sourceHandle, _ = payload["handle"].(string)
		if sourceHandle == "" && stringsHasPrefix(mon, "vendor:") {
			sourceHandle = mon[len("vendor:"):]
		}
		if hasA && hasB {
			ok = true
			return
		}
		// keep first single-tag hit as fallback
		if !ok {
			ok = true
			// continue searching for both-tag hit
			if hasA && hasB {
				return
			}
		}
	}
	return
}

// FindDiscoveryPostLocation returns the Instagram venue tag (location) from the
// discovery post for a couple. Used to feed post location into detective queries.
func (s *Store) FindDiscoveryPostLocation(handleA, handleB string) (location string, ok bool) {
	row := s.DB.QueryRow(`
		SELECT raw_payload FROM social_observations
		WHERE observation_type IN ('post','vendor_post')
		  AND (
		    raw_payload ILIKE '%' || $1 || '%'
		    OR raw_payload ILIKE '%' || $2 || '%'
		  )
		ORDER BY observed_at DESC LIMIT 1`, handleA, handleB)
	var raw string
	if err := row.Scan(&raw); err != nil {
		return "", false
	}
	var payload map[string]any
	if json.Unmarshal([]byte(raw), &payload) != nil {
		return "", false
	}
	loc, _ := payload["location"].(string)
	return loc, loc != ""
}

func lower(s string) string {
	b := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		b[i] = c
	}
	// also strip @
	out := string(b)
	if len(out) > 0 && out[0] == '@' {
		return out[1:]
	}
	return out
}

func stringsHasPrefix(s, p string) bool {
	return len(s) >= len(p) && s[:len(p)] == p
}
