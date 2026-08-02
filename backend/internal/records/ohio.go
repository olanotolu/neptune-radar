package records

import "strings"

// Ohio city → zip code mappings for smarter heuristic candidates.
// Sourced from USPS data. Only cities with population > 5,000 included.
var ohioCityZips = map[string][]string{
	// Central Ohio
	"columbus":      {"43085", "43201", "43202", "43203", "43204", "43205", "43206", "43207", "43209", "43210", "43211", "43212", "43213", "43214", "43215", "43219", "43220", "43221", "43222", "43224", "43227", "43228", "43229", "43230", "43231", "43232", "43234", "43235", "43236"},
	"dublin":        {"43016", "43017"},
	"westerville":   {"43081", "43082"},
	"hilliard":      {"43026"},
	"gahanna":       {"43230"},
	"reynoldsburg":  {"43068", "43069"},
	"grove city":    {"43123"},
	"groveport":     {"43125"},
	"canal winchester": {"43110"},
	"obetz":         {"43137"},
	"blacklick":     {"43004"},
	"new albany":    {"43054"},
	"powell":        {"43065"},
	"lewis center":  {"43035"},
	"delaware":      {"43015"},
	"marion":        {"43302"},
	"marysville":    {"43040"},
	"london":        {"43140"},
	"sunbury":       {"43074"},

	// Northeast Ohio
	"cleveland":     {"44101", "44102", "44103", "44104", "44105", "44106", "44107", "44108", "44109", "44110", "44111", "44112", "44113", "44114", "44115", "44118", "44119", "44120", "44121", "44122", "44124", "44125", "44126", "44127", "44128", "44129", "44130", "44131", "44134", "44135", "44137", "44138", "44139", "44140", "44141", "44143", "44144", "44145", "44146"},
	"cleveland heights": {"44118"},
	"shaker heights": {"44120", "44122"},
	"parma":         {"44129", "44130", "44134"},
	"parma heights": {"44130"},
	"strongsville":  {"44136"},
	"brunswick":     {"44212"},
	"north royalton": {"44133"},
	"solon":         {"44139"},
	"chagrin falls": {"44022"},
	"mayfield heights": {"44124"},
	"lyndhurst":     {"44124"},
	"beachwood":     {"44122"},
	"university heights": {"44118"},
	" south euclid": {"44121"},
	"cuyahoga falls": {"44221", "44223"},
	"akron":         {"44301", "44302", "44303", "44304", "44305", "44306", "44307", "44308", "44310", "44311", "44312", "44313", "44314", "44319", "44320", "44321"},
	"barberton":     {"44203"},
	"green":         {"44232"},
	"hudson":        {"44236"},
	"stow":          {"44224"},
	"kent":          {"44240"},
	"ravenna":       {"44266"},
	"mansfield":     {"44901", "44902", "44903", "44904", "44905", "44906", "44907"},
	"ashland":       {"44805"},
	"wooster":       {"44691"},
	"canton":        {"44701", "44702", "44703", "44704", "44705", "44706", "44707", "44708", "44709", "44710"},
	"massillon":     {"44646", "44647"},
	"north canton":  {"44720"},

	// Northwest Ohio
	"toledo":        {"43601", "43602", "43603", "43604", "43605", "43606", "43607", "43608", "43609", "43610", "43611", "43612", "43613", "43614", "43615", "43616", "43617", "43619", "43620"},
	"sylvania":      {"43560"},
	"oregon":        {"43616"},
	"maumee":        {"43537"},
	"perrysburg":    {"43551"},
	"bowling green": {"43402"},
	"findlay":       {"45839", "45840"},
	"lima":          {"45801", "45802", "45804", "45805"},
	"defiance":      {"43512"},
	"napoleon":      {"43545"},
	"wauson":        {"43567"},
	"port clinton":  {"43452"},
	"sandusky":      {"44870", "44871"},

	// Southwest Ohio
	"cincinnati":    {"45201", "45202", "45203", "45204", "45205", "45206", "45207", "45208", "45209", "45210", "45211", "45212", "45213", "45214", "45215", "45216", "45217", "45218", "45219", "45220", "45221", "45222", "45223", "45224", "45225", "45226", "45227", "45228", "45229", "45230", "45231", "45232", "45233", "45234", "45235", "45236", "45237", "45238", "45239", "45240", "45241", "45242", "45243", "45244", "45245", "45246", "45247", "45248", "45249", "45250", "45251", "45252", "45255"},
	"norwood":       {"45212"},
	"blue ash":      {"45242"},
	"madeira":       {"45243"},
	"mariemont":     {"45227"},
	"hyde park":     {"45208"},
	"oakley":        {"45209"},
	"anderson township": {"45230", "45244"},
	"clifton":       {"45220"},
	"price hill":    {"45205"},
	"westwood":      {"45211"},
	"hartwell":      {"45215"},
	"dayton":        {"45401", "45402", "45403", "45404", "45405", "45406", "45409", "45410", "45412", "45414", "45415", "45416", "45417", "45419", "45420", "45424", "45426", "45427", "45429", "45430", "45431", "45432", "45433", "45434", "45437", "45440", "45449"},
	"kettering":     {"45419", "45420", "45429"},
	"beavercreek":   {"45430", "45431", "45432"},
	"centerville":   {"45459"},
	"springboro":    {"45066"},
	"miamisburg":    {"45342"},
	"vandalia":      {"45377"},
	"troy":          {"45373"},
	"piqua":         {"45356"},
	"sidney":        {"45365"},
	"greenville":    {"45331"},
	"eaton":         {"45320"},
	"springfield":   {"45501", "45502", "45503", "45504", "45505", "45506"},

	// Southeast Ohio
	"athens":        {"45701"},
	"chillicothe":   {"45601"},
	"portsmouth":    {"45662"},
	"ironton":       {"45638"},
	"zanesville":    {"43701"},
	"cambridge":     {"43725"},
	"new philadelphia": {"44663"},
	"dover":         {"44622"},
	"coshocton":     {"43812"},
	"martins ferry": {"43935"},
	"st. clairsville": {"43950"},
	"steubenville":  {"43952"},
	"youngstown":    {"44501", "44502", "44503", "44504", "44505", "44506", "44507", "44509", "44510", "44511", "44512", "44513", "44514", "44515"},
	"warren":        {"44481", "44482", "44483", "44484", "44485"},
	"niles":         {"44446"},
	"campbell":      {"44405"},
	"struthers":     {"44471"},
	"new castle":    {"44460"},
	"sharon":        {"16146"},
	"hermitage":     {"16148"},

	// Other notable cities
	"fairborn":      {"45324"},
	"trotwood":      {"45426"},
	"huber heights": {"45424"},
	"mason":         {"45040"},
	"lebanon":       {"45036"},
	"franklin":      {"45005"},
	"middletown":    {"45042", "45044"},
	"hamilton":      {"45011", "45012", "45013"},
	"oxford":        {"45056"},
	"morrow":        {"45152"},
	"wilmington":    {"45177"},
	"washington ch": {"43160"},
	"circleville":   {"43113"},
	"logan":         {"43138"},
	"nelsonville":   {"45764"},
	"pomeroy":       {"45769"},
	"gallipolis":    {"45631"},
	"rio grande":    {"45674"},
}

