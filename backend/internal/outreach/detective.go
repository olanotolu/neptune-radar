package outreach

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"neptune-social-radar/backend/internal/mail"
	"neptune-social-radar/backend/internal/records"
	"neptune-social-radar/backend/internal/store"
)

// RunDetective calls people-search providers and writes address candidates onto the kit.
func (a *Agent) RunDetective(ctx context.Context, kitID string) (store.CongratulateKit, error) {
	k, err := a.Store.GetCongratulateKit(kitID)
	if err != nil {
		return k, err
	}
	prov := a.Records
	if prov == nil {
		prov = records.NewMulti()
	}

	// Prefer structured kit fields (operator-editable), then person_a_name split
	firstA := firstNonEmpty(k.FirstNameA, splitNameFirst(k.PersonAName))
	lastA := firstNonEmpty(k.LastNameA, splitNameLast(k.PersonAName))
	firstB := firstNonEmpty(k.FirstNameB, splitNameFirst(k.PersonBName))
	lastB := firstNonEmpty(k.LastNameB, splitNameLast(k.PersonBName))

	// --- Enrich query with ALL available location signals ---
	q := records.Query{
		FirstName:    firstA,
		LastName:     lastA,
		PartnerFirst: firstB,
		PartnerLast:  lastB,
		City:         firstNonEmpty(k.AddressCity, k.MarketCity),
		Region:       firstNonEmpty(k.AddressRegion, k.MarketRegion),
		Handle:       k.HandleA,
		BioA:         k.BioA,
		BioB:         k.BioB,
	}

	// Load couple + account data for location signal enrichment
	if k.CoupleID != "" {
		if couple, err := a.Store.GetCouple(k.CoupleID); err == nil {
			// Couple-level inferred city (from post geotags + both bios)
			if q.City == "" && couple.InferredCity != "" {
				q.City = couple.InferredCity
				q.Region = couple.InferredRegion
			}
			// Load each partner's individually-inferred city from their Instagram bio
			if acctA, err := a.Store.GetAccountByPersonID(couple.PersonAID); err == nil {
				q.AccountCityA = acctA.InferredCity
				q.AccountRegionA = acctA.InferredRegion
				if q.BioA == "" {
					q.BioA = acctA.BioText
				}
			}
			if acctB, err := a.Store.GetAccountByPersonID(couple.PersonBID); err == nil {
				q.AccountCityB = acctB.InferredCity
				q.AccountRegionB = acctB.InferredRegion
				if q.BioB == "" {
					q.BioB = acctB.BioText
				}
			}
		}
	}

	// Vendor/photographer city from watched_sources registry
	if k.SourceHandle != "" {
		if src, err := a.Store.GetWatchedSource(k.SourceHandle); err == nil {
			q.VendorCity = src.City
			q.VendorState = src.State
		}
	}

	// Post venue location from discovery post's raw payload
	if k.DiscoveryPostURL != "" {
		if postLoc, _ := a.Store.FindDiscoveryPostLocation(k.HandleA, k.HandleB); postLoc != "" {
			q.PostLocation = postLoc
		}
	}

	// Stronger note when last names missing
	if lastA == "" && lastB == "" {
		k.ResearchNotes = strings.TrimSpace(k.ResearchNotes + "\n\n⚠ Detective: no last names — street hits unlikely. Fill Last name A/B and re-run.")
	}
	res, err := prov.Search(ctx, q)
	// Also try partner as primary if first search empty and partner named
	if (err != nil || len(res.Candidates) == 0) && firstB != "" {
		q2 := q
		q2.FirstName, q2.LastName = firstB, lastB
		q2.PartnerFirst, q2.PartnerLast = firstA, lastA
		q2.Handle = k.HandleB
		if res2, err2 := prov.Search(ctx, q2); err2 == nil && len(res2.Candidates) > 0 {
			res = res2
			err = nil
		}
	}

	// Bio-to-address regex: extract street addresses from Instagram bios (free, no API)
	for _, bio := range []string{k.BioA, k.BioB} {
		if addr := records.ExtractAddressFromBio(bio); addr != nil {
			res.Candidates = append([]records.Candidate{{
				Line1:      addr.Line1,
				Line2:      addr.Line2,
				City:       firstNonEmpty(addr.City, q.City),
				Region:     firstNonEmpty(addr.Region, q.Region),
				Postal:     addr.Postal,
				Country:    "US",
				Confidence: addr.Confidence,
				Source:     "bio_regex",
				Note:       fmt.Sprintf("Street address parsed from Instagram bio (%s). Verify before mail.", addr.Source),
			}}, res.Candidates...)
			res.Provider = res.Provider + "+bio_regex"
		}
	}

	// Persist audit row
	qJSON, _ := json.Marshal(q)
	cJSON, _ := json.Marshal(res.Candidates)
	st := res.Status
	if st == "" {
		st = "ok"
	}
	if err != nil && st == "ok" {
		st = "error"
	}
	_, _ = a.Store.InsertAddressLookup(store.AddressLookup{
		KitID: k.ID, CoupleID: k.CoupleID, Provider: res.Provider,
		QueryJSON: string(qJSON), ResponseJSON: res.RawJSON,
		CandidatesJSON: string(cJSON), Status: st,
		ErrorMessage: res.Error, CostCents: res.CostCents,
	})
	if err != nil && len(res.Candidates) == 0 {
		return k, fmt.Errorf("detective search: %w", err)
	}

	// Map to store candidates
	var cands []store.AddressCandidate
	bestConf := 0.0
	for _, c := range res.Candidates {
		cands = append(cands, store.AddressCandidate{
			Line1: c.Line1, Line2: c.Line2, City: c.City, Region: c.Region,
			Postal: c.Postal, Country: firstNonEmpty(c.Country, "US"),
			Confidence: c.Confidence, Source: c.Source, Note: c.Note,
		})
		if c.Confidence > bestConf {
			bestConf = c.Confidence
		}
	}

	// Lob verification post-step: verify each candidate with a street address
	if a.Mail != nil && a.Mail.Available() && len(cands) > 0 {
		for i := range cands {
			c := &cands[i]
			if c.Line1 == "" || c.City == "" {
				continue
			}
			vr, err := a.Mail.VerifyAddress(ctx, mail.Address{
				Name:           strings.TrimSpace(k.PersonAName + " & " + k.PersonBName),
				AddressLine1:   c.Line1,
				AddressLine2:   c.Line2,
				AddressCity:    c.City,
				AddressState:   c.Region,
				AddressZip:     c.Postal,
				AddressCountry: firstNonEmpty(c.Country, "US"),
			})
			if err != nil {
				continue // skip verification errors silently
			}
			// Apply verified components
			if vr.Address.AddressLine1 != "" {
				c.Line1 = vr.Address.AddressLine1
			}
			if vr.Address.AddressCity != "" {
				c.City = vr.Address.AddressCity
			}
			if vr.Address.AddressState != "" {
				c.Region = vr.Address.AddressState
			}
			if vr.Address.AddressZip != "" {
				c.Postal = vr.Address.AddressZip
			}
			if vr.Deliverable {
				c.Confidence = 0.90
				c.Note = "Lob verified deliverable — safe to mail."
			} else {
				c.Confidence = 0.65
				c.Note = "Lob verified but deliverability uncertain — review."
			}
			c.Source = "lob_verified_" + c.Source
		}
	}

	k.AddressCandidates = cands
	if bestConf > k.AddressConfidence {
		k.AddressConfidence = bestConf
	}
	k.AddressSource = res.Provider
	// Prefill city from top candidate without claiming street until user picks
	if len(cands) > 0 {
		top := cands[0]
		if k.AddressCity == "" {
			k.AddressCity = top.City
		}
		if k.AddressRegion == "" {
			k.AddressRegion = top.Region
		}
		// Only auto-fill street when confidence is strong AND line present
		// Still require human verify before mail.
		if top.Line1 != "" && top.Confidence >= 0.5 && k.AddressLine1 == "" {
			// leave street for user pick in UI — do not auto-write line1
		}
	}

	// Research notes append — include all location signals + county records for operator context
	var signalSummary []string
	if q.VendorCity != "" {
		signalSummary = append(signalSummary, fmt.Sprintf("vendor=%s,%s", q.VendorCity, q.VendorState))
	}
	if q.AccountCityA != "" {
		signalSummary = append(signalSummary, fmt.Sprintf("acctA=%s,%s", q.AccountCityA, q.AccountRegionA))
	}
	if q.AccountCityB != "" {
		signalSummary = append(signalSummary, fmt.Sprintf("acctB=%s,%s", q.AccountCityB, q.AccountRegionB))
	}
	if q.PostLocation != "" {
		signalSummary = append(signalSummary, fmt.Sprintf("postLoc=%s", q.PostLocation))
	}
	sigStr := ""
	if len(signalSummary) > 0 {
		sigStr = "\nSignals: " + strings.Join(signalSummary, " · ")
	}

	// County record links (Ohio only)
	countyStr := ""
	searchCity := firstNonEmpty(q.City, q.Region)
	if searchCity != "" {
		countyName := records.CountyName(searchCity, q.Region)
		if countyName != "" {
			countyURL := records.CountySearchURL(searchCity, q.Region)
			countyStr = fmt.Sprintf("\nCounty: %s County, Ohio — %s", countyName, countyURL)
		}
	}

	k.ResearchNotes = strings.TrimSpace(k.ResearchNotes + "\n\n--- Detective run " + time.Now().UTC().Format(time.RFC3339) + " ---\n" +
		fmt.Sprintf("Provider: %s · status: %s · candidates: %d\n", res.Provider, st, len(cands)) +
		fmt.Sprintf("Query: %s %s · %s, %s\n", firstA, lastA, q.City, q.Region) +
		sigStr + countyStr + "\n" +
		"Pick a candidate in the UI, then Verify address. Configure TRESTLE_API_KEY or PDL_API_KEY for street hits.")

	// Mark detective step done in research steps
	steps := k.ResearchSteps
	found := false
	for i := range steps {
		if steps[i].ID == "people" || steps[i].ID == "detective" {
			steps[i].Status = "done"
			steps[i].Detail = fmt.Sprintf("Ran %s → %d candidates", res.Provider, len(cands))
			found = true
		}
	}
	if !found {
		steps = append([]store.ResearchStep{{
			ID: "detective", Label: "Detective run",
			Detail: fmt.Sprintf("%s returned %d candidates", res.Provider, len(cands)),
			Status: "done",
		}}, steps...)
	}
	k.ResearchSteps = steps
	k.PostcardHTML = RenderPostcardHTML(k)
	k.MailPayload = mailPayload(k)
	if k.Status == "draft" {
		k.Status = "ready_review"
	}
	return a.Store.UpsertCongratulateKit(k)
}

