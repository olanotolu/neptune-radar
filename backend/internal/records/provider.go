// Package records finds mailing-address candidates from name + market signals.
// Implementations: Trestle IQ (primary), People Data Labs, Cleanlist, Melissa,
// Google CSE, TruePeopleSearch, county property, voter, deterministic heuristic.
package records

import (
	"context"
	"os"
	"regexp"
	"sort"
	"strings"
)

// Candidate kinds — research URLs must never live in Line1 (that false-triggers street hits).
const (
	KindStreet       = "street"
	KindLocality     = "locality"
	KindResearchLink = "research_link"
)

// Query is a people-search request built from a congratulate kit dossier.
type Query struct {
	FirstName string
	LastName  string
	// Optional second person (partner) — improves household ranking.
	PartnerFirst string
	PartnerLast  string
	City         string
	Region       string // state abbrev
	// Social handles for providers that accept them (PDL).
	Handle string

	// --- Enriched location signals (used by heuristic to boost confidence) ---
	PostLocation   string // Instagram venue tag, e.g. "The Joseph Hotel, Columbus OH"
	VendorCity     string // photographer/vendor city from watched_sources
	VendorState    string // photographer/vendor state from watched_sources
	AccountCityA   string // person A's individually-inferred city from their bio
	AccountRegionA string // person A's state
	AccountCityB   string // person B's individually-inferred city from their bio
	AccountRegionB string // person B's state
	BioA           string // person A bio text (heuristic can check for location mentions)
	BioB           string // person B bio text
}

// Candidate is one possible mailing address — never mailed until human confirms.
type Candidate struct {
	Line1      string  `json:"line1,omitempty"`
	Line2      string  `json:"line2,omitempty"`
	City       string  `json:"city,omitempty"`
	Region     string  `json:"region,omitempty"`
	Postal     string  `json:"postal,omitempty"`
	Country    string  `json:"country,omitempty"`
	Confidence float64 `json:"confidence"`
	Source     string  `json:"source"`
	Note       string  `json:"note,omitempty"`
	// Kind: street | locality | research_link. Empty = inferred from fields.
	Kind string `json:"kind,omitempty"`
	// URL is for research_link candidates only — never put http URLs in Line1.
	URL string `json:"url,omitempty"`
	// Optional identity hints from the provider
	FullName string `json:"full_name,omitempty"`
	Phone    string `json:"phone,omitempty"`
	// Property asset data from county auditor HTML (internal operator use only)
	Asset PropertyAsset `json:"asset,omitempty"`
}

// Result is one provider call outcome for audit storage.
type Result struct {
	Provider   string
	Candidates []Candidate
	RawJSON    string
	Status     string // ok | empty | error
	Error      string
	CostCents  int
	// PaidCalls is how many paid people-data providers were invoked this Search.
	PaidCalls int
}

// Provider searches for people/addresses.
type Provider interface {
	Name() string
	Available() bool
	Search(ctx context.Context, q Query) (Result, error)
}

// NewProvider picks the best configured people-search backend.
// Order: Trestle → PDL → Melissa → Cleanlist → Google Search → Heuristic.
func NewProvider() Provider {
	if k := strings.TrimSpace(os.Getenv("TRESTLE_API_KEY")); k != "" {
		return &Trestle{APIKey: k}
	}
	if k := strings.TrimSpace(os.Getenv("PDL_API_KEY")); k != "" {
		return &PDL{APIKey: k}
	}
	if k := strings.TrimSpace(os.Getenv("MELISSA_LICENSE_KEY")); k != "" {
		return &Melissa{LicenseKey: k}
	}
	if k := strings.TrimSpace(os.Getenv("CLEANLIST_API_KEY")); k != "" {
		return &Cleanlist{APIKey: k}
	}
	if k := strings.TrimSpace(os.Getenv("GOOGLE_SEARCH_API_KEY")); k != "" {
		cx := strings.TrimSpace(os.Getenv("GOOGLE_SEARCH_CX"))
		if cx != "" {
			return &GoogleSearch{APIKey: k, Cx: cx}
		}
	}
	return &Heuristic{}
}

