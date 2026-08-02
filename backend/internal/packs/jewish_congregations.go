package packs

// Jewish federations and community organizations for all 50 states + DC, plus
// prominent synagogues for the top 6 metro areas.
//
// The Jewish federations of North America (JFNA) is the umbrella for ~146 US
// federations. Not every state has a formal federation; small-state entries use
// the nearest major federation or a state-level Jewish community council, as
// noted in each entry's comment.
//
// Federation names and URLs were verified from the JFNA public directory
// (https://docslib.org/doc/854103/jewish-federations-of-north-america-leadership)
// and individual web searches confirming each organization's current homepage.
//
// Synagogue names and addresses are well-known, prominent congregations verified
// via web search. BulletinURL is empty (no verified bulletin archives).

// JewishFederationsFor returns Jewish federations/community organizations whose
// primary state matches st.
func JewishFederationsFor(st string) []DioceseDef {
	var out []DioceseDef
	for _, d := range jewishFederations {
		if d.State == st {
			out = append(out, d)
		}
	}
	return out
}

// JewishSynagoguesFor returns synagogues whose DioceseSlug matches a federation
// assigned to state st.
func JewishSynagoguesFor(st string) []ParishDef {
	slugSet := map[string]bool{}
	for _, d := range jewishFederations {
		if d.State == st {
			slugSet[d.Slug] = true
		}
	}
	var out []ParishDef
	for _, p := range jewishSynagogues {
		if slugSet[p.DioceseSlug] {
			out = append(out, p)
		}
	}
	return out
}

