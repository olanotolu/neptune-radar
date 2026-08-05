import { useMemo } from "react";
import { useCoupleDossier, useCoupleJourney } from "../api/hooks";
import type { JourneyTimelineEvent } from "../api/types";
import { EmptyState } from "../components/EmptyState";
import { LoadingState } from "../components/LoadingState";

function pct(n: number): string {
  return `${Math.round((n || 0) * 100)}%`;
}

// Status tone per event — green for positive milestones, amber for pending /
// in-progress, red for failures. Anything unrecognized is neutral.
function eventTone(ev: JourneyTimelineEvent): "pos" | "pending" | "neg" | "neutral" {
  const t = ev.event_type;
  // Failures / suppressions.
  if (
    t === "closed_lost" ||
    t === "marked_mistaken" ||
    t === "suppressed" ||
    t === "automation_paused"
  ) {
    return "neg";
  }
  // Positive, completed milestones.
  if (
    t === "signal_detected" ||
    t === "evidence_added" ||
    t === "kit_built" ||
    t === "address_found" ||
    t === "postcard_mailed" ||
    t === "follow_up_sent" ||
    t === "closed_won" ||
    t === "consult_booked" ||
    t === "chat_started" ||
    t === "handoff_clicked" ||
    t === "handoff_issued" ||
    t === "address_verified" ||
    t === "automation_resumed"
  ) {
    return "pos";
  }
  // Pending / in-progress detective work.
  if (t.includes("detective") || t.includes("prep") || t === "journey_stage_set") {
    return "pending";
  }
  return "neutral";
}

// Human label for the event_type chip.
function eventLabel(t: string): string {
  const map: Record<string, string> = {
    signal_detected: "Signal",
    evidence_added: "Evidence",
    kit_built: "Kit built",
    address_found: "Address",
    postcard_mailed: "Mailed",
    follow_up_sent: "Follow-up",
    chat_started: "Chat",
    consult_booked: "Booked",
    closed_won: "Won",
    closed_lost: "Lost",
    handoff_clicked: "Clicked",
    handoff_issued: "Handoff",
    automation_paused: "Paused",
    automation_resumed: "Resumed",
    marked_mistaken: "Mistaken",
    suppressed: "Suppressed",
    journey_stage_set: "Stage set",
  };
  return map[t] ?? t.replace(/_/g, " ");
}

function fmtTime(ts: string): string {
  const d = new Date(ts);
  if (Number.isNaN(d.getTime())) return ts;
  // Geist Mono, compact: YYYY-MM-DD HH:mm
  const pad = (n: number) => String(n).padStart(2, "0");
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`;
}

function SummaryStat({ label, value }: { label: string; value: string }) {
  return (
    <div className="journey__stat">
      <span className="journey__stat-label">{label}</span>
      <strong className="journey__stat-value">{value}</strong>
    </div>
  );
}

export function JourneyView({ coupleId }: { coupleId?: string }) {
  const { data: dossier } = useCoupleDossier(coupleId);
  const { data: events, error, isLoading } = useCoupleJourney(coupleId);

  const sorted = useMemo(
    () => [...(events || [])].sort((a, b) => a.timestamp.localeCompare(b.timestamp)),
    [events],
  );

  if (!coupleId) {
    return (
      <EmptyState
        variant="empty"
        title="No couple selected"
        message="Open a couple's dossier and choose View Journey to see their timeline."
      />
    );
  }
  if (isLoading) return <LoadingState variant="spinner" message="Loading journey…" />;
  if (error) {
    return <EmptyState variant="warning" title="Journey unavailable" message={(error as Error).message} />;
  }

  const nameA = dossier?.person_a_name || dossier?.handle_a || "A";
  const nameB = dossier?.person_b_name || dossier?.handle_b || "B";
  const evidenceCount = dossier?.evidence?.length ?? 0;
  const signalCount = sorted.filter((e) => e.event_type === "evidence_added").length;
  const confidence = dossier?.hypothesis_score ?? dossier?.engagement_score ?? 0;
  const kitStatus = dossier?.latest_kit_status || "none";

  return (
    <div className="journey">
      <header className="journey__head">
        <div className="journey__identity">
          <h1 className="journey__title">
            {nameA} <span className="journey__amp">&</span> {nameB}
          </h1>
          {dossier?.handle_a && dossier?.handle_b && (
            <p className="journey__handles">
              @{dossier.handle_a} · @{dossier.handle_b}
              {(dossier.city || dossier.region) && (
                <span>
                  {" · "}
                  {dossier.city}
                  {dossier.region ? `, ${dossier.region}` : ""}
                </span>
              )}
            </p>
          )}
        </div>
      </header>

      <div className="journey__summary">
        <SummaryStat label="Signals" value={String(signalCount)} />
        <SummaryStat label="Evidence" value={String(evidenceCount)} />
        <SummaryStat label="Confidence" value={pct(confidence)} />
        <SummaryStat label="Kit" value={kitStatus.replace(/_/g, " ")} />
        <SummaryStat label="Stage" value={(dossier?.journey_stage || "detected").replace(/_/g, " ")} />
      </div>

      {sorted.length === 0 ? (
        <EmptyState
          variant="empty"
          title="No journey events yet"
          message="This couple has no recorded milestones. Signals will appear here as they're detected."
        />
      ) : (
        <ol className="journey__timeline" aria-label="Couple journey timeline">
          {sorted.map((ev, i) => {
            const tone = eventTone(ev);
            return (
              <li key={`${ev.timestamp}-${i}`} className={`journey__node journey__node--${tone}`}>
                <span className="journey__dot" aria-hidden />
                <div className="journey__card">
                  <div className="journey__card-head">
                    <span className="journey__chip">{eventLabel(ev.event_type)}</span>
                    <time className="journey__time">{fmtTime(ev.timestamp)}</time>
                  </div>
                  <p className="journey__desc">{ev.description}</p>
                  <div className="journey__meta">
                    {ev.confidence !== undefined && ev.confidence > 0 && (
                      <span className="journey__conf">{pct(ev.confidence)} confidence</span>
                    )}
                    {ev.source && <span className="journey__src">{ev.source}</span>}
                  </div>
                </div>
              </li>
            );
          })}
        </ol>
      )}
    </div>
  );
}
