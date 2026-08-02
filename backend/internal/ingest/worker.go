package ingest

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"log"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"neptune-social-radar/backend/internal/llm"
	"neptune-social-radar/backend/internal/ontology"
	"neptune-social-radar/backend/internal/pipeline"
	"neptune-social-radar/backend/internal/pipeline/watchtower"
	"neptune-social-radar/backend/internal/signals"
	"neptune-social-radar/backend/internal/store"
)

// WorkerConfig tunes the watch loop. All knobs are env-wired in cmd/server.
type WorkerConfig struct {
	PollInterval        time.Duration
	DryRun              bool // fetch and log, but never store or process events
	DailyBudget         int  // max provider results per UTC day; 0 = block all fetching
	ActiveMarkets       []string
	ResultsPerRun       int // per-actor-call results limit (the spend quantum)
	FollowChecksPerTick int
	Client              SocialProvider
	Vision              llm.VisionClassifier
}

// Worker is the live watch loop. Every tick it polls the monitored hashtags,
// the curated vendor accounts, and the bios of known-couple accounts; lazily
// re-verifies mutual-follow state for couples that need it; and feeds every
// new event into the pipeline.
//
// Pause/Resume is operator-controlled via the dashboard: when paused, ticks
// are skipped (no provider spend) but the loop keeps running so Resume is
// instant — no process restart.
type Worker struct {
	store  *store.Store
	orch   *pipeline.Orchestrator
	cfg    WorkerConfig
	vision llm.VisionClassifier
	paused atomic.Bool
	jobs   *jobStore
}

func NewWorker(s *store.Store, orch *pipeline.Orchestrator, cfg WorkerConfig) *Worker {
	if cfg.ResultsPerRun <= 0 {
		cfg.ResultsPerRun = 30
	}
	if cfg.FollowChecksPerTick <= 0 {
		cfg.FollowChecksPerTick = 2
	}
	if cfg.Vision == nil {
		cfg.Vision = llm.NoopVision{}
	}
	return &Worker{store: s, orch: orch, cfg: cfg, vision: cfg.Vision, jobs: newJobStore()}
}

// Pause stops provider polling until Resume. Safe for concurrent API use.
func (w *Worker) Pause() {
	if w.paused.CompareAndSwap(false, true) {
		log.Println("[watchtower] PAUSED — ticks will skip until resume")
	}
}

// Resume re-enables provider polling on the next tick.
func (w *Worker) Resume() {
	if w.paused.CompareAndSwap(true, false) {
		log.Println("[watchtower] RESUMED — polling active")
	}
}

// IsPaused reports whether the operator has paused the watch loop.
func (w *Worker) IsPaused() bool { return w.paused.Load() }

// ProviderAvailable is true when Bright Data or Apify is configured.
func (w *Worker) ProviderAvailable() bool {
	return w.cfg.Client != nil && w.cfg.Client.Available()
}

// ProviderName is "brightdata", "apify", or "none".
func (w *Worker) ProviderName() string {
	if w.cfg.Client == nil || !w.cfg.Client.Available() {
		return "none"
	}
	return w.cfg.Client.Name()
}

// PollInterval returns the configured tick interval.
func (w *Worker) PollInterval() time.Duration { return w.cfg.PollInterval }

// DailyBudget returns the configured daily provider result cap.
func (w *Worker) DailyBudget() int { return w.cfg.DailyBudget }

