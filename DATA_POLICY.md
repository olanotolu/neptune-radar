# Data policy

This applies to the source registry (government/church/social connectors
behind the Ohio coverage map) and, more broadly, to how Neptune Radar
monitors and stores signals.

## Public sources only

Every connector in `internal/connectors` targets a public source: a
government office's public record-search page, a church's public website, or
a business's public social account. Nothing here accesses a private account,
bypasses authentication, or scrapes a login-gated page.

## What we don't do

- No access to private Instagram accounts.
- No bypassing authentication or CAPTCHAs.
- No stealing or reusing sessions.
- No facial recognition.
- No inferring religious identity or belief from parish/diocese data — a
  parish's bulletin publishing a bann is evidence a wedding was publicly
  announced, nothing more.
- No classifying people by protected characteristics.
- No automatic outreach to any person or business.
- No treating public social content as ground truth, or a marriage-license
  application/issuance as proof a wedding happened — see the
  `identified → configured → active → error` connector-status model and the
  event-type distinctions in `PRODUCTION_GAPS.md`.

## Honesty about status

A connector's `status` in the database is never set except as the direct
result of a real check recorded in `connector_runs`
(`internal/store/map.go`'s `RecordConnectorRun`). Concretely:

- `setup` — registered, no successful check has run yet.
- `healthy` — the most recent real check succeeded.
- `degraded` — reachable but recent checks are failing, or checks are stale.
- `offline` — three consecutive real check failures.

The UI never shows "Active"/"Healthy" without a real check behind it. Where
information is missing, the UI says so plainly: "Not configured," "No
successful check yet," "No bulletin archive discovered yet."

## Fixture data stays in tests

Go's `_test.go` files are excluded from any built binary by the toolchain
itself — fixture data used in pipeline tests (`internal/pipeline/*_test.go`)
can never be served by the live API. The source registry additionally has a
`data_mode` column (`live` / `verified_import` / `manual` / `fixture`) for
defense in depth going forward; production reads exclude `fixture` rows
unless the server is explicitly started with fixtures enabled for local
development.

## Provenance

Every `source_organizations` row records a `provenance` — where the fact
that this real institution/business exists was confirmed (its own official
website, an official government or diocese directory, or manual curation
against a cited source). `cmd/bootstrap-ohio/ohio_data.go` documents the
exact source and verification date for every fact it seeds.

## Retention and deletion

Not yet built for the source registry specifically — see the "Legal & data
protection" section of `PRODUCTION_GAPS.md` for the broader retention/
deletion posture this needs to inherit before handling real couples' data
end-to-end.
