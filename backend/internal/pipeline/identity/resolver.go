// Package identity resolves handles/tags/follows to known CRM Persons and
// couples. It deliberately does NOT use facial recognition or any biometric
// signal as ground truth — a tagged handle, a linked profile, or a confirmed
// CRM record is what promotes a SocialAccount to a Person. An account with no
// such link stays "unknown" no matter how visually similar its photos are to
// someone else's.
package identity

import (
	"strings"

	"neptune-social-radar/backend/internal/ontology"
	"neptune-social-radar/backend/internal/pipeline/watchtower"
	"neptune-social-radar/backend/internal/signals"
	"neptune-social-radar/backend/internal/store"
)

type Resolved struct {
	Account        ontology.SocialAccount
	PartnerAccount *ontology.SocialAccount  // set when a tagged/followed counterpart is resolvable
	Couple         *ontology.Couple         // set once both sides have a Person
	TaggedAccounts []ontology.SocialAccount // every account tagged in this event, in order — mechanical fact, no interpretation
	// MentionedAccounts are accounts @mentioned in a post's caption — a
	// weaker reference than an image tag (spec §8: the most important
	// "tags" may be @mentions), recorded as mentioned_by edges and never
	// used by the reciprocal-identity check.
	MentionedAccounts []ontology.SocialAccount
}

func Resolve(s *store.Store, raw watchtower.RawEvent, observationID string) (Resolved, error) {
	acct, err := s.EnsureAccount(ontology.SocialAccount{Handle: raw.Handle})
	if err != nil {
		return Resolved{}, err
	}

	switch raw.Type {
	case "bio_change":
		bio, _ := raw.Payload["bio"].(string)
		if err := s.UpdateAccountBio(acct.ID, bio, raw.OccurredAt); err != nil {
			return Resolved{}, err
		}
		acct.BioText = bio

	case "account_disabled":
		disabled := true
		if v, ok := raw.Payload["disabled"].(bool); ok {
			disabled = v
		}
		if err := s.SetAccountDisabled(acct.ID, disabled); err != nil {
			return Resolved{}, err
		}
		acct.IsDisabled = disabled

	case "account_private":
		private := true
		if v, ok := raw.Payload["private"].(bool); ok {
			private = v
		}
		if err := s.SetAccountPrivate(acct.ID, private); err != nil {
			return Resolved{}, err
		}
		acct.IsPrivate = private

	case "username_change":
		// The account identity (ID, person link) is preserved across a
		// rename — a naive system might read "old handle disappeared" as
		// evidence of something meaningful. It isn't.
		newHandle, _ := raw.Payload["new_handle"].(string)
		if newHandle != "" {
			if _, err := s.DB.Exec(`UPDATE social_accounts SET handle = $1 WHERE id = $2`, newHandle, acct.ID); err == nil {
				acct.Handle = newHandle
			}
		}

	case "follow_change":
		targetHandle, _ := raw.Payload["target_handle"].(string)
		active := true
		if v, ok := raw.Payload["active"].(bool); ok {
			active = v
		}
		if targetHandle != "" {
			target, err := s.EnsureAccount(ontology.SocialAccount{Handle: targetHandle})
			if err != nil {
				return Resolved{}, err
			}
			if _, err := s.UpsertEdge(ontology.EdgeFollows, acct.ID, target.ID, active, raw.OccurredAt, observationID); err != nil {
				return Resolved{}, err
			}
		}

	case "post":
		var tagged []ontology.SocialAccount
		// Payloads carry []string from the mapper/worker, []any from JSON
		// round-trips and tests — accept both (the []any-only assertion once
		// silently dropped every tag edge in production).
		for _, handle := range payloadStrings(raw.Payload["tags"]) {
			if handle == "" {
				continue
			}
			taggedAcct, err := s.EnsureAccount(ontology.SocialAccount{Handle: handle})
			if err != nil {
				return Resolved{}, err
			}
			if _, err := s.UpsertEdge(ontology.EdgeTaggedWith, acct.ID, taggedAcct.ID, true, raw.OccurredAt, observationID); err != nil {
				return Resolved{}, err
			}
			tagged = append(tagged, taggedAcct)
		}
		// Caption @mentions are recorded as mentioned_by edges (from the
		// author to the mentioned account). They never promote identity on
		// their own — resolveCoupleFor only reads tag/follow edges — but
		// they are available to event-first couple resolution once language
		// analysis says the post is worth it.
		var mentioned []ontology.SocialAccount
		for _, handle := range signals.ExtractFromPayload(raw.Payload).MentionedHandles {
			if handle == acct.Handle {
				continue
			}
			mentionedAcct, err := s.EnsureAccount(ontology.SocialAccount{Handle: handle})
			if err != nil {
				return Resolved{}, err
			}
			if _, err := s.UpsertEdge(ontology.EdgeMentionedBy, acct.ID, mentionedAcct.ID, true, raw.OccurredAt, observationID); err != nil {
				return Resolved{}, err
			}
			mentioned = append(mentioned, mentionedAcct)
		}
		// A collaboration co-author is a first-class referenced account —
		// stronger than a tag (both accounts chose to publish together).
		// Recorded as a tag edge so the reciprocal-identity check can see it.
		if collab := signals.ExtractFromPayload(raw.Payload).Collab; collab != "" && collab != acct.Handle {
			collabAcct, err := s.EnsureAccount(ontology.SocialAccount{Handle: collab})
			if err != nil {
				return Resolved{}, err
			}
			if _, err := s.UpsertEdge(ontology.EdgeTaggedWith, acct.ID, collabAcct.ID, true, raw.OccurredAt, observationID); err != nil {
				return Resolved{}, err
			}
			tagged = append(tagged, collabAcct)
		}
		res, err := resolveCoupleFor(s, acct)
		if err != nil {
			return res, err
		}
		res.TaggedAccounts = tagged
		res.MentionedAccounts = mentioned
		return res, nil
	}

	return resolveCoupleFor(s, acct)
}