func (w *Worker) Run(ctx context.Context) {
	if !w.ProviderAvailable() {
		log.Println("[watchtower] no social provider (set BRIGHTDATA_API_KEY or APIFY_TOKEN) — watch loop idle")
		// Stay alive so Pause/Resume state is meaningful even without a token;
		// ticks no-op until a provider is configured on restart.
		<-ctx.Done()
		log.Println("[watchtower] watch loop stopped")
		return
	}
	// Leader election: a second replica would double-spend the provider budget
	// and double-process events (idempotency saves correctness, not money).
	// The session-level advisory lock auto-releases on disconnect, so a crashed
	// worker frees it without a heartbeat. If we're not the leader, idle the
	// poll loop but stay alive — the API and Pause/Resume still work.
	leader, err := w.store.TryAcquireLeaderLock()
	if err != nil {
		log.Printf("[watchtower] leader lock acquire failed, refusing to poll: %v", err)
		<-ctx.Done()
		return
	}
	if !leader {
		log.Println("[watchtower] another replica holds the leader lock — poll loop idle (API still serving)")
		<-ctx.Done()
		log.Println("[watchtower] watch loop stopped (non-leader)")
		return
	}
	log.Println("[watchtower] acquired leader lock — this replica owns the poll loop")
	log.Printf("[watchtower] watch loop started provider=%s (interval=%s, budget=%d results/day, markets=%v, dry_run=%v)",
		w.ProviderName(), w.cfg.PollInterval, w.cfg.DailyBudget, w.cfg.ActiveMarkets, w.cfg.DryRun)
	w.Tick(ctx) // first poll immediately, then on the interval
	ticker := time.NewTicker(w.cfg.PollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			log.Println("[watchtower] watch loop stopped")
			return
		case <-ticker.C:
			w.Tick(ctx)
		}
	}
}

// Tick runs one full poll across all sources. Exported for tests and ops
// triggers.
func (w *Worker) Tick(ctx context.Context) {
	if w.paused.Load() {
		log.Println("[watchtower] tick skipped — paused")
		return
	}
	if !w.ProviderAvailable() {
		return
	}
	// One query for the handle→class map used by SourceClassForHandle across
	// the whole tick — replaces up to 6 SELECTs per post (worker stamp +
	// roles + scorer).
	if err := w.store.RefreshSourceClassCache(); err != nil {
		log.Printf("[watchtower] refresh source class cache: %v", err)
	}
	w.pollHashtags(ctx)
	w.pollVendors(ctx)
	w.pollProfiles(ctx)
	w.checkFollowStates(ctx)
}

// budgetRemaining reports how many provider results are still affordable
// today; fetching stops at zero. Scraping APIs bill per result — this cap is
// the difference between a monitoring system and a surprise invoice.
func (w *Worker) budgetRemaining(ctx context.Context) int {
	if w.cfg.DailyBudget <= 0 {
		return 0
	}
	provider := w.ProviderName()
	if provider == "none" {
		return 0
	}
	used, err := w.store.UsageToday(provider)
	if err != nil {
		log.Printf("[watchtower] usage check failed, refusing to spend: %v", err)
		return 0
	}
	return w.cfg.DailyBudget - used
}

func (w *Worker) pollHashtags(ctx context.Context) {
	hashtags := signals.MonitoredHashtags(w.cfg.ActiveMarkets)
	if len(hashtags) == 0 {
		return
	}
	if w.budgetRemaining(ctx) < w.cfg.ResultsPerRun {
		log.Printf("[watchtower] daily budget exhausted — skipping hashtag monitor")
		return
	}
	items, err := w.cfg.Client.FetchHashtagPosts(ctx, hashtags, w.cfg.ResultsPerRun)
	if err != nil {
		log.Printf("[watchtower] hashtag fetch failed: %v", err)
		return
	}
	w.recordUsage("hashtag:batch", len(items))
	// Schema drift canary: catch Apify actor upgrades that silently rename/drop
	// fields before the whole batch degrades to zero-signal.
	if report := CheckSchemaDrift("hashtag", items); report.Drifted {
		log.Printf("[watchtower] SCHEMA DRIFT on hashtag batch: %v", report.MissingFields)
	}
	newest := time.Time{}
	for _, item := range items {
		raw, imageURL, err := MapPost(item, "hashtag:batch")
		if err != nil {
			w.deadLetter("apify:hashtag", "hashtag:batch", item, err)
			continue
		}
		raw.Monitor = attributeHashtagMonitor(raw.Payload)
		w.processPost(ctx, raw, imageURL)
		if raw.OccurredAt.After(newest) {
			newest = raw.OccurredAt
		}
	}
	w.advance("hashtag:batch", newest)
}

