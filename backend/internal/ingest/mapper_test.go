package ingest

import (
	"encoding/json"
	"testing"
)

func TestMapPost_CollabFields(t *testing.T) {
	// Apify actors have shipped the co-author under different field names;
	// each must land in payload["collab_with"] for the signals package.
	cases := []struct {
		name string
		item string
		want string
	}{
		{"coauthorProducers", `{"id":"1","ownerUsername":"janedoe","coauthorProducers":[{"username":"alexsmith"}]}`, "alexsmith"},
		{"coauthors", `{"id":"2","ownerUsername":"janedoe","coauthors":[{"username":"alexsmith"}]}`, "alexsmith"},
		{"collabWith", `{"id":"3","ownerUsername":"janedoe","collabWith":"@alexsmith"}`, "alexsmith"},
		{"no collab", `{"id":"4","ownerUsername":"janedoe"}`, ""},
		{"self collab ignored", `{"id":"5","ownerUsername":"janedoe","coauthors":[{"username":"janedoe"}]}`, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			raw, _, err := MapPost(json.RawMessage(tc.item), "m")
			if err != nil {
				t.Fatalf("map: %v", err)
			}
			got, _ := raw.Payload["collab_with"].(string)
			if got != tc.want {
				t.Errorf("collab_with = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestMapPost_FullItem(t *testing.T) {
	item := json.RawMessage(`{
		"id": "3649123456789012345",
		"shortCode": "CxyzABC123",
		"url": "https://www.instagram.com/p/CxyzABC123/",
		"caption": "She said yes! @mayak @jordanl #JustEngaged",
		"ownerUsername": "weddingsbynoor",
		"hashtags": ["JustEngaged"],
		"mentions": ["mayak", "jordanl"],
		"taggedUsers": [{"username": "mayak"}, {"username": "jordanl"}],
		"locationName": "Central Park",
		"timestamp": "2026-06-03T18:30:00.000Z",
		"displayUrl": "https://cdn.example.com/img.jpg",
		"type": "Image"
	}`)
	raw, imageURL, err := MapPost(item, "vendor:weddingsbynoor")
	if err != nil {
		t.Fatalf("map: %v", err)
	}
	if raw.ExternalEventID != "3649123456789012345" {
		t.Errorf("external id: %q", raw.ExternalEventID)
	}
	if raw.Handle != "weddingsbynoor" {
		t.Errorf("handle: %q", raw.Handle)
	}
	if raw.Monitor != "vendor:weddingsbynoor" || raw.Source != "apify" {
		t.Errorf("monitor/source: %q %q", raw.Monitor, raw.Source)
	}
	if raw.OccurredAt.Year() != 2026 || raw.OccurredAt.Month() != 6 {
		t.Errorf("occurred_at: %v", raw.OccurredAt)
	}
	tags, _ := raw.Payload["tags"].([]string)
	if len(tags) != 2 || tags[0] != "mayak" {
		t.Errorf("tags: %v", raw.Payload["tags"])
	}
	if raw.Payload["location"] != "Central Park" {
		t.Errorf("location: %v", raw.Payload["location"])
	}
	if imageURL != "https://cdn.example.com/img.jpg" {
		t.Errorf("image url: %q", imageURL)
	}
}

func TestMapPost_SchemaDriftFallbacks(t *testing.T) {
	// Older actor shape: string hashtags, no taggedUsers, unix timestamp,
	// shortCode only.
	item := json.RawMessage(`{
		"shortCode": "Drift99",
		"caption": "I said yes!",
		"ownerUsername": "priya",
		"takenAtTimestamp": 1780000000
	}`)
	raw, _, err := MapPost(item, "hashtag:isaidyes")
	if err != nil {
		t.Fatalf("map: %v", err)
	}
	if raw.ExternalEventID != "Drift99" {
		t.Errorf("fallback id: %q", raw.ExternalEventID)
	}
	if raw.OccurredAt.Unix() != 1780000000 {
		t.Errorf("unix timestamp fallback: %v", raw.OccurredAt)
	}
}

func TestMapPost_RejectsNonPost(t *testing.T) {
	if _, _, err := MapPost(json.RawMessage(`{"error": "Private profile"}`), "m"); err == nil {
		t.Error("expected an error for an error-shaped item")
	}
}

func TestParseProfileBio(t *testing.T) {
	if bio, ok := ParseProfileBio(json.RawMessage(`{"biography": "fiancée of jordan 💍"}`)); !ok || bio != "fiancée of jordan 💍" {
		t.Errorf("biography: %q ok=%v", bio, ok)
	}
	if _, ok := ParseProfileBio(json.RawMessage(`{"error": "not found"}`)); ok {
		t.Error("error item should not parse")
	}
	if bio, ok := ParseProfileBio(json.RawMessage(`{"biography": ""}`)); !ok || bio != "" {
		t.Errorf("empty bio is a valid observation: %q ok=%v", bio, ok)
	}
}

func TestParseFollowingUsernames(t *testing.T) {
	items := []json.RawMessage{
		json.RawMessage(`{"username": "mayak"}`),
		json.RawMessage(`{"username": "jordanl"}`),
		json.RawMessage(`{"no_username": true}`),
	}
	got := ParseFollowingUsernames(items)
	if len(got) != 2 || got[0] != "mayak" {
		t.Errorf("following: %v", got)
	}
}

func TestAttributeHashtagMonitor(t *testing.T) {
	m := attributeHashtagMonitor(map[string]any{"caption": "She said yes! #JustEngaged"})
	if m != "hashtag:justengaged" {
		t.Errorf("monitor attribution: %q", m)
	}
	if m := attributeHashtagMonitor(map[string]any{"caption": "nice day"}); m != "hashtag:batch" {
		t.Errorf("unattributed monitor: %q", m)
	}
}
