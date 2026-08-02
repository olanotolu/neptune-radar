// Command verify-sources re-checks that vendor Instagram handles and government
// URLs in the Neptune Radar source registry are still valid. It is read-only:
// no DB writes, just a report to stdout.
//
// Usage:
//
//	go run ./cmd/verify-sources -states=TX,FL,NY
//	go run ./cmd/verify-sources -all
//	go run ./cmd/verify-sources -states=TX -layer=vendors -stale-days=180
//
// Output is tab-separated: state, source_id, status, last_verified, url, note.
// Statuses: STALE, HEALTHY, DEGRADED, OFFLINE.
package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"neptune-social-radar/backend/internal/packs"
)

const userAgent = "NeptuneRadar-SourceReverify/1.0 (+https://meetneptune.com; source re-verification probe)"

var allStates = []string{
	"AK", "AL", "AR", "AZ", "CA", "CO", "CT", "DC", "DE", "FL",
	"GA", "HI", "IA", "ID", "IL", "IN", "KS", "KY", "LA", "MA",
	"MD", "ME", "MI", "MN", "MO", "MS", "MT", "NC", "ND", "NE",
	"NH", "NJ", "NM", "NV", "NY", "OH", "OK", "OR", "PA", "RI",
	"SC", "SD", "TN", "TX", "UT", "VA", "VT", "WA", "WI", "WV", "WY",
}

const requestDelay = 1500 * time.Millisecond

func main() {
	statesFlag := flag.String("states", "", "comma-separated USPS codes to verify (TX,FL,NY)")
	allFlag := flag.Bool("all", false, "verify all 51 jurisdictions")
	staleDaysFlag := flag.Int("stale-days", 90, "flag sources older than this many days")
	layerFlag := flag.String("layer", "vendors,gov,church", "comma-separated layers: vendors,gov,church")
	flag.Parse()

	var states []string
	if *allFlag {
		states = allStates
	} else if *statesFlag != "" {
		for _, p := range strings.Split(*statesFlag, ",") {
			p = strings.ToUpper(strings.TrimSpace(p))
			if p != "" {
				states = append(states, p)
			}
		}
	}
	if len(states) == 0 {
		fmt.Fprintln(os.Stderr, "usage: verify-sources -states=TX,FL,NY | -all")
		os.Exit(1)
	}

	layers := map[string]bool{}
	for _, l := range strings.Split(*layerFlag, ",") {
		l = strings.ToLower(strings.TrimSpace(l))
		if l != "" {
			layers[l] = true
		}
	}

	client := &http.Client{Timeout: 15 * time.Second}
	staleThreshold := time.Now().AddDate(0, 0, -*staleDaysFlag)

	fmt.Println("state\tsource_id\tstatus\tlast_verified\turl\tnote")

	for _, st := range states {
		pack := packs.PackFor(st)
		if pack == nil {
			fmt.Printf("%s\t\tOFFLINE\t\t\tno pack defined for state\n", st)
			continue
		}
		if layers["vendors"] {
			verifyVendors(st, pack.Vendors, client, staleThreshold)
		}
		if layers["gov"] {
			verifyGov(st, pack.Government, client)
		}
		if layers["church"] {
			verifyChurch(st, pack, client)
		}
	}
}

// verifyVendors re-checks each vendor's OfficialURL for the presence of the
// registered Instagram handle. If the page is reachable but the handle is no
// longer present, the source is DEGRADED. If the page is unreachable, OFFLINE.
// Sources whose Verified date is older than the stale threshold are STALE
// (the URL is still checked and the result reported, but STALE takes
// precedence as the headline status).
func verifyVendors(state string, vendors []packs.VendorDef, client *http.Client, staleThreshold time.Time) {
	for _, v := range vendors {
		sourceID := "vendor_" + strings.ToLower(state) + "_" + v.Handle
		lastVerified := v.Verified
		url := v.OfficialURL

		stale := false
		if verified, err := time.Parse("2006-01-02", v.Verified); err == nil {
			if verified.Before(staleThreshold) {
				stale = true
			}
		}

		// If the OfficialURL is itself an Instagram URL, just health-check it.
		isIGURL := strings.Contains(url, "instagram.com/")

		status, body, note := httpGet(client, url)
		time.Sleep(requestDelay)

		if status == "OFFLINE" {
			printRow(state, sourceID, "OFFLINE", lastVerified, url, note)
			continue
		}

		if isIGURL {
			// ponytail: for vendors whose only URL is their IG profile, we
			// can't check whether the handle is "still present" — the page IS
			// the profile. Report HEALTHY if reachable.
			if stale {
				printRow(state, sourceID, "STALE", lastVerified, url, "reachable; verification date expired")
			} else {
				printRow(state, sourceID, "HEALTHY", lastVerified, url, "reachable")
			}
			continue
		}

		// Check if the IG handle is still present on the page.
		handleFound := igHandleOnPage(body, v.Handle)

		switch {
		case stale && handleFound:
			printRow(state, sourceID, "STALE", lastVerified, url, "reachable; handle present; verification date expired")
		case stale && !handleFound:
			printRow(state, sourceID, "STALE", lastVerified, url, "reachable; handle NOT found on page")
		case !stale && handleFound:
			printRow(state, sourceID, "HEALTHY", lastVerified, url, "reachable; handle present")
		case !stale && !handleFound:
			printRow(state, sourceID, "DEGRADED", lastVerified, url, "reachable; handle NOT found on page")
		}
	}
}

