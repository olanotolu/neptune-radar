package ingest

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"neptune-social-radar/backend/internal/ontology"
	"neptune-social-radar/backend/internal/pipeline/watchtower"
	"neptune-social-radar/backend/internal/signals"
)

// SourceScanResult is the operator-facing summary after running the agent on
// one watched source: posts fetched, tagged couples discovered, and any
// approval actions created for the human queue.
type SourceScanResult struct {
	Handle           string              `json:"handle"`
	PostsFetched     int                 `json:"posts_fetched"`
	PostsProcessed   int                 `json:"posts_processed"`
	Duplicates       int                 `json:"duplicates"`
	TaggedPosts      int                 `json:"tagged_posts"`
	ActionsCreated   int                 `json:"actions_created"`
	City             string              `json:"city,omitempty"`
	State            string              `json:"state,omitempty"`
	FullName         string              `json:"full_name,omitempty"`
	ProfilePicURL    string              `json:"profile_pic_url,omitempty"`
	FollowerCount    *int                `json:"follower_count,omitempty"`
	Couples          []ScannedCouple     `json:"couples"`
	PendingApprovals []ScannedApproval   `json:"pending_approvals"`
	Errors           []string            `json:"errors,omitempty"`
	DurationMs       int64               `json:"duration_ms"`
}

// ScannedCouple is a pair discovered from tags on this source's posts.
type ScannedCouple struct {
	CoupleID     string   `json:"couple_id,omitempty"`
	HandleA      string   `json:"handle_a"`
	HandleB      string   `json:"handle_b"`
	Tags         []string `json:"tags"`         // person tags used for the pair
	VendorTags   []string `json:"vendor_tags,omitempty"` // filtered-out business tags
	PostURL      string   `json:"post_url,omitempty"`
	Caption      string   `json:"caption,omitempty"`
	ImageURL     string   `json:"image_url,omitempty"`
	ActionID     string   `json:"action_id,omitempty"`
	ActionType   string   `json:"action_type,omitempty"`
	Confidence   float64  `json:"confidence,omitempty"`
	HypothesisID string   `json:"hypothesis_id,omitempty"`
	// Quality 0–100: how likely this is a real couple vs vendor soup.
	Quality       int    `json:"quality"`
	QualityLabel  string `json:"quality_label"` // "strong_couple" | "likely_couple" | "weak" | "vendor_noise"
	HasPeopleShot bool   `json:"has_people_shot"`
	SkippedReason string `json:"skipped_reason,omitempty"`
}

// ScannedApproval is a recommended_action that needs human approve/ignore.
type ScannedApproval struct {
	ActionID   string  `json:"action_id"`
	ActionType string  `json:"action_type"`
	CoupleID   string  `json:"couple_id,omitempty"`
	HandleA    string  `json:"handle_a,omitempty"`
	HandleB    string  `json:"handle_b,omitempty"`
	Confidence float64 `json:"confidence,omitempty"`
}

// EnrichAccountProfile fetches IG profile for any handle (prospect side) and
// stores pic/bio/location on social_accounts.
func (w *Worker) EnrichAccountProfile(ctx context.Context, handle string) error {
	handle = strings.TrimPrefix(strings.TrimSpace(handle), "@")
	if handle == "" || !w.ProviderAvailable() {
		return fmt.Errorf("profile enrich unavailable")
	}
	if w.budgetRemaining(ctx) < 1 {
		return fmt.Errorf("daily provider budget exhausted")
	}
	items, err := w.cfg.Client.FetchProfile(ctx, handle)
	if err != nil {
		return err
	}
	w.recordUsage("profile:"+handle, max(1, len(items)))
	if len(items) == 0 {
		return fmt.Errorf("no profile data for @%s", handle)
	}
	prof, ok := ParseProfile(items[0])
	if !ok {
		return fmt.Errorf("unparseable profile for @%s", handle)
	}
	acct, err := w.store.EnsureAccount(ontology.SocialAccount{Handle: handle})
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	if err := w.store.UpdateAccountProfile(acct.ID, prof.DisplayName, prof.Bio, prof.ProfilePicURL,
		prof.FollowerCount, prof.FollowingCount, prof.IsPrivate, now); err != nil {
		return err
	}
	if loc, ok := signals.InferLocationFromText(prof.Bio, "bio"); ok {
		_ = w.store.UpdateAccountLocation(acct.ID, loc.City, loc.Region, loc.Source)
	} else if prof.City != "" {
		_ = w.store.UpdateAccountLocation(acct.ID, prof.City, prof.Region, prof.LocationSource)
	}
	return nil
}

