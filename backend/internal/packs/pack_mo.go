package packs

// Missouri source pack — verified 2026-08-01.
//
// Government: Missouri marriage records are held by each county's Recorder of
// Deeds. Search URLs for the top 7 counties by population were verified against
// each county's official .gov site or its fidlar/iCounty search portal.
//
// Church: all 4 Missouri Catholic dioceses/archdioceses verified via USCCB +
// each diocese's own website. St. Louis-area parishes (Archdiocese of St. Louis)
// were verified against the archdiocese's own parish directory
// (archstl.org/parish-directory) + direct bulletin-archive URL discovery.
//
// Social: Instagram handles verified from each business's own public website
// social links (or from IG search results where the site is JS-rendered and
// the handle was visible in the search snippet). Verification date recorded
// per vendor.

var moPack = StatePack{
	State: "MO",

	// --- Cities ----------------------------------------------------------
	Cities: []CityDef{
		{
			ID: "city_st_louis_mo", State: "MO", County: "29510", Name: "St. Louis",
			Lat: 38.6270, Lng: -90.1994,
			Markets: []string{"stlouis", "stl", "missouri"},
		},
		{
			ID: "city_kansas_city_mo", State: "MO", County: "29095", Name: "Kansas City",
			Lat: 39.0997, Lng: -94.5786,
			Markets: []string{"kansascity", "kc", "jackson", "missouri"},
		},
	},

	// --- Government (county recorder marriage-record searches) ------------
	Government: []GovSource{
		{
			// St. Louis City (independent city, FIPS 29510) — Recorder of
			// Deeds Marriage Department; free fidlar search portal covering
			// 1932–present.
			CountyFIPS: "29510",
			CourtName:  "St. Louis City Recorder of Deeds",
			CourtURL:   "https://www.stlouis-mo.gov/government/departments/recorder/marriage/index.cfm",
			SearchURL:  "https://mostlouiscity.fidlar.com/MOStLouisVS/Apex.WebPortal/search",
			Note:       "Free fidlar marriage search portal (1932–present); enumeration candidate.",
		},
		{
			// St. Louis County (FIPS 29189) — Recorder of Deeds; marriage
			// license index searchable via web-based deed search app.
			CountyFIPS: "29189",
			CourtName:  "St. Louis County Recorder of Deeds",
			CourtURL:   "https://stlouiscountymo.gov/st-louis-county-departments/revenue/recorder-of-deeds/",
			SearchURL:  "https://stlouiscountymo.gov/st-louis-county-departments/revenue/recorder-of-deeds/marriage-copy-order-form/",
			Note:       "Web-based app with marriage license index (1877–present); marriage filtering needs testing.",
		},
		{
			// Jackson County (Kansas City, FIPS 29095) — Recorder of Deeds;
			// Aumentum public access search with marriage index back to 1826.
			CountyFIPS: "29095",
			CourtName:  "Jackson County Recorder of Deeds",
			CourtURL:   "https://www.jacksongov.org/Government/Departments/Recorder-of-Deeds",
			SearchURL:  "https://aumentumweb.jacksongov.org/users/basket.aspx",
			Note:       "Aumentum public access portal with Search Marriage Index (1826–present); enumeration candidate.",
		},
		{
			// Clay County (FIPS 29047) — Recorder of Deeds; iRecord search
			// via iCounty platform.
			CountyFIPS: "29047",
			CourtName:  "Clay County Recorder of Deeds",
			CourtURL:   "https://www.claycountymo.gov/253/Recorder-of-Deeds",
			SearchURL:  "https://claymo.icounty.com/",
			Note:       "iRecord search via iCounty; marriage records searchable; enumeration capability needs testing.",
		},
		{
			// Greene County (Springfield, FIPS 29077) — Recorder of Deeds;
			// marriage license search accessible from recorder navigation.
			CountyFIPS: "29077",
			CourtName:  "Greene County Recorder of Deeds",
			CourtURL:   "https://www.greenecountymo.gov/recorder/",
			SearchURL:  "https://www.greenecountymo.gov/recorder/marriage_licenses.php",
			Note:       "Marriage License Search in recorder nav menu; dedicated search tool; enumeration capability needs testing.",
		},
		{
			// Boone County (Columbia, FIPS 29019) — Recorder of Deeds;
			// iRecord search via iCounty with marriage records from 1865.
			CountyFIPS: "29019",
			CourtName:  "Boone County Recorder of Deeds",
			CourtURL:   "https://www.boonemo.gov/recorder/",
			SearchURL:  "https://boonemo.icounty.com/login/login",
			Note:       "iRecord search via iCounty; marriage licenses from 1865 included; enumeration capability needs testing.",
		},
		{
			// Jefferson County (FIPS 29099) — Recorder of Deeds; certified
			// marriage copy orders via OfficialRecordsOnline.
			CountyFIPS: "29099",
			CourtName:  "Jefferson County Recorder of Deeds",
			CourtURL:   "https://www.jeffcomo.gov/315/Recorder-of-Deeds",
			SearchURL:  "https://www.officialrecordsonline.com/Select/Index.html?state=MO&county=JEFFERSON",
			Note:       "OfficialRecordsOnline portal for marriage license copies; land records via Tapestry; enumeration capability needs testing.",
		},
	},

	// --- Church (Catholic dioceses + parishes) ---------------------------
	Dioceses: []DioceseDef{
		{
			Slug: "st_louis", Name: "Archdiocese of St. Louis", Type: "archdiocese",
			Website: "https://archstl.org", Directory: "https://archstl.org/parish-directory",
			HubCityID: "city_st_louis_mo",
		},
		{
			Slug: "kc_st_joseph", Name: "Diocese of Kansas City-St. Joseph", Type: "diocese",
			Website: "https://www.diocese-kcsj.org", Directory: "https://www.diocese-kcsj.org/parishes",
			HubCityID: "city_kansas_city_mo",
		},
		{
			Slug: "springfield_cg", Name: "Diocese of Springfield-Cape Girardeau", Type: "diocese",
			Website: "https://www.dioscg.org", Directory: "https://www.dioscg.org/parishes",
		},
		{
			Slug: "jefferson_city", Name: "Diocese of Jefferson City", Type: "diocese",
			Website: "https://www.diojeffcity.org", Directory: "https://www.diojeffcity.org/parishes",
		},
	},

	// St. Louis-area parishes in the Archdiocese of St. Louis.
	// Names and addresses verified from the archdiocese's own parish directory
	// (archstl.org/parish-directory). Bulletin URLs verified by direct search
	// for each parish's bulletin archive.
	Parishes: []ParishDef{
		{
			DioceseSlug: "st_louis", Name: "Cathedral Basilica of Saint Louis",
			Address:     "4431 Lindell Blvd, St. Louis, MO 63108",
			BulletinURL: "https://cathedralstl.org/bulletins",
		},
		{
			DioceseSlug: "st_louis", Name: "Basilica of St. Louis, King of France",
			Address:     "209 Walnut St, St. Louis, MO 63102",
			BulletinURL: "https://www.stlouiskingoffrance.org/bulletin-archive/",
		},
		{
			DioceseSlug: "st_louis", Name: "Immacolata Catholic Church",
			Address:     "8900 Clayton Rd, St. Louis, MO 63117",
			BulletinURL: "https://immacolata.org/bulletins",
		},
		{
			DioceseSlug: "st_louis", Name: "Christ the King Catholic Church",
			Address:     "7316 Balson Ave, St. Louis, MO 63130",
			BulletinURL: "https://www.ctkstl.com/bulletins",
		},
		{
			DioceseSlug: "st_louis", Name: "Holy Infant Catholic Church",
			Address: "627 Dennison Dr, Ballwin, MO 63021",
		},
		{
			DioceseSlug: "st_louis", Name: "St. Peter Catholic Church",
			Address:     "243 W Argonne Dr, Kirkwood, MO 63122",
			BulletinURL: "https://stpeterkirkwood.org/bulletin",
		},
		{
			DioceseSlug: "st_louis", Name: "St. Gerard Majella Catholic Church",
			Address: "1969 Dougherty Ferry Rd, St. Louis, MO 63122",
		},
	},

	// --- Social (wedding-industry Instagram vendors) ---------------------
	Vendors: []VendorDef{
		// St. Louis photographers
		{
			Name: "Kara Hoganson Photo", OfficialURL: "https://karahogansonphoto.com/",
			Handle: "karahogansonphoto", SourceClass: "engagement_photographer",
			CityID: "city_st_louis_mo", State: "MO", City: "St. Louis", Verified: "2026-08-01",
			TikTokHandle: "karahogansonphoto",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-photographers/kara-hoganson-photo-6446870",
		},
		{
			Name: "Josh and Lindsey Photo", OfficialURL: "https://www.joshandlindseyphoto.com/",
			Handle: "joshandlindseyphoto", SourceClass: "engagement_photographer",
			CityID: "city_st_louis_mo", State: "MO", City: "St. Louis", Verified: "2026-08-01",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-photographers/josh-and-lindsey-photo-9975680",
		},
		{
			Name: "Abigail Munshaw Photography", OfficialURL: "https://www.abigailmunshawphotography.com/",
			Handle: "abigailmunshawphoto", SourceClass: "engagement_photographer",
			CityID: "city_st_louis_mo", State: "MO", City: "St. Louis", Verified: "2026-08-01",
			TikTokHandle: "abailmunshawphoto",
		},
		{
			Name: "Lindsey Tyler Weddings", OfficialURL: "https://lindseytylerweddings.com/",
			Handle: "lindseytylerphoto_film", SourceClass: "engagement_photographer",
			CityID: "city_st_louis_mo", State: "MO", City: "St. Louis", Verified: "2026-08-01",
		},
		// St. Louis venues
		{
			Name: "Silver Oaks Chateau", OfficialURL: "https://www.silveroakschateau.com/",
			Handle: "silveroakschateau", SourceClass: "wedding_venue",
			CityID: "city_st_louis_mo", State: "MO", City: "St. Louis", Verified: "2026-08-01",
			TikTokHandle: "silveroakschateau",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-venues/silver-oaks-chateau-8505520",
		},
		{
			Name: "The Venue at Wildflower Ridge", OfficialURL: "https://www.thevenueatwildflowerridge.com/",
			Handle: "wildflowerridgeweddings", SourceClass: "wedding_venue",
			CityID: "city_st_louis_mo", State: "MO", City: "St. Louis", Verified: "2026-08-01",
		},
		{
			Name: "Knotting Hills Wedding Venue & Resort", OfficialURL: "https://www.knottinghills.com/",
			Handle: "knotting.hills", SourceClass: "wedding_venue",
			CityID: "city_st_louis_mo", State: "MO", City: "St. Louis", Verified: "2026-08-01",
			TikTokHandle: "knotting.hills",
		},
		// St. Louis jewelers
		{
			Name: "Simons Jewelers", OfficialURL: "https://www.simonsjewelers.com/",
			Handle: "simonsjewelers", SourceClass: "jeweler",
			CityID: "city_st_louis_mo", State: "MO", City: "St. Louis", Verified: "2026-08-01",
		},
		{
			Name: "Clarkson Jewelers", OfficialURL: "https://clarksonjewelers.com/",
			Handle: "clarksonjewelers", SourceClass: "jeweler",
			CityID: "city_st_louis_mo", State: "MO", City: "St. Louis", Verified: "2026-08-01",
		},
		{
			Name: "White Klump Photography", OfficialURL: "https://whiteklumpphotography.com/home",
			Handle: "whiteklump_photog", SourceClass: "engagement_photographer",
			CityID: "city_st_louis_mo", State: "MO", City: "St. Louis", Verified: "2026-08-03",
		},
		{
			Name: "The Coronado", OfficialURL: "https://thecoronado.com/",
			Handle: "thecoronadostl", SourceClass: "wedding_venue",
			CityID: "city_st_louis_mo", State: "MO", City: "St. Louis", Verified: "2026-08-03",
		},
		{
			Name: "Haue Valley", OfficialURL: "https://hauevalleyweddings.com/",
			Handle: "hauevalleyweddings", SourceClass: "wedding_venue",
			CityID: "city_st_louis_mo", State: "MO", City: "St. Louis", Verified: "2026-08-03",
		},
		{
			Name: "Four Seasons Hotel St. Louis", OfficialURL: "https://www.fourseasons.com/stlouis/weddings/venues/",
			Handle: "fourseasonsstlouis", SourceClass: "wedding_venue",
			CityID: "city_st_louis_mo", State: "MO", City: "St. Louis", Verified: "2026-08-03",
		},
		{
			Name: "The Christy", OfficialURL: "https://www.thechristy.com/weddings/weddings-st-louis/",
			Handle: "thechristystl", SourceClass: "wedding_venue",
			CityID: "city_st_louis_mo", State: "MO", City: "St. Louis", Verified: "2026-08-03",
		},
	},
}
