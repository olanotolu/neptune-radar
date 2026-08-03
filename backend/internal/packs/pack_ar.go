package packs

// Arkansas source pack — verified 2026-08-01.
//
// Government: Arkansas marriage records are held by the county clerk. Most
// counties participate in the statewide CIS Arkansas marriage-license lookup
// (marriage.cisarkansas.com), which is the actual searchable portal; Pulaski
// County runs its own vitals search. Search URLs for the top 7 counties by
// population were verified against each county's official site + the CIS
// portal.
//
// Church: the Diocese of Little Rock covers the entire state of Arkansas
// (verified via USCCB + dolr.org). Little Rock-area parishes were verified
// against the DOLR parish directory + each parish's own website. Bulletin
// archive URLs set where a real, reachable archive was found.
//
// Social: Instagram handles verified from each business's own public website
// social links (or from IG search results / directory listings where the site
// is JS-rendered). Verification date recorded per vendor.

var arPack = StatePack{
	State: "AR",

	// --- Cities ----------------------------------------------------------
	Cities: []CityDef{
		{ID: "city_little_rock_ar", State: "AR", County: "05119", Name: "Little Rock",
			Lat: 34.7465, Lng: -92.2896, Markets: []string{"littlerock", "lr", "ar", "pulaski"}},
	},

	// --- Government (county clerk marriage-record searches) --------------
	Government: []GovSource{
		{
			// Pulaski County (Little Rock) — circuit/county clerk vitals
			// search portal with a dedicated marriage page.
			CountyFIPS: "05119",
			CourtName:  "Pulaski Circuit/County Clerk",
			CourtURL:   "https://www.pulaskiclerk.com",
			SearchURL:  "http://realestatesearch.pulaskiclerk.com/vitals/search.php?page=marriage",
			Note:       "Dedicated marriage-license search portal; enumeration candidate.",
		},
		{
			// Benton County — county clerk marriage licenses; search via the
			// statewide CIS Arkansas portal.
			CountyFIPS: "05007",
			CourtName:  "Benton County Clerk",
			CourtURL:   "https://bentoncountyar.gov/county-clerk/marriage-licenses/",
			SearchURL:  "https://marriage.cisarkansas.com/?C=Benton",
			Note:       "CIS statewide marriage-license lookup, Benton County selected; enumeration candidate.",
		},
		{
			// Washington County (Fayetteville) — county clerk marriage info;
			// search via the statewide CIS Arkansas portal.
			CountyFIPS: "05143",
			CourtName:  "Washington County Clerk",
			CourtURL:   "https://www.washingtoncountyar.gov/government/departments-a-e/county-clerk/marriage-information",
			SearchURL:  "https://marriage.cisarkansas.com/?c=washington",
			Note:       "CIS statewide marriage-license lookup, Washington County selected; enumeration candidate.",
		},
		{
			// Faulkner County (Conway) — county clerk marriage records; search
			// via the statewide CIS Arkansas portal.
			CountyFIPS: "05045",
			CourtName:  "Faulkner County Clerk",
			CourtURL:   "https://www.faulknercountyar.gov/government/departments/county-clerk/marriage-records/",
			SearchURL:  "https://marriage.cisarkansas.com/?C=Faulkner",
			Note:       "CIS statewide marriage-license lookup, Faulkner County selected; enumeration candidate.",
		},
		{
			// Saline County — county clerk marriage licenses; search via the
			// statewide CIS Arkansas portal.
			CountyFIPS: "05125",
			CourtName:  "Saline County Clerk",
			CourtURL:   "https://www.salinecounty.org/government/county_clerk/marriage/index.php",
			SearchURL:  "https://marriage.cisarkansas.com/?C=Saline",
			Note:       "CIS statewide marriage-license lookup, Saline County selected; enumeration candidate.",
		},
		{
			// Garland County (Hot Springs) — county clerk marriage licenses.
			// No online search portal located; copies are request-based
			// (in-person or mail-in form).
			CountyFIPS: "05051",
			CourtName:  "Garland County Clerk",
			CourtURL:   "https://www.garlandcounty.org/178/Marriage",
			SearchURL:  "https://www.garlandcounty.org/178/Marriage",
			Note:       "No online search; marriage-license copies by mail/in-person request form only.",
		},
		{
			// Craighead County (Jonesboro) — county clerk marriage licenses;
			// search via the statewide CIS Arkansas portal.
			CountyFIPS: "05031",
			CourtName:  "Craighead County Clerk",
			CourtURL:   "https://www.craigheadclerk.com/HowToMarriage",
			SearchURL:  "https://marriage.cisarkansas.com/?C=Craighead",
			Note:       "CIS statewide marriage-license lookup, Craighead County selected; enumeration candidate.",
		},
	},

	// --- Church (Catholic dioceses + parishes) ---------------------------
	Dioceses: []DioceseDef{
		{Slug: "little_rock", Name: "Diocese of Little Rock", Type: "diocese",
			Website: "https://www.dolr.org", Directory: "https://www.dolr.org/parishes", HubCityID: "city_little_rock_ar"},
	},

	// Little Rock-area parishes in the Diocese of Little Rock. Names and
	// addresses verified against the DOLR parish directory (dolr.org/parishes)
	// + each parish's own website. Bulletin archive URLs verified by direct
	// search for each parish's bulletin page.
	Parishes: []ParishDef{
		{
			DioceseSlug: "little_rock", Name: "Cathedral of St. Andrew",
			Address: "617 S. Louisiana St, Little Rock, AR 72201",
		},
		{
			DioceseSlug: "little_rock", Name: "Christ the King Church",
			Address:     "4000 N. Rodney Parham Rd, Little Rock, AR 72212",
			BulletinURL: "https://www.ctklr.org/bulletin-board",
		},
		{
			DioceseSlug: "little_rock", Name: "Our Lady of Good Counsel Church",
			Address: "1321 S. Van Buren St, Little Rock, AR 72204",
		},
		{
			DioceseSlug: "little_rock", Name: "Our Lady of the Holy Souls Church",
			Address:     "1003 N. Tyler St, Little Rock, AR 72205",
			BulletinURL: "https://holysouls.org/sunday-bulletins",
		},
		{
			DioceseSlug: "little_rock", Name: "St. Theresa Church",
			Address: "6219 Baseline Rd, Little Rock, AR 72209",
		},
		{
			DioceseSlug: "little_rock", Name: "St. Edward Church",
			Address:     "801 Sherman St, Little Rock, AR 72202",
			BulletinURL: "https://www.stedwardchurchlr.org/sunday-bulletin/",
		},
		{
			DioceseSlug: "little_rock", Name: "St. Bartholomew Church",
			Address:     "1622 Marshall St, Little Rock, AR 72202",
			BulletinURL: "https://stbartholomewchurch-lr.org/church-bulletin",
		},
	},

	// --- Social (wedding-industry Instagram vendors) ---------------------
	Vendors: []VendorDef{
		// Little Rock photographers
		{
			Name: "Hannah Lee Co", OfficialURL: "https://hannahleeco.net/",
			Handle: "hannahleeco", SourceClass: "engagement_photographer",
			CityID: "city_little_rock_ar", State: "AR", City: "Little Rock", Verified: "2026-08-01",
			TikTokHandle: "hannahleeco",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-photographers/hannah-lee-co-4844043",
		},
		{
			Name: "Bailey Burton Photography", OfficialURL: "https://baileyburtonphoto.com/",
			Handle: "baileyburtonphoto", SourceClass: "engagement_photographer",
			CityID: "city_little_rock_ar", State: "AR", City: "Little Rock", Verified: "2026-08-01",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-photographers/bailey-burton-photography-4068415",
		},
		{
			Name: "Bethany Grace Photography", OfficialURL: "https://www.bethanygrace.photography/",
			Handle: "bethanygracecummings", SourceClass: "engagement_photographer",
			CityID: "city_little_rock_ar", State: "AR", City: "Little Rock", Verified: "2026-08-01",
			TikTokHandle: "bethanygracecummings",
		},
		{
			Name: "The Linns", OfficialURL: "https://thelinns.co/",
			Handle: "thelinns.co", SourceClass: "engagement_photographer",
			CityID: "city_little_rock_ar", State: "AR", City: "Little Rock", Verified: "2026-08-01",
		},
		// Little Rock venues
		{
			Name: "The DeCantillon", OfficialURL: "https://decantillon.com/",
			Handle: "thedecantillon", SourceClass: "wedding_venue",
			CityID: "city_little_rock_ar", State: "AR", City: "Little Rock", Verified: "2026-08-01",
			TikTokHandle: "thedecantillon",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-venues/the-decantillon-1553056",
		},
		{
			Name: "An Enchanting Evening", OfficialURL: "https://anenchantingevening.com/",
			Handle: "anenchantingevening", SourceClass: "wedding_venue",
			CityID: "city_little_rock_ar", State: "AR", City: "Little Rock", Verified: "2026-08-01",
		},
		{
			Name: "The Empress of Little Rock", OfficialURL: "https://www.theempress.com/",
			Handle: "theempressoflr", SourceClass: "wedding_venue",
			CityID: "city_little_rock_ar", State: "AR", City: "Little Rock", Verified: "2026-08-01",
			TikTokHandle: "theempressoflr",
		},
		// Little Rock jewelers
		{
			Name: "Danwerke Jewelers", OfficialURL: "https://www.danwerkejewelers.com/",
			Handle: "danwerkejewelers", SourceClass: "jeweler",
			CityID: "city_little_rock_ar", State: "AR", City: "Little Rock", Verified: "2026-08-01",
		},
		{
			Name: "Jones & Son Fine Jewelry", OfficialURL: "https://jonesandson.com/",
			Handle: "jonesandson", SourceClass: "jeweler",
			CityID: "city_little_rock_ar", State: "AR", City: "Little Rock", Verified: "2026-08-01",
		},
		{
			Name: "Tyler Rosenthal Photography", OfficialURL: "https://www.tylerrosenthalphotography.com/",
			Handle: "tylerrosenthalphotography", SourceClass: "engagement_photographer",
			CityID: "city_little_rock_ar", State: "AR", City: "Little Rock", Verified: "2026-08-03",
		},
		{
			Name: "Grandeur House", OfficialURL: "https://grandeurhouse.com/",
			Handle: "grandeurhouse", SourceClass: "wedding_venue",
			CityID: "city_little_rock_ar", State: "AR", City: "Little Rock", Verified: "2026-08-03",
		},
		{
			Name: "Alda's Magnolia Hill", OfficialURL: "https://aldasmagnoliahill.com/",
			Handle: "aldasmagnoliahill", SourceClass: "wedding_venue",
			CityID: "city_little_rock_ar", State: "AR", City: "Little Rock", Verified: "2026-08-03",
		},
	},
}
