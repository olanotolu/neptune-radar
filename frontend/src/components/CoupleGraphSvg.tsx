import type { CoupleGraph, GraphEdge } from "../api/types";

const WIDTH = 800;
const HEIGHT = 400;

const EDGE_COLORS: Record<string, string> = {
  tagged: "var(--cove-deep)",
  commented: "var(--mesa-deep)",
  mentioned: "#7c5cff",
  liked: "#e1306c",
  co_posted: "var(--green)",
};

function edgeColor(kind: string, active: boolean): string {
  return active ? EDGE_COLORS[kind] ?? "var(--ink-soft)" : "var(--ink-faint)";
}

export function CoupleGraphSvg({ graph, confidence }: { graph: CoupleGraph; confidence?: number }) {
  const persons = graph.nodes.filter((n) => n.type === "person");
  const accounts = graph.nodes.filter((n) => n.type === "account");

  if (persons.length < 2 || accounts.length < 2) {
    return <EmptyGraph />;
  }

  const [personA, personB] = persons;
  const accountFor = (personId: string) =>
    accounts.find((a) =>
      graph.edges.some((e) => e.kind === "owns_account" && e.from === personId && e.to === a.id),
    );
  const accountA = accountFor(personA.id) ?? accounts[0];
  const accountB = accountFor(personB.id) ?? accounts[1];

  const aX = 180;
  const bX = 620;
  const personY = 120;
  const accountY = 290;

  const crossEdges = graph.edges.filter(
    (e) =>
      e.kind !== "owns_account" &&
      ((e.from === accountA.id && e.to === accountB.id) ||
        (e.from === accountB.id && e.to === accountA.id)),
  );

  const conf = confidence ?? 0;
  const confPct = Math.round(Math.max(0, Math.min(1, conf)) * 100);
  const ringR = 54;
  const ringC = 2 * Math.PI * ringR;

  return (
    <svg
      viewBox={`0 0 ${WIDTH} ${HEIGHT}`}
      className="couple-graph-svg"
      role="img"
      aria-label={`Relationship graph for ${personA.label} and ${personB.label}`}
    >
      <defs>
        <radialGradient id="bgGlow" cx="50%" cy="42%" r="65%">
          <stop offset="0%" stopColor="color-mix(in srgb, var(--cove) 60%, var(--surface))" />
          <stop offset="100%" stopColor="var(--surface)" />
        </radialGradient>
        <pattern id="dotGrid" x="0" y="0" width="22" height="22" patternUnits="userSpaceOnUse">
          <circle cx="2" cy="2" r="1.1" fill="var(--ink-faint)" opacity="0.35" />
        </pattern>
        <linearGradient id="coveGrad" x1="0%" y1="0%" x2="100%" y2="100%">
          <stop offset="0%" stopColor="var(--cove-deep)" />
          <stop offset="100%" stopColor="var(--cove-ink)" />
        </linearGradient>
        <linearGradient id="mesaGrad" x1="0%" y1="0%" x2="100%" y2="100%">
          <stop offset="0%" stopColor="var(--mesa)" />
          <stop offset="100%" stopColor="var(--mesa-deep)" />
        </linearGradient>
        <linearGradient id="igGrad" x1="0%" y1="0%" x2="100%" y2="100%">
          <stop offset="0%" stopColor="#833ab4" />
          <stop offset="50%" stopColor="#e1306c" />
          <stop offset="100%" stopColor="#f77737" />
        </linearGradient>
        <linearGradient id="ownsA" x1="0%" y1="0%" x2="0%" y2="100%">
          <stop offset="0%" stopColor="var(--cove-deep)" />
          <stop offset="100%" stopColor="#e1306c" />
        </linearGradient>
        <linearGradient id="ownsB" x1="0%" y1="0%" x2="0%" y2="100%">
          <stop offset="0%" stopColor="var(--mesa-deep)" />
          <stop offset="100%" stopColor="#f77737" />
        </linearGradient>
        <filter id="nodeShadow" x="-40%" y="-40%" width="180%" height="180%">
          <feDropShadow dx="0" dy="3" stdDeviation="5" floodColor="var(--ink)" floodOpacity="0.16" />
        </filter>
        <filter id="softShadow" x="-40%" y="-40%" width="180%" height="180%">
          <feDropShadow dx="0" dy="2" stdDeviation="3" floodColor="var(--ink)" floodOpacity="0.12" />
        </filter>
      </defs>

      {/* background */}
      <rect x="0" y="0" width={WIDTH} height={HEIGHT} fill="url(#bgGlow)" />
      <rect x="0" y="0" width={WIDTH} height={HEIGHT} fill="url(#dotGrid)" />

      {/* owns_account curved edges */}
      <path
        d={`M${aX} ${personY + 36} C ${aX} ${personY + 150}, ${aX} ${accountY - 90}, ${aX} ${accountY - 20}`}
        className="graph-edge-owns"
        stroke="url(#ownsA)"
        fill="none"
      />
      <path
        d={`M${bX} ${personY + 36} C ${bX} ${personY + 150}, ${bX} ${accountY - 90}, ${bX} ${accountY - 20}`}
        className="graph-edge-owns"
        stroke="url(#ownsB)"
        fill="none"
      />

      {/* cross edges as curved arcs */}
      {crossEdges.map((e, i) => {
        const fromLeft = e.from === accountA.id;
        const x1 = fromLeft ? aX + 60 : bX - 60;
        const x2 = fromLeft ? bX - 60 : aX + 60;
        const sweep = fromLeft ? 1 : 0;
        const lift = 60 + i * 26;
        const mx = (x1 + x2) / 2;
        const my = accountY + 28 + lift;
        return (
          <g key={`${e.kind}-${e.from}-${e.to}-${i}`} className="graph-cross-group">
            <path
              d={`M${x1} ${accountY} Q ${mx} ${accountY + lift + 40}, ${x2} ${accountY}`}
              className={`graph-cross-edge ${e.active ? "is-active" : "is-inactive"}`}
              stroke={edgeColor(e.kind, e.active)}
              fill="none"
              style={{ animationDelay: `${i * 0.4}s` }}
            />
            {/* label pill at midpoint */}
            <g transform={`translate(${mx} ${my})`}>
              <rect
                x={-(labelWidth(e.kind) / 2)}
                y={-11}
                width={labelWidth(e.kind)}
                height={22}
                rx={11}
                className="graph-edge-pill"
                filter="url(#softShadow)"
              />
              <text textAnchor="middle" y={4} className="graph-edge-pill-text">
                {e.kind}
              </text>
            </g>
          </g>
        );
      })}

      {/* confidence ring (center) */}
      <g transform={`translate(${WIDTH / 2} ${personY + 70})`} className="graph-conf-ring">
        <circle r={ringR} className="graph-conf-track" />
        <circle
          r={ringR}
          className="graph-conf-fill"
          strokeDasharray={ringC}
          strokeDashoffset={ringC * (1 - conf / 100)}
        />
        <text textAnchor="middle" y={-2} className="graph-conf-pct">
          {confPct}%
        </text>
        <text textAnchor="middle" y={16} className="graph-conf-cap">
          confidence
        </text>
      </g>

      {/* person A */}
      <PersonNode x={aX} y={personY} label={personA.label} gradId="coveGrad" />
      {/* person B */}
      <PersonNode x={bX} y={personY} label={personB.label} gradId="mesaGrad" />

      {/* account A */}
      <AccountNode x={aX} y={accountY} label={accountA.label} />
      {/* account B */}
      <AccountNode x={bX} y={accountY} label={accountB.label} />

      {/* signal timeline */}
      <SignalTimeline edges={crossEdges} />
    </svg>
  );
}

