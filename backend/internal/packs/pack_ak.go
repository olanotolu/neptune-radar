package packs

// Alaska source pack — verified 2026-08-01.
//
// Government: Alaska is a state-administered vital-records system. Marriage
// records are held by the state's Health Analytics & Vital Records (HAVRS),
// NOT by boroughs. All boroughs redirect to the same state office. Historical
// marriage-license applications (open to the public) are indexed in the Alaska
// State Archives vital-statistics spreadsheet, scanned by FamilySearch.
//
// Church: Archdiocese of Anchorage-Juneau (aoaj.org) + Diocese of Fairbanks
// (dioceseoffairbanks.org) verified via USCCB + each diocese's own website.
// Anchorage-area parishes verified against the archdiocese parish-finder and
// each parish's own website. Bulletin URLs verified by direct discovery.
//
// Social: Instagram handles verified from each business's own public website
// social links. Verification date recorded per vendor.

var akPack = StatePack{
	State: "AK",

	// --- Cities ----------------------------------------------------------
	Cities: []CityDef{
		{ID: "city_anchorage_ak", State: "AK", County: "02020", Name: "Anchorage",
			Lat: 61.2181, Lng: -149.9003, Markets: []string{"anchorage", "ak", "alaska", "anc"}},
	},

	// --- Government (state-administered vital records) -------------------
	// Alaska marriage records are centralized at the state level. Each borough
	// entry below links to the state HAVRS portal; the SearchURL points at the
	// Alaska State Archives vital-statistics spreadsheet (public historical
	// marriage-license applications) or the borough clerk where applicable.
	Government: []GovSource{
		{
			// Anchorage Borough — state HAVRS Anchorage office.
			CountyFIPS: "02020",
			CourtName:  "Alaska HAVRS — Anchorage Office",
			CourtURL:   "https://health.alaska.gov/en/services/vital-records-orders/",
			SearchURL:  "https://archives.alaska.gov/documents/indexes/vital-stats-index.xlsx",
			Note:       "State-administered; Anchorage office at 3901 Old Seward Hwy. Historical marriage-license applications in Archives spreadsheet.",
		},
		{
			// Fairbanks North Star Borough — state HAVRS; courts issue licenses.
			CountyFIPS: "02090",
			CourtName:  "Alaska HAVRS — Fairbanks Court",
			CourtURL:   "https://health.alaska.gov/en/services/marriage-license/",
			SearchURL:  "https://archives.alaska.gov/documents/indexes/vital-stats-index.xlsx",
			Note:       "State-administered; Fairbanks court issues licenses. Historical records in Archives spreadsheet.",
		},
		{
			// Matanuska-Susitna Borough — state HAVRS; courts issue licenses.
			CountyFIPS: "02170",
			CourtName:  "Alaska HAVRS — Palmer/Wasilla Court",
			CourtURL:   "https://health.alaska.gov/en/services/marriage-license/",
			SearchURL:  "https://archives.alaska.gov/documents/indexes/vital-stats-index.xlsx",
			Note:       "State-administered; Palmer/Wasilla courts issue licenses. Historical records in Archives spreadsheet.",
		},
		{
			// Juneau (City and Borough) — state HAVRS Juneau office.
			CountyFIPS: "02110",
			CourtName:  "Alaska HAVRS — Juneau Office",
			CourtURL:   "https://health.alaska.gov/en/services/vital-records-orders/",
			SearchURL:  "https://archives.alaska.gov/documents/indexes/vital-stats-index.xlsx",
			Note:       "State-administered; Juneau office at 5441 Commercial Blvd. Historical marriage-license applications in Archives spreadsheet.",
		},
		{
			// Kenai Peninsula Borough — state HAVRS; borough clerk redirects.
			CountyFIPS: "02122",
			CourtName:  "Kenai Peninsula Borough Clerk",
			CourtURL:   "https://www.kpb.us/local-governance-and-permitting/borough-information",
			SearchURL:  "https://archives.alaska.gov/documents/indexes/vital-stats-index.xlsx",
			Note:       "Borough clerk redirects to state HAVRS for marriage records. Historical records in Archives spreadsheet.",
		},
		{
			// Kodiak Island Borough — state HAVRS; borough clerk redirects.
			CountyFIPS: "02150",
			CourtName:  "Kodiak Island Borough Clerk",
			CourtURL:   "https://www.kodiakak.us/288/Records-and-Public-Information",
			SearchURL:  "https://archives.alaska.gov/documents/indexes/vital-stats-index.xlsx",
			Note:       "Borough clerk redirects to state HAVRS for marriage records. Historical records in Archives spreadsheet.",
		},
	},

	// --- Church (Catholic dioceses + parishes) ---------------------------
	Dioceses: []DioceseDef{
		{Slug: "anchorage_juneau", Name: "Archdiocese of Anchorage-Juneau", Type: "archdiocese",
			Website: "https://aoaj.org", Directory: "https://aoaj.org/parishfinder", HubCityID: "city_anchorage_ak"},
		{Slug: "fairbanks", Name: "Diocese of Fairbanks", Type: "diocese",
			Website: "https://dioceseoffairbanks.org", Directory: "https://dioceseoffairbanks.org/parish-profiles"},
	},

	// Anchorage-area parishes in the Archdiocese of Anchorage-Juneau.
	// Names and addresses verified against the archdiocese parish-finder
	// (aoaj.org/parishfinder) and each parish's own website. Bulletin URLs
	// verified by direct discovery.
	Parishes: []ParishDef{
		{
			DioceseSlug: "anchorage_juneau", Name: "Cathedral of Our Lady of Guadalupe",
			Address:     "3900 Wisconsin St, Anchorage, AK 99517",
			BulletinURL: "https://olgak.org/bulletin",
		},
		{
			DioceseSlug: "anchorage_juneau", Name: "Holy Family Parish",
			Address:     "800 W 5th Ave, Anchorage, AK 99501",
			BulletinURL: "https://holyfamilyalaska.org/bulletins",
		},
		{
			DioceseSlug: "anchorage_juneau", Name: "Holy Cross Catholic Church",
			Address:     "2627 Lore Rd, Anchorage, AK 99507",
			BulletinURL: "https://www.holycrossalaska.net/bulletin",
		},
		{
			DioceseSlug: "anchorage_juneau", Name: "St. Anthony Catholic Church",
			Address:     "825 S Klevin St, Anchorage, AK 99508",
			BulletinURL: "https://stanthonyak.org/bulletins",
		},
		{
			DioceseSlug: "anchorage_juneau", Name: "St. Benedict Catholic Church",
			Address: "8110 Jewel Lake Rd, Anchorage, AK 99502",
		},
		{
			DioceseSlug: "anchorage_juneau", Name: "St. Patrick Catholic Church",
			Address: "2111 Muldoon Rd, Anchorage, AK 99504",
		},
		{
			DioceseSlug: "anchorage_juneau", Name: "St. Elizabeth Ann Seton Catholic Church",
			Address: "2901 Huffman Rd, Anchorage, AK 99516",
		},
	},

	// --- Social (wedding-industry Instagram vendors) ---------------------
	Vendors: []VendorDef{
		// Anchorage photographers
		{
			Name: "Emelia K Photography", OfficialURL: "https://emeliakphotography.com/",
			Handle: "emeliakphoto", SourceClass: "engagement_photographer",
			CityID: "city_anchorage_ak", State: "AK", City: "Anchorage", Verified: "2026-08-01",
			TikTokHandle: "emeliakphoto",
		},
		{
			Name: "Lace & Fern Creative", OfficialURL: "https://laceandfern.com/",
			Handle: "laceandfern", SourceClass: "engagement_photographer",
			CityID: "city_anchorage_ak", State: "AK", City: "Anchorage", Verified: "2026-08-01",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-photographers/lace-fern-creative-9253949",
		},
		{
			Name: "Erica Rose Photography", OfficialURL: "https://erosephoto.com/",
			Handle: "erosephoto", SourceClass: "engagement_photographer",
			CityID: "city_anchorage_ak", State: "AK", City: "Anchorage", Verified: "2026-08-01",
			TikTokHandle: "erosephoto",
		},
		// Anchorage venues
		{
			Name: "The Wildbirch Hotel", OfficialURL: "https://wildbirchhotel.com/",
			Handle: "wildbirchhotel", SourceClass: "wedding_venue",
			CityID: "city_anchorage_ak", State: "AK", City: "Anchorage", Verified: "2026-08-01",
		},
		{
			Name: "Ptarmigan Roost", OfficialURL: "https://ptarmiganroost.com/",
			Handle: "ptarmiganroost", SourceClass: "wedding_venue",
			CityID: "city_anchorage_ak", State: "AK", City: "Anchorage", Verified: "2026-08-01",
			TikTokHandle: "ptarmanroost",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-venues/ptarmigan-roost-8454337",
		},
		// Anchorage jeweler
		{
			Name: "5th Avenue Jewelers", OfficialURL: "https://www.akdiamondco.com/",
			Handle: "akdiamondco", SourceClass: "jeweler",
			CityID: "city_anchorage_ak", State: "AK", City: "Anchorage", Verified: "2026-08-01",
		},
	},
}
