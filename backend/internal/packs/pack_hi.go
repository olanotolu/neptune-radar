package packs

// Hawaii source pack — verified 2026-08-01.
//
// Government: Hawaii is unique — marriage records are held by the state
// Department of Health (not county clerks). The Electronic Marriage
// Registration System (EMRS) at emrs.ehawaii.gov is the license search
// portal; vitrec.ehawaii.gov is the certificate ordering portal. Each
// county has a District Health Office that assists with vital records.
// Kalawao County (former Kalaupapa settlement, ~80 residents) has no
// separate office and routes through the state DOH.
//
// Church: Diocese of Honolulu verified via USCCB + catholichawaii.org.
// Parish names/addresses verified from the diocese directory, gcatholic.org,
// and each parish's own website. Bulletin URLs verified by direct search.
//
// Social: Instagram handles verified from each business's own public website
// social links (or from IG search results where the site is JS-rendered and
// the handle was visible in the search snippet). Verification date recorded
// per vendor.

var hiPack = StatePack{
	State: "HI",

	// --- Cities ----------------------------------------------------------
	Cities: []CityDef{
	{
		ID: "city_maui_hi", State: "HI", County: "15009", Name: "Maui",
		Lat: 20.8893, Lng: -156.4739,
		Markets: []string{"maui", "lahaina", "kahului", "wailea"},
	},
		{ID: "city_honolulu_hi", State: "HI", County: "15003", Name: "Honolulu",
			Lat: 21.3069, Lng: -157.8583, Markets: []string{"honolulu", "hi", "hawaii", "oahu"}},
	},

	// --- Government (state DOH marriage-record searches) -----------------
	// Hawaii centralizes marriage records at the state Dept of Health.
	// The EMRS portal searches marriage license applications; the vitrec
	// portal orders certified copies. District Health Offices assist
	// neighbor-island residents with pickup.
	Government: []GovSource{
		{
			// Honolulu County (Oahu) — state DOH main office at
			// 1250 Punchbowl Street. EMRS is the license search portal.
			CountyFIPS: "15003",
			CourtName:  "Hawaii Dept of Health — Vital Records (Honolulu)",
			CourtURL:   "https://health.hawaii.gov/vitalrecords/marriage-licenses/",
			SearchURL:  "https://emrs.ehawaii.gov/emrs/public/application-search.html?tab=app",
			Note:       "EMRS marriage license application search; statewide portal hosted on Oahu.",
		},
		{
			// Hawaii County (Big Island) — Hawaii District Health Office
			// with Hilo and Kamuela locations; orders via vitrec portal.
			CountyFIPS: "15001",
			CourtName:  "Hawaii District Health Office — Vital Records",
			CourtURL:   "https://health.hawaii.gov/big-island/home/vital-statistics/",
			SearchURL:  "https://vitrec.ehawaii.gov/vitalrecords/",
			Note:       "Big Island district office; certified copies ordered via statewide vitrec portal.",
		},
		{
			// Maui County (Maui, Molokai, Lanai) — Maui District Health
			// Office in Wailuku; orders via vitrec portal.
			CountyFIPS: "15009",
			CourtName:  "Maui District Health Office — Vital Records",
			CourtURL:   "https://health.hawaii.gov/maui/vital-records/",
			SearchURL:  "https://vitrec.ehawaii.gov/vitalrecords/",
			Note:       "Maui district office covers Maui, Molokai, Lanai; certified copies via vitrec portal.",
		},
		{
			// Kauai County — Kauai District Health Office in Lihue;
			// orders via vitrec portal.
			CountyFIPS: "15007",
			CourtName:  "Kauai District Health Office — Vital Records",
			CourtURL:   "https://health.hawaii.gov/kauai/vital-records/",
			SearchURL:  "https://vitrec.ehawaii.gov/vitalrecords/",
			Note:       "Kauai district office; certified copies ordered via statewide vitrec portal.",
		},
		{
			// Kalawao County (Kalaupapa peninsula, Molokai) — no separate
			// office; ~80 residents. Routes through state DOH directly.
			CountyFIPS: "15005",
			CourtName:  "Hawaii Dept of Health — Vital Records (Kalawao)",
			CourtURL:   "https://health.hawaii.gov/vitalrecords/",
			SearchURL:  "https://vitrec.ehawaii.gov/vitalrecords/",
			Note:       "Kalawao County has no district office; records handled by state DOH in Honolulu.",
		},
	},

	// --- Church (Catholic dioceses + parishes) ---------------------------
	Dioceses: []DioceseDef{
		{Slug: "honolulu", Name: "Diocese of Honolulu", Type: "diocese",
			Website: "https://www.catholichawaii.org", Directory: "https://www.catholichawaii.org/parishes", HubCityID: "city_honolulu_hi"},
	},

	// Honolulu-area parishes in the Diocese of Honolulu.
	// Names/addresses verified from the diocese directory, gcatholic.org,
	// and each parish's own website. Bulletin URLs verified by direct
	// search for each parish's bulletin archive.
	Parishes: []ParishDef{
		{
			DioceseSlug: "honolulu", Name: "Cathedral Basilica of Our Lady of Peace",
			Address:     "1184 Bishop St, Honolulu, HI 96813",
			BulletinURL: "https://honolulucathedral.org/2022-weekly-bulletin/",
		},
		{
			DioceseSlug: "honolulu", Name: "Co-Cathedral of St. Theresa of the Child Jesus",
			Address:     "712 N School St, Honolulu, HI 96817",
			BulletinURL: "https://www.cocathedral.org/bulletin",
		},
		{
			DioceseSlug: "honolulu", Name: "Sacred Heart Catholic Church",
			Address:     "1701 Wilder Ave, Honolulu, HI 96822",
			BulletinURL: "https://www.sacredhearthnl.org/bulletin",
		},
		{
			DioceseSlug: "honolulu", Name: "Holy Family Catholic Church",
			Address:     "830 Main St, Honolulu, HI 96818",
			BulletinURL: "https://www.holyfamilyhonolulu.org/",
		},
		{
			DioceseSlug: "honolulu", Name: "St. Pius X Catholic Church, Manoa",
			Address:     "2821 Lowrey Ave, Honolulu, HI 96822",
			BulletinURL: "https://stpiusxmanoa.com/",
		},
		{
			DioceseSlug: "honolulu", Name: "St. Augustine by the Sea Catholic Church",
			Address: "130 Ohua Ave, Honolulu, HI 96815",
		},
		{
			DioceseSlug: "honolulu", Name: "St. Anthony of Padua Catholic Church",
			Address: "148A Makawao St, Kailua, HI 96734",
		},
		{
			DioceseSlug: "honolulu", Name: "St. John Vianney Catholic Church",
			Address:     "920 Keolu Dr, Kailua, HI 96734",
			BulletinURL: "https://www.saintjohnvianneyhawaii.org/sunday-bulletin.html",
		},
	},

	// --- Social (wedding-industry Instagram vendors) ---------------------
	Vendors: []VendorDef{
		// Honolulu photographers
		{
			Name: "Megan Moura Photography", OfficialURL: "https://meganmoura.com/",
			Handle: "megan_moura", SourceClass: "engagement_photographer",
			CityID: "city_honolulu_hi", State: "HI", City: "Honolulu", Verified: "2026-08-01",
			TikTokHandle: "megan_moura",
		},
		{
			Name: "Keoni Michael Photography", OfficialURL: "https://keonimichael.com/",
			Handle: "keoni_michael", SourceClass: "engagement_photographer",
			CityID: "city_honolulu_hi", State: "HI", City: "Honolulu", Verified: "2026-08-01",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-photographers/keoni-michael-photography-3040051",
		},
		// Honolulu venues
		{
			Name: "Sunset Ranch Hawaii", OfficialURL: "https://www.sunsetranchhawaii.com/",
			Handle: "sunsetranchhawaii", SourceClass: "wedding_venue",
			CityID: "city_honolulu_hi", State: "HI", City: "Honolulu", Verified: "2026-08-01",
			TikTokHandle: "sunsetranchhawaii",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-venues/sunset-ranch-hawaii-3091915",
		},
		{
			Name: "Loulu Palm Estate", OfficialURL: "https://loulupalm.com/",
			Handle: "loulupalm", SourceClass: "wedding_venue",
			CityID: "city_honolulu_hi", State: "HI", City: "Honolulu", Verified: "2026-08-01",
		},
		// Honolulu jewelers
		{
			Name: "Diamond Guy Hawaii", OfficialURL: "https://diamondguyhawaii.com/",
			Handle: "diamondguy808", SourceClass: "jeweler",
			CityID: "city_honolulu_hi", State: "HI", City: "Honolulu", Verified: "2026-08-01",
		},
		{
			Name: "The Wedding Ring Shop", OfficialURL: "https://www.weddingringshop.com/",
			Handle: "weddingringshop", SourceClass: "jeweler",
			CityID: "city_honolulu_hi", State: "HI", City: "Honolulu", Verified: "2026-08-01",
		},
		{
			Name: "Jenny Vargas Photography", OfficialURL: "https://jennyvargasphotography.com/",
			Handle: "jennyvargasphotography", SourceClass: "engagement_photographer",
			CityID: "city_maui_hi", State: "HI", City: "Maui", Verified: "2026-08-03",
		},
		{
			Name: "Tara Lee Photography", OfficialURL: "https://taraleephotography.net/",
			Handle: "taralee.photo", SourceClass: "engagement_photographer",
			CityID: "city_maui_hi", State: "HI", City: "Maui", Verified: "2026-08-03",
		},
		{
			Name: "Kylene Morgan Photography", OfficialURL: "https://kylenemorganphotography.com/",
			Handle: "kylenemorganphotography", SourceClass: "engagement_photographer",
			CityID: "city_honolulu_hi", State: "HI", City: "Honolulu", Verified: "2026-08-03",
		},
	},
}
