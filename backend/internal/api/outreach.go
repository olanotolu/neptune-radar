package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"neptune-social-radar/backend/internal/auth"
	"neptune-social-radar/backend/internal/mail"
	"neptune-social-radar/backend/internal/ops"
	"neptune-social-radar/backend/internal/outreach"
	"neptune-social-radar/backend/internal/records"
	"neptune-social-radar/backend/internal/store"
)

func apiFirstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

type queueItem struct {
	KitID      string  `json:"kit_id"`
	CoupleID   string  `json:"couple_id"`
	PersonA    string  `json:"person_a"`
	PersonB    string  `json:"person_b"`
	City       string  `json:"city"`
	Region     string  `json:"region"`
	Status     string  `json:"status"`
	Confidence float64 `json:"confidence"`
	Source     string  `json:"source"`
	Tier       string  `json:"tier"`
	HasStreet  bool    `json:"has_street"`
}

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
	// Celebrate-first journey step
	_ = s.Store.SetJourneyStage(coupleID, "congratulated")
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
	s.Store.Audit("kit", kit.ID, "ready_to_mail",
		map[string]any{"couple_id": kit.CoupleID, "by": "human:" + auth.UserFromContext(r.Context()).Email}, "", 0)
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
	s.Store.Audit("kit", kit.ID, "mailed",
		map[string]any{"couple_id": kit.CoupleID, "by": "human:" + auth.UserFromContext(r.Context()).Email}, "", 0)
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
	s.Store.Audit("kit", kit.ID, "address_verified",
		map[string]any{"couple_id": kit.CoupleID, "verified_by": body.VerifiedBy,
			"by": "human:" + auth.UserFromContext(r.Context()).Email}, "", 0)
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
	s.Store.Audit("kit", kit.ID, "postcard_sent",
		map[string]any{"couple_id": kit.CoupleID, "by": "human:" + auth.UserFromContext(r.Context()).Email}, "", 0)
	writeJSON(w, http.StatusOK, kit)
}

// listGreetingTemplates returns the curated postcard greeting templates.
func (s *Server) listGreetingTemplates(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, outreach.TemplateLibrary())
}

// applyGreetingTemplate applies a curated greeting template to a kit.
func (s *Server) applyGreetingTemplate(w http.ResponseWriter, r *http.Request) {
	if s.Outreach == nil {
		writeError(w, http.StatusServiceUnavailable, errorString("outreach agent not configured"))
		return
	}
	kitID := r.PathValue("id")
	var body struct {
		TemplateID string `json:"template_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	existing, err := s.Store.GetCongratulateKit(kitID)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	tpls := outreach.TemplateLibrary()
	var tpl *outreach.GreetingTemplate
	for i := range tpls {
		if tpls[i].ID == body.TemplateID {
			tpl = &tpls[i]
			break
		}
	}
	if tpl == nil {
		writeError(w, http.StatusBadRequest, errorString("template not found: "+body.TemplateID))
		return
	}
	data := outreach.TemplateData{
		NameA:    existing.FirstNameA,
		NameB:    existing.FirstNameB,
		Location: existing.MarketCity,
	}
	existing.BodyMessage = outreach.RenderTemplate(*tpl, data)
	existing.Headline = "Congratulations"
	existing.InternalNote = existing.InternalNote + fmt.Sprintf("\n\nApplied template: %s (%s)", tpl.Name, tpl.Tone)
	existing.PostcardHTML = outreach.RenderPostcardHTML(existing)
	existing.MailPayload = make(map[string]any) // re-render on next export
	kit, err := s.Store.UpsertCongratulateKit(existing)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, kit)
}

// countyRecordLinks returns Ohio county marriage/property/voter record links for a kit.
func (s *Server) countyRecordLinks(w http.ResponseWriter, r *http.Request) {
	kit, err := s.Store.GetCongratulateKit(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	firstA := kit.FirstNameA
	if firstA == "" {
		firstA = kit.PersonAName
	}
	firstB := kit.FirstNameB
	if firstB == "" {
		firstB = kit.PersonBName
	}
	city := kit.AddressCity
	if city == "" {
		city = kit.MarketCity
	}
	region := kit.AddressRegion
	if region == "" {
		region = kit.MarketRegion
	}
	links := records.CountyRecordLinks(firstA, firstB, city, region)
	countyName := records.CountyName(city, region)
	writeJSON(w, http.StatusOK, map[string]any{
		"county":     countyName,
		"city":       city,
		"region":     region,
		"links":      links,
		"has_county": countyName != "",
	})
}

// batchDetective runs the detective on multiple kits in one call.
func (s *Server) batchDetective(w http.ResponseWriter, r *http.Request) {
	if s.Outreach == nil {
		writeError(w, http.StatusServiceUnavailable, errorString("outreach agent not configured"))
		return
	}
	var body struct {
		KitIDs []string `json:"kit_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if len(body.KitIDs) > 50 {
		writeError(w, http.StatusBadRequest, errorString("max 50 kits per batch"))
		return
	}
	type kitResult struct {
		KitID      string `json:"kit_id"`
		Status     string `json:"status"`
		Candidates int    `json:"candidates"`
		Error      string `json:"error,omitempty"`
	}
	var results []kitResult
	for _, kid := range body.KitIDs {
		kit, err := s.Outreach.RunDetective(r.Context(), kid)
		kr := kitResult{KitID: kid}
		if err != nil {
			kr.Status = "error"
			kr.Error = err.Error()
		} else {
			kr.Status = "ok"
			kr.Candidates = len(kit.AddressCandidates)
		}
		results = append(results, kr)
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"batch_size": len(results),
		"results":    results,
	})
}

