package policy_test

import (
	"testing"

	"neptune-social-radar/backend/internal/ontology"
	"neptune-social-radar/backend/internal/pipeline/analyst"
	"neptune-social-radar/backend/internal/pipeline/policy"
	"neptune-social-radar/backend/internal/store"
	"neptune-social-radar/backend/internal/store/storetest"
)

func setupHypothesis(t *testing.T, s *store.Store, eventType string, consentActions []string) ontology.LifeEventHypothesis {
	t.Helper()
	p, err := s.CreatePerson(ontology.Person{DisplayName: "Test Person"})
	if err != nil {
		t.Fatalf("create person: %v", err)
	}
	if consentActions != nil {
		if _, err := s.CreateConsentPolicy(ontology.ConsentPolicy{
			PersonID: p.ID, Scope: ontology.ScopeNeptuneInternal, AllowedActions: consentActions,
		}); err != nil {
			t.Fatalf("create consent: %v", err)
		}
	}
	hyp, err := s.CreateHypothesis(ontology.LifeEventHypothesis{
		PersonID: p.ID, EventType: eventType, ProposedStage: ontology.StageEngaged,
		ConsentScope: ontology.ScopeUnconfirmedInfer,
	})
	if err != nil {
		t.Fatalf("create hypothesis: %v", err)
	}
	return hyp
}

func TestDecide_BelowThreshold_NeverActs(t *testing.T) {
	s := storetest.Open(t)
	defer s.Close()
	hyp := setupHypothesis(t, s, analyst.CandidateEngagement, []string{"create_lead"})

	d, err := policy.Decide(s, hyp, 0.59)
	if err != nil {
		t.Fatalf("decide: %v", err)
	}
	if d.ShouldAct {
		t.Error("expected no action below the review threshold, even with valid consent")
	}
}

func TestDecide_AtThreshold_NoConsent_Denied(t *testing.T) {
	s := storetest.Open(t)
	defer s.Close()
	hyp := setupHypothesis(t, s, analyst.CandidateEngagement, nil) // no consent policy at all

	d, err := policy.Decide(s, hyp, 0.9)
	if err != nil {
		t.Fatalf("decide: %v", err)
	}
	if d.ShouldAct {
		t.Error("expected no action when the person has no consent policy on file")
	}
	if d.ConsentOK {
		t.Error("expected ConsentOK=false with no consent policy")
	}
}

func TestDecide_ConsentScopeMismatch_Denied(t *testing.T) {
	s := storetest.Open(t)
	defer s.Close()
	// consent exists, but doesn't include the action this event type needs
	hyp := setupHypothesis(t, s, analyst.CandidateRelationshipChange, []string{"create_lead"})

	d, err := policy.Decide(s, hyp, 0.9)
	if err != nil {
		t.Fatalf("decide: %v", err)
	}
	if d.ShouldAct {
		t.Error("expected no action when consent doesn't cover pause_automation")
	}
}

func TestDecide_EngagementInCreateTier_RecommendsReview(t *testing.T) {
	s := storetest.Open(t)
	defer s.Close()
	hyp := setupHypothesis(t, s, analyst.CandidateEngagement, []string{"create_lead"})

	d, err := policy.Decide(s, hyp, 0.95) // 95 points: create-prospect tier
	if err != nil {
		t.Fatalf("decide: %v", err)
	}
	if !d.ShouldAct {
		t.Fatalf("expected ShouldAct=true, reasons: %v", d.Reasons)
	}
	if d.ActionType != ontology.ActionReview {
		t.Errorf("expected action_type=review, got %s", d.ActionType)
	}
}

func TestDecide_EngagementInvestigationTier_RoutesToInvestigate(t *testing.T) {
	s := storetest.Open(t)
	defer s.Close()
	hyp := setupHypothesis(t, s, analyst.CandidateEngagement, []string{"create_lead"})

	d, err := policy.Decide(s, hyp, 0.85) // 85 points: 70–89 investigation queue
	if err != nil {
		t.Fatalf("decide: %v", err)
	}
	if !d.ShouldAct {
		t.Fatalf("expected ShouldAct=true, reasons: %v", d.Reasons)
	}
	if d.ActionType != ontology.ActionInvestigate {
		t.Errorf("expected action_type=investigate, got %s", d.ActionType)
	}
}

