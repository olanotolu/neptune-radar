package packs

// Delaware source pack — verified 2026-08-01.
//
// Government: Delaware marriage records are held by each county's Clerk of
// the Peace (not a county clerk). New Castle, Sussex, and Kent counties each
// maintain their own marriage bureau. Search/request URLs verified against
// each county's official .gov site.
//
// Church: Diocese of Wilmington verified via USCCB + cdow.org. Parish names
// and addresses verified against Wikipedia's list of churches in the Diocese
// of Wilmington (which cites the diocese's own records) + each parish's own
// website. Bulletin URLs verified by direct search for each parish's bulletin
// archive.
//
// Social: Instagram handles verified from each business's own public website
// social links (or from IG search results where the handle was visible in
// the search snippet). Verification date recorded per vendor.

var dePack = StatePack{
	State: "DE",

	// --- Cities ----------------------------------------------------------
	Cities: []CityDef{
	{
		ID: "city_dover_de", State: "DE", County: "10001", Name: "Dover",
		Lat: 39.1582, Lng: -75.5244,
		Markets: []string{"dover", "kentcounty"},
	},
		{ID: "city_wilmington_de", State: "DE", County: "10003", Name: "Wilmington",
			Lat: 39.7391, Lng: -75.5398, Markets: []string{"wilmington", "de", "delaware", "newcastle"}},
	},

	// --- Government (Clerk of the Peace marriage-record searches) --------
	Government: []GovSource{
		{
			// New Castle County (Wilmington) — Clerk of the Peace marriage
			// license application + certified copy request portal.
			CountyFIPS: "10003",
			CourtName:  "New Castle County Clerk of the Peace",
			CourtURL:   "https://www.newcastlede.gov/3013/Clerk-of-the-Peace",
			SearchURL:  "https://www.newcastlede.gov/128/Certified-Copies-of-Records",
			Note:       "Clerk of the Peace issues marriage licenses and certified copies; online request portal via GovPilot; request-oriented, no public enumeration.",
		},
		{
			// Sussex County — Clerk of the Peace marriage bureau, certified
			// copy request page with online portal.
			CountyFIPS: "10005",
			CourtName:  "Sussex County Clerk of the Peace",
			CourtURL:   "https://sussexcountyde.gov/clerk-peace",
			SearchURL:  "https://sussexcountyde.gov/certified-copies-marriage-license",
			Note:       "Marriage Bureau issues licenses and certified copies for Sussex County; online request portal available; request-oriented, no public enumeration.",
		},
		{
			// Kent County — Clerk of the Peace marriage licenses page.
			CountyFIPS: "10001",
			CourtName:  "Kent County Clerk of the Peace",
			CourtURL:   "https://www.kentcountyde.gov/My-Government/Departments/Clerk-of-the-Peace",
			SearchURL:  "https://www.kentcountyde.gov/Residents/Marriage-Licenses",
			Note:       "Clerk of the Peace issues marriage licenses and performs ceremonies; in-person only, no online search portal.",
		},
	},

	// --- Church (Catholic dioceses + parishes) ---------------------------
	Dioceses: []DioceseDef{
		{Slug: "wilmington", Name: "Diocese of Wilmington", Type: "diocese",
			Website: "https://www.cdow.org", Directory: "https://cdow.org/find-a-parish/", HubCityID: "city_wilmington_de"},
	},

	// Wilmington-area parishes in the Diocese of Wilmington.
	// Names and addresses verified from Wikipedia's list of churches in the
	// diocese (which cites the diocese's own records) + each parish's own
	// website. Bulletin URLs verified by direct search for each parish's
	// bulletin archive.
	Parishes: []ParishDef{
		{
			DioceseSlug: "wilmington", Name: "Cathedral of St. Peter",
			Address:     "500 West St, Wilmington, DE 19801",
			BulletinURL: "https://www.downtowncatholic.com/bulletin.html",
		},
		{
			DioceseSlug: "wilmington", Name: "St. Anthony of Padua Catholic Church",
			Address:     "901 N. Dupont St, Wilmington, DE 19805",
			BulletinURL: "https://sapde.org/bulletin-archive/",
		},
		{
			DioceseSlug: "wilmington", Name: "St. Elizabeth Catholic Church",
			Address:     "809 S. Broom St, Wilmington, DE 19805",
			BulletinURL: "https://steparish.org/bulletins",
		},
		{
			DioceseSlug: "wilmington", Name: "St. Joseph on the Brandywine Catholic Church",
			Address:     "10 Old Church Rd, Greenville, DE 19807",
			BulletinURL: "https://stjosephonthebrandywine.org/bulletins",
		},
		{
			DioceseSlug: "wilmington", Name: "St. Mary Magdalen Catholic Church",
			Address:     "7 Sharpley Rd, Wilmington, DE 19803",
			BulletinURL: "https://smmchurch.org",
		},
		{
			DioceseSlug: "wilmington", Name: "Immaculate Heart of Mary Catholic Church",
			Address:     "4701 Weldin Rd, Wilmington, DE 19803",
			BulletinURL: "https://www.ihm.org",
		},
		{
			DioceseSlug: "wilmington", Name: "Holy Rosary Catholic Church",
			Address:     "3200 Philadelphia Pike, Claymont, DE 19703",
			BulletinURL: "https://www.hrparish.com",
		},
		{
			DioceseSlug: "wilmington", Name: "St. Ann Catholic Church",
			Address: "2013 Gilpin Ave, Wilmington, DE 19806",
		},
	},

	// --- Social (wedding-industry Instagram vendors) ---------------------
	Vendors: []VendorDef{
		// Wilmington photographers
		{
			Name: "Becca Mathias Photography", OfficialURL: "https://www.beccamathiasphoto.com/",
			Handle: "beccamathiasphoto", SourceClass: "engagement_photographer",
			CityID: "city_wilmington_de", State: "DE", City: "Wilmington", Verified: "2026-08-01",
			TikTokHandle: "beccamathiasphoto",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-photographers/becca-mathias-photography-5124625",
		},
		{
			Name: "Kerry Harrison Photography", OfficialURL: "https://www.kerryharrison.net/",
			Handle: "kerryharrisonphotography", SourceClass: "engagement_photographer",
			CityID: "city_wilmington_de", State: "DE", City: "Wilmington", Verified: "2026-08-01",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-photographers/kerry-harrison-photography-6136126",
		},
		{
			Name: "Creative Image Weddings", OfficialURL: "https://creativeimageweddings.com/",
			Handle: "creativeimageweddings", SourceClass: "engagement_photographer",
			CityID: "city_wilmington_de", State: "DE", City: "Wilmington", Verified: "2026-08-01",
			TikTokHandle: "creativeimageweddings",
		},
		{
			Name: "Gretchen Johnson Weddings", OfficialURL: "https://gretchenjohnsonweddings.com/",
			Handle: "gretchenjohnsonphoto", SourceClass: "engagement_photographer",
			CityID: "city_wilmington_de", State: "DE", City: "Wilmington", Verified: "2026-08-01",
		},
		// Wilmington venues
		{
			Name: "The Farmhouse", OfficialURL: "https://www.thefarmhousede.com/",
			Handle: "thefarmhousede", SourceClass: "wedding_venue",
			CityID: "city_wilmington_de", State: "DE", City: "Wilmington", Verified: "2026-08-01",
			TikTokHandle: "thefarmhousede",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-venues/the-farmhouse-5355466",
		},
		{
			Name: "Hotel Du Pont", OfficialURL: "https://www.hoteldupont.com/weddings-events/weddings",
			Handle: "hotel_dupont", SourceClass: "wedding_venue",
			CityID: "city_wilmington_de", State: "DE", City: "Wilmington", Verified: "2026-08-01",
		},
		// Wilmington jewelers
		{
			Name: "A.R. Morris Jewelers", OfficialURL: "https://www.armorrisjewelers.com/",
			Handle: "armorrisjewelers", SourceClass: "jeweler",
			CityID: "city_wilmington_de", State: "DE", City: "Wilmington", Verified: "2026-08-01",
		},
		{
			Name: "Stephen's Jewelers", OfficialURL: "https://stephensjewelers.com/",
			Handle: "stephensjewelers", SourceClass: "jeweler",
			CityID: "city_wilmington_de", State: "DE", City: "Wilmington", Verified: "2026-08-01",
		},
		{
			Name: "Stone Weddings", OfficialURL: "https://www.stoneweddings.com/",
			Handle: "stoneweddings", SourceClass: "engagement_photographer",
			CityID: "city_wilmington_de", State: "DE", City: "Wilmington", Verified: "2026-08-03",
		},
		{
			Name: "King Cole Farm", OfficialURL: "https://www.kingcolefarm.com/",
			Handle: "kingcolefarm", SourceClass: "wedding_venue",
			CityID: "city_dover_de", State: "DE", City: "Dover", Verified: "2026-08-03",
		},
		{
			Name: "Belak Flowers", OfficialURL: "https://www.belak-flowers.com/",
			Handle: "belakflowers.wilmington", SourceClass: "florist",
			CityID: "city_wilmington_de", State: "DE", City: "Wilmington", Verified: "2026-08-03",
		},
	},
}
