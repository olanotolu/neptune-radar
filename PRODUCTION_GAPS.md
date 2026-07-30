# What's required before this is truly production

This build is live software: real Postgres, real provider ingestion, auth on the
API, versioned migrations. The following remain honestly open — ordered roughly
by "would stop a lawyer" to "would stop an SRE."

## Legal & data protection (the top of the list)

- **The data provider violates no law by existing, but Meta's ToS prohibit
  scraping, and Meta litigates.** Apify actors are a third-party scraping
  layer. Legal needs to sign off on the provider's compliance posture (and
  indemnify/business-associate terms), and the system must be designed to swap
  providers or go vendor-opt-in (official Meta API) without pipeline changes —
  the `ingest` package is the only provider-coupled code, keep it that way.
- **Monitoring couples' relationship status is sensitive personal data.** GDPR/
  CCPA posture is not optional: documented lawful basis, a data-retention and
  deletion policy for `social_observations`/`evidence`/`audit_events`
  (especially anything at `unconfirmed_inference` scope about people who never
  interacted with Neptune), and a real DSAR delete path.
- **Consent capture at the source.** `ConsentPolicy` rows are written by hand/
  import right now. Production needs the auditable flow: what was the person
  told, when, and can they see/revoke it themselves.

## Identity & model quality

- **A documented appeals/override path for identity resolution.** Reciprocal
  tag/follow + co-tag is deliberately conservative, but it will be wrong
  sometimes, and a human needs a way to mark a couple/hypothesis as mistaken
  that the scorer respects permanently.
- **Vision classifier accuracy is unmeasured.** The Baseten multimodal model's
  precision on ring/proposal detection needs a labeled sample before its +10
  means anything. Until then it's an uncalibrated evidence source.
- **Prompt injection resistance.** Captions/bios are untrusted content flowing
  into `internal/llm` prompts. The policy layer structurally can't trust the
  model past a confidence number, which limits blast radius, but the prompts
  have no injection hardening.
- **Model calibration sampling.** Nothing audits whether the analyst's
  confidence numbers stay calibrated over time.

## Security

- **The admin token is a shared secret, not identity.** Real RBAC: who is "the
  concierge," per-user attribution on approve/ignore (today `decided_by` is a
  hardcoded string), and audit of *dashboard* actions, not just pipeline ones.
- **`visibility_scope` must gate reads, not just label rows.** `attorney_only`
  data is currently queryable by anyone with the admin token.
- **Audit log integrity** — `audit_events` should be append-only at the
  database permission level, not just by convention.
- **Secrets management** — fly secrets is fine at this size; rotation policy
  is not written down anywhere.

## Reliability & scale

- **Provider failure modes are logged, not alerted.** Consecutive-fetch-failure
  alerting, dead-lettering for unmappable items, and a provider-outage runbook.
- **Apify actor schema drift.** The mapper is defensive and unit-tested, but an
  actor upgrade that changes shapes silently degrades ingestion to zero-signal;
  add a daily canary assertion (expected fields present in N% of items).
- **One worker, no lease.** Running two replicas double-spends the provider
  budget and double-processes events (idempotency saves correctness, not
  money). Add a leader lock before scaling beyond one machine.
- **Rate limiting on the API** — the bearer token is the only throttle.

## Product surface

- **No frontend tests** — the approval flow and sources management deserve
  component tests and one Playwright smoke test.
- **No accessibility audit** of the dashboard.
- **CRM import path** — persons/consent currently arrive via direct DB writes.
  The real intake sync (guide downloads, signups) needs a documented importer
  or webhook.

## Ohio source registry (government/church/social map)

- **Only Franklin County has a government connector.** The other 87 Ohio
  counties honestly show "not yet configured" — extending coverage means
  identifying each county's real probate-court (or equivalent) marriage-record
  endpoint one by one; there is no generic pattern across Ohio's 88 counties'
  websites to automate this.
- **Government/church connectors check reachability only, not records.**
  `internal/connectors.HTTPHealthConnector` confirms the endpoint is up and
  hashes the page for drift detection — it does not search, extract, or store
  any actual marriage-license or bulletin record yet. That's a materially
  bigger, separately-scoped build (real parsing, real record-vs-no-record
  logic, the `LicenseApplied`/`LicenseIssued`/`MarriageRecorded` distinction).
- **No parish has a verified bulletin-archive URL yet.** The 9 real Columbus
  parishes are registered from the diocese's own directory, but no individual
  parish website was verified, so `BulletinDiscoveryConnector` has never been
  run against them — the map honestly shows "no bulletin archive discovered
  yet" for all nine. Running discovery (and, later, real PDF banns extraction)
  is the next real step, not a placeholder to quietly fill with guessed URLs.
- **Instagram connector health depends on `APIFY_TOKEN`.** Without it, all 13
  Columbus vendor connectors correctly read "degraded" (a real check attempt
  failed for a real, stated reason) rather than a fabricated "healthy" — set
  the token to get real per-vendor health from the same Apify pipeline that
  already powers ingestion.
- **Diocese/city coordinates are hand-verified public facts, not geocoded.**
  Fine at one city; will need a real geocoding step before this covers more
  than a handful of Ohio cities.
- **No automatic "review task" on connector degradation yet.** The Operations-
  agent responsibility described in the design doc (flag a source when it goes
  offline, a bulletin archive stops updating, etc.) isn't wired to
  `recommended_actions` — today a human has to notice a red status in the UI.
