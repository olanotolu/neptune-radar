package signals

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Minimum process days for a full Neptune prenup path (consult → disclose → draft → sign).
// Couples under this runway are deprioritized or suppressed for outreach.
const PrenupProcessFloorDays = 45

// Soft ideal runway: enough time for alignment + attorney work without rush.
const PrenupIdealRunwayDays = 120

// WeddingRunway is a wedding-date inference used to rank prenup prospects.
type WeddingRunway struct {
	// Date is the best-guess wedding date (UTC midnight). Nil if unknown.
	Date *time.Time `json:"date,omitempty"`
	// Raw is the matched phrase (e.g. "October 2026", "10.12.26").
	Raw string `json:"raw,omitempty"`
	// Source is caption | bio | hashtag | unknown.
	Source string `json:"source,omitempty"`
	// DaysUntil is calendar days from now to Date (negative if past).
	DaysUntil *int `json:"days_until,omitempty"`
	// Band: green (>= Ideal), amber (floor..Ideal), red (< floor), unknown.
	Band string `json:"band"`
	// Factor is a 0–1 multiplier for Neptune rank (unknown defaults to 0.75).
	Factor float64 `json:"factor"`
	// SuppressOutreach is true when runway is known and below process floor.
	SuppressOutreach bool `json:"suppress_outreach"`
	// Confidence 0–1 for the date extraction itself.
	Confidence float64 `json:"confidence"`
}

var (
	// Month name + optional day + year: "October 12, 2026", "Oct 2026", "october 12th 2026"
	reMonthDayYear = regexp.MustCompile(`(?i)\b(january|february|march|april|may|june|july|august|september|october|november|december|jan|feb|mar|apr|jun|jul|aug|sep|sept|oct|nov|dec)\.?\s+(\d{1,2})(?:st|nd|rd|th)?(?:,?\s+|\s+)(20\d{2})\b`)
	reMonthYear    = regexp.MustCompile(`(?i)\b(january|february|march|april|may|june|july|august|september|october|november|december|jan|feb|mar|apr|jun|jul|aug|sep|sept|oct|nov|dec)\.?\s+(20\d{2})\b`)
	// Numeric: 10/12/2026, 10-12-26, 10.12.2026
	reNumericDate = regexp.MustCompile(`\b(0?[1-9]|1[0-2])[./-](0?[1-9]|[12]\d|3[01])[./-]((?:20)?\d{2})\b`)
	// Relative: "next June", "this fall", "summer 2027"
	reSeasonYear = regexp.MustCompile(`(?i)\b(spring|summer|fall|autumn|winter)\s+(20\d{2})\b`)
	reNextMonth  = regexp.MustCompile(`(?i)\bnext\s+(january|february|march|april|may|june|july|august|september|october|november|december|jan|feb|mar|apr|jun|jul|aug|sep|sept|oct|nov|dec)\b`)
	// Hashtag-ish: #2026bride, wedding2026
	reYearHashtag = regexp.MustCompile(`(?i)\b(?:wedding|married|bride|groom)?#?(20\d{2})(?:bride|groom|wedding)?\b`)
	// "getting married in 2026" / "wedding in May"
	reMarriedInYear = regexp.MustCompile(`(?i)\b(?:getting\s+married|wedding|tie\s+the\s+knot|say\s+i\s+do)\s+(?:in\s+)?(20\d{2})\b`)
)

var monthNum = map[string]int{
	"january": 1, "jan": 1, "february": 2, "feb": 2, "march": 3, "mar": 3,
	"april": 4, "apr": 4, "may": 5, "june": 6, "jun": 6, "july": 7, "jul": 7,
	"august": 8, "aug": 8, "september": 9, "sep": 9, "sept": 9,
	"october": 10, "oct": 10, "november": 11, "nov": 11, "december": 12, "dec": 12,
}

var seasonMonth = map[string]int{
	"spring": 4, "summer": 7, "fall": 10, "autumn": 10, "winter": 1,
}

