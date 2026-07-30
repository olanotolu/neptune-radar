package store

import (
	"database/sql"
	"time"

	"neptune-social-radar/backend/internal/ontology"
)

// --- watched_sources: the curated vendor/account list -------------------------

const watchedSourceSelect = `SELECT id, handle, source_class, active, state, city,
	follower_count, following_count, post_count, full_name, profile_pic_url, verified, profile_checked_at,
	last_scanned_at, last_scan_couples, last_scan_actions,
	created_at FROM watched_sources`

func scanWatchedSource(row interface{ Scan(...any) error }) (ontology.WatchedSource, error) {
	var w ontology.WatchedSource
	var state, city, fullName, picURL sql.NullString
	var followers, following, posts sql.NullInt64
	var verified sql.NullBool
	var checkedAt, scannedAt sql.NullTime
	var scanCouples, scanActions sql.NullInt64
	err := row.Scan(&w.ID, &w.Handle, &w.SourceClass, &w.Active, &state, &city,
		&followers, &following, &posts, &fullName, &picURL, &verified, &checkedAt,
		&scannedAt, &scanCouples, &scanActions,
		&w.CreatedAt)
	if err != nil {
		return w, err
	}
	w.State, w.City = state.String, city.String
	w.FullName, w.ProfilePicURL = fullName.String, picURL.String
	if followers.Valid {
		n := int(followers.Int64)
		w.FollowerCount = &n
	}
	if following.Valid {
		n := int(following.Int64)
		w.FollowingCount = &n
	}
	if posts.Valid {
		n := int(posts.Int64)
		w.PostCount = &n
	}
	if verified.Valid {
		v := verified.Bool
		w.Verified = &v
	}
	if checkedAt.Valid {
		t := checkedAt.Time
		w.ProfileCheckedAt = &t
	}
	if scannedAt.Valid {
		t := scannedAt.Time
		w.LastScannedAt = &t
	}
	if scanCouples.Valid {
		n := int(scanCouples.Int64)
		w.LastScanCouples = &n
	}
	if scanActions.Valid {
		n := int(scanActions.Int64)
		w.LastScanActions = &n
	}
	return w, nil
}

// TouchSourceScan records that an agent scan finished for this vendor.
func (s *Store) TouchSourceScan(handle string, couples, actions int) error {
	_, err := s.DB.Exec(
		`UPDATE watched_sources SET last_scanned_at = now(), last_scan_couples = $2, last_scan_actions = $3 WHERE handle = $1`,
		handle, couples, actions,
	)
	return err
}

// UpdateWatchedSourceProfileStats persists the real result of an Apify
// profile fetch — only ever called after a successful check, so these
// columns are either real numbers or NULL, never a placeholder.
func (s *Store) UpdateWatchedSourceProfileStats(handle string, followers, following, posts int, fullName, profilePicURL string, verified bool) error {
	_, err := s.DB.Exec(
		`UPDATE watched_sources SET follower_count = $2, following_count = $3, post_count = $4,
		   full_name = $5, profile_pic_url = $6, verified = $7, profile_checked_at = now()
		 WHERE handle = $1`,
		handle, followers, following, posts, fullName, profilePicURL, verified,
	)
	return err
}

// UpdateWatchedSourceProfileAndGeo saves profile stats AND city/state so the
// map and prospect scoring can use the vendor's market.
func (s *Store) UpdateWatchedSourceProfileAndGeo(handle string, followers, following, posts int, fullName, profilePicURL string, verified bool, city, state string) error {
	_, err := s.DB.Exec(
		`UPDATE watched_sources SET
		   follower_count = $2, following_count = $3, post_count = $4,
		   full_name = $5, profile_pic_url = $6, verified = $7, profile_checked_at = now(),
		   city = CASE WHEN $8 <> '' THEN $8 ELSE city END,
		   state = CASE WHEN $9 <> '' THEN $9 ELSE state END
		 WHERE handle = $1`,
		handle, followers, following, posts, fullName, profilePicURL, verified, city, state,
	)
	return err
}

// SetWatchedSourceLocation manually or programmatically sets geo without a full profile refresh.
func (s *Store) SetWatchedSourceLocation(handle, city, state string) error {
	_, err := s.DB.Exec(
		`UPDATE watched_sources SET
		   city = CASE WHEN $2 <> '' THEN $2 ELSE city END,
		   state = CASE WHEN $3 <> '' THEN $3 ELSE state END
		 WHERE handle = $1`,
		handle, city, state,
	)
	return err
}

func (s *Store) AddWatchedSource(handle, sourceClass string) (ontology.WatchedSource, error) {
	return s.AddWatchedSourceWithGeo(handle, sourceClass, "", "")
}

