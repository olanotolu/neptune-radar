package store

import (
	"encoding/json"
	"time"
)

// VisionClassification records one vision-classifier call for calibration
// tracking. The labels returned are stored as JSON so we can audit the
// model's precision/recall over time.
type VisionClassification struct {
	ID               string  `json:"id"`
	ObservationID    string  `json:"observation_id,omitempty"`
	ExternalEventID  string  `json:"external_event_id,omitempty"`
	ImageURL         string  `json:"image_url"`
	Labels           string  `json:"labels"` // JSON array
	Model            string  `json:"model"`
	Error            string  `json:"error,omitempty"`
	RingConfidence   float64 `json:"ring_confidence,omitempty"`
	PhotoLabel       string  `json:"photo_label,omitempty"`
	PhotoConfidence  float64   `json:"photo_confidence,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
}

// RecordVisionClassification logs one vision call. labels is the JSON array
// of returned labels (empty array if none); errMsg is empty on success.
func (s *Store) RecordVisionClassification(observationID, externalEventID, imageURL, model string, labels []string, errMsg string) error {
	labelsJSON, _ := json.Marshal(labels)
	if errMsg == "" && labels == nil {
		labelsJSON = []byte("[]")
	}
	_, err := s.DB.Exec(
		`INSERT INTO vision_classifications (id, observation_id, external_event_id, image_url, labels, model, error) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		NewID("vis"), observationID, externalEventID, imageURL, string(labelsJSON), model, errMsg,
	)
	return err
}

// VisionClassificationStats returns calibration stats: total calls, success
// rate, and the most common labels. Used by the ops dashboard to spot a
// mis-calibrated model (e.g. one that never returns "ring" anymore).
type VisionClassificationStats struct {
	TotalCalls  int            `json:"total_calls"`
	ErrorCount  int            `json:"error_count"`
	LabelCounts map[string]int `json:"label_counts"`
}

func (s *Store) VisionClassificationStats() (VisionClassificationStats, error) {
	var stats VisionClassificationStats
	stats.LabelCounts = make(map[string]int)
	err := s.DB.QueryRow(`SELECT COUNT(*), COUNT(*) FILTER (WHERE error != '') FROM vision_classifications`).Scan(&stats.TotalCalls, &stats.ErrorCount)
	if err != nil {
		return stats, err
	}
	if stats.TotalCalls == 0 {
		return stats, nil
	}
	rows, err := s.DB.Query(`SELECT labels FROM vision_classifications WHERE error = ''`)
	if err != nil {
		return stats, err
	}
	defer rows.Close()
	for rows.Next() {
		var labelsJSON string
		if err := rows.Scan(&labelsJSON); err != nil {
			return stats, err
		}
		var labels []string
		if json.Unmarshal([]byte(labelsJSON), &labels) == nil {
			for _, l := range labels {
				stats.LabelCounts[l]++
			}
		}
	}
	return stats, rows.Err()
}

// RecordVisionAnalysis logs a combined vision call (visual labels + ring
// detection + photo classification) in one row. ringConfidence, photoLabel,
// and photoConfidence are 0/"" when those sub-analyses weren't run or failed.
func (s *Store) RecordVisionAnalysis(externalEventID, imageURL, model string, labels []string, errMsg string, ringConfidence float64, photoLabel string, photoConfidence float64) error {
	labelsJSON, _ := json.Marshal(labels)
	if errMsg == "" && labels == nil {
		labelsJSON = []byte("[]")
	}
	_, err := s.DB.Exec(
		`INSERT INTO vision_classifications (id, external_event_id, image_url, labels, model, error, ring_confidence, photo_label, photo_confidence)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		NewID("vis"), externalEventID, imageURL, string(labelsJSON), model, errMsg, ringConfidence, photoLabel, photoConfidence,
	)
	return err
}

// ListVisionAnalysis returns the most recent vision classification rows for
// the dashboard. limit caps the result count; 0 defaults to 100.
func (s *Store) ListVisionAnalysis(limit int) ([]VisionClassification, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := s.DB.Query(
		`SELECT id, COALESCE(observation_id,''), COALESCE(external_event_id,''), image_url, labels, model, COALESCE(error,''),
		        COALESCE(ring_confidence,0), COALESCE(photo_label,''), COALESCE(photo_confidence,0), created_at
		 FROM vision_classifications ORDER BY created_at DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []VisionClassification
	for rows.Next() {
		var v VisionClassification
		if err := rows.Scan(&v.ID, &v.ObservationID, &v.ExternalEventID, &v.ImageURL, &v.Labels, &v.Model, &v.Error,
			&v.RingConfidence, &v.PhotoLabel, &v.PhotoConfidence, &v.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
