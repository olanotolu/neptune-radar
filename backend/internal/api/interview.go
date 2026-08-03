package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"neptune-social-radar/backend/internal/llm"
	"neptune-social-radar/backend/internal/store"
)

var errEmptyText = errors.New("text is required")

// createInterviewSession starts a new two-couple interview session.
func (s *Server) createInterviewSession(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CoupleALabel string `json:"couple_a_label"`
		CoupleBLabel string `json:"couple_b_label"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req) // empty body is fine — defaults apply
	sess, err := s.Store.CreateInterviewSession(r.Context(), req.CoupleALabel, req.CoupleBLabel)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, sess)
}

// listInterviewSessions returns recent sessions.
func (s *Server) listInterviewSessions(w http.ResponseWriter, r *http.Request) {
	limit := 50
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	sessions, err := s.Store.ListInterviewSessions(r.Context(), limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, sessions)
}

// getInterviewSession returns a session with its messages and extractions.
func (s *Server) getInterviewSession(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	sess, err := s.Store.GetInterviewSession(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	messages, err := s.Store.ListInterviewMessages(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	extractions, err := s.Store.ListExtractions(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"session":     sess,
		"messages":    messages,
		"extractions": extractions,
	})
}

// addInterviewMessage appends one utterance to a session.
func (s *Server) addInterviewMessage(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		Speaker  string `json:"speaker"`
		Couple   string `json:"couple"`
		Text     string `json:"text"`
		AudioURL string `json:"audio_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if req.Text == "" {
		writeError(w, http.StatusBadRequest, errEmptyText)
		return
	}
	msg, err := s.Store.AddInterviewMessage(r.Context(), id, req.Speaker, req.Couple, req.Text, req.AudioURL)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, msg)
}

// runExtraction runs all extraction agents over the session's messages and
// persists each result.
func (s *Server) runExtraction(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	messages, err := s.Store.ListInterviewMessages(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	conv := make([]llm.ConversationMessage, 0, len(messages))
	for _, m := range messages {
		conv = append(conv, llm.ConversationMessage{Speaker: m.Speaker, Couple: m.Couple, Text: m.Text})
	}
	results, err := llm.RunExtractionAgents(r.Context(), conv)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, err)
		return
	}
	saved := make([]store.InterviewExtraction, 0, len(results))
	for _, res := range results {
		e, err := s.Store.SaveExtraction(r.Context(), id, res.AgentType, res.Findings, res.Confidence, res.Summary)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		saved = append(saved, e)
	}
	writeJSON(w, http.StatusOK, saved)
}

// endInterviewSession marks a session completed.
func (s *Server) endInterviewSession(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := s.Store.UpdateSessionStatus(r.Context(), id, "completed"); err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"id": id, "status": "completed"})
}
