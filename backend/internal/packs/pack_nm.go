package packs

// New Mexico source pack — verified 2026-08-01.
//
// Government: New Mexico marriage records are held by the county clerk. Search
// URLs for the top 5 counties were verified against each county's official .gov
// site or its publicsearch.us portal.
//
// Church: the Archdiocese of Santa Fe was verified via USCCB + the archdiocese's
// own website. Albuquerque-area parishes were verified against the archdiocese
// parish directory + direct bulletin-archive URL discovery. Santa Fe parish
// (Cristo Rey) verified via its own website.
//
// Social: Instagram handles verified from each business's own public website
// social links (or from IG search results where the site is JS-rendered and
// the handle was visible in the search snippet). Verification date recorded
// per vendor.

var nmPack = StatePack{
	State: "NM",

	// --- Cities ----------------------------------------------------------
	Cities: []CityDef{
		{
			ID: "city_albuquerque_nm", State: "NM", County: "35001", Name: "Albuquerque",
			Lat: 35.0844, Lng: -106.6504,
			Markets: []string{"albuquerque", "abq", "nm", "bernalillo"},
		},
		{
			ID: "city_santa_fe_nm", State: "NM", County: "35049", Name: "Santa Fe",
			Lat: 35.6870, Lng: -105.9378,
			Markets: []string{"santafe", "nm", "santa_fe_county"},
		},
	},

	// --- Government (county clerk marriage-record searches) --------------
	Government: []GovSource{
		{
			// Bernalillo County (Albuquerque) — county clerk public records
			// search portal with marriage record category.
			CountyFIPS: "35001",
			CourtName:  "Bernalillo County Clerk",
			CourtURL:   "https://www.berncoclerk.gov",
			SearchURL:  "https://www.berncoclerk.gov/recording-and-filing/public-records-search/",
			Note:       "Public records search with marriage category filter; Tylerhost web search backend; enumeration candidate.",
		},
		{
			// Santa Fe County — ClerkTrackWeb portal with PUBLIC/PUBLIC login.
			CountyFIPS: "35049",
			CourtName:  "Santa Fe County Clerk",
			CourtURL:   "https://www.santafecountynm.gov/clerk",
			SearchURL:  "https://www.santafecountynm.gov/clerk/divisions/public-records-access",
			Note:       "ClerkTrackWeb portal (login PUBLIC/PUBLIC) with marriage license section; enumeration capability needs testing.",
		},
		{
			// Doña Ana County (Las Cruces) — official record search via
			// publicsearch.us.
			CountyFIPS: "35013",
			CourtName:  "Doña Ana County Clerk",
			CourtURL:   "https://www.donaana.gov/departments/elected_officials/clerks_office/",
			SearchURL:  "https://donaana.nm.publicsearch.us/",
			Note:       "Official record search portal; marriage records searchable; enumeration capability needs testing.",
		},
		{
			// Sandoval County — self-service web search portal.
			CountyFIPS: "35043",
			CourtName:  "Sandoval County Clerk",
			CourtURL:   "https://www.sandovalcountynm.gov/countyclerk/",
			SearchURL:  "https://scclerk.sandovalcountynm.gov/web/",
			Note:       "Self-service web search portal; marriage records indexed; enumeration capability needs testing.",
		},
		{
			// Otero County — clerk's office page; no online search portal.
			CountyFIPS: "35035",
			CourtName:  "Otero County Clerk",
			CourtURL:   "https://co.otero.nm.us/225/Clerk",
			SearchURL:  "https://co.otero.nm.us/225/Clerk",
			Note:       "No online search portal; marriage records by request form or in-person only.",
		},
	},

	// --- Church (Catholic dioceses + parishes) ---------------------------
	Dioceses: []DioceseDef{
		{
			Slug: "santa_fe", Name: "Archdiocese of Santa Fe", Type: "archdiocese",
			Website: "https://www.archdiosf.org", Directory: "https://www.archdiosf.org/parishes",
			HubCityID: "city_santa_fe_nm",
		},
		{
			Slug: "gallup", Name: "Diocese of Gallup", Type: "diocese",
			Website: "https://www.dioceseofgallup.org", Directory: "https://www.dioceseofgallup.org/parishes",
		},
		{
			Slug: "las_cruces", Name: "Diocese of Las Cruces", Type: "diocese",
			Website: "https://www.dioceseoflascruces.org", Directory: "https://www.dioceseoflascruces.org/parishes",
		},
	},

	// Albuquerque-area parishes in the Archdiocese of Santa Fe.
	// Bulletin URLs verified by direct search for each parish's bulletin
	// archive. Cristo Rey in Santa Fe also verified via its own website.
	Parishes: []ParishDef{
		{
			DioceseSlug: "santa_fe", Name: "Immaculate Conception Church",
			Address:     "619 Copper Ave NW, Albuquerque, NM 87102",
			BulletinURL: "https://www.iccabq.org/icc-bulletins----archives",
		},
		{
			DioceseSlug: "santa_fe", Name: "Our Lady of Fatima Catholic Church",
			Address:     "409 Golf Ave NE, Albuquerque, NM 87112",
			BulletinURL: "https://fatimachurchabq.org/bulletins",
		},
		{
			DioceseSlug: "santa_fe", Name: "St. John XXIII Catholic Church",
			Address:     "4831 Tramway Ridge Dr NE, Albuquerque, NM 87111",
			BulletinURL: "https://www.johnxxiiicc.org/bulletin",
		},
		{
			DioceseSlug: "santa_fe", Name: "Risen Savior Catholic Church",
			Address:     "7701 Wyoming Blvd NE, Albuquerque, NM 87109",
			BulletinURL: "https://www.risensaviorcc.org/bulletin",
		},
		{
			DioceseSlug: "santa_fe", Name: "Prince of Peace Catholic Church",
			Address:     "8216 Comanche Rd NE, Albuquerque, NM 87109",
			BulletinURL: "https://popabq.org/bulletin",
		},
		{
			DioceseSlug: "santa_fe", Name: "Cristo Rey Catholic Church",
			Address:     "1420 De Vargas St, Santa Fe, NM 87501",
			BulletinURL: "https://www.cristoreyparish.org/bulletin",
		},
	},

	// --- Social (wedding-industry Instagram vendors) ---------------------
	Vendors: []VendorDef{
		// Albuquerque photographers
		{
			Name: "Alicia Lucia Photography", OfficialURL: "https://www.alicialucia.com/",
			Handle: "alicialuciaphotos", SourceClass: "engagement_photographer",
			CityID: "city_albuquerque_nm", State: "NM", City: "Albuquerque", Verified: "2026-08-01",
			TikTokHandle: "alicialuciaphotos",
		},
		{
			Name: "Carissa & Ben Photography", OfficialURL: "https://carissaandben.com/",
			Handle: "carissaandben", SourceClass: "engagement_photographer",
			CityID: "city_albuquerque_nm", State: "NM", City: "Albuquerque", Verified: "2026-08-01",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-photographers/carissa-ben-photography-2377691",
		},
		{
			Name: "Erika Durdle Photography", OfficialURL: "https://www.durdlephotography.com/",
			Handle: "durdlephotography", SourceClass: "engagement_photographer",
			CityID: "city_albuquerque_nm", State: "NM", City: "Albuquerque", Verified: "2026-08-01",
			TikTokHandle: "durdlephotography",
		},
		{
			Name: "David Jesse Photography", OfficialURL: "https://www.davidjessephoto.com/",
			Handle: "davidjessephoto", SourceClass: "engagement_photographer",
			CityID: "city_albuquerque_nm", State: "NM", City: "Albuquerque", Verified: "2026-08-01",
		},
		// Albuquerque venues
		{
			Name: "Los Poblanos Historic Inn & Organic Farm", OfficialURL: "https://lospoblanos.com/",
			Handle: "lospoblanos", SourceClass: "wedding_venue",
			CityID: "city_albuquerque_nm", State: "NM", City: "Albuquerque", Verified: "2026-08-01",
			TikTokHandle: "lospoblanos",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-venues/los-poblanos-historic-inn-organic-farm-5512311",
		},
		{
			Name: "Hotel Andaluz", OfficialURL: "https://www.hotelandaluz.com/",
			Handle: "hotelandaluzalbuquerque", SourceClass: "wedding_venue",
			CityID: "city_albuquerque_nm", State: "NM", City: "Albuquerque", Verified: "2026-08-01",
		},
		// Santa Fe photographers
		{
			Name: "Molly Morgan Photography", OfficialURL: "https://www.mollymorganphotography.com/",
			Handle: "mollymorganphotography", SourceClass: "engagement_photographer",
			CityID: "city_santa_fe_nm", State: "NM", City: "Santa Fe", Verified: "2026-08-01",
			TikTokHandle: "mollymorganphotography",
		},
		{
			Name: "Caitlin Elizabeth Photography", OfficialURL: "https://www.caitlinelizabeth.com/",
			Handle: "caitlinephoto", SourceClass: "engagement_photographer",
			CityID: "city_santa_fe_nm", State: "NM", City: "Santa Fe", Verified: "2026-08-01",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-photographers/caitlin-elizabeth-photography-7677841",
		},
	},
}