// Multi merges paid → free public records → heuristic so UI always gets something useful.
// Unlike first-street-wins, it collects candidates across providers (budget-aware early stop
// only when a high-confidence real street is already found).
type Multi struct {
	Primary  Provider
	Paid     []Provider // additional paid providers (secondary, melissa, etc.)
	Free     []Provider // free public-records providers, tried in order
	Fallback Provider

	// MaxPaidCalls caps paid provider invocations for this Search (0 = unlimited).
	// Detective fan-out sets this from DETECTIVE_PAID_CAP remaining budget.
	MaxPaidCalls int
	// SkipFree skips Google/TPS/property/voter (paid-only fan-out pass).
	SkipFree bool
	// SkipFallback skips the OH zip heuristic.
	SkipFallback bool
	// PrimaryOnlyPaid: on fan-out, run only the primary paid provider (stretch budget
	// across more name×loc pairs). Full free pass sets false to merge all paid.
	PrimaryOnlyPaid bool

	// Accuracy is the Bayesian provider accuracy map (provider→state→rate).
	// When non-nil, finish() re-ranks candidates via FuseCandidates.
	// Set by the detective from the store; nil = cold-start 0.5 prior for all.
	Accuracy map[string]map[string]float64
}

func NewMulti() *Multi {
	p := NewProvider()
	// Additional paid providers that aren't the primary (each runs if available).
	var paid []Provider
	addPaid := func(pr Provider) {
		if pr == nil || !pr.Available() {
			return
		}
		if p != nil && pr.Name() == p.Name() {
			return
		}
		for _, existing := range paid {
			if existing.Name() == pr.Name() {
				return
			}
		}
		paid = append(paid, pr)
	}
	if k := strings.TrimSpace(os.Getenv("TRESTLE_API_KEY")); k != "" {
		addPaid(&Trestle{APIKey: k})
	}
	if k := strings.TrimSpace(os.Getenv("PDL_API_KEY")); k != "" {
		addPaid(&PDL{APIKey: k})
	}
	if k := strings.TrimSpace(os.Getenv("MELISSA_LICENSE_KEY")); k != "" {
		addPaid(&Melissa{LicenseKey: k})
	}
	if k := strings.TrimSpace(os.Getenv("CLEANLIST_API_KEY")); k != "" {
		addPaid(&Cleanlist{APIKey: k})
	}
	// If primary is one of the paid list, remove it from Paid to avoid double-call.
	if p != nil {
		filtered := paid[:0]
		for _, pr := range paid {
			if pr.Name() != p.Name() {
				filtered = append(filtered, pr)
			}
		}
		paid = filtered
	}

	// Free / pay-as-you-go public-records hunters — always after paid, before heuristic.
	var free []Provider
	// Apify TPS: only if explicitly enabled (APIFY_TPS_ENABLED=true) — default PAUSED to save credits.
	if ap := apifyTPSFromEnv(); ap.Available() {
		free = append(free, ap)
	}
	// Free SERP always on (DuckDuckGo HTML) — research links + rare snippet streets
	free = append(free, &DDGSerp{})
	if k := strings.TrimSpace(os.Getenv("GOOGLE_SEARCH_API_KEY")); k != "" {
		cx := strings.TrimSpace(os.Getenv("GOOGLE_SEARCH_CX"))
		if cx != "" {
			// Google only in free tier when not already primary (no paid keys case).
			if _, isGoogle := p.(*GoogleSearch); !isGoogle {
				free = append(free, &GoogleSearch{APIKey: k, Cx: cx})
			}
		}
	}
	if tps := tpsFromEnv(); tps.Available() {
		free = append(free, tps)
	}
	free = append(free, &PropertyRecords{})
	free = append(free, &VoterRegistration{})
	return &Multi{Primary: p, Paid: paid, Free: free, Fallback: &Heuristic{}}
}

func (m *Multi) Name() string {
	if m.Primary != nil && m.Primary.Available() {
		return m.Primary.Name()
	}
	if len(m.Paid) > 0 && m.Paid[0].Available() {
		return m.Paid[0].Name()
	}
	if len(m.Free) > 0 && m.Free[0].Available() {
		return m.Free[0].Name()
	}
	return "heuristic"
}

func (m *Multi) Available() bool { return true }

// highConfStreet is confidence at which we stop burning more provider budget.
const highConfStreet = 0.85

