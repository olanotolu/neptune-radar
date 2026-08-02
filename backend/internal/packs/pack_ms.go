package packs

// Mississippi source pack — verified 2026-08-01.
//
// Government: Mississippi marriage records are held by the county Circuit
// Clerk. Most MS counties do not offer online marriage-record search portals;
// records are request-oriented (in-person or mail). Where an online search
// exists (Harrison, DeSoto), the SearchURL points at the actual search page.
// For counties without online search, SearchURL points at the Circuit Clerk's
// official page where marriage-license info and contact details are listed.
//
// Church: both Mississippi dioceses (Jackson, Biloxi) verified via USCCB +
// each diocese's own website. Jackson-area parishes (Diocese of Jackson) were
// verified against the diocese's metro-jackson-parishes page + each parish's
// own website. Bulletin URLs verified by direct search for each parish's
// bulletin archive.
//
// Social: Instagram handles verified from each business's own public website
// social links (or from IG/Threads search results where the site is
// JS-rendered and the handle was visible in the search snippet). Verification
// date recorded per vendor.

var msPack = StatePack{
	State: "MS",

	// --- Cities ----------------------------------------------------------
	Cities: []CityDef{
		{ID: "city_jackson_ms", State: "MS", County: "28049", Name: "Jackson",
			Lat: 32.2988, Lng: -90.1848, Markets: []string{"jackson", "ms", "hinds", "mississippi"}},
		{ID: "city_gulfport_ms", State: "MS", County: "28059", Name: "Gulfport",
			Lat: 30.3674, Lng: -89.0926, Markets: []string{"gulfport", "harrison", "ms", "gulfcoast"}},
	},

	// --- Government (circuit clerk marriage-record sources) --------------
	Government: []GovSource{
		{
			// Hinds County (Jackson) — Circuit Clerk page; no online marriage
			// search portal. Records are request-oriented (in-person or mail).
			CountyFIPS: "28049",
			CourtName:  "Hinds County Circuit Clerk",
			CourtURL:   "https://www.hindscountyms.com/elected-offices/circuit-clerk",
			SearchURL:  "https://www.hindscountyms.com/elected-offices/circuit-clerk",
			Note:       "Marriage records via Circuit Clerk; request-oriented, no online search portal.",
		},
		{
			// Harrison County (Gulfport) — online marriage license search by
			// groom/bride name, split by judicial district.
			CountyFIPS: "28059",
			CourtName:  "Harrison County Circuit Clerk",
			CourtURL:   "https://www.harrisoncountyms.gov/how_do_i_/apply_for/a_marriage_license.php",
			SearchURL:  "http://harrison2.co.harrison.ms.us/marriage/",
			Note:       "Online marriage license search by name, split by judicial district; enumeration candidate.",
		},
		{
			// Rankin County — Circuit Clerk page; no online marriage search.
			CountyFIPS: "28121",
			CourtName:  "Rankin County Circuit Clerk",
			CourtURL:   "https://www.rankincountyms.org/",
			SearchURL:  "https://www.rankincountyms.org/",
			Note:       "Marriage records via Circuit Clerk; request-oriented, no online search portal.",
		},
		{
			// DeSoto County — online marriage license search via Delta Computer
			// Systems portal (name, date, or book/page).
			CountyFIPS: "28033",
			CourtName:  "DeSoto County Circuit Clerk",
			CourtURL:   "https://www.desotocountyms.gov/639/Marriage-License",
			SearchURL:  "https://www.deltacomputersystems.com/MS/MS17/mllinkquerym2.html",
			Note:       "Online marriage license search by name/date/book-page via Delta Computer Systems; enumeration candidate.",
		},
		{
			// Jackson County (Pascagoula) — Circuit Clerk page; no online
			// marriage search portal.
			CountyFIPS: "28109",
			CourtName:  "Jackson County Circuit Clerk",
			CourtURL:   "https://www.co.jackson.ms.us/296/Circuit-Clerk",
			SearchURL:  "https://www.co.jackson.ms.us/300/Marriage-Licenses",
			Note:       "Marriage licenses page with requirements; request-oriented, no online search portal.",
		},
		{
			// Madison County — Circuit Clerk page; marriage records search is
			// $10 per name (request-oriented).
			CountyFIPS: "28089",
			CourtName:  "Madison County Circuit Clerk",
			CourtURL:   "https://www.madisoncountycircuitclerk.com/",
			SearchURL:  "https://www.madisoncountycircuitclerk.com/fees-forms",
			Note:       "Marriage records search $10 per name; request-oriented, no online search portal.",
		},
		{
			// Forrest County (Hattiesburg) — Circuit Clerk page; no online
			// marriage search portal.
			CountyFIPS: "28035",
			CourtName:  "Forrest County Circuit Clerk",
			CourtURL:   "https://forrestcountyms.us/directory/circuit-clerk/",
			SearchURL:  "https://forrestcountyms.us/directory/circuit-clerk/",
			Note:       "Marriage records via Circuit Clerk; request-oriented, no online search portal.",
		},
	},

	// --- Church (Catholic dioceses + parishes) ---------------------------
	Dioceses: []DioceseDef{
		{Slug: "jackson", Name: "Diocese of Jackson", Type: "diocese",
			Website: "https://www.jacksondiocese.org", Directory: "https://www.jacksondiocese.org/parishes", HubCityID: "city_jackson_ms"},
		{Slug: "biloxi", Name: "Diocese of Biloxi", Type: "diocese",
			Website: "https://www.biloxidiocese.org", Directory: "https://www.biloxidiocese.org/parishes", HubCityID: "city_gulfport_ms"},
	},

	// Jackson-area parishes in the Diocese of Jackson. Names and addresses
	// verified from the diocese's metro-jackson-parishes page + each parish's
	// own website. Bulletin URLs verified by direct search for each parish's
	// bulletin archive.
	Parishes: []ParishDef{
		{DioceseSlug: "jackson", Name: "Cathedral of St. Peter the Apostle", Address: "123 N. West St., Jackson, MS 39201"},
		{DioceseSlug: "jackson", Name: "Christ the King Catholic Church", Address: "2303 John R. Lynch St., Jackson, MS 39209"},
		{DioceseSlug: "jackson", Name: "Holy Family Catholic Church", Address: "820 Forest Ave., Jackson, MS 39206"},
		{DioceseSlug: "jackson", Name: "Holy Ghost Catholic Church", Address: "1151 Cloister St., Jackson, MS 39202"},
		{
			DioceseSlug: "jackson", Name: "St. Richard Catholic Church",
			Address:     "1242 Lynwood Dr., Jackson, MS 39206",
			BulletinURL: "https://saintrichard.com/bulletins",
		},
		{
			DioceseSlug: "jackson", Name: "St. Therese Catholic Church",
			Address:     "329 W. McDowell Rd., Jackson, MS 39204",
			BulletinURL: "https://sttheresejackson.org/bulletins",
		},
		{
			DioceseSlug: "jackson", Name: "St. Paul Catholic Church",
			Address:     "5971 Hwy 25, Flowood, MS 39232",
			BulletinURL: "https://saintpaulcatholicchurch.com/announcements",
		},
		{
			DioceseSlug: "jackson", Name: "St. Joseph Catholic Church",
			Address:     "127 Church Rd., Gluckstadt, MS 39110",
			BulletinURL: "https://stjosephgluckstadt.com/bulletins",
		},
		{DioceseSlug: "jackson", Name: "St. Francis of Assisi Catholic Church", Address: "4000 W Tidewater Ln., Madison, MS 39110"},
	},

	// --- Social (wedding-industry Instagram vendors) ---------------------
	Vendors: []VendorDef{
		// Jackson photographers
		{
			Name: "Jessica Greene Photography", OfficialURL: "https://jessicagreenephoto.com/",
			Handle: "jessicagreenephoto", SourceClass: "engagement_photographer",
			CityID: "city_jackson_ms", State: "MS", City: "Jackson", Verified: "2026-08-01",
			TikTokHandle: "jessicagreenephoto",
		},
		{
			Name: "Shelby Mullins Photography", OfficialURL: "https://shelbymullinsphotography.com/",
			Handle: "shelbymullinsphotography", SourceClass: "engagement_photographer",
			CityID: "city_jackson_ms", State: "MS", City: "Jackson", Verified: "2026-08-01",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-photographers/shelby-mullins-photography-9703123",
		},
		{
			Name: "Patrick Remington Photography", OfficialURL: "https://patrickremingtonphotography.com/",
			Handle: "theremingtonphotography", SourceClass: "engagement_photographer",
			CityID: "city_jackson_ms", State: "MS", City: "Jackson", Verified: "2026-08-01",
			TikTokHandle: "theremingtonphotography",
		},
		{
			Name: "Mason Graves Photography", OfficialURL: "https://masongravesphoto.com/",
			Handle: "masongravesphoto", SourceClass: "engagement_photographer",
			CityID: "city_jackson_ms", State: "MS", City: "Jackson", Verified: "2026-08-01",
		},
		{
			Name: "Lindsey Jamison Photography", OfficialURL: "https://lindseyjamison.com/",
			Handle: "lindseyjamisonphoto", SourceClass: "engagement_photographer",
			CityID: "city_jackson_ms", State: "MS", City: "Jackson", Verified: "2026-08-01",
			TikTokHandle: "lindseyjamisonphoto",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-photographers/lindsey-jamison-photography-1467073",
		},
		// Jackson venues
		{
			Name: "Duling Hall", OfficialURL: "https://dulinghall.com/",
			Handle: "dulinghall", SourceClass: "wedding_venue",
			CityID: "city_jackson_ms", State: "MS", City: "Jackson", Verified: "2026-08-01",
			TikTokHandle: "dulinghall",
		},
		{
			Name: "The Rookery", OfficialURL: "https://therookeryms.com/",
			Handle: "therookeryms", SourceClass: "wedding_venue",
			CityID: "city_jackson_ms", State: "MS", City: "Jackson", Verified: "2026-08-01",
			TikTokHandle: "therookeryms",
		},
		{
			Name: "The Reed House at Live Oaks", OfficialURL: "https://thereedhouseatliveoaks.com/",
			Handle: "thereedhouseatliveoaks", SourceClass: "wedding_venue",
			CityID: "city_jackson_ms", State: "MS", City: "Jackson", Verified: "2026-08-01",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-venues/the-reed-house-at-live-oaks-5968046",
		},
		// Jackson jeweler
		{
			Name: "Albriton's Jewelry", OfficialURL: "https://www.albritons.com/",
			Handle: "albritonsjewelry", SourceClass: "jeweler",
			CityID: "city_jackson_ms", State: "MS", City: "Jackson", Verified: "2026-08-01",
		},
	},
}
