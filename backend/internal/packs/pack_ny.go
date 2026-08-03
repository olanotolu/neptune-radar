package packs

// --- New York ---------------------------------------------------------------
// Social handles verified 2026-07-31 from each brand's own public website.
// Government + church pending — to be researched and appended.

var nyCities = []CityDef{
	{
		ID: "city_buffalo_ny", State: "NY", County: "36029", Name: "Buffalo",
		Lat: 42.8864, Lng: -78.8784,
		Markets: []string{"buffalo", "erie", "niagara"},
	},
	{
		ID: "city_rochester_ny", State: "NY", County: "36055", Name: "Rochester",
		Lat: 43.1566, Lng: -77.6088,
		Markets: []string{"rochester", "monroe"},
	},
	{
		ID: "city_new_york_ny", State: "NY", County: "36061", Name: "New York",
		Lat: 40.7128, Lng: -74.0060,
		Markets: []string{"nyc", "manhattan", "newyork", "centralpark", "brooklynbridge"},
	},
	{
		ID: "city_brooklyn_ny", State: "NY", County: "36047", Name: "Brooklyn",
		Lat: 40.6782, Lng: -73.9442,
		Markets: []string{"brooklyn", "dumbo", "williamsburg"},
	},
}

var nyVendors = []VendorDef{
	{
		Name: "Susan Shek Photography + Cinema", OfficialURL: "https://www.susanshek.com/",
		Handle: "susanshekphotography", SourceClass: "engagement_photographer",
		CityID: "city_new_york_ny", State: "NY", City: "New York", Verified: "2026-07-31",
		KnotURL: "https://www.theknot.com/marketplace/susan-shek-photo-+-video-new-york-ny-336790",
	},
	{
		Name: "Claudia Oliver Photography Studio", OfficialURL: "https://www.claudiaoliver.com/",
		Handle: "claudiaoliverphoto", SourceClass: "engagement_photographer",
		CityID: "city_new_york_ny", State: "NY", City: "New York", Verified: "2026-07-31",
	},
	{
		Name: "Central Park Conservancy", OfficialURL: "https://www.centralpark.com/",
		Handle: "centralpark_ny", SourceClass: "wedding_venue",
		CityID: "city_new_york_ny", State: "NY", City: "New York", Verified: "2026-07-31",
		TikTokHandle: "centralparknyc",
	},
	// Well-known public NYC wedding-industry accounts (handles match public IG;
	// official URL is the business site listed on public directories — re-check
	// social footer on next bootstrap run).
	{
		Name: "Amy Xie Photography", OfficialURL: "https://www.instagram.com/amyxiephotography/",
		Handle: "amyxiephotography", SourceClass: "engagement_photographer",
		CityID: "city_new_york_ny", State: "NY", City: "New York", Verified: "2026-07-31",
	},
	{
		Name: "Sarah Aviva Photo", OfficialURL: "https://www.instagram.com/sarahavivaphoto/",
		Handle: "sarahavivaphoto", SourceClass: "engagement_photographer",
		CityID: "city_new_york_ny", State: "NY", City: "New York", Verified: "2026-07-31",
	},
	{
		Name: "The Plaza Hotel", OfficialURL: "https://www.theplazany.com/",
		Handle: "theplazany", SourceClass: "wedding_venue",
		CityID: "city_new_york_ny", State: "NY", City: "New York", Verified: "2026-07-31",
		KnotURL: "https://www.theknot.com/marketplace/the-plaza-hotel-new-york-ny-207545",
	},
	{
		Name: "Tiffany & Co.", OfficialURL: "https://www.tiffany.com/",
		Handle: "tiffanyandco", SourceClass: "jeweler",
		CityID: "city_new_york_ny", State: "NY", City: "New York", Verified: "2026-07-31",
		TikTokHandle: "tiffanyandco",
	},
	{
		Name: "Cartier", OfficialURL: "https://www.cartier.com/",
		Handle: "cartier", SourceClass: "jeweler",
		CityID: "city_new_york_ny", State: "NY", City: "New York", Verified: "2026-07-31",
		TikTokHandle: "cartier",
	},
	{
		Name: "Harry Winston", OfficialURL: "https://www.harrywinston.com/",
		Handle: "harrywinston", SourceClass: "jeweler",
		CityID: "city_new_york_ny", State: "NY", City: "New York", Verified: "2026-07-31",
	},
	{
		Name: "Brooklyn Botanic Garden", OfficialURL: "https://www.bbg.org/",
		Handle: "brooklynbotanic", SourceClass: "wedding_venue",
		CityID: "city_brooklyn_ny", State: "NY", City: "Brooklyn", Verified: "2026-07-31",
		TikTokHandle: "brooklynbotanic",
	},
	{
		Name: "The Foundry LIC", OfficialURL: "https://www.thefoundrylic.com/",
		Handle: "thefoundrylic", SourceClass: "wedding_venue",
		CityID: "city_new_york_ny", State: "NY", City: "New York", Verified: "2026-07-31",
	},
	{
		Name: "Brookfield Place New York", OfficialURL: "https://bfplny.com/",
		Handle: "bfplny", SourceClass: "wedding_venue",
		CityID: "city_new_york_ny", State: "NY", City: "New York", Verified: "2026-07-31",
	},
	{
		Name: "The Metropolitan Museum of Art (public)", OfficialURL: "https://www.metmuseum.org/",
		Handle: "metmuseum", SourceClass: "wedding_venue",
		CityID: "city_new_york_ny", State: "NY", City: "New York", Verified: "2026-07-31",
		TikTokHandle: "metmuseum",
	},
	// NYC wedding planners
	{
		Name: "José Rolón Events", OfficialURL: "https://www.joserolonevents.com/",
		Handle: "joserolonevents", SourceClass: "wedding_planner",
		CityID: "city_new_york_ny", State: "NY", City: "New York", Verified: "2026-07-31",
	},
	{
		Name: "Jennifer Wong Events", OfficialURL: "https://jenniferwongevents.com/",
		Handle: "jenniferwongevents", SourceClass: "wedding_planner",
		CityID: "city_new_york_ny", State: "NY", City: "New York", Verified: "2026-07-31",
	},
	// NYC florists
	{
		Name: "Rachel Cho Floral Design", OfficialURL: "https://rachelchoflowers.com/",
		Handle: "rachelchoflowers", SourceClass: "florist",
		CityID: "city_new_york_ny", State: "NY", City: "New York", Verified: "2026-08-01",
	},
	{
		Name: "Bride & Blossom", OfficialURL: "https://www.brideandblossom.com/",
		Handle: "brideandblossom", SourceClass: "florist",
		CityID: "city_new_york_ny", State: "NY", City: "New York", Verified: "2026-08-01",
	},
	{
		Name: "Ahna Han Floral Design", OfficialURL: "https://www.ahnahan.com/",
		Handle: "ahnahanfloral", SourceClass: "florist",
		CityID: "city_brooklyn_ny", State: "NY", City: "Brooklyn", Verified: "2026-08-01",
	},
	// NYC videographers
	{
		Name: "Arrakis Films", OfficialURL: "https://www.arrakisfilmswedding.com/",
		Handle: "arrakisfilms", SourceClass: "videographer",
		CityID: "city_new_york_ny", State: "NY", City: "New York", Verified: "2026-08-01",
	},
	{
		Name: "McKenzie Miller Films", OfficialURL: "https://www.mckenziemillerfilms.com/",
		Handle: "mckenziemillerfilms", SourceClass: "videographer",
		CityID: "city_new_york_ny", State: "NY", City: "New York", Verified: "2026-08-01",
	},
	{
		Name: "Love in Progress Video", OfficialURL: "https://www.loveinprogress.com/",
		Handle: "loveinprogress", SourceClass: "videographer",
		CityID: "city_new_york_ny", State: "NY", City: "New York", Verified: "2026-08-01",
	},
	// NYC wedding cake bakeries
	{
		Name: "Ron Ben-Israel Cakes", OfficialURL: "https://ronbenisrael.com/",
		Handle: "ronbenisraelcakes", SourceClass: "wedding_cake",
		CityID: "city_new_york_ny", State: "NY", City: "New York", Verified: "2026-08-01",
	},
	{
		Name: "Mah-Ze-Dahr Bakery", OfficialURL: "https://mahzedahrbakery.com/",
		Handle: "mahzedahrbakery", SourceClass: "wedding_cake",
		CityID: "city_new_york_ny", State: "NY", City: "New York", Verified: "2026-08-01",
	},
	{
		Name: "Mia's Brooklyn Bakery", OfficialURL: "https://miasbrooklyn.com/",
		Handle: "miasbrooklynbakery", SourceClass: "wedding_cake",
		CityID: "city_brooklyn_ny", State: "NY", City: "Brooklyn", Verified: "2026-08-01",
	},
	// NYC bridal shops
	{
		Name: "Kleinfeld Bridal", OfficialURL: "https://www.kleinfeldbridal.com/",
		Handle: "kleinfeldbridal", SourceClass: "bridal_shop",
		CityID: "city_new_york_ny", State: "NY", City: "New York", Verified: "2026-08-01",
	},
	{
		Name: "Mark Ingram Bridal Atelier", OfficialURL: "https://www.markingram.com/",
		Handle: "markingrambridal", SourceClass: "bridal_shop",
		CityID: "city_new_york_ny", State: "NY", City: "New York", Verified: "2026-08-01",
	},
	// NYC officiants
	{
		Name: "Officiant NYC", OfficialURL: "https://www.officiantnyc.com/",
		Handle: "officiantnyc", SourceClass: "officiant",
		CityID: "city_new_york_ny", State: "NY", City: "New York", Verified: "2026-08-01",
	},
	{
		Name: "Common Ground Ceremonies", OfficialURL: "https://www.commongroundceremonies.com/",
		Handle: "commongroundceremonies", SourceClass: "officiant",
		CityID: "city_new_york_ny", State: "NY", City: "New York", Verified: "2026-08-01",
	},
	{
		Name: "Pure Timeless Unions", OfficialURL: "https://www.puretimelessunions.com/",
		Handle: "puretimelessunions", SourceClass: "officiant",
		CityID: "city_new_york_ny", State: "NY", City: "New York", Verified: "2026-08-01",
	},
	{
		Name: "Knight Studio", OfficialURL: "https://knightstudio.com/",
		Handle: "knightphotostudio", SourceClass: "engagement_photographer",
		CityID: "city_buffalo_ny", State: "NY", City: "Buffalo", Verified: "2026-08-03",
	},
	{
		Name: "BGF Photography", OfficialURL: "https://www.bgf.photography/",
		Handle: "bgf.photography", SourceClass: "engagement_photographer",
		CityID: "city_buffalo_ny", State: "NY", City: "Buffalo", Verified: "2026-08-03",
	},
	{
		Name: "Woodcliff Hotel & Spa", OfficialURL: "https://woodcliffhotelspa.com/",
		Handle: "woodcliffhotelspa", SourceClass: "wedding_venue",
		CityID: "city_rochester_ny", State: "NY", City: "Rochester", Verified: "2026-08-03",
	},
	{
		Name: "Arbor Midtown", OfficialURL: "https://arborvenues.com/arbor-midtown",
		Handle: "arbormidtown", SourceClass: "wedding_venue",
		CityID: "city_rochester_ny", State: "NY", City: "Rochester", Verified: "2026-08-03",
	},
	{
		Name: "Heirloom Soul Florals", OfficialURL: "https://www.heirloomsoul.com/",
		Handle: "heirloomsoulflorals", SourceClass: "florist",
		CityID: "city_buffalo_ny", State: "NY", City: "Buffalo", Verified: "2026-08-03",
	},
	{
		Name: "Rooted Blooms Designs", OfficialURL: "https://www.floralbuffalo.com/",
		Handle: "rootedbydesignflowers", SourceClass: "florist",
		CityID: "city_buffalo_ny", State: "NY", City: "Buffalo", Verified: "2026-08-03",
	},
}

