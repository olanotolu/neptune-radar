// National coverage-map endpoints. Every response is assembled from real
// rows in the source registry (see internal/store/map.go) — nothing here is
// invented. Routes are parametrized by USPS state code; empty registries
// return honest empty arrays ("not yet configured"), never fabricated status.
package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"neptune-social-radar/backend/internal/ontology"
)

// denominationOf extracts the denomination from a source_organizations
// metadata JSON string, defaulting to "catholic" when NULL/empty/missing —
// the bootstrap default for dioceses/parishes without explicit metadata.
func denominationOf(metadata string) string {
	if metadata == "" {
		return "catholic"
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(metadata), &m); err != nil {
		return "catholic"
	}
	if d, ok := m["denomination"].(string); ok && d != "" {
		return strings.ToLower(d)
	}
	return "catholic"
}

type countyGovernmentView struct {
	County       ontology.County              `json:"county"`
	Organization *ontology.SourceOrganization `json:"organization,omitempty"`
	Endpoint     *ontology.SourceEndpoint     `json:"endpoint,omitempty"`
	Connector    *ontology.Connector          `json:"connector,omitempty"`
}

// mapGovernment lists every real county for a state and, where registered,
// government marriage-license connectors. Counties with no org show as
// "not yet configured."
func (s *Server) mapGovernment(w http.ResponseWriter, r *http.Request) {
	st, ok := s.requireState(w, r)
	if !ok {
		return
	}
	counties, err := s.Store.ListCountiesByState(st)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	orgs, err := s.Store.ListSourceOrganizationsByType("government_office", false)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	// Index orgs by county; only those whose county is in this state appear.
	countySet := make(map[string]bool, len(counties))
	for _, c := range counties {
		countySet[c.ID] = true
	}
	byCounty := make(map[string]ontology.SourceOrganization)
	for _, o := range orgs {
		if o.CountyID != "" && countySet[o.CountyID] {
			byCounty[o.CountyID] = o
		}
	}

	out := make([]countyGovernmentView, 0, len(counties))
	for _, c := range counties {
		view := countyGovernmentView{County: c}
		if org, ok := byCounty[c.ID]; ok {
			orgCopy := org
			view.Organization = &orgCopy
			endpoints, err := s.Store.ListSourceEndpointsByOrg(org.ID)
			if err == nil {
				for _, e := range endpoints {
					if e.EndpointType != "marriage_record_search" {
						continue
					}
					eCopy := e
					view.Endpoint = &eCopy
					if conn, err := s.Store.GetConnectorForEndpoint(e.ID); err == nil {
						connCopy := conn
						view.Connector = &connCopy
					}
					break
				}
			}
		}
		out = append(out, view)
	}
	writeJSON(w, http.StatusOK, out)
}

type parishView struct {
	Organization ontology.SourceOrganization `json:"organization"`
	Parish       ontology.Parish             `json:"parish"`
}
type dioceseView struct {
	Organization       ontology.SourceOrganization `json:"organization"`
	Jurisdiction       ontology.ChurchJurisdiction `json:"jurisdiction"`
	DirectoryEndpoint  *ontology.SourceEndpoint    `json:"directory_endpoint,omitempty"`
	DirectoryConnector *ontology.Connector         `json:"directory_connector,omitempty"`
	Parishes           []parishView                `json:"parishes"`
}

// mapChurches lists dioceses whose city (or first parish city) is in the state.
func (s *Server) mapChurches(w http.ResponseWriter, r *http.Request) {
	st, ok := s.requireState(w, r)
	if !ok {
		return
	}
	cities, err := s.Store.ListCitiesByState(st)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	citySet := make(map[string]bool, len(cities))
	for _, c := range cities {
		citySet[c.ID] = true
	}

	// ?denomination=catholic|episcopal|methodist|jewish filters dioceses/parishes
	// by the metadata.denomination field (default "catholic" when absent).
	denom := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("denomination")))

	orgs, err := s.Store.ListSourceOrganizationsByType("diocese", false)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	out := make([]dioceseView, 0)
	for _, org := range orgs {
		if denom != "" && denominationOf(org.Metadata) != denom {
			continue
		}
		// Include if diocese org is in a city of this state, or any parish is.
		inState := org.CityID != "" && citySet[org.CityID]
		jurisdiction, err := s.Store.GetChurchJurisdictionByOrg(org.ID)
		if err != nil {
			continue
		}
		view := dioceseView{Organization: org, Jurisdiction: jurisdiction, Parishes: make([]parishView, 0)}
		parishes, _ := s.Store.ListParishesByJurisdiction(jurisdiction.ID)
		for _, p := range parishes {
			porg, err := s.Store.GetSourceOrganization(p.SourceOrganizationID)
			if err != nil {
				continue
			}
			if denom != "" && denominationOf(porg.Metadata) != denom {
				continue
			}
			if porg.CityID != "" && citySet[porg.CityID] {
				inState = true
			}
			view.Parishes = append(view.Parishes, parishView{Organization: porg, Parish: p})
		}
		if !inState && len(cities) > 0 {
			// No city match — skip for non-empty states. If state has zero cities,
			// still skip dioceses (honest: not configured here).
			continue
		}
		if !inState {
			continue
		}

		if endpoints, err := s.Store.ListSourceEndpointsByOrg(org.ID); err == nil {
			for _, e := range endpoints {
				if e.EndpointType != "parish_directory" {
					continue
				}
				eCopy := e
				view.DirectoryEndpoint = &eCopy
				if conn, err := s.Store.GetConnectorForEndpoint(e.ID); err == nil {
					connCopy := conn
					view.DirectoryConnector = &connCopy
				}
			}
		}
		out = append(out, view)
	}
	if out == nil {
		out = []dioceseView{}
	}
	writeJSON(w, http.StatusOK, out)
}

