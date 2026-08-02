package api

import (
	"encoding/csv"
	"net/http"
	"strconv"
	"time"
)

// exportCouples streams couples as CSV: id, partner_a_handle, partner_b_handle,
// stage, confidence, city, state, created_at. Optional ?state=<USPS>&stage=<stage>.
func (s *Server) exportCouples(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("format") != "csv" {
		writeError(w, http.StatusBadRequest, errorString("format=csv is required"))
		return
	}
	rows, err := s.Store.ExportCouples(r.URL.Query().Get("state"), r.URL.Query().Get("stage"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", `attachment; filename="couples.csv"`)
	cw := csv.NewWriter(w)
	_ = cw.Write([]string{"id", "partner_a_handle", "partner_b_handle", "stage", "confidence", "city", "state", "created_at"})
	for _, r := range rows {
		_ = cw.Write([]string{
			r.ID, r.PartnerAHandle, r.PartnerBHandle, r.Stage,
			strconv.FormatFloat(r.Confidence, 'f', -1, 64),
			r.City, r.State, r.CreatedAt.Format(time.RFC3339),
		})
	}
	cw.Flush()
}

// exportLeads streams leads as CSV: id, couple_id, name, handle, stage, created_at.
func (s *Server) exportLeads(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("format") != "csv" {
		writeError(w, http.StatusBadRequest, errorString("format=csv is required"))
		return
	}
	rows, err := s.Store.ExportLeads()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", `attachment; filename="leads.csv"`)
	cw := csv.NewWriter(w)
	_ = cw.Write([]string{"id", "couple_id", "name", "handle", "stage", "created_at"})
	for _, r := range rows {
		_ = cw.Write([]string{r.ID, r.CoupleID, r.Name, r.Handle, r.Stage, r.CreatedAt.Format(time.RFC3339)})
	}
	cw.Flush()
}

// exportAudit streams audit events as CSV: id, entity_type, entity_id, event,
// monitor, created_at. Optional ?limit=1000.
func (s *Server) exportAudit(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("format") != "csv" {
		writeError(w, http.StatusBadRequest, errorString("format=csv is required"))
		return
	}
	limit := 1000
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			limit = n
		}
	}
	rows, err := s.Store.ExportAudit(limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", `attachment; filename="audit.csv"`)
	cw := csv.NewWriter(w)
	_ = cw.Write([]string{"id", "entity_type", "entity_id", "event", "monitor", "created_at"})
	for _, r := range rows {
		_ = cw.Write([]string{r.ID, r.EntityType, r.EntityID, r.Event, r.Monitor, r.CreatedAt.Format(time.RFC3339)})
	}
	cw.Flush()
}
