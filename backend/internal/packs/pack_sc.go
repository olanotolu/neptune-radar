package packs

// South Carolina source pack — verified 2026-08-02.
//
// Government: SC marriage records are held by the county probate court.
// Search URLs for the top 7 counties by population were verified against each
// county's official .gov site or its southcarolinaprobate.net portal.
//
// Church: Diocese of Charleston verified via USCCB + the diocese's own
// directory (directory.charlestondiocese.org). Charleston-deanery parishes
// verified against the diocesan directory + each parish's own website.
// Bulletin URLs verified by direct search for each parish's bulletin archive.
//
// Social: Instagram handles verified from each business's own public website
// social links (or from IG search results where the site is JS-rendered and
// the handle was visible in the search snippet). Verification date recorded
// per vendor.

var scPack = StatePack{
	State: "SC",

	// --- Cities ----------------------------------------------------------
	Cities: []CityDef{
		{ID: "city_charleston_sc", State: "SC", County: "45019", Name: "Charleston",
			Lat: 32.7765, Lng: -79.9311, Markets: []string{"charleston", "chs", "sc", "lowcountry"}},
		{ID: "city_greenville_sc", State: "SC", County: "45045", Name: "Greenville",
			Lat: 34.8526, Lng: -82.3940, Markets: []string{"greenville", "gvl", "sc", "upstate"}},
		{ID: "city_columbia_sc", State: "SC", County: "45079", Name: "Columbia",
			Lat: 34.0007, Lng: -81.0348, Markets: []string{"columbia", "cola", "sc", "midlands"}},
		{ID: "city_myrtle_beach_sc", State: "SC", County: "45051", Name: "Myrtle Beach",
			Lat: 33.6891, Lng: -78.8867, Markets: []string{"myrtlebeach", "grandstrand", "sc"}},
		{ID: "city_spartanburg_sc", State: "SC", County: "45083", Name: "Spartanburg",
			Lat: 34.9496, Lng: -81.9320, Markets: []string{"spartanburg", "spart", "sc", "upstate"}},
		{ID: "city_lexington_sc", State: "SC", County: "45063", Name: "Lexington",
			Lat: 33.9815, Lng: -81.2362, Markets: []string{"lexington", "sc", "midlands"}},
		{ID: "city_rock_hill_sc", State: "SC", County: "45091", Name: "Rock Hill",
			Lat: 34.9249, Lng: -81.0251, Markets: []string{"rockhill", "sc", "upstate"}},
	},

	// --- Government (county probate court marriage-record searches) ------
	Government: []GovSource{
		{
			// Charleston County — probate court marriage license division with
			// online search page.
			CountyFIPS: "45019",
			CourtName:  "Charleston County Probate Court",
			CourtURL:   "https://ccprobate.charlestoncounty.gov/marriage-license.php",
			SearchURL:  "http://www3.charlestoncounty.org/docs/ProbateMain.html",
			Note:       "Probate Court online search page with marriage license search; enumeration candidate.",
		},
		{
			// Greenville County — probate court marriage license division with
			// online marriage license search portal.
			CountyFIPS: "45045",
			CourtName:  "Greenville County Probate Court",
			CourtURL:   "https://www.greenvillecounty.org/Probate/MarriageLicense.aspx",
			SearchURL:  "https://www.greenvillecounty.org/disclaimer/PublicRecords.aspx?DirURL=MLSearch",
			Note:       "Online marriage license search for certified copies; enumeration candidate.",
		},
		{
			// Richland County (Columbia) — probate court with online marriage
			// license certified copy search by party name.
			CountyFIPS: "45079",
			CourtName:  "Richland County Probate Court",
			CourtURL:   "https://www.richlandcountysc.gov/Courts-Safety/Probate-Court/Marriage",
			SearchURL:  "https://www.richlandcountysc.gov/Online-Services/Marriage-License-Inquiry",
			Note:       "Online marriage license search by party name; enumeration candidate.",
		},
		{
			// Horry County (Myrtle Beach) — probate court marriage license
			// division; no online search portal.
			CountyFIPS: "45051",
			CourtName:  "Horry County Probate Court",
			CourtURL:   "https://horrycountysc.gov/departments/probate-court",
			SearchURL:  "https://horrycountysc.gov/departments/probate-court/services/marriage-license/",
			Note:       "Marriage license info page; no online search portal, request-oriented.",
		},
		{
			// Spartanburg County — probate court; no online search, certified
			// copies by mail or in-person request only.
			CountyFIPS: "45083",
			CourtName:  "Spartanburg County Probate Court",
			CourtURL:   "https://www.spartanburgcounty.gov/153/Probate-Court",
			SearchURL:  "https://www.spartanburgcounty.org/433/Marriage-License-Requirements",
			Note:       "No online search; certified copies by mail or in-person request only.",
		},
		{
			// Lexington County — probate court with online marriage license
			// search (1986–present).
			CountyFIPS: "45063",
			CourtName:  "Lexington County Probate Court",
			CourtURL:   "https://www.lex-co.sc.gov/departments/probate-court",
			SearchURL:  "https://www.lex-co.sc.gov/departments/probate-court",
			Note:       "Online marriage license search (1986–present); enumeration candidate.",
		},
		{
			// York County (Rock Hill) — probate court; marriage licenses and
			// estate records searchable via southcarolinaprobate.net.
			CountyFIPS: "45091",
			CourtName:  "York County Probate Court",
			CourtURL:   "https://www.yorkcountygov.com/169/Probate-Court",
			SearchURL:  "https://www.southcarolinaprobate.net/search",
			Note:       "Marriage licenses and estate records searchable via southcarolinaprobate.net; enumeration candidate.",
		},
	},

	// --- Church (Catholic dioceses + parishes) ---------------------------
	Dioceses: []DioceseDef{
		{Slug: "charleston", Name: "Diocese of Charleston", Type: "diocese",
			Website: "https://charlestondiocese.org", Directory: "https://directory.charlestondiocese.org/directories/parish-directory/", HubCityID: "city_charleston_sc"},
	},

	// Charleston-deanery parishes in the Diocese of Charleston. Names and
	// addresses verified from the diocesan directory
	// (directory.charlestondiocese.org). Bulletin URLs verified by direct
	// search for each parish's bulletin archive.
	Parishes: []ParishDef{
		{
			DioceseSlug: "charleston", Name: "Cathedral of St. John the Baptist",
			Address:     "120 Broad Street, Charleston, SC 29401",
			BulletinURL: "https://charlestoncathedral.com",
		},
		{
			DioceseSlug: "charleston", Name: "St. Mary of the Annunciation",
			Address:     "89 Hasell Street, Charleston, SC 29401",
			BulletinURL: "https://www.sma.church/parish-bulletins.html",
		},
		{
			DioceseSlug: "charleston", Name: "Blessed Sacrament Catholic Church",
			Address:     "5 Saint Teresa Dr, Charleston, SC 29407",
			BulletinURL: "https://www.blsac.org/bulletin",
		},
		{
			DioceseSlug: "charleston", Name: "Church of the Nativity",
			Address:     "1061 Folly Road, Charleston, SC 29412",
			BulletinURL: "https://www.nativitycharleston.org/bulletin",
		},
		{
			DioceseSlug: "charleston", Name: "St. Patrick Catholic Church",
			Address:     "134 St. Philip Street, Charleston, SC 29403",
			BulletinURL: "https://www.stpatrickcharleston.org/bulletin",
		},
		{
			DioceseSlug: "charleston", Name: "St. Joseph Catholic Church",
			Address: "1695 Wallenberg Boulevard, Charleston, SC 29407",
		},
		{
			DioceseSlug: "charleston", Name: "Sacred Heart Catholic Church",
			Address: "888 King Street, Charleston, SC 29403",
		},
	},

	// --- Social (wedding-industry Instagram vendors) ---------------------
	Vendors: []VendorDef{
		// Charleston photographers
		{
			Name: "Lauren Jonas Photography", OfficialURL: "https://www.laurenjonas.com/",
			Handle: "laurenjonasweddings", SourceClass: "engagement_photographer",
			CityID: "city_charleston_sc", State: "SC", City: "Charleston", Verified: "2026-08-02",
			TikTokHandle: "laurenjonasweddings",
		},
		{
			Name: "Taylor Rae Photography", OfficialURL: "https://taylorraephotography.com/",
			Handle: "taylorraephoto", SourceClass: "engagement_photographer",
			CityID: "city_charleston_sc", State: "SC", City: "Charleston", Verified: "2026-08-02",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-photographers/taylor-rae-photography-3932171",
		},
		{
			Name: "Kendra Martin Photography", OfficialURL: "https://kendramartinphotography.com/",
			Handle: "kendramartinphotography", SourceClass: "engagement_photographer",
			CityID: "city_charleston_sc", State: "SC", City: "Charleston", Verified: "2026-08-02",
			TikTokHandle: "kendramartinphotography",
		},
		// Charleston venues
		{
			Name: "Cannon Green", OfficialURL: "https://cannongreencharleston.com/",
			Handle: "cannongreenevents", SourceClass: "wedding_venue",
			CityID: "city_charleston_sc", State: "SC", City: "Charleston", Verified: "2026-08-02",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-venues/cannon-green-6812017",
		},
		{
			Name: "Le James", OfficialURL: "https://lejamesvenue.com/",
			Handle: "lejamescharleston", SourceClass: "wedding_venue",
			CityID: "city_charleston_sc", State: "SC", City: "Charleston", Verified: "2026-08-02",
			TikTokHandle: "lejamescharleston",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-venues/le-james-6188842",
		},
		{
			Name: "Harborside East", OfficialURL: "https://www.harborsideeast.com/",
			Handle: "harborsideeast", SourceClass: "wedding_venue",
			CityID: "city_charleston_sc", State: "SC", City: "Charleston", Verified: "2026-08-02",
		},
		{
			Name: "The Governor Thomas Bennett House", OfficialURL: "https://www.governorthomasbennetthouse.com/",
			Handle: "govthomasbennetthouse", SourceClass: "wedding_venue",
			CityID: "city_charleston_sc", State: "SC", City: "Charleston", Verified: "2026-08-02",
			TikTokHandle: "govthomasbennetthouse",
		},
		// Charleston jeweler
		{
			Name: "Croghan's Jewel Box", OfficialURL: "https://www.croghansjewelbox.com/",
			Handle: "croghans", SourceClass: "jeweler",
			CityID: "city_charleston_sc", State: "SC", City: "Charleston", Verified: "2026-08-02",
		},
		{
			Name: "Taylor Jordan Photography", OfficialURL: "https://taylorjordanphotography.com/",
			Handle: "taylorjordanphoto", SourceClass: "engagement_photographer",
			CityID: "city_charleston_sc", State: "SC", City: "Charleston", Verified: "2026-08-03",
		},
		{
			Name: "Georgia & Micah Photography", OfficialURL: "https://georgiaandmicah.com/",
			Handle: "_gmphotoandfilm_", SourceClass: "engagement_photographer",
			CityID: "city_charleston_sc", State: "SC", City: "Charleston", Verified: "2026-08-03",
		},
		{
			Name: "Lowndes Grove", OfficialURL: "https://www.pphgcharleston.com/venues/lowndes-grove/",
			Handle: "pphgcharleston", SourceClass: "wedding_venue",
			CityID: "city_charleston_sc", State: "SC", City: "Charleston", Verified: "2026-08-03",
		},
		{
			Name: "Festoon Charleston", OfficialURL: "https://festooncharleston.com/",
			Handle: "festooncharleston", SourceClass: "florist",
			CityID: "city_charleston_sc", State: "SC", City: "Charleston", Verified: "2026-08-03",
		},
		{
			Name: "Firefly Weddings & Events", OfficialURL: "https://fireflywed.com/",
			Handle: "fireflywed_charleston", SourceClass: "wedding_planner",
			CityID: "city_charleston_sc", State: "SC", City: "Charleston", Verified: "2026-08-03",
		},
		{
			Name: "Emma Ivey Events", OfficialURL: "https://emmaiveyevents.com/",
			Handle: "emmaiveyevents", SourceClass: "wedding_planner",
			CityID: "city_charleston_sc", State: "SC", City: "Charleston", Verified: "2026-08-03",
		},
		{
			Name: "Vue 1919", OfficialURL: "https://vue1919.com/",
			Handle: "vue1919gvl", SourceClass: "wedding_venue",
			CityID: "city_greenville_sc", State: "SC", City: "Greenville", Verified: "2026-08-03",
		},
		{
			Name: "Statice Floral", OfficialURL: "https://www.staticeflowers.com/",
			Handle: "staticefloral", SourceClass: "florist",
			CityID: "city_greenville_sc", State: "SC", City: "Greenville", Verified: "2026-08-03",
		},
		{
			Name: "Nikki Morgan Photography", OfficialURL: "https://www.nikkimorganphotography.com/",
			Handle: "nikkimorganphotography", SourceClass: "engagement_photographer",
			CityID: "city_columbia_sc", State: "SC", City: "Columbia", Verified: "2026-08-03",
		},
		{
			Name: "Stone River Columbia", OfficialURL: "https://stonerivercolumbia.com/",
			Handle: "stonerivercolumbia", SourceClass: "wedding_venue",
			CityID: "city_columbia_sc", State: "SC", City: "Columbia", Verified: "2026-08-03",
		},
	},
}
