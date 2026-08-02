// Package outreach builds human-reviewed congratulate kits: dossier from
// discovery signals, address research plan (never auto-mail), and postcard draft.
package outreach

import (
	"context"
	"fmt"
	"html"
	"math"
	"net/url"
	"strings"
	"time"

	"neptune-social-radar/backend/internal/llm"
	"neptune-social-radar/backend/internal/mail"
	"neptune-social-radar/backend/internal/records"
	"neptune-social-radar/backend/internal/signals"
	"neptune-social-radar/backend/internal/store"
)

// Agent builds and updates congratulate kits.
type Agent struct {
	Store   *store.Store
	LLM     llm.Interpreter
	Records records.Provider // people-search (Trestle / PDL / Cleanlist / heuristic)
	Mail    *mail.Client     // Lob verify + send (optional)
}

// BuildKit gathers all known data for a couple and produces a ready_review kit.
// Address street lines are never invented — only market-level candidates plus
// a research checklist for human public-record lookup.
func (a *Agent) BuildKit(ctx context.Context, coupleID string) (store.CongratulateKit, error) {
	couple, err := a.Store.GetCouple(coupleID)
	if err != nil {
		return store.CongratulateKit{}, fmt.Errorf("couple: %w", err)
	}

	// Prefer board projection (already joins accounts/bios/pics).
	board, _ := a.Store.ListProspectBoard(300)
	var card *store.ProspectCard
	for i := range board {
		if board[i].CoupleID == coupleID {
			card = &board[i]
			break
		}
	}

	kit := store.CongratulateKit{
		CoupleID: coupleID,
		Status:   "draft",
	}
	if card != nil {
		kit.HandleA, kit.HandleB = card.HandleA, card.HandleB
		kit.PersonAName = firstNonEmpty(card.PersonALabel, card.HandleA)
		kit.PersonBName = firstNonEmpty(card.PersonBLabel, card.HandleB)
		kit.BioA, kit.BioB = card.BioA, card.BioB
		kit.ProfilePicA, kit.ProfilePicB = card.ProfilePicA, card.ProfilePicB
		kit.MarketCity, kit.MarketRegion = card.City, card.Region
	} else {
		// Fallback: accounts by person
		if acct, err := a.Store.GetAccountByPersonID(couple.PersonAID); err == nil {
			kit.HandleA = acct.Handle
			kit.PersonAName = firstNonEmpty(acct.DisplayName, acct.Handle)
			kit.BioA = acct.BioText
			kit.ProfilePicA = acct.ProfilePicURL
			if kit.MarketCity == "" {
				kit.MarketCity, kit.MarketRegion = acct.InferredCity, acct.InferredRegion
			}
		}
		if acct, err := a.Store.GetAccountByPersonID(couple.PersonBID); err == nil {
			kit.HandleB = acct.Handle
			kit.PersonBName = firstNonEmpty(acct.DisplayName, acct.Handle)
			kit.BioB = acct.BioText
			kit.ProfilePicB = acct.ProfilePicURL
			if kit.MarketCity == "" {
				kit.MarketCity, kit.MarketRegion = acct.InferredCity, acct.InferredRegion
			}
		}
		if couple.InferredCity != "" {
			kit.MarketCity, kit.MarketRegion = couple.InferredCity, couple.InferredRegion
			kit.MarketSource = couple.LocationSource
		}
	}

	// Discovery post + photographer market
	cap, img, url, postLoc, srcHandle, found := a.Store.FindDiscoveryPost(kit.HandleA, kit.HandleB)
	if found {
		kit.DiscoveryCaption = truncate(cap, 400)
		kit.DiscoveryImageURL = img
		kit.DiscoveryPostURL = url
		kit.SourceHandle = srcHandle
		if src, err := a.Store.GetWatchedSource(srcHandle); err == nil {
			kit.SourceClass = src.SourceClass
			if kit.MarketCity == "" && src.City != "" {
				kit.MarketCity, kit.MarketRegion = src.City, src.State
				kit.MarketSource = "photographer_profile"
			} else if src.City != "" && kit.MarketSource == "" {
				kit.MarketSource = "photographer_profile"
			}
		}
		if postLoc != "" {
			if g, ok := signals.InferLocationFromText(postLoc, "post"); ok {
				if kit.MarketCity == "" {
					kit.MarketCity, kit.MarketRegion = g.City, g.Region
				}
				if kit.MarketSource == "" {
					kit.MarketSource = "post_location"
				}
			}
		}
	}

	// Bio location inference
	if g, ok := signals.BestLocation(postLoc, kit.BioA, kit.BioB, cap); ok {
		if kit.MarketCity == "" {
			kit.MarketCity, kit.MarketRegion = g.City, g.Region
			kit.MarketSource = g.Source
		}
	}

	// Resolve first names: caption > display_name > bio > handle.
	// Caption like "between Alida and Andrew" is the gold source for postcards.
	fullCap := cap
	if fullCap == "" {
		fullCap = kit.DiscoveryCaption
	}
	// Prefer live account display names when available
	displayA, displayB := kit.PersonAName, kit.PersonBName
	if acct, err := a.Store.GetAccountByPersonID(couple.PersonAID); err == nil {
		if acct.DisplayName != "" {
			displayA = acct.DisplayName
		}
		if kit.BioA == "" {
			kit.BioA = acct.BioText
		}
	}
	if acct, err := a.Store.GetAccountByPersonID(couple.PersonBID); err == nil {
		if acct.DisplayName != "" {
			displayB = acct.DisplayName
		}
		if kit.BioB == "" {
			kit.BioB = acct.BioText
		}
	}

	rnA, rnB := signals.ResolveCoupleFirstNames(
		fullCap,
		displayA, displayB,
		kit.BioA, kit.BioB,
		kit.HandleA, kit.HandleB,
	)

	// Persist structured first + last (last is what makes people-search work)
	kit.FirstNameA, kit.LastNameA = rnA.First, rnA.Last
	kit.FirstNameB, kit.LastNameB = rnB.First, rnB.Last
	kit.NameSourceA, kit.NameSourceB = rnA.Source, rnB.Source
	// Postcard greeting uses first name; person_* holds best display for UI
	if rnA.First != "" {
		kit.PersonAName = rnA.First
		if rnA.Last != "" {
			kit.PersonAName = rnA.First // keep first for Dear X; last lives in last_name_a
		}
	}
	if rnB.First != "" {
		kit.PersonBName = rnB.First
	}

	// Evidence trail (after names so we can note them)
	kit.Evidence = buildEvidence(kit, found, rnA, rnB)

	// Address research (honest: city/state only until human verifies street)
	kit.AddressCandidates, kit.ResearchSteps, kit.ResearchNotes, kit.AddressConfidence, kit.AddressSource =
		researchAddress(kit, rnA, rnB)

	// Prefer top candidate for city/region fields (not street)
	if len(kit.AddressCandidates) > 0 {
		top := kit.AddressCandidates[0]
		kit.AddressCity = top.City
		kit.AddressRegion = top.Region
		kit.AddressCountry = firstNonEmpty(top.Country, "US")
		kit.AddressLine1 = top.Line1
		kit.AddressLine2 = top.Line2
		kit.AddressPostal = top.Postal
	}

	// Curated copy — always first names on the card
	nameA := firstNonEmpty(kit.FirstNameA, rnA.First, prettyName(kit.PersonAName, kit.HandleA), "there")
	nameB := firstNonEmpty(kit.FirstNameB, rnB.First, prettyName(kit.PersonBName, kit.HandleB), "partner")
	locLabel := ""
	if kit.MarketCity != "" {
		locLabel = kit.MarketCity
		if kit.MarketRegion != "" {
			locLabel += ", " + kit.MarketRegion
		}
	}
	kit.PersonAName = nameA
	kit.PersonBName = nameB
	kit.FirstNameA, kit.FirstNameB = nameA, nameB

	copyOut, err := a.LLM.DraftCopy(ctx, llm.CopyRequest{
		ActionType:      "postcard",
		EventType:       "engagement",
		PersonName:      nameA,
		PartnerName:     nameB,
		Confidence:      kit.AddressConfidence,
		EvidenceSummary: kit.Evidence,
		Location:        locLabel,
	})
	if err != nil {
		copyOut = postcardFallback(nameA, nameB, locLabel)
	}
	kit.Headline = "Congratulations"
	kit.BodyMessage = preferFirstNameCopy(copyOut.CustomerFacing, nameA, nameB, locLabel)
	nameNote := fmt.Sprintf("\n\nName resolution: %s %s (%s) & %s %s (%s)",
		nameA, kit.LastNameA, rnA.Source, nameB, kit.LastNameB, rnB.Source)
	if kit.LastNameA == "" && kit.LastNameB == "" {
		nameNote += "\n⚠ No last names yet — people-search will be weak. Edit last names before Run detective, or rely on IG display_name."
	}
	kit.InternalNote = copyOut.InternalNote + nameNote
	if kit.BodyMessage == "" {
		kit.BodyMessage = postcardFallback(nameA, nameB, locLabel).CustomerFacing
	}

	kit.PostcardHTML = RenderPostcardHTML(kit)
	kit.MailPayload = mailPayload(kit)
	kit.Status = "ready_review"

	// Compute priority score (higher = operator should work this kit first)
	kit.PriorityScore = ComputePriorityScore(kit)

	// Schedule follow-up card 14 days after first mail (if address gets verified)
	if kit.AddressConfidence >= 0.50 {
		followUp := time.Now().UTC().AddDate(0, 0, 14)
		kit.FollowUpAt = &followUp
		kit.FollowUpTemplate = "bright_casual" // different template for follow-up
	}

	saved, err := a.Store.UpsertCongratulateKit(kit)
	if err != nil {
		return saved, err
	}

	// Auto-detective: run detective immediately if we have enough data (last names + city)
	if saved.LastNameA != "" || saved.LastNameB != "" {
		if saved.MarketCity != "" || saved.AddressCity != "" {
			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
				defer cancel()
				_, _ = a.RunDetective(ctx, saved.ID)
			}()
		}
	}

	return saved, nil
}

