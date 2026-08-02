package packs

// Virginia source pack — verified 2026-08-01.
//
// Government: Virginia marriage records are held by the Circuit Court Clerk
// in each county/city. Search URLs for the top 7 jurisdictions by population
// were verified against each locality's official .gov site or its online
// records portal.
//
// Church: both Virginia Catholic dioceses verified via USCCB + each diocese's
// own website. Northern Virginia parishes (Diocese of Arlington) were verified
// against the diocese's parish directory + Wikipedia's list of churches in the
// diocese (which cites the diocese's own records). Bulletin URLs verified by
// direct search for each parish's bulletin archive.
//
// Social: Instagram handles verified from each business's own public website
// social links (or from IG search results where the site is JS-rendered and
// the handle was visible in the search snippet). Verification date recorded
// per vendor.

var vaPack = StatePack{
	State: "VA",

	// --- Cities ----------------------------------------------------------
	Cities: []CityDef{
		{
			ID: "city_arlington_va", State: "VA", County: "51013", Name: "Arlington",
			Lat: 38.8816, Lng: -77.0927,
			Markets: []string{"arlington", "nova", "northernvirginia", "dc"},
		},
		{
			ID: "city_richmond_va", State: "VA", County: "51760", Name: "Richmond",
			Lat: 37.5407, Lng: -77.4360,
			Markets: []string{"richmond", "rva", "henrico"},
		},
		{
			ID: "city_virginia_beach_va", State: "VA", County: "51810", Name: "Virginia Beach",
			Lat: 36.8529, Lng: -75.9780,
			Markets: []string{"virginiabeach", "vb", "hamptonroads"},
		},
	},

	// --- Government (circuit court clerk marriage-record searches) -------
	Government: []GovSource{
		{
			// Fairfax County — eCaseSearch online case/records portal;
			// marriage license copy requests via the circuit court site.
			CountyFIPS: "51059",
			CourtName:  "Fairfax County Circuit Court Clerk",
			CourtURL:   "https://www.fairfaxcounty.gov/circuit/marriage/marriage-license-copy",
			SearchURL:  "https://www.fairfaxcounty.gov/apps/ECS_Public/",
			Note:       "eCaseSearch online portal; marriage license copy requests via court site; enumeration capability needs testing.",
		},
		{
			// Virginia Beach — online marriage license application portal.
			CountyFIPS: "51810",
			CourtName:  "Virginia Beach Circuit Court Clerk",
			CourtURL:   "https://courts.virginiabeach.gov/circuit-court-clerks-office/marriage",
			SearchURL:  "https://vbmarriage.org/",
			Note:       "Online marriage license application portal; records search is request-oriented.",
		},
		{
			// Prince William County — document books with marriage index
			// images via Courthouse Computer Systems; copy requests online.
			CountyFIPS: "51153",
			CourtName:  "Prince William County Circuit Court Clerk",
			CourtURL:   "https://www.pwcva.gov/department/circuit-court/record-copy-requests/",
			SearchURL:  "https://us6.courthousecomputersystems.com/PrinceWilliamVA/",
			Note:       "Document books with marriage index images; Historical Online Portal (HOP) for older records; enumeration candidate.",
		},
		{
			// Loudoun County — online certified copy request system;
			// marriage license info on the county site.
			CountyFIPS: "51107",
			CourtName:  "Loudoun County Circuit Court Clerk",
			CourtURL:   "https://www.loudoun.gov/1162/Marriage-Licenses",
			SearchURL:  "https://lisweb.loudoun.gov/Kiosk/51107/CertifiedCopy/Instructions",
			Note:       "Online certified copy request kiosk; historic marriage index PDFs available; enumeration needs testing.",
		},
		{
			// Chesterfield County — public records database (CCC Land
			// Records); online marriage license application.
			CountyFIPS: "51041",
			CourtName:  "Chesterfield County Circuit Court Clerk",
			CourtURL:   "https://www.chesterfield.gov/1127/Circuit-Court",
			SearchURL:  "https://www.ccclandrecords.org/Opening.asp",
			Note:       "CCC Land Records public records database; marriage license application online; enumeration capability needs testing.",
		},
		{
			// Henrico County — marriage license info and copy requests
			// via the clerk's office.
			CountyFIPS: "51087",
			CourtName:  "Henrico County Circuit Court Clerk",
			CourtURL:   "https://henrico.gov/services/marriage-license/",
			SearchURL:  "https://henrico.gov/clerk/",
			Note:       "Marriage license info + copy requests via clerk's office; online case info via Virginia Judiciary; enumeration needs testing.",
		},
		{
			// Arlington County — marriage license info and certified
			// copy requests via the clerk's office.
			CountyFIPS: "51013",
			CourtName:  "Arlington County Circuit Court Clerk",
			CourtURL:   "https://www.arlingtonva.us/Government/Departments/Courts/Circuit-Court/Marriage",
			SearchURL:  "https://www.arlingtonva.us/Government/Departments/Courts/Circuit-Court",
			Note:       "Marriage license by appointment; certified copies via clerk or Vital Records; request-oriented.",
		},
	},

	// --- Church (Catholic dioceses + parishes) ---------------------------
	Dioceses: []DioceseDef{
		{
			Slug: "arlington", Name: "Diocese of Arlington", Type: "diocese",
			Website: "https://www.arlingtondiocese.org", Directory: "https://www.arlingtondiocese.org/parish-finder/",
			HubCityID: "city_arlington_va",
		},
		{
			Slug: "richmond", Name: "Diocese of Richmond", Type: "diocese",
			Website: "https://www.richmonddiocese.org", Directory: "https://www.richmonddiocese.org/parishes",
			HubCityID: "city_richmond_va",
		},
	},

	// Northern Virginia parishes in the Diocese of Arlington. Names and
	// addresses verified from the diocese's parish directory + Wikipedia's
	// list of churches in the diocese (which cites the diocese's own
	// records). Bulletin URLs verified by direct search for each parish's
	// bulletin archive.
	Parishes: []ParishDef{
		{
			DioceseSlug: "arlington", Name: "Cathedral of Saint Thomas More",
			Address:     "3901 Cathedral Ln, Arlington, VA 22203",
			BulletinURL: "https://www.cathedralstm.org/bulletin/",
		},
		{
			DioceseSlug: "arlington", Name: "Saint Agnes Catholic Church",
			Address:     "2002 N. Randolph St, Arlington, VA 22207",
			BulletinURL: "https://saintagnes.org/bulletins/",
		},
		{
			DioceseSlug: "arlington", Name: "Saint James Catholic Church",
			Address:     "905 Park Ave, Falls Church, VA 22046",
			BulletinURL: "https://stjamescatholic.org/other-parish-diocesan-news/bulletins/",
		},
		{
			DioceseSlug: "arlington", Name: "Saint John the Beloved Catholic Church",
			Address:     "6420 Linway Ter, McLean, VA 22101",
			BulletinURL: "https://www.stjohncatholicmclean.org/bulletins1/",
		},
		{
			DioceseSlug: "arlington", Name: "Saint Rita Catholic Church",
			Address:     "3815 Russell Rd, Alexandria, VA 22305",
			BulletinURL: "https://stritaalexandria.com/weekly-bulletin/",
		},
		{
			DioceseSlug: "arlington", Name: "Saint Luke Catholic Church",
			Address: "7001 Georgetown Pike, McLean, VA 22101",
		},
		{
			DioceseSlug: "arlington", Name: "Saint Philip Catholic Church",
			Address:     "7500 Saint Philip Ct, Falls Church, VA 22042",
			BulletinURL: "https://church-bulletin.org/?id=714",
			Aggregator:  true,
		},
		{
			DioceseSlug: "arlington", Name: "Saint Charles Borromeo Catholic Church",
			Address: "3304 N. Washington Blvd, Arlington, VA 22201",
		},
	},

	// --- Social (wedding-industry Instagram vendors) ---------------------
	Vendors: []VendorDef{
		// Northern Virginia photographers
		{
			Name: "Danielle Towle Photography", OfficialURL: "https://danielletowlephotography.com/",
			Handle: "danielletowlephotography", SourceClass: "engagement_photographer",
			CityID: "city_arlington_va", State: "VA", City: "Arlington", Verified: "2026-08-01",
			TikTokHandle: "danielletowlephotography",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-photographers/danielle-towle-photography-4115793",
		},
		{
			Name: "Anna Wright Photography", OfficialURL: "https://annakayphotography.net/",
			Handle: "annawrightphoto", SourceClass: "engagement_photographer",
			CityID: "city_arlington_va", State: "VA", City: "Arlington", Verified: "2026-08-01",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-photographers/anna-wright-photography-7801178",
		},
		{
			Name: "Monica Roberts", OfficialURL: "https://monicaroberts.com/",
			Handle: "monicaroberts_", SourceClass: "engagement_photographer",
			CityID: "city_arlington_va", State: "VA", City: "Arlington", Verified: "2026-08-01",
			TikTokHandle: "monicaroberts",
		},
		{
			Name: "Kelly Loss Photography", OfficialURL: "https://kellylossphoto.com/",
			Handle: "kellylossphoto", SourceClass: "engagement_photographer",
			CityID: "city_arlington_va", State: "VA", City: "Arlington", Verified: "2026-08-01",
		},
		// Richmond venues
		{
			Name: "The Mill at Fine Creek", OfficialURL: "https://themillatfinecreek.com/",
			Handle: "millatfinecreek", SourceClass: "wedding_venue",
			CityID: "city_richmond_va", State: "VA", City: "Richmond", Verified: "2026-08-01",
			TikTokHandle: "millatfinecreek",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-venues/the-mill-at-fine-creek-3315138",
		},
		{
			Name: "Mankin Mansion", OfficialURL: "https://www.mankinmansion.com/",
			Handle: "mankinmansion", SourceClass: "wedding_venue",
			CityID: "city_richmond_va", State: "VA", City: "Richmond", Verified: "2026-08-01",
		},
		{
			Name: "The Jefferson Hotel", OfficialURL: "https://www.jeffersonhotel.com/",
			Handle: "thejeffersonhotel", SourceClass: "wedding_venue",
			CityID: "city_richmond_va", State: "VA", City: "Richmond", Verified: "2026-08-01",
			TikTokHandle: "thejeffersonhotel",
		},
		// Northern Virginia jewelers
		{
			Name: "Washington Diamond", OfficialURL: "https://washingtondiamond.com/",
			Handle: "mywashingtondiamond", SourceClass: "jeweler",
			CityID: "city_arlington_va", State: "VA", City: "Arlington", Verified: "2026-08-01",
		},
		{
			Name: "Midtown Jewelers", OfficialURL: "https://midtownjewelersinc.com/",
			Handle: "midtownjewelers", SourceClass: "jeweler",
			CityID: "city_arlington_va", State: "VA", City: "Arlington", Verified: "2026-08-01",
		},
	},
}
