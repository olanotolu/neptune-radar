package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"neptune-social-radar/backend/internal/auth"
	"neptune-social-radar/backend/internal/ingest"
	"neptune-social-radar/backend/internal/ontology"
)

type signalResponse struct {
	ID              string `json:"id"`
	ObservationType string `json:"observation_type"`
	Monitor         string `json:"monitor"`
	Handle          string `json:"handle"`
	Summary         string `json:"summary"`
	ObservedAt      string `json:"observed_at"`
	ConsentScope    string `json:"consent_scope"`
}

// listSignals is the live signal feed: newest observations first, optionally
// filtered to one monitor (?monitor=hashtag:justengaged). With no filter it
// streams across every watch source.
func (s *Server) listSignals(w http.ResponseWriter, r *http.Request) {
	monitor := r.URL.Query().Get("monitor")
	limit := 200
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			limit = n
		}
	}
	obs, err := s.Store.ListObservations(monitor, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	out := make([]signalResponse, 0, len(obs))
	for _, o := range obs {
		var payload map[string]any
		_ = json.Unmarshal([]byte(o.RawPayload), &payload)
		handle, _ := payload["handle"].(string)
		out = append(out, signalResponse{
			ID: o.ID, ObservationType: o.ObservationType, Monitor: o.Monitor, Handle: handle,
			Summary:    summarizeObservation(o.ObservationType, payload),
			ObservedAt: o.ObservedAt.Format("2006-01-02T15:04:05Z07:00"), ConsentScope: string(o.ConsentScope),
		})
	}
	writeJSON(w, http.StatusOK, out)
}

func summarizeObservation(obsType string, payload map[string]any) string {
	switch obsType {
	case "post":
		caption, _ := payload["caption"].(string)
		if caption == "" {
			return "posted a photo"
		}
		return "posted: \"" + caption + "\""
	case "bio_change":
		bio, _ := payload["bio"].(string)
		if bio == "" {
			return "cleared their bio"
		}
		return "updated bio to: \"" + bio + "\""
	case "follow_change":
		target, _ := payload["target_handle"].(string)
		active, _ := payload["active"].(bool)
		if active {
			return "followed @" + target
		}
		return "unfollowed @" + target
	case "post_archived":
		return "archived a previous post"
	case "account_disabled":
		disabled, _ := payload["disabled"].(bool)
		if disabled {
			return "account became disabled/unreachable"
		}
		return "account re-enabled"
	case "account_private":
		private, _ := payload["private"].(bool)
		if private {
			return "account switched to private"
		}
		return "account switched to public"
	case "username_change":
		newHandle, _ := payload["new_handle"].(string)
		return "renamed to @" + newHandle
	default:
		return obsType
	}
}

type coupleSummary struct {
	ID           string `json:"id"`
	PersonALabel string `json:"person_a_label"`
	PersonBLabel string `json:"person_b_label"`
}

func (s *Server) listCouples(w http.ResponseWriter, r *http.Request) {
	couples, err := s.Store.ListCouples()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	out := make([]coupleSummary, 0, len(couples))
	for _, c := range couples {
		personA, _ := s.Store.GetPerson(c.PersonAID)
		personB, _ := s.Store.GetPerson(c.PersonBID)
		out = append(out, coupleSummary{ID: c.ID, PersonALabel: personA.DisplayName, PersonBLabel: personB.DisplayName})
	}
	writeJSON(w, http.StatusOK, out)
}

type graphNode struct {
	ID    string `json:"id"`
	Type  string `json:"type"` // "person" or "account"
	Label string `json:"label"`
}
type graphEdge struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Kind   string `json:"kind"`
	Active bool   `json:"active"`
}
type coupleGraphResponse struct {
	Nodes []graphNode `json:"nodes"`
	Edges []graphEdge `json:"edges"`
}

