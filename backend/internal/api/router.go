// Package api is the HTTP surface the dashboard drives. Handlers are thin:
// they translate requests into store/operator/verifier calls and serialize
// the result. No pipeline logic lives here.
package api

import (
	"encoding/json"
	"log"
	"net/http"

	"neptune-social-radar/backend/internal/auth"
	"neptune-social-radar/backend/internal/ingest"
	"neptune-social-radar/backend/internal/notify"
	"neptune-social-radar/backend/internal/outreach"
	"neptune-social-radar/backend/internal/ratelimit"
	"neptune-social-radar/backend/internal/store"
)

type Server struct {
	Store    *store.Store
	Watch    *ingest.Worker // global watch-loop pause/play
	Outreach *outreach.Agent
	Hub      *notify.Hub // SSE live updates (nil = endpoint returns 503)
}

func NewRouter(s *store.Store, worker *ingest.Worker, agent *outreach.Agent, hub *notify.Hub) http.Handler {
	srv := &Server{Store: s, Watch: worker, Outreach: agent, Hub: hub}
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/signals", srv.listSignals)
	mux.HandleFunc("GET /api/couples", srv.listCouples)
	mux.HandleFunc("GET /api/couples/{id}/graph", srv.coupleGraph)
	mux.HandleFunc("GET /api/couples/{id}/relationship", srv.coupleRelationship)
	mux.HandleFunc("POST /api/couples/{id}/pause", srv.pauseCouple)
	mux.HandleFunc("POST /api/couples/{id}/resume", srv.resumeCouple)
	mux.HandleFunc("POST /api/couples/{id}/suppress", srv.suppressCouple)
	mux.HandleFunc("POST /api/couples/bulk", srv.bulkCouple)

	mux.HandleFunc("GET /api/ops/summary", srv.opsSummary)
	mux.HandleFunc("POST /api/prospects/enrich-missing", srv.enrichMissingProfiles)
	mux.HandleFunc("POST /api/prospects/backfill-locations", srv.backfillLocations)
	mux.HandleFunc("GET /api/hypotheses/{id}/evidence", srv.hypothesisEvidence)
	mux.HandleFunc("GET /api/hypotheses/{id}/confidence", srv.hypothesisConfidence)

	mux.HandleFunc("GET /api/cases", srv.listCases)
	mux.HandleFunc("GET /api/cases/{id}", srv.getCase)
	mux.HandleFunc("GET /api/leads", srv.listLeads)

	// Universal search across couples, leads, and cases.
	mux.HandleFunc("GET /api/search", srv.search)

	// CSV exports for operational reporting.
	mux.HandleFunc("GET /api/export/couples", srv.exportCouples)
	mux.HandleFunc("GET /api/export/leads", srv.exportLeads)
	mux.HandleFunc("GET /api/export/audit", srv.exportAudit)

	// Dead-letter queue visibility + replay/retry controls.
	mux.HandleFunc("GET /api/dlq", srv.listDLQ)
	mux.HandleFunc("POST /api/dlq/{id}/replay", srv.replayDLQ)
	mux.HandleFunc("POST /api/dlq/{id}/retry", srv.retryDLQ)

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
	// GET /api/scan-jobs (list) is a distinct pattern from /api/scan-jobs/{id}.
	mux.HandleFunc("GET /api/scan-jobs", srv.listScanJobs)
	mux.HandleFunc("GET /api/scan-jobs/{id}", srv.scanJobStatus)
	mux.HandleFunc("POST /api/prospects/suppress-vendor-pairs", srv.suppressVendorPairs)
	mux.HandleFunc("GET /api/prospects/board", srv.listProspectBoard)
	mux.HandleFunc("GET /api/map/prospects", srv.listProspectPins)
	mux.HandleFunc("GET /api/ingest/status", srv.ingestStatus)
	mux.HandleFunc("POST /api/ingest/pause", srv.pauseIngest)
	mux.HandleFunc("POST /api/ingest/resume", srv.resumeIngest)

	// National map: any USPS state (seed-geography required). OH legacy paths
	// remain identical via {state}=OH.
	mux.HandleFunc("GET /api/map/coverage/summary", srv.mapCoverageSummary)
	mux.HandleFunc("GET /api/map/coverage", srv.mapNationalCoverage)
	mux.HandleFunc("GET /api/map/states/{state}/overview", srv.mapOverview)
	mux.HandleFunc("GET /api/map/states/{state}/government", srv.mapGovernment)
	mux.HandleFunc("GET /api/map/states/{state}/churches", srv.mapChurches)
	mux.HandleFunc("GET /api/map/states/{state}/social", srv.mapSocial)
	mux.HandleFunc("GET /api/map/organizations/{id}", srv.mapOrganization)
	mux.HandleFunc("GET /api/map/connectors/{id}", srv.mapConnector)
	mux.HandleFunc("GET /api/map/connectors/{id}/runs", srv.mapConnectorRuns)

	mux.HandleFunc("GET /api/audit", srv.listAudit)
	mux.HandleFunc("GET /api/health", srv.health)
	mux.HandleFunc("GET /api/events/stream", srv.eventsStream)
	mux.HandleFunc("GET /api/pipeline/metrics", srv.pipelineMetrics)
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
	mux.HandleFunc("POST /api/couples/{id}/handoff", srv.createHandoff)
	mux.HandleFunc("POST /api/couples/{id}/journey", srv.setJourneyStage)
	mux.HandleFunc("POST /api/ops/janitor", srv.runJanitor)

	// Congratulate kit upgrades — templates, county lookup, batch detect, operator queue, follow-up
	mux.HandleFunc("GET /api/kits/templates", srv.listGreetingTemplates)
	mux.HandleFunc("POST /api/kits/{id}/apply-template", srv.applyGreetingTemplate)
	mux.HandleFunc("GET /api/kits/{id}/county-records", srv.countyRecordLinks)
	mux.HandleFunc("POST /api/kits/batch-detective", srv.batchDetective)
	mux.HandleFunc("POST /api/kits/batch-verify", srv.batchVerifyAddresses)
	mux.HandleFunc("GET /api/kits/operator-queue", srv.operatorQueue)
	mux.HandleFunc("GET /api/kits/follow-up-queue", srv.followUpQueue)
	mux.HandleFunc("POST /api/kits/{id}/send-follow-up", srv.sendFollowUp)

	// Closed-loop funnel (Meet Neptune app → Radar) + trust autopsy
	mux.HandleFunc("POST /api/webhooks/neptune", srv.ingestFunnelWebhook)
	mux.HandleFunc("GET /api/funnel/events", srv.listFunnelEvents)
	mux.HandleFunc("GET /api/funnel/stats", srv.funnelStats)
	mux.HandleFunc("GET /api/trust/autopsies", srv.listAutopsies)
	mux.HandleFunc("POST /api/trust/autopsy", srv.generateAutopsy)
	mux.HandleFunc("GET /api/trust/autopsies/{id}", srv.getAutopsy)

	// User management (admin-only — handler checks role).
	mux.HandleFunc("GET /api/users", srv.listUsers)
	mux.HandleFunc("POST /api/users", srv.createUser)
	mux.HandleFunc("POST /api/users/{id}/rotate-key", srv.rotateAPIKey)
	mux.HandleFunc("POST /api/users/{id}/disable", srv.disableUser)
	mux.HandleFunc("POST /api/users/{id}/enable", srv.enableUser)

	// DSAR: GDPR/CCPA right-to-erasure (admin-only — handler checks role).
	mux.HandleFunc("POST /api/dsar/delete", srv.dsarDelete)

	// Identity override: human marks couple as mistaken or hypothesis as rejected.
	mux.HandleFunc("POST /api/couple/mistaken", srv.markCoupleMistaken)
	mux.HandleFunc("POST /api/hypothesis/reject", srv.rejectHypothesis)

	// Agent run ledger — one summary row per pipeline execution, with the
	// per-stage audit/timing detail joinable by observation_id.
	mux.HandleFunc("GET /api/runs", srv.listRuns)
	mux.HandleFunc("GET /api/runs/{id}", srv.getRun)

	// Retention policy — admin-only config + purge preview.
	mux.HandleFunc("GET /api/retention", srv.listRetention)
	mux.HandleFunc("PUT /api/retention", srv.setRetention)
	mux.HandleFunc("GET /api/retention/preview", srv.purgePreview)

	return mux
}

// Wrap applies the production middleware chain: origin-locked CORS, per-user
// auth (with legacy shared-token fallback), rate limiting, request logging.
func Wrap(handler http.Handler, s *store.Store, adminToken, dashboardOrigin string, limiter *ratelimit.Limiter) http.Handler {
	return withCORS(dashboardOrigin, auth.Middleware(s, adminToken, ratelimit.Middleware(limiter, withLogging(handler))))
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
