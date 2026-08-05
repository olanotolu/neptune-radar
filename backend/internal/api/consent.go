package api

import (
	"encoding/json"
	"errors"
	"net/http"
)

// consentGrant is called when a couple lands on the celebrate page and consents
// to data processing. Public — couples reach it via the postcard QR code, no
// API key. Resolves the couple from the handoff_code, creates consent policies
// for both persons, records an audit event.
func (s *Server) consentGrant(w http.ResponseWriter, r *http.Request) {
	var req struct {
		HandoffCode    string   `json:"handoff_code"`
		ConsentActions []string `json:"consent_actions"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if req.HandoffCode == "" {
		writeError(w, http.StatusBadRequest, errors.New("handoff_code is required"))
		return
	}
	if len(req.ConsentActions) == 0 {
		// ponytail: default to the standard celebrate-flow actions so the
		// frontend can omit the list; ceiling — if new actions are added the
		// client must send them explicitly.
		req.ConsentActions = []string{"postcard", "follow_up", "data_processing"}
	}
	coupleID, err := s.resolveCoupleByHandoff(req.HandoffCode)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	if _, err := s.Store.CreateConsentForCouple(coupleID, req.ConsentActions); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	s.Store.Audit("couple", coupleID, "consent_granted",
		map[string]any{"handoff_code": req.HandoffCode, "actions": req.ConsentActions}, "consent", 0)
	names := s.coupleDisplayNames(coupleID)
	st, _ := s.Store.GetConsentStatus(coupleID)
	writeJSON(w, http.StatusOK, map[string]any{
		"couple_id": coupleID,
		"person_a":  names[0],
		"person_b":  names[1],
		"granted":   st.Granted,
		"actions":   st.AllowedActions,
	})
}

// consentRevoke is called when a couple opts out. Public. Revokes consent for
// both persons (RevokeConsent also suppresses the couple — no more postcards or
// follow-ups) and records an audit event.
func (s *Server) consentRevoke(w http.ResponseWriter, r *http.Request) {
	var req struct {
		HandoffCode string `json:"handoff_code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if req.HandoffCode == "" {
		writeError(w, http.StatusBadRequest, errors.New("handoff_code is required"))
		return
	}
	coupleID, err := s.resolveCoupleByHandoff(req.HandoffCode)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	c, err := s.Store.GetCouple(coupleID)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	// RevokeConsent suppresses the couple and cancels pending actions per person.
	s.Store.RevokeConsent(c.PersonAID)
	s.Store.RevokeConsent(c.PersonBID)
	s.Store.Audit("couple", coupleID, "consent_revoked",
		map[string]any{"handoff_code": req.HandoffCode}, "consent", 0)
	writeJSON(w, http.StatusOK, map[string]any{
		"couple_id": coupleID,
		"revoked":   true,
		"message":   "Your consent has been revoked. We will not contact you again.",
	})
}

// consentStatus returns the current consent state for a couple by handoff code.
// Public — lets the celebrate page decide which state to render.
func (s *Server) consentStatus(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("handoffCode")
	if code == "" {
		writeError(w, http.StatusBadRequest, errors.New("handoff code required"))
		return
	}
	coupleID, err := s.resolveCoupleByHandoff(code)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	st, err := s.Store.GetConsentStatus(coupleID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"couple_id":       coupleID,
		"granted":         st.Granted,
		"revoked":         st.Revoked,
		"allowed_actions": st.AllowedActions,
		"granted_at":      st.GrantedAt,
	})
}

// resolveCoupleByHandoff maps a handoff_code to its couple id.
func (s *Server) resolveCoupleByHandoff(code string) (string, error) {
	var coupleID string
	err := s.Store.DB.QueryRow(`SELECT id FROM couples WHERE handoff_code = $1`, code).Scan(&coupleID)
	if err != nil {
		return "", errors.New("no couple found for handoff code")
	}
	return coupleID, nil
}

// coupleDisplayNames returns [personA, personB] display names for a couple.
func (s *Server) coupleDisplayNames(coupleID string) [2]string {
	c, err := s.Store.GetCouple(coupleID)
	if err != nil {
		return [2]string{}
	}
	a, _ := s.Store.GetPerson(c.PersonAID)
	b, _ := s.Store.GetPerson(c.PersonBID)
	return [2]string{a.DisplayName, b.DisplayName}
}
