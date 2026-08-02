package packs

// State packs for the Neptune Radar source registry (Phase 2: national
// expansion). Each pack contains the three source layers for one state:
//
//   - Government: county marriage-record offices with real, verified search
//     URLs + http_health connectors.
//   - Church: Catholic dioceses/archdioceses covering the state, with parish
//     directory endpoints; parishes (with bulletin archives where located) for
//     the primary-metro diocese.
//   - Social: wedding-industry Instagram vendors, handles verified from each
//     business's own website.
//
// Every URL below was verified against a real public source before being
// written — see the comment above each block for exactly where it came from.
// Nothing here is invented or guessed. SourceClass values must match
// signals.WatchedSourceClasses.

// --- Types ------------------------------------------------------------------

type CityDef struct {
	ID      string
	State   string
	County  string // 5-digit FIPS
	Name    string
	Lat     float64
	Lng     float64
	Markets []string // ACTIVE_MARKETS-style tokens for hashtag generation
}

type VendorDef struct {
	Name         string
	OfficialURL  string
	Handle       string // no "@"
	SourceClass  string
	CityID       string
	State        string
	City         string
	TikTokHandle string // optional: TikTok handle if the vendor has one
	KnotURL      string // optional: The Knot profile URL if the vendor has one
	// Verified is ISO date the official site was checked for this handle.
	Verified string
}

// GovSource is one verified county marriage-record office. Every URL and
// capability note comes from the source review — the note states honestly
// what the endpoint can do and what still needs testing. FIPS ids match
// geo.Counties.
type GovSource struct {
	CountyFIPS string
	CourtName  string
	CourtURL   string // official office/site URL
	SearchURL  string // the actual search/index page
	Note       string // honest capability note
}

// DioceseDef is a Catholic diocese or archdiocese covering part of a state.
// For non-Catholic denominations (episcopal, methodist), State identifies the
// primary state the conference belongs to so it can be filtered per-state.
type DioceseDef struct {
	Slug         string // URL-safe slug, unique within the state
	Name         string
	Type         string // "diocese" | "archdiocese" | "annual_conference"
	Website      string
	Directory    string // parish-finder / directory URL
	HubCityID    string // optional: cathedral city ID
	Denomination string // "catholic" (default when empty) | "episcopal" | "methodist"
	State        string // optional: for non-Catholic denominations, the primary state
}

// ParishDef is a real parish in a diocese. BulletinURL is set only where a
// real, reachable bulletin archive was located; Aggregator marks third-party
// listings (Parishes Online / Discover Mass) rather than the parish's own site.
type ParishDef struct {
	DioceseSlug  string
	Name         string
	Address      string
	BulletinURL  string
	Aggregator   bool
	CityID       string  // optional: parish city; empty = fall back to diocese HubCityID
	CountyFIPS   string  // optional: 5-digit FIPS; empty = leave CountyID unset
	Denomination string  // "catholic" (default when empty) | "episcopal" | "methodist"
	GeoLat       float64 // optional: parish latitude for geographic accuracy
	GeoLng       float64 // optional: parish longitude for geographic accuracy
}

// StatePack bundles all three source layers for one state.
type StatePack struct {
	State      string
	Cities     []CityDef
	Government []GovSource
	Dioceses   []DioceseDef
	Parishes   []ParishDef
	Vendors    []VendorDef
}

// packFor returns the state pack for a USPS code, or nil if not yet defined.
func PackFor(st string) *StatePack {
	switch st {
	case "NY":
		return &nyPack
	case "CA":
		return &caPack
	case "TX":
		return &txPack
	case "MI":
		return &miPack
	case "IL":
		return &ilPack
	case "OH":
		return &ohPack
	case "FL":
		return &flPack
	case "PA":
		return &paPack
	case "GA":
		return &gaPack
	case "NC":
		return &ncPack
	case "NJ":
		return &njPack
	case "VA":
		return &vaPack
	case "WA":
		return &waPack
	case "MA":
		return &maPack
	case "CO":
		return &coPack
	case "AZ":
		return &azPack
	case "TN":
		return &tnPack
	case "IN":
		return &inPack
	case "MO":
		return &moPack
	case "WI":
		return &wiPack
	case "MN":
		return &mnPack
	case "MD":
		return &mdPack
	case "CT":
		return &ctPack
	case "OR":
		return &orPack
	case "NV":
		return &nvPack
	case "KY":
		return &kyPack
	case "LA":
		return &laPack
	case "AL":
		return &alPack
	case "SC":
		return &scPack
	case "OK":
		return &okPack
	case "UT":
		return &utPack
	case "IA":
		return &iaPack
	case "KS":
		return &ksPack
	case "NE":
		return &nePack
	case "AR":
		return &arPack
	case "MS":
		return &msPack
	case "NM":
		return &nmPack
	case "ID":
		return &idPack
	case "MT":
		return &mtPack
	case "WY":
		return &wyPack
	case "AK":
		return &akPack
	case "NH":
		return &nhPack
	case "ME":
		return &mePack
	case "VT":
		return &vtPack
	case "RI":
		return &riPack
	case "DE":
		return &dePack
	case "HI":
		return &hiPack
	case "DC":
		return &dcPack
	case "ND":
		return &ndPack
	case "SD":
		return &sdPack
	case "WV":
		return &wvPack
	default:
		return nil
	}
}

// --- New York ---------------------------------------------------------------
// Defined in pack_ny.go.

// --- California -------------------------------------------------------------
// Defined in pack_ca.go.

// --- Texas ------------------------------------------------------------------
// Defined in pack_tx.go.

// marketsForStates returns ACTIVE_MARKETS tokens to merge into server env docs.
func MarketsForStates(states []string) []string {
	var out []string
	seen := map[string]bool{}
	for _, st := range states {
		p := PackFor(st)
		if p == nil {
			continue
		}
		for _, c := range p.Cities {
			for _, m := range c.Markets {
				if !seen[m] {
					seen[m] = true
					out = append(out, m)
				}
			}
		}
	}
	return out
}