func (m *Multi) Search(ctx context.Context, q Query) (Result, error) {
	var (
		all       []Candidate
		providers []string
		cost      int
		paidCalls int
		lastErr   error
		rawParts  []string
	)

	paidBudgetLeft := func() bool {
		if m.MaxPaidCalls <= 0 {
			return true // unlimited
		}
		return paidCalls < m.MaxPaidCalls
	}

	run := func(pr Provider, countAsPaid bool) {
		if pr == nil || !pr.Available() {
			return
		}
		if countAsPaid && !paidBudgetLeft() {
			return
		}
		res, err := pr.Search(ctx, q)
		providers = append(providers, pr.Name())
		cost += res.CostCents
		if countAsPaid {
			paidCalls++
		}
		if res.RawJSON != "" {
			rawParts = append(rawParts, pr.Name()+":"+truncate(res.RawJSON, 500))
		}
		if err != nil {
			lastErr = err
		}
		all = append(all, normalizeCandidates(res.Candidates)...)
	}

	// Tier 1: primary (paid API or heuristic when no keys)
	if m.Primary != nil {
		run(m.Primary, isPaidProvider(m.Primary))
	}
	// Prefer one paid provider per Search when PrimaryFanout is set (budget stretch).
	// Only cascade additional paid if no street yet.
	if maxStreetConf(all) >= highConfStreet {
		return m.finish(all, providers, cost, paidCalls, rawParts, nil)
	}

	// Tier 2: other paid providers — only if primary returned no real street
	// (or PrimaryFanout is false for full multi-key merge on free pass).
	needMorePaid := !hasStreetCandidates(all) || !m.PrimaryOnlyPaid
	if needMorePaid {
		for _, pr := range m.Paid {
			if m.PrimaryOnlyPaid && hasStreetCandidates(all) {
				break
			}
			run(pr, true)
			if maxStreetConf(all) >= highConfStreet {
				return m.finish(all, providers, cost, paidCalls, rawParts, nil)
			}
			// Fan-out mode: stop after first paid that yields any street
			if m.PrimaryOnlyPaid && hasStreetCandidates(all) {
				break
			}
		}
	}

	// Tier 3: free public-records
	if !m.SkipFree {
		for _, fp := range m.Free {
			run(fp, false)
			if maxStreetConf(all) >= highConfStreet {
				return m.finish(all, providers, cost, paidCalls, rawParts, nil)
			}
		}
	}

	// Tier 4: heuristic locality/zip
	if !m.SkipFallback {
		needFallback := m.Fallback != nil && !hasStreetCandidates(all)
		if needFallback {
			primaryIsFallback := m.Primary != nil && m.Fallback != nil && m.Primary.Name() == m.Fallback.Name()
			if !primaryIsFallback {
				run(m.Fallback, false)
			} else if len(all) == 0 {
				// primary already ran as heuristic and returned nothing — nothing more to do
			}
		}
	}

	return m.finish(all, providers, cost, paidCalls, rawParts, lastErr)
}

func isPaidProvider(pr Provider) bool {
	if pr == nil {
		return false
	}
	switch pr.Name() {
	case "trestle", "pdl", "melissa", "cleanlist":
		return true
	default:
		return false
	}
}

func (m *Multi) finish(all []Candidate, providers []string, cost, paidCalls int, rawParts []string, lastErr error) (Result, error) {
	all = dedupeCandidates(all)
	rankCandidates(all)
	if m.Accuracy != nil {
		all = FuseCandidates(all, m.Accuracy)
	}
	name := strings.Join(providers, "+")
	if name == "" {
		name = "multi"
	}
	st := "ok"
	if len(all) == 0 {
		st = "empty"
		if lastErr != nil {
			st = "error"
			return Result{Provider: name, Status: st, Error: lastErr.Error(), CostCents: cost,
				PaidCalls: paidCalls, RawJSON: strings.Join(rawParts, "\n---\n")}, lastErr
		}
	}
	return Result{
		Provider:   name,
		Candidates: all,
		Status:     st,
		RawJSON:    strings.Join(rawParts, "\n---\n"),
		CostCents:  cost,
		PaidCalls:  paidCalls,
	}, nil
}