// jewishFederations lists major Jewish federations and community organizations
// covering all 50 states + DC. Organized alphabetically by state code.
var jewishFederations = []DioceseDef{
	// --- Alabama ---
	{Slug: "jewish_fed_birmingham", Name: "The Birmingham Jewish Federation", Type: "federation",
		Website: "https://www.bjf.org/", Directory: "https://www.bjf.org/",
		Denomination: "jewish", State: "AL"},

	// --- Alaska ---
	// No formal federation; Alaska Jewish Campus is the primary Jewish community organization.
	{Slug: "jewish_fed_alaska", Name: "Alaska Jewish Campus", Type: "federation",
		Website: "https://alaskajewishcampus.org/", Directory: "https://alaskajewishcampus.org/",
		Denomination: "jewish", State: "AK"},

	// --- Arizona ---
	{Slug: "jewish_fed_phoenix", Name: "Jewish Federation of Greater Phoenix", Type: "federation",
		Website: "https://www.jewishphoenix.org/", Directory: "https://www.jewishphoenix.org/",
		Denomination: "jewish", State: "AZ"},
	{Slug: "jewish_fed_tucson", Name: "Jewish Federation of Southern Arizona", Type: "federation",
		Website: "https://www.jewishtucson.org/", Directory: "https://www.jewishtucson.org/",
		Denomination: "jewish", State: "AZ"},

	// --- Arkansas ---
	{Slug: "jewish_fed_arkansas", Name: "Jewish Federation of Arkansas", Type: "federation",
		Website: "https://www.jewisharkansas.org/", Directory: "https://www.jewisharkansas.org/",
		Denomination: "jewish", State: "AR"},

	// --- California ---
	{Slug: "jewish_fed_la", Name: "Jewish Federation of Greater Los Angeles", Type: "federation",
		Website: "https://www.jewishla.org/", Directory: "https://www.jewishla.org/",
		Denomination: "jewish", State: "CA"},
	{Slug: "jewish_fed_sd", Name: "Jewish Federation of San Diego", Type: "federation",
		Website: "https://jewishinsandiego.org/", Directory: "https://jewishinsandiego.org/",
		Denomination: "jewish", State: "CA"},
	{Slug: "jewish_fed_sf", Name: "Jewish Community Federation of San Francisco", Type: "federation",
		Website: "https://www.jewishfed.org/", Directory: "https://www.jewishfed.org/",
		Denomination: "jewish", State: "CA"},
	{Slug: "jewish_fed_sacramento", Name: "Jewish Federation of the Sacramento Region", Type: "federation",
		Website: "https://www.jewishsacramento.org/", Directory: "https://www.jewishsacramento.org/",
		Denomination: "jewish", State: "CA"},
	{Slug: "jewish_fed_silicon_valley", Name: "Jewish Federation of Silicon Valley", Type: "federation",
		Website: "https://www.jewishsiliconvalley.org/", Directory: "https://www.jewishsiliconvalley.org/",
		Denomination: "jewish", State: "CA"},

	// --- Colorado ---
	{Slug: "jewish_fed_colorado", Name: "JEWISHcolorado", Type: "federation",
		Website: "https://www.jewishcolorado.org/", Directory: "https://www.jewishcolorado.org/",
		Denomination: "jewish", State: "CO"},

	// --- Connecticut ---
	{Slug: "jewish_fed_hartford", Name: "Jewish Federation of Greater Hartford", Type: "federation",
		Website: "https://www.jewishhartford.org/", Directory: "https://www.jewishhartford.org/",
		Denomination: "jewish", State: "CT"},
	{Slug: "jewish_fed_new_haven", Name: "Jewish Federation of Greater New Haven", Type: "federation",
		Website: "https://www.jewishnewhaven.org/", Directory: "https://www.jewishnewhaven.org/",
		Denomination: "jewish", State: "CT"},

	// --- Delaware ---
	{Slug: "jewish_fed_delaware", Name: "Jewish Federation of Delaware", Type: "federation",
		Website: "https://www.shalomdelaware.org/", Directory: "https://www.shalomdelaware.org/",
		Denomination: "jewish", State: "DE"},

	// --- District of Columbia ---
	{Slug: "jewish_fed_dc", Name: "The Jewish Federation of Greater Washington", Type: "federation",
		Website: "https://www.shalomdc.org/", Directory: "https://www.shalomdc.org/",
		Denomination: "jewish", State: "DC"},

	// --- Florida ---
	{Slug: "jewish_fed_miami", Name: "Greater Miami Jewish Federation", Type: "federation",
		Website: "https://jewishmiami.org/", Directory: "https://jewishmiami.org/",
		Denomination: "jewish", State: "FL"},
	{Slug: "jewish_fed_broward", Name: "Jewish Federation of Broward County", Type: "federation",
		Website: "https://jewishbroward.org/", Directory: "https://jewishbroward.org/",
		Denomination: "jewish", State: "FL"},
	{Slug: "jewish_fed_orlando", Name: "Jewish Federation of Greater Orlando", Type: "federation",
		Website: "https://www.jfgo.org/", Directory: "https://www.jfgo.org/",
		Denomination: "jewish", State: "FL"},
	{Slug: "jewish_fed_palm_beach", Name: "Jewish Federation of Palm Beach County", Type: "federation",
		Website: "https://www.jewishpalmbeach.org/", Directory: "https://www.jewishpalmbeach.org/",
		Denomination: "jewish", State: "FL"},
	{Slug: "jewish_fed_south_palm_beach", Name: "Jewish Federation of South Palm Beach County", Type: "federation",
		Website: "https://jewishboca.org/", Directory: "https://jewishboca.org/",
		Denomination: "jewish", State: "FL"},

	// --- Georgia ---
	{Slug: "jewish_fed_atlanta", Name: "Jewish Federation of Greater Atlanta", Type: "federation",
		Website: "https://www.jewishatlanta.org/", Directory: "https://www.jewishatlanta.org/",
		Denomination: "jewish", State: "GA"},

	// --- Hawaii ---
	// The Jewish Federation of Hawaii closed in 1998; Jewish Community Services of
	// Hawaii is the primary Jewish community organization.
	{Slug: "jewish_fed_hawaii", Name: "Jewish Community Services of Hawaii", Type: "federation",
		Website: "https://jcs-hi.org/", Directory: "https://jcs-hi.org/",
		Denomination: "jewish", State: "HI"},

	// --- Idaho ---
	// No formal federation; Idaho Jewish Alliance is the state-level Jewish community organization.
	{Slug: "jewish_fed_idaho", Name: "Idaho Jewish Alliance", Type: "federation",
		Website: "https://www.idahojewishalliance.com/", Directory: "https://www.idahojewishalliance.com/",
		Denomination: "jewish", State: "ID"},

	// --- Illinois ---
	{Slug: "jewish_fed_chicago", Name: "Jewish United Fund / Jewish Federation of Metropolitan Chicago", Type: "federation",
		Website: "https://www.juf.org/", Directory: "https://www.juf.org/",
		Denomination: "jewish", State: "IL"},

	// --- Indiana ---
	{Slug: "jewish_fed_indianapolis", Name: "Jewish Federation of Greater Indianapolis", Type: "federation",
		Website: "https://www.jfgi.org/", Directory: "https://www.jfgi.org/",
		Denomination: "jewish", State: "IN"},

	// --- Iowa ---
	{Slug: "jewish_fed_des_moines", Name: "Jewish Federation of Greater Des Moines", Type: "federation",
		Website: "https://www.jewishdesmoines.org/", Directory: "https://www.jewishdesmoines.org/",
		Denomination: "jewish", State: "IA"},

	// --- Kansas ---
	{Slug: "jewish_fed_kansas_city", Name: "Jewish Federation of Greater Kansas City", Type: "federation",
		Website: "https://www.jewishkansascity.org/", Directory: "https://www.jewishkansascity.org/",
		Denomination: "jewish", State: "KS"},

	// --- Kentucky ---
	{Slug: "jewish_fed_louisville", Name: "Jewish Community of Louisville", Type: "federation",
		Website: "https://www.jewishlouisville.org/", Directory: "https://www.jewishlouisville.org/",
		Denomination: "jewish", State: "KY"},

	// --- Louisiana ---
	{Slug: "jewish_fed_new_orleans", Name: "Jewish Federation of Greater New Orleans", Type: "federation",
		Website: "https://www.jewishnola.com/", Directory: "https://www.jewishnola.com/",
		Denomination: "jewish", State: "LA"},

	// --- Maine ---
	{Slug: "jewish_fed_maine", Name: "Jewish Community Alliance of Southern Maine", Type: "federation",
		Website: "https://www.mainejewish.org/", Directory: "https://www.mainejewish.org/",
		Denomination: "jewish", State: "ME"},

	// --- Maryland ---
	{Slug: "jewish_fed_baltimore", Name: "The Associated: Jewish Community Federation of Baltimore", Type: "federation",
		Website: "https://www.associated.org/", Directory: "https://www.associated.org/",
		Denomination: "jewish", State: "MD"},

	// --- Massachusetts ---
	{Slug: "jewish_fed_boston", Name: "Combined Jewish Philanthropies of Greater Boston", Type: "federation",
		Website: "https://www.cjp.org/", Directory: "https://www.cjp.org/",
		Denomination: "jewish", State: "MA"},

	// --- Michigan ---
	{Slug: "jewish_fed_detroit", Name: "Jewish Federation of Metropolitan Detroit", Type: "federation",
		Website: "https://www.thisisfederation.org/", Directory: "https://www.thisisfederation.org/",
		Denomination: "jewish", State: "MI"},

	// --- Minnesota ---
	{Slug: "jewish_fed_minneapolis", Name: "Minneapolis Jewish Federation", Type: "federation",
		Website: "https://www.jewishminneapolis.org/", Directory: "https://www.jewishminneapolis.org/",
		Denomination: "jewish", State: "MN"},
	{Slug: "jewish_fed_st_paul", Name: "St. Paul Jewish Federation", Type: "federation",
		Website: "https://www.jewishminnesota.org/", Directory: "https://www.jewishminnesota.org/",
		Denomination: "jewish", State: "MN"},

	// --- Mississippi ---
	// No formal federation; the nearest is the Jewish Federation of Greater New Orleans.
	{Slug: "jewish_fed_mississippi", Name: "Jewish Federation of Greater New Orleans (Mississippi)", Type: "federation",
		Website: "https://www.jewishnola.com/", Directory: "https://www.jewishnola.com/",
		Denomination: "jewish", State: "MS"},

	// --- Missouri ---
	{Slug: "jewish_fed_st_louis", Name: "Jewish Federation of St. Louis", Type: "federation",
		Website: "https://www.jewishinstlouis.org/", Directory: "https://www.jewishinstlouis.org/",
		Denomination: "jewish", State: "MO"},

	// --- Montana ---
	// No formal federation; Montana Jewish Project is the state-level Jewish community organization.
	{Slug: "jewish_fed_montana", Name: "Montana Jewish Project", Type: "federation",
		Website: "https://www.montanajewishproject.org/", Directory: "https://www.montanajewishproject.org/",
		Denomination: "jewish", State: "MT"},

	// --- North Carolina ---
	{Slug: "jewish_fed_charlotte", Name: "Jewish Federation of Greater Charlotte", Type: "federation",
		Website: "https://www.jewishcharlotte.org/", Directory: "https://www.jewishcharlotte.org/",
		Denomination: "jewish", State: "NC"},
	{Slug: "jewish_fed_raleigh", Name: "Jewish Federation of Raleigh-Cary", Type: "federation",
		Website: "https://www.shalomraleigh.org/", Directory: "https://www.shalomraleigh.org/",
		Denomination: "jewish", State: "NC"},

	// --- North Dakota ---
	// No formal federation; Chabad Jewish Center of North Dakota is the primary Jewish organization.
	{Slug: "jewish_fed_nd", Name: "Chabad Jewish Center of North Dakota", Type: "federation",
		Website: "https://jewishnorthdakota.com/", Directory: "https://jewishnorthdakota.com/",
		Denomination: "jewish", State: "ND"},

	// --- Nebraska ---
	{Slug: "jewish_fed_omaha", Name: "Jewish Federation of Omaha", Type: "federation",
		Website: "https://www.jewishomaha.org/", Directory: "https://www.jewishomaha.org/",
		Denomination: "jewish", State: "NE"},

	// --- Nevada ---
	{Slug: "jewish_fed_nevada", Name: "JewishNevada", Type: "federation",
		Website: "https://www.jewishnevada.org/", Directory: "https://www.jewishnevada.org/",
		Denomination: "jewish", State: "NV"},

	// --- New Hampshire ---
	{Slug: "jewish_fed_nh", Name: "Jewish Federation of New Hampshire", Type: "federation",
		Website: "https://jewishnh.org/", Directory: "https://jewishnh.org/",
		Denomination: "jewish", State: "NH"},

	// --- New Jersey ---
	{Slug: "jewish_fed_north_nj", Name: "Jewish Federation of Northern New Jersey", Type: "federation",
		Website: "https://www.jfnnj.org/", Directory: "https://www.jfnnj.org/",
		Denomination: "jewish", State: "NJ"},
	{Slug: "jewish_fed_metrowest_nj", Name: "Jewish Federation of Greater MetroWest NJ", Type: "federation",
		Website: "https://www.jfedgmw.org/", Directory: "https://www.jfedgmw.org/",
		Denomination: "jewish", State: "NJ"},
	{Slug: "jewish_fed_south_nj", Name: "Jewish Federation of Southern New Jersey", Type: "federation",
		Website: "https://www.jfedsnj.org/", Directory: "https://www.jfedsnj.org/",
		Denomination: "jewish", State: "NJ"},

	// --- New Mexico ---
	{Slug: "jewish_fed_nm", Name: "Jewish Federation of New Mexico", Type: "federation",
		Website: "https://www.jewishnewmexico.org/", Directory: "https://www.jewishnewmexico.org/",
		Denomination: "jewish", State: "NM"},

	// --- New York ---
	{Slug: "jewish_fed_nyc", Name: "UJA-Federation of New York", Type: "federation",
		Website: "https://www.ujafedny.org/", Directory: "https://www.ujafedny.org/",
		Denomination: "jewish", State: "NY"},
	{Slug: "jewish_fed_rochester", Name: "Jewish Federation of Greater Rochester", Type: "federation",
		Website: "https://www.jewishrochester.org/", Directory: "https://www.jewishrochester.org/",
		Denomination: "jewish", State: "NY"},
	{Slug: "jewish_fed_buffalo", Name: "Buffalo Jewish Federation", Type: "federation",
		Website: "https://www.buffalojewishfederation.org/", Directory: "https://www.buffalojewishfederation.org/",
		Denomination: "jewish", State: "NY"},
	{Slug: "jewish_fed_ne_ny", Name: "Jewish Federation of Northeastern New York", Type: "federation",
		Website: "https://www.jewishfedny.org/", Directory: "https://www.jewishfedny.org/",
		Denomination: "jewish", State: "NY"},

	// --- Ohio ---
	{Slug: "jewish_fed_cincinnati", Name: "Jewish Federation of Cincinnati", Type: "federation",
		Website: "https://www.jewishcincinnati.org/", Directory: "https://www.jewishcincinnati.org/",
		Denomination: "jewish", State: "OH"},
	{Slug: "jewish_fed_cleveland", Name: "Jewish Federation of Cleveland", Type: "federation",
		Website: "https://www.jewishcleveland.org/", Directory: "https://www.jewishcleveland.org/",
		Denomination: "jewish", State: "OH"},
	{Slug: "jewish_fed_columbus", Name: "Jewish Federation of Columbus", Type: "federation",
		Website: "https://www.jewishcolumbus.org/", Directory: "https://www.jewishcolumbus.org/",
		Denomination: "jewish", State: "OH"},

	// --- Oklahoma ---
	{Slug: "jewish_fed_okc", Name: "Jewish Federation of Greater Oklahoma City", Type: "federation",
		Website: "https://www.jfedokc.org/", Directory: "https://www.jfedokc.org/",
		Denomination: "jewish", State: "OK"},
	{Slug: "jewish_fed_tulsa", Name: "Jewish Federation of Tulsa", Type: "federation",
		Website: "https://www.jewishtulsa.org/", Directory: "https://www.jewishtulsa.org/",
		Denomination: "jewish", State: "OK"},

	// --- Oregon ---
	{Slug: "jewish_fed_portland", Name: "Jewish Federation of Greater Portland", Type: "federation",
		Website: "https://www.jewishportland.org/", Directory: "https://www.jewishportland.org/",
		Denomination: "jewish", State: "OR"},

	// --- Pennsylvania ---
	{Slug: "jewish_fed_philly", Name: "Jewish Federation of Greater Philadelphia", Type: "federation",
		Website: "https://www.jewishphilly.org/", Directory: "https://www.jewishphilly.org/",
		Denomination: "jewish", State: "PA"},
	{Slug: "jewish_fed_pittsburgh", Name: "Jewish Federation of Greater Pittsburgh", Type: "federation",
		Website: "https://www.jfedpgh.org/", Directory: "https://www.jfedpgh.org/",
		Denomination: "jewish", State: "PA"},

	// --- Rhode Island ---
	{Slug: "jewish_fed_ri", Name: "Jewish Alliance of Greater Rhode Island", Type: "federation",
		Website: "https://jewishallianceri.org/", Directory: "https://jewishallianceri.org/",
		Denomination: "jewish", State: "RI"},

	// --- South Carolina ---
	{Slug: "jewish_fed_charleston_sc", Name: "Charleston Jewish Federation", Type: "federation",
		Website: "https://jewishcharleston.org/", Directory: "https://jewishcharleston.org/",
		Denomination: "jewish", State: "SC"},
	{Slug: "jewish_fed_columbia_sc", Name: "Columbia Jewish Federation", Type: "federation",
		Website: "https://www.jewishcolumbia.org/", Directory: "https://www.jewishcolumbia.org/",
		Denomination: "jewish", State: "SC"},

	// --- South Dakota ---
	// No formal federation; South Dakota Jewish Center is the primary Jewish organization.
	{Slug: "jewish_fed_sd", Name: "South Dakota Jewish Center", Type: "federation",
		Website: "https://www.jewishsd.org/", Directory: "https://www.jewishsd.org/",
		Denomination: "jewish", State: "SD"},

	// --- Tennessee ---
	{Slug: "jewish_fed_nashville", Name: "Jewish Federation of Nashville and Middle Tennessee", Type: "federation",
		Website: "https://www.jewishnashville.org/", Directory: "https://www.jewishnashville.org/",
		Denomination: "jewish", State: "TN"},
	{Slug: "jewish_fed_memphis", Name: "Memphis Jewish Federation", Type: "federation",
		Website: "https://www.jcpmemphis.org/", Directory: "https://www.jcpmemphis.org/",
		Denomination: "jewish", State: "TN"},

	// --- Texas ---
	{Slug: "jewish_fed_houston", Name: "Jewish Federation of Greater Houston", Type: "federation",
		Website: "https://www.houstonjewish.org/", Directory: "https://www.houstonjewish.org/",
		Denomination: "jewish", State: "TX"},
	{Slug: "jewish_fed_dallas", Name: "Jewish Federation of Greater Dallas", Type: "federation",
		Website: "https://www.jewishdallas.org/", Directory: "https://www.jewishdallas.org/",
		Denomination: "jewish", State: "TX"},
	{Slug: "jewish_fed_austin", Name: "Shalom Austin", Type: "federation",
		Website: "https://www.shalomaustin.org/", Directory: "https://www.shalomaustin.org/",
		Denomination: "jewish", State: "TX"},

	// --- Utah ---
	{Slug: "jewish_fed_utah", Name: "United Jewish Federation of Utah", Type: "federation",
		Website: "https://shalomutah.org/", Directory: "https://shalomutah.org/",
		Denomination: "jewish", State: "UT"},

	// --- Virginia ---
	{Slug: "jewish_fed_richmond", Name: "Jewish Community Federation of Richmond", Type: "federation",
		Website: "https://jewishrichmond.org/", Directory: "https://jewishrichmond.org/",
		Denomination: "jewish", State: "VA"},
	{Slug: "jewish_fed_tidewater", Name: "United Jewish Federation of Tidewater", Type: "federation",
		Website: "https://federation.jewishva.org/", Directory: "https://federation.jewishva.org/",
		Denomination: "jewish", State: "VA"},

	// --- Vermont ---
	// No formal federation; Jewish Communities of Vermont is the state-level Jewish community organization.
	{Slug: "jewish_fed_vermont", Name: "Jewish Communities of Vermont", Type: "federation",
		Website: "https://www.jewishcommunitiesofvermont.org/", Directory: "https://www.jewishcommunitiesofvermont.org/",
		Denomination: "jewish", State: "VT"},

	// --- Washington ---
	{Slug: "jewish_fed_seattle", Name: "Jewish Federation of Greater Seattle", Type: "federation",
		Website: "https://jewishinseattle.org/", Directory: "https://jewishinseattle.org/",
		Denomination: "jewish", State: "WA"},

	// --- Wisconsin ---
	{Slug: "jewish_fed_milwaukee", Name: "Milwaukee Jewish Federation", Type: "federation",
		Website: "https://milwaukeejewish.org/", Directory: "https://milwaukeejewish.org/",
		Denomination: "jewish", State: "WI"},

	// --- West Virginia ---
	{Slug: "jewish_fed_wv", Name: "Federated Jewish Charities of Charleston", Type: "federation",
		Website: "http://www.fjcofcharleston.org/", Directory: "http://www.fjcofcharleston.org/",
		Denomination: "jewish", State: "WV"},

	// --- Wyoming ---
	// No formal federation; Chabad Lubavitch of Wyoming is the primary Jewish organization.
	{Slug: "jewish_fed_wyoming", Name: "Chabad Lubavitch of Wyoming", Type: "federation",
		Website: "https://www.jewishwyoming.com/", Directory: "https://www.jewishwyoming.com/",
		Denomination: "jewish", State: "WY"},
}

