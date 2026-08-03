package packs

// North Dakota source pack — verified 2026-08-01.
//
// Government: North Dakota marriage records are held by the county recorder
// (or treasurer/finance office in Cass County). Search URLs for the top 7
// counties were verified against each county's official .gov site.
//
// Church: Diocese of Fargo and Diocese of Bismarck verified via USCCB + each
// diocese's own website. Fargo-area parishes verified against the diocese
// parish directory + direct bulletin-archive URL discovery.
//
// Social: Instagram handles verified from each business's own public website
// social links. Verification date recorded per vendor.

var ndPack = StatePack{
	State: "ND",

	// --- Cities ----------------------------------------------------------
	Cities: []CityDef{
		{ID: "city_fargo_nd", State: "ND", County: "38017", Name: "Fargo",
			Lat: 46.8772, Lng: -96.7898, Markets: []string{"fargo", "nd", "northdakota", "cass"}},
		{ID: "city_bismarck_nd", State: "ND", County: "38015", Name: "Bismarck",
			Lat: 46.8083, Lng: -100.7837, Markets: []string{"bismarck", "nd", "burleigh"}},
	},

	// --- Government (county recorder marriage-record searches) -----------
	Government: []GovSource{
		{
			// Cass County (Fargo) — marriage licenses handled by the Finance
			// Office / County Treasurer since 2001.
			CountyFIPS: "38017",
			CourtName:  "Cass County Finance Office",
			CourtURL:   "https://www.casscountynd.gov",
			SearchURL:  "https://www.casscountynd.gov/our-county/finance-office/marriage-licenses-and-weddings/marriage-licenses",
			Note:       "Marriage license application page; records request-oriented, no online search portal.",
		},
		{
			// Burleigh County (Bismarck) — County Recorder marriage information.
			CountyFIPS: "38015",
			CourtName:  "Burleigh County Recorder",
			CourtURL:   "https://www.burleigh.gov/departments/recorder/",
			SearchURL:  "https://www.burleigh.gov/departments/recorder/marriage-information/",
			Note:       "Marriage information page; certified copies by request form, no online search portal.",
		},
		{
			// Grand Forks County — Tax Equalization Office handles marriage
			// licenses.
			CountyFIPS: "38035",
			CourtName:  "Grand Forks County Tax Equalization",
			CourtURL:   "https://www.gfcounty.nd.gov",
			SearchURL:  "https://www.gfcounty.nd.gov/government/tax-equalization-office/marriage-license-applications",
			Note:       "Marriage license application page; certified copies by request, no online search portal.",
		},
		{
			// Ward County (Minot) — County Recorder marriage license requirements.
			CountyFIPS: "38101",
			CourtName:  "Ward County Recorder",
			CourtURL:   "https://www.co.ward.nd.us/194/Recorder",
			SearchURL:  "https://www.co.ward.nd.us/199/1710/Marriage-License-Requirements",
			Note:       "Marriage license requirements page; recorded copy request form available, no online search portal.",
		},
		{
			// Morton County (Mandan) — County Recorder marriage licenses.
			CountyFIPS: "38059",
			CourtName:  "Morton County Recorder",
			CourtURL:   "https://www.mortonnd.gov",
			SearchURL:  "https://www.mortonnd.gov/marriage-passport",
			Note:       "Marriage license & passport services page; request-oriented, no online search portal.",
		},
		{
			// Stutsman County (Jamestown) — County Recorder / Treasurer.
			CountyFIPS: "38093",
			CourtName:  "Stutsman County Recorder",
			CourtURL:   "https://www.stutsmancounty.gov/departments/recorder/",
			SearchURL:  "https://www.stutsmancounty.gov/departments/recorder/marriage-license-information/",
			Note:       "Marriage license information page; statewide lookup available in-office, no public online search portal.",
		},
		{
			// Richland County (Wahpeton) — County Recorder.
			CountyFIPS: "38079",
			CourtName:  "Richland County Recorder",
			CourtURL:   "https://www.co.richland.nd.us/recorder/",
			SearchURL:  "https://www.co.richland.nd.us/recorder/",
			Note:       "Recorder page with marriage license info; request-oriented, no online search portal.",
		},
	},

	// --- Church (Catholic dioceses + parishes) ---------------------------
	Dioceses: []DioceseDef{
		{Slug: "fargo", Name: "Diocese of Fargo", Type: "diocese",
			Website: "https://fargodiocese.org", Directory: "https://fargodiocese.org/parishes", HubCityID: "city_fargo_nd"},
		{Slug: "bismarck", Name: "Diocese of Bismarck", Type: "diocese",
			Website: "https://www.bismarckdiocese.com", Directory: "https://www.bismarckdiocese.com/parishes", HubCityID: "city_bismarck_nd"},
	},

	// Fargo-area parishes in the Diocese of Fargo. Names and addresses
	// verified against the diocese parish directory (fargodiocese.net) +
	// each parish's own website. Bulletin URLs verified by direct search.
	Parishes: []ParishDef{
		{
			DioceseSlug: "fargo", Name: "Cathedral of St. Mary",
			Address:     "604 Broadway N, Fargo, ND 58102",
			BulletinURL: "https://cathedralofstmary.com/bulletins",
		},
		{
			DioceseSlug: "fargo", Name: "Holy Spirit Catholic Church",
			Address:     "1420 7th St N, Fargo, ND 58102",
			BulletinURL: "https://holyspiritfargo.com/bulletin",
		},
		{
			DioceseSlug: "fargo", Name: "Church of the Nativity",
			Address: "1825 11th St S, Fargo, ND 58103",
		},
		{
			DioceseSlug: "fargo", Name: "Sts. Anne & Joachim Catholic Church",
			Address:     "5202 25th St S, Fargo, ND 58104",
			BulletinURL: "https://stsaaj.org/bulletins/",
		},
		{
			DioceseSlug: "fargo", Name: "St. Anthony of Padua Catholic Church",
			Address: "710 10th St S, Fargo, ND 58103",
		},
		{
			DioceseSlug: "fargo", Name: "St. Paul's Catholic Newman Center",
			Address: "1141 University Dr N, Fargo, ND 58102",
		},
		{
			DioceseSlug: "fargo", Name: "Holy Cross Catholic Church",
			Address: "2711 7th St E, West Fargo, ND 58078",
		},
	},

	// --- Social (wedding-industry Instagram vendors) ---------------------
	Vendors: []VendorDef{
		// Fargo photographers
		{
			Name: "Abby Anderson Photography", OfficialURL: "https://abbyanderson.com/",
			Handle: "abbyanders", SourceClass: "engagement_photographer",
			CityID: "city_fargo_nd", State: "ND", City: "Fargo", Verified: "2026-08-01",
			TikTokHandle: "abbyanders",
		},
		{
			Name: "Kiella Lawrence Photography", OfficialURL: "https://kiellalawrence.com/",
			Handle: "kiellalawrencephoto", SourceClass: "engagement_photographer",
			CityID: "city_fargo_nd", State: "ND", City: "Fargo", Verified: "2026-08-01",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-photographers/kiella-lawrence-photography-2002015",
		},
		// Fargo venues
		{
			Name: "The Pines Weddings & Events", OfficialURL: "https://thepinesvenue.com/",
			Handle: "thepinesvenue", SourceClass: "wedding_venue",
			CityID: "city_fargo_nd", State: "ND", City: "Fargo", Verified: "2026-08-01",
			TikTokHandle: "thepinesvenue",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-venues/the-pines-weddings-events-1594917",
		},
		{
			Name: "The Yard Weddings & Events", OfficialURL: "https://www.theyardweddings.com/",
			Handle: "theyardweddings", SourceClass: "wedding_venue",
			CityID: "city_fargo_nd", State: "ND", City: "Fargo", Verified: "2026-08-01",
		},
		// Fargo jeweler
		{
			Name: "Schmidt's Gems & Fine Jewelry", OfficialURL: "https://schmidtsjewelry.us/",
			Handle: "schmidtsjewelry", SourceClass: "jeweler",
			CityID: "city_fargo_nd", State: "ND", City: "Fargo", Verified: "2026-08-01",
		},
		// Bismarck jeweler
		{
			Name: "Zorells Jewelry", OfficialURL: "https://zorells.com/",
			Handle: "zorellsjewelry", SourceClass: "jeweler",
			CityID: "city_bismarck_nd", State: "ND", City: "Bismarck", Verified: "2026-08-01",
		},
		{
			Name: "Abby Anderson", OfficialURL: "https://abbyanderson.com/",
			Handle: "abbyandersonphoto", SourceClass: "engagement_photographer",
			CityID: "city_fargo_nd", State: "ND", City: "Fargo", Verified: "2026-08-03",
		},
		{
			Name: "Officiant Amber", OfficialURL: "https://www.officiantamber.com/",
			Handle: "officiantamber", SourceClass: "officiant",
			CityID: "city_fargo_nd", State: "ND", City: "Fargo", Verified: "2026-08-03",
		},
		{
			Name: "Your Day by Nicole", OfficialURL: "https://shopydbn.com/",
			Handle: "yourdaybynicole", SourceClass: "bridal_shop",
			CityID: "city_fargo_nd", State: "ND", City: "Fargo", Verified: "2026-08-03",
		},
		{
			Name: "Klaudia & Co.", OfficialURL: "https://klaudiaandco.com/",
			Handle: "klaudiaandcobridal", SourceClass: "bridal_shop",
			CityID: "city_fargo_nd", State: "ND", City: "Fargo", Verified: "2026-08-03",
		},
	},
}
