package packs

// Louisiana source pack — verified 2026-08-01.
//
// Government: Louisiana marriage records are held by the parish Clerk of
// Court, except Orleans Parish which is unique — its records are maintained
// by the LA Vital Records Registry (<50 yrs) and the State Archives (>50 yrs).
// Search URLs for the top 7 parishes by population were verified against each
// clerk's official site or the statewide eClerksLA / LCRAA portals.
//
// Church: all 7 Louisiana Catholic dioceses/archdiocese verified via USCCB +
// each diocese's own website. New Orleans-area parishes (Archdiocese of New
// Orleans) were verified against the archdiocese's parish finder
// (nolacatholic.org/parishfinder) + direct bulletin-archive URL discovery.
//
// Social: Instagram handles verified from each business's own public website
// social links. Verification date recorded per vendor.

var laPack = StatePack{
	State: "LA",

	// --- Cities ----------------------------------------------------------
	Cities: []CityDef{
		{ID: "city_new_orleans_la", State: "LA", County: "22071", Name: "New Orleans",
			Lat: 29.9511, Lng: -90.0715, Markets: []string{"neworleans", "nola", "orleans", "la"}},
		{ID: "city_baton_rouge_la", State: "LA", County: "22033", Name: "Baton Rouge",
			Lat: 30.4515, Lng: -91.1871, Markets: []string{"batonrouge", "ebr", "la"}},
	},

	// --- Government (parish clerk marriage-record searches) --------------
	Government: []GovSource{
		{
			// Orleans Parish — unique in Louisiana: marriage records are
			// maintained by the LA Vital Records Registry (<50 yrs) and the
			// State Archives (>50 yrs), not the Clerk of Court. The State
			// Archives search covers Orleans 1870–1975.
			CountyFIPS: "22071",
			CourtName:  "Louisiana Vital Records Registry (Orleans Parish)",
			CourtURL:   "https://ldh.la.gov/page/marriage",
			SearchURL:  "https://vitalrecords.sos.la.gov/Marriages/DataView.aspx",
			Note:       "State Archives marriage search (Orleans 1870–1975); <50 yrs via Vital Records; enumeration candidate.",
		},
		{
			// Jefferson Parish — Clerk of Court marriage license page;
			// historical indices via JeffNet remote access + LCRAA.
			CountyFIPS: "22051",
			CourtName:  "Jefferson Parish Clerk of Court",
			CourtURL:   "https://www.jpclerkofcourt.us",
			SearchURL:  "https://www.jpclerkofcourt.us/marriage-licenses-passports/marriage-licenses/",
			Note:       "Marriage license indices 1840s–present; images pre-1951 via JeffNet; post-1950 index-only.",
		},
		{
			// East Baton Rouge Parish — Clerk of Court online access via
			// ClerkConnect (civil, family, probate, criminal, property).
			CountyFIPS: "22033",
			CourtName:  "East Baton Rouge Parish Clerk of Court",
			CourtURL:   "https://www.ebrclerk.com",
			SearchURL:  "https://www.ebrclerk.com/online-access",
			Note:       "ClerkConnect portal; marriage records from 1980; subscription-based for full access.",
		},
		{
			// Caddo Parish (Shreveport) — Clerk of Court marriage page with
			// online marriage index dating back to 1838.
			CountyFIPS: "22017",
			CourtName:  "Caddo Parish Clerk of Court",
			CourtURL:   "https://www.caddoclerk.com",
			SearchURL:  "https://www.caddoclerk.com/marriage.htm",
			Note:       "Marriage records back to 1838; online name index; certified copies via e-commerce.",
		},
		{
			// St. Tammany Parish — Clerk of Court online service with
			// marriage license data and indexes from 1923 to present.
			CountyFIPS: "22103",
			CourtName:  "St. Tammany Parish Clerk of Court",
			CourtURL:   "https://www.sttammanyclerk.org",
			SearchURL:  "https://ssl.sttammanyclerk.org/securenew/login.asp",
			Note:       "Online marriage license data 1923–present; archival records 1810–1969 via Archives Dept.",
		},
		{
			// Lafayette Parish — Clerk of Court; free index search via the
			// statewide eClerksLA portal (land, civil, marriage, probate).
			CountyFIPS: "22055",
			CourtName:  "Lafayette Parish Clerk of Court",
			CourtURL:   "https://www.lpclerk.com",
			SearchURL:  "https://eclerksla.com/home",
			Note:       "eClerksLA free index search for marriage licenses; subscription for images via ClerkConnect.",
		},
		{
			// Calcasieu Parish (Lake Charles) — Clerk of Court online search
			// services with marriage record category.
			CountyFIPS: "22019",
			CourtName:  "Calcasieu Parish Clerk of Court",
			CourtURL:   "https://www.calcasieuclerk.gov",
			SearchURL:  "https://www.calcasieuclerk.gov/online-search-services",
			Note:       "Online records search with marriage category; marriage licenses 1910–present.",
		},
	},

	// --- Church (Catholic dioceses + parishes) ---------------------------
	Dioceses: []DioceseDef{
		{Slug: "new_orleans", Name: "Archdiocese of New Orleans", Type: "archdiocese",
			Website: "https://nolacatholic.org", Directory: "https://nolacatholic.org/parishfinder", HubCityID: "city_new_orleans_la"},
		{Slug: "baton_rouge", Name: "Diocese of Baton Rouge", Type: "diocese",
			Website: "https://www.diobr.org", Directory: "https://www.diobr.org/parishes", HubCityID: "city_baton_rouge_la"},
		{Slug: "lafayette_la", Name: "Diocese of Lafayette in Louisiana", Type: "diocese",
			Website: "https://www.diolaf.org", Directory: "https://www.diolaf.org/parishes"},
		{Slug: "alexandria", Name: "Diocese of Alexandria", Type: "diocese",
			Website: "https://www.diocesealex.org", Directory: "https://www.diocesealex.org/parishes"},
		{Slug: "houma_thibodaux", Name: "Diocese of Houma-Thibodaux", Type: "diocese",
			Website: "https://www.htdiocese.org", Directory: "https://www.htdiocese.org/parishes"},
		{Slug: "lake_charles", Name: "Diocese of Lake Charles", Type: "diocese",
			Website: "https://www.dioceseoflakecharles.org", Directory: "https://www.dioceseoflakecharles.org/parishes"},
		{Slug: "shreveport", Name: "Diocese of Shreveport", Type: "diocese",
			Website: "https://www.dioshpt.org", Directory: "https://www.dioshpt.org/parishes"},
	},

	// New Orleans-area parishes in the Archdiocese of New Orleans. Names
	// and addresses verified from the archdiocese's parish finder
	// (nolacatholic.org/parishfinder) and each parish's own website.
	// Bulletin URLs verified by direct search for each parish's bulletin
	// archive on its own site.
	Parishes: []ParishDef{
		{DioceseSlug: "new_orleans", Name: "St. Louis Cathedral (Cathedral-Basilica of St. Louis King of France)",
			Address: "615 Pere Antoine Alley, New Orleans, LA 70116"},
		{DioceseSlug: "new_orleans", Name: "Holy Name of Jesus Parish",
			Address: "6367 St. Charles Ave, New Orleans, LA 70118"},
		{DioceseSlug: "new_orleans", Name: "St. Dominic Parish",
			Address: "775 Harrison Ave, New Orleans, LA 70124"},
		{
			DioceseSlug: "new_orleans", Name: "Mater Dolorosa",
			Address:     "1230 S. Carrollton Ave, New Orleans, LA 70118",
			BulletinURL: "https://mdolorosa.com/bulletin",
		},
		{DioceseSlug: "new_orleans", Name: "St. Francis of Assisi",
			Address: "631 State Street, New Orleans, LA 70118"},
		{
			DioceseSlug: "new_orleans", Name: "Immaculate Conception Jesuit Church",
			Address:     "130 Baronne Street, New Orleans, LA 70112",
			BulletinURL: "https://jesuitchurch.net/bulletins-1",
		},
		{DioceseSlug: "new_orleans", Name: "St. Augustine",
			Address: "1210 Governor Nicholls Street, New Orleans, LA 70116"},
		{DioceseSlug: "new_orleans", Name: "St. Louis King of France",
			Address: "1609 Carrollton Ave, Metairie, LA 70005"},
	},

	// --- Social (wedding-industry Instagram vendors) ---------------------
	Vendors: []VendorDef{
		// New Orleans photographers
		{
			Name: "Dee Olmstead Photography", OfficialURL: "https://deeolmstead.com/",
			Handle: "deeolmstead_photography", SourceClass: "engagement_photographer",
			CityID: "city_new_orleans_la", State: "LA", City: "New Orleans", Verified: "2026-08-01",
			TikTokHandle: "deeolmstead_photography",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-photographers/dee-olmstead-photography-7049021",
		},
		{
			Name: "Olivia Yuen Photo", OfficialURL: "https://oliviayuenphoto.com/",
			Handle: "_oliviayuen", SourceClass: "engagement_photographer",
			CityID: "city_new_orleans_la", State: "LA", City: "New Orleans", Verified: "2026-08-01",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-photographers/olivia-yuen-photo-1846175",
		},
		{
			Name: "Alyse and Ben Photography", OfficialURL: "https://alyseandben.com/",
			Handle: "alyseandbenphotography", SourceClass: "engagement_photographer",
			CityID: "city_new_orleans_la", State: "LA", City: "New Orleans", Verified: "2026-08-01",
			TikTokHandle: "alyseandbenphotography",
		},
		{
			Name: "Linka Odom Photography", OfficialURL: "https://linkaodom.com/",
			Handle: "linkaodomphotography", SourceClass: "engagement_photographer",
			CityID: "city_new_orleans_la", State: "LA", City: "New Orleans", Verified: "2026-08-01",
		},
		// New Orleans venues
		{
			Name: "Baroness at 512 Conti", OfficialURL: "https://www.baronessfq.com/",
			Handle: "baronessfqnola", SourceClass: "wedding_venue",
			CityID: "city_new_orleans_la", State: "LA", City: "New Orleans", Verified: "2026-08-01",
			TikTokHandle: "baronessfqnola",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-venues/baroness-at-512-conti-6101200",
		},
		{
			Name: "The Chicory", OfficialURL: "https://chicoryvenue.com/",
			Handle: "thechicory", SourceClass: "wedding_venue",
			CityID: "city_new_orleans_la", State: "LA", City: "New Orleans", Verified: "2026-08-01",
		},
		{
			Name: "Soulet Muse", OfficialURL: "https://www.souletmuse.com/",
			Handle: "souletmuse", SourceClass: "wedding_venue",
			CityID: "city_new_orleans_la", State: "LA", City: "New Orleans", Verified: "2026-08-01",
			TikTokHandle: "souletmuse",
		},
		// New Orleans jewelers
		{
			Name: "Valobra Jewelers", OfficialURL: "https://www.valobra.net/",
			Handle: "valobramasterjewelers", SourceClass: "jeweler",
			CityID: "city_new_orleans_la", State: "LA", City: "New Orleans", Verified: "2026-08-01",
		},
		{
			Name: "Friend & Company Fine Jewelers", OfficialURL: "https://friendandcompany.com/",
			Handle: "friendandcompany", SourceClass: "jeweler",
			CityID: "city_new_orleans_la", State: "LA", City: "New Orleans", Verified: "2026-08-01",
		},
		{
			Name: "Marc Pagani Photography", OfficialURL: "https://www.paganiphoto.com/",
			Handle: "marcpagani", SourceClass: "engagement_photographer",
			CityID: "city_new_orleans_la", State: "LA", City: "New Orleans", Verified: "2026-08-03",
		},
		{
			Name: "Sarah Mattix Photography", OfficialURL: "https://www.neworleansweddingphotography.com/",
			Handle: "sarahmattixphotography", SourceClass: "engagement_photographer",
			CityID: "city_new_orleans_la", State: "LA", City: "New Orleans", Verified: "2026-08-03",
		},
		{
			Name: "Cheryl Cole Photography", OfficialURL: "https://cherylcolephotography.com/",
			Handle: "cherylcolephotography", SourceClass: "engagement_photographer",
			CityID: "city_new_orleans_la", State: "LA", City: "New Orleans", Verified: "2026-08-03",
		},
		{
			Name: "The Crawfords Photography", OfficialURL: "https://wearethecrawfords.com/",
			Handle: "wearethecrawfords", SourceClass: "engagement_photographer",
			CityID: "city_new_orleans_la", State: "LA", City: "New Orleans", Verified: "2026-08-03",
		},
		{
			Name: "The Cannery", OfficialURL: "https://cannerynola.com/",
			Handle: "cannerynola", SourceClass: "wedding_venue",
			CityID: "city_new_orleans_la", State: "LA", City: "New Orleans", Verified: "2026-08-03",
		},
		{
			Name: "NOPSI Hotel", OfficialURL: "https://www.nopsihotel.com/weddings",
			Handle: "nopsihotel", SourceClass: "wedding_venue",
			CityID: "city_new_orleans_la", State: "LA", City: "New Orleans", Verified: "2026-08-03",
		},
		{
			Name: "Baroness", OfficialURL: "https://www.baronessfq.com/",
			Handle: "baronessfq", SourceClass: "wedding_venue",
			CityID: "city_new_orleans_la", State: "LA", City: "New Orleans", Verified: "2026-08-03",
		},
	},
}
