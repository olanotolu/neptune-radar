package api

import (
	"net/http"
	"strconv"

	"neptune-social-radar/backend/internal/store"
)

// search is the universal search endpoint: ?q=<query>&type=couples|leads|cases|all
// &state=<USPS>&min_confidence=<0-1>. type=all returns grouped results; a single
// type returns just that array.
func (s *Server) search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		writeError(w, http.StatusBadRequest, errorString("q (query) is required"))
		return
	}
	p := store.SearchParams{
		Query:        q,
		Type:         r.URL.Query().Get("type"),
		State:        r.URL.Query().Get("state"),
		MinConfidence: 0,
		Limit:        50,
	}
	if v := r.URL.Query().Get("min_confidence"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			p.MinConfidence = f
		}
	}
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			p.Limit = n
		}
	}
	if p.Type == "" {
		p.Type = "all"
	}
	res, err := s.Store.Search(p)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	// type=all → grouped object; single type → just that array.
	switch p.Type {
	case "couples":
		writeJSON(w, http.StatusOK, res.Couples)
	case "leads":
		writeJSON(w, http.StatusOK, res.Leads)
	case "cases":
		writeJSON(w, http.StatusOK, res.Cases)
	default:
		writeJSON(w, http.StatusOK, res)
	}
}