// profileRefreshInterval caps how often a vendor's profile (follower/post
// counts) is re-fetched. Those numbers move slowly — there's no reason to
// spend an Apify result on them every 15-minute tick like post-polling does.
const profileRefreshInterval = 24 * time.Hour

// pollVendors fetches recent posts for ALL active vendor accounts in one
// batched Apify call, then partitions results by owner. One round-trip
// replaces N sequential blocking calls — the actor accepts username: [] and
// returns mixed results, each tagged with ownerUsername.
func (w *Worker) pollVendors(ctx context.Context) {
	sources, err := w.store.ListWatchedSources(true)
	if err != nil {
		log.Printf("[watchtower] list watched sources: %v", err)
		return
	}
	if len(sources) == 0 {
		return
	}

	handles := make([]string, len(sources))
	for i, src := range sources {
		handles[i] = src.Handle
	}

	// Cap the request at the budget remaining — the actor's resultsLimit is a
	// total cap across all handles, so this never overspends.
	limit := w.cfg.ResultsPerRun * len(handles)
	remaining := w.budgetRemaining(ctx)
	if remaining < w.cfg.ResultsPerRun {
		log.Printf("[watchtower] daily budget exhausted — skipping vendor batch")
		return
	}
	if limit > remaining {
		limit = remaining
	}

	items, err := w.cfg.Client.FetchAccountPosts(ctx, handles, limit)
	if err != nil {
		log.Printf("[watchtower] vendor batch fetch failed: %v", err)
		return
	}

	// One pass: map, process, and track per-vendor newest timestamp and item
	// count for cursor advancement and usage accounting.
	newestByVendor := make(map[string]time.Time, len(sources))
	countByVendor := make(map[string]int, len(sources))
	if report := CheckSchemaDrift("vendor", items); report.Drifted {
		log.Printf("[watchtower] SCHEMA DRIFT on vendor batch: %v", report.MissingFields)
	}
	for _, item := range items {
		raw, imageURL, err := MapPost(item, "vendor:batch")
		if err != nil {
			w.deadLetter("apify:vendor", "vendor:batch", item, err)
			continue
		}
		// Re-attribute to the specific vendor's monitor — the audit trail
		// still shows which vendor surfaced each post.
		raw.Monitor = "vendor:" + raw.Handle
		w.processPost(ctx, raw, imageURL)
		countByVendor[raw.Handle]++
		if raw.OccurredAt.After(newestByVendor[raw.Handle]) {
			newestByVendor[raw.Handle] = raw.OccurredAt
		}
	}

	// Advance cursors and refresh profiles per vendor — vendors with no posts
	// in this batch still get TouchCursor so their staleness is recorded.
	for _, src := range sources {
		monitor := "vendor:" + src.Handle
		w.recordUsage(monitor, countByVendor[src.Handle])
		w.advance(monitor, newestByVendor[src.Handle])
		w.refreshVendorProfileIfStale(ctx, src)
	}
}

// refreshVendorProfileIfStale fetches a vendor's real follower/post counts
// via the same Apify pipeline used everywhere else, if they've never been
// checked or the last check is older than profileRefreshInterval — every
// watched source gets this, not just ones seeded by a one-off script.
func (w *Worker) refreshVendorProfileIfStale(ctx context.Context, src ontology.WatchedSource) {
	if src.ProfileCheckedAt != nil && time.Since(*src.ProfileCheckedAt) < profileRefreshInterval {
		return
	}
	if w.budgetRemaining(ctx) < 1 {
		return
	}
	// Full profile enrich (stats + city/state from bio/address) so every
	// watched source carries a market location for map + prospect scoring.
	if err := w.EnrichSourceProfile(ctx, src.Handle); err != nil {
		log.Printf("[watchtower] profile refresh %s failed: %v", src.Handle, err)
	}
}