// HasPaidProviders reports whether Multi would invoke any paid people-data API.
func (m *Multi) HasPaidProviders() bool {
	if m.Primary != nil && m.Primary.Available() && isPaidProvider(m.Primary) {
		return true
	}
	for _, p := range m.Paid {
		if p != nil && p.Available() {
			return true
		}
	}
	return false
}

// ResearchLink builds a non-street candidate that points the operator at a URL.
// Line1 is intentionally empty so hasStreetCandidates never fires.
func ResearchLink(url, city, region, fullName, source, note string) Candidate {
	if note == "" {
		note = "Operator research link — visit and paste street if found."
	}
	if url != "" && !strings.Contains(note, url) {
		note = strings.TrimSpace(note + " " + url)
	}
	return Candidate{
		Kind:       KindResearchLink,
		URL:        url,
		City:       city,
		Region:     region,
		Country:    "US",
		Confidence: 0.30,
		Source:     source,
		FullName:   fullName,
		Note:       note,
	}
}

// IsRealStreet reports whether Line1 looks like a US street address (not a URL or empty).
// Requires a street number and a recognized street-type suffix (St, Ave, Rd, …).
func IsRealStreet(line1 string) bool {
	line1 = strings.TrimSpace(line1)
	if line1 == "" {
		return false
	}
	lower := strings.ToLower(line1)
	if strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") {
		return false
	}
	if strings.Contains(lower, "://") {
		return false
	}
	if !realStreetRe.MatchString(line1) {
		return false
	}
	return HasStreetType(line1)
}

// realStreetRe: number + at least one name token.
var realStreetRe = regexp.MustCompile(`(?i)^\d{1,6}\s+[A-Za-z0-9.'\-]+`)

// hasStreetCandidates returns true if any candidate is a real mailing street.
func hasStreetCandidates(cands []Candidate) bool {
	for _, c := range cands {
		if candidateIsStreet(c) {
			return true
		}
	}
	return false
}

// HasConfirmableStreet is true when we have a street good enough to skip free scrapers:
// real street type + city + (zip or conf ≥ 0.65).
func HasConfirmableStreet(cands []Candidate) bool {
	for _, c := range cands {
		if !candidateIsStreet(c) {
			continue
		}
		if c.City == "" {
			continue
		}
		if c.Postal != "" || c.Confidence >= 0.65 {
			return true
		}
	}
	return false
}

func candidateIsStreet(c Candidate) bool {
	if c.Kind == KindResearchLink {
		return false
	}
	if c.Kind == KindLocality {
		return false
	}
	return IsRealStreet(c.Line1)
}

func maxStreetConf(cands []Candidate) float64 {
	max := 0.0
	for _, c := range cands {
		if candidateIsStreet(c) && c.Confidence > max {
			max = c.Confidence
		}
	}
	return max
}

// normalizeCandidates sets Kind and strips accidental URLs from Line1.
func normalizeCandidates(cands []Candidate) []Candidate {
	out := make([]Candidate, 0, len(cands))
	for _, c := range cands {
		c = normalizeOne(c)
		out = append(out, c)
	}
	return out
}

func normalizeOne(c Candidate) Candidate {
	line := strings.TrimSpace(c.Line1)
	// Rescue mis-tagged research URLs that older code put in Line1.
	if c.Kind == KindResearchLink || isHTTPURL(line) {
		if c.URL == "" && isHTTPURL(line) {
			c.URL = line
		}
		c.Line1 = ""
		c.Kind = KindResearchLink
		if c.Confidence <= 0 {
			c.Confidence = 0.30
		}
		return c
	}
	if IsRealStreet(line) {
		if c.Kind == "" {
			c.Kind = KindStreet
		}
		return c
	}
	if line != "" {
		// Non-street garbage in Line1 (e.g. free text) — demote
		c.Line1 = ""
		if c.Note == "" {
			c.Note = "Dropped non-street Line1: " + truncate(line, 80)
		}
	}
	if c.Kind == "" {
		if c.City != "" || c.Postal != "" {
			c.Kind = KindLocality
		}
	}
	return c
}

func isHTTPURL(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}