// jewishSynagogues lists prominent synagogues in the top 6 metro areas.
// DioceseSlug matches the Slug field in jewishFederations above.
var jewishSynagogues = []ParishDef{
	// New York City (UJA-Federation of New York)
	{DioceseSlug: "jewish_fed_nyc", Name: "Central Synagogue",
		Address: "652 Lexington Ave, New York, NY 10022", BulletinURL: "",
		Denomination: "jewish"},
	{DioceseSlug: "jewish_fed_nyc", Name: "Temple Emanu-El",
		Address: "1 E 65th St, New York, NY 10065", BulletinURL: "",
		Denomination: "jewish"},
	{DioceseSlug: "jewish_fed_nyc", Name: "B'nai Jeshurun",
		Address: "270 W 89th St, New York, NY 10024", BulletinURL: "",
		Denomination: "jewish"},

	// Los Angeles (Jewish Federation of Greater Los Angeles)
	{DioceseSlug: "jewish_fed_la", Name: "Wilshire Boulevard Temple",
		Address: "3663 Wilshire Blvd, Los Angeles, CA 90010", BulletinURL: "",
		Denomination: "jewish"},
	{DioceseSlug: "jewish_fed_la", Name: "Sinai Temple",
		Address: "10400 Wilshire Blvd, Los Angeles, CA 90024", BulletinURL: "",
		Denomination: "jewish"},
	{DioceseSlug: "jewish_fed_la", Name: "Valley Beth Shalom",
		Address: "15739 Ventura Blvd, Encino, CA 91436", BulletinURL: "",
		Denomination: "jewish"},

	// Chicago (Jewish United Fund / Jewish Federation of Metropolitan Chicago)
	{DioceseSlug: "jewish_fed_chicago", Name: "Anshe Emet Synagogue",
		Address: "3751 N. Broadway, Chicago, IL 60613", BulletinURL: "",
		Denomination: "jewish"},
	{DioceseSlug: "jewish_fed_chicago", Name: "Temple Sholom of Chicago",
		Address: "3480 N. Lake Shore Dr, Chicago, IL 60657", BulletinURL: "",
		Denomination: "jewish"},

	// Houston (Jewish Federation of Greater Houston)
	{DioceseSlug: "jewish_fed_houston", Name: "Congregation Beth Israel",
		Address: "5600 N. Braeswood Blvd, Houston, TX 77096", BulletinURL: "",
		Denomination: "jewish"},
	{DioceseSlug: "jewish_fed_houston", Name: "Congregation Emanu El",
		Address: "1500 Sunset Blvd, Houston, TX 77005", BulletinURL: "",
		Denomination: "jewish"},

	// Washington, DC (Jewish Federation of Greater Washington)
	{DioceseSlug: "jewish_fed_dc", Name: "Washington Hebrew Congregation",
		Address: "3935 Macomb St NW, Washington, DC 20016", BulletinURL: "",
		Denomination: "jewish"},
	{DioceseSlug: "jewish_fed_dc", Name: "Adas Israel Congregation",
		Address: "2850 Quebec St NW, Washington, DC 20008", BulletinURL: "",
		Denomination: "jewish"},

	// Atlanta (Jewish Federation of Greater Atlanta)
	{DioceseSlug: "jewish_fed_atlanta", Name: "The Temple (Hebrew Benevolent Congregation)",
		Address: "1589 Peachtree St NE, Atlanta, GA 30309", BulletinURL: "",
		Denomination: "jewish"},
	{DioceseSlug: "jewish_fed_atlanta", Name: "Ahavath Achim Synagogue",
		Address: "600 Highland Pkwy NE, Atlanta, GA 30342", BulletinURL: "",
		Denomination: "jewish"},
}