// pollProfiles re-reads the bios of known-couple accounts and emits a
// bio_change event when one changed — the live input to both the engagement
// path (fiancé bio appears) and the relationship-state path (bio cleared).
func (w *Worker) pollProfiles(ctx context.Context) {
	handles, err := w.store.ListProfileWatchHandles()
	if err != nil {
		log.Printf("[watchtower] list profile-watch handles: %v", err)
		return
	}
	for _, handle := range handles {
		if w.budgetRemaining(ctx) < 1 {
			return
		}
		items, err := w.cfg.Client.FetchProfile(ctx, handle)
		if err != nil {
			log.Printf("[watchtower] profile fetch %s failed: %v", handle, err)
			continue
		}
		w.recordUsage("profile:"+handle, len(items))
		if len(items) == 0 {
			continue
		}
		prof, ok := ParseProfile(items[0])
		if !ok {
			continue
		}
		acct, err := w.store.GetAccountByHandle("instagram", handle)
		if err != nil {
			continue
		}
		// Always refresh avatar/stats when we paid for a profile fetch.
		_ = w.store.UpdateAccountProfile(acct.ID, prof.DisplayName, prof.Bio, prof.ProfilePicURL,
			prof.FollowerCount, prof.FollowingCount, prof.IsPrivate, time.Now().UTC())
		if loc, ok := signals.InferLocationFromText(prof.Bio, "bio"); ok {
			_ = w.store.UpdateAccountLocation(acct.ID, loc.City, loc.Region, loc.Source)
		}
		// Persist Instagram business address (free street-level data when present)
		if prof.StreetAddress != "" {
			_ = w.store.UpdateAccountBusinessAddress(acct.ID, prof.StreetAddress, prof.BusinessCity, prof.BusinessState, prof.BusinessPostal)
		}
		if acct.BioText == prof.Bio {
			continue // bio unchanged — no bio_change event
		}
		w.process(ctx, watchtower.RawEvent{
			Monitor:         "profile:" + handle,
			Source:          "apify",
			ExternalEventID: "bio:" + handle + ":" + shortHash(prof.Bio),
			Handle:          handle,
			Type:            "bio_change",
			Payload:         map[string]any{"bio": prof.Bio},
			OccurredAt:      time.Now().UTC(),
		})
	}
}

// checkFollowStates lazily re-verifies mutual follows for couples with an
// open hypothesis or a live high-stakes stage. This is the ONLY place
// follower-list pulls happen — they are the most expensive provider call per
// result, so they run at most FollowChecksPerTick couples per tick and never
// speculatively.
func (w *Worker) checkFollowStates(ctx context.Context) {
	targets, err := w.store.ListCouplesForFollowCheck(w.cfg.FollowChecksPerTick)
	if err != nil {
		log.Printf("[watchtower] list follow-check couples: %v", err)
		return
	}
	for _, t := range targets {
		for _, pair := range [][2]string{{t.HandleA, t.HandleB}, {t.HandleB, t.HandleA}} {
			budget := w.budgetRemaining(ctx)
			if budget < 1 {
				return
			}
			// Spend at most what remains: the gate used to check "≥30 left"
			// then fire a 1000-result fetch — 3× the daily cap in one call.
			limit := 1000
			if budget < limit {
				limit = budget
			}
			items, err := w.cfg.Client.FetchFollowing(ctx, pair[0], limit)
			if err != nil {
				log.Printf("[watchtower] following fetch %s failed: %v", pair[0], err)
				continue
			}
			w.recordUsage("follow:"+pair[0], len(items))
			active := false
			for _, h := range ParseFollowingUsernames(items) {
				if strings.EqualFold(h, pair[1]) {
					active = true
					break
				}
			}
			// Only a real state CHANGE is an event. The old dedup key
			// embedded the date, so an unchanged state minted a fresh
			// observation row per couple-direction per day forever — and a
			// true→false→true flap inside one day collided with the day's
			// earlier key and was silently suppressed.
			current, err := w.store.FollowState(pair[0], pair[1])
			if err != nil {
				log.Printf("[watchtower] follow state read %s→%s: %v", pair[0], pair[1], err)
				continue
			}
			if current == nil || *current == active {
				continue // no change — nothing to record
			}
			w.process(ctx, watchtower.RawEvent{
				Monitor:         "follow:" + pair[0],
				Source:          "apify",
				ExternalEventID: fmt.Sprintf("follow:%s>%s:%v:%d", pair[0], pair[1], active, time.Now().UTC().Unix()),
				Handle:          pair[0],
				Type:            "follow_change",
				Payload:         map[string]any{"target_handle": pair[1], "active": active},
				OccurredAt:      time.Now().UTC(),
			})
		}
	}
}

