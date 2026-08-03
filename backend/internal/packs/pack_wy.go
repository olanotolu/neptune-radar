package packs

// Wyoming source pack — verified 2026-08-01.
//
// Government: Wyoming marriage records are held by the county clerk. The top
// 7 counties by population were verified against each county's official .gov
// site. Most Wyoming counties do not offer online marriage-record search;
// where an online index exists (Campbell) it is noted, otherwise the clerk's
// marriage-license page is the SearchURL.
//
// Church: the Diocese of Cheyenne covers all of Wyoming. Parish list verified
// via gcatholic.org + the diocese's own parish directory. Cheyenne- and
// Casper-area parishes verified against each parish's own website; bulletin
// URLs set where a real archive was located.
//
// Social: Instagram handles verified from each business's own public website
// social links. Verification date recorded per vendor.

var wyPack = StatePack{
	State: "WY",

	// --- Cities ----------------------------------------------------------
	Cities: []CityDef{
		{ID: "city_cheyenne_wy", State: "WY", County: "56021", Name: "Cheyenne",
			Lat: 41.1400, Lng: -104.8202, Markets: []string{"cheyenne", "wy", "laramie", "wyoming"}},
		{ID: "city_casper_wy", State: "WY", County: "56025", Name: "Casper",
			Lat: 42.8666, Lng: -106.3131, Markets: []string{"casper", "natrona", "wy"}},
	},

	// --- Government (county clerk marriage-record searches) --------------
	Government: []GovSource{
		{
			// Laramie County (Cheyenne) — county clerk marriage-license page.
			CountyFIPS: "56021",
			CourtName:  "Laramie County Clerk",
			CourtURL:   "https://www.laramiecountywy.gov/County-Government/Elected-Officials/County-Clerk",
			SearchURL:  "https://www.laramiecountywy.gov/County-Government/Elected-Officials/County-Clerk/Marriage-Licenses",
			Note:       "Marriage licenses issued by appointment; no online search portal, request-oriented.",
		},
		{
			// Natrona County (Casper) — county clerk; online land-records search
			// via iDoc Market but marriage licenses not available online.
			CountyFIPS: "56025",
			CourtName:  "Natrona County Clerk",
			CourtURL:   "https://www.natronacounty-wy.gov/18/Clerk",
			SearchURL:  "https://natrona.net/141/Marriage-Licenses",
			Note:       "Marriage licenses issued by clerk; online records search excludes marriage licenses, request-oriented.",
		},
		{
			// Campbell County (Gillette) — marriage index by surname letter,
			// published online by the county genealogical society.
			CountyFIPS: "56005",
			CourtName:  "Campbell County Clerk",
			CourtURL:   "https://campbellcountywy.gov/163/County-Clerk",
			SearchURL:  "https://www.campbellcountywy.gov/488/Campbell-County-Marriage-Index",
			Note:       "Alphabetical marriage index online; copies obtained from clerk's office; enumeration candidate.",
		},
		{
			// Sweetwater County (Rock Springs) — county clerk marriage-license
			// page; no online search portal.
			CountyFIPS: "56037",
			CourtName:  "Sweetwater County Clerk",
			CourtURL:   "https://sweetwatercountywy.gov/departments/county_clerk/marriage_licenses.php",
			SearchURL:  "https://sweetwatercountywy.gov/departments/county_clerk/marriage_licenses.php",
			Note:       "Marriage licenses issued by clerk; no online search portal, request-oriented.",
		},
		{
			// Albany County (Laramie) — county clerk; marriage-license info
			// page with downloadable application forms.
			CountyFIPS: "56001",
			CourtName:  "Albany County Clerk",
			CourtURL:   "https://www.albanycountywy.gov/157/Clerk",
			SearchURL:  "https://www.co.albany.wy.us/225/Marriage-License",
			Note:       "Marriage licenses issued by clerk; no online search portal, request-oriented.",
		},
		{
			// Sheridan County — county clerk marriage-license information page.
			CountyFIPS: "56033",
			CourtName:  "Sheridan County Clerk",
			CourtURL:   "https://www.sheridancountywy.gov/county_government/elected_offices/clerk_and_recorder/index.php",
			SearchURL:  "https://www.sheridancountywy.gov/county_government/elected_offices/clerk_and_recorder/marriage_license_information.php",
			Note:       "Marriage license info page; no online search portal, request-oriented.",
		},
		{
			// Fremont County (Riverton/Lander) — county clerk; marriage
			// licenses recorded at clerk's office, certified copies by request.
			CountyFIPS: "56013",
			CourtName:  "Fremont County Clerk",
			CourtURL:   "https://fremontcountywy.gov/government/elected_officials/clerk/vital_records.php",
			SearchURL:  "https://fremontcountywy.org/government/elected_officials/clerk/marriage_license.php",
			Note:       "Marriage licenses recorded at clerk's office; certified copies by request, no online search.",
		},
	},

	// --- Church (Catholic dioceses + parishes) ---------------------------
	Dioceses: []DioceseDef{
		{Slug: "cheyenne", Name: "Diocese of Cheyenne", Type: "diocese",
			Website: "https://www.dioceseofcheyenne.org", Directory: "https://www.dioceseofcheyenne.org/parishes", HubCityID: "city_cheyenne_wy"},
	},

	// Cheyenne- and Casper-area parishes in the Diocese of Cheyenne.
	// Names verified via gcatholic.org + each parish's own website. Bulletin
	// URLs verified by direct search for each parish's bulletin archive.
	Parishes: []ParishDef{
		{
			DioceseSlug: "cheyenne", Name: "Cathedral of St. Mary",
			Address:     "2107 Capitol Ave, Cheyenne, WY 82001",
			BulletinURL: "https://cathedralofstmary.com/bulletins",
		},
		{
			DioceseSlug: "cheyenne", Name: "Church of the Holy Trinity",
			Address: "1836 Hot Springs Ave, Cheyenne, WY 82001",
		},
		{
			DioceseSlug: "cheyenne", Name: "St. Joseph Catholic Church",
			Address:     "603 House Ave, Cheyenne, WY 82007",
			BulletinURL: "https://www.parishesonline.com/organization/st-joseph-church-82001",
			Aggregator:  true,
		},
		{
			DioceseSlug: "cheyenne", Name: "Our Lady of Fatima Catholic Church",
			Address:     "1401 CY Ave, Casper, WY 82604",
			BulletinURL: "https://fatimaincasper.org/our-parish/bulletins/",
		},
		{
			DioceseSlug: "cheyenne", Name: "St. Anthony of Padua Catholic Church",
			Address: "604 S Center St, Casper, WY 82601",
		},
		{
			DioceseSlug: "cheyenne", Name: "St. Patrick Catholic Church",
			Address: "400 Country Club Rd, Casper, WY 82609",
		},
	},

	// --- Social (wedding-industry Instagram vendors) ---------------------
	Vendors: []VendorDef{
		// Cheyenne photographers
		{
			Name: "Ardent Photography", OfficialURL: "https://www.ardentphotographyinc.com/",
			Handle: "ardentphoto", SourceClass: "engagement_photographer",
			CityID: "city_cheyenne_wy", State: "WY", City: "Cheyenne", Verified: "2026-08-01",
			TikTokHandle: "ardentphoto",
		},
		{
			Name: "Jasmine Mallo Imagery", OfficialURL: "https://www.jasminemalloimagery.com/",
			Handle: "jasminemalloimagery", SourceClass: "engagement_photographer",
			CityID: "city_cheyenne_wy", State: "WY", City: "Cheyenne", Verified: "2026-08-01",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-photographers/jasmine-mallo-imagery-6153494",
		},
		// Casper photographer
		{
			Name: "Tayler Ford Photography", OfficialURL: "https://taylerfordphotography.com/",
			Handle: "taylerfordphotography", SourceClass: "engagement_photographer",
			CityID: "city_casper_wy", State: "WY", City: "Casper", Verified: "2026-08-01",
			TikTokHandle: "taylerfordphotography",
		},
		// Cheyenne venues
		{
			Name: "Terry Bison Ranch Resort", OfficialURL: "https://terrybisonranch.com/",
			Handle: "terrybisonranch", SourceClass: "wedding_venue",
			CityID: "city_cheyenne_wy", State: "WY", City: "Cheyenne", Verified: "2026-08-01",
		},
		{
			Name: "White Antelope Barn", OfficialURL: "https://www.whiteantelopebarn.com/",
			Handle: "whiteantelopeweddings", SourceClass: "wedding_venue",
			CityID: "city_cheyenne_wy", State: "WY", City: "Cheyenne", Verified: "2026-08-01",
			TikTokHandle: "whiteantelopeweddings",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-venues/white-antelope-barn-4218279",
		},
		// Cheyenne jeweler
		{
			Name: "Burri Jewelers", OfficialURL: "https://burrijewelers.com/",
			Handle: "jburri1922", SourceClass: "jeweler",
			CityID: "city_cheyenne_wy", State: "WY", City: "Cheyenne", Verified: "2026-08-01",
		},
		{
			Name: "Three Crowns Golf Club", OfficialURL: "https://www.threecrownsgolfclub.com/",
			Handle: "threecrownsgolf", SourceClass: "wedding_venue",
			CityID: "city_casper_wy", State: "WY", City: "Casper", Verified: "2026-08-03",
		},
		{
			Name: "Trout On Inn", OfficialURL: "https://troutoninn.com/",
			Handle: "rusticelegance_at_troutoninn", SourceClass: "wedding_venue",
			CityID: "city_casper_wy", State: "WY", City: "Casper", Verified: "2026-08-03",
		},
		{
			Name: "Wild One Floral", OfficialURL: "https://www.wildonefloral.com/",
			Handle: "wildonefloral", SourceClass: "florist",
			CityID: "city_cheyenne_wy", State: "WY", City: "Cheyenne", Verified: "2026-08-03",
		},
	},
}