// ohioNeighborhoods maps major cities to their known neighborhoods/districts.
// Used to generate more specific address candidates.
var ohioNeighborhoods = map[string][]struct {
	Name    string
	Zip     string
	ConfMod float64 // adjustment to confidence
}{
	"columbus": {
		{"German Village", "43206", 0.05},
		{"Short North", "43215", 0.05},
		{"Victorian Village", "43201", 0.03},
		{"Clintonville", "43202", 0.03},
		{"Bexley", "43209", 0.05},
		{"Grandview", "43212", 0.03},
		{"Dublin", "43017", 0.05},
		{"Westerville", "43082", 0.05},
		{"Upper Arlington", "43221", 0.05},
		{"Worthington", "43085", 0.05},
		{"Hilliard", "43026", 0.04},
		{"Gahanna", "43230", 0.04},
		{"Reynoldsburg", "43068", 0.04},
		{"Grove City", "43123", 0.04},
		{"Pickerington", "43147", 0.04},
		{"Canal Winchester", "43110", 0.04},
		{"New Albany", "43054", 0.05},
		{"Powell", "43065", 0.05},
		{"Lewis Center", "43035", 0.04},
		{"Blacklick", "43004", 0.03},
	},
	"cleveland": {
		{"Ohio City", "44113", 0.03},
		{"Tremont", "44113", 0.03},
		{"University Circle", "44106", 0.03},
		{"Coventry Village", "44118", 0.03},
		{"Shaker Heights", "44120", 0.05},
		{"Beachwood", "44122", 0.05},
		{"Lyndhurst", "44124", 0.04},
		{"South Euclid", "44121", 0.04},
		{"Parma", "44129", 0.04},
		{"Strongsville", "44136", 0.04},
		{"Solon", "44139", 0.05},
		{"Chagrin Falls", "44022", 0.05},
		{"Broadview Heights", "44147", 0.04},
		{"North Royalton", "44133", 0.04},
	},
	"cincinnati": {
		{"Hyde Park", "45208", 0.05},
		{"Oakley", "45209", 0.03},
		{"Mt. Lookout", "45208", 0.04},
		{"Indian Hill", "45243", 0.06},
		{"Madeira", "45243", 0.05},
		{"Blue Ash", "45242", 0.05},
		{"Mason", "45040", 0.05},
		{"West Chester", "45069", 0.05},
		{"Anderson Township", "45230", 0.04},
		{"Mt. Washington", "45230", 0.03},
		{"Northside", "45223", 0.03},
		{"Clifton", "45220", 0.03},
		{"Gaslight District", "45219", 0.04},
		{"Pleasant Ridge", "45213", 0.03},
		{"Norwood", "45212", 0.03},
	},
	"dayton": {
		{"Oakwood", "45419", 0.05},
		{"Kettering", "45419", 0.05},
		{"Centerville", "45459", 0.05},
		{"Beavercreek", "45431", 0.05},
		{"Springboro", "45066", 0.05},
		{"Huber Heights", "45424", 0.04},
		{"Vandalia", "45377", 0.04},
		{"Trotwood", "45426", 0.03},
		{"West Carrollton", "45449", 0.03},
		{"Moraine", "45439", 0.03},
	},
	"akron": {
		{"Highland Square", "44303", 0.03},
		{"Firestone Park", "44305", 0.03},
		{"Fairlawn", "44333", 0.05},
		{"Copley", "44321", 0.04},
		{"Green", "44232", 0.05},
		{"Hudson", "44236", 0.05},
		{"Stow", "44224", 0.04},
		{"Cuyahoga Falls", "44221", 0.04},
		{"Tallmadge", "44278", 0.04},
		{"Munroe Falls", "44262", 0.04},
	},
	"toledo": {
		{"West Toledo", "43615", 0.03},
		{"Ottawa Hills", "43606", 0.05},
		{"Sylvania", "43560", 0.05},
		{"Maumee", "43537", 0.05},
		{"Perrysburg", "43551", 0.05},
		{"Oregon", "43616", 0.03},
		{"Rossford", "43460", 0.04},
		{"Whitehouse", "43571", 0.04},
		{"Waterville", "43566", 0.04},
	},
	"youngstown": {
		{"Boardman", "44512", 0.04},
		{"Canfield", "44406", 0.05},
		{"Poland", "44514", 0.05},
		{"Austintown", "44515", 0.03},
		{"Liberty", "44505", 0.03},
		{"Campbell", "44405", 0.03},
		{"Struthers", "44471", 0.03},
	},
	"springfield": {
		{"Enon", "45323", 0.04},
		{"Yellow Springs", "45387", 0.05},
		{"Fairborn", "45324", 0.04},
		{"New Carlisle", "45344", 0.03},
		{"North Hampton", "45349", 0.03},
	},
}

