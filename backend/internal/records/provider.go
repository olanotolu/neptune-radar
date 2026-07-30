// Package records finds mailing-address candidates from name + market signals.
// Implementations: People Data Labs, Melissa (optional), deterministic heuristic fallback.
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
func NewProvider() Provider {
	if k := strings.TrimSpace(os.Getenv("PDL_API_KEY")); k != "" {
		return &PDL{APIKey: k}
	}
	if k := strings.TrimSpace(os.Getenv("MELISSA_LICENSE_KEY")); k != "" {
		return &Melissa{LicenseKey: k}
	}
	return &Heuristic{}
}

// Multi tries primary then falls back to heuristic so UI always gets something useful.
type Multi struct {
	Primary Provider
	Fallback Provider
}

func NewMulti() *Multi {
	p := NewProvider()
	return &Multi{Primary: p, Fallback: &Heuristic{}}
}

func (m *Multi) Name() string {
	if m.Primary != nil && m.Primary.Available() {
		return m.Primary.Name()
	}
	return "heuristic"
}

func (m *Multi) Available() bool { return true }

func (m *Multi) Search(ctx context.Context, q Query) (Result, error) {
	if m.Primary != nil && m.Primary.Available() {
		res, err := m.Primary.Search(ctx, q)
		if err == nil && len(res.Candidates) > 0 {
			return res, nil
		}
		// Fall back so operators always get market-level guidance.
		fb, fbErr := m.Fallback.Search(ctx, q)
		if fbErr == nil && len(fb.Candidates) > 0 {
			if res.Error != "" && len(fb.Candidates) > 0 {
				fb.Candidates[0].Note = strings.TrimSpace(fb.Candidates[0].Note + " | primary: " + res.Error)
			} else if err != nil && len(fb.Candidates) > 0 {
				fb.Candidates[0].Note = strings.TrimSpace(fb.Candidates[0].Note + " | primary: " + err.Error())
			}
			fb.Status = "ok"
			return fb, nil
		}
		if err != nil {
			return res, err
		}
		return res, nil
	}
	return m.Fallback.Search(ctx, q)
}
