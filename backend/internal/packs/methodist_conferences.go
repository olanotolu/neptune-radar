package packs

// United Methodist Church annual conferences for all 50 states + DC.
//
// The UMC has 51 annual conferences in the US (including 3 missionary
// conferences). Each conference has a church directory at the official UMC
// Find-A-Church tool (https://www.umc.org/en/find-a-church).
//
// All conference names and website URLs were verified from the official UMC
// directory page: https://www.umc.org/en/content/annual-conferences-directory-us
// (fetched 2026-08-01). The Find-A-Church directory URL was verified at
// https://www.umc.org/en/find-a-church.
//
// Conferences are assigned to their primary state (where the conference office
// is located). Several conferences span multiple states (e.g. Dakotas covers
// ND+SD, Great Plains covers KS+NE, Mountain Sky covers CO+MT+WY+UT).

// umcFindAChurch is the official UMC church-finder, used as the Directory URL
// for every annual conference.
const umcFindAChurch = "https://www.umc.org/en/find-a-church"

// MethodistConferencesFor returns UMC annual conferences whose primary state
// matches st. Some conferences span multiple states (e.g. Dakotas covers ND+SD)
// but are assigned to one primary state for filtering.
func MethodistConferencesFor(st string) []DioceseDef {
	var out []DioceseDef
	for _, d := range methodistConferences {
		if d.State == st {
			out = append(out, d)
		}
	}
	return out
}

// MethodistChurchesFor returns UMC churches whose diocese slug matches a
// conference assigned to state st.
func MethodistChurchesFor(st string) []ParishDef {
	slugSet := map[string]bool{}
	for _, d := range methodistConferences {
		if d.State == st {
			slugSet[d.Slug] = true
		}
	}
	var out []ParishDef
	for _, p := range methodistChurches {
		if slugSet[p.DioceseSlug] {
			out = append(out, p)
		}
	}
	return out
}

