// Package api is the HTTP surface the dashboard drives. Handlers are thin:
// they translate requests into store/operator/verifier calls and serialize
// the result. No pipeline logic lives here.
package api

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"neptune-social-radar/backend/internal/ingest"
	"neptune-social-radar/backend/internal/outreach"
	"neptune-social-radar/backend/internal/store"
)

type Server struct {
	Store    *store.Store
	Watch    *ingest.Worker // global watch-loop pause/play
	Outreach *outreach.Agent
}

func NewRouter(s *store.Store, worker *ingest.Worker, agent *outreach.Agent) http.Handler {
	srv := &Server{Store: s, Watch: worker, Outreach: agent}
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/signals", srv.listSignals)
	mux.HandleFunc("GET /api/couples", srv.listCouples)
	mux.HandleFunc("GET /api/couples/{id}/graph", srv.coupleGraph)
	mux.HandleFunc("GET /api/couples/{id}/relationship", srv.coupleRelationship)
	mux.HandleFunc("POST /api/couples/{id}/pause", srv.pauseCouple)
	mux.HandleFunc("POST /api/couples/{id}/resume", srv.resumeCouple)
	mux.HandleFunc("POST /api/couples/{id}/suppress", srv.suppressCouple)

	mux.HandleFunc("GET /api/ops/summary", srv.opsSummary)
	mux.HandleFunc("POST /api/prospects/enrich-missing", srv.enrichMissingProfiles)
	mux.HandleFunc("POST /api/prospects/backfill-locations", srv.backfillLocations)
	mux.HandleFunc("GET /api/hypotheses/{id}/evidence", srv.hypothesisEvidence)
	mux.HandleFunc("GET /api/hypotheses/{id}/confidence", srv.hypothesisConfidence)

	mux.HandleFunc("GET /api/cases", srv.listCases)
	mux.HandleFunc("GET /api/cases/{id}", srv.getCase)
	mux.HandleFunc("GET /api/leads", srv.listLeads)

	mux.HandleFunc("GET /api/actions", srv.listActions)
	mux.HandleFunc("POST /api/actions/{id}/approve", srv.approveAction)
	mux.HandleFunc("POST /api/actions/{id}/ignore", srv.ignoreAction)

	mux.HandleFunc("GET /api/sources", srv.listSources)
	mux.HandleFunc("POST /api/sources", srv.addSource)
	// Static paths under /api/sources must register before {handle} wildcards.
	mux.HandleFunc("POST /api/sources/scan-bulk", srv.scanBulk)
	mux.HandleFunc("DELETE /api/sources/{handle}", srv.removeSource)
	mux.HandleFunc("GET /api/sources/{handle}/posts", srv.listSourcePosts)
	mux.HandleFunc("POST /api/sources/{handle}/scan", srv.scanSource)
	mux.HandleFunc("POST /api/sources/{handle}/enrich", srv.enrichSource)
	mux.HandleFunc("PATCH /api/sources/{handle}/location", srv.patchSourceLocation)
	// Job polling lives outside /api/sources/{…} to avoid Go ServeMux wildcard conflicts
	// (e.g. scan-jobs/{id} vs {handle}/posts both match …/scan-jobs/posts).
	mux.HandleFunc("GET /api/scan-jobs/{id}", srv.scanJobStatus)
	mux.HandleFunc("POST /api/prospects/suppress-vendor-pairs", srv.suppressVendorPairs)
	mux.HandleFunc("GET /api/prospects/board", srv.listProspectBoard)
	mux.HandleFunc("GET /api/map/prospects", srv.listProspectPins)
	mux.HandleFunc("GET /api/ingest/status", srv.ingestStatus)
	mux.HandleFunc("POST /api/ingest/pause", srv.pauseIngest)
	mux.HandleFunc("POST /api/ingest/resume", srv.resumeIngest)

	mux.HandleFunc("GET /api/map/states/OH/overview", srv.mapOverview)
	mux.HandleFunc("GET /api/map/states/OH/government", srv.mapGovernment)
	mux.HandleFunc("GET /api/map/states/OH/churches", srv.mapChurches)
	mux.HandleFunc("GET /api/map/states/OH/social", srv.mapSocial)
	mux.HandleFunc("GET /api/map/organizations/{id}", srv.mapOrganization)
	mux.HandleFunc("GET /api/map/connectors/{id}", srv.mapConnector)
	mux.HandleFunc("GET /api/map/connectors/{id}/runs", srv.mapConnectorRuns)

	mux.HandleFunc("GET /api/audit", srv.listAudit)
	mux.HandleFunc("GET /api/health", srv.health)
	// Image proxy — Instagram CDN blocks cross-origin paint (CORP same-origin).
	mux.HandleFunc("GET /api/media", srv.mediaProxy)

	// Congratulate kits — dossier + address research + postcard (human-reviewed mail).
	mux.HandleFunc("GET /api/kits", srv.listCongratulateKits)
	mux.HandleFunc("POST /api/couples/{id}/congratulate", srv.buildCongratulateKit)
	mux.HandleFunc("GET /api/couples/{id}/kit", srv.latestKitForCouple)
	mux.HandleFunc("GET /api/kits/{id}", srv.getCongratulateKit)
	mux.HandleFunc("PATCH /api/kits/{id}", srv.patchCongratulateKit)
	mux.HandleFunc("POST /api/kits/{id}/ready-to-mail", srv.kitReadyToMail)
	mux.HandleFunc("POST /api/kits/{id}/mailed", srv.kitMarkMailed)
	mux.HandleFunc("GET /api/kits/{id}/postcard", srv.kitPostcardHTML)
	mux.HandleFunc("GET /api/kits/{id}/export", srv.kitMailExport)
	mux.HandleFunc("POST /api/kits/{id}/run-detective", srv.runDetective)
	mux.HandleFunc("POST /api/kits/{id}/apply-candidate", srv.applyKitCandidate)
	mux.HandleFunc("POST /api/kits/{id}/verify-address", srv.verifyKitAddress)
	mux.HandleFunc("POST /api/kits/{id}/send-postcard", srv.sendKitPostcard)
	mux.HandleFunc("GET /api/couples/{id}/dossier", srv.coupleDossier)
	mux.HandleFunc("POST /api/ops/janitor", srv.runJanitor)

	return mux
}

