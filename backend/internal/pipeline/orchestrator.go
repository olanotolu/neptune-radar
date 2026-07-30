// Package pipeline wires the eleven-stage loop described in the project
// brief: observe -> normalize -> resolve identity -> compare with prior state
// -> hypothesize -> corroborate/score -> policy check -> recommend -> (human
// approves) -> execute -> verify -> audit. Orchestrator is the only place
// that sees every stage; individual stage packages stay narrow and testable
// in isolation.
package pipeline

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"neptune-social-radar/backend/internal/llm"
	"neptune-social-radar/backend/internal/ontology"
	"neptune-social-radar/backend/internal/pipeline/analyst"
	"neptune-social-radar/backend/internal/pipeline/identity"
	"neptune-social-radar/backend/internal/pipeline/normalize"
	"neptune-social-radar/backend/internal/pipeline/policy"
	"neptune-social-radar/backend/internal/pipeline/roles"
	"neptune-social-radar/backend/internal/pipeline/scorer"
	"neptune-social-radar/backend/internal/pipeline/watchtower"
	"neptune-social-radar/backend/internal/store"
)

type Orchestrator struct {
	Store  *store.Store
	Interp llm.Interpreter
}

func New(s *store.Store, interp llm.Interpreter) *Orchestrator {
	return &Orchestrator{Store: s, Interp: interp}
}

type StepResult struct {
	Duplicate       bool
	NoSignal        bool
	ObservationID   string
	HypothesisID    string
	FinalConfidence float64
	ActionCreated   string // recommended_action id, if any
}

