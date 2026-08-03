package packs

// Vermont source pack — verified 2026-08-01.
//
// Government: Vermont marriage records are held by the state Dept of Health
// (vital records from 1909+) and town/city clerks. The VT State Archives
// (VSARA) also issues certified copies for marriages 2013 and earlier.
// Noncertified copies are free from town clerks.
//
// Church: Diocese of Burlington verified via USCCB + vermontcatholic.org.
// Burlington-area parishes verified against the diocese's own parish finder
// (vermontcatholic.org/parishes). The diocese publishes a weekly bulletin,
// "The Inland See," at vermontcatholic.org.
//
// Social: Instagram handles verified from each business's own public website
// social links (or from IG search results where the site is JS-rendered).

var vtPack = StatePack{
	State: "VT",

	// --- Cities ----------------------------------------------------------
	Cities: []CityDef{
		{ID: "city_burlington_vt", State: "VT", County: "50007", Name: "Burlington",
			Lat: 44.4759, Lng: -73.2121, Markets: []string{"burlington", "vt", "vermont", "chittenden"}},
	},

	// --- Government (state vital records + town clerks) -----------------
	Government: []GovSource{
		{
			// Chittenden County (Burlington) — city clerk + state vital records.
			CountyFIPS: "50007",
			CourtName:  "Burlington City Clerk",
			CourtURL:   "https://www.burlingtonvt.gov/Clerk",
			SearchURL:  "https://healthvermont.org/stats/vital-records/order-vital-records",
			Note:       "City clerk issues marriage licenses & certificates; VT Dept of Health holds statewide records from 1909.",
		},
		{
			// Rutland County — VT Dept of Health statewide vital records.
			CountyFIPS: "50021",
			CourtName:  "Vermont Dept of Health — Vital Records",
			CourtURL:   "https://healthvermont.org/stats/vital-records",
			SearchURL:  "https://secure.vermont.gov/VSARA/vitalrecords/applicant.php",
			Note:       "Statewide vital records; town clerks also hold records; VSARA handles pre-2014 marriages.",
		},
		{
			// Washington County — VT Dept of Health statewide vital records.
			CountyFIPS: "50023",
			CourtName:  "Vermont Dept of Health — Vital Records",
			CourtURL:   "https://healthvermont.org/stats/vital-records",
			SearchURL:  "https://secure.vermont.gov/VSARA/vitalrecords/applicant.php",
			Note:       "Statewide vital records; Montpelier City Clerk also holds local records.",
		},
		{
			// Windsor County — VT Dept of Health statewide vital records.
			CountyFIPS: "50027",
			CourtName:  "Vermont Dept of Health — Vital Records",
			CourtURL:   "https://healthvermont.org/stats/vital-records",
			SearchURL:  "https://secure.vermont.gov/VSARA/vitalrecords/applicant.php",
			Note:       "Statewide vital records; town clerks also hold records.",
		},
	},

	// --- Church (Catholic dioceses + parishes) ---------------------------
	Dioceses: []DioceseDef{
		{Slug: "burlington", Name: "Diocese of Burlington", Type: "diocese",
			Website: "https://www.vermontcatholic.org", Directory: "https://www.vermontcatholic.org/parishes", HubCityID: "city_burlington_vt"},
	},

	// Burlington-area parishes in the Diocese of Burlington.
	// Names verified from the diocese's parish finder
	// (vermontcatholic.org/parishes/find-parish-mass-time).
	Parishes: []ParishDef{
		{
			DioceseSlug: "burlington", Name: "Cathedral of St. Joseph",
			Address: "29 Allen St, Burlington, VT 05401",
		},
		{
			DioceseSlug: "burlington", Name: "St. Mark Church",
			Address:     "1251 North Ave, Burlington, VT 05401",
			BulletinURL: "https://www.stmarksvt.com",
		},
		{
			DioceseSlug: "burlington", Name: "Christ the King Church",
			Address: "136 Locust St, Burlington, VT 05401",
		},
	},

	// --- Social (wedding-industry Instagram vendors) ---------------------
	Vendors: []VendorDef{
		{
			Name: "Zack Griswold Photography", OfficialURL: "https://zackgriswold.com/",
			Handle: "zgriswold", SourceClass: "engagement_photographer",
			CityID: "city_burlington_vt", State: "VT", City: "Burlington", Verified: "2026-08-01",
			TikTokHandle: "zgriswold",
		},
		{
			Name: "The Portrait Gallery", OfficialURL: "https://portraitgallery-vt.com/",
			Handle: "portraitgalleryvt", SourceClass: "engagement_photographer",
			CityID: "city_burlington_vt", State: "VT", City: "Burlington", Verified: "2026-08-01",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-photographers/the-portrait-gallery-8709187",
		},
		{
			Name: "Shelburne Farms", OfficialURL: "https://shelburnefarms.org/",
			Handle: "shelburnefarms", SourceClass: "wedding_venue",
			CityID: "city_burlington_vt", State: "VT", City: "Burlington", Verified: "2026-08-01",
			TikTokHandle: "shelburnefarms",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-venues/shelburne-farms-5260807",
		},
		{
			Name: "Clayton Floral", OfficialURL: "https://claytonandcovt.com/",
			Handle: "claytonfloral", SourceClass: "florist",
			CityID: "city_burlington_vt", State: "VT", City: "Burlington", Verified: "2026-08-03",
		},
		{
			Name: "Hotel Champlain", OfficialURL: "https://www.hotelchamplainvermont.com/gather/weddings-celebrations/",
			Handle: "hotelchamplainvt", SourceClass: "wedding_venue",
			CityID: "city_burlington_vt", State: "VT", City: "Burlington", Verified: "2026-08-03",
		},
	},
}
