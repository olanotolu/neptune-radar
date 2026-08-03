package packs

// West Virginia source pack — verified 2026-08-01.
//
// Government: WV marriage records are held by each county clerk. The WV
// Culture Center's Vital Research Records project
// (archive.wvculture.org/vrr/va_mcsearch.aspx) provides a statewide
// searchable index of historical marriage records (varies by county, most
// through ~1970). Several counties also maintain their own IDX/online
// record-search portals; those are used as SearchURL where available.
//
// Church: the Diocese of Wheeling-Charleston covers the entire state.
// Parishes in Charleston and Wheeling were verified against each parish's
// own website + the diocese's parish directory. Bulletin URLs point at
// each parish's own bulletin archive page.
//
// Social: Instagram handles verified from each business's own public website
// social links (or from IG search results where the site is JS-rendered and
// the handle was visible in the search snippet). Verification date recorded
// per vendor.

var wvPack = StatePack{
	State: "WV",

	// --- Cities ----------------------------------------------------------
	Cities: []CityDef{
	{
		ID: "city_morgantown_wv", State: "WV", County: "54061", Name: "Morgantown",
		Lat: 39.6296, Lng: -79.956,
		Markets: []string{"morgantown", "monongalia"},
	},
		{ID: "city_charleston_wv", State: "WV", County: "54039", Name: "Charleston",
			Lat: 38.3498, Lng: -81.6326, Markets: []string{"charleston", "wv", "westvirginia", "kanawha"}},
		{ID: "city_huntington_wv", State: "WV", County: "54011", Name: "Huntington",
			Lat: 38.4192, Lng: -82.4452, Markets: []string{"huntington", "wv", "cabell"}},
	},

	// --- Government (county clerk marriage-record searches) --------------
	Government: []GovSource{
		{
			// Kanawha County (Charleston) — county clerk vital records; no
			// dedicated online marriage search portal, but the WV Culture
			// Center statewide vital-records index covers Kanawha through 1967.
			CountyFIPS: "54039",
			CourtName:  "Kanawha County Clerk",
			CourtURL:   "https://kanawha.us/county-clerk/",
			SearchURL:  "https://archive.wvculture.org/vrr/va_mcsearch.aspx",
			Note:       "County clerk holds marriage records; statewide historical index covers Kanawha 1789–1967; post-1967 records request-oriented.",
		},
		{
			// Cabell County (Huntington) — county clerk vital statistics page
			// + IDX online record search with Marriage category.
			CountyFIPS: "54011",
			CourtName:  "Cabell County Clerk",
			CourtURL:   "https://www.cabellcountyclerk.org/departments/vital_statistics/marriage_records_licenses.php",
			SearchURL:  "https://www.recordscabellcountyclerk.org/",
			Note:       "IDX online record search with Marriage book/record category; enumeration capability needs testing.",
		},
		{
			// Monongalia County (Morgantown) — county clerk vital statistics
			// + IDX online record search with Marriage category.
			CountyFIPS: "54061",
			CourtName:  "Monongalia County Clerk",
			CourtURL:   "https://www.monongaliacountyclerk.org/index.php/11-birth-death-marriage-certificates",
			SearchURL:  "https://searchrecords.monongaliacountyclerk.com/Default.aspx",
			Note:       "IDX online record search with Marriage license/book category; enumeration capability needs testing.",
		},
		{
			// Berkeley County (Martinsburg) — county clerk vital statistics
			// + electronic record search (subscription-based, $15/mo).
			CountyFIPS: "54003",
			CourtName:  "Berkeley County Clerk",
			CourtURL:   "https://berkeleywv.org/267/County-Clerk",
			SearchURL:  "https://www.berkeleywv.org/541/Electronic-Record-Search",
			Note:       "Electronic record search available by subscription; marriage records from 1781; enumeration capability needs testing.",
		},
		{
			// Wood County (Parkersburg) — county clerk holds marriage records
			// from 1801; online document imaging covers marriages from 1899.
			// Statewide historical index covers Wood through 1971.
			CountyFIPS: "54107",
			CourtName:  "Wood County Clerk",
			CourtURL:   "https://woodcountywv.com/county-offices/county-clerk/",
			SearchURL:  "https://archive.wvculture.org/vrr/va_mcsearch.aspx",
			Note:       "County clerk has marriage records from 1801; online document imaging covers 1899+; statewide historical index covers Wood through 1971.",
		},
		{
			// Raleigh County (Beckley) — county clerk vital statistics;
			// request-oriented, no dedicated online search portal.
			CountyFIPS: "54081",
			CourtName:  "Raleigh County Clerk",
			CourtURL:   "https://raleighcountyclerk.com/vital-statistics-2/",
			SearchURL:  "https://archive.wvculture.org/vrr/va_mcsearch.aspx",
			Note:       "County clerk holds vital records from 1850; request-oriented; statewide historical index covers Raleigh through 1971.",
		},
		{
			// Harrison County (Clarksburg) — county clerk + IDX online record
			// search with Marriage category.
			CountyFIPS: "54033",
			CourtName:  "Harrison County Clerk",
			CourtURL:   "https://www.harrisoncountywv.gov/county-clerk.html",
			SearchURL:  "http://lookup.harrisoncountywv.com/",
			Note:       "IDX online record search with Marriage record/bonds category; marriage records from 1784; enumeration capability needs testing.",
		},
	},

	// --- Church (Catholic dioceses + parishes) ---------------------------
	Dioceses: []DioceseDef{
		{Slug: "wheeling_charleston", Name: "Diocese of Wheeling-Charleston", Type: "diocese",
			Website: "https://dwc.org", Directory: "https://dwc.org/parishes", HubCityID: "city_charleston_wv"},
	},

	// Charleston- and Wheeling-area parishes in the Diocese of Wheeling-
	// Charleston. Names and addresses verified from each parish's own website.
	// Bulletin URLs verified by direct search for each parish's bulletin
	// archive.
	Parishes: []ParishDef{
		{
			DioceseSlug: "wheeling_charleston", Name: "Basilica of the Co-Cathedral of the Sacred Heart",
			Address:     "1114 Virginia St. E, Charleston, WV 25301",
			BulletinURL: "https://sacredheartcocathedral.com/bulletin-archive/",
		},
		{
			DioceseSlug: "wheeling_charleston", Name: "St. Anthony Parish",
			Address:     "1000 6th Street, Charleston, WV 25302",
			BulletinURL: "https://stanthonywv.com/about-us/",
		},
		{
			DioceseSlug: "wheeling_charleston", Name: "St. Agnes Catholic Church",
			Address:     "49th Street & Staunton Avenue, Charleston, WV 25304",
			BulletinURL: "https://stagnescharlestonwv.org/bulletins/",
		},
		{
			DioceseSlug: "wheeling_charleston", Name: "St. James the Greater",
			Address:     "49 Crosswinds Drive, Charles Town, WV 25414",
			BulletinURL: "https://www.stjameswv.org/bulletin9ab2b387",
		},
		{
			DioceseSlug: "wheeling_charleston", Name: "Cathedral of St. Joseph",
			Address:     "1300 Eoff Street, Wheeling, WV 26003",
			BulletinURL: "https://saintjosephcathedral.com/courier-archive/",
		},
		{
			DioceseSlug: "wheeling_charleston", Name: "St. Michael Parish",
			Address:     "1225 National Road, Wheeling, WV 26003",
			BulletinURL: "https://stmikesparish.org/bulletins/",
		},
	},

	// --- Social (wedding-industry Instagram vendors) ---------------------
	Vendors: []VendorDef{
		// Charleston photographers
		{
			Name: "Jessica Ellis Photography", OfficialURL: "https://jessicaellisphotography.com/",
			Handle: "jessicaellisphotography", SourceClass: "engagement_photographer",
			CityID: "city_charleston_wv", State: "WV", City: "Charleston", Verified: "2026-08-01",
			TikTokHandle: "jessicaellisphotography",
		},
		{
			Name: "Melissa Kincaid Photography", OfficialURL: "https://www.melissakincaidphoto.com/",
			Handle: "melissaweddingphotographer", SourceClass: "engagement_photographer",
			CityID: "city_charleston_wv", State: "WV", City: "Charleston", Verified: "2026-08-01",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-photographers/melissa-kincaid-photography-8802976",
		},
		{
			Name: "Erin Hurst Photography", OfficialURL: "https://www.erinhurstphotography.com/",
			Handle: "erinhurstphotography", SourceClass: "engagement_photographer",
			CityID: "city_charleston_wv", State: "WV", City: "Charleston", Verified: "2026-08-01",
			TikTokHandle: "erinhurstphotography",
		},
		// Charleston venues
		{
			Name: "Edgewood Country Club", OfficialURL: "https://www.edgewoodcc.com/private-events",
			Handle: "edgewoodccwv", SourceClass: "wedding_venue",
			CityID: "city_charleston_wv", State: "WV", City: "Charleston", Verified: "2026-08-01",
			TikTokHandle: "edgewoodccwv",
		},
		{
			Name: "The Greenbrier", OfficialURL: "https://www.greenbrier.com/gather/weddings/",
			Handle: "the_greenbrier", SourceClass: "wedding_venue",
			CityID: "city_charleston_wv", State: "WV", City: "Charleston", Verified: "2026-08-01",
			TikTokHandle: "the_greenbrier",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-venues/the-greenbrier-1192734",
		},
		// Huntington venues
		{
			Name: "Heritage Farm Museum & Village", OfficialURL: "https://www.heritagefarmwv.com/",
			Handle: "heritagefarmhtgtnwv", SourceClass: "wedding_venue",
			CityID: "city_huntington_wv", State: "WV", City: "Huntington", Verified: "2026-08-01",
		},
		// Huntington jeweler
		{
			Name: "T.K. Dodrill Jewelers", OfficialURL: "https://dodrilljewelers.com/",
			Handle: "dodrilljewelers", SourceClass: "jeweler",
			CityID: "city_huntington_wv", State: "WV", City: "Huntington", Verified: "2026-08-01",
		},
		{
			Name: "The Oberports", OfficialURL: "https://theoberports.com/",
			Handle: "theoberports", SourceClass: "engagement_photographer",
			CityID: "city_charleston_wv", State: "WV", City: "Charleston", Verified: "2026-08-03",
		},
		{
			Name: "Lakeview Golf Resort", OfficialURL: "https://www.lakeviewresort.com/",
			Handle: "lakeviewgolfwv", SourceClass: "wedding_venue",
			CityID: "city_morgantown_wv", State: "WV", City: "Morgantown", Verified: "2026-08-03",
		},
		{
			Name: "Hotel Morgan", OfficialURL: "https://hotelmorgan.com/",
			Handle: "thehotelmorgan", SourceClass: "wedding_venue",
			CityID: "city_morgantown_wv", State: "WV", City: "Morgantown", Verified: "2026-08-03",
		},
	},
}
