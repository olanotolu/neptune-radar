package signals

import (
	"regexp"
	"strings"
	"unicode"
)

// ResolvedName is a human-facing name with provenance for research + postcards.
type ResolvedName struct {
	First  string // preferred first name for "Dear Alida"
	Last   string // surname when known — critical for people-search / address
	Full   string // "Alida" or "Alida Smith" when known
	Source string // caption | display_name | bio | handle | unknown
}

// Full-name couple patterns: "Alida Smith and Andrew Jones"
var coupleFullNamePatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\b(?:between|of|featuring|introducing|capturing|photographing|congratulations?\s+to|congrats(?:\s+to)?|with|for)\s+([A-Z][a-z]{1,18})\s+([A-Z][a-z]{1,20})\s+(?:and|&)\s+([A-Z][a-z]{1,18})\s+([A-Z][a-z]{1,20})\b`),
	regexp.MustCompile(`(?i)\b([A-Z][a-z]{1,18})\s+([A-Z][a-z]{1,20})\s+(?:and|&)\s+([A-Z][a-z]{1,18})\s+([A-Z][a-z]{1,20})(?:'s|\s+got\s+|\s+engagement|\s+wedding|\s+proposal)`),
}

// Patterns ordered by confidence. Captions from wedding photographers almost
// always name the couple in prose — that beats Instagram handles every time.
var coupleNamePatterns = []*regexp.Regexp{
	// "between Alida and Andrew" / "moments of Alissa and Jon"
	regexp.MustCompile(`(?i)\b(?:between|of|featuring|introducing|celebrating|announcing)\s+([A-Z][a-z]{1,20})\s+(?:and|&)\s+([A-Z][a-z]{1,20})\b`),
	// "congratulations to Alida and Andrew" / "congrats Alida & Andrew"
	regexp.MustCompile(`(?i)\b(?:congratulations?\s+to|congrats(?:\s+to)?|so\s+happy\s+for)\s+([A-Z][a-z]{1,20})\s+(?:and|&)\s+([A-Z][a-z]{1,20})\b`),
	// "capturing Alida and Andrew" / "with Alida and Andrew" / "for Alida and Andrew"
	regexp.MustCompile(`(?i)\b(?:capturing|photographing|documenting|celebrating|to|for|with)\s+([A-Z][a-z]{1,20})\s+(?:and|&)\s+([A-Z][a-z]{1,20})\b`),
	// "Alida and Andrew's engagement" / "Alida & Andrew got engaged"
	regexp.MustCompile(`(?i)\b([A-Z][a-z]{1,20})\s+(?:and|&)\s+([A-Z][a-z]{1,20})(?:'s|\s+got\s+|\s+are\s+|\s+said\s+|\s+during\s+|\s+on\s+their\s+|\s+engagement|\s+wedding|\s+proposal)`),
	// Bare "Alida and Andrew" near engagement language (last, lower confidence)
	regexp.MustCompile(`(?i)\b([A-Z][a-z]{1,20})\s+(?:and|&)\s+([A-Z][a-z]{1,20})\b`),
}

// Words that look like names but aren't people.
var notAPersonName = map[string]bool{
	"The": true, "This": true, "That": true, "Their": true, "These": true,
	"New": true, "York": true, "Los": true, "Angeles": true, "San": true,
	"Ohio": true, "City": true, "Park": true, "Farm": true, "House": true,
	"Studio": true, "Photo": true, "Wedding": true, "Bride": true, "Groom": true,
	"Happy": true, "Love": true, "Light": true, "Natural": true, "Bright": true,
	"Airy": true, "Sweet": true, "Intimate": true, "Moments": true, "Shared": true,
	"During": true, "Session": true, "Engagement": true, "Proposal": true,
	"Congratulations": true, "Congrats": true, "Mr": true, "Mrs": true, "Ms": true,
	"Dr": true, "And": true, "With": true, "From": true, "Your": true, "Our": true,
	"Best": true, "Day": true, "Night": true, "Morning": true, "Weekend": true,
	"Central": true, "North": true, "South": true, "East": true, "West": true,
	"Jersey": true, "Shore": true, "Picnic": true, "Co": true,
}