type vendorView struct {
	Organization  ontology.SourceOrganization `json:"organization"`
	SocialSource  ontology.SocialSource       `json:"social_source"`
	WatchedSource *ontology.WatchedSource     `json:"watched_source,omitempty"`
	Connector     *ontology.Connector         `json:"connector,omitempty"`
}
type socialMarketView struct {
	City    ontology.City `json:"city"`
	Vendors []vendorView  `json:"vendors"`
}

// mapSocial lists wedding-industry accounts for every city market in the state.
// Returns [] when no cities/markets are configured (honest empty).
func (s *Server) mapSocial(w http.ResponseWriter, r *http.Request) {
	st, ok := s.requireState(w, r)
	if !ok {
		return
	}
	cities, err := s.Store.ListCitiesByState(st)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	// Backward-compat: older clients expected a single object for OH/Columbus.
	// New clients should pass ?all=1 or accept array; we always return array
	// unless Accept legacy — always array for multi-city readiness.
	// ?class=engagement_photographer|wedding_venue|florist|... filters vendors
	// by social_sources.category.
	classFilter := strings.TrimSpace(r.URL.Query().Get("class"))
	out := make([]socialMarketView, 0, len(cities))
	for _, city := range cities {
		sources, err := s.Store.ListSocialSourcesByCityMarket(city.ID)
		if err != nil {
			continue
		}
		view := socialMarketView{City: city, Vendors: make([]vendorView, 0)}
		for _, sc := range sources {
			if classFilter != "" && sc.Category != classFilter {
				continue
			}
			org, err := s.Store.GetSourceOrganization(sc.SourceOrganizationID)
			if err != nil {
				continue
			}
			vv := vendorView{Organization: org, SocialSource: sc}
			if sc.WatchedSourceID != "" {
				if ws, err := s.Store.GetWatchedSourceByID(sc.WatchedSourceID); err == nil {
					vv.WatchedSource = &ws
				}
			}
			if endpoints, err := s.Store.ListSourceEndpointsByOrg(org.ID); err == nil {
				for _, e := range endpoints {
					if e.EndpointType != "social_profile" {
						continue
					}
					if conn, err := s.Store.GetConnectorForEndpoint(e.ID); err == nil {
						connCopy := conn
						vv.Connector = &connCopy
					}
				}
			}
			view.Vendors = append(view.Vendors, vv)
		}
		out = append(out, view)
	}
	// Legacy single-object response for OH when client expects old shape:
	// if exactly one city market, also support ?format=legacy
	if r.URL.Query().Get("format") == "legacy" && len(out) == 1 {
		writeJSON(w, http.StatusOK, out[0])
		return
	}
	// For OH single Columbus market, frontend historically expected object —
	// keep that when there's exactly one city with vendors or one city total.
	if st == "OH" && len(out) == 1 && r.URL.Query().Get("all") != "1" {
		writeJSON(w, http.StatusOK, out[0])
		return
	}
	writeJSON(w, http.StatusOK, out)
}

type overviewCounts struct {
	Government int `json:"government"`
	Church     int `json:"church"`
	Social     int `json:"social"`
	Healthy    int `json:"healthy"`
	Degraded   int `json:"degraded"`
	Setup      int `json:"setup"`
	Offline    int `json:"offline"`
}
type overviewCityView struct {
	City   ontology.City  `json:"city"`
	Counts overviewCounts `json:"counts"`
}

