package packs

// Tennessee source pack — verified 2026-08-01.
//
// Government: Tennessee marriage records are held by the county clerk. Most
// counties use the statewide TNCountyClerk.com marriage-lookup portal
// (secure.tncountyclerk.com/marriagelookup) with a per-county numeric id.
// Hamilton County runs its own dedicated marriage-license search portal.
// Countylist ids were verified by fetching each county's marriage-form page
// and confirming the county name in the response.
//
// Church: all 3 Tennessee Catholic dioceses verified via USCCB + each
// diocese's own website. Nashville-area parishes (Diocese of Nashville) were
// verified against the diocese's own mass-schedule parish list + direct
// bulletin-archive URL discovery on each parish's website.
//
// Social: Instagram handles verified from each business's own public website
// social links (or from IG search results where the site is JS-rendered and
// the handle was visible in the search snippet). Verification date recorded
// per vendor.

var tnPack = StatePack{
	State: "TN",

	// --- Cities ----------------------------------------------------------
	Cities: []CityDef{
		{ID: "city_nashville_tn", State: "TN", County: "47037", Name: "Nashville",
			Lat: 36.1627, Lng: -86.7816, Markets: []string{"nashville", "nash", "davidson", "musiccity"}},
		{ID: "city_memphis_tn", State: "TN", County: "47157", Name: "Memphis",
			Lat: 35.1495, Lng: -90.0490, Markets: []string{"memphis", "shelby", "midsouth"}},
	},

	// --- Government (county clerk marriage-record searches) --------------
	Government: []GovSource{
		{
			// Davidson County (Nashville) — statewide TNCountyClerk marriage
			// lookup portal, countylist=19 (verified via marriage-form page).
			CountyFIPS: "47037",
			CourtName:  "Davidson County Clerk",
			CourtURL:   "https://www.nashville.gov/departments/county-clerk",
			SearchURL:  "https://secure.tncountyclerk.com/marriagelookup/index.php?countylist=19",
			Note:       "Statewide marriage-lookup portal with Davidson County pre-selected; name-based search; enumeration candidate.",
		},
		{
			// Shelby County (Memphis) — statewide TNCountyClerk marriage
			// lookup portal, countylist=79 (verified via marriage-form page).
			CountyFIPS: "47157",
			CourtName:  "Shelby County Clerk",
			CourtURL:   "https://www.shelbycountytn.gov/74/County-Clerk/",
			SearchURL:  "https://secure.tncountyclerk.com/marriagelookup/index.php?countylist=79",
			Note:       "Statewide marriage-lookup portal with Shelby County pre-selected; name-based search; enumeration candidate.",
		},
		{
			// Knox County (Knoxville) — statewide TNCountyClerk marriage
			// lookup portal, countylist=47 (verified via marriage-form page).
			CountyFIPS: "47093",
			CourtName:  "Knox County Clerk",
			CourtURL:   "https://www.knoxcounty.org/clerk/marriagelicense.php",
			SearchURL:  "https://secure.tncountyclerk.com/marriagelookup/index.php?countylist=47",
			Note:       "Statewide marriage-lookup portal with Knox County pre-selected; name-based search; enumeration candidate.",
		},
		{
			// Hamilton County (Chattanooga) — dedicated marriage-license
			// search portal hosted by the county; searchable by name, date,
			// or officiant. Records 1857–8/3/2021 on the site.
			CountyFIPS: "47065",
			CourtName:  "Hamilton County Clerk",
			CourtURL:   "https://www.countyclerkanytime.com/",
			SearchURL:  "https://marriagelicense.hamiltontn.gov/PublicScreens/LicenseSearch.aspx",
			Note:       "Dedicated marriage-license search by name/date/officiant; records through 8/3/2021; enumeration candidate.",
		},
		{
			// Rutherford County (Murfreesboro) — statewide TNCountyClerk
			// marriage lookup portal, countylist=75 (verified via marriage-
			// form page).
			CountyFIPS: "47149",
			CourtName:  "Rutherford County Clerk",
			CourtURL:   "https://rutherfordcountytn.gov/county-clerk/marriage-license",
			SearchURL:  "https://secure.tncountyclerk.com/marriagelookup/index.php?countylist=75",
			Note:       "Statewide marriage-lookup portal with Rutherford County pre-selected; name-based search; enumeration candidate.",
		},
		{
			// Williamson County (Franklin) — statewide TNCountyClerk
			// marriage lookup portal, countylist=94 (verified via marriage-
			// form page).
			CountyFIPS: "47187",
			CourtName:  "Williamson County Clerk",
			CourtURL:   "https://williamsoncounty-tn.gov/161/County-Clerk",
			SearchURL:  "https://secure.tncountyclerk.com/marriagelookup/index.php?countylist=94",
			Note:       "Statewide marriage-lookup portal with Williamson County pre-selected; name-based search; enumeration candidate.",
		},
		{
			// Sumner County (Gallatin) — statewide TNCountyClerk marriage
			// lookup portal, countylist=83 (verified via marriage-form page).
			CountyFIPS: "47165",
			CourtName:  "Sumner County Clerk",
			CourtURL:   "https://sumnertags.com/",
			SearchURL:  "https://secure.tncountyclerk.com/marriagelookup/index.php?countylist=83",
			Note:       "Statewide marriage-lookup portal with Sumner County pre-selected; name-based search; enumeration candidate.",
		},
	},

	// --- Church (Catholic dioceses + parishes) ---------------------------
	Dioceses: []DioceseDef{
		{Slug: "nashville", Name: "Diocese of Nashville", Type: "diocese",
			Website: "https://dioceseofnashville.com", Directory: "https://dioceseofnashville.com/parishes", HubCityID: "city_nashville_tn"},
		{Slug: "knoxville", Name: "Diocese of Knoxville", Type: "diocese",
			Website: "https://www.dioknox.org", Directory: "https://www.dioknox.org/parishes"},
		{Slug: "memphis", Name: "Diocese of Memphis", Type: "diocese",
			Website: "https://www.cdom.org", Directory: "https://www.cdom.org/parishes", HubCityID: "city_memphis_tn"},
	},

	// Nashville-area parishes in the Diocese of Nashville. Names and
	// addresses verified from the diocese's own mass-schedule parish list
	// + each parish's own website. Bulletin URLs verified by direct
	// discovery on each parish's site; Christ the King uses the Parishes
	// Online aggregator (marked Aggregator: true).
	Parishes: []ParishDef{
		{
			DioceseSlug: "nashville", Name: "Cathedral of the Incarnation",
			Address:     "2015 West End Ave, Nashville, TN 37203",
			BulletinURL: "https://www.cathedralnashville.org/weekly-bulletin",
		},
		{
			DioceseSlug: "nashville", Name: "Christ the King Catholic Church",
			Address:     "3001 Belmont Blvd, Nashville, TN 37212",
			BulletinURL: "https://ctk.org/bulletin",
			Aggregator:  true, // redirects to parishesonline.com
		},
		{
			DioceseSlug: "nashville", Name: "Saint Henry Catholic Church",
			Address: "6401 Harding Pike, Nashville, TN 37205",
		},
		{
			DioceseSlug: "nashville", Name: "St. Ann Catholic Church",
			Address:     "5101 Charlotte Ave, Nashville, TN 37209",
			BulletinURL: "https://saintannparish.com/bulletins",
		},
		{
			DioceseSlug: "nashville", Name: "St. Edward Church",
			Address: "188 Thompson Ln, Nashville, TN 37211",
		},
		{
			DioceseSlug: "nashville", Name: "St. Pius X Catholic Church",
			Address: "2800 Tucker Rd, Nashville, TN 37218",
		},
		{
			DioceseSlug: "nashville", Name: "St. Vincent de Paul Catholic Church",
			Address: "1700 Heiman St, Nashville, TN 37208",
		},
		{
			DioceseSlug: "nashville", Name: "Holy Family Catholic Church",
			Address: "9100 Crockett Rd, Brentwood, TN 37027",
		},
	},

	// --- Social (wedding-industry Instagram vendors) ---------------------
	Vendors: []VendorDef{
		// Nashville photographers
		{
			Name: "Haley Maria Photography", OfficialURL: "https://haleymaria.com/",
			Handle: "haleymariaphotography", SourceClass: "engagement_photographer",
			CityID: "city_nashville_tn", State: "TN", City: "Nashville", Verified: "2026-08-01",
			TikTokHandle: "haleymariaphotography",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-photographers/haley-maria-photography-1479847",
		},
		{
			Name: "Kelsey Alex Photography", OfficialURL: "https://kelseyalex.com/",
			Handle: "kelseyalexphoto", SourceClass: "engagement_photographer",
			CityID: "city_nashville_tn", State: "TN", City: "Nashville", Verified: "2026-08-01",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-photographers/kelsey-alex-photography-3764041",
		},
		{
			Name: "Nicole Lenia Kiser", OfficialURL: "https://nicoleleniakiser.com/",
			Handle: "nicolelkiser", SourceClass: "engagement_photographer",
			CityID: "city_nashville_tn", State: "TN", City: "Nashville", Verified: "2026-08-01",
			TikTokHandle: "nicolelkiser",
		},
		{
			Name: "Kiley B Photos", OfficialURL: "https://kileybphotos.com/",
			Handle: "kileyb.photos", SourceClass: "engagement_photographer",
			CityID: "city_nashville_tn", State: "TN", City: "Nashville", Verified: "2026-08-01",
		},
		// Nashville venues
		{
			Name: "The White Dove Barn", OfficialURL: "https://thewhitedovebarn.com/",
			Handle: "thewhitedovebarn", SourceClass: "wedding_venue",
			CityID: "city_nashville_tn", State: "TN", City: "Nashville", Verified: "2026-08-01",
			TikTokHandle: "thewhitedovebarn",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-venues/the-white-dove-barn-6171823",
		},
		{
			Name: "Mint Springs Farm", OfficialURL: "https://mintspringsfarmtn.com/",
			Handle: "mintspringsfarm", SourceClass: "wedding_venue",
			CityID: "city_nashville_tn", State: "TN", City: "Nashville", Verified: "2026-08-01",
		},
		{
			Name: "The Estate at Cherokee Dock", OfficialURL: "https://cherokeedock.com/",
			Handle: "cherokee.dock", SourceClass: "wedding_venue",
			CityID: "city_nashville_tn", State: "TN", City: "Nashville", Verified: "2026-08-01",
			TikTokHandle: "cherokee.dock",
		},
		{
			Name: "Firefly Lane Weddings", OfficialURL: "https://fireflylaneevents.com/",
			Handle: "fireflylaneweddings", SourceClass: "wedding_venue",
			CityID: "city_nashville_tn", State: "TN", City: "Nashville", Verified: "2026-08-01",
			// ponytail: KnotURL placeholder — verify on theknot.com before production use
			KnotURL: "https://www.theknot.com/marketplace/wedding-venues/firefly-lane-weddings-7086751",
		},
		// Nashville jewelers
		{
			Name: "King Jewelers", OfficialURL: "https://kingjewelers.com/",
			Handle: "kingjewelers", SourceClass: "jeweler",
			CityID: "city_nashville_tn", State: "TN", City: "Nashville", Verified: "2026-08-01",
		},
		{
			Name: "Genesis Diamonds", OfficialURL: "https://genesisdiamonds.com/",
			Handle: "genesisdiamonds", SourceClass: "jeweler",
			CityID: "city_nashville_tn", State: "TN", City: "Nashville", Verified: "2026-08-01",
		},
		{
			Name: "Stokes Del Rio Photography", OfficialURL: "https://stokesdelrio.com/",
			Handle: "stokesdelrio", SourceClass: "engagement_photographer",
			CityID: "city_nashville_tn", State: "TN", City: "Nashville", Verified: "2026-08-03",
		},
		{
			Name: "Brandon Allan Photography", OfficialURL: "https://brandonallanphotography.com/",
			Handle: "brandonallanphotography", SourceClass: "engagement_photographer",
			CityID: "city_nashville_tn", State: "TN", City: "Nashville", Verified: "2026-08-03",
		},
		{
			Name: "Harp and Olive Photography", OfficialURL: "https://harpandolive.com/",
			Handle: "harpandolive", SourceClass: "engagement_photographer",
			CityID: "city_nashville_tn", State: "TN", City: "Nashville", Verified: "2026-08-03",
		},
		{
			Name: "Cheekwood Estate & Gardens", OfficialURL: "https://cheekwood.org/weddings-facility-rentals/weddings/",
			Handle: "cheekwood", SourceClass: "wedding_venue",
			CityID: "city_nashville_tn", State: "TN", City: "Nashville", Verified: "2026-08-03",
		},
		{
			Name: "The Parthenon Nashville", OfficialURL: "https://www.nashvilleparthenon.com/nashville-wedding-venue",
			Handle: "nashvilleparthenon", SourceClass: "wedding_venue",
			CityID: "city_nashville_tn", State: "TN", City: "Nashville", Verified: "2026-08-03",
		},
		{
			Name: "Rose Hill Flowers", OfficialURL: "https://www.rosehillflowers.com/",
			Handle: "rosehillflowersnashville", SourceClass: "florist",
			CityID: "city_nashville_tn", State: "TN", City: "Nashville", Verified: "2026-08-03",
		},
		{
			Name: "Frances and Jane Floral Design", OfficialURL: "https://francesandjane.com/",
			Handle: "frances_and_jane", SourceClass: "florist",
			CityID: "city_nashville_tn", State: "TN", City: "Nashville", Verified: "2026-08-03",
		},
		{
			Name: "Genesis Diamonds", OfficialURL: "https://genesisdiamonds.com/pages/nashville",
			Handle: "genesisdiamonds", SourceClass: "jeweler",
			CityID: "city_nashville_tn", State: "TN", City: "Nashville", Verified: "2026-08-03",
		},
		{
			Name: "King Jewelers", OfficialURL: "https://kingjewelers.com/",
			Handle: "kingjewelers", SourceClass: "jeweler",
			CityID: "city_nashville_tn", State: "TN", City: "Nashville", Verified: "2026-08-03",
		},
		{
			Name: "Peerless Films", OfficialURL: "https://peerlessfilms.com/",
			Handle: "peerlessfilms", SourceClass: "videographer",
			CityID: "city_nashville_tn", State: "TN", City: "Nashville", Verified: "2026-08-03",
		},
		{
			Name: "With This Ring Wedding Films", OfficialURL: "https://withthisringvideo.com/",
			Handle: "wtrfilms", SourceClass: "videographer",
			CityID: "city_nashville_tn", State: "TN", City: "Nashville", Verified: "2026-08-03",
		},
		{
			Name: "Baked in Nashville", OfficialURL: "http://www.bakedinnashville.com/",
			Handle: "bakedinnashville", SourceClass: "wedding_cake",
			CityID: "city_nashville_tn", State: "TN", City: "Nashville", Verified: "2026-08-03",
		},
		{
			Name: "Ivey Cake", OfficialURL: "https://www.iveycake.com/",
			Handle: "iveycakestore", SourceClass: "wedding_cake",
			CityID: "city_nashville_tn", State: "TN", City: "Nashville", Verified: "2026-08-03",
		},
		{
			Name: "Elizabeth Hoard Photography", OfficialURL: "https://www.elizabethhoardphotography.com/",
			Handle: "elizabethhoardphotography", SourceClass: "engagement_photographer",
			CityID: "city_memphis_tn", State: "TN", City: "Memphis", Verified: "2026-08-03",
		},
		{
			Name: "Kevin Barre Photography", OfficialURL: "https://www.kevinbarrephoto.com/",
			Handle: "kevinbarre", SourceClass: "engagement_photographer",
			CityID: "city_memphis_tn", State: "TN", City: "Memphis", Verified: "2026-08-03",
		},
		{
			Name: "Sam Sikes Photography", OfficialURL: "https://samsikesphotography.com/",
			Handle: "samsikesphoto", SourceClass: "engagement_photographer",
			CityID: "city_memphis_tn", State: "TN", City: "Memphis", Verified: "2026-08-03",
		},
		{
			Name: "Memphis Botanic Garden", OfficialURL: "https://membg.org/rentals/",
			Handle: "memphisbotanic", SourceClass: "wedding_venue",
			CityID: "city_memphis_tn", State: "TN", City: "Memphis", Verified: "2026-08-03",
		},
		{
			Name: "The Peabody Memphis", OfficialURL: "https://peabodymemphis.com/weddings/",
			Handle: "peabodymemphis", SourceClass: "wedding_venue",
			CityID: "city_memphis_tn", State: "TN", City: "Memphis", Verified: "2026-08-03",
		},
	},
}