// ExtractWeddingRunway scans caption, bio texts, tags, and location for a wedding date.
// now is injectable for tests; pass time.Time{} to use time.Now().UTC().
func ExtractWeddingRunway(caption, bioA, bioB string, tags []string, now time.Time) WeddingRunway {
	if now.IsZero() {
		now = time.Now().UTC()
	}
	now = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	type hit struct {
		date time.Time
		raw  string
		src  string
		conf float64
	}
	var best *hit

	consider := func(d time.Time, raw, src string, conf float64) {
		// Ignore clearly past dates more than 30 days ago (old posts / throwbacks).
		if d.Before(now.AddDate(0, 0, -30)) {
			return
		}
		// Ignore absurd future (> 5 years).
		if d.After(now.AddDate(5, 0, 0)) {
			return
		}
		if best == nil || conf > best.conf || (conf == best.conf && d.Before(best.date)) {
			best = &hit{date: d, raw: raw, src: src, conf: conf}
		}
	}

	scan := func(text, src string, confBoost float64) {
		if strings.TrimSpace(text) == "" {
			return
		}
		for _, m := range reMonthDayYear.FindAllStringSubmatch(text, -1) {
			mon := monthNum[strings.ToLower(m[1])]
			day, _ := strconv.Atoi(m[2])
			year, _ := strconv.Atoi(m[3])
			if mon == 0 || day < 1 || day > 31 {
				continue
			}
			d := time.Date(year, time.Month(mon), day, 0, 0, 0, 0, time.UTC)
			consider(d, m[0], src, 0.92+confBoost)
		}
		for _, m := range reMonthYear.FindAllStringSubmatch(text, -1) {
			mon := monthNum[strings.ToLower(m[1])]
			year, _ := strconv.Atoi(m[2])
			if mon == 0 {
				continue
			}
			// Mid-month proxy when day unknown.
			d := time.Date(year, time.Month(mon), 15, 0, 0, 0, 0, time.UTC)
			consider(d, m[0], src, 0.78+confBoost)
		}
		for _, m := range reNumericDate.FindAllStringSubmatch(text, -1) {
			mon, _ := strconv.Atoi(m[1])
			day, _ := strconv.Atoi(m[2])
			year, _ := strconv.Atoi(m[3])
			if year < 100 {
				year += 2000
			}
			if mon < 1 || mon > 12 || day < 1 || day > 31 {
				continue
			}
			d := time.Date(year, time.Month(mon), day, 0, 0, 0, 0, time.UTC)
			consider(d, m[0], src, 0.88+confBoost)
		}
		for _, m := range reSeasonYear.FindAllStringSubmatch(text, -1) {
			mon := seasonMonth[strings.ToLower(m[1])]
			year, _ := strconv.Atoi(m[2])
			d := time.Date(year, time.Month(mon), 15, 0, 0, 0, 0, time.UTC)
			consider(d, m[0], src, 0.65+confBoost)
		}
		for _, m := range reNextMonth.FindAllStringSubmatch(text, -1) {
			mon := monthNum[strings.ToLower(m[1])]
			if mon == 0 {
				continue
			}
			year := now.Year()
			if int(now.Month()) >= mon {
				year++
			}
			d := time.Date(year, time.Month(mon), 15, 0, 0, 0, 0, time.UTC)
			consider(d, m[0], src, 0.7+confBoost)
		}
		for _, m := range reMarriedInYear.FindAllStringSubmatch(text, -1) {
			year, _ := strconv.Atoi(m[1])
			d := time.Date(year, 6, 15, 0, 0, 0, 0, time.UTC)
			consider(d, m[0], src, 0.55+confBoost)
		}
	}

	scan(caption, "caption", 0)
	scan(bioA, "bio", -0.05)
	scan(bioB, "bio", -0.05)
	for _, t := range tags {
		tag := strings.TrimPrefix(strings.ToLower(t), "#")
		scan(tag, "hashtag", -0.1)
		// #2026bride style
		if m := reYearHashtag.FindStringSubmatch(tag); len(m) == 2 {
			year, _ := strconv.Atoi(m[1])
			if year >= now.Year() && year <= now.Year()+5 {
				d := time.Date(year, 6, 15, 0, 0, 0, 0, time.UTC)
				consider(d, "#"+tag, "hashtag", 0.5)
			}
		}
	}

	out := WeddingRunway{Band: "unknown", Factor: 0.75, Confidence: 0}
	if best == nil {
		return out
	}
	days := int(best.date.Sub(now).Hours() / 24)
	out.Date = &best.date
	out.Raw = best.raw
	out.Source = best.src
	out.DaysUntil = &days
	out.Confidence = best.conf
	if best.conf > 1 {
		out.Confidence = 1
	}
	if best.conf < 0 {
		out.Confidence = 0
	}

	switch {
	case days < 0:
		out.Band = "past"
		out.Factor = 0.35
		out.SuppressOutreach = true
	case days < PrenupProcessFloorDays:
		out.Band = "red"
		out.Factor = 0.25
		out.SuppressOutreach = true
	case days < PrenupIdealRunwayDays:
		out.Band = "amber"
		// Linear ramp from 0.55 at floor to 0.95 at ideal
		span := float64(PrenupIdealRunwayDays - PrenupProcessFloorDays)
		t := float64(days-PrenupProcessFloorDays) / span
		out.Factor = 0.55 + 0.40*t
	default:
		out.Band = "green"
		// Slight peak around ideal, gentle decay for very far dates
		if days > 400 {
			out.Factor = 0.85
		} else {
			out.Factor = 1.0
		}
	}
	return out
}

// FormatRunwayLabel is a short operator-facing string.
func FormatRunwayLabel(r WeddingRunway) string {
	if r.DaysUntil == nil {
		return "Runway unknown"
	}
	d := *r.DaysUntil
	if d < 0 {
		return "Wedding date passed"
	}
	if d == 0 {
		return "Wedding today"
	}
	if d < 30 {
		return strconv.Itoa(d) + "d runway"
	}
	months := d / 30
	if months < 2 {
		return strconv.Itoa(d) + "d runway"
	}
	return strconv.Itoa(months) + "mo runway"
}