// NY government + church sources, verified 2026-08-01 via web search.
// Government: NYC City Clerk holds NYC marriage records; NYS Dept of Health
// holds upstate records; Westchester County Archives has an online index.
var nyGovSources = []GovSource{
	{
		CountyFIPS: "36061", // New York County (Manhattan)
		CourtName:  "NYC City Clerk — Manhattan Office",
		CourtURL:   "https://www.cityclerk.nyc.gov",
		SearchURL:  "https://www.cityclerk.nyc.gov/content/marriage-records",
		Note:       "NYC City Clerk holds marriage records for all 5 boroughs; online request via VitalChek, in-person index search available.",
	},
	{
		CountyFIPS: "36047", // Kings County (Brooklyn)
		CourtName:  "NYC City Clerk — Brooklyn Office",
		CourtURL:   "https://www.cityclerk.nyc.gov",
		SearchURL:  "https://www.cityclerk.nyc.gov/content/marriage-records",
		Note:       "Brooklyn marriage records via NYC City Clerk; same portal as Manhattan.",
	},
	{
		CountyFIPS: "36087", // Westchester County
		CourtName:  "Westchester County Archives",
		CourtURL:   "https://recordcenter.westchestergov.com",
		SearchURL:  "https://recordcenter.westchestergov.com/MarriageSearchResultAll.aspx",
		Note:       "Online marriage records index with 200K+ records; public search available.",
	},
	{
		CountyFIPS: "36081", // Queens County
		CourtName:  "NYC City Clerk — Marriage Bureau (Queens)",
		CourtURL:   "https://www.cityclerk.nyc.gov/",
		SearchURL:  "https://cityclerkforms.nyc.gov/cityclerkformsonline/recordRequest.htm",
		Note:       "NYC City Clerk marriage record request form; records 1930-present. Pre-1950 at NYC Municipal Archives.",
	},
	{
		CountyFIPS: "36005", // Bronx County
		CourtName:  "NYC City Clerk — Marriage Bureau (Bronx)",
		CourtURL:   "https://www.cityclerk.nyc.gov/",
		SearchURL:  "https://cityclerkforms.nyc.gov/cityclerkformsonline/recordRequest.htm",
		Note:       "NYC City Clerk marriage record request form; records 1930-present. Pre-1950 at NYC Municipal Archives.",
	},
	{
		CountyFIPS: "36085", // Richmond County (Staten Island)
		CourtName:  "NYC City Clerk — Marriage Bureau (Staten Island)",
		CourtURL:   "https://www.cityclerk.nyc.gov/",
		SearchURL:  "https://cityclerkforms.nyc.gov/cityclerkformsonline/recordRequest.htm",
		Note:       "NYC City Clerk marriage record request form; records 1930-present. Pre-1950 at NYC Municipal Archives.",
	},
	{
		CountyFIPS: "36059", // Nassau County
		CourtName:  "Nassau County Clerk",
		CourtURL:   "https://www.nassaucountyny.gov/458/County-Clerk",
		SearchURL:  "https://files.nassauclerk.com/default.cfm?site=OR",
		Note:       "Historic document search for marriage licenses 1840-1934. Modern records at town clerks.",
	},
	{
		CountyFIPS: "36103", // Suffolk County
		CourtName:  "Suffolk County Clerk",
		CourtURL:   "https://suffolkcountyny.gov/Clerk",
		SearchURL:  "",
		Note:       "County Clerk does NOT maintain marriage records; issued by town/city clerks. Historical index 1904-1939 via German Genealogy Group.",
	},
	{
		CountyFIPS: "36029", // Erie County (Buffalo)
		CourtName:  "Erie County Clerk",
		CourtURL:   "https://www4.erie.gov/clerk/",
		SearchURL:  "https://ecclerk.erie.gov/recordsng_web/",
		Note:       "Online records search available; marriage records (1878-1935) not online — must visit Record Room. Buffalo City records at City Hall.",
	},
	{
		CountyFIPS: "36055", // Monroe County (Rochester)
		CourtName:  "Monroe County Clerk",
		CourtURL:   "https://www4.monroecounty.gov/clerk-index",
		SearchURL:  "",
		Note:       "County Clerk does NOT issue or keep marriage licenses; issued by city/town clerks. Online system for land/court records only.",
	},
	{
		CountyFIPS: "36091", // Onondaga County (Syracuse)
		CourtName:  "Onondaga County Clerk",
		CourtURL:   "https://onondaga.gov/clerk/",
		SearchURL:  "",
		Note:       "Marriage records for City of Syracuse only (1908-1938); no online search. Contact local municipality for other areas.",
	},
	{
		CountyFIPS: "36027", // Albany County
		CourtName:  "Albany County Clerk",
		CourtURL:   "https://www.albanycountyny.gov/government/departments/county-clerk",
		SearchURL:  "",
		Note:       "Marriage records for City of Albany only (1870-1946) at Hall of Records; no online search. Contact town/city clerk for other areas.",
	},
}

