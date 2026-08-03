package packs

// Alabama source pack — verified 2026-08-01.
//
// Government: Alabama marriage records are held by the county probate court.
// Since Act 2019-340 (effective Aug 29 2019) probate courts record marriage
// certificates rather than issuing licenses. Search URLs for the top 7 counties
// by population were verified against each county's official .gov site or its
// Landmark WEB / ROAM / publicsearch.us portal.
//
// Church: the Archdiocese of Mobile and the Diocese of Birmingham cover the
// state (dioceses verified via USCCB). Mobile-area parishes were verified
// against the archdiocese's own parish directory + gcatholic.org church list;
// bulletin URLs verified by direct fetch of each parish's bulletin archive.
//
// Social: Instagram handles verified from each business's own public website
// social links. Verification date recorded per vendor.

var alPack = StatePack{
	State: "AL",

	// --- Cities ----------------------------------------------------------
	Cities: []CityDef{
		{ID: "city_birmingham_al", State: "AL", County: "01073", Name: "Birmingham",
			Lat: 33.5207, Lng: -86.8025, Markets: []string{"birmingham", "bham", "jefferson", "al"}},
		{ID: "city_mobile_al", State: "AL", County: "01097", Name: "Mobile",
			Lat: 30.6954, Lng: -88.0430, Markets: []string{"mobile", "al", "gulfcoast"}},
	},

	// --- Government (county probate court marriage-record searches) ------
	Government: []GovSource{
		{
			// Jefferson County (Birmingham) — Landmark WEB portal includes
			// marriage records (Birmingham index 1987–current, images 1993–
			// current; Bessemer 1965–current). Free account required to print.
			CountyFIPS: "01073",
			CourtName:  "Jefferson County Probate Court",
			CourtURL:   "https://jeffcoprobatecourt.com",
			SearchURL:  "http://landmarkweb.jccal.org/landmarkweb",
			Note:       "Landmark WEB portal; marriage records indexed 1987+, images 1993+; enumeration candidate.",
		},
		{
			// Madison County (Huntsville) — recorded documents online portal;
			// free account required. Marriage certificates recorded post-2019.
			CountyFIPS: "01089",
			CourtName:  "Madison County Probate Court",
			CourtURL:   "https://www.madisoncountyal.gov/departments/probate-judge",
			SearchURL:  "https://www.madisoncountyal.gov/departments/probate-judge/recorded-documents",
			Note:       "Online record portal (account required); marriage filtering needs testing.",
		},
		{
			// Mobile County — Landmark WEB portal; marriage records indexed
			// 1813–present, viewable Sep 2019–present. Free account required.
			CountyFIPS: "01097",
			CourtName:  "Mobile County Probate Court",
			CourtURL:   "https://probate.mobilecountyal.gov",
			SearchURL:  "https://probate.mobilecountyal.gov/public-records/records-search/",
			Note:       "Landmark WEB portal; marriage indexed 1813+, images Sep 2019+; enumeration candidate.",
		},
		{
			// Montgomery County — probate records search; marriage license
			// indexes from ~1975 to present. Pre-1975 records being added.
			CountyFIPS: "01101",
			CourtName:  "Montgomery County Probate Court",
			CourtURL:   "https://www.montgomeryprobatecourtal.gov",
			SearchURL:  "https://www.montgomeryprobatecourtal.gov/divisions/records-recording/probate-records-search",
			Note:       "Probate records search; marriage indexes ~1975+; enumeration capability needs testing.",
		},
		{
			// Shelby County — ROAM public records search with Marriage
			// Licenses index type; returns MARR APP & CERT doc types.
			CountyFIPS: "01117",
			CourtName:  "Shelby County Probate Court",
			CourtURL:   "https://www.shelbyal.com/285/Probate-Court",
			SearchURL:  "https://probaterecords.shelbyal.com/shelby/",
			Note:       "ROAM portal with Marriage Licenses index; enumeration candidate.",
		},
		{
			// Baldwin County — Open Baldwin portal with dedicated marriage
			// license search (pre-2019) and marriage certificate search
			// (2019–present). Also available via publicsearch.us.
			CountyFIPS: "01003",
			CourtName:  "Baldwin County Probate Court",
			CourtURL:   "https://baldwincountyal.gov/government/probate-office",
			SearchURL:  "https://open.baldwincountyal.gov/",
			Note:       "Open Baldwin portal; dedicated marriage license + certificate search; enumeration candidate.",
		},
		{
			// Tuscaloosa County — probate records search; free account
			// required to view/print. Records from 3/20/1975 to current.
			CountyFIPS: "01125",
			CourtName:  "Tuscaloosa County Probate Court",
			CourtURL:   "https://www.tuscco.com/probate/",
			SearchURL:  "https://www.tuscco.com/probate/records-and-recordings/",
			Note:       "Probate records search (account required); records from 1975+; marriage filtering needs testing.",
		},
	},

	// --- Church (Catholic dioceses + parishes) ---------------------------
	Dioceses: []DioceseDef{
		{Slug: "mobile", Name: "Archdiocese of Mobile", Type: "archdiocese",
			Website: "https://mobarch.org", Directory: "https://mobarch.org/parishes", HubCityID: "city_mobile_al"},
		{Slug: "birmingham", Name: "Diocese of Birmingham", Type: "diocese",
			Website: "https://www.bhmdiocese.org", Directory: "https://www.bhmdiocese.org/parishes", HubCityID: "city_birmingham_al"},
		{Slug: "montgomery", Name: "Diocese of Montgomery", Type: "diocese",
			Website: "https://www.dioceseofmontgomery.org", Directory: "https://www.dioceseofmontgomery.org/parishes"},
	},

	// Mobile-area parishes in the Archdiocese of Mobile. Names and addresses
	// verified against the archdiocese's own parish directory + gcatholic.org
	// church list. Bulletin URLs verified by direct fetch of each parish's
	// bulletin archive.
	Parishes: []ParishDef{
		{
			DioceseSlug: "mobile", Name: "Cathedral Basilica of the Immaculate Conception",
			Address:     "307 Conti Street, Mobile, AL 36602",
			BulletinURL: "https://mobilecathedral.org/bulletins",
		},
		{
			DioceseSlug: "mobile", Name: "Corpus Christi Parish",
			Address: "6300 McKenna Dr, Mobile, AL 36608",
		},
		{
			DioceseSlug: "mobile", Name: "St. Pius X Catholic Church",
			Address: "217 S Sage Avenue, Mobile, AL 36606",
		},
		{
			DioceseSlug: "mobile", Name: "Our Savior Parish",
			Address: "1801 Cody Road S, Mobile, AL 36695",
		},
		{
			DioceseSlug: "mobile", Name: "St. Mary's Parish",
			Address: "1453 Old Shell Road, Mobile, AL 36604",
		},
		{
			DioceseSlug: "mobile", Name: "Our Lady of Lourdes Parish",
			Address:     "1621 Boykin Blvd, Mobile, AL 36605",
			BulletinURL: "https://ourladyoflourdesmobile.weebly.com/bulletin.html",
		},
		{
			DioceseSlug: "mobile", Name: "St. Vincent de Paul Parish",
			Address: "6625 Three Notch Rd, Mobile, AL 36619",
		},
	},

	// --- Social (wedding-industry Instagram vendors) ---------------------
	Vendors: []VendorDef{
		// Birmingham photographers
		{
			Name: "Maddie Moore Photography", OfficialURL: "https://maddiemoore.com/",
			Handle: "maddiemoorephoto", SourceClass: "engagement_photographer",
			CityID: "city_birmingham_al", State: "AL", City: "Birmingham", Verified: "2026-08-01",
			TikTokHandle: "maddiemoorephoto",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-photographers/maddie-moore-photography-3541789",
		},
		{
			Name: "Krista Suzanne Photography", OfficialURL: "https://kristasuzanne.com/",
			Handle: "kristasuzannephotography", SourceClass: "engagement_photographer",
			CityID: "city_birmingham_al", State: "AL", City: "Birmingham", Verified: "2026-08-01",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-photographers/krista-suzanne-photography-1844258",
		},
		{
			Name: "Elizabeth Hall Photography", OfficialURL: "https://elizabethhallphoto.com/",
			Handle: "elizabethhallphotography", SourceClass: "engagement_photographer",
			CityID: "city_birmingham_al", State: "AL", City: "Birmingham", Verified: "2026-08-01",
			TikTokHandle: "elizabethhallphotography",
		},
		{
			Name: "Kevin Roberts Photography", OfficialURL: "http://www.kevinrobertsimages.com/",
			Handle: "kevinrobertsphotography", SourceClass: "engagement_photographer",
			CityID: "city_birmingham_al", State: "AL", City: "Birmingham", Verified: "2026-08-01",
		},
		// Birmingham venues
		{
			Name: "The Sonnet House", OfficialURL: "https://www.thesonnethouse.com/",
			Handle: "thesonnethouse", SourceClass: "wedding_venue",
			CityID: "city_birmingham_al", State: "AL", City: "Birmingham", Verified: "2026-08-01",
			TikTokHandle: "thesonnethouse",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-venues/the-sonnet-house-8101728",
		},
		// Birmingham jeweler
		{
			Name: "Levy's Fine Jewelry", OfficialURL: "https://levysfinejewelry.com/",
			Handle: "levysfj", SourceClass: "jeweler",
			CityID: "city_birmingham_al", State: "AL", City: "Birmingham", Verified: "2026-08-01",
		},
		// Mobile venues
		{
			Name: "The Pillars Mobile", OfficialURL: "https://www.pillarsmobile.com/",
			Handle: "pillarsmobile", SourceClass: "wedding_venue",
			CityID: "city_mobile_al", State: "AL", City: "Mobile", Verified: "2026-08-01",
		},
		{
			Name: "Crown Hall", OfficialURL: "https://crownhall.events/",
			Handle: "crownhallevents", SourceClass: "wedding_venue",
			CityID: "city_mobile_al", State: "AL", City: "Mobile", Verified: "2026-08-01",
			TikTokHandle: "crownhallevents",
		},
		// Mobile jeweler
		{
			Name: "Zundel's Jewelry", OfficialURL: "https://zundelsjewelry.com/",
			Handle: "zundels_jewelry", SourceClass: "jeweler",
			CityID: "city_mobile_al", State: "AL", City: "Mobile", Verified: "2026-08-01",
		},
		{
			Name: "Katie & Alec Photography", OfficialURL: "https://katieandalec.com/",
			Handle: "katieandalecphoto", SourceClass: "engagement_photographer",
			CityID: "city_birmingham_al", State: "AL", City: "Birmingham", Verified: "2026-08-03",
		},
		{
			Name: "Stanley Parrish Photography", OfficialURL: "https://www.stanleyparrish.com/",
			Handle: "stanleyparrish", SourceClass: "engagement_photographer",
			CityID: "city_birmingham_al", State: "AL", City: "Birmingham", Verified: "2026-08-03",
		},
		{
			Name: "Hatcher Farms Venue", OfficialURL: "https://hatcherfarmsvenue.com/",
			Handle: "hatcherfarmsvenue", SourceClass: "wedding_venue",
			CityID: "city_mobile_al", State: "AL", City: "Mobile", Verified: "2026-08-03",
		},
	},
}