// ExtractCoupleFullNamesFromCaption finds "Alida Smith and Andrew Jones" pairs.
func ExtractCoupleFullNamesFromCaption(caption string) (aFirst, aLast, bFirst, bLast string, ok bool) {
	caption = strings.TrimSpace(caption)
	if caption == "" {
		return "", "", "", "", false
	}
	for _, re := range coupleFullNamePatterns {
		all := re.FindAllStringSubmatch(caption, -1)
		for _, m := range all {
			if len(m) < 5 {
				continue
			}
			af, al := titleCaseName(m[1]), titleCaseName(m[2])
			bf, bl := titleCaseName(m[3]), titleCaseName(m[4])
			if !isPlausibleFirstName(af) || !isPlausibleLastName(al) {
				continue
			}
			if !isPlausibleFirstName(bf) || !isPlausibleLastName(bl) {
				continue
			}
			if af == bf && al == bl {
				continue
			}
			return af, al, bf, bl, true
		}
	}
	return "", "", "", "", false
}

// ExtractCoupleNamesFromCaption finds first-name pairs in photographer captions.
// Prefer higher-confidence patterns; reject non-person tokens.
func ExtractCoupleNamesFromCaption(caption string) (a, b string, ok bool) {
	caption = strings.TrimSpace(caption)
	if caption == "" {
		return "", "", false
	}
	// Prefer full names when present
	if af, _, bf, _, ok := ExtractCoupleFullNamesFromCaption(caption); ok {
		return af, bf, true
	}
	// Walk patterns in order; first valid hit wins.
	for i, re := range coupleNamePatterns {
		// Last pattern (bare "X and Y") only when engagement-ish language present.
		if i == len(coupleNamePatterns)-1 && !hasEngagementContext(caption) {
			continue
		}
		all := re.FindAllStringSubmatch(caption, -1)
		for _, m := range all {
			if len(m) < 3 {
				continue
			}
			x, y := titleCaseName(m[1]), titleCaseName(m[2])
			if !isPlausibleFirstName(x) || !isPlausibleFirstName(y) {
				continue
			}
			if x == y {
				continue
			}
			return x, y, true
		}
	}
	return "", "", false
}

func isPlausibleLastName(s string) bool {
	if !isPlausibleFirstName(s) {
		return false
	}
	// Last names are usually not the engagement-filler words
	switch strings.ToLower(s) {
	case "and", "the", "engagement", "wedding", "session", "moments":
		return false
	}
	return true
}

func hasEngagementContext(caption string) bool {
	c := strings.ToLower(caption)
	for _, k := range []string{
		"engag", "propos", "fiancé", "fiance", "wedding", "bride", "groom",
		"said yes", "she said", "he said", "ring", "just engaged", "got engaged",
	} {
		if strings.Contains(c, k) {
			return true
		}
	}
	return false
}

func isPlausibleFirstName(s string) bool {
	s = strings.TrimSpace(s)
	if len(s) < 2 || len(s) > 20 {
		return false
	}
	if notAPersonName[s] || notAPersonName[strings.ToLower(s)] {
		// map is mixed case; also check lower
		for k := range notAPersonName {
			if strings.EqualFold(k, s) {
				return false
			}
		}
	}
	// Must be letters only (allow hyphen / apostrophe inside)
	letters := 0
	for _, r := range s {
		if unicode.IsLetter(r) {
			letters++
			continue
		}
		if r == '-' || r == '\'' {
			continue
		}
		return false
	}
	return letters >= 2
}

func titleCaseName(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}
	// Preserve Mc/Mac-ish simple title case
	runes := []rune(strings.ToLower(s))
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