func dedupeCandidates(cands []Candidate) []Candidate {
	seen := map[string]int{} // key → index of best
	var out []Candidate
	for _, c := range cands {
		key := candidateKey(c)
		if i, ok := seen[key]; ok {
			prev := out[i]
			// Merge: take higher conf, union sources, fill missing postal/line2
			merged := prev
			if c.Confidence > prev.Confidence {
				merged = c
				// keep better postal/line2 from either
				if merged.Postal == "" {
					merged.Postal = prev.Postal
				}
				if merged.Line2 == "" {
					merged.Line2 = prev.Line2
				}
			} else {
				if merged.Postal == "" {
					merged.Postal = c.Postal
				}
				if merged.Line2 == "" {
					merged.Line2 = c.Line2
				}
			}
			// Multi-source corroboration boost
			srcA, srcB := prev.Source, c.Source
			if srcA != "" && srcB != "" && !strings.Contains(srcA, srcB) && !strings.Contains(srcB, srcA) {
				// Count independent sources in merged source string
				nSources := 1 + strings.Count(merged.Source, "+")
				if !strings.Contains(merged.Source, srcB) && srcB != merged.Source {
					merged.Source = strings.Trim(merged.Source+"+"+srcB, "+")
					nSources++
				}
				// +0.08 per additional independent source, cap +0.24
				extra := float64(nSources-1) * 0.08
				if extra > 0.24 {
					extra = 0.24
				}
				base := merged.Confidence
				if prev.Confidence > base {
					base = prev.Confidence
				}
				if c.Confidence > base {
					base = c.Confidence
				}
				merged.Confidence = base + extra
				if merged.Confidence > 0.92 {
					merged.Confidence = 0.92
				}
				merged.Note = strings.TrimSpace(merged.Note + " · multi-source (" + merged.Source + ")")
			}
			out[i] = merged
			continue
		}
		seen[key] = len(out)
		out = append(out, c)
	}
	return out
}

func candidateKey(c Candidate) string {
	if c.Kind == KindResearchLink || c.URL != "" {
		return "url:" + strings.ToLower(c.URL)
	}
	if IsRealStreet(c.Line1) {
		// Postal-agnostic so "123 Main St" and "123 Main St 43215" merge
		return "st:" + normalizeStreetKey(c.Line1) + "|" +
			strings.ToLower(strings.TrimSpace(c.City)) + "|" + strings.ToUpper(strings.TrimSpace(c.Region))
	}
	return "loc:" + strings.ToLower(c.City) + "|" + strings.ToUpper(c.Region) + "|" + c.Source
}

func normalizeStreetKey(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, ".", "")
	s = strings.Join(strings.Fields(s), " ")
	// Normalize common suffixes
	repl := []struct{ a, b string }{
		{" street", " st"}, {" avenue", " ave"}, {" road", " rd"},
		{" drive", " dr"}, {" lane", " ln"}, {" court", " ct"},
		{" boulevard", " blvd"}, {" place", " pl"}, {" circle", " cir"},
	}
	for _, r := range repl {
		if strings.HasSuffix(s, r.a) {
			s = strings.TrimSuffix(s, r.a) + r.b
		}
	}
	return s
}

// rankCandidates: streets first (by conf), then locality, then research links.
// Cap pure locality confidence so zip-only never outranks a real street.
func rankCandidates(cands []Candidate) {
	for i := range cands {
		if !candidateIsStreet(cands[i]) && cands[i].Kind != KindResearchLink && cands[i].URL == "" {
			if cands[i].Kind == "" {
				cands[i].Kind = KindLocality
			}
			if cands[i].Confidence > 0.49 {
				cands[i].Confidence = 0.49
			}
		}
	}
	rank := func(c Candidate) int {
		switch {
		case candidateIsStreet(c):
			return 0
		case c.Kind == KindResearchLink || c.URL != "":
			return 2
		default:
			return 1
		}
	}
	sort.SliceStable(cands, func(i, j int) bool {
		ri, rj := rank(cands[i]), rank(cands[j])
		if ri != rj {
			return ri < rj
		}
		// Prefer multi-source
		si := strings.Count(cands[i].Source, "+")
		sj := strings.Count(cands[j].Source, "+")
		if si != sj {
			return si > sj
		}
		return cands[i].Confidence > cands[j].Confidence
	})
}
