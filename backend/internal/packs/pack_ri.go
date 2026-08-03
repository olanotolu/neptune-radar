package packs

// Rhode Island source pack — verified 2026-08-01.
//
// Government: RI marriage records are held by the state. The State Archives
// maintains open vital records (marriages >100 years old) with an online
// digital archive. The Dept of Health Center for Vital Records holds recent
// records. City/town clerks also issue certified copies.
//
// Church: Diocese of Providence verified via USCCB + dioceseofprovidence.org.
// Providence-area parishes verified against the diocese's parish finder
// (dioceseofprovidence.org/parishfinder) + each parish's own website.
// Bulletin URLs verified by direct search.
//
// Social: Instagram handles verified from each business's own public website
// social links (or from IG search results where the site is JS-rendered).

var riPack = StatePack{
	State: "RI",

	// --- Cities ----------------------------------------------------------
	Cities: []CityDef{
	{
		ID: "city_newport_ri", State: "RI", County: "44005", Name: "Newport",
		Lat: 41.4901, Lng: -71.3128,
		Markets: []string{"newport", "aquidneck"},
	},
		{ID: "city_providence_ri", State: "RI", County: "44007", Name: "Providence",
			Lat: 41.8240, Lng: -71.4128, Markets: []string{"providence", "ri", "rhodeisland", "providencecounty"}},
	},

	// --- Government (state archives + dept of health) --------------------
	Government: []GovSource{
		{
			// Providence County — RI State Archives digital archive + DOH.
			CountyFIPS: "44007",
			CourtName:  "RI State Archives — Vital Records",
			CourtURL:   "https://www.sos.ri.gov/divisions/state-archives/VitalRecords",
			SearchURL:  "https://sosri.access.preservica.com",
			Note:       "State Archives has marriage records 1853-1925 online; DOH holds recent records (health.ri.gov/records).",
		},
		{
			// Kent County — RI State Archives + DOH.
			CountyFIPS: "44003",
			CourtName:  "RI State Archives — Vital Records",
			CourtURL:   "https://www.sos.ri.gov/divisions/state-archives/VitalRecords",
			SearchURL:  "https://sosri.access.preservica.com",
			Note:       "State Archives has marriage records 1853-1925 online; town clerks also hold records.",
		},
		{
			// Newport County — RI State Archives + DOH.
			CountyFIPS: "44005",
			CourtName:  "RI State Archives — Vital Records",
			CourtURL:   "https://www.sos.ri.gov/divisions/state-archives/VitalRecords",
			SearchURL:  "https://sosri.access.preservica.com",
			Note:       "State Archives has marriage records 1853-1925 online; town clerks also hold records.",
		},
		{
			// Washington County — RI State Archives + DOH.
			CountyFIPS: "44009",
			CourtName:  "RI State Archives — Vital Records",
			CourtURL:   "https://www.sos.ri.gov/divisions/state-archives/VitalRecords",
			SearchURL:  "https://sosri.access.preservica.com",
			Note:       "State Archives has marriage records 1853-1925 online; town clerks also hold records.",
		},
	},

	// --- Church (Catholic dioceses + parishes) ---------------------------
	Dioceses: []DioceseDef{
		{Slug: "providence", Name: "Diocese of Providence", Type: "diocese",
			Website: "https://www.dioceseofprovidence.org", Directory: "https://dioceseofprovidence.org/parishfinder", HubCityID: "city_providence_ri"},
	},

	// Providence-area parishes in the Diocese of Providence.
	// Names verified from the diocese's parish finder
	// (dioceseofprovidence.org/parishfinder). Bulletin URLs verified by
	// direct search for each parish's bulletin archive.
	Parishes: []ParishDef{
		{
			DioceseSlug: "providence", Name: "Cathedral of SS. Peter and Paul",
			Address: "30 Fenner St, Providence, RI 02903",
		},
		{
			DioceseSlug: "providence", Name: "St. Augustine Church",
			Address:     "636 Mount Pleasant Ave, Providence, RI 02908",
			BulletinURL: "https://churchofsaintaugustineprov.com/bulletin/",
		},
		{
			DioceseSlug: "providence", Name: "St. Joseph Church",
			Address:     "92 Hope St, Providence, RI 02906",
			BulletinURL: "https://www.stjosephprovidence.org/bulletin",
		},
		{
			DioceseSlug: "providence", Name: "St. Sebastian Church",
			Address: "67 Cole Ave, Providence, RI 02906",
		},
	},

	// --- Social (wedding-industry Instagram vendors) ---------------------
	Vendors: []VendorDef{
		{
			Name: "Sara Zarrella Photography", OfficialURL: "https://sarazarrella.com/",
			Handle: "sarazarrellaphotography", SourceClass: "engagement_photographer",
			CityID: "city_providence_ri", State: "RI", City: "Providence", Verified: "2026-08-01",
			TikTokHandle: "sarazarrellaphotography",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-photographers/sara-zarrella-photography-7847303",
		},
		{
			Name: "Graduate Providence", OfficialURL: "https://graduatehotels.com/providence/",
			Handle: "graduateprovidence", SourceClass: "wedding_venue",
			CityID: "city_providence_ri", State: "RI", City: "Providence", Verified: "2026-08-01",
			TikTokHandle: "graduateprovidence",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-venues/graduate-providence-9887102",
		},
		{
			Name: "Providence Diamond", OfficialURL: "https://www.providencediamond.com/",
			Handle: "providencediamond", SourceClass: "jeweler",
			CityID: "city_providence_ri", State: "RI", City: "Providence", Verified: "2026-08-01",
		},
		{
			Name: "Madilacie Photography", OfficialURL: "https://madilaciephotography.com/",
			Handle: "madilaciephotography", SourceClass: "engagement_photographer",
			CityID: "city_providence_ri", State: "RI", City: "Providence", Verified: "2026-08-03",
		},
		{
			Name: "OceanCliff", OfficialURL: "https://www.newportexperience.com/venues/oceancliff/",
			Handle: "oceancliffnewport", SourceClass: "wedding_venue",
			CityID: "city_newport_ri", State: "RI", City: "Newport", Verified: "2026-08-03",
		},
	},
}
