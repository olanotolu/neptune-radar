package packs

// Oregon source pack — verified 2026-08-01.
//
// Government: Oregon marriage records are held by the county clerk (Recording
// Division). Search URLs for the top 7 counties by population were verified
// against each county's official .gov site or its Digital Research Room portal.
// Oregon marriage records have a 50-year access restriction (ORS 432.121).
//
// Church: both Oregon Catholic jurisdictions verified via USCCB + each
// diocese's own website. Portland-area parishes (Archdiocese of Portland in
// Oregon) verified against Wikipedia's list of churches in the archdiocese
// (which cites the archdiocese's own records) + direct bulletin-archive URL
// discovery.
//
// Social: Instagram handles verified from each business's own public website
// social links. Verification date recorded per vendor.

var orPack = StatePack{
	State: "OR",

	// --- Cities ----------------------------------------------------------
	Cities: []CityDef{
		{
			ID: "city_portland_or", State: "OR", County: "41051", Name: "Portland",
			Lat: 45.5152, Lng: -122.6784,
			Markets: []string{"portland", "pdx", "oregon", "multnomah"},
		},
		{
			ID: "city_eugene_or", State: "OR", County: "41039", Name: "Eugene",
			Lat: 44.0521, Lng: -123.0868,
			Markets: []string{"eugene", "lane", "oregon"},
		},
	},

	// --- Government (county clerk marriage-record searches) --------------
	Government: []GovSource{
		{
			// Multnomah County (Portland) — MultcoRecords.com recorded document
			// search; marriage records digitally available 1990–present.
			CountyFIPS: "41051",
			CourtName:  "Multnomah County Clerk",
			CourtURL:   "https://multco.us/services/marriage-licenses",
			SearchURL:  "https://multcorecords.com/MarriageSearch",
			Note:       "Dedicated marriage search portal; records 1990–present, 1980–1989 index only; 50-yr access restriction.",
		},
		{
			// Washington County (Hillsboro) — Recording Division; copies by
			// phone/email or in-office public terminals.
			CountyFIPS: "41067",
			CourtName:  "Washington County Clerk",
			CourtURL:   "https://www.washingtoncountyor.gov/at/recording/marriage-licenses",
			SearchURL:  "https://www.washingtoncountyor.gov/at/recording/copies-documents",
			Note:       "Marriage licenses issued by Recording Division; copies via phone/email or in-office terminal; no free online search.",
		},
		{
			// Clackamas County (Oregon City) — Recording Division of the County
			// Clerk; marriage records dating back to the 1840s.
			CountyFIPS: "41005",
			CourtName:  "Clackamas County Clerk",
			CourtURL:   "https://www.clackamas.us/recording",
			SearchURL:  "https://www.clackamas.us/how-to-get-a-marriage-license",
			Note:       "Recording Division issues marriage licenses; public viewing area for research; no free online search.",
		},
		{
			// Lane County (Eugene) — Deeds & Records Marriage License Section;
			// certified copies by phone or mail.
			CountyFIPS: "41039",
			CourtName:  "Lane County Clerk",
			CourtURL:   "https://www.lanecountyor.gov/government/county_departments/county_administration/general_county_administration/operations/county_clerk/marriage_licenses",
			SearchURL:  "https://www.lanecountyor.gov/government/county_departments/county_administration/general_county_administration/operations/county_clerk/marriage_licenses/ordering_a_certified_copy_of_your_marriage_license",
			Note:       "Deeds & Records office; certified copies by phone (541-682-3653) or mail; no free online search.",
		},
		{
			// Marion County (Salem) — online DeedSearch portal + dedicated
			// marriage license search application.
			CountyFIPS: "41047",
			CourtName:  "Marion County Clerk",
			CourtURL:   "https://www.co.marion.or.us/CO/records/Pages/marriage.aspx",
			SearchURL:  "https://apps.co.marion.or.us/MarriageLicenseSearch/Disclaimer.aspx",
			Note:       "Dedicated marriage license search app; historical 1849–1977 index also available; enumeration candidate.",
		},
		{
			// Deschutes County (Bend) — Clerk's Office Digital Research Room;
			// marriage record copies ordered online or by mail.
			CountyFIPS: "41017",
			CourtName:  "Deschutes County Clerk",
			CourtURL:   "https://www.deschutescounty.gov/470/Marriage-Licenses",
			SearchURL:  "https://www.deschutescounty.gov/525/Marriage-Record-Copy-Request",
			Note:       "Marriage record copy request page; online ordering at deschutescounty.gov/marriagecopies; Digital Research Room for recorded docs.",
		},
		{
			// Jackson County (Medford) — Digital Research Room for recorded
			// documents; marriage licenses issued by Recording Office.
			CountyFIPS: "41029",
			CourtName:  "Jackson County Clerk",
			CourtURL:   "https://jacksoncountyor.gov/departments/clerk/recording/services/marriage_licenses/index.php",
			SearchURL:  "https://apps.jacksoncountyor.gov/DigitalResearchRoomPublic",
			Note:       "Digital Research Room for recorded docs (1984–present); marriage licenses via Recording Office; enumeration needs testing.",
		},
	},

	// --- Church (Catholic dioceses + parishes) ---------------------------
	Dioceses: []DioceseDef{
		{
			Slug: "portland_or", Name: "Archdiocese of Portland in Oregon", Type: "archdiocese",
			Website: "https://www.archdpdx.org", Directory: "https://archdpdx.org/parishfinder",
			HubCityID: "city_portland_or",
		},
		{
			Slug: "baker", Name: "Diocese of Baker", Type: "diocese",
			Website: "https://www.dioceseofbaker.org", Directory: "https://www.dioceseofbaker.org/parishes",
		},
	},

	// Portland-area parishes in the Archdiocese of Portland in Oregon.
	// Names verified from Wikipedia's list of churches in the archdiocese
	// (which cites the archdiocese's own records). Bulletin URLs verified by
	// direct search for each parish's bulletin archive.
	Parishes: []ParishDef{
		{
			DioceseSlug: "portland_or", Name: "St. Mary's Cathedral of the Immaculate Conception",
			Address:     "1716 NW Davis St, Portland, OR 97209",
			BulletinURL: "https://www.maryscathedral.com/bulletin",
		},
		{
			DioceseSlug: "portland_or", Name: "St. Philip Neri Catholic Church",
			Address:     "2408 SE 16th Ave, Portland, OR 97214",
			BulletinURL: "https://www.stphilipneripdx.org/news/category/bulletin",
		},
		{
			DioceseSlug: "portland_or", Name: "All Saints Parish",
			Address: "3847 NE Glisan St, Portland, OR 97232",
		},
		{
			DioceseSlug: "portland_or", Name: "Holy Rosary Church",
			Address:     "375 NE Clackamas St, Portland, OR 97232",
			BulletinURL: "https://holyrosarypdx.org/bulletins",
		},
		{
			DioceseSlug: "portland_or", Name: "St. Thomas More Catholic Church",
			Address: "3525 SW Patton Rd, Portland, OR 97221",
		},
		{
			DioceseSlug: "portland_or", Name: "St. Charles Borromeo Parish",
			Address: "5310 NE 42nd Ave, Portland, OR 97218",
		},
		{
			DioceseSlug: "portland_or", Name: "Holy Redeemer Catholic Church",
			Address: "25 N Rosa Parks Way, Portland, OR 97217",
		},
		{
			DioceseSlug: "portland_or", Name: "Holy Family Catholic Church",
			Address: "7525 SE Cesar Chavez Blvd, Portland, OR 97206",
		},
	},

	// --- Social (wedding-industry Instagram vendors) ---------------------
	Vendors: []VendorDef{
		// Portland photographers
		{
			Name: "Catalina Jean Photography", OfficialURL: "https://catalinajean.com/",
			Handle: "catalinajeanphoto", SourceClass: "engagement_photographer",
			CityID: "city_portland_or", State: "OR", City: "Portland", Verified: "2026-08-01",
			TikTokHandle: "catalinajeanphoto",
		},
		{
			Name: "Maria Lamb Photography", OfficialURL: "https://marialamb.co/",
			Handle: "maria.lamb", SourceClass: "engagement_photographer",
			CityID: "city_portland_or", State: "OR", City: "Portland", Verified: "2026-08-01",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-photographers/maria-lamb-photography-1055922",
		},
		{
			Name: "Chris Brodell Photography", OfficialURL: "https://www.chrisbrodell.com/",
			Handle: "chrisbrodell", SourceClass: "engagement_photographer",
			CityID: "city_portland_or", State: "OR", City: "Portland", Verified: "2026-08-01",
			TikTokHandle: "chrisbrodell",
		},
		// Portland venues
		{
			Name: "Easton Broad", OfficialURL: "https://www.eastonbroadpdx.com/",
			Handle: "eastonbroadpdx", SourceClass: "wedding_venue",
			CityID: "city_portland_or", State: "OR", City: "Portland", Verified: "2026-08-01",
			TikTokHandle: "eastonbroadpdx",
		},
		{
			Name: "Everett West", OfficialURL: "https://www.everettwest.com/",
			Handle: "everettwestpdx", SourceClass: "wedding_venue",
			CityID: "city_portland_or", State: "OR", City: "Portland", Verified: "2026-08-01",
			TikTokHandle: "everettwestpdx",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-venues/everett-west-9988076",
		},
		{
			Name: "Highland Farms Oregon", OfficialURL: "https://highlandfarmsoregon.com/",
			Handle: "highlandfarmsor", SourceClass: "wedding_venue",
			CityID: "city_portland_or", State: "OR", City: "Portland", Verified: "2026-08-01",
		},
		// Portland jewelers
		{
			Name: "Alchemy Jeweler", OfficialURL: "https://alchemyjeweler.com/",
			Handle: "alchemyjeweler", SourceClass: "jeweler",
			CityID: "city_portland_or", State: "OR", City: "Portland", Verified: "2026-08-01",
		},
		{
			Name: "Malka Diamonds & Jewelry", OfficialURL: "https://www.malkadiamonds.com/",
			Handle: "malkadiamonds", SourceClass: "jeweler",
			CityID: "city_portland_or", State: "OR", City: "Portland", Verified: "2026-08-01",
		},
		{
			Name: "Starflower", OfficialURL: "https://www.starflowerpassion.com/",
			Handle: "starflowerpdx", SourceClass: "florist",
			CityID: "city_portland_or", State: "OR", City: "Portland", Verified: "2026-08-03",
		},
		{
			Name: "Wrong Number Floral", OfficialURL: "https://wrongnumberfloral.com/",
			Handle: "wrongnumberfloral", SourceClass: "florist",
			CityID: "city_eugene_or", State: "OR", City: "Eugene", Verified: "2026-08-03",
		},
		{
			Name: "Darling Dahlia Floral", OfficialURL: "https://darlingdahliafloral.com/",
			Handle: "darling_dahlia_floral", SourceClass: "florist",
			CityID: "city_portland_or", State: "OR", City: "Portland", Verified: "2026-08-03",
		},
		{
			Name: "Little Banana Bakery", OfficialURL: "https://littlebananabakery.com/",
			Handle: "littlebananabakery", SourceClass: "wedding_cake",
			CityID: "city_portland_or", State: "OR", City: "Portland", Verified: "2026-08-03",
		},
		{
			Name: "Next Dimension Bakery", OfficialURL: "https://www.nextdimensionbakery.com/",
			Handle: "next_dimension_bakery", SourceClass: "wedding_cake",
			CityID: "city_portland_or", State: "OR", City: "Portland", Verified: "2026-08-03",
		},
		{
			Name: "Banquette Bakery", OfficialURL: "https://www.banquettepdx.com/",
			Handle: "banquettepdx", SourceClass: "wedding_cake",
			CityID: "city_portland_or", State: "OR", City: "Portland", Verified: "2026-08-03",
		},
	},
}