// ParseDisplayName returns first (+ last when present) from IG full_name.
// "Alida Smith | NYC" → Alida, Smith. "Alida" → Alida, "".
func ParseDisplayName(display string) (first, last string, ok bool) {
	display = strings.TrimSpace(display)
	if display == "" {
		return "", "", false
	}
	// Strip emoji / trailing role junk after | or –
	for _, sep := range []string{"|", "–", "—", "•", "·"} {
		if i := strings.Index(display, sep); i > 0 {
			left := strings.TrimSpace(display[:i])
			if left != "" {
				display = left
			}
			break
		}
	}
	// "Name - Photographer" only strip hyphen when second half is role-ish
	if i := strings.Index(display, " - "); i > 0 {
		right := strings.ToLower(display[i+3:])
		if strings.Contains(right, "photo") || strings.Contains(right, "wedding") ||
			strings.Contains(right, "planner") || strings.Contains(right, "official") {
			display = strings.TrimSpace(display[:i])
		}
	}
	display = strings.TrimPrefix(display, "@")
	if strings.Contains(display, "_") && !strings.Contains(display, " ") {
		return "", "", false // still a handle
	}
	if LooksLikeBusinessHandle(strings.ReplaceAll(strings.ToLower(display), " ", "")) {
		return "", "", false
	}
	parts := strings.Fields(display)
	if len(parts) == 0 {
		return "", "", false
	}
	first = titleCaseName(stripNonName(parts[0]))
	if !isPlausibleFirstName(first) {
		return "", "", false
	}
	if len(parts) >= 2 {
		// Skip middle initials "A." 
		cand := parts[len(parts)-1]
		if len(cand) == 2 && strings.HasSuffix(cand, ".") {
			if len(parts) >= 3 {
				cand = parts[len(parts)-2]
			} else {
				cand = ""
			}
		}
		last = titleCaseName(stripNonName(cand))
		if !isPlausibleLastName(last) || strings.EqualFold(last, first) {
			last = ""
		}
	}
	return first, last, true
}

// FirstNameFromDisplay takes an Instagram full_name / display_name and returns
// a clean first name when it looks human ("Alida Smith", "Alida | NYC").
func FirstNameFromDisplay(display string) (string, bool) {
	f, _, ok := ParseDisplayName(display)
	return f, ok
}

func stripNonName(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || r == '-' || r == '\'' {
			return r
		}
		return -1
	}, s)
}

// FirstNameFromBio looks for "Name | city" or leading Name patterns in bios.
func FirstNameFromBio(bio string) (string, bool) {
	bio = strings.TrimSpace(bio)
	if bio == "" {
		return "", false
	}
	// First line often "Alida 💍" or "Alida | Columbus"
	line := bio
	if i := strings.IndexAny(bio, "\n\r"); i > 0 {
		line = bio[:i]
	}
	if n, ok := FirstNameFromDisplay(line); ok {
		return n, true
	}
	// "I'm Alida" / "Hi, I'm Andrew"
	re := regexp.MustCompile(`(?i)\b(?:i'?m|i am|hi[, ]+i'?m)\s+([A-Z][a-z]{1,20})\b`)
	if m := re.FindStringSubmatch(bio); len(m) >= 2 {
		n := titleCaseName(m[1])
		if isPlausibleFirstName(n) {
			return n, true
		}
	}
	return "", false
}

// FirstNameFromHandle is last resort: ale.alejandra92 → Ale, mmccrohan11 → Mmccrohan (weak).
// Only used when nothing better exists; confidence is low.
func FirstNameFromHandle(handle string) (string, bool) {
	h := strings.ToLower(strings.TrimPrefix(strings.TrimSpace(handle), "@"))
	if h == "" || LooksLikeBusinessHandle(h) {
		return "", false
	}
	// strip trailing digits
	for len(h) > 0 {
		last := h[len(h)-1]
		if last >= '0' && last <= '9' {
			h = h[:len(h)-1]
			continue
		}
		break
	}
	// split on _ . -
	parts := strings.FieldsFunc(h, func(r rune) bool {
		return r == '_' || r == '.' || r == '-'
	})
	if len(parts) == 0 {
		return "", false
	}
	// Prefer the more name-like token (longer alpha, not "the", "official")
	best := parts[0]
	for _, p := range parts {
		if len(p) >= 3 && len(p) > len(best) && !notAPersonName[titleCaseName(p)] {
			// prefer first token still if it's long enough
			if len(parts[0]) >= 3 {
				best = parts[0]
				break
			}
			best = p
		}
	}
	n := titleCaseName(best)
	if !isPlausibleFirstName(n) {
		return "", false
	}
	return n, true
}

