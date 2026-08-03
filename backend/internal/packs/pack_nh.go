package packs

// New Hampshire source pack — verified 2026-08-01.
//
// Government: NH marriage records are held at the town/city clerk level.
// The state Division of Vital Records Administration (DVRA) operates
// NHVRIN, a statewide vital records index. Marriage records >50 years old
// are public domain; recent records require "direct and tangible interest."
//
// Church: Diocese of Manchester verified via USCCB + catholicnh.org.
// Manchester-area parishes verified against the diocese's own parish
// directory (directory.catholicnh.org) + each parish's own website.
// Bulletin URLs verified by direct search for each parish's bulletin page.
//
// Social: Instagram handles verified from each business's own public website
// social links (or from IG search results where the site is JS-rendered).

var nhPack = StatePack{
	State: "NH",

	// --- Cities ----------------------------------------------------------
	Cities: []CityDef{
		{ID: "city_manchester_nh", State: "NH", County: "33011", Name: "Manchester",
			Lat: 43.0087, Lng: -71.4541, Markets: []string{"manchester", "nh", "hillsborough", "newhampshire"}},
	},

	// --- Government (town/city clerk + statewide vital records) -----------
	Government: []GovSource{
		{
			// Hillsborough County (Manchester) — city clerk vital records.
			CountyFIPS: "33011",
			CourtName:  "Manchester City Clerk",
			CourtURL:   "https://www.manchesternh.gov/Departments/City-Clerk",
			SearchURL:  "https://www.manchesternh.gov/Departments/City-Clerk/Vital-Records-and-Genealogy",
			Note:       "City clerk holds marriage licenses & certificates; public records pre-1960, request for recent.",
		},
		{
			// Rockingham County — NH statewide vital records system (NHVRIN).
			CountyFIPS: "33015",
			CourtName:  "NH Division of Vital Records Administration",
			CourtURL:   "https://www.sos.nh.gov/vital-records-0",
			SearchURL:  "https://nhvrinweb.sos.nh.gov/nhivs_marriage_query.aspx",
			Note:       "Statewide NHVRIN marriage query; town clerks also hold records; access restricted to direct interest.",
		},
		{
			// Merrimack County — NH statewide vital records system (NHVRIN).
			CountyFIPS: "33013",
			CourtName:  "NH Division of Vital Records Administration",
			CourtURL:   "https://www.sos.nh.gov/vital-records-0",
			SearchURL:  "https://nhvrinweb.sos.nh.gov/nhivs_marriage_query.aspx",
			Note:       "Statewide NHVRIN marriage query; Concord City Clerk also holds local records.",
		},
		{
			// Strafford County — NH statewide vital records system (NHVRIN).
			CountyFIPS: "33017",
			CourtName:  "NH Division of Vital Records Administration",
			CourtURL:   "https://www.sos.nh.gov/vital-records-0",
			SearchURL:  "https://nhvrinweb.sos.nh.gov/nhivs_marriage_query.aspx",
			Note:       "Statewide NHVRIN marriage query; Dover City Clerk also holds local records.",
		},
	},

	// --- Church (Catholic dioceses + parishes) ---------------------------
	Dioceses: []DioceseDef{
		{Slug: "manchester", Name: "Diocese of Manchester", Type: "diocese",
			Website: "https://www.catholicnh.org", Directory: "https://www.catholicnh.org/parishes", HubCityID: "city_manchester_nh"},
	},

	// Manchester-area parishes in the Diocese of Manchester.
	// Names verified from the diocese's parish directory
	// (directory.catholicnh.org). Bulletin URLs verified by direct search.
	Parishes: []ParishDef{
		{
			DioceseSlug: "manchester", Name: "St. Joseph Cathedral Parish",
			Address:     "145 Lowell St, Manchester, NH 03104",
			BulletinURL: "https://www.stjosephcathedralnh.org/bulletin",
		},
		{
			DioceseSlug: "manchester", Name: "Blessed Sacrament Parish",
			Address:     "14 Elm St, Manchester, NH 03103",
			BulletinURL: "https://www.blessedsacramentnh.org/images/bulletins/",
		},
		{
			DioceseSlug: "manchester", Name: "St. Anthony of Padua Parish",
			Address: "172 Belmont St, Manchester, NH 03103",
		},
		{
			DioceseSlug: "manchester", Name: "St. Hedwig Parish",
			Address:     "147 Walnut St, Manchester, NH 03104",
			BulletinURL: "https://www.sainthedwignh.org/bulletins.html",
		},
	},

	// --- Social (wedding-industry Instagram vendors) ---------------------
	Vendors: []VendorDef{
		{
			Name: "Mike Indi Photography", OfficialURL: "https://www.mikeindiphotography.com/",
			Handle: "mike_indi_photography", SourceClass: "engagement_photographer",
			CityID: "city_manchester_nh", State: "NH", City: "Manchester", Verified: "2026-08-01",
			TikTokHandle: "mike_indi_photography",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-photographers/mike-indi-photography-8896515",
		},
		{
			Name: "The Venues at The Factory", OfficialURL: "https://www.thevenuesatthefactory.com/",
			Handle: "thevenuesatthefactory", SourceClass: "wedding_venue",
			CityID: "city_manchester_nh", State: "NH", City: "Manchester", Verified: "2026-08-01",
			TikTokHandle: "thevenuesatthefactory",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-venues/the-venues-at-the-factory-5015373",
		},
		{
			Name: "Bellman Jewelers", OfficialURL: "https://bellmans.com/",
			Handle: "bellmanjewelers", SourceClass: "jeweler",
			CityID: "city_manchester_nh", State: "NH", City: "Manchester", Verified: "2026-08-01",
		},
		{
			Name: "Studio Elizabeth", OfficialURL: "https://www.studioelizabeth.net/",
			Handle: "studioelizabethnh", SourceClass: "engagement_photographer",
			CityID: "city_manchester_nh", State: "NH", City: "Manchester", Verified: "2026-08-03",
		},
		{
			Name: "Flowers By Jennifer", OfficialURL: "https://www.flowersbyjennifer.com/",
			Handle: "flowersbyjennifernh", SourceClass: "florist",
			CityID: "city_manchester_nh", State: "NH", City: "Manchester", Verified: "2026-08-03",
		},
	},
}
