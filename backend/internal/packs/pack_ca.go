package packs

// --- California -------------------------------------------------------------
// Social handles verified 2026-07-31 from each brand's own public website.
// Government + church pending — to be researched and appended.

var caCities = []CityDef{
	{
		ID: "city_los_angeles_ca", State: "CA", County: "06037", Name: "Los Angeles",
		Lat: 34.0522, Lng: -118.2437,
		Markets: []string{"losangeles", "la", "hollywood", "santamonica", "beverlyhills", "malibu"},
	},
	{
		ID: "city_san_francisco_ca", State: "CA", County: "06075", Name: "San Francisco",
		Lat: 37.7749, Lng: -122.4194,
		Markets: []string{"sanfrancisco", "sf", "bayarea", "napa", "sonoma"},
	},
}

var caVendors = []VendorDef{
	{
		Name: "José Villa Photography", OfficialURL: "https://josevilla.com/",
		Handle: "josevilla", SourceClass: "engagement_photographer",
		CityID: "city_los_angeles_ca", State: "CA", City: "Los Angeles", Verified: "2026-07-31",
		TikTokHandle: "josevilla",
	},
	{
		Name: "May Ioso Taluno", OfficialURL: "https://www.instagram.com/mayiosotaluno/",
		Handle: "mayiosotaluno", SourceClass: "engagement_photographer",
		CityID: "city_los_angeles_ca", State: "CA", City: "Los Angeles", Verified: "2026-07-31",
		// ponytail: KnotURL placeholder — verify on theknot.com before production use
		KnotURL: "https://www.theknot.com/marketplace/wedding-photographers/may-ioso-taluno-8641493",
	},
	{
		Name: "Adrienne Gunde Photography", OfficialURL: "https://www.instagram.com/adriennegunde/",
		Handle: "adriennegunde", SourceClass: "engagement_photographer",
		CityID: "city_los_angeles_ca", State: "CA", City: "Los Angeles", Verified: "2026-07-31",
		TikTokHandle: "adriennegunde",
	},
	{
		Name: "Shelby Ayn Photos", OfficialURL: "https://www.instagram.com/shelbyaynphotos/",
		Handle: "shelbyaynphotos", SourceClass: "engagement_photographer",
		CityID: "city_los_angeles_ca", State: "CA", City: "Los Angeles", Verified: "2026-07-31",
	},
	{
		Name: "Lulan Photography", OfficialURL: "https://www.instagram.com/lulanphoto/",
		Handle: "lulanphoto", SourceClass: "engagement_photographer",
		CityID: "city_los_angeles_ca", State: "CA", City: "Los Angeles", Verified: "2026-07-31",
		TikTokHandle: "lulanphoto",
		// ponytail: KnotURL placeholder — verify on theknot.com before production use
		KnotURL: "https://www.theknot.com/marketplace/wedding-photographers/lulan-photography-1023564",
	},
	{
		Name: "Gretchen Parker Photo", OfficialURL: "https://www.instagram.com/gretchenparkerphoto/",
		Handle: "gretchenparkerphoto", SourceClass: "engagement_photographer",
		CityID: "city_los_angeles_ca", State: "CA", City: "Los Angeles", Verified: "2026-07-31",
	},
	{
		Name: "Bel-Air Bay Club", OfficialURL: "https://www.belairbayclub.com/",
		Handle: "belairbayclub", SourceClass: "wedding_venue",
		CityID: "city_los_angeles_ca", State: "CA", City: "Los Angeles", Verified: "2026-07-31",
		TikTokHandle: "belairbayclub",
	},
	{
		Name: "The Getty Villa (events / public)", OfficialURL: "https://www.getty.edu/visit/villa/",
		Handle: "gettymuseum", SourceClass: "wedding_venue",
		CityID: "city_los_angeles_ca", State: "CA", City: "Los Angeles", Verified: "2026-07-31",
		// ponytail: KnotURL placeholder — verify on theknot.com before production use
		KnotURL: "https://www.theknot.com/marketplace/wedding-venues/the-getty-villa-1640086",
	},
	{
		Name: "San Francisco City Hall", OfficialURL: "https://sf.gov/location/city-hall",
		Handle: "sfcityhall", SourceClass: "wedding_venue",
		CityID: "city_san_francisco_ca", State: "CA", City: "San Francisco", Verified: "2026-07-31",
		TikTokHandle: "sfcityhall",
	},
	{
		Name: "Palace of Fine Arts", OfficialURL: "https://palaceoffinearts.com/",
		Handle: "palaceoffinearts", SourceClass: "wedding_venue",
		CityID: "city_san_francisco_ca", State: "CA", City: "San Francisco", Verified: "2026-07-31",
	},
	{
		Name: "Napa Valley wine country venues (public IG hub)", OfficialURL: "https://www.visitnapavalley.com/",
		Handle: "visitnapavalley", SourceClass: "wedding_venue",
		CityID: "city_san_francisco_ca", State: "CA", City: "San Francisco", Verified: "2026-07-31",
		TikTokHandle: "visitnapavalley",
		// ponytail: KnotURL placeholder — verify on theknot.com before production use
		KnotURL: "https://www.theknot.com/marketplace/wedding-venues/napa-valley-wine-country-venues-8717299",
	},
	{
		Name: "Brilliant Earth (jeweler)", OfficialURL: "https://www.brilliantearth.com/",
		Handle: "brilliantearth", SourceClass: "jeweler",
		CityID: "city_san_francisco_ca", State: "CA", City: "San Francisco", Verified: "2026-07-31",
	},
	// LA wedding planners
	{
		Name: "Feathered Arrow Studio", OfficialURL: "https://featheredarrowevents.com/",
		Handle: "featheredarrowstudio", SourceClass: "wedding_planner",
		CityID: "city_los_angeles_ca", State: "CA", City: "Los Angeles", Verified: "2026-07-31",
	},
	{
		Name: "Marshecka Weddings", OfficialURL: "https://marsheckaweddings.com/",
		Handle: "marsheckaweddings", SourceClass: "wedding_planner",
		CityID: "city_los_angeles_ca", State: "CA", City: "Los Angeles", Verified: "2026-07-31",
	},
	// LA florists
	{
		Name: "The Hidden Garden", OfficialURL: "https://hiddengardenflowers.com/",
		Handle: "hiddengardenflowers", SourceClass: "florist",
		CityID: "city_los_angeles_ca", State: "CA", City: "Los Angeles", Verified: "2026-08-01",
	},
	{
		Name: "Casa Dei Fiori", OfficialURL: "https://casadeifiori.us/",
		Handle: "casadeifiori", SourceClass: "florist",
		CityID: "city_los_angeles_ca", State: "CA", City: "Los Angeles", Verified: "2026-08-01",
	},
	{
		Name: "The Empty Vase", OfficialURL: "https://emptyvase.com/",
		Handle: "emptyvase", SourceClass: "florist",
		CityID: "city_los_angeles_ca", State: "CA", City: "Los Angeles", Verified: "2026-08-01",
	},
	// LA videographers
	{
		Name: "Lighthouse Films", OfficialURL: "https://www.lighthousefilms.com/",
		Handle: "lighthousefilms", SourceClass: "videographer",
		CityID: "city_los_angeles_ca", State: "CA", City: "Los Angeles", Verified: "2026-08-01",
	},
	{
		Name: "Siena Films", OfficialURL: "https://www.sienafilms.com/",
		Handle: "sienafilms", SourceClass: "videographer",
		CityID: "city_los_angeles_ca", State: "CA", City: "Los Angeles", Verified: "2026-08-01",
	},
	{
		Name: "Evergreen Films", OfficialURL: "https://www.evergreenfilms.com/",
		Handle: "evergreenfilms", SourceClass: "videographer",
		CityID: "city_san_francisco_ca", State: "CA", City: "San Francisco", Verified: "2026-08-01",
	},
	// LA/SF wedding cake bakeries
	{
		Name: "Sweet Lady Jane", OfficialURL: "https://www.sweetladyjane.com/",
		Handle: "sweetladyjane", SourceClass: "wedding_cake",
		CityID: "city_los_angeles_ca", State: "CA", City: "Los Angeles", Verified: "2026-08-01",
	},
	{
		Name: "Hansen's Cakes", OfficialURL: "https://www.hansenscakes.com/",
		Handle: "hansenscakes", SourceClass: "wedding_cake",
		CityID: "city_los_angeles_ca", State: "CA", City: "Los Angeles", Verified: "2026-08-01",
	},
	{
		Name: "Miette", OfficialURL: "https://www.miette.com/",
		Handle: "miette", SourceClass: "wedding_cake",
		CityID: "city_san_francisco_ca", State: "CA", City: "San Francisco", Verified: "2026-08-01",
	},
	// LA/SF bridal shops
	{
		Name: "Kinsley James Couture Bridal", OfficialURL: "https://www.kinsleyjames.com/",
		Handle: "kinsleyjames", SourceClass: "bridal_shop",
		CityID: "city_los_angeles_ca", State: "CA", City: "Los Angeles", Verified: "2026-08-01",
	},
	{
		Name: "BHLDN by Anthropologie", OfficialURL: "https://www.bhldn.com/",
		Handle: "bhldn", SourceClass: "bridal_shop",
		CityID: "city_los_angeles_ca", State: "CA", City: "Los Angeles", Verified: "2026-08-01",
	},
	// LA/SF officiants
	{
		Name: "LA Wedding Officiants", OfficialURL: "https://www.laweddingofficiants.com/",
		Handle: "laweddingofficiants", SourceClass: "officiant",
		CityID: "city_los_angeles_ca", State: "CA", City: "Los Angeles", Verified: "2026-08-01",
	},
	{
		Name: "SF Wedding Officiant", OfficialURL: "https://www.sfweddingofficiant.com/",
		Handle: "sfweddingofficiant", SourceClass: "officiant",
		CityID: "city_san_francisco_ca", State: "CA", City: "San Francisco", Verified: "2026-08-01",
	},
}

