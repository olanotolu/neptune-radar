// Command bootstrap-state seeds city markets + government marriage-record
// sources + Catholic dioceses/parishes + public wedding-industry social
// sources for one or more USPS states (national radar expansion).
//
// Prerequisites: run seed-geography first so states/counties exist.
//
// Usage:
//
//	DATABASE_URL=… go run ./cmd/bootstrap-state -states=NY,CA
//	DATABASE_URL=… go run ./cmd/bootstrap-state -states=TX   # single state
//
// Idempotent. Optional APIFY_TOKEN / Bright Data enables live Instagram health
// checks; without a provider, connectors stay at whatever RecordConnectorRun writes.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"neptune-social-radar/backend/internal/packs"
	"neptune-social-radar/backend/internal/connectors"
	"neptune-social-radar/backend/internal/ingest"
	"neptune-social-radar/backend/internal/ontology"
	"neptune-social-radar/backend/internal/store"
)

func main() {
	statesFlag := flag.String("states", "NY,CA", "comma-separated USPS codes to bootstrap (NY,CA)")
	flag.Parse()

	var states []string
	for _, p := range strings.Split(*statesFlag, ",") {
		p = strings.ToUpper(strings.TrimSpace(p))
		if p != "" {
			states = append(states, p)
		}
	}
	if len(states) == 0 {
		log.Fatal("no states given")
	}

	dsn := os.Getenv("DATABASE_URL")
	s, err := store.Open(dsn)
	if err != nil {
		log.Fatalf("open store: %v", err)
	}
	defer s.Close()

	ctx := context.Background()
	apify := ingest.NewApifyClient(os.Getenv("APIFY_TOKEN"))
	httpHealth := connectors.NewHTTPHealthConnector()

	totalGov := 0
	totalChurch := 0
	totalVendors := 0
	for _, st := range states {
		// Ensure state row exists (seed-geography may not have run).
		if _, err := s.GetState(st); err != nil {
			log.Fatalf("state %s not in registry — run seed-geography first: %v", st, err)
		}
		pack := packs.PackFor(st)
		if pack == nil {
			log.Fatalf("no pack defined for state %s", st)
		}
		log.Printf("[bootstrap-state] %s…", st)

		// Cities
		for _, c := range pack.Cities {
			lat, lng := c.Lat, c.Lng
			must(s.UpsertCity(c.ID, c.State, c.County, c.Name, &lat, &lng))
			log.Printf("[bootstrap-state]   city %s (%s)", c.Name, c.ID)
		}

		// Government
		nGov := bootstrapGovernment(ctx, s, httpHealth, pack)
		totalGov += nGov

		// Church (Catholic)
		nChurch := bootstrapChurch(ctx, s, httpHealth, pack)
		// Church (Episcopal — first non-Catholic layer)
		nEpis := bootstrapEpiscopal(ctx, s, httpHealth, st)
		// Church (Methodist — UMC annual conferences)
		nMeth := bootstrapMethodist(ctx, s, httpHealth, st)
		// Church (Jewish — federations and community organizations)
		nJew := bootstrapJewish(ctx, s, httpHealth, st)
		totalChurch += nChurch + nEpis + nMeth + nJew

		// Social
		nVendors := bootstrapSocial(ctx, s, apify, pack.Vendors)
		totalVendors += nVendors

		log.Printf("[bootstrap-state]   %s: %d cities, %d gov, %d church (+%d episcopal, +%d methodist, +%d jewish), %d vendors",
			st, len(pack.Cities), nGov, nChurch, nEpis, nMeth, nJew, nVendors)
	}

	markets := packs.MarketsForStates(states)
	if _, err := s.Audit("bootstrap", "states", "bootstrap_state_completed", map[string]any{
		"states":         states,
		"gov_sources":    totalGov,
		"church_sources": totalChurch,
		"vendors":        totalVendors,
		"markets":        markets,
		"apify_set":      apify.Available(),
		"suggested_env":  "ACTIVE_MARKETS append: " + strings.Join(markets, ","),
	}, "", -1); err != nil {
		log.Printf("[bootstrap-state] audit failed: %v", err)
	}

	fmt.Printf("bootstrap-state ok: states=%v gov=%d church=%d vendors=%d\n",
		states, totalGov, totalChurch, totalVendors)
	fmt.Printf("Suggested ACTIVE_MARKETS additions:\n  %s\n", strings.Join(markets, ","))
}