// EnrichSourceProfile fetches Instagram profile for a watched vendor and
// persists stats + location (city/state from profile address or bio).
func (w *Worker) EnrichSourceProfile(ctx context.Context, handle string) error {
	handle = strings.TrimPrefix(strings.TrimSpace(handle), "@")
	if handle == "" || !w.ProviderAvailable() {
		return fmt.Errorf("profile enrich unavailable")
	}
	if w.budgetRemaining(ctx) < 1 {
		return fmt.Errorf("daily provider budget exhausted")
	}
	items, err := w.cfg.Client.FetchProfile(ctx, handle)
	if err != nil {
		return err
	}
	w.recordUsage("profile:"+handle, max(1, len(items)))
	if len(items) == 0 {
		return fmt.Errorf("no profile data for @%s", handle)
	}
	prof, ok := ParseProfile(items[0])
	if !ok {
		return fmt.Errorf("unparseable profile for @%s", handle)
	}
	// Prefer signals package for richer bio location when mapper only got partial.
	city, state := prof.City, prof.Region
	if city == "" && prof.Bio != "" {
		if loc, ok := signals.InferLocationFromText(prof.Bio, "bio"); ok {
			city, state = loc.City, loc.Region
			prof.LocationSource = loc.Source
		}
	}
	followers, following, posts := 0, 0, 0
	if prof.FollowerCount != nil {
		followers = *prof.FollowerCount
	}
	if prof.FollowingCount != nil {
		following = *prof.FollowingCount
	}
	if prof.PostCount != nil {
		posts = *prof.PostCount
	}
	if err := w.store.UpdateWatchedSourceProfileAndGeo(handle, followers, following, posts,
		prof.DisplayName, prof.ProfilePicURL, prof.Verified, city, state); err != nil {
		return err
	}
	log.Printf("[watchtower] enriched source @%s location=%s,%s followers=%d", handle, city, state, followers)
	return nil
}

