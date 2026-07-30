// Command bootstrap-ohio populates the source registry with real Ohio
// geography and real, verified government/church/social sources, then runs
// a real health check against every one of them. It is idempotent — safe to
// re-run, and re-running is useful: it re-executes every check and updates
// each connector's status/timestamps from what actually happened this time.
//
// Usage: go run ./cmd/bootstrap-ohio
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"neptune-social-radar/backend/internal/connectors"
	"neptune-social-radar/backend/internal/ingest"
	"neptune-social-radar/backend/internal/ontology"
	"neptune-social-radar/backend/internal/store"
)

const columbusCityID = "city_columbus_oh"

func main() {
	ctx := context.Background()

	dsn := os.Getenv("DATABASE_URL")
	s, err := store.Open(dsn)
	if err != nil {
		log.Fatalf("open store: %v", err)
	}
	defer s.Close()

	apifyClient := ingest.NewApifyClient(os.Getenv("APIFY_TOKEN"))
	httpHealth := connectors.NewHTTPHealthConnector()

	log.Println("[bootstrap-ohio] geography...")
	bootstrapGeography(s)

	log.Println("[bootstrap-ohio] government: Franklin County Probate Court...")
	bootstrapGovernment(ctx, s, httpHealth)

	log.Println("[bootstrap-ohio] church: Diocese of Columbus + parishes...")
	bootstrapChurch(ctx, s, httpHealth)

	log.Println("[bootstrap-ohio] social: Columbus wedding-industry accounts...")
	bootstrapSocial(ctx, s, apifyClient)

	if _, err := s.Audit("bootstrap", "ohio", "bootstrap_run_completed", map[string]any{
		"counties":  len(ohioCounties),
		"parishes":  len(columbusParishes),
		"vendors":   len(columbusVendors),
		"apify_set": apifyClient.Available(),
	}, "", -1); err != nil {
		log.Printf("[bootstrap-ohio] audit write failed: %v", err)
	}

	log.Println("[bootstrap-ohio] done")
}

func bootstrapGeography(s *store.Store) {
	must(s.UpsertState("OH", "Ohio"))
	for _, c := range ohioCounties {
		must(s.UpsertCounty(c.FIPS, "OH", c.Name))
	}
	lat, lng := 39.9612, -82.9988
	must(s.UpsertCity(columbusCityID, "OH", franklinCountyFIPS, "Columbus", &lat, &lng))
}

func bootstrapGovernment(ctx context.Context, s *store.Store, hc connectors.SourceConnector) {
	orgID := "org_franklin_probate"
	meta, _ := json.Marshal(map[string]any{
		"phone":               franklinPhone,
		"coverage_start_date": franklinCoverageStart,
		"coverage_note":       franklinCoverageNote,
	})
	must(s.UpsertSourceOrganization(ontology.SourceOrganization{
		ID: orgID, OrgType: "government_office", Name: franklinCourtName,
		CityID: columbusCityID, CountyID: franklinCountyFIPS, OfficialURL: franklinCourtURL,
		Provenance: "official_government_website", DataMode: ontology.DataModeManual, Metadata: string(meta),
	}))

	endpointID := "endpoint_franklin_marriage_search"
	must(s.UpsertSourceEndpoint(ontology.SourceEndpoint{
		ID: endpointID, OrganizationID: orgID, EndpointType: "marriage_record_search",
		URL: franklinSearchURL, AccessMethod: "html_search_form", IsOfficial: true, DataMode: ontology.DataModeManual,
	}))

	connID := "connector_franklin_marriage_search"
	must(s.UpsertConnector(ontology.Connector{
		ID: connID, SourceEndpointID: endpointID, ConnectorType: "http_health", Provider: "neptune-http-health-v1",
	}))

	runCheck(s, connID, franklinSearchURL, hc.CheckHealth(ctx, franklinSearchURL))

	// The remaining verified county probate-court sources (see ohio_data.go).
	// Each gets a real HTTP health check against its search endpoint, so the
	// map shows what actually answered — not an assumed status.
	for _, g := range ohioGovSources {
		gOrgID := "org_gov_county_" + g.CountyFIPS
		gMeta, _ := json.Marshal(map[string]any{"capability_note": g.Note})
		must(s.UpsertSourceOrganization(ontology.SourceOrganization{
			ID: gOrgID, OrgType: "government_office", Name: g.CourtName,
			CountyID: g.CountyFIPS, OfficialURL: g.CourtURL,
			Provenance: "official_government_website", DataMode: ontology.DataModeManual, Metadata: string(gMeta),
		}))

		gEndpointID := "endpoint_gov_marriage_search_" + g.CountyFIPS
		must(s.UpsertSourceEndpoint(ontology.SourceEndpoint{
			ID: gEndpointID, OrganizationID: gOrgID, EndpointType: "marriage_record_search",
			URL: g.SearchURL, AccessMethod: "html_search_form", IsOfficial: true, DataMode: ontology.DataModeManual,
		}))

		gConnID := "connector_gov_marriage_search_" + g.CountyFIPS
		must(s.UpsertConnector(ontology.Connector{
			ID: gConnID, SourceEndpointID: gEndpointID, ConnectorType: "http_health", Provider: "neptune-http-health-v1",
		}))

		runCheck(s, gConnID, g.SearchURL, hc.CheckHealth(ctx, g.SearchURL))
	}
}

