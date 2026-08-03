package packs

// South Dakota source pack — verified 2026-08-01.
//
// Government: South Dakota marriage records are held by the county Register of
// Deeds. Per SD law, birth/death/marriage records are NOT open to the public;
// copies are issued as certified or informational copies to eligible requesters
// for a $15 fee. There is no online public search portal — all seven county
// URLs below point at each Register of Deeds' vital-records request page. The
// SD Dept of Health (doh.sd.gov) performs statewide searches by mail for $15.
// FamilySearch hosts a free marriage index (1950–2016) sourced from the SD DOH.
//
// Church: both SD dioceses verified via USCCB + each diocese's own website.
// Sioux Falls-area parishes verified against the diocese parish directory
// (sfcatholic.org/parishes) + direct parish website / bulletin-archive
// discovery.
//
// Social: Instagram handles verified from each business's own public website
// social links. Verification date recorded per vendor.

var sdPack = StatePack{
	State: "SD",

	// --- Cities ----------------------------------------------------------
	Cities: []CityDef{
		{ID: "city_sioux_falls_sd", State: "SD", County: "46099", Name: "Sioux Falls",
			Lat: 43.5510, Lng: -96.7003, Markets: []string{"siouxfalls", "sd", "southdakota", "minnehaha"}},
		{ID: "city_rapid_city_sd", State: "SD", County: "46081", Name: "Rapid City",
			Lat: 44.0805, Lng: -103.2310, Markets: []string{"rapidcity", "sd", "pennington"}},
	},

	// --- Government (county Register of Deeds vital-records requests) -----
	// ponytail: SD has no online marriage-record search; URLs point at the
	// ROD vital-records request page for each county. Enumeration would require
	// a research card from the SD Genealogical Society + in-person index access.
	Government: []GovSource{
		{
			// Minnehaha County (Sioux Falls) — Register of Deeds vital records.
			CountyFIPS: "46099",
			CourtName:  "Minnehaha County Register of Deeds",
			CourtURL:   "https://www.minnehahacounty.gov/dept/rd/rd.php",
			SearchURL:  "https://www.minnehahacounty.gov/dept/rd/vital_records/vital_records.php",
			Note:       "Marriage records from 1872; request-oriented, $15 fee, not public. VitalChek online ordering available.",
		},
		{
			// Pennington County (Rapid City) — Register of Deeds vital records.
			CountyFIPS: "46081",
			CourtName:  "Pennington County Register of Deeds",
			CourtURL:   "https://www.pennco.org/services/register_of_deeds/index.php",
			SearchURL:  "https://pennco.org/?SEC=79563912-2DBA-4F55-B295-E3CCDF58A465",
			Note:       "Marriage licenses and certified copies; request-oriented, $15 fee, not public.",
		},
		{
			// Lincoln County (Canton) — Register of Deeds vital records.
			CountyFIPS: "46083",
			CourtName:  "Lincoln County Register of Deeds",
			CourtURL:   "https://www.lincolncountysd.gov/190/Register-of-Deeds",
			SearchURL:  "https://lincolncountysd.gov/197/Vital-Records---Birth-Death-Marriage",
			Note:       "Marriage records from 1887; request-oriented, $15 fee, not public. VitalChek online ordering available.",
		},
		{
			// Brown County (Aberdeen) — Register of Deeds vital records.
			CountyFIPS: "46013",
			CourtName:  "Brown County Register of Deeds",
			CourtURL:   "https://www.brown.sd.us/department/register-deeds",
			SearchURL:  "https://www.brown.sd.us/department/register-deeds/vital-records",
			Note:       "Marriage records from 1885; request-oriented, $15 fee, not public.",
		},
		{
			// Brookings County — Register of Deeds (marriage licenses + vital records).
			CountyFIPS: "46011",
			CourtName:  "Brookings County Register of Deeds",
			CourtURL:   "https://www.brookingscountysd.gov/228/Register-of-Deeds",
			SearchURL:  "https://www.brookingscountysd.gov/228/Register-of-Deeds",
			Note:       "Issues marriage licenses ($40) and certified copies ($15); request-oriented, not public.",
		},
		{
			// Codington County (Watertown) — Register of Deeds.
			CountyFIPS: "46029",
			CourtName:  "Codington County Register of Deeds",
			CourtURL:   "https://codington.sdcounty.gov/register-of-deeds/",
			SearchURL:  "https://codington.sdcounty.gov/register-of-deeds/",
			Note:       "Marriage licenses ($40 cash) and certified copies ($15); request-oriented, not public.",
		},
		{
			// Yankton County — Register of Deeds.
			CountyFIPS: "46135",
			CourtName:  "Yankton County Register of Deeds",
			CourtURL:   "http://www.co.yankton.sd.us/custom/register-of-deeds",
			SearchURL:  "http://www.co.yankton.sd.us/custom/register-of-deeds",
			Note:       "Marriage records from 1889; request-oriented, $15 fee, not public. VitalChek online ordering available.",
		},
	},

	// --- Church (Catholic dioceses + parishes) ---------------------------
	Dioceses: []DioceseDef{
		{Slug: "sioux_falls", Name: "Diocese of Sioux Falls", Type: "diocese",
			Website: "https://sfcatholic.org", Directory: "https://sfcatholic.org/parishes/", HubCityID: "city_sioux_falls_sd"},
		{Slug: "rapid_city", Name: "Diocese of Rapid City", Type: "diocese",
			Website: "https://www.rapidcitydiocese.org", Directory: "https://www.rapidcitydiocese.org/parishes", HubCityID: "city_rapid_city_sd"},
	},

	// Sioux Falls-area parishes in the Diocese of Sioux Falls. Names and
	// addresses verified against the diocese parish directory
	// (sfcatholic.org/parishes) and each parish's own website. Bulletin URLs
	// verified by direct search for each parish's bulletin archive.
	Parishes: []ParishDef{
		{
			DioceseSlug: "sioux_falls", Name: "Cathedral of Saint Joseph",
			Address:     "521 N Duluth Ave, Sioux Falls, SD 57104",
			BulletinURL: "https://stjosephcathedral.net/",
		},
		{
			DioceseSlug: "sioux_falls", Name: "Christ the King Catholic Church",
			Address:     "1501 W 26th St, Sioux Falls, SD 57105",
			BulletinURL: "https://www.divinemercysf.org/",
		},
		{
			DioceseSlug: "sioux_falls", Name: "Holy Spirit Catholic Church",
			Address:     "3601 E Dudley Ln, Sioux Falls, SD 57103",
			BulletinURL: "https://www.holyspiritsf.org/bulletins",
		},
		{
			DioceseSlug: "sioux_falls", Name: "St. Mary Catholic Church",
			Address:     "2109 S 5th Ave, Sioux Falls, SD 57105",
			BulletinURL: "https://stmarysf.org/",
		},
		{
			DioceseSlug: "sioux_falls", Name: "St. Lambert Catholic Church",
			Address:     "1000 S Bahnson Ave, Sioux Falls, SD 57103",
			BulletinURL: "https://sites.google.com/sfcatholic.org/stlambertnew/home",
		},
		{
			DioceseSlug: "sioux_falls", Name: "St. Katharine Drexel Catholic Church",
			Address:     "1800 S Katie Ave Suite 1, Sioux Falls, SD 57106",
			BulletinURL: "https://www.stkdsfsd.org/",
		},
		{
			DioceseSlug: "sioux_falls", Name: "St. Michael Catholic Church",
			Address:     "1600 S Marion Rd, Sioux Falls, SD 57106",
			BulletinURL: "https://www.stmichaelsfsd.org/",
		},
		{
			DioceseSlug: "sioux_falls", Name: "Our Lady of Guadalupe Catholic Church",
			Address:     "1220 E 8th St, Sioux Falls, SD 57103",
			BulletinURL: "https://ourladyofguadalupesf.org/",
		},
	},

	// --- Social (wedding-industry Instagram vendors) ---------------------
	Vendors: []VendorDef{
		// Sioux Falls photographers
		{
			Name: "Ethan Wiese Photography", OfficialURL: "https://www.ethanwiese.com/",
			Handle: "wiese.ethan", SourceClass: "engagement_photographer",
			CityID: "city_sioux_falls_sd", State: "SD", City: "Sioux Falls", Verified: "2026-08-01",
			TikTokHandle: "wiese.ethan",
		},
		{
			Name: "Kylee Warrick Photography", OfficialURL: "https://www.kyleewarrickphotography.com/",
			Handle: "kyleewarrickphotography", SourceClass: "engagement_photographer",
			CityID: "city_sioux_falls_sd", State: "SD", City: "Sioux Falls", Verified: "2026-08-01",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-photographers/kylee-warrick-photography-1564825",
		},
		{
			Name: "Michael Liedtke Photography", OfficialURL: "https://michaelliedtke.com/",
			Handle: "michaelliedtke", SourceClass: "engagement_photographer",
			CityID: "city_sioux_falls_sd", State: "SD", City: "Sioux Falls", Verified: "2026-08-01",
			TikTokHandle: "michaelliedtke",
		},
		// Sioux Falls venues
		{
			Name: "The Social", OfficialURL: "https://www.thesocialsiouxfalls.com/",
			Handle: "thesocial_siouxfalls", SourceClass: "wedding_venue",
			CityID: "city_sioux_falls_sd", State: "SD", City: "Sioux Falls", Verified: "2026-08-01",
			TikTokHandle: "thesocial_siouxfalls",
		},
		{
			Name: "LuxeFalls Venue", OfficialURL: "https://www.luxefallsvenue.com/",
			Handle: "luxefalls_venue", SourceClass: "wedding_venue",
			CityID: "city_sioux_falls_sd", State: "SD", City: "Sioux Falls", Verified: "2026-08-01",
			TikTokHandle: "luxefalls_venue",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-venues/luxefalls-venue-2939679",
		},
		{
			Name: "Monick Yards", OfficialURL: "https://www.monickyards.com/",
			Handle: "monickyards", SourceClass: "wedding_venue",
			CityID: "city_sioux_falls_sd", State: "SD", City: "Sioux Falls", Verified: "2026-08-01",
		},
		// Sioux Falls jewelers
		{
			Name: "The Diamond Room by Spektor", OfficialURL: "https://www.thediamondroom.com/",
			Handle: "thediamondroombyspektor", SourceClass: "jeweler",
			CityID: "city_sioux_falls_sd", State: "SD", City: "Sioux Falls", Verified: "2026-08-01",
		},
		{
			Name: "Faini Designs Jewelry Studio", OfficialURL: "https://www.fainidesigns.com/",
			Handle: "faini_designs", SourceClass: "jeweler",
			CityID: "city_sioux_falls_sd", State: "SD", City: "Sioux Falls", Verified: "2026-08-01",
		},
		{
			Name: "Jase Dewald Photography", OfficialURL: "https://jasedewaldphoto.com/",
			Handle: "jasedewaldphoto", SourceClass: "engagement_photographer",
			CityID: "city_sioux_falls_sd", State: "SD", City: "Sioux Falls", Verified: "2026-08-03",
		},
		{
			Name: "Luke and Savannah Photography", OfficialURL: "https://lukeandsavannah.com/",
			Handle: "lukeandsavannah", SourceClass: "engagement_photographer",
			CityID: "city_sioux_falls_sd", State: "SD", City: "Sioux Falls", Verified: "2026-08-03",
		},
		{
			Name: "Thistle & Dot Floral", OfficialURL: "https://thistledotfloral.com/",
			Handle: "thistle.dot.floral", SourceClass: "florist",
			CityID: "city_sioux_falls_sd", State: "SD", City: "Sioux Falls", Verified: "2026-08-03",
		},
		{
			Name: "Bridal Gallery", OfficialURL: "https://www.bridalgallerysf.com/",
			Handle: "bridalgallerysiouxfalls", SourceClass: "bridal_shop",
			CityID: "city_sioux_falls_sd", State: "SD", City: "Sioux Falls", Verified: "2026-08-03",
		},
		{
			Name: "Diamond Spur Events Center", OfficialURL: "https://www.diamondspurevents.com/",
			Handle: "diamondspurevents", SourceClass: "wedding_venue",
			CityID: "city_rapid_city_sd", State: "SD", City: "Rapid City", Verified: "2026-08-03",
		},
	},
}