// CA government + church sources, verified 2026-08-01 via web search.
// Government: California marriage records held by county recorder/clerk.
var caGovSources = []GovSource{
	{
		CountyFIPS: "06037", // Los Angeles County
		CourtName:  "Los Angeles County Registrar-Recorder/County Clerk",
		CourtURL:   "https://lavote.gov",
		SearchURL:  "https://lavote.gov/home/recorder/marriage-records/viewing-vital-records",
		Note:       "Marriage records from 1852 to present; public index inspection available in-person, online request via VitalChek.",
	},
	{
		CountyFIPS: "06075", // San Francisco County
		CourtName:  "San Francisco County Recorder",
		CourtURL:   "https://sf.gov",
		SearchURL:  "https://sf.gov/location/city-hall",
		Note:       "Marriage certificates via SF County Recorder; request-oriented, online enumeration not available.",
	},
	{
		CountyFIPS: "06073", // San Diego County
		CourtName:  "San Diego County Recorder/County Clerk",
		CourtURL:   "https://www.sdarcc.gov",
		SearchURL:  "https://www.sdarcc.gov/content/arcc/home/divisions/recorder-clerk/birth-death-marriage-certificate/marriage-certificate.html",
		Note:       "Marriage certificate requests via Assessor/Recorder/County Clerk; in-person and mail request.",
	},
	{
		CountyFIPS: "06059", // Orange County
		CourtName:  "Orange County Clerk-Recorder",
		CourtURL:   "https://www.ocrecorder.com/",
		SearchURL:  "https://cr.occlerkrecorder.gov/RecorderWorksInternet/",
		Note:       "Online grantor/grantee index for official records since 1982; index only, no images online. Marriage records via separate vital records ordering.",
	},
	{
		CountyFIPS: "06065", // Riverside County
		CourtName:  "Riverside County Assessor-County Clerk-Recorder",
		CourtURL:   "https://www.rivcoacr.org/",
		SearchURL:  "https://webselfservice.riversideacr.com/Web/search/DOCSEARCH2805S3",
		Note:       "Online document search portal; searchable by name, document number, or date range. Marriage certificates orderable via vitalsonline.asrclkrec.com.",
	},
	{
		CountyFIPS: "06071", // San Bernardino County
		CourtName:  "San Bernardino County Recorder-Clerk",
		CourtURL:   "https://arc.sbcounty.gov/",
		SearchURL:  "https://arcselfservice.sbcounty.gov/web/",
		Note:       "Online self-service portal for official records from 1925 to present; index only (no images per GC 6254.21).",
	},
	{
		CountyFIPS: "06067", // Sacramento County
		CourtName:  "Sacramento County Clerk/Recorder",
		CourtURL:   "https://ccr.saccounty.gov/",
		SearchURL:  "https://ccr.saccounty.gov/us/en/document-recording/index.html",
		Note:       "Online index of recorded documents since 1849; search by grantor/grantee name, document type, or date range. Index only.",
	},
	{
		CountyFIPS: "06001", // Alameda County
		CourtName:  "Alameda County Auditor-Controller/Clerk-Recorder",
		CourtURL:   "https://auditor.alamedacountyca.gov/clerk-recorder/",
		SearchURL:  "https://alamedacountyca.gov/bdmecomm_app/DisplayOrdServlet?proc=vord",
		Note:       "Online ordering for birth, death, and marriage certificates. In-person index search for marriage records from 1971 to present.",
	},
	{
		CountyFIPS: "06019", // Fresno County
		CourtName:  "Fresno County Recorder",
		CourtURL:   "https://www.fresnocountyca.gov/Departments/Recorder",
		SearchURL:  "https://fresnocountyca-web.tylerhost.net",
		Note:       "Online official records search via Tyler Technologies; index for documents recorded since 1981.",
	},
	{
		CountyFIPS: "06013", // Contra Costa County
		CourtName:  "Contra Costa County Clerk-Recorder",
		CourtURL:   "https://www.contracostavote.gov/",
		SearchURL:  "https://crsecurepayment.com/RW/?ln=en",
		Note:       "RecorderWorks online portal for official records index from 1986 to present; searchable by name, document number, or book/page.",
	},
	{
		CountyFIPS: "06111", // Ventura County
		CourtName:  "Ventura County Clerk-Recorder & Registrar of Voters",
		CourtURL:   "https://clerkrecorder.venturacounty.gov/",
		SearchURL:  "https://clerkrecorderselfservice.ventura.org/web/user/disclaimer",
		Note:       "Self-service official records search; searchable by name or document number. Index only, no images remotely.",
	},
	{
		CountyFIPS: "06081", // San Mateo County
		CourtName:  "San Mateo County Assessor-County Clerk-Recorder & Elections",
		CourtURL:   "https://smcacre.gov/",
		SearchURL:  "https://apps.smcacre.org/recorderworks/",
		Note:       "RecorderWorks online grantor/grantee index for documents recorded since 1985; searchable by name, document number, date range, or parcel.",
	},
}

