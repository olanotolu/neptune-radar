package records

import (
	"fmt"
	"strings"
)

// OhioCounty maps Ohio county names to their FIPS codes and marriage record URLs.
type OhioCounty struct {
	Name        string
	FIPS        string
	MarriageURL string // county clerk marriage record search page
}

// ohioCounties maps city names to their county (most common county for multi-county cities).
var ohioCounties = map[string]OhioCounty{
	"columbus":        {"Franklin", "39049", "https://www FranklinCountyOhio.gov/recorder/marriage-search"},
	"dublin":          {"Franklin", "39049", "https://www.franklincountyo.gov/recorder/marriage-search"},
	"westerville":     {"Franklin", "39049", "https://www.franklincountyo.gov/recorder/marriage-search"},
	"hilliard":        {"Franklin", "39049", "https://www.franklincountyo.gov/recorder/marriage-search"},
	"gahanna":         {"Franklin", "39049", "https://www.franklincountyo.gov/recorder/marriage-search"},
	"reynoldsburg":    {"Franklin", "39049", "https://www.franklincountyo.gov/recorder/marriage-search"},
	"grove city":      {"Franklin", "39049", "https://www.franklincountyo.gov/recorder/marriage-search"},
	"groveport":       {"Franklin", "39049", "https://www.franklincountyo.gov/recorder/marriage-search"},
	"canal winchester": {"Franklin", "39049", "https://www.franklincountyo.gov/recorder/marriage-search"},
	"obetz":           {"Franklin", "39049", "https://www.franklincountyo.gov/recorder/marriage-search"},
	"blacklick":       {"Franklin", "39049", "https://www.franklincountyo.gov/recorder/marriage-search"},
	"new albany":      {"Franklin", "39049", "https://www.franklincountyo.gov/recorder/marriage-search"},
	"powell":          {"Delaware", "39041", "https://www.co.delaware.oh.us/recording/marriage-license"},
	"lewis center":    {"Delaware", "39041", "https://www.co.delaware.oh.us/recording/marriage-license"},
	"delaware":        {"Delaware", "39041", "https://www.co.delaware.oh.us/recording/marriage-license"},
	"marion":          {"Marion", "39101", "https://www.co.marion.oh.us/recording/marriage-license"},
	"marysville":      {"Union", "39159", "https://www.co.union.oh.us/recording/marriage-license"},
	"london":          {"Madison", "39099", "https://www.co.madison.oh.us/recording/marriage-license"},
	"sunbury":         {"Delaware", "39041", "https://www.co.delaware.oh.us/recording/marriage-license"},

	// Northeast Ohio
	"cleveland":        {"Cuyahoga", "39035", "https://www.cuyahogacounty.gov/recorder/marriage-search"},
	"cleveland heights": {"Cuyahoga", "39035", "https://www.cuyahogacounty.gov/recorder/marriage-search"},
	"shaker heights":   {"Cuyahoga", "39035", "https://www.cuyahogacounty.gov/recorder/marriage-search"},
	"parma":            {"Cuyahoga", "39035", "https://www.cuyahogacounty.gov/recorder/marriage-search"},
	"parma heights":    {"Cuyahoga", "39035", "https://www.cuyahogacounty.gov/recorder/marriage-search"},
	"strongsville":     {"Cuyahoga", "39035", "https://www.cuyahogacounty.gov/recorder/marriage-search"},
	"brunswick":        {"Medina", "39103", "https://www.co.medina.oh.us/recording/marriage-license"},
	"north royalton":   {"Cuyahoga", "39035", "https://www.cuyahogacounty.gov/recorder/marriage-search"},
	"solon":            {"Cuyahoga", "39035", "https://www.cuyahogacounty.gov/recorder/marriage-search"},
	"chagrin falls":    {"Cuyahoga", "39035", "https://www.cuyahogacounty.gov/recorder/marriage-search"},
	"mayfield heights": {"Cuyahoga", "39035", "https://www.cuyahogacounty.gov/recorder/marriage-search"},
	"lyndhurst":        {"Cuyahoga", "39035", "https://www.cuyahogacounty.gov/recorder/marriage-search"},
	"beachwood":        {"Cuyahoga", "39035", "https://www.cuyahogacounty.gov/recorder/marriage-search"},
	"university heights": {"Cuyahoga", "39035", "https://www.cuyahogacounty.gov/recorder/marriage-search"},
	"south euclid":     {"Cuyahoga", "39035", "https://www.cuyahogacounty.gov/recorder/marriage-search"},
	"cuyahoga falls":   {"Summit", "39153", "https://www.co.summit.oh.us/recording/marriage-license"},
	"akron":            {"Summit", "39153", "https://www.co.summit.oh.us/recording/marriage-license"},
	"barberton":        {"Summit", "39153", "https://www.co.summit.oh.us/recording/marriage-license"},
	"green":            {"Summit", "39153", "https://www.co.summit.oh.us/recording/marriage-license"},
	"hudson":           {"Summit", "39153", "https://www.co.summit.oh.us/recording/marriage-license"},
	"stow":             {"Summit", "39153", "https://www.co.summit.oh.us/recording/marriage-license"},
	"kent":             {"Portage", "39133", "https://www.co.portage.oh.us/recording/marriage-license"},
	"ravenna":          {"Portage", "39133", "https://www.co.portage.oh.us/recording/marriage-license"},
	"mansfield":        {"Richland", "39139", "https://www.co.richland.oh.us/recording/marriage-license"},
	"ashland":          {"Ashland", "39005", "https://www.co.ashland.oh.us/recording/marriage-license"},
	"wooster":          {"Wayne", "39169", "https://www.co.wayne.oh.us/recording/marriage-license"},
	"canton":           {"Stark", "39151", "https://www.co.stark.oh.us/recording/marriage-license"},
	"massillon":        {"Stark", "39151", "https://www.co.stark.oh.us/recording/marriage-license"},
	"north canton":     {"Stark", "39151", "https://www.co.stark.oh.us/recording/marriage-license"},

	// Northwest Ohio
	"toledo":        {"Lucas", "39095", "https://www.co.lucas.oh.us/recording/marriage-license"},
	"sylvania":      {"Lucas", "39095", "https://www.co.lucas.oh.us/recording/marriage-license"},
	"oregon":        {"Lucas", "39095", "https://www.co.lucas.oh.us/recording/marriage-license"},
	"maumee":        {"Lucas", "39095", "https://www.co.lucas.oh.us/recording/marriage-license"},
	"perrysburg":    {"Wood", "39173", "https://www.co.wood.oh.us/recording/marriage-license"},
	"bowling green": {"Wood", "39173", "https://www.co.wood.oh.us/recording/marriage-license"},
	"findlay":       {"Hancock", "39063", "https://www.co.hancock.oh.us/recording/marriage-license"},
	"lima":          {"Allen", "39003", "https://www.co.allen.oh.us/recording/marriage-license"},
	"defiance":      {"Defiance", "39039", "https://www.co.defiance.oh.us/recording/marriage-license"},
	"napoleon":      {"Henry", "39069", "https://www.co.henry.oh.us/recording/marriage-license"},
	"sandusky":      {"Erie", "39043", "https://www.co.er.oh.us/recording/marriage-license"},

	// Southwest Ohio
	"cincinnati":     {"Hamilton", "39061", "https://www.hamilton-co.org/recording/marriage-license"},
	"norwood":        {"Hamilton", "39061", "https://www.hamilton-co.org/recording/marriage-license"},
	"blue ash":       {"Hamilton", "39061", "https://www.hamilton-co.org/recording/marriage-license"},
	"madeira":        {"Hamilton", "39061", "https://www.hamilton-co.org/recording/marriage-license"},
	"mariemont":      {"Hamilton", "39061", "https://www.hamilton-co.org/recording/marriage-license"},
	"hyde park":      {"Hamilton", "39061", "https://www.hamilton-co.org/recording/marriage-license"},
	"oakley":         {"Hamilton", "39061", "https://www.hamilton-co.org/recording/marriage-license"},
	"anderson township": {"Hamilton", "39061", "https://www.hamilton-co.org/recording/marriage-license"},
	"clifton":        {"Greene", "39057", "https://www.co.greene.oh.us/recording/marriage-license"},
	"price hill":     {"Hamilton", "39061", "https://www.hamilton-co.org/recording/marriage-license"},
	"westwood":       {"Hamilton", "39061", "https://www.hamilton-co.org/recording/marriage-license"},
	"hartwell":       {"Hamilton", "39061", "https://www.hamilton-co.org/recording/marriage-license"},
	"dayton":         {"Montgomery", "39113", "https://www.mcohio.org/recording/marriage-license"},
	"kettering":      {"Montgomery", "39113", "https://www.mcohio.org/recording/marriage-license"},
	"beavercreek":    {"Greene", "39057", "https://www.co.greene.oh.us/recording/marriage-license"},
	"centerville":    {"Montgomery", "39113", "https://www.mcohio.org/recording/marriage-license"},
	"springboro":     {"Warren", "39165", "https://www.co.warren.oh.us/recording/marriage-license"},
	"miamisburg":     {"Montgomery", "39113", "https://www.mcohio.org/recording/marriage-license"},
	"vandalia":       {"Montgomery", "39113", "https://www.mcohio.org/recording/marriage-license"},
	"troy":           {"Miami", "39109", "https://www.co.miami.oh.us/recording/marriage-license"},
	"piqua":          {"Miami", "39109", "https://www.co.miami.oh.us/recording/marriage-license"},
	"sidney":         {"Shelby", "39149", "https://www.co.shelby.oh.us/recording/marriage-license"},
	"greenville":     {"Darke", "39037", "https://www.co.darke.oh.us/recording/marriage-license"},
	"eaton":          {"Preble", "39135", "https://www.co.preble.oh.us/recording/marriage-license"},
	"springfield":    {"Clark", "39023", "https://www.co.clark.oh.us/recording/marriage-license"},

	// Southeast Ohio
	"athens":           {"Athens", "39009", "https://www.co.athens.oh.us/recording/marriage-license"},
	"chillicothe":      {"Ross", "39141", "https://www.co.ross.oh.us/recording/marriage-license"},
	"portsmouth":       {"Scioto", "39145", "https://www.co.scioto.oh.us/recording/marriage-license"},
	"ironton":          {"Lawrence", "39087", "https://www.co.lawrence.oh.us/recording/marriage-license"},
	"zanesville":       {"Muskingum", "39119", "https://www.co.muskingum.oh.us/recording/marriage-license"},
	"cambridge":        {"Guernsey", "39059", "https://www.co.guernsey.oh.us/recording/marriage-license"},
	"new philadelphia": {"Tuscarawas", "39157", "https://www.co.tuscarawas.oh.us/recording/marriage-license"},
	"dover":            {"Tuscarawas", "39157", "https://www.co.tuscarawas.oh.us/recording/marriage-license"},
	"coshocton":        {"Coshocton", "39031", "https://www.co.coshocton.oh.us/recording/marriage-license"},
	"steubenville":     {"Jefferson", "39081", "https://www.co.jefferson.oh.us/recording/marriage-license"},
	"youngstown":       {"Mahoning", "39099", "https://www.co.mahoning.oh.us/recording/marriage-license"},
	"warren":           {"Trumbull", "39155", "https://www.co.trumbull.oh.us/recording/marriage-license"},
	"niles":            {"Trumbull", "39155", "https://www.co.trumbull.oh.us/recording/marriage-license"},
	"campbell":         {"Mahoning", "39099", "https://www.co.mahoning.oh.us/recording/marriage-license"},
	"struthers":        {"Mahoning", "39099", "https://www.co.mahoning.oh.us/recording/marriage-license"},

	// Other notable cities
	"fairborn":       {"Greene", "39057", "https://www.co.greene.oh.us/recording/marriage-license"},
	"mason":          {"Warren", "39165", "https://www.co.warren.oh.us/recording/marriage-license"},
	"lebanon":        {"Warren", "39165", "https://www.co.warren.oh.us/recording/marriage-license"},
	"franklin":       {"Warren", "39165", "https://www.co.warren.oh.us/recording/marriage-license"},
	"middletown":     {"Butler", "39017", "https://www.co.butler.oh.us/recording/marriage-license"},
	"hamilton":       {"Butler", "39017", "https://www.co.butler.oh.us/recording/marriage-license"},
	"oxford":         {"Butler", "39017", "https://www.co.butler.oh.us/recording/marriage-license"},
	"wilmington":     {"Clinton", "39027", "https://www.co.clinton.oh.us/recording/marriage-license"},
	"circleville":    {"Pickaway", "39129", "https://www.co.pickaway.oh.us/recording/marriage-license"},
	"logan":          {"Hocking", "39067", "https://www.co.hocking.oh.us/recording/marriage-license"},
}

