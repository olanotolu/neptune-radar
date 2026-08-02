package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"neptune-social-radar/backend/internal/signals"
	"neptune-social-radar/backend/internal/store"
)

// health is the one unauthenticated route — load balancer/uptime probe.
// It reports DB connectivity, provider availability, ingest loop state,
// pending action and DLQ counts, and the build version.
func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	dbOK := true
	if err := s.Store.DB.Ping(); err != nil {
		dbOK = false
	}

	providerStatus := "unknown"
	ingestRunning := false
	if s.Watch != nil {
		ingestRunning = !s.Watch.IsPaused()
		if s.Watch.ProviderAvailable() {
			providerStatus = "ok"
		} else {
			providerStatus = "unavailable"
		}
	}

	pendingActions := 0
	dlqPending := 0
	if dbOK {
		if n, err := s.Store.CountPendingActions(); err == nil {
			pendingActions = n
		}
		if n, err := s.Store.CountDLQPending(); err == nil {
			dlqPending = n
		}
	}

	dbCheck := "ok"
	if !dbOK {
		dbCheck = "fail"
	}

	status := "ok"
	httpCode := http.StatusOK
	if !dbOK {
		status = "down"
		httpCode = http.StatusServiceUnavailable
	} else if providerStatus == "unavailable" || dlqPending > 0 {
		status = "degraded"
	}

	writeJSON(w, httpCode, map[string]any{
		"status": status,
		"checks": map[string]any{
			"database":        dbCheck,
			"provider":        providerStatus,
			"ingest_running":  ingestRunning,
			"pending_actions": pendingActions,
			"dlq_pending":     dlqPending,
		},
		"version": BuildVersion,
	})
}

// BuildVersion is set at link time via -ldflags "-X api.BuildVersion=…".
// Defaults to "dev" when not overridden.
var BuildVersion = "dev"

type sourceRequest struct {
	Handle      string `json:"handle"`
	SourceClass string `json:"source_class"`
	City        string `json:"city,omitempty"`
	State       string `json:"state,omitempty"`
}

// listSources returns the curated watched-accounts list (vendors) with SLA
// fields: posts_stored, last_post_at, stale (no posts in 7 days).
func (s *Server) listSources(w http.ResponseWriter, r *http.Request) {
	src, err := s.Store.ListWatchedSources(r.URL.Query().Get("active") != "false")
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	staleAfter := 7 * 24 * time.Hour
	out := make([]map[string]any, 0, len(src))
	for _, wsrc := range src {
		n, _ := s.Store.CountSourcePosts(wsrc.Handle)
		last, _ := s.Store.SourceLastPostAt(wsrc.Handle)
		stale := true
		if last != nil {
			stale = time.Since(*last) > staleAfter
		}
		// Venues are "monitor" sources — agent discovery is for photographers.
		scanMode := "find_couples"
		if wsrc.SourceClass == "wedding_venue" || wsrc.SourceClass == "jeweler" ||
			wsrc.SourceClass == "wedding_publication" || wsrc.SourceClass == "registry_provider" {
			scanMode = "monitor_only"
		}
		m := map[string]any{
			"id": wsrc.ID, "handle": wsrc.Handle, "source_class": wsrc.SourceClass,
			"active": wsrc.Active, "state": wsrc.State, "city": wsrc.City,
			"full_name": wsrc.FullName, "profile_pic_url": wsrc.ProfilePicURL,
			"created_at":   wsrc.CreatedAt.UTC().Format(time.RFC3339),
			"posts_stored": n, "stale": stale, "scan_mode": scanMode,
		}
		if wsrc.LastScannedAt != nil {
			m["last_scanned_at"] = wsrc.LastScannedAt.UTC().Format(time.RFC3339)
		}
		if wsrc.LastScanCouples != nil {
			m["last_scan_couples"] = *wsrc.LastScanCouples
		}
		if wsrc.LastScanActions != nil {
			m["last_scan_actions"] = *wsrc.LastScanActions
		}
		if wsrc.FollowerCount != nil {
			m["follower_count"] = *wsrc.FollowerCount
		}
		if wsrc.FollowingCount != nil {
			m["following_count"] = *wsrc.FollowingCount
		}
		if wsrc.PostCount != nil {
			m["post_count"] = *wsrc.PostCount
		}
		if wsrc.Verified != nil {
			m["verified"] = *wsrc.Verified
		}
		if wsrc.ProfileCheckedAt != nil {
			m["profile_checked_at"] = wsrc.ProfileCheckedAt.UTC().Format(time.RFC3339)
		}
		if last != nil {
			m["last_post_at"] = last.UTC().Format(time.RFC3339)
		}
		out = append(out, m)
	}
	writeJSON(w, http.StatusOK, out)
}

