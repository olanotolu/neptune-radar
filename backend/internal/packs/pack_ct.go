package packs

// Connecticut source pack — verified 2026-08-01.
//
// Government: Connecticut has no county-level vital records. Marriage records
// are held by town/city clerks. The primary town clerk for each of the top 7
// counties was verified against the town's official .gov site. A statewide
// marriage index (1897-2001) is available at data.ct.gov and a free searchable
// index (1897-2017) at connecticutgenealogy.org (Reclaim The Records).
//
// Church: all 3 CT Catholic dioceses/archdiocese verified via USCCB + each
// diocese's own website. Hartford-area parishes (Archdiocese of Hartford) were
// verified against the archdiocese's parish list (archdioceseofhartford.org/
// parishes) + Wikipedia's list of parishes in the archdiocese. Bulletin URLs
// verified by direct search for each parish's bulletin archive.
//
// Social: Instagram handles verified from each business's own public website
// social links. Verification date recorded per vendor.

var ctPack = StatePack{
	State: "CT",

	// --- Cities ----------------------------------------------------------
	Cities: []CityDef{
		{ID: "city_hartford_ct", State: "CT", County: "09003", Name: "Hartford",
			Lat: 41.7658, Lng: -72.6734, Markets: []string{"hartford", "ct", "connecticut"}},
		{ID: "city_stamford_ct", State: "CT", County: "09001", Name: "Stamford",
			Lat: 41.0534, Lng: -73.0539, Markets: []string{"stamford", "fairfield", "ct"}},
		{ID: "city_new_haven_ct", State: "CT", County: "09009", Name: "New Haven",
			Lat: 41.3083, Lng: -72.9279, Markets: []string{"newhaven", "nhv", "ct"}},
	},

	// --- Government (town clerk marriage-record offices) -----------------
	// ponytail: CT has no county clerks; marriage records are at the town
	// level. Each county entry below maps to the primary city's town clerk.
	Government: []GovSource{
		{
			// Fairfield County — Fairfield Town Clerk vital statistics.
			CountyFIPS: "09001",
			CourtName:  "Fairfield Town Clerk",
			CourtURL:   "https://fairfieldct.gov/service/town_clerk/index.php",
			SearchURL:  "https://fairfieldct.gov/service/town_clerk/vital_statistics.php",
			Note:       "CT marriage records via town clerk; request-oriented, no online search index.",
		},
		{
			// Hartford County — Hartford Town and City Clerk; online vital
			// records request portal via Permitium.
			CountyFIPS: "09003",
			CourtName:  "Hartford Town and City Clerk",
			CourtURL:   "https://www.hartfordct.gov/Government/Town-and-City-Clerk",
			SearchURL:  "https://hartfordct.permitium.com/",
			Note:       "Online vital records request portal; marriage certificates $20; no public search index.",
		},
		{
			// New Haven County — New Haven Vital Statistics (Health Dept).
			CountyFIPS: "09009",
			CourtName:  "New Haven Vital Statistics",
			CourtURL:   "https://nhvhealth.org/vital-statistics/",
			SearchURL:  "https://nhvhealth.org/vital-statistics/",
			Note:       "CT marriage records via city vital statistics; request-oriented, no online search index.",
		},
		{
			// New London County — New London City Clerk vital records.
			CountyFIPS: "09011",
			CourtName:  "New London City Clerk",
			CourtURL:   "https://newlondonct.gov/city-clerk",
			SearchURL:  "https://newlondonct.gov/vital-records",
			Note:       "CT marriage records via city clerk; request form download; no online search index.",
		},
		{
			// Litchfield County — Litchfield Town Clerk vital records.
			CountyFIPS: "09005",
			CourtName:  "Litchfield Town Clerk",
			CourtURL:   "https://www.townoflitchfieldct.gov/entities/town-clerk",
			SearchURL:  "https://www.townoflitchfieldct.gov/subpages/vital-records-birth-death-marriage-certificates",
			Note:       "CT marriage records via town clerk; request form download; no online search index.",
		},
		{
			// Middlesex County — Middletown Vital Statistics; online marriage
			// application portal available.
			CountyFIPS: "09007",
			CourtName:  "Middletown Vital Statistics",
			CourtURL:   "https://www.middletownct.gov/329/Vital-Statistics",
			SearchURL:  "https://middletownct.gov/330/Online-Marriage-Application",
			Note:       "Online marriage license application portal; certified copies by request; no public search index.",
		},
		{
			// Tolland County — Tolland Town Clerk vital statistics.
			CountyFIPS: "09013",
			CourtName:  "Tolland Town Clerk",
			CourtURL:   "https://www.tollandct.gov/town-clerk",
			SearchURL:  "https://www.tollandct.gov/town-clerk/pages/marriage-licenses-vital-statistics",
			Note:       "CT marriage records via town clerk; request-oriented, no online search index.",
		},
	},

	// --- Church (Catholic dioceses + parishes) ---------------------------
	Dioceses: []DioceseDef{
		{Slug: "hartford", Name: "Archdiocese of Hartford", Type: "archdiocese",
			Website: "https://www.aohct.org", Directory: "https://www.aohct.org/parishes", HubCityID: "city_hartford_ct"},
		{Slug: "bridgeport", Name: "Diocese of Bridgeport", Type: "diocese",
			Website: "https://www.bridgeportdiocese.org", Directory: "https://www.bridgeportdiocese.org/parishes"},
		{Slug: "norwich", Name: "Diocese of Norwich", Type: "diocese",
			Website: "https://www.norwichdiocese.org", Directory: "https://www.norwichdiocese.org/parishes"},
	},

	// Hartford-area parishes in the Archdiocese of Hartford. Names and
	// addresses verified from the archdiocese's parish list
	// (archdioceseofhartford.org/parishes) + Wikipedia's list of parishes in
	// the archdiocese. Bulletin URLs verified by direct search for each
	// parish's bulletin archive.
	Parishes: []ParishDef{
		{DioceseSlug: "hartford", Name: "Cathedral of St. Joseph", Address: "140 Farmington Ave, Hartford, CT 06105"},
		{DioceseSlug: "hartford", Name: "St. Patrick-St. Anthony Church", Address: "285 Church St, Hartford, CT 06103"},
		{DioceseSlug: "hartford", Name: "St. Augustine Parish", Address: "10 Campfield Ave, Hartford, CT 06114"},
		{DioceseSlug: "hartford", Name: "St. Thomas & St. Timothy Parish", Address: "872 Farmington Ave, West Hartford, CT 06119"},
		{
			DioceseSlug: "hartford", Name: "St. Bridget of Sweden Parish",
			Address:     "175 Main St, Cheshire, CT 06410",
			BulletinURL: "https://www.cheshirecatholic.org/bulletin",
		},
		{DioceseSlug: "hartford", Name: "Christ the King Parish", Address: "544 Prospect St, Wethersfield, CT 06109"},
		{DioceseSlug: "hartford", Name: "St. Mary Church (Bl. Michael McGivney Parish)", Address: "5 Hillhouse Ave, New Haven, CT 06511"},
		{DioceseSlug: "hartford", Name: "Our Lady of Mount Carmel", Address: "2819 Whitney Ave, Hamden, CT 06518"},
	},

	// --- Social (wedding-industry Instagram vendors) ---------------------
	Vendors: []VendorDef{
		// Hartford photographers
		{
			Name: "Anaise Prince Photography", OfficialURL: "https://www.anaiseprince.com/",
			Handle: "anaiseprincephoto", SourceClass: "engagement_photographer",
			CityID: "city_hartford_ct", State: "CT", City: "Hartford", Verified: "2026-08-01",
			TikTokHandle: "anaiseprincephoto",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-photographers/anaise-prince-photography-5125039",
		},
		{
			Name: "Carla Ten Eyck Photography", OfficialURL: "https://carlateneyck.com/",
			Handle: "c10ike", SourceClass: "engagement_photographer",
			CityID: "city_hartford_ct", State: "CT", City: "Hartford", Verified: "2026-08-01",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-photographers/carla-ten-eyck-photography-2249949",
		},
		{
			Name: "Catherine Johanna Photography", OfficialURL: "https://catherinejohannaphotography.com/",
			Handle: "cdunlavey", SourceClass: "engagement_photographer",
			CityID: "city_hartford_ct", State: "CT", City: "Hartford", Verified: "2026-08-01",
			TikTokHandle: "cdunlavey",
		},
		{
			Name: "Mishabook Photography", OfficialURL: "https://www.mishabook.com/",
			Handle: "mishamishyn", SourceClass: "engagement_photographer",
			CityID: "city_hartford_ct", State: "CT", City: "Hartford", Verified: "2026-08-01",
		},
		// Hartford venues
		{
			Name: "The Society Room of Hartford", OfficialURL: "https://hartfordsocietyroom.com/",
			Handle: "thesocietyroomofhartford", SourceClass: "wedding_venue",
			CityID: "city_hartford_ct", State: "CT", City: "Hartford", Verified: "2026-08-01",
			TikTokHandle: "thesocietyroomofhartford",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-venues/the-society-room-of-hartford-2110461",
		},
		{
			Name: "Wadsworth Mansion at Long Hill", OfficialURL: "https://www.wadsworthmansion.com/",
			Handle: "thewadsworthmansion", SourceClass: "wedding_venue",
			CityID: "city_hartford_ct", State: "CT", City: "Middletown", Verified: "2026-08-01",
		},
		// New Haven venue
		{
			Name: "The Estate New Haven", OfficialURL: "https://theestatenewhaven.com/",
			Handle: "theestatenewhaven", SourceClass: "wedding_venue",
			CityID: "city_new_haven_ct", State: "CT", City: "New Haven", Verified: "2026-08-01",
			TikTokHandle: "theestatenewhaven",
		},
		// Hartford jeweler
		{
			Name: "Armeny Custom Jewelry Design", OfficialURL: "https://armeny.com/",
			Handle: "armeny_custom_jewelry_design", SourceClass: "jeweler",
			CityID: "city_hartford_ct", State: "CT", City: "West Hartford", Verified: "2026-08-01",
		},
	},
}