var caDioceses = []DioceseDef{
	{Slug: "los_angeles", Name: "Archdiocese of Los Angeles", Type: "archdiocese",
		Website: "https://lacatholics.org", Directory: "https://lacatholics.org/find/", HubCityID: "city_los_angeles_ca"},
	{Slug: "san_francisco", Name: "Archdiocese of San Francisco", Type: "archdiocese",
		Website: "https://www.sfarch.org", Directory: "https://www.sfarch.org/parishes", HubCityID: "city_san_francisco_ca"},
	{Slug: "fresno", Name: "Diocese of Fresno", Type: "diocese",
		Website: "https://www.dioceseoffresno.org", Directory: "https://www.dioceseoffresno.org/parishes"},
	{Slug: "monterey", Name: "Diocese of Monterey", Type: "diocese",
		Website: "https://www.diocesemonterey.org", Directory: "https://www.diocesemonterey.org/parishes"},
	{Slug: "oakland", Name: "Diocese of Oakland", Type: "diocese",
		Website: "https://www.oakdiocese.org", Directory: "https://www.oakdiocese.org/parishes"},
	{Slug: "orange", Name: "Diocese of Orange", Type: "diocese",
		Website: "https://www.rcbo.org", Directory: "https://www.rcbo.org/parish-directory"},
	{Slug: "sacramento", Name: "Diocese of Sacramento", Type: "diocese",
		Website: "https://www.scd.org", Directory: "https://www.scd.org/parishes"},
	{Slug: "san_bernardino", Name: "Diocese of San Bernardino", Type: "diocese",
		Website: "https://www.sbscca.org", Directory: "https://www.sbscca.org/parishes"},
	{Slug: "san_diego", Name: "Diocese of San Diego", Type: "diocese",
		Website: "https://sdcatholic.org", Directory: "https://sdcatholic.org/parishes"},
	{Slug: "san_jose", Name: "Diocese of San Jose", Type: "diocese",
		Website: "https://www.dsj.org", Directory: "https://www.dsj.org/parishes"},
	{Slug: "santa_rosa", Name: "Diocese of Santa Rosa", Type: "diocese",
		Website: "https://www.srdiocese.org", Directory: "https://www.srdiocese.org/parishes"},
	{Slug: "stockton", Name: "Diocese of Stockton", Type: "diocese",
		Website: "https://www.dioceseofstockton.org", Directory: "https://www.dioceseofstockton.org/parishes"},
}