// listSourcePosts returns actual ingested posts for one watched vendor.
func (s *Server) listSourcePosts(w http.ResponseWriter, r *http.Request) {
	handle := strings.TrimPrefix(r.PathValue("handle"), "@")
	if handle == "" {
		writeError(w, http.StatusBadRequest, errorString("handle required"))
		return
	}
	limit := 40
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			limit = n
		}
	}
	posts, err := s.Store.ListSourcePosts(handle, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if posts == nil {
		posts = []store.SourcePost{}
	}
	writeJSON(w, http.StatusOK, posts)
}

// listProspectBoard is the stage Kanban for couples/prospects.
func (s *Server) listProspectBoard(w http.ResponseWriter, r *http.Request) {
	cards, err := s.Store.ListProspectBoard(200)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if cards == nil {
		cards = []store.ProspectCard{}
	}
	// Group by column for convenient frontend consumption.
	board := map[string][]store.ProspectCard{
		store.ColTaggedPair:     {},
		store.ColInvestigating:  {},
		store.ColEngagedSignal:  {},
		store.ColReadyOutreach:  {},
		store.ColApprovedPaused: {},
	}
	for _, c := range cards {
		board[c.Column] = append(board[c.Column], c)
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"columns": []map[string]string{
			{"id": store.ColTaggedPair, "label": "Tagged pair"},
			{"id": store.ColInvestigating, "label": "Investigating"},
			{"id": store.ColEngagedSignal, "label": "Engaged signal"},
			{"id": store.ColReadyOutreach, "label": "Ready for outreach"},
			{"id": store.ColApprovedPaused, "label": "Approved / Paused"},
		},
		"cards": board,
		"total": len(cards),
	})
}

// listProspectPins feeds the map layer of geo-located prospects.
func (s *Server) listProspectPins(w http.ResponseWriter, r *http.Request) {
	pins, err := s.Store.ListProspectPins()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if pins == nil {
		pins = []store.ProspectPin{}
	}
	writeJSON(w, http.StatusOK, pins)
}

// addSource registers a public account to watch. source_class must be one of
// the known vendor classes — an unclassified source gets no vendor scoring
// weight, so accepting junk classes here would silently corrupt scores.
// Optional city/state are stored immediately; profile fetch then fills gaps.
func (s *Server) addSource(w http.ResponseWriter, r *http.Request) {
	var req sourceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	req.Handle = strings.TrimPrefix(strings.TrimSpace(req.Handle), "@")
	if req.Handle == "" || !signals.WatchedSourceClasses[req.SourceClass] {
		writeError(w, http.StatusBadRequest, errBadSource)
		return
	}
	src, err := s.Store.AddWatchedSourceWithGeo(req.Handle, req.SourceClass, strings.ToUpper(strings.TrimSpace(req.State)), strings.TrimSpace(req.City))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	// Best-effort profile + location enrich so the source is useful immediately.
	if s.Watch != nil {
		if err := s.Watch.EnrichSourceProfile(r.Context(), req.Handle); err != nil {
			log.Printf("enrich new source @%s: %v", req.Handle, err)
		} else if refreshed, err := s.Store.GetWatchedSource(req.Handle); err == nil {
			src = refreshed
		}
	}
	writeJSON(w, http.StatusCreated, src)
}