// bootstrapGovernment registers county marriage-record offices and runs a real
// HTTP health check against each search endpoint. Ported from bootstrap-ohio.
func bootstrapGovernment(ctx context.Context, s *store.Store, hc connectors.SourceConnector, pack *packs.StatePack) int {
	n := 0
	for _, g := range pack.Government {
		orgID := "org_gov_county_" + g.CountyFIPS
		meta, _ := json.Marshal(map[string]any{"capability_note": g.Note})
		must(s.UpsertSourceOrganization(ontology.SourceOrganization{
			ID: orgID, OrgType: "government_office", Name: g.CourtName,
			CountyID: g.CountyFIPS, OfficialURL: g.CourtURL,
			Provenance: "official_government_website", DataMode: ontology.DataModeManual, Metadata: string(meta),
		}))

		endpointID := "endpoint_gov_marriage_search_" + g.CountyFIPS
		must(s.UpsertSourceEndpoint(ontology.SourceEndpoint{
			ID: endpointID, OrganizationID: orgID, EndpointType: "marriage_record_search",
			URL: g.SearchURL, AccessMethod: "html_search_form", IsOfficial: true, DataMode: ontology.DataModeManual,
		}))

		connID := "connector_gov_marriage_search_" + g.CountyFIPS
		must(s.UpsertConnector(ontology.Connector{
			ID: connID, SourceEndpointID: endpointID, ConnectorType: "http_health", Provider: "neptune-http-health-v1",
		}))

		runCheck(s, connID, g.SearchURL, hc.CheckHealth(ctx, g.SearchURL))
		n++
	}
	return n
}

// bootstrapChurch registers Catholic dioceses/archdioceses with parish-directory
// endpoints, plus parishes (with bulletin archives where located) for the
// primary-metro diocese. Ported from bootstrap-ohio.
func bootstrapChurch(ctx context.Context, s *store.Store, hc connectors.SourceConnector, pack *packs.StatePack) int {
	return registerDioceses(ctx, s, hc, pack.State, pack.Dioceses, pack.Parishes)
}

// bootstrapEpiscopal registers Episcopal dioceses (and curated metro parishes)
// for one state — the first non-Catholic church layer. Data lives in
// episcopal_dioceses.go; generated IDs are namespaced with an "episcopal_"
// prefix so they never collide with Catholic diocese/parish rows.
func bootstrapEpiscopal(ctx context.Context, s *store.Store, hc connectors.SourceConnector, st string) int {
	return registerDioceses(ctx, s, hc, st, packs.EpiscopalDiocesesFor(st), packs.EpiscopalParishesFor(st))
}

// bootstrapMethodist registers United Methodist Church annual conferences (and
// curated metro churches) for one state. Data lives in methodist_conferences.go;
// generated IDs are namespaced with a "methodist_" prefix.
func bootstrapMethodist(ctx context.Context, s *store.Store, hc connectors.SourceConnector, st string) int {
	return registerDioceses(ctx, s, hc, st, packs.MethodistConferencesFor(st), packs.MethodistChurchesFor(st))
}

// bootstrapJewish registers Jewish federations/community organizations (and
// curated metro synagogues) for one state. Data lives in jewish_congregations.go;
// generated IDs are namespaced with a "jewish_" prefix.
func bootstrapJewish(ctx context.Context, s *store.Store, hc connectors.SourceConnector, st string) int {
	return registerDioceses(ctx, s, hc, st, packs.JewishFederationsFor(st), packs.JewishSynagoguesFor(st))
}