func TestDecide_EngagementBelowInvestigateTier_NeverActs(t *testing.T) {
	s := storetest.Open(t)
	defer s.Close()
	hyp := setupHypothesis(t, s, analyst.CandidateEngagement, []string{"create_lead"})

	d, err := policy.Decide(s, hyp, 0.65) // 65 points: retained/discarded per spec
	if err != nil {
		t.Fatalf("decide: %v", err)
	}
	if d.ShouldAct {
		t.Error("expected no action below the 0.70 investigation tier, even with valid consent")
	}
}

func TestDecide_StateChangeAboveThreshold_RecommendsConciergeReview(t *testing.T) {
	s := storetest.Open(t)
	defer s.Close()
	hyp := setupHypothesis(t, s, analyst.CandidateRelationshipChange, []string{"pause_automation"})

	d, err := policy.Decide(s, hyp, 0.8)
	if err != nil {
		t.Fatalf("decide: %v", err)
	}
	if !d.ShouldAct {
		t.Fatalf("expected ShouldAct=true, reasons: %v", d.Reasons)
	}
	if d.ActionType != ontology.ActionConciergeReview {
		t.Errorf("expected action_type=concierge_review, got %s", d.ActionType)
	}
}

func TestDecide_FreshDiscovery_EngagementSurfacesWithoutConsent(t *testing.T) {
	s := storetest.Open(t)
	defer s.Close()
	// A person who exists ONLY because a third party tagged them — never a
	// Neptune lead, so there is nothing they could have consented to yet.
	p, err := s.CreatePerson(ontology.Person{DisplayName: "Stranger", CRMSource: "inferred_from_co_tag"})
	if err != nil {
		t.Fatalf("create person: %v", err)
	}
	hyp, err := s.CreateHypothesis(ontology.LifeEventHypothesis{
		PersonID: p.ID, EventType: analyst.CandidateEngagement, ProposedStage: ontology.StageEngaged,
		ConsentScope: ontology.ScopeUnconfirmedInfer,
	})
	if err != nil {
		t.Fatalf("create hypothesis: %v", err)
	}

	d, err := policy.Decide(s, hyp, 0.95) // no consent policy created at all
	if err != nil {
		t.Fatalf("decide: %v", err)
	}
	if !d.ShouldAct {
		t.Fatalf("expected a freshly-discovered prospect to surface for human review without pre-existing consent, reasons: %v", d.Reasons)
	}
	if d.ActionType != ontology.ActionReview {
		t.Errorf("expected action_type=review, got %s", d.ActionType)
	}
}

func TestDecide_FreshDiscovery_StateChangeStillRequiresConsent(t *testing.T) {
	s := storetest.Open(t)
	defer s.Close()
	// The consent bypass is specifically for surfacing a NEW prospect. It must
	// not quietly extend to monitoring an existing relationship's stability —
	// that's a different, higher-stakes action even for an inferred person.
	p, err := s.CreatePerson(ontology.Person{DisplayName: "Stranger", CRMSource: "inferred_from_reciprocal_tag"})
	if err != nil {
		t.Fatalf("create person: %v", err)
	}
	hyp, err := s.CreateHypothesis(ontology.LifeEventHypothesis{
		PersonID: p.ID, EventType: analyst.CandidateRelationshipChange, ProposedStage: ontology.StageStatusUncertain,
		ConsentScope: ontology.ScopeUnconfirmedInfer,
	})
	if err != nil {
		t.Fatalf("create hypothesis: %v", err)
	}

	d, err := policy.Decide(s, hyp, 0.8)
	if err != nil {
		t.Fatalf("decide: %v", err)
	}
	if d.ShouldAct {
		t.Error("expected relationship-state-change monitoring to still require consent, even for an inferred person")
	}
}

func TestDecide_Idempotent_NoDuplicatePendingAction(t *testing.T) {
	s := storetest.Open(t)
	defer s.Close()
	hyp := setupHypothesis(t, s, analyst.CandidateEngagement, []string{"create_lead"})

	d1, err := policy.Decide(s, hyp, 0.9)
	if err != nil || !d1.ShouldAct {
		t.Fatalf("first decide should act: %v %+v", err, d1)
	}
	if _, err := s.CreateRecommendedAction(ontology.RecommendedAction{HypothesisID: hyp.ID, ActionType: d1.ActionType}); err != nil {
		t.Fatalf("create recommended action: %v", err)
	}

	d2, err := policy.Decide(s, hyp, 0.9)
	if err != nil {
		t.Fatalf("second decide: %v", err)
	}
	if d2.ShouldAct {
		t.Error("expected duplicate suppression: a pending action for this hypothesis already exists")
	}
}