// computePriorityScore combines multiple signals into a single score (0-1).
// Higher = more valuable kit to work first.
// ComputePriorityScore combines multiple signals into a single score (0-1).
// Higher = more valuable kit to work first.
func ComputePriorityScore(kit store.CongratulateKit) float64 {
	score := 0.0

	// Address confidence (40% weight) — the most important signal
	score += kit.AddressConfidence * 0.40

	// Name completeness (25% weight) — last names make people-search work
	nameScore := 0.0
	if kit.LastNameA != "" {
		nameScore += 0.5
	}
	if kit.LastNameB != "" {
		nameScore += 0.5
	}
	score += nameScore * 0.25

	// Market location quality (15% weight) — Ohio cities score higher than unknown
	if kit.MarketCity != "" {
		marketScore := 0.5 // generic city
		if kit.MarketSource == "photographer_profile" || kit.MarketSource == "post_location" {
			marketScore = 0.8 // strong source
		}
		if kit.MarketSource == "bio_agreement" {
			marketScore = 1.0 // both bios agree
		}
		score += marketScore * 0.15
	}

	// Vendor quality (10% weight) — kits from known vendors are better
	if kit.SourceHandle != "" {
		score += 0.08 // has a source vendor
	}

	// Recency (10% weight) — newer kits are more timely
	hoursSinceCreation := time.Since(kit.CreatedAt).Hours()
	if hoursSinceCreation < 24 {
		score += 0.10 // fresh
	} else if hoursSinceCreation < 72 {
		score += 0.07
	} else if hoursSinceCreation < 168 {
		score += 0.04
	}

	return math.Min(score, 1.0)
}