// ScanSource runs the agent on one watched source: refresh profile+location,
// fetch recent posts, process each through the pipeline (tags → couples →
// approval actions). Safe to call from the API (operator-initiated).
func (w *Worker) ScanSource(ctx context.Context, handle string, postsLimit int) (SourceScanResult, error) {
	start := time.Now()
	handle = strings.TrimPrefix(strings.TrimSpace(strings.ToLower(handle)), "@")
	out := SourceScanResult{Handle: handle, Couples: []ScannedCouple{}, PendingApprovals: []ScannedApproval{}}
	if handle == "" {
		return out, fmt.Errorf("handle required")
	}
	if !w.ProviderAvailable() {
		return out, fmt.Errorf("social provider not configured (BRIGHTDATA_API_KEY) — cannot scan")
	}
	if w.paused.Load() {
		// Operator scan is explicit; allow it even when global loop is paused.
		log.Printf("[watchtower] scan @%s while loop paused (operator override)", handle)
	}
	if postsLimit <= 0 {
		postsLimit = 20
	}
	if postsLimit > 50 {
		postsLimit = 50
	}

	// 1) Profile + location (best-effort; always load what we already stored)
	if err := w.EnrichSourceProfile(ctx, handle); err != nil {
		out.Errors = append(out.Errors, "profile: "+err.Error())
	}
	if src, err := w.store.GetWatchedSource(handle); err == nil {
		out.City, out.State = src.City, src.State
		out.FullName, out.ProfilePicURL = src.FullName, src.ProfilePicURL
		out.FollowerCount = src.FollowerCount
	}

	// 2) Ensure source is active + class stamped for scoring
	if class := w.store.SourceClassForHandle(handle); class == "" {
		// Still allow scan if they exist in watched_sources
		if _, err := w.store.GetWatchedSource(handle); err != nil {
			return out, fmt.Errorf("source @%s is not watched — add it first", handle)
		}
	}
	_ = w.store.RefreshSourceClassCache()

	// Known vendor handles (watched sources) — never treat as couple partners.
	knownVendors := map[string]bool{}
	if all, err := w.store.ListWatchedSources(false); err == nil {
		for _, s := range all {
			knownVendors[strings.ToLower(s.Handle)] = true
		}
	}
	knownVendors[handle] = true

	// 3) Fetch posts from provider when budget allows; on failure fall back
	// to already-stored observations so the agent still surfaces couples.
	seenCouples := map[string]bool{}
	liveFetchOK := false
	if w.budgetRemaining(ctx) >= 1 {
		limit := postsLimit
		if rem := w.budgetRemaining(ctx); rem < limit {
			limit = rem
		}
		items, err := w.cfg.Client.FetchAccountPosts(ctx, []string{handle}, limit)
		if err != nil {
			out.Errors = append(out.Errors, "live fetch: "+err.Error()+" — using stored posts")
		} else {
			liveFetchOK = true
			out.PostsFetched = len(items)
			w.recordUsage("vendor:"+handle, len(items))
			for _, item := range items {
				if ctx.Err() != nil {
					break
				}
				raw, imageURL, err := MapPost(item, "vendor:"+handle)
				if err != nil {
					out.Errors = append(out.Errors, "map: "+err.Error())
					continue
				}
				if raw.Handle == "" {
					raw.Handle = handle
				}
				raw.Monitor = "vendor:" + handle
				if class := w.store.SourceClassForHandle(handle); class != "" {
					raw.Payload["source_account_type"] = class
				}
				if imageURL != "" {
					sig := signals.ExtractFromPayload(raw.Payload)
					if sig.CreatesCandidate() {
						if labels, err := w.vision.ClassifyVisualSignals(ctx, imageURL); err == nil && len(labels) > 0 {
							raw.Payload["visual_signals"] = labels
						}
					}
				}
				w.applyScanPost(ctx, handle, raw, imageURL, &out, seenCouples, knownVendors)
			}
			w.advance("vendor:"+handle, time.Now().UTC())
		}
	} else {
		out.Errors = append(out.Errors, "daily budget exhausted — using stored posts only")
	}

	// 4) Always re-read stored posts for this author so tags/couples show even
	// when live fetch is blocked or everything was already ingested.
	if !liveFetchOK || len(out.Couples) == 0 {
		stored, err := w.store.ListSourcePosts(handle, postsLimit)
		if err != nil {
			out.Errors = append(out.Errors, "stored posts: "+err.Error())
		} else {
			for _, p := range stored {
				payload := map[string]any{
					"caption": p.Caption, "url": p.URL, "image_url": p.ImageURL,
					"location": p.Location, "handle": handle,
				}
				if len(p.Tags) > 0 {
					tags := make([]any, len(p.Tags))
					for i, t := range p.Tags {
						tags[i] = t
					}
					payload["tags"] = tags
				}
				if len(p.Mentions) > 0 {
					ms := make([]any, len(p.Mentions))
					for i, m := range p.Mentions {
						ms[i] = m
					}
					payload["provider_mentions"] = ms
				}
				if class := w.store.SourceClassForHandle(handle); class != "" {
					payload["source_account_type"] = class
				}
				// Skip vision on stored-post re-scan (slow); tag/caption heuristics are enough.
				sc := coupleFromRawEventWithVendors(handle, payload, p.ImageURL, knownVendors)
				if sc == nil {
					continue
				}
				out.TaggedPosts++
				key := sc.HandleA + "|" + sc.HandleB
				if seenCouples[key] {
					// Keep higher quality duplicate
					for i := range out.Couples {
						if out.Couples[i].HandleA+"|"+out.Couples[i].HandleB == key && sc.Quality > out.Couples[i].Quality {
							out.Couples[i] = *sc
						}
					}
					continue
				}
				seenCouples[key] = true
				if a, errA := w.store.GetAccountByHandle("instagram", sc.HandleA); errA == nil {
					if b, errB := w.store.GetAccountByHandle("instagram", sc.HandleB); errB == nil {
						if c, err := w.store.GetCoupleForAccountPair(a.ID, b.ID); err == nil {
							sc.CoupleID = c.ID
							if act, err := w.store.LatestPendingActionForCouple(c.ID); err == nil {
								sc.ActionID = act.ID
								sc.ActionType = string(act.ActionType)
								out.PendingApprovals = append(out.PendingApprovals, ScannedApproval{
									ActionID: act.ID, ActionType: string(act.ActionType),
									CoupleID: c.ID, HandleA: sc.HandleA, HandleB: sc.HandleB,
								})
							}
							if src, err := w.store.GetWatchedSource(handle); err == nil && src.City != "" && c.InferredCity == "" {
								lat, lng := cityCoords(src.City, src.State)
								_ = w.store.UpdateCoupleLocation(c.ID, src.City, src.State, "source", lat, lng)
							}
						}
					}
				}
				out.Couples = append(out.Couples, *sc)
			}
		}
	}

	// Rank: real couples first; drop residual vendor_noise from the list shown
	sort.SliceStable(out.Couples, func(i, j int) bool {
		return out.Couples[i].Quality > out.Couples[j].Quality
	})
	filtered := out.Couples[:0]
	for _, c := range out.Couples {
		if c.QualityLabel == "vendor_noise" && c.Quality < 45 {
			continue
		}
		filtered = append(filtered, c)
	}
	out.Couples = filtered

	out.DurationMs = time.Since(start).Milliseconds()
	_ = w.store.TouchSourceScan(handle, len(out.Couples), out.ActionsCreated)
	_, _ = w.store.Audit("watched_source", handle, "source_scan", map[string]any{
		"posts_fetched": out.PostsFetched, "posts_processed": out.PostsProcessed,
		"couples": len(out.Couples), "actions": out.ActionsCreated, "live": liveFetchOK,
	}, "vendor:"+handle, 0)
	return out, nil
}

