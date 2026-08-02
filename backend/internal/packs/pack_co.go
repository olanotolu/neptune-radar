package packs

// Colorado source pack — verified 2026-08-01.
//
// Government: Colorado marriage records are held by the county clerk &
// recorder. Search URLs for the top 8 counties by population were verified
// against each county's official .gov site or its Kofile/publicsearch portal.
//
// Church: all 3 Colorado Catholic dioceses/archdioceses verified via USCCB +
// each diocese's own website. Denver-area parishes (Archdiocese of Denver)
// were verified against the archdiocese's parish locator + each parish's own
// website. Bulletin URLs verified by direct search for each parish's bulletin
// archive.
//
// Social: Instagram handles verified from each business's own public website
// social links. Verification date recorded per vendor.

var coPack = StatePack{
	State: "CO",

	// --- Cities ----------------------------------------------------------
	Cities: []CityDef{
		{ID: "city_denver_co", State: "CO", County: "08031", Name: "Denver",
			Lat: 39.7392, Lng: -104.9903, Markets: []string{"denver", "den", "denvercounty", "colorado"}},
		{ID: "city_colorado_springs_co", State: "CO", County: "08041", Name: "Colorado Springs",
			Lat: 38.8339, Lng: -104.8214, Markets: []string{"coloradosprings", "cos", "elpasocounty"}},
	},

	// --- Government (county clerk marriage-record searches) --------------
	Government: []GovSource{
		{
			// Denver County — Kofile online document search portal; marriage
			// licenses digitised back to 1859.
			CountyFIPS: "08031",
			CourtName:  "Denver Clerk & Recorder",
			CourtURL:   "https://denvergov.org/Government/Agencies-Departments-Offices/Agencies-Departments-Offices-Directory/Denver-Clerk-and-Recorder/Recording-Division",
			SearchURL:  "https://countyfusion3.kofiletech.us/countyweb/loginDisplay.action?countyname=Denver",
			Note:       "Kofile online document search; marriage licenses fully digitised; enumeration candidate.",
		},
		{
			// El Paso County — public record search portal; marriages from
			// 05/01/1991 to present.
			CountyFIPS: "08041",
			CourtName:  "El Paso County Clerk & Recorder",
			CourtURL:   "https://clerkandrecorder.elpasoco.com/recording/",
			SearchURL:  "https://publicrecordsearch.elpasoco.com/",
			Note:       "Public record search with marriage records from 1991; enumeration candidate.",
		},
		{
			// Arapahoe County — official record search via publicsearch.us;
			// marriage licenses from Feb 1996 to present.
			CountyFIPS: "08005",
			CourtName:  "Arapahoe County Clerk & Recorder",
			CourtURL:   "https://www.arapahoeco.gov/your_county/county_departments/clerk_and_recorder/recording/index.php",
			SearchURL:  "https://arapahoe.co.publicsearch.us/",
			Note:       "Official record search portal; marriage licenses from 1996; enumeration capability needs testing.",
		},
		{
			// Jefferson County — Landmark Web marriage index search;
			// marriages from 2000 forward (older via subscription).
			CountyFIPS: "08059",
			CourtName:  "Jefferson County Clerk & Recorder",
			CourtURL:   "https://www.jeffco.us/1023/Marriage-Licenses-Civil-Unions",
			SearchURL:  "https://landrecords.co.jefferson.co.us/Marriage/SearchEntry.aspx",
			Note:       "Dedicated marriage index search page; records from 2000; enumeration candidate.",
		},
		{
			// Adams County — Landmark Web quick search; marriage licenses
			// included in recorded documents.
			CountyFIPS: "08001",
			CourtName:  "Adams County Clerk & Recorder",
			CourtURL:   "https://adamscountyco.gov/our-county/elected-officials/clerk-recorder/recording/",
			SearchURL:  "https://recording.adcogov.org/",
			Note:       "Landmark Web document search; marriage licenses searchable; enumeration capability needs testing.",
		},
		{
			// Boulder County — Kofile ds search; marriage licenses and civil
			// unions included in recorded documents.
			CountyFIPS: "08013",
			CourtName:  "Boulder County Clerk & Recorder",
			CourtURL:   "https://bouldercounty.gov/records/licenses/marriages-and-civil-unions/",
			SearchURL:  "https://boulder.co.ds.kofile.systems/",
			Note:       "Kofile public records search; marriage licenses searchable; enumeration capability needs testing.",
		},
		{
			// Douglas County — Landmark Web official records search;
			// marriage licenses included in recorded documents.
			CountyFIPS: "08035",
			CourtName:  "Douglas County Clerk & Recorder",
			CourtURL:   "https://www.douglasco.gov/recording/marriage-licenses/",
			SearchURL:  "https://apps.douglas.co.us/LandmarkWeb/",
			Note:       "Landmark Web document search; marriage licenses searchable; enumeration capability needs testing.",
		},
		{
			// Larimer County — Easy Access to recorded documents; marriage
			// licenses included in searchable document types.
			CountyFIPS: "08069",
			CourtName:  "Larimer County Clerk & Recorder",
			CourtURL:   "https://www.larimer.gov/clerk/recording/marriage",
			SearchURL:  "https://www.larimer.gov/clerk/recording/easy-access",
			Note:       "Easy Access recorded documents search; marriage licenses included; enumeration capability needs testing.",
		},
	},

	// --- Church (Catholic dioceses + parishes) ---------------------------
	Dioceses: []DioceseDef{
		{Slug: "denver", Name: "Archdiocese of Denver", Type: "archdiocese",
			Website: "https://archden.org", Directory: "https://archden.org/parish-locator/", HubCityID: "city_denver_co"},
		{Slug: "colorado_springs", Name: "Diocese of Colorado Springs", Type: "diocese",
			Website: "https://www.diocs.org", Directory: "https://www.diocs.org/parishes", HubCityID: "city_colorado_springs_co"},
		{Slug: "pueblo", Name: "Diocese of Pueblo", Type: "diocese",
			Website: "https://www.dioceseofpueblo.org", Directory: "https://www.dioceseofpueblo.org/parishes"},
	},

	// Denver-area parishes in the Archdiocese of Denver. Names and addresses
	// verified from the archdiocese's parish locator + each parish's own
	// website. Bulletin URLs verified by direct search for each parish's
	// bulletin archive.
	Parishes: []ParishDef{
		{DioceseSlug: "denver", Name: "Cathedral Basilica of the Immaculate Conception", Address: "1535 Logan St, Denver, CO 80203"},
		{
			DioceseSlug: "denver", Name: "Notre Dame Catholic Parish",
			Address:     "2190 S Sheridan Blvd, Denver, CO 80219",
			BulletinURL: "https://www.denvernotredame.org/bulletins",
		},
		{
			DioceseSlug: "denver", Name: "St. James Catholic Church",
			Address:     "1311 Oneida St, Denver, CO 80220",
			BulletinURL: "https://stjamesdenver.org/bulletins/",
		},
		{
			DioceseSlug: "denver", Name: "St. Ignatius of Loyola Catholic Church",
			Address:     "2309 N Gaylord St, Denver, CO 80205",
			BulletinURL: "https://www.stignatiusdenver.org/bulletins",
		},
		{
			DioceseSlug: "denver", Name: "Holy Name Catholic Parish",
			Address:     "3290 W Milan Ave, Sheridan, CO 80110",
			BulletinURL: "https://www.holynamedenver.org/bulletins",
		},
		{
			DioceseSlug: "denver", Name: "St. Gianna Beretta Molla Catholic Church",
			Address:     "6890 Argonne St, Unit B, Denver, CO 80249",
			BulletinURL: "https://www.stgiannadenver.org/bulletins",
		},
		{DioceseSlug: "denver", Name: "Spirit of Christ Catholic Community", Address: "7400 W 80th Ave, Arvada, CO 80003"},
		{DioceseSlug: "denver", Name: "St. Thomas Aquinas Catholic Center", Address: "898 14th St, Boulder, CO 80302"},
	},

	// --- Social (wedding-industry Instagram vendors) ---------------------
	Vendors: []VendorDef{
		// Denver photographers
		{
			Name: "Brittany Ann Photography", OfficialURL: "https://brittanyannphotography.com/",
			Handle: "brittany.ann.photos", SourceClass: "engagement_photographer",
			CityID: "city_denver_co", State: "CO", City: "Denver", Verified: "2026-08-01",
			TikTokHandle: "brittany.ann.photos",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-photographers/brittany-ann-photography-4551811",
		},
		{
			Name: "Mat Schramm Photography", OfficialURL: "https://matschrammphoto.com/",
			Handle: "matschrammphoto", SourceClass: "engagement_photographer",
			CityID: "city_denver_co", State: "CO", City: "Denver", Verified: "2026-08-01",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-photographers/mat-schramm-photography-9768007",
		},
		{
			Name: "Madeline J Studios", OfficialURL: "https://www.madelinejstudios.com/",
			Handle: "madelinejstudios", SourceClass: "engagement_photographer",
			CityID: "city_denver_co", State: "CO", City: "Denver", Verified: "2026-08-01",
			TikTokHandle: "madelinejstudios",
		},
		// Boulder photographer
		{
			Name: "Abby Shepard Photography", OfficialURL: "https://abbyshepardphotography.com/",
			Handle: "abby.shepard.photography", SourceClass: "engagement_photographer",
			CityID: "city_denver_co", State: "CO", City: "Boulder", Verified: "2026-08-01",
		},
		// Denver venues
		{
			Name: "Mile High Station", OfficialURL: "https://milehighstation.com/",
			Handle: "milehighstation", SourceClass: "wedding_venue",
			CityID: "city_denver_co", State: "CO", City: "Denver", Verified: "2026-08-01",
			TikTokHandle: "milehhstation",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-venues/mile-high-station-7792058",
		},
		{
			Name: "Ironworks", OfficialURL: "https://ironworksdenver.co/",
			Handle: "ironworksdenver", SourceClass: "wedding_venue",
			CityID: "city_denver_co", State: "CO", City: "Denver", Verified: "2026-08-01",
		},
		// Boulder venue
		{
			Name: "Hotel Boulderado", OfficialURL: "https://www.boulderado.com/",
			Handle: "hotel.boulderado", SourceClass: "wedding_venue",
			CityID: "city_denver_co", State: "CO", City: "Boulder", Verified: "2026-08-01",
			TikTokHandle: "hotel.boulderado",
		},
		// Denver jewelers
		{
			Name: "Sarah O. Jewelry", OfficialURL: "https://www.sarahojewelry.com/",
			Handle: "sarahojewelry", SourceClass: "jeweler",
			CityID: "city_denver_co", State: "CO", City: "Denver", Verified: "2026-08-01",
		},
		{
			Name: "Abby Sparks Jewelry", OfficialURL: "https://www.abbysparks.com/",
			Handle: "abbysparksjewelry", SourceClass: "jeweler",
			CityID: "city_denver_co", State: "CO", City: "Denver", Verified: "2026-08-01",
		},
	},
}
