package outreach

import (
	"strings"
	"testing"

	neptunestore "neptune-social-radar/backend/internal/store"
)

func TestRenderPostcardHTML(t *testing.T) {
	k := neptunestore.CongratulateKit{
		PersonAName:      "Alice Johnson",
		PersonBName:      "Bob Smith",
		Headline:         "Congratulations",
		BodyMessage:      "So happy for you both!\nWishing you a lifetime of joy.",
		MarketCity:       "Portland",
		MarketRegion:     "OR",
		AddressLine1:     "123 Main St",
		AddressCity:      "Portland",
		AddressRegion:    "OR",
		AddressPostal:    "97201",
		DiscoveryImageURL: "https://example.com/photo.jpg",
	}
	html := RenderPostcardHTML(k)
	checks := []struct{ name, want string }{
		{"headline", "Congratulations"},
		{"first names", "Alice &amp; Bob"},
		{"body message", "So happy for you both!"},
		{"location", "Portland, OR"},
		{"address", "123 Main St"},
		{"black bg", "#0a0a0a"},
		{"geist font", `"Geist"`},
		{"serif for names", `"Iowan Old Style"`},
		{"mono for labels", `"Geist Mono"`},
		{"grayscale filter", "grayscale(100%)"},
		{"deliver to label", "Deliver to"},
		{"stamp area", "pc-back__stamp"},
		{"hairline rule", "pc-front__rule"},
	}
	for _, c := range checks {
		if !strings.Contains(html, c.want) {
			t.Errorf("postcard HTML missing %s: want %q", c.name, c.want)
		}
	}
	// Should NOT contain old warm palette colors
	banned := []string{"#1a3a3c", "#0f2628", "#f4f1ea", "#fdfbf7", "#57534e", "#78716c", "#a8a29e"}
	for _, b := range banned {
		if strings.Contains(html, b) {
			t.Errorf("postcard HTML still contains old warm color %q", b)
		}
	}
}

func TestRenderPostcardHTML_EmptyAddress(t *testing.T) {
	k := neptunestore.CongratulateKit{
		PersonAName: "Alice",
		PersonBName: "Bob",
	}
	html := RenderPostcardHTML(k)
	if !strings.Contains(html, "pc-back__addr-pending") {
		t.Error("empty address should use pc-back__addr-pending class")
	}
	if !strings.Contains(html, "Address pending human verification") {
		t.Error("empty address should show pending message")
	}
}

func TestRenderPostcardHTML_CelebrateQR(t *testing.T) {
	k := neptunestore.CongratulateKit{
		PersonAName:  "Alice",
		PersonBName:  "Bob",
		CelebrateURL: "https://app.meetneptune.com/chat?utm_medium=postcard&ref=abc",
	}
	html := RenderPostcardHTML(k)
	if !strings.Contains(html, "pc-back__qr") {
		t.Error("celebrate URL should render QR block")
	}
	if !strings.Contains(html, "meetneptune.com") {
		t.Error("celebrate QR should brand meetneptune.com")
	}
	if !strings.Contains(html, "utm_medium=postcard") {
		t.Error("QR should preserve celebrate tracking URL")
	}
}
