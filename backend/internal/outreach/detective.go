package outreach

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"neptune-social-radar/backend/internal/llm"
	"neptune-social-radar/backend/internal/mail"
	"neptune-social-radar/backend/internal/records"
	"neptune-social-radar/backend/internal/store"
)

// RunDetective calls people-search providers and writes address candidates onto the kit.
// Strategy:
//  0. Prep agent gate (identity + home market) — refuse if score < MinRunThreshold
//  1. Free text extract (bios + discovery caption + recent posts) — no API cost
//  2. Paid fan-out: name variants × location variants under DETECTIVE_PAID_CAP
//  3. Free multi-loc pass if still no confirmable street
//  4. Merge, rank, Lob-verify streets, persist
func (a *Agent) RunDetective(ctx context.Context, kitID string) (store.CongratulateKit, error) {
	k, err := a.Store.GetCongratulateKit(kitID)
	if err != nil {
		return k, err
	}

	// Prep agent — recompute every run so name/city edits count
	prep := RunPrep(k)
	ApplyPrepToKit(&k, prep)
	if prep.Score < MinRunThreshold {
		// Persist prep feedback so UI can show blockers
		saved, uerr := a.Store.UpsertCongratulateKit(k)
		if uerr != nil {
			return k, uerr
		}
		return saved, fmt.Errorf("prep blocked detective (%.0f%%): %s — fix: %s",
			prep.Score*100, prep.Summary, strings.Join(prep.Blockers, ", "))
	}

	// Prefer structured kit fields (operator-editable), then person_a_name split
	firstA := firstNonEmpty(k.FirstNameA, splitNameFirst(k.PersonAName))
	lastA := firstNonEmpty(k.LastNameA, splitNameLast(k.PersonAName))
	firstB := firstNonEmpty(k.FirstNameB, splitNameFirst(k.PersonBName))
	lastB := firstNonEmpty(k.LastNameB, splitNameLast(k.PersonBName))

	// Prefer prep home market over garbage market_city
	city := firstNonEmpty(k.AddressCity, prep.HomeCity, k.MarketCity)
	region := firstNonEmpty(k.AddressRegion, prep.HomeRegion, k.MarketRegion)
	if isGarbageCity(city) {
		city = prep.HomeCity
		region = prep.HomeRegion
	}

	// --- Enrich query with ALL available location signals ---
	q := records.Query{
		FirstName:    firstA,
		LastName:     lastA,
		PartnerFirst: firstB,
		PartnerLast:  lastB,
		City:         city,
		Region:       region,
		Handle:       k.HandleA,
		BioA:         k.BioA,
		BioB:         k.BioB,
	}

	if k.CoupleID != "" {
		if couple, err := a.Store.GetCouple(k.CoupleID); err == nil {
			if q.City == "" && couple.InferredCity != "" {
				q.City = couple.InferredCity
				q.Region = couple.InferredRegion
			}
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

	if k.SourceHandle != "" {
		if src, err := a.Store.GetWatchedSource(k.SourceHandle); err == nil {
			q.VendorCity = src.City
			q.VendorState = src.State
		}
	}

	if k.DiscoveryPostURL != "" || k.HandleA != "" || k.HandleB != "" {
		if postLoc, _ := a.Store.FindDiscoveryPostLocation(k.HandleA, k.HandleB); postLoc != "" {
			q.PostLocation = postLoc
		}
	}

	if lastA == "" && lastB == "" {
		k.ResearchNotes = strings.TrimSpace(k.ResearchNotes + "\n\n⚠ Detective: no last names — street hits unlikely. Fill Last name A/B and re-run.")
	}

	// --- Extra location evidence: multi-post geotags ---
	if locsExtra := a.Store.ListCouplePostLocations(k.HandleA, k.HandleB, 20); len(locsExtra) > 0 {
		for _, locName := range locsExtra {
			if q.PostLocation == "" {
				q.PostLocation = locName
			}
			if city, region, ok := records.ExtractLocationFromText(locName); ok {
				if q.AccountCityA == "" {
					q.AccountCityA, q.AccountRegionA = city, region
				} else if q.AccountCityB == "" && !strings.EqualFold(city, q.AccountCityA) {
					q.AccountCityB, q.AccountRegionB = city, region
				}
			}
		}
	}

	// --- 0b. AI home market (Baseten) when city weak / vendor-only / missing ---
	// Never invents streets — city+state only to aim TruePeopleSearch.
	needAICity := q.City == "" || isGarbageCity(q.City) || prep.HomeSource == "vendor_only" || prep.HomeSource == "none" || prep.HomeSource == ""
	if needAICity || prep.HomeSource == "bio_a" || prep.HomeSource == "bio_b" {
		// Always try AI when we have Baseten — can refine A&M-style hints too
		guess, gerr := llm.InferHomeMarket(ctx, llm.HomeMarketInput{
			PersonA: strings.TrimSpace(firstA + " " + lastA), PersonB: strings.TrimSpace(firstB + " " + lastB),
			HandleA: k.HandleA, HandleB: k.HandleB,
			BioA: q.BioA, BioB: q.BioB, Caption: k.DiscoveryCaption,
			VendorCity: q.VendorCity, VendorState: q.VendorState,
			PostLocation: q.PostLocation,
			MarketHint:   strings.Trim(prep.HomeCity+", "+prep.HomeRegion, ", "),
			EvidenceLines: k.Evidence,
		})
		if gerr == nil && guess.City != "" && !isGarbageCity(guess.City) {
			k.ResearchNotes = strings.TrimSpace(k.ResearchNotes + fmt.Sprintf(
				"\n\n--- AI home market ---\n%s, %s (conf %.0f%%) · %s\n%s",
				guess.City, guess.Region, guess.Confidence*100, guess.Source, guess.Reason))
			// Prefer AI when higher conf than prep-only vendor, or city was empty
			if q.City == "" || isGarbageCity(q.City) || guess.Confidence >= 0.55 {
				q.City, q.Region = guess.City, firstNonEmpty(guess.Region, q.Region)
				if k.AddressCity == "" || isGarbageCity(k.AddressCity) {
					k.AddressCity, k.AddressRegion = q.City, q.Region
				}
				if k.MarketCity == "" || isGarbageCity(k.MarketCity) {
					k.MarketCity, k.MarketRegion = q.City, q.Region
					k.MarketSource = "ai_home_market"
				}
			}
		} else if gerr != nil {
			k.ResearchNotes = strings.TrimSpace(k.ResearchNotes + "\nAI home market skipped: "+gerr.Error())
		}
	}

	// --- 1. Free text extract (before any paid spend) ---
	texts := []records.TextSource{
		{Text: firstNonEmpty(q.BioA, k.BioA), Source: "bio_a"},
		{Text: firstNonEmpty(q.BioB, k.BioB), Source: "bio_b"},
		{Text: k.DiscoveryCaption, Source: "discovery_caption"},
	}
	if capBlob, _ := a.Store.ListRecentCaptionsForCouple(k.HandleA, k.HandleB, 15); capBlob != "" {
		texts = append(texts, records.TextSource{Text: capBlob, Source: "recent_caption"})
	}
	extracted := records.ExtractStreetsFromTexts(texts, q.City, q.Region)
	// Business profile streets
	if k.CoupleID != "" {
		if couple, err := a.Store.GetCouple(k.CoupleID); err == nil {
			for i, pid := range []string{couple.PersonAID, couple.PersonBID} {
				if pid == "" {
					continue
				}
				acct, err := a.Store.GetAccountByPersonID(pid)
				if err != nil {
					continue
				}
				st, city, state, postal, ok := a.Store.GetAccountBusinessAddress(acct.ID)
				if !ok || !records.IsRealStreet(st) {
					continue
				}
				src := "ig_business_a"
				note := "Instagram business profile street — may be workplace/studio, not home. Confirm."
				if i == 1 {
					src = "ig_business_b"
				}
				extracted = append(extracted, records.Candidate{
					Line1: st, City: firstNonEmpty(city, q.City), Region: firstNonEmpty(state, q.Region),
					Postal: postal, Country: "US", Kind: records.KindStreet,
					Confidence: 0.55, Source: src, Note: note,
				})
			}
		}
	}
	var parts []records.Result
	var searchErr error
	if len(extracted) > 0 {
		parts = append(parts, records.Result{
			Provider: "text_extract", Status: "ok", Candidates: extracted,
		})
	}

	// --- 1b. MULTI-HUNTER swarm: person A then B, every viable route ---
	// Order: Apify TPS (works with APIFY_TOKEN alone) → Bright Data unlocker TPS →
	// later Multi free (Google/property/voter). First+last required per person.
	type personHit struct {
		first, last, handle, label string
	}
	var people []personHit
	if lastA != "" && firstA != "" {
		people = append(people, personHit{firstA, lastA, k.HandleA, "person_a"})
	}
	if lastB != "" && firstB != "" {
		people = append(people, personHit{firstB, lastB, k.HandleB, "person_b"})
	}
	if lastA != "" && firstB != "" && lastB == "" {
		people = append(people, personHit{firstB, lastA, k.HandleB, "person_b_as_last_a"})
	}
	if lastB != "" && firstA != "" && lastA == "" {
		people = append(people, personHit{firstA, lastB, k.HandleA, "person_a_as_last_b"})
	}

	tpsCities := []records.LocVariant{}
	if q.City != "" && !isGarbageCity(q.City) {
		tpsCities = append(tpsCities, records.LocVariant{City: q.City, Region: q.Region, Source: "home"})
	}
	for _, lv := range records.LocationVariants(q) {
		if isGarbageCity(lv.City) {
			continue
		}
		dup := false
		for _, e := range tpsCities {
			if strings.EqualFold(e.City, lv.City) {
				dup = true
				break
			}
		}
		if !dup {
			tpsCities = append(tpsCities, lv)
		}
	}
	if len(tpsCities) == 0 {
		tpsCities = []records.LocVariant{{City: "", Region: q.Region, Source: "none"}}
	}
	if len(tpsCities) > 2 {
		tpsCities = tpsCities[:2]
	}

	// Build hunter list — every available route (first-principles: try ALL that can fire)
	var hunters []records.Provider
	if ap := records.NewApifyTPSFromEnv(); ap != nil && ap.Available() {
		hunters = append(hunters, ap)
	}
	if tps := records.NewTruePeopleSearchFromEnv(); tps.Available() {
		hunters = append(hunters, tps)
	}
	// Free SERP always — research links + snippet streets for each person×city
	hunters = append(hunters, &records.DDGSerp{})

	k.ResearchNotes = strings.TrimSpace(k.ResearchNotes + "\n\n--- Multi-hunter swarm ---\n" +
		fmt.Sprintf("Home aim: %s, %s · people=%d · cities=%d · hunters=%d\n", q.City, q.Region, len(people), len(tpsCities), len(hunters)) +
		"Apify: "+records.ApifyTPSStatus()+"\nBrightData TPS: "+records.TPSStatus())

	if len(people) > 0 && len(hunters) > 0 {
		for _, person := range people {
			gotStreet := false
			for _, loc := range tpsCities {
				if gotStreet {
					break
				}
				for _, hunter := range hunters {
					if err := ctx.Err(); err != nil {
						searchErr = err
						break
					}
					// If we already have confirmable streets from a stronger hunter, skip weaker
					if records.HasConfirmableStreet(records.MergeResults(parts...).Candidates) {
						gotStreet = true
						// Still try partner person for multi-source — break cities only
						break
					}
					qv := q
					qv.FirstName, qv.LastName, qv.Handle = person.first, person.last, person.handle
					qv.City, qv.Region = loc.City, loc.Region
					if person.label == "person_a" || person.label == "person_a_as_last_b" {
						qv.PartnerFirst, qv.PartnerLast = firstB, lastB
					} else {
						qv.PartnerFirst, qv.PartnerLast = firstA, lastA
					}
					rv, e := hunter.Search(ctx, qv)
					if e != nil {
						searchErr = e
					}
					tag := fmt.Sprintf("%s %s as %s %s @ %s", hunter.Name(), person.label, person.first, person.last, loc.City)
					for i := range rv.Candidates {
						if rv.Candidates[i].Note == "" {
							rv.Candidates[i].Note = tag
						} else {
							rv.Candidates[i].Note = tag + " · " + rv.Candidates[i].Note
						}
						rv.Candidates[i].Source = hunter.Name() + "+" + person.label
					}
					streetN := 0
					for _, c := range rv.Candidates {
						if records.IsRealStreet(c.Line1) {
							streetN++
						}
					}
					k.ResearchNotes = strings.TrimSpace(k.ResearchNotes + fmt.Sprintf(
						"\n• %s → status=%s streets=%d err=%s", tag, rv.Status, streetN, rv.Error))
					parts = append(parts, rv)
					if streetN > 0 {
						gotStreet = true
						// Don't burn more hunters for this person at this city
						break
					}
				}
			}
		}
	} else {
		k.ResearchNotes = strings.TrimSpace(k.ResearchNotes + "\nNo hunters available or missing last names.")
	}

	// --- 2–3. Name × location fan-out with paid budget ---
	names := records.NameVariants(firstA, lastA, k.HandleA, firstB, lastB, k.HandleB)
	locs := records.LocationVariants(q)
	if len(locs) == 0 {
		locs = []records.LocVariant{{City: "", Region: q.Region, Source: "none"}}
	}

	paidCap := records.DetectivePaidCap()
	paidUsed := 0
	var fanoutPairs int
	hasLastName := lastA != "" || lastB != ""

	baseMulti, multiOK := a.Records.(*records.Multi)
	if a.Records == nil {
		baseMulti = records.NewMulti()
		multiOK = true
	}

	// Paid-only fan-out — skip entirely without last names (saves budget)
	if multiOK && baseMulti.HasPaidProviders() && hasLastName {
		for _, loc := range locs {
			for _, nv := range names {
				if err := ctx.Err(); err != nil {
					searchErr = err
					break
				}
				// Early stop only on confirmable multi-source-ready streets
				if records.HasConfirmableStreet(records.MergeResults(parts...).Candidates) &&
					records.MaxStreetConf(records.MergeResults(parts...).Candidates) >= 0.85 {
					break
				}
				if paidUsed >= paidCap {
					break
				}
				// Skip paid name variants with empty last (waste)
				if nv.Last == "" {
					continue
				}
				qv := q
				qv.FirstName, qv.LastName, qv.Handle = nv.First, nv.Last, nv.Handle
				qv.City, qv.Region = loc.City, loc.Region
				if strings.EqualFold(nv.First, firstA) {
					qv.PartnerFirst, qv.PartnerLast = firstB, lastB
				} else {
					qv.PartnerFirst, qv.PartnerLast = firstA, lastA
				}
				m := cloneMultiPaidOnly(baseMulti, paidCap-paidUsed)
				rv, e := m.Search(ctx, qv)
				if e != nil {
					searchErr = e
				}
				tagParts := []string{}
				if nv.Note != "" {
					tagParts = append(tagParts, nv.Note)
				}
				if loc.Source != "" && loc.Source != "kit" {
					tagParts = append(tagParts, "location="+loc.Source+":"+loc.City)
				}
				tag := strings.Join(tagParts, " · ")
				if tag != "" {
					for i := range rv.Candidates {
						if rv.Candidates[i].Note == "" {
							rv.Candidates[i].Note = tag
						} else {
							rv.Candidates[i].Note = tag + " · " + rv.Candidates[i].Note
						}
					}
				}
				paidUsed += rv.PaidCalls
				fanoutPairs++
				if len(rv.Candidates) > 0 || rv.Status == "ok" {
					parts = append(parts, rv)
				}
			}
			if paidUsed >= paidCap || (records.HasConfirmableStreet(records.MergeResults(parts...).Candidates) &&
				records.MaxStreetConf(records.MergeResults(parts...).Candidates) >= 0.85) {
				break
			}
		}
	}

	// Free-tier multi-location recovery when no confirmable street yet
	mergedSoFar := records.MergeResults(parts...)
	if !records.HasConfirmableStreet(mergedSoFar.Candidates) {
		// Cap free location fan-out: top 3 locs × primary A + primary B (max 6 free Multis)
		freeLocs := locs
		if len(freeLocs) > 3 {
			freeLocs = freeLocs[:3]
		}
		freeNames := names
		if len(freeNames) > 2 {
			freeNames = freeNames[:2] // A + B primaries only
		}
		if len(freeNames) == 0 {
			freeNames = []records.NameVariant{{First: firstA, Last: lastA, Handle: k.HandleA}}
		}
		freeTried := 0
		const maxFreePasses = 4
		for _, loc := range freeLocs {
			for _, nv := range freeNames {
				if freeTried >= maxFreePasses {
					break
				}
				if records.HasConfirmableStreet(records.MergeResults(parts...).Candidates) {
					break
				}
				if err := ctx.Err(); err != nil {
					searchErr = err
					break
				}
				qv := q
				qv.FirstName, qv.LastName, qv.Handle = nv.First, nv.Last, nv.Handle
				qv.City, qv.Region = loc.City, loc.Region
				qv.PartnerFirst, qv.PartnerLast = firstB, lastB

				var rv records.Result
				var e error
				if multiOK {
					// Free recovery: free scrapers + heuristic; residual paid only if budget left
					var m *records.Multi
					if paidUsed >= paidCap {
						m = cloneMultiFreeOnly(baseMulti)
					} else {
						m = cloneMultiFull(baseMulti, paidCap-paidUsed)
						m.PrimaryOnlyPaid = true
					}
					rv, e = m.Search(ctx, qv)
					paidUsed += rv.PaidCalls
				} else if a.Records != nil {
					rv, e = a.Records.Search(ctx, qv)
				} else {
					m := records.NewMulti()
					rv, e = m.Search(ctx, qv)
					paidUsed += rv.PaidCalls
				}
				if e != nil {
					searchErr = e
				}
				if loc.Source != "" {
					for i := range rv.Candidates {
						tag := "free_pass location=" + loc.Source
						if rv.Candidates[i].Note == "" {
							rv.Candidates[i].Note = tag
						} else {
							rv.Candidates[i].Note = tag + " · " + rv.Candidates[i].Note
						}
					}
				}
				parts = append(parts, rv)
				fanoutPairs++
				freeTried++
			}
			if freeTried >= maxFreePasses || records.HasConfirmableStreet(records.MergeResults(parts...).Candidates) {
				break
			}
		}
	}

	res := records.MergeResults(parts...)
	err = searchErr
	if res.Provider == "" || res.Provider == "merged" {
		res.Provider = fmt.Sprintf("detective(pairs=%d,paid=%d/%d)", fanoutPairs, paidUsed, paidCap)
	} else {
		res.Provider = fmt.Sprintf("%s·fanout(pairs=%d,paid=%d/%d)", res.Provider, fanoutPairs, paidUsed, paidCap)
	}
	res.PaidCalls = paidUsed

	// Audit: log primary query shape + fan-out metadata
	qJSON, _ := json.Marshal(map[string]any{
		"base":           q,
		"names":          names,
		"locations":      locs,
		"fanout_pairs":   fanoutPairs,
		"paid_used":      paidUsed,
		"paid_cap":       paidCap,
		"text_extract_n": len(extracted),
	})
	cJSON, _ := json.Marshal(res.Candidates)
	st := res.Status
	if st == "" {
		st = "ok"
	}
	if err != nil && st == "ok" && len(res.Candidates) == 0 {
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

	// Map to store candidates (street first already ranked by Multi)
	var cands []store.AddressCandidate
	bestConf := 0.0
	for _, c := range res.Candidates {
		// Never persist research URLs into Line1
		line1 := c.Line1
		if c.Kind == records.KindResearchLink || strings.HasPrefix(strings.ToLower(line1), "http") {
			line1 = ""
		}
		ac := store.AddressCandidate{
			Line1: line1, Line2: c.Line2, City: c.City, Region: c.Region,
			Postal: c.Postal, Country: firstNonEmpty(c.Country, "US"),
			Confidence: c.Confidence, Source: c.Source, Note: c.Note,
			Kind: c.Kind, URL: c.URL,
		}
		if ac.Kind == "" {
			if records.IsRealStreet(ac.Line1) {
				ac.Kind = records.KindStreet
			} else if ac.URL != "" {
				ac.Kind = records.KindResearchLink
			} else if ac.City != "" {
				ac.Kind = records.KindLocality
			}
		}
		cands = append(cands, ac)
		if records.IsRealStreet(ac.Line1) && c.Confidence > bestConf {
			bestConf = c.Confidence
		}
	}

	// Lob verification post-step: verify each real street candidate
	if a.Mail != nil && a.Mail.Available() && len(cands) > 0 {
		for i := range cands {
			c := &cands[i]
			if !records.IsRealStreet(c.Line1) || c.City == "" {
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
				// Boost without erasing identity score; USPS ≠ right person
				if c.Confidence < 0.75 {
					c.Confidence = 0.75
				} else {
					c.Confidence += 0.08
					if c.Confidence > 0.92 {
						c.Confidence = 0.92
					}
				}
				c.Note = strings.TrimSpace(c.Note + " · Lob USPS deliverable (identity still operator-confirmed)")
			} else {
				if c.Confidence > 0.55 {
					c.Confidence = 0.55
				}
				c.Note = strings.TrimSpace(c.Note + " · Lob USPS NOT deliverable — review unit/street")
			}
			c.Source = "lob_check_" + c.Source
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

	streetN, linkN := 0, 0
	for _, c := range cands {
		if records.IsRealStreet(c.Line1) {
			streetN++
		} else if c.URL != "" || c.Kind == records.KindResearchLink {
			linkN++
		}
	}
	locList := make([]string, 0, len(records.LocationVariants(q)))
	for _, lv := range records.LocationVariants(q) {
		locList = append(locList, fmt.Sprintf("%s(%s)", lv.City, lv.Source))
	}
	k.ResearchNotes = strings.TrimSpace(k.ResearchNotes + "\n\n--- Detective run " + time.Now().UTC().Format(time.RFC3339) + " ---\n" +
		fmt.Sprintf("Provider: %s · status: %s · candidates: %d (%d street, %d research links)\n", res.Provider, st, len(cands), streetN, linkN) +
		fmt.Sprintf("Query: %s %s · %s, %s\n", firstA, lastA, q.City, q.Region) +
		fmt.Sprintf("Fan-out locations: %s\n", strings.Join(locList, ", ")) +
		sigStr + countyStr + "\n" +
		"Pick a street candidate, then Verify address. Fan-out: names×locations under DETECTIVE_PAID_CAP; free text extract first; free scrapers once if no street.")

	// Mark detective step done in research steps
	steps := k.ResearchSteps
	found := false
	for i := range steps {
		if steps[i].ID == "people" || steps[i].ID == "detective" {
			steps[i].Status = "done"
			steps[i].Detail = fmt.Sprintf("Ran %s → %d candidates (%d street)", res.Provider, len(cands), streetN)
			found = true
		}
	}
	if !found {
		steps = append([]store.ResearchStep{{
			ID: "detective", Label: "Detective run",
			Detail: fmt.Sprintf("%s returned %d candidates (%d street)", res.Provider, len(cands), streetN),
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

// ParseAddressText extracts a street from operator-pasted research text and appends
// it as a candidate with source operator_paste (does not auto-verify).
func (a *Agent) ParseAddressText(kitID, text string) (store.CongratulateKit, error) {
	k, err := a.Store.GetCongratulateKit(kitID)
	if err != nil {
		return k, err
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return k, fmt.Errorf("paste address text from research (TruePeopleSearch, whitepages, etc.)")
	}
	if strings.HasPrefix(strings.ToLower(text), "http") {
		return k, fmt.Errorf("paste the address block, not a URL")
	}
	addr := records.ExtractAddressFromBio(text)
	if addr == nil || !records.IsRealStreet(addr.Line1) {
		return k, fmt.Errorf("could not parse a US street address from pasted text")
	}
	c := store.AddressCandidate{
		Line1: addr.Line1, Line2: addr.Line2,
		City: firstNonEmpty(addr.City, k.AddressCity, k.MarketCity),
		Region: firstNonEmpty(addr.Region, k.AddressRegion, k.MarketRegion),
		Postal: addr.Postal, Country: "US",
		Confidence: 0.70, Source: "operator_paste", Kind: records.KindStreet,
		Note: "Parsed from operator paste — verify with Lob before mail.",
	}
	// Prepend as top candidate and fill form fields
	k.AddressCandidates = append([]store.AddressCandidate{c}, k.AddressCandidates...)
	k.AddressLine1 = c.Line1
	k.AddressLine2 = c.Line2
	if c.City != "" {
		k.AddressCity = c.City
	}
	if c.Region != "" {
		k.AddressRegion = c.Region
	}
	if c.Postal != "" {
		k.AddressPostal = c.Postal
	}
	k.AddressSource = "operator_paste"
	k.AddressConfidence = c.Confidence
	if k.Status == "address_verified" || k.Status == "ready_to_mail" {
		k.Status = "ready_review"
		k.VerifiedAt = nil
		k.VerifiedBy = ""
	}
	setMailMeta(&k, "lob_deliverable", false)
	setMailMeta(&k, "verified_fp", "")
	k.PostcardHTML = RenderPostcardHTML(k)
	mp := mailPayload(k)
	if k.MailPayload != nil {
		for key, val := range k.MailPayload {
			if key == "lob_deliverable" || key == "verified_fp" {
				mp[key] = val
			}
		}
	}
	k.MailPayload = mp
	return a.Store.UpsertCongratulateKit(k)
}

// ApplyCandidate copies one ranked street candidate onto the kit mailing fields (not yet verified).
func (a *Agent) ApplyCandidate(kitID string, idx int) (store.CongratulateKit, error) {
	k, err := a.Store.GetCongratulateKit(kitID)
	if err != nil {
		return k, err
	}
	if idx < 0 || idx >= len(k.AddressCandidates) {
		return k, fmt.Errorf("candidate index %d out of range", idx)
	}
	c := k.AddressCandidates[idx]
	if c.Kind == records.KindResearchLink || strings.HasPrefix(strings.ToLower(c.Line1), "http") {
		return k, fmt.Errorf("cannot apply research link as street — open the link and paste the address")
	}
	if !records.IsRealStreet(c.Line1) {
		return k, fmt.Errorf("candidate has no real street (got %q) — pick a street candidate or type an address", c.Line1)
	}
	k.AddressLine1 = c.Line1
	k.AddressLine2 = c.Line2
	k.AddressCity = c.City
	k.AddressRegion = c.Region
	k.AddressPostal = c.Postal
	k.AddressCountry = firstNonEmpty(c.Country, "US")
	k.AddressConfidence = c.Confidence
	k.AddressSource = c.Source
	// Address change invalidates prior verification
	if k.Status == "address_verified" || k.Status == "ready_to_mail" {
		k.Status = "ready_review"
		k.VerifiedAt = nil
		k.VerifiedBy = ""
	}
	setMailMeta(&k, "lob_deliverable", false)
	setMailMeta(&k, "verified_fp", "")
	k.PostcardHTML = RenderPostcardHTML(k)
	mp := mailPayload(k)
	// Preserve meta flags we just set
	if k.MailPayload != nil {
		for key, val := range k.MailPayload {
			if key == "lob_deliverable" || key == "verified_fp" {
				mp[key] = val
			}
		}
	}
	k.MailPayload = mp
	return a.Store.UpsertCongratulateKit(k)
}

// VerifyAndConfirm runs Lob USPS verify when available, then marks address_verified
// only when deliverable (or audited override when Lob is unavailable).
func (a *Agent) VerifyAndConfirm(ctx context.Context, kitID, verifiedBy string) (store.CongratulateKit, error) {
	k, err := a.Store.GetCongratulateKit(kitID)
	if err != nil {
		return k, err
	}
	if !records.IsRealStreet(k.AddressLine1) || k.AddressCity == "" {
		return k, fmt.Errorf("real street and city required — pick a candidate or type an address first")
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
			return k, fmt.Errorf("address verify failed — kit NOT marked verified: %w", err)
		}
		// Apply standardized components
		k.AddressLine1 = firstNonEmpty(vr.Address.AddressLine1, k.AddressLine1)
		if vr.Address.AddressLine2 != "" {
			k.AddressLine2 = vr.Address.AddressLine2
		}
		k.AddressCity = firstNonEmpty(vr.Address.AddressCity, k.AddressCity)
		k.AddressRegion = firstNonEmpty(vr.Address.AddressState, k.AddressRegion)
		k.AddressPostal = firstNonEmpty(vr.Address.AddressZip, k.AddressPostal)
		k.AddressSource = "lob_usps"
		setMailMeta(&k, "lob_deliverable", vr.Deliverable)
		setMailMeta(&k, "verified_fp", addressFingerprint(k.AddressLine1, k.AddressLine2, k.AddressCity, k.AddressRegion, k.AddressPostal))
		if !vr.Deliverable {
			k.AddressConfidence = 0.55
			k.ResearchNotes = strings.TrimSpace(k.ResearchNotes + "\nLob USPS: NOT deliverable — fix unit/street before verify.")
			// Persist notes + standardized fields but do NOT mark verified
			k.PostcardHTML = RenderPostcardHTML(k)
			mp := mailPayload(k)
			if k.MailPayload != nil {
				for key, val := range k.MailPayload {
					mp[key] = val
				}
			}
			k.MailPayload = mp
			if k.Status == "address_verified" || k.Status == "ready_to_mail" {
				k.Status = "ready_review"
			}
			saved, err := a.Store.UpsertCongratulateKit(k)
			if err != nil {
				return k, err
			}
			return saved, fmt.Errorf("Lob USPS reports address not deliverable — review unit/street and try again")
		}
		// Deliverable: boost conf without erasing multi-source identity score
		if k.AddressConfidence < 0.88 {
			k.AddressConfidence = 0.88
		}
		if k.AddressConfidence > 0.95 {
			k.AddressConfidence = 0.95
		}
		return a.UpdateKitAddress(kitID, k, true, verifiedBy)
	}

	// Lob unavailable: allow operator assert with explicit source tag (dev / offline).
	// Still require real street+city+zip.
	if k.AddressPostal == "" {
		return k, fmt.Errorf("postal code required when Lob is not configured")
	}
	k.AddressSource = "operator_asserted_no_lob"
	k.ResearchNotes = strings.TrimSpace(k.ResearchNotes + "\n⚠ Verified without Lob USPS — LOB_API_KEY missing. Confirm deliverability manually before mail.")
	setMailMeta(&k, "lob_deliverable", false)
	setMailMeta(&k, "verified_fp", addressFingerprint(k.AddressLine1, k.AddressLine2, k.AddressCity, k.AddressRegion, k.AddressPostal))
	setMailMeta(&k, "operator_asserted", true)
	return a.UpdateKitAddress(kitID, k, true, verifiedBy)
}

// SendPostcard mails via Lob when configured. Requires ready_to_mail (or address_verified
// with deliverable flag) and fingerprint match.
func (a *Agent) SendPostcard(ctx context.Context, kitID string) (store.CongratulateKit, error) {
	k, err := a.Store.GetCongratulateKit(kitID)
	if err != nil {
		return k, err
	}
	if k.Status != "address_verified" && k.Status != "ready_to_mail" {
		return k, fmt.Errorf("verify address before sending (status=%s)", k.Status)
	}
	if !records.IsRealStreet(k.AddressLine1) || k.AddressPostal == "" {
		return k, fmt.Errorf("complete mailing street address required")
	}
	if a.Mail == nil || !a.Mail.Available() {
		return k, fmt.Errorf("LOB_API_KEY not configured — set key + LOB_FROM_* return address")
	}
	// Fingerprint: block send if address changed after verify
	fp := addressFingerprint(k.AddressLine1, k.AddressLine2, k.AddressCity, k.AddressRegion, k.AddressPostal)
	if stored := mailMetaString(k, "verified_fp"); stored != "" && stored != fp {
		return k, fmt.Errorf("address changed after verify — re-run Verify address before send")
	}
	// Prefer deliverable; allow operator_asserted only if explicitly ready_to_mail
	if !mailMetaBool(k, "lob_deliverable") && !mailMetaBool(k, "operator_asserted") {
		// Re-verify immediately before send (final preflight)
		vr, err := a.Mail.VerifyAddress(ctx, mail.Address{
			Name: strings.TrimSpace(k.PersonAName + " & " + k.PersonBName),
			AddressLine1: k.AddressLine1, AddressLine2: k.AddressLine2,
			AddressCity: k.AddressCity, AddressState: k.AddressRegion,
			AddressZip: k.AddressPostal, AddressCountry: firstNonEmpty(k.AddressCountry, "US"),
		})
		if err != nil {
			return k, fmt.Errorf("preflight verify failed: %w", err)
		}
		if !vr.Deliverable {
			return k, fmt.Errorf("preflight: Lob reports not deliverable — fix address before send")
		}
		setMailMeta(&k, "lob_deliverable", true)
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

// hasStreetCandidates reports whether any candidate has a real mailing street
// (not a research URL or city-only locality).
func hasStreetCandidates(cs []records.Candidate) bool {
	return records.HasStreetCandidates(cs)
}

// cloneMultiPaidOnly copies Multi for name×location fan-out: one paid API per pair.
func cloneMultiPaidOnly(src *records.Multi, maxPaid int) *records.Multi {
	if src == nil {
		src = records.NewMulti()
	}
	return &records.Multi{
		Primary:         src.Primary,
		Paid:            src.Paid,
		Free:            nil,
		Fallback:        nil,
		MaxPaidCalls:    maxPaid,
		SkipFree:        true,
		SkipFallback:    true,
		PrimaryOnlyPaid: true, // stretch budget across more name×loc pairs
	}
}

// cloneMultiFull runs the full cascade (paid remaining + free + heuristic).
func cloneMultiFull(src *records.Multi, maxPaid int) *records.Multi {
	if src == nil {
		src = records.NewMulti()
	}
	return &records.Multi{
		Primary:      src.Primary,
		Paid:         src.Paid,
		Free:         src.Free,
		Fallback:     src.Fallback,
		MaxPaidCalls: maxPaid,
		SkipFree:     false,
		SkipFallback: false,
	}
}

// cloneMultiFreeOnly runs Google/TPS/property/voter/heuristic with no paid APIs.
func cloneMultiFreeOnly(src *records.Multi) *records.Multi {
	if src == nil {
		src = records.NewMulti()
	}
	return &records.Multi{
		Primary:      &records.Heuristic{},
		Paid:         nil,
		Free:         src.Free,
		Fallback:     src.Fallback,
		SkipFree:     false,
		SkipFallback: false,
	}
}

func addressFingerprint(line1, line2, city, region, postal string) string {
	return strings.ToLower(strings.Join([]string{
		strings.TrimSpace(line1),
		strings.TrimSpace(line2),
		strings.TrimSpace(city),
		strings.ToUpper(strings.TrimSpace(region)),
		strings.TrimSpace(postal),
	}, "|"))
}

func setMailMeta(k *store.CongratulateKit, key string, val any) {
	if k.MailPayload == nil {
		k.MailPayload = map[string]any{}
	}
	k.MailPayload[key] = val
}

func mailMetaBool(k store.CongratulateKit, key string) bool {
	if k.MailPayload == nil {
		return false
	}
	v, ok := k.MailPayload[key]
	if !ok {
		return false
	}
	b, _ := v.(bool)
	return b
}

func mailMetaString(k store.CongratulateKit, key string) string {
	if k.MailPayload == nil {
		return ""
	}
	v, _ := k.MailPayload[key].(string)
	return v
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
