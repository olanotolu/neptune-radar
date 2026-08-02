package api

import (
	"net/http"
	"strconv"

	"neptune-social-radar/backend/internal/auth"
)

// listNotifications returns notifications, optionally filtered to unread only.
func (s *Server) listNotifications(w http.ResponseWriter, r *http.Request) {
	unreadOnly := r.URL.Query().Get("unread") == "true"
	limit := 50
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			limit = n
		}
	}
	notifications, err := s.Store.ListNotifications(unreadOnly, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, notifications)
}

func (s *Server) markNotificationRead(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := s.Store.MarkNotificationRead(id); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"id": id, "status": "read"})
}

func (s *Server) ackNotification(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	ackedBy := "human:" + auth.UserFromContext(r.Context()).Email
	if err := s.Store.MarkNotificationAcked(id, ackedBy); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"id": id, "status": "acked"})
}

func (s *Server) markAllNotificationsRead(w http.ResponseWriter, r *http.Request) {
	n, err := s.Store.MarkAllNotificationsRead()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]int64{"read": n})
}
