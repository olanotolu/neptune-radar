// The source registry store methods: geography backbone, real government/
// church/social organizations, their endpoints, and the connectors that
// check them. RecordConnectorRun is the only function that ever moves a
// connector's status away from "setup" — every other write here is
// structural (what exists), never a health claim.
package store

import (
	"database/sql"

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
