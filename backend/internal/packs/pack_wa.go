package packs

// Washington source pack — verified 2026-08-01.
//
// Government: Washington marriage records are held by the county auditor
// (recording division). Search URLs for the top 7 counties by population were
// verified against each county's official .gov site or its EagleWeb/Landmark
// search portal.
//
// Church: all 3 Washington Catholic dioceses/archdioceses verified via USCCB +
// each diocese's own website. Seattle-area parishes (Archdiocese of Seattle)
// were verified against the archdiocese's parish finder (archseattle.org) +
// direct bulletin-archive URL discovery for each parish's own site.
//
// Social: Instagram handles verified from each business's own public website
// social links (or from IG search results where the site is JS-rendered and
// the handle was visible in the search snippet). Verification date recorded
// per vendor.

var waPack = StatePack{
	State: "WA",

	// --- Cities ----------------------------------------------------------
	Cities: []CityDef{
		{ID: "city_seattle_wa", State: "WA", County: "53033", Name: "Seattle",
			Lat: 47.6062, Lng: -122.3321, Markets: []string{"seattle", "sea", "kingcounty", "pnw"}},
		{ID: "city_spokane_wa", State: "WA", County: "53063", Name: "Spokane",
			Lat: 47.6588, Lng: -117.4260, Markets: []string{"spokane", "spokanecounty", "inlandnw"}},
	},

	// --- Government (county auditor marriage-record searches) -------------
	Government: []GovSource{
		{
			// King County (Seattle) — Recorder's Office Landmark Web official
			// records search; marriage documents indexed alongside land records.
			CountyFIPS: "53033",
			CourtName:  "King County Recorder",
			CourtURL:   "https://www.kingcounty.gov/en/dept/executive-services/certificates-permits-licenses/records-licensing/recorders-office",
			SearchURL:  "https://recordsearch.kingcounty.gov/LandmarkWeb/search/index",
			Note:       "Landmark Web official records search; marriage filtering by document type needs testing.",
		},
		{
			// Pierce County (Tacoma) — Auditor marriage record search portal.
			CountyFIPS: "53053",
			CourtName:  "Pierce County Auditor",
			CourtURL:   "https://www.piercecountywa.gov/356/Marriage",
			SearchURL:  "https://armsweb.co.pierce.wa.us/Marriage/SearchEntry.aspx",
			Note:       "Dedicated marriage record search application; enumeration candidate.",
		},
		{
			// Snohomish County (Everett) — Auditor recorded documents search.
			CountyFIPS: "53061",
			CourtName:  "Snohomish County Auditor",
			CourtURL:   "https://snohomishcountywa.gov/278/Recording",
			SearchURL:  "https://www.snohomishcountywa.gov/1983/Recorded-Documents-Search",
			Note:       "Recorded documents search; marriage certificates searchable by name; enumeration capability needs testing.",
		},
		{
			// Spokane County — Auditor EagleWeb recorded document search with
			// dedicated marriage category.
			CountyFIPS: "53063",
			CourtName:  "Spokane County Auditor",
			CourtURL:   "https://www.spokanecounty.gov/299/Recording",
			SearchURL:  "https://recording.spokanecounty.org/recorder/eagleweb/customSearch.jsp?pageId=Marriage",
			Note:       "EagleWeb marriage-specific search page; enumeration candidate.",
		},
		{
			// Clark County (Vancouver) — Auditor Landmark Web official records
			// search; marriage licenses recorded from 1890-present.
			CountyFIPS: "53011",
			CourtName:  "Clark County Auditor",
			CourtURL:   "https://clark.wa.gov/auditor/marriage-license",
			SearchURL:  "https://e-docs.clark.wa.gov/LandmarkWeb",
			Note:       "Landmark Web official records search; marriage filtering by document type needs testing.",
		},
		{
			// Thurston County (Olympia) — Auditor EagleWeb document search.
			CountyFIPS: "53067",
			CourtName:  "Thurston County Auditor",
			CourtURL:   "https://www.thurstoncountywa.gov/departments/auditor/licensing-services/marriage-licenses",
			SearchURL:  "https://eagleweb.co.thurston.wa.us/thurstonrecorder/web/",
			Note:       "EagleWeb document search portal; marriage filtering by document type needs testing.",
		},
		{
			// Kitsap County (Port Orchard) — Auditor EagleWeb document search
			// with dedicated marriage category.
			CountyFIPS: "53035",
			CourtName:  "Kitsap County Auditor",
			CourtURL:   "https://www.kitsap.gov/auditor/Pages/recording.aspx",
			SearchURL:  "https://kcwaimg.kitsap.gov/recorder/eagleweb/customSearch.jsp?pageId=Marriages",
			Note:       "EagleWeb marriage-specific search page; enumeration candidate.",
		},
	},

	// --- Church (Catholic dioceses + parishes) ---------------------------
	Dioceses: []DioceseDef{
		{Slug: "seattle", Name: "Archdiocese of Seattle", Type: "archdiocese",
			Website: "https://seattlearchdiocese.org", Directory: "https://seattlearchdiocese.org/parishes", HubCityID: "city_seattle_wa"},
		{Slug: "spokane", Name: "Diocese of Spokane", Type: "diocese",
			Website: "https://www.dioceseofspokane.org", Directory: "https://www.dioceseofspokane.org/parishes", HubCityID: "city_spokane_wa"},
		{Slug: "yakima", Name: "Diocese of Yakima", Type: "diocese",
			Website: "https://www.yakimadiocese.org", Directory: "https://www.yakimadiocese.org/parishes"},
	},

	// Seattle-area parishes in the Archdiocese of Seattle. Names and
	// addresses verified from the archdiocese's parish finder
	// (archseattle.org) + each parish's own website. Bulletin URLs verified
	// by direct search for each parish's bulletin archive.
	Parishes: []ParishDef{
		{
			DioceseSlug: "seattle", Name: "St. James Cathedral",
			Address:     "804 Ninth Ave, Seattle, WA 98104",
			BulletinURL: "https://www.stjames-cathedral.org/Bulletin/bulletins.aspx",
		},
		{
			DioceseSlug: "seattle", Name: "St. Joseph Parish",
			Address:     "732 18th Ave E, Seattle, WA 98112",
			BulletinURL: "https://www.stjosephparish.org/943",
		},
		{
			DioceseSlug: "seattle", Name: "Blessed Sacrament Church",
			Address:     "5041 9th Ave NE, Seattle, WA 98105",
			BulletinURL: "https://www.blessed-sacrament.org/bulletin",
		},
		{
			DioceseSlug: "seattle", Name: "St. Therese Parish",
			Address:     "3416 E Marion St, Seattle, WA 98122",
			BulletinURL: "https://st-therese.cc/bulletin",
		},
		{
			DioceseSlug: "seattle", Name: "Holy Rosary Catholic Church",
			Address: "4210 SW Genesee St, Seattle, WA 98116",
		},
		{
			DioceseSlug: "seattle", Name: "Christ Our Hope Catholic Church",
			Address: "1902 2nd Ave, Seattle, WA 98101",
		},
		{
			DioceseSlug: "seattle", Name: "Our Lady of the Lake Parish",
			Address: "8900 35th Ave NE, Seattle, WA 98115",
		},
	},

	// --- Social (wedding-industry Instagram vendors) ---------------------
	Vendors: []VendorDef{
		// Seattle photographers
		{
			Name: "Weddings By Andre", OfficialURL: "https://www.weddingsbyandre.com/",
			Handle: "weddingsbyandre", SourceClass: "engagement_photographer",
			CityID: "city_seattle_wa", State: "WA", City: "Seattle", Verified: "2026-08-01",
			TikTokHandle: "weddingsbyandre",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-photographers/weddings-by-andre-3963899",
		},
		{
			Name: "Captured by Candace", OfficialURL: "https://capturedbycandacephoto.com/",
			Handle: "capturedbycandacephoto", SourceClass: "engagement_photographer",
			CityID: "city_seattle_wa", State: "WA", City: "Seattle", Verified: "2026-08-01",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-photographers/captured-by-candace-6129030",
		},
		{
			Name: "Natalie Jayne Photography", OfficialURL: "https://nataliejaynephotography.com/",
			Handle: "nataliejaynephotography", SourceClass: "engagement_photographer",
			CityID: "city_seattle_wa", State: "WA", City: "Seattle", Verified: "2026-08-01",
			TikTokHandle: "nataliejaynephotography",
		},
		{
			Name: "Emett Joseph Photography", OfficialURL: "https://emettjoseph.com/",
			Handle: "emettjoseph", SourceClass: "engagement_photographer",
			CityID: "city_seattle_wa", State: "WA", City: "Seattle", Verified: "2026-08-01",
		},
		// Seattle venues
		{
			Name: "Chateau Lill", OfficialURL: "https://chateaulill.com/",
			Handle: "chateaulill", SourceClass: "wedding_venue",
			CityID: "city_seattle_wa", State: "WA", City: "Seattle", Verified: "2026-08-01",
			TikTokHandle: "chateaulill",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-venues/chateau-lill-3712473",
		},
		{
			Name: "The Sanctuary at Lotte Hotel Seattle", OfficialURL: "https://www.sanctuaryweddings.com/",
			Handle: "lottehotel_seattle", SourceClass: "wedding_venue",
			CityID: "city_seattle_wa", State: "WA", City: "Seattle", Verified: "2026-08-01",
		},
		{
			Name: "Imperia Lake Union", OfficialURL: "https://imperiaseattle.com/",
			Handle: "imperialakeunion", SourceClass: "wedding_venue",
			CityID: "city_seattle_wa", State: "WA", City: "Seattle", Verified: "2026-08-01",
			TikTokHandle: "imperialakeunion",
		},
		// Seattle jewelers
		{
			Name: "L T Denny Jewelers", OfficialURL: "https://www.ltdenny.com/",
			Handle: "ltdenny", SourceClass: "jeweler",
			CityID: "city_seattle_wa", State: "WA", City: "Seattle", Verified: "2026-08-01",
		},
		{
			Name: "Valerie Madison Fine Jewelry", OfficialURL: "https://valeriemadison.com/",
			Handle: "valeriemadisonjewelry", SourceClass: "jeweler",
			CityID: "city_seattle_wa", State: "WA", City: "Seattle", Verified: "2026-08-01",
		},
		{
			Name: "Stórica Studio", OfficialURL: "https://storicastudio.com/",
			Handle: "storicastudio", SourceClass: "jeweler",
			CityID: "city_seattle_wa", State: "WA", City: "Seattle", Verified: "2026-08-01",
		},
	},
}