function PersonNode({ x, y, label, gradId }: { x: number; y: number; label: string; gradId: string }) {
  return (
    <g className="graph-person" transform={`translate(${x} ${y})`}>
      <circle r={36} className="graph-person-ring" />
      <circle r={36} fill={`url(#${gradId})`} filter="url(#nodeShadow)" className="graph-person-core" />
      <text textAnchor="middle" y={6} className="graph-person-initials">
        {initials(label)}
      </text>
      <text textAnchor="middle" y={58} className="graph-person-name">
        {label}
      </text>
    </g>
  );
}

function AccountNode({ x, y, label }: { x: number; y: number; label: string }) {
  const w = 120;
  const h = 40;
  return (
    <g className="graph-account" transform={`translate(${x} ${y})`}>
      <rect
        x={-w / 2}
        y={-h / 2}
        width={w}
        height={h}
        rx={12}
        fill="url(#igGrad)"
        filter="url(#nodeShadow)"
        className="graph-account-rect"
      />
      {/* simple camera glyph */}
      <g transform="translate(-38 0)" className="graph-account-icon">
        <rect x={-9} y={-7} width={18} height={14} rx={3} fill="#fff" opacity="0.95" />
        <circle cx={0} cy={0} r={4} fill="#e1306c" />
        <rect x={-4} y={-9} width={8} height={3} rx={1} fill="#fff" opacity="0.95" />
      </g>
      <text x={10} y={5} textAnchor="middle" className="graph-account-handle">
        {label}
      </text>
    </g>
  );
}

