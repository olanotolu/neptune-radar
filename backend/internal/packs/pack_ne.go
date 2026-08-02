package packs

// Nebraska source pack — verified 2026-08-01.
//
// Government: Nebraska marriage records are held by the county clerk. The top
// 7 counties by population were verified against each county's official .gov
// site. Douglas and Lancaster have online search portals; the rest are
// request-oriented (in-person or mail-in certified copy requests).
//
// Church: all 3 Nebraska Catholic dioceses/archdiocese verified via USCCB +
// each diocese's own website. Omaha-area parishes (Archdiocese of Omaha) were
// verified against the archdiocese's own parish directory (archomaha.org) +
// direct bulletin-archive URL discovery on each parish's own website.
//
// Social: Instagram handles verified from each business's own public website
// social links. Verification date recorded per vendor.

var nePack = StatePack{
	State: "NE",

	// --- Cities ----------------------------------------------------------
	Cities: []CityDef{
		{ID: "city_omaha_ne", State: "NE", County: "31055", Name: "Omaha",
			Lat: 41.2565, Lng: -95.9345, Markets: []string{"omaha", "oma", "ne", "douglas"}},
		{ID: "city_lincoln_ne", State: "NE", County: "31109", Name: "Lincoln",
			Lat: 40.8258, Lng: -96.6852, Markets: []string{"lincoln", "lnk", "lancaster", "ne"}},
	},

	// --- Government (county clerk marriage-record searches) --------------
	Government: []GovSource{
		{
			// Douglas County (Omaha) — county clerk marriage license search.
			CountyFIPS: "31055",
			CourtName:  "Douglas County Clerk",
			CourtURL:   "https://clerk.douglascounty-ne.gov",
			SearchURL:  "https://clerk.douglascounty-ne.gov/find-a-marriage-license/",
			Note:       "Online marriage license search portal; records from 1856 to present; enumeration candidate.",
		},
		{
			// Lancaster County (Lincoln) — county clerk marriage license search.
			CountyFIPS: "31109",
			CourtName:  "Lancaster County Clerk",
			CourtURL:   "https://www.lancaster.ne.gov/clerk",
			SearchURL:  "https://app.lincoln.ne.gov/aspx/cnty/clerk/marrsrch.aspx",
			Note:       "Online marriage license search portal; records from mid-June 1976 to present; enumeration candidate.",
		},
		{
			// Sarpy County — county clerk marriage licenses; no online search,
			// request-oriented (in-person or mail-in certified copy).
			CountyFIPS: "31155",
			CourtName:  "Sarpy County Clerk",
			CourtURL:   "https://www.sarpy.gov/183/County-Clerk-Register-of-Deeds",
			SearchURL:  "https://www.sarpy.gov/214/Marriage-Licenses",
			Note:       "Marriage license info page; no online search; certified copies by mail or in-person only.",
		},
		{
			// Hall County (Grand Island) — county clerk marriage licenses;
			// request-oriented.
			CountyFIPS: "31079",
			CourtName:  "Hall County Clerk",
			CourtURL:   "https://hallcountyne.gov/departments/county_clerk/index.php",
			SearchURL:  "https://hallcountyne.gov/departments/county_clerk/marriage_licenses.php",
			Note:       "Marriage license info page; no online search; certified copies by mail or in-person only.",
		},
		{
			// Buffalo County (Kearney) — county clerk marriage licenses;
			// request-oriented.
			CountyFIPS: "31019",
			CourtName:  "Buffalo County Clerk",
			CourtURL:   "https://buffalocounty.ne.gov/county-offices/clerk",
			SearchURL:  "https://buffalocounty.ne.gov/county-offices/clerk/marriage-license",
			Note:       "Marriage license info page; no online search; certified copies by mail or in-person only.",
		},
		{
			// Dodge County (Fremont) — county clerk marriage licenses;
			// request-oriented.
			CountyFIPS: "31053",
			CourtName:  "Dodge County Clerk",
			CourtURL:   "https://dodgecounty.nebraska.gov/clerk",
			SearchURL:  "https://dodgecounty.nebraska.gov/clerk",
			Note:       "Clerk page with marriage license info; no online search; certified copies by mail or in-person only.",
		},
		{
			// Scotts Bluff County — county clerk marriage licenses;
			// request-oriented.
			CountyFIPS: "31157",
			CourtName:  "Scotts Bluff County Clerk",
			CourtURL:   "https://scottsbluffcountyne.gov/county-clerk/",
			SearchURL:  "https://scottsbluffcountyne.gov/county-clerk/",
			Note:       "Clerk page with marriage license info; no online search; certified copies by mail or in-person only.",
		},
	},

	// --- Church (Catholic dioceses + parishes) ---------------------------
	Dioceses: []DioceseDef{
		{Slug: "omaha", Name: "Archdiocese of Omaha", Type: "archdiocese",
			Website: "https://archomaha.org", Directory: "https://archomaha.org/find-a-parish", HubCityID: "city_omaha_ne"},
		{Slug: "lincoln", Name: "Diocese of Lincoln", Type: "diocese",
			Website: "https://www.lincolndiocese.org", Directory: "https://www.lincolndiocese.org/parishes", HubCityID: "city_lincoln_ne"},
		{Slug: "grand_island", Name: "Diocese of Grand Island", Type: "diocese",
			Website: "https://www.gidiocese.org", Directory: "https://www.gidiocese.org/parishes"},
	},

	// Omaha-area parishes in the Archdiocese of Omaha. Names and addresses
	// verified from the archdiocese's own parish directory (archomaha.org) and
	// each parish's own website. Bulletin URLs verified by direct discovery on
	// each parish's own site.
	Parishes: []ParishDef{
		{
			DioceseSlug: "omaha", Name: "St. Cecilia Cathedral",
			Address:     "701 N 40th St, Omaha, NE 68131",
			BulletinURL: "https://stceciliacathedral.org",
		},
		{
			DioceseSlug: "omaha", Name: "St. Margaret Mary Catholic Church",
			Address:     "6116 Dodge St, Omaha, NE 68132",
			BulletinURL: "https://www.smmomaha.org/about-us/bulletin",
		},
		{
			DioceseSlug: "omaha", Name: "St. Stephen the Martyr Catholic Church",
			Address:     "16701 S St, Omaha, NE 68135",
			BulletinURL: "https://stephen.org",
		},
		{
			DioceseSlug: "omaha", Name: "St. Patrick Catholic Church",
			Address:     "18602 W Maple Rd, Elkhorn, NE 68022",
			BulletinURL: "https://www.stpatselkhorn.org/bulletin.html",
		},
		{
			DioceseSlug: "omaha", Name: "St. Pius X Catholic Church",
			Address:     "6905 Blondo St, Omaha, NE 68104",
			BulletinURL: "https://www.stpiusxomaha.org/publications/",
		},
		{
			DioceseSlug: "omaha", Name: "St. Robert Bellarmine Catholic Parish",
			Address:     "11802 Pacific St, Omaha, NE 68154",
			BulletinURL: "https://www.stroberts.com",
		},
	},

	// --- Social (wedding-industry Instagram vendors) ---------------------
	Vendors: []VendorDef{
		// Omaha photographers
		{
			Name: "Julie Trinh Photography", OfficialURL: "https://julietrinhphotography.com/",
			Handle: "julietrinhphotography", SourceClass: "engagement_photographer",
			CityID: "city_omaha_ne", State: "NE", City: "Omaha", Verified: "2026-08-01",
			TikTokHandle: "julietrinhphotography",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-photographers/julie-trinh-photography-7999804",
		},
		{
			Name: "Caitlin and Camera", OfficialURL: "https://caitlinandcamera.com/",
			Handle: "caitlinandcamera", SourceClass: "engagement_photographer",
			CityID: "city_omaha_ne", State: "NE", City: "Omaha", Verified: "2026-08-01",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-photographers/caitlin-and-camera-1229252",
		},
		{
			Name: "Addie Zelasko Photography", OfficialURL: "https://addiezelaskophotography.com/",
			Handle: "addiezelaskophotography", SourceClass: "engagement_photographer",
			CityID: "city_omaha_ne", State: "NE", City: "Omaha", Verified: "2026-08-01",
			TikTokHandle: "addiezelaskophotography",
		},
		// Omaha venues
		{
			Name: "The Players Club Omaha", OfficialURL: "https://www.theplayersclubomaha.com/",
			Handle: "theplayersclubomaha_events", SourceClass: "wedding_venue",
			CityID: "city_omaha_ne", State: "NE", City: "Omaha", Verified: "2026-08-01",
			TikTokHandle: "theplayersclubomaha_events",
		},
		{
			Name: "Leo Ballroom", OfficialURL: "https://www.leoballroom.com/",
			Handle: "leoballroom.omaha", SourceClass: "wedding_venue",
			CityID: "city_omaha_ne", State: "NE", City: "Omaha", Verified: "2026-08-01",
			TikTokHandle: "leoballroom.omaha",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-venues/leo-ballroom-8987193",
		},
		{
			Name: "Vintage Ballroom", OfficialURL: "https://www.vintageballroom.com/",
			Handle: "vintageballroom", SourceClass: "wedding_venue",
			CityID: "city_omaha_ne", State: "NE", City: "Omaha", Verified: "2026-08-01",
		},
		// Omaha jewelers
		{
			Name: "Martin Jewelry", OfficialURL: "https://www.martinjewelry.net/",
			Handle: "martinjewelryne", SourceClass: "jeweler",
			CityID: "city_omaha_ne", State: "NE", City: "Omaha", Verified: "2026-08-01",
		},
		{
			Name: "Hériter Gems", OfficialURL: "https://www.heritergems.com/",
			Handle: "heritergems", SourceClass: "jeweler",
			CityID: "city_omaha_ne", State: "NE", City: "Omaha", Verified: "2026-08-01",
		},
		// Lincoln photographer
		{
			Name: "Emily Moore Creative", OfficialURL: "https://emilymoorecreative.com/",
			Handle: "emilymoorecreative", SourceClass: "engagement_photographer",
			CityID: "city_lincoln_ne", State: "NE", City: "Lincoln", Verified: "2026-08-01",
		},
		// Lincoln jeweler
		{
			Name: "Sartor Hamann Jewelers", OfficialURL: "https://sartorhamann.com/",
			Handle: "sartorhamann", SourceClass: "jeweler",
			CityID: "city_lincoln_ne", State: "NE", City: "Lincoln", Verified: "2026-08-01",
		},
	},
}
