package pipeline_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"neptune-social-radar/backend/internal/llm"
	"neptune-social-radar/backend/internal/pipeline"
	"neptune-social-radar/backend/internal/pipeline/watchtower"
	"neptune-social-radar/backend/internal/store/storetest"
)

// TestVendorExcludedFromPair drives the spec's Step 3: a photographer's post
// tags the couple AND a florist. The florist — a registry-known vendor — must
// be excluded from the candidate pair; the couple is Jane + Alex, never
// Jane + florist.
func TestVendorExcludedFromPair(t *testing.T) {
	s := storetest.Open(t)
	orch := pipeline.New(s, llm.NewTemplateInterpreter())
	ctx := context.Background()

	if _, err := s.AddWatchedSource("columbusproposalphoto", "engagement_photographer"); err != nil {
		t.Fatalf("add photographer: %v", err)
	}
	if _, err := s.AddWatchedSource("flowerstudio", "florist"); err != nil {
		t.Fatalf("add florist: %v", err)
	}

	post := watchtower.RawEvent{
		Monitor:         "vendor:columbusproposalphoto",
		Source:          "apify",
		ExternalEventID: "post_roles_001",
		Handle:          "columbusproposalphoto",
		Type:            "post",
		OccurredAt:      time.Now().UTC(),
		Payload: map[string]any{
			// Florist mentioned FIRST in the caption — a naive first-two-
			// referenced implementation would pair Jane with the florist.
			"caption":             "Alex surprised Jane at sunset and she said yes. Flowers by @flowerstudio. Congrats @janedoe and @alexsmith! #JustEngaged",
			"tags":                []any{"janedoe", "alexsmith", "flowerstudio"},
			"source_account_type": "engagement_photographer",
			"location":            "Columbus, OH",
		},
	}
	if _, err := orch.ProcessEvent(ctx, post); err != nil {
		t.Fatalf("process post: %v", err)
	}

	// The couple must be janedoe + alexsmith — the florist was excluded.
	couples, err := s.ListCouples()
	if err != nil {
		t.Fatalf("list couples: %v", err)
	}
	if len(couples) != 1 {
		t.Fatalf("expected exactly one couple, got %d", len(couples))
	}
	c := couples[0]
	pa, err := s.GetPerson(c.PersonAID)
	if err != nil {
		t.Fatalf("get person A: %v", err)
	}
	pb, err := s.GetPerson(c.PersonBID)
	if err != nil {
		t.Fatalf("get person B: %v", err)
	}
	names := strings.ToLower(pa.DisplayName + " " + pb.DisplayName)
	if !strings.Contains(names, "janedoe") || !strings.Contains(names, "alexsmith") {
		t.Errorf("couple is %q + %q, want janedoe + alexsmith", pa.DisplayName, pb.DisplayName)
	}
	if strings.Contains(names, "flowerstudio") {
		t.Errorf("florist ended up in the couple: %q + %q", pa.DisplayName, pb.DisplayName)
	}
}