func (s *Server) coupleGraph(w http.ResponseWriter, r *http.Request) {
	coupleID := r.PathValue("id")
	couple, err := s.Store.GetCouple(coupleID)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	personA, _ := s.Store.GetPerson(couple.PersonAID)
	personB, _ := s.Store.GetPerson(couple.PersonBID)
	acctA, errA := s.Store.GetAccountByPersonID(couple.PersonAID)
	acctB, errB := s.Store.GetAccountByPersonID(couple.PersonBID)

	resp := coupleGraphResponse{
		Nodes: []graphNode{
			{ID: personA.ID, Type: "person", Label: personA.DisplayName},
			{ID: personB.ID, Type: "person", Label: personB.DisplayName},
		},
	}
	if errA == nil {
		resp.Nodes = append(resp.Nodes, graphNode{ID: acctA.ID, Type: "account", Label: "@" + acctA.Handle})
		resp.Edges = append(resp.Edges, graphEdge{From: personA.ID, To: acctA.ID, Kind: "owns_account", Active: true})
	}
	if errB == nil {
		resp.Nodes = append(resp.Nodes, graphNode{ID: acctB.ID, Type: "account", Label: "@" + acctB.Handle})
		resp.Edges = append(resp.Edges, graphEdge{From: personB.ID, To: acctB.ID, Kind: "owns_account", Active: true})
	}
	if errA == nil && errB == nil {
		edges, err := s.Store.EdgesForAccount(acctA.ID)
		if err == nil {
			for _, e := range edges {
				if e.FromAccountID == acctA.ID && e.ToAccountID == acctB.ID ||
					e.FromAccountID == acctB.ID && e.ToAccountID == acctA.ID {
					resp.Edges = append(resp.Edges, graphEdge{From: e.FromAccountID, To: e.ToAccountID, Kind: string(e.Kind), Active: e.Active})
				}
			}
		}
	}
	writeJSON(w, http.StatusOK, resp)
}

type relationshipResponse struct {
	Current ontology.Relationship   `json:"current"`
	History []ontology.Relationship `json:"history"`
}

