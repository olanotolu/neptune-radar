package connectors

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// BulletinDiscoveryConnector fetches a parish's real official website and
// looks for a real link to its bulletin archive. It returns a discovered URL
// only when an actual link is found on the actual page — it never guesses
// or constructs a plausible-looking URL from the parish name.
type BulletinDiscoveryConnector struct {
	Client *http.Client
}

func NewBulletinDiscoveryConnector() *BulletinDiscoveryConnector {
	return &BulletinDiscoveryConnector{Client: &http.Client{Timeout: 15 * time.Second}}
}

var bulletinLinkPattern = regexp.MustCompile(`(?is)<a[^>]+href=["']([^"']+)["'][^>]*>[^<]*bulletin[^<]*</a>`)

// DiscoverBulletinURL fetches parishURL and returns the first link whose
// anchor text mentions "bulletin", or "" if the real page has no such link.
func (c *BulletinDiscoveryConnector) DiscoverBulletinURL(ctx context.Context, parishURL string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, parishURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "NeptuneRadar-SourceHealthCheck/1.0 (+https://meetneptune.com; bulletin discovery)")
	resp, err := c.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 512*1024))
	if err != nil {
		return "", err
	}
	m := bulletinLinkPattern.FindSubmatch(body)
	if m == nil {
		return "", nil
	}
	link := string(m[1])
	if strings.HasPrefix(link, "/") {
		if u, err := url.Parse(parishURL); err == nil {
			link = u.Scheme + "://" + u.Host + link
		}
	}
	return link, nil
}

// CheckHealth treats "the bulletin archive URL is reachable" as the health
// signal — discovery of the URL itself happens once via DiscoverBulletinURL,
// not on every health check.
func (c *BulletinDiscoveryConnector) CheckHealth(ctx context.Context, endpointURL string) CheckResult {
	return NewHTTPHealthConnector().CheckHealth(ctx, endpointURL)
}
