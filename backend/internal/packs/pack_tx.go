package packs

// Texas source pack — verified 2026-08-01.
//
// Government: Texas marriage records are held by the county clerk. Search URLs
// for the top 7 counties by population were verified against each county's
// official .gov site or its publicsearch.us portal.
//
// Church: all 15 Texas Catholic dioceses/archdioceses verified via USCCB +
// each diocese's own website. Parish directory URLs point at each
// jurisdiction's own parish-finder. Houston-area parishes (Archdiocese of
// Galveston-Houston) were verified against the Wikipedia list of churches in
// the archdiocese (which cites the archdiocese's own records) + direct
// bulletin-archive URL discovery.
//
// Social: Instagram handles verified from each business's own public website
// social links (or from IG search results where the site is JS-rendered and
// the handle was visible in the search snippet). Verification date recorded
// per vendor.

var txPack = StatePack{
	State: "TX",

	// --- Cities ----------------------------------------------------------
	Cities: []CityDef{
		{
			ID: "city_houston_tx", State: "TX", County: "48201", Name: "Houston",
			Lat: 29.7604, Lng: -95.3698,
			Markets: []string{"houston", "htx", "galveston", "sugarland", "katy"},
		},
		{
			ID: "city_dallas_tx", State: "TX", County: "48113", Name: "Dallas",
			Lat: 32.7767, Lng: -96.7970,
			Markets: []string{"dallas", "dtx", "fortworth", "plano", "frisco"},
		},
		{
			ID: "city_austin_tx", State: "TX", County: "48453", Name: "Austin",
			Lat: 30.2672, Lng: -97.7431,
			Markets: []string{"austin", "atx", "drippingsprings", "hillcountry"},
		},
		{
			ID: "city_san_antonio_tx", State: "TX", County: "48029", Name: "San Antonio",
			Lat: 29.4241, Lng: -98.4936,
			Markets: []string{"sanantonio", "satx", "boerne", "newbraunfels"},
		},
	},

	// --- Government (county clerk marriage-record searches) --------------
	Government: []GovSource{
		{
			// Harris County (Houston) — county clerk document search portal,
			// marriage (ceremonial & informal) category.
			CountyFIPS: "48201",
			CourtName:  "Harris County Clerk",
			CourtURL:   "https://www.cclerk.hctx.net",
			SearchURL:  "https://www.cclerk.hctx.net/applications/websearch/mal.aspx",
			Note:       "Official document search portal with marriage category; enumeration candidate.",
		},
		{
			// Dallas County — county clerk official record search via
			// publicsearch.us.
			CountyFIPS: "48113",
			CourtName:  "Dallas County Clerk",
			CourtURL:   "https://www.dallascounty.org/government/county-clerk/vital-records/marriage-license.php",
			SearchURL:  "https://dallas.tx.publicsearch.us/",
			Note:       "Official record search portal; marriage records searchable; enumeration capability needs testing.",
		},
		{
			// Travis County (Austin) — county clerk recording search.
			CountyFIPS: "48453",
			CourtName:  "Travis County Clerk",
			CourtURL:   "https://countyclerk.traviscountytx.gov",
			SearchURL:  "https://countyclerk.traviscountytx.gov/departments/recording/search-copies-of-records/",
			Note:       "Recording search + copies of records; marriage filtering needs testing.",
		},
		{
			// Bexar County (San Antonio) — public record search.
			CountyFIPS: "48029",
			CourtName:  "Bexar County Clerk",
			CourtURL:   "https://www.bexar.org/2946/County-Clerk",
			SearchURL:  "https://www.bexar.org/2984/Public-Record-Search",
			Note:       "Public record search page; marriage-record capability needs testing.",
		},
		{
			// Tarrant County (Fort Worth) — county clerk vital records.
			CountyFIPS: "48439",
			CourtName:  "Tarrant County Clerk",
			CourtURL:   "https://www.tarrantcountytx.gov/en/county-clerk/vital-records/marriage-licenses.html",
			SearchURL:  "https://countyfusion.tarrantcounty.com/",
			Note:       "CountyFusion record search portal; enumeration capability needs testing.",
		},
		{
			// Collin County — official record search via publicsearch.us.
			CountyFIPS: "48085",
			CourtName:  "Collin County Clerk",
			CourtURL:   "https://www.collincountytx.gov/county-clerk",
			SearchURL:  "https://collin.tx.publicsearch.us/",
			Note:       "Official record search portal; marriage records searchable; enumeration capability needs testing.",
		},
		{
			// El Paso County — marriage records search application.
			CountyFIPS: "48141",
			CourtName:  "El Paso County Clerk",
			CourtURL:   "https://www.epcounty.com/583/Vitals-Division-Birth-Death-Marriage",
			SearchURL:  "https://apps.epcountytx.gov/publicrecords/Marriages",
			Note:       "Dedicated marriage records search application; enumeration candidate.",
		},
		{
			CountyFIPS: "48157", // Fort Bend County
			CourtName:  "Fort Bend County Clerk",
			CourtURL:   "https://www.fortbendcountytx.gov/government/departments/county-clerk",
			SearchURL:  "https://ccweb.co.fort-bend.tx.us/Marriage/SearchEntry.aspx",
			Note:       "Online marriage license index search; requires applicant name or date of marriage; unofficial copies printable online.",
		},
		{
			CountyFIPS: "48339", // Montgomery County
			CourtName:  "Montgomery County Clerk",
			CourtURL:   "https://www.mctx.org/countyclerk",
			SearchURL:  "https://montgomery.tx.publicsearch.us/",
			Note:       "Public records portal via publicsearch.us; search by document number or party name with date range.",
		},
		{
			CountyFIPS: "48121", // Denton County
			CourtName:  "Denton County Clerk",
			CourtURL:   "https://www.dentoncounty.gov/173/County-Clerk",
			SearchURL:  "https://dentontx.search.kofile.com/48121/Home/Index/1",
			Note:       "Online marriage license index search at no cost; images not viewable online, must visit office.",
		},
		{
			CountyFIPS: "48491", // Williamson County
			CourtName:  "Williamson County Clerk",
			CourtURL:   "https://www.wilcotx.gov/countyclerk",
			SearchURL:  "https://williamsoncountytx-web.tylerhost.net/williamsonweb/user/disclaimer",
			Note:       "Official public records search via Tyler Technologies; marriage, birth, and death records searchable.",
		},
		{
			CountyFIPS: "48167", // Galveston County
			CourtName:  "Galveston County Clerk",
			CourtURL:   "https://www.galvestoncountytx.gov/our-county/county-clerk/county-clerk-86",
			SearchURL:  "https://txgalveston.fidlar.com/TXGalveston/Apex.WebPortal/search",
			Note:       "Marriage license and assumed name record search via Fidlar Technologies portal.",
		},
		{
			CountyFIPS: "48309", // McLennan County (Waco)
			CourtName:  "McLennan County Clerk",
			CourtURL:   "https://www.mclennan.gov/166/County-Clerk",
			SearchURL:  "https://mclennancountytx-web.tylerhost.net/web/user/disclaimer",
			Note:       "Online marriage records search from April 1878 to present; plain copies $1, certified copies $6.",
		},
		{
			CountyFIPS: "48039", // Brazoria County
			CourtName:  "Brazoria County Clerk",
			CourtURL:   "https://www.brazoriacountyclerktx.gov/",
			SearchURL:  "https://www.brazoriacountyclerktx.gov/search-records",
			Note:       "Requires registration and $5 fee per name for marriage record search; vital records require payment before search.",
		},
		{
			CountyFIPS: "48027", // Bell County (Killeen/Temple)
			CourtName:  "Bell County Clerk",
			CourtURL:   "https://bellcountytx.com/county_government/county_clerk/",
			SearchURL:  "https://bell.tx.publicsearch.us/",
			Note:       "Public records portal via publicsearch.us; certified copies ordered through Permitium.",
		},
		{
			CountyFIPS: "48245", // Jefferson County (Beaumont)
			CourtName:  "Jefferson County Clerk",
			CourtURL:   "https://jeffersoncountytx.gov/cclerk",
			SearchURL:  "https://jefferson.tx.publicsearch.us/",
			Note:       "Official records search via publicsearch.us; over 2.8 million records available free of charge.",
		},
		{
			CountyFIPS: "48355", // Nueces County (Corpus Christi)
			CourtName:  "Nueces County Clerk",
			CourtURL:   "https://www.co.nueces.tx.us/countyclerk/",
			SearchURL:  "https://nueces.tx.publicsearch.us/",
			Note:       "Official records search via publicsearch.us; search by department (Marriage) and party names.",
		},
	},

	// --- Church (Catholic dioceses + parishes) ---------------------------
	Dioceses: []DioceseDef{
		{
			Slug: "galveston_houston", Name: "Archdiocese of Galveston-Houston",
			Type: "archdiocese", Website: "https://archgh.org",
			Directory: "https://www.archgh.org/parishfinder",
			HubCityID: "city_houston_tx",
		},
		{
			Slug: "san_antonio", Name: "Archdiocese of San Antonio",
			Type: "archdiocese", Website: "https://archsa.org",
			Directory: "https://archsa.org/parishes/",
			HubCityID: "city_san_antonio_tx",
		},
		{
			Slug: "dallas", Name: "Diocese of Dallas",
			Type: "diocese", Website: "https://dallascatholic.org",
			Directory: "https://dallascatholic.org/community-finder/",
			HubCityID: "city_dallas_tx",
		},
		{
			Slug: "austin", Name: "Diocese of Austin",
			Type: "diocese", Website: "https://austindiocese.org",
			Directory: "https://www.austindiocese.org/parishfinder",
			HubCityID: "city_austin_tx",
		},
		{
			Slug: "fort_worth", Name: "Diocese of Fort Worth",
			Type: "diocese", Website: "https://fwdioc.org",
			Directory: "https://fwdioc.org/parish-finder",
		},
		{
			Slug: "el_paso", Name: "Diocese of El Paso",
			Type: "diocese", Website: "https://www.elpasodiocese.org",
			Directory: "https://www.elpasodiocese.org/parishes.html",
		},
		{
			Slug: "corpus_christi", Name: "Diocese of Corpus Christi",
			Type: "diocese", Website: "https://diocesecc.org",
			Directory: "https://diocesecc.org/parishfinder",
		},
		{
			Slug: "brownsville", Name: "Diocese of Brownsville",
			Type: "diocese", Website: "https://cdob.org",
			Directory: "https://cdob.org/parish-info",
		},
		{
			Slug: "laredo", Name: "Diocese of Laredo",
			Type: "diocese", Website: "https://dioceseoflaredo.org",
			Directory: "https://dioceseoflaredo.org/parishes/",
		},
		{
			Slug: "lubbock", Name: "Diocese of Lubbock",
			Type: "diocese", Website: "https://www.catholiclubbock.org",
			Directory: "https://www.catholiclubbock.org/Parishes.html",
		},
		{
			Slug: "san_angelo", Name: "Diocese of San Angelo",
			Type: "diocese", Website: "https://sanangelodiocese.org",
			Directory: "https://sanangelodiocese.org/parishes",
		},
		{
			Slug: "tyler", Name: "Diocese of Tyler",
			Type: "diocese", Website: "https://www.dioceseoftyler.org",
			Directory: "https://www.dioceseoftyler.org/parishes/",
		},
		{
			Slug: "victoria", Name: "Diocese of Victoria in Texas",
			Type: "diocese", Website: "https://victoriadiocese.org",
			Directory: "https://victoriadiocese.org/parishes-by-deanery",
		},
		{
			Slug: "beaumont", Name: "Diocese of Beaumont",
			Type: "diocese", Website: "https://dioceseofbmt.org",
			Directory: "https://dioceseofbmt.org/priest-directory",
		},
		{
			Slug: "amarillo", Name: "Diocese of Amarillo",
			Type: "diocese", Website: "https://amarillodiocese.org",
			Directory: "https://amarillodiocese.org/parishfinder",
		},
	},

	// Houston-area parishes in the Archdiocese of Galveston-Houston.
	// Names verified from Wikipedia's list of churches in the archdiocese
	// (which cites the archdiocese's own records). Bulletin URLs verified by
	// direct search for each parish's bulletin archive.
	Parishes: []ParishDef{
		{DioceseSlug: "galveston_houston", Name: "Co-Cathedral of the Sacred Heart", Address: "1111 St Joseph Pkwy, Houston, TX 77002"},
		{DioceseSlug: "galveston_houston", Name: "Annunciation Church", Address: "1618 Texas Ave, Houston, TX 77001"},
		{DioceseSlug: "galveston_houston", Name: "All Saints Church", Address: "204 E 10th St, Houston, TX 77007"},
		{
			DioceseSlug: "galveston_houston", Name: "Christ the Redeemer Catholic Church",
			Address:     "11507 Huffmeister Rd, Houston, TX 77065",
			BulletinURL: "https://ctrcc.com/bulletin",
		},
		{
			DioceseSlug: "galveston_houston", Name: "Saint Cecilia Catholic Church",
			Address:     "11720 Joan of Arc Dr, Houston, TX 77024",
			BulletinURL: "https://saintcecilia.org/bulletin",
		},
		{
			DioceseSlug: "galveston_houston", Name: "St. Thomas More Catholic Church",
			Address:     "11703 Holmes Rd, Houston, TX 77031",
			BulletinURL: "https://stmhouston.org/bulletins",
		},
		{
			DioceseSlug: "galveston_houston", Name: "St. Clare of Assisi Catholic Church",
			Address:     "3131 El Dorado Blvd, Houston, TX 77072",
			BulletinURL: "https://stclarehouston.org/bulletins",
		},
		{
			DioceseSlug: "galveston_houston", Name: "St. Augustine Catholic Church",
			Address:     "4009 Avenue B, Bellaire, TX 77401",
			BulletinURL: "https://staugustinecc.org/bulletins",
		},
		{
			DioceseSlug: "galveston_houston", Name: "Holy Rosary Catholic Church",
			Address:     "3617 Travis St, Houston, TX 77006",
			BulletinURL: "https://holyrosarycatholic.org/bulletins",
		},
		{
			DioceseSlug: "galveston_houston", Name: "St. Francis of Assisi Catholic Church",
			Address:     "8000 Roos Rd, Houston, TX 77036",
			BulletinURL: "https://www.stfrancisofhouston.org/archives/",
		},
	},

	// --- Social (wedding-industry Instagram vendors) ---------------------
	Vendors: []VendorDef{
		// Houston photographers
		{
			Name: "Good Omen Co", OfficialURL: "https://goodomenco.com/",
			Handle: "goodomen.weddings", SourceClass: "engagement_photographer",
			CityID: "city_houston_tx", State: "TX", City: "Houston", Verified: "2026-08-01",
			TikTokHandle: "goodomen.weddings",
		},
		{
			Name: "RKM Photography", OfficialURL: "https://www.rkmphotography.com/",
			Handle: "rkm_photography_", SourceClass: "engagement_photographer",
			CityID: "city_houston_tx", State: "TX", City: "Houston", Verified: "2026-08-01",
			TikTokHandle: "rkm_photography",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-photographers/rkm-photography-9583397",
		},
		// Houston venues
		{
			Name: "The Bell Tower on 34th", OfficialURL: "https://thebelltoweron34th.com/",
			Handle: "belltowerhouston", SourceClass: "wedding_venue",
			CityID: "city_houston_tx", State: "TX", City: "Houston", Verified: "2026-08-01",
			TikTokHandle: "belltowerhouston",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-venues/the-bell-tower-on-34th-3931911",
		},
		{
			Name: "Chateau Nouvelle", OfficialURL: "https://www.chateaunouvelle.com/",
			Handle: "chateaunouvelle", SourceClass: "wedding_venue",
			CityID: "city_houston_tx", State: "TX", City: "Houston", Verified: "2026-08-01",
		},
		{
			Name: "The Springs Wedding & Event Venues", OfficialURL: "https://springsvenue.com/",
			Handle: "springsvenue", SourceClass: "wedding_venue",
			CityID: "city_houston_tx", State: "TX", City: "Houston", Verified: "2026-08-01",
			TikTokHandle: "springsvenue",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-venues/the-springs-wedding-event-venues-7679936",
		},
		// Houston jeweler
		{
			Name: "Reiner's Fine Jewelry", OfficialURL: "https://www.reinersjewelry.com/",
			Handle: "reinersjewelry", SourceClass: "jeweler",
			CityID: "city_houston_tx", State: "TX", City: "Houston", Verified: "2026-08-01",
		},
		// Austin photographers
		{
			Name: "Colette Elyse Photography", OfficialURL: "https://coletteelysephotography.com/",
			Handle: "coletteelysephotography", SourceClass: "engagement_photographer",
			CityID: "city_austin_tx", State: "TX", City: "Austin", Verified: "2026-08-01",
		},
		{
			Name: "Southern Love Creative", OfficialURL: "https://www.southernlovecreative.com/",
			Handle: "southernlovecreative", SourceClass: "engagement_photographer",
			CityID: "city_austin_tx", State: "TX", City: "Austin", Verified: "2026-08-01",
			TikTokHandle: "southernlovecreative",
		},
		{
			Name: "Eryn Chandler Photography", OfficialURL: "https://erynchandler.com/",
			Handle: "erynchandlerphoto", SourceClass: "engagement_photographer",
			CityID: "city_austin_tx", State: "TX", City: "Austin", Verified: "2026-08-01",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-photographers/eryn-chandler-photography-8735930",
		},
		// Austin venues
		{
			Name: "Ma Maison", OfficialURL: "https://themamaison.com/",
			Handle: "themamaison", SourceClass: "wedding_venue",
			CityID: "city_austin_tx", State: "TX", City: "Austin", Verified: "2026-08-01",
			TikTokHandle: "themamaison",
		},
		{
			Name: "Pecan Springs Ranch", OfficialURL: "https://pecanspringsranch.com/",
			Handle: "pecanspringsranch", SourceClass: "wedding_venue",
			CityID: "city_austin_tx", State: "TX", City: "Austin", Verified: "2026-08-01",
		},
		// Houston wedding planners
		{
			Name: "Timeless Rose Events", OfficialURL: "https://timelessroseevents.com/",
			Handle: "timeless_rose_events", SourceClass: "wedding_planner",
			CityID: "city_houston_tx", State: "TX", City: "Houston", Verified: "2026-08-01",
		},
		// Dallas wedding planners
		{
			Name: "Lottie & Co. Events", OfficialURL: "https://lottieandcoevents.com/",
			Handle: "lottieandcoevents", SourceClass: "wedding_planner",
			CityID: "city_dallas_tx", State: "TX", City: "Dallas", Verified: "2026-08-01",
		},
		{
			Name: "Mayfield Events", OfficialURL: "https://mayfieldevents.com/",
			Handle: "mayfield_events", SourceClass: "wedding_planner",
			CityID: "city_dallas_tx", State: "TX", City: "Dallas", Verified: "2026-08-01",
		},
		// Texas florists
		{
			Name: "VIP Floristry", OfficialURL: "https://vipfloristry.com/",
			Handle: "vipfloristry", SourceClass: "florist",
			CityID: "city_houston_tx", State: "TX", City: "Houston", Verified: "2026-08-01",
		},
		{
			Name: "Bring Joy Florals", OfficialURL: "https://bringjoytexas.com/",
			Handle: "bringjoytexas", SourceClass: "florist",
			CityID: "city_houston_tx", State: "TX", City: "Houston", Verified: "2026-08-01",
		},
		{
			Name: "Dr Delphinium", OfficialURL: "https://drdelphinium.com/",
			Handle: "drdelphinium", SourceClass: "florist",
			CityID: "city_dallas_tx", State: "TX", City: "Dallas", Verified: "2026-08-01",
		},
		// Texas videographers
		{
			Name: "The Wedding Filmer", OfficialURL: "https://www.theweddingfilmer.com/",
			Handle: "theweddingfilmer", SourceClass: "videographer",
			CityID: "city_austin_tx", State: "TX", City: "Austin", Verified: "2026-08-01",
		},
		{
			Name: "Dallas Wedding Films", OfficialURL: "https://www.dallasweddingfilms.com/",
			Handle: "dallasweddingfilms", SourceClass: "videographer",
			CityID: "city_dallas_tx", State: "TX", City: "Dallas", Verified: "2026-08-01",
		},
		// Texas wedding cake bakeries
		{
			Name: "Fancy Cakes by Lauren", OfficialURL: "https://fancycakesbylauren.com/",
			Handle: "fancycakesbylauren", SourceClass: "wedding_cake",
			CityID: "city_dallas_tx", State: "TX", City: "Dallas", Verified: "2026-08-01",
		},
		{
			Name: "Wedding Cakes by Tammy Allen", OfficialURL: "https://www.weddingcakesbytammyallen.com/",
			Handle: "weddingcakesbytammyallen", SourceClass: "wedding_cake",
			CityID: "city_houston_tx", State: "TX", City: "Houston", Verified: "2026-08-01",
		},
		// Texas bridal shops
		{
			Name: "Stanley Korshak Bridal", OfficialURL: "https://www.stanleykorshak.com/",
			Handle: "stanleykorshak", SourceClass: "bridal_shop",
			CityID: "city_dallas_tx", State: "TX", City: "Dallas", Verified: "2026-08-01",
		},
		{
			Name: "Posh Bridal Couture", OfficialURL: "https://www.poshbridalcouture.com/",
			Handle: "poshbridalcouture", SourceClass: "bridal_shop",
			CityID: "city_austin_tx", State: "TX", City: "Austin", Verified: "2026-08-01",
		},
		// Texas officiants
		{
			Name: "Austin Wedding Officiants", OfficialURL: "https://www.austinweddingofficiants.com/",
			Handle: "austinweddingofficiants", SourceClass: "officiant",
			CityID: "city_austin_tx", State: "TX", City: "Austin", Verified: "2026-08-01",
		},
		{
			Name: "Houston Wedding Officiant", OfficialURL: "https://www.houstonweddingofficiant.com/",
			Handle: "houstonweddingofficiant", SourceClass: "officiant",
			CityID: "city_houston_tx", State: "TX", City: "Houston", Verified: "2026-08-01",
		},
	},
}
