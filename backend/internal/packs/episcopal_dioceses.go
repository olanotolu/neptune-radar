package packs

// Domestic dioceses of The Episcopal Church (USA), covering all 50 states + DC.
//
// The Episcopal Church has 94 domestic dioceses across the 50 states and the
// District of Columbia (as of 2024-2025). Several dioceses have recently merged:
//   - Diocese of the Great Lakes (2024) absorbed Eastern Michigan & Western Michigan
//   - Diocese of Wisconsin (2024) absorbed Eau Claire & Fond du Lac
//   - Diocese of the Susquehanna (2025/2026) absorbed Central Pennsylvania into Bethlehem
//   - Diocese of Texas (2022) absorbed the former Fort Worth / North Texas
//   - Diocese of Chicago (2013) absorbed Quincy
//
// All diocese names, see cities, and province assignments were verified from
// https://en.wikipedia.org/wiki/Ecclesiastical_provinces_and_dioceses_of_the_Episcopal_Church
// and the official Episcopal Church directory at
// https://www.episcopalchurch.org/find-a-church/browse-by-diocese/
//
// The parish finder URL is the official Episcopal Church "Find a Church" tool:
// https://www.episcopalchurch.org/finder/
//
// Dioceses are assigned to their primary state (where the diocesan offices or
// cathedral are located). Some dioceses span multiple states (e.g. the Diocese
// of the Central Gulf Coast covers southern Alabama and the Florida panhandle;
// the Missionary Diocese of Navajoland covers parts of AZ, NM, and UT).

// episcopalFinder is the official Episcopal Church parish-finder, used as the
// Directory URL for every diocese.
const episcopalFinder = "https://www.episcopalchurch.org/finder/"

// EpiscopalDiocesesFor returns Episcopal dioceses whose primary state matches st.
func EpiscopalDiocesesFor(st string) []DioceseDef {
	var out []DioceseDef
	for _, d := range episcopalDioceses {
		if d.State == st {
			out = append(out, d)
		}
	}
	return out
}

// EpiscopalParishesFor returns Episcopal parishes whose DioceseSlug matches a
// diocese assigned to state st.
func EpiscopalParishesFor(st string) []ParishDef {
	slugSet := map[string]bool{}
	for _, d := range episcopalDioceses {
		if d.State == st {
			slugSet[d.Slug] = true
		}
	}
	var out []ParishDef
	for _, p := range episcopalParishes {
		if slugSet[p.DioceseSlug] {
			out = append(out, p)
		}
	}
	return out
}