// TestRepeatedCooccurrenceEvidence drives the spec's +10: the same pair
// referenced together by TWO different source accounts earns the repeated
// co-occurrence evidence kind; one source posting twice does not.
func TestRepeatedCooccurrenceEvidence(t *testing.T) {
	s := storetest.Open(t)
	orch := pipeline.New(s, llm.NewTemplateInterpreter())
	ctx := context.Background()

	for _, src := range []struct{ handle, class string }{
		{"photog_one", "engagement_photographer"},
		{"photog_two", "proposal_planner"},
	} {
		if _, err := s.AddWatchedSource(src.handle, src.class); err != nil {
			t.Fatalf("add source %s: %v", src.handle, err)
		}
	}

	mkPost := func(id, author, caption string) watchtower.RawEvent {
		return watchtower.RawEvent{
			Monitor:         "vendor:" + author,
			Source:          "apify",
			ExternalEventID: id,
			Handle:          author,
			Type:            "post",
			OccurredAt:      time.Now().UTC(),
			Payload: map[string]any{
				"caption":             caption,
				"tags":                []any{"mayak", "jordanl"},
				"source_account_type": "engagement_photographer",
			},
		}
	}

	// Post 1 from photog_one creates the pair.
	if _, err := orch.ProcessEvent(ctx, mkPost("co_1", "photog_one", "She said yes! @mayak and @jordanl #JustEngaged")); err != nil {
		t.Fatalf("post 1: %v", err)
	}
	// Post 2 from the SAME source: shared posts = 2 but distinct sources = 1
	// — no co-occurrence evidence yet.
	if _, err := orch.ProcessEvent(ctx, mkPost("co_2", "photog_one", "More of @mayak and @jordanl's engagement session #JustEngaged")); err != nil {
		t.Fatalf("post 2: %v", err)
	}
	// Post 3 from a DIFFERENT source: distinct sources = 2 — evidence fires.
	if _, err := orch.ProcessEvent(ctx, mkPost("co_3", "photog_two", "Loved planning this proposal for @mayak and @jordanl #JustEngaged")); err != nil {
		t.Fatalf("post 3: %v", err)
	}

	acctA, err := s.GetAccountByHandle("instagram", "mayak")
	if err != nil {
		t.Fatalf("account A: %v", err)
	}
	acctB, err := s.GetAccountByHandle("instagram", "jordanl")
	if err != nil {
		t.Fatalf("account B: %v", err)
	}
	cooc, err := s.GetPairCooccurrence(acctA.ID, acctB.ID)
	if err != nil {
		t.Fatalf("get co-occurrence: %v", err)
	}
	if cooc.SharedPosts != 3 {
		t.Errorf("shared posts = %d, want 3", cooc.SharedPosts)
	}
	if cooc.DistinctSources != 2 {
		t.Errorf("distinct sources = %d, want 2", cooc.DistinctSources)
	}

	couple, err := s.GetCoupleForAccountPair(acctA.ID, acctB.ID)
	if err != nil {
		t.Fatalf("couple lookup: %v", err)
	}
	hyp, err := s.LatestHypothesisForCouple(couple.ID)
	if err != nil {
		t.Fatalf("hypothesis: %v", err)
	}
	ev, err := s.EvidenceForHypothesis(hyp.ID)
	if err != nil {
		t.Fatalf("evidence: %v", err)
	}
	found := false
	for _, e := range ev {
		if e.Kind == "repeated_cooccurrence" {
			found = true
			if e.Weight <= 0 {
				t.Errorf("co-occurrence evidence weight = %v, want positive", e.Weight)
			}
		}
	}
	if !found {
		t.Error("expected repeated_cooccurrence evidence after two independent sources")
	}
}

// TestVendorInPairPenalty: if a vendor account somehow ends up inside a
// pair (e.g. paired before it was added to the registry), the −30
// vendor-in-pair contradiction lands on the score.
func TestVendorInPairPenalty(t *testing.T) {
	s := storetest.Open(t)
	orch := pipeline.New(s, llm.NewTemplateInterpreter())
	ctx := context.Background()

	if _, err := s.AddWatchedSource("photog_x", "engagement_photographer"); err != nil {
		t.Fatalf("add source: %v", err)
	}

	post := watchtower.RawEvent{
		Monitor:         "vendor:photog_x",
		Source:          "apify",
		ExternalEventID: "vip_1",
		Handle:          "photog_x",
		Type:            "post",
		OccurredAt:      time.Now().UTC(),
		Payload: map[string]any{
			"caption":             "She said yes! Congrats @janek and @tomm #JustEngaged",
			"tags":                []any{"janek", "tomm"},
			"source_account_type": "engagement_photographer",
		},
	}
	if _, err := orch.ProcessEvent(ctx, post); err != nil {
		t.Fatalf("process: %v", err)
	}

	//tomm turns out to be a vendor — he joins the registry AFTER pairing.
	if _, err := s.AddWatchedSource("tomm", "wedding_venue"); err != nil {
		t.Fatalf("register tomm as vendor: %v", err)
	}

	// A corroborating post re-touches the hypothesis, re-deriving evidence.
	post2 := post
	post2.ExternalEventID = "vip_2"
	post2.Payload["caption"] = "Another angle of @janek and @tomm's big moment #JustEngaged"
	if _, err := orch.ProcessEvent(ctx, post2); err != nil {
		t.Fatalf("process 2: %v", err)
	}

	acctA, _ := s.GetAccountByHandle("instagram", "janek")
	acctB, _ := s.GetAccountByHandle("instagram", "tomm")
	couple, err := s.GetCoupleForAccountPair(acctA.ID, acctB.ID)
	if err != nil {
		t.Fatalf("couple: %v", err)
	}
	hyp, err := s.LatestHypothesisForCouple(couple.ID)
	if err != nil {
		t.Fatalf("hypothesis: %v", err)
	}
	ev, err := s.EvidenceForHypothesis(hyp.ID)
	if err != nil {
		t.Fatalf("evidence: %v", err)
	}
	for _, e := range ev {
		if e.Kind == "vendor_in_pair" {
			if e.Weight >= 0 {
				t.Errorf("vendor_in_pair weight = %v, want negative", e.Weight)
			}
			return
		}
	}
	t.Error("expected vendor_in_pair penalty evidence")
}