// ohioZipCities is the reverse mapping: zip → city name(s).
var ohioZipCities = map[string][]string{
	"43085": {"westerville"},
	"43017": {"dublin"},
	"43026": {"hilliard"},
	"43054": {"new albany"},
	"43065": {"powell"},
	"43068": {"reynoldsburg"},
	"43110": {"canal winchester"},
	"43123": {"grove city"},
	"44118": {"cleveland heights"},
	"44120": {"shaker heights"},
	"44122": {"beachwood", "shaker heights"},
	"44124": {"lyndhurst", "mayfield heights"},
	"44129": {"parma"},
	"44136": {"strongsville"},
	"44139": {"solon"},
	"44221": {"cuyahoga falls"},
	"44232": {"green"},
	"44236": {"hudson"},
	"45208": {"cincinnati"},
	"45242": {"blue ash"},
	"45243": {"madeira", "indian hill"},
	"45419": {"kettering", "oakwood"},
	"45431": {"beavercreek"},
	"45459": {"centerville"},
	"43560": {"sylvania"},
	"43537": {"maumee"},
	"43551": {"perrysburg"},
}

// getOhioZipCandidates returns zip+city candidates for an Ohio city lookup.
func getOhioZipCandidates(city, region string) []Candidate {
	if region != "" && region != "OH" && region != "Ohio" {
		return nil
	}
	cityKey := normalizeOhioCity(city)
	zips, ok := ohioCityZips[cityKey]
	if !ok {
		return nil
	}
	var cands []Candidate
	// Return top 3 zips with base confidence
	for i, zip := range zips {
		if i >= 3 {
			break
		}
		conf := 0.50
		if i == 0 {
			conf = 0.55 // primary zip gets higher confidence
		}
		cands = append(cands, Candidate{
			City:       city,
			Region:     "OH",
			Postal:     zip,
			Country:    "US",
			Confidence: conf,
			Source:     "ohio_zipdb",
			Note:       "Ohio zip code database — most common zip for " + city + ". Verify with Lob.",
		})
	}
	return cands
}