func nameSourceRank(src string) int {
	switch src {
	case "caption":
		return 4
	case "display_name":
		return 3
	case "bio":
		return 2
	case "handle":
		return 1
	default:
		return 0
	}
}

// preferFirstNameCopy ensures the postcard body greets with first names even
// if the LLM echoed handles or odd tokens.
func preferFirstNameCopy(body, nameA, nameB, loc string) string {
	body = strings.TrimSpace(body)
	if body == "" {
		return postcardFallback(nameA, nameB, loc).CustomerFacing
	}
	// If message already opens with Dear <NameA>, keep it.
	lower := strings.ToLower(body)
	if strings.Contains(lower, strings.ToLower(nameA)) && strings.Contains(lower, strings.ToLower(nameB)) {
		return body
	}
	// Replace a weak opener
	return postcardFallback(nameA, nameB, loc).CustomerFacing
}

func buildEvidence(kit store.CongratulateKit, foundPost bool, rnA, rnB signals.ResolvedName) []string {
	var ev []string
	if rnA.First != "" && rnB.First != "" {
		if rnA.Last != "" || rnB.Last != "" {
			ev = append(ev, fmt.Sprintf("Names: %s %s & %s %s (via %s / %s)",
				rnA.First, rnA.Last, rnB.First, rnB.Last, rnA.Source, rnB.Source))
		} else {
			ev = append(ev, fmt.Sprintf("First names: %s & %s (via %s / %s) — last names missing for strong address search",
				rnA.First, rnB.First, rnA.Source, rnB.Source))
		}
	}
	if kit.HandleA != "" && kit.HandleB != "" {
		ev = append(ev, fmt.Sprintf("Tagged pair @%s & @%s", kit.HandleA, kit.HandleB))
	}
	if foundPost && kit.SourceHandle != "" {
		ev = append(ev, fmt.Sprintf("Discovered via photographer/vendor @%s", kit.SourceHandle))
	}
	if kit.DiscoveryCaption != "" {
		ev = append(ev, "Discovery caption: "+truncate(kit.DiscoveryCaption, 120))
	}
	if kit.MarketCity != "" {
		ev = append(ev, fmt.Sprintf("Market signal: %s, %s (%s)", kit.MarketCity, kit.MarketRegion, kit.MarketSource))
	}
	if kit.BioA != "" {
		ev = append(ev, "Bio A: "+truncate(kit.BioA, 80))
	}
	if kit.BioB != "" {
		ev = append(ev, "Bio B: "+truncate(kit.BioB, 80))
	}
	if kit.DiscoveryImageURL != "" {
		ev = append(ev, "Discovery photo available for postcard front")
	}
	return ev
}

