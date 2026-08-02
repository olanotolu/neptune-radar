package store

// KitStats summarizes the celebration operations pipeline — how many kits
// are at each stage, how many are ready to mail, how many are overdue.
type KitStats struct {
	ReadyToMail  int `json:"ready_to_mail"`
	Mailed       int `json:"mailed"`
	Verified     int `json:"verified"`
	PendingAddress int `json:"pending_address"`
	Overdue      int `json:"overdue"` // ready but not mailed in 7+ days
	Total        int `json:"total"`
}

// GetKitStats returns a summary of the celebration operations pipeline.
func (s *Store) GetKitStats() (KitStats, error) {
	var stats KitStats
	// Count by status.
	rows, err := s.DB.Query(`
		SELECT status, COUNT(*)
		FROM congratulate_kits
		GROUP BY status`)
	if err != nil {
		return stats, err
	}
	for rows.Next() {
		var status string
		var n int
		if err := rows.Scan(&status, &n); err != nil {
			rows.Close()
			return stats, err
		}
		stats.Total += n
		switch status {
		case "ready_to_mail":
			stats.ReadyToMail = n
		case "mailed":
			stats.Mailed = n
		case "verified":
			stats.Verified = n
		case "pending_address":
			stats.PendingAddress = n
		}
	}
	rows.Close()
	// Count overdue: ready_to_mail but created > 7 days ago.
	err = s.DB.QueryRow(`
		SELECT COUNT(*) FROM congratulate_kits
		WHERE status = 'ready_to_mail'
		  AND created_at < now() - interval '7 days'`).Scan(&stats.Overdue)
	return stats, err
}