// getOhioNeighborhoodCandidates returns neighborhood-level candidates for a city.
func getOhioNeighborhoodCandidates(city, region string) []Candidate {
	if region != "" && region != "OH" && region != "Ohio" {
		return nil
	}
	cityKey := normalizeOhioCity(city)
	hoods, ok := ohioNeighborhoods[cityKey]
	if !ok {
		return nil
	}
	var cands []Candidate
	for _, h := range hoods {
		cands = append(cands, Candidate{
			City:       city,
			Region:     "OH",
			Postal:     h.Zip,
			Country:    "US",
			Confidence: 0.45 + h.ConfMod,
			Source:     "ohio_neighborhoods",
			Note:       h.Name + " neighborhood, " + city + " — likely residential area. Verify with Lob.",
		})
	}
	return cands
}

func normalizeOhioCity(city string) string {
	city = strings.ToLower(strings.TrimSpace(city))
	// Common aliases
	switch city {
	case "cbus", "cow town":
		return "columbus"
	case "cincy":
		return "cincinnati"
	case "c-inci":
		return "cincinnati"
	case "the land":
		return "cleveland"
	case "the 330":
		return "akron"
	case "bow town":
		return "bowling green"
	case "b-g":
		return "bowling green"
	}
	return city
}

// parsePostLocationCity extracts a city from an Instagram venue tag like
// "The Joseph Hotel, Columbus OH" or "Schumacher Place, Columbus, Ohio".
func parsePostLocationCity(loc string) (city, region string, ok bool) {
	loc = strings.TrimSpace(loc)
	if loc == "" {
		return "", "", false
	}
	// Try "Venue Name, City, ST" or "Venue Name, City ST" patterns
	lower := strings.ToLower(loc)
	// Known Ohio cities in venue tags
	ohioCities := []string{
		"columbus", "cleveland", "cincinnati", "dayton", "akron", "toledo",
		"youngstown", "canton", "springfield", "mansfield", "findlay",
		"delaware", "dublin", "westerville", "hilliard", "gahanna",
		"worthington", "bexley", "grove city", "pickerington",
		"new albany", "powell", "marion", "chillicothe", "zanesville",
		"lima", "sandusky", "bowling green", "perrysburg", "maumee",
		"norwood", "blue ash", "madeira", "kettering", "beavercreek",
		"centerville", "mason", "west chester", "hamilton", "middletown",
	}
	for _, c := range ohioCities {
		if strings.Contains(lower, c) {
			// Try to extract state after the city
			idx := strings.Index(lower, c)
			rest := loc[idx+len(c):]
			rest = strings.TrimLeft(rest, " ,")
			if len(rest) >= 2 {
				if rest[:2] == "OH" || strings.HasPrefix(strings.ToLower(rest), "ohio") {
					return c, "OH", true
				}
			}
			return c, "OH", true
		}
	}
	// Generic "City, ST" pattern
	re := strings.LastIndex(loc, ",")
	if re > 0 {
		after := strings.TrimSpace(loc[re+1:])
		parts := strings.Fields(after)
		if len(parts) >= 2 {
			st := parts[len(parts)-1]
			if len(st) == 2 && strings.ToUpper(st) == st {
				cityGuess := strings.TrimSpace(strings.Join(parts[:len(parts)-1], " "))
				if cityGuess != "" {
					return cityGuess, st, true
				}
			}
		}
	}
	return "", "", false
}
