package api

import (
	"net/http"
	"strconv"

	"neptune-social-radar/backend/internal/ingest"
)

// listScanJobs returns recent scan jobs. ?limit=50&status=running|completed|failed.
// Registered at GET /api/scan-jobs (distinct from the existing GET /api/scan-jobs/{id}).
func (s *Server) listScanJobs(w http.ResponseWriter, r *http.Request) {
	if s.Watch == nil {
		writeError(w, http.StatusServiceUnavailable, errorString("watch loop not configured"))
		return
	}
	limit := 50
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			limit = n
		}
	}
	jobs := s.Watch.ListScanJobs(r.URL.Query().Get("status"), limit)
	if jobs == nil {
		jobs = []ingest.ScanJob{}
	}
	writeJSON(w, http.StatusOK, jobs)
}