func bootstrapChurch(ctx context.Context, s *store.Store, hc connectors.SourceConnector) {
	dioceseOrgID := "org_diocese_columbus"
	meta, _ := json.Marshal(map[string]any{
		"deaneries":       dioceseColumbusDeaneries,
		"parishes_total":  dioceseColumbusParishesN,
		"counties_served": dioceseColumbusCountiesN,
	})
	must(s.UpsertSourceOrganization(ontology.SourceOrganization{
		ID: dioceseOrgID, OrgType: "diocese", Name: dioceseColumbusName,
		CityID: columbusCityID, OfficialURL: dioceseColumbusWebsite,
		Provenance: "official_directory", DataMode: ontology.DataModeManual, Metadata: string(meta),
	}))

	jurisdictionID := "jurisdiction_diocese_columbus"
	must(s.UpsertChurchJurisdiction(ontology.ChurchJurisdiction{
		ID: jurisdictionID, SourceOrganizationID: dioceseOrgID, JurisdictionType: "diocese", HubCityID: columbusCityID,
	}))

	directoryEndpointID := "endpoint_diocese_columbus_directory"
	must(s.UpsertSourceEndpoint(ontology.SourceEndpoint{
		ID: directoryEndpointID, OrganizationID: dioceseOrgID, EndpointType: "parish_directory",
		URL: dioceseColumbusDirectory, AccessMethod: "public_index", IsOfficial: true, DataMode: ontology.DataModeManual,
	}))

	directoryConnID := "connector_diocese_columbus_directory"
	must(s.UpsertConnector(ontology.Connector{
		ID: directoryConnID, SourceEndpointID: directoryEndpointID, ConnectorType: "http_health", Provider: "neptune-http-health-v1",
	}))
	runCheck(s, directoryConnID, dioceseColumbusDirectory, hc.CheckHealth(ctx, dioceseColumbusDirectory))

	for i, p := range columbusParishes {
		orgID := fmt.Sprintf("org_parish_columbus_%02d", i+1)
		meta := map[string]any{"source": parishSourceURL}
		if p.Address != "" {
			meta["address"] = p.Address
		}
		if p.BulletinURL != "" {
			meta["bulletin_url"] = p.BulletinURL
		}
		if p.BannsEvidence != "" {
			meta["banns_evidence"] = p.BannsEvidence
		}
		metaJSON, _ := json.Marshal(meta)
		must(s.UpsertSourceOrganization(ontology.SourceOrganization{
			ID: orgID, OrgType: "parish", Name: p.Name, CityID: columbusCityID,
			Provenance: "manually_curated", DataMode: ontology.DataModeManual, Metadata: string(metaJSON),
		}))

		// Register a real bulletin-archive endpoint only where one was
		// actually located and health-checked; parishes without one keep an
		// empty BulletinEndpointID and the map honestly shows "no bulletin
		// archive discovered yet".
		var bulletinEndpointID string
		if p.BulletinURL != "" {
			bulletinEndpointID = fmt.Sprintf("endpoint_parish_columbus_%02d_bulletin", i+1)
			// Aggregator URLs (Parishes Online / Discover Mass) are third-party
			// listings, not the parish's own site — hence IsOfficial=false.
			must(s.UpsertSourceEndpoint(ontology.SourceEndpoint{
				ID: bulletinEndpointID, OrganizationID: orgID, EndpointType: "bulletin_archive",
				URL: p.BulletinURL, AccessMethod: "public_index", IsOfficial: !p.Aggregator, DataMode: ontology.DataModeManual,
			}))
			bConnID := fmt.Sprintf("connector_parish_columbus_%02d_bulletin", i+1)
			must(s.UpsertConnector(ontology.Connector{
				ID: bConnID, SourceEndpointID: bulletinEndpointID, ConnectorType: "http_health", Provider: "neptune-http-health-v1",
			}))
			runCheck(s, bConnID, p.BulletinURL, hc.CheckHealth(ctx, p.BulletinURL))
		}

		parishID := fmt.Sprintf("parish_columbus_%02d", i+1)
		must(s.UpsertParish(ontology.Parish{
			ID: parishID, SourceOrganizationID: orgID, JurisdictionID: jurisdictionID,
			BulletinEndpointID: bulletinEndpointID,
		}))
	}

	// The other five Ohio Catholic jurisdictions. Directory connectors get
	// real health checks; parish inventories come from the future directory
	// crawl, so these honestly start at "0 parishes registered".
	for _, d := range ohioDioceses {
		dOrgID := "org_diocese_" + d.Slug
		must(s.UpsertSourceOrganization(ontology.SourceOrganization{
			ID: dOrgID, OrgType: "diocese", Name: d.Name, OfficialURL: d.Website,
			Provenance: "official_directory", DataMode: ontology.DataModeManual,
		}))

		dJurID := "jurisdiction_diocese_" + d.Slug
		must(s.UpsertChurchJurisdiction(ontology.ChurchJurisdiction{
			ID: dJurID, SourceOrganizationID: dOrgID, JurisdictionType: d.Type,
		}))

		dEndpointID := "endpoint_diocese_" + d.Slug + "_directory"
		must(s.UpsertSourceEndpoint(ontology.SourceEndpoint{
			ID: dEndpointID, OrganizationID: dOrgID, EndpointType: "parish_directory",
			URL: d.Directory, AccessMethod: "public_index", IsOfficial: true, DataMode: ontology.DataModeManual,
		}))

		dConnID := "connector_diocese_" + d.Slug + "_directory"
		must(s.UpsertConnector(ontology.Connector{
			ID: dConnID, SourceEndpointID: dEndpointID, ConnectorType: "http_health", Provider: "neptune-http-health-v1",
		}))
		runCheck(s, dConnID, d.Directory, hc.CheckHealth(ctx, d.Directory))
	}
}