// mapOverview computes real counts per city market in the state.
func (s *Server) mapOverview(w http.ResponseWriter, r *http.Request) {
	st, ok := s.requireState(w, r)
	if !ok {
		return
	}
	cities, err := s.Store.ListCitiesByState(st)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if len(cities) == 0 {
		writeJSON(w, http.StatusOK, []overviewCityView{})
		return
	}

	govOrgs, _ := s.Store.ListSourceOrganizationsByType("government_office", false)
	dioceseOrgs, _ := s.Store.ListSourceOrganizationsByType("diocese", false)

	out := make([]overviewCityView, 0, len(cities))
	for _, city := range cities {
		counts := overviewCounts{}
		tally := func(conn *ontology.Connector) {
			if conn == nil {
				counts.Setup++
				return
			}
			switch conn.Status {
			case ontology.ConnectorHealthy:
				counts.Healthy++
			case ontology.ConnectorDegraded:
				counts.Degraded++
			case ontology.ConnectorOffline:
				counts.Offline++
			default:
				counts.Setup++
			}
		}

		for _, org := range govOrgs {
			if org.CityID != city.ID {
				continue
			}
			counts.Government++
			endpoints, _ := s.Store.ListSourceEndpointsByOrg(org.ID)
			for _, e := range endpoints {
				conn, err := s.Store.GetConnectorForEndpoint(e.ID)
				if err != nil {
					tally(nil)
					continue
				}
				tally(&conn)
			}
		}

		for _, org := range dioceseOrgs {
			jurisdiction, err := s.Store.GetChurchJurisdictionByOrg(org.ID)
			if err != nil {
				continue
			}
			parishes, _ := s.Store.ListParishesByJurisdiction(jurisdiction.ID)
			for _, p := range parishes {
				porg, err := s.Store.GetSourceOrganization(p.SourceOrganizationID)
				if err != nil || porg.CityID != city.ID {
					continue
				}
				counts.Church++
				if p.BulletinEndpointID == "" {
					tally(nil)
					continue
				}
				conn, err := s.Store.GetConnectorForEndpoint(p.BulletinEndpointID)
				if err != nil {
					tally(nil)
					continue
				}
				tally(&conn)
			}
		}

		socialSources, _ := s.Store.ListSocialSourcesByCityMarket(city.ID)
		for _, sc := range socialSources {
			counts.Social++
			org, err := s.Store.GetSourceOrganization(sc.SourceOrganizationID)
			if err != nil {
				tally(nil)
				continue
			}
			endpoints, _ := s.Store.ListSourceEndpointsByOrg(org.ID)
			found := false
			for _, e := range endpoints {
				if e.EndpointType != "social_profile" {
					continue
				}
				if conn, err := s.Store.GetConnectorForEndpoint(e.ID); err == nil {
					tally(&conn)
					found = true
				}
			}
			if !found {
				tally(nil)
			}
		}

		out = append(out, overviewCityView{City: city, Counts: counts})
	}
	writeJSON(w, http.StatusOK, out)
}

// mapCoverageSummary returns national aggregate coverage stats for the
// dashboard widget — totals, denomination/SourceClass breakdowns, connector
// health, and top states by alive score.
func (s *Server) mapCoverageSummary(w http.ResponseWriter, r *http.Request) {
	sum, err := s.Store.CoverageSummary()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, sum)
}

// mapNationalCoverage returns choropleth-ready counts for all states.
func (s *Server) mapNationalCoverage(w http.ResponseWriter, r *http.Request) {
	cov, err := s.Store.ListStateCoverage()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if cov == nil {
		writeJSON(w, http.StatusOK, []any{})
		return
	}
	writeJSON(w, http.StatusOK, cov)
}

// mapOrganization, mapConnector, mapConnectorRuns back the detail panel.
func (s *Server) mapOrganization(w http.ResponseWriter, r *http.Request) {
	org, err := s.Store.GetSourceOrganization(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeJSON(w, http.StatusOK, org)
}

func (s *Server) mapConnector(w http.ResponseWriter, r *http.Request) {
	conn, err := s.Store.GetConnector(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeJSON(w, http.StatusOK, conn)
}

func (s *Server) mapConnectorRuns(w http.ResponseWriter, r *http.Request) {
	runs, err := s.Store.ListConnectorRuns(r.PathValue("id"), 50)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, runs)
}

// requireState validates USPS state from path and that geography exists.
func (s *Server) requireState(w http.ResponseWriter, r *http.Request) (string, bool) {
	st := strings.ToUpper(strings.TrimSpace(r.PathValue("state")))
	if st == "" {
		// Legacy routes registered without {state} — default OH.
		st = "OH"
	}
	if len(st) != 2 {
		writeError(w, http.StatusBadRequest, errorString("state must be a 2-letter USPS code"))
		return "", false
	}
	if _, err := s.Store.GetState(st); err != nil {
		// State not seeded yet — still allow empty responses if we know it's valid USPS?
		// Prefer seed; return 404 so ops know to run seed-geography.
		writeError(w, http.StatusNotFound, errorString("state not in registry — run seed-geography"))
		return "", false
	}
	return st, true
}