// applyScanPost processes one mapped post into the pipeline and updates the scan summary.
func (w *Worker) applyScanPost(ctx context.Context, handle string, raw watchtower.RawEvent, imageURL string, out *SourceScanResult, seenCouples map[string]bool, knownVendors map[string]bool) {
	// Vision only when the post already looks engagement-shaped (avoids
	// classifying every chandelier shot and keeps scan under a minute).
	if imageURL != "" && w.vision != nil {
		sig := signals.ExtractFromPayload(raw.Payload)
		if sig.CreatesCandidate() || sig.IsKnownVendor() {
			if labels, err := w.vision.ClassifyVisualSignals(ctx, imageURL); err == nil && len(labels) > 0 {
				raw.Payload["visual_signals"] = labels
			}
		}
	}

	exists, err := w.store.ObservationExists(raw.ExternalEventID)
	if err == nil && exists {
		out.Duplicates++
		if sc := coupleFromRawEventWithVendors(handle, raw.Payload, imageURL, knownVendors); sc != nil {
			key := sc.HandleA + "|" + sc.HandleB
			if !seenCouples[key] {
				seenCouples[key] = true
				if a, errA := w.store.GetAccountByHandle("instagram", sc.HandleA); errA == nil {
					if b, errB := w.store.GetAccountByHandle("instagram", sc.HandleB); errB == nil {
						if c, err := w.store.GetCoupleForAccountPair(a.ID, b.ID); err == nil {
							sc.CoupleID = c.ID
						}
					}
				}
				out.Couples = append(out.Couples, *sc)
			}
		}
		return
	}
	if w.cfg.DryRun {
		out.PostsProcessed++
		return
	}
	result, err := w.orch.ProcessEvent(ctx, raw)
	if err != nil {
		out.Errors = append(out.Errors, fmt.Sprintf("process %s: %v", raw.ExternalEventID, err))
		return
	}
	if result.Duplicate {
		out.Duplicates++
		return
	}
	out.PostsProcessed++

	sc := coupleFromRawEventWithVendors(handle, raw.Payload, imageURL, knownVendors)
	if sc != nil {
		out.TaggedPosts++
		sc.HypothesisID = result.HypothesisID
		sc.Confidence = result.FinalConfidence
		sc.ActionID = result.ActionCreated
		if result.ActionCreated != "" {
			if act, err := w.store.GetAction(result.ActionCreated); err == nil {
				sc.ActionType = string(act.ActionType)
			}
		}
		if a, errA := w.store.GetAccountByHandle("instagram", sc.HandleA); errA == nil {
			if b, errB := w.store.GetAccountByHandle("instagram", sc.HandleB); errB == nil {
				if c, err := w.store.GetCoupleForAccountPair(a.ID, b.ID); err == nil {
					sc.CoupleID = c.ID
					if src, err := w.store.GetWatchedSource(handle); err == nil && src.City != "" && c.InferredCity == "" {
						lat, lng := cityCoords(src.City, src.State)
						_ = w.store.UpdateCoupleLocation(c.ID, src.City, src.State, "source", lat, lng)
					}
				}
			}
		}
		key := sc.HandleA + "|" + sc.HandleB
		if !seenCouples[key] {
			seenCouples[key] = true
			out.Couples = append(out.Couples, *sc)
		}
		if result.ActionCreated != "" {
			out.ActionsCreated++
			out.PendingApprovals = append(out.PendingApprovals, ScannedApproval{
				ActionID: result.ActionCreated, ActionType: sc.ActionType,
				CoupleID: sc.CoupleID, HandleA: sc.HandleA, HandleB: sc.HandleB,
				Confidence: result.FinalConfidence,
			})
		}
		// Only enrich person handles, not every vendor tag
		w.enrichTaggedProfiles(ctx, raw)
	} else if result.ActionCreated != "" {
		out.ActionsCreated++
		out.PendingApprovals = append(out.PendingApprovals, ScannedApproval{
			ActionID: result.ActionCreated, Confidence: result.FinalConfidence,
		})
	}
}