// ResolvePersonName stacks display → bio → handle (caption handled at couple level).
func ResolvePersonName(display, bio, handle string) ResolvedName {
	// 1) Instagram display name (may include last)
	if f, l, ok := ParseDisplayName(display); ok {
		full := f
		if l != "" {
			full = f + " " + l
		}
		return ResolvedName{First: f, Last: l, Full: full, Source: "display_name"}
	}
	// 2) Bio
	if n, ok := FirstNameFromBio(bio); ok {
		return ResolvedName{First: n, Full: n, Source: "bio"}
	}
	// 3) Handle — weak, first only
	if n, ok := FirstNameFromHandle(handle); ok {
		return ResolvedName{First: n, Full: n, Source: "handle"}
	}
	return ResolvedName{Source: "unknown"}
}

// ResolveCoupleFirstNames returns best first (+ last when known) for both partners.
// Priority: caption full names → caption first names + display last → display/bio/handle.
func ResolveCoupleFirstNames(
	caption string,
	displayA, displayB string,
	bioA, bioB string,
	handleA, handleB string,
) (a, b ResolvedName) {
	// 1) Full names in caption
	if af, al, bf, bl, ok := ExtractCoupleFullNamesFromCaption(caption); ok {
		a = ResolvedName{First: af, Last: al, Full: af + " " + al, Source: "caption"}
		b = ResolvedName{First: bf, Last: bl, Full: bf + " " + bl, Source: "caption"}
		return alignCoupleToHandles(a, b, handleA, handleB)
	}

	// 2) First names from caption + last from display when first matches
	if ca, cb, ok := ExtractCoupleNamesFromCaption(caption); ok {
		a = ResolvedName{First: ca, Full: ca, Source: "caption"}
		b = ResolvedName{First: cb, Full: cb, Source: "caption"}
		a, b = alignCoupleToHandles(a, b, handleA, handleB)
		// Merge last names from display/bio if first name matches
		if f, l, ok := ParseDisplayName(displayA); ok {
			if strings.EqualFold(f, a.First) && l != "" {
				a.Last, a.Full = l, a.First+" "+l
			} else if strings.EqualFold(f, b.First) && l != "" {
				b.Last, b.Full = l, b.First+" "+l
			}
		}
		if f, l, ok := ParseDisplayName(displayB); ok {
			if strings.EqualFold(f, b.First) && l != "" && b.Last == "" {
				b.Last, b.Full = l, b.First+" "+l
			} else if strings.EqualFold(f, a.First) && l != "" && a.Last == "" {
				a.Last, a.Full = l, a.First+" "+l
			}
		}
		return a, b
	}

	// 3) Per-person resolution
	a = ResolvePersonName(displayA, bioA, handleA)
	b = ResolvePersonName(displayB, bioB, handleB)
	return a, b
}

func alignCoupleToHandles(a, b ResolvedName, handleA, handleB string) (ResolvedName, ResolvedName) {
	ha, hb := strings.ToLower(handleA), strings.ToLower(handleB)
	ca, cb := strings.ToLower(a.First), strings.ToLower(b.First)
	if (strings.Contains(ha, cb) && !strings.Contains(ha, ca)) ||
		(strings.Contains(hb, ca) && !strings.Contains(hb, cb)) {
		return b, a
	}
	return a, b
}

// HasLastName reports whether detective can do a strong people-search.
func (r ResolvedName) HasLastName() bool {
	return strings.TrimSpace(r.Last) != ""
}
