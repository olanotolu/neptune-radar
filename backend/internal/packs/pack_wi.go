package packs

// Wisconsin source pack — verified 2026-08-01.
//
// Government: Wisconsin marriage records are held by the county Register of
// Deeds (the County Clerk issues licenses but does not retain them). Search
// URLs for the top 7 counties by population were verified against each
// county's official .gov site. Most Wisconsin counties do not expose an
// online searchable marriage database — records are obtained via application
// form (in person, by mail, or via VitalChek). Racine County's LandShark
// portal is the closest to a true online search. Notes state this honestly.
//
// Church: all 5 Wisconsin Catholic dioceses/archdioceses verified via USCCB +
// each diocese's own website. Milwaukee-area parishes (Archdiocese of
// Milwaukee) were verified against the archdiocese's own parish map
// (archmil.org/maps/parishes) + direct bulletin-archive URL discovery.
// Parishes Online / Discover Mass aggregator URLs are flagged with
// Aggregator=true.
//
// Social: Instagram handles verified from each business's own public website
// social links. Verification date recorded per vendor.

var wiPack = StatePack{
	State: "WI",

	// --- Cities ----------------------------------------------------------
	Cities: []CityDef{
		{
			ID: "city_milwaukee_wi", State: "WI", County: "55079", Name: "Milwaukee",
			Lat: 43.0389, Lng: -87.9065,
			Markets: []string{"milwaukee", "mke", "wisconsin", "milwaukeecounty"},
		},
		{
			ID: "city_madison_wi", State: "WI", County: "55025", Name: "Madison",
			Lat: 43.0731, Lng: -89.4012,
			Markets: []string{"madison", "mad", "danecounty"},
		},
	},

	// --- Government (county register of deeds marriage-record searches) ---
	Government: []GovSource{
		{
			// Milwaukee County — Register of Deeds vital records page.
			// Marriage certificates from the 1830s to present. No online
			// search portal; records obtained via application form or
			// VitalChek. County Clerk handles license issuance only.
			CountyFIPS: "55079",
			CourtName:  "Milwaukee County Register of Deeds",
			CourtURL:   "https://county.milwaukee.gov/EN/Register-of-Deeds/Vital-Records",
			SearchURL:  "https://county.milwaukee.gov/EN/County-Clerk/Marriage-License",
			Note:       "Vital records page with marriage certificate application; no online search, request-oriented. Clerk page handles license issuance.",
		},
		{
			// Dane County (Madison) — Register of Deeds vital records.
			// Marriage certificates from Oct 1907 to present (statewide
			// issuance). Genealogy search by appointment only.
			CountyFIPS: "55025",
			CourtName:  "Dane County Register of Deeds",
			CourtURL:   "https://rod.danecounty.gov/vital-records",
			SearchURL:  "https://rod.danecounty.gov/vital-records/genealogy",
			Note:       "Vital records + genealogy search page; in-person search by appointment, otherwise application-based. No online enumeration.",
		},
		{
			// Waukesha County — Register of Deeds vital records.
			// Marriage certificates from Oct 1907 to present. Genealogy
			// research by appointment.
			CountyFIPS: "55133",
			CourtName:  "Waukesha County Register of Deeds",
			CourtURL:   "https://www.waukeshacounty.gov/register-of-deeds/vital-records/",
			SearchURL:  "https://www.waukeshacounty.gov/register-of-deeds/genealogy-research/",
			Note:       "Vital records + genealogy research page; appointment-based search, no online portal.",
		},
		{
			// Brown County (Green Bay) — Register of Deeds vital records.
			// Marriage certificates available; genealogy by appointment.
			CountyFIPS: "55009",
			CourtName:  "Brown County Register of Deeds",
			CourtURL:   "https://www.browncountywi.gov/departments/register-of-deeds/vital-records/services/",
			SearchURL:  "https://www.browncountywi.gov/services/genealogy/?department=cd96ad5fd32c",
			Note:       "Vital records services + genealogy search page; appointment-based, no online portal.",
		},
		{
			// Racine County — Register of Deeds vital records + LandShark
			// online records search (the closest to a true online search
			// among Wisconsin counties).
			CountyFIPS: "55101",
			CourtName:  "Racine County Register of Deeds",
			CourtURL:   "https://www.racinecounty.gov/our-county/visiting/birth-marriage-death-divorce-certificates",
			SearchURL:  "https://www.racinecounty.gov/departments/register-of-deeds/records-certificates/online-records-search-landshark",
			Note:       "LandShark online records search includes vital records; enumeration capability needs testing. Marriage certificates from Oct 1907.",
		},
		{
			// Outagamie County (Appleton) — Register of Deeds vital
			// records. Marriage certificates from Oct 1907 to present.
			CountyFIPS: "55087",
			CourtName:  "Outagamie County Register of Deeds",
			CourtURL:   "https://www.outagamie.gov/County-Services/Register-of-Deeds/Vital-Records",
			SearchURL:  "https://www.outagamie.gov/Our-County/County-Clerk/Marriage-Licenses",
			Note:       "Vital records page with marriage certificate info; no online search, application-based. Clerk page handles license issuance.",
		},
		{
			// Kenosha County — Register of Deeds vital records.
			// Marriage certificates from Oct 1907 to present.
			CountyFIPS: "55059",
			CourtName:  "Kenosha County Register of Deeds",
			CourtURL:   "https://www.kenoshacountywi.gov/547/Vital-Records",
			SearchURL:  "https://www.kenoshacountywi.gov/140/Marriage-License",
			Note:       "Vital records page with marriage certificate application; no online search, request-oriented. Clerk page handles license issuance.",
		},
	},

	// --- Church (Catholic dioceses + parishes) ---------------------------
	Dioceses: []DioceseDef{
		{
			Slug: "milwaukee", Name: "Archdiocese of Milwaukee", Type: "archdiocese",
			Website: "https://www.archmil.org", Directory: "https://www.archmil.org/maps/parishes",
			HubCityID: "city_milwaukee_wi",
		},
		{
			Slug: "green_bay", Name: "Diocese of Green Bay", Type: "diocese",
			Website: "https://www.gbdioc.org", Directory: "https://www.gbdioc.org/parishes",
		},
		{
			Slug: "madison", Name: "Diocese of Madison", Type: "diocese",
			Website: "https://www.madisondiocese.org", Directory: "https://www.madisondiocese.org/parishes",
			HubCityID: "city_madison_wi",
		},
		{
			Slug: "la_crosse", Name: "Diocese of La Crosse", Type: "diocese",
			Website: "https://www.diolc.org", Directory: "https://www.diolc.org/parishes",
		},
		{
			Slug: "superior", Name: "Diocese of Superior", Type: "diocese",
			Website: "https://www.catholicdioceseofsuperior.org", Directory: "https://www.catholicdioceseofsuperior.org/parishes",
		},
	},

	// Milwaukee-area parishes in the Archdiocese of Milwaukee. Names and
	// addresses verified from the archdiocese's own parish map
	// (archmil.org/maps/parishes) and each parish's own website. Bulletin
	// URLs verified by direct search for each parish's bulletin archive.
	// Parishes Online aggregator URLs are flagged with Aggregator=true.
	Parishes: []ParishDef{
		{
			DioceseSlug: "milwaukee", Name: "Cathedral of St. John the Evangelist",
			Address:     "812 N Jackson St, Milwaukee, WI 53202",
			BulletinURL: "https://www.stjohncathedral.org/index.php/bulletin/",
		},
		{
			DioceseSlug: "milwaukee", Name: "Basilica of St. Josaphat",
			Address:     "2333 S 6th St, Milwaukee, WI 53215",
			BulletinURL: "https://parishesonline.com/publication-page/basilica-of-saint-josaphat",
			Aggregator:  true,
		},
		{
			DioceseSlug: "milwaukee", Name: "Gesu Parish",
			Address:     "1145 W Wisconsin Ave, Milwaukee, WI 53233",
			BulletinURL: "https://gesuparish.org/bulletin",
		},
		{
			DioceseSlug: "milwaukee", Name: "Three Holy Women Catholic Parish",
			Address:     "1716 N Humboldt Ave, Milwaukee, WI 53202",
			BulletinURL: "https://www.threeholywomenparish.org/bulletin/",
		},
		{
			DioceseSlug: "milwaukee", Name: "St. Matthias Catholic Parish",
			Address:     "9306 W Beloit Rd, Milwaukee, WI 53227",
			BulletinURL: "https://www.stmatthias-milw.org/bulletin-1",
		},
		{
			DioceseSlug: "milwaukee", Name: "St. Stanislaus Parish and Oratory",
			Address:     "524 W Historic Mitchell St, Milwaukee, WI 53204",
			BulletinURL: "https://institute-christ-king.org/bulletins-milwaukee",
		},
		{
			DioceseSlug: "milwaukee", Name: "St. Sebastian Catholic Church",
			Address:     "5400 Washington Blvd, Milwaukee, WI 53208",
			BulletinURL: "https://parishesonline.com/find/st-sebastian-church-53208",
			Aggregator:  true,
		},
		{
			DioceseSlug: "milwaukee", Name: "Our Lady Queen of Peace",
			Address: "3222 S 29th St, Milwaukee, WI 53215",
		},
	},

	// --- Social (wedding-industry Instagram vendors) ---------------------
	Vendors: []VendorDef{
		// Milwaukee photographers
		{
			Name: "Bailey Bryn Photography", OfficialURL: "https://baileybrynphotography.com/",
			Handle: "baileybrynphotography", SourceClass: "engagement_photographer",
			CityID: "city_milwaukee_wi", State: "WI", City: "Milwaukee", Verified: "2026-08-01",
			TikTokHandle: "baileybrynphotography",
		},
		{
			Name: "Roost Photography", OfficialURL: "https://www.roostmke.com/",
			Handle: "roostmke", SourceClass: "engagement_photographer",
			CityID: "city_milwaukee_wi", State: "WI", City: "Milwaukee", Verified: "2026-08-01",
			TikTokHandle: "roostmke",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-photographers/roost-photography-8405646",
		},
		// Milwaukee venues
		{
			Name: "The Gage", OfficialURL: "https://www.thegagemke.com/",
			Handle: "thegagemke", SourceClass: "wedding_venue",
			CityID: "city_milwaukee_wi", State: "WI", City: "Milwaukee", Verified: "2026-08-01",
			TikTokHandle: "thegagemke",
		},
		// Milwaukee jeweler
		{
			Name: "Schwanke-Kasten Jewelers", OfficialURL: "https://schwanke-kasten.com/",
			Handle: "schwankekasten", SourceClass: "jeweler",
			CityID: "city_milwaukee_wi", State: "WI", City: "Milwaukee", Verified: "2026-08-01",
		},
		// Madison photographers
		{
			Name: "Sara Baillies Photography", OfficialURL: "https://www.sarabaillies.com/",
			Handle: "sarabaillies", SourceClass: "engagement_photographer",
			CityID: "city_madison_wi", State: "WI", City: "Madison", Verified: "2026-08-01",
		},
		// Madison venues
		{
			Name: "The Eloise", OfficialURL: "https://theeloiseevents.com/",
			Handle: "theeloiseevents", SourceClass: "wedding_venue",
			CityID: "city_madison_wi", State: "WI", City: "Madison", Verified: "2026-08-01",
			TikTokHandle: "theeloiseevents",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-venues/the-eloise-1365164",
		},
		{
			Name: "The Fields Reserve", OfficialURL: "https://www.fieldsreserve.com/",
			Handle: "fieldsreserve", SourceClass: "wedding_venue",
			CityID: "city_madison_wi", State: "WI", City: "Madison", Verified: "2026-08-01",
		},
		// Madison jewelers
		{
			Name: "Jewelers Workshop", OfficialURL: "https://www.jewelersworkshop.com/",
			Handle: "jewelersworkshop", SourceClass: "jeweler",
			CityID: "city_madison_wi", State: "WI", City: "Madison", Verified: "2026-08-01",
		},
		{
			Name: "BR Diamond Suite", OfficialURL: "https://www.brdiamondsuite.com/",
			Handle: "brdiamondsuite", SourceClass: "jeweler",
			CityID: "city_madison_wi", State: "WI", City: "Madison", Verified: "2026-08-01",
		},
		{
			Name: "McKenna Marie Photography", OfficialURL: "https://www.mckennamariephoto.com/",
			Handle: "mckennamariephoto", SourceClass: "engagement_photographer",
			CityID: "city_milwaukee_wi", State: "WI", City: "Milwaukee", Verified: "2026-08-03",
		},
		{
			Name: "St James 1868", OfficialURL: "https://stjames1868.com/",
			Handle: "stjames1868", SourceClass: "wedding_venue",
			CityID: "city_milwaukee_wi", State: "WI", City: "Milwaukee", Verified: "2026-08-03",
		},
		{
			Name: "Tim Fitch Photography", OfficialURL: "https://www.timfitchweddings.com/",
			Handle: "timfitchphotography", SourceClass: "engagement_photographer",
			CityID: "city_madison_wi", State: "WI", City: "Madison", Verified: "2026-08-03",
		},
		{
			Name: "James Stokes & Co.", OfficialURL: "https://www.james-stokes.com/",
			Handle: "jamesstokesphoto", SourceClass: "engagement_photographer",
			CityID: "city_madison_wi", State: "WI", City: "Madison", Verified: "2026-08-03",
		},
		{
			Name: "Monona Terrace", OfficialURL: "https://www.mononaterrace.com/weddings/",
			Handle: "mononaterrace", SourceClass: "wedding_venue",
			CityID: "city_madison_wi", State: "WI", City: "Madison", Verified: "2026-08-03",
		},
		{
			Name: "The Tinsmith", OfficialURL: "https://www.thetinsmith.com/",
			Handle: "tinsmithevents", SourceClass: "wedding_venue",
			CityID: "city_madison_wi", State: "WI", City: "Madison", Verified: "2026-08-03",
		},
	},
}