var methodistConferences = []DioceseDef{
	// --- North Central Jurisdiction ---
	{Slug: "dakotas", Name: "Dakotas Annual Conference", Type: "annual_conference",
		Website: "https://www.dakotasumc.org/", Directory: umcFindAChurch,
		Denomination: "methodist", State: "SD"}, // covers ND + SD
	{Slug: "east_ohio", Name: "East Ohio Annual Conference", Type: "annual_conference",
		Website: "http://www.eocumc.com/", Directory: umcFindAChurch,
		Denomination: "methodist", State: "OH"},
	{Slug: "illinois_great_rivers", Name: "Illinois Great Rivers Annual Conference", Type: "annual_conference",
		Website: "https://www.igrc.org/", Directory: umcFindAChurch,
		Denomination: "methodist", State: "IL"},
	{Slug: "indiana", Name: "Indiana Annual Conference", Type: "annual_conference",
		Website: "http://www.inumc.org/", Directory: umcFindAChurch,
		Denomination: "methodist", State: "IN"},
	{Slug: "iowa", Name: "Iowa Annual Conference", Type: "annual_conference",
		Website: "https://www.iaumc.org/", Directory: umcFindAChurch,
		Denomination: "methodist", State: "IA"},
	{Slug: "michigan", Name: "Michigan Annual Conference", Type: "annual_conference",
		Website: "https://michiganumc.org/", Directory: umcFindAChurch,
		Denomination: "methodist", State: "MI"},
	{Slug: "minnesota", Name: "Minnesota Annual Conference", Type: "annual_conference",
		Website: "https://www.minnesotaumc.org/", Directory: umcFindAChurch,
		Denomination: "methodist", State: "MN"},
	{Slug: "northern_illinois", Name: "Northern Illinois Annual Conference", Type: "annual_conference",
		Website: "https://www.umcnic.org/", Directory: umcFindAChurch,
		Denomination: "methodist", State: "IL"},
	{Slug: "west_ohio", Name: "West Ohio Annual Conference", Type: "annual_conference",
		Website: "https://www.westohioumc.org/", Directory: umcFindAChurch,
		Denomination: "methodist", State: "OH"},
	{Slug: "wisconsin", Name: "Wisconsin Annual Conference", Type: "annual_conference",
		Website: "https://www.wisconsinumc.org/", Directory: umcFindAChurch,
		Denomination: "methodist", State: "WI"},

	// --- Northeastern Jurisdiction ---
	{Slug: "baltimore_washington", Name: "Baltimore-Washington Annual Conference", Type: "annual_conference",
		Website: "http://www.bwcumc.org/", Directory: umcFindAChurch,
		Denomination: "methodist", State: "MD"}, // covers MD + DC
	{Slug: "eastern_pennsylvania", Name: "Eastern Pennsylvania Annual Conference", Type: "annual_conference",
		Website: "https://www.epaumc.org/", Directory: umcFindAChurch,
		Denomination: "methodist", State: "PA"},
	{Slug: "greater_new_jersey", Name: "Greater New Jersey Annual Conference", Type: "annual_conference",
		Website: "https://www.gnjumc.org/", Directory: umcFindAChurch,
		Denomination: "methodist", State: "NJ"},
	{Slug: "new_england", Name: "New England Annual Conference", Type: "annual_conference",
		Website: "https://www.neumc.org/", Directory: umcFindAChurch,
		Denomination: "methodist", State: "MA"}, // covers CT, MA, ME, NH, RI, VT
	{Slug: "new_york", Name: "New York Annual Conference", Type: "annual_conference",
		Website: "https://www.nyac.com/", Directory: umcFindAChurch,
		Denomination: "methodist", State: "NY"},
	{Slug: "peninsula_delaware", Name: "Peninsula-Delaware Annual Conference", Type: "annual_conference",
		Website: "https://www.pen-del.org/", Directory: umcFindAChurch,
		Denomination: "methodist", State: "DE"}, // covers DE + parts of MD
	{Slug: "susquehanna", Name: "Susquehanna Annual Conference", Type: "annual_conference",
		Website: "http://www.susumc.org/", Directory: umcFindAChurch,
		Denomination: "methodist", State: "PA"},
	{Slug: "upper_new_york", Name: "Upper New York Annual Conference", Type: "annual_conference",
		Website: "http://www.unyumc.org/", Directory: umcFindAChurch,
		Denomination: "methodist", State: "NY"},
	{Slug: "west_virginia", Name: "West Virginia Annual Conference", Type: "annual_conference",
		Website: "https://www.wvumc.org/", Directory: umcFindAChurch,
		Denomination: "methodist", State: "WV"},
	{Slug: "western_pennsylvania", Name: "Western Pennsylvania Annual Conference", Type: "annual_conference",
		Website: "https://www.wpaumc.org/", Directory: umcFindAChurch,
		Denomination: "methodist", State: "PA"},

	// --- South Central Jurisdiction ---
	{Slug: "arkansas", Name: "Arkansas Annual Conference", Type: "annual_conference",
		Website: "https://arumc.org/", Directory: umcFindAChurch,
		Denomination: "methodist", State: "AR"},
	{Slug: "great_plains", Name: "Great Plains Annual Conference", Type: "annual_conference",
		Website: "https://www.greatplainsumc.org/", Directory: umcFindAChurch,
		Denomination: "methodist", State: "KS"}, // covers KS + NE
	{Slug: "horizon_texas", Name: "Horizon Texas Conference", Type: "annual_conference",
		Website: "https://www.htcumc.org/", Directory: umcFindAChurch,
		Denomination: "methodist", State: "TX"},
	{Slug: "louisiana", Name: "Louisiana Annual Conference", Type: "annual_conference",
		Website: "https://www.la-umc.org/", Directory: umcFindAChurch,
		Denomination: "methodist", State: "LA"},
	{Slug: "missouri", Name: "Missouri Annual Conference", Type: "annual_conference",
		Website: "https://www.moumethodist.org/", Directory: umcFindAChurch,
		Denomination: "methodist", State: "MO"},
	{Slug: "new_mexico", Name: "New Mexico Annual Conference", Type: "annual_conference",
		Website: "http://www.nmconfum.com/", Directory: umcFindAChurch,
		Denomination: "methodist", State: "NM"},
	{Slug: "oklahoma", Name: "Oklahoma Annual Conference", Type: "annual_conference",
		Website: "https://www.okumc.org/", Directory: umcFindAChurch,
		Denomination: "methodist", State: "OK"},
	{Slug: "oklahoma_indian_missionary", Name: "Oklahoma Indian Missionary Annual Conference", Type: "missionary_conference",
		Website: "http://www.umc-oimc.org/", Directory: umcFindAChurch,
		Denomination: "methodist", State: "OK"},
	{Slug: "rio_texas", Name: "Rio Texas Annual Conference", Type: "annual_conference",
		Website: "https://riotexas.org/", Directory: umcFindAChurch,
		Denomination: "methodist", State: "TX"},
	{Slug: "texas", Name: "Texas Annual Conference", Type: "annual_conference",
		Website: "https://www.txcumc.org/", Directory: umcFindAChurch,
		Denomination: "methodist", State: "TX"},

	// --- Southeastern Jurisdiction ---
	{Slug: "alabama_west_florida", Name: "Alabama-West Florida Annual Conference", Type: "annual_conference",
		Website: "https://www.awfumc.org/", Directory: umcFindAChurch,
		Denomination: "methodist", State: "AL"}, // covers AL + FL panhandle
	{Slug: "central_appalachian", Name: "Central Appalachian Missionary Annual Conference", Type: "missionary_conference",
		Website: "https://www.centralappalachianumc.org/", Directory: umcFindAChurch,
		Denomination: "methodist", State: "KY"}, // covers KY, TN, VA
	{Slug: "florida", Name: "Florida Annual Conference", Type: "annual_conference",
		Website: "https://www.flumc.org/", Directory: umcFindAChurch,
		Denomination: "methodist", State: "FL"},
	{Slug: "holston", Name: "Holston Annual Conference", Type: "annual_conference",
		Website: "http://www.holston.org/", Directory: umcFindAChurch,
		Denomination: "methodist", State: "TN"}, // covers TN + VA
	{Slug: "kentucky", Name: "Kentucky Annual Conference", Type: "annual_conference",
		Website: "https://www.kyumc.org/", Directory: umcFindAChurch,
		Denomination: "methodist", State: "KY"},
	{Slug: "mississippi", Name: "Mississippi Annual Conference", Type: "annual_conference",
		Website: "https://www.mississippi-umc.org/", Directory: umcFindAChurch,
		Denomination: "methodist", State: "MS"},
	{Slug: "north_alabama", Name: "North Alabama Annual Conference", Type: "annual_conference",
		Website: "https://www.umcna.org/", Directory: umcFindAChurch,
		Denomination: "methodist", State: "AL"},
	{Slug: "north_carolina", Name: "North Carolina Annual Conference", Type: "annual_conference",
		Website: "https://nccumc.org/", Directory: umcFindAChurch,
		Denomination: "methodist", State: "NC"},
	{Slug: "north_georgia", Name: "North Georgia Annual Conference", Type: "annual_conference",
		Website: "https://www.ngumc.org/", Directory: umcFindAChurch,
		Denomination: "methodist", State: "GA"},
	{Slug: "south_carolina", Name: "South Carolina Annual Conference", Type: "annual_conference",
		Website: "http://www.umcsc.org/home/", Directory: umcFindAChurch,
		Denomination: "methodist", State: "SC"},
	{Slug: "south_georgia", Name: "South Georgia Annual Conference", Type: "annual_conference",
		Website: "https://www.sgaumc.org/", Directory: umcFindAChurch,
		Denomination: "methodist", State: "GA"},
	{Slug: "tennessee_western_kentucky", Name: "Tennessee-Western Kentucky Annual Conference", Type: "annual_conference",
		Website: "https://twkumc.org/", Directory: umcFindAChurch,
		Denomination: "methodist", State: "TN"}, // covers TN + KY
	{Slug: "virginia", Name: "Virginia Annual Conference", Type: "annual_conference",
		Website: "https://vaumc.org/", Directory: umcFindAChurch,
		Denomination: "methodist", State: "VA"},
	{Slug: "western_north_carolina", Name: "Western North Carolina Annual Conference", Type: "annual_conference",
		Website: "https://wnccumc.org/", Directory: umcFindAChurch,
		Denomination: "methodist", State: "NC"},

	// --- Western Jurisdiction ---
	{Slug: "alaska", Name: "Alaska Annual Conference", Type: "missionary_conference",
		Website: "https://alaskaumc.org/", Directory: umcFindAChurch,
		Denomination: "methodist", State: "AK"},
	{Slug: "california_nevada", Name: "California-Nevada Annual Conference", Type: "annual_conference",
		Website: "https://www.cnumc.org/", Directory: umcFindAChurch,
		Denomination: "methodist", State: "CA"}, // covers CA + NV
	{Slug: "california_pacific", Name: "California-Pacific Annual Conference", Type: "annual_conference",
		Website: "https://calpacumc.org/", Directory: umcFindAChurch,
		Denomination: "methodist", State: "CA"}, // covers CA, HI, Pacific Islands
	{Slug: "desert_southwest", Name: "Desert Southwest Annual Conference", Type: "annual_conference",
		Website: "https://dscumc.org/", Directory: umcFindAChurch,
		Denomination: "methodist", State: "AZ"}, // covers AZ + parts of CA/NV
	{Slug: "mountain_sky", Name: "Mountain Sky Annual Conference", Type: "annual_conference",
		Website: "https://mtnskyumc.org/", Directory: umcFindAChurch,
		Denomination: "methodist", State: "CO"}, // covers CO, MT, WY, UT
	{Slug: "oregon_idaho", Name: "Oregon-Idaho Annual Conference", Type: "annual_conference",
		Website: "https://umoi.org/", Directory: umcFindAChurch,
		Denomination: "methodist", State: "OR"}, // covers OR + ID
	{Slug: "pacific_northwest", Name: "Pacific Northwest Annual Conference", Type: "annual_conference",
		Website: "https://www.pnwumc.org/", Directory: umcFindAChurch,
		Denomination: "methodist", State: "WA"},
	// --- Missing states (multi-state conferences + single-state) ---
	{Slug: "new_england", Name: "New England Annual Conference", Type: "annual_conference",
		Website: "https://www.neumc.org/", Directory: umcFindAChurch,
		Denomination: "methodist", State: "CT"}, // covers CT, ME, MA, NH, RI, VT
	{Slug: "new_england_me", Name: "New England Annual Conference (Maine)", Type: "annual_conference",
		Website: "https://www.neumc.org/", Directory: umcFindAChurch,
		Denomination: "methodist", State: "ME"},
	{Slug: "new_england_nh", Name: "New England Annual Conference (NH)", Type: "annual_conference",
		Website: "https://www.neumc.org/", Directory: umcFindAChurch,
		Denomination: "methodist", State: "NH"},
	{Slug: "new_england_ri", Name: "New England Annual Conference (RI)", Type: "annual_conference",
		Website: "https://www.neumc.org/", Directory: umcFindAChurch,
		Denomination: "methodist", State: "RI"},
	{Slug: "new_england_vt", Name: "New England Annual Conference (VT)", Type: "annual_conference",
		Website: "https://www.neumc.org/", Directory: umcFindAChurch,
		Denomination: "methodist", State: "VT"},
	{Slug: "greater_new_jersey", Name: "Greater New Jersey Annual Conference", Type: "annual_conference",
		Website: "https://www.gnjumc.org/", Directory: umcFindAChurch,
		Denomination: "methodist", State: "NJ"},
	{Slug: "dakotas_nd", Name: "Dakotas Annual Conference (ND)", Type: "annual_conference",
		Website: "https://www.dakotasumc.org/", Directory: umcFindAChurch,
		Denomination: "methodist", State: "ND"}, // covers ND + SD
	{Slug: "great_plains_ne", Name: "Great Plains Annual Conference (NE)", Type: "annual_conference",
		Website: "https://www.greatplainsumc.org/", Directory: umcFindAChurch,
		Denomination: "methodist", State: "NE"}, // covers KS + NE
	{Slug: "mountain_sky_mt", Name: "Mountain Sky Annual Conference (MT)", Type: "annual_conference",
		Website: "https://mtnskyumc.org/", Directory: umcFindAChurch,
		Denomination: "methodist", State: "MT"}, // covers CO, MT, WY, UT
	{Slug: "mountain_sky_wy", Name: "Mountain Sky Annual Conference (WY)", Type: "annual_conference",
		Website: "https://mtnskyumc.org/", Directory: umcFindAChurch,
		Denomination: "methodist", State: "WY"},
	{Slug: "mountain_sky_ut", Name: "Mountain Sky Annual Conference (UT)", Type: "annual_conference",
		Website: "https://mtnskyumc.org/", Directory: umcFindAChurch,
		Denomination: "methodist", State: "UT"},
	{Slug: "desert_southwest_nv", Name: "Desert Southwest Annual Conference (NV)", Type: "annual_conference",
		Website: "https://dscumc.org/", Directory: umcFindAChurch,
		Denomination: "methodist", State: "NV"}, // covers AZ + NV
	{Slug: "california_pacific_hi", Name: "California-Pacific Annual Conference (HI)", Type: "annual_conference",
		Website: "https://calpacumc.org/", Directory: umcFindAChurch,
		Denomination: "methodist", State: "HI"},
	{Slug: "baltimore_washington_dc", Name: "Baltimore-Washington Annual Conference (DC)", Type: "annual_conference",
		Website: "https://www.bwcumc.org/", Directory: umcFindAChurch,
		Denomination: "methodist", State: "DC"},
	{Slug: "oregon_idaho_id", Name: "Oregon-Idaho Annual Conference (ID)", Type: "annual_conference",
		Website: "https://umoi.org/", Directory: umcFindAChurch,
		Denomination: "methodist", State: "ID"}, // covers OR + ID
}

