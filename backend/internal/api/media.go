package api

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Allowed media hosts for the image proxy — Instagram / Meta CDNs only.
// Matching is exact or dot-suffixed subdomains ONLY. A substring match here
// once made this an unauthenticated open proxy: "cdninstagram.com.evil.com"
// and "scontent.attacker.io" both passed validation.
var allowedMediaHosts = []string{
	"cdninstagram.com",
	"fbcdn.net",
	"instagram.com",
}

var mediaClient = &http.Client{Timeout: 25 * time.Second}

func hostAllowed(host string) bool {
	host = strings.ToLower(host)
	for _, h := range allowedMediaHosts {
		if host == h || strings.HasSuffix(host, "."+h) {
			return true
		}
	}
	return false
}

// mediaProxy streams Instagram CDN images through our origin so the browser
// is not blocked by Cross-Origin-Resource-Policy: same-origin on fbcdn.
// GET /api/media?url=<encoded absolute URL>
func (s *Server) mediaProxy(w http.ResponseWriter, r *http.Request) {
	raw := r.URL.Query().Get("url")
	if raw == "" {
		writeError(w, http.StatusBadRequest, errorString("url required"))
		return
	}
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" {
		writeError(w, http.StatusBadRequest, errorString("invalid url"))
		return
	}
	if !hostAllowed(u.Host) {
		writeError(w, http.StatusForbidden, errorString("host not allowed"))
		return
	}

	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, u.String(), nil)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	// Instagram is picky about UA; no Referer avoids hotlink denials.
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; NeptuneRadar/1.0)")
	req.Header.Set("Accept", "image/avif,image/webp,image/apng,image/*,*/*;q=0.8")

	resp, err := mediaClient.Do(req)
	if err != nil {
		writeError(w, http.StatusBadGateway, fmt.Errorf("fetch media: %w", err))
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		writeError(w, http.StatusBadGateway, fmt.Errorf("upstream %d", resp.StatusCode))
		return
	}

	ct := resp.Header.Get("Content-Type")
	if ct == "" || !strings.HasPrefix(ct, "image/") {
		// Some edges return octet-stream; still serve if body looks ok
		if ct == "" {
			ct = "image/jpeg"
		}
	}
	w.Header().Set("Content-Type", ct)
	w.Header().Set("Cache-Control", "public, max-age=86400")
	// Open CORP so the dashboard origin can paint the image
	w.Header().Set("Cross-Origin-Resource-Policy", "cross-origin")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if cl := resp.Header.Get("Content-Length"); cl != "" {
		w.Header().Set("Content-Length", cl)
	}
	w.WriteHeader(http.StatusOK)
	_, _ = io.Copy(w, resp.Body)
}
