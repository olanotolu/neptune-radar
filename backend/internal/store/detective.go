package store

import (
	"encoding/json"
	"time"
)

// AddressLookup is one people-search / records call.
type AddressLookup struct {
	ID           string    `json:"id"`
	KitID        string    `json:"kit_id,omitempty"`
	CoupleID     string    `json:"couple_id,omitempty"`
	Provider     string    `json:"provider"`
	QueryJSON    string    `json:"query_json"`
	ResponseJSON string    `json:"response_json,omitempty"`
	CandidatesJSON string  `json:"candidates_json"`
	Status       string    `json:"status"`
	ErrorMessage string    `json:"error_message,omitempty"`
	CostCents    int       `json:"cost_cents"`
	CreatedAt    time.Time `json:"created_at"`
}

// MailSend is one Lob/PostGrid postcard send.
type MailSend struct {
	ID                   string    `json:"id"`
	KitID                string    `json:"kit_id"`
	CoupleID             string    `json:"couple_id,omitempty"`
	Provider             string    `json:"provider"`
	ExternalID           string    `json:"external_id,omitempty"`
	Status               string    `json:"status"`
	ToAddressJSON        string    `json:"to_address_json,omitempty"`
	FromAddressJSON      string    `json:"from_address_json,omitempty"`
	RawResponse          string    `json:"raw_response,omitempty"`
	ErrorMessage         string    `json:"error_message,omitempty"`
	CostCents            int       `json:"cost_cents"`
	ExpectedDeliveryDate string    `json:"expected_delivery_date,omitempty"`
	CreatedAt            time.Time `json:"created_at"`
}

