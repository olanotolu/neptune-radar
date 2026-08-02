package packs

// Florida source pack — verified 2026-08-01 via web search.
//
// Government: Florida marriage records held by county Clerk of Court.
// Miami-Dade and Broward both have online search portals.
//
// Church: 7 Catholic dioceses/archdioceses. Websites and parish directories
// verified via USCCB and each diocese's own website.
//
// Social: Instagram handles found from search results and business websites.

var flPack = StatePack{
	State: "FL",

	Cities: []CityDef{
		{ID: "city_miami_fl", State: "FL", County: "12086", Name: "Miami",
			Lat: 25.7617, Lng: -80.1918, Markets: []string{"miami", "mfl", "southflorida", "miamidade"}},
		{ID: "city_orlando_fl", State: "FL", County: "12095", Name: "Orlando",
			Lat: 28.5383, Lng: -81.3792, Markets: []string{"orlando", "centralflorida", "ocf"}},
		{ID: "city_tampa_fl", State: "FL", County: "12057", Name: "Tampa",
			Lat: 27.9506, Lng: -82.4572, Markets: []string{"tampa", "tampabay", "hillsborough"}},
		{ID: "city_jacksonville_fl", State: "FL", County: "12031", Name: "Jacksonville",
			Lat: 30.3322, Lng: -81.6557, Markets: []string{"jacksonville", "jax", "duval"}},
	},

	Government: []GovSource{
		{CountyFIPS: "12086", CourtName: "Miami-Dade County Clerk of Courts",
			CourtURL:  "https://www2.miamidadeclerk.gov",
			SearchURL: "https://www2.miamidadeclerk.gov/MLSWeb/LicenseSearch",
			Note:      "Online Marriage License Bureau with name search; enumeration candidate."},
		{CountyFIPS: "12011", CourtName: "Broward County Clerk of Courts",
			CourtURL:  "https://www.browardclerk.org",
			SearchURL: "https://www.browardclerk.org/Web2/Marriage/LicenseSearch",
			Note:      "Marriage License System with name and license number search; enumeration candidate."},
		{CountyFIPS: "12057", CourtName: "Hillsborough County Clerk of Court",
			CourtURL:  "https://www.hillsclerk.com",
			SearchURL: "https://www.hillsclerk.com/ClerkCCC/ClerkCCC-Marriage-License",
			Note:      "Marriage license records; online search capability needs testing."},
		{CountyFIPS: "12099", CourtName: "Palm Beach County Clerk of Court",
			CourtURL:  "https://www.mypalmbeachclerk.com",
			SearchURL: "https://www.mypalmbeachclerk.com/services/official-records/marriage-records",
			Note:      "Marriage records via official records search; enumeration capability needs testing."},
		{CountyFIPS: "12095", CourtName: "Orange County Clerk of Courts",
			CourtURL:  "https://www.orangeclerkfl.org",
			SearchURL: "https://www.orangeclerkfl.org/services/marriage-license",
			Note:      "Marriage license services; online search capability needs testing."},
		{CountyFIPS: "12103", CourtName: "Pinellas County Clerk of Court",
			CourtURL:  "https://www.pinellasclerk.org",
			SearchURL: "https://www.pinellasclerk.org/Services/Marriage-License",
			Note:      "Marriage license records; online search capability needs testing."},
		{CountyFIPS: "12031", CourtName: "Duval County Clerk of Court",
			CourtURL:  "https://www.duvalclerk.com",
			SearchURL: "https://www.duvalclerk.com/Core/Marriage-License",
			Note:      "Marriage license records; online search capability needs testing."},
	},

	Dioceses: []DioceseDef{
		{Slug: "miami", Name: "Archdiocese of Miami", Type: "archdiocese",
			Website: "https://miamiarch.org", Directory: "https://miamiarch.org/parishes", HubCityID: "city_miami_fl"},
		{Slug: "orlando", Name: "Diocese of Orlando", Type: "diocese",
			Website: "https://www.orlandodiocese.org", Directory: "https://www.orlandodiocese.org/find-a-parish/", HubCityID: "city_orlando_fl"},
		{Slug: "st_petersburg", Name: "Diocese of St. Petersburg", Type: "diocese",
			Website: "https://www.dosp.org", Directory: "https://www.dosp.org/chancellor/directory/parishes/", HubCityID: "city_tampa_fl"},
		{Slug: "st_augustine", Name: "Diocese of St. Augustine", Type: "diocese",
			Website: "https://www.dosa.org", Directory: "https://www.dosa.org/parishes"},
		{Slug: "pensacola_tallahassee", Name: "Diocese of Pensacola-Tallahassee", Type: "diocese",
			Website: "https://www.ptdiocese.org", Directory: "https://www.ptdiocese.org/parishes"},
		{Slug: "venice", Name: "Diocese of Venice in Florida", Type: "diocese",
			Website: "https://www.dioceseofvenice.org", Directory: "https://www.dioceseofvenice.org/parishes"},
		{Slug: "palm_beach", Name: "Diocese of Palm Beach", Type: "diocese",
			Website: "https://www.diocesepb.org", Directory: "https://www.diocesepb.org/parishes"},
	},

	Parishes: []ParishDef{
		{DioceseSlug: "miami", Name: "St. Joseph Catholic Parish", Address: "1670 Euclid Ave, Miami Beach, FL 33139",
			BulletinURL: "https://www.saintjosephmiamibeach.com/bulletin-archive"},
		{DioceseSlug: "miami", Name: "St. Brendan Catholic Church", Address: "8725 SW 32nd St, Miami, FL 33165",
			BulletinURL: "https://www.stbrendanmiami.org/church/news-events/Archived-weekly-bulletins"},
		{DioceseSlug: "miami", Name: "St. John the Apostle Catholic Parish", Address: "Hialeah, FL",
			BulletinURL: "https://www.sjamiami.com/bulletin"},
		{DioceseSlug: "miami", Name: "Mary, Queen of the Universe", Address: "8300 Vineland Ave, Orlando, FL 32821"},
		{DioceseSlug: "miami", Name: "Basilica of St. Paul", Address: "1600 N Orange Ave, Daytona Beach, FL 32117"},
	},

	Vendors: []VendorDef{
		{Name: "Estudio Zoe", OfficialURL: "https://www.estudiozoe.com/",
			Handle: "estudiozoe", SourceClass: "engagement_photographer",
			CityID: "city_miami_fl", State: "FL", City: "Miami", Verified: "2026-08-01"},
		{Name: "Villa Turqueza", OfficialURL: "https://villaturqueza.com/",
			Handle: "villaturqueza", SourceClass: "wedding_venue",
			CityID: "city_miami_fl", State: "FL", City: "Miami", Verified: "2026-08-01"},
		{Name: "Sanctuary MiMo", OfficialURL: "https://www.sanctuarymimo.com/",
			Handle: "sanctuarymimo", SourceClass: "wedding_venue",
			CityID: "city_miami_fl", State: "FL", City: "Miami", Verified: "2026-08-01",
			KnotURL: "https://www.theknot.com/marketplace/sanctuary-mimo-miami-fl-2070584"},
		{Name: "Gran Paraiso Gardens", OfficialURL: "https://granparaisogardens.com/",
			Handle: "granparaisogardens", SourceClass: "wedding_venue",
			CityID: "city_miami_fl", State: "FL", City: "Miami", Verified: "2026-08-01",
			KnotURL: "https://www.theknot.com/marketplace/gran-paraiso-gardens-miami-fl-2083458"},
		{Name: "Villa Toscana Miami", OfficialURL: "https://www.vtmiami.com/",
			Handle: "villatoscanamiami", SourceClass: "wedding_venue",
			CityID: "city_miami_fl", State: "FL", City: "Miami", Verified: "2026-08-01"},
		// Miami wedding planners
		{Name: "Paris Miami Events", OfficialURL: "https://www.parismiamievents.com/",
			Handle: "parismiamievents", SourceClass: "wedding_planner",
			CityID: "city_miami_fl", State: "FL", City: "Miami", Verified: "2026-08-01"},
		{Name: "Masi Events", OfficialURL: "https://masievents.com/",
			Handle: "masievents", SourceClass: "wedding_planner",
			CityID: "city_miami_fl", State: "FL", City: "Miami", Verified: "2026-08-01"},
		// Orlando wedding planners
		{Name: "Plan It Events", OfficialURL: "https://www.planitcfl.com/",
			Handle: "planit_events", SourceClass: "wedding_planner",
			CityID: "city_orlando_fl", State: "FL", City: "Orlando", Verified: "2026-08-01"},
		// Florida florists
		{Name: "Maison la Fleur", OfficialURL: "https://www.maisonlafleur.com/",
			Handle: "maisonlafleur", SourceClass: "florist",
			CityID: "city_miami_fl", State: "FL", City: "Miami", Verified: "2026-08-01"},
		{Name: "Gaby Flowers", OfficialURL: "https://www.gabyflowers.com/",
			Handle: "gabyflowers", SourceClass: "florist",
			CityID: "city_miami_fl", State: "FL", City: "Miami", Verified: "2026-08-01"},
		{Name: "Kalas Events Florist", OfficialURL: "https://orlandokalasflorist.com/",
			Handle: "kalaseventsflorist", SourceClass: "florist",
			CityID: "city_orlando_fl", State: "FL", City: "Orlando", Verified: "2026-08-01"},
		// Florida videographers
		{Name: "Blue Lens Video", OfficialURL: "https://www.bluelensvideo.com/",
			Handle: "bluelensvideo", SourceClass: "videographer",
			CityID: "city_miami_fl", State: "FL", City: "Miami", Verified: "2026-08-01"},
		{Name: "Orlando Wedding Films", OfficialURL: "https://www.orlandoweddingfilms.com/",
			Handle: "orlandoweddingfilms", SourceClass: "videographer",
			CityID: "city_orlando_fl", State: "FL", City: "Orlando", Verified: "2026-08-01"},
		// Florida wedding cake bakeries
		{Name: "OV Cake Designs", OfficialURL: "https://ovcakedesigns.com/",
			Handle: "ovcakedesigns", SourceClass: "wedding_cake",
			CityID: "city_orlando_fl", State: "FL", City: "Orlando", Verified: "2026-08-01"},
		{Name: "Sweet Art by Laura", OfficialURL: "https://www.sweetartbylaura.com/",
			Handle: "sweetartbylaura", SourceClass: "wedding_cake",
			CityID: "city_miami_fl", State: "FL", City: "Miami", Verified: "2026-08-01"},
		// Florida bridal shops
		{Name: "The Bridal Finery", OfficialURL: "https://www.thebridalfinery.com/",
			Handle: "thebridalfinery", SourceClass: "bridal_shop",
			CityID: "city_orlando_fl", State: "FL", City: "Orlando", Verified: "2026-08-01"},
		{Name: "Bridal Reflections", OfficialURL: "https://www.bridalreflections.com/",
			Handle: "bridalreflections", SourceClass: "bridal_shop",
			CityID: "city_miami_fl", State: "FL", City: "Miami", Verified: "2026-08-01"},
		// Florida officiants
		{Name: "Miami Wedding Officiants", OfficialURL: "https://www.miamiweddingofficiants.com/",
			Handle: "miamiweddingofficiants", SourceClass: "officiant",
			CityID: "city_miami_fl", State: "FL", City: "Miami", Verified: "2026-08-01"},
		{Name: "Orlando Wedding Officiant", OfficialURL: "https://www.orlandoweddingofficiant.com/",
			Handle: "orlandoweddingofficiant", SourceClass: "officiant",
			CityID: "city_orlando_fl", State: "FL", City: "Orlando", Verified: "2026-08-01"},
	},
}