// CountySearchURL returns the marriage record search URL for a given Ohio city.
func CountySearchURL(city, region string) string {
	cityKey := normalizeOhioCity(city)
	if region != "" && region != "OH" && region != "Ohio" {
		return ""
	}
	if county, ok := ohioCounties[cityKey]; ok {
		return county.MarriageURL
	}
	return ""
}

// CountyName returns the county name for an Ohio city.
func CountyName(city, region string) string {
	cityKey := normalizeOhioCity(city)
	if region != "" && region != "OH" && region != "Ohio" {
		return ""
	}
	if county, ok := ohioCounties[cityKey]; ok {
		return county.Name
	}
	return ""
}

// CountyMarriageSearchQuery returns a Google search query for marriage records
// for a couple in a specific Ohio county.
func CountyMarriageSearchQuery(firstName, lastName, city, region string) string {
	county := CountyName(city, region)
	if county == "" {
		return fmt.Sprintf(`"%s" "%s" marriage license %s %s site:.gov`,
			firstName, lastName, city, region)
	}
	return fmt.Sprintf(`"%s" "%s" marriage license %s County Ohio site:.gov`,
		firstName, lastName, county)
}

// CountyRecordLinks returns all relevant record search links for a couple.
func CountyRecordLinks(firstName, lastName, city, region string) map[string]string {
	cityKey := normalizeOhioCity(city)
	county, _ := ohioCounties[cityKey]
	links := map[string]string{}

	if county.MarriageURL != "" {
		links["county_clerk"] = county.MarriageURL
	}
	if county.Name != "" {
		links["marriage_search"] = fmt.Sprintf(
			"https://www.google.com/search?q=%s",
			strings.ReplaceAll(
				fmt.Sprintf(`"%s" "%s" marriage license %s County Ohio`,
					firstName, lastName, county.Name),
				" ", "+",
			),
		)
		links["property_search"] = fmt.Sprintf(
			"https://www.google.com/search?q=%s",
			strings.ReplaceAll(
				fmt.Sprintf(`"%s" "%s" property owner %s County Ohio`,
					firstName, lastName, county.Name),
				" ", "+",
			),
		)
		links["voter_record"] = fmt.Sprintf(
			"https://www.google.com/search?q=%s",
			strings.ReplaceAll(
				fmt.Sprintf(`"%s" "%s" voter registration %s County Ohio`,
					firstName, lastName, county.Name),
				" ", "+",
			),
		)
	}
	return links
}