func (s *Store) InsertAddressLookup(l AddressLookup) (AddressLookup, error) {
	if l.ID == "" {
		l.ID = NewID("alook")
	}
	if l.Status == "" {
		l.Status = "ok"
	}
	if l.CandidatesJSON == "" {
		l.CandidatesJSON = "[]"
	}
	if l.QueryJSON == "" {
		l.QueryJSON = "{}"
	}
	l.CreatedAt = time.Now().UTC()
	_, err := s.DB.Exec(`
		INSERT INTO address_lookups (id, kit_id, couple_id, provider, query_json, response_json, candidates_json, status, error_message, cost_cents, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		l.ID, nullIfEmpty(l.KitID), nullIfEmpty(l.CoupleID), l.Provider,
		l.QueryJSON, nullIfEmpty(l.ResponseJSON), l.CandidatesJSON,
		l.Status, nullIfEmpty(l.ErrorMessage), l.CostCents, l.CreatedAt,
	)
	return l, err
}

func (s *Store) InsertMailSend(m MailSend) (MailSend, error) {
	if m.ID == "" {
		m.ID = NewID("mail")
	}
	if m.Status == "" {
		m.Status = "queued"
	}
	m.CreatedAt = time.Now().UTC()
	_, err := s.DB.Exec(`
		INSERT INTO mail_sends (id, kit_id, couple_id, provider, external_id, status, to_address_json, from_address_json, raw_response, error_message, cost_cents, expected_delivery_date, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		m.ID, m.KitID, nullIfEmpty(m.CoupleID), m.Provider, nullIfEmpty(m.ExternalID), m.Status,
		nullIfEmpty(m.ToAddressJSON), nullIfEmpty(m.FromAddressJSON), nullIfEmpty(m.RawResponse),
		nullIfEmpty(m.ErrorMessage), m.CostCents, nullIfEmpty(m.ExpectedDeliveryDate), m.CreatedAt,
	)
	return m, err
}

// ExtractObservationFacts backfills structured columns from raw_payload for recent rows.
func (s *Store) ExtractObservationFacts(limit int) (int, error) {
	if limit <= 0 || limit > 5000 {
		limit = 500
	}
	rows, err := s.DB.Query(`
		SELECT id, raw_payload FROM social_observations
		WHERE facts_extracted_at IS NULL
		ORDER BY observed_at DESC LIMIT $1`, limit)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	n := 0
	for rows.Next() {
		var id, raw string
		if err := rows.Scan(&id, &raw); err != nil {
			return n, err
		}
		var payload map[string]any
		if json.Unmarshal([]byte(raw), &payload) != nil {
			_, _ = s.DB.Exec(`UPDATE social_observations SET facts_extracted_at = now() WHERE id = $1`, id)
			continue
		}
		cap, _ := payload["caption"].(string)
		img, _ := payload["image_url"].(string)
		if img == "" {
			img, _ = payload["display_url"].(string)
		}
		url, _ := payload["url"].(string)
		loc, _ := payload["location"].(string)
		handle, _ := payload["handle"].(string)
		tags, _ := json.Marshal(stringSlice(payload["tags"]))
		mentions := stringSlice(payload["provider_mentions"])
		if len(mentions) == 0 {
			mentions = stringSlice(payload["mentions"])
		}
		mentJSON, _ := json.Marshal(mentions)
		_, err := s.DB.Exec(`
			UPDATE social_observations SET
			  caption = $2, image_url = $3, post_url = $4, location_name = $5,
			  source_handle = $6, tags_json = $7, mentions_json = $8, facts_extracted_at = now()
			WHERE id = $1`,
			id, nullIfEmpty(cap), nullIfEmpty(img), nullIfEmpty(url), nullIfEmpty(loc),
			nullIfEmpty(handle), string(tags), string(mentJSON),
		)
		if err != nil {
			return n, err
		}
		n++
	}
	return n, rows.Err()
}

// CoupleDossier is agent-ready evidence for one couple.
type CoupleDossier struct {
	CoupleID      string           `json:"couple_id"`
	HandleA       string           `json:"handle_a,omitempty"`
	HandleB       string           `json:"handle_b,omitempty"`
	PersonAName   string           `json:"person_a_name,omitempty"`
	PersonBName   string           `json:"person_b_name,omitempty"`
	BioA          string           `json:"bio_a,omitempty"`
	BioB          string           `json:"bio_b,omitempty"`
	ProfilePicA   string           `json:"profile_pic_a,omitempty"`
	ProfilePicB   string           `json:"profile_pic_b,omitempty"`
	City          string           `json:"city,omitempty"`
	Region        string           `json:"region,omitempty"`
	Observations  []DossierPost    `json:"observations"`
	LatestKitID   string           `json:"latest_kit_id,omitempty"`
}

type DossierPost struct {
	ID        string   `json:"id"`
	Caption   string   `json:"caption,omitempty"`
	ImageURL  string   `json:"image_url,omitempty"`
	PostURL   string   `json:"post_url,omitempty"`
	Location  string   `json:"location,omitempty"`
	Handle    string   `json:"source_handle,omitempty"`
	Tags      []string `json:"tags,omitempty"`
	ObservedAt string  `json:"observed_at,omitempty"`
}

// GetCoupleDossier loads structured evidence for detective agents.
func (s *Store) GetCoupleDossier(coupleID string) (CoupleDossier, error) {
	var d CoupleDossier
	d.CoupleID = coupleID
	// Prefer board projection fields
	board, err := s.ListProspectBoard(300)
	if err == nil {
		for _, c := range board {
			if c.CoupleID == coupleID {
				d.HandleA, d.HandleB = c.HandleA, c.HandleB
				d.PersonAName, d.PersonBName = c.PersonALabel, c.PersonBLabel
				d.BioA, d.BioB = c.BioA, c.BioB
				d.ProfilePicA, d.ProfilePicB = c.ProfilePicA, c.ProfilePicB
				d.City, d.Region = c.City, c.Region
				break
			}
		}
	}
	if kit, err := s.GetLatestKitForCouple(coupleID); err == nil {
		d.LatestKitID = kit.ID
		if kit.PersonAName != "" {
			d.PersonAName = kit.PersonAName
		}
		if kit.PersonBName != "" {
			d.PersonBName = kit.PersonBName
		}
		if kit.MarketCity != "" {
			d.City, d.Region = kit.MarketCity, kit.MarketRegion
		}
	}

	// Related observations by handle tags in payload / structured cols
	if d.HandleA != "" || d.HandleB != "" {
		ha, hb := d.HandleA, d.HandleB
		if ha == "" {
			ha = hb
		}
		if hb == "" {
			hb = ha
		}
		rows, err := s.DB.Query(`
			SELECT id, COALESCE(caption,''), COALESCE(image_url,''), COALESCE(post_url,''),
			  COALESCE(location_name,''), COALESCE(source_handle,''), COALESCE(tags_json,'[]'),
			  raw_payload, observed_at
			FROM social_observations
			WHERE raw_payload ILIKE '%' || $1 || '%'
			   OR raw_payload ILIKE '%' || $2 || '%'
			ORDER BY observed_at DESC LIMIT 30`, ha, hb)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var p DossierPost
				var tagsRaw, raw string
				var observed time.Time
				if err := rows.Scan(&p.ID, &p.Caption, &p.ImageURL, &p.PostURL, &p.Location, &p.Handle, &tagsRaw, &raw, &observed); err != nil {
					continue
				}
				// Fall back to raw_payload parse when structured cols empty
				if p.Caption == "" || p.ImageURL == "" {
					var payload map[string]any
					if json.Unmarshal([]byte(raw), &payload) == nil {
						if p.Caption == "" {
							p.Caption, _ = payload["caption"].(string)
						}
						if p.ImageURL == "" {
							p.ImageURL, _ = payload["image_url"].(string)
							if p.ImageURL == "" {
								p.ImageURL, _ = payload["display_url"].(string)
							}
						}
						if p.PostURL == "" {
							p.PostURL, _ = payload["url"].(string)
						}
						if p.Location == "" {
							p.Location, _ = payload["location"].(string)
						}
						if p.Handle == "" {
							p.Handle, _ = payload["handle"].(string)
						}
						if tagsRaw == "" || tagsRaw == "[]" {
							if t := stringSlice(payload["tags"]); len(t) > 0 {
								p.Tags = t
							}
						}
					}
				}
				if len(p.Tags) == 0 {
					_ = json.Unmarshal([]byte(tagsRaw), &p.Tags)
				}
				p.ObservedAt = observed.UTC().Format(time.RFC3339)
				d.Observations = append(d.Observations, p)
			}
		}
	}
	if d.Observations == nil {
		d.Observations = []DossierPost{}
	}
	return d, nil
}
