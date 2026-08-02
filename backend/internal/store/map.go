// The source registry store methods: geography backbone, real government/
// church/social organizations, their endpoints, and the connectors that
// check them. RecordConnectorRun is the only function that ever moves a
// connector's status away from "setup" — every other write here is
// structural (what exists), never a health claim.
package store

import (
	"database/sql"
	"sort"
	"strings"

	"neptune-social-radar/backend/internal/ontology"
)

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func nullIfZero(n int) any {
	if n == 0 {
		return nil
	}
	return n
}

// --- states / counties / cities -------------------------------------------

func (s *Store) UpsertState(id, name string) error {
	_, err := s.DB.Exec(
		`INSERT INTO states (id, name) VALUES ($1, $2) ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name`,
		id, name)
	return err
}

func (s *Store) GetState(id string) (ontology.State, error) {
	var st ontology.State
	err := s.DB.QueryRow(`SELECT id, name FROM states WHERE id = $1`, id).Scan(&st.ID, &st.Name)
	return st, err
}

func (s *Store) ListStates() ([]ontology.State, error) {
	rows, err := s.DB.Query(`SELECT id, name FROM states ORDER BY name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ontology.State
	for rows.Next() {
		var st ontology.State
		if err := rows.Scan(&st.ID, &st.Name); err != nil {
			return nil, err
		}
		out = append(out, st)
	}
	return out, rows.Err()
}

// StateCoverage is national map choropleth / coverage strip data.
type StateCoverage struct {
	StateID            string `json:"state_id"`
	Name               string `json:"name"`
	CountyCount        int    `json:"county_count"`
	CountiesConfigured int    `json:"counties_configured"` // gov marriage endpoints present
	Cities             int    `json:"cities"`
	GovernmentSources  int    `json:"government_sources"`
	ChurchSources      int    `json:"church_sources"`
	SocialSources      int    `json:"social_sources"`
	WatchedSources     int    `json:"watched_sources"`
	// AliveScore 0–1 for UI choropleth (social-weighted: radar value first).
	AliveScore float64 `json:"alive_score"`
	// DenominationBreakdown counts dioceses/parishes per denomination
	// (defaults to "catholic" when metadata.denomination is NULL/empty).
	DenominationBreakdown map[string]int `json:"denomination_breakdown"` // {"catholic": N, "episcopal": N, ...}
	// SourceClassBreakdown counts social vendors per social_sources.category.
	SourceClassBreakdown map[string]int `json:"source_class_breakdown"` // {"engagement_photographer": N, ...}
	// ConnectorHealth summarizes connector status for the state's sources.
	ConnectorHealth ConnectorHealthSummary `json:"connector_health"`
}

// ConnectorHealthSummary counts connectors by status for a state.
type ConnectorHealthSummary struct {
	Healthy  int `json:"healthy"`
	Degraded int `json:"degraded"`
	Offline  int `json:"offline"`
	Setup    int `json:"setup"`
}

// ListStateCoverage returns one row per state with real source counts.
func (s *Store) ListStateCoverage() ([]StateCoverage, error) {
	states, err := s.ListStates()
	if err != nil {
		return nil, err
	}
	out := make([]StateCoverage, 0, len(states))
	for _, st := range states {
		c := StateCoverage{
			StateID:               st.ID,
			Name:                  st.Name,
			DenominationBreakdown: map[string]int{},
			SourceClassBreakdown:  map[string]int{},
		}
		_ = s.DB.QueryRow(`SELECT COUNT(*) FROM counties WHERE state_id = $1`, st.ID).Scan(&c.CountyCount)
		_ = s.DB.QueryRow(`SELECT COUNT(*) FROM cities WHERE state_id = $1`, st.ID).Scan(&c.Cities)
		_ = s.DB.QueryRow(`
			SELECT COUNT(DISTINCT o.county_id) FROM source_organizations o
			JOIN counties c ON c.id = o.county_id
			WHERE c.state_id = $1 AND o.org_type = 'government_office' AND o.data_mode != 'fixture'
			  AND o.county_id IS NOT NULL AND o.county_id <> ''`, st.ID).Scan(&c.CountiesConfigured)
		_ = s.DB.QueryRow(`
			SELECT COUNT(*) FROM source_organizations o
			LEFT JOIN counties c ON c.id = o.county_id
			LEFT JOIN cities ci ON ci.id = o.city_id
			WHERE o.org_type = 'government_office' AND o.data_mode != 'fixture'
			  AND (c.state_id = $1 OR ci.state_id = $1)`, st.ID).Scan(&c.GovernmentSources)
		_ = s.DB.QueryRow(`
			SELECT COUNT(*) FROM source_organizations o
			LEFT JOIN cities ci ON ci.id = o.city_id
			WHERE o.org_type IN ('diocese','parish') AND o.data_mode != 'fixture'
			  AND ci.state_id = $1`, st.ID).Scan(&c.ChurchSources)
		_ = s.DB.QueryRow(`
			SELECT COUNT(*) FROM source_organizations o
			LEFT JOIN cities ci ON ci.id = o.city_id
			WHERE o.org_type = 'business' AND o.data_mode != 'fixture'
			  AND ci.state_id = $1`, st.ID).Scan(&c.SocialSources)
		_ = s.DB.QueryRow(`
			SELECT COUNT(*) FROM watched_sources WHERE active AND upper(state) = upper($1)`, st.ID).Scan(&c.WatchedSources)

		// Denomination breakdown: dioceses/parishes grouped by metadata.denomination
		// (defaulting to "catholic" when NULL/empty).
		if rows, err := s.DB.Query(`
			SELECT COALESCE(o.metadata::json->>'denomination','catholic'), COUNT(*)
			FROM source_organizations o
			LEFT JOIN cities ci ON ci.id = o.city_id
			WHERE o.org_type IN ('diocese','parish') AND o.data_mode != 'fixture'
			  AND ci.state_id = $1
			GROUP BY 1`, st.ID); err == nil {
			for rows.Next() {
				var denom string
				var n int
				if err := rows.Scan(&denom, &n); err == nil {
					c.DenominationBreakdown[strings.ToLower(denom)] = n
				}
			}
			rows.Close()
		}

		// SourceClass breakdown: social vendors grouped by social_sources.category.
		if rows, err := s.DB.Query(`
			SELECT ss.category, COUNT(*)
			FROM social_sources ss
			JOIN cities ci ON ci.id = ss.city_market_id
			WHERE ci.state_id = $1
			GROUP BY 1`, st.ID); err == nil {
			for rows.Next() {
				var cat string
				var n int
				if err := rows.Scan(&cat, &n); err == nil {
					c.SourceClassBreakdown[cat] = n
				}
			}
			rows.Close()
		}

		// Connector health: connectors for this state's sources grouped by status.
		if rows, err := s.DB.Query(`
			SELECT c.status, COUNT(*)
			FROM connectors c
			JOIN source_endpoints se ON se.id = c.source_endpoint_id
			JOIN source_organizations o ON o.id = se.organization_id
			LEFT JOIN cities ci ON ci.id = o.city_id
			LEFT JOIN counties co ON co.id = o.county_id
			WHERE ci.state_id = $1 OR co.state_id = $1
			GROUP BY 1`, st.ID); err == nil {
			for rows.Next() {
				var status string
				var n int
				if err := rows.Scan(&status, &n); err == nil {
					switch status {
					case "healthy":
						c.ConnectorHealth.Healthy = n
					case "degraded":
						c.ConnectorHealth.Degraded = n
					case "offline":
						c.ConnectorHealth.Offline = n
					default:
						c.ConnectorHealth.Setup += n
					}
				}
			}
			rows.Close()
		}
		// Score: any social/watched lights the state; gov/church add depth.
		score := 0.0
		if c.WatchedSources > 0 || c.SocialSources > 0 {
			score = 0.55 + 0.05*float64(min(c.WatchedSources, 8))
		}
		if c.GovernmentSources > 0 {
			score += 0.15
		}
		if c.ChurchSources > 0 {
			score += 0.1
		}
		if c.Cities > 0 {
			score += 0.05
		}
		if score > 1 {
			score = 1
		}
		c.AliveScore = score
		out = append(out, c)
	}
	return out, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// CoverageSummary is the national coverage dashboard widget: aggregate counts,
// denomination/SourceClass breakdowns, connector health, and top states.
type CoverageSummary struct {
	StatesCovered          int                    `json:"states_covered"`
	TotalGovernmentSources int                    `json:"total_government_sources"`
	TotalChurchSources     int                    `json:"total_church_sources"`
	TotalSocialSources     int                    `json:"total_social_sources"`
	TotalWatchedSources    int                    `json:"total_watched_sources"`
	AvgAliveScore          float64                `json:"avg_alive_score"`
	DenominationBreakdown  map[string]int         `json:"denomination_breakdown"`
	SourceClassBreakdown   map[string]int         `json:"source_class_breakdown"`
	ConnectorHealth        ConnectorHealthSummary `json:"connector_health"`
	TopStates              []topState             `json:"top_states"`
}

type topState struct {
	StateID    string  `json:"state_id"`
	Name       string  `json:"name"`
	AliveScore float64 `json:"alive_score"`
}

// CoverageSummary aggregates national coverage stats by reusing
// ListStateCoverage and summing in Go — no duplicate SQL.
func (s *Store) CoverageSummary() (*CoverageSummary, error) {
	states, err := s.ListStateCoverage()
	if err != nil {
		return nil, err
	}
	sum := &CoverageSummary{
		DenominationBreakdown: map[string]int{},
		SourceClassBreakdown:  map[string]int{},
		TopStates:             []topState{},
	}
	if len(states) == 0 {
		return sum, nil
	}
	for _, st := range states {
		sum.StatesCovered++
		sum.TotalGovernmentSources += st.GovernmentSources
		sum.TotalChurchSources += st.ChurchSources
		sum.TotalSocialSources += st.SocialSources
		sum.TotalWatchedSources += st.WatchedSources
		sum.AvgAliveScore += st.AliveScore
		sum.ConnectorHealth.Healthy += st.ConnectorHealth.Healthy
		sum.ConnectorHealth.Degraded += st.ConnectorHealth.Degraded
		sum.ConnectorHealth.Offline += st.ConnectorHealth.Offline
		sum.ConnectorHealth.Setup += st.ConnectorHealth.Setup
		for d, n := range st.DenominationBreakdown {
			sum.DenominationBreakdown[d] += n
		}
		for c, n := range st.SourceClassBreakdown {
			sum.SourceClassBreakdown[c] += n
		}
	}
	sum.AvgAliveScore /= float64(sum.StatesCovered)
	// Top states by alive score (descending).
	top := make([]topState, len(states))
	for i, st := range states {
		top[i] = topState{StateID: st.StateID, Name: st.Name, AliveScore: st.AliveScore}
	}
	sort.SliceStable(top, func(i, j int) bool { return top[i].AliveScore > top[j].AliveScore })
	if len(top) > 5 {
		top = top[:5]
	}
	sum.TopStates = top
	return sum, nil
}

func (s *Store) UpsertCounty(id, stateID, name string) error {
	_, err := s.DB.Exec(
		`INSERT INTO counties (id, state_id, name) VALUES ($1, $2, $3)
		 ON CONFLICT (id) DO UPDATE SET state_id = EXCLUDED.state_id, name = EXCLUDED.name`,
		id, stateID, name)
	return err
}

func (s *Store) ListCountiesByState(stateID string) ([]ontology.County, error) {
	rows, err := s.DB.Query(`SELECT id, state_id, name FROM counties WHERE state_id = $1 ORDER BY name ASC`, stateID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ontology.County
	for rows.Next() {
		var c ontology.County
		if err := rows.Scan(&c.ID, &c.StateID, &c.Name); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *Store) UpsertCity(id, stateID, countyID, name string, lat, lng *float64) error {
	_, err := s.DB.Exec(
		`INSERT INTO cities (id, state_id, primary_county_id, name, lat, lng) VALUES ($1, $2, $3, $4, $5, $6)
		 ON CONFLICT (id) DO UPDATE SET state_id = EXCLUDED.state_id, primary_county_id = EXCLUDED.primary_county_id,
		   name = EXCLUDED.name, lat = EXCLUDED.lat, lng = EXCLUDED.lng`,
		id, stateID, nullIfEmpty(countyID), name, lat, lng)
	return err
}

func (s *Store) ListCitiesByState(stateID string) ([]ontology.City, error) {
	rows, err := s.DB.Query(`SELECT id, state_id, COALESCE(primary_county_id,''), name, lat, lng FROM cities WHERE state_id = $1 ORDER BY name ASC`, stateID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ontology.City
	for rows.Next() {
		var c ontology.City
		var lat, lng sql.NullFloat64
		if err := rows.Scan(&c.ID, &c.StateID, &c.PrimaryCountyID, &c.Name, &lat, &lng); err != nil {
			return nil, err
		}
		if lat.Valid {
			c.Lat = &lat.Float64
		}
		if lng.Valid {
			c.Lng = &lng.Float64
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *Store) GetCity(id string) (ontology.City, error) {
	var c ontology.City
	var lat, lng sql.NullFloat64
	err := s.DB.QueryRow(`SELECT id, state_id, COALESCE(primary_county_id,''), name, lat, lng FROM cities WHERE id = $1`, id).
		Scan(&c.ID, &c.StateID, &c.PrimaryCountyID, &c.Name, &lat, &lng)
	if err != nil {
		return c, err
	}
	if lat.Valid {
		c.Lat = &lat.Float64
	}
	if lng.Valid {
		c.Lng = &lng.Float64
	}
	return c, nil
}

// --- source_organizations / source_endpoints -------------------------------

func (s *Store) UpsertSourceOrganization(o ontology.SourceOrganization) error {
	_, err := s.DB.Exec(
		`INSERT INTO source_organizations (id, org_type, name, city_id, county_id, official_url, provenance, data_mode, metadata)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		 ON CONFLICT (id) DO UPDATE SET org_type=EXCLUDED.org_type, name=EXCLUDED.name, city_id=EXCLUDED.city_id,
		   county_id=EXCLUDED.county_id, official_url=EXCLUDED.official_url, provenance=EXCLUDED.provenance,
		   data_mode=EXCLUDED.data_mode, metadata=EXCLUDED.metadata`,
		o.ID, o.OrgType, o.Name, nullIfEmpty(o.CityID), nullIfEmpty(o.CountyID), nullIfEmpty(o.OfficialURL),
		o.Provenance, string(o.DataMode), nullIfEmpty(o.Metadata),
	)
	return err
}

func (s *Store) GetSourceOrganization(id string) (ontology.SourceOrganization, error) {
	var o ontology.SourceOrganization
	var cityID, countyID, url, meta sql.NullString
	var dataMode string
	err := s.DB.QueryRow(
		`SELECT id, org_type, name, city_id, county_id, official_url, provenance, data_mode, metadata, created_at
		 FROM source_organizations WHERE id = $1`, id,
	).Scan(&o.ID, &o.OrgType, &o.Name, &cityID, &countyID, &url, &o.Provenance, &dataMode, &meta, &o.CreatedAt)
	if err != nil {
		return o, err
	}
	o.CityID, o.CountyID, o.OfficialURL, o.Metadata = cityID.String, countyID.String, url.String, meta.String
	o.DataMode = ontology.DataMode(dataMode)
	return o, nil
}

// ListSourceOrganizationsByType returns real organizations of one kind,
// excluding fixture rows unless includeFixtures is set — the guarantee that
// keeps demo/test data from ever appearing in the live map.
func (s *Store) ListSourceOrganizationsByType(orgType string, includeFixtures bool) ([]ontology.SourceOrganization, error) {
	q := `SELECT id, org_type, name, city_id, county_id, official_url, provenance, data_mode, metadata, created_at
	      FROM source_organizations WHERE org_type = $1`
	if !includeFixtures {
		q += ` AND data_mode != 'fixture'`
	}
	q += ` ORDER BY name ASC`
	rows, err := s.DB.Query(q, orgType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ontology.SourceOrganization
	for rows.Next() {
		var o ontology.SourceOrganization
		var cityID, countyID, url, meta sql.NullString
		var dataMode string
		if err := rows.Scan(&o.ID, &o.OrgType, &o.Name, &cityID, &countyID, &url, &o.Provenance, &dataMode, &meta, &o.CreatedAt); err != nil {
			return nil, err
		}
		o.CityID, o.CountyID, o.OfficialURL, o.Metadata = cityID.String, countyID.String, url.String, meta.String
		o.DataMode = ontology.DataMode(dataMode)
		out = append(out, o)
	}
	return out, rows.Err()
}

func (s *Store) UpsertSourceEndpoint(e ontology.SourceEndpoint) error {
	_, err := s.DB.Exec(
		`INSERT INTO source_endpoints (id, organization_id, endpoint_type, url, access_method, is_official, data_mode)
		 VALUES ($1,$2,$3,$4,$5,$6,$7)
		 ON CONFLICT (id) DO UPDATE SET organization_id=EXCLUDED.organization_id, endpoint_type=EXCLUDED.endpoint_type,
		   url=EXCLUDED.url, access_method=EXCLUDED.access_method, is_official=EXCLUDED.is_official, data_mode=EXCLUDED.data_mode`,
		e.ID, e.OrganizationID, e.EndpointType, e.URL, e.AccessMethod, e.IsOfficial, string(e.DataMode),
	)
	return err
}

func (s *Store) ListSourceEndpointsByOrg(orgID string) ([]ontology.SourceEndpoint, error) {
	rows, err := s.DB.Query(
		`SELECT id, organization_id, endpoint_type, url, access_method, is_official, data_mode, created_at
		 FROM source_endpoints WHERE organization_id = $1 ORDER BY created_at ASC`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ontology.SourceEndpoint
	for rows.Next() {
		var e ontology.SourceEndpoint
		var dataMode string
		if err := rows.Scan(&e.ID, &e.OrganizationID, &e.EndpointType, &e.URL, &e.AccessMethod, &e.IsOfficial, &dataMode, &e.CreatedAt); err != nil {
			return nil, err
		}
		e.DataMode = ontology.DataMode(dataMode)
		out = append(out, e)
	}
	return out, rows.Err()
}

// --- connectors / connector_runs -------------------------------------------

// UpsertConnector registers a connector's structure only — it never touches
// status/timestamps, which belong exclusively to RecordConnectorRun.
func (s *Store) UpsertConnector(c ontology.Connector) error {
	_, err := s.DB.Exec(
		`INSERT INTO connectors (id, source_endpoint_id, connector_type, provider, schedule)
		 VALUES ($1,$2,$3,$4,$5)
		 ON CONFLICT (id) DO UPDATE SET connector_type=EXCLUDED.connector_type, provider=EXCLUDED.provider, schedule=EXCLUDED.schedule`,
		c.ID, c.SourceEndpointID, c.ConnectorType, c.Provider, nullIfEmpty(c.Schedule),
	)
	return err
}

func scanConnector(row interface{ Scan(...any) error }) (ontology.Connector, error) {
	var c ontology.Connector
	var status string
	var schedule, errMsg sql.NullString
	var lastChecked, lastSuccess, lastFailure sql.NullTime
	err := row.Scan(&c.ID, &c.SourceEndpointID, &c.ConnectorType, &c.Provider, &status, &schedule,
		&lastChecked, &lastSuccess, &lastFailure, &errMsg, &c.CreatedAt)
	if err != nil {
		return c, err
	}
	c.Status = ontology.ConnectorStatus(status)
	c.Schedule, c.ErrorMessage = schedule.String, errMsg.String
	if lastChecked.Valid {
		t := lastChecked.Time
		c.LastCheckedAt = &t
	}
	if lastSuccess.Valid {
		t := lastSuccess.Time
		c.LastSuccessAt = &t
	}
	if lastFailure.Valid {
		t := lastFailure.Time
		c.LastFailureAt = &t
	}
	return c, nil
}

const connectorSelect = `SELECT id, source_endpoint_id, connector_type, provider, status, schedule,
	last_checked_at, last_success_at, last_failure_at, error_message, created_at FROM connectors`

func (s *Store) GetConnector(id string) (ontology.Connector, error) {
	return scanConnector(s.DB.QueryRow(connectorSelect+` WHERE id = $1`, id))
}

// GetConnectorForEndpoint returns the connector for a source endpoint, or
// sql.ErrNoRows if none has been registered yet.
func (s *Store) GetConnectorForEndpoint(endpointID string) (ontology.Connector, error) {
	return scanConnector(s.DB.QueryRow(connectorSelect+` WHERE source_endpoint_id = $1`, endpointID))
}

// RecordConnectorRun is the only function that moves a connector's status
// away from "setup". Success marks it healthy immediately. Failure looks at
// the 3 most recent runs (including this one): 1-2 recent failures is
// "degraded", 3 consecutive is "offline".
func (s *Store) RecordConnectorRun(run ontology.ConnectorRun) error {
	_, err := s.DB.Exec(
		`INSERT INTO connector_runs (id, connector_id, started_at, completed_at, status, http_status, response_time_ms, structure_signature, error_message)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		NewID("run"), run.ConnectorID, run.StartedAt, run.CompletedAt, run.Status,
		nullIfZero(run.HTTPStatus), nullIfZero(run.ResponseTimeMs), nullIfEmpty(run.StructureSignature), nullIfEmpty(run.ErrorMessage),
	)
	if err != nil {
		return err
	}

	if run.Status == "success" {
		_, err = s.DB.Exec(
			`UPDATE connectors SET status = 'healthy', last_checked_at = $2, last_success_at = $2, error_message = NULL WHERE id = $1`,
			run.ConnectorID, run.StartedAt,
		)
		return err
	}

	var recentFailures int
	if err := s.DB.QueryRow(
		`SELECT COUNT(*) FROM (
		   SELECT status FROM connector_runs WHERE connector_id = $1 ORDER BY started_at DESC LIMIT 3
		 ) recent WHERE status = 'failure'`, run.ConnectorID,
	).Scan(&recentFailures); err != nil {
		return err
	}
	status := "degraded"
	if recentFailures >= 3 {
		status = "offline"
	}
	_, err = s.DB.Exec(
		`UPDATE connectors SET status = $2, last_checked_at = $3, last_failure_at = $3, error_message = $4 WHERE id = $1`,
		run.ConnectorID, status, run.StartedAt, run.ErrorMessage,
	)
	return err
}