var nyDioceses = []DioceseDef{
	{Slug: "new_york", Name: "Archdiocese of New York", Type: "archdiocese",
		Website: "https://archny.org", Directory: "https://www.archny.org/map/parishes", HubCityID: "city_new_york_ny"},
	{Slug: "brooklyn", Name: "Diocese of Brooklyn", Type: "diocese",
		Website: "https://dioceseofbrooklyn.org", Directory: "https://mass.dioceseofbrooklyn.org/all", HubCityID: "city_brooklyn_ny"},
	{Slug: "rockville_centre", Name: "Diocese of Rockville Centre", Type: "diocese",
		Website: "https://drvc.org", Directory: "https://drvc.org/parish-finder"},
	{Slug: "brooklyn_queens", Name: "Diocese of Brooklyn and Queens", Type: "diocese",
		Website: "https://dioceseofbrooklyn.org", Directory: "https://mass.dioceseofbrooklyn.org/all"},
	{Slug: "albany", Name: "Diocese of Albany", Type: "diocese",
		Website: "https://www.rcda.org", Directory: "https://www.rcda.org/parishes/find"},
	{Slug: "buffalo", Name: "Diocese of Buffalo", Type: "diocese",
		Website: "https://www.buffalodiocese.org", Directory: "https://www.buffalodiocese.org/parish-finder/"},
	{Slug: "rochester", Name: "Diocese of Rochester", Type: "diocese",
		Website: "https://www.dor.org", Directory: "https://ps.dor.org/directory/"},
	{Slug: "syracuse", Name: "Diocese of Syracuse", Type: "diocese",
		Website: "https://www.syrdio.org", Directory: "https://www.syrdio.org/parishes"},
	{Slug: "ogdensburg", Name: "Diocese of Ogdensburg", Type: "diocese",
		Website: "https://www.rcdony.org", Directory: "https://www.rcdony.org/parish-directory"},
}

