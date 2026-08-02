package packs

// Massachusetts source pack — verified 2026-08-02.
//
// Government: Massachusetts vital records are decentralized — marriage records
// are held by the city/town clerk where the marriage intention was filed, not
// at the county level. The top 7 counties by population are represented by
// their largest city's clerk office (Suffolk→Boston, Middlesex→Cambridge,
// Worcester→Worcester, Essex→Salem, Norfolk→Quincy, Bristol→Taunton,
// Plymouth→Plymouth). The statewide Registry of Vital Records and Statistics
// (RVRS) is also included. All URLs verified against each municipality's
// official .gov site.
//
// Church: all 4 Massachusetts Catholic dioceses/archdioceses verified via
// USCCB + each diocese's own website. Boston-area parishes (Archdiocese of
// Boston) were verified against the Boston Catholic Directory + each parish's
// own website. Bulletin-archive URLs verified by direct fetch.
//
// Social: Instagram handles verified from each business's own public website
// social links (or from the IG profile page where the site is JS-rendered).
// Verification date recorded per vendor.

var maPack = StatePack{
	State: "MA",

	// --- Cities ----------------------------------------------------------
	Cities: []CityDef{
		{ID: "city_boston_ma", State: "MA", County: "25025", Name: "Boston",
			Lat: 42.3601, Lng: -71.0589, Markets: []string{"boston", "bos", "suffolk", "mass"}},
		{ID: "city_worcester_ma", State: "MA", County: "25027", Name: "Worcester",
			Lat: 42.2626, Lng: -71.8023, Markets: []string{"worcester", "worcestercounty"}},
	},

	// --- Government (city/town clerk marriage-record offices) ------------
	// MA has no county-level clerk for vital records; each city/town clerk
	// holds marriage records for licenses filed in that municipality.
	Government: []GovSource{
		{
			// Suffolk County (Boston) — Registry Division issues marriage
			// certificates and licenses; online ordering via registry.boston.gov.
			CountyFIPS: "25025",
			CourtName:  "Boston Registry Division (City Clerk)",
			CourtURL:   "https://www.boston.gov/departments/registry-birth-death-and-marriage",
			SearchURL:  "https://registry.boston.gov/marriage",
			Note:       "Online marriage certificate request portal; request-oriented, no public browse.",
		},
		{
			// Middlesex County — Cambridge City Clerk (largest city in the
			// county). Marriage certificates for licenses filed in Cambridge.
			CountyFIPS: "25017",
			CourtName:  "Cambridge City Clerk",
			CourtURL:   "https://www.cambridgema.gov/Departments/cityclerksoffice",
			SearchURL:  "https://www.cambridgema.gov/iwantto/orderacertifiedcopyofamarriagecertificate",
			Note:       "Online/mail/in-person marriage certificate requests; no public browse.",
		},
		{
			// Worcester County — Worcester City Clerk. Marriage certificates
			// for licenses filed in Worcester; records from 1848 to present.
			CountyFIPS: "25027",
			CourtName:  "Worcester City Clerk",
			CourtURL:   "https://www.worcesterma.gov/city-clerk",
			SearchURL:  "https://www.worcesterma.gov/city-clerk/certificates-licenses/marriage-certificates",
			Note:       "Online ordering via payment partner; request-oriented, no public browse.",
		},
		{
			// Essex County — Salem City Clerk (county seat). Marriage licenses
			// and records for licenses filed in Salem.
			CountyFIPS: "25009",
			CourtName:  "Salem City Clerk",
			CourtURL:   "https://salemma.gov/181/City-Clerk-Elections",
			SearchURL:  "https://salemma.gov/376/Marriage-Licenses-Records",
			Note:       "Marriage licenses & records page; online certificate ordering via UniPay; no public browse.",
		},
		{
			// Norfolk County — Quincy City Clerk (largest city in the county).
			// Vital statistics from 1792 to present.
			CountyFIPS: "25021",
			CourtName:  "Quincy City Clerk",
			CourtURL:   "https://www.quincyma.gov/departments/city_clerk/index.php",
			SearchURL:  "https://www.quincyma.gov/departments/city_clerk/births_deaths_and_marriages.php",
			Note:       "Vital statistics page with online request form; request-oriented, no public browse.",
		},
		{
			// Bristol County — Taunton City Clerk (county seat). Marriage
			// records available in person, by mail, or online via City Hall
			// Systems.
			CountyFIPS: "25005",
			CourtName:  "Taunton City Clerk",
			CourtURL:   "https://www.taunton-ma.gov/183/City-Clerk",
			SearchURL:  "https://www.taunton-ma.gov/184/Birth-Death-Marriage-Records",
			Note:       "Vital records page with online ordering via epay.cityhallsystems.com; no public browse.",
		},
		{
			// Plymouth County — Plymouth Town Clerk (county seat). Marriage
			// certificates for licenses filed in Plymouth; online ordering via
			// City Hall Systems.
			CountyFIPS: "25023",
			CourtName:  "Plymouth Town Clerk",
			CourtURL:   "https://plymouth-ma.gov/549/Town-Clerk",
			SearchURL:  "https://www.plymouth-ma.gov/750/Vital-Records-Births-Deaths-Marriages",
			Note:       "Vital records page with online ordering via epay.cityhallsystems.com; no public browse.",
		},
	},

	// --- Church (Catholic dioceses + parishes) ---------------------------
	Dioceses: []DioceseDef{
		{Slug: "boston", Name: "Archdiocese of Boston", Type: "archdiocese",
			Website: "https://www.bostoncatholic.org", Directory: "https://www.bostoncatholic.org/parishes", HubCityID: "city_boston_ma"},
		{Slug: "worcester", Name: "Diocese of Worcester", Type: "diocese",
			Website: "https://www.worcesterdiocese.org", Directory: "https://www.worcesterdiocese.org/parishes", HubCityID: "city_worcester_ma"},
		{Slug: "springfield", Name: "Diocese of Springfield", Type: "diocese",
			Website: "https://www.diospringfield.org", Directory: "https://www.diospringfield.org/parishes"},
		{Slug: "fall_river", Name: "Diocese of Fall River", Type: "diocese",
			Website: "https://www.fallriverdiocese.org", Directory: "https://www.fallriverdiocese.org/parishes"},
	},

	// Boston-area parishes in the Archdiocese of Boston. Names and addresses
	// verified against the Boston Catholic Directory (thebostonpilot.com/bcd)
	// and each parish's own website. Bulletin URLs verified by direct fetch.
	Parishes: []ParishDef{
		{
			DioceseSlug: "boston", Name: "Saint Cecilia Parish",
			Address:     "18 Belvidere Street, Boston, MA 02115",
			BulletinURL: "https://www.stceciliaboston.org/weekly-bulletin/",
		},
		{
			DioceseSlug: "boston", Name: "St. Theresa of Avila Parish",
			Address:     "2078 Centre Street, West Roxbury, MA 02132",
			BulletinURL: "https://www.sttheresaparishboston.com/weekly-bulletin",
		},
		{
			DioceseSlug: "boston", Name: "St. Martin de Porres Parish",
			Address:     "243 Neponset Avenue, Dorchester, MA 02122",
			BulletinURL: "https://stmartindeporresparishdorchester.org/bulletins",
		},
		{
			DioceseSlug: "boston", Name: "St. Ann Parish",
			Address:     "399 Medford Street, Somerville, MA 02145",
			BulletinURL: "https://saintann-somerville.org/bulletins",
		},
		{
			DioceseSlug: "boston", Name: "Sacred Heart Parish",
			Address:     "169 Cummins Highway, Roslindale, MA 02131",
			BulletinURL: "https://www.sh-roslindale.org/bulletin/old",
		},
		{
			DioceseSlug: "boston", Name: "Saint Julia Parish",
			Address:     "374 Boston Post Road, Weston, MA 02493",
			BulletinURL: "https://stjulia.org/livestream-archive-masses/",
		},
		{
			DioceseSlug: "boston", Name: "St. Peter Lithuanian Parish",
			Address:     "75 Flaherty Way, South Boston, MA 02127",
			BulletinURL: "http://www.stpeterlithuanianparish.org/parish_bulletins.htm",
		},
	},

	// --- Social (wedding-industry Instagram vendors) ---------------------
	Vendors: []VendorDef{
		// Boston engagement/wedding photographers
		{
			Name: "Erin of Boston Photography", OfficialURL: "https://erinofboston.com/",
			Handle: "erinofboston", SourceClass: "engagement_photographer",
			CityID: "city_boston_ma", State: "MA", City: "Boston", Verified: "2026-08-02",
			TikTokHandle: "erinofboston",
		},
		{
			Name: "Caroline Giuliano Photography", OfficialURL: "https://carolinegiulianophoto.com/",
			Handle: "carolinegiulianophoto", SourceClass: "engagement_photographer",
			CityID: "city_boston_ma", State: "MA", City: "Boston", Verified: "2026-08-02",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-photographers/caroline-giuliano-photography-7016532",
		},
		{
			Name: "Katarina Houghton Photography", OfficialURL: "https://katarinaphotography.com/",
			Handle: "katarinaphotography", SourceClass: "engagement_photographer",
			CityID: "city_boston_ma", State: "MA", City: "Boston", Verified: "2026-08-02",
			TikTokHandle: "katarinaphotography",
		},
		{
			Name: "Lena Mirisola Photography", OfficialURL: "https://lenamirisolaphoto.com/",
			Handle: "lenamirisola", SourceClass: "engagement_photographer",
			CityID: "city_boston_ma", State: "MA", City: "Boston", Verified: "2026-08-02",
		},
		{
			Name: "Stephanie Berenson Photography", OfficialURL: "https://stephanieberenson.com/",
			Handle: "stephanieberensonphotography", SourceClass: "engagement_photographer",
			CityID: "city_boston_ma", State: "MA", City: "Boston", Verified: "2026-08-02",
			TikTokHandle: "stephanieberensonphotography",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-photographers/stephanie-berenson-photography-5322157",
		},
		// Boston wedding venues
		{
			Name: "The Bradford Collection", OfficialURL: "https://www.thebradfordcollectionboston.com/",
			Handle: "thebradfordcollection", SourceClass: "wedding_venue",
			CityID: "city_boston_ma", State: "MA", City: "Boston", Verified: "2026-08-02",
			TikTokHandle: "thebradfordcollection",
		},
		{
			Name: "Seaport Hotel Boston", OfficialURL: "https://www.seaportboston.com/weddings",
			Handle: "seaportboston", SourceClass: "wedding_venue",
			CityID: "city_boston_ma", State: "MA", City: "Boston", Verified: "2026-08-02",
			TikTokHandle: "seaportboston",
		},
		{
			Name: "The Newbury Boston", OfficialURL: "https://www.thenewburyboston.com/weddings",
			Handle: "thenewburyboston", SourceClass: "wedding_venue",
			CityID: "city_boston_ma", State: "MA", City: "Boston", Verified: "2026-08-02",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-venues/the-newbury-boston-9279487",
		},
		// Boston jewelers
		{
			Name: "Bostonian Jewelers", OfficialURL: "https://bostonianjewelers.com/",
			Handle: "bostonianjewelers", SourceClass: "jeweler",
			CityID: "city_boston_ma", State: "MA", City: "Boston", Verified: "2026-08-02",
		},
		{
			Name: "Boston Diamond Company", OfficialURL: "https://bostondiamond.com/",
			Handle: "bostondiamondcompany", SourceClass: "jeweler",
			CityID: "city_boston_ma", State: "MA", City: "Boston", Verified: "2026-08-02",
		},
	},
}
