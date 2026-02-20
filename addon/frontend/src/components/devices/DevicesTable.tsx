import { memo } from "react";
import {
  flexRender,
  type Row as TanStackRow,
  type Table as TanStackTable
} from "@tanstack/react-table";

import {
  Pagination,
  PaginationContent,
  PaginationEllipsis,
  PaginationItem,
  PaginationLink,
  PaginationNext,
  PaginationPrevious
} from "@/components/ui/pagination";
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

function buildPageItems(pageIndex: number, pageCount: number): Array<number | "ellipsis"> {
  if (pageCount <= 7) {
    return Array.from({ length: pageCount }, (_, index) => index);
  }

  const pages = new Set<number>([0, pageCount - 1, pageIndex - 1, pageIndex, pageIndex + 1]);
  const ordered = Array.from(pages)
    .filter((page) => page >= 0 && page < pageCount)
    .sort((a, b) => a - b);

  const items: Array<number | "ellipsis"> = [];
  for (let index = 0; index < ordered.length; index += 1) {
    const current = ordered[index];
    const previous = index > 0 ? ordered[index - 1] : null;
    if (previous !== null && current-previous > 1) {
      items.push("ellipsis");
    }
    items.push(current);
  }
  return items;
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
  const totalRows = table.getPrePaginationRowModel().rows.length;
  const pageIndex = table.getState().pagination.pageIndex;
  const pageCount = Math.max(table.getPageCount(), 1);
  const pageItems = buildPageItems(pageIndex, pageCount);
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
        <p className="text-sm text-muted-foreground">
          Showing {rows.length} of {totalRows} rows
        </p>

        <Pagination className="mx-0 w-auto justify-end">
          <PaginationContent>
            <PaginationItem>
              <PaginationPrevious
                onClick={() => table.previousPage()}
                disabled={!table.getCanPreviousPage()}
              />
            </PaginationItem>

            {pageItems.map((item, index) => (
              <PaginationItem key={`${item}-${index}`}>
                {item === "ellipsis" ? (
                  <PaginationEllipsis />
                ) : (
                  <PaginationLink
                    isActive={item === pageIndex}
                    onClick={() => table.setPageIndex(item)}
                  >
                    {item + 1}
                  </PaginationLink>
                )}
              </PaginationItem>
            ))}

            <PaginationItem>
              <PaginationNext
                onClick={() => table.nextPage()}
                disabled={!table.getCanNextPage()}
              />
            </PaginationItem>
          </PaginationContent>
        </Pagination>
      </div>
    </div>
  );
}
