import type { CoupleGraph } from "../api/types";

const WIDTH = 560;
const HEIGHT = 260;

export function CoupleGraphSvg({ graph }: { graph: CoupleGraph }) {
  const persons = graph.nodes.filter((n) => n.type === "person");
  const accounts = graph.nodes.filter((n) => n.type === "account");
  if (persons.length < 2 || accounts.length < 2) {
    return <div className="graph-empty">Not enough resolved identity yet — couple hasn't formed.</div>;
  }
  const [personA, personB] = persons;
  const accountFor = (personId: string) =>
    accounts.find((a) => graph.edges.some((e) => e.kind === "owns_account" && e.from === personId && e.to === a.id));
  const accountA = accountFor(personA.id) ?? accounts[0];
  const accountB = accountFor(personB.id) ?? accounts[1];

  const leftX = 120;
  const rightX = WIDTH - 120;
  const personY = 50;
  const accountY = HEIGHT - 60;

  const crossEdges = graph.edges.filter(
    (e) => e.kind !== "owns_account" &&
      ((e.from === accountA.id && e.to === accountB.id) || (e.from === accountB.id && e.to === accountA.id))
  );

  return (
    <svg viewBox={`0 0 ${WIDTH} ${HEIGHT}`} className="couple-graph-svg" role="img" aria-label="Couple identity graph">
      <line x1={leftX} y1={personY} x2={leftX} y2={accountY} className="graph-edge graph-edge--owns" />
      <line x1={rightX} y1={personY} x2={rightX} y2={accountY} className="graph-edge graph-edge--owns" />

      {crossEdges.map((e, i) => {
        const fromLeft = e.from === accountA.id;
        const y = accountY + 26 + i * 20;
        return (
          <g key={`${e.kind}-${e.from}-${e.to}`}>
            <line
              x1={fromLeft ? leftX : rightX}
              y1={y}
              x2={fromLeft ? rightX : leftX}
              y2={y}
              className={`graph-edge graph-edge--${e.kind} ${e.active ? "graph-edge--active" : "graph-edge--inactive"}`}
              markerEnd={fromLeft ? "url(#arrow)" : undefined}
              markerStart={!fromLeft ? "url(#arrow)" : undefined}
            />
            <text x={WIDTH / 2} y={y - 6} className="graph-edge-label" textAnchor="middle">
              {e.kind}
              {!e.active ? " (inactive)" : ""}
            </text>
          </g>
        );
      })}

      <defs>
        <marker id="arrow" markerWidth="8" markerHeight="8" refX="6" refY="4" orient="auto">
          <path d="M0,0 L8,4 L0,8 z" className="graph-arrowhead" />
        </marker>
      </defs>

      <g>
        <circle cx={leftX} cy={personY} r={22} className="graph-node graph-node--person" />
        <text x={leftX} y={personY + 5} textAnchor="middle" className="graph-node-label">
          {initials(personA.label)}
        </text>
        <text x={leftX} y={personY - 32} textAnchor="middle" className="graph-caption">
          {personA.label}
        </text>

        <rect x={leftX - 45} y={accountY - 16} width={90} height={32} rx={8} className="graph-node graph-node--account" />
        <text x={leftX} y={accountY + 5} textAnchor="middle" className="graph-node-label graph-node-label--account">
          {accountA.label}
        </text>
      </g>

      <g>
        <circle cx={rightX} cy={personY} r={22} className="graph-node graph-node--person" />
        <text x={rightX} y={personY + 5} textAnchor="middle" className="graph-node-label">
          {initials(personB.label)}
        </text>
        <text x={rightX} y={personY - 32} textAnchor="middle" className="graph-caption">
          {personB.label}
        </text>

        <rect x={rightX - 45} y={accountY - 16} width={90} height={32} rx={8} className="graph-node graph-node--account" />
        <text x={rightX} y={accountY + 5} textAnchor="middle" className="graph-node-label graph-node-label--account">
          {accountB.label}
        </text>
      </g>
    </svg>
  );
}

function initials(name: string): string {
  return name
    .split(" ")
    .map((p) => p[0])
    .join("")
    .slice(0, 2)
    .toUpperCase();
}
