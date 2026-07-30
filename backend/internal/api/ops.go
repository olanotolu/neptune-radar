package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"neptune-social-radar/backend/internal/ontology"
	"neptune-social-radar/backend/internal/signals"
)

// opsSummary is the Today workbench KPI strip.
func (s *Server) opsSummary(w http.ResponseWriter, r *http.Request) {
	sum, err := s.Store.GetOpsSummary()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	out := map[string]any{
		"couples_total":      sum.CouplesTotal,
		"couples_24h":        sum.Couples24h,
		"pending_actions":    sum.PendingActions,
		"needs_pics":         sum.NeedsPics,
		"needs_location":     sum.NeedsLocation,
		"sources_total":      sum.SourcesTotal,
		"sources_with_loc":   sum.SourcesWithLoc,
		"sources_stale":      sum.SourcesStale,
		"map_pins":           sum.MapPins,
		"results_used_today": sum.ResultsUsedToday,
	}
	if s.Watch != nil {
		out["paused"] = s.Watch.IsPaused()
		out["provider_available"] = s.Watch.ProviderAvailable()
		out["running"] = s.Watch.ProviderAvailable() && !s.Watch.IsPaused()
		out["daily_budget"] = s.Watch.DailyBudget()
		out["poll_interval"] = s.Watch.PollInterval().String()
	}
	writeJSON(w, http.StatusOK, out)
}

// suppressCouple marks a false positive so it leaves the board permanently.
func (s *Server) suppressCouple(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var body struct {
		Reason string `json:"reason"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	if err := s.Store.SuppressCouple(id, body.Reason); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	_, _ = s.Store.Audit("couple", id, "suppressed", map[string]any{"reason": body.Reason, "by": "human:concierge"}, "", 0)
	if act, err := s.Store.LatestPendingActionForCouple(id); err == nil && act.ID != "" {
		_ = s.Store.DecideAction(act.ID, ontology.ActionIgnored, "human:suppress")
	}
	writeJSON(w, http.StatusOK, map[string]string{"id": id, "status": "suppressed"})
}

// enrichMissingProfiles pulls Instagram profile pics/bios for couples missing them.
func (s *Server) enrichMissingProfiles(w http.ResponseWriter, r *http.Request) {
	if s.Watch == nil {
		writeError(w, http.StatusServiceUnavailable, errorString("watch loop not configured"))
		return
	}
	limit := 15
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	accts, err := s.Store.ListAccountsNeedingProfile(limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	type row struct {
		Handle string `json:"handle"`
		OK     bool   `json:"ok"`
		Error  string `json:"error,omitempty"`
	}
	var results []row
	okN := 0
	for _, a := range accts {
		if r.Context().Err() != nil {
			break
		}
		err := s.Watch.EnrichAccountProfile(r.Context(), a.Handle)
		rec := row{Handle: a.Handle, OK: err == nil}
		if err != nil {
			rec.Error = err.Error()
		} else {
			okN++
		}
		results = append(results, rec)
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"attempted": len(results),
		"succeeded": okN,
		"results":   results,
	})
}

// backfillLocations infers city from bios for couples missing geo.
func (s *Server) backfillLocations(w http.ResponseWriter, r *http.Request) {
	limit := 100
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			limit = n
		}
	}
	couples, err := s.Store.ListCouplesMissingLocation(limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	coords := map[string][2]float64{
		"columbus|OH": {39.9612, -82.9988}, "cleveland|OH": {41.4993, -81.6944},
		"cincinnati|OH": {39.1031, -84.5120}, "brooklyn|NY": {40.6782, -73.9442},
		"new york|NY": {40.7128, -74.0060}, "manhattan|NY": {40.7831, -73.9712},
		"los angeles|CA": {34.0522, -118.2437}, "chicago|IL": {41.8781, -87.6298},
		"miami|FL": {25.7617, -80.1918}, "austin|TX": {30.2672, -97.7431},
		"dallas|TX": {32.7767, -96.7970}, "houston|TX": {29.7604, -95.3698},
		"seattle|WA": {47.6062, -122.3321}, "boston|MA": {42.3601, -71.0589},
		"philadelphia|PA": {39.9526, -75.1652}, "denver|CO": {39.7392, -104.9903},
		"atlanta|GA": {33.7490, -84.3880}, "dublin|OH": {40.0992, -83.1141},
		"worthington|OH": {40.0931, -83.0180},
	}
	updated := 0
	for _, c := range couples {
		a, _ := s.Store.GetAccountByPersonID(c.PersonAID)
		b, _ := s.Store.GetAccountByPersonID(c.PersonBID)
		loc, ok := signals.BestLocation("", a.BioText, b.BioText, "")
		if !ok {
			if a.InferredCity != "" {
				loc = signals.LocationGuess{City: a.InferredCity, Region: a.InferredRegion, Source: "bio"}
				ok = true
			} else if b.InferredCity != "" {
				loc = signals.LocationGuess{City: b.InferredCity, Region: b.InferredRegion, Source: "bio"}
				ok = true
			}
		}
		if !ok {
			continue
		}
		key := strings.ToLower(loc.City) + "|" + strings.ToUpper(loc.Region)
		var lat, lng *float64
		if xy, hit := coords[key]; hit {
			la, ln := xy[0], xy[1]
			lat, lng = &la, &ln
		} else {
			for k, xy := range coords {
				if strings.HasPrefix(k, strings.ToLower(loc.City)+"|") {
					la, ln := xy[0], xy[1]
					lat, lng = &la, &ln
					break
				}
			}
		}
		if err := s.Store.UpdateCoupleLocation(c.ID, loc.City, loc.Region, loc.Source, lat, lng); err == nil {
			updated++
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"checked": len(couples), "updated": updated})
}