// Wrap applies the production middleware chain: origin-locked CORS, bearer
// auth, request logging.
func Wrap(handler http.Handler, adminToken, dashboardOrigin string) http.Handler {
	return withCORS(dashboardOrigin, withAuth(adminToken, withLogging(handler)))
}

func withCORS(origin string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// withAuth enforces a shared bearer token on every /api/* route except
// health. The dashboard is Neptune-internal; this is the whole auth model
// for now — per-user concierge identity is a PRODUCTION_GAPS item.
func withAuth(adminToken string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if adminToken == "" {
			writeError(w, http.StatusServiceUnavailable, errors.New("server misconfigured: NEPTUNE_ADMIN_TOKEN is not set"))
			return
		}
		// Health is public; media proxy is also public so <img> tags work
		// without Authorization headers (browsers never send our bearer on img).
		if r.URL.Path == "/api/health" || r.URL.Path == "/api/media" {
			next.ServeHTTP(w, r)
			return
		}
		const prefix = "Bearer "
		h := r.Header.Get("Authorization")
		if len(h) <= len(prefix) || h[:len(prefix)] != prefix || subtle.ConstantTimeCompare([]byte(h[len(prefix):]), []byte(adminToken)) != 1 {
			writeError(w, http.StatusUnauthorized, errors.New("missing or invalid bearer token"))
			return
		}
		next.ServeHTTP(w, r)
	})
}

func withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("encode response: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}
