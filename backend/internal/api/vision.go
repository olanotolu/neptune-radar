package api

import (
	"net/http"
	"strconv"
)

// visionAnalysis returns recent vision classification rows (ring detection +
// CLIP photo classification) for the dashboard. ?limit caps the result count.
func (s *Server) visionAnalysis(w http.ResponseWriter, r *http.Request) {
	limit := 100
	if q := r.URL.Query().Get("limit"); q != "" {
		if n, err := strconv.Atoi(q); err == nil && n > 0 {
			limit = n
		}
	}
	rows, err := s.Store.ListVisionAnalysis(limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, rows)
}