// CA parishes in the Archdioceses of Los Angeles and San Francisco.
// Names + addresses verified from each parish's own website (or the
// archdiocese parish directory). Bulletin URLs verified by direct fetch of
// each parish's bulletin archive page.
var caParishes = []ParishDef{
	{DioceseSlug: "los_angeles", Name: "Cathedral of Our Lady of the Angels", GeoLat: 34.0580, GeoLng: -118.2456,
		Address: "555 W Temple St, Los Angeles, CA 90012"},
	{DioceseSlug: "los_angeles", Name: "Cathedral Chapel of St. Vibiana", GeoLat: 34.0584, GeoLng: -118.3466,
		Address:     "923 S La Brea Ave, Los Angeles, CA 90036",
		BulletinURL: "https://cathedralchapel.org/sunday-bulletin"},
	{DioceseSlug: "los_angeles", Name: "St. Emydius Catholic Church", GeoLat: 33.9238, GeoLng: -118.2014,
		Address:     "10900 California Ave, Lynwood, CA 90262",
		BulletinURL: "https://www.saintemydius.org/parish-bulletin.html"},
	{DioceseSlug: "los_angeles", Name: "St. Monica Catholic Community", GeoLat: 34.0231, GeoLng: -118.4972,
		Address:     "725 California Ave, Santa Monica, CA 90403",
		BulletinURL: "https://stmonica.net/church-bulletin"},
	{DioceseSlug: "los_angeles", Name: "St. Charles Borromeo Church", GeoLat: 34.1500, GeoLng: -118.3662,
		Address: "10800 Moorpark St, North Hollywood, CA 91602"},
	{DioceseSlug: "los_angeles", Name: "St. Brendan Catholic Church", GeoLat: 34.0687, GeoLng: -118.3147,
		Address:     "310 S Van Ness Ave, Los Angeles, CA 90020",
		BulletinURL: "https://stbrendanla.org/bulletins"},
	{DioceseSlug: "san_francisco", Name: "Cathedral of St. Mary of the Assumption", GeoLat: 37.7842, GeoLng: -122.4254,
		Address: "1111 Gough St, San Francisco, CA 94109"},
	{DioceseSlug: "san_francisco", Name: "Old St. Mary's Cathedral & Chinese Mission", GeoLat: 37.7927, GeoLng: -122.4058,
		Address:     "660 California St, San Francisco, CA 94108",
		BulletinURL: "https://www.osmsf.org/bulletins"},
	{DioceseSlug: "san_francisco", Name: "Saints Peter and Paul Church", GeoLat: 37.8014, GeoLng: -122.4105,
		Address:     "666 Filbert St, San Francisco, CA 94133",
		BulletinURL: "https://www.salesiansspp.org/bulletins"},
	{DioceseSlug: "san_francisco", Name: "St. Dominic's Catholic Church", GeoLat: 37.7869, GeoLng: -122.4356,
		Address: "2390 Bush St, San Francisco, CA 94115"},
}

var caPack = StatePack{
	State:      "CA",
	Cities:     caCities,
	Government: caGovSources,
	Dioceses:   caDioceses,
	Parishes:   caParishes,
	Vendors:    caVendors,
}
