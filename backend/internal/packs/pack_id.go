package packs

// Idaho source pack — verified 2026-08-01.
//
// Government: Idaho marriage records are held by the county recorder/clerk.
// Search URLs for the top 5 counties were verified against each county's
// official .gov site or its Tylerhost self-service portal.
//
// Church: the Diocese of Boise was verified via USCCB + the diocese's own
// website. Boise-area parishes were verified against the diocese parish
// directory + direct bulletin-archive URL discovery.
//
// Social: Instagram handles verified from each business's own public website
// social links (or from IG search results where the site is JS-rendered and
// the handle was visible in the search snippet). Verification date recorded
// per vendor.

var idPack = StatePack{
	State: "ID",

	// --- Cities ----------------------------------------------------------
	Cities: []CityDef{
		{
			ID: "city_boise_id", State: "ID", County: "16001", Name: "Boise",
			Lat: 43.6150, Lng: -116.2023,
			Markets: []string{"boise", "id", "ada", "idaho"},
		},
	},

	// --- Government (county clerk marriage-record searches) --------------
	Government: []GovSource{
		{
			// Ada County (Boise) — county clerk marriage license page;
			// public records request form for copies.
			CountyFIPS: "16001",
			CourtName:  "Ada County Clerk",
			CourtURL:   "https://adacounty.id.gov/clerk/",
			SearchURL:  "https://adacounty.id.gov/clerk/marriage-license/",
			Note:       "Marriage license page + public records request form; no online index search; request-oriented.",
		},
		{
			// Canyon County — online marriage application + recorded document
			// index (1984–present) available at recorder's office terminals.
			CountyFIPS: "16027",
			CourtName:  "Canyon County Recorder",
			CourtURL:   "https://www.canyoncounty.id.gov/elected-officials/clerk/recorder/",
			SearchURL:  "https://rec-search.canyoncounty.id.gov/Marriage/Home.aspx",
			Note:       "Online marriage application system; document index from 1984 available in-office; online index search needs testing.",
		},
		{
			// Kootenai County — recorder search portal with marriage
			// application and document search.
			CountyFIPS: "16055",
			CourtName:  "Kootenai County Recorder",
			CourtURL:   "https://kcgov.us/351/Marriage-Licenses",
			SearchURL:  "https://www.kcgov.us/372/Recorder-Search",
			Note:       "Recorder search portal with real estate and marriage sections; enumeration capability needs testing.",
		},
		{
			// Bonneville County — Tylerhost self-service web with document
			// search and marriage application.
			CountyFIPS: "16019",
			CourtName:  "Bonneville County Clerk",
			CourtURL:   "https://www.bonnevillecountyidaho.gov/administration/clerk-auditor-and-recorder",
			SearchURL:  "https://bonnevillecountyid-web.tylerhost.net/bonnevillekiosk/",
			Note:       "Tylerhost self-service portal with document search; marriage records indexed; enumeration candidate.",
		},
		{
			// Bannock County — recorder page; no online search portal.
			CountyFIPS: "16005",
			CourtName:  "Bannock County Recorder",
			CourtURL:   "https://www.bannockcounty.gov/recorder/",
			SearchURL:  "https://www.bannockcounty.gov/recorder/",
			Note:       "No online search portal; marriage records by public records request form or in-person only.",
		},
	},

	// --- Church (Catholic dioceses + parishes) ---------------------------
	Dioceses: []DioceseDef{
		{
			Slug: "boise", Name: "Diocese of Boise", Type: "diocese",
			Website: "https://www.catholicidaho.org", Directory: "https://www.catholicidaho.org/parishes",
			HubCityID: "city_boise_id",
		},
	},

	// Boise-area parishes in the Diocese of Boise.
	// Bulletin URLs verified by direct search for each parish's bulletin
	// archive.
	Parishes: []ParishDef{
		{
			DioceseSlug: "boise", Name: "Cathedral of St. John the Evangelist",
			Address:     "707 N 8th St, Boise, ID 83702",
			BulletinURL: "https://www.boisecathedral.org/bulletins",
		},
		{
			DioceseSlug: "boise", Name: "Our Lady of the Rosary",
			Address:     "1500 E Wright St, Boise, ID 83706",
			BulletinURL: "https://olrboise.org/bulletin",
		},
		{
			DioceseSlug: "boise", Name: "Our Lady of Good Counsel",
			Address:     "6244 W State St, Boise, ID 83714",
			BulletinURL: "https://goodcounselidaho.org/ParishBulletin.html",
		},
		{
			DioceseSlug: "boise", Name: "St. Mark's Catholic Church",
			Address:     "7960 Northview St, Boise, ID 83704",
			BulletinURL: "https://www.stmarksboise.org/bulletin",
		},
	},

	// --- Social (wedding-industry Instagram vendors) ---------------------
	Vendors: []VendorDef{
		// Boise photographers
		{
			Name: "Erin Egleston Photography", OfficialURL: "https://erinegleston.com/",
			Handle: "erin_egleston", SourceClass: "engagement_photographer",
			CityID: "city_boise_id", State: "ID", City: "Boise", Verified: "2026-08-01",
			TikTokHandle: "erin_egleston",
		},
		{
			Name: "Shayne Center Photo", OfficialURL: "https://shaynecenterphoto.com/",
			Handle: "shaynecenterphoto", SourceClass: "engagement_photographer",
			CityID: "city_boise_id", State: "ID", City: "Boise", Verified: "2026-08-01",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-photographers/shayne-center-photo-2306822",
		},
		{
			Name: "Kassie Gunn Photo", OfficialURL: "https://kassiegunn.com/",
			Handle: "kassiegunnphoto", SourceClass: "engagement_photographer",
			CityID: "city_boise_id", State: "ID", City: "Boise", Verified: "2026-08-01",
			TikTokHandle: "kassiegunnphoto",
		},
		{
			Name: "Jenesis SC Photography", OfficialURL: "https://jenesisscphotography.com/",
			Handle: "jenesisscphotography", SourceClass: "engagement_photographer",
			CityID: "city_boise_id", State: "ID", City: "Boise", Verified: "2026-08-01",
		},
		{
			Name: "Yasinski Photography", OfficialURL: "https://www.yasinskiphotography.com/",
			Handle: "yasinskiphotography", SourceClass: "engagement_photographer",
			CityID: "city_boise_id", State: "ID", City: "Boise", Verified: "2026-08-01",
			TikTokHandle: "yasinskiphotography",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-photographers/yasinski-photography-7597301",
		},
		// Boise venues
		{
			Name: "Stone Crossing", OfficialURL: "https://stonecrossing.com/",
			Handle: "stone_crossing", SourceClass: "wedding_venue",
			CityID: "city_boise_id", State: "ID", City: "Boise", Verified: "2026-08-01",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-venues/stone-crossing-6221625",
		},
		{
			Name: "The Avery Hotel & Brasserie", OfficialURL: "https://theaveryboise.com/",
			Handle: "theaveryhotelboise", SourceClass: "wedding_venue",
			CityID: "city_boise_id", State: "ID", City: "Boise", Verified: "2026-08-01",
			TikTokHandle: "theaveryhotelboise",
		},
		{
			Name: "Pastor Dave Wedding Officiant", OfficialURL: "https://weddingsbypastordave.com/",
			Handle: "weddingpastordave", SourceClass: "officiant",
			CityID: "city_boise_id", State: "ID", City: "Boise", Verified: "2026-08-03",
		},
		{
			Name: "Ivoire Bridal Atelier", OfficialURL: "https://ivoirebridalboise.com/",
			Handle: "ivoirebridalatelier", SourceClass: "bridal_shop",
			CityID: "city_boise_id", State: "ID", City: "Boise", Verified: "2026-08-03",
		},
		{
			Name: "Karli & David Photography", OfficialURL: "https://karlianddavid.com/",
			Handle: "karlianddavid", SourceClass: "engagement_photographer",
			CityID: "city_boise_id", State: "ID", City: "Boise", Verified: "2026-08-03",
		},
		{
			Name: "Katy Kahla Photography", OfficialURL: "https://www.katykahlaphotography.com/",
			Handle: "katykahlaphotography", SourceClass: "engagement_photographer",
			CityID: "city_boise_id", State: "ID", City: "Boise", Verified: "2026-08-03",
		},
	},
}
