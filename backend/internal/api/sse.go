package api

import (
	"encoding/json"
	"net/http"
	"time"
)

const sseHeartbeatInterval = 30 * time.Second

// eventsStream is the SSE endpoint. It subscribes to the hub, writes events
// as `data: {json}\n\n`, sends a heartbeat comment every 30s, and closes when
// the client disconnects (r.Context() cancellation).
func (s *Server) eventsStream(w http.ResponseWriter, r *http.Request) {
	if s.Hub == nil {
		writeError(w, http.StatusServiceUnavailable, errorString("event hub not configured"))
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, errorString("streaming unsupported"))
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", w.Header().Get("Access-Control-Allow-Origin"))

	ch := s.Hub.Subscribe()
	defer s.Hub.Unsubscribe(ch)

	heartbeat := time.NewTicker(sseHeartbeatInterval)
	defer heartbeat.Stop()

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case <-heartbeat.C:
			if _, err := w.Write([]byte(": heartbeat\n\n")); err != nil {
				return
			}
			flusher.Flush()
		case ev, ok := <-ch:
			if !ok {
				return
			}
			payload, err := json.Marshal(struct {
				Type string      `json:"type"`
				Data interface{} `json:"data"`
				Time time.Time   `json:"time"`
			}{ev.Type, ev.Data, ev.Time})
			if err != nil {
				continue
			}
			if _, err := w.Write([]byte("data: ")); err != nil {
				return
			}
			if _, err := w.Write(payload); err != nil {
				return
			}
			if _, err := w.Write([]byte("\n\n")); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

// pipelineMetrics returns avg duration per stage, the last 100 timings, and
// throughput (events/min over the last hour).
func (s *Server) pipelineMetrics(w http.ResponseWriter, r *http.Request) {
	summary, err := s.Store.GetPipelineSummary()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	recent, err := s.Store.GetStageTimings("", 100)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	throughput, err := s.Store.PipelineThroughput()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"avg_duration_ms": summary,
		"recent_timings":  recent,
		"events_per_min":  throughput,
	})
}