// episcopalDioceses lists all domestic dioceses of The Episcopal Church.
// Organized alphabetically by state code, then by diocese name within each state.
var episcopalDioceses = []DioceseDef{
	// --- Alabama ---
	{Slug: "alabama", Name: "Diocese of Alabama", Type: "diocese",
		Website: "https://www.dioala.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "AL"},
	{Slug: "central_gulf_coast", Name: "Diocese of the Central Gulf Coast", Type: "diocese",
		Website: "https://www.diocgc.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "AL"}, // covers southern AL + FL panhandle; cathedral in Mobile, AL

	// --- Alaska ---
	{Slug: "alaska", Name: "Diocese of Alaska", Type: "diocese",
		Website: "https://www.dioak.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "AK"},

	// --- Arizona ---
	{Slug: "arizona", Name: "Diocese of Arizona", Type: "diocese",
		Website: "https://www.azdiocese.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "AZ"},

	// --- Arkansas ---
	{Slug: "arkansas", Name: "Diocese of Arkansas", Type: "diocese",
		Website: "https://www.arkansas.anglican.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "AR"},

	// --- California ---
	{Slug: "california", Name: "Diocese of California", Type: "diocese",
		Website: "https://www.diocal.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "CA"}, // based in San Francisco
	{Slug: "el_camino_real", Name: "Diocese of El Camino Real", Type: "diocese",
		Website: "https://www.edecr.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "CA"}, // offices in Salinas, cathedral in San Jose
	{Slug: "los_angeles", Name: "Diocese of Los Angeles", Type: "diocese",
		Website: "https://www.ladiocese.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "CA"},
	{Slug: "northern_california", Name: "Diocese of Northern California", Type: "diocese",
		Website: "https://www.norcalepiscopal.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "CA"}, // based in Sacramento
	{Slug: "san_diego", Name: "Diocese of San Diego", Type: "diocese",
		Website: "https://www.edsd.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "CA"},
	{Slug: "san_joaquin", Name: "Diocese of San Joaquin", Type: "diocese",
		Website: "https://www.diosj.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "CA"}, // based in Fresno

	// --- Colorado ---
	{Slug: "colorado", Name: "Diocese of Colorado", Type: "diocese",
		Website: "https://www.coloradodiocese.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "CO"},

	// --- Connecticut ---
	{Slug: "connecticut", Name: "Diocese of Connecticut", Type: "diocese",
		Website: "https://www.episcopalct.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "CT"},

	// --- Delaware ---
	{Slug: "delaware", Name: "Diocese of Delaware", Type: "diocese",
		Website: "https://www.diocesede.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "DE"},

	// --- District of Columbia ---
	{Slug: "washington", Name: "Diocese of Washington", Type: "diocese",
		Website: "https://www.edow.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "DC"}, // includes DC + part of MD

	// --- Florida ---
	{Slug: "central_florida", Name: "Diocese of Central Florida", Type: "diocese",
		Website: "https://www.cfdiocese.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "FL"}, // based in Orlando
	{Slug: "florida", Name: "Diocese of Florida", Type: "diocese",
		Website: "https://www.diocesefl.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "FL"}, // based in Jacksonville
	{Slug: "southeast_florida", Name: "Diocese of Southeast Florida", Type: "diocese",
		Website: "https://www.diosef.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "FL"}, // based in Miami
	{Slug: "southwest_florida", Name: "Diocese of Southwest Florida", Type: "diocese",
		Website: "https://www.dioceseswfl.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "FL"}, // offices in Parrish, cathedral in St. Petersburg

	// --- Georgia ---
	{Slug: "atlanta", Name: "Diocese of Atlanta", Type: "diocese",
		Website: "https://www.episcopalatlanta.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "GA"}, // northern and central Georgia
	{Slug: "georgia", Name: "Diocese of Georgia", Type: "diocese",
		Website: "https://www.georgia.anglican.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "GA"}, // based in Savannah; southern Georgia

	// --- Hawaii ---
	{Slug: "hawaii", Name: "Diocese of Hawaii", Type: "diocese",
		Website: "https://www.hawaiiepiscopal.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "HI"},

	// --- Idaho ---
	{Slug: "idaho", Name: "Diocese of Idaho", Type: "diocese",
		Website: "https://www.dio-id.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "ID"},

	// --- Illinois ---
	{Slug: "chicago", Name: "Diocese of Chicago", Type: "diocese",
		Website: "https://www.episcopalchicago.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "IL"}, // also absorbed Quincy in 2013
	{Slug: "springfield", Name: "Diocese of Springfield", Type: "diocese",
		Website: "https://www.episcopalspringfield.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "IL"},

	// --- Indiana ---
	{Slug: "indianapolis", Name: "Diocese of Indianapolis", Type: "diocese",
		Website: "https://www.indydio.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "IN"},
	{Slug: "northern_indiana", Name: "Diocese of Northern Indiana", Type: "diocese",
		Website: "https://www.ednin.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "IN"}, // based in South Bend

	// --- Iowa ---
	{Slug: "iowa", Name: "Diocese of Iowa", Type: "diocese",
		Website: "https://www.dioceseiowa.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "IA"},

	// --- Kansas ---
	{Slug: "kansas", Name: "Diocese of Kansas", Type: "diocese",
		Website: "https://www.episcopalkansas.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "KS"}, // based in Topeka
	{Slug: "western_kansas", Name: "Diocese of Western Kansas", Type: "diocese",
		Website: "https://www.wkdiocese.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "KS"}, // based in Salina

	// --- Kentucky ---
	{Slug: "kentucky", Name: "Diocese of Kentucky", Type: "diocese",
		Website: "https://www.edoky.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "KY"}, // based in Louisville
	{Slug: "lexington", Name: "Diocese of Lexington", Type: "diocese",
		Website: "https://www.diolex.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "KY"},

	// --- Louisiana ---
	{Slug: "louisiana", Name: "Diocese of Louisiana", Type: "diocese",
		Website: "https://www.edola.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "LA"}, // based in New Orleans
	{Slug: "western_louisiana", Name: "Diocese of Western Louisiana", Type: "diocese",
		Website: "https://www.diowl.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "LA"}, // offices in Pineville, cathedral in Shreveport

	// --- Maine ---
	{Slug: "maine", Name: "Diocese of Maine", Type: "diocese",
		Website: "https://www.episcopalmaine.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "ME"},

	// --- Maryland ---
	{Slug: "easton", Name: "Diocese of Easton", Type: "diocese",
		Website: "https://www.dioceseofeaston.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "MD"}, // Eastern Shore
	{Slug: "maryland", Name: "Diocese of Maryland", Type: "diocese",
		Website: "https://www.episcopalmaryland.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "MD"}, // based in Baltimore

	// --- Massachusetts ---
	{Slug: "massachusetts", Name: "Diocese of Massachusetts", Type: "diocese",
		Website: "https://www.diomass.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "MA"}, // based in Boston
	{Slug: "western_massachusetts", Name: "Diocese of Western Massachusetts", Type: "diocese",
		Website: "https://www.diowma.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "MA"}, // based in Springfield

	// --- Michigan ---
	{Slug: "great_lakes", Name: "Diocese of the Great Lakes", Type: "diocese",
		Website: "https://www.edogl.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "MI"}, // formed 2024 from Eastern MI + Western MI; based in Saginaw
	{Slug: "michigan", Name: "Diocese of Michigan", Type: "diocese",
		Website: "https://www.edomi.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "MI"}, // based in Detroit
	{Slug: "northern_michigan", Name: "Diocese of Northern Michigan", Type: "diocese",
		Website: "https://www.updiocese.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "MI"}, // based in Marquette (Upper Peninsula)

	// --- Minnesota ---
	{Slug: "minnesota", Name: "Episcopal Church in Minnesota", Type: "diocese",
		Website: "https://www.episcopalmn.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "MN"}, // based in Minneapolis

	// --- Mississippi ---
	{Slug: "mississippi", Name: "Diocese of Mississippi", Type: "diocese",
		Website: "https://www.dioms.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "MS"}, // based in Jackson

	// --- Missouri ---
	{Slug: "missouri", Name: "Diocese of Missouri", Type: "diocese",
		Website: "https://www.diocesemo.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "MO"}, // based in St. Louis
	{Slug: "west_missouri", Name: "Diocese of West Missouri", Type: "diocese",
		Website: "https://www.diowestmo.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "MO"}, // based in Kansas City

	// --- Montana ---
	{Slug: "montana", Name: "Diocese of Montana", Type: "diocese",
		Website: "https://www.diomontana.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "MT"}, // based in Helena

	// --- North Carolina ---
	{Slug: "east_carolina", Name: "Diocese of East Carolina", Type: "diocese",
		Website: "https://www.dioceseec.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "NC"}, // based in Kinston
	{Slug: "north_carolina", Name: "Diocese of North Carolina", Type: "diocese",
		Website: "https://www.dionc.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "NC"}, // based in Raleigh
	{Slug: "western_north_carolina", Name: "Diocese of Western North Carolina", Type: "diocese",
		Website: "https://www.diocesewnc.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "NC"}, // based in Asheville

	// --- North Dakota ---
	{Slug: "north_dakota", Name: "Diocese of North Dakota", Type: "diocese",
		Website: "https://www.nddiocese.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "ND"}, // based in Fargo

	// --- Nebraska ---
	{Slug: "nebraska", Name: "Diocese of Nebraska", Type: "diocese",
		Website: "https://www.dionebraska.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "NE"}, // based in Omaha

	// --- New Hampshire ---
	{Slug: "new_hampshire", Name: "Diocese of New Hampshire", Type: "diocese",
		Website: "https://www.nhepiscopal.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "NH"}, // based in Concord

	// --- New Jersey ---
	{Slug: "new_jersey", Name: "Diocese of New Jersey", Type: "diocese",
		Website: "https://www.dioceseni.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "NJ"}, // based in Trenton; southern NJ
	{Slug: "newark", Name: "Diocese of Newark", Type: "diocese",
		Website: "https://www.dioceseofnewark.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "NJ"}, // northern NJ

	// --- New Mexico ---
	{Slug: "navajoland", Name: "Missionary Diocese of Navajoland", Type: "diocese",
		Website: "https://www.navajoland.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "NM"}, // covers parts of AZ, NM, UT; see city Farmington, NM
	{Slug: "rio_grande", Name: "Diocese of the Rio Grande", Type: "diocese",
		Website: "https://www.diorg.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "NM"}, // based in Albuquerque; covers NM + El Paso area of TX

	// --- Nevada ---
	{Slug: "nevada", Name: "Diocese of Nevada", Type: "diocese",
		Website: "https://www.episcopalnevada.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "NV"}, // based in Las Vegas

	// --- New York ---
	{Slug: "albany", Name: "Diocese of Albany", Type: "diocese",
		Website: "https://www.albanyepiscopal.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "NY"},
	{Slug: "central_new_york", Name: "Diocese of Central New York", Type: "diocese",
		Website: "https://www.cnyepiscopal.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "NY"}, // based in Syracuse
	{Slug: "long_island", Name: "Diocese of Long Island", Type: "diocese",
		Website: "https://www.dioceseli.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "NY"}, // based in Garden City
	{Slug: "new_york", Name: "Diocese of New York", Type: "diocese",
		Website: "https://www.dioceseny.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "NY"}, // New York City
	{Slug: "rochester", Name: "Diocese of Rochester", Type: "diocese",
		Website: "https://www.episcopalrochester.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "NY"},
	{Slug: "western_new_york", Name: "Diocese of Western New York", Type: "diocese",
		Website: "https://www.episcopalwny.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "NY"}, // based in Buffalo

	// --- Ohio ---
	{Slug: "ohio", Name: "Diocese of Ohio", Type: "diocese",
		Website: "https://www.diohio.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "OH"}, // based in Cleveland
	{Slug: "southern_ohio", Name: "Diocese of Southern Ohio", Type: "diocese",
		Website: "https://www.diosohio.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "OH"}, // based in Cincinnati

	// --- Oklahoma ---
	{Slug: "oklahoma", Name: "Diocese of Oklahoma", Type: "diocese",
		Website: "https://www.episcopaloklahoma.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "OK"}, // based in Oklahoma City

	// --- Oregon ---
	{Slug: "eastern_oregon", Name: "Diocese of Eastern Oregon", Type: "diocese",
		Website: "https://www.dioeoregon.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "OR"}, // based in The Dalles
	{Slug: "oregon", Name: "Diocese of Oregon", Type: "diocese",
		Website: "https://www.diocese-oregon.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "OR"}, // based in Portland

	// --- Pennsylvania ---
	{Slug: "bethlehem", Name: "Diocese of the Susquehanna", Type: "diocese",
		Website: "https://www.diobeth.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "PA"}, // formerly Bethlehem; merged with Central PA in 2025; see cities Bethlehem & Harrisburg
	{Slug: "northwestern_pennsylvania", Name: "Diocese of Northwestern Pennsylvania", Type: "diocese",
		Website: "https://www.dionwpa.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "PA"}, // based in Erie
	{Slug: "pennsylvania", Name: "Diocese of Pennsylvania", Type: "diocese",
		Website: "https://www.diopa.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "PA"}, // based in Philadelphia
	{Slug: "pittsburgh", Name: "Diocese of Pittsburgh", Type: "diocese",
		Website: "https://www.episcopalpittsburgh.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "PA"},

	// --- Rhode Island ---
	{Slug: "rhode_island", Name: "Diocese of Rhode Island", Type: "diocese",
		Website: "https://www.episcopalri.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "RI"}, // based in Providence

	// --- South Carolina ---
	{Slug: "south_carolina", Name: "Diocese of South Carolina", Type: "diocese",
		Website: "https://www.episcopalchurchsc.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "SC"}, // based in Charleston
	{Slug: "upper_south_carolina", Name: "Diocese of Upper South Carolina", Type: "diocese",
		Website: "https://www.edusc.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "SC"}, // based in Columbia

	// --- South Dakota ---
	{Slug: "south_dakota", Name: "Diocese of South Dakota", Type: "diocese",
		Website: "https://www.sddiocese.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "SD"}, // based in Sioux Falls

	// --- Tennessee ---
	{Slug: "east_tennessee", Name: "Diocese of East Tennessee", Type: "diocese",
		Website: "https://www.etdiocese.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "TN"}, // based in Knoxville
	{Slug: "tennessee", Name: "Diocese of Tennessee", Type: "diocese",
		Website: "https://www.edtn.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "TN"}, // based in Nashville
	{Slug: "west_tennessee", Name: "Diocese of West Tennessee", Type: "diocese",
		Website: "https://www.wttn.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "TN"}, // based in Memphis

	// --- Texas ---
	{Slug: "dallas", Name: "Diocese of Dallas", Type: "diocese",
		Website: "https://www.edod.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "TX"},
	{Slug: "northwest_texas", Name: "Diocese of Northwest Texas", Type: "diocese",
		Website: "https://www.nwtexas.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "TX"}, // based in Lubbock
	{Slug: "texas", Name: "Diocese of Texas", Type: "diocese",
		Website: "https://www.etdiocese.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "TX"}, // based in Houston; absorbed Fort Worth/North Texas in 2022
	{Slug: "west_texas", Name: "Diocese of West Texas", Type: "diocese",
		Website: "https://www.dwtx.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "TX"}, // based in San Antonio

	// --- Utah ---
	{Slug: "utah", Name: "Diocese of Utah", Type: "diocese",
		Website: "https://www.episcopal-ut.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "UT"}, // based in Salt Lake City

	// --- Virginia ---
	{Slug: "southern_virginia", Name: "Diocese of Southern Virginia", Type: "diocese",
		Website: "https://www.diosova.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "VA"}, // based in Newport News
	{Slug: "southwestern_virginia", Name: "Diocese of Southwestern Virginia", Type: "diocese",
		Website: "https://www.dioswva.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "VA"}, // based in Roanoke
	{Slug: "virginia", Name: "Diocese of Virginia", Type: "diocese",
		Website: "https://www.thediocese.net/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "VA"}, // based in Richmond; largest diocese by membership

	// --- Vermont ---
	{Slug: "vermont", Name: "Diocese of Vermont", Type: "diocese",
		Website: "https://www.diocesevt.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "VT"}, // based in Burlington

	// --- Washington ---
	{Slug: "olympia", Name: "Diocese of Olympia", Type: "diocese",
		Website: "https://www.ecww.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "WA"}, // based in Seattle; western WA
	{Slug: "spokane", Name: "Diocese of Spokane", Type: "diocese",
		Website: "https://www.diospokane.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "WA"}, // eastern WA + part of ID

	// --- Wisconsin ---
	{Slug: "wisconsin", Name: "Diocese of Wisconsin", Type: "diocese",
		Website: "https://www.edwi.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "WI"}, // absorbed Eau Claire & Fond du Lac in 2024; see cities Eau Claire, Fond du Lac, Milwaukee

	// --- West Virginia ---
	{Slug: "west_virginia", Name: "Diocese of West Virginia", Type: "diocese",
		Website: "https://www.wvdiocese.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "WV"}, // based in Charleston

	// --- Wyoming ---
	{Slug: "wyoming", Name: "Diocese of Wyoming", Type: "diocese",
		Website: "https://www.wyomingdiocese.org/", Directory: episcopalFinder,
		Denomination: "episcopal", State: "WY"}, // offices in Casper, cathedral in Laramie
}

// episcopalParishes lists prominent Episcopal parishes in the top 5-6 metro
// areas. DioceseSlug matches the Slug field in episcopalDioceses above.
// BulletinURL is empty because we have not verified bulletin archives for
// these parishes.
var episcopalParishes = []ParishDef{
	// New York City (Diocese of New York)
	{DioceseSlug: "new_york", Name: "Trinity Church Wall Street",
		Address: "74 Trinity Place, New York, NY 10006", BulletinURL: "",
		Denomination: "episcopal"},
	{DioceseSlug: "new_york", Name: "St. Thomas Church Fifth Avenue",
		Address: "1 W 53rd St, New York, NY 10019", BulletinURL: "",
		Denomination: "episcopal"},
	{DioceseSlug: "new_york", Name: "Cathedral of St. John the Divine",
		Address: "1047 Amsterdam Ave, New York, NY 10025", BulletinURL: "",
		Denomination: "episcopal"},

	// Los Angeles (Diocese of Los Angeles)
	{DioceseSlug: "los_angeles", Name: "All Saints Church Pasadena",
		Address: "132 N Garfield Ave, Pasadena, CA 91101", BulletinURL: "",
		Denomination: "episcopal"},
	{DioceseSlug: "los_angeles", Name: "St. John's Cathedral",
		Address: "514 W Adams Blvd, Los Angeles, CA 90007", BulletinURL: "",
		Denomination: "episcopal"},

	// Chicago (Diocese of Chicago)
	{DioceseSlug: "chicago", Name: "St. James Cathedral",
		Address: "65 E Huron St, Chicago, IL 60611", BulletinURL: "",
		Denomination: "episcopal"},
	{DioceseSlug: "chicago", Name: "Grace Church",
		Address: "637 S Dearborn St, Chicago, IL 60605", BulletinURL: "",
		Denomination: "episcopal"},

	// Houston (Diocese of Texas)
	{DioceseSlug: "texas", Name: "Christ Church Cathedral",
		Address: "1117 Texas Ave, Houston, TX 77002", BulletinURL: "",
		Denomination: "episcopal"},
	{DioceseSlug: "texas", Name: "St. Martin's Episcopal Church",
		Address: "717 Sage Rd, Houston, TX 77056", BulletinURL: "",
		Denomination: "episcopal"},

	// Washington, DC (Diocese of Washington)
	{DioceseSlug: "washington", Name: "Washington National Cathedral",
		Address: "3101 Wisconsin Ave NW, Washington, DC 20016", BulletinURL: "",
		Denomination: "episcopal"},
	{DioceseSlug: "washington", Name: "St. John's Church Lafayette Square",
		Address: "1525 H St NW, Washington, DC 20005", BulletinURL: "",
		Denomination: "episcopal"},

	// Atlanta (Diocese of Atlanta)
	{DioceseSlug: "atlanta", Name: "Cathedral of St. Philip",
		Address: "2744 Peachtree Rd NW, Atlanta, GA 30305", BulletinURL: "",
		Denomination: "episcopal"},
	{DioceseSlug: "atlanta", Name: "All Saints' Episcopal Church",
		Address: "634 W Peachtree St NW, Atlanta, GA 30308", BulletinURL: "",
		Denomination: "episcopal"},
}