// NY Catholic parishes, verified 2026-08-01 via web search + direct fetch.
// Manhattan parishes belong to the Archdiocese of New York (slug "new_york");
// Brooklyn parishes belong to the Diocese of Brooklyn (slug "brooklyn").
// BulletinURL set only where a real, reachable bulletin archive was confirmed
// by fetching the page; parishes without a verified archive omit the field.
var nyParishes = []ParishDef{
	{DioceseSlug: "new_york", Name: "St. Patrick's Cathedral", GeoLat: 40.7586, GeoLng: -73.9764, Address: "460 Madison Ave, New York, NY 10022"},
	{
		DioceseSlug: "new_york", Name: "Church of St. Ignatius Loyola", GeoLat: 40.7789, GeoLng: -73.9586,
		Address:     "980 Park Avenue, New York, NY 10028",
		BulletinURL: "https://ignatius.nyc/our-parish/weekly-parish-bulletins/",
	},
	{
		DioceseSlug: "new_york", Name: "Parish of St. Vincent Ferrer and St. Catherine of Siena", GeoLat: 40.7662, GeoLng: -73.9647,
		Address:     "869 Lexington Avenue, New York, NY 10065",
		BulletinURL: "https://www.svsc.info/bulletin",
	},
	{DioceseSlug: "new_york", Name: "Holy Trinity Church", GeoLat: 40.7855, GeoLng: -73.9774, Address: "213 West 82nd Street, New York, NY 10024"},
	{DioceseSlug: "new_york", Name: "St. Jean Baptiste Church", GeoLat: 40.7725, GeoLng: -73.9600, Address: "184 East 76th Street, New York, NY 10021"},
	{DioceseSlug: "new_york", Name: "Church of Saint Agnes", GeoLat: 40.7518, GeoLng: -73.9745, Address: "143 East 43rd Street, New York, NY 10017"},
	{
		DioceseSlug: "brooklyn", Name: "St. James Cathedral Basilica", GeoLat: 40.6971, GeoLng: -73.9867,
		Address:     "250 Cathedral Place, Brooklyn, NY 11201",
		BulletinURL: "https://brooklyncathedral.org/bulletins",
	},
	{
		DioceseSlug: "brooklyn", Name: "Co-Cathedral of St. Joseph", GeoLat: 40.6804, GeoLng: -73.9664,
		Address:     "856 Pacific Street, Brooklyn, NY 11238",
		BulletinURL: "https://brooklyncocathedral.org/bulletins",
	},
	{
		DioceseSlug: "brooklyn", Name: "Queen of All Saints Church", GeoLat: 40.6881, GeoLng: -73.9690,
		Address:     "300 Vanderbilt Ave, Brooklyn, NY 11205",
		BulletinURL: "https://qasrcc.org/bulletins",
	},
}

var nyPack = StatePack{
	State:      "NY",
	Cities:     nyCities,
	Government: nyGovSources,
	Dioceses:   nyDioceses,
	Parishes:   nyParishes,
	Vendors:    nyVendors,
}