func (s *Server) coupleRelationship(w http.ResponseWriter, r *http.Request) {
	coupleID := r.PathValue("id")
	role := auth.UserFromContext(r.Context()).Role
	current, err := s.Store.CurrentRelationship(coupleID)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	// Gate the current relationship: if it's attorney_only and the caller
	// isn't attorney/admin, return 404 (don't even confirm the couple exists).
	if !auth.ScopeVisible(role, string(current.VisibilityScope)) {
		writeError(w, http.StatusNotFound, errors.New("relationship not found"))
		return
	}
	history, err := s.Store.RelationshipHistory(coupleID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	// Filter history: drop attorney_only rows the caller can't see.
	visible := history[:0]
	for _, h := range history {
		if auth.ScopeVisible(role, string(h.VisibilityScope)) {
			visible = append(visible, h)
		}
	}
	writeJSON(w, http.StatusOK, relationshipResponse{Current: current, History: visible})
}

// pauseCouple flips automation_paused to true on the couple's current
// relationship — Neptune stops acting on new signals for this couple until
// a human resumes. The stage/confidence/scope are preserved unchanged.
func (s *Server) pauseCouple(w http.ResponseWriter, r *http.Request) {
	coupleID := r.PathValue("id")
	rel, err := s.Store.SetAutomationPaused(coupleID, true)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	s.Store.Audit("relationship", rel.ID, "automation_paused",
		map[string]any{"couple_id": coupleID, "by": "human:" + auth.UserFromContext(r.Context()).Email}, "", -1)
	writeJSON(w, http.StatusOK, rel)
}

// resumeCouple flips automation_paused back to false — Neptune resumes
// responding to signals for this couple.
func (s *Server) resumeCouple(w http.ResponseWriter, r *http.Request) {
	coupleID := r.PathValue("id")
	rel, err := s.Store.SetAutomationPaused(coupleID, false)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	s.Store.Audit("relationship", rel.ID, "automation_resumed",
		map[string]any{"couple_id": coupleID, "by": "human:" + auth.UserFromContext(r.Context()).Email}, "", -1)
	writeJSON(w, http.StatusOK, rel)
}

func (s *Server) hypothesisEvidence(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	role := auth.UserFromContext(r.Context()).Role
	// Gate: evidence for an attorney_only hypothesis is invisible to concierge.
	if hyp, err := s.Store.GetHypothesis(id); err == nil {
		if !auth.ScopeVisible(role, string(hyp.VisibilityScope)) {
			writeError(w, http.StatusNotFound, errors.New("hypothesis not found"))
			return
		}
	}
	ev, err := s.Store.EvidenceForHypothesis(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, ev)
}

type confidenceComponent struct {
	Kind        string  `json:"kind"`
	Weight      float64 `json:"weight"`
	Description string  `json:"description"`
}
type confidenceResponse struct {
	Final      float64               `json:"final"`
	Components []confidenceComponent `json:"components"`
}

func (s *Server) hypothesisConfidence(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	role := auth.UserFromContext(r.Context()).Role
	hyp, err := s.Store.GetHypothesis(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	// Gate: an attorney_only hypothesis is invisible to concierge.
	if !auth.ScopeVisible(role, string(hyp.VisibilityScope)) {
		writeError(w, http.StatusNotFound, errors.New("hypothesis not found"))
		return
	}
	ev, err := s.Store.EvidenceForHypothesis(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	resp := confidenceResponse{Final: hyp.Confidence}
	for _, e := range ev {
		resp.Components = append(resp.Components, confidenceComponent{Kind: e.Kind, Weight: e.Weight, Description: e.Description})
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) listCases(w http.ResponseWriter, r *http.Request) {
	cases, err := s.Store.ListCases(r.URL.Query().Get("status"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, cases)
}

func (s *Server) getCase(w http.ResponseWriter, r *http.Request) {
	c, err := s.Store.GetCase(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeJSON(w, http.StatusOK, c)
}

func (s *Server) listLeads(w http.ResponseWriter, r *http.Request) {
	leads, err := s.Store.ListLeads(r.URL.Query().Get("status"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, leads)
}

// listFenrisEvents returns recent Fenris Digital life events for the dashboard.
func (s *Server) listFenrisEvents(w http.ResponseWriter, r *http.Request) {
	limit := 100
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	events, err := s.Store.ListFenrisEvents(limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, events)
}

func (s *Server) listAudit(w http.ResponseWriter, r *http.Request) {
	f := ontologyAuditFilterFromQuery(r)
	events, err := s.Store.ListAudit(f)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, events)
}

// assetProfile returns the financial profile from county property records
// for a couple. Internal operator use only — never exposed on postcards.
func (s *Server) assetProfile(w http.ResponseWriter, r *http.Request) {
	coupleID := r.PathValue("coupleId")
	dossier, err := s.Store.GetGodTierDossier(coupleID)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	if dossier.AssetProfile == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"couple_id": coupleID,
			"estimated_home_value": 0,
			"confidence": 0,
			"source": "",
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"couple_id":           coupleID,
		"estimated_home_value": dossier.AssetProfile.EstimatedHomeValue,
		"property_asset":       dossier.AssetProfile.PropertyAsset,
		"confidence":           dossier.AssetProfile.Confidence,
		"source":               dossier.AssetProfile.Source,
	})
}

// marriageLicenseResponse is one row in the Perfect Timing dashboard view.
type marriageLicenseResponse struct {
	ID                  string  `json:"id"`
	PersonAName         string  `json:"person_a_name"`
	PersonBName         string  `json:"person_b_name"`
	County              string  `json:"county"`
	FilingDate          string  `json:"filing_date"`
	PredictedWeddingDate string  `json:"predicted_wedding_date"`
	WeddingDate         *string `json:"wedding_date,omitempty"`
	DaysUntilWedding    *int    `json:"days_until_wedding,omitempty"`
	Priority            string  `json:"priority"`
}

// listMarriageLicenses returns recent marriage-license filings as couples,
// newest first. Each row carries the predicted wedding date + a priority bucket
// (urgent/priority/early/monitor) that drives outreach timing.
func (s *Server) listMarriageLicenses(w http.ResponseWriter, r *http.Request) {
	limit := 100
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	couples, err := s.Store.ListMarriageLicenseCouples(limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	out := make([]marriageLicenseResponse, 0, len(couples))
	for _, c := range couples {
		pA, _ := s.Store.GetPerson(c.PersonAID)
		pB, _ := s.Store.GetPerson(c.PersonBID)
		row := marriageLicenseResponse{
			ID: c.ID, PersonAName: pA.DisplayName, PersonBName: pB.DisplayName,
			County: c.LicenseCounty, Priority: ingest.PriorityBucket(c),
		}
		if c.LicenseFilingDate != nil {
			row.FilingDate = c.LicenseFilingDate.Format(time.RFC3339)
		}
		ref := c.PredictedWeddingDate
		if c.WeddingDate != nil {
			ref = c.WeddingDate
		}
		if ref != nil {
			row.PredictedWeddingDate = ref.Format(time.RFC3339)
			d := int(time.Until(*ref).Hours() / 24)
			row.DaysUntilWedding = &d
		}
		if c.WeddingDate != nil {
			wd := c.WeddingDate.Format(time.RFC3339)
			row.WeddingDate = &wd
		}
		out = append(out, row)
	}
	writeJSON(w, http.StatusOK, out)
}
