package api

import (
	"net/http"
)

// organismStatus is the agentic Meet Neptune growth OS: swarm, guarantees,
// yield close-loop, risk sentinel, and morning briefing.
// GET /api/organism
func (s *Server) organismStatus(w http.ResponseWriter, r *http.Request) {
	o, err := s.Store.GetOrganism()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, o)
}
