import { useState } from "react";
import { AuditTable } from "../components/AuditTable";
import {
  useAudit,
  useAutopsies,
  useFunnelEvents,
  useFunnelStats,
  useGenerateAutopsy,
} from "../api/hooks";
import type { AutopsyReport } from "../api/types";
import { useToast } from "../components/Toast";
import { EmptyState } from "../components/EmptyState";
import { LoadingState } from "../components/LoadingState";

function pct(n: number): string {
  return `${Math.round((n || 0) * 100)}%`;
}

function FunnelPanel() {
  const { data: stats, isLoading } = useFunnelStats();
  const { data: events } = useFunnelEvents();

  return (
    <section className="trust-panel">
      <header className="trust-panel__header">
        <div>
          <h3 className="trust-panel__title">Closed-loop funnel</h3>
          <p className="trust-panel__sub">
            Meet Neptune product events → journey advances. Wire{" "}
            <code>POST /api/webhooks/neptune</code> with{" "}
            <code>NEPTUNE_WEBHOOK_SECRET</code>.
          </p>
        </div>
      </header>

      {isLoading && <LoadingState variant="dots" message="Loading funnel…" />}
      {stats && (
        <div className="funnel-kpis">
          <div className="funnel-kpi">
            <strong>{stats.handoffs_issued}</strong>
            <span>Handoffs issued</span>
          </div>
          <div className="funnel-kpi">
            <strong>{stats.chat_started_7d}</strong>
            <span>Chats (7d)</span>
          </div>
          <div className="funnel-kpi">
            <strong>{stats.consult_booked_7d}</strong>
            <span>Booked (7d)</span>
          </div>
          <div className="funnel-kpi">
            <strong>{stats.closed_won_7d}</strong>
            <span>Closed won (7d)</span>
          </div>
          <div className="funnel-kpi">
            <strong>{pct(stats.chat_rate)}</strong>
            <span>Chat rate</span>
          </div>
          <div className="funnel-kpi">
            <strong>{pct(stats.book_rate)}</strong>
            <span>Book rate</span>
          </div>
        </div>
      )}

      <div className="funnel-events">
        <h4>Recent funnel events</h4>
        {(events ?? []).length === 0 ? (
          <EmptyState
            variant="empty"
            title="No product events yet"
            message="When a couple starts chat or books a consult, post to the webhook with handoff_code or utm_content (couple id)."
          />
        ) : (
          <table className="dossier-ledger">
            <thead>
              <tr>
                <th>When</th>
                <th>Event</th>
                <th>Couple</th>
                <th>Matched</th>
                <th>Journey</th>
              </tr>
            </thead>
            <tbody>
              {(events ?? []).slice(0, 20).map((e) => (
                <tr key={e.id}>
                  <td>{e.occurred_at?.slice(0, 16)?.replace("T", " ")}</td>
                  <td>
                    <code>{e.event_type}</code>
                  </td>
                  <td>
                    {e.couple_id ? (
                      <a href={`#/dossier/${encodeURIComponent(e.couple_id)}`}>{e.couple_id.slice(0, 14)}…</a>
                    ) : (
                      <em>unresolved</em>
                    )}
                  </td>
                  <td>{e.matched_by}</td>
                  <td>
                    {e.journey_stage_before || "—"} → {e.journey_stage_after || "—"}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      <details className="webhook-docs">
        <summary>Webhook contract</summary>
        <pre className="webhook-docs__pre">{`POST /api/webhooks/neptune
Authorization: Bearer $NEPTUNE_WEBHOOK_SECRET
# or: X-Neptune-Webhook-Secret: $NEPTUNE_WEBHOOK_SECRET

{
  "event": "chat_started",
  "handoff_code": "abc123",
  "utm_content": "couple_…",
  "external_id": "product-event-uuid",
  "occurred_at": "2026-07-31T15:00:00Z",
  "metadata": { "session_id": "…" }
}

Events: chat_started | consult_booked | closed_won | closed_lost | handoff_clicked
Journey: invited → in_chat → booked → closed_*`}</pre>
      </details>
    </section>
  );
}

function AutopsyPanel() {
  const { data: reports, isLoading } = useAutopsies();
  const generate = useGenerateAutopsy();
  const toast = useToast();
  const [selected, setSelected] = useState<AutopsyReport | null>(null);

  const latest = selected ?? reports?.[0] ?? null;

  return (
    <section className="trust-panel">
      <header className="trust-panel__header">
        <div>
          <h3 className="trust-panel__title">False-positive autopsy</h3>
          <p className="trust-panel__sub">
            Weekly trust report for legal + ops: suppressions, ignores, reject rate, lessons per case.
          </p>
        </div>
        <button
          type="button"
          className="btn btn--primary"
          disabled={generate.isPending}
          onClick={() =>
            generate.mutate(
              { days: 7 },
              {
                onSuccess: (r) => {
                  toast.push("Autopsy generated", "ok");
                  setSelected(r);
                },
                onError: (e) => toast.push((e as Error).message, "err"),
              },
            )
          }
        >
          {generate.isPending ? "Generating…" : "Run 7-day autopsy"}
        </button>
      </header>

      {isLoading && <LoadingState variant="dots" message="Loading reports…" />}

      {!isLoading && (reports ?? []).length === 0 && !latest && (
        <EmptyState
          variant="empty"
          title="No autopsy reports yet"
          message="Run a 7-day autopsy to generate the weekly trust report for legal and ops."
        />
      )}

      {(reports ?? []).length > 0 && (
        <div className="autopsy-list">
          {(reports ?? []).map((r) => (
            <button
              key={r.id}
              type="button"
              className={`autopsy-list__item ${latest?.id === r.id ? "autopsy-list__item--active" : ""}`}
              onClick={() => setSelected(r)}
            >
              <strong>{r.period_start.slice(0, 10)} → {r.period_end.slice(0, 10)}</strong>
              <span>
                reject {pct(r.summary.human_reject_rate)} · {r.summary.suppressed_couples} suppressed ·{" "}
                {r.summary.ignored_actions} ignored
              </span>
            </button>
          ))}
        </div>
      )}

      {latest && (
        <div className="autopsy-detail">
          <div className="funnel-kpis">
            <div className="funnel-kpi">
              <strong>{pct(latest.summary.human_reject_rate)}</strong>
              <span>Human reject rate</span>
            </div>
            <div className="funnel-kpi">
              <strong>{latest.summary.suppressed_couples}</strong>
              <span>Suppressed</span>
            </div>
            <div className="funnel-kpi">
              <strong>{latest.summary.ignored_actions}</strong>
              <span>Ignored</span>
            </div>
            <div className="funnel-kpi">
              <strong>{latest.summary.approved_actions}</strong>
              <span>Approved</span>
            </div>
            <div className="funnel-kpi">
              <strong>{latest.summary.rejected_hypotheses}</strong>
              <span>Rejected hyps</span>
            </div>
          </div>

          {latest.summary.top_failure_modes?.length > 0 && (
            <div className="autopsy-modes">
              <h4>Top failure modes</h4>
              <ul>
                {latest.summary.top_failure_modes.map((m) => (
                  <li key={m}>{m}</li>
                ))}
              </ul>
            </div>
          )}

          {latest.summary.notes?.length > 0 && (
            <ul className="autopsy-notes">
              {latest.summary.notes.map((n) => (
                <li key={n}>{n}</li>
              ))}
            </ul>
          )}

          <h4>Cases ({latest.cases?.length ?? 0})</h4>
          {(latest.cases ?? []).length === 0 ? (
            <p className="work-drawer__muted">No suppressions/ignores in this window.</p>
          ) : (
            <div className="autopsy-cases">
              {latest.cases.slice(0, 40).map((c, i) => (
                <article key={`${c.couple_id}-${c.kind}-${i}`} className={`autopsy-case autopsy-case--${c.kind}`}>
                  <header>
                    <span className="autopsy-case__kind">{c.kind}</span>
                    <code>{c.reason}</code>
                    {c.score != null && c.score > 0 && <span className="autopsy-case__score">{pct(c.score)}</span>}
                  </header>
                  <p className="autopsy-case__meta">
                    {c.handles?.length ? c.handles.map((h) => `@${h}`).join(" · ") : "—"}
                    {c.couple_id && (
                      <>
                        {" · "}
                        <a href={`#/dossier/${encodeURIComponent(c.couple_id)}`}>dossier</a>
                      </>
                    )}
                    {c.occurred_at && <> · {c.occurred_at.slice(0, 10)}</>}
                  </p>
                  <p className="autopsy-case__lesson">{c.lesson}</p>
                </article>
              ))}
            </div>
          )}
        </div>
      )}

      {!latest && !isLoading && (
        <p className="work-drawer__muted">No autopsy reports yet — run a 7-day autopsy to seed the trust log.</p>
      )}
    </section>
  );
}

export function AuditTrailView() {
  const [monitorFilter, setMonitorFilter] = useState("");
  const { data: events, error, isLoading } = useAudit(monitorFilter || undefined);

  return (
    <div className="view view--system">
      <h2 className="view__title">Audit trail</h2>
      <p className="view__subtitle">
        Every decision, every action, every outcome. Trust is built on transparency.
      </p>

      <FunnelPanel />
      <AutopsyPanel />

      <section className="trust-panel">
        <header className="trust-panel__header">
          <div>
            <h3 className="trust-panel__title">Pipeline audit trail</h3>
            <p className="trust-panel__sub">
              Every stage logs its decision — including “do nothing”. Observe → normalize → resolve → score → policy → recommend.
            </p>
          </div>
        </header>
        <div className="feed-toolbar">
          <input
            className="feed-filter"
            placeholder="Filter by monitor (e.g. vendor:weddingsbynoor)…"
            value={monitorFilter}
            onChange={(e) => setMonitorFilter(e.target.value)}
          />
        </div>
        {error ? (
          <EmptyState variant="warning" title="Audit trail unavailable" message={(error as Error).message} />
        ) : isLoading ? (
          <LoadingState variant="skeleton" message="Loading audit trail…" />
        ) : (
          <AuditTable events={events ?? []} />
        )}
      </section>
    </div>
  );
}