func (o *Orchestrator) ProcessEvent(ctx context.Context, raw watchtower.RawEvent) (StepResult, error) {
	obs, err := normalize.Normalize(raw)
	if err != nil {
		return StepResult{}, err
	}

	obs, err = o.Store.InsertObservation(obs)
	if err != nil {
		if errors.Is(err, store.ErrDuplicateObservation) {
			o.Store.Audit("social_observation", raw.ExternalEventID, "duplicate_suppressed",
				map[string]any{"handle": raw.Handle, "type": raw.Type}, raw.Monitor, -1)
			return StepResult{Duplicate: true}, nil
		}
		return StepResult{}, err
	}
	o.Store.Audit("social_observation", obs.ID, "observed", map[string]any{"handle": raw.Handle, "type": raw.Type}, raw.Monitor, -1)

	res, err := identity.Resolve(o.Store, raw, obs.ID)
	if err != nil {
		return StepResult{}, err
	}
	// Event-first resolution below may replace res.Account with a couple
	// member — capture the post author's account id now, because
	// co-occurrence sightings must be sourced by WHO OBSERVED the pair.
	authorAccountID := res.Account.ID
	o.Store.Audit("social_account", res.Account.ID, "identity_resolved",
		map[string]any{"handle": res.Account.Handle, "person_id": res.Account.PersonID, "couple_resolved": res.Couple != nil},
		raw.Monitor, -1)

	// Event-first discovery: an engagement-shaped post doesn't require the
	// couple to already be known. When two people are referenced (image tags
	// or caption @mentions):
	//  1. Attachment before creation — if the referenced pair is ALREADY a
	//     known couple, attach the event to it no matter what the post looks
	//     like: an ad or styled shoot ABOUT a known couple must still be
	//     scored and suppressed on the ledger, not silently dropped.
	//  2. Creation — only if the combined signal vocabulary says the post is
	//     worth resolving identity for (explicit language, a known vendor
	//     referencing exactly two people, or proposal-shaped visuals — never
	//     ads, styled shoots, or reposts), name the pair as a candidate
	//     couple now. This is the "public event detected -> couple resolved"
	//     step, deliberately separate from the language interpretation that
	//     happens in analyst.Detect below.
	if raw.Type == "post" {
		// Co-occurrence bookkeeping runs for EVERY post with 2+ referenced
		// accounts, regardless of what the language says — this is the
		// spec's PairCooccurrence object. Every unordered pair among the
		// referenced accounts is recorded (a post tagging 5 people is 10
		// pair sightings, not one) — recording only the first two made tag
		// ORDER silently bias the whole graph.
		referenced := dedupAccounts(append(res.TaggedAccounts, res.MentionedAccounts...))
		recordedPairs := map[string]bool{}
		for i := 0; i < len(referenced); i++ {
			for j := i + 1; j < len(referenced); j++ {
				if err := o.Store.RecordPairCooccurrence(referenced[i].ID, referenced[j].ID, res.Account.ID, raw.OccurredAt); err != nil {
					return StepResult{}, err
				}
				a, b := referenced[i].ID, referenced[j].ID
				if a > b {
					a, b = b, a
				}
				recordedPairs[a+"|"+b] = true
			}
		}

		if res.Couple == nil && len(referenced) >= 2 {
			// 1. Attachment before creation, on the RAW referenced pair: if
			//    these two are already a known couple, attach no matter what
			//    the post looks like — the couple record is vetted state,
			//    and even an ad ABOUT a known couple must be scored and
			//    suppressed on the ledger, not silently dropped.
			if couple, err := o.Store.GetCoupleForAccountPair(referenced[0].ID, referenced[1].ID); err == nil {
				a, errA := o.Store.GetAccount(referenced[0].ID)
				b, errB := o.Store.GetAccount(referenced[1].ID)
				if errA != nil || errB != nil {
					return StepResult{}, fmt.Errorf("attach to known couple: %w", errors.Join(errA, errB))
				}
				res = identity.Resolved{Account: a, PartnerAccount: &b, Couple: &couple, TaggedAccounts: res.TaggedAccounts, MentionedAccounts: res.MentionedAccounts}
				o.Store.Audit("couple", couple.ID, "attached_to_known_couple",
					map[string]any{"account_a": a.Handle, "account_b": b.Handle, "via": raw.Handle}, raw.Monitor, -1)
			} else {
				// 2. Creation path — Step 3, role resolution: a vendor
				//    tagged on the post (the florist, the venue) is NOT a
				//    couple candidate. Only person-role accounts may form a
				//    NEW pair.
				classified, err := roles.ClassifyReferenced(o.Store, referenced)
				if err != nil {
					return StepResult{}, err
				}
				candidates := roles.PairCandidates(classified)
				if vendorList := roles.Vendors(classified); len(vendorList) > 0 {
					vendorHandles := make([]string, 0, len(vendorList))
					for _, v := range vendorList {
						vendorHandles = append(vendorHandles, v.Account.Handle)
					}
					o.Store.Audit("social_observation", obs.ID, "vendor_refs_excluded_from_pair",
						map[string]any{"vendors": vendorHandles}, raw.Monitor, -1)
				}

				if len(candidates) >= 2 {
					// The role-filtered first-two may differ from the raw
					// tag order and still match an existing couple (e.g. one
					// created via the reciprocal-tag path) — check before
					// minting anything new.
					if couple, err := o.Store.GetCoupleForAccountPair(candidates[0].ID, candidates[1].ID); err == nil {
						a, errA := o.Store.GetAccount(candidates[0].ID)
						b, errB := o.Store.GetAccount(candidates[1].ID)
						if errA != nil || errB != nil {
							return StepResult{}, fmt.Errorf("attach to known couple: %w", errors.Join(errA, errB))
						}
						res = identity.Resolved{Account: a, PartnerAccount: &b, Couple: &couple, TaggedAccounts: res.TaggedAccounts, MentionedAccounts: res.MentionedAccounts}
						o.Store.Audit("couple", couple.ID, "attached_to_known_couple",
							map[string]any{"account_a": a.Handle, "account_b": b.Handle, "via": raw.Handle}, raw.Monitor, -1)
					} else if analyst.WorthResolvingIdentity(raw) {
						upgraded, err := identity.ResolveCoupleFromPair(o.Store, candidates[0], candidates[1])
						if err != nil {
							return StepResult{}, err
						}
						res = upgraded
						o.Store.Audit("couple", res.Couple.ID, "event_first_couple_resolved",
							map[string]any{"account_a": res.Account.Handle, "account_b": res.PartnerAccount.Handle}, raw.Monitor, -1)
					}
				}
			}
		}

		// The scorer reads co-occurrence for the FINAL resolved pair
		// (res.Account × res.PartnerAccount) — which may not be among the
		// referenced pairs recorded above (e.g. the reciprocal path resolves
		// the post's author × a tagging counterpart, and the author is not
		// in the referenced list). Record that pair too, sourced by the
		// post's AUTHOR (raw.Handle → authorAccountID captured before
		// event-first resolution may have replaced res.Account), or the +10
		// repeated-co-occurrence evidence reads a row that was never written.
		if res.Couple != nil && res.PartnerAccount != nil {
			a, b := res.Account.ID, res.PartnerAccount.ID
			if a > b {
				a, b = b, a
			}
			if !recordedPairs[a+"|"+b] {
				if err := o.Store.RecordPairCooccurrence(res.Account.ID, res.PartnerAccount.ID, authorAccountID, raw.OccurredAt); err != nil {
					return StepResult{}, err
				}
			}
		}
	}

	var prior *ontology.Relationship
	if res.Couple != nil {
		r, err := o.Store.CurrentRelationship(res.Couple.ID)
		if err == nil {
			prior = &r
		} else if !errors.Is(err, sql.ErrNoRows) {
			return StepResult{}, err
		}
	}

	cand, err := analyst.Detect(o.Store, res, raw, prior)
	if err != nil {
		return StepResult{}, err
	}
	if cand == nil {
		o.Store.Audit("social_observation", obs.ID, "no_candidate_signal", nil, raw.Monitor, -1)
		return StepResult{ObservationID: obs.ID, NoSignal: true}, nil
	}

	partnerHandle := ""
	if res.PartnerAccount != nil {
		partnerHandle = res.PartnerAccount.Handle
	}
	priorStage := string(ontology.StageUnknown)
	if prior != nil {
		priorStage = string(prior.Stage)
	}

	hyp, isNew, err := o.findOrCreateHypothesis(res, *cand)
	if err != nil {
		return StepResult{}, err
	}

	var existingEvidence []string
	if !isNew {
		ev, err := o.Store.EvidenceForHypothesis(hyp.ID)
		if err != nil {
			return StepResult{}, err
		}
		for _, e := range ev {
			existingEvidence = append(existingEvidence, e.Description)
		}
	}

	interp, err := analyst.Interpret(ctx, o.Interp, *cand, raw, partnerHandle, priorStage, existingEvidence)
	if err != nil {
		return StepResult{}, err
	}
	if isNew {
		if err := o.Store.UpdateHypothesisModelRule(hyp.ID, interp.Source); err != nil {
			return StepResult{}, err
		}
	}
	o.Store.Audit("hypothesis", hyp.ID, "analyst_interpreted",
		map[string]any{"source": interp.Source, "confidence": interp.Confidence, "rationale": interp.Rationale}, raw.Monitor, -1)

	evidence, err := scorer.CollectEvidence(o.Store, hyp, res, *cand, raw, interp.Confidence)
	if err != nil {
		return StepResult{}, err
	}

	var finalScore float64
	// Per-evidence (kind → weight) snapshot in the audit detail: evidence
	// rows are overwritten in place as understanding changes, so without
	// this the "why was confidence X at time T" trail was unreconstructable.
	evidenceWeights := map[string]float64{}
	for _, e := range evidence {
		evidenceWeights[e.Kind] = e.Weight
	}
	auditDetail := map[string]any{"evidence_count": len(evidence), "evidence_weights": evidenceWeights}
	if cand.EventType == ontology.EventTypeEngagement {
		// Engagement prospects run on the spec's points table: every point
		// is an auditable deterministic signal (explicit language +40, both
		// partners tagged +25, vendor +15, visual +10, reciprocal +10,
		// registry +15, recent +10; styled shoot/ad −50, no second person
		// −30, old/reposted −25, identity conflict −40). The two sub-scores
		// — "did this happen" and "did we get the right two people" — are
		// still reported separately so a confident caption tagging the wrong
		// person is visible as exactly that.
		var engagementConf, partnerConf float64
		finalScore, engagementConf, partnerConf = scorer.ProspectScore(evidence)
		if err := o.Store.UpdateHypothesisSubScores(hyp.ID, engagementConf, partnerConf); err != nil {
			return StepResult{}, err
		}
		hyp.EngagementConfidence, hyp.PartnerConfidence = &engagementConf, &partnerConf
		auditDetail["engagement_confidence"] = engagementConf
		auditDetail["partner_confidence"] = partnerConf
	} else {
		finalScore = scorer.Score(interp.Confidence, evidence)
	}
	if err := o.Store.UpdateHypothesisConfidence(hyp.ID, finalScore); err != nil {
		return StepResult{}, err
	}
	hyp.Confidence = finalScore // keep the in-memory copy in sync for the copy drafted below
	auditDetail["final_confidence"] = finalScore
	o.Store.Audit("hypothesis", hyp.ID, "scored", auditDetail, raw.Monitor, -1)

	// Internal belief update: the ontology's own understanding of the couple
	// changes as soon as evidence crosses the bar. This is Neptune's
	// internal state, not a customer-visible or irreversible action, so it
	// does not wait for human approval — only the recommended_action below
	// does. Engagement belief requires the investigation tier (0.70): the
	// spec discards anything below it, so it must not silently become "what
	// Neptune believes" either. State changes keep the 0.60 bar.
	stageBar := policy.ThresholdActOnStage
	if cand.EventType == ontology.EventTypeEngagement {
		stageBar = policy.ThresholdInvestigate
	}
	if res.Couple != nil && finalScore >= stageBar && (prior == nil || prior.Stage != cand.ProposedStage) {
		paused := false
		if prior != nil {
			paused = prior.AutomationPaused
		}
		if _, err := o.Store.TransitionRelationship(res.Couple.ID, cand.ProposedStage, finalScore, ontology.ScopeNeptuneInternal, paused); err != nil {
			return StepResult{}, err
		}
		if err := o.Store.UpdateHypothesisStatus(hyp.ID, ontology.HypothesisCorroborating); err != nil {
			return StepResult{}, err
		}
		o.Store.Audit("relationship", res.Couple.ID, "stage_transition",
			map[string]any{"stage": cand.ProposedStage, "confidence": finalScore}, raw.Monitor, -1)
	}

	decision, err := policy.Decide(o.Store, hyp, finalScore)
	if err != nil {
		return StepResult{}, err
	}
	o.Store.Audit("hypothesis", hyp.ID, "policy_decision",
		map[string]any{"should_act": decision.ShouldAct, "action_type": decision.ActionType, "reasons": decision.Reasons}, raw.Monitor, -1)

	result := StepResult{ObservationID: obs.ID, HypothesisID: hyp.ID, FinalConfidence: finalScore}
	if !decision.ShouldAct {
		return result, nil
	}

	action, err := o.proposeAction(ctx, decision, hyp, res, evidence, raw)
	if err != nil {
		return StepResult{}, err
	}
	result.ActionCreated = action.ID
	o.Store.Audit("recommended_action", action.ID, "action_recommended",
		map[string]any{"action_type": action.ActionType}, raw.Monitor, -1)
	return result, nil
}

