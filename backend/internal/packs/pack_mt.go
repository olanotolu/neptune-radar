package packs

// Montana source pack — verified 2026-08-02.
//
// Government: Montana marriage records are held by the Clerk of District Court
// in each county (the state DPHHS maintains only a statewide index, not
// certificates). Search URLs for the top 7 counties by population were
// verified against each county's official .gov site.
//
// Church: both Montana dioceses verified via USCCB + each diocese's own
// website. Billings- and Great Falls-area parishes (Diocese of Great
// Falls-Billings) verified against the diocese parish directory + each
// parish's own website. Bulletin URLs verified by direct search for each
// parish's bulletin archive.
//
// Social: Instagram handles verified from each business's own public website
// social links (or from IG search results where the site is JS-rendered and
// the handle was visible in the search snippet). Verification date recorded
// per vendor.

var mtPack = StatePack{
	State: "MT",

	// --- Cities ----------------------------------------------------------
	Cities: []CityDef{
		{ID: "city_billings_mt", State: "MT", County: "30111", Name: "Billings",
			Lat: 45.7833, Lng: -108.5007, Markets: []string{"billings", "mt", "yellowstone", "montana"}},
		{ID: "city_missoula_mt", State: "MT", County: "30063", Name: "Missoula",
			Lat: 46.8787, Lng: -113.9966, Markets: []string{"missoula", "mt", "montana"}},
	},

	// --- Government (clerk of district court marriage-record searches) ---
	Government: []GovSource{
		{
			// Yellowstone County (Billings) — Clerk of District Court marriage
			// license search page.
			CountyFIPS: "30111",
			CourtName:  "Yellowstone County Clerk of District Court",
			CourtURL:   "https://www.yellowstonecountymt.gov/clerk_court/",
			SearchURL:  "https://www.yellowstonecountymt.gov/clerk_court/ML_csearch.asp",
			Note:       "Marriage license search page; records sealed 30 years; request-oriented for copies.",
		},
		{
			// Missoula County — District Court records / marriage licenses page.
			CountyFIPS: "30063",
			CourtName:  "Missoula County Clerk of District Court",
			CourtURL:   "https://www.missoulacounty.gov/departments/district-court/court-records/",
			SearchURL:  "https://www.missoulacounty.gov/departments/district-court/marriage-licenses/",
			Note:       "Marriage licenses page with copy/search fee schedule; request-oriented, no online search portal.",
		},
		{
			// Cascade County (Great Falls) — Clerk of Court copies & search
			// request documents page.
			CountyFIPS: "30013",
			CourtName:  "Cascade County Clerk of District Court",
			CourtURL:   "https://www.cascadecountymt.gov/158/Clerk-of-Courts-Office",
			SearchURL:  "https://cascadecountymt.gov/459/Copies-and-Search-Request-Documents",
			Note:       "Copies and search request documents page; marriage license copy forms downloadable; no online search portal.",
		},
		{
			// Gallatin County (Bozeman) — Clerk of District Court records
			// requests page; also offers public records portal.
			CountyFIPS: "30031",
			CourtName:  "Gallatin County Clerk of District Court",
			CourtURL:   "https://www.gallatinmt.gov/272/Clerk-of-District-Court",
			SearchURL:  "https://www.gallatinmt.gov/293/Records-Requests",
			Note:       "Records requests page with fee schedule; public records portal at dcportal.pubcourts.mt.gov for case index; marriage copies request-oriented.",
		},
		{
			// Flathead County (Kalispell) — Clerk of Court department page;
			// marriage application start page for online applications.
			CountyFIPS: "30029",
			CourtName:  "Flathead County Clerk of District Court",
			CourtURL:   "https://flatheadcounty.gov/department-directory/clerk_of_court",
			SearchURL:  "https://flatheadcounty.gov/department-directory/clerk_of_court/marriage_application_start",
			Note:       "Marriage application start page; copy requests via mail/email with PDF form; no online search portal.",
		},
		{
			// Lewis and Clark County (Helena) — Clerk of District Court copies
			// and document search requests page.
			CountyFIPS: "30049",
			CourtName:  "Lewis and Clark County Clerk of District Court",
			CourtURL:   "https://www.lccountymt.gov/Government/Clerk-of-District-Court",
			SearchURL:  "https://www.lccountymt.gov/Government/Clerk-of-District-Court/Copies-and-Document-Search-Requests",
			Note:       "Copies and document search requests page with downloadable forms; search via email/fax/mail; no online search portal.",
		},
		{
			// Silver Bow County (Butte) — Clerk of District Court staff
			// directory page; marriage licenses issued in person.
			CountyFIPS: "30093",
			CourtName:  "Silver Bow County Clerk of District Court",
			CourtURL:   "https://www.co.silverbow.mt.us/Directory.aspx?DID=19",
			SearchURL:  "https://www.co.silverbow.mt.us/Directory.aspx?DID=19",
			Note:       "Clerk of District Court directory page; marriage licenses issued in person; copies and searches request-oriented via phone/mail.",
		},
	},

	// --- Church (Catholic dioceses + parishes) ---------------------------
	Dioceses: []DioceseDef{
		{Slug: "helena", Name: "Diocese of Helena", Type: "diocese",
			Website: "https://www.diocesehelena.org", Directory: "https://www.diocesehelena.org/parishes"},
		{Slug: "great_falls_billings", Name: "Diocese of Great Falls-Billings", Type: "diocese",
			Website: "https://www.gfbdiocese.org", Directory: "https://www.gfbdiocese.org/parishes", HubCityID: "city_billings_mt"},
	},

	// Billings- and Great Falls-area parishes in the Diocese of Great
	// Falls-Billings. Names verified from the diocese parish directory
	// (diocesegfb.org) + gcatholic.org listing. Bulletin URLs verified by
	// direct search for each parish's bulletin archive.
	Parishes: []ParishDef{
		{
			DioceseSlug: "great_falls_billings", Name: "St. Patrick Co-Cathedral",
			Address: "215 N 31st St, Billings, MT 59101",
		},
		{
			DioceseSlug: "great_falls_billings", Name: "Mary Queen of Peace Parish",
			Address:     "3411 3rd Ave S, Billings, MT 59101",
			BulletinURL: "https://www.mqpbillings.org/bulletins",
		},
		{
			DioceseSlug: "great_falls_billings", Name: "St. Bernard Catholic Church",
			Address:     "226 Wicks Ln, Billings, MT 59105",
			BulletinURL: "https://stbernardblgs.org/Resources/Sunday-Bulletin",
		},
		{
			DioceseSlug: "great_falls_billings", Name: "St. Pius X Parish",
			Address:     "717 18th St W, Billings, MT 59102",
			BulletinURL: "https://www.stpiusxblgs.org/weekly-bulletins/",
		},
		{
			DioceseSlug: "great_falls_billings", Name: "St. Thomas the Apostle Catholic Community",
			Address:     "2055 Woody Dr, Billings, MT 59102",
			BulletinURL: "https://www.stthomasbillings.org/bulletins",
		},
		{
			DioceseSlug: "great_falls_billings", Name: "Holy Spirit Catholic Parish",
			Address: "201 44th St S, Great Falls, MT 59405",
		},
		{
			DioceseSlug: "great_falls_billings", Name: "Corpus Christi Catholic Parish",
			Address:     "410 22nd Ave NE, Great Falls, MT 59404",
			BulletinURL: "https://corpuschristigreatfalls.blogspot.com/p/latest-bulletin.html",
		},
	},

	// --- Social (wedding-industry Instagram vendors) ---------------------
	Vendors: []VendorDef{
		// Billings photographers
		{
			Name: "Alaina Faith Photography", OfficialURL: "https://alainafaithphotography.com/",
			Handle: "alaina.faithphotography", SourceClass: "engagement_photographer",
			CityID: "city_billings_mt", State: "MT", City: "Billings", Verified: "2026-08-02",
			TikTokHandle: "alaina.faithphotography",
		},
		{
			Name: "Sara Nagel Photography", OfficialURL: "https://saranagelphotography.com/",
			Handle: "saranagelphotography", SourceClass: "engagement_photographer",
			CityID: "city_billings_mt", State: "MT", City: "Billings", Verified: "2026-08-02",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-photographers/sara-nagel-photography-6901252",
		},
		// Billings venues
		{
			Name: "Billings Depot", OfficialURL: "https://www.billingsdepot.org/",
			Handle: "historicbillingsdepot", SourceClass: "wedding_venue",
			CityID: "city_billings_mt", State: "MT", City: "Billings", Verified: "2026-08-02",
			TikTokHandle: "historicbillingsdepot",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-venues/billings-depot-2080410",
		},
		// Billings jewelers
		{
			Name: "Montague's Jewelers", OfficialURL: "https://www.montaguesjewelers.com/",
			Handle: "montaguesjewelers", SourceClass: "jeweler",
			CityID: "city_billings_mt", State: "MT", City: "Billings", Verified: "2026-08-02",
		},
		{
			Name: "Goldsmith Gallery Jewelers", OfficialURL: "https://www.goldsmithgalleryjewelers.com/",
			Handle: "goldsmithgalleryjewelers", SourceClass: "jeweler",
			CityID: "city_billings_mt", State: "MT", City: "Billings", Verified: "2026-08-02",
		},
		// Missoula photographers
		{
			Name: "Jazzer Rae Photos", OfficialURL: "https://www.jazzerraephotos.com/",
			Handle: "jazzerraephotos", SourceClass: "engagement_photographer",
			CityID: "city_missoula_mt", State: "MT", City: "Missoula", Verified: "2026-08-02",
		},
		{
			Name: "Esther Grace Photo", OfficialURL: "https://esthergracephoto.com/",
			Handle: "esthergracephoto", SourceClass: "engagement_photographer",
			CityID: "city_missoula_mt", State: "MT", City: "Missoula", Verified: "2026-08-02",
			TikTokHandle: "esthergracephoto",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-photographers/esther-grace-photo-8152924",
		},
		// Missoula venues
		{
			Name: "White Raven Venue & Retreat", OfficialURL: "https://www.whiteravenmontana.com/",
			Handle: "whiteravenvenue", SourceClass: "wedding_venue",
			CityID: "city_missoula_mt", State: "MT", City: "Missoula", Verified: "2026-08-02",
		},
		{
			Name: "Kristin Jean Photography", OfficialURL: "https://kristinjeanphotographer.com/",
			Handle: "kristinjeanphoto", SourceClass: "engagement_photographer",
			CityID: "city_billings_mt", State: "MT", City: "Billings", Verified: "2026-08-03",
		},
		{
			Name: "Samuel Roland Films", OfficialURL: "https://samuelrolandfilms.com/",
			Handle: "samuelrolandfilms", SourceClass: "videographer",
			CityID: "city_billings_mt", State: "MT", City: "Billings", Verified: "2026-08-03",
		},
		{
			Name: "Rogers & Co. Fine Jewelry", OfficialURL: "https://www.rogerscojewelry.com/",
			Handle: "rogersandco.finejewelry", SourceClass: "jeweler",
			CityID: "city_missoula_mt", State: "MT", City: "Missoula", Verified: "2026-08-03",
		},
		{
			Name: "The Barn on Mullan", OfficialURL: "https://www.barnonmullanmt.com/",
			Handle: "thebarnonmullan", SourceClass: "wedding_venue",
			CityID: "city_missoula_mt", State: "MT", City: "Missoula", Verified: "2026-08-03",
		},
		{
			Name: "A & A Weddings", OfficialURL: "https://aandaweddings.com/",
			Handle: "aandaweddings", SourceClass: "videographer",
			CityID: "city_billings_mt", State: "MT", City: "Billings", Verified: "2026-08-03",
		},
	},
}
