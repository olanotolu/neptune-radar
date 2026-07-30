package main

// Real, verified Ohio facts. Every value here was confirmed against a real
// public source before being written — see the comment above each block for
// exactly where it came from. Nothing here is invented or guessed.

// ohioCounty is one of Ohio's 88 real counties. This list (FIPS id + name)
// was extracted directly from us-atlas/counties-10m.json (the same US
// Census TIGER-derived dataset the frontend map already renders), filtered
// to FIPS ids starting with "39" (Ohio's state FIPS code) — not typed from
// memory.
type ohioCounty struct {
	FIPS string
	Name string
}

var ohioCounties = []ohioCounty{
	{"39001", "Adams"}, {"39003", "Allen"}, {"39005", "Ashland"}, {"39007", "Ashtabula"},
	{"39009", "Athens"}, {"39011", "Auglaize"}, {"39013", "Belmont"}, {"39015", "Brown"},
	{"39017", "Butler"}, {"39019", "Carroll"}, {"39021", "Champaign"}, {"39023", "Clark"},
	{"39025", "Clermont"}, {"39027", "Clinton"}, {"39029", "Columbiana"}, {"39031", "Coshocton"},
	{"39033", "Crawford"}, {"39035", "Cuyahoga"}, {"39037", "Darke"}, {"39039", "Defiance"},
	{"39041", "Delaware"}, {"39043", "Erie"}, {"39045", "Fairfield"}, {"39047", "Fayette"},
	{"39049", "Franklin"}, {"39051", "Fulton"}, {"39053", "Gallia"}, {"39055", "Geauga"},
	{"39057", "Greene"}, {"39059", "Guernsey"}, {"39061", "Hamilton"}, {"39063", "Hancock"},
	{"39065", "Hardin"}, {"39067", "Harrison"}, {"39069", "Henry"}, {"39071", "Highland"},
	{"39073", "Hocking"}, {"39075", "Holmes"}, {"39077", "Huron"}, {"39079", "Jackson"},
	{"39081", "Jefferson"}, {"39083", "Knox"}, {"39085", "Lake"}, {"39087", "Lawrence"},
	{"39089", "Licking"}, {"39091", "Logan"}, {"39093", "Lorain"}, {"39095", "Lucas"},
	{"39097", "Madison"}, {"39099", "Mahoning"}, {"39101", "Marion"}, {"39103", "Medina"},
	{"39105", "Meigs"}, {"39107", "Mercer"}, {"39109", "Miami"}, {"39111", "Monroe"},
	{"39113", "Montgomery"}, {"39115", "Morgan"}, {"39117", "Morrow"}, {"39119", "Muskingum"},
	{"39121", "Noble"}, {"39123", "Ottawa"}, {"39125", "Paulding"}, {"39127", "Perry"},
	{"39129", "Pickaway"}, {"39131", "Pike"}, {"39133", "Portage"}, {"39135", "Preble"},
	{"39137", "Putnam"}, {"39139", "Richland"}, {"39141", "Ross"}, {"39143", "Sandusky"},
	{"39145", "Scioto"}, {"39147", "Seneca"}, {"39149", "Shelby"}, {"39151", "Stark"},
	{"39153", "Summit"}, {"39155", "Trumbull"}, {"39157", "Tuscarawas"}, {"39159", "Union"},
	{"39161", "Van Wert"}, {"39163", "Vinton"}, {"39165", "Warren"}, {"39167", "Washington"},
	{"39169", "Wayne"}, {"39171", "Williams"}, {"39173", "Wood"}, {"39175", "Wyandot"},
}

const franklinCountyFIPS = "39049"

// Franklin County Probate Court's real online marriage-license search,
// verified 2026-07-29: https://probate.franklincountyohio.gov/Record-Search/Marriage-License-Search
// Coverage and phone number confirmed on the court's own site.
const (
	franklinCourtName     = "Franklin County Probate Court"
	franklinCourtURL      = "https://probate.franklincountyohio.gov"
	franklinSearchURL     = "https://probate.franklincountyohio.gov/Record-Search/Marriage-License-Search"
	franklinPhone         = "(614) 525-3894"
	franklinCoverageStart = "1994-01-03"
	franklinCoverageNote  = "Online search covers marriage licenses issued January 3, 1994 to present."
)

// ohioGovSource is one verified county probate-court source. Every URL and
// capability note below comes from the 2026-07-29 source review (see
// PRODUCTION_GAPS.md) — the note states honestly what the endpoint can do
// and what still needs testing. FIPS ids match the ohioCounties table above.
type ohioGovSource struct {
	CountyFIPS string
	CourtName  string
	CourtURL   string
	SearchURL  string
	Note       string
}