// batchVerifyAddresses runs Lob batch verification on multiple kits.
func (s *Server) batchVerifyAddresses(w http.ResponseWriter, r *http.Request) {
	if s.Outreach == nil || s.Outreach.Mail == nil || !s.Outreach.Mail.Available() {
		writeError(w, http.StatusServiceUnavailable, errorString("LOB_API_KEY not configured"))
		return
	}
	var body struct {
		KitIDs []string `json:"kit_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if len(body.KitIDs) > 100 {
		writeError(w, http.StatusBadRequest, errorString("max 100 kits per batch"))
		return
	}

	type verifyResult struct {
		KitID       string `json:"kit_id"`
		Deliverable bool   `json:"deliverable"`
		Line1       string `json:"line1,omitempty"`
		City        string `json:"city,omitempty"`
		Region      string `json:"region,omitempty"`
		Postal      string `json:"postal,omitempty"`
		Error       string `json:"error,omitempty"`
	}

	var addresses []mail.Address
	var kitIDs []string
	for _, kid := range body.KitIDs {
		k, err := s.Store.GetCongratulateKit(kid)
		if err != nil {
			continue
		}
		if k.AddressLine1 == "" || k.AddressCity == "" {
			continue
		}
		addresses = append(addresses, mail.Address{
			Name:           strings.TrimSpace(k.PersonAName + " & " + k.PersonBName),
			AddressLine1:   k.AddressLine1,
			AddressLine2:   k.AddressLine2,
			AddressCity:    k.AddressCity,
			AddressState:   k.AddressRegion,
			AddressZip:     k.AddressPostal,
			AddressCountry: apiFirstNonEmpty(k.AddressCountry, "US"),
		})
		kitIDs = append(kitIDs, kid)
	}

	if len(addresses) == 0 {
		writeJSON(w, http.StatusOK, map[string]any{"batch_size": 0, "results": []verifyResult{}})
		return
	}

	results, err := s.Outreach.Mail.VerifyBatch(r.Context(), addresses)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	var out []verifyResult
	for i, r := range results {
		kr := verifyResult{KitID: kitIDs[i]}
		if i < len(results) {
			kr.Deliverable = r.Deliverable
			kr.Line1 = r.Address.AddressLine1
			kr.City = r.Address.AddressCity
			kr.Region = r.Address.AddressState
			kr.Postal = r.Address.AddressZip
			kr.Error = r.Error
		}
		out = append(out, kr)
	}
	writeJSON(w, http.StatusOK, map[string]any{"batch_size": len(out), "results": out})
}

// operatorQueue returns kits filtered by confidence tier for operator review.
func (s *Server) operatorQueue(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	if status == "" {
		status = "ready_review"
	}
	minConf := 0.0
	if v := r.URL.Query().Get("min_confidence"); v != "" {
		fmt.Sscanf(v, "%f", &minConf)
	}
	kits, err := s.Store.ListCongratulateKits(status, 100)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	var queue []queueItem
	for _, k := range kits {
		// Compute priority on-the-fly for old kits that don't have it stored
		priority := k.PriorityScore
		if priority == 0 && (k.LastNameA != "" || k.LastNameB != "" || k.MarketCity != "") {
			priority = outreach.ComputePriorityScore(k)
		}
		if priority < minConf {
			continue
		}
		qi := queueItem{
			KitID:      k.ID,
			CoupleID:   k.CoupleID,
			PersonA:    k.PersonAName,
			PersonB:    k.PersonBName,
			City:       apiFirstNonEmpty(k.AddressCity, k.MarketCity),
			Region:     apiFirstNonEmpty(k.AddressRegion, k.MarketRegion),
			Status:     k.Status,
			Confidence: priority,
			Source:     k.AddressSource,
			HasStreet:  k.AddressLine1 != "",
		}
		switch {
		case priority >= 0.75:
			qi.Tier = "gold"
		case priority >= 0.50:
			qi.Tier = "silver"
		default:
			qi.Tier = "bronze"
		}
		queue = append(queue, qi)
	}
	if queue == nil {
		queue = []queueItem{}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"queue":  queue,
		"total":  len(queue),
		"gold":   countTier(queue, "gold"),
		"silver": countTier(queue, "silver"),
		"bronze": countTier(queue, "bronze"),
	})
}

func countTier(q []queueItem, tier string) int {
	n := 0
	for _, i := range q {
		if i.Tier == tier {
			n++
		}
	}
	return n
}

// followUpQueue returns kits that are due for a follow-up card.
func (s *Server) followUpQueue(w http.ResponseWriter, r *http.Request) {
	kits, err := s.Store.ListCongratulateKits("mailed", 200)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	type followUpItem struct {
		KitID     string  `json:"kit_id"`
		CoupleID  string  `json:"couple_id"`
		PersonA   string  `json:"person_a"`
		PersonB   string  `json:"person_b"`
		MailedAt  string  `json:"mailed_at"`
		DaysSince int     `json:"days_since_mail"`
		Template  string  `json:"template"`
		Priority  float64 `json:"priority_score"`
	}
	var queue []followUpItem
	now := time.Now().UTC()
	for _, k := range kits {
		if k.FollowUpAt == nil || k.FollowUpAt.After(now) {
			continue
		}
		if k.FollowUpSentAt != nil {
			continue // already sent follow-up
		}
		if k.FollowUpCount >= 2 {
			continue // max 2 follow-ups
		}
		days := int(now.Sub(*k.MailedAt).Hours() / 24)
		queue = append(queue, followUpItem{
			KitID:     k.ID,
			CoupleID:  k.CoupleID,
			PersonA:   k.PersonAName,
			PersonB:   k.PersonBName,
			MailedAt:  k.MailedAt.Format("2006-01-02"),
			DaysSince: days,
			Template:  k.FollowUpTemplate,
			Priority:  k.PriorityScore,
		})
	}
	if queue == nil {
		queue = []followUpItem{}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"queue": queue,
		"total": len(queue),
	})
}

// sendFollowUp sends a follow-up postcard for a kit.
func (s *Server) sendFollowUp(w http.ResponseWriter, r *http.Request) {
	if s.Outreach == nil {
		writeError(w, http.StatusServiceUnavailable, errorString("outreach agent not configured"))
		return
	}
	kitID := r.PathValue("id")
	k, err := s.Store.GetCongratulateKit(kitID)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	if k.Status != "mailed" {
		writeError(w, http.StatusBadRequest, errorString("kit must be mailed before follow-up"))
		return
	}
	if k.FollowUpCount >= 2 {
		writeError(w, http.StatusBadRequest, errorString("max 2 follow-ups per kit"))
		return
	}
	if k.AddressLine1 == "" || k.AddressPostal == "" {
		writeError(w, http.StatusBadRequest, errorString("complete address required for follow-up"))
		return
	}

	// Use the follow-up template (different from first card)
	tpls := outreach.TemplateLibrary()
	tplID := k.FollowUpTemplate
	if tplID == "" {
		tplID = "bright_casual"
	}
	var tpl *outreach.GreetingTemplate
	for i := range tpls {
		if tpls[i].ID == tplID {
			tpl = &tpls[i]
			break
		}
	}
	if tpl == nil {
		tpl = &tpls[0] // fallback to first template
	}

	data := outreach.TemplateData{
		NameA:    k.FirstNameA,
		NameB:    k.FirstNameB,
		Location: k.MarketCity,
	}
	followUpBody := outreach.RenderTemplate(*tpl, data)

	// Send via Lob
	if s.Outreach.Mail != nil && s.Outreach.Mail.Available() {
		front := outreach.RenderPostcardHTML(k)
		back := fmt.Sprintf(`<html><body style="margin:0;font-family:Georgia,serif;padding:24px;font-size:13px;line-height:1.5;color:#1c1917">%s<p style="margin-top:20px;font-size:12px;color:#57534e">Neptune · with care (follow-up)</p></body></html>`,
			strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(followUpBody, "&", "&amp;"), "<", "&lt;"), ">", "&gt;"))
		to := mail.Address{
			Name:           strings.TrimSpace(k.PersonAName + " & " + k.PersonBName),
			AddressLine1:   k.AddressLine1,
			AddressLine2:   k.AddressLine2,
			AddressCity:    k.AddressCity,
			AddressState:   k.AddressRegion,
			AddressZip:     k.AddressPostal,
			AddressCountry: apiFirstNonEmpty(k.AddressCountry, "US"),
		}
		res, err := s.Outreach.Mail.SendPostcard(r.Context(), to, front, back,
			fmt.Sprintf("Neptune follow-up %s & %s (#%d)", k.PersonAName, k.PersonBName, k.FollowUpCount+1))
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		// Record follow-up
		k.FollowUpCount++
		now := time.Now().UTC()
		k.FollowUpSentAt = &now
		k.InternalNote = strings.TrimSpace(k.InternalNote + fmt.Sprintf(
			"\n\nFollow-up #%d sent via Lob: %s", k.FollowUpCount, res.ExternalID))
		k, err = s.Store.UpsertCongratulateKit(k)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"kit_id":          k.ID,
			"follow_up_count": k.FollowUpCount,
			"external_id":     res.ExternalID,
			"status":          "sent",
		})
	} else {
		writeError(w, http.StatusServiceUnavailable, errorString("LOB_API_KEY not configured"))
	}
}

// coupleDossier returns the god-tier operator dossier (evidence, runway, ICP,
// brand-safe journey, handoff, audit). ?lite=1 returns the older slim shape.
func (s *Server) coupleDossier(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if r.URL.Query().Get("lite") == "1" {
		d, err := s.Store.GetCoupleDossier(id)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, d)
		return
	}
	d, err := s.Store.GetGodTierDossier(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, d)
}

// createHandoff issues a tracked Meet Neptune chat deep link for a couple.
func (s *Server) createHandoff(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	code, url, utm, err := s.Store.EnsureHandoff(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	// Soft advance journey when handoff is issued (don't regress later stages).
	var cur string
	_ = s.Store.DB.QueryRow(`SELECT COALESCE(journey_stage,'detected') FROM couples WHERE id = $1`, id).Scan(&cur)
	stage := "invited"
	switch cur {
	case "in_chat", "booked", "closed_won", "closed_lost", "do_not_contact":
		stage = cur
	default:
		_ = s.Store.SetJourneyStage(id, "invited")
	}
	s.Store.Audit("couple", id, "handoff_issued",
		map[string]any{"handoff_code": code, "journey_stage": stage,
			"by": "human:" + auth.UserFromContext(r.Context()).Email}, "", 0)
	writeJSON(w, http.StatusOK, map[string]string{
		"couple_id": id, "handoff_code": code, "handoff_url": url, "handoff_utm": utm,
		"journey_stage": stage,
	})
}

// setJourneyStage updates brand-safe funnel stage (celebrate → invite → chat…).
func (s *Server) setJourneyStage(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var body struct {
		Stage string `json:"stage"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if body.Stage == "" {
		writeError(w, http.StatusBadRequest, errorString("stage required"))
		return
	}
	if err := s.Store.SetJourneyStage(id, body.Stage); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	s.Store.Audit("couple", id, "journey_stage_set",
		map[string]any{"stage": body.Stage, "by": "human:" + auth.UserFromContext(r.Context()).Email}, "", 0)
	writeJSON(w, http.StatusOK, map[string]string{"couple_id": id, "journey_stage": body.Stage})
}

// runJanitor executes maintenance cleanup.
func (s *Server) runJanitor(w http.ResponseWriter, r *http.Request) {
	if !auth.RequireAdmin(w, r) {
		return
	}
	j := &ops.Janitor{Store: s.Store}
	res := j.Run(r.Context())
	writeJSON(w, http.StatusOK, res)
}

// kitStats returns the celebration operations pipeline summary.
func (s *Server) kitStats(w http.ResponseWriter, r *http.Request) {
	stats, err := s.Store.GetKitStats()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, stats)
}
