# Neptune Radar — single-service image: built dashboard + Go API + watch loop.
FROM node:22-alpine AS frontend
WORKDIR /build/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
# The dashboard calls the same origin it was served from in production.
RUN npm run build

FROM golang:1.26-alpine AS backend
WORKDIR /build/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN CGO_ENABLED=0 go build -o /out/server ./cmd/server
RUN CGO_ENABLED=0 go build -o /out/seed-geography ./cmd/seed-geography
RUN CGO_ENABLED=0 go build -o /out/bootstrap-state ./cmd/bootstrap-state

FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=backend /out/server /app/server
COPY --from=backend /out/seed-geography /app/seed-geography
COPY --from=backend /out/bootstrap-state /app/bootstrap-state
COPY --from=frontend /build/frontend/dist /app/public
ENV ADDR=":8080" \
    STATIC_DIR="/app/public" \
    DASHBOARD_ORIGIN="https://neptune-radar.fly.dev"
EXPOSE 8080
CMD ["/app/server"]
