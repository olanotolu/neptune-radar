package packs

// District of Columbia source pack — verified 2026-08-01.
//
// Government: DC marriage records are held by the Superior Court Marriage
// Bureau. The marriage-records landing page and certified-copy order page
// were verified against dccourts.gov.
//
// Church: the Archdiocese of Washington was verified via USCCB + adw.org.
// DC-area parishes were verified against the archdiocese's own parish
// directory PDF and each parish's own website. Bulletin URLs verified by
// direct search for each parish's bulletin archive.
//
// Social: Instagram handles verified from each business's own public website
// social links (or from IG search results where the site is JS-rendered and
// the handle was visible in the search snippet). Verification date recorded
// per vendor.

var dcPack = StatePack{
	State: "DC",

	// --- Cities ----------------------------------------------------------
	Cities: []CityDef{
		{ID: "city_washington_dc", State: "DC", County: "11001", Name: "Washington",
			Lat: 38.9072, Lng: -77.0369, Markets: []string{"washington", "dc", "district", "capitol"}},
	},

	// --- Government (Superior Court marriage-record search) --------------
	Government: []GovSource{
		{
			// DC Superior Court Marriage Bureau — marriage records from
			// 1811 to present. The certified-copy order page is the
			// primary public-facing record request interface.
			CountyFIPS: "11001",
			CourtName:  "DC Superior Court Marriage Bureau",
			CourtURL:   "https://www.dccourts.gov",
			SearchURL:  "https://www.dccourts.gov/superior-court/superior-court-divisions/family-court-operations-division/marriage/order-certified-copy-marriage-record",
			Note:       "Marriage records 1811–present; certified-copy request page; request-oriented, no online enumeration.",
		},
	},

	// --- Church (Catholic dioceses + parishes) ---------------------------
	Dioceses: []DioceseDef{
		{Slug: "washington", Name: "Archdiocese of Washington", Type: "archdiocese",
			Website: "https://adw.org", Directory: "https://adw.org/parishes", HubCityID: "city_washington_dc"},
	},

	// Washington DC-area parishes in the Archdiocese of Washington.
	// Names and addresses verified from the archdiocese's own parish directory
	// (adw.org) and each parish's website. Bulletin URLs verified by direct
	// search for each parish's bulletin archive.
	Parishes: []ParishDef{
		{
			DioceseSlug: "washington", Name: "Cathedral of St. Matthew the Apostle",
			Address:     "1725 Rhode Island Ave NW, Washington, DC 20036",
			BulletinURL: "https://www.stmatthewscathedral.org/news/bulletin",
		},
		{
			DioceseSlug: "washington", Name: "Holy Trinity Catholic Church",
			Address:     "1315 36th St NW, Washington, DC 20007",
			BulletinURL: "https://trinity.org/bulletins/",
		},
		{
			DioceseSlug: "washington", Name: "St. Patrick Catholic Church",
			Address: "619 10th St NW, Washington, DC 20001",
		},
		{
			DioceseSlug: "washington", Name: "St. Dominic Catholic Church",
			Address: "630 E St SW, Washington, DC 20024",
		},
		{
			DioceseSlug: "washington", Name: "St. Augustine Catholic Church",
			Address: "1419 V St NW, Washington, DC 20009",
		},
		{
			DioceseSlug: "washington", Name: "Church of the Annunciation",
			Address: "3810 Massachusetts Ave NW, Washington, DC 20016",
		},
	},

	// --- Social (wedding-industry Instagram vendors) ---------------------
	Vendors: []VendorDef{
		// DC photographers
		{
			Name: "Akbar Sayed Photography", OfficialURL: "https://akbarsayed.com/",
			Handle: "akbarsayedphotography", SourceClass: "engagement_photographer",
			CityID: "city_washington_dc", State: "DC", City: "Washington", Verified: "2026-08-01",
			TikTokHandle: "akbarsayedphotography",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-photographers/akbar-sayed-photography-3172842",
		},
		{
			Name: "DCorzo Photography", OfficialURL: "https://www.dcorzo.com/",
			Handle: "dcorzo_photography", SourceClass: "engagement_photographer",
			CityID: "city_washington_dc", State: "DC", City: "Washington", Verified: "2026-08-01",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-photographers/dcorzo-photography-9173407",
		},
		{
			Name: "Vness Photography", OfficialURL: "https://vnessphotography.com/",
			Handle: "vnessphotography", SourceClass: "engagement_photographer",
			CityID: "city_washington_dc", State: "DC", City: "Washington", Verified: "2026-08-01",
			TikTokHandle: "vnessphotography",
		},
		{
			Name: "Ayanah George Photography", OfficialURL: "https://ayanahgeorge.com/",
			Handle: "ayanahgeorge", SourceClass: "engagement_photographer",
			CityID: "city_washington_dc", State: "DC", City: "Washington", Verified: "2026-08-01",
		},
		// DC venues
		{
			Name: "District Winery", OfficialURL: "https://www.districtwinery.com/",
			Handle: "districtwinery", SourceClass: "wedding_venue",
			CityID: "city_washington_dc", State: "DC", City: "Washington", Verified: "2026-08-01",
			TikTokHandle: "districtwinery",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-venues/district-winery-4529669",
		},
		{
			Name: "Riggs Washington DC", OfficialURL: "https://www.riggsdc.com/",
			Handle: "riggshotel", SourceClass: "wedding_venue",
			CityID: "city_washington_dc", State: "DC", City: "Washington", Verified: "2026-08-01",
		},
		{
			Name: "Pendry Washington DC - The Wharf", OfficialURL: "https://www.pendry.com/washington-dc/",
			Handle: "pendrywharfdc", SourceClass: "wedding_venue",
			CityID: "city_washington_dc", State: "DC", City: "Washington", Verified: "2026-08-01",
			TikTokHandle: "pendrywharfdc",
		},
		// DC jewelers
		{
			Name: "Mervis Diamond Importers", OfficialURL: "https://www.mervisdiamond.com/",
			Handle: "mervisdiamond", SourceClass: "jeweler",
			CityID: "city_washington_dc", State: "DC", City: "Washington", Verified: "2026-08-01",
		},
		{
			Name: "Market Street Diamonds", OfficialURL: "https://www.marketstreetdiamonds.com/",
			Handle: "marketstreetdiamonds", SourceClass: "jeweler",
			CityID: "city_washington_dc", State: "DC", City: "Washington", Verified: "2026-08-01",
		},
		{
			Name: "Tiny Jewel Box", OfficialURL: "https://www.tinyjewelbox.com/",
			Handle: "tinyjewelboxdc", SourceClass: "jeweler",
			CityID: "city_washington_dc", State: "DC", City: "Washington", Verified: "2026-08-01",
		},
	},
}
