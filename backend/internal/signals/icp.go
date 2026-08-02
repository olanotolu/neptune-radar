package signals

import (
	"regexp"
	"strings"
)

// ICPFit scores how well a couple matches Meet Neptune's public ICP:
// high-earning professionals (tech equity, finance, medicine, founders)
// and priority metros where attorney coverage / SEO density is strong.
type ICPFit struct {
	// Score 0–1 combined ICP quality.
	Score float64 `json:"score"`
	// Tags are machine ids: tech_equity, physician, founder, finance, etc.
	Tags []string `json:"tags"`
	// Labels are human-readable chips for the dossier UI.
	Labels []string `json:"labels"`
	// Employers are matched company names from bios (public text only).
	Employers []string `json:"employers,omitempty"`
	// Roles are matched profession keywords.
	Roles []string `json:"roles,omitempty"`
	// MarketPriority 0–1 from city/region vs Neptune priority markets.
	MarketPriority float64 `json:"market_priority"`
	// MarketLabel e.g. "priority: nyc" or "secondary" or "unknown".
	MarketLabel string `json:"market_label"`
}

// Employer patterns aligned to meetneptune.com /lp/prenup-for-*-employees pages.
var employerPatterns = []struct {
	tag   string
	label string
	re    *regexp.Regexp
	w     float64
}{
	{"employer_google", "Google", regexp.MustCompile(`(?i)\b(google|alphabet|youtube|waymo)\b`), 0.22},
	{"employer_meta", "Meta", regexp.MustCompile(`(?i)\b(meta|facebook)\b`), 0.22},
	{"employer_openai", "OpenAI", regexp.MustCompile(`(?i)\bopenai\b`), 0.24},
	{"employer_anthropic", "Anthropic", regexp.MustCompile(`(?i)\banthropic\b`), 0.24},
	{"employer_amazon", "Amazon", regexp.MustCompile(`(?i)\b(amazon|aws)\b`), 0.18},
	{"employer_apple", "Apple", regexp.MustCompile(`(?i)\bapple\b`), 0.2},
	{"employer_microsoft", "Microsoft", regexp.MustCompile(`(?i)\b(microsoft|azure)\b`), 0.18},
	{"employer_goldman", "Goldman Sachs", regexp.MustCompile(`(?i)\bgoldman\b`), 0.24},
	{"employer_jpmorgan", "JPMorgan", regexp.MustCompile(`(?i)\b(jpmorgan|jp\s*morgan|jpm)\b`), 0.22},
	{"employer_mckinsey", "McKinsey", regexp.MustCompile(`(?i)\bmckinsey\b`), 0.22},
	{"employer_stripe", "Stripe", regexp.MustCompile(`(?i)\bstripe\b`), 0.2},
	{"employer_coinbase", "Coinbase", regexp.MustCompile(`(?i)\bcoinbase\b`), 0.2},
	{"employer_linkedin", "LinkedIn", regexp.MustCompile(`(?i)\blinkedin\b`), 0.18},
	{"employer_uber", "Uber", regexp.MustCompile(`(?i)\buber\b`), 0.16},
	{"employer_netflix", "Netflix", regexp.MustCompile(`(?i)\bnetflix\b`), 0.18},
	{"employer_disney", "Disney", regexp.MustCompile(`(?i)\bdisney\b`), 0.16},
	{"employer_stanford", "Stanford", regexp.MustCompile(`(?i)\bstanford\b`), 0.16},
	{"employer_block", "Block", regexp.MustCompile(`(?i)\b(block|square)\b.*\b(eng|engineer|product|design)`), 0.16},
}

var rolePatterns = []struct {
	tag   string
	label string
	re    *regexp.Regexp
	w     float64
}{
	{"role_swe", "Software engineer", regexp.MustCompile(`(?i)\b(software\s+engineer|swe\b|full[\s-]?stack|backend|frontend|staff\s+engineer|principal\s+engineer|ml\s+engineer|sre\b)\b`), 0.2},
	{"role_pm", "Product manager", regexp.MustCompile(`(?i)\b(product\s+manager|group\s+pm|director\s+of\s+product)\b`), 0.16},
	{"role_founder", "Founder", regexp.MustCompile(`(?i)\b(founder|co[\s-]?founder|ceo\b|startup\s+founder)\b`), 0.24},
	{"role_physician", "Physician", regexp.MustCompile(`(?i)\b(md\b|physician|surgeon|resident\s+physician|attending|dermatologist|cardiologist|anesthesiologist)\b`), 0.24},
	{"role_attorney", "Attorney", regexp.MustCompile(`(?i)\b(attorney|esq\.?|lawyer|counsel\b)\b`), 0.18},
	{"role_finance", "Finance", regexp.MustCompile(`(?i)\b(investment\s+bank|private\s+equity|hedge\s+fund|venture\s+capital|\bvc\b|portfolio\s+manager|equity\s+research|quant\b)\b`), 0.24},
	{"role_consultant", "Consultant", regexp.MustCompile(`(?i)\b(management\s+consultant|strategy\s+consultant|mbb\b)\b`), 0.18},
	{"role_pilot", "Pilot", regexp.MustCompile(`(?i)\b(airline\s+pilot|commercial\s+pilot|captain\b.*\b(air|flight)|first\s+officer)\b`), 0.16},
	{"role_nurse", "Nurse", regexp.MustCompile(`(?i)\b(rn\b|nurse\s+practitioner|\bnp\b|registered\s+nurse)\b`), 0.12},
	{"role_exec", "Executive", regexp.MustCompile(`(?i)\b(vp\b|vice\s+president|managing\s+director|general\s+counsel|cfo\b|cto\b|coo\b)\b`), 0.2},
	{"role_rsu", "Equity / RSU", regexp.MustCompile(`(?i)\b(rsu|stock\s+options|equity\s+comp|ipo\b|pre[\s-]?ipo)\b`), 0.18},
	{"role_crypto", "Crypto", regexp.MustCompile(`(?i)\b(crypto|bitcoin|ethereum|web3)\b`), 0.14},
	{"role_business_owner", "Business owner", regexp.MustCompile(`(?i)\b(business\s+owner|practice\s+owner|small\s+business\s+owner)\b`), 0.16},
}