// ApplyCandidate copies one ranked candidate onto the kit mailing fields (not yet verified).
func (a *Agent) ApplyCandidate(kitID string, idx int) (store.CongratulateKit, error) {
	k, err := a.Store.GetCongratulateKit(kitID)
	if err != nil {
		return k, err
	}
	if idx < 0 || idx >= len(k.AddressCandidates) {
		return k, fmt.Errorf("candidate index %d out of range", idx)
	}
	c := k.AddressCandidates[idx]
	k.AddressLine1 = c.Line1
	k.AddressLine2 = c.Line2
	k.AddressCity = c.City
	k.AddressRegion = c.Region
	k.AddressPostal = c.Postal
	k.AddressCountry = firstNonEmpty(c.Country, "US")
	k.AddressConfidence = c.Confidence
	k.AddressSource = c.Source
	k.PostcardHTML = RenderPostcardHTML(k)
	k.MailPayload = mailPayload(k)
	return a.Store.UpsertCongratulateKit(k)
}

// VerifyAndConfirm runs Lob USPS verify when available, then marks address_verified.
func (a *Agent) VerifyAndConfirm(ctx context.Context, kitID, verifiedBy string) (store.CongratulateKit, error) {
	k, err := a.Store.GetCongratulateKit(kitID)
	if err != nil {
		return k, err
	}
	if k.AddressLine1 == "" || k.AddressCity == "" {
		return k, fmt.Errorf("street and city required — pick a candidate or type an address first")
	}
	if k.AddressPostal == "" && a.Mail != nil && a.Mail.Available() {
		// allow verify without zip — Lob may fill it
	} else if k.AddressPostal == "" {
		return k, fmt.Errorf("postal code required to verify")
	}

	if a.Mail != nil && a.Mail.Available() {
		vr, err := a.Mail.VerifyAddress(ctx, mail.Address{
			Name:           strings.TrimSpace(k.PersonAName + " & " + k.PersonBName),
			AddressLine1:   k.AddressLine1,
			AddressLine2:   k.AddressLine2,
			AddressCity:    k.AddressCity,
			AddressState:   k.AddressRegion,
			AddressZip:     k.AddressPostal,
			AddressCountry: firstNonEmpty(k.AddressCountry, "US"),
		})
		if err != nil {
			return k, fmt.Errorf("address verify: %w", err)
		}
		// Apply standardized components
		k.AddressLine1 = firstNonEmpty(vr.Address.AddressLine1, k.AddressLine1)
		k.AddressLine2 = vr.Address.AddressLine2
		k.AddressCity = firstNonEmpty(vr.Address.AddressCity, k.AddressCity)
		k.AddressRegion = firstNonEmpty(vr.Address.AddressState, k.AddressRegion)
		k.AddressPostal = firstNonEmpty(vr.Address.AddressZip, k.AddressPostal)
		k.AddressSource = "lob_verified"
		if vr.Deliverable {
			k.AddressConfidence = 0.95
		} else {
			k.AddressConfidence = 0.7
			k.ResearchNotes += "\nLob deliverability uncertain — review carefully."
		}
	}

	return a.UpdateKitAddress(kitID, k, true, verifiedBy)
}

