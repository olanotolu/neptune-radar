package packs

// Maryland source pack — verified 2026-08-01.
//
// Government: Maryland marriage records are held by the Circuit Court Clerk
// in each county (and Baltimore City). Unlike Texas, Maryland counties do not
// host per-county online record-search portals; marriage records are
// request-oriented (certified copies by mail or in person). The statewide
// Maryland Judiciary Case Search (casesearch.mdcourts.gov) includes some
// marriage-license filings but is not a dedicated marriage-record index.
// URLs for the top 7 jurisdictions by population were verified against each
// county's official .gov or mdcourts.gov site.
//
// Church: the Archdiocese of Baltimore and the Diocese of Wilmington were
// verified via USCCB + each diocese's own website. Baltimore-area parishes
// (Archdiocese of Baltimore) were verified against the archdiocese's own
// parish directory (archbalt.org/parishes) + direct bulletin-archive URL
// discovery on each parish's website.
//
// Social: Instagram handles verified from each business's own public website
// social links (or from IG search results where the site is JS-rendered and
// the handle was visible in the search snippet). Verification date recorded
// per vendor.

var mdPack = StatePack{
	State: "MD",

	// --- Cities ----------------------------------------------------------
	Cities: []CityDef{
		{ID: "city_baltimore_md", State: "MD", County: "24510", Name: "Baltimore",
			Lat: 39.2904, Lng: -76.6122, Markets: []string{"baltimore", "bmore", "md"}},
		{ID: "city_bethesda_md", State: "MD", County: "24031", Name: "Bethesda",
			Lat: 38.9847, Lng: -77.0947, Markets: []string{"bethesda", "montgomery", "md"}},
	},

	// --- Government (circuit court clerk marriage-record offices) ---------
	Government: []GovSource{
		{
			// Baltimore City — independent city; Land Records & Licenses
			// Division at the Clarence M. Mitchell, Jr. Courthouse handles
			// marriage licenses (Room 627).
			CountyFIPS: "24510",
			CourtName:  "Baltimore City Circuit Court Clerk",
			CourtURL:   "https://baltimorecitycourt.org/clerks-office/land-records-licenses-division/",
			SearchURL:  "https://baltimorecitycourt.org/clerks-office/land-records-licenses-division/",
			Note:       "Marriage licenses via Land Records & Licenses Division; no online index — request-oriented. Statewide case search at casesearch.mdcourts.gov includes some filings.",
		},
		{
			// Montgomery County — marriage license info + certified copy
			// request form; records from 1993 forward at the court.
			CountyFIPS: "24031",
			CourtName:  "Montgomery County Circuit Court Clerk",
			CourtURL:   "https://www.montgomerycountymd.gov/circuit-court",
			SearchURL:  "https://www.montgomerycountymd.gov/circuit-court/how-do-i/how-do-i-get-marriage-license",
			Note:       "Marriage license page with certified copy request form (PDF); records from 1993 forward; no online index — request-oriented.",
		},
		{
			// Prince George's County — marriage license page with
			// certified copy request form (PDF).
			CountyFIPS: "24033",
			CourtName:  "Prince George's County Circuit Court Clerk",
			CourtURL:   "https://www.princegeorgescourts.org/178/Clerk-of-the-Circuit-Court",
			SearchURL:  "https://www.princegeorgescourts.org/225/Marriage-License",
			Note:       "Marriage license page with certified copy request form (PDF); no online index — request-oriented.",
		},
		{
			// Baltimore County — marriage/divorce records page with
			// copy request info.
			CountyFIPS: "24005",
			CourtName:  "Baltimore County Circuit Court Clerk",
			CourtURL:   "https://www.baltimorecountymd.gov/departments/circuit/clerk/marriage-divorce",
			SearchURL:  "https://www.baltimorecountymd.gov/departments/circuit/clerk/marriage-divorce",
			Note:       "Marriage/divorce records page with copy request info; no online index — request-oriented.",
		},
		{
			// Howard County — marriage license info + certified copy
			// request form (PDF) on mdcourts.gov.
			CountyFIPS: "24027",
			CourtName:  "Howard County Circuit Court Clerk",
			CourtURL:   "https://www.mdcourts.gov/clerks/howard",
			SearchURL:  "https://www.mdcourts.gov/clerks/howard/marriage",
			Note:       "Marriage license info + certified copy request form (PDF); no online index — request-oriented.",
		},
		{
			// Anne Arundel County — marriage licenses page; records from
			// 1905 to present.
			CountyFIPS: "24003",
			CourtName:  "Anne Arundel County Circuit Court Clerk",
			CourtURL:   "https://www.circuitcourt.org/clerk-circuit-court",
			SearchURL:  "https://www.circuitcourt.org/clerk-circuit-court/marriage-licenses",
			Note:       "Marriage licenses page; records from 1905 to present; certified copies by mail or in person; no online index.",
		},
		{
			// Frederick County — marriage license info + copy request
			// form on mdcourts.gov.
			CountyFIPS: "24021",
			CourtName:  "Frederick County Circuit Court Clerk",
			CourtURL:   "https://www.mdcourts.gov/clerks/frederick",
			SearchURL:  "https://www.mdcourts.gov/clerks/frederick/marriage",
			Note:       "Marriage license info + copy request form; no online index — request-oriented.",
		},
	},

	// --- Church (Catholic dioceses + parishes) ---------------------------
	Dioceses: []DioceseDef{
		{Slug: "baltimore", Name: "Archdiocese of Baltimore", Type: "archdiocese",
			Website: "https://www.archbalt.org", Directory: "https://www.archbalt.org/parishes", HubCityID: "city_baltimore_md"},
		{Slug: "wilmington", Name: "Diocese of Wilmington", Type: "diocese",
			Website: "https://www.cdow.org", Directory: "https://www.cdow.org/parishes"},
	},

	// Baltimore-area parishes in the Archdiocese of Baltimore. Names and
	// addresses verified from the archdiocese's own parish directory
	// (archbalt.org/parishes/all-parishes). Bulletin URLs verified by
	// direct discovery on each parish's own website.
	Parishes: []ParishDef{
		{
			DioceseSlug: "baltimore", Name: "Cathedral of Mary Our Queen",
			Address:     "5200 N Charles St, Baltimore, MD 21210",
			BulletinURL: "https://cathedralofmary.org/news",
		},
		{
			DioceseSlug: "baltimore", Name: "Basilica of the Assumption",
			Address:     "409 Cathedral St, Baltimore, MD 21201",
			BulletinURL: "https://americasfirstcathedral.org/parish-newsletters",
		},
		{
			DioceseSlug: "baltimore", Name: "St. Ignatius Catholic Community",
			Address: "740 N Calvert St, Baltimore, MD 21202",
		},
		{
			DioceseSlug: "baltimore", Name: "St. Casimir Church at Canton & Patterson Park",
			Address:     "2736 O'Donnell St, Baltimore, MD 21224",
			BulletinURL: "https://stcasimir.org/bulletin/",
		},
		{
			DioceseSlug: "baltimore", Name: "St. Agnes Catholic Church",
			Address:     "5422 Old Frederick Rd, Baltimore, MD 21229",
			BulletinURL: "https://www.stagnescatholicchurch.org/bulletin",
		},
		{
			DioceseSlug: "baltimore", Name: "St. Mark Parish (Catonsville)",
			Address:     "30 Melvin Ave, Baltimore, MD 21228",
			BulletinURL: "https://www.stmarkchurch-catonsville.org/events/parish-bulletins/",
		},
		{
			DioceseSlug: "baltimore", Name: "St. John the Evangelist (Severna Park)",
			Address:     "689 Ritchie Hwy, Severna Park, MD 21146",
			BulletinURL: "https://stjohnsp.org/bulletin",
		},
	},

	// --- Social (wedding-industry Instagram vendors) ---------------------
	Vendors: []VendorDef{
		// Baltimore photographers
		{
			Name: "Nicole Simensky Photography", OfficialURL: "https://nicolesimenskyphotography.com/",
			Handle: "nicolesimensky.photo", SourceClass: "engagement_photographer",
			CityID: "city_baltimore_md", State: "MD", City: "Baltimore", Verified: "2026-08-01",
			TikTokHandle: "nicolesimensky.photo",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-photographers/nicole-simensky-photography-1735235",
		},
		{
			Name: "Cait Kramer Photography", OfficialURL: "https://www.caitkramer.com/",
			Handle: "caitkramer", SourceClass: "engagement_photographer",
			CityID: "city_baltimore_md", State: "MD", City: "Baltimore", Verified: "2026-08-01",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-photographers/cait-kramer-photography-5847913",
		},
		{
			Name: "Kimberly Dean Photos", OfficialURL: "https://kimberlydeanphotos.com/",
			Handle: "kimberlydeanphotos", SourceClass: "engagement_photographer",
			CityID: "city_baltimore_md", State: "MD", City: "Baltimore", Verified: "2026-08-01",
			TikTokHandle: "kimberlydeanphotos",
		},
		{
			Name: "Jenna Davis Photography", OfficialURL: "https://jennadavisphoto.com/",
			Handle: "jennadavisphotographymd", SourceClass: "engagement_photographer",
			CityID: "city_baltimore_md", State: "MD", City: "Baltimore", Verified: "2026-08-01",
		},
		// Baltimore venues
		{
			Name: "The Loom Baltimore", OfficialURL: "https://www.loombaltimore.com/",
			Handle: "theloombaltimore", SourceClass: "wedding_venue",
			CityID: "city_baltimore_md", State: "MD", City: "Baltimore", Verified: "2026-08-01",
			TikTokHandle: "theloombaltimore",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-venues/the-loom-baltimore-5382252",
		},
		{
			Name: "Gramercy Mansion", OfficialURL: "https://www.gramercymansion.com/",
			Handle: "gramercymansion", SourceClass: "wedding_venue",
			CityID: "city_baltimore_md", State: "MD", City: "Baltimore", Verified: "2026-08-01",
		},
		{
			Name: "Pendry Baltimore", OfficialURL: "https://www.pendry.com/baltimore/",
			Handle: "sagamorependrybaltimore", SourceClass: "wedding_venue",
			CityID: "city_baltimore_md", State: "MD", City: "Baltimore", Verified: "2026-08-01",
			TikTokHandle: "sagamorependrybaltimore",
		},
		{
			Name: "Lord Baltimore Hotel", OfficialURL: "https://www.lordbaltimorehotel.com/",
			Handle: "lordbaltimorehotel", SourceClass: "wedding_venue",
			CityID: "city_baltimore_md", State: "MD", City: "Baltimore", Verified: "2026-08-01",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-venues/lord-baltimore-hotel-6473777",
		},
		// Baltimore jewelers
		{
			Name: "Nelson Coleman Jewelers", OfficialURL: "https://nelsoncoleman.com/",
			Handle: "nelsoncoleman_jewelers", SourceClass: "jeweler",
			CityID: "city_baltimore_md", State: "MD", City: "Baltimore", Verified: "2026-08-01",
		},
		{
			Name: "Gabe & Rubens Jewelers", OfficialURL: "https://gabeandrubensjewelry.com/",
			Handle: "gabeandrubensjewelry", SourceClass: "jeweler",
			CityID: "city_baltimore_md", State: "MD", City: "Baltimore", Verified: "2026-08-01",
		},
		{
			Name: "Arpasi Photography", OfficialURL: "https://arpasiphotography.com/",
			Handle: "arpasiphotography", SourceClass: "engagement_photographer",
			CityID: "city_baltimore_md", State: "MD", City: "Baltimore", Verified: "2026-08-03",
		},
		{
			Name: "Jen Harvey Photography", OfficialURL: "https://jenharveyphotography.com/",
			Handle: "jenharveyphotography", SourceClass: "engagement_photographer",
			CityID: "city_baltimore_md", State: "MD", City: "Baltimore", Verified: "2026-08-03",
		},
		{
			Name: "Jenna Davis Photography", OfficialURL: "https://jennadavisphoto.com/",
			Handle: "jennaphotography_1", SourceClass: "engagement_photographer",
			CityID: "city_baltimore_md", State: "MD", City: "Baltimore", Verified: "2026-08-03",
		},
		{
			Name: "The Grand Baltimore", OfficialURL: "https://thegrandbaltimore.com/",
			Handle: "thegrandbaltimore", SourceClass: "wedding_venue",
			CityID: "city_baltimore_md", State: "MD", City: "Baltimore", Verified: "2026-08-03",
		},
		{
			Name: "Kimpton Hotel Monaco Baltimore", OfficialURL: "https://monaco-baltimore.com/",
			Handle: "KimptonMonacoBaltimore", SourceClass: "wedding_venue",
			CityID: "city_baltimore_md", State: "MD", City: "Baltimore", Verified: "2026-08-03",
		},
		{
			Name: "Blossoms of Bethesda", OfficialURL: "https://bethesdablossom.com/",
			Handle: "blossomsofbethesda", SourceClass: "florist",
			CityID: "city_bethesda_md", State: "MD", City: "Bethesda", Verified: "2026-08-03",
		},
		{
			Name: "Nelson Coleman Jewelers", OfficialURL: "https://nelsoncoleman.com/",
			Handle: "nelsoncolemantowson", SourceClass: "jeweler",
			CityID: "city_baltimore_md", State: "MD", City: "Baltimore", Verified: "2026-08-03",
		},
		{
			Name: "Smyth Jewelers", OfficialURL: "https://www.smythjewelers.com/",
			Handle: "smythjewelers", SourceClass: "jeweler",
			CityID: "city_baltimore_md", State: "MD", City: "Baltimore", Verified: "2026-08-03",
		},
	},
}
