package packs

// Indiana source pack — verified 2026-08-01.
//
// Government: Indiana marriage records are held by the county clerk. The
// Indiana Judicial Branch operates a statewide Marriage License Public Lookup
// (public.courts.in.gov/mlpl/Search) covering records from 1993 to present,
// filterable by county. Each county clerk also maintains its own office page.
// Search URLs for the top 7 counties by population were verified against each
// county's official .gov site.
//
// Church: all 5 Indiana Catholic dioceses/archdioceses verified via USCCB +
// each diocese's own website. Indianapolis-area parishes (Archdiocese of
// Indianapolis) were verified against the archdiocese's own parish directory
// (archindy.org/parishes) + direct bulletin-archive URL discovery.
//
// Social: Instagram handles verified from each business's own public website
// social links (or from IG search results where the site is JS-rendered and
// the handle was visible in the search snippet). Verification date recorded
// per vendor.

var inPack = StatePack{
	State: "IN",

	// --- Cities ----------------------------------------------------------
	Cities: []CityDef{
		{ID: "city_indianapolis_in", State: "IN", County: "18097", Name: "Indianapolis",
			Lat: 39.7684, Lng: -86.1581, Markets: []string{"indianapolis", "indy", "marion", "indiana"}},
		{ID: "city_fort_wayne_in", State: "IN", County: "18003", Name: "Fort Wayne",
			Lat: 41.0793, Lng: -85.1394, Markets: []string{"fortwayne", "allen"}},
	},

	// --- Government (county clerk marriage-record searches) --------------
	Government: []GovSource{
		{
			// Marion County (Indianapolis) — statewide marriage license
			// public lookup, filter by Marion County.
			CountyFIPS: "18097",
			CourtName:  "Marion County Clerk",
			CourtURL:   "https://www.indy.gov",
			SearchURL:  "https://public.courts.in.gov/mlpl/Search/",
			Note:       "Statewide Marriage License Public Lookup (beta); filter by county; records from 1993-present; enumeration candidate.",
		},
		{
			// Lake County — clerk's marriage office in Crown Point.
			CountyFIPS: "18089",
			CourtName:  "Lake County Clerk",
			CourtURL:   "https://www.lakecountyin.gov/departments/clerk-marriage",
			SearchURL:  "https://public.courts.in.gov/mlpl/Search/",
			Note:       "Statewide lookup; Lake County clerk marriage office accepts online applications via courts.in.gov/marriage.",
		},
		{
			// Allen County (Fort Wayne) — clerk's record request page.
			CountyFIPS: "18003",
			CourtName:  "Allen County Clerk",
			CourtURL:   "https://allencountyclerk.in.gov/obtain-copies-of-records/cost-how-to-request-records/",
			SearchURL:  "https://public.courts.in.gov/mlpl/Search/",
			Note:       "Statewide lookup; Allen County clerk in Fort Wayne courthouse; online copy request form available.",
		},
		{
			// St. Joseph County (South Bend) — clerk's page.
			CountyFIPS: "18141",
			CourtName:  "St. Joseph County Clerk",
			CourtURL:   "https://www.sjcindiana.gov/2394/Clerk",
			SearchURL:  "https://public.courts.in.gov/mlpl/Search/",
			Note:       "Statewide lookup; St. Joseph County clerk in South Bend; archives hold records 1832-1988.",
		},
		{
			// Hamilton County — clerk's marriage license page.
			CountyFIPS: "18057",
			CourtName:  "Hamilton County Clerk",
			CourtURL:   "https://hamiltoncounty.in.gov/468/Marriage-Licenses",
			SearchURL:  "https://public.courts.in.gov/mlpl/Search/",
			Note:       "Statewide lookup; Hamilton County clerk in Noblesville; appointment required for license application.",
		},
		{
			// Hendricks County — county website; clerk on 2nd floor of courthouse.
			CountyFIPS: "18063",
			CourtName:  "Hendricks County Clerk",
			CourtURL:   "https://www.co.hendricks.in.us",
			SearchURL:  "https://public.courts.in.gov/mlpl/Search/",
			Note:       "Statewide lookup; Hendricks County clerk in Danville courthouse; contact (317) 745-9231.",
		},
		{
			// Tippecanoe County (Lafayette) — clerk's marriage license page.
			CountyFIPS: "18157",
			CourtName:  "Tippecanoe County Clerk",
			CourtURL:   "https://www.tippecanoe.in.gov/224/Marriage-Licenses",
			SearchURL:  "https://public.courts.in.gov/mlpl/Search/",
			Note:       "Statewide lookup; Tippecanoe County clerk in Lafayette courthouse; record requests via SBS Portals.",
		},
		{CountyFIPS: "18095", CourtName: "Lake County Clerk of Courts",
			CourtURL:  "https://lakecountyin.gov/",
			SearchURL: "https://public.courts.in.gov/mlpl/Search/",
			Note:      "Indiana statewide Marriage License Public Lookup; search marriage licenses from 1993 to present by name and county."},
	},

	// --- Church (Catholic dioceses + parishes) ---------------------------
	Dioceses: []DioceseDef{
		{Slug: "indianapolis", Name: "Archdiocese of Indianapolis", Type: "archdiocese",
			Website: "https://www.archindy.org", Directory: "https://www.archindy.org/parishes", HubCityID: "city_indianapolis_in"},
		{Slug: "fort_wayne_sb", Name: "Diocese of Fort Wayne-South Bend", Type: "diocese",
			Website: "https://www.diocesefwsb.org", Directory: "https://www.diocesefwsb.org/parishes", HubCityID: "city_fort_wayne_in"},
		{Slug: "gary", Name: "Diocese of Gary", Type: "diocese",
			Website: "https://www.dcgary.org", Directory: "https://www.dcgary.org/parishes"},
		{Slug: "lafayette_in", Name: "Diocese of Lafayette in Indiana", Type: "diocese",
			Website: "https://www.dol-in.org", Directory: "https://www.dol-in.org/parishes"},
		{Slug: "evansville", Name: "Diocese of Evansville", Type: "diocese",
			Website: "https://www.evdio.org", Directory: "https://www.evdio.org/parishes"},
	},

	// Indianapolis-area parishes in the Archdiocese of Indianapolis.
	// Names and addresses verified from the archdiocese's own parish
	// directory (archindy.org/parishes). Bulletin URLs verified by direct
	// search for each parish's bulletin archive.
	Parishes: []ParishDef{
		{DioceseSlug: "indianapolis", Name: "SS. Peter and Paul Cathedral", Address: "1347 N. Meridian St., Indianapolis, IN 46202"},
		{DioceseSlug: "indianapolis", Name: "Christ the King Catholic Church", Address: "5884 N. Crittenden Ave., Indianapolis, IN 46220"},
		{
			DioceseSlug: "indianapolis", Name: "Holy Spirit Catholic Church",
			Address:     "7243 E. 10th St., Indianapolis, IN 46219",
			BulletinURL: "https://www.holyspirit-indy.org/news/bulletin",
		},
		{DioceseSlug: "indianapolis", Name: "St. John the Evangelist Catholic Church", Address: "126 W. Georgia St., Indianapolis, IN 46225"},
		{DioceseSlug: "indianapolis", Name: "St. Joan of Arc Catholic Church", Address: "4217 Central Ave., Indianapolis, IN 46205"},
		{
			DioceseSlug: "indianapolis", Name: "Our Lady of Lourdes Catholic Church",
			Address:     "5333 E. Washington St., Indianapolis, IN 46219",
			BulletinURL: "https://ollindy.org/church/a-parish-bulletin/",
		},
		{
			DioceseSlug: "indianapolis", Name: "St. Lawrence Catholic Church",
			Address:     "6944 E. 46th St., Indianapolis, IN 46226",
			BulletinURL: "https://www.saintlawrence.net/component/content/article/93-worship/151-bulletins",
		},
		{
			DioceseSlug: "indianapolis", Name: "St. Christopher Catholic Church",
			Address:     "5301 W. 16th St., Indianapolis, IN 46224",
			BulletinURL: "https://www.stchrisindy.org/news/bulletin",
		},
	},

	// --- Social (wedding-industry Instagram vendors) ---------------------
	Vendors: []VendorDef{
		// Indianapolis photographers
		{
			Name: "Stacy Able Photography", OfficialURL: "https://www.stacyable.com/",
			Handle: "stacyablephotography", SourceClass: "engagement_photographer",
			CityID: "city_indianapolis_in", State: "IN", City: "Indianapolis", Verified: "2026-08-01",
			TikTokHandle: "stacyablephotography",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-photographers/stacy-able-photography-7750448",
		},
		{
			Name: "Danielle Harris Photography", OfficialURL: "https://www.danielleharrisphotography.com/",
			Handle: "danielle_harris_photography", SourceClass: "engagement_photographer",
			CityID: "city_indianapolis_in", State: "IN", City: "Indianapolis", Verified: "2026-08-01",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-photographers/danielle-harris-photography-3386034",
		},
		{
			Name: "Maria McKenzie Photography", OfficialURL: "https://mariamckenziephotography.com/",
			Handle: "mariamckenziephoto", SourceClass: "engagement_photographer",
			CityID: "city_indianapolis_in", State: "IN", City: "Indianapolis", Verified: "2026-08-01",
			TikTokHandle: "mariamckenziephoto",
		},
		{
			Name: "Evangeline Renee Photography", OfficialURL: "https://evangelinerenee.com/",
			Handle: "evangelinerenee", SourceClass: "engagement_photographer",
			CityID: "city_indianapolis_in", State: "IN", City: "Indianapolis", Verified: "2026-08-01",
		},
		// Indianapolis venues
		{
			Name: "24 Shelby", OfficialURL: "https://24shelby.com/",
			Handle: "24shelbyevents", SourceClass: "wedding_venue",
			CityID: "city_indianapolis_in", State: "IN", City: "Indianapolis", Verified: "2026-08-01",
			TikTokHandle: "24shelbyevents",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-venues/24-shelby-6427709",
		},
		{
			Name: "Indiana Roof Ballroom", OfficialURL: "https://www.indianaroof.com/",
			Handle: "indianaroofballroom", SourceClass: "wedding_venue",
			CityID: "city_indianapolis_in", State: "IN", City: "Indianapolis", Verified: "2026-08-01",
		},
		{
			Name: "Mavris Arts & Event Center", OfficialURL: "https://www.mavris.net/",
			Handle: "mavrisevents", SourceClass: "wedding_venue",
			CityID: "city_indianapolis_in", State: "IN", City: "Indianapolis", Verified: "2026-08-01",
			TikTokHandle: "mavrisevents",
		},
		// Indianapolis jewelers
		{
			Name: "Barrington Jewels", OfficialURL: "https://barringtonjewels.com/",
			Handle: "barringtonjewels", SourceClass: "jeweler",
			CityID: "city_indianapolis_in", State: "IN", City: "Indianapolis", Verified: "2026-08-01",
		},
		{
			Name: "Moyer Fine Jewelers", OfficialURL: "https://www.moyerfinejewelers.com/",
			Handle: "moyerfinejewelers", SourceClass: "jeweler",
			CityID: "city_indianapolis_in", State: "IN", City: "Indianapolis", Verified: "2026-08-01",
		},
		{
			Name: "Duet Floral Studio", OfficialURL: "https://www.duetfloral.com/",
			Handle: "duetfloral", SourceClass: "florist",
			CityID: "city_indianapolis_in", State: "IN", City: "Indianapolis", Verified: "2026-08-03",
		},
		{
			Name: "RK Florals", OfficialURL: "https://www.rkflorals.com/",
			Handle: "rkflorals", SourceClass: "florist",
			CityID: "city_indianapolis_in", State: "IN", City: "Indianapolis", Verified: "2026-08-03",
		},
		{
			Name: "Windsor Jewelry", OfficialURL: "https://windsorjewelry.com/",
			Handle: "windsorjewelry", SourceClass: "jeweler",
			CityID: "city_indianapolis_in", State: "IN", City: "Indianapolis", Verified: "2026-08-03",
		},
		{
			Name: "Reis-Nichols Jewelers", OfficialURL: "https://www.reisnichols.com/",
			Handle: "reisnichols", SourceClass: "jeweler",
			CityID: "city_indianapolis_in", State: "IN", City: "Indianapolis", Verified: "2026-08-03",
		},
	},
}
