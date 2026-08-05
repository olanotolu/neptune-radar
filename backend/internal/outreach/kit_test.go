package outreach

import (
	"strings"
	"testing"

	"neptune-social-radar/backend/internal/store"
)

// TestQRRedirectURL verifies the postcard QR encodes /r/{code} (the scan-tracking
// redirect) instead of the direct CelebrateURL. Old postcards with direct URLs
// still work — this only checks that new renders use the redirect path.
func TestQRRedirectURL(t *testing.T) {
	kit := store.CongratulateKit{
		CelebrateURL: "https://app.meetneptune.com/chat?utm_source=neptune_radar&utm_medium=postcard&utm_campaign=celebrate_first&utm_content=cpl_abc&ref=deadbeef",
	}
	html := RenderPostcardHTML(kit)
	// The QR image data= param should encode /r/deadbeef, not the full celebrate URL.
	// url.QueryEscape encodes / as %2F, so we check for the handoff code + redirect base.
	qrIdx := strings.Index(html, "api.qrserver.com")
	if qrIdx == -1 {
		t.Fatalf("no QR image in rendered postcard")
	}
	// Extract just the QR img src (ends at the next quote after the data= URL).
	qrEnd := strings.Index(html[qrIdx:], "\"")
	if qrEnd == -1 {
		t.Fatalf("malformed QR img src")
	}
	qrSrc := html[qrIdx : qrIdx+qrEnd]
	if !strings.Contains(qrSrc, "deadbeef") {
		t.Errorf("QR should contain handoff code deadbeef, got: %s", qrSrc)
	}
	if !strings.Contains(qrSrc, "radar.meetneptune.com") {
		t.Errorf("QR should use redirect base radar.meetneptune.com, got: %s", qrSrc)
	}
	if strings.Contains(qrSrc, "app.meetneptune.com") {
		t.Errorf("QR should not encode the direct celebrate URL — it should use /r/{code}")
	}
}
