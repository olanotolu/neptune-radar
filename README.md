# Neptune Radar

An agentic CRM that watches authorized/curated social signals, maintains a
living relationship ontology for couples, and turns detected life events into
governed CRM/workflow actions — always with a human in the loop for anything
customer-facing.

**This is the production build.** A background watch loop ingests real posts
from a data provider (Apify) on an interval; the deterministic signal
vocabulary scores them; prospects land in an approval queue. There is no demo
mode, no replay, no seeded fixtures.

## The two triggers, and why they're different

- **Engagement detected** → possible prenup opportunity → internal alert with
  `[approve] [ignore]`.
- **Unfollow / relationship-state change detected** → the *existing* workflow may
  need to slow down → internal alert recommending "pause automation + concierge
  review", never a pitch.

You do not offer someone a prenup because they unfollowed their partner. The two
hypothesis types are scored independently and routed to different actions —
confusing them is structurally impossible.

## How the radar reads posts

Not hashtags alone (Instagram capped those at five per post in 2026). The
`internal/signals` package classifies a **combined signal vocabulary**: explicit
caption phrases (word-boundary matched, with the model judging semantic
variations), tiered hashtags (high-intent / supporting / status / vendor /
location / inclusive / cultural / weak / exclusion), image tags AND caption
@mentions, curated vendor source accounts, visual/on-screen signals (a
multimodal model on Baseten, gated so it only runs on candidate posts),
locations, and linked evidence (registry matches).

Scoring is an explicit points table: +40 explicit language, +25 both partners
referenced, +15 known vendor, +15 registry match, +10 visual / reciprocal /
recent — and −50 styled shoot, −50 advertisement, −30 no identifiable second
person, −25 old/reposted, −40 conflicting identity.

- **90+** → create-prospect card in the approval queue
- **70–89** → human investigation queue
- **below 70** → retained as an unconfirmed inference, never surfaced

Exclusion signals never mint new couples at the identity gate, and when they
hit an already-known couple they land on the audit trail as −50 evidence rows.
The model's only scoring influence is crediting explicit language the
deterministic phrase list missed; it cannot invent points, and the policy and
operator packages are build-enforced (`go list -deps` tests) to never import
the LLM package.

## Running it

Requires Go 1.24+, Node 22+, and Postgres 16+.

```bash
cp .env.example backend/.env   # fill in DATABASE_URL, NEPTUNE_ADMIN_TOKEN, APIFY_TOKEN, BASETEN_API_KEY
make test-db                   # throwaway local Postgres in Docker (prints the DSNs to export)
make backend                   # API + watch loop on :8080
make frontend                  # dashboard on :5173 (dev)
```

Open `http://localhost:5173`, enter the admin token, add a watched vendor under
**Sources**, and the next poll tick starts flowing the feed. With no
`APIFY_TOKEN` the watch loop idles and the API/dashboard still run.

### Key configuration

| Env | Effect |
|---|---|
| `DATABASE_URL` | Postgres connection string (required) |
| `NEPTUNE_ADMIN_TOKEN` | Bearer token for all `/api/*` routes and the dashboard (required) |
| `APIFY_TOKEN` | Apify provider token; unset = watch loop idle |
| `BASETEN_API_KEY` / `BASETEN_MODEL` | Analyst + copy model (falls back to deterministic template on error/absence) |
| `BASETEN_VISION_MODEL` | Multimodal model for ring/proposal visual signals; unset = visual evidence never fires |
| `ACTIVE_MARKETS` | Comma-separated markets for location hashtag generation (e.g. `nyc,brooklyn,centralpark`) |
| `DAILY_BUDGET_CAP` | Max provider results per UTC day (default 500) — the spend guardrail |
| `DASHBOARD_ORIGIN` | CORS origin for the dashboard |
| `STATIC_DIR` | Serve a built dashboard from the Go server (set in the Docker image) |

### Deploy

**Current production: TeamC AWS** — EC2 `i-0a7a301139acb9e87` (t3.micro,
us-east-1), app + Postgres 17 + Caddy in Docker with `--restart=always`.
Public: **`https://54.196.158.87.sslip.io`** (Let's Encrypt TLS, auto-renewed;
HTTP redirects to HTTPS; port 8080 is not public).

```bash
make deploy   # rsync source → rebuild on the instance → restart container
```

Secrets live in `~/neptune-radar.env` on the instance (chmod 600, gitignored).
SSH: `ssh -i ~/.ssh/neptune-radar.pem ec2-user@54.196.158.87`.

**Fly.io alternative:** `fly.toml` is ready — `fly apps create`, attach
Postgres, `fly secrets set ...`, `fly deploy` once billing exists on the
account.

The Dockerfile builds one service: the Go server serves the built dashboard and
runs the API + watch loop. Health check: `GET /api/health`.

## What's in here

- **Signal vocabulary**: [`backend/internal/signals/`](backend/internal/signals)
- **Ingestion** (Apify client, provider→event mapping, the watch loop worker):
  [`backend/internal/ingest/`](backend/internal/ingest),
  [`backend/internal/pipeline/watchtower/event.go`](backend/internal/pipeline/watchtower/event.go)
- **Pipeline** (Identity Resolver → Relationship Analyst → Relevance Scorer →
  Policy Guard → Conversation Agent → Workflow Operator → Verifier):
  [`backend/internal/pipeline/`](backend/internal/pipeline)
- **Postgres store + migrations**:
  [`backend/internal/store/`](backend/internal/store)
- **Vision classifier**: [`backend/internal/llm/vision.go`](backend/internal/llm/vision.go)
- **Ops dashboard** (live signal feed, approval queue, couple graph, sources,
  audit trail): [`frontend/`](frontend)
- **Tests**: pure unit tests (signal vocabulary, scorer math, policy tiers,
  no-LLM-import invariants), inline end-to-end pipeline tests, and ingest
  mapper tests — all against a real Postgres (`TEST_DATABASE_URL`).

See [`ARCHITECTURE.md`](ARCHITECTURE.md) for the pipeline diagram and design
decisions, and [`PRODUCTION_GAPS.md`](PRODUCTION_GAPS.md) for the honest list
of what still isn't covered (consent capture, data retention, RBAC, and the
compliance posture of the data provider).