// SendPostcard mails via Lob when configured. Requires address_verified or ready_to_mail.
func (a *Agent) SendPostcard(ctx context.Context, kitID string) (store.CongratulateKit, error) {
	k, err := a.Store.GetCongratulateKit(kitID)
	if err != nil {
		return k, err
	}
	if k.Status != "address_verified" && k.Status != "ready_to_mail" {
		return k, fmt.Errorf("verify address before sending (status=%s)", k.Status)
	}
	if k.AddressLine1 == "" || k.AddressPostal == "" {
		return k, fmt.Errorf("complete mailing address required")
	}
	if a.Mail == nil || !a.Mail.Available() {
		return k, fmt.Errorf("LOB_API_KEY not configured — set key + LOB_FROM_* return address")
	}

	// Ensure ready_to_mail
	if k.Status == "address_verified" {
		k, err = a.MarkReadyToMail(kitID)
		if err != nil {
			return k, err
		}
	}

	front := postcardFrontHTML(k)
	back := postcardBackHTML(k)
	to := mail.Address{
		Name:           strings.TrimSpace(k.PersonAName + " & " + k.PersonBName),
		AddressLine1:   k.AddressLine1,
		AddressLine2:   k.AddressLine2,
		AddressCity:    k.AddressCity,
		AddressState:   k.AddressRegion,
		AddressZip:     k.AddressPostal,
		AddressCountry: firstNonEmpty(k.AddressCountry, "US"),
	}
	res, err := a.Mail.SendPostcard(ctx, to, front, back, fmt.Sprintf("Neptune congratulate %s & %s", k.PersonAName, k.PersonBName))
	toJSON, _ := json.Marshal(to)
	ms := store.MailSend{
		KitID: k.ID, CoupleID: k.CoupleID, Provider: "lob",
		ExternalID: res.ExternalID, Status: firstNonEmpty(res.Status, "created"),
		ToAddressJSON: string(toJSON), RawResponse: res.RawJSON,
		CostCents: res.CostCents, ExpectedDeliveryDate: res.ExpectedDeliveryDate,
	}
	if err != nil {
		ms.Status = "error"
		ms.ErrorMessage = err.Error()
		_, _ = a.Store.InsertMailSend(ms)
		return k, err
	}
	_, _ = a.Store.InsertMailSend(ms)

	// Mark mailed
	k, err = a.MarkMailed(kitID)
	if err != nil {
		return k, err
	}
	k.InternalNote = strings.TrimSpace(k.InternalNote + "\n\nLob postcard id: " + res.ExternalID +
		" expected_delivery: " + res.ExpectedDeliveryDate)
	// Best-effort store external id if column exists
	_, _ = a.Store.DB.Exec(`UPDATE congratulate_kits SET mail_external_id = $2, mail_provider = 'lob', updated_at = now() WHERE id = $1`,
		k.ID, res.ExternalID)
	return a.Store.UpsertCongratulateKit(k)
}

