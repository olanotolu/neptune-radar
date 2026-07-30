// Ohio coverage-map endpoints. Every response is assembled from real rows in
// the source registry (see internal/store/map.go) — nothing here is
// computed in the frontend or hardcoded. Only Ohio has data today, so these
// routes are literal ("OH"), not parametrized by state.
package api

import (
	"net/http"

	"neptune-social-radar/backend/internal/ontology"
)

const columbusCityID = "city_columbus_oh"

type countyGovernmentView struct {
	County       ontology.County              `json:"county"`
	Organization *ontology.SourceOrganization `json:"organization,omitempty"`
	Endpoint     *ontology.SourceEndpoint     `json:"endpoint,omitempty"`
	Connector    *ontology.Connector          `json:"connector,omitempty"`
}

// mapGovernment lists every real Ohio county and, for the ones that have a
// registered government connector, its real endpoint/connector detail. A
// county with no matching organization simply has nil fields — the frontend
// renders that as "not yet configured," never a fabricated status.
func (s *Server) mapGovernment(w http.ResponseWriter, r *http.Request) {
	counties, err := s.Store.ListCountiesByState("OH")
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	orgs, err := s.Store.ListSourceOrganizationsByType("government_office", false)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	byCounty := make(map[string]ontology.SourceOrganization, len(orgs))
	for _, o := range orgs {
		if o.CountyID != "" {
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

// mapChurches lists every real diocese registered for Ohio and its real
// parishes. A parish with no bulletin_endpoint_id has honestly had no
// bulletin archive discovered yet — never a fabricated URL.
func (s *Server) mapChurches(w http.ResponseWriter, r *http.Request) {
	orgs, err := s.Store.ListSourceOrganizationsByType("diocese", false)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	out := make([]dioceseView, 0, len(orgs))
	for _, org := range orgs {
		jurisdiction, err := s.Store.GetChurchJurisdictionByOrg(org.ID)
		if err != nil {
			continue
		}
		// Parishes must serialize as [] (not null) — a diocese with no
		// parishes yet is a valid, honest state and the frontend renders it.
		view := dioceseView{Organization: org, Jurisdiction: jurisdiction, Parishes: make([]parishView, 0)}

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

		parishes, err := s.Store.ListParishesByJurisdiction(jurisdiction.ID)
		if err == nil {
			for _, p := range parishes {
				porg, err := s.Store.GetSourceOrganization(p.SourceOrganizationID)
				if err != nil {
					continue
				}
				view.Parishes = append(view.Parishes, parishView{Organization: porg, Parish: p})
			}
		}
		out = append(out, view)
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

// mapSocial lists the real Columbus wedding-industry accounts and their
// current status straight from the same watched_sources row the live Apify
// worker polls — never a separate, decorative record.
func (s *Server) mapSocial(w http.ResponseWriter, r *http.Request) {
	city, err := s.Store.GetCity(columbusCityID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	sources, err := s.Store.ListSocialSourcesByCityMarket(city.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	view := socialMarketView{City: city}
	for _, sc := range sources {
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
	writeJSON(w, http.StatusOK, view)
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

// mapOverview computes every count from real rows — government connectors,
// parishes, and social sources currently registered for Columbus — never a
// number assembled in the frontend.
func (s *Server) mapOverview(w http.ResponseWriter, r *http.Request) {
	city, err := s.Store.GetCity(columbusCityID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
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

	govOrgs, _ := s.Store.ListSourceOrganizationsByType("government_office", false)
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

	dioceseOrgs, _ := s.Store.ListSourceOrganizationsByType("diocese", false)
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

	writeJSON(w, http.StatusOK, []overviewCityView{{City: city, Counts: counts}})
}

// mapOrganization, mapConnector, mapConnectorRuns back the detail panel's
// "inspect source" / "inspect connector" / "view connector history" links.
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
