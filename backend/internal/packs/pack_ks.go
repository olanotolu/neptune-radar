package packs

// Kansas source pack — verified 2026-08-01.
//
// Government: Kansas marriage records are held by the district court in each
// county (not the county clerk). All 7 top counties use the statewide Kansas
// CaseSearch portal (https://casesearch.kscourts.gov) for online record
// searches; each county's own district court page is the CourtURL.
//
// Church: all 4 Kansas Catholic dioceses/archdioceses verified via USCCB +
// each diocese's own website. Kansas City KS-area parishes (Archdiocese of
// Kansas City in Kansas) were verified against the archkck.org parish finder
// + direct bulletin-archive URL discovery.
//
// Social: Instagram handles verified from each business's own public website
// social links. Verification date recorded per vendor.

var ksPack = StatePack{
	State: "KS",

	// --- Cities ----------------------------------------------------------
	Cities: []CityDef{
		{ID: "city_wichita_ks", State: "KS", County: "20173", Name: "Wichita",
			Lat: 37.6872, Lng: -97.3301, Markets: []string{"wichita", "ict", "ks", "sedgwick"}},
		{ID: "city_kansas_city_ks", State: "KS", County: "20209", Name: "Kansas City",
			Lat: 39.1147, Lng: -94.6270, Markets: []string{"kansascity", "kck", "wyandotte", "ks"}},
	},

	// --- Government (district court marriage-record searches) -------------
	// Kansas marriage records are filed with the district court in the county
	// where the license was issued. All counties below participate in the
	// statewide Kansas CaseSearch portal for online public access.
	Government: []GovSource{
		{
			// Sedgwick County (Wichita) — 18th Judicial District Court.
			CountyFIPS: "20173",
			CourtName:  "Sedgwick County District Court (18th JD)",
			CourtURL:   "https://www.dc18.org/records",
			SearchURL:  "https://casesearch.kscourts.gov",
			Note:       "Marriage records held by 18th Judicial District Court; searchable via statewide Kansas CaseSearch portal by party name.",
		},
		{
			// Johnson County — Johnson County District Court.
			CountyFIPS: "20091",
			CourtName:  "Johnson County District Court",
			CourtURL:   "https://courts.jocogov.org/dc_accessrecs.aspx",
			SearchURL:  "https://casesearch.kscourts.gov",
			Note:       "Marriage records held by Johnson County District Court; searchable via statewide Kansas CaseSearch portal by party name.",
		},
		{
			// Wyandotte County (Kansas City) — 29th Judicial District Court.
			CountyFIPS: "20209",
			CourtName:  "Wyandotte County District Court (29th JD)",
			CourtURL:   "https://www.wycodistrictcourt.org/marriage-license",
			SearchURL:  "https://casesearch.kscourts.gov",
			Note:       "Marriage records held by 29th Judicial District Court; searchable via statewide Kansas CaseSearch portal by party name.",
		},
		{
			// Shawnee County (Topeka) — Third Judicial District.
			CountyFIPS: "20177",
			CourtName:  "Shawnee County District Court (3rd JD)",
			CourtURL:   "https://ks-shawneecountycourts.civicplus.com/355/Marriage-Licenses",
			SearchURL:  "https://casesearch.kscourts.gov",
			Note:       "Marriage records held by Shawnee County District Court; searchable via statewide Kansas CaseSearch portal by party name.",
		},
		{
			// Douglas County (Lawrence) — Douglas County District Court.
			CountyFIPS: "20045",
			CourtName:  "Douglas County District Court",
			CourtURL:   "https://www.dgcoks.gov/district-court/clerk-of-the-district-court/marriage-divorce-and-protection-orders",
			SearchURL:  "https://casesearch.kscourts.gov",
			Note:       "Marriage records held by Douglas County District Court; searchable via statewide Kansas CaseSearch portal by party name.",
		},
		{
			// Leavenworth County — Leavenworth County District Court.
			CountyFIPS: "20103",
			CourtName:  "Leavenworth County District Court",
			CourtURL:   "https://files.leavenworthcounty.gov/departments/district_court/marriage_licenses.php",
			SearchURL:  "https://casesearch.kscourts.gov",
			Note:       "Marriage records held by Leavenworth County District Court; searchable via statewide Kansas CaseSearch portal by party name.",
		},
		{
			// Riley County (Manhattan) — Riley County District Court.
			CountyFIPS: "20161",
			CourtName:  "Riley County District Court",
			CourtURL:   "https://www.rileycountyks.gov/1621/Marriage-License",
			SearchURL:  "https://casesearch.kscourts.gov",
			Note:       "Marriage records held by Riley County District Court; searchable via statewide Kansas CaseSearch portal by party name.",
		},
	},

	// --- Church (Catholic dioceses + parishes) ---------------------------
	Dioceses: []DioceseDef{
		{Slug: "kc_ks", Name: "Archdiocese of Kansas City in Kansas", Type: "archdiocese",
			Website: "https://www.archkck.org", Directory: "https://archkck.org/parish-school/catholic-parishes/find-your-parish/", HubCityID: "city_kansas_city_ks"},
		{Slug: "wichita", Name: "Diocese of Wichita", Type: "diocese",
			Website: "https://www.dow-ks.org", Directory: "https://www.dow-ks.org/parishes", HubCityID: "city_wichita_ks"},
		{Slug: "dodge_city", Name: "Diocese of Dodge City", Type: "diocese",
			Website: "https://www.dcdiocese.org", Directory: "https://www.dcdiocese.org/parishes"},
		{Slug: "salina", Name: "Diocese of Salina", Type: "diocese",
			Website: "https://salinadiocese.org", Directory: "https://salinadiocese.org/parishes"},
	},

	// Kansas City KS-area parishes in the Archdiocese of Kansas City in
	// Kansas. Names and addresses verified against the archkck.org parish
	// finder. Bulletin URLs verified by direct visit to each parish's own
	// website; Aggregator=true where the bulletin is hosted on a third-party
	// platform (Parishes Online) rather than the parish's own site.
	Parishes: []ParishDef{
		{
			DioceseSlug: "kc_ks", Name: "Cathedral of St. Peter the Apostle",
			Address: "409 N. 15th St, Kansas City, KS 66102",
		},
		{
			DioceseSlug: "kc_ks", Name: "St. John the Baptist Catholic Church",
			Address:     "708 N. 4th St, Kansas City, KS 66101",
			BulletinURL: "http://www.stjohnthebaptistcatholicchurch.com/bulletins/",
		},
		{
			DioceseSlug: "kc_ks", Name: "Holy Family Church",
			Address:     "274 Orchard St, Kansas City, KS 66101",
			BulletinURL: "https://holyfamilychurchkck.org/bulletins",
		},
		{
			DioceseSlug: "kc_ks", Name: "St. Patrick Catholic Church",
			Address:     "1086 N. 94th St, Kansas City, KS 66112",
			BulletinURL: "https://stpatrickkck.org/bulletins",
		},
		{
			DioceseSlug: "kc_ks", Name: "St. Rose Philippine Duchesne Parish",
			Address:     "5035 Rainbow Blvd, Westwood, KS 66205",
			BulletinURL: "https://spdlatinmass.com/bulletins/",
		},
		{
			DioceseSlug: "kc_ks", Name: "St. Agnes Catholic Parish",
			Address:     "5250 Mission Rd, Roeland Park, KS 66205",
			BulletinURL: "https://parishesonline.com/find/st-agnes-church-66205",
			Aggregator:  true,
		},
		{
			DioceseSlug: "kc_ks", Name: "Church of the Nativity",
			Address: "3800 W. 119th St, Leawood, KS 66209",
		},
		{
			DioceseSlug: "kc_ks", Name: "Queen of the Holy Rosary Catholic Church",
			Address:     "7023 W. 71st St, Overland Park, KS 66204",
			BulletinURL: "https://www.queenoftheholyrosary.org/bulletin",
		},
	},

	// --- Social (wedding-industry Instagram vendors) ---------------------
	Vendors: []VendorDef{
		// Wichita photographers
		{
			Name: "Alex Bo Photo", OfficialURL: "https://alexbophoto.co/",
			Handle: "alexbo.photo", SourceClass: "engagement_photographer",
			CityID: "city_wichita_ks", State: "KS", City: "Wichita", Verified: "2026-08-01",
			TikTokHandle: "alexbo.photo",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-photographers/alex-bo-photo-7937946",
		},
		{
			Name: "Ashley x Ashley", OfficialURL: "https://ashleyxashley.com/",
			Handle: "itsashleyxashley", SourceClass: "engagement_photographer",
			CityID: "city_wichita_ks", State: "KS", City: "Wichita", Verified: "2026-08-01",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-photographers/ashley-x-ashley-7870509",
		},
		{
			Name: "Lyndi Mishe Photography", OfficialURL: "https://lyndimishephotography.com/",
			Handle: "lyndimishephoto", SourceClass: "engagement_photographer",
			CityID: "city_wichita_ks", State: "KS", City: "Wichita", Verified: "2026-08-01",
			TikTokHandle: "lyndimishephoto",
		},
		// Wichita venues
		{
			Name: "Prairie Hill Vineyard", OfficialURL: "https://prairiehillvineyard.com/",
			Handle: "prairiehillvineyard", SourceClass: "wedding_venue",
			CityID: "city_wichita_ks", State: "KS", City: "Wichita", Verified: "2026-08-01",
		},
		{
			Name: "Mark Arts", OfficialURL: "https://markartsks.com/",
			Handle: "markartsks", SourceClass: "wedding_venue",
			CityID: "city_wichita_ks", State: "KS", City: "Wichita", Verified: "2026-08-01",
			TikTokHandle: "markartsks",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-venues/mark-arts-1073371",
		},
		// Wichita jewelers
		{
			Name: "Mike Seltzer Jewelers", OfficialURL: "https://mikeseltzerjewelers.com/",
			Handle: "mikeseltzerjewelers", SourceClass: "jeweler",
			CityID: "city_wichita_ks", State: "KS", City: "Wichita", Verified: "2026-08-01",
		},
		{
			Name: "Burnell's Fine Jewelry", OfficialURL: "https://burnells.com/",
			Handle: "burnellsjewelry", SourceClass: "jeweler",
			CityID: "city_wichita_ks", State: "KS", City: "Wichita", Verified: "2026-08-01",
		},
		// Kansas City photographers
		{
			Name: "Mariam Saifan Photography", OfficialURL: "https://www.mariamsaifan.com/",
			Handle: "mariamsaifan", SourceClass: "engagement_photographer",
			CityID: "city_kansas_city_ks", State: "KS", City: "Kansas City", Verified: "2026-08-01",
		},
		// Kansas City venues
		{
			Name: "The Mint", OfficialURL: "https://www.kcmint.com/",
			Handle: "thekcmint", SourceClass: "wedding_venue",
			CityID: "city_kansas_city_ks", State: "KS", City: "Kansas City", Verified: "2026-08-01",
			TikTokHandle: "thekcmint",
		},
	},
}