func researchAddress(kit store.CongratulateKit, rnA, rnB signals.ResolvedName) (cands []store.AddressCandidate, steps []store.ResearchStep, notes string, conf float64, source string) {
	city, region := kit.MarketCity, kit.MarketRegion
	source = kit.MarketSource
	firstA := firstNonEmpty(rnA.First, kit.PersonAName, kit.HandleA)
	firstB := firstNonEmpty(rnB.First, kit.PersonBName, kit.HandleB)
	nameOK := rnA.Source == "caption" || rnA.Source == "display_name" || rnA.Source == "bio" ||
		rnB.Source == "caption" || rnB.Source == "display_name" || rnB.Source == "bio"

	if city == "" {
		notes = "No city/region signal yet. Enrich profiles, re-scan the photographer, or enter market manually before public-record lookup."
		steps = []store.ResearchStep{
			{ID: "market", Label: "Establish market", Detail: "Need city + state from bio, post geotag, or photographer location.", Status: "blocked"},
			{
				ID: "names", Label: "First names",
				Detail: fmt.Sprintf("%s & %s", firstA, firstB),
				Status: map[bool]string{true: "done", false: "suggested"}[nameOK],
			},
			{ID: "county", Label: "County property / voter / whitepages", Detail: "Once market + names are known, search county assessor and people-search directories.", Status: "suggested"},
			{ID: "verify", Label: "Human verify before mail", Detail: "Never mail on model confidence alone. Confirm address with a second source.", Status: "suggested"},
		}
		return nil, steps, notes, 0, ""
	}

	// City-level candidate only — street unknown
	cands = append(cands, store.AddressCandidate{
		City: city, Region: region, Country: "US",
		Confidence: 0.35,
		Source:     firstNonEmpty(source, "market_inference"),
		Note:       "City/region only — street address not yet verified. Not mail-ready.",
	})
	conf = 0.35

	// Stronger if both bios agree on city
	if gA, okA := signals.InferLocationFromText(kit.BioA, "bio"); okA {
		if gB, okB := signals.InferLocationFromText(kit.BioB, "bio"); okB {
			if strings.EqualFold(gA.City, gB.City) {
				cands[0].Confidence = 0.55
				conf = 0.55
				cands[0].Note = "Both bios mention the same city. Still need street address."
				source = "bio_agreement"
			}
		}
	}

	// Name-aware research tools (public web search — human runs them)
	countyURL := countySearchURL(city, region)
	peopleURL := googleQ(fmt.Sprintf(`"%s" "%s" %s %s`, firstA, firstB, city, region))
	whitepagesURL := googleQ(fmt.Sprintf("%s %s %s %s phone address", firstA, firstB, city, region))
	propertyURL := googleQ(fmt.Sprintf("%s %s %s %s property owner OR deed OR assessor", firstA, firstB, city, region))
	marriageURL := googleQ(fmt.Sprintf("%s %s marriage license OR engagement %s %s", firstA, firstB, city, region))

	nameStatus := "suggested"
	nameDetail := fmt.Sprintf("Working first names: %s & %s.", firstA, firstB)
	if nameOK {
		nameStatus = "done"
		nameDetail = fmt.Sprintf("First names from %s / %s: %s & %s. Confirm surnames for public records if mailing.",
			rnA.Source, rnB.Source, firstA, firstB)
	}

	steps = []store.ResearchStep{
		{
			ID: "market", Label: "Market established",
			Detail: fmt.Sprintf("%s, %s from %s", city, region, firstNonEmpty(source, "signals")),
			Status: "done",
		},
		{
			ID: "names", Label: "First names for postcard",
			Detail: nameDetail,
			Status: nameStatus,
		},
		{
			ID: "people", Label: "People search (name + city)",
			Detail: fmt.Sprintf("Search directories for %s & %s near %s, %s. Treat hits as candidates only.", firstA, firstB, city, region),
			Status: "suggested",
			URL:    peopleURL,
		},
		{
			ID: "whitepages", Label: "Address / phone directories",
			Detail: "Cross-check whitepages-style results. Prefer matches that list both partners or shared surname.",
			Status: "suggested",
			URL:    whitepagesURL,
		},
		{
			ID: "property", Label: "Property / deed records",
			Detail: "County assessor / auditor under either first+last name in the market city.",
			Status: "suggested",
			URL:    firstNonEmpty(propertyURL, countyURL),
		},
		{
			ID: "marriage", Label: "Public marriage / engagement mentions",
			Detail: "Optional corroboration: license indexes, newspaper, registry pages.",
			Status: "suggested",
			URL:    marriageURL,
		},
		{
			ID: "verify", Label: "Human verify street address",
			Detail: "Enter line1 + ZIP, mark Address verified, then Ready to mail. Neptune never mails automatically.",
			Status: "suggested",
		},
	}

	notes = fmt.Sprintf(
		"Research brief for %s & %s\n\n"+
			"First names: %s (%s) & %s (%s)\n"+
			"Handles: @%s & @%s\n"+
			"Market: %s, %s (source: %s)\n\n"+
			"Postcard should greet: Dear %s & %s\n\n"+
			"What we have: first names, social handles, bios, discovery post, city-level location.\n"+
			"What we do NOT invent: street address, apartment number, or ZIP.\n\n"+
			"Research tools (open links in the kit UI):\n"+
			"1. People search with both first names + city\n"+
			"2. Whitepages / address directories\n"+
			"3. County property / deed under surnames once known\n"+
			"4. Optional marriage-license index for surname confirmation\n\n"+
			"Then paste verified street + ZIP → Verify address → Ready to mail.\n"+
			"Mail export is Lob/PostGrid shaped for later automation.",
		firstA, firstB,
		firstA, rnA.Source, firstB, rnB.Source,
		kit.HandleA, kit.HandleB,
		city, region, firstNonEmpty(source, "unknown"),
		firstA, firstB,
	)
	return cands, steps, notes, conf, source
}

