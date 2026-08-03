package packs

// Arizona source pack — verified 2026-08-01.
//
// Government: Arizona marriage records are held by the Clerk of the Superior
// Court in each county. Search URLs for the top 7 counties by population were
// verified against each county's official .gov or civicplus site. Most AZ
// counties do not offer a public online marriage-record search portal;
// instead they provide request-oriented pages with instructions for obtaining
// copies. Pima County is the exception, offering an Agave Online public case
// search that includes marriage license cases.
//
// Church: both Arizona dioceses (Phoenix and Tucson) verified via USCCB +
// each diocese's own website. Phoenix-area parishes (Diocese of Phoenix) were
// verified against the Wikipedia list of churches in the diocese (which cites
// the diocese's own records) + direct bulletin-archive URL discovery.
//
// Social: Instagram handles verified from each business's own public website
// social links (or from IG search results where the site is JS-rendered and
// the handle was visible in the search snippet). Verification date recorded
// per vendor.

var azPack = StatePack{
	State: "AZ",

	// --- Cities ----------------------------------------------------------
	Cities: []CityDef{
		{ID: "city_phoenix_az", State: "AZ", County: "04013", Name: "Phoenix",
			Lat: 33.4484, Lng: -112.0740, Markets: []string{"phoenix", "phx", "maricopa", "az"}},
		{ID: "city_tucson_az", State: "AZ", County: "04019", Name: "Tucson",
			Lat: 32.2226, Lng: -110.9747, Markets: []string{"tucson", "pima", "southernaz"}},
	},

	// --- Government (Clerk of Superior Court marriage-record sources) ----
	Government: []GovSource{
		{
			// Maricopa County (Phoenix) — Clerk of Superior Court. Marriage
			// license copies page; public access terminals offer a Marriage
			// License Look Up via iCIS.
			CountyFIPS: "04013",
			CourtName:  "Maricopa County Clerk of Superior Court",
			CourtURL:   "https://www.clerkofcourt.maricopa.gov",
			SearchURL:  "https://www.clerkofcourt.maricopa.gov/records/obtaining-records/marriage-license-copies",
			Note:       "Marriage license copies page; iCIS terminal lookup available in-person; no public online search portal.",
		},
		{
			// Pima County (Tucson) — Clerk of Superior Court. Agave Online
			// public record search includes marriage license cases (case
			// numbers starting with "M").
			CountyFIPS: "04019",
			CourtName:  "Pima County Clerk of Superior Court",
			CourtURL:   "https://www.cosc.pima.gov",
			SearchURL:  "https://agave.cosc.pima.gov/PublicDocs",
			Note:       "Agave Online public case search; marriage license cases searchable by name; enumeration candidate.",
		},
		{
			// Pinal County — Clerk of Superior Court. Marriage license copies
			// page with request instructions.
			CountyFIPS: "04021",
			CourtName:  "Pinal County Clerk of Superior Court",
			CourtURL:   "https://www.coscpinalcountyaz.gov",
			SearchURL:  "https://www.coscpinalcountyaz.gov/161/Obtaining-Copies-of-a-Marriage-License",
			Note:       "Marriage license copies request page; no public online search portal; records from 1875 to present.",
		},
		{
			// Yavapai County — Clerk of Superior Court. Marriage licenses page
			// with certified copy request instructions.
			CountyFIPS: "04025",
			CourtName:  "Yavapai County Clerk of Superior Court",
			CourtURL:   "https://courts.yavapaiaz.gov",
			SearchURL:  "https://courts.yavapaiaz.gov/Departments/Clerk/Marriage-Licenses",
			Note:       "Marriage licenses page; certified copies $35; search fee $35 if Clerk searches; no public online search portal.",
		},
		{
			// Mohave County — Clerk of Superior Court. Marriage licenses page
			// with abstract request instructions.
			CountyFIPS: "04015",
			CourtName:  "Mohave County Clerk of Superior Court",
			CourtURL:   "https://www.mohavecourts.com",
			SearchURL:  "https://www.mohavecourts.com/court-departments/clerk-superior-court/marriage-licenses",
			Note:       "Marriage licenses page; abstracts $30–$35; records request form for copies; no public online search portal.",
		},
		{
			// Yuma County — Clerk of Superior Court. Marriage license page
			// with copy request instructions.
			CountyFIPS: "04027",
			CourtName:  "Yuma County Clerk of Superior Court",
			CourtURL:   "https://www.yumacountyaz.gov/government/courts/clerk-of-superior-court",
			SearchURL:  "https://www.yumacountyaz.gov/government/courts/clerk-of-superior-court/marriage-license",
			Note:       "Marriage license page; certified copies $35; records on paper/microfilm/microfiche; no public online search portal.",
		},
		{
			// Coconino County (Flagstaff) — Clerk of Superior Court. Marriage
			// licenses page with affidavit ordering instructions.
			CountyFIPS: "04005",
			CourtName:  "Coconino County Clerk of Superior Court",
			CourtURL:   "https://az-coconinocounty2.civicplus.com/132/Clerk-of-the-Superior-Court",
			SearchURL:  "https://az-coconinocounty2.civicplus.com/134/Marriage-Licenses",
			Note:       "Marriage licenses page; affidavit of record $35; online ordering available for a fee; no public online search portal.",
		},
	},

	// --- Church (Catholic dioceses + parishes) ---------------------------
	Dioceses: []DioceseDef{
		{Slug: "phoenix", Name: "Diocese of Phoenix", Type: "diocese",
			Website: "https://dphx.org", Directory: "https://dphx.org/parishes", HubCityID: "city_phoenix_az"},
		{Slug: "tucson", Name: "Diocese of Tucson", Type: "diocese",
			Website: "https://www.diocesetucson.org", Directory: "https://www.diocesetucson.org/parishes", HubCityID: "city_tucson_az"},
	},

	// Phoenix-area parishes in the Diocese of Phoenix. Names and addresses
	// verified from Wikipedia's list of churches in the diocese (which cites
	// the diocese's own records) + each parish's own website. Bulletin URLs
	// verified by direct access to each parish's bulletin archive page.
	Parishes: []ParishDef{
		{
			DioceseSlug: "phoenix", Name: "Ss. Simon and Jude Cathedral",
			Address: "6351 N 27th Ave, Phoenix, AZ 85017",
		},
		{
			DioceseSlug: "phoenix", Name: "St. Mary's Basilica",
			Address: "231 N 3rd St, Phoenix, AZ 85004",
		},
		{
			DioceseSlug: "phoenix", Name: "St. Francis Xavier Catholic Church",
			Address:     "4715 N Central Ave, Phoenix, AZ 85012",
			BulletinURL: "https://www.sfxphx.org/event",
		},
		{
			DioceseSlug: "phoenix", Name: "St. Gregory Catholic Church",
			Address:     "3424 N 18th Ave, Phoenix, AZ 85015",
			BulletinURL: "https://www.stgregoryphx.org/bulletin9ab2b387",
		},
		{
			DioceseSlug: "phoenix", Name: "Our Lady of Mount Carmel Catholic Church",
			Address: "2121 S Rural Rd, Tempe, AZ 85282",
		},
		{
			DioceseSlug: "phoenix", Name: "St. Thomas More Catholic Church",
			Address: "6180 W Utopia Rd, Glendale, AZ 85308",
		},
		{
			DioceseSlug: "phoenix", Name: "St. Anne Roman Catholic Parish",
			Address: "440 E Elliot Rd, Gilbert, AZ 85234",
		},
		{
			DioceseSlug: "phoenix", Name: "St. Jerome Catholic Church",
			Address:     "10815 N 35th Ave, Phoenix, AZ 85029",
			BulletinURL: "https://www.saintjerome.org/bulletins/",
		},
	},

	// --- Social (wedding-industry Instagram vendors) ---------------------
	Vendors: []VendorDef{
		// Phoenix photographers
		{
			Name: "Patrick Julian Photography", OfficialURL: "https://www.patrickjulianphotography.com/",
			Handle: "patrick_julian_photography", SourceClass: "engagement_photographer",
			CityID: "city_phoenix_az", State: "AZ", City: "Phoenix", Verified: "2026-08-01",
			TikTokHandle: "patrick_julian_photography",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-photographers/patrick-julian-photography-2725329",
		},
		{
			Name: "David Orr Photography", OfficialURL: "https://www.davidorrphotography.com/",
			Handle: "davidorrphotography", SourceClass: "engagement_photographer",
			CityID: "city_phoenix_az", State: "AZ", City: "Phoenix", Verified: "2026-08-01",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-photographers/david-orr-photography-4584003",
		},
		{
			Name: "Mariel Jenny Photography", OfficialURL: "https://www.marieljennyphotography.com/",
			Handle: "marieljenny.photography", SourceClass: "engagement_photographer",
			CityID: "city_phoenix_az", State: "AZ", City: "Phoenix", Verified: "2026-08-01",
			TikTokHandle: "marieljenny.photography",
		},
		{
			Name: "Ernesto Jase", OfficialURL: "https://ernestojase.com/",
			Handle: "ernestojase.photography", SourceClass: "engagement_photographer",
			CityID: "city_phoenix_az", State: "AZ", City: "Phoenix", Verified: "2026-08-01",
		},
		// Phoenix venues
		{
			Name: "The Paseo", OfficialURL: "https://thepaseovenue.com/",
			Handle: "thepaseovenue", SourceClass: "wedding_venue",
			CityID: "city_phoenix_az", State: "AZ", City: "Phoenix", Verified: "2026-08-01",
			TikTokHandle: "thepaseovenue",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-venues/the-paseo-8009171",
		},
		{
			Name: "Venue at the Grove", OfficialURL: "https://venueatthegrove.com/",
			Handle: "venueatthegrove", SourceClass: "wedding_venue",
			CityID: "city_phoenix_az", State: "AZ", City: "Phoenix", Verified: "2026-08-01",
		},
		// Phoenix jewelers
		{
			Name: "Schmitt Jewelers", OfficialURL: "https://schmittjewelers.com/",
			Handle: "schmittjewelers", SourceClass: "jeweler",
			CityID: "city_phoenix_az", State: "AZ", City: "Phoenix", Verified: "2026-08-01",
		},
		{
			Name: "Demirjian Jewelry Design", OfficialURL: "https://demirjiandesign.com/",
			Handle: "demirjianjewelry", SourceClass: "jeweler",
			CityID: "city_phoenix_az", State: "AZ", City: "Phoenix", Verified: "2026-08-01",
		},
		{
			Name: "Christopher Fine Diamonds", OfficialURL: "https://christopherfinediamonds.com/",
			Handle: "christopherfinediamonds", SourceClass: "jeweler",
			CityID: "city_phoenix_az", State: "AZ", City: "Phoenix", Verified: "2026-08-01",
		},
		{
			Name: "Snapdragon Bloom Bar", OfficialURL: "https://snapdragonbloombar.com/",
			Handle: "arizonaweddingflorist", SourceClass: "florist",
			CityID: "city_phoenix_az", State: "AZ", City: "Phoenix", Verified: "2026-08-03",
		},
		{
			Name: "Bloom and Blueprint", OfficialURL: "https://bloomandblueprint.com/",
			Handle: "bloomandblueprint", SourceClass: "wedding_planner",
			CityID: "city_phoenix_az", State: "AZ", City: "Phoenix", Verified: "2026-08-03",
		},
		{
			Name: "C West Entertainment", OfficialURL: "https://www.djcwest.com/",
			Handle: "cwestent", SourceClass: "videographer",
			CityID: "city_phoenix_az", State: "AZ", City: "Phoenix", Verified: "2026-08-03",
		},
		{
			Name: "Perfectly Planned Celebrations", OfficialURL: "https://www.perfectlyplannedbycandida.com/",
			Handle: "perfectlyplannedcelebrations", SourceClass: "wedding_planner",
			CityID: "city_tucson_az", State: "AZ", City: "Tucson", Verified: "2026-08-03",
		},
		{
			Name: "Shua Photography", OfficialURL: "https://www.yourtucsonwedding.com/",
			Handle: "shuaphoto520", SourceClass: "engagement_photographer",
			CityID: "city_tucson_az", State: "AZ", City: "Tucson", Verified: "2026-08-03",
		},
	},
}
