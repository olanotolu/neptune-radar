package api

import (
	"net/http"

	"neptune-social-radar/backend/internal/auth"
	"neptune-social-radar/backend/internal/llm"
)

// calibration returns confidence bands with actual outcome distribution.
func (s *Server) calibration(w http.ResponseWriter, r *http.Request) {
	bands, err := s.Store.GetCalibrationData()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, bands)
}

// sourceYield returns per-monitor signal yield and approval rate.
func (s *Server) sourceYield(w http.ResponseWriter, r *http.Request) {
	yield, err := s.Store.GetSourceYield()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, yield)
}

// providerAccuracy returns per-provider accuracy scores by state for the
// Bayesian Provider Fusion dashboard.
func (s *Server) providerAccuracy(w http.ResponseWriter, r *http.Request) {
	rows, err := s.Store.ListProviderAccuracy()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, rows)
}

// runEval runs the golden eval cases against the current interpreter and
// returns the results. Admin-only — this makes LLM calls that cost money.
func (s *Server) runEval(w http.ResponseWriter, r *http.Request) {
	if !auth.RequireAdmin(w, r) {
		return
	}
	interp := llm.NewInterpreter()
	results := llm.RunEval(r.Context(), interp)
	writeJSON(w, http.StatusOK, results)
}

// handoffPacket returns the evidence-cited alignment narrative for a couple.
func (s *Server) handoffPacket(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	pkt, err := s.Store.GenerateHandoffPacket(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, pkt)
}
