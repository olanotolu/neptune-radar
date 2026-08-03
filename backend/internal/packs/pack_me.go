package packs

// Maine source pack — verified 2026-08-01.
//
// Government: Maine marriage records are held by the state CDC Division of
// Data, Research, and Vital Statistics (DRVS) and most municipal offices.
// The DRVS online index (DocuWare) covers births, deaths, and marriages
// from 1892 to present. Marriage records >50 years old are public.
//
// Church: Diocese of Portland verified via USCCB + portlanddiocese.org.
// Portland-area parishes verified against the Portland Peninsula & Island
// Parishes cluster website (portlandcatholic.org) + the diocese's own
// parish finder. Bulletin URLs verified by direct search.
//
// Social: Instagram handles verified from each business's own public website
// social links (or from IG search results where the site is JS-rendered).

var mePack = StatePack{
	State: "ME",

	// --- Cities ----------------------------------------------------------
	Cities: []CityDef{
	{
		ID: "city_bangor_me", State: "ME", County: "23019", Name: "Bangor",
		Lat: 44.8015, Lng: -68.7778,
		Markets: []string{"bangor", "penobscot"},
	},
		{ID: "city_portland_me", State: "ME", County: "23005", Name: "Portland",
			Lat: 43.6591, Lng: -70.2568, Markets: []string{"portland", "me", "maine", "cumberland"}},
	},

	// --- Government (statewide vital records + municipal clerks) --------
	Government: []GovSource{
		{
			// Cumberland County (Portland) — Maine CDC DRVS statewide index.
			CountyFIPS: "23005",
			CourtName:  "Maine CDC — Division of Vital Records",
			CourtURL:   "https://www.maine.gov/dhhs/mecdc/vital-records",
			SearchURL:  "https://docuware.maine.gov/DocuWare/Platform/WebClient/",
			Note:       "Statewide DocuWare vital records index; marriage records from 1892; municipal clerks also hold records.",
		},
		{
			// York County — Maine CDC DRVS statewide index.
			CountyFIPS: "23031",
			CourtName:  "Maine CDC — Division of Vital Records",
			CourtURL:   "https://www.maine.gov/dhhs/mecdc/vital-records",
			SearchURL:  "https://docuware.maine.gov/DocuWare/Platform/WebClient/",
			Note:       "Statewide DocuWare vital records index; town clerks also hold records.",
		},
		{
			// Penobscot County — Maine CDC DRVS statewide index.
			CountyFIPS: "23019",
			CourtName:  "Maine CDC — Division of Vital Records",
			CourtURL:   "https://www.maine.gov/dhhs/mecdc/vital-records",
			SearchURL:  "https://docuware.maine.gov/DocuWare/Platform/WebClient/",
			Note:       "Statewide DocuWare vital records index; Bangor City Clerk also holds local records.",
		},
		{
			// Kennebec County — Maine CDC DRVS statewide index.
			CountyFIPS: "23011",
			CourtName:  "Maine CDC — Division of Vital Records",
			CourtURL:   "https://www.maine.gov/dhhs/mecdc/vital-records",
			SearchURL:  "https://docuware.maine.gov/DocuWare/Platform/WebClient/",
			Note:       "Statewide DocuWare vital records index; Augusta City Clerk also holds local records.",
		},
	},

	// --- Church (Catholic dioceses + parishes) ---------------------------
	Dioceses: []DioceseDef{
		{Slug: "portland_me", Name: "Diocese of Portland", Type: "diocese",
			Website: "https://www.portlanddiocese.org", Directory: "https://portlanddiocese.org/find-a-parish", HubCityID: "city_portland_me"},
	},

	// Portland-area parishes in the Diocese of Portland.
	// Part of the Portland Peninsula & Island Parishes cluster
	// (portlandcatholic.org). Names verified from the diocese's parish
	// finder + the cluster website.
	Parishes: []ParishDef{
		{
			DioceseSlug: "portland_me", Name: "Cathedral of the Immaculate Conception",
			Address: "307 Congress St, Portland, ME 04101",
		},
		{
			DioceseSlug: "portland_me", Name: "St. Peter Church",
			Address: "672 Federal St, Portland, ME 04101",
		},
		{
			DioceseSlug: "portland_me", Name: "St. Louis Church",
			Address: "279 Danforth St, Portland, ME 04101",
		},
		{
			DioceseSlug: "portland_me", Name: "Sacred Heart / St. Dominic Parish",
			Address: "65 Mellen St, Portland, ME 04101",
		},
	},

	// --- Social (wedding-industry Instagram vendors) ---------------------
	Vendors: []VendorDef{
		{
			Name: "Casey Durgin Photography", OfficialURL: "https://caseydurginphotography.com/",
			Handle: "caseydurginphoto", SourceClass: "engagement_photographer",
			CityID: "city_portland_me", State: "ME", City: "Portland", Verified: "2026-08-01",
			TikTokHandle: "caseydurginphoto",
		},
		{
			Name: "Emily Leonard Photography", OfficialURL: "https://emilyleonardphotography.com/",
			Handle: "emilyleonardphotography", SourceClass: "engagement_photographer",
			CityID: "city_portland_me", State: "ME", City: "Portland", Verified: "2026-08-01",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-photographers/emily-leonard-photography-1405115",
		},
		{
			Name: "O'Maine Studios", OfficialURL: "https://omainestudios.com/",
			Handle: "omainestudios", SourceClass: "wedding_venue",
			CityID: "city_portland_me", State: "ME", City: "Portland", Verified: "2026-08-01",
			TikTokHandle: "omainestudios",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-venues/omaine-studios-2906935",
		},
		{
			Name: "Linda Barry Photography", OfficialURL: "https://lindabarryphotography.com/",
			Handle: "linda__barry", SourceClass: "engagement_photographer",
			CityID: "city_portland_me", State: "ME", City: "Portland", Verified: "2026-08-03",
		},
		{
			Name: "Wintertide Floral", OfficialURL: "https://www.wintertidefloral.com/",
			Handle: "wintertidefloral", SourceClass: "florist",
			CityID: "city_portland_me", State: "ME", City: "Portland", Verified: "2026-08-03",
		},
		{
			Name: "Field Floral Studio", OfficialURL: "https://www.fieldfloralstudio.com/",
			Handle: "fieldfloralstudio", SourceClass: "florist",
			CityID: "city_portland_me", State: "ME", City: "Portland", Verified: "2026-08-03",
		},
		{
			Name: "The Tarratine", OfficialURL: "https://www.tarratinebangor.com/",
			Handle: "thetarratine", SourceClass: "wedding_venue",
			CityID: "city_bangor_me", State: "ME", City: "Bangor", Verified: "2026-08-03",
		},
		{
			Name: "Olive and Co Events", OfficialURL: "https://oliveandcoevents.com/",
			Handle: "oliveandcoevents", SourceClass: "wedding_planner",
			CityID: "city_portland_me", State: "ME", City: "Portland", Verified: "2026-08-03",
		},
		{
			Name: "Andrea's Bridal", OfficialURL: "https://www.andreasbridalmaine.com/",
			Handle: "andreasbridal", SourceClass: "bridal_shop",
			CityID: "city_portland_me", State: "ME", City: "Portland", Verified: "2026-08-03",
		},
	},
}