func dedupAccounts(accounts []ontology.SocialAccount) []ontology.SocialAccount {
	seen := map[string]bool{}
	var out []ontology.SocialAccount
	for _, a := range accounts {
		if !seen[a.ID] {
			seen[a.ID] = true
			out = append(out, a)
		}
	}
	return out
}

func (o *Orchestrator) findOrCreateHypothesis(res identity.Resolved, cand analyst.Candidate) (ontology.LifeEventHypothesis, bool, error) {
	if res.Couple == nil {
		// Shouldn't happen — Detect requires a resolved couple — but guard
		// defensively rather than panic on a nil dereference below.
		return ontology.LifeEventHypothesis{}, false, sql.ErrNoRows
	}
	existing, err := o.Store.LatestHypothesisForCouple(res.Couple.ID)
	if err == nil && existing.EventType == cand.EventType &&
		(existing.Status == ontology.HypothesisUnconfirmed || existing.Status == ontology.HypothesisCorroborating) {
		return existing, false, nil
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return ontology.LifeEventHypothesis{}, false, err
	}
	hyp, err := o.Store.CreateHypothesis(ontology.LifeEventHypothesis{
		CoupleID:        res.Couple.ID,
		PersonID:        res.Account.PersonID,
		EventType:       cand.EventType,
		ProposedStage:   cand.ProposedStage,
		Confidence:      0,
		ModelOrRule:     "pending", // updated to the interpreter's identity right after this call
		Status:          ontology.HypothesisUnconfirmed,
		VisibilityScope: ontology.ScopeUnconfirmedInfer,
		ConsentScope:    ontology.ScopeUnconfirmedInfer,
	})
	return hyp, true, err
}
