package packs

// Nevada source pack — verified 2026-08-01.
//
// Government: Nevada marriage records are held by the county clerk (license)
// and county recorder (certificate). Search URLs for Clark, Washoe, Carson
// City, Douglas, and Lyon counties were verified against each county's
// official .gov site or its records-search portal.
//
// Church: both Nevada dioceses verified via USCCB + each diocese's own
// website. Reno-area parishes (Diocese of Reno) verified against the
// diocese's parish directory (highdesertcatholic.org) + Wikipedia's list of
// churches in the Diocese of Reno. Bulletin URLs verified by direct fetch of
// each parish's own bulletin-archive page.
//
// Social: Instagram handles verified from each business's own public website
// social links. Verification date recorded per vendor.

var nvPack = StatePack{
	State: "NV",

	// --- Cities ----------------------------------------------------------
	Cities: []CityDef{
		{ID: "city_las_vegas_nv", State: "NV", County: "32003", Name: "Las Vegas",
			Lat: 36.1699, Lng: -115.1398, Markets: []string{"lasvegas", "vegas", "lv", "clark", "nv"}},
		{ID: "city_reno_nv", State: "NV", County: "32031", Name: "Reno",
			Lat: 39.5296, Lng: -119.8138, Markets: []string{"reno", "washoe", "nv"}},
	},

	// --- Government (county clerk / recorder marriage-record searches) ----
	Government: []GovSource{
		{
			// Clark County (Las Vegas) — county clerk AcclaimWeb record search
			// system with a dedicated marriage-record search page.
			CountyFIPS: "32003",
			CourtName:  "Clark County Clerk",
			CourtURL:   "https://clerk.clarkcountynv.gov/AcclaimWeb/",
			SearchURL:  "https://clerk.clarkcountynv.gov/AcclaimWeb/Marriage/FindMyMarriageRecordSearch",
			Note:       "Dedicated marriage-record search by name; enumeration candidate.",
		},
		{
			// Washoe County (Reno) — county clerk marriage license bureau
			// search page.
			CountyFIPS: "32031",
			CourtName:  "Washoe County Clerk",
			CourtURL:   "https://washoecounty.gov/clerks/mlb/",
			SearchURL:  "https://washoecounty.gov/clerks/mlb/search_marriage_records.php",
			Note:       "Marriage application search portal; enumeration capability needs testing.",
		},
		{
			// Carson City — clerk-recorder marriage department; records
			// searchable via Landmark Web portal.
			CountyFIPS: "32510",
			CourtName:  "Carson City Clerk-Recorder",
			CourtURL:   "https://www.carsoncity.gov/government/departments-a-f/clerk-recorder/marriage-department",
			SearchURL:  "https://landmark.carsoncity.gov/LandmarkWeb/",
			Note:       "Landmark Web document search with marriage category; enumeration capability needs testing.",
		},
		{
			// Douglas County — recorder's office marriage search portal.
			CountyFIPS: "32005",
			CourtName:  "Douglas County Recorder",
			CourtURL:   "https://www.douglascountynv.gov/GOVERNMENT/departments/recorder",
			SearchURL:  "https://recorder-search.douglasnv.us/Recording",
			Note:       "Dedicated marriage search page; index only 1949-present; enumeration candidate.",
		},
		{
			// Lyon County — recorder's office self-service records search;
			// marriage certificates filterable by document type.
			CountyFIPS: "32019",
			CourtName:  "Lyon County Recorder",
			CourtURL:   "https://www.lyon-county.org/108/Recorder",
			SearchURL:  "https://records.lyon-county.org/web/",
			Note:       "Self-service search; choose Document Type 'Marriage Certificate'; enumeration capability needs testing.",
		},
	},

	// --- Church (Catholic dioceses + parishes) ---------------------------
	Dioceses: []DioceseDef{
		{Slug: "las_vegas", Name: "Archdiocese of Las Vegas", Type: "archdiocese",
			Website: "https://www.lasvegasdiocese.org", Directory: "https://www.lasvegasdiocese.org/parishes", HubCityID: "city_las_vegas_nv"},
		{Slug: "reno", Name: "Diocese of Reno", Type: "diocese",
			Website: "https://www.renodiocese.org", Directory: "https://www.renodiocese.org/parishes", HubCityID: "city_reno_nv"},
	},

	// Reno-area parishes in the Diocese of Reno. Names and addresses verified
	// from the diocese's parish directory (highdesertcatholic.org) + Wikipedia's
	// list of churches in the Diocese of Reno. Bulletin URLs verified by direct
	// fetch of each parish's own bulletin-archive page.
	Parishes: []ParishDef{
		{
			DioceseSlug: "reno", Name: "St. Thomas Aquinas Cathedral",
			Address:     "310 W 2nd St, Reno, NV 89503",
			BulletinURL: "https://stacathedral.com/weekly-bulletins",
		},
		{
			DioceseSlug: "reno", Name: "Our Lady of the Snows Catholic Church",
			Address:     "1138 Wright St, Reno, NV 89509",
			BulletinURL: "https://www.olsparish.com/bulletin/",
		},
		{
			DioceseSlug: "reno", Name: "St. Albert the Great Catholic Church",
			Address:     "1250 Wyoming Ave, Reno, NV 89503",
			BulletinURL: "https://www.stalbertreno.org/bulletin",
		},
		{
			DioceseSlug: "reno", Name: "Holy Cross Catholic Church",
			Address: "5650 Vista Blvd, Sparks, NV 89436",
		},
		{
			DioceseSlug: "reno", Name: "Immaculate Conception Catholic Church",
			Address: "2900 N McCarran Blvd, Sparks, NV 89431",
		},
		{
			DioceseSlug: "reno", Name: "St. Gall Catholic Community",
			Address:     "1343 Centerville Ln, Gardnerville, NV 89410",
			BulletinURL: "https://saintgall.org/weekly_bulletin_new",
		},
		{
			DioceseSlug: "reno", Name: "Our Lady of Tahoe Catholic Church",
			Address: "1 Elks Point Rd, Zephyr Cove, NV 89448",
		},
		{
			DioceseSlug: "reno", Name: "Corpus Christi Catholic Church",
			Address: "3597 N Sunridge Dr, Carson City, NV 89706",
		},
	},

	// --- Social (wedding-industry Instagram vendors) ---------------------
	Vendors: []VendorDef{
		// Las Vegas photographers
		{
			Name: "Tasha Caliz Photo", OfficialURL: "https://www.tashacalizphoto.com/",
			Handle: "tashacaliz", SourceClass: "engagement_photographer",
			CityID: "city_las_vegas_nv", State: "NV", City: "Las Vegas", Verified: "2026-08-01",
			TikTokHandle: "tashacaliz",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-photographers/tasha-caliz-photo-3224115",
		},
		{
			Name: "Karissa Russ & Co", OfficialURL: "https://karissaruss.co/",
			Handle: "karissaruss.co", SourceClass: "engagement_photographer",
			CityID: "city_las_vegas_nv", State: "NV", City: "Las Vegas", Verified: "2026-08-01",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-photographers/karissa-russ-co-7923108",
		},
		{
			Name: "Katelyn Faye Photography", OfficialURL: "https://katelynfaye.com/",
			Handle: "katelynfayephoto", SourceClass: "engagement_photographer",
			CityID: "city_las_vegas_nv", State: "NV", City: "Las Vegas", Verified: "2026-08-01",
			TikTokHandle: "katelynfayephoto",
		},
		{
			Name: "Mackenna D Photography", OfficialURL: "https://www.mackennad.com/",
			Handle: "mackennadphotography", SourceClass: "engagement_photographer",
			CityID: "city_las_vegas_nv", State: "NV", City: "Las Vegas", Verified: "2026-08-01",
		},
		// Las Vegas venues
		{
			Name: "Sunset Gardens", OfficialURL: "https://sunsetgardens.com/",
			Handle: "sunsetgardensweddings", SourceClass: "wedding_venue",
			CityID: "city_las_vegas_nv", State: "NV", City: "Las Vegas", Verified: "2026-08-01",
			TikTokHandle: "sunsetgardensweddings",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-venues/sunset-gardens-5290326",
		},
		{
			Name: "Chapel of the Flowers", OfficialURL: "https://www.littlechapel.com/",
			Handle: "littlechapel", SourceClass: "wedding_venue",
			CityID: "city_las_vegas_nv", State: "NV", City: "Las Vegas", Verified: "2026-08-01",
		},
		// Las Vegas jewelers
		{
			Name: "The Jewelers of Las Vegas", OfficialURL: "https://www.thejewelers.com/",
			Handle: "thejewelersoflasvegas", SourceClass: "jeweler",
			CityID: "city_las_vegas_nv", State: "NV", City: "Las Vegas", Verified: "2026-08-01",
		},
		{
			Name: "Lyght Jewelers", OfficialURL: "https://www.lyght.com/",
			Handle: "lyght", SourceClass: "jeweler",
			CityID: "city_las_vegas_nv", State: "NV", City: "Las Vegas", Verified: "2026-08-01",
		},
		{
			Name: "Gaby J Photography", OfficialURL: "https://www.gabyjphotography.com/",
			Handle: "hellogabyj", SourceClass: "engagement_photographer",
			CityID: "city_las_vegas_nv", State: "NV", City: "Las Vegas", Verified: "2026-08-03",
		},
		{
			Name: "Karissa Russ & Co", OfficialURL: "https://karissaruss.co/",
			Handle: "karissaruss", SourceClass: "engagement_photographer",
			CityID: "city_las_vegas_nv", State: "NV", City: "Las Vegas", Verified: "2026-08-03",
		},
		{
			Name: "Bowties Bridal", OfficialURL: "https://bowtiesbridal.com/",
			Handle: "bowtiesbridal", SourceClass: "bridal_shop",
			CityID: "city_las_vegas_nv", State: "NV", City: "Las Vegas", Verified: "2026-08-03",
		},
		{
			Name: "MaidenWhite", OfficialURL: "https://maidenwhite.com/",
			Handle: "maidenwhitebride", SourceClass: "bridal_shop",
			CityID: "city_las_vegas_nv", State: "NV", City: "Las Vegas", Verified: "2026-08-03",
		},
		{
			Name: "The Elm Estate", OfficialURL: "https://theelmestate.com/",
			Handle: "theelmestate", SourceClass: "wedding_venue",
			CityID: "city_reno_nv", State: "NV", City: "Reno", Verified: "2026-08-03",
		},
		{
			Name: "Lavender Ridge", OfficialURL: "https://www.lavenderridgereno.com/",
			Handle: "lavenderridgeweddings", SourceClass: "wedding_venue",
			CityID: "city_reno_nv", State: "NV", City: "Reno", Verified: "2026-08-03",
		},
	},
}