// verifyGov runs an HTTP health check against each government SearchURL.
func verifyGov(state string, sources []packs.GovSource, client *http.Client) {
	for _, g := range sources {
		sourceID := "gov_county_" + g.CountyFIPS
		url := g.SearchURL

		status, _, note := httpGet(client, url)
		time.Sleep(requestDelay)

		switch status {
		case "OFFLINE":
			printRow(state, sourceID, "OFFLINE", "unknown", url, note)
		default:
			printRow(state, sourceID, "HEALTHY", "unknown", url, "reachable")
		}
	}
}

// verifyChurch runs HTTP health checks against each diocese Directory URL and
// each parish BulletinURL.
func verifyChurch(state string, pack *packs.StatePack, client *http.Client) {
	parishesByDiocese := make(map[string][]packs.ParishDef)
	for _, p := range pack.Parishes {
		parishesByDiocese[p.DioceseSlug] = append(parishesByDiocese[p.DioceseSlug], p)
	}

	for _, d := range pack.Dioceses {
		sourceID := "diocese_" + strings.ToLower(state) + "_" + d.Slug
		url := d.Directory

		status, _, note := httpGet(client, url)
		time.Sleep(requestDelay)

		switch status {
		case "OFFLINE":
			printRow(state, sourceID, "OFFLINE", "unknown", url, note)
		default:
			printRow(state, sourceID, "HEALTHY", "unknown", url, "reachable")
		}

		// Parishes for this diocese.
		for i, p := range parishesByDiocese[d.Slug] {
			if p.BulletinURL == "" {
				continue
			}
			sourceID := fmt.Sprintf("parish_%s_%s_%02d", strings.ToLower(state), d.Slug, i+1)
			url := p.BulletinURL

			status, _, note := httpGet(client, url)
			time.Sleep(requestDelay)

			switch status {
			case "OFFLINE":
				printRow(state, sourceID, "OFFLINE", "unknown", url, note)
			default:
				printRow(state, sourceID, "HEALTHY", "unknown", url, "reachable")
			}
		}
	}
}

// httpGet performs a GET request and returns a status string ("OK" or
// "OFFLINE"), the response body (up to 256 KB), and a note explaining the
// outcome.
func httpGet(client *http.Client, url string) (status string, body string, note string) {
	if url == "" {
		return "OFFLINE", "", "empty URL"
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "OFFLINE", "", "invalid URL: " + err.Error()
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req)
	if err != nil {
		return "OFFLINE", "", "request failed: " + err.Error()
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(io.LimitReader(resp.Body, 256*1024))

	if resp.StatusCode >= 400 {
		return "OFFLINE", string(data), fmt.Sprintf("HTTP %d", resp.StatusCode)
	}
	return "OK", string(data), ""
}

// igHandleOnPage checks whether the Instagram handle appears on the page.
// It looks for "instagram.com/{handle}" (with or without trailing slash) to
// avoid false positives from the handle appearing in unrelated text.
func igHandleOnPage(body, handle string) bool {
	if handle == "" {
		return false
	}
	lower := strings.ToLower(body)
	handle = strings.ToLower(handle)
	// Match instagram.com/{handle} with optional trailing slash or path.
	needle := "instagram.com/" + handle
	if strings.Contains(lower, needle) {
		return true
	}
	// Some sites use www.instagram.com.
	needle = "www.instagram.com/" + handle
	return strings.Contains(lower, needle)
}

func printRow(state, sourceID, status, lastVerified, url, note string) {
	fmt.Printf("%s\t%s\t%s\t%s\t%s\t%s\n", state, sourceID, status, lastVerified, url, note)
}
