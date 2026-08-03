package packs

// Michigan source pack — verified 2026-08-01.
//
// Government: Michigan marriage records are held by the county clerk. Search
// URLs for the top 8 counties by population were verified against each county's
// official .gov / .org site or its third-party search portal (fidlar, OakGnlg).
//
// Church: all 7 Michigan Catholic dioceses/archdioceses verified via USCCB +
// each diocese's own website. Parish directory URLs point at each
// jurisdiction's own parish-finder. Detroit-area parishes (Archdiocese of
// Detroit) were verified against the archdiocese's own parish directory
// (aod.org/parishes) + direct bulletin-archive URL discovery.
//
// Social: Instagram handles verified from each business's own public website
// social links (or from IG search results where the site is JS-rendered and
// the handle was visible in the search snippet). Verification date recorded
// per vendor.

var miPack = StatePack{
	State: "MI",

	// --- Cities ----------------------------------------------------------
	Cities: []CityDef{
		{
			ID: "city_detroit_mi", State: "MI", County: "26163", Name: "Detroit",
			Lat: 42.3314, Lng: -83.0458,
			Markets: []string{"detroit", "detroitmi", "metrodet", "wayne", "midtown"},
		},
		{
			ID: "city_grand_rapids_mi", State: "MI", County: "26081", Name: "Grand Rapids",
			Lat: 42.9634, Lng: -85.6681,
			Markets: []string{"grandrapids", "gr", "westmichigan", "wmi", "kentcounty"},
		},
	},

	// --- Government (county clerk marriage-record searches) --------------
	Government: []GovSource{
		{
			// Wayne County (Detroit) — county clerk marriage license page.
			// Certified copies available online via VitalChek; no public
			// search portal, but the marriage license page is the official
			// entry point for records requests.
			CountyFIPS: "26163",
			CourtName:  "Wayne County Clerk",
			CourtURL:   "https://www.waynecountymi.gov/Government/Elected-Officials/Clerk/General-Office/Marriage-Licenses",
			SearchURL:  "https://www.waynecountymi.gov/Government/Elected-Officials/Clerk/General-Office/Marriage-Licenses",
			Note:       "Marriage license page with online ordering via VitalChek; no public browse search; enumeration needs FOIA or VitalChek API.",
		},
		{
			// Oakland County — county clerk genealogy search service for
			// marriage and death records, searchable online.
			CountyFIPS: "26125",
			CourtName:  "Oakland County Clerk",
			CourtURL:   "https://www.oakgov.com/government/clerk-register-of-deeds/life-events-services/marriage-records",
			SearchURL:  "https://courts.oakgov.com/OakGnlg/",
			Note:       "Genealogy search service for marriage + death records; free search, fee for certified copies; enumeration candidate.",
		},
		{
			// Macomb County — county clerk marriage license + genealogy page.
			// Marriage records 1819–1925 open for inspection; indexes through
			// 1962 viewable.
			CountyFIPS: "26099",
			CourtName:  "Macomb County Clerk",
			CourtURL:   "https://www.macombgov.org/departments/clerk-register-deeds/marriage-licenses",
			SearchURL:  "https://www.macombgov.org/node/31/genealogy",
			Note:       "Genealogy page with marriage record indexes 1819–1962; no online search portal; enumeration needs in-person or mail request.",
		},
		{
			// Kent County (Grand Rapids) — county clerk marriage records with
			// online self-service search portal (deedsselfservice).
			CountyFIPS: "26081",
			CourtName:  "Kent County Clerk",
			CourtURL:   "https://www.kentcountymi.gov/807/Marriage-Records",
			SearchURL:  "https://deedsselfservice.kentcountymi.gov/clerkweb/search/DOCSEARCH248S2",
			Note:       "Online self-service marriage record search portal; enumeration candidate.",
		},
		{
			// Genesee County (Flint) — county clerk vital records search for
			// death, marriage, notary and DBA records.
			CountyFIPS: "26049",
			CourtName:  "Genesee County Clerk",
			CourtURL:   "https://www.geneseecountymi.gov/departments/county_clerk/vital_records_division/index.php",
			SearchURL:  "https://www.geneseecountymi.gov/departments/county_clerk/vital_records_division/death,_marriage,_dba_search.php",
			Note:       "Online search for marriage records by name; enumeration candidate.",
		},
		{
			// Washtenaw County (Ann Arbor) — county clerk certified copies of
			// marriage licenses + records search page.
			CountyFIPS: "26161",
			CourtName:  "Washtenaw County Clerk",
			CourtURL:   "https://www.washtenaw.org/2342/Certified-Copies-of-Marriage-Licenses",
			SearchURL:  "https://www.washtenaw.org/302/Search-Request-Records",
			Note:       "Search & request records page; marriage-record search capability needs testing.",
		},
		{
			// Ingham County (Lansing) — county clerk vital records with online
			// marriage and death record search (1959/1961–present).
			CountyFIPS: "26065",
			CourtName:  "Ingham County Clerk",
			CourtURL:   "https://clerk.ingham.org/departments_and_officials/county_clerk/vital_records.php",
			SearchURL:  "https://clerk.ingham.org/departments_and_officials/county_clerk/genealogy_research.php",
			Note:       "Online record search for marriage records 1961–present; enumeration candidate for recent records.",
		},
		{
			// Ottawa County — county clerk marriage records with online search
			// via fidlar Apex WebPortal + online ordering.
			CountyFIPS: "26139",
			CourtName:  "Ottawa County Clerk",
			CourtURL:   "https://miottawa.org/clerk/vital-records/marriage/",
			SearchURL:  "https://miottawa.fidlar.com/MIOttawa/Apex.WebPortal/search",
			Note:       "fidlar Apex WebPortal for marriage record search; enumeration candidate.",
		},
		{CountyFIPS: "26115", CourtName: "Monroe County Clerk",
			CourtURL:  "https://co.monroe.mi.us/",
			SearchURL: "https://fidlar.monroemi.org/MIMonroeVS/Apex.WebPortal/applications",
			Note:      "Online marriage license application portal; apply online then visit Clerk's office to complete."},
	},

	// --- Church (Catholic dioceses + parishes) ---------------------------
	Dioceses: []DioceseDef{
		{
			Slug: "detroit", Name: "Archdiocese of Detroit",
			Type: "archdiocese", Website: "https://www.aod.org",
			Directory: "https://www.aod.org/parishes",
			HubCityID: "city_detroit_mi",
		},
		{
			Slug: "grand_rapids", Name: "Diocese of Grand Rapids",
			Type: "diocese", Website: "https://grdiocese.org",
			Directory: "https://grdiocese.org/parishes/",
			HubCityID: "city_grand_rapids_mi",
		},
		{
			Slug: "kalamazoo", Name: "Diocese of Kalamazoo",
			Type: "diocese", Website: "https://diokzoo.org",
			Directory: "https://diokzoo.org/parishfinder",
		},
		{
			Slug: "lansing", Name: "Diocese of Lansing",
			Type: "diocese", Website: "https://www.dioceseoflansing.org",
			Directory: "https://www.dioceseoflansing.org/find-a-parish",
		},
		{
			Slug: "saginaw", Name: "Diocese of Saginaw",
			Type: "diocese", Website: "https://saginaw.org",
			Directory: "https://saginaw.org/churches",
		},
		{
			Slug: "gaylord", Name: "Diocese of Gaylord",
			Type: "diocese", Website: "https://dioceseofgaylord.org",
			Directory: "https://dioceseofgaylord.org/find-a-parish",
		},
		{
			Slug: "marquette", Name: "Diocese of Marquette",
			Type: "diocese", Website: "https://dioceseofmarquette.org",
			Directory: "https://dioceseofmarquette.org/directory",
		},
	},

	// Detroit-area parishes in the Archdiocese of Detroit.
	// Names + addresses verified from the archdiocese's own parish directory
	// (aod.org/parishes) and each parish's own website. Bulletin URLs verified
	// by direct search for each parish's bulletin archive.
	Parishes: []ParishDef{
		{
			DioceseSlug: "detroit", Name: "Cathedral of the Most Blessed Sacrament",
			Address:     "9844 Woodward Ave, Detroit, MI 48202",
			BulletinURL: "https://www.parishesonline.com/organization/cathedral-of-the-most-blessed-sacrament",
			Aggregator:  true,
		},
		{
			DioceseSlug: "detroit", Name: "St. Aloysius Parish",
			Address:     "1234 Washington Blvd, Detroit, MI 48226",
			BulletinURL: "https://www.staloysiusdetroit.com/bulletin",
		},
		{
			DioceseSlug: "detroit", Name: "Sacred Heart Church",
			Address:     "1000 Eliot St, Detroit, MI 48207",
			BulletinURL: "https://sacredheartdetroit.com/bulletins/",
		},
		{
			DioceseSlug: "detroit", Name: "St. Charles Borromeo Parish",
			Address:     "1491 Baldwin St, Detroit, MI 48214",
			BulletinURL: "https://www.stcharlesdetroit.org/bulletins/",
		},
		{
			DioceseSlug: "detroit", Name: "St. Jude Parish",
			Address:     "15889 E. Seven Mile Rd, Detroit, MI 48205",
			BulletinURL: "https://stjudedetroit.org/download-our-church-bulletins/",
		},
		{
			DioceseSlug: "detroit", Name: "St. Matthew Catholic Church",
			Address:     "6021 Whittier Ave, Detroit, MI 48224",
			BulletinURL: "https://www.stmatthewdetroit.com/parish-bulletin.php",
		},
		{
			DioceseSlug: "detroit", Name: "St. Elizabeth Catholic Church",
			Address:     "3138 E. Canfield St, Detroit, MI 48207",
			BulletinURL: "https://www.stelizabethdetroit.org/bulletins",
		},
		{
			DioceseSlug: "detroit", Name: "St. Joseph Shrine",
			Address:     "1828 Jay St, Detroit, MI 48207",
			BulletinURL: "https://institute-christ-king.org/detroit-bulletins",
		},
		{
			DioceseSlug: "detroit", Name: "National Shrine of the Little Flower Basilica",
			Address:     "2123 Roseland Ave, Royal Oak, MI 48073",
			BulletinURL: "https://shrinechurch.com/bulletin",
		},
		{
			DioceseSlug: "detroit", Name: "St. Mary, Cause of Our Joy Catholic Church",
			Address:     "8200 N. Wayne Rd, Westland, MI 48185",
			BulletinURL: "https://stmarycooj.org/our-parish/joyful-news/",
		},
	},

	// --- Social (wedding-industry Instagram vendors) ---------------------
	Vendors: []VendorDef{
		// Detroit photographers
		{
			Name: "Rosy & Shaun Wedding Photography", OfficialURL: "https://rosyandshaun.com/",
			Handle: "rosyandshaun", SourceClass: "engagement_photographer",
			CityID: "city_detroit_mi", State: "MI", City: "Detroit", Verified: "2026-08-01",
			TikTokHandle: "rosyandshaun",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-photographers/rosy-shaun-wedding-photography-6241779",
		},
		{
			Name: "Samantha Sutarova Photography", OfficialURL: "https://samanthasutarova.com/",
			Handle: "samanthasutarova", SourceClass: "engagement_photographer",
			CityID: "city_detroit_mi", State: "MI", City: "Detroit", Verified: "2026-08-01",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-photographers/samantha-sutarova-photography-3614842",
		},
		{
			Name: "Mariah Kasten Photo", OfficialURL: "https://mariahkastenphoto.com/",
			Handle: "mariahkastenphoto", SourceClass: "engagement_photographer",
			CityID: "city_detroit_mi", State: "MI", City: "Detroit", Verified: "2026-08-01",
			TikTokHandle: "mariahkastenphoto",
		},
		{
			Name: "Sheree Danel Photography", OfficialURL: "https://shereedanel.com/",
			Handle: "shereedanel", SourceClass: "engagement_photographer",
			CityID: "city_detroit_mi", State: "MI", City: "Detroit", Verified: "2026-08-01",
		},
		// Detroit venues
		{
			Name: "The Roostertail", OfficialURL: "https://roostertail.com/",
			Handle: "roostertail", SourceClass: "wedding_venue",
			CityID: "city_detroit_mi", State: "MI", City: "Detroit", Verified: "2026-08-01",
			TikTokHandle: "roostertail",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-venues/the-roostertail-6027028",
		},
		{
			Name: "Garden Theater Detroit", OfficialURL: "https://www.thegardendetroit.com/",
			Handle: "gardentheaterdetroit", SourceClass: "wedding_venue",
			CityID: "city_detroit_mi", State: "MI", City: "Detroit", Verified: "2026-08-01",
		},
		{
			Name: "Shinola Hotel", OfficialURL: "https://www.shinolahotel.com/",
			Handle: "shinolahotel", SourceClass: "wedding_venue",
			CityID: "city_detroit_mi", State: "MI", City: "Detroit", Verified: "2026-08-01",
			TikTokHandle: "shinolahotel",
		},
		{
			Name: "The Imperial House", OfficialURL: "https://theimperialhouse.com/",
			Handle: "theimperialhouse", SourceClass: "wedding_venue",
			CityID: "city_detroit_mi", State: "MI", City: "Detroit", Verified: "2026-08-01",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-venues/the-imperial-house-3783100",
		},
		// Detroit jeweler
		{
			Name: "Chapman's Jewelry", OfficialURL: "https://chapmansjewelry.com/",
			Handle: "chapmansjewelry", SourceClass: "jeweler",
			CityID: "city_detroit_mi", State: "MI", City: "Detroit", Verified: "2026-08-01",
		},
		// Grand Rapids photographers
		{
			Name: "Kelly Sweet Photography", OfficialURL: "https://www.kellysweet.com/",
			Handle: "kellysweetphoto", SourceClass: "engagement_photographer",
			CityID: "city_grand_rapids_mi", State: "MI", City: "Grand Rapids", Verified: "2026-08-01",
			TikTokHandle: "kellysweetphoto",
		},
		{
			Name: "Kara Hanes Photography", OfficialURL: "https://karahanesphotography.com/",
			Handle: "karahanesphotography", SourceClass: "engagement_photographer",
			CityID: "city_grand_rapids_mi", State: "MI", City: "Grand Rapids", Verified: "2026-08-01",
		},
		// Grand Rapids venue
		{
			Name: "The Goei Center", OfficialURL: "https://www.thegoeicenter.com/",
			Handle: "goeicentergr", SourceClass: "wedding_venue",
			CityID: "city_grand_rapids_mi", State: "MI", City: "Grand Rapids", Verified: "2026-08-01",
			TikTokHandle: "goeicentergr",
		},
		// Detroit wedding planners
		{
			Name: "Blissfully Organized Events", OfficialURL: "https://blissfullyorganizedevents.com/",
			Handle: "blissfully.organized.events", SourceClass: "wedding_planner",
			CityID: "city_detroit_mi", State: "MI", City: "Detroit", Verified: "2026-08-01",
		},
		{
			Name: "Detroit Cultivated", OfficialURL: "https://www.detroitcultivated.com/",
			Handle: "detroitcultivated", SourceClass: "wedding_planner",
			CityID: "city_detroit_mi", State: "MI", City: "Detroit", Verified: "2026-08-01",
		},
	},
}