func coupleFromRawEventWithVendors(handle string, payload map[string]any, imageURL string, knownVendors map[string]bool) *ScannedCouple {
	sig := signals.ExtractFromPayload(payload)
	rawTags := append([]string{}, sig.TaggedHandles...)
	if len(rawTags) == 0 {
		rawTags = stringListAny(payload["tags"])
	}
	// Mentions can include co-planners; only use as fallback after person filter
	allRaw := append([]string{}, rawTags...)
	for _, m := range sig.MentionedHandles {
		allRaw = append(allRaw, m)
	}

	// Collect vendor/business tags for UI transparency
	var vendorTags []string
	src := strings.ToLower(strings.TrimPrefix(handle, "@"))
	for _, t := range allRaw {
		t = strings.ToLower(strings.TrimPrefix(strings.TrimSpace(t), "@"))
		if t == "" || t == src {
			continue
		}
		if knownVendors[t] || signals.LooksLikeBusinessHandle(t) {
			vendorTags = append(vendorTags, t)
		}
	}

	a, b, people, ok := signals.PickCouplePair(handle, allRaw, knownVendors)
	if !ok {
		return nil
	}

	cap, _ := payload["caption"].(string)
	url, _ := payload["url"].(string)
	img, _ := payload["image_url"].(string)
	if img == "" {
		img = imageURL
	}

	// Hard-exclude styled shoots / ads / editorials at scan time
	if signals.IsStyledOrAdContent(cap, stringListAny(payload["hashtags"])) {
		return nil
	}

	// Visual people signal from vision labels if present
	visualPeople := false
	if vs := stringListAny(payload["visual_signals"]); len(vs) > 0 {
		for _, v := range vs {
			switch v {
			case "couple_portrait", "people_present", "proposal_scene", "ring":
				visualPeople = true
			case "venue_only", "product_only":
				// strong negative for "this is a couple photo"
			}
		}
		for _, v := range vs {
			if v == "venue_only" || v == "product_only" {
				// only kill if no positive people signal
				if !visualPeople {
					// still allow if caption names a couple
					if _, _, nameOK := signals.ExtractCoupleNamesFromCaption(cap); !nameOK {
						return nil
					}
				}
			}
		}
	}

	q := signals.CoupleQualityScore(cap, a, b, allRaw, img != "", visualPeople)
	label := "weak"
	switch {
	case q >= 75:
		label = "strong_couple"
	case q >= 55:
		label = "likely_couple"
	case q < 40:
		label = "vendor_noise"
	}

	// Drop pure vendor noise from results (quality floor)
	if label == "vendor_noise" && !visualPeople {
		if _, _, nameOK := signals.ExtractCoupleNamesFromCaption(cap); !nameOK {
			return nil
		}
	}

	tagShow := people
	if len(tagShow) > 6 {
		tagShow = tagShow[:6]
	}
	return &ScannedCouple{
		HandleA:       a,
		HandleB:       b,
		Tags:          tagShow,
		VendorTags:    vendorTags[:min(len(vendorTags), 8)],
		PostURL:       url,
		Caption:       truncateRunes(cap, 180),
		ImageURL:      img,
		Quality:       q,
		QualityLabel:  label,
		HasPeopleShot: visualPeople,
	}
}

func stringListAny(v any) []string {
	switch t := v.(type) {
	case []string:
		return t
	case []any:
		var out []string
		for _, x := range t {
			if s, ok := x.(string); ok && s != "" {
				out = append(out, s)
			}
		}
		return out
	default:
		return nil
	}
}

func truncateRunes(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}

