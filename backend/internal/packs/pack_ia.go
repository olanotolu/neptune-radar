package packs

// Iowa source pack — verified 2026-08-01.
//
// Government: Iowa marriage records are held by the county recorder (not the
// clerk). Unlike many states, Iowa counties do NOT offer public online search
// portals for marriage records — records are closed to public inspection at
// the state level (Iowa Code §144) but open at the county level (IAC 144.43).
// Certified copies require in-person or mail requests with proof of
// entitlement and notarized application. The SearchURL below points at each
// county recorder's marriage/vital-records page where the process and
// application forms are hosted.
//
// Church: all 4 Iowa Catholic dioceses/archdioceses verified via USCCB +
// each diocese's own website. Des Moines-area parishes (Diocese of Des Moines)
// verified against the diocese's own parish finder
// (dmdiocese.org/worship/parishes-and-mass-times). Bulletin URLs verified by
// direct search for each parish's bulletin archive.
//
// Social: Instagram handles verified from each business's own public website
// social links (or from IG search results where the site is JS-rendered and
// the handle was visible in the search snippet). Verification date recorded
// per vendor.

var iaPack = StatePack{
	State: "IA",

	// --- Cities ----------------------------------------------------------
	Cities: []CityDef{
		{ID: "city_des_moines_ia", State: "IA", County: "19049", Name: "Des Moines",
			Lat: 41.5868, Lng: -93.6250, Markets: []string{"desmoines", "dsm", "iowa", "polk"}},
		{ID: "city_cedar_rapids_ia", State: "IA", County: "19055", Name: "Cedar Rapids",
			Lat: 42.0083, Lng: -91.6436, Markets: []string{"cedarrapids", "linn", "iowa"}},
	},

	// --- Government (county recorder marriage-record pages) --------------
	// Iowa has no public online marriage search; these are the recorder
	// vital-records pages where applications and instructions are hosted.
	Government: []GovSource{
		{
			// Polk County (Des Moines) — recorder vital records / marriage.
			CountyFIPS: "19153",
			CourtName:  "Polk County Recorder",
			CourtURL:   "https://www.polkcountyiowa.gov/county-recorder/",
			SearchURL:  "https://www.polkcountyiowa.gov/county-recorder/vital-records/marriage-records/",
			Note:       "Marriage records 1880–present; no online search, in-person/mail with entitlement proof.",
		},
		{
			// Linn County (Cedar Rapids) — recorder marriage license page.
			CountyFIPS: "19113",
			CourtName:  "Linn County Recorder",
			CourtURL:   "http://www.linncounty.org/149",
			SearchURL:  "http://www.linncounty.org/1145/Marriage-License",
			Note:       "Marriage records 1880–1920 & 1942–present; no online search, in-person/mail with entitlement proof.",
		},
		{
			// Scott County (Davenport) — recorder marriage page.
			CountyFIPS: "19163",
			CourtName:  "Scott County Recorder",
			CourtURL:   "https://www.scottcountyiowa.gov/recorder",
			SearchURL:  "https://www.scottcountyiowa.gov/recorder/marriage",
			Note:       "Marriage records 1942–present; no online search, in-person/mail with entitlement proof.",
		},
		{
			// Black Hawk County (Waterloo) — recorder vital records.
			CountyFIPS: "19013",
			CourtName:  "Black Hawk County Recorder",
			CourtURL:   "https://blackhawkcounty.iowa.gov/239/Recorder",
			SearchURL:  "https://blackhawkcounty.iowa.gov/401/Vital-Records",
			Note:       "Marriage records via recorder; no online search, in-person/mail with entitlement proof.",
		},
		{
			// Johnson County (Iowa City) — recorder marriage records page.
			CountyFIPS: "19103",
			CourtName:  "Johnson County Recorder",
			CourtURL:   "https://johnsoncountyiowa.gov/recorder",
			SearchURL:  "https://johnsoncountyiowa.gov/recorder/marriage-records",
			Note:       "Marriage records via recorder; no online search, in-person/mail with entitlement proof.",
		},
		{
			// Woodbury County (Sioux City) — recorder marriage records page.
			CountyFIPS: "19193",
			CourtName:  "Woodbury County Recorder",
			CourtURL:   "https://www.woodburycountyiowa.gov/recorder/",
			SearchURL:  "https://www.woodburycountyiowa.gov/recorder/marriage_records_licenses/",
			Note:       "Marriage records 1880–present; no online search, in-person/mail with entitlement proof.",
		},
		{
			// Dubuque County — recorder vital records / marriage certificates.
			CountyFIPS: "19061",
			CourtName:  "Dubuque County Recorder",
			CourtURL:   "https://www.dubuquecountyiowa.gov/212/Recorder",
			SearchURL:  "https://www.dubuquecountyiowa.gov/222/Birth-Death-Marriage-Certificates",
			Note:       "Marriage records 1839–present; no online search, in-person/mail with entitlement proof.",
		},
	},

	// --- Church (Catholic dioceses + parishes) ---------------------------
	Dioceses: []DioceseDef{
		{Slug: "dubuque", Name: "Archdiocese of Dubuque", Type: "archdiocese",
			Website: "https://www.dbqarch.org", Directory: "https://www.dbqarch.org/parishes"},
		{Slug: "davenport", Name: "Diocese of Davenport", Type: "diocese",
			Website: "https://www.davenportdiocese.org", Directory: "https://www.davenportdiocese.org/parishes"},
		{Slug: "des_moines", Name: "Diocese of Des Moines", Type: "diocese",
			Website: "https://www.dmdiocese.org", Directory: "https://www.dmdiocese.org/parishes", HubCityID: "city_des_moines_ia"},
		{Slug: "sioux_city", Name: "Diocese of Sioux City", Type: "diocese",
			Website: "https://www.scdiocese.org", Directory: "https://www.scdiocese.org/parishes"},
	},

	// Des Moines-area parishes in the Diocese of Des Moines. Names and
	// addresses verified from the diocese's own parish finder
	// (dmdiocese.org/worship/parishes-and-mass-times). Bulletin URLs verified
	// by direct search for each parish's bulletin archive.
	Parishes: []ParishDef{
		{
			DioceseSlug: "des_moines", Name: "St. Ambrose Cathedral",
			Address: "607 High St, Des Moines, IA 50309",
		},
		{
			DioceseSlug: "des_moines", Name: "Basilica of St. John",
			Address:     "1915 University Ave, Des Moines, IA 50314",
			BulletinURL: "https://basilicaofstjohn.org/bulletins-1",
		},
		{
			DioceseSlug: "des_moines", Name: "Holy Trinity Catholic Church",
			Address:     "2926 Beaver Ave, Des Moines, IA 50310",
			BulletinURL: "https://holytrinitydm.org/home/bulletin/",
		},
		{
			DioceseSlug: "des_moines", Name: "St. Augustin Catholic Church",
			Address: "545 42nd St, Des Moines, IA 50312",
		},
		{
			DioceseSlug: "des_moines", Name: "Christ the King Catholic Church",
			Address: "5711 SW 9th St, Des Moines, IA 50315",
		},
		{
			DioceseSlug: "des_moines", Name: "All Saints Catholic Church",
			Address: "650 NE 52nd Ave, Des Moines, IA 50313",
		},
	},

	// --- Social (wedding-industry Instagram vendors) ---------------------
	Vendors: []VendorDef{
		// Des Moines photographers
		{
			Name: "Morgan Moon Photography", OfficialURL: "https://morganmoonphotography.com/",
			Handle: "morganmoonphotography", SourceClass: "engagement_photographer",
			CityID: "city_des_moines_ia", State: "IA", City: "Des Moines", Verified: "2026-08-01",
			TikTokHandle: "morganmoonphotography",
		},
		{
			Name: "Christina Ney Photography", OfficialURL: "https://christinaney.com/",
			Handle: "christinaneyphotography", SourceClass: "engagement_photographer",
			CityID: "city_des_moines_ia", State: "IA", City: "Des Moines", Verified: "2026-08-01",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-photographers/christina-ney-photography-9324326",
		},
		{
			Name: "Jenna McEntee Photography", OfficialURL: "https://jennamcenteephotography.com/",
			Handle: "jennamcenteephotography", SourceClass: "engagement_photographer",
			CityID: "city_des_moines_ia", State: "IA", City: "Des Moines", Verified: "2026-08-01",
			TikTokHandle: "jennamcenteephotography",
		},
		// Des Moines venues
		{
			Name: "Willow on Grand", OfficialURL: "https://willowongrand.com/",
			Handle: "willowongrand_dsm", SourceClass: "wedding_venue",
			CityID: "city_des_moines_ia", State: "IA", City: "Des Moines", Verified: "2026-08-01",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-venues/willow-on-grand-9261803",
		},
		{
			Name: "Greater Des Moines Botanical Garden", OfficialURL: "https://dmbotanicalgarden.com/",
			Handle: "dmbotanicalgarden", SourceClass: "wedding_venue",
			CityID: "city_des_moines_ia", State: "IA", City: "Des Moines", Verified: "2026-08-01",
			TikTokHandle: "dmbotanicalgarden",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-venues/greater-des-moines-botanical-garden-5010938",
		},
		{
			Name: "Des Moines Heritage Center", OfficialURL: "https://www.desmoinesheritagecenter.org/",
			Handle: "dsmheritagecenter", SourceClass: "wedding_venue",
			CityID: "city_des_moines_ia", State: "IA", City: "Des Moines", Verified: "2026-08-01",
		},
		{
			Name: "The Vineyard at St. Charles", OfficialURL: "https://www.vineyardatstcharles.com/",
			Handle: "vineyardatstcharles", SourceClass: "wedding_venue",
			CityID: "city_des_moines_ia", State: "IA", City: "Des Moines", Verified: "2026-08-01",
			TikTokHandle: "vineyardatstcharles",
		},
		// Des Moines jewelers
		{
			Name: "Josephs Jewelers", OfficialURL: "https://josephsjewelers.com/",
			Handle: "josephsjewelers1871", SourceClass: "jeweler",
			CityID: "city_des_moines_ia", State: "IA", City: "Des Moines", Verified: "2026-08-01",
		},
		{
			Name: "Christopher's Fine Jewelry", OfficialURL: "https://christophersjewelry.com/",
			Handle: "christophersjewelrydsm", SourceClass: "jeweler",
			CityID: "city_des_moines_ia", State: "IA", City: "Des Moines", Verified: "2026-08-01",
		},
	},
}
