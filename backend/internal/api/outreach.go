package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"neptune-social-radar/backend/internal/ops"
	"neptune-social-radar/backend/internal/outreach"
	"neptune-social-radar/backend/internal/store"
)

// buildCongratulateKit runs the outreach agent for one couple.
func (s *Server) buildCongratulateKit(w http.ResponseWriter, r *http.Request) {
	if s.Outreach == nil {
		writeError(w, http.StatusServiceUnavailable, errorString("outreach agent not configured"))
		return
	}
	coupleID := r.PathValue("id")
	if coupleID == "" {
		writeError(w, http.StatusBadRequest, errorString("couple id required"))
		return
	}
	kit, err := s.Outreach.BuildKit(r.Context(), coupleID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, kit)
}

func (s *Server) getCongratulateKit(w http.ResponseWriter, r *http.Request) {
	kit, err := s.Store.GetCongratulateKit(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeJSON(w, http.StatusOK, kit)
}

func (s *Server) listCongratulateKits(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	kits, err := s.Store.ListCongratulateKits(status, 50)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if kits == nil {
		kits = []store.CongratulateKit{}
	}
	writeJSON(w, http.StatusOK, kits)
}

func (s *Server) latestKitForCouple(w http.ResponseWriter, r *http.Request) {
	kit, err := s.Store.GetLatestKitForCouple(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeJSON(w, http.StatusOK, kit)
}

// patchCongratulateKit updates address/message/names; optional verify flag.
func (s *Server) patchCongratulateKit(w http.ResponseWriter, r *http.Request) {
	if s.Outreach == nil {
		writeError(w, http.StatusServiceUnavailable, errorString("outreach agent not configured"))
		return
	}
	var body struct {
		AddressLine1   string `json:"address_line1"`
		AddressLine2   string `json:"address_line2"`
		AddressCity    string `json:"address_city"`
		AddressRegion  string `json:"address_region"`
		AddressPostal  string `json:"address_postal"`
		AddressCountry string `json:"address_country"`
		Headline       string `json:"headline"`
		BodyMessage    string `json:"body_message"`
		FirstNameA     string `json:"first_name_a"`
		LastNameA      string `json:"last_name_a"`
		FirstNameB     string `json:"first_name_b"`
		LastNameB      string `json:"last_name_b"`
		Verify         bool   `json:"verify"`
		VerifiedBy     string `json:"verified_by"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	// Load + merge names so detective can use last names next run
	existing, err := s.Store.GetCongratulateKit(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	if body.FirstNameA != "" {
		existing.FirstNameA = strings.TrimSpace(body.FirstNameA)
		existing.PersonAName = existing.FirstNameA
		existing.NameSourceA = "operator"
	}
	if body.LastNameA != "" {
		existing.LastNameA = strings.TrimSpace(body.LastNameA)
		existing.NameSourceA = "operator"
	}
	if body.FirstNameB != "" {
		existing.FirstNameB = strings.TrimSpace(body.FirstNameB)
		existing.PersonBName = existing.FirstNameB
		existing.NameSourceB = "operator"
	}
	if body.LastNameB != "" {
		existing.LastNameB = strings.TrimSpace(body.LastNameB)
		existing.NameSourceB = "operator"
	}
	patch := store.CongratulateKit{
		AddressLine1: body.AddressLine1, AddressLine2: body.AddressLine2,
		AddressCity: body.AddressCity, AddressRegion: body.AddressRegion,
		AddressPostal: body.AddressPostal, AddressCountry: body.AddressCountry,
		Headline: body.Headline, BodyMessage: body.BodyMessage,
		FirstNameA: existing.FirstNameA, LastNameA: existing.LastNameA,
		FirstNameB: existing.FirstNameB, LastNameB: existing.LastNameB,
		NameSourceA: existing.NameSourceA, NameSourceB: existing.NameSourceB,
		PersonAName: existing.PersonAName, PersonBName: existing.PersonBName,
	}
	kit, err := s.Outreach.UpdateKitAddress(r.PathValue("id"), patch, body.Verify, body.VerifiedBy)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	// Persist names even when only names changed
	kit.FirstNameA, kit.LastNameA = existing.FirstNameA, existing.LastNameA
	kit.FirstNameB, kit.LastNameB = existing.FirstNameB, existing.LastNameB
	kit.NameSourceA, kit.NameSourceB = existing.NameSourceA, existing.NameSourceB
	kit.PersonAName, kit.PersonBName = existing.PersonAName, existing.PersonBName
	kit, err = s.Store.UpsertCongratulateKit(kit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, kit)
}

func (s *Server) kitReadyToMail(w http.ResponseWriter, r *http.Request) {
	if s.Outreach == nil {
		writeError(w, http.StatusServiceUnavailable, errorString("outreach agent not configured"))
		return
	}
	kit, err := s.Outreach.MarkReadyToMail(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, kit)
}

func (s *Server) kitMarkMailed(w http.ResponseWriter, r *http.Request) {
	if s.Outreach == nil {
		writeError(w, http.StatusServiceUnavailable, errorString("outreach agent not configured"))
		return
	}
	kit, err := s.Outreach.MarkMailed(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, kit)
}

// kitPostcardHTML serves the print-ready postcard preview.
func (s *Server) kitPostcardHTML(w http.ResponseWriter, r *http.Request) {
	kit, err := s.Store.GetCongratulateKit(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	html := kit.PostcardHTML
	if html == "" {
		html = outreach.RenderPostcardHTML(kit)
	}
	// Proxy IG images through our media endpoint if present in HTML
	// (browser can load /api/media without auth).
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(html))
}

// kitMailExport returns Lob/PostGrid-shaped JSON + CSV-ish fields.
func (s *Server) kitMailExport(w http.ResponseWriter, r *http.Request) {
	kit, err := s.Store.GetCongratulateKit(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"kit_id":       kit.ID,
		"status":       kit.Status,
		"mail_payload": kit.MailPayload,
		"csv_row": map[string]string{
			"name":    strings.TrimSpace(kit.PersonAName + " & " + kit.PersonBName),
			"line1":   kit.AddressLine1,
			"line2":   kit.AddressLine2,
			"city":    kit.AddressCity,
			"state":   kit.AddressRegion,
			"zip":     kit.AddressPostal,
			"country": kit.AddressCountry,
			"message": kit.BodyMessage,
		},
		"print_url": "/api/kits/" + kit.ID + "/postcard",
		"note":      "Human must verify address before mail. Connect Lob/PostGrid with this payload when ready.",
	})
}

// runDetective searches people-data providers for address candidates.
func (s *Server) runDetective(w http.ResponseWriter, r *http.Request) {
	if s.Outreach == nil {
		writeError(w, http.StatusServiceUnavailable, errorString("outreach agent not configured"))
		return
	}
	kit, err := s.Outreach.RunDetective(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, kit)
}

// applyKitCandidate selects one address candidate onto the kit.
func (s *Server) applyKitCandidate(w http.ResponseWriter, r *http.Request) {
	if s.Outreach == nil {
		writeError(w, http.StatusServiceUnavailable, errorString("outreach agent not configured"))
		return
	}
	var body struct {
		Index int `json:"index"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	if v := r.URL.Query().Get("index"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			body.Index = n
		}
	}
	kit, err := s.Outreach.ApplyCandidate(r.PathValue("id"), body.Index)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, kit)
}

// verifyKitAddress USPS-verifies (Lob) and marks address_verified.
func (s *Server) verifyKitAddress(w http.ResponseWriter, r *http.Request) {
	if s.Outreach == nil {
		writeError(w, http.StatusServiceUnavailable, errorString("outreach agent not configured"))
		return
	}
	var body struct {
		VerifiedBy string `json:"verified_by"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	kit, err := s.Outreach.VerifyAndConfirm(r.Context(), r.PathValue("id"), body.VerifiedBy)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, kit)
}

// sendKitPostcard sends physical mail via Lob.
func (s *Server) sendKitPostcard(w http.ResponseWriter, r *http.Request) {
	if s.Outreach == nil {
		writeError(w, http.StatusServiceUnavailable, errorString("outreach agent not configured"))
		return
	}
	kit, err := s.Outreach.SendPostcard(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, kit)
}

// coupleDossier returns structured evidence for agents/UI.
func (s *Server) coupleDossier(w http.ResponseWriter, r *http.Request) {
	d, err := s.Store.GetCoupleDossier(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, d)
}

// runJanitor executes maintenance cleanup.
func (s *Server) runJanitor(w http.ResponseWriter, r *http.Request) {
	j := &ops.Janitor{Store: s.Store}
	res := j.Run(r.Context())
	writeJSON(w, http.StatusOK, res)
}

