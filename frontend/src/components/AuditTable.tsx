import { createColumnHelper, flexRender, getCoreRowModel, useReactTable } from "@tanstack/react-table";
import type { AuditEvent } from "../api/types";
import { EmptyState } from "./EmptyState";

const columnHelper = createColumnHelper<AuditEvent>();

const columns = [
  columnHelper.accessor("step_index", { header: "Step" }),
  columnHelper.accessor("entity_type", { header: "Entity" }),
  columnHelper.accessor("event", { header: "Event", cell: (info) => <code>{info.getValue()}</code> }),
  columnHelper.accessor("detail", {
    header: "Detail",
    cell: (info) => {
      const raw = info.getValue();
      if (!raw) return "—";
      try {
        const parsed = JSON.parse(raw);
        return <span className="audit-detail">{JSON.stringify(parsed)}</span>;
      } catch {
        return raw;
      }
    },
  }),
  columnHelper.accessor("created_at", { header: "When" }),
];

export function AuditTable({ events }: { events: AuditEvent[] }) {
  const table = useReactTable({ data: events, columns, getCoreRowModel: getCoreRowModel() });

  if (events.length === 0) {
    return <EmptyState variant="empty" title="No audit events yet" message="The watch loop logs every stage as events arrive." />;
  }

  return (
    <table className="data-table data-table--audit">
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