func (s *Server) removeSource(w http.ResponseWriter, r *http.Request) {
	if err := s.Store.DeactivateWatchedSource(r.PathValue("handle")); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deactivated"})
}

// patchSourceLocation lets ops set/correct city+state on a watched source.
func (s *Server) patchSourceLocation(w http.ResponseWriter, r *http.Request) {
	handle := strings.TrimPrefix(r.PathValue("handle"), "@")
	var req struct {
		City  string `json:"city"`
		State string `json:"state"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := s.Store.SetWatchedSourceLocation(handle, strings.TrimSpace(req.City), strings.ToUpper(strings.TrimSpace(req.State))); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	src, err := s.Store.GetWatchedSource(handle)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeJSON(w, http.StatusOK, src)
}

// scanSource starts an async agent job (preferred) or runs sync if ?sync=1.
func (s *Server) scanSource(w http.ResponseWriter, r *http.Request) {
	if s.Watch == nil {
		writeError(w, http.StatusServiceUnavailable, errorString("watch loop not configured"))
		return
	}
	handle := strings.TrimPrefix(r.PathValue("handle"), "@")
	limit := 15
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			limit = n
		}
	}
	// Default async so the UI can show progress without timing out.
	if r.URL.Query().Get("sync") == "1" {
		result, err := s.Watch.ScanSource(r.Context(), handle, limit)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusOK, result)
		return
	}
	jobID := s.Watch.StartScanJob(handle, limit)
	writeJSON(w, http.StatusAccepted, map[string]any{"job_id": jobID, "status": "queued"})
}

// scanJobStatus polls an async scan job.
func (s *Server) scanJobStatus(w http.ResponseWriter, r *http.Request) {
	if s.Watch == nil {
		writeError(w, http.StatusServiceUnavailable, errorString("watch loop not configured"))
		return
	}
	job, ok := s.Watch.GetScanJob(r.PathValue("id"))
	if !ok {
		writeError(w, http.StatusNotFound, errorString("job not found"))
		return
	}
	writeJSON(w, http.StatusOK, job)
}

// scanBulk starts agent scans for many sources (default: stale photographers).
func (s *Server) scanBulk(w http.ResponseWriter, r *http.Request) {
	if s.Watch == nil {
		writeError(w, http.StatusServiceUnavailable, errorString("watch loop not configured"))
		return
	}
	var body struct {
		StaleOnly   bool     `json:"stale_only"`
		Classes     []string `json:"classes"`
		Limit       int      `json:"limit"`
		PostsPerSrc int      `json:"posts_per_source"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	if body.PostsPerSrc <= 0 {
		body.PostsPerSrc = 12
	}
	if len(body.Classes) == 0 {
		// Photographers are the discovery engine; venues use "monitor only".
		body.Classes = []string{"engagement_photographer", "proposal_planner"}
	}
	classSet := map[string]bool{}
	for _, c := range body.Classes {
		classSet[c] = true
	}
	src, err := s.Store.ListWatchedSources(true)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	staleAfter := 7 * 24 * time.Hour
	var handles []string
	for _, wsrc := range src {
		if !classSet[wsrc.SourceClass] {
			continue
		}
		if body.StaleOnly {
			last, _ := s.Store.SourceLastPostAt(wsrc.Handle)
			if last != nil && time.Since(*last) < staleAfter {
				continue
			}
		}
		handles = append(handles, wsrc.Handle)
		if body.Limit > 0 && len(handles) >= body.Limit {
			break
		}
	}
	if len(handles) == 0 {
		writeJSON(w, http.StatusOK, map[string]any{"job_id": "", "handles": []string{}, "message": "nothing to scan"})
		return
	}
	jobID := s.Watch.StartBulkScanJob(handles, body.PostsPerSrc)
	writeJSON(w, http.StatusAccepted, map[string]any{"job_id": jobID, "handles": handles, "count": len(handles)})
}

// suppressVendorPairs removes board noise where both partners look like businesses.
func (s *Server) suppressVendorPairs(w http.ResponseWriter, r *http.Request) {
	// Build the registered-vendor set from the watched_sources table.
	registered := map[string]bool{}
	if all, err := s.Store.ListWatchedSources(false); err == nil {
		for _, ws := range all {
			registered[strings.ToLower(ws.Handle)] = true
		}
	}
	n, err := s.Store.SuppressVendorVendorCouples(registered)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"suppressed": n})
}

