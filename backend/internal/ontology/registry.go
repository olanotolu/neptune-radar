// Package ontology (continued) — the source registry: real government,
// church, and social/wedding-industry entities Neptune monitors, plus the
// connectors that check them and the runs that record what actually
// happened. Government and church are new source hierarchies with no
// extraction pipeline yet (see ConnectorStatus); Instagram/social reuses the
// existing WatchedSource below via SocialSource.WatchedSourceID rather than a
// parallel table.
package ontology

import "time"

// ConnectorStatus reports how far Neptune has actually gotten with a
// monitored target. It must never be set to anything but ConnectorSetup
// except as the direct result of a real ConnectorRun.
type ConnectorStatus string

const (
	ConnectorSetup    ConnectorStatus = "setup"    // registered, no successful check has run yet
	ConnectorHealthy  ConnectorStatus = "healthy"  // most recent check succeeded
	ConnectorDegraded ConnectorStatus = "degraded" // reachable but recent checks are failing or stale
	ConnectorOffline  ConnectorStatus = "offline"  // repeated hard failures
)

type DataMode string

const (
	DataModeLive           DataMode = "live"
	DataModeVerifiedImport DataMode = "verified_import"
	DataModeManual         DataMode = "manual"
	DataModeFixture        DataMode = "fixture"
)

type State struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type County struct {
	ID      string `json:"id"`
	StateID string `json:"state_id"`
	Name    string `json:"name"`
}

type City struct {
	ID              string   `json:"id"`
	StateID         string   `json:"state_id"`
	PrimaryCountyID string   `json:"primary_county_id,omitempty"`
	Name            string   `json:"name"`
	Lat             *float64 `json:"lat,omitempty"`
	Lng             *float64 `json:"lng,omitempty"`
}

// SourceOrganization is a real government office, diocese, parish, or
// business — the "who" a source_endpoint belongs to.
type SourceOrganization struct {
	ID          string    `json:"id"`
	OrgType     string    `json:"org_type"` // government_office | diocese | parish | business
	Name        string    `json:"name"`
	CityID      string    `json:"city_id,omitempty"`
	CountyID    string    `json:"county_id,omitempty"`
	OfficialURL string    `json:"official_url,omitempty"`
	Provenance  string    `json:"provenance"`
	DataMode    DataMode  `json:"data_mode"`
	Metadata    string    `json:"metadata,omitempty"` // JSON: org_type-specific facts (phone, coverage dates, deanery, ...)
	CreatedAt   time.Time `json:"created_at"`
}

// SourceEndpoint is a real, checkable URL belonging to an organization — a
// marriage-record search page, a bulletin archive, or a social profile.
type SourceEndpoint struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organization_id"`
	EndpointType   string    `json:"endpoint_type"` // marriage_record_search | bulletin_archive | social_profile
	URL            string    `json:"url"`
	AccessMethod   string    `json:"access_method"`
	IsOfficial     bool      `json:"is_official"`
	DataMode       DataMode  `json:"data_mode"`
	CreatedAt      time.Time `json:"created_at"`
}

// Connector is the thing that actually checks a source_endpoint. Status only
// ever changes as the result of a real ConnectorRun — see internal/connectors.
type Connector struct {
	ID               string          `json:"id"`
	SourceEndpointID string          `json:"source_endpoint_id"`
	ConnectorType    string          `json:"connector_type"` // http_health | bulletin_discovery | apify_instagram
	Provider         string          `json:"provider"`
	Status           ConnectorStatus `json:"status"`
	Schedule         string          `json:"schedule,omitempty"`
	LastCheckedAt    *time.Time      `json:"last_checked_at,omitempty"`
	LastSuccessAt    *time.Time      `json:"last_success_at,omitempty"`
	LastFailureAt    *time.Time      `json:"last_failure_at,omitempty"`
	ErrorMessage     string          `json:"error_message,omitempty"`
	CreatedAt        time.Time       `json:"created_at"`
}

// ConnectorRun is one real check execution — the measurement record.
type ConnectorRun struct {
	ID                 string     `json:"id"`
	ConnectorID        string     `json:"connector_id"`
	StartedAt          time.Time  `json:"started_at"`
	CompletedAt        *time.Time `json:"completed_at,omitempty"`
	Status             string     `json:"status"` // success | failure
	HTTPStatus         int        `json:"http_status,omitempty"`
	ResponseTimeMs     int        `json:"response_time_ms,omitempty"`
	StructureSignature string     `json:"structure_signature,omitempty"`
	ErrorMessage       string     `json:"error_message,omitempty"`
}

// ChurchJurisdiction is a real Catholic diocese or archdiocese — a regional
// hub marker on the map (cathedral-city coordinates), not a border polygon.
type ChurchJurisdiction struct {
	ID                   string `json:"id"`
	SourceOrganizationID string `json:"source_organization_id"`
	JurisdictionType     string `json:"jurisdiction_type"` // diocese | archdiocese
	HubCityID            string `json:"hub_city_id,omitempty"`
}

// Parish is a real parish within a jurisdiction, optionally linked to a
// discovered bulletin-archive endpoint.
type Parish struct {
	ID                   string `json:"id"`
	SourceOrganizationID string `json:"source_organization_id"`
	JurisdictionID       string `json:"jurisdiction_id"`
	BulletinEndpointID   string `json:"bulletin_endpoint_id,omitempty"`
}

// SocialSource links a real wedding-industry business to the actual
// WatchedSource row the existing Apify pipeline polls — this is never a
// second, disconnected source of truth for whether an account is monitored.
type SocialSource struct {
	ID                   string `json:"id"`
	SourceOrganizationID string `json:"source_organization_id"`
	Platform             string `json:"platform"`
	Category             string `json:"category"` // one of signals.WatchedSourceClasses
	CityMarketID         string `json:"city_market_id,omitempty"`
	ManuallyVerified     bool   `json:"manually_verified"`
	WatchedSourceID      string `json:"watched_source_id,omitempty"`
}
