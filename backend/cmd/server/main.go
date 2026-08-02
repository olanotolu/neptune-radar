// Command server runs the Neptune Radar API and ingestion worker:
// Postgres-backed store, the Baseten-backed interpreter (template fallback),
// the full pipeline orchestrator, and the social watch loop (Bright Data primary).
package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"neptune-social-radar/backend/internal/api"
	"neptune-social-radar/backend/internal/ingest"
	"neptune-social-radar/backend/internal/llm"
	"neptune-social-radar/backend/internal/mail"
	"neptune-social-radar/backend/internal/notify"
	"neptune-social-radar/backend/internal/outreach"
	"neptune-social-radar/backend/internal/pipeline"
	"neptune-social-radar/backend/internal/ratelimit"
	"neptune-social-radar/backend/internal/records"
	"neptune-social-radar/backend/internal/store"
	"neptune-social-radar/backend/internal/vocab"
)

func main() {
	addr := flag.String("addr", envOr("ADDR", ":8080"), "HTTP listen address")
	pollInterval := flag.Duration("poll-interval", 15*time.Minute, "how often the watch loop polls each source")
	dryRun := flag.Bool("dry-run", false, "fetch from the provider but only log events, don't store or process them")
	bootstrapUser := flag.String("bootstrap-user", "", "create the first admin user (email) and print their API key, then exit")
	flag.Parse()

	dsn := os.Getenv("DATABASE_URL")
	s, err := store.Open(dsn)
	if err != nil {
		log.Fatalf("open store: %v", err)
	}
	defer s.Close()

	// Merge DB-driven signal vocabulary into the in-memory defaults so ops
	// can add new engagement phrases/hashtags without a redeploy.
	if err := vocab.LoadFromStore(s); err != nil {
		log.Printf("load external vocabulary: %v (continuing with defaults)", err)
	}

	// Bootstrap the first admin user. This is the only way to create the first
	// user since the API's POST /api/users requires an existing admin. After
	// this, the shared admin token stops working (UserCount > 0 switches auth
	// to per-user API keys).
	if *bootstrapUser != "" {
		user, plaintext, err := s.CreateUser(*bootstrapUser, *bootstrapUser, store.RoleAdmin)
		if err != nil {
			log.Fatalf("bootstrap user: %v", err)
		}
		log.Printf("created admin user %s (%s)", user.Email, user.ID)
		log.Printf("API key (store this — it won't be shown again): %s", plaintext)
		return
	}

	interp := llm.NewInterpreter()
	switch it := interp.(type) {
	case *llm.TemplateInterpreter:
		log.Println("No LLM API key configured — running with the deterministic template interpreter")
	case *llm.FallbackInterpreter:
		if it.HasBaseten() {
			log.Println("BASETEN: Relationship Analyst and Conversation Agent will call Baseten")
		}
		if it.HasClaude() {
			log.Println("ANTHROPIC: Claude available as fallback")
		}
		log.Println("Template fallback available on error")
	default:
		log.Println("LLM interpreter active with template fallback on error")
	}

	orch := pipeline.New(s, interp)

	// SSE live updates hub — pipeline events are pushed to dashboard subscribers.
	hub := notify.NewHub()
	orch.SetHub(hub)

	// Slack/webhook alerting for high-confidence couples. Empty URL = no-op.
	notifier := notify.NewNotifier(os.Getenv("SLACK_WEBHOOK_URL"))
	if notifier.Enabled() {
		orch.SetNotifier(notifier)
		log.Println("SLACK: webhook alerts enabled for high-confidence couples")
	}

	// The watch loop: monitors run on an interval and their events flow
	// straight through the pipeline. With no provider token it simply idles —
	// the API stays up and every other surface keeps working.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	provider := ingest.NewSocialProvider()
	log.Printf("social provider: %s (available=%v)", provider.Name(), provider.Available())
	worker := ingest.NewWorker(s, orch, ingest.WorkerConfig{
		PollInterval:  *pollInterval,
		DryRun:        *dryRun,
		DailyBudget:   envInt("DAILY_BUDGET_CAP", 500),
		ActiveMarkets: envList("ACTIVE_MARKETS"),
		Client:        provider,
		Vision:        llm.NewVisionClassifier(),
	})
	go worker.Run(ctx)

	adminToken := os.Getenv("NEPTUNE_ADMIN_TOKEN")
	if adminToken == "" {
		log.Println("WARNING: NEPTUNE_ADMIN_TOKEN is not set — all /api routes will return 503")
	}
	origin := envOr("DASHBOARD_ORIGIN", "http://localhost:5173")

	agent := &outreach.Agent{
		Store:   s,
		LLM:     interp,
		Records: records.NewMulti(),
		Mail:    mail.NewFromEnv(),
	}
	if agent.Mail != nil && agent.Mail.Available() {
		mode := "live"
		if agent.Mail.IsTest() {
			mode = "test"
		}
		log.Printf("LOB: postcard verify+send enabled (%s mode)", mode)
	} else {
		log.Println("LOB: not configured — set LOB_API_KEY (+ LOB_FROM_*) to send postcards")
	}
	if p := records.NewProvider(); p.Available() {
		log.Printf("records provider: %s", p.Name())
	} else {
		log.Println("records provider: heuristic only — set TRESTLE_API_KEY or PDL_API_KEY for street hits")
	}
	// Rate limiter: 10 req/sec per identity, burst of 20. Tunable via env.
	// ponytail: ceiling — in-memory, per-process. With leader election only
	// the leader polls, but all replicas serve the API, so the effective
	// per-user limit is replicas × burst. Fine at 5 users.
	limiter := ratelimit.New(envFloat("NEPTUNE_API_RATE", 10), envFloat("NEPTUNE_API_BURST", 20))

	handler := api.Wrap(api.NewRouter(s, worker, agent, hub), s, adminToken, origin, limiter)
	// In the deployed single-service image the Go server also serves the
	// built dashboard; in local dev the Vite dev server owns the frontend.
	if staticDir := os.Getenv("STATIC_DIR"); staticDir != "" {
		handler = withStatic(handler, staticDir)
		log.Printf("serving dashboard from %s", staticDir)
	}

	srv := &http.Server{
		Addr:              *addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	log.Printf("neptune-radar listening on %s (poll=%s dry_run=%v)", *addr, *pollInterval, *dryRun)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// withStatic serves the built dashboard (Vite output) at / with SPA fallback
// to index.html for client-side routes, while /api/* stays on the API.
func withStatic(apiHandler http.Handler, dir string) http.Handler {
	fileServer := http.FileServer(http.Dir(dir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			apiHandler.ServeHTTP(w, r)
			return
		}
		// Serve the file if it exists; otherwise fall back to index.html.
		path := filepath.Join(dir, filepath.Clean(r.URL.Path))
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			fileServer.ServeHTTP(w, r)
			return
		}
		http.ServeFile(w, r, filepath.Join(dir, "index.html"))
	})
}

func envInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func envFloat(key string, fallback float64) float64 {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return fallback
}

func envList(key string) []string {
	v := os.Getenv(key)
	if v == "" {
		return nil
	}
	var out []string
	for _, part := range strings.Split(v, ",") {
		if p := strings.TrimSpace(part); p != "" {
			out = append(out, p)
		}
	}
	return out
}
