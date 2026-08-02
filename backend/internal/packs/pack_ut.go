package packs

// Utah source pack — verified 2026-08-01.
//
// Government: Utah marriage records are held by the county clerk. Search URLs
// for the top 7 counties by population were verified against each county's
// official .gov site. Utah County and Weber County have dedicated online
// marriage license search portals; others are request-oriented.
//
// Church: the Diocese of Salt Lake City (the only Catholic diocese in Utah)
// was verified via USCCB + dioslc.org. Salt Lake City-area parishes were
// verified against the diocese parish directory (dioslc.org/parishes) + each
// parish's own website. Bulletin URLs verified by direct search for each
// parish's bulletin archive.
//
// Social: Instagram handles verified from each business's own public website
// social links. Verification date recorded per vendor.

var utPack = StatePack{
	State: "UT",

	// --- Cities ----------------------------------------------------------
	Cities: []CityDef{
		{ID: "city_salt_lake_city_ut", State: "UT", County: "49035", Name: "Salt Lake City",
			Lat: 40.7608, Lng: -111.8910, Markets: []string{"saltlakecity", "slc", "utah", "slccounty"}},
		{ID: "city_park_city_ut", State: "UT", County: "49043", Name: "Park City",
			Lat: 40.6461, Lng: -111.4980, Markets: []string{"parkcity", "summit", "utah"}},
	},

	// --- Government (county clerk marriage-record searches) --------------
	Government: []GovSource{
		{
			// Salt Lake County — county clerk marriage division page with
			// license appointment scheduling; historical records search at
			// the county archives (1887–1904).
			CountyFIPS: "49035",
			CourtName:  "Salt Lake County Clerk",
			CourtURL:   "https://www.saltlakecounty.gov/clerk/",
			SearchURL:  "https://www.saltlakecounty.gov/clerk/marriage/",
			Note:       "Marriage division page with license appointment scheduling; certified copies via request. Historical search at /archives/records-online/marriage-licenses/.",
		},
		{
			// Utah County — dedicated marriage license search portal,
			// searchable by name; 724K+ records.
			CountyFIPS: "49049",
			CourtName:  "Utah County Clerk",
			CourtURL:   "https://www.utahcounty.gov/dept/clerk/marriage/",
			SearchURL:  "https://www.utahcounty.gov/dept/clerk/marriage/licensesearch.html",
			Note:       "Online marriage license search by name; 724K+ records; enumeration candidate.",
		},
		{
			// Davis County — marriage licenses page with online search and
			// request form.
			CountyFIPS: "49011",
			CourtName:  "Davis County Clerk",
			CourtURL:   "https://daviscountyutah.gov/clerk/marriage-licenses",
			SearchURL:  "https://daviscountyutah.gov/clerk/marriage-licenses",
			Note:       "Search and request marriage records page; online search by name; enumeration capability needs testing.",
		},
		{
			// Weber County — marriage license search with records since 1887;
			// online search for records >75 years old; certified copies for
			// recent records.
			CountyFIPS: "49057",
			CourtName:  "Weber County Clerk/Auditor",
			CourtURL:   "https://webercountyutah.gov/Clerk_Auditor/marriage.php",
			SearchURL:  "https://webercountyutah.gov/Clerk_Auditor/Marriage_License/",
			Note:       "Marriage license search with records since 1887; online search for records >75 years old; certified copies for recent records.",
		},
		{
			// Cache County — marriage license info page; no online search
			// portal; request-oriented.
			CountyFIPS: "49005",
			CourtName:  "Cache County Clerk",
			CourtURL:   "https://www.cachecounty.gov/clerk/",
			SearchURL:  "https://www.cachecounty.gov/clerk/marriage-license/",
			Note:       "Marriage license info page; no online search portal; request-oriented.",
		},
		{
			// Washington County — online records search portal for marriage
			// records and business licenses.
			CountyFIPS: "49053",
			CourtName:  "Washington County Clerk",
			CourtURL:   "https://www.washco.utah.gov/departments/clerk/",
			SearchURL:  "https://qdocs.washco.utah.gov/clerk/web",
			Note:       "Online records search portal for marriage records and business licenses; enumeration capability needs testing.",
		},
		{
			// Tooele County — marriage licenses page with online application
			// and copy requests; no public search portal.
			CountyFIPS: "49045",
			CourtName:  "Tooele County Clerk",
			CourtURL:   "https://www.tooeleco.gov/departments/elected_officials/clerk/",
			SearchURL:  "https://www.tooeleco.gov/departments/administration/clerk/marriage_licenses.php",
			Note:       "Marriage licenses page with online application and copy requests; no public search portal.",
		},
	},

	// --- Church (Catholic dioceses + parishes) ---------------------------
	Dioceses: []DioceseDef{
		{Slug: "salt_lake_city", Name: "Diocese of Salt Lake City", Type: "diocese",
			Website: "https://www.dioslc.org", Directory: "https://www.dioslc.org/parishes", HubCityID: "city_salt_lake_city_ut"},
	},

	// Salt Lake City-area parishes in the Diocese of Salt Lake City.
	// Names and addresses verified from the diocese parish directory
	// (dioslc.org/parishes) + each parish's own website. Bulletin URLs
	// verified by direct search for each parish's bulletin archive.
	Parishes: []ParishDef{
		{
			DioceseSlug: "salt_lake_city", Name: "Cathedral of the Madeleine",
			Address:     "331 E. South Temple, Salt Lake City, UT 84111",
			BulletinURL: "https://www.utcotm.org/about/weekly-bulletins",
		},
		{
			DioceseSlug: "salt_lake_city", Name: "Saint Ambrose Catholic Church",
			Address:     "2315 E. Redondo Ave, Salt Lake City, UT 84108",
			BulletinURL: "https://www.stambrosecatholicchurch.org/bulletin",
		},
		{
			DioceseSlug: "salt_lake_city", Name: "Our Lady of Lourdes Catholic Church",
			Address:     "1085 E 700 S, Salt Lake City, UT 84102",
			BulletinURL: "https://lourdes-slc.org/content.cfm?id=9010",
		},
		{
			DioceseSlug: "salt_lake_city", Name: "Saint Ann Catholic Church",
			Address: "450 E 2100 S, Salt Lake City, UT 84115",
		},
		{
			DioceseSlug: "salt_lake_city", Name: "Saint Patrick Catholic Church",
			Address:     "1058 W 400 S, Salt Lake City, UT 84104",
			BulletinURL: "https://www.stpatrickslc.org/bulletins",
		},
		{
			DioceseSlug: "salt_lake_city", Name: "Blessed Sacrament Catholic Church",
			Address:     "9757 S 1700 East, Sandy, UT 84092",
			BulletinURL: "https://parishesonline.com/find/blessed-sacrament-church-84092",
			Aggregator:  true,
		},
		{
			DioceseSlug: "salt_lake_city", Name: "Saint Catherine of Siena Catholic Church",
			Address: "170 S. University St., Salt Lake City, UT 84102",
		},
	},

	// --- Social (wedding-industry Instagram vendors) ---------------------
	Vendors: []VendorDef{
		// Salt Lake City photographers
		{
			Name: "Blake Hogge Photography", OfficialURL: "https://blakehogge.com/",
			Handle: "blakehogge", SourceClass: "engagement_photographer",
			CityID: "city_salt_lake_city_ut", State: "UT", City: "Salt Lake City", Verified: "2026-08-01",
			TikTokHandle: "blakehogge",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-photographers/blake-hogge-photography-8801346",
		},
		{
			Name: "Branson Maxwell Photo & Video", OfficialURL: "https://www.bransonmaxwell.com/",
			Handle: "bransonmaxwell.photo", SourceClass: "engagement_photographer",
			CityID: "city_salt_lake_city_ut", State: "UT", City: "Salt Lake City", Verified: "2026-08-01",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-photographers/branson-maxwell-photo-video-1777933",
		},
		{
			Name: "Jordan Varela Photography", OfficialURL: "https://www.jordanvarela.com/",
			Handle: "jordanvarelaphotography", SourceClass: "engagement_photographer",
			CityID: "city_salt_lake_city_ut", State: "UT", City: "Salt Lake City", Verified: "2026-08-01",
			TikTokHandle: "jordanvarelaphotography",
		},
		{
			Name: "Haylee Baker Photo", OfficialURL: "https://hayleebakerphoto.com/",
			Handle: "hayleebakerphoto", SourceClass: "engagement_photographer",
			CityID: "city_salt_lake_city_ut", State: "UT", City: "Salt Lake City", Verified: "2026-08-01",
		},
		// Salt Lake City venues
		{
			Name: "Millcreek Inn", OfficialURL: "https://www.millcreekinn.com/",
			Handle: "millcreekinn", SourceClass: "wedding_venue",
			CityID: "city_salt_lake_city_ut", State: "UT", City: "Salt Lake City", Verified: "2026-08-01",
			TikTokHandle: "millcreekinn",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-venues/millcreek-inn-6121855",
		},
		{
			Name: "Siempre Utah", OfficialURL: "https://www.siempreutah.com/",
			Handle: "siempreutah", SourceClass: "wedding_venue",
			CityID: "city_salt_lake_city_ut", State: "UT", City: "Salt Lake City", Verified: "2026-08-01",
		},
		{
			Name: "The Hearth", OfficialURL: "https://www.thehearthutah.com/",
			Handle: "thehearthutah", SourceClass: "wedding_venue",
			CityID: "city_salt_lake_city_ut", State: "UT", City: "Salt Lake City", Verified: "2026-08-01",
			TikTokHandle: "thehearthutah",
		},
		// Salt Lake City jewelers
		{
			Name: "Morgan Jewelers", OfficialURL: "https://www.morganjewelers.com/",
			Handle: "officialmorganjewelers", SourceClass: "jeweler",
			CityID: "city_salt_lake_city_ut", State: "UT", City: "Salt Lake City", Verified: "2026-08-01",
		},
		{
			Name: "9th & 9th Jewelers", OfficialURL: "https://www.9thand9thjewelers.com/",
			Handle: "9thand9thjewelers", SourceClass: "jeweler",
			CityID: "city_salt_lake_city_ut", State: "UT", City: "Salt Lake City", Verified: "2026-08-01",
		},
		{
			Name: "Forge Jewelry Works", OfficialURL: "https://www.forgejewelryworks.com/",
			Handle: "forgejewelryworks", SourceClass: "jeweler",
			CityID: "city_salt_lake_city_ut", State: "UT", City: "Salt Lake City", Verified: "2026-08-01",
		},
	},
}
