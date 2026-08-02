package packs

// North Carolina source pack — verified 2026-08-01 via web search.
// Government: NC marriage records held by county Register of Deeds.
// Mecklenburg has an online search portal; Wake has an index search.

var ncPack = StatePack{
	State: "NC",
	Cities: []CityDef{
		{ID: "city_charlotte_nc", State: "NC", County: "37119", Name: "Charlotte",
			Lat: 35.2271, Lng: -80.8431, Markets: []string{"charlotte", "clt", "mecklenburg", "queen_city"}},
		{ID: "city_raleigh_nc", State: "NC", County: "37183", Name: "Raleigh",
			Lat: 35.7796, Lng: -78.6382, Markets: []string{"raleigh", "rdu", "wake", "triangle"}},
		{ID: "city_asheville_nc", State: "NC", County: "37021", Name: "Asheville",
			Lat: 35.5951, Lng: -82.5515, Markets: []string{"asheville", "wnc", "buncombe"}},
	},
	Government: []GovSource{
		{CountyFIPS: "37119", CourtName: "Mecklenburg County Register of Deeds",
			CourtURL:  "https://deeds.mecknc.gov",
			SearchURL: "https://meckrod.manatron.com/Marriage/SearchEntry.aspx?e=newSession",
			Note:      "Online marriage record search portal; enumeration candidate."},
		{CountyFIPS: "37183", CourtName: "Wake County Register of Deeds",
			CourtURL:  "https://www.wake.gov/departments-government/register-deeds",
			SearchURL: "https://www.wake.gov/departments-government/register-deeds/vital-records",
			Note:      "Marriage index search available; enumeration capability needs testing."},
		{CountyFIPS: "37081", CourtName: "Guilford County Register of Deeds",
			CourtURL:  "https://www.guilfordcountync.gov/our-county/register-of-deeds",
			SearchURL: "https://www.guilfordcountync.gov/our-county/register-of-deeds",
			Note:      "Marriage records via Register of Deeds; request-oriented."},
		{CountyFIPS: "37067", CourtName: "Forsyth County Register of Deeds",
			CourtURL:  "https://www.co.forsyth.nc.us",
			SearchURL: "https://www.co.forsyth.nc.us",
			Note:      "Marriage records via Register of Deeds; request-oriented."},
		{CountyFIPS: "37051", CourtName: "Cumberland County Register of Deeds",
			CourtURL:  "https://www.cumberlandcountync.gov",
			SearchURL: "https://www.cumberlandcountync.gov",
			Note:      "Marriage records via Register of Deeds; request-oriented."},
		{CountyFIPS: "37063", CourtName: "Durham County Register of Deeds",
			CourtURL:  "https://www.dconc.gov",
			SearchURL: "https://www.dconc.gov",
			Note:      "Marriage records via Register of Deeds; request-oriented."},
		{CountyFIPS: "37021", CourtName: "Buncombe County Register of Deeds",
			CourtURL:  "https://www.buncombecounty.org/governing/depts/register-of-deeds",
			SearchURL: "https://www.buncombecounty.org/governing/depts/register-of-deeds",
			Note:      "Marriage records via Register of Deeds; request-oriented."},
	},
	Dioceses: []DioceseDef{
		{Slug: "charlotte", Name: "Diocese of Charlotte", Type: "diocese",
			Website: "https://charlottediocese.org", Directory: "https://charlottediocese.org/find-parish/", HubCityID: "city_charlotte_nc"},
		{Slug: "raleigh", Name: "Diocese of Raleigh", Type: "diocese",
			Website: "https://dioceseofraleigh.org", Directory: "https://dioceseofraleigh.org/find-a-parish", HubCityID: "city_raleigh_nc"},
	},
	Parishes: []ParishDef{
		{DioceseSlug: "charlotte", Name: "Basilica of St. Lawrence", Address: "408 Biltmore Ave, Asheville, NC 28801"},
		{DioceseSlug: "charlotte", Name: "Church of the Epiphany", Address: "7800 Shelby St, Charlotte, NC 28210"},
		{DioceseSlug: "raleigh", Name: "Holy Name of Jesus Cathedral", Address: "715 Nazareth St, Raleigh, NC 27606"},
		{DioceseSlug: "raleigh", Name: "Our Lady of Lourdes", Address: "2710 Overbrook Dr, Raleigh, NC 27608"},
		{DioceseSlug: "raleigh", Name: "Sacred Heart Church", Address: "224 N Wilmington St, Raleigh, NC 27601"},
		{DioceseSlug: "raleigh", Name: "Saint Francis of Assisi", Address: "11401 Leesville Rd, Raleigh, NC 27614"},
	},
	Vendors: []VendorDef{
		// Charlotte photographers
		{Name: "Charlotte Wedding Photographer", OfficialURL: "https://atlweddingphotos.com/",
			Handle: "charlotteweddings", SourceClass: "engagement_photographer",
			CityID: "city_charlotte_nc", State: "NC", City: "Charlotte", Verified: "2026-08-01"},
		// Charlotte venues
		{Name: "The Yorkmont", OfficialURL: "https://www.theyorkmont.com/",
			Handle: "theyorkmont", SourceClass: "wedding_venue",
			CityID: "city_charlotte_nc", State: "NC", City: "Charlotte", Verified: "2026-08-01"},
		{Name: "The Ivy Place", OfficialURL: "https://www.ivyplaceweddings.com/",
			Handle: "theivyplace", SourceClass: "wedding_venue",
			CityID: "city_charlotte_nc", State: "NC", City: "Charlotte", Verified: "2026-08-01"},
		// Charlotte jewelers
		{Name: "Ballantyne Jewelers", OfficialURL: "https://ballantynejewelers.com/",
			Handle: "ballantynejewelers", SourceClass: "jeweler",
			CityID: "city_charlotte_nc", State: "NC", City: "Charlotte", Verified: "2026-08-01"},
		{Name: "Ascot Diamonds Charlotte", OfficialURL: "https://www.ascotdiamonds.com/locations/charlotte",
			Handle: "ascotdiamondscharlotte", SourceClass: "jeweler",
			CityID: "city_charlotte_nc", State: "NC", City: "Charlotte", Verified: "2026-08-01"},
		// Raleigh photographers
		{Name: "Brian Mullins Photography", OfficialURL: "https://brianmullinsphotography.com/",
			Handle: "brianmullinsphotography", SourceClass: "engagement_photographer",
			CityID: "city_raleigh_nc", State: "NC", City: "Raleigh", Verified: "2026-08-01"},
		{Name: "Rae Marshall Photography", OfficialURL: "https://www.raemarshall.com/",
			Handle: "raemarshall", SourceClass: "engagement_photographer",
			CityID: "city_raleigh_nc", State: "NC", City: "Raleigh", Verified: "2026-08-01"},
		// Raleigh venues
		{Name: "The Maxwell", OfficialURL: "https://themaxwellraleigh.com/",
			Handle: "themaxwellraleigh", SourceClass: "wedding_venue",
			CityID: "city_raleigh_nc", State: "NC", City: "Raleigh", Verified: "2026-08-01"},
		{Name: "Highgrove Estate", OfficialURL: "https://highgrove-nc.com/",
			Handle: "highgrovenc", SourceClass: "wedding_venue",
			CityID: "city_raleigh_nc", State: "NC", City: "Raleigh", Verified: "2026-08-01"},
		// Raleigh jeweler
		{Name: "Raleigh Diamond", OfficialURL: "https://www.raleighdiamond.com/",
			Handle: "raleighdiamond", SourceClass: "jeweler",
			CityID: "city_raleigh_nc", State: "NC", City: "Raleigh", Verified: "2026-08-01"},
		// Asheville photographer
		{Name: "Nick Levine Photography", OfficialURL: "https://www.nicklevinephoto.com/",
			Handle: "nicklevinephoto", SourceClass: "engagement_photographer",
			CityID: "city_asheville_nc", State: "NC", City: "Asheville", Verified: "2026-08-01"},
		// Asheville venue
		{Name: "The Farm Events", OfficialURL: "https://thefarmevents.com/",
			Handle: "thefarmevents", SourceClass: "wedding_venue",
			CityID: "city_asheville_nc", State: "NC", City: "Asheville", Verified: "2026-08-01"},
		// Charlotte wedding planner
		{Name: "J'adore Jay", OfficialURL: "https://jadorejay.com/",
			Handle: "jadorejayweddings", SourceClass: "wedding_planner",
			CityID: "city_charlotte_nc", State: "NC", City: "Charlotte", Verified: "2026-08-01"},
		// Raleigh wedding planner
		{Name: "Southern Oak Events", OfficialURL: "https://www.southernoakevents.com/",
			Handle: "southernoakevents", SourceClass: "wedding_planner",
			CityID: "city_raleigh_nc", State: "NC", City: "Raleigh", Verified: "2026-08-01"},
	},
}