// processPost enriches and gates one post-shaped event: cross-monitor dedupe,
// vendor source classification, then — only for posts that are already
// engagement candidates — a vision-classifier call to add visual signals.
func (w *Worker) processPost(ctx context.Context, raw watchtower.RawEvent, imageURL string) {
	if raw.Handle == "" {
		return // a post with no author account is unusable
	}
	exists, err := w.store.ObservationExists(raw.ExternalEventID)
	if err != nil {
		log.Printf("[watchtower] dedupe check failed: %v", err)
		return
	}
	if exists {
		return // same post surfaced under another monitor already
	}
	if class := w.store.SourceClassForHandle(raw.Handle); class != "" {
		raw.Payload["source_account_type"] = class
	}
	// Vision gating: the classifier only spends a call on posts the cheap
	// deterministic vocabulary already considers engagement-shaped. A #Love
	// post never costs a model call.
	if imageURL != "" {
		sig := signals.ExtractFromPayload(raw.Payload)
		if sig.CreatesCandidate() {
			labels, err := w.vision.ClassifyVisualSignals(ctx, imageURL)
			// Record the classification for calibration tracking — even
			// failures, so we can see the error rate over time.
			model := w.visionModelName()
			if logErr := w.store.RecordVisionClassification("", raw.ExternalEventID, imageURL, model, labels, errToString(err)); logErr != nil {
				log.Printf("[watchtower] vision log failed: %v", logErr)
			}
			if err != nil {
				log.Printf("[watchtower] vision classify failed for %s: %v", raw.ExternalEventID, err)
			} else if len(labels) > 0 {
				raw.Payload["visual_signals"] = labels
			}
		}
	}
	w.process(ctx, raw)
	// After a post is processed, enrich tagged people (profile pic + bio) so
	// the prospect board and map can show real avatars and locations.
	if raw.Type == "post" {
		w.enrichTaggedProfiles(ctx, raw)
	}
}

func (w *Worker) process(ctx context.Context, raw watchtower.RawEvent) {
	if w.cfg.DryRun {
		log.Printf("[watchtower] DRY RUN event: monitor=%s type=%s handle=%s payload=%v",
			raw.Monitor, raw.Type, raw.Handle, raw.Payload)
		return
	}
	// Retry transient DB errors (connection blips, deadlocks). A single
	// hiccup no longer drops an entire event permanently — the event is
	// only lost after maxRetries attempts. Non-retryable errors (e.g.
	// ErrDuplicateObservation) are returned immediately by ProcessEvent
	// and not retried.
	const maxRetries = 2
	var result pipeline.StepResult
	var err error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		result, err = w.orch.ProcessEvent(ctx, raw)
		if err == nil {
			break
		}
		if !isRetryableDBError(err) {
			break
		}
		if attempt < maxRetries {
			wait := time.Duration(1<<attempt) * time.Second // 1s, 2s
			log.Printf("[watchtower] transient error on %s (attempt %d), retrying in %s: %v",
				raw.ExternalEventID, attempt+1, wait, err)
			select {
			case <-ctx.Done():
				return
			case <-time.After(wait):
			}
		}
	}
	if err != nil {
		log.Printf("[watchtower] process event %s failed after %d attempts: %v", raw.ExternalEventID, maxRetries+1, err)
		return
	}
	switch {
	case result.Duplicate:
	case result.ActionCreated != "":
		log.Printf("[watchtower] ACTION %s created for %s (confidence %.2f)", result.ActionCreated, raw.ExternalEventID, result.FinalConfidence)
	case result.HypothesisID != "":
		log.Printf("[watchtower] hypothesis %s scored %.2f (no action)", result.HypothesisID, result.FinalConfidence)
	}
}