// resolveCoupleFor promotes accounts to Persons and Persons to a Couple once
// there's reciprocal identity evidence: a mutual tag (each has tagged the
// other) or a mutual follow. Either alone is not identity, but reciprocity
// between two already-tagged handles is the strongest non-biometric signal
// available and is what the spec explicitly recommends over face matching.
func resolveCoupleFor(s *store.Store, acct ontology.SocialAccount) (Resolved, error) {
	res := Resolved{Account: acct}

	edges, err := s.EdgesForAccount(acct.ID)
	if err != nil {
		return res, err
	}

	partnerAccountID := ""
	for _, e := range edges {
		other := e.ToAccountID
		if other == acct.ID {
			other = e.FromAccountID
		}
		if other == acct.ID || other == "" {
			continue
		}
		if hasReciprocalTagOrFollow(edges, acct.ID, other) {
			partnerAccountID = other
			break
		}
	}
	if partnerAccountID == "" {
		return res, nil
	}
	partner, err := s.GetAccount(partnerAccountID)
	if err != nil {
		return res, nil //nolint:nilerr // partner not resolvable yet is not an error
	}
	res.PartnerAccount = &partner

	// Ensure both sides have a Person. A handle promoted to Person purely via
	// reciprocal tag/follow evidence is tagged as such in crm_source, and
	// stays a lightweight/unconfirmed record — it is not the same trust tier
	// as a person who arrived via CRM intake (e.g. downloaded a guide).
	acctPerson, err := ensurePersonForAccount(s, acct, "inferred_from_reciprocal_tag")
	if err != nil {
		return res, err
	}
	partnerPerson, err := ensurePersonForAccount(s, partner, "inferred_from_reciprocal_tag")
	if err != nil {
		return res, err
	}

	couple, err := s.EnsureCouple(acctPerson.ID, partnerPerson.ID)
	if err != nil {
		return res, err
	}
	res.Couple = &couple
	return res, nil
}

// ResolveCoupleFromPair is the event-first identity entry point: given two
// accounts a third party tagged together (e.g. a wedding photographer's
// post), it names them as a candidate couple immediately — without requiring
// the two accounts to have ever interacted with each other directly. This is
// deliberately a WEAKER identity claim than resolveCoupleFor's reciprocal-tag/
// follow check: the caller (analyst/orchestrator) only reaches for this after
// language analysis confirms the post is worth resolving identity for, and
// the resulting hypothesis's partner-confidence score stays low until the two
// people corroborate the pairing themselves (a mutual follow, a bio update).
func ResolveCoupleFromPair(s *store.Store, a, b ontology.SocialAccount) (Resolved, error) {
	aPerson, err := ensurePersonForAccount(s, a, "inferred_from_co_tag")
	if err != nil {
		return Resolved{}, err
	}
	bPerson, err := ensurePersonForAccount(s, b, "inferred_from_co_tag")
	if err != nil {
		return Resolved{}, err
	}
	couple, err := s.EnsureCouple(aPerson.ID, bPerson.ID)
	if err != nil {
		return Resolved{}, err
	}
	// Re-fetch so callers see person_id populated on both accounts.
	aResolved, err := s.GetAccount(a.ID)
	if err != nil {
		return Resolved{}, err
	}
	bResolved, err := s.GetAccount(b.ID)
	if err != nil {
		return Resolved{}, err
	}
	return Resolved{Account: aResolved, PartnerAccount: &bResolved, Couple: &couple}, nil
}

func ensurePersonForAccount(s *store.Store, acct ontology.SocialAccount, crmSource string) (ontology.Person, error) {
	if acct.PersonID != "" {
		return s.GetPerson(acct.PersonID)
	}
	display := strings.ToUpper(acct.Handle[:1]) + acct.Handle[1:]
	p, err := s.CreatePerson(ontology.Person{DisplayName: display, CRMSource: crmSource})
	if err != nil {
		return p, err
	}
	if err := s.SetAccountPersonID(acct.ID, p.ID); err != nil {
		return p, err
	}
	return p, nil
}

// payloadStrings reads a string list from a payload field regardless of
// whether the producer stored []string (mapper/worker) or []any (JSON
// round-trip, hand-built test payloads).
func payloadStrings(v any) []string {
	switch list := v.(type) {
	case []string:
		return list
	case []any:
		out := make([]string, 0, len(list))
		for _, item := range list {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
		return out
	}
	return nil
}

func hasReciprocalTagOrFollow(edges []ontology.Edge, a, b string) bool {
	var aToB, bToA bool
	for _, e := range edges {
		if e.Kind != ontology.EdgeTaggedWith && e.Kind != ontology.EdgeFollows {
			continue
		}
		if !e.Active {
			continue
		}
		if e.FromAccountID == a && e.ToAccountID == b {
			aToB = true
		}
		if e.FromAccountID == b && e.ToAccountID == a {
			bToA = true
		}
	}
	return aToB && bToA
}
