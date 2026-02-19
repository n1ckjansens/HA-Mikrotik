import { memo } from "react";
import {
  flexRender,
  type Row as TanStackRow,
  type Table as TanStackTable
} from "@tanstack/react-table";
import { ChevronLeft, ChevronRight } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow
} from "@/components/ui/table";
import type { Device } from "@/types/device";

type Props = {
  table: TanStackTable<Device>;
  isLoading: boolean;
  density: "comfortable" | "compact";
};

type RowProps = {
  row: TanStackRow<Device>;
  rowHeight: string;
  columnsVersion: string;
};

function columnWidthClass(columnId: string) {
  if (columnId === "name") {
    return "w-72";
  }
  if (columnId === "last_ip") {
    return "w-36";
  }
  if (columnId === "last_subnet") {
    return "w-40";
  }
  if (columnId === "vendor") {
    return "w-44";
  }
  if (columnId === "source") {
    return "w-28";
  }
  if (columnId === "last_seen") {
    return "w-32";
  }
  if (columnId === "registration") {
    return "w-32";
  }
  if (columnId === "actions") {
    return "w-12";
  }
  return "";
}

const DeviceDataRow = memo(
  function DeviceDataRow({ row, rowHeight }: RowProps) {
    return (
      <TableRow
        data-state={row.getIsSelected() && "selected"}
        className="hover:bg-muted/40"
      >
        {row.getVisibleCells().map((cell) => (
          <TableCell
            key={cell.id}
            className={`${rowHeight} ${columnWidthClass(cell.column.id)} align-top overflow-hidden`}
          >
            {flexRender(cell.column.columnDef.cell, cell.getContext())}
          </TableCell>
        ))}
      </TableRow>
    );
  },
  (previous, next) =>
    previous.row.original === next.row.original &&
    previous.rowHeight === next.rowHeight &&
    previous.columnsVersion === next.columnsVersion
);

export function DevicesTable({
  table,
  isLoading,
  density
}: Props) {
  const rows = table.getRowModel().rows;
  const rowHeight = density === "compact" ? "py-1" : "py-2";
  const columnsVersion = table
    .getVisibleLeafColumns()
    .map((column) => column.id)
    .join("|");

  return (
    <div className="space-y-3">
      <div className="rounded-md border">
        <Table className="table-fixed">
          <TableHeader>
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow key={headerGroup.id}>
                {headerGroup.headers.map((header) => (
                  <TableHead
                    key={header.id}
                    className={`${columnWidthClass(header.column.id)} overflow-hidden`}
                  >
                    {header.isPlaceholder
                      ? null
                      : flexRender(header.column.columnDef.header, header.getContext())}
                  </TableHead>
                ))}
              </TableRow>
            ))}
          </TableHeader>

          <TableBody>
            {isLoading
              ? Array.from({ length: 8 }).map((_, index) => (
                  <TableRow key={`skeleton-${index}`}>
                    {table.getAllLeafColumns().map((column) => (
                      <TableCell
                        key={`${column.id}-${index}`}
                        className={`${rowHeight} ${columnWidthClass(column.id)} overflow-hidden`}
                      >
                        <Skeleton className="h-5 w-full" />
                      </TableCell>
                    ))}
                  </TableRow>
                ))
              : null}

            {!isLoading && rows.length === 0 ? (
              <TableRow>
                <TableCell colSpan={table.getAllLeafColumns().length} className="h-24 text-center">
                  No devices to display.
                </TableCell>
              </TableRow>
            ) : null}

            {!isLoading
              ? rows.map((row) => (
                  <DeviceDataRow
                    key={row.id}
                    row={row}
                    rowHeight={rowHeight}
                    columnsVersion={columnsVersion}
                  />
                ))
              : null}
          </TableBody>
        </Table>
      </div>

      <div className="flex items-center justify-between">
        <p className="text-sm text-muted-foreground">{rows.length} rows</p>
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => table.previousPage()}
            disabled={!table.getCanPreviousPage()}
          >
            <ChevronLeft className="mr-1 h-4 w-4" /> Prev
          </Button>
          <p className="text-sm text-muted-foreground">
            Page {table.getState().pagination.pageIndex + 1} of {table.getPageCount()}
          </p>
          <Button
            variant="outline"
            size="sm"
            onClick={() => table.nextPage()}
            disabled={!table.getCanNextPage()}
          >
            Next <ChevronRight className="ml-1 h-4 w-4" />
          </Button>
        </div>
      </div>
    </div>
  );
}
