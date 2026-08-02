package packs

// Kentucky source pack — verified 2026-08-02.
//
// Government: Kentucky marriage records are held by the county clerk. Search
// URLs for the top 7 counties by population were verified against each
// county's official .ky.gov site, its eCCLIX portal, or its dedicated
// records-search site. Most KY counties use the eCCLIX (Software Management
// LLC) subscriber system; Kenton offers a free marriage-index search and
// Fayette uses a dedicated FayetteDeeds portal.
//
// Church: all 5 Kentucky Catholic dioceses/archdioceses verified via USCCB +
// each diocese's own website. Louisville-area parishes (Archdiocese of
// Louisville) were verified against the archdiocese's own parish directory
// (archlou.org/parishes) + direct bulletin-archive URL discovery on each
// parish's own website.
//
// Social: Instagram handles verified from each business's own public website
// social links. Verification date recorded per vendor.

var kyPack = StatePack{
	State: "KY",

	// --- Cities ----------------------------------------------------------
	Cities: []CityDef{
		{ID: "city_louisville_ky", State: "KY", County: "21111", Name: "Louisville",
			Lat: 38.2527, Lng: -85.7585, Markets: []string{"louisville", "lou", "jefferson", "ky"}},
		{ID: "city_lexington_ky", State: "KY", County: "21067", Name: "Lexington",
			Lat: 38.0406, Lng: -84.5037, Markets: []string{"lexington", "fayette", "ky"}},
	},

	// --- Government (county clerk marriage-record searches) --------------
	Government: []GovSource{
		{
			// Jefferson County (Louisville) — marriage license copy request
			// page; no online search portal, copies via mail/in-person.
			CountyFIPS: "21111",
			CourtName:  "Jefferson County Clerk",
			CourtURL:   "https://jeffersoncountyclerk.org",
			SearchURL:  "https://www.jeffersoncountyclerk.org/Legal-Records/Marriage-License-Copy",
			Note:       "Marriage license copy request page; no online index search, copies via mail or in-person.",
		},
		{
			// Fayette County (Lexington) — FayetteDeeds online portal with
			// marriage search (July 1989 to present) + pre-1989 records.
			CountyFIPS: "21067",
			CourtName:  "Fayette County Clerk",
			CourtURL:   "https://fayettekyclerk.gov",
			SearchURL:  "https://fayettedeeds.com",
			Note:       "FayetteDeeds portal; marriage search July 1989–present + pre-1989 index; enumeration candidate.",
		},
		{
			// Boone County — eCCLIX subscriber portal.
			CountyFIPS: "21015",
			CourtName:  "Boone County Clerk",
			CourtURL:   "https://boonecountyclerk.ky.gov",
			SearchURL:  "https://ecclix.com/ecclix/login.aspx",
			Note:       "eCCLIX subscriber login; marriage records searchable with paid account.",
		},
		{
			// Kenton County — free marriage record index search.
			CountyFIPS: "21117",
			CourtName:  "Kenton County Clerk",
			CourtURL:   "https://kentonkyclerk.gov",
			SearchURL:  "https://ccspublicsearchfree.kentoncounty.org/",
			Note:       "Free public marriage index search; certified copies via online request form.",
		},
		{
			// Campbell County — eCCLIX residential search (free, 5/day).
			CountyFIPS: "21037",
			CourtName:  "Campbell County Clerk",
			CourtURL:   "https://campbellcountyclerk.ky.gov",
			SearchURL:  "https://www.ecclix.com/ecclix/Residential/Signup.aspx?id=campbell",
			Note:       "eCCLIX residential search (free, 5 searches/day); commercial subscription available.",
		},
		{
			// Daviess County (Owensboro) — eCCLIX free search tool.
			CountyFIPS: "21059",
			CourtName:  "Daviess County Clerk",
			CourtURL:   "https://www.daviessky.org",
			SearchURL:  "https://daviess.ecclix.com/",
			Note:       "eCCLIX free search tool; marriage records being digitized back to 1966.",
		},
		{
			// Warren County (Bowling Green) — eCCLIX subscriber portal;
			// marriages indexed back to 1797.
			CountyFIPS: "21227",
			CourtName:  "Warren County Clerk",
			CourtURL:   "https://warrencountyclerk.ky.gov",
			SearchURL:  "https://ecclix.com/ecclix/login.aspx",
			Note:       "eCCLIX subscriber login; marriages indexed and scanned back to 1797.",
		},
	},

	// --- Church (Catholic dioceses + parishes) ---------------------------
	Dioceses: []DioceseDef{
		{Slug: "louisville", Name: "Archdiocese of Louisville", Type: "archdiocese",
			Website: "https://www.archlou.org", Directory: "https://www.archlou.org/parishes", HubCityID: "city_louisville_ky"},
		{Slug: "covington", Name: "Diocese of Covington", Type: "diocese",
			Website: "https://www.covingtondiocese.org", Directory: "https://www.covingtondiocese.org/parishes"},
		{Slug: "lexington", Name: "Diocese of Lexington", Type: "diocese",
			Website: "https://www.cdlex.org", Directory: "https://www.cdlex.org/parishes", HubCityID: "city_lexington_ky"},
		{Slug: "owensboro", Name: "Diocese of Owensboro", Type: "diocese",
			Website: "https://www.owensborodiocese.org", Directory: "https://www.owensborodiocese.org/parishes"},
		{Slug: "paducah", Name: "Diocese of Paducah", Type: "diocese",
			Website: "https://www.paducahdiocese.org", Directory: "https://www.paducahdiocese.org/parishes"},
	},

	// Louisville-area parishes in the Archdiocese of Louisville.
	// Names and addresses verified from the archdiocese's own parish
	// directory (archlou.org/parishes). Bulletin URLs verified by direct
	// discovery on each parish's own website.
	Parishes: []ParishDef{
		{DioceseSlug: "louisville", Name: "Cathedral of the Assumption",
			Address: "433 S. Fifth Street, Louisville, KY 40202"},
		{DioceseSlug: "louisville", Name: "St. Agnes Catholic Church",
			Address: "1920 Newburg Road, Louisville, KY 40205"},
		{
			DioceseSlug: "louisville", Name: "St. Margaret Mary Catholic Church",
			Address:     "7813 Shelbyville Rd, Louisville, KY 40222",
			BulletinURL: "https://www.stmm.org/bulletins/",
		},
		{
			DioceseSlug: "louisville", Name: "Catholic Community of St. Francis of Assisi",
			Address:     "1960 Bardstown Road, Louisville, KY 40205",
			BulletinURL: "https://www.ccsfachurch.org/bulletin",
		},
		{
			DioceseSlug: "louisville", Name: "Holy Spirit Catholic Church",
			Address:     "3345 Lexington Rd, Louisville, KY 40206",
			BulletinURL: "https://www.hspirit.org/bulletins",
		},
		{
			DioceseSlug: "louisville", Name: "St. Raphael the Archangel Catholic Church",
			Address:     "2141 Lancashire Ave, Louisville, KY 40205",
			BulletinURL: "https://www.sraparish.org/bulletin",
		},
	},

	// --- Social (wedding-industry Instagram vendors) ---------------------
	Vendors: []VendorDef{
		// Louisville photographers
		{
			Name: "Sarah Katherine Davis Photography", OfficialURL: "https://www.sarahkatherinedavis.com/",
			Handle: "sarahkatherinedavisphotography", SourceClass: "engagement_photographer",
			CityID: "city_louisville_ky", State: "KY", City: "Louisville", Verified: "2026-08-02",
			TikTokHandle: "sarahkatherinedavisphotography",
		},
		{
			Name: "Destiny Rae Photography", OfficialURL: "https://destinyraephotography.com/",
			Handle: "destinyraephotography", SourceClass: "engagement_photographer",
			CityID: "city_louisville_ky", State: "KY", City: "Louisville", Verified: "2026-08-02",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-photographers/destiny-rae-photography-1725811",
		},
		{
			Name: "Amanda Karis Photography", OfficialURL: "https://amandakarisphoto.com/",
			Handle: "amandakarisphotography", SourceClass: "engagement_photographer",
			CityID: "city_louisville_ky", State: "KY", City: "Louisville", Verified: "2026-08-02",
			TikTokHandle: "amandakarisphotography",
		},
		// Louisville venues
		{
			Name: "The Kentucky Rose", OfficialURL: "https://kentuckyroseweddings.com/",
			Handle: "kentuckyroseweddings", SourceClass: "wedding_venue",
			CityID: "city_louisville_ky", State: "KY", City: "Louisville", Verified: "2026-08-02",
			TikTokHandle: "kentuckyroseweddings",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-venues/the-kentucky-rose-6215628",
		},
		// Louisville jewelers
		{
			Name: "Seng Jewelers", OfficialURL: "https://www.sengjewelers.com/",
			Handle: "sengjewelerssince1889", SourceClass: "jeweler",
			CityID: "city_louisville_ky", State: "KY", City: "Louisville", Verified: "2026-08-02",
		},
		{
			Name: "Merkley Kendrick Jewelers", OfficialURL: "https://www.mkjewelers.com/",
			Handle: "merkleykendrickjewelers", SourceClass: "jeweler",
			CityID: "city_louisville_ky", State: "KY", City: "Louisville", Verified: "2026-08-02",
		},
		{
			Name: "Davis Jewelers", OfficialURL: "https://www.davisjewelers.com/",
			Handle: "shopdavisjewelers", SourceClass: "jeweler",
			CityID: "city_louisville_ky", State: "KY", City: "Louisville", Verified: "2026-08-02",
		},
		// Lexington photographers
		{
			Name: "Emily Faith Photography", OfficialURL: "https://emilyfaithphotography.com/",
			Handle: "emilyfaithphotography", SourceClass: "engagement_photographer",
			CityID: "city_lexington_ky", State: "KY", City: "Lexington", Verified: "2026-08-02",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-photographers/emily-faith-photography-5916558",
		},
		// Lexington venues
		{
			Name: "The Manchester Reserve", OfficialURL: "https://themanchesterreserve.com/",
			Handle: "manchesterreservelex", SourceClass: "wedding_venue",
			CityID: "city_lexington_ky", State: "KY", City: "Lexington", Verified: "2026-08-02",
		},
		{
			Name: "The Campbell House", OfficialURL: "https://www.thecampbellhouse.com/",
			Handle: "thecampbellhouselex", SourceClass: "wedding_venue",
			CityID: "city_lexington_ky", State: "KY", City: "Lexington", Verified: "2026-08-02",
			TikTokHandle: "thecampbellhouselex",
		},
	},
}