// isRetryableDBError reports whether an error from ProcessEvent is worth
// retrying: connection errors, deadlocks, and serialization failures are
// transient. Unique violations (duplicate events) and bad input are not.
func isRetryableDBError(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	// ponytail: string matching instead of typed errors because the pgx
	// error types are wrapped through multiple layers and the typed check
	// (pgconn.PgError code) only works on the unwrapped leaf. The strings
	// are stable Postgres error messages.
	for _, pattern := range []string{
		"connection reset",
		"connection refused",
		"deadlock detected",
		"could not serialize",
		"server closed the connection",
		"i/o timeout",
		"context deadline exceeded",
	} {
		if strings.Contains(s, pattern) {
			return true
		}
	}
	return false
}

// profileEnrichMaxAge avoids re-fetching the same tagged account every tick.
const profileEnrichMaxAge = 72 * time.Hour

// enrichTaggedProfiles pulls Instagram profile pic + bio for people tagged on
// a post (budget-gated). Location is inferred from bios / post geotag and
// written onto the couple when one exists.
func (w *Worker) enrichTaggedProfiles(ctx context.Context, raw watchtower.RawEvent) {
	if w.paused.Load() || !w.ProviderAvailable() {
		return
	}
	sig := signals.ExtractFromPayload(raw.Payload)
	handles := append([]string{}, sig.TaggedHandles...)
	if len(handles) == 0 {
		return
	}
	// Cap enrichment per post so one viral post can't blow the budget.
	const maxPerPost = 4
	if len(handles) > maxPerPost {
		handles = handles[:maxPerPost]
	}

	var bios []string
	for _, handle := range handles {
		if strings.EqualFold(handle, raw.Handle) {
			continue // skip the vendor author
		}
		acct, err := w.store.GetAccountByHandle("instagram", handle)
		if err != nil {
			// Identity stage usually created them; if not, skip.
			continue
		}
		need, err := w.store.NeedsProfileRefresh(acct.ID, profileEnrichMaxAge)
		if err != nil || !need {
			if acct.BioText != "" {
				bios = append(bios, acct.BioText)
			}
			continue
		}
		if w.budgetRemaining(ctx) < 1 {
			return
		}
		items, err := w.cfg.Client.FetchProfile(ctx, handle)
		if err != nil || len(items) == 0 {
			log.Printf("[watchtower] enrich profile %s: %v", handle, err)
			continue
		}
		w.recordUsage("profile:"+handle, len(items))
		prof, ok := ParseProfile(items[0])
		if !ok {
			continue
		}
		now := time.Now().UTC()
		if err := w.store.UpdateAccountProfile(acct.ID, prof.DisplayName, prof.Bio, prof.ProfilePicURL,
			prof.FollowerCount, prof.FollowingCount, prof.IsPrivate, now); err != nil {
			log.Printf("[watchtower] save profile %s: %v", handle, err)
			continue
		}
		if loc, ok := signals.InferLocationFromText(prof.Bio, "bio"); ok {
			_ = w.store.UpdateAccountLocation(acct.ID, loc.City, loc.Region, loc.Source)
		}
		if prof.Bio != "" {
			bios = append(bios, prof.Bio)
		}
	}

	// Couple location from post geotag + both bios.
	postLoc, _ := raw.Payload["location"].(string)
	caption, _ := raw.Payload["caption"].(string)
	bioA, bioB := "", ""
	if len(bios) > 0 {
		bioA = bios[0]
	}
	if len(bios) > 1 {
		bioB = bios[1]
	}
	if loc, ok := signals.BestLocation(postLoc, bioA, bioB, caption); ok {
		// Attach to couple if the two tagged accounts form one.
		if len(handles) >= 2 {
			a, errA := w.store.GetAccountByHandle("instagram", handles[0])
			b, errB := w.store.GetAccountByHandle("instagram", handles[1])
			if errA == nil && errB == nil {
				if couple, err := w.store.GetCoupleForAccountPair(a.ID, b.ID); err == nil {
					lat, lng := cityCoords(loc.City, loc.Region)
					_ = w.store.UpdateCoupleLocation(couple.ID, loc.City, loc.Region, loc.Source, lat, lng)
				}
			}
		}
	}
}

