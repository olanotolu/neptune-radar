// Package records finds mailing-address candidates from name + market signals.
// Implementations: Trestle IQ (primary), People Data Labs, Cleanlist, deterministic heuristic fallback.
package records

import (
	"context"
	"os"
	"strings"
)

// Query is a people-search request built from a congratulate kit dossier.
type Query struct {
	FirstName string
	LastName  string
	// Optional second person (partner) — improves household ranking.
	PartnerFirst string
	PartnerLast  string
	City         string
	Region       string // state abbrev
	// Social handles for providers that accept them (PDL).
	Handle string

	// --- Enriched location signals (used by heuristic to boost confidence) ---
	PostLocation    string // Instagram venue tag, e.g. "The Joseph Hotel, Columbus OH"
	VendorCity      string // photographer/vendor city from watched_sources
	VendorState     string // photographer/vendor state from watched_sources
	AccountCityA    string // person A's individually-inferred city from their bio
	AccountRegionA  string // person A's state
	AccountCityB    string // person B's individually-inferred city from their bio
	AccountRegionB  string // person B's state
	BioA            string // person A bio text (heuristic can check for location mentions)
	BioB            string // person B bio text
}

// Candidate is one possible mailing address — never mailed until human confirms.
type Candidate struct {
	Line1      string  `json:"line1"`
	Line2      string  `json:"line2,omitempty"`
	City       string  `json:"city"`
	Region     string  `json:"region"`
	Postal     string  `json:"postal,omitempty"`
	Country    string  `json:"country"`
	Confidence float64 `json:"confidence"`
	Source     string  `json:"source"`
	Note       string  `json:"note,omitempty"`
	// Optional identity hints from the provider
	FullName string `json:"full_name,omitempty"`
	Phone    string `json:"phone,omitempty"`
}

// Result is one provider call outcome for audit storage.
type Result struct {
	Provider   string
	Candidates []Candidate
	RawJSON    string
	Status     string // ok | empty | error
	Error      string
	CostCents  int
}

// Provider searches for people/addresses.
type Provider interface {
	Name() string
	Available() bool
	Search(ctx context.Context, q Query) (Result, error)
}

// NewProvider picks the best configured people-search backend.
// Cascade: Trestle → PDL → Cleanlist → Google Search → Heuristic.
func NewProvider() Provider {
	if k := strings.TrimSpace(os.Getenv("TRESTLE_API_KEY")); k != "" {
		return &Trestle{APIKey: k}
	}
	if k := strings.TrimSpace(os.Getenv("PDL_API_KEY")); k != "" {
		return &PDL{APIKey: k}
	}
	if k := strings.TrimSpace(os.Getenv("CLEANLIST_API_KEY")); k != "" {
		return &Cleanlist{APIKey: k}
	}
	if k := strings.TrimSpace(os.Getenv("GOOGLE_SEARCH_API_KEY")); k != "" {
		cx := strings.TrimSpace(os.Getenv("GOOGLE_SEARCH_CX"))
		if cx != "" {
			return &GoogleSearch{APIKey: k, Cx: cx}
		}
	}
	return &Heuristic{}
}

// Multi tries primary → secondary → heuristic so UI always gets something useful.
type Multi struct {
	Primary   Provider
	Fallback  Provider
	Secondary Provider // optional second-tier provider before heuristic
}

func NewMulti() *Multi {
	p := NewProvider()
	// Second tier: pick the next best provider that isn't the primary
	var second Provider
	switch p.(type) {
	case *Trestle:
		if k := strings.TrimSpace(os.Getenv("PDL_API_KEY")); k != "" {
			second = &PDL{APIKey: k}
		} else if k := strings.TrimSpace(os.Getenv("CLEANLIST_API_KEY")); k != "" {
			second = &Cleanlist{APIKey: k}
		}
	case *PDL:
		if k := strings.TrimSpace(os.Getenv("CLEANLIST_API_KEY")); k != "" {
			second = &Cleanlist{APIKey: k}
		}
	case *Cleanlist:
		if k := strings.TrimSpace(os.Getenv("PDL_API_KEY")); k != "" {
			second = &PDL{APIKey: k}
		}
	}
	// If primary is heuristic (no API keys), try Google Search as secondary
	if _, isHeuristic := p.(*Heuristic); isHeuristic {
		if k := strings.TrimSpace(os.Getenv("GOOGLE_SEARCH_API_KEY")); k != "" {
			cx := strings.TrimSpace(os.Getenv("GOOGLE_SEARCH_CX"))
			if cx != "" {
				second = &GoogleSearch{APIKey: k, Cx: cx}
			}
		}
	}
	return &Multi{Primary: p, Secondary: second, Fallback: &Heuristic{}}
}

func (m *Multi) Name() string {
	if m.Primary != nil && m.Primary.Available() {
		return m.Primary.Name()
	}
	if m.Secondary != nil && m.Secondary.Available() {
		return m.Secondary.Name()
	}
	return "heuristic"
}

func (m *Multi) Available() bool { return true }

func (m *Multi) Search(ctx context.Context, q Query) (Result, error) {
	// Tier 1: primary provider
	if m.Primary != nil && m.Primary.Available() {
		res, err := m.Primary.Search(ctx, q)
		if err == nil && len(res.Candidates) > 0 {
			return res, nil
		}
		// Tier 2: secondary provider (if primary empty/errored)
		if m.Secondary != nil && m.Secondary.Available() {
			res2, err2 := m.Secondary.Search(ctx, q)
			if err2 == nil && len(res2.Candidates) > 0 {
				if res.Error != "" {
					res2.Candidates[0].Note = strings.TrimSpace(res2.Candidates[0].Note + " | primary(" + m.Primary.Name() + "): " + res.Error)
				}
				res2.Status = "ok"
				return res2, nil
			}
		}
		// Tier 3: heuristic fallback
		fb, fbErr := m.Fallback.Search(ctx, q)
		if fbErr == nil && len(fb.Candidates) > 0 {
			if res.Error != "" && len(fb.Candidates) > 0 {
				fb.Candidates[0].Note = strings.TrimSpace(fb.Candidates[0].Note + " | primary(" + m.Primary.Name() + "): " + res.Error)
			} else if err != nil && len(fb.Candidates) > 0 {
				fb.Candidates[0].Note = strings.TrimSpace(fb.Candidates[0].Note + " | primary(" + m.Primary.Name() + "): " + err.Error())
			}
			fb.Status = "ok"
			return fb, nil
		}
		if err != nil {
			return res, err
		}
		return res, nil
	}
	// No primary — try secondary directly
	if m.Secondary != nil && m.Secondary.Available() {
		res, err := m.Secondary.Search(ctx, q)
		if err == nil && len(res.Candidates) > 0 {
			return res, nil
		}
	}
	return m.Fallback.Search(ctx, q)
}