function SignalTimeline({ edges }: { edges: GraphEdge[] }) {
  const y = HEIGHT - 26;
  const startX = 60;
  const endX = WIDTH - 60;
  const span = endX - startX;
  if (edges.length === 0) {
    return (
      <g className="graph-timeline">
        <line x1={startX} y1={y} x2={endX} y2={y} className="graph-timeline-axis" />
        <text x={WIDTH / 2} y={y - 10} textAnchor="middle" className="graph-timeline-empty">
          no cross-account signals yet
        </text>
      </g>
    );
  }
  return (
    <g className="graph-timeline">
      <line x1={startX} y1={y} x2={endX} y2={y} className="graph-timeline-axis" />
      {edges.map((e, i) => {
        const x = edges.length === 1 ? WIDTH / 2 : startX + (span * i) / (edges.length - 1);
        return (
          <g key={`${e.kind}-${i}`} transform={`translate(${x} ${y})`}>
            <circle r={5} className="graph-timeline-dot" fill={edgeColor(e.kind, e.active)} />
            <text y={18} textAnchor="middle" className="graph-timeline-label">
              {e.kind}
            </text>
          </g>
        );
      })}
    </g>
  );
}

function EmptyGraph() {
  return (
    <div className="graph-empty">
      <svg viewBox="0 0 320 160" className="graph-empty-svg" aria-hidden="true">
        <defs>
          <linearGradient id="emptyGrad" x1="0%" y1="0%" x2="100%" y2="100%">
            <stop offset="0%" stopColor="var(--cove-deep)" />
            <stop offset="100%" stopColor="var(--mesa-deep)" />
          </linearGradient>
        </defs>
        <circle cx="110" cy="80" r="26" fill="none" stroke="url(#emptyGrad)" strokeWidth="2" strokeDasharray="4 4" />
        <circle cx="210" cy="80" r="26" fill="none" stroke="url(#emptyGrad)" strokeWidth="2" strokeDasharray="4 4" />
        <path d="M136 80 Q160 50 184 80" fill="none" stroke="var(--ink-faint)" strokeWidth="1.5" strokeDasharray="3 3" />
        <text x="160" y="130" textAnchor="middle" className="graph-empty-title">
          identity not yet resolved
        </text>
      </svg>
      <p className="graph-empty-body">
        Neptune is still watching. A couple graph appears here once both identities and their accounts resolve.
      </p>
    </div>
  );
}

function labelWidth(kind: string): number {
  return Math.max(48, kind.length * 7.2 + 18);
}

function initials(name: string): string {
  return name
    .split(" ")
    .map((p) => p[0])
    .join("")
    .slice(0, 2)
    .toUpperCase();
}
