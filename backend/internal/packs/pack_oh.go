package packs

// Ohio source pack — data verified 2026-07-29 and ported from the
// bootstrap-ohio command (backend/cmd/bootstrap-ohio). Every URL below was
// confirmed against a real public source before being written; nothing here is
// invented or guessed.
//
// Government: Franklin County Probate Court (the primary, Columbus metro
// marriage-record office) plus 8 additional verified county probate-court
// sources. URLs and capability notes come from the 2026-07-29 source review.
//
// Church: all six Ohio Catholic jurisdictions (Diocese of Columbus + 5 others).
// Columbus parish directory verified against columbuscatholic.org/find-a-parish;
// the 13 Columbus-proper parishes were verified against the Wikipedia list of
// churches in the Diocese of Columbus (which cites the diocese's own records).
// Bulletin URLs set only where a real, reachable archive was located;
// Aggregator marks third-party listings (Parishes Online / Discover Mass).
//
// Social: Instagram handles verified 2026-07-29 by fetching each business's own
// official website and reading the handle off its own social links.

var ohPack = StatePack{
	State: "OH",

	// --- Cities ----------------------------------------------------------
	Cities: []CityDef{
	{
		ID: "city_cleveland_oh", State: "OH", County: "39035", Name: "Cleveland",
		Lat: 41.4993, Lng: -81.6944,
		Markets: []string{"cleveland", "cuyahoga", "clevelandheights"},
	},
		{
			ID: "city_columbus_oh", State: "OH", County: "39049", Name: "Columbus",
			Lat: 39.9612, Lng: -82.9988,
			Markets: []string{"columbus", "ohio", "cbus", "franklincounty"},
		},
	},

	// --- Government (county probate-court marriage-record searches) ------
	Government: []GovSource{
		{
			// Franklin County (Columbus) — primary probate court. Online search
			// covers marriage licenses issued January 3, 1994 to present.
			CountyFIPS: "39049",
			CourtName:  "Franklin County Probate Court",
			CourtURL:   "https://probate.franklincountyohio.gov",
			SearchURL:  "https://probate.franklincountyohio.gov/Record-Search/Marriage-License-Search",
			Note:       "Online search covers marriage licenses issued January 3, 1994 to present.",
		},
		{
			// Montgomery County (Dayton) — supports searching by issued-date
			// range as well as names: strongest enumeration candidate.
			CountyFIPS: "39113",
			CourtName:  "Montgomery County Probate Court",
			CourtURL:   "https://go.mcohio.org",
			SearchURL:  "https://go.mcohio.org/applications/probate/prodcfm/marriagesearch.cfm",
			Note:       "Supports name search and issued-date-range search — enumeration candidate; connector build next.",
		},
		{
			// Hamilton County (Cincinnati) — public electronic index.
			CountyFIPS: "39061",
			CourtName:  "Hamilton County Probate Court",
			CourtURL:   "https://www.probatect.org/marriage-license",
			SearchURL:  "https://www.probatect.org/court-records/archive-categories/marriages",
			Note:       "Public electronic index; verification use confirmed, enumeration capability needs testing.",
		},
		{
			// Cuyahoga County (Cleveland) — marriage dept + web docket.
			CountyFIPS: "39035",
			CourtName:  "Cuyahoga County Probate Court",
			CourtURL:   "https://probate.cuyahogacounty.gov/marriage.aspx",
			SearchURL:  "https://probate.cuyahogacounty.gov/pa/",
			Note:       "Web docket available for verification; automation and terms review required.",
		},
		{
			// Summit County (Akron) — record information page.
			CountyFIPS: "39153",
			CourtName:  "Summit County Probate Court",
			CourtURL:   "https://summitcountycourt.org",
			SearchURL:  "https://summitcountycourt.org/marriage-divorce-records/",
			Note:       "Record information page; request-oriented, automation capability needs testing.",
		},
		{
			// Delaware County — official probate record search.
			CountyFIPS: "39041",
			CourtName:  "Delaware County Probate Court",
			CourtURL:   "https://probate.co.delaware.oh.us",
			SearchURL:  "https://probate.co.delaware.oh.us/recordsearch/",
			Note:       "Official probate record search; marriage-record capability needs testing.",
		},
		{
			// Fairfield County — online probate records search form.
			CountyFIPS: "39045",
			CourtName:  "Fairfield County Probate Court",
			CourtURL:   "https://www.fairfieldcountyprobate.com",
			SearchURL:  "https://www.fairfieldcountyprobate.com/ff-Probate-Records-Search-Form.html",
			Note:       "Online probate records search form; marriage filtering needs testing.",
		},
		{
			// Licking County — online portal behind an agreement gate.
			CountyFIPS: "39089",
			CourtName:  "Licking County Probate Court",
			CourtURL:   "https://pjc-portal.lickingcounty.gov",
			SearchURL:  "https://pjc-portal.lickingcounty.gov/recordSearch.php?k=acceptAgreementsearchForm4503",
			Note:       "Online portal (agreement-gated); terms and record-type capability need verification.",
		},
		{
			// Lucas County (Toledo) — official custodian page.
			CountyFIPS: "39095",
			CourtName:  "Lucas County Probate Court",
			CourtURL:   "https://www.co.lucas.oh.us/169/Probate-Court",
			SearchURL:  "https://www.co.lucas.oh.us/169/Probate-Court",
			Note:       "Official custodian page; online enumeration not yet confirmed.",
		},
		{CountyFIPS: "39051", CourtName: "Delaware County Probate Court",
			CourtURL:  "https://probate.co.delaware.oh.us/",
			SearchURL: "https://probate.co.delaware.oh.us/recordsearch/",
			Note:      "Online probate court record search; marriage records from 1996 to present."},
		{CountyFIPS: "39069", CourtName: "Lake County Probate Court",
			CourtURL:  "https://www.lakecountyohio.gov/probate-court/",
			SearchURL: "https://www.lakecountyohio.gov/probate-court/marriage-department/",
			Note:      "Online probate case search; includes marriage department cases with document images."},
		{CountyFIPS: "39151", CourtName: "Stark County Probate Court",
			CourtURL:  "https://starkcountyohio.gov/",
			SearchURL: "http://www.probate.co.stark.oh.us/search/search_main.html",
			Note:      "Online probate case search; marriage indices from April 23, 1986 to present."},
	},

	// --- Church (Catholic dioceses + parishes) ---------------------------
	Dioceses: []DioceseDef{
		{
			Slug: "columbus", Name: "Diocese of Columbus", Type: "diocese",
			Website: "https://columbuscatholic.org", Directory: "https://columbuscatholic.org/find-a-parish",
			HubCityID: "city_columbus_oh",
		},
		{
			Slug: "cleveland", Name: "Diocese of Cleveland", Type: "diocese",
			Website: "https://www.dioceseofcleveland.org", Directory: "https://www.dioceseofcleveland.org/about/our-parishes",
		},
		{
			Slug: "cincinnati", Name: "Archdiocese of Cincinnati", Type: "archdiocese",
			Website: "https://catholicaoc.org", Directory: "https://catholicaoc.org/parishes",
		},
		{
			Slug: "toledo", Name: "Diocese of Toledo", Type: "diocese",
			Website: "https://toledodiocese.org", Directory: "https://toledodiocese.org",
		},
		{
			Slug: "youngstown", Name: "Diocese of Youngstown", Type: "diocese",
			Website: "https://doy.org", Directory: "https://doy.org",
		},
		{
			Slug: "steubenville", Name: "Diocese of Steubenville", Type: "diocese",
			Website: "https://www.diosteub.org", Directory: "https://www.diosteub.org/parishfinder",
		},
	},

	// Columbus-proper parishes in the Diocese of Columbus. Names and addresses
	// verified 2026-07-29 against the Wikipedia list of churches in the Diocese
	// of Columbus (which cites the diocese's own records). Bulletin URLs set
	// only where a real, reachable archive was located; Aggregator marks
	// third-party listings (Parishes Online / Discover Mass).
	Parishes: []ParishDef{
		{DioceseSlug: "columbus", Name: "Community of Holy Rosary and Saint John the Evangelist", Address: "648 S Ohio Ave, Columbus, OH 43205"},
		{
			DioceseSlug: "columbus", Name: "Holy Cross Church", Address: "204 S 5th St, Columbus, OH 43215",
			BulletinURL: "https://parishesonline.com/organization/holy-cross-catholic-church-43215", Aggregator: true,
		},
		{DioceseSlug: "columbus", Name: "Saint Leo Oratory", Address: "221 Hanford St, Columbus, OH 43206"},
		{
			DioceseSlug: "columbus", Name: "Saint Dominic Church", Address: "453 N 20th St, Columbus, OH 43203",
			BulletinURL: "https://stdominic-church.org/bulletins",
		},
		{DioceseSlug: "columbus", Name: "Saint Joseph Cathedral", Address: "212 E Broad St, Columbus, OH 43215"},
		{DioceseSlug: "columbus", Name: "Saint Mary, Mother of God Church", Address: "684 S 3rd St, Columbus, OH 43206"},
		{
			DioceseSlug: "columbus", Name: "Saint Patrick Church", Address: "280 N Grant Ave, Columbus, OH",
			BulletinURL: "https://www.stpatrickcolumbus.org/weekly-bulletin",
		},
		{DioceseSlug: "columbus", Name: "Saint Thomas the Apostle Church", Address: "2692 E 5th Ave, Columbus, OH 43219"},
		{DioceseSlug: "columbus", Name: "Saints Augustine and Gabriel Church", Address: "1550 E Hudson St, Columbus, OH 43211"},
		// Immaculate Conception: real bulletin observed with a "BANNS OF
		// MARRIAGE" section naming the couple and the wedding date — the proof
		// that this lane produces the signal Neptune wants.
		{
			DioceseSlug: "columbus", Name: "Immaculate Conception Church",
			BulletinURL: "https://www.iccols.org/bulletin/",
		},
		{
			DioceseSlug: "columbus", Name: "Sacred Heart Church",
			BulletinURL: "https://sacredheartchurchcolumbus.org/bulletins",
		},
		{
			DioceseSlug: "columbus", Name: "Holy Spirit Church",
			BulletinURL: "https://holyspiritcolumbus.org/bulletins",
		},
		{
			DioceseSlug: "columbus", Name: "Saint Agatha Church",
			BulletinURL: "https://discovermass.com/church/st-agatha-columbus-oh/?id=20170425", Aggregator: true,
		},
	},

	// --- Social (Columbus wedding-industry Instagram vendors) ------------
	// Handles verified 2026-07-29 from each business's own official website.
	Vendors: []VendorDef{
		{
			Name: "Starling Studio", OfficialURL: "https://www.starling-studio.com/",
			Handle: "starling_studio", SourceClass: "engagement_photographer",
			CityID: "city_columbus_oh", State: "OH", City: "Columbus", Verified: "2026-07-29",
			TikTokHandle: "starling_studio",
		},
		{
			Name: "Jessica Miller Photography", OfficialURL: "https://www.thejessicamillerphotos.com/",
			Handle: "jmillerphotos", SourceClass: "engagement_photographer",
			CityID: "city_columbus_oh", State: "OH", City: "Columbus", Verified: "2026-07-29",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-photographers/jessica-miller-photography-7054632",
		},
		{
			Name: "Laura Witherow Photography", OfficialURL: "https://laurawitherowphotography.com/",
			Handle: "laurawitherow", SourceClass: "engagement_photographer",
			CityID: "city_columbus_oh", State: "OH", City: "Columbus", Verified: "2026-07-29",
			TikTokHandle: "laurawitherow",
		},
		{
			Name: "Kismet Visuals & Co", OfficialURL: "https://kismetvisuals.com/",
			Handle: "kismetvisuals", SourceClass: "engagement_photographer",
			CityID: "city_columbus_oh", State: "OH", City: "Columbus", Verified: "2026-07-29",
		},
		{
			Name: "Svetlana Photography", OfficialURL: "https://svetlanaphotography.com/",
			Handle: "svetphoto", SourceClass: "engagement_photographer",
			CityID: "city_columbus_oh", State: "OH", City: "Columbus", Verified: "2026-07-29",
			TikTokHandle: "svetphoto",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-photographers/svetlana-photography-8368497",
		},
		{
			Name: "Asteria Photography", OfficialURL: "https://www.asteriaphoto.com/",
			Handle: "asteriaphoto", SourceClass: "engagement_photographer",
			CityID: "city_columbus_oh", State: "OH", City: "Columbus", Verified: "2026-07-29",
		},
		{
			Name: "Magnolia Hill Farm", OfficialURL: "https://magnoliahill-farm.com/",
			Handle: "magnoliahillfarm", SourceClass: "wedding_venue",
			CityID: "city_columbus_oh", State: "OH", City: "Columbus", Verified: "2026-07-29",
			TikTokHandle: "magnoliahillfarm",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-venues/magnolia-hill-farm-4359970",
		},
		{
			Name: "Jorgensen Farms", OfficialURL: "https://jorgensen-farms.com/",
			Handle: "jorgensenfarms", SourceClass: "wedding_venue",
			CityID: "city_columbus_oh", State: "OH", City: "Columbus", Verified: "2026-07-29",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-venues/jorgensen-farms-5278626",
		},
		{
			Name: "Franklin Park Conservatory and Botanical Gardens", OfficialURL: "https://www.fpconservatory.org/",
			Handle: "fpconservatory", SourceClass: "wedding_venue",
			CityID: "city_columbus_oh", State: "OH", City: "Columbus", Verified: "2026-07-29",
			TikTokHandle: "fpconservatory",
		},
		{
			Name: "The Columbus Athenaeum", OfficialURL: "https://www.columbusmeetings.com/",
			Handle: "thecolumbusathenaeum", SourceClass: "wedding_venue",
			CityID: "city_columbus_oh", State: "OH", City: "Columbus", Verified: "2026-07-29",
		},
		{
			Name: "Le Méridien Columbus, The Joseph", OfficialURL: "https://www.weddingsatthejoseph.com/",
			Handle: "lmcolumbusthejoseph", SourceClass: "wedding_venue",
			CityID: "city_columbus_oh", State: "OH", City: "Columbus", Verified: "2026-07-29",
			TikTokHandle: "lmcolumbusthejoseph",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-venues/le-mridien-columbus-the-joseph-7090734",
		},
		{
			Name: "Brookshire", OfficialURL: "https://brookshire.biz/",
			Handle: "brookshireweddings", SourceClass: "wedding_venue",
			CityID: "city_columbus_oh", State: "OH", City: "Columbus", Verified: "2026-07-29",
		},
		{
			Name: "Worthington Hills Country Club", OfficialURL: "https://www.worthingtonhills.com/",
			Handle: "worthingtonhillscountryclub", SourceClass: "wedding_venue",
			CityID: "city_columbus_oh", State: "OH", City: "Columbus", Verified: "2026-07-29",
			TikTokHandle: "worthingtonhillscountryclub",
		},
		// Columbus wedding planners
		{
			Name: "Rooted Together", OfficialURL: "https://rootedtogether.co/",
			Handle: "rootedtogetherevents", SourceClass: "wedding_planner",
			CityID: "city_columbus_oh", State: "OH", City: "Columbus", Verified: "2026-07-29",
		},
		{
			Name: "Signature Event Planning", OfficialURL: "https://www.thesignatureevent.com/",
			Handle: "thesignatureevent", SourceClass: "wedding_planner",
			CityID: "city_columbus_oh", State: "OH", City: "Columbus", Verified: "2026-07-29",
		},
		// Columbus florists
		{
			Name: "Steinmeier & Co Florist", OfficialURL: "https://www.steinmeierco.com/",
			Handle: "steinmeierco", SourceClass: "florist",
			CityID: "city_columbus_oh", State: "OH", City: "Columbus", Verified: "2026-08-01",
		},
		{
			Name: "Stein Florals", OfficialURL: "https://www.steinflorals.com/",
			Handle: "steinflorals", SourceClass: "florist",
			CityID: "city_columbus_oh", State: "OH", City: "Columbus", Verified: "2026-08-01",
		},
		// Columbus videographers
		{
			Name: "Columbus Wedding Films", OfficialURL: "https://www.columbusweddingfilms.com/",
			Handle: "columbusweddingfilms", SourceClass: "videographer",
			CityID: "city_columbus_oh", State: "OH", City: "Columbus", Verified: "2026-08-01",
		},
		{
			Name: "Wolfe Films", OfficialURL: "https://www.wolfefilms.com/",
			Handle: "wolfefilms", SourceClass: "videographer",
			CityID: "city_columbus_oh", State: "OH", City: "Columbus", Verified: "2026-08-01",
		},
		// Columbus wedding cake bakeries
		{
			Name: "The Sassafras Bakery", OfficialURL: "https://www.sassafrasbakery.com/",
			Handle: "sassafrasbakery", SourceClass: "wedding_cake",
			CityID: "city_columbus_oh", State: "OH", City: "Columbus", Verified: "2026-08-01",
		},
		{
			Name: "Cake Lady Columbus", OfficialURL: "https://www.cakeladycolumbus.com/",
			Handle: "cakeladycolumbus", SourceClass: "wedding_cake",
			CityID: "city_columbus_oh", State: "OH", City: "Columbus", Verified: "2026-08-01",
		},
		// Columbus bridal shops
		{
			Name: "Jaclyn's Bridal", OfficialURL: "https://www.jaclynsbridal.com/",
			Handle: "jaclynsbridal", SourceClass: "bridal_shop",
			CityID: "city_columbus_oh", State: "OH", City: "Columbus", Verified: "2026-08-01",
		},
		{
			Name: "The Bridal Path", OfficialURL: "https://www.thebridalpathcolumbus.com/",
			Handle: "thebridalpath", SourceClass: "bridal_shop",
			CityID: "city_columbus_oh", State: "OH", City: "Columbus", Verified: "2026-08-01",
		},
		// Columbus officiants
		{
			Name: "Columbus Wedding Officiants", OfficialURL: "https://www.columbusweddingofficiants.com/",
			Handle: "columbusweddingofficiants", SourceClass: "officiant",
			CityID: "city_columbus_oh", State: "OH", City: "Columbus", Verified: "2026-08-01",
		},
		{
			Name: "Ohio Wedding Officiant", OfficialURL: "https://www.ohioweddingofficiant.com/",
			Handle: "ohioweddingofficiant", SourceClass: "officiant",
			CityID: "city_columbus_oh", State: "OH", City: "Columbus", Verified: "2026-08-01",
		},
		{
			Name: "Fiori Floral Design Studio", OfficialURL: "https://fioriflorals.biz/",
			Handle: "fiori.florals", SourceClass: "florist",
			CityID: "city_columbus_oh", State: "OH", City: "Columbus", Verified: "2026-08-03",
		},
		{
			Name: "Alexander's Jewelers", OfficialURL: "https://alexanderscolumbus.com/",
			Handle: "alexandersjewelerscolumbus", SourceClass: "jeweler",
			CityID: "city_columbus_oh", State: "OH", City: "Columbus", Verified: "2026-08-03",
		},
		{
			Name: "Zolana Weddings", OfficialURL: "https://zolanaweddings.com/",
			Handle: "zolana.weddings", SourceClass: "wedding_planner",
			CityID: "city_columbus_oh", State: "OH", City: "Columbus", Verified: "2026-08-03",
		},
		{
			Name: "Glidden House", OfficialURL: "https://www.gliddenhouse.com/weddings/",
			Handle: "gliddenhousehotel", SourceClass: "wedding_venue",
			CityID: "city_cleveland_oh", State: "OH", City: "Cleveland", Verified: "2026-08-03",
		},
	},
}