func bootstrapSocial(ctx context.Context, s *store.Store, apify *ingest.ApifyClient) {
	for _, v := range columbusVendors {
		orgID := "org_vendor_" + v.Handle
		must(s.UpsertSourceOrganization(ontology.SourceOrganization{
			ID: orgID, OrgType: "business", Name: v.Name, CityID: columbusCityID, OfficialURL: v.OfficialURL,
			Provenance: "public_business_website", DataMode: ontology.DataModeManual,
		}))

		watched, err := s.UpsertWatchedSourceGeo(v.Handle, v.SourceClass, "OH", "Columbus")
		must(err)

		socialSourceID := "social_" + v.Handle
		must(s.UpsertSocialSource(ontology.SocialSource{
			ID: socialSourceID, SourceOrganizationID: orgID, Platform: "instagram", Category: v.SourceClass,
			CityMarketID: columbusCityID, ManuallyVerified: true, WatchedSourceID: watched.ID,
		}))

		endpointID := "endpoint_social_" + v.Handle
		must(s.UpsertSourceEndpoint(ontology.SourceEndpoint{
			ID: endpointID, OrganizationID: orgID, EndpointType: "social_profile",
			URL: "https://instagram.com/" + v.Handle, AccessMethod: "structured_api", IsOfficial: true, DataMode: ontology.DataModeManual,
		}))

		connID := "connector_social_" + v.Handle
		must(s.UpsertConnector(ontology.Connector{
			ID: connID, SourceEndpointID: endpointID, ConnectorType: "apify_instagram", Provider: "apify",
		}))

		result, stats := connectors.CheckInstagramHandle(ctx, apify, v.Handle)
		runCheck(s, connID, "https://instagram.com/"+v.Handle, result)
		if stats != nil {
			if err := s.UpdateWatchedSourceProfileStats(v.Handle, stats.FollowerCount, stats.FollowingCount, stats.PostCount,
				stats.FullName, stats.ProfilePicURL, stats.Verified); err != nil {
				log.Printf("[bootstrap-ohio] save profile stats for %s failed: %v", v.Handle, err)
			}
		}
	}
}

// runCheck persists one real check result and mirrors it into audit_events
// for the connector's inspection trail.
func runCheck(s *store.Store, connectorID, endpointURL string, result connectors.CheckResult) {
	started := time.Now().UTC()
	err := s.RecordConnectorRun(ontology.ConnectorRun{
		ConnectorID: connectorID, StartedAt: started, CompletedAt: &started,
		Status: result.Status, HTTPStatus: result.HTTPStatus, ResponseTimeMs: result.ResponseTimeMs,
		StructureSignature: result.StructureSignature, ErrorMessage: result.ErrorMessage,
	})
	if err != nil {
		log.Printf("[bootstrap-ohio] record run for %s failed: %v", connectorID, err)
		return
	}
	if _, err := s.Audit("connector", connectorID, "health_check", map[string]any{
		"endpoint_url": endpointURL, "status": result.Status, "http_status": result.HTTPStatus,
		"response_time_ms": result.ResponseTimeMs, "error": result.ErrorMessage,
	}, "", -1); err != nil {
		log.Printf("[bootstrap-ohio] audit write for %s failed: %v", connectorID, err)
	}
	if result.Status == "success" {
		log.Printf("[bootstrap-ohio]   OK   %s (%dms)", endpointURL, result.ResponseTimeMs)
	} else {
		log.Printf("[bootstrap-ohio]   FAIL %s: %s", endpointURL, result.ErrorMessage)
	}
}

func must(err error) {
	if err != nil {
		log.Fatalf("[bootstrap-ohio] %v", err)
	}
}