// Priority markets mirror Neptune location SEO + dual-attorney density.
// Keys are lowercase city or region tokens found in inferred location / bios.
var priorityMarkets = map[string]float64{
	// Tier 1
	"new york": 1.0, "nyc": 1.0, "manhattan": 1.0, "brooklyn": 0.95, "ny": 0.9,
	"san francisco": 1.0, "sf": 1.0, "bay area": 0.95, "oakland": 0.85, "palo alto": 0.95,
	"los angeles": 0.95, "la": 0.9, "santa monica": 0.9,
	"boston": 0.95, "cambridge": 0.9, "ma": 0.85,
	"miami": 0.9, "fl": 0.8,
	"chicago": 0.9, "il": 0.8,
	// Tier 2
	"seattle": 0.85, "wa": 0.75,
	"austin": 0.85, "tx": 0.75, "dallas": 0.8, "houston": 0.8,
	"denver": 0.8, "co": 0.7,
	"washington": 0.85, "dc": 0.85, "arlington": 0.8,
	"atlanta": 0.8, "ga": 0.7,
	"philadelphia": 0.75, "pa": 0.7,
	"san diego": 0.8, "orange county": 0.8,
	"columbus": 0.7, "cleveland": 0.65, "cincinnati": 0.65, "oh": 0.65,
	"portland": 0.7, "nashville": 0.75, "charlotte": 0.75,
	"minneapolis": 0.7, "detroit": 0.65,
	"ca": 0.8, "california": 0.85, "new york state": 0.85,
	"massachusetts": 0.85, "florida": 0.8,
}

// ExtractICPFit classifies public bio/location text for Neptune ICP fit.
func ExtractICPFit(bioA, bioB, city, region string) ICPFit {
	text := strings.TrimSpace(bioA + "\n" + bioB)
	out := ICPFit{
		Tags:           nil,
		Labels:         nil,
		MarketPriority: 0.45,
		MarketLabel:    "unknown",
	}

	score := 0.15 // baseline — any real couple with engagement signal is non-zero
	seen := map[string]bool{}

	addTag := func(tag, label string, w float64) {
		if seen[tag] {
			return
		}
		seen[tag] = true
		out.Tags = append(out.Tags, tag)
		if label != "" {
			out.Labels = append(out.Labels, label)
		}
		score += w
	}

	for _, p := range employerPatterns {
		if p.re.MatchString(text) {
			addTag(p.tag, p.label, p.w)
			out.Employers = append(out.Employers, p.label)
		}
	}
	for _, p := range rolePatterns {
		if p.re.MatchString(text) {
			addTag(p.tag, p.label, p.w)
			out.Roles = append(out.Roles, p.label)
		}
	}

	// Market priority from city + region + bios (people put "NYC" in bio).
	locBlob := strings.ToLower(strings.TrimSpace(city + " " + region + " " + text))
	bestM := 0.45
	bestLabel := "unknown"
	for token, pri := range priorityMarkets {
		if strings.Contains(locBlob, token) {
			if pri > bestM {
				bestM = pri
				bestLabel = token
			}
		}
	}
	out.MarketPriority = bestM
	out.MarketLabel = bestLabel
	// Market contributes up to +0.25
	score += (bestM - 0.45) * 0.55
	if bestM >= 0.9 {
		addTag("market_priority", "Priority market", 0)
	} else if bestM >= 0.75 {
		addTag("market_secondary", "Secondary market", 0)
	}

	// Dual-professional boost: both bios non-empty and at least one role each side.
	if strings.TrimSpace(bioA) != "" && strings.TrimSpace(bioB) != "" {
		aHit, bHit := false, false
		for _, p := range rolePatterns {
			if p.re.MatchString(bioA) {
				aHit = true
			}
			if p.re.MatchString(bioB) {
				bHit = true
			}
		}
		for _, p := range employerPatterns {
			if p.re.MatchString(bioA) {
				aHit = true
			}
			if p.re.MatchString(bioB) {
				bHit = true
			}
		}
		if aHit && bHit {
			addTag("dual_professional", "Dual professional", 0.12)
		}
	}

	if score > 1 {
		score = 1
	}
	if score < 0 {
		score = 0
	}
	out.Score = score
	if out.Tags == nil {
		out.Tags = []string{}
	}
	if out.Labels == nil {
		out.Labels = []string{}
	}
	return out
}

// NeptuneRank multiplies engagement confidence with ICP and wedding runway.
// deliverability is 0–1 (pics present, location known, not suppressed).
// All inputs should already be clamped 0–1 where applicable.
func NeptuneRank(engagementConf, icpScore, runwayFactor, deliverability float64) float64 {
	clamp := func(x float64) float64 {
		if x < 0 {
			return 0
		}
		if x > 1 {
			return 1
		}
		return x
	}
	engagementConf = clamp(engagementConf)
	icpScore = clamp(icpScore)
	runwayFactor = clamp(runwayFactor)
	deliverability = clamp(deliverability)
	// Geometric blend so a zero runway kills rank; weak ICP still allows
	// high-engagement prospects through but deprioritizes them.
	r := engagementConf * (0.35 + 0.65*icpScore) * runwayFactor * (0.5 + 0.5*deliverability)
	return clamp(r)
}