func (s *Store) ListConnectorRuns(connectorID string, limit int) ([]ontology.ConnectorRun, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.DB.Query(
		`SELECT id, connector_id, started_at, completed_at, status, http_status, response_time_ms, structure_signature, error_message
		 FROM connector_runs WHERE connector_id = $1 ORDER BY started_at DESC LIMIT $2`, connectorID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ontology.ConnectorRun
	for rows.Next() {
		var r ontology.ConnectorRun
		var completed sql.NullTime
		var httpStatus, respTime sql.NullInt64
		var sig, errMsg sql.NullString
		if err := rows.Scan(&r.ID, &r.ConnectorID, &r.StartedAt, &completed, &r.Status, &httpStatus, &respTime, &sig, &errMsg); err != nil {
			return nil, err
		}
		if completed.Valid {
			t := completed.Time
			r.CompletedAt = &t
		}
		r.HTTPStatus = int(httpStatus.Int64)
		r.ResponseTimeMs = int(respTime.Int64)
		r.StructureSignature, r.ErrorMessage = sig.String, errMsg.String
		out = append(out, r)
	}
	return out, rows.Err()
}

// --- church_jurisdictions / parishes ----------------------------------------

func (s *Store) UpsertChurchJurisdiction(j ontology.ChurchJurisdiction) error {
	_, err := s.DB.Exec(
		`INSERT INTO church_jurisdictions (id, source_organization_id, jurisdiction_type, hub_city_id) VALUES ($1,$2,$3,$4)
		 ON CONFLICT (id) DO UPDATE SET source_organization_id=EXCLUDED.source_organization_id,
		   jurisdiction_type=EXCLUDED.jurisdiction_type, hub_city_id=EXCLUDED.hub_city_id`,
		j.ID, j.SourceOrganizationID, j.JurisdictionType, nullIfEmpty(j.HubCityID),
	)
	return err
}

func (s *Store) GetChurchJurisdictionByOrg(orgID string) (ontology.ChurchJurisdiction, error) {
	var j ontology.ChurchJurisdiction
	var hubCity sql.NullString
	err := s.DB.QueryRow(
		`SELECT id, source_organization_id, jurisdiction_type, hub_city_id FROM church_jurisdictions WHERE source_organization_id = $1`,
		orgID,
	).Scan(&j.ID, &j.SourceOrganizationID, &j.JurisdictionType, &hubCity)
	j.HubCityID = hubCity.String
	return j, err
}

func (s *Store) UpsertParish(p ontology.Parish) error {
	_, err := s.DB.Exec(
		`INSERT INTO parishes (id, source_organization_id, jurisdiction_id, bulletin_endpoint_id) VALUES ($1,$2,$3,$4)
		 ON CONFLICT (id) DO UPDATE SET source_organization_id=EXCLUDED.source_organization_id,
		   jurisdiction_id=EXCLUDED.jurisdiction_id, bulletin_endpoint_id=EXCLUDED.bulletin_endpoint_id`,
		p.ID, p.SourceOrganizationID, p.JurisdictionID, nullIfEmpty(p.BulletinEndpointID),
	)
	return err
}

func (s *Store) ListParishesByJurisdiction(jurisdictionID string) ([]ontology.Parish, error) {
	rows, err := s.DB.Query(
		`SELECT id, source_organization_id, jurisdiction_id, COALESCE(bulletin_endpoint_id,'')
		 FROM parishes WHERE jurisdiction_id = $1`, jurisdictionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ontology.Parish
	for rows.Next() {
		var p ontology.Parish
		if err := rows.Scan(&p.ID, &p.SourceOrganizationID, &p.JurisdictionID, &p.BulletinEndpointID); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// --- social_sources ----------------------------------------------------------

func (s *Store) UpsertSocialSource(sc ontology.SocialSource) error {
	_, err := s.DB.Exec(
		`INSERT INTO social_sources (id, source_organization_id, platform, category, city_market_id, manually_verified, watched_source_id)
		 VALUES ($1,$2,$3,$4,$5,$6,$7)
		 ON CONFLICT (id) DO UPDATE SET source_organization_id=EXCLUDED.source_organization_id, platform=EXCLUDED.platform,
		   category=EXCLUDED.category, city_market_id=EXCLUDED.city_market_id, manually_verified=EXCLUDED.manually_verified,
		   watched_source_id=EXCLUDED.watched_source_id`,
		sc.ID, sc.SourceOrganizationID, sc.Platform, sc.Category, nullIfEmpty(sc.CityMarketID), sc.ManuallyVerified, nullIfEmpty(sc.WatchedSourceID),
	)
	return err
}

func (s *Store) ListSocialSourcesByCityMarket(cityID string) ([]ontology.SocialSource, error) {
	rows, err := s.DB.Query(
		`SELECT id, source_organization_id, platform, category, COALESCE(city_market_id,''), manually_verified, COALESCE(watched_source_id,'')
		 FROM social_sources WHERE city_market_id = $1`, cityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ontology.SocialSource
	for rows.Next() {
		var sc ontology.SocialSource
		if err := rows.Scan(&sc.ID, &sc.SourceOrganizationID, &sc.Platform, &sc.Category, &sc.CityMarketID, &sc.ManuallyVerified, &sc.WatchedSourceID); err != nil {
			return nil, err
		}
		out = append(out, sc)
	}
	return out, rows.Err()
}
