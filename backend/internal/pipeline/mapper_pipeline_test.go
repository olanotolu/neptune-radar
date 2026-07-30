package pipeline_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"neptune-social-radar/backend/internal/ingest"
	"neptune-social-radar/backend/internal/llm"
	"neptune-social-radar/backend/internal/pipeline"
	"neptune-social-radar/backend/internal/store/storetest"
)

// TestMapperOutputDrivesPipeline is the regression test for the []string vs
// []any kill chain: every other test hand-builds []any payloads, but the real
// mapper emits []string — and for months that mismatch silently meant zero
// tag edges from real provider posts. This test pipes MapPost's ACTUAL output
// into ProcessEvent and asserts the tag path works end to end.
func TestMapperOutputDrivesPipeline(t *testing.T) {
	s := storetest.Open(t)
	orch := pipeline.New(s, llm.NewTemplateInterpreter())
	ctx := context.Background()

	if _, err := s.AddWatchedSource("columbusproposalphoto", "engagement_photographer"); err != nil {
		t.Fatalf("add source: %v", err)
	}

	item := json.RawMessage(`{
		"id": "3649123456789012346",
		"shortCode": "CxyzREAL1",
		"url": "https://www.instagram.com/p/CxyzREAL1/",
		"caption": "She said yes! Congrats @sarab and @mikeT #JustEngaged",
		"ownerUsername": "columbusproposalphoto",
		"hashtags": ["JustEngaged"],
		"mentions": ["sarab", "mikeT"],
		"taggedUsers": [{"username": "sarab"}, {"username": "mikeT"}],
		"locationName": "Columbus, OH",
		"timestamp": "2026-07-29T18:30:00.000Z",
		"type": "Image"
	}`)
	raw, _, err := ingest.MapPost(item, "vendor:columbusproposalphoto")
	if err != nil {
		t.Fatalf("map: %v", err)
	}
	raw.OccurredAt = time.Now().UTC()

	res, err := orch.ProcessEvent(ctx, raw)
	if err != nil {
		t.Fatalf("process: %v", err)
	}
	if res.HypothesisID == "" {
		t.Fatalf("expected a hypothesis from a mapper-produced post, got %+v", res)
	}

	// The tag edges must exist — this is the exact thing that was silently
	// broken: tagged_with edges from a real provider payload.
	acct, err := s.GetAccountByHandle("instagram", "sarab")
	if err != nil {
		t.Fatalf("sarab account: %v", err)
	}
	edges, err := s.EdgesForAccount(acct.ID)
	if err != nil {
		t.Fatalf("edges: %v", err)
	}
	tagFound := false
	for _, e := range edges {
		if e.Kind == "tagged_with" {
			tagFound = true
		}
	}
	if !tagFound {
		t.Error("no tagged_with edge created from real mapper output — the []string/[]any regression is back")
	}
}
