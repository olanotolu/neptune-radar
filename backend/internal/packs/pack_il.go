package packs

// Illinois source pack — verified 2026-08-01.
//
// Government: Illinois marriage records are held by the county clerk. Search
// URLs for the top 8 counties by population were verified against each county's
// official .gov site. Several counties offer genealogy search portals (Cook,
// Kane, Lake) for historical marriage records; others require VitalChek or
// in-person/mail requests for certified copies.
//
// Church: all 6 Illinois Catholic dioceses/archdioceses verified via USCCB +
// each diocese's own website. Parish directory URLs point at each
// jurisdiction's own parish-finder. Chicago-area parishes (Archdiocese of
// Chicago) were verified against the archdiocese's own parish directory PDF +
// direct bulletin-archive URL discovery from each parish's website.
//
// Social: Instagram handles verified from each business's own public website
// social links (or from directory listings / IG search results where the site
// is JS-rendered and the handle was visible in the search snippet). Verification
// date recorded per vendor.

var ilPack = StatePack{
	State: "IL",

	// --- Cities ----------------------------------------------------------
	Cities: []CityDef{
		{
			ID: "city_chicago_il", State: "IL", County: "17031", Name: "Chicago",
			Lat: 41.8781, Lng: -87.6298,
			Markets: []string{"chicago", "chi", "chicagoland", "rivernorth", "westloop", "lincolnpark"},
		},
		{
			ID: "city_aurora_il", State: "IL", County: "17089", Name: "Aurora",
			Lat: 41.7606, Lng: -88.3201,
			Markets: []string{"aurora", "foxvalley", "naperville", "kane"},
		},
		{
			ID: "city_naperville_il", State: "IL", County: "17043", Name: "Naperville",
			Lat: 41.7508, Lng: -88.1535,
			Markets: []string{"naperville", "dupage", "wheaton", "foxvalley"},
		},
	},

	// --- Government (county clerk marriage-record searches) --------------
	Government: []GovSource{
		{
			// Cook County (Chicago) — county clerk vital records; genealogy
			// search portal for marriage licenses 50+ years old.
			CountyFIPS: "17031",
			CourtName:  "Cook County Clerk",
			CourtURL:   "https://www.cookcountyclerkil.gov/vital-records/marriage-civil-union",
			SearchURL:  "https://cookcountygenealogy.com/Default.aspx",
			Note:       "Genealogy online search for marriage licenses 50+ years old; modern records require VitalChek or in-person request.",
		},
		{
			// DuPage County — county clerk vital records; no online search
			// portal, request via form, VitalChek, or in-person.
			CountyFIPS: "17043",
			CourtName:  "DuPage County Clerk",
			CourtURL:   "https://www.dupagecounty.gov/elected_officials/county_clerk/vital_records.php",
			SearchURL:  "https://www.dupagecounty.gov/elected_officials/county_clerk/vital_records.php",
			Note:       "Marriage certificates open to public but no online search; request via form, VitalChek, or in-person.",
		},
		{
			// Lake County — county clerk vital records; genealogical records
			// search for pre-1916 marriages.
			CountyFIPS: "17097",
			CourtName:  "Lake County Clerk",
			CourtURL:   "https://www.lakecountyil.gov/3964/Marriage-Records",
			SearchURL:  "https://il-lakecounty.civicplus.com/398/Genealogical-Records-Search",
			Note:       "Genealogical records search for pre-1916 marriages; modern records via VitalChek or in-person.",
		},
		{
			// Will County — county clerk vital records; no online search
			// portal, VitalChek for certified copies.
			CountyFIPS: "17197",
			CourtName:  "Will County Clerk",
			CourtURL:   "https://www.willcountyclerk.gov/vital-records-2/",
			SearchURL:  "https://www.willcountyclerk.gov/vital-records-2/",
			Note:       "Marriage records from 1836; no online search portal; VitalChek for certified copies or in-person/mail request.",
		},
		{
			// Kane County — county clerk vital records; genealogy search
			// portal for marriage licenses 50+ years old (1836–1962).
			CountyFIPS: "17089",
			CourtName:  "Kane County Clerk",
			CourtURL:   "https://clerk.kanecountyil.gov/VitalRecords/Pages/Marriage.aspx",
			SearchURL:  "https://genealogy.kanecountyclerk.org/",
			Note:       "Genealogy online search for marriage licenses 50+ years old (1836–1962); modern records via VitalChek or in-person.",
		},
		{
			// McHenry County — county clerk vital records; no online search
			// portal, VitalChek for certified copies.
			CountyFIPS: "17111",
			CourtName:  "McHenry County Clerk",
			CourtURL:   "https://www.mchenrycountyil.gov/departments/county-clerk/vital-records/marriage-civil-union",
			SearchURL:  "https://www.mchenrycountyil.gov/departments/county-clerk/vital-records/marriage-civil-union",
			Note:       "Marriage records from 1837; no online search portal; VitalChek or in-person/mail request.",
		},
		{
			// Winnebago County — county clerk vital records; no public search
			// portal, Official Records Online for certified copies.
			CountyFIPS: "17201",
			CourtName:  "Winnebago County Clerk",
			CourtURL:   "https://wincoil.gov/departments/clerks-office/vital-records",
			SearchURL:  "https://wincoil.gov/departments/clerks-office/vital-records",
			Note:       "Marriage records from 1836; no public search portal; Official Records Online or in-person/mail request.",
		},
		{
			// St. Clair County — county clerk vital records; no online search
			// portal, VitalChek for certified copies.
			CountyFIPS: "17163",
			CourtName:  "St. Clair County Clerk",
			CourtURL:   "https://www.co.st-clair.il.us/departments/county-clerk/vital-records",
			SearchURL:  "https://www.co.st-clair.il.us/departments/county-clerk/vital-records",
			Note:       "Marriage records from 1763; no online search portal; VitalChek or in-person/mail request.",
		},
	},

	// --- Church (Catholic dioceses + parishes) ---------------------------
	Dioceses: []DioceseDef{
		{
			Slug: "chicago", Name: "Archdiocese of Chicago",
			Type: "archdiocese", Website: "https://www.archchicago.org",
			Directory: "https://www.archchicago.org/parish-map",
			HubCityID: "city_chicago_il",
		},
		{
			Slug: "joliet", Name: "Diocese of Joliet",
			Type: "diocese", Website: "https://www.diojoliet.org",
			Directory: "https://www.diojoliet.org/parishes",
		},
		{
			Slug: "rockford", Name: "Diocese of Rockford",
			Type: "diocese", Website: "https://www.rockforddiocese.org",
			Directory: "https://www.rockforddiocese.org/parishes/",
		},
		{
			Slug: "peoria", Name: "Diocese of Peoria",
			Type: "diocese", Website: "https://www.cdop.org",
			Directory: "https://www.cdop.org/parishes",
		},
		{
			Slug: "springfield", Name: "Diocese of Springfield in Illinois",
			Type: "diocese", Website: "https://dio.org",
			Directory: "https://dio.org/view/parish-directory/",
		},
		{
			Slug: "belleville", Name: "Diocese of Belleville",
			Type: "diocese", Website: "https://www.diobelle.org",
			Directory: "https://www.diobelle.org/directory/parishes",
		},
	},

	// Chicago-area parishes in the Archdiocese of Chicago.
	// Names verified from the archdiocese's own parish directory PDF (March
	// 2026 edition, archchicago.org/documents/). Bulletin URLs verified by
	// direct search for each parish's bulletin archive page.
	Parishes: []ParishDef{
		{
			DioceseSlug: "chicago", Name: "Holy Name Cathedral",
			Address:     "735 N. State St., Chicago, IL 60654",
			BulletinURL: "https://holynamecathedral.org/bulletin-and-newsletter/",
		},
		{
			DioceseSlug: "chicago", Name: "Old St. Patrick's Church",
			Address:     "700 West Adams Street, Chicago, IL 60661",
			BulletinURL: "https://www.oldstpats.org/crossroads-publication.html",
		},
		{
			DioceseSlug: "chicago", Name: "Immaculate Conception Parish",
			Address:     "7211 West Talcott Avenue, Chicago, IL 60631",
			BulletinURL: "https://www.icchicago.org/discover-ic/bulletin-archive.cfm",
		},
		{
			DioceseSlug: "chicago", Name: "Our Lady of Kibeho",
			Address:     "11159 South Loomis, Chicago, IL 60643",
			BulletinURL: "https://olk.archchicago.org/bulletins-and-additional-publications.html",
		},
		{
			DioceseSlug: "chicago", Name: "St. Monica and St. Rosalie Parish",
			Address:     "5136 N. Nottingham, Chicago, IL 60656",
			BulletinURL: "https://www.stsmonicarosalie.org/bulletins.html",
		},
		{
			DioceseSlug: "chicago", Name: "Blessed Sacrament Parish",
			Address:     "3745 S Paulina St, Chicago, IL 60609",
			BulletinURL: "https://www.bspchicago.org/bulletins",
		},
		{
			DioceseSlug: "chicago", Name: "St. John Bosco-St. James Parish",
			Address:     "2250 N. McVicker Ave., Chicago, IL 60639",
			BulletinURL: "https://www.sjbchicago.org/bulletins",
		},
		{
			DioceseSlug: "chicago", Name: "Saint Linus Catholic Church",
			Address:     "10300 S. Lawler Avenue, Oak Lawn, IL 60453",
			BulletinURL: "https://www.stlinusoaklawn.org/Bulletins.php",
		},
		{
			DioceseSlug: "chicago", Name: "St. Clement Parish",
			Address: "642 West Deming Place, Chicago, IL 60614",
		},
		{
			DioceseSlug: "chicago", Name: "Queen of All Saints Basilica",
			Address:     "6280 North Sauganash Avenue, Chicago, IL 60646",
			BulletinURL: "https://qaschurch.org/bulletins",
		},
	},

	// --- Social (wedding-industry Instagram vendors) ---------------------
	Vendors: []VendorDef{
		// Chicago photographers
		{
			Name: "Jeremy Glickstein Photography", OfficialURL: "https://jeremyglicksteinphotography.com/",
			Handle: "glickofficial", SourceClass: "engagement_photographer",
			CityID: "city_chicago_il", State: "IL", City: "Chicago", Verified: "2026-08-01",
			TikTokHandle: "glickofficial",
		},
		{
			Name: "Ashlee Cole Photography", OfficialURL: "https://ashleecole.com/",
			Handle: "ashleelcole", SourceClass: "engagement_photographer",
			CityID: "city_chicago_il", State: "IL", City: "Chicago", Verified: "2026-08-01",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-photographers/ashlee-cole-photography-5411786",
		},
		{
			Name: "Marissa Kelly Photography", OfficialURL: "https://marissakellyphotography.com/",
			Handle: "marissakellyphotography", SourceClass: "engagement_photographer",
			CityID: "city_chicago_il", State: "IL", City: "Chicago", Verified: "2026-08-01",
			TikTokHandle: "marissakellyphotography",
		},
		{
			Name: "Karly Ellis Photography", OfficialURL: "https://www.karlyellisphotography.com/",
			Handle: "karlyellisphotography", SourceClass: "engagement_photographer",
			CityID: "city_chicago_il", State: "IL", City: "Chicago", Verified: "2026-08-01",
		},
		{
			Name: "Alex Maldonado Photography", OfficialURL: "https://www.alexmaldonadophotography.com/",
			Handle: "alexmaldonadophotography", SourceClass: "engagement_photographer",
			CityID: "city_chicago_il", State: "IL", City: "Chicago", Verified: "2026-08-01",
			TikTokHandle: "alexmaldonadophotography",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-photographers/alex-maldonado-photography-3271017",
		},
		{
			Name: "The Adamkovi", OfficialURL: "https://theadamkovi.com/",
			Handle: "theadamkovi", SourceClass: "engagement_photographer",
			CityID: "city_chicago_il", State: "IL", City: "Chicago", Verified: "2026-08-01",
		},
		// Chicago venues
		{
			Name: "The Arbory", OfficialURL: "https://thearborychicago.com/",
			Handle: "the.arbory", SourceClass: "wedding_venue",
			CityID: "city_chicago_il", State: "IL", City: "Chicago", Verified: "2026-08-01",
			TikTokHandle: "the.arbory",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-venues/the-arbory-4711241",
		},
		{
			Name: "Stan Mansion", OfficialURL: "https://www.stanmansion.com/",
			Handle: "stanmansion", SourceClass: "wedding_venue",
			CityID: "city_chicago_il", State: "IL", City: "Chicago", Verified: "2026-08-01",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-venues/stan-mansion-3595664",
		},
		{
			Name: "Galleria Marchetti", OfficialURL: "https://www.galleriamarchetti.com/",
			Handle: "galleriamarchetti", SourceClass: "wedding_venue",
			CityID: "city_chicago_il", State: "IL", City: "Chicago", Verified: "2026-08-01",
			TikTokHandle: "galleriamarchetti",
		},
		{
			Name: "The Wellsley", OfficialURL: "https://www.wellsleychicago.com/",
			Handle: "thewellsley", SourceClass: "wedding_venue",
			CityID: "city_chicago_il", State: "IL", City: "Chicago", Verified: "2026-08-01",
		},
		// Chicago jewelers
		{
			Name: "Gregg Helfer Ltd.", OfficialURL: "https://gregghelfer.com/",
			Handle: "gregghelferltd", SourceClass: "jeweler",
			CityID: "city_chicago_il", State: "IL", City: "Chicago", Verified: "2026-08-01",
		},
		{
			Name: "Ivy & Rose Fine Jewelry", OfficialURL: "https://ivyandrose.com/",
			Handle: "ivyandrosevintage", SourceClass: "jeweler",
			CityID: "city_chicago_il", State: "IL", City: "Chicago", Verified: "2026-08-01",
		},
		// Chicago wedding planners
		{
			Name: "JPB Designs + Co.", OfficialURL: "https://jpbdesigns.com/",
			Handle: "jpbdesigns", SourceClass: "wedding_planner",
			CityID: "city_chicago_il", State: "IL", City: "Chicago", Verified: "2026-08-01",
		},
		{
			Name: "BWEDDINGS", OfficialURL: "https://bweddingsplanner.com/",
			Handle: "bweddingschicago", SourceClass: "wedding_planner",
			CityID: "city_chicago_il", State: "IL", City: "Chicago", Verified: "2026-08-01",
		},
		// Chicago florists
		{
			Name: "Flowers for Dreams", OfficialURL: "https://flowersfordreams.com/",
			Handle: "flowersfordreams", SourceClass: "florist",
			CityID: "city_chicago_il", State: "IL", City: "Chicago", Verified: "2026-08-01",
		},
		{
			Name: "Life In Bloom", OfficialURL: "https://lifeinbloomchicago.com/",
			Handle: "lifeinbloomchicago", SourceClass: "florist",
			CityID: "city_chicago_il", State: "IL", City: "Chicago", Verified: "2026-08-01",
		},
		// Chicago videographers
		{
			Name: "MK Films", OfficialURL: "https://www.mkfilms.com/",
			Handle: "mkfilms", SourceClass: "videographer",
			CityID: "city_chicago_il", State: "IL", City: "Chicago", Verified: "2026-08-01",
		},
		{
			Name: "Visual Films Chicago", OfficialURL: "https://www.visualfilmschicago.com/",
			Handle: "visualfilmschicago", SourceClass: "videographer",
			CityID: "city_chicago_il", State: "IL", City: "Chicago", Verified: "2026-08-01",
		},
		// Chicago wedding cake bakeries
		{
			Name: "Roeser's Bakery", OfficialURL: "https://www.roesersbakery.com/",
			Handle: "roesersbakery", SourceClass: "wedding_cake",
			CityID: "city_chicago_il", State: "IL", City: "Chicago", Verified: "2026-08-01",
		},
		{
			Name: "Brown Sugar Bakery", OfficialURL: "https://www.brownsugarbakerychicago.com/",
			Handle: "brownsugarbakery", SourceClass: "wedding_cake",
			CityID: "city_chicago_il", State: "IL", City: "Chicago", Verified: "2026-08-01",
		},
		// Chicago bridal shops
		{
			Name: "Bella Bianca Bridal Couture", OfficialURL: "https://www.bellabianca.com/",
			Handle: "bellabianca", SourceClass: "bridal_shop",
			CityID: "city_chicago_il", State: "IL", City: "Chicago", Verified: "2026-08-01",
		},
		{
			Name: "Aria Bridal Couture", OfficialURL: "https://www.ariabridal.com/",
			Handle: "ariabridal", SourceClass: "bridal_shop",
			CityID: "city_chicago_il", State: "IL", City: "Chicago", Verified: "2026-08-01",
		},
		// Chicago officiants
		{
			Name: "Chicago Wedding Officiants", OfficialURL: "https://www.chicagoweddingofficiants.com/",
			Handle: "chicagoweddingofficiants", SourceClass: "officiant",
			CityID: "city_chicago_il", State: "IL", City: "Chicago", Verified: "2026-08-01",
		},
		{
			Name: "Reverend Dan's Wedding Service", OfficialURL: "https://www.revdans.com/",
			Handle: "revdans", SourceClass: "officiant",
			CityID: "city_chicago_il", State: "IL", City: "Chicago", Verified: "2026-08-01",
		},
	},
}
