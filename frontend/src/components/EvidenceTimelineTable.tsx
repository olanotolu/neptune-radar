import { createColumnHelper, flexRender, getCoreRowModel, useReactTable } from "@tanstack/react-table";
import type { Evidence } from "../api/types";

const columnHelper = createColumnHelper<Evidence>();

const columns = [
  columnHelper.accessor("kind", { header: "Signal", cell: (info) => <code>{info.getValue()}</code> }),
  columnHelper.accessor("description", { header: "Evidence" }),
  columnHelper.accessor("weight", {
    header: "Weight",
    cell: (info) => {
      const w = info.getValue();
      return <span className={w >= 0 ? "weight-positive" : "weight-negative"}>{w >= 0 ? "+" : ""}{w.toFixed(2)}</span>;
    },
  }),
  columnHelper.accessor("confirmed", {
    header: "Confirmed",
    cell: (info) => (info.getValue() ? "✓" : "—"),
  }),
];

export function EvidenceTimelineTable({ evidence }: { evidence: Evidence[] }) {
  const table = useReactTable({ data: evidence, columns, getCoreRowModel: getCoreRowModel() });

  if (evidence.length === 0) {
    return <div className="empty-state">No evidence collected yet.</div>;
  }

  return (
    <table className="data-table">
      <thead>
        {table.getHeaderGroups().map((hg) => (
          <tr key={hg.id}>
            {hg.headers.map((h) => (
              <th key={h.id}>{flexRender(h.column.columnDef.header, h.getContext())}</th>
            ))}
          </tr>
        ))}
      </thead>
      <tbody>
        {table.getRowModel().rows.map((row) => (
          <tr key={row.id}>
            {row.getVisibleCells().map((cell) => (
              <td key={cell.id}>{flexRender(cell.column.columnDef.cell, cell.getContext())}</td>
            ))}
          </tr>
        ))}
      </tbody>
    </table>
  );
}