// cityCoords is a tiny static table for map pins without a geocoder.
func cityCoords(city, region string) (*float64, *float64) {
	key := strings.ToLower(strings.TrimSpace(city)) + "|" + strings.ToUpper(strings.TrimSpace(region))
	table := map[string][2]float64{
		"columbus|OH":     {39.9612, -82.9988},
		"cleveland|OH":    {41.4993, -81.6944},
		"cincinnati|OH":   {39.1031, -84.5120},
		"dublin|OH":       {40.0992, -83.1141},
		"worthington|OH":  {40.0931, -83.0180},
		"westerville|OH":  {40.1262, -82.9291},
		"brooklyn|NY":     {40.6782, -73.9442},
		"manhattan|NY":    {40.7831, -73.9712},
		"new york|NY":     {40.7128, -74.0060},
		"los angeles|CA":  {34.0522, -118.2437},
		"chicago|IL":      {41.8781, -87.6298},
		"miami|FL":        {25.7617, -80.1918},
		"austin|TX":       {30.2672, -97.7431},
		"dallas|TX":       {32.7767, -96.7970},
		"houston|TX":      {29.7604, -95.3698},
		"seattle|WA":      {47.6062, -122.3321},
		"boston|MA":       {42.3601, -71.0589},
		"philadelphia|PA": {39.9526, -75.1652},
		"denver|CO":       {39.7392, -104.9903},
		"atlanta|GA":      {33.7490, -84.3880},
	}
	if c, ok := table[key]; ok {
		lat, lng := c[0], c[1]
		return &lat, &lng
	}
	// Try city only
	for k, c := range table {
		parts := strings.SplitN(k, "|", 2)
		if parts[0] == strings.ToLower(strings.TrimSpace(city)) {
			lat, lng := c[0], c[1]
			return &lat, &lng
		}
	}
	return nil, nil
}

func (w *Worker) recordUsage(monitor string, results int) {
	if results <= 0 {
		return
	}
	provider := w.ProviderName()
	if provider == "none" {
		provider = "brightdata"
	}
	if err := w.store.RecordUsage(provider, monitor, results); err != nil {
		log.Printf("[watchtower] record usage: %v", err)
	}
}

// deadLetter persists an unmappable provider item to the DLQ instead of
// silently dropping it. The raw item JSON is stored so it can be inspected
// or replayed later (e.g. after a mapper fix).
func (w *Worker) deadLetter(source, monitor string, item json.RawMessage, mapErr error) {
	log.Printf("[watchtower] dead-letter %s item: %v", source, mapErr)
	raw := ""
	if item != nil {
		raw = string(item)
	}
	if err := w.store.InsertDLQ(source, monitor, "", raw, mapErr.Error()); err != nil {
		log.Printf("[watchtower] dlq insert failed: %v", err)
	}
}

// visionModelName returns the name of the active vision classifier for the
// calibration log. Falls back to "noop" when no model is configured.
func (w *Worker) visionModelName() string {
	if v, ok := w.vision.(interface{ Name() string }); ok {
		return v.Name()
	}
	return "noop"
}

func errToString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func (w *Worker) advance(monitor string, newest time.Time) {
	var err error
	if newest.IsZero() {
		err = w.store.TouchCursor(monitor)
	} else {
		err = w.store.AdvanceCursor(monitor, newest)
	}
	if err != nil {
		log.Printf("[watchtower] advance cursor %s: %v", monitor, err)
	}
}

// attributeHashtagMonitor names the monitor after the first watched hashtag
// the post actually carries, so the audit trail shows which signal found it.
func attributeHashtagMonitor(payload map[string]any) string {
	sig := signals.ExtractFromPayload(payload)
	var hits []string
	hits = append(hits, sig.HighIntentHits...)
	if len(hits) > 0 {
		sort.Strings(hits)
		return "hashtag:" + hits[0]
	}
	return "hashtag:batch"
}

func shortHash(s string) string {
	h := fnv.New32a()
	_, _ = h.Write([]byte(s))
	return strconv.FormatUint(uint64(h.Sum32()), 16)
}