// enrichSource re-fetches profile stats + location without a full post scan.
func (s *Server) enrichSource(w http.ResponseWriter, r *http.Request) {
	if s.Watch == nil {
		writeError(w, http.StatusServiceUnavailable, errorString("watch loop not configured"))
		return
	}
	handle := strings.TrimPrefix(r.PathValue("handle"), "@")
	if err := s.Watch.EnrichSourceProfile(r.Context(), handle); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	src, err := s.Store.GetWatchedSource(handle)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeJSON(w, http.StatusOK, src)
}

var errBadSource = errorString("handle required and source_class must be a known vendor class (see signals.WatchedSourceClasses)")

type errorString string

func (e errorString) Error() string { return string(e) }

// ingestStatus reports per-monitor cursors, today's provider usage, and the
// global pause/play state — the ops view of what the watcher is doing.
func (s *Server) ingestStatus(w http.ResponseWriter, r *http.Request) {
	providerName := "none"
	paused := false
	providerOK := false
	poll := ""
	budget := 0
	if s.Watch != nil {
		providerName = s.Watch.ProviderName()
		paused = s.Watch.IsPaused()
		providerOK = s.Watch.ProviderAvailable()
		poll = s.Watch.PollInterval().String()
		budget = s.Watch.DailyBudget()
	}
	usageKey := providerName
	if usageKey == "none" {
		usageKey = "brightdata"
	}
	used, err := s.Store.UsageToday(usageKey)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	cursors, err := s.Store.ListCursors()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	running := providerOK && !paused
	writeJSON(w, http.StatusOK, map[string]any{
		"provider":           providerName,
		"provider_available": providerOK,
		"paused":             paused,
		"running":            running,
		"poll_interval":      poll,
		"daily_budget":       budget,
		"results_used_today": used,
		"cursors":            cursors,
	})
}

// pauseIngest freezes the global watch loop — no provider fetches until resume.
func (s *Server) pauseIngest(w http.ResponseWriter, r *http.Request) {
	if s.Watch == nil {
		writeError(w, http.StatusServiceUnavailable, errorString("watch loop not configured"))
		return
	}
	s.Watch.Pause()
	if _, err := s.Store.Audit("ingest", "watch_loop", "paused", "operator paused global watch loop via dashboard", "", 0); err != nil {
		log.Printf("audit pause: %v", err)
	}
	writeJSON(w, http.StatusOK, map[string]any{"paused": true, "running": false})
}

// resumeIngest re-enables the global watch loop on the next tick.
func (s *Server) resumeIngest(w http.ResponseWriter, r *http.Request) {
	if s.Watch == nil {
		writeError(w, http.StatusServiceUnavailable, errorString("watch loop not configured"))
		return
	}
	s.Watch.Resume()
	if _, err := s.Store.Audit("ingest", "watch_loop", "resumed", "operator resumed global watch loop via dashboard", "", 0); err != nil {
		log.Printf("audit resume: %v", err)
	}
	running := s.Watch.ProviderAvailable()
	writeJSON(w, http.StatusOK, map[string]any{"paused": false, "running": running})
}
