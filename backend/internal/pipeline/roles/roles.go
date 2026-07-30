// Package roles implements the spec's Step 3 (assign roles): given the
// accounts referenced on a post, decide which ones are plausible couple
// members and which are vendors (photographer, planner, venue, florist…)
// that merely got tagged alongside them. Without this, a photographer's post
// tagging @coupleA @coupleB @florist would happily mint a "couple" out of
// Jane and the florist.
//
// Two deterministic vendor tests, no model judgment:
//
//  1. Registry test — the account is on the curated watched_sources list
//     with a vendor class. Photographers/planners/venues are the radar's
//     sensors; a sensor is never its own signal's subject.
//
//  2. Structural test — the account's outbound tag/mention edges fan out to
//     MANY distinct accounts (vendor node: many weak client relationships),
//     while a personal profile's edges concentrate on a small circle
//     (partner node: one strong reciprocal relationship).
//
// Everything here is re-derivable from edges/registry state, so a vendor
// classification can change as evidence accumulates — it is never a
// permanent label on the account.
package roles

import (
	"neptune-social-radar/backend/internal/ontology"
	"neptune-social-radar/backend/internal/store"
)

// StructuralVendorMinTargets is how many distinct accounts an account must
// tag/mention before it looks like a vendor node rather than a person. Well
// below any real photographer's client count, well above a personal
// profile's partner references.
const StructuralVendorMinTargets = 12

// ClassifiedAccount pairs an account with its resolved role.
type ClassifiedAccount struct {
	Account  ontology.SocialAccount
	IsVendor bool
	Reason   string // audit-friendly explanation, "" for person-role
}

// ClassifyReferenced assigns each referenced account a role. Order is
// preserved — the caller still prefers the first two person-roles as the
// candidate pair, keeping tag order meaningful (couples are usually tagged
// before vendors on photographer posts).
func ClassifyReferenced(s *store.Store, referenced []ontology.SocialAccount) ([]ClassifiedAccount, error) {
	out := make([]ClassifiedAccount, 0, len(referenced))
	for _, acct := range referenced {
		c := ClassifiedAccount{Account: acct}

		if class := s.SourceClassForHandle(acct.Handle); class != "" {
			c.IsVendor = true
			c.Reason = "on the curated source registry as " + class
			out = append(out, c)
			continue
		}

		targets, err := countDistinctOutboundRefs(s, acct.ID)
		if err != nil {
			return nil, err
		}
		if targets >= StructuralVendorMinTargets {
			c.IsVendor = true
			c.Reason = "tags/mentions many distinct accounts — vendor-shaped node"
		}
		out = append(out, c)
	}
	return out, nil
}

// PairCandidates returns the referenced accounts that plausibly ARE the
// couple — vendors excluded — in original tag order.
func PairCandidates(classified []ClassifiedAccount) []ontology.SocialAccount {
	var out []ontology.SocialAccount
	for _, c := range classified {
		if !c.IsVendor {
			out = append(out, c.Account)
		}
	}
	return out
}

// Vendors returns the referenced accounts classified as vendors, for audit
// detail ("we saw the florist and left her out of the couple").
func Vendors(classified []ClassifiedAccount) []ClassifiedAccount {
	var out []ClassifiedAccount
	for _, c := range classified {
		if c.IsVendor {
			out = append(out, c)
		}
	}
	return out
}

// countDistinctOutboundRefs counts how many distinct accounts this account
// has ever tagged or mentioned — the structural vendor signal.
func countDistinctOutboundRefs(s *store.Store, accountID string) (int, error) {
	var n int
	err := s.DB.QueryRow(
		`SELECT COUNT(DISTINCT to_account_id) FROM edges
		 WHERE from_account_id = $1 AND kind IN ('tagged_with', 'mentioned_by')`,
		accountID,
	).Scan(&n)
	return n, err
}
