package packs

// Minnesota source pack — verified 2026-08-01.
//
// Government: Minnesota marriage records are held by the county in which the
// license was issued. The Minnesota Official Marriage System (MOMS) at
// moms.mn.gov/Search is a statewide search portal covering most counties.
// Search URLs for the top 7 counties by population were verified against each
// county's official .gov/.us site.
//
// Church: all 6 Minnesota Catholic dioceses/archdiocese verified via USCCB +
// each diocese's own website. Parish directory URL points at the Archdiocese
// of Saint Paul and Minneapolis's own parish-finder. Twin Cities-area
// parishes were verified against the archdiocese's parish listings + direct
// bulletin-archive URL discovery.
//
// Social: Instagram handles verified from each business's own public website
// social links (or from IG search results where the site is JS-rendered and
// the handle was visible in the search snippet). Verification date recorded
// per vendor.

var mnPack = StatePack{
	State: "MN",

	// --- Cities ----------------------------------------------------------
	Cities: []CityDef{
		{
			ID: "city_minneapolis_mn", State: "MN", County: "27053", Name: "Minneapolis",
			Lat: 44.9778, Lng: -93.2650,
			Markets: []string{"minneapolis", "msp", "hennepin", "twincities"},
		},
		{
			ID: "city_st_paul_mn", State: "MN", County: "27123", Name: "St. Paul",
			Lat: 44.9537, Lng: -93.0900,
			Markets: []string{"stpaul", "ramsey", "twincities"},
		},
	},

	// --- Government (county marriage-record searches) ---------------------
	Government: []GovSource{
		{
			// Hennepin County (Minneapolis) — county vital records office;
			// statewide search via MOMS.
			CountyFIPS: "27053",
			CourtName:  "Hennepin County Vital Records",
			CourtURL:   "https://www.hennepincounty.gov/services/licenses-certificates/marriage",
			SearchURL:  "https://moms.mn.gov/Search",
			Note:       "MOMS statewide search covers Hennepin; county office issues certified copies; enumeration candidate via MOMS.",
		},
		{
			// Ramsey County (St. Paul) — county vital records office;
			// Ramsey is not fully in MOMS, so the county page is the primary source.
			CountyFIPS: "27123",
			CourtName:  "Ramsey County Vital Records",
			CourtURL:   "https://www.ramseycountymn.gov/residents/licenses-permits-records/marriage-records",
			SearchURL:  "https://www.ramseycountymn.gov/residents/licenses-permits-records/marriage-records",
			Note:       "Ramsey County marriage records from ~1850; not fully in MOMS; request-oriented, no online search index.",
		},
		{
			// Dakota County — county vital records office; MOMS covers Dakota.
			CountyFIPS: "27037",
			CourtName:  "Dakota County Vital Records",
			CourtURL:   "https://www.co.dakota.mn.us/Permits/MarriageLicensesCertificates/Certificates/Pages/default.aspx",
			SearchURL:  "https://moms.mn.gov/Search",
			Note:       "MOMS statewide search covers Dakota; county office issues certified copies; enumeration candidate via MOMS.",
		},
		{
			// Anoka County — county vital records office; MOMS covers Anoka.
			CountyFIPS: "27003",
			CourtName:  "Anoka County Vital Records",
			CourtURL:   "https://anokacountymn.gov/241/Marriage-Certificates",
			SearchURL:  "https://moms.mn.gov/Search",
			Note:       "MOMS statewide search covers Anoka; county office issues certified copies; enumeration candidate via MOMS.",
		},
		{
			// Washington County — county vital records office;
			// Washington does not participate in MOMS.
			CountyFIPS: "27163",
			CourtName:  "Washington County Vital Records",
			CourtURL:   "https://washingtoncountymn.gov/1231/Marriage-Certificate",
			SearchURL:  "https://washingtoncountymn.gov/1231/Marriage-Certificate",
			Note:       "Washington County does not participate in MOMS; records from 1870; request-oriented, no online search index.",
		},
		{
			// St. Louis County (Duluth) — county recorder office;
			// MOMS covers St. Louis County.
			CountyFIPS: "27137",
			CourtName:  "St. Louis County Recorder",
			CourtURL:   "https://www.stlouiscountymn.gov/departments-a-z/public-records/records/marriage-records",
			SearchURL:  "https://moms.mn.gov/Search",
			Note:       "MOMS statewide search covers St. Louis; county recorder issues certified copies; enumeration candidate via MOMS.",
		},
		{
			// Stearns County (St. Cloud) — county virtual license center;
			// MOMS covers Stearns County.
			CountyFIPS: "27145",
			CourtName:  "Stearns County License Center",
			CourtURL:   "https://www.co.stearns.mn.us/1426/Virtual-License-Center",
			SearchURL:  "https://moms.mn.gov/Search",
			Note:       "MOMS statewide search covers Stearns; county license center issues certified copies; enumeration candidate via MOMS.",
		},
	},

	// --- Church (Catholic dioceses + parishes) ----------------------------
	Dioceses: []DioceseDef{
		{
			Slug: "stpaul_minneapolis", Name: "Archdiocese of Saint Paul and Minneapolis",
			Type: "archdiocese", Website: "https://www.archspm.org",
			Directory: "https://www.archspm.org/parishes", HubCityID: "city_minneapolis_mn",
		},
		{
			Slug: "duluth", Name: "Diocese of Duluth", Type: "diocese",
			Website: "https://www.dioceseduluth.org", Directory: "https://www.dioceseduluth.org/parishes",
		},
		{
			Slug: "st_cloud", Name: "Diocese of Saint Cloud", Type: "diocese",
			Website: "https://stclouddiocese.org", Directory: "https://stclouddiocese.org/parishes",
		},
		{
			Slug: "new_ulm", Name: "Diocese of New Ulm", Type: "diocese",
			Website: "https://www.dnu.org", Directory: "https://www.dnu.org/parishes",
		},
		{
			Slug: "winona_rochester", Name: "Diocese of Winona-Rochester", Type: "diocese",
			Website: "https://www.dowr.org", Directory: "https://www.dowr.org/parishes",
		},
		{
			Slug: "crookston", Name: "Diocese of Crookston", Type: "diocese",
			Website: "https://www.crookston.org", Directory: "https://www.crookston.org/parishes",
		},
	},

	// Twin Cities-area parishes in the Archdiocese of Saint Paul and
	// Minneapolis. Names and addresses verified from the archdiocese's own
	// parish listings (archspm.org/location/…). Bulletin URLs verified by
	// direct search for each parish's bulletin archive.
	Parishes: []ParishDef{
		{
			DioceseSlug: "stpaul_minneapolis", Name: "Basilica of Saint Mary",
			Address:     "1600 Hennepin Ave, Minneapolis, MN 55403",
			BulletinURL: "https://mary.org/bulletins/",
		},
		{
			DioceseSlug: "stpaul_minneapolis", Name: "Cathedral of Saint Paul",
			Address:     "239 Selby Ave, Saint Paul, MN 55102",
			BulletinURL: "https://cathedralsaintpaul.org/bulletins",
		},
		{
			DioceseSlug: "stpaul_minneapolis", Name: "Our Lady of Lourdes Catholic Church",
			Address:     "1 Lourdes Pl, Minneapolis, MN 55414",
			BulletinURL: "https://lourdesmpls.org/sunday-bulletins-2026-year-a",
		},
		{
			DioceseSlug: "stpaul_minneapolis", Name: "Saint Olaf Catholic Church",
			Address:     "215 S 8th St, Minneapolis, MN 55402",
			BulletinURL: "https://www.saintolaf.org/weekly-bulletin/recent-bulletins",
		},
		{
			DioceseSlug: "stpaul_minneapolis", Name: "St. Joan of Arc Catholic Community",
			Address:     "4537 3rd Ave S, Minneapolis, MN 55419",
			BulletinURL: "https://www.saintjoanofarc.org/bull_LPI",
		},
		{
			DioceseSlug: "stpaul_minneapolis", Name: "Holy Name of Jesus Catholic Church",
			Address:     "3637 11th Ave S, Minneapolis, MN 55407",
			BulletinURL: "https://www.hnoj.org/about/bulletin/",
		},
		{
			DioceseSlug: "stpaul_minneapolis", Name: "St. Helena Catholic Church",
			Address: "3201 43rd St E, Minneapolis, MN 55406",
		},
		{
			DioceseSlug: "stpaul_minneapolis", Name: "St. Stephen–Holy Rosary Catholic Church",
			Address: "2211 Clinton Ave S, Minneapolis, MN 55404",
		},
	},

	// --- Social (wedding-industry Instagram vendors) ----------------------
	Vendors: []VendorDef{
		// Minneapolis photographers
		{
			Name: "JM Photography", OfficialURL: "https://www.jmphotomn.com/",
			Handle: "jmphotographymn", SourceClass: "engagement_photographer",
			CityID: "city_minneapolis_mn", State: "MN", City: "Minneapolis", Verified: "2026-08-01",
			TikTokHandle: "jmphotographymn",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-photographers/jm-photography-7353696",
		},
		{
			Name: "Alexandra Robyn Photo + Design", OfficialURL: "https://alexandrarobyn.com/",
			Handle: "alexandrarobynphoto", SourceClass: "engagement_photographer",
			CityID: "city_minneapolis_mn", State: "MN", City: "Minneapolis", Verified: "2026-08-01",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-photographers/alexandra-robyn-photo-design-4382736",
		},
		{
			Name: "Rohana Olson Photography", OfficialURL: "https://rohanaolson.com/",
			Handle: "rohanaolson", SourceClass: "engagement_photographer",
			CityID: "city_minneapolis_mn", State: "MN", City: "Minneapolis", Verified: "2026-08-01",
			TikTokHandle: "rohanaolson",
		},
		{
			Name: "Kadence Cruse Photo", OfficialURL: "https://kadencecrusephoto.com/",
			Handle: "kadencecrusephoto", SourceClass: "engagement_photographer",
			CityID: "city_minneapolis_mn", State: "MN", City: "Minneapolis", Verified: "2026-08-01",
		},
		// Minneapolis venues
		{
			Name: "Watson Block", OfficialURL: "https://www.watsonblock.com/",
			Handle: "watson_block", SourceClass: "wedding_venue",
			CityID: "city_minneapolis_mn", State: "MN", City: "Minneapolis", Verified: "2026-08-01",
			TikTokHandle: "watson_block",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-venues/watson-block-5767406",
		},
		{
			Name: "FIVE Event Center", OfficialURL: "https://www.fiveeventcenter.com/",
			Handle: "fiveeventcenter", SourceClass: "wedding_venue",
			CityID: "city_minneapolis_mn", State: "MN", City: "Minneapolis", Verified: "2026-08-01",
		},
		{
			Name: "Day Block Event Center", OfficialURL: "https://www.dayblock.com/",
			Handle: "dayblockeventcenter", SourceClass: "wedding_venue",
			CityID: "city_minneapolis_mn", State: "MN", City: "Minneapolis", Verified: "2026-08-01",
			TikTokHandle: "dayblockeventcenter",
		},
		{
			Name: "Semple Mansion", OfficialURL: "https://semplemansion.com/",
			Handle: "semplemansion", SourceClass: "wedding_venue",
			CityID: "city_minneapolis_mn", State: "MN", City: "Minneapolis", Verified: "2026-08-01",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-venues/semple-mansion-7829502",
		},
		{
			Name: "Leopold's Mississippi Gardens", OfficialURL: "https://www.leopoldsmn.com/",
			Handle: "leopoldsmississippigardens", SourceClass: "wedding_venue",
			CityID: "city_minneapolis_mn", State: "MN", City: "Minneapolis", Verified: "2026-08-01",
			TikTokHandle: "leopoldsmississippardens",
		},
		// Minneapolis jeweler
		{
			Name: "Gittelson Jewelers", OfficialURL: "https://gittelsonjewelers.com/",
			Handle: "gittelsonj", SourceClass: "jeweler",
			CityID: "city_minneapolis_mn", State: "MN", City: "Minneapolis", Verified: "2026-08-01",
		},
	},
}