// registerDioceses is the shared Catholic/Episcopal diocese+parish registrar.
// Denomination is read from each DioceseDef (empty defaults to "catholic"); a
// non-catholic denomination prefixes every generated ID so the layers coexist.
// Catholic callers produce byte-identical IDs to the pre-refactor bootstrapChurch.
func registerDioceses(ctx context.Context, s *store.Store, hc connectors.SourceConnector, state string, dioceses []packs.DioceseDef, parishes []packs.ParishDef) int {
	n := 0
	st := strings.ToLower(state)

	// Index parishes by diocese slug for grouping.
	parishesByDiocese := make(map[string][]packs.ParishDef)
	for _, p := range parishes {
		parishesByDiocese[p.DioceseSlug] = append(parishesByDiocese[p.DioceseSlug], p)
	}

	for _, d := range dioceses {
		denom := d.Denomination
		if denom == "" {
			denom = "catholic"
		}
		prefix := ""
		if denom != "catholic" {
			prefix = denom + "_"
		}
		dOrgID := "org_diocese_" + prefix + st + "_" + d.Slug
		dMeta, _ := json.Marshal(map[string]any{"denomination": denom})
		must(s.UpsertSourceOrganization(ontology.SourceOrganization{
			ID: dOrgID, OrgType: "diocese", Name: d.Name, CityID: d.HubCityID, OfficialURL: d.Website,
			Provenance: "official_directory", DataMode: ontology.DataModeManual, Metadata: string(dMeta),
		}))

		dJurID := "jurisdiction_diocese_" + prefix + st + "_" + d.Slug
		must(s.UpsertChurchJurisdiction(ontology.ChurchJurisdiction{
			ID: dJurID, SourceOrganizationID: dOrgID, JurisdictionType: d.Type, HubCityID: d.HubCityID,
		}))

		dEndpointID := "endpoint_diocese_" + prefix + st + "_" + d.Slug + "_directory"
		must(s.UpsertSourceEndpoint(ontology.SourceEndpoint{
			ID: dEndpointID, OrganizationID: dOrgID, EndpointType: "parish_directory",
			URL: d.Directory, AccessMethod: "public_index", IsOfficial: true, DataMode: ontology.DataModeManual,
		}))

		dConnID := "connector_diocese_" + prefix + st + "_" + d.Slug + "_directory"
		must(s.UpsertConnector(ontology.Connector{
			ID: dConnID, SourceEndpointID: dEndpointID, ConnectorType: "http_health", Provider: "neptune-http-health-v1",
		}))
		runCheck(s, dConnID, d.Directory, hc.CheckHealth(ctx, d.Directory))
		n++

		// Parishes for this diocese (if any were curated).
		parishes := parishesByDiocese[d.Slug]
		for i, p := range parishes {
			orgID := fmt.Sprintf("org_parish_%s%s_%s_%02d", prefix, st, d.Slug, i+1)
			meta := map[string]any{"denomination": denom}
			if p.Address != "" {
				meta["address"] = p.Address
			}
			if p.BulletinURL != "" {
				meta["bulletin_url"] = p.BulletinURL
			}
			if p.GeoLat != 0 {
				meta["geo_lat"] = p.GeoLat
				meta["geo_lng"] = p.GeoLng
			}
			metaJSON, _ := json.Marshal(meta)
			parishCityID := p.CityID
			if parishCityID == "" {
				parishCityID = d.HubCityID
			}
			must(s.UpsertSourceOrganization(ontology.SourceOrganization{
				ID: orgID, OrgType: "parish", Name: p.Name, CityID: parishCityID, CountyID: p.CountyFIPS,
				Provenance: "manually_curated", DataMode: ontology.DataModeManual, Metadata: string(metaJSON),
			}))

			var bulletinEndpointID string
			if p.BulletinURL != "" {
				bulletinEndpointID = fmt.Sprintf("endpoint_parish_%s%s_%s_%02d_bulletin", prefix, st, d.Slug, i+1)
				must(s.UpsertSourceEndpoint(ontology.SourceEndpoint{
					ID: bulletinEndpointID, OrganizationID: orgID, EndpointType: "bulletin_archive",
					URL: p.BulletinURL, AccessMethod: "public_index", IsOfficial: !p.Aggregator, DataMode: ontology.DataModeManual,
				}))
				bConnID := fmt.Sprintf("connector_parish_%s%s_%s_%02d_bulletin", prefix, st, d.Slug, i+1)
				must(s.UpsertConnector(ontology.Connector{
					ID: bConnID, SourceEndpointID: bulletinEndpointID, ConnectorType: "http_health", Provider: "neptune-http-health-v1",
				}))
				runCheck(s, bConnID, p.BulletinURL, hc.CheckHealth(ctx, p.BulletinURL))
				n++
			}

			parishID := fmt.Sprintf("parish_%s%s_%s_%02d", prefix, st, d.Slug, i+1)
			must(s.UpsertParish(ontology.Parish{
				ID: parishID, SourceOrganizationID: orgID, JurisdictionID: dJurID,
				BulletinEndpointID: bulletinEndpointID,
			}))
		}
	}
	return n
}

