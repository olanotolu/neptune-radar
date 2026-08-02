package packs

// Oklahoma source pack — verified 2026-08-01.
//
// Government: Oklahoma marriage records are held by the Court Clerk in each
// county. Most counties route online search through the Oklahoma State Courts
// Network (OSCN) docket search, which covers all 77 counties and supports a
// "Marriage License" case-type filter. Individual county websites provide
// marriage-license info and records-request forms. SearchURLs point at the
// OSCN docket search (the central enumeration-capable endpoint) where the
// county's own site lacks an online search portal.
//
// Church: both Oklahoma dioceses verified via USCCB + each diocese's own
// website. Oklahoma City-area parishes (Archdiocese of Oklahoma City) verified
// against the archdiocese parish finder + DiscoverMass bulletin archives.
// Bulletin URLs use DiscoverMass (aggregator) where the parish's own site does
// not host a public bulletin archive.
//
// Social: Instagram handles verified from each business's own public website
// social links (or from IG search results where the site is JS-rendered and
// the handle was visible in the search snippet). Verification date recorded
// per vendor.

var okPack = StatePack{
	State: "OK",

	// --- Cities ----------------------------------------------------------
	Cities: []CityDef{
		{
			ID: "city_oklahoma_city_ok", State: "OK", County: "40109", Name: "Oklahoma City",
			Lat: 35.4676, Lng: -97.5164,
			Markets: []string{"oklahomacity", "okc", "oklahoma", "ok"},
		},
		{
			ID: "city_tulsa_ok", State: "OK", County: "40143", Name: "Tulsa",
			Lat: 36.1540, Lng: -95.9928,
			Markets: []string{"tulsa", "ok", "creekcounty"},
		},
	},

	// --- Government (court clerk marriage-record searches) ---------------
	Government: []GovSource{
		{
			// Oklahoma County — Court Clerk records request page; search
			// via OSCN docket search with Marriage License case type.
			CountyFIPS: "40109",
			CourtName:  "Oklahoma County Court Clerk",
			CourtURL:   "https://www.oklahomacounty.org/elected-offices/court-clerk/request-records",
			SearchURL:  "https://www.oscn.net/dockets/Search.aspx",
			Note:       "Records request form + OSCN docket search with Marriage License filter; enumeration candidate via OSCN.",
		},
		{
			// Tulsa County — Court Clerk marriage license division page;
			// search via OSCN docket search.
			CountyFIPS: "40143",
			CourtName:  "Tulsa County Court Clerk",
			CourtURL:   "https://courtclerk.tulsacounty.org/Home/Marriage",
			SearchURL:  "https://www.oscn.net/dockets/Search.aspx",
			Note:       "Marriage license division page; OSCN docket search with Marriage License filter; enumeration candidate via OSCN.",
		},
		{
			// Cleveland County — District Court Clerk page; historic
			// marriage index via Kofile QuickLink, recent via OSCN.
			CountyFIPS: "40027",
			CourtName:  "Cleveland County District Court Clerk",
			CourtURL:   "https://www.clevelandcountyok.com/198/District-Court-Clerk",
			SearchURL:  "https://kofilequicklinks.com/Cleveland-OK_DCC/",
			Note:       "Kofile QuickLink historic index with Marriages document type; recent records via OSCN docket search.",
		},
		{
			// Canadian County — Court Clerk page; record search via OSCN.
			CountyFIPS: "40015",
			CourtName:  "Canadian County Court Clerk",
			CourtURL:   "https://canadiancounty.org/157/Court-Clerk",
			SearchURL:  "https://www.oscn.net/dockets/Search.aspx",
			Note:       "Court Clerk page with OSCN record search instructions; Marriage License filter on OSCN; enumeration candidate.",
		},
		{
			// Comanche County — Court Clerk marriage license page; search
			// via OSCN docket search.
			CountyFIPS: "40031",
			CourtName:  "Comanche County Court Clerk",
			CourtURL:   "https://www.comanchecountyok.gov/196/Marriage-License",
			SearchURL:  "https://www.oscn.net/dockets/Search.aspx",
			Note:       "Marriage license page with records request form; OSCN docket search with Marriage License filter; enumeration candidate.",
		},
		{
			// Rogers County — Court Clerk page; search via OSCN.
			CountyFIPS: "40129",
			CourtName:  "Rogers County Court Clerk",
			CourtURL:   "https://www.rogerscounty.org/181/Court-Clerk",
			SearchURL:  "https://www.oscn.net/dockets/Search.aspx",
			Note:       "Court Clerk page links to OSCN docket search; Marriage License filter on OSCN; enumeration candidate.",
		},
		{
			// Muskogee County — Court Clerk page; search via ODCR.
			CountyFIPS: "40101",
			CourtName:  "Muskogee County Court Clerk",
			CourtURL:   "https://muskogee.okcounties.org/offices/court-clerk/",
			SearchURL:  "https://www1.odcr.com/",
			Note:       "Court Clerk page links to ODCR court records search; marriage records searchable by name; enumeration capability needs testing.",
		},
	},

	// --- Church (Catholic dioceses + parishes) ---------------------------
	Dioceses: []DioceseDef{
		{
			Slug: "oklahoma_city", Name: "Archdiocese of Oklahoma City", Type: "archdiocese",
			Website: "https://archokc.org", Directory: "https://archokc.org/parishfinder",
			HubCityID: "city_oklahoma_city_ok",
		},
		{
			Slug: "tulsa", Name: "Diocese of Tulsa", Type: "diocese",
			Website: "https://www.dioceseoftulsa.org", Directory: "https://www.dioceseoftulsa.org/parishes",
			HubCityID: "city_tulsa_ok",
		},
	},

	// Oklahoma City-area parishes in the Archdiocese of Oklahoma City.
	// Names and addresses verified from the archdiocese parish finder +
	// DiscoverMass/gcatholic listings. Bulletin URLs use DiscoverMass
	// (aggregator) where the parish's own site does not host a public
	// bulletin archive.
	Parishes: []ParishDef{
		{
			DioceseSlug: "oklahoma_city", Name: "Cathedral of Our Lady of Perpetual Help",
			Address:     "3214 N Lake Ave, Oklahoma City, OK 73118",
			BulletinURL: "https://discovermass.com/church/cathedral-of-our-lady-of-perpetual-help-oklahoma-city-ok/#bulletins",
			Aggregator:  true,
		},
		{
			DioceseSlug: "oklahoma_city", Name: "Christ the King Catholic Church",
			Address:     "8005 Dorset Dr, Oklahoma City, OK 73120",
			BulletinURL: "https://discovermass.com/church/christ-the-king-oklahoma-city-ok/#bulletins",
			Aggregator:  true,
		},
		{
			DioceseSlug: "oklahoma_city", Name: "St. Eugene Catholic Church",
			Address:     "2400 W Hefner Rd, Oklahoma City, OK 73120",
			BulletinURL: "https://discovermass.com/church/saint-eugene-oklahoma-city-ok/#bulletins",
			Aggregator:  true,
		},
		{
			DioceseSlug: "oklahoma_city", Name: "St. James the Greater Catholic Church",
			Address:     "4201 S McKinley Ave, Oklahoma City, OK 73109",
			BulletinURL: "https://discovermass.com/church/saint-james-the-greater-oklahoma-city-ok/#bulletins",
			Aggregator:  true,
		},
		{
			DioceseSlug: "oklahoma_city", Name: "St. Francis of Assisi Catholic Church",
			Address: "1901 NW 18th St, Oklahoma City, OK 73106",
		},
		{
			DioceseSlug: "oklahoma_city", Name: "St. Patrick Catholic Church",
			Address:     "2121 N Portland Ave, Oklahoma City, OK 73107",
			BulletinURL: "https://www.saintpatrickokc.org/bulletin",
		},
		{
			DioceseSlug: "oklahoma_city", Name: "Holy Angels Catholic Church",
			Address: "317 N Blackwelder Ave, Oklahoma City, OK 73106",
		},
		{
			DioceseSlug: "oklahoma_city", Name: "Epiphany of the Lord Catholic Church",
			Address: "7336 W Britton Rd, Oklahoma City, OK 73132",
		},
	},

	// --- Social (wedding-industry Instagram vendors) ---------------------
	Vendors: []VendorDef{
		// Oklahoma City photographers
		{
			Name: "Leia Smethurst Photography", OfficialURL: "https://leiaphoto.com/",
			Handle: "leiasmethurst", SourceClass: "engagement_photographer",
			CityID: "city_oklahoma_city_ok", State: "OK", City: "Oklahoma City", Verified: "2026-08-01",
			TikTokHandle: "leiasmethurst",
		},
		{
			Name: "Wild Thistle Photo", OfficialURL: "https://wildthistlephoto.com/",
			Handle: "wildthistlephoto", SourceClass: "engagement_photographer",
			CityID: "city_oklahoma_city_ok", State: "OK", City: "Oklahoma City", Verified: "2026-08-01",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-photographers/wild-thistle-photo-5048481",
		},
		{
			Name: "Caro Liz Creative", OfficialURL: "https://www.carolizcreative.com/",
			Handle: "carolineeliza.co", SourceClass: "engagement_photographer",
			CityID: "city_oklahoma_city_ok", State: "OK", City: "Oklahoma City", Verified: "2026-08-01",
			TikTokHandle: "carolineeliza.co",
		},
		// Oklahoma City venues
		{
			Name: "Coles Garden", OfficialURL: "https://colesgarden.net/",
			Handle: "colesgardenokc", SourceClass: "wedding_venue",
			CityID: "city_oklahoma_city_ok", State: "OK", City: "Oklahoma City", Verified: "2026-08-01",
			TikTokHandle: "colesgardenokc",
		},
		{
			Name: "Rose Briar Place", OfficialURL: "https://www.rosebriarplace.com/",
			Handle: "rosebriarplace", SourceClass: "wedding_venue",
			CityID: "city_oklahoma_city_ok", State: "OK", City: "Oklahoma City", Verified: "2026-08-01",
			TikTokHandle: "rosebriarplace",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-venues/rose-briar-place-3235550",
		},
		{
			Name: "36 Acres", OfficialURL: "https://36acres.net/",
			Handle: "36acresvenue", SourceClass: "wedding_venue",
			CityID: "city_oklahoma_city_ok", State: "OK", City: "Oklahoma City", Verified: "2026-08-01",
		},
		// Oklahoma City jewelers
		{
			Name: "Naifeh Fine Jewelry", OfficialURL: "https://www.naifehfinejewelry.com/",
			Handle: "naifehfinejewelry", SourceClass: "jeweler",
			CityID: "city_oklahoma_city_ok", State: "OK", City: "Oklahoma City", Verified: "2026-08-01",
		},
		{
			Name: "Huntington Fine Jewelers", OfficialURL: "https://www.huntingtonfinejewelers.com/",
			Handle: "huntingtonfinejewelers", SourceClass: "jeweler",
			CityID: "city_oklahoma_city_ok", State: "OK", City: "Oklahoma City", Verified: "2026-08-01",
		},
	},
}
