package pipeline_test

import (
	"context"
	"testing"
	"time"

	"neptune-social-radar/backend/internal/llm"
	"neptune-social-radar/backend/internal/ontology"
	"neptune-social-radar/backend/internal/pipeline"
	"neptune-social-radar/backend/internal/pipeline/operator"
	"neptune-social-radar/backend/internal/pipeline/verifier"
	"neptune-social-radar/backend/internal/pipeline/watchtower"
	"neptune-social-radar/backend/internal/store/storetest"
)

// TestEndToEndVendorDiscovery drives the full live loop with inline events
// (no fixtures, no replay engine): a watched vendor's post discovers a couple
// the CRM never knew, the points table scores it into the create-prospect
// tier, and a human approval opens the prenup case.
func TestEndToEndVendorDiscovery(t *testing.T) {
	s := storetest.Open(t)
	orch := pipeline.New(s, llm.NewTemplateInterpreter())
	ctx := context.Background()

	if _, err := s.AddWatchedSource("weddingsbynoor", "engagement_photographer"); err != nil {
		t.Fatalf("add watched source: %v", err)
	}

	post := watchtower.RawEvent{
		Monitor:         "vendor:weddingsbynoor",
		Source:          "apify",
		ExternalEventID: "post_001",
		Handle:          "weddingsbynoor",
		Type:            "post",
		OccurredAt:      time.Now().UTC(),
		Payload: map[string]any{
			"caption":             "She said yes! Congratulations to @mayak and @jordanl #JustEngaged #NYCEngagement",
			"tags":                []any{"mayak", "jordanl"},
			"source_account_type": "engagement_photographer",
			"visual_signals":      []any{"ring"},
			"location":            "Central Park, NYC",
		},
	}
	res, err := orch.ProcessEvent(ctx, post)
	if err != nil {
		t.Fatalf("process vendor post: %v", err)
	}
	if res.ActionCreated == "" {
		t.Fatalf("expected a prospect action from a 100-point post, got %+v", res)
	}
	if res.FinalConfidence < 0.90 {
		t.Errorf("expected create-tier confidence, got %.2f", res.FinalConfidence)
	}

	action, err := s.GetAction(res.ActionCreated)
	if err != nil {
		t.Fatalf("get action: %v", err)
	}
	if action.ActionType != ontology.ActionReview {
		t.Errorf("expected review action, got %s", action.ActionType)
	}

	// Mutual follows corroborate the pairing — partner evidence rises, and
	// the duplicate-pending-action guard must hold (still exactly one card).
	for _, f := range []watchtower.RawEvent{
		{Monitor: "follow:jordanl", Source: "apify", ExternalEventID: "fw_1", Handle: "jordanl", Type: "follow_change", OccurredAt: time.Now().UTC(), Payload: map[string]any{"target_handle": "mayak", "active": true}},
		{Monitor: "follow:mayak", Source: "apify", ExternalEventID: "fw_2", Handle: "mayak", Type: "follow_change", OccurredAt: time.Now().UTC(), Payload: map[string]any{"target_handle": "jordanl", "active": true}},
	} {
		if _, err := orch.ProcessEvent(ctx, f); err != nil {
			t.Fatalf("process follow: %v", err)
		}
	}
	pending, err := s.ListActions("pending")
	if err != nil {
		t.Fatalf("list actions: %v", err)
	}
	if len(pending) != 1 {
		t.Fatalf("expected exactly one pending card (duplicates suppressed), got %d", len(pending))
	}

	exec, err := operator.Approve(s, action.ID, "human:concierge")
	if err != nil {
		t.Fatalf("approve: %v", err)
	}
	verified, err := verifier.Confirm(s, exec, action)
	if err != nil || !verified {
		t.Fatalf("expected verified execution, got %v, %v", verified, err)
	}
	leads, err := s.ListLeads("")
	if err != nil || len(leads) != 1 {
		t.Fatalf("expected one lead, got %+v (%v)", leads, err)
	}
	cases, err := s.ListCases("intake")
	if err != nil || len(cases) != 1 {
		t.Fatalf("expected one intake case, got %+v (%v)", cases, err)
	}
}

// TestEndToEndAdSuppression drives an ad about a known couple through the
// same loop and proves the −50 exclusion suppresses it on the ledger.
func TestEndToEndAdSuppression(t *testing.T) {
	s := storetest.Open(t)
	orch := pipeline.New(s, llm.NewTemplateInterpreter())
	ctx := context.Background()

	// A known couple: maya16 is CRM-known with consent; the pair mutually follow.
	maya, err := s.CreatePerson(ontology.Person{DisplayName: "Maya K.", CRMSource: "prenup_guide_download"})
	if err != nil {
		t.Fatalf("create person: %v", err)
	}
	if _, err := s.EnsureAccount(ontology.SocialAccount{Handle: "maya16", PersonID: maya.ID}); err != nil {
		t.Fatalf("link account: %v", err)
	}
	if _, err := s.CreateConsentPolicy(ontology.ConsentPolicy{PersonID: maya.ID, Scope: ontology.ScopeNeptuneInternal, AllowedActions: []string{"create_lead", "pause_automation", "draft_outreach"}}); err != nil {
		t.Fatalf("create consent: %v", err)
	}
	for _, f := range []watchtower.RawEvent{
		{Monitor: "follow:jordan16", Source: "apify", ExternalEventID: "fa_1", Handle: "jordan16", Type: "follow_change", OccurredAt: time.Now().UTC(), Payload: map[string]any{"target_handle": "maya16", "active": true}},
		{Monitor: "follow:maya16", Source: "apify", ExternalEventID: "fa_2", Handle: "maya16", Type: "follow_change", OccurredAt: time.Now().UTC(), Payload: map[string]any{"target_handle": "jordan16", "active": true}},
	} {
		if _, err := orch.ProcessEvent(ctx, f); err != nil {
			t.Fatalf("process follow: %v", err)
		}
	}

	ad := watchtower.RawEvent{
		Monitor:         "hashtag:shesaidyes",
		Source:          "apify",
		ExternalEventID: "post_ad_1",
		Handle:          "brilliantearth.nyc",
		Type:            "post",
		OccurredAt:      time.Now().UTC(),
		Payload: map[string]any{
			"caption":             "The ring that made her say yes 💍 #SheSaidYes #Ad #JewelryAd",
			"tags":                []any{"maya16", "jordan16"},
			"source_account_type": "jeweler",
			"visual_signals":      []any{"ring"},
		},
	}
	res, err := orch.ProcessEvent(ctx, ad)
	if err != nil {
		t.Fatalf("process ad post: %v", err)
	}
	if res.ActionCreated != "" {
		t.Fatalf("an advertisement must never produce an action, got %s", res.ActionCreated)
	}
	if res.FinalConfidence >= 0.70 {
		t.Errorf("expected the ad to score below the investigation tier, got %.2f", res.FinalConfidence)
	}
	// The suppression must be visible on the evidence ledger.
	ev, err := s.EvidenceForHypothesis(res.HypothesisID)
	if err != nil {
		t.Fatalf("evidence: %v", err)
	}
	foundAd := false
	for _, e := range ev {
		if e.Kind == "advertisement" && e.Weight == -0.50 {
			foundAd = true
		}
	}
	if !foundAd {
		t.Errorf("expected an advertisement −50 evidence row, have %+v", ev)
	}
}