var ohioGovSources = []ohioGovSource{
	{
		// Montgomery County (Dayton) — supports searching by issued-date
		// range as well as names: strongest enumeration candidate.
		CountyFIPS: "39113",
		CourtName:  "Montgomery County Probate Court",
		CourtURL:   "https://go.mcohio.org",
		SearchURL:  "https://go.mcohio.org/applications/probate/prodcfm/marriagesearch.cfm",
		Note:       "Supports name search and issued-date-range search — enumeration candidate; connector build next.",
	},
	{
		// Hamilton County (Cincinnati) — public electronic index.
		CountyFIPS: "39061",
		CourtName:  "Hamilton County Probate Court",
		CourtURL:   "https://www.probatect.org/marriage-license",
		SearchURL:  "https://www.probatect.org/court-records/archive-categories/marriages",
		Note:       "Public electronic index; verification use confirmed, enumeration capability needs testing.",
	},
	{
		// Cuyahoga County (Cleveland) — marriage dept + web docket.
		CountyFIPS: "39035",
		CourtName:  "Cuyahoga County Probate Court",
		CourtURL:   "https://probate.cuyahogacounty.gov/marriage.aspx",
		SearchURL:  "https://probate.cuyahogacounty.gov/pa/",
		Note:       "Web docket available for verification; automation and terms review required.",
	},
	{
		// Summit County (Akron) — record information page.
		CountyFIPS: "39153",
		CourtName:  "Summit County Probate Court",
		CourtURL:   "https://summitcountycourt.org",
		SearchURL:  "https://summitcountycourt.org/marriage-divorce-records/",
		Note:       "Record information page; request-oriented, automation capability needs testing.",
	},
	{
		// Delaware County — official probate record search.
		CountyFIPS: "39041",
		CourtName:  "Delaware County Probate Court",
		CourtURL:   "https://probate.co.delaware.oh.us",
		SearchURL:  "https://probate.co.delaware.oh.us/recordsearch/",
		Note:       "Official probate record search; marriage-record capability needs testing.",
	},
	{
		// Fairfield County — online probate records search form.
		CountyFIPS: "39045",
		CourtName:  "Fairfield County Probate Court",
		CourtURL:   "https://www.fairfieldcountyprobate.com",
		SearchURL:  "https://www.fairfieldcountyprobate.com/ff-Probate-Records-Search-Form.html",
		Note:       "Online probate records search form; marriage filtering needs testing.",
	},
	{
		// Licking County — online portal behind an agreement gate.
		CountyFIPS: "39089",
		CourtName:  "Licking County Probate Court",
		CourtURL:   "https://pjc-portal.lickingcounty.gov",
		SearchURL:  "https://pjc-portal.lickingcounty.gov/recordSearch.php?k=acceptAgreementsearchForm4503",
		Note:       "Online portal (agreement-gated); terms and record-type capability need verification.",
	},
	{
		// Lucas County (Toledo) — official custodian page.
		CountyFIPS: "39095",
		CourtName:  "Lucas County Probate Court",
		CourtURL:   "https://www.co.lucas.oh.us/169/Probate-Court",
		SearchURL:  "https://www.co.lucas.oh.us/169/Probate-Court",
		Note:       "Official custodian page; online enumeration not yet confirmed.",
	},
}

// Diocese of Columbus, verified 2026-07-29: official site and parish
// directory tool at https://columbuscatholic.org/find-a-parish. Deanery/
// parish/county counts as published on that page.
const (
	dioceseColumbusName      = "Diocese of Columbus"
	dioceseColumbusWebsite   = "https://columbuscatholic.org"
	dioceseColumbusDirectory = "https://columbuscatholic.org/find-a-parish"
	dioceseColumbusDeaneries = 10
	dioceseColumbusParishesN = 81
	dioceseColumbusCountiesN = 23
)

// All six Ohio Catholic jurisdictions, verified 2026-07-29. Directory URLs
// point at each jurisdiction's own parish-finder; for Toledo and Youngstown
// the directory is the main site (their finders are search forms).
type ohioDiocese struct {
	Slug      string
	Name      string
	Type      string // "diocese" or "archdiocese"
	Website   string
	Directory string
}

var ohioDioceses = []ohioDiocese{
	{"cleveland", "Diocese of Cleveland", "diocese", "https://www.dioceseofcleveland.org", "https://www.dioceseofcleveland.org/about/our-parishes"},
	{"cincinnati", "Archdiocese of Cincinnati", "archdiocese", "https://catholicaoc.org", "https://catholicaoc.org/parishes"},
	{"toledo", "Diocese of Toledo", "diocese", "https://toledodiocese.org", "https://toledodiocese.org"},
	{"youngstown", "Diocese of Youngstown", "diocese", "https://doy.org", "https://doy.org"},
	{"steubenville", "Diocese of Steubenville", "diocese", "https://www.diosteub.org", "https://www.diosteub.org/parishfinder"},
}

