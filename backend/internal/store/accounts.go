package store

import (
	"database/sql"
	"errors"
	"time"

	"neptune-social-radar/backend/internal/ontology"
)

func (s *Store) EnsureAccount(a ontology.SocialAccount) (ontology.SocialAccount, error) {
	existing, err := s.GetAccountByHandle(a.Platform, a.Handle)
	if err == nil {
		return existing, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return ontology.SocialAccount{}, err
	}
	if a.ID == "" {
		a.ID = NewID("acct")
	}
	if a.Platform == "" {
		a.Platform = "instagram"
	}
	var personID any
	if a.PersonID != "" {
		personID = a.PersonID
	}
	// ON CONFLICT: a concurrent worker can insert the same handle between our
	// existence check and our INSERT. Losing that race must return the
	// winner's row, not drop the event with a unique-violation error.
	err = s.DB.QueryRow(
		`INSERT INTO social_accounts (id, person_id, platform, handle, display_name, bio_text, is_private, is_disabled,
		  profile_pic_url, follower_count, following_count, inferred_city, inferred_region, location_source)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		 ON CONFLICT (platform, handle) DO NOTHING
		 RETURNING id`,
		a.ID, personID, a.Platform, a.Handle, a.DisplayName, a.BioText, a.IsPrivate, a.IsDisabled,
		nullIfEmpty(a.ProfilePicURL), a.FollowerCount, a.FollowingCount,
		nullIfEmpty(a.InferredCity), nullIfEmpty(a.InferredRegion), nullIfEmpty(a.LocationSource),
	).Scan(&a.ID)
	if err == sql.ErrNoRows {
		return s.GetAccountByHandle(a.Platform, a.Handle)
	}
	if err != nil {
		return ontology.SocialAccount{}, err
	}
	return s.GetAccount(a.ID)
}

func (s *Store) GetAccount(id string) (ontology.SocialAccount, error) {
	return s.scanAccount(s.DB.QueryRow(accountSelect + ` WHERE id = $1`, id))
}

func (s *Store) GetAccountByPersonID(personID string) (ontology.SocialAccount, error) {
	return s.scanAccount(s.DB.QueryRow(accountSelect+` WHERE person_id = $1 LIMIT 1`, personID))
}

func (s *Store) GetAccountByHandle(platform, handle string) (ontology.SocialAccount, error) {
	if platform == "" {
		platform = "instagram"
	}
	return s.scanAccount(s.DB.QueryRow(accountSelect+` WHERE platform = $1 AND handle = $2`, platform, handle))
}

const accountSelect = `SELECT id, COALESCE(person_id,''), platform, handle, COALESCE(display_name,''), COALESCE(bio_text,''),
	is_private, is_disabled, last_seen_at, COALESCE(profile_pic_url,''), follower_count, following_count,
	profile_checked_at, COALESCE(inferred_city,''), COALESCE(inferred_region,''), COALESCE(location_source,'')
	FROM social_accounts`

func (s *Store) scanAccount(row *sql.Row) (ontology.SocialAccount, error) {
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

func (s *Store) UpdateAccountBio(accountID, bio string, seenAt time.Time) error {
	_, err := s.DB.Exec(`UPDATE social_accounts SET bio_text = $1, last_seen_at = $2 WHERE id = $3`,
		bio, seenAt.UTC(), accountID)
	return err
}

// UpdateAccountProfile stores Instagram avatar + bio + counts for prospect cards.
func (s *Store) UpdateAccountProfile(accountID string, displayName, bio, profilePicURL string, followers, following *int, private bool, seenAt time.Time) error {
	_, err := s.DB.Exec(
		`UPDATE social_accounts SET
		   display_name = CASE WHEN $1 <> '' THEN $1 ELSE display_name END,
		   bio_text = $2,
		   profile_pic_url = CASE WHEN $3 <> '' THEN $3 ELSE profile_pic_url END,
		   follower_count = COALESCE($4, follower_count),
		   following_count = COALESCE($5, following_count),
		   is_private = $6,
		   last_seen_at = $7,
		   profile_checked_at = $7
		 WHERE id = $8`,
		displayName, bio, profilePicURL, followers, following, private, seenAt.UTC(), accountID,
	)
	return err
}

// UpdateAccountLocation sets bio-inferred city/region on a social account.
func (s *Store) UpdateAccountLocation(accountID, city, region, source string) error {
	_, err := s.DB.Exec(
		`UPDATE social_accounts SET inferred_city = $1, inferred_region = $2, location_source = $3 WHERE id = $4`,
		nullIfEmpty(city), nullIfEmpty(region), nullIfEmpty(source), accountID,
	)
	return err
}

func (s *Store) SetAccountPersonID(accountID, personID string) error {
	_, err := s.DB.Exec(`UPDATE social_accounts SET person_id = $1 WHERE id = $2`, personID, accountID)
	return err
}

func (s *Store) SetAccountDisabled(accountID string, disabled bool) error {
	_, err := s.DB.Exec(`UPDATE social_accounts SET is_disabled = $1 WHERE id = $2`, disabled, accountID)
	return err
}

func (s *Store) SetAccountPrivate(accountID string, private bool) error {
	_, err := s.DB.Exec(`UPDATE social_accounts SET is_private = $1 WHERE id = $2`, private, accountID)
	return err
}

// NeedsProfileRefresh is true when we've never checked the profile, or the
// last check is older than maxAge.
func (s *Store) NeedsProfileRefresh(accountID string, maxAge time.Duration) (bool, error) {
	var checked sql.NullTime
	err := s.DB.QueryRow(`SELECT profile_checked_at FROM social_accounts WHERE id = $1`, accountID).Scan(&checked)
	if err != nil {
		return false, err
	}
	if !checked.Valid {
		return true, nil
	}
	return time.Since(checked.Time) > maxAge, nil
}