func splitName(full string) (first, last string) {
	full = strings.TrimSpace(full)
	if full == "" {
		return "", ""
	}
	parts := strings.Fields(full)
	if len(parts) == 1 {
		return firstNameOnly(parts[0]), ""
	}
	return firstNameOnly(parts[0]), parts[len(parts)-1]
}

func splitNameFirst(full string) string {
	f, _ := splitName(full)
	return f
}

func splitNameLast(full string) string {
	_, l := splitName(full)
	return l
}

func postcardFrontHTML(k store.CongratulateKit) string {
	// Lob accepts HTML for front; keep simple and print-safe.
	nameA := htmlEscape(firstNameOnly(k.PersonAName))
	nameB := htmlEscape(firstNameOnly(k.PersonBName))
	return fmt.Sprintf(`<html><body style="margin:0;font-family:Georgia,serif;background:#1a3a3c;color:#faf7f2;text-align:center;padding:40px 20px">
<h1 style="font-weight:500;letter-spacing:0.04em">%s</h1>
<p style="font-style:italic;font-size:18px">%s &amp; %s</p>
<p style="font-size:12px;opacity:0.7;text-transform:uppercase;letter-spacing:0.08em">%s</p>
</body></html>`, htmlEscape(firstNonEmpty(k.Headline, "Congratulations")), nameA, nameB,
		htmlEscape(strings.Trim(k.MarketCity+", "+k.MarketRegion, ", ")))
}

func postcardBackHTML(k store.CongratulateKit) string {
	body := htmlEscape(k.BodyMessage)
	body = strings.ReplaceAll(body, "\n", "<br/>")
	return fmt.Sprintf(`<html><body style="margin:0;font-family:Georgia,serif;padding:24px;font-size:13px;line-height:1.5;color:#1c1917">
%s
<p style="margin-top:20px;font-size:12px;color:#57534e">Neptune · with care</p>
</body></html>`, body)
}

func htmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	return s
}