// ohioParish is a real Columbus-proper parish. Names and addresses verified
// 2026-07-29 against https://en.wikipedia.org/wiki/List_of_churches_in_the_Diocese_of_Columbus
// (which cites the diocese's own records). BulletinURL is set only where a
// real, reachable bulletin archive was located during the 2026-07-29 source
// review; Aggregator marks URLs that are third-party listings (Parishes
// Online / Discover Mass) rather than the parish's own site. BannsEvidence
// links a real bulletin PDF observed to contain a "Banns of Marriage"
// section with couple names and a wedding date.
type ohioParish struct {
	Name          string
	Address       string
	BulletinURL   string
	Aggregator    bool
	BannsEvidence string
}

const parishSourceURL = "https://en.wikipedia.org/wiki/List_of_churches_in_the_Diocese_of_Columbus"

var columbusParishes = []ohioParish{
	{Name: "Community of Holy Rosary and Saint John the Evangelist", Address: "648 S Ohio Ave, Columbus, OH 43205"},
	{Name: "Holy Cross Church", Address: "204 S 5th St, Columbus, OH 43215",
		BulletinURL: "https://parishesonline.com/organization/holy-cross-catholic-church-43215", Aggregator: true},
	{Name: "Saint Leo Oratory", Address: "221 Hanford St, Columbus, OH 43206"},
	{Name: "Saint Dominic Church", Address: "453 N 20th St, Columbus, OH 43203",
		BulletinURL: "https://stdominic-church.org/bulletins"},
	{Name: "Saint Joseph Cathedral", Address: "212 E Broad St, Columbus, OH 43215"},
	{Name: "Saint Mary, Mother of God Church", Address: "684 S 3rd St, Columbus, OH 43206"},
	{Name: "Saint Patrick Church", Address: "280 N Grant Ave, Columbus, OH",
		BulletinURL: "https://www.stpatrickcolumbus.org/weekly-bulletin"},
	{Name: "Saint Thomas the Apostle Church", Address: "2692 E 5th Ave, Columbus, OH 43219"},
	{Name: "Saints Augustine and Gabriel Church", Address: "1550 E Hudson St, Columbus, OH 43211"},
	// Immaculate Conception: real bulletin observed with a "BANNS OF
	// MARRIAGE" section naming the couple and the wedding date — the proof
	// that this lane produces the signal Neptune wants.
	{Name: "Immaculate Conception Church",
		BulletinURL:   "https://www.iccols.org/bulletin/",
		BannsEvidence: "https://www.iccols.org/wp-content/uploads/2025/10/IC-Columbus-10-12.pdf"},
	{Name: "Sacred Heart Church", BulletinURL: "https://sacredheartchurchcolumbus.org/bulletins"},
	{Name: "Holy Spirit Church", BulletinURL: "https://holyspiritcolumbus.org/bulletins"},
	{Name: "Saint Agatha Church",
		BulletinURL: "https://discovermass.com/church/st-agatha-columbus-oh/?id=20170425", Aggregator: true},
}

// ohioVendor is a real Columbus wedding-industry business. Instagram handles
// were verified 2026-07-29 by fetching each business's own official website
// and reading the handle off its own social links — never guessed from the
// business name. SourceClass values match signals.WatchedSourceClasses.
type ohioVendor struct {
	Name        string
	OfficialURL string
	Handle      string // verified real Instagram handle, no "@"
	SourceClass string
}

var columbusVendors = []ohioVendor{
	{"Starling Studio", "https://www.starling-studio.com/", "starling_studio", "engagement_photographer"},
	{"Jessica Miller Photography", "https://www.thejessicamillerphotos.com/", "jmillerphotos", "engagement_photographer"},
	{"Laura Witherow Photography", "https://laurawitherowphotography.com/", "laurawitherow", "engagement_photographer"},
	{"Kismet Visuals & Co", "https://kismetvisuals.com/", "kismetvisuals", "engagement_photographer"},
	{"Svetlana Photography", "https://svetlanaphotography.com/", "svetphoto", "engagement_photographer"},
	{"Asteria Photography", "https://www.asteriaphoto.com/", "asteriaphoto", "engagement_photographer"},
	{"Magnolia Hill Farm", "https://magnoliahill-farm.com/", "magnoliahillfarm", "wedding_venue"},
	{"Jorgensen Farms", "https://jorgensen-farms.com/", "jorgensenfarms", "wedding_venue"},
	{"Franklin Park Conservatory and Botanical Gardens", "https://www.fpconservatory.org/", "fpconservatory", "wedding_venue"},
	{"The Columbus Athenaeum", "https://www.columbusmeetings.com/", "thecolumbusathenaeum", "wedding_venue"},
	{"Le Méridien Columbus, The Joseph", "https://www.weddingsatthejoseph.com/", "lmcolumbusthejoseph", "wedding_venue"},
	{"Brookshire", "https://brookshire.biz/", "brookshireweddings", "wedding_venue"},
	{"Worthington Hills Country Club", "https://www.worthingtonhills.com/", "worthingtonhillscountryclub", "wedding_venue"},
}