func googleQ(q string) string {
	return "https://www.google.com/search?q=" + url.QueryEscape(q)
}

func countySearchURL(city, region string) string {
	return googleQ(fmt.Sprintf("%s %s county property search auditor assessor", city, region))
}

func postcardFallback(nameA, nameB, loc string) llm.Copy {
	where := ""
	if loc != "" {
		where = fmt.Sprintf(" in %s", loc)
	}
	a, b := firstNameOnly(nameA), firstNameOnly(nameB)
	return llm.Copy{
		InternalNote: fmt.Sprintf(
			"POSTCARD DRAFT\nCouple: %s & %s\nLocation signal: %s\nStatus: human review required before mail\nNo street address claimed by the agent.",
			a, b, firstNonEmpty(loc, "unknown"),
		),
		CustomerFacing: fmt.Sprintf(
			"Dear %s & %s,\n\nCongratulations on your engagement%s! May this season of planning be full of joy, good light, and the people you love most.\n\nWith warm regards,\nNeptune",
			a, b, where,
		),
	}
}

func firstNameOnly(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}
	// Already a first name, or "Alida Smith" → Alida
	parts := strings.Fields(s)
	if len(parts) == 0 {
		return s
	}
	// Strip @handle leftovers
	p := strings.TrimPrefix(parts[0], "@")
	if i := strings.IndexAny(p, "._"); i > 0 && !strings.Contains(p, " ") {
		// looks like handle fragment
		p = p[:i]
	}
	if len(p) > 0 {
		r := []rune(strings.ToLower(p))
		r[0] = []rune(strings.ToUpper(string(r[0])))[0]
		return string(r)
	}
	return s
}