// AddWatchedSourceWithGeo registers a vendor and optional market location.
func (s *Store) AddWatchedSourceWithGeo(handle, sourceClass, state, city string) (ontology.WatchedSource, error) {
	w := ontology.WatchedSource{ID: NewID("src"), Handle: handle, SourceClass: sourceClass, Active: true, State: state, City: city}
	_, err := s.DB.Exec(
		`INSERT INTO watched_sources (id, handle, source_class, active, state, city) VALUES ($1, $2, $3, TRUE, NULLIF($4,''), NULLIF($5,''))
		 ON CONFLICT (handle) DO UPDATE SET
		   source_class = EXCLUDED.source_class,
		   active = TRUE,
		   state = COALESCE(NULLIF(EXCLUDED.state,''), watched_sources.state),
		   city = COALESCE(NULLIF(EXCLUDED.city,''), watched_sources.city)`,
		w.ID, w.Handle, w.SourceClass, state, city,
	)
	if err != nil {
		return w, err
	}
	return s.GetWatchedSource(handle)
}

// UpsertWatchedSourceGeo registers (or updates) a watched source with a
// state/city tag, so the map's Instagram layer can filter to it — the row
// is the same one internal/ingest's worker polls, not a parallel record.
func (s *Store) UpsertWatchedSourceGeo(handle, sourceClass, state, city string) (ontology.WatchedSource, error) {
	_, err := s.DB.Exec(
		`INSERT INTO watched_sources (id, handle, source_class, active, state, city) VALUES ($1, $2, $3, TRUE, $4, $5)
		 ON CONFLICT (handle) DO UPDATE SET source_class = EXCLUDED.source_class, active = TRUE, state = EXCLUDED.state, city = EXCLUDED.city`,
		NewID("src"), handle, sourceClass, state, city,
	)
	if err != nil {
		return ontology.WatchedSource{}, err
	}
	return s.GetWatchedSource(handle)
}

func (s *Store) GetWatchedSourceByID(id string) (ontology.WatchedSource, error) {
	return scanWatchedSource(s.DB.QueryRow(watchedSourceSelect+` WHERE id = $1`, id))
}

func (s *Store) GetWatchedSource(handle string) (ontology.WatchedSource, error) {
	return scanWatchedSource(s.DB.QueryRow(watchedSourceSelect+` WHERE handle = $1`, handle))
}

func (s *Store) DeactivateWatchedSource(handle string) error {
	_, err := s.DB.Exec(`UPDATE watched_sources SET active = FALSE WHERE handle = $1`, handle)
	return err
}

func (s *Store) ListWatchedSources(activeOnly bool) ([]ontology.WatchedSource, error) {
	q := watchedSourceSelect
	if activeOnly {
		q += ` WHERE active = TRUE`
	}
	q += ` ORDER BY created_at ASC`
	rows, err := s.DB.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ontology.WatchedSource
	for rows.Next() {
		w, err := scanWatchedSource(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, w)
	}
	return out, rows.Err()
}

// ListWatchedSourcesByState returns watched sources geo-tagged to a state —
// the real, already-polled accounts the map's Instagram layer displays.
func (s *Store) ListWatchedSourcesByState(state string, activeOnly bool) ([]ontology.WatchedSource, error) {
	q := watchedSourceSelect + ` WHERE state = $1`
	if activeOnly {
		q += ` AND active = TRUE`
	}
	q += ` ORDER BY city ASC, handle ASC`
	rows, err := s.DB.Query(q, state)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ontology.WatchedSource
	for rows.Next() {
		w, err := scanWatchedSource(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, w)
	}
	return out, rows.Err()
}

// SourceClassForHandle returns the curated source class for an account that
// produced a post, if it is on the watched list — this is how a vendor's post
// gets its source_account_type. Unknown accounts return "".
//
// When the per-tick cache is populated (worker path), this is a map lookup —
// no DB round-trip. Without the cache (tests, API paths that don't run the
// worker), it falls back to a direct query.
func (s *Store) SourceClassForHandle(handle string) string {
	s.mu.RLock()
	if s.sourceClassCache != nil {
		_, ok := s.sourceClassCache[handle]
		s.mu.RUnlock()
		if ok {
			return s.sourceClassCache[handle]
		}
		return ""
	}
	s.mu.RUnlock()
	var class string
	err := s.DB.QueryRow(`SELECT source_class FROM watched_sources WHERE handle = $1 AND active = TRUE`, handle).Scan(&class)
	if err != nil {
		return ""
	}
	return class
}

// RefreshSourceClassCache loads handle→class for all active watched_sources in
// one query, replacing the per-post SELECT that SourceClassForHandle used to
// make 6× per post (worker stamp + roles + scorer). Called once per worker
// tick; callers that don't invoke this get the DB fallback.
func (s *Store) RefreshSourceClassCache() error {
	rows, err := s.DB.Query(`SELECT handle, source_class FROM watched_sources WHERE active = TRUE`)
	if err != nil {
		return err
	}
	defer rows.Close()
	m := make(map[string]string, 64)
	for rows.Next() {
		var handle, class string
		if err := rows.Scan(&handle, &class); err != nil {
			return err
		}
		m[handle] = class
	}
	if err := rows.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	s.sourceClassCache = m
	s.mu.Unlock()
	return nil
}

// --- ingest_cursors: per-monitor read progress --------------------------------

