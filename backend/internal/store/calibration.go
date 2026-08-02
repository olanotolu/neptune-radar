package store

import "time"

// CalibrationBand groups hypotheses by confidence range and shows the
// actual outcome distribution within that band. This is the core of the
// calibration cockpit: "when we said 80-90% confident, what actually happened?"
type CalibrationBand struct {
	MinConfidence  float64 `json:"min_confidence"`
	MaxConfidence  float64 `json:"max_confidence"`
	HypothesisCount int    `json:"hypothesis_count"`
	ApprovedCount  int    `json:"approved_count"`
	RejectedCount  int    `json:"rejected_count"`
	PendingCount   int    `json:"pending_count"`
	// Journey outcomes for couples linked to hypotheses in this band.
	Congratulated  int `json:"congratulated"`
	Invited       int `json:"invited"`
	InChat        int `json:"in_chat"`
	Booked        int `json:"booked"`
	ClosedWon     int `json:"closed_won"`
	ClosedLost    int `json:"closed_lost"`
}

// GetCalibrationData returns hypothesis confidence bands with actual
// outcome distribution. This connects "what we predicted" to "what happened"
// so the team can see if confidence is calibrated.
func (s *Store) GetCalibrationData() ([]CalibrationBand, error) {
	// 5 bands: 0-20%, 20-40%, 40-60%, 60-80%, 80-100%
	bands := []CalibrationBand{
		{MinConfidence: 0.0, MaxConfidence: 0.2},
		{MinConfidence: 0.2, MaxConfidence: 0.4},
		{MinConfidence: 0.4, MaxConfidence: 0.6},
		{MinConfidence: 0.6, MaxConfidence: 0.8},
		{MinConfidence: 0.8, MaxConfidence: 1.01},
	}
	for i := range bands {
		b := &bands[i]
		// Count hypotheses in this band.
		if err := s.DB.QueryRow(
			`SELECT COUNT(*),
			   COUNT(*) FILTER (WHERE status = 'confirmed'),
			   COUNT(*) FILTER (WHERE status = 'rejected'),
			   COUNT(*) FILTER (WHERE status = 'unconfirmed' OR status = 'corroborating')
			 FROM life_event_hypotheses
			 WHERE confidence >= $1 AND confidence < $2`,
			b.MinConfidence, b.MaxConfidence,
		).Scan(&b.HypothesisCount, &b.ApprovedCount, &b.RejectedCount, &b.PendingCount); err != nil {
			return nil, err
		}
		// Count journey outcomes for couples linked to hypotheses in this band.
		rows, err := s.DB.Query(
			`SELECT c.journey_stage, COUNT(*)
			 FROM couples c
			 JOIN life_event_hypotheses h ON h.couple_id = c.id
			 WHERE h.confidence >= $1 AND h.confidence < $2
			   AND c.suppressed_at IS NULL
			 GROUP BY c.journey_stage`, b.MinConfidence, b.MaxConfidence)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var stage string
			var n int
			if err := rows.Scan(&stage, &n); err != nil {
				rows.Close()
				return nil, err
			}
			switch stage {
			case "congratulated":
				b.Congratulated = n
			case "invited":
				b.Invited = n
			case "in_chat":
				b.InChat = n
			case "booked":
				b.Booked = n
			case "closed_won":
				b.ClosedWon = n
			case "closed_lost":
				b.ClosedLost = n
			}
		}
		rows.Close()
	}
	return bands, nil
}

// SourceYieldRow is per-source prospect yield for the calibration cockpit.
type SourceYieldRow struct {
	Monitor         string  `json:"monitor"`
	TotalSignals    int     `json:"total_signals"`
	ApprovedSignals int     `json:"approved_signals"`
	ApprovalRate    float64 `json:"approval_rate"`
}

// GetSourceYield returns per-monitor signal yield and approval rate.
func (s *Store) GetSourceYield() ([]SourceYieldRow, error) {
	rows, err := s.DB.Query(
		`SELECT COALESCE(monitor,''), COUNT(*) as total,
		   COUNT(*) FILTER (WHERE status = 'confirmed') as approved
		 FROM life_event_hypotheses
		 WHERE created_at > $1
		 GROUP BY COALESCE(monitor,'')
		 ORDER BY total DESC`,
		time.Now().UTC().AddDate(0, 0, -30))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []SourceYieldRow
	for rows.Next() {
		var r SourceYieldRow
		if err := rows.Scan(&r.Monitor, &r.TotalSignals, &r.ApprovedSignals); err != nil {
			return nil, err
		}
		if r.TotalSignals > 0 {
			r.ApprovalRate = float64(r.ApprovedSignals) / float64(r.TotalSignals)
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
