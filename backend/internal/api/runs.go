package api

import (
	"errors"
	"net/http"
	"strconv"

	"neptune-social-radar/backend/internal/store"
)

// listRuns returns pipeline run summary rows, newest first. Optional query
// params: couple_id, monitor, stop_reason, limit.
func (s *Server) listRuns(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	f := store.RunFilter{
		CoupleID:   q.Get("couple_id"),
		Monitor:    q.Get("monitor"),
		StopReason: q.Get("stop_reason"),
	}
	if n, err := strconv.Atoi(q.Get("limit")); err == nil && n > 0 {
		f.Limit = n
	}
	runs, err := s.Store.ListPipelineRuns(f)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, runs)
}

// getRun returns one run by id, with the per-stage timings and audit events
// that belong to it so the viewer can render the full trace in one round-trip.
func (s *Server) getRun(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, errors.New("missing run id"))
		return
	}
	detail, err := s.Store.GetPipelineRunDetail(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeJSON(w, http.StatusOK, detail)
}