func (s *Store) GetCursor(monitor string) (ontology.IngestCursor, error) {
	var c ontology.IngestCursor
	var lastSeen, lastRun sql.NullTime
	err := s.DB.QueryRow(
		`SELECT monitor, last_seen_at, last_run_at, updated_at FROM ingest_cursors WHERE monitor = $1`, monitor,
	).Scan(&c.Monitor, &lastSeen, &lastRun, &c.UpdatedAt)
	if err != nil {
		return c, err
	}
	if lastSeen.Valid {
		t := lastSeen.Time
		c.LastSeenAt = &t
	}
	if lastRun.Valid {
		t := lastRun.Time
		c.LastRunAt = &t
	}
	return c, nil
}

// AdvanceCursor moves a monitor's high-water mark forward (never backward)
// and records the run time.
func (s *Store) AdvanceCursor(monitor string, lastSeen time.Time) error {
	_, err := s.DB.Exec(
		`INSERT INTO ingest_cursors (monitor, last_seen_at, last_run_at, updated_at)
		 VALUES ($1, $2, now(), now())
		 ON CONFLICT (monitor) DO UPDATE SET
		   last_seen_at = GREATEST(ingest_cursors.last_seen_at, EXCLUDED.last_seen_at),
		   last_run_at = now(), updated_at = now()`,
		monitor, lastSeen.UTC(),
	)
	return err
}

// TouchCursor records a run that found nothing new.
func (s *Store) TouchCursor(monitor string) error {
	_, err := s.DB.Exec(
		`INSERT INTO ingest_cursors (monitor, last_run_at, updated_at) VALUES ($1, now(), now())
		 ON CONFLICT (monitor) DO UPDATE SET last_run_at = now(), updated_at = now()`,
		monitor,
	)
	return err
}

func (s *Store) ListCursors() ([]ontology.IngestCursor, error) {
	rows, err := s.DB.Query(`SELECT monitor, last_seen_at, last_run_at, updated_at FROM ingest_cursors ORDER BY monitor ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ontology.IngestCursor
	for rows.Next() {
		var c ontology.IngestCursor
		var lastSeen, lastRun sql.NullTime
		if err := rows.Scan(&c.Monitor, &lastSeen, &lastRun, &c.UpdatedAt); err != nil {
			return nil, err
		}
		if lastSeen.Valid {
			t := lastSeen.Time
			c.LastSeenAt = &t
		}
		if lastRun.Valid {
			t := lastRun.Time
			c.LastRunAt = &t
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// --- api_usage: provider spend accounting -------------------------------------

func (s *Store) RecordUsage(provider, monitor string, results int) error {
	_, err := s.DB.Exec(
		`INSERT INTO api_usage (id, provider, monitor, results_count) VALUES ($1, $2, $3, $4)`,
		NewID("usage"), provider, monitor, results,
	)
	return err
}

// UsageToday returns how many provider results have been consumed since
// midnight UTC — the number the daily budget cap is enforced against.
func (s *Store) UsageToday(provider string) (int, error) {
	var n int
	err := s.DB.QueryRow(
		`SELECT COALESCE(SUM(results_count),0) FROM api_usage WHERE provider = $1 AND created_at >= date_trunc('day', now())`,
		provider,
	).Scan(&n)
	return n, err
}

// ListProfileWatchHandles returns handles of accounts linked to a person in a
// known couple — the set whose bios the profile monitor polls for changes.
func (s *Store) ListProfileWatchHandles() ([]string, error) {
	rows, err := s.DB.Query(
		`SELECT DISTINCT handle FROM social_accounts
		 WHERE person_id IN (SELECT person_a_id FROM couples UNION SELECT person_b_id FROM couples)
		   AND is_disabled = FALSE AND is_private = FALSE`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var h string
		if err := rows.Scan(&h); err != nil {
			return nil, err
		}
		out = append(out, h)
	}
	return out, rows.Err()
}

// FollowCheckTarget is one couple whose mutual-follow state should be
// re-verified: they have an open hypothesis, or a live engaged/married/
// status_uncertain relationship (the stages where an unfollow matters).
type FollowCheckTarget struct {
	CoupleID string
	HandleA  string
	HandleB  string
}

func (s *Store) ListCouplesForFollowCheck(limit int) ([]FollowCheckTarget, error) {
	rows, err := s.DB.Query(
		`SELECT c.id, a.handle, b.handle FROM couples c
		 JOIN social_accounts a ON a.person_id = c.person_a_id
		 JOIN social_accounts b ON b.person_id = c.person_b_id
		 WHERE (
		   EXISTS (SELECT 1 FROM life_event_hypotheses h WHERE h.couple_id = c.id AND h.status IN ('unconfirmed','corroborating'))
		   OR EXISTS (SELECT 1 FROM relationships r WHERE r.couple_id = c.id AND r.effective_to IS NULL AND r.stage IN ('engaged','married','status_uncertain'))
		 )
		 AND a.is_disabled = FALSE AND b.is_disabled = FALSE
		 LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []FollowCheckTarget
	for rows.Next() {
		var t FollowCheckTarget
		if err := rows.Scan(&t.CoupleID, &t.HandleA, &t.HandleB); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}