// methodistChurches lists prominent UMC churches in the top 5 metros.
// Website URLs verified 2026-08-01 via web search + direct fetch.
// DioceseSlug matches the Slug field in methodistConferences above.
var methodistChurches = []ParishDef{
	// New York (New York Annual Conference)
	{DioceseSlug: "new_york", Name: "John Street United Methodist Church",
		Address:     "44 John Street, New York, NY 10038",
		BulletinURL: "https://www.johnstreetchurch.org/"},
	{DioceseSlug: "new_york", Name: "Christ Church United Methodist",
		Address:     "524 Park Avenue, New York, NY 10065",
		BulletinURL: "https://christchurchnyc.org/"},
	{DioceseSlug: "new_york", Name: "Salem United Methodist Church",
		Address:     "2190 Adam Clayton Powell Jr. Blvd, New York, NY 10027",
		BulletinURL: "https://salem-harlem.org/"},

	// Los Angeles (California-Pacific Annual Conference)
	{DioceseSlug: "california_pacific", Name: "Los Angeles First United Methodist Church",
		Address:     "500 S. Hope St, Los Angeles, CA 90071",
		BulletinURL: "https://www.lafirstumc.org/"},
	{DioceseSlug: "california_pacific", Name: "Hollywood United Methodist Church",
		Address:     "6817 Franklin Ave, Los Angeles, CA 90028",
		BulletinURL: "https://hollywoodumc.org/"},
	{DioceseSlug: "california_pacific", Name: "Westwood United Methodist Church",
		Address:     "10497 Wilshire Blvd, Los Angeles, CA 90024",
		BulletinURL: "https://westwoodumc.org/"},

	// Chicago (Northern Illinois Annual Conference)
	{DioceseSlug: "northern_illinois", Name: "The Chicago Temple — First United Methodist Church",
		Address:     "77 W. Washington St, Chicago, IL 60602",
		BulletinURL: "https://chicagotemple.org/"},
	{DioceseSlug: "northern_illinois", Name: "St. Mark United Methodist Church",
		Address:     "8441 S. St. Lawrence Avenue, Chicago, IL 60619",
		BulletinURL: "https://www.stmarkumcchicago.org/"},

	// Miami (Florida Annual Conference)
	{DioceseSlug: "florida", Name: "First Church Miami",
		Address:     "398 NE 5th Street Suite 100, Miami, FL 33132",
		BulletinURL: "https://firstchurchmiami.org/"},

	// Philadelphia (Eastern Pennsylvania Annual Conference)
	{DioceseSlug: "eastern_pennsylvania", Name: "Historic St. George's United Methodist Church",
		Address:     "235 N 4th St, Philadelphia, PA 19106",
		BulletinURL: "https://www.historicstgeorges.org/"},
	{DioceseSlug: "eastern_pennsylvania", Name: "Tindley Temple United Methodist Church",
		Address:     "750-62 S. Broad St, Philadelphia, PA 19146",
		BulletinURL: "https://www.tindleytemple.net/"},
	{DioceseSlug: "eastern_pennsylvania", Name: "Arch Street United Methodist Church",
		Address:     "55 N. Broad St, Philadelphia, PA 19139",
		BulletinURL: "https://archstreetumc.org/"},
}
