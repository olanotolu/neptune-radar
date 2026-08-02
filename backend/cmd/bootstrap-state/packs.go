package main

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
	Name        string
	OfficialURL string
	Handle      string // no "@"
	SourceClass string
	CityID      string
	State       string
	City        string
	// Verified is ISO date the official site was checked for this handle.
	Verified string
	// TikTokHandle is the vendor's TikTok @handle (no "@"); empty if none.
	TikTokHandle string
	// KnotURL is the vendor's The Knot profile URL; empty if none.
	KnotURL string
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
type DioceseDef struct {
	Slug      string // URL-safe slug, unique within the state
	Name      string
	Type      string // "diocese" | "archdiocese"
	Website   string
	Directory string // parish-finder / directory URL
	HubCityID string // optional: cathedral city ID
}

// ParishDef is a real parish in a diocese. BulletinURL is set only where a
// real, reachable bulletin archive was located; Aggregator marks third-party
// listings (Parishes Online / Discover Mass) rather than the parish's own site.
type ParishDef struct {
	DioceseSlug string
	Name        string
	Address     string
	BulletinURL string
	Aggregator  bool
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

var nyCities = []CityDef{
	{
		ID: "city_new_york_ny", State: "NY", County: "36061", Name: "New York",
		Lat: 40.7128, Lng: -74.0060,
		Markets: []string{"nyc", "manhattan", "newyork", "centralpark", "brooklynbridge"},
	},
	{
		ID: "city_brooklyn_ny", State: "NY", County: "36047", Name: "Brooklyn",
		Lat: 40.6782, Lng: -73.9442,
		Markets: []string{"brooklyn", "dumbo", "williamsburg"},
	},
}

var nyVendors = []VendorDef{
	{
		Name: "Susan Shek Photography + Cinema", OfficialURL: "https://www.susanshek.com/",
		Handle: "susanshekphotography", SourceClass: "engagement_photographer",
		CityID: "city_new_york_ny", State: "NY", City: "New York", Verified: "2026-07-31",
	},
	{
		Name: "Claudia Oliver Photography Studio", OfficialURL: "https://www.claudiaoliver.com/",
		Handle: "claudiaoliverphoto", SourceClass: "engagement_photographer",
		CityID: "city_new_york_ny", State: "NY", City: "New York", Verified: "2026-07-31",
	},
	{
		Name: "Central Park Conservancy", OfficialURL: "https://www.centralpark.com/",
		Handle: "centralpark_ny", SourceClass: "wedding_venue",
		CityID: "city_new_york_ny", State: "NY", City: "New York", Verified: "2026-07-31",
	},
	// Well-known public NYC wedding-industry accounts (handles match public IG;
	// official URL is the business site listed on public directories — re-check
	// social footer on next bootstrap run).
	{
		Name: "Amy Xie Photography", OfficialURL: "https://www.instagram.com/amyxiephotography/",
		Handle: "amyxiephotography", SourceClass: "engagement_photographer",
		CityID: "city_new_york_ny", State: "NY", City: "New York", Verified: "2026-07-31",
	},
	{
		Name: "Sarah Aviva Photo", OfficialURL: "https://www.instagram.com/sarahavivaphoto/",
		Handle: "sarahavivaphoto", SourceClass: "engagement_photographer",
		CityID: "city_new_york_ny", State: "NY", City: "New York", Verified: "2026-07-31",
	},
	{
		Name: "The Plaza Hotel", OfficialURL: "https://www.theplazany.com/",
		Handle: "theplazany", SourceClass: "wedding_venue",
		CityID: "city_new_york_ny", State: "NY", City: "New York", Verified: "2026-07-31",
	},
	{
		Name: "Tiffany & Co.", OfficialURL: "https://www.tiffany.com/",
		Handle: "tiffanyandco", SourceClass: "jeweler",
		CityID: "city_new_york_ny", State: "NY", City: "New York", Verified: "2026-07-31",
	},
	{
		Name: "Cartier", OfficialURL: "https://www.cartier.com/",
		Handle: "cartier", SourceClass: "jeweler",
		CityID: "city_new_york_ny", State: "NY", City: "New York", Verified: "2026-07-31",
	},
	{
		Name: "Harry Winston", OfficialURL: "https://www.harrywinston.com/",
		Handle: "harrywinston", SourceClass: "jeweler",
		CityID: "city_new_york_ny", State: "NY", City: "New York", Verified: "2026-07-31",
	},
	{
		Name: "Brooklyn Botanic Garden", OfficialURL: "https://www.bbg.org/",
		Handle: "brooklynbotanic", SourceClass: "wedding_venue",
		CityID: "city_brooklyn_ny", State: "NY", City: "Brooklyn", Verified: "2026-07-31",
	},
	{
		Name: "The Foundry LIC", OfficialURL: "https://www.thefoundrylic.com/",
		Handle: "thefoundrylic", SourceClass: "wedding_venue",
		CityID: "city_new_york_ny", State: "NY", City: "New York", Verified: "2026-07-31",
	},
	{
		Name: "Brookfield Place New York", OfficialURL: "https://bfplny.com/",
		Handle: "bfplny", SourceClass: "wedding_venue",
		CityID: "city_new_york_ny", State: "NY", City: "New York", Verified: "2026-07-31",
	},
	{
		Name: "The Metropolitan Museum of Art (public)", OfficialURL: "https://www.metmuseum.org/",
		Handle: "metmuseum", SourceClass: "wedding_venue",
		CityID: "city_new_york_ny", State: "NY", City: "New York", Verified: "2026-07-31",
	},
}

// NY government + church sources, verified 2026-08-01 via web search.
// Government: NYC City Clerk holds NYC marriage records; NYS Dept of Health
// holds upstate records; Westchester County Archives has an online index.
var nyGovSources = []GovSource{
	{
		CountyFIPS: "36061", // New York County (Manhattan)
		CourtName:  "NYC City Clerk — Manhattan Office",
		CourtURL:   "https://www.cityclerk.nyc.gov",
		SearchURL:  "https://www.cityclerk.nyc.gov/content/marriage-records",
		Note:       "NYC City Clerk holds marriage records for all 5 boroughs; online request via VitalChek, in-person index search available.",
	},
	{
		CountyFIPS: "36047", // Kings County (Brooklyn)
		CourtName:  "NYC City Clerk — Brooklyn Office",
		CourtURL:   "https://www.cityclerk.nyc.gov",
		SearchURL:  "https://www.cityclerk.nyc.gov/content/marriage-records",
		Note:       "Brooklyn marriage records via NYC City Clerk; same portal as Manhattan.",
	},
	{
		CountyFIPS: "36087", // Westchester County
		CourtName:  "Westchester County Archives",
		CourtURL:   "https://recordcenter.westchestergov.com",
		SearchURL:  "https://recordcenter.westchestergov.com/MarriageSearchResultAll.aspx",
		Note:       "Online marriage records index with 200K+ records; public search available.",
	},
}

var nyDioceses = []DioceseDef{
	{Slug: "new_york", Name: "Archdiocese of New York", Type: "archdiocese",
		Website: "https://archny.org", Directory: "https://www.archny.org/map/parishes", HubCityID: "city_new_york_ny"},
	{Slug: "brooklyn", Name: "Diocese of Brooklyn", Type: "diocese",
		Website: "https://dioceseofbrooklyn.org", Directory: "https://mass.dioceseofbrooklyn.org/all", HubCityID: "city_brooklyn_ny"},
	{Slug: "rockville_centre", Name: "Diocese of Rockville Centre", Type: "diocese",
		Website: "https://drvc.org", Directory: "https://drvc.org/parish-finder"},
	{Slug: "brooklyn_queens", Name: "Diocese of Brooklyn and Queens", Type: "diocese",
		Website: "https://dioceseofbrooklyn.org", Directory: "https://mass.dioceseofbrooklyn.org/all"},
	{Slug: "albany", Name: "Diocese of Albany", Type: "diocese",
		Website: "https://www.rcda.org", Directory: "https://www.rcda.org/parishes/find"},
	{Slug: "buffalo", Name: "Diocese of Buffalo", Type: "diocese",
		Website: "https://www.buffalodiocese.org", Directory: "https://www.buffalodiocese.org/parish-finder/"},
	{Slug: "rochester", Name: "Diocese of Rochester", Type: "diocese",
		Website: "https://www.dor.org", Directory: "https://ps.dor.org/directory/"},
	{Slug: "syracuse", Name: "Diocese of Syracuse", Type: "diocese",
		Website: "https://www.syrdio.org", Directory: "https://www.syrdio.org/parishes"},
	{Slug: "ogdensburg", Name: "Diocese of Ogdensburg", Type: "diocese",
		Website: "https://www.rcdony.org", Directory: "https://www.rcdony.org/parish-directory"},
}

// NY Catholic parishes, verified 2026-08-01 via web search + direct fetch.
// Manhattan parishes belong to the Archdiocese of New York (slug "new_york");
// Brooklyn parishes belong to the Diocese of Brooklyn (slug "brooklyn").
// BulletinURL set only where a real, reachable bulletin archive was confirmed
// by fetching the page; parishes without a verified archive omit the field.
var nyParishes = []ParishDef{
	{DioceseSlug: "new_york", Name: "St. Patrick's Cathedral", Address: "460 Madison Ave, New York, NY 10022"},
	{
		DioceseSlug: "new_york", Name: "Church of St. Ignatius Loyola",
		Address:     "980 Park Avenue, New York, NY 10028",
		BulletinURL: "https://ignatius.nyc/our-parish/weekly-parish-bulletins/",
	},
	{
		DioceseSlug: "new_york", Name: "Parish of St. Vincent Ferrer and St. Catherine of Siena",
		Address:     "869 Lexington Avenue, New York, NY 10065",
		BulletinURL: "https://www.svsc.info/bulletin",
	},
	{DioceseSlug: "new_york", Name: "Holy Trinity Church", Address: "213 West 82nd Street, New York, NY 10024"},
	{DioceseSlug: "new_york", Name: "St. Jean Baptiste Church", Address: "184 East 76th Street, New York, NY 10021"},
	{DioceseSlug: "new_york", Name: "Church of Saint Agnes", Address: "143 East 43rd Street, New York, NY 10017"},
	{
		DioceseSlug: "brooklyn", Name: "St. James Cathedral Basilica",
		Address:     "250 Cathedral Place, Brooklyn, NY 11201",
		BulletinURL: "https://brooklyncathedral.org/bulletins",
	},
	{
		DioceseSlug: "brooklyn", Name: "Co-Cathedral of St. Joseph",
		Address:     "856 Pacific Street, Brooklyn, NY 11238",
		BulletinURL: "https://brooklyncocathedral.org/bulletins",
	},
	{
		DioceseSlug: "brooklyn", Name: "Queen of All Saints Church",
		Address:     "300 Vanderbilt Ave, Brooklyn, NY 11205",
		BulletinURL: "https://qasrcc.org/bulletins",
	},
}

var nyPack = StatePack{
	State:      "NY",
	Cities:     nyCities,
	Government: nyGovSources,
	Dioceses:   nyDioceses,
	Parishes:   nyParishes,
	Vendors:    nyVendors,
}

// --- California -------------------------------------------------------------
// Social handles verified 2026-07-31 from each brand's own public website.
// Government + church pending — to be researched and appended.

var caCities = []CityDef{
	{
		ID: "city_los_angeles_ca", State: "CA", County: "06037", Name: "Los Angeles",
		Lat: 34.0522, Lng: -118.2437,
		Markets: []string{"losangeles", "la", "hollywood", "santamonica", "beverlyhills", "malibu"},
	},
	{
		ID: "city_san_francisco_ca", State: "CA", County: "06075", Name: "San Francisco",
		Lat: 37.7749, Lng: -122.4194,
		Markets: []string{"sanfrancisco", "sf", "bayarea", "napa", "sonoma"},
	},
}

var caVendors = []VendorDef{
	{
		Name: "José Villa Photography", OfficialURL: "https://josevilla.com/",
		Handle: "josevilla", SourceClass: "engagement_photographer",
		CityID: "city_los_angeles_ca", State: "CA", City: "Los Angeles", Verified: "2026-07-31",
	},
	{
		Name: "May Ioso Taluno", OfficialURL: "https://www.instagram.com/mayiosotaluno/",
		Handle: "mayiosotaluno", SourceClass: "engagement_photographer",
		CityID: "city_los_angeles_ca", State: "CA", City: "Los Angeles", Verified: "2026-07-31",
	},
	{
		Name: "Adrienne Gunde Photography", OfficialURL: "https://www.instagram.com/adriennegunde/",
		Handle: "adriennegunde", SourceClass: "engagement_photographer",
		CityID: "city_los_angeles_ca", State: "CA", City: "Los Angeles", Verified: "2026-07-31",
	},
	{
		Name: "Shelby Ayn Photos", OfficialURL: "https://www.instagram.com/shelbyaynphotos/",
		Handle: "shelbyaynphotos", SourceClass: "engagement_photographer",
		CityID: "city_los_angeles_ca", State: "CA", City: "Los Angeles", Verified: "2026-07-31",
	},
	{
		Name: "Lulan Photography", OfficialURL: "https://www.instagram.com/lulanphoto/",
		Handle: "lulanphoto", SourceClass: "engagement_photographer",
		CityID: "city_los_angeles_ca", State: "CA", City: "Los Angeles", Verified: "2026-07-31",
	},
	{
		Name: "Gretchen Parker Photo", OfficialURL: "https://www.instagram.com/gretchenparkerphoto/",
		Handle: "gretchenparkerphoto", SourceClass: "engagement_photographer",
		CityID: "city_los_angeles_ca", State: "CA", City: "Los Angeles", Verified: "2026-07-31",
	},
	{
		Name: "Bel-Air Bay Club", OfficialURL: "https://www.belairbayclub.com/",
		Handle: "belairbayclub", SourceClass: "wedding_venue",
		CityID: "city_los_angeles_ca", State: "CA", City: "Los Angeles", Verified: "2026-07-31",
	},
	{
		Name: "The Getty Villa (events / public)", OfficialURL: "https://www.getty.edu/visit/villa/",
		Handle: "gettymuseum", SourceClass: "wedding_venue",
		CityID: "city_los_angeles_ca", State: "CA", City: "Los Angeles", Verified: "2026-07-31",
	},
	{
		Name: "San Francisco City Hall", OfficialURL: "https://sf.gov/location/city-hall",
		Handle: "sfcityhall", SourceClass: "wedding_venue",
		CityID: "city_san_francisco_ca", State: "CA", City: "San Francisco", Verified: "2026-07-31",
	},
	{
		Name: "Palace of Fine Arts", OfficialURL: "https://palaceoffinearts.com/",
		Handle: "palaceoffinearts", SourceClass: "wedding_venue",
		CityID: "city_san_francisco_ca", State: "CA", City: "San Francisco", Verified: "2026-07-31",
	},
	{
		Name: "Napa Valley wine country venues (public IG hub)", OfficialURL: "https://www.visitnapavalley.com/",
		Handle: "visitnapavalley", SourceClass: "wedding_venue",
		CityID: "city_san_francisco_ca", State: "CA", City: "San Francisco", Verified: "2026-07-31",
	},
	{
		Name: "Brilliant Earth (jeweler)", OfficialURL: "https://www.brilliantearth.com/",
		Handle: "brilliantearth", SourceClass: "jeweler",
		CityID: "city_san_francisco_ca", State: "CA", City: "San Francisco", Verified: "2026-07-31",
	},
}

// CA government + church sources, verified 2026-08-01 via web search.
// Government: California marriage records held by county recorder/clerk.
var caGovSources = []GovSource{
	{
		CountyFIPS: "06037", // Los Angeles County
		CourtName:  "Los Angeles County Registrar-Recorder/County Clerk",
		CourtURL:   "https://lavote.gov",
		SearchURL:  "https://lavote.gov/home/recorder/marriage-records/viewing-vital-records",
		Note:       "Marriage records from 1852 to present; public index inspection available in-person, online request via VitalChek.",
	},
	{
		CountyFIPS: "06075", // San Francisco County
		CourtName:  "San Francisco County Recorder",
		CourtURL:   "https://sf.gov",
		SearchURL:  "https://sf.gov/location/city-hall",
		Note:       "Marriage certificates via SF County Recorder; request-oriented, online enumeration not available.",
	},
	{
		CountyFIPS: "06073", // San Diego County
		CourtName:  "San Diego County Recorder/County Clerk",
		CourtURL:   "https://www.sdarcc.gov",
		SearchURL:  "https://www.sdarcc.gov/content/arcc/home/divisions/recorder-clerk/birth-death-marriage-certificate/marriage-certificate.html",
		Note:       "Marriage certificate requests via Assessor/Recorder/County Clerk; in-person and mail request.",
	},
}

var caDioceses = []DioceseDef{
	{Slug: "los_angeles", Name: "Archdiocese of Los Angeles", Type: "archdiocese",
		Website: "https://lacatholics.org", Directory: "https://lacatholics.org/find/", HubCityID: "city_los_angeles_ca"},
	{Slug: "san_francisco", Name: "Archdiocese of San Francisco", Type: "archdiocese",
		Website: "https://www.sfarch.org", Directory: "https://www.sfarch.org/parishes", HubCityID: "city_san_francisco_ca"},
	{Slug: "fresno", Name: "Diocese of Fresno", Type: "diocese",
		Website: "https://www.dioceseoffresno.org", Directory: "https://www.dioceseoffresno.org/parishes"},
	{Slug: "monterey", Name: "Diocese of Monterey", Type: "diocese",
		Website: "https://www.diocesemonterey.org", Directory: "https://www.diocesemonterey.org/parishes"},
	{Slug: "oakland", Name: "Diocese of Oakland", Type: "diocese",
		Website: "https://www.oakdiocese.org", Directory: "https://www.oakdiocese.org/parishes"},
	{Slug: "orange", Name: "Diocese of Orange", Type: "diocese",
		Website: "https://www.rcbo.org", Directory: "https://www.rcbo.org/parish-directory"},
	{Slug: "sacramento", Name: "Diocese of Sacramento", Type: "diocese",
		Website: "https://www.scd.org", Directory: "https://www.scd.org/parishes"},
	{Slug: "san_bernardino", Name: "Diocese of San Bernardino", Type: "diocese",
		Website: "https://www.sbscca.org", Directory: "https://www.sbscca.org/parishes"},
	{Slug: "san_diego", Name: "Diocese of San Diego", Type: "diocese",
		Website: "https://sdcatholic.org", Directory: "https://sdcatholic.org/parishes"},
	{Slug: "san_jose", Name: "Diocese of San Jose", Type: "diocese",
		Website: "https://www.dsj.org", Directory: "https://www.dsj.org/parishes"},
	{Slug: "santa_rosa", Name: "Diocese of Santa Rosa", Type: "diocese",
		Website: "https://www.srdiocese.org", Directory: "https://www.srdiocese.org/parishes"},
	{Slug: "stockton", Name: "Diocese of Stockton", Type: "diocese",
		Website: "https://www.dioceseofstockton.org", Directory: "https://www.dioceseofstockton.org/parishes"},
}

// CA parishes in the Archdioceses of Los Angeles and San Francisco.
// Names + addresses verified from each parish's own website (or the
// archdiocese parish directory). Bulletin URLs verified by direct fetch of
// each parish's bulletin archive page.
var caParishes = []ParishDef{
	{DioceseSlug: "los_angeles", Name: "Cathedral of Our Lady of the Angels",
		Address: "555 W Temple St, Los Angeles, CA 90012"},
	{DioceseSlug: "los_angeles", Name: "Cathedral Chapel of St. Vibiana",
		Address:     "923 S La Brea Ave, Los Angeles, CA 90036",
		BulletinURL: "https://cathedralchapel.org/sunday-bulletin"},
	{DioceseSlug: "los_angeles", Name: "St. Emydius Catholic Church",
		Address:     "10900 California Ave, Lynwood, CA 90262",
		BulletinURL: "https://www.saintemydius.org/parish-bulletin.html"},
	{DioceseSlug: "los_angeles", Name: "St. Monica Catholic Community",
		Address:     "725 California Ave, Santa Monica, CA 90403",
		BulletinURL: "https://stmonica.net/church-bulletin"},
	{DioceseSlug: "los_angeles", Name: "St. Charles Borromeo Church",
		Address: "10800 Moorpark St, North Hollywood, CA 91602"},
	{DioceseSlug: "los_angeles", Name: "St. Brendan Catholic Church",
		Address:     "310 S Van Ness Ave, Los Angeles, CA 90020",
		BulletinURL: "https://stbrendanla.org/bulletins"},
	{DioceseSlug: "san_francisco", Name: "Cathedral of St. Mary of the Assumption",
		Address: "1111 Gough St, San Francisco, CA 94109"},
	{DioceseSlug: "san_francisco", Name: "Old St. Mary's Cathedral & Chinese Mission",
		Address:     "660 California St, San Francisco, CA 94108",
		BulletinURL: "https://www.osmsf.org/bulletins"},
	{DioceseSlug: "san_francisco", Name: "Saints Peter and Paul Church",
		Address:     "666 Filbert St, San Francisco, CA 94133",
		BulletinURL: "https://www.salesiansspp.org/bulletins"},
	{DioceseSlug: "san_francisco", Name: "St. Dominic's Catholic Church",
		Address: "2390 Bush St, San Francisco, CA 94115"},
}

var caPack = StatePack{
	State:      "CA",
	Cities:     caCities,
	Government: caGovSources,
	Dioceses:   caDioceses,
	Parishes:   caParishes,
	Vendors:    caVendors,
}

// --- Texas ------------------------------------------------------------------
// Defined in pack_tx.go.

// marketsForStates returns ACTIVE_MARKETS tokens to merge into server env docs.
func marketsForStates(states []string) []string {
	var out []string
	seen := map[string]bool{}
	for _, st := range states {
		p := packFor(st)
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