// bootstrapSocial registers wedding-industry Instagram vendors and runs Apify
// health checks. Unchanged from the original bootstrapVendors.
func bootstrapSocial(ctx context.Context, s *store.Store, apify *ingest.ApifyClient, vendors []packs.VendorDef) int {
	n := 0
	for _, v := range vendors {
		orgID := "org_vendor_" + v.State + "_" + v.Handle
		meta := map[string]any{"verified": v.Verified, "handle": v.Handle}
		if v.TikTokHandle != "" {
			meta["tiktok_handle"] = v.TikTokHandle
		}
		if v.KnotURL != "" {
			meta["knot_url"] = v.KnotURL
		}
		metaJSON, _ := json.Marshal(meta)
		must(s.UpsertSourceOrganization(ontology.SourceOrganization{
			ID: orgID, OrgType: "business", Name: v.Name, CityID: v.CityID, OfficialURL: v.OfficialURL,
			Provenance: "public_business_website", DataMode: ontology.DataModeManual,
			Metadata: string(metaJSON),
		}))

		watched, err := s.UpsertWatchedSourceGeo(v.Handle, v.SourceClass, v.State, v.City)
		must(err)

		socialSourceID := "social_" + strings.ToLower(v.State) + "_" + v.Handle
		must(s.UpsertSocialSource(ontology.SocialSource{
			ID: socialSourceID, SourceOrganizationID: orgID, Platform: "instagram", Category: v.SourceClass,
			CityMarketID: v.CityID, ManuallyVerified: true, WatchedSourceID: watched.ID,
		}))

		endpointID := "endpoint_social_" + strings.ToLower(v.State) + "_" + v.Handle
		must(s.UpsertSourceEndpoint(ontology.SourceEndpoint{
			ID: endpointID, OrganizationID: orgID, EndpointType: "social_profile",
			URL: "https://instagram.com/" + v.Handle, AccessMethod: "structured_api", IsOfficial: true, DataMode: ontology.DataModeManual,
		}))

		connID := "connector_social_" + strings.ToLower(v.State) + "_" + v.Handle
		must(s.UpsertConnector(ontology.Connector{
			ID: connID, SourceEndpointID: endpointID, ConnectorType: "apify_instagram", Provider: "apify",
		}))

		result, stats := connectors.CheckInstagramHandle(ctx, apify, v.Handle)
		runCheck(s, connID, "https://instagram.com/"+v.Handle, result)
		if stats != nil {
			if err := s.UpdateWatchedSourceProfileStats(v.Handle, stats.FollowerCount, stats.FollowingCount, stats.PostCount,
				stats.FullName, stats.ProfilePicURL, stats.Verified); err != nil {
				log.Printf("[bootstrap-state] save profile stats for %s failed: %v", v.Handle, err)
			}
		}
		n++
	}
	return n
}

func runCheck(s *store.Store, connectorID, endpointURL string, result connectors.CheckResult) {
	started := time.Now().UTC()
	err := s.RecordConnectorRun(ontology.ConnectorRun{
		ConnectorID: connectorID, StartedAt: started, CompletedAt: &started,
		Status: result.Status, HTTPStatus: result.HTTPStatus, ResponseTimeMs: result.ResponseTimeMs,
		StructureSignature: result.StructureSignature, ErrorMessage: result.ErrorMessage,
	})
	if err != nil {
		log.Printf("[bootstrap-state] record run for %s failed: %v", connectorID, err)
		return
	}
	if result.Status == "success" {
		log.Printf("[bootstrap-state]   OK   %s", endpointURL)
	} else {
		log.Printf("[bootstrap-state]   FAIL %s: %s", endpointURL, result.ErrorMessage)
	}
}

func must(err error) {
	if err != nil {
		log.Fatalf("[bootstrap-state] %v", err)
	}
}
