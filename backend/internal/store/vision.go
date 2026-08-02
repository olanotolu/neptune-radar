package store

import "encoding/json"

// VisionClassification records one vision-classifier call for calibration
// tracking. The labels returned are stored as JSON so we can audit the
// model's precision/recall over time.
type VisionClassification struct {
	ID               string `json:"id"`
	ObservationID    string `json:"observation_id,omitempty"`
	ExternalEventID  string `json:"external_event_id,omitempty"`
	ImageURL         string `json:"image_url"`
	Labels           string `json:"labels"` // JSON array
	Model            string `json:"model"`
	Error            string `json:"error,omitempty"`
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