// RenderPostcardHTML produces a print-friendly 6×4 preview (HTML/CSS).
// Images are rewritten through /api/media so Instagram CDN CORP does not blank them.
// Names on the card are first names (PersonAName/PersonBName already resolved).
func RenderPostcardHTML(k store.CongratulateKit) string {
	nameA := html.EscapeString(firstNonEmpty(firstNameOnly(k.PersonAName), prettyName(k.PersonAName, k.HandleA), "Partner"))
	nameB := html.EscapeString(firstNonEmpty(firstNameOnly(k.PersonBName), prettyName(k.PersonBName, k.HandleB), "Partner"))
	headline := html.EscapeString(firstNonEmpty(k.Headline, "Congratulations"))
	body := html.EscapeString(k.BodyMessage)
	body = strings.ReplaceAll(body, "\n", "<br/>")
	loc := html.EscapeString(strings.TrimSpace(k.MarketCity + ", " + k.MarketRegion))
	if strings.Trim(loc, ", ") == "" {
		loc = ""
	}
	addr := formatAddressHTML(k)
	rawImg := k.DiscoveryImageURL
	if rawImg == "" {
		rawImg = k.ProfilePicA
	}
	img := html.EscapeString(proxyMediaURL(rawImg))

	frontPhoto := `<div class="pc-front__photo pc-front__photo--empty">♥</div>`
	if img != "" {
		frontPhoto = fmt.Sprintf(`<div class="pc-front__photo" style="background-image:url('%s')"></div>`, img)
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html><head><meta charset="utf-8"/><title>Postcard — %s & %s</title>
<style>
  @page { size: 6in 4in; margin: 0; }
  * { box-sizing: border-box; }
  body { margin: 0; font-family: "Iowan Old Style", "Palatino Linotype", Palatino, Georgia, serif; background: #f4f1ea; color: #1c1917; }
  .pc-sheet { display: flex; flex-wrap: wrap; gap: 16px; padding: 16px; justify-content: center; }
  .pc-card { width: 6in; height: 4in; border-radius: 4px; overflow: hidden; box-shadow: 0 8px 28px rgba(28,25,23,.12); background: #fff; position: relative; }
  .pc-front { display: grid; grid-template-rows: 1fr auto; height: 100%%; background: linear-gradient(160deg, #1a3a3c 0%%, #0f2628 100%%); color: #faf7f2; }
  .pc-front__photo { background-size: cover; background-position: center; min-height: 0; }
  .pc-front__photo--empty { display:flex; align-items:center; justify-content:center; font-size: 64px; opacity: .5; background: #244a4c; }
  .pc-front__banner { padding: 14px 20px 18px; text-align: center; }
  .pc-front__headline { font-size: 28px; letter-spacing: .04em; font-weight: 500; margin: 0 0 4px; }
  .pc-front__pair { font-size: 15px; opacity: .9; margin: 0; font-style: italic; }
  .pc-front__loc { font-size: 11px; opacity: .65; margin: 6px 0 0; font-family: ui-monospace, monospace; letter-spacing: .06em; text-transform: uppercase; }
  .pc-back { display: grid; grid-template-columns: 1.15fr 0.85fr; height: 100%%; background: #fdfbf7; }
  .pc-back__message { padding: 22px 20px; font-size: 13px; line-height: 1.55; border-right: 1px dashed #d6d0c4; }
  .pc-back__message p { margin: 0 0 10px; white-space: normal; }
  .pc-back__from { margin-top: 16px; font-size: 12px; color: #57534e; }
  .pc-back__addr { padding: 22px 18px; display: flex; flex-direction: column; justify-content: space-between; }
  .pc-back__stamp { width: 48px; height: 56px; border: 1.5px dashed #a8a29e; align-self: flex-end; border-radius: 2px; opacity: .5; }
  .pc-back__to { font-size: 13px; line-height: 1.45; margin-top: 24px; }
  .pc-back__to strong { display: block; margin-bottom: 4px; }
  .pc-label { font-size: 10px; text-transform: uppercase; letter-spacing: .08em; color: #78716c; font-family: ui-monospace, monospace; margin-bottom: 8px; }
  @media print { body { background: #fff; } .pc-sheet { padding: 0; gap: 0; } .pc-card { box-shadow: none; page-break-after: always; } }
</style></head><body>
<div class="pc-sheet">
  <div class="pc-card pc-front">
    %s
    <div class="pc-front__banner">
      <h1 class="pc-front__headline">%s</h1>
      <p class="pc-front__pair">%s &amp; %s</p>
      %s
    </div>
  </div>
  <div class="pc-card pc-back">
    <div class="pc-back__message">
      <div class="pc-label">Message</div>
      <p>%s</p>
      <div class="pc-back__from">Neptune · with care</div>
    </div>
    <div class="pc-back__addr">
      <div class="pc-back__stamp" title="Stamp area"></div>
      <div class="pc-back__to">
        <div class="pc-label">To</div>
        <strong>%s &amp; %s</strong>
        %s
      </div>
    </div>
  </div>
</div>
</body></html>`,
		nameA, nameB,
		frontPhoto,
		headline, nameA, nameB,
		func() string {
			if loc == "" || loc == ", " {
				return ""
			}
			return `<p class="pc-front__loc">` + loc + `</p>`
		}(),
		body, nameA, nameB, addr,
	)
}

func formatAddressHTML(k store.CongratulateKit) string {
	if k.AddressLine1 == "" && k.AddressCity == "" {
		return `<span style="color:#a8a29e;font-style:italic">Address pending human verification</span>`
	}
	var parts []string
	if k.AddressLine1 != "" {
		parts = append(parts, html.EscapeString(k.AddressLine1))
	}
	if k.AddressLine2 != "" {
		parts = append(parts, html.EscapeString(k.AddressLine2))
	}
	line := strings.TrimSpace(fmt.Sprintf("%s, %s %s", k.AddressCity, k.AddressRegion, k.AddressPostal))
	line = strings.Trim(line, ", ")
	if line != "" {
		parts = append(parts, html.EscapeString(line))
	}
	if k.AddressCountry != "" && k.AddressCountry != "US" {
		parts = append(parts, html.EscapeString(k.AddressCountry))
	}
	return strings.Join(parts, "<br/>")
}

func mailPayload(k store.CongratulateKit) map[string]any {
	return map[string]any{
		"description": fmt.Sprintf("Neptune congratulate postcard — %s & %s", k.PersonAName, k.PersonBName),
		"to": map[string]any{
			"name":            strings.TrimSpace(k.PersonAName + " & " + k.PersonBName),
			"address_line1":   k.AddressLine1,
			"address_line2":   k.AddressLine2,
			"address_city":    k.AddressCity,
			"address_state":   k.AddressRegion,
			"address_zip":     k.AddressPostal,
			"address_country": firstNonEmpty(k.AddressCountry, "US"),
		},
		"from": map[string]any{
			"name": "Neptune",
			// Operator fills return address at mail time.
		},
		"front":        "see postcard_html front panel",
		"back":         k.BodyMessage,
		"size":         "4x6",
		"metadata":     map[string]string{"couple_id": k.CoupleID, "kit_id": k.ID},
		"export_ready": k.AddressLine1 != "" && k.AddressPostal != "" && (k.Status == "ready_to_mail" || k.Status == "address_verified"),
		"provider_hint": "Lob, PostGrid, or print-at-home via postcard_html",
		"generated_at": time.Now().UTC().Format(time.RFC3339),
	}
}

// UpdateKitAddress applies human-edited mailing fields and re-renders preview.
func (a *Agent) UpdateKitAddress(kitID string, patch store.CongratulateKit, verified bool, verifiedBy string) (store.CongratulateKit, error) {
	k, err := a.Store.GetCongratulateKit(kitID)
	if err != nil {
		return k, err
	}
	if patch.AddressLine1 != "" {
		k.AddressLine1 = patch.AddressLine1
	}
	if patch.AddressLine2 != "" || patch.AddressLine1 != "" {
		k.AddressLine2 = patch.AddressLine2
	}
	if patch.AddressCity != "" {
		k.AddressCity = patch.AddressCity
	}
	if patch.AddressRegion != "" {
		k.AddressRegion = patch.AddressRegion
	}
	if patch.AddressPostal != "" {
		k.AddressPostal = patch.AddressPostal
	}
	if patch.AddressCountry != "" {
		k.AddressCountry = patch.AddressCountry
	}
	if patch.Headline != "" {
		k.Headline = patch.Headline
	}
	if patch.BodyMessage != "" {
		k.BodyMessage = patch.BodyMessage
	}
	if patch.FirstNameA != "" {
		k.FirstNameA = patch.FirstNameA
		k.PersonAName = patch.FirstNameA
	}
	if patch.LastNameA != "" {
		k.LastNameA = patch.LastNameA
	}
	if patch.FirstNameB != "" {
		k.FirstNameB = patch.FirstNameB
		k.PersonBName = patch.FirstNameB
	}
	if patch.LastNameB != "" {
		k.LastNameB = patch.LastNameB
	}
	if patch.NameSourceA != "" {
		k.NameSourceA = patch.NameSourceA
	}
	if patch.NameSourceB != "" {
		k.NameSourceB = patch.NameSourceB
	}
	if verified {
		if k.AddressLine1 == "" || k.AddressCity == "" || k.AddressPostal == "" {
			return k, fmt.Errorf("street, city, and postal code required to verify")
		}
		k.Status = "address_verified"
		k.AddressSource = "human_verified"
		k.AddressConfidence = 0.95
		now := time.Now().UTC()
		k.VerifiedAt = &now
		k.VerifiedBy = firstNonEmpty(verifiedBy, "operator")
	}
	k.PostcardHTML = RenderPostcardHTML(k)
	k.MailPayload = mailPayload(k)
	return a.Store.UpsertCongratulateKit(k)
}

// MarkReadyToMail requires address_verified first.
func (a *Agent) MarkReadyToMail(kitID string) (store.CongratulateKit, error) {
	k, err := a.Store.GetCongratulateKit(kitID)
	if err != nil {
		return k, err
	}
	if k.Status != "address_verified" && k.Status != "ready_to_mail" {
		return k, fmt.Errorf("verify address before marking ready to mail (status=%s)", k.Status)
	}
	if k.AddressLine1 == "" || k.AddressPostal == "" {
		return k, fmt.Errorf("complete street address required")
	}
	k.Status = "ready_to_mail"
	k.MailPayload = mailPayload(k)
	k.PostcardHTML = RenderPostcardHTML(k)
	return a.Store.UpsertCongratulateKit(k)
}

// MarkMailed records that a human completed the physical mail step.
func (a *Agent) MarkMailed(kitID string) (store.CongratulateKit, error) {
	k, err := a.Store.GetCongratulateKit(kitID)
	if err != nil {
		return k, err
	}
	if k.Status != "ready_to_mail" && k.Status != "mailed" {
		return k, fmt.Errorf("kit must be ready_to_mail (status=%s)", k.Status)
	}
	now := time.Now().UTC()
	k.MailedAt = &now
	k.Status = "mailed"
	return a.Store.UpsertCongratulateKit(k)
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func prettyName(label, handle string) string {
	label = strings.TrimSpace(label)
	if label == "" || strings.EqualFold(label, handle) {
		if handle == "" {
			return ""
		}
		// soft-prettify handle: jane_doe → Jane
		h := strings.TrimPrefix(handle, "@")
		parts := strings.FieldsFunc(h, func(r rune) bool {
			return r == '_' || r == '.' || r == '-'
		})
		if len(parts) == 0 {
			return h
		}
		p := parts[0]
		if len(p) == 0 {
			return h
		}
		return strings.ToUpper(p[:1]) + strings.ToLower(p[1:])
	}
	return label
}

func looksLikeHandleLabel(label, handle string) bool {
	label = strings.TrimSpace(label)
	if label == "" {
		return true
	}
	if strings.EqualFold(label, handle) {
		return true
	}
	// vendor-ish compound handles used as display names
	if strings.Contains(label, "_") || strings.Count(label, ".") >= 1 {
		return true
	}
	if len(label) > 22 && !strings.Contains(label, " ") {
		return true
	}
	return false
}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}

func proxyMediaURL(u string) string {
	u = strings.TrimSpace(u)
	if u == "" {
		return ""
	}
	if strings.HasPrefix(u, "/api/media") {
		return u
	}
	// Relative proxy so the print view works same-origin.
	return "/api/media?url=" + url.QueryEscape(u)
}
