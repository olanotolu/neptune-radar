.PHONY: backend frontend test test-db test-db-down build deploy

BACKEND_ENV := $(shell test -f backend/.env && echo backend/.env)
LOAD_ENV = set -a && [ -f $(BACKEND_ENV) ] && . $(BACKEND_ENV); true

# Local dev: requires DATABASE_URL (see test-db for a throwaway local one),
# NEPTUNE_ADMIN_TOKEN, and optionally APIFY_TOKEN / BASETEN_API_KEY.
backend:
	$(LOAD_ENV); cd backend && go run ./cmd/server

frontend:
	cd frontend && npm install && npm run dev

# Throwaway Postgres for local dev/tests in Docker.
test-db:
	docker rm -f neptune-pg-test >/dev/null 2>&1 || true
	docker run -d --name neptune-pg-test -e POSTGRES_PASSWORD=test -e POSTGRES_DB=neptune_test -p 55432:5432 postgres:17-alpine
	@sleep 3
	@echo 'export DATABASE_URL=postgres://postgres:test@localhost:55432/neptune_test?sslmode=disable'
	@echo 'export TEST_DATABASE_URL=$$DATABASE_URL'

test-db-down:
	docker rm -f neptune-pg-test

test:
	@# The store/pipeline tests skip silently without a real Postgres —
	@# a green run with zero DB tests is a lie. Fail loudly instead.
	@$(LOAD_ENV); [ -n "$$TEST_DATABASE_URL" ] || { \
		echo "TEST_DATABASE_URL not set — run 'make test-db' first and export the printed URL."; \
		echo "(Refusing to run a suite that would skip every database test.)"; exit 1; }
	$(LOAD_ENV); cd backend && go test ./...
	cd frontend && npm run lint && npx tsc --noEmit -p tsconfig.app.json

build:
	docker build -t neptune-radar .

# Deploys to the TeamC AWS EC2 instance (i-0a7a301139acb9e87). The box is a
# 1 GB t3.micro: building the Go/npm stages in-image there wedges it (OOM),
# so artifacts are built LOCALLY (frontend bundle + linux/amd64 Go binaries)
# and the remote docker build is a trivial COPY from Dockerfile.deploy.
# Secrets live in ~/neptune-radar.env on the instance — edit there, never here.
EC2_HOST = ec2-user@54.196.158.87
EC2_KEY  = ~/.ssh/neptune-radar.pem

deploy:
	cd frontend && npm ci && npm run build
	mkdir -p build
	cd backend && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ../build/server ./cmd/server
	cd backend && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ../build/bootstrap-ohio ./cmd/bootstrap-ohio
	rsync -az --delete -e "ssh -i $(EC2_KEY)" --exclude node_modules --exclude .git --exclude ".env" ./ $(EC2_HOST):neptune-radar/
	ssh -i $(EC2_KEY) $(EC2_HOST) 'cd neptune-radar && docker build -q -f Dockerfile.deploy -t neptune-radar . && \
		docker rm -f neptune-radar >/dev/null 2>&1 || true; \
		docker run -d --name neptune-radar --network neptune-net --restart=always \
		  --log-opt max-size=10m --log-opt max-file=3 \
		  -p 127.0.0.1:8080:8080 --env-file ~/neptune-radar.env neptune-radar >/dev/null && \
		sleep 2 && curl -sf --retry 10 --retry-all-errors --retry-delay 1 http://localhost:8080/api/health && \
		docker image prune -f >/dev/null'
	@# Hard gate: the bundle must never contain the dev API URL (incident 2026-07-29).
	@! ssh -i $(EC2_KEY) $(EC2_HOST) 'docker run --rm neptune-radar sh -c "grep -rq localhost:8080 /app/public"' && echo "bundle clean"
