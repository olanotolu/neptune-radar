package records

import (
	"testing"
)

// TestParseWeddingWebsiteHTML verifies the conservative HTML extractor: both
// names must match, date/venue/registry are parsed, and a single-name page is
// rejected. ponytail: one test covers the parsing logic — the live scrapers are
// network-dependent and not unit-tested.
func TestParseWeddingWebsiteHTML(t *testing.T) {
	html := `<html><body>
		<h1>Jane Smith &amp; Alex Doe are getting married!</h1>
		<div class="event-date">October 12, 2025</div>
		<div>Venue: The Grand Ballroom</div>
		<a href="https://www.zola.com/registry/jane-alex">Registry</a>
		<a href="https://www.amazon.com/wedding/jane-alex">Amazon</a>
	</body></html>`

	hits := parseWeddingWebsiteHTML(html, "knot", "https://example.com/p", "Jane", "Smith", "Alex", "Doe")
	if len(hits) != 1 {
		t.Fatalf("expected 1 hit, got %d", len(hits))
	}
	h := hits[0]
	if h.Platform != "knot" {
		t.Errorf("platform: want knot, got %s", h.Platform)
	}
	if h.WeddingDate != "2025-10-12" {
		t.Errorf("wedding_date: want 2025-10-12, got %q", h.WeddingDate)
	}
	if h.VenueName != "The Grand Ballroom" {
		t.Errorf("venue: want %q, got %q", "The Grand Ballroom", h.VenueName)
	}
	if len(h.RegistryURLs) != 2 {
		t.Errorf("registry_urls: want 2, got %d", len(h.RegistryURLs))
	}

	// Conservative guard: only one name present → no hit.
	miss := parseWeddingWebsiteHTML(html, "knot", "https://example.com/p", "Jane", "Smith", "Nobody", "Here")
	if len(miss) != 0 {
		t.Errorf("expected 0 hits when only one name matches, got %d", len(miss))
	}

	// ISO date passthrough.
	iso := parseWeddingWebsiteHTML(
		`<div>Jane Smith and Alex Doe — 2025-10-12 · Venue: The Plaza</div>`,
		"zola", "https://example.com/z", "Jane", "Smith", "Alex", "Doe")
	if len(iso) != 1 || iso[0].WeddingDate != "2025-10-12" {
		t.Fatalf("iso date passthrough: got %d hits, date %q", len(iso), func() string {
			if len(iso) > 0 {
				return iso[0].WeddingDate
			}
			return ""
		}())
	}
}
