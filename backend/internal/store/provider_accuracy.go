package store

import (
	"strings"
)

// ProviderAccuracy is one provider×state accuracy row for the dashboard.
type ProviderAccuracy struct {
	Provider      string  `json:"provider"`
	State         string  `json:"state"`
	TotalAttempts int     `json:"total_attempts"`
	Successful    int     `json:"successful"`
	Accuracy      float64 `json:"accuracy"`
}

// RecordProviderAttempt increments the provider×state counter. success=true
// counts as a hit. Idempotent upsert via ON CONFLICT.
func (s *Store) RecordProviderAttempt(provider, state string, success bool) {
	provider = strings.TrimSpace(provider)
	state = strings.ToUpper(strings.TrimSpace(state))
	if provider == "" || state == "" {
		return
	}
	hit := 0
	if success {
		hit = 1
	}
	_, _ = s.DB.Exec(
		`INSERT INTO provider_accuracy (provider_name, state, total_attempts, successful)
		 VALUES ($1, $2, 1, $3)
		 ON CONFLICT (provider_name, state) DO UPDATE SET
		     total_attempts = provider_accuracy.total_attempts + 1,
		     successful     = provider_accuracy.successful + $3,
		     last_updated   = now()`,
		provider, state, hit,
	)
}

// GetProviderAccuracy returns accuracy per provider per state:
// map[provider]map[state]float64. Empty map on cold start (caller uses 0.5 prior).
func (s *Store) GetProviderAccuracy() (map[string]map[string]float64, error) {
	rows, err := s.DB.Query(
		`SELECT provider_name, state, total_attempts, successful FROM provider_accuracy`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make(map[string]map[string]float64)
	for rows.Next() {
		var p, st string
		var total, success int
		if err := rows.Scan(&p, &st, &total, &success); err != nil {
			return nil, err
		}
		if total <= 0 {
			continue
		}
		if out[p] == nil {
			out[p] = make(map[string]float64)
		}
		out[p][strings.ToUpper(st)] = float64(success) / float64(total)
	}
	return out, rows.Err()
}

// ListProviderAccuracy returns flat rows for the dashboard endpoint.
func (s *Store) ListProviderAccuracy() ([]ProviderAccuracy, error) {
	rows, err := s.DB.Query(
		`SELECT provider_name, state, total_attempts, successful FROM provider_accuracy
		 ORDER BY provider_name, state`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ProviderAccuracy
	for rows.Next() {
		var pa ProviderAccuracy
		if err := rows.Scan(&pa.Provider, &pa.State, &pa.TotalAttempts, &pa.Successful); err != nil {
			return nil, err
		}
		if pa.TotalAttempts > 0 {
			pa.Accuracy = float64(pa.Successful) / float64(pa.TotalAttempts)
		}
		out = append(out, pa)
	}
	return out, rows.Err()
}
