// GoPCA Suite
//
// Copyright © 2025-2026 Rune Mathisen <devel@bitjungle.com>
//
// This file is part of GoPCA Suite.
//
// GoPCA Suite is source-available software with free binary redistribution.
// Official compiled binary releases may be used and redistributed free of charge
// under the GoPCA Suite Source-Available Freeware License.
//
// The source code is provided for viewing, review, education, security analysis,
// research, interoperability analysis, and evaluation only.
//
// Modification, redistribution, publication, sublicensing, reuse, incorporation
// into another project, or creation of derivative works based on the source code
// is not permitted without prior written permission from the copyright holder.
//
// Usage Restriction: GoPCA Suite may not be used, directly or indirectly, for
// military, warfare, weapons, intelligence, surveillance, targeting, or
// law-enforcement surveillance applications.
//
// See LICENSE for the full license terms.

import React from 'react';
import {
  useReactTable,
  getCoreRowModel,
  flexRender,
  ColumnDef,
  RowSelectionState
} from '@tanstack/react-table';

// Interface for row data in the table
interface TableRowData {
  rowName?: string;
  [key: string]: string | number | undefined;
}

// Component for the select-all checkbox to avoid React Hooks in non-component functions
const SelectAllCheckbox: React.FC<{ table: any }> = ({ table }) => {
  const checkboxRef = React.useRef<HTMLInputElement>(null);

  React.useEffect(() => {
    if (checkboxRef.current) {
      checkboxRef.current.indeterminate = table.getIsSomeRowsSelected();
    }
  });

  return (
    <input
      ref={checkboxRef}
      type="checkbox"
      checked={table.getIsAllRowsSelected()}
      onChange={table.getToggleAllRowsSelectedHandler()}
      className="w-4 h-4 text-blue-600 bg-white dark:bg-gray-700 border-gray-300 dark:border-gray-600 rounded focus:ring-blue-500"
    />
  );
};

interface DataTableProps {
  headers: string[];
  rowNames: string[];
  data: number[][];
  title?: string;
  enableRowSelection?: boolean;
  enableColumnSelection?: boolean;
  onRowSelectionChange?: (selectedRows: number[]) => void;
  onColumnSelectionChange?: (selectedColumns: number[]) => void;
  externalSelectedRows?: number[];
  highlightExternalSelections?: boolean;
}

export const DataTable: React.FC<DataTableProps> = ({
  headers,
  rowNames,
  data,
  title,
  enableRowSelection = false,
  enableColumnSelection = false,
  onRowSelectionChange,
  onColumnSelectionChange,
  externalSelectedRows,
  highlightExternalSelections: _highlightExternalSelections
}) => {
  const hasRowNames = rowNames && rowNames.length > 0;
  const [rowSelection, setRowSelection] = React.useState<RowSelectionState>({});
  const [columnSelection, setColumnSelection] = React.useState<Record<string, boolean>>({});
  // Removed unused variable: const isFirstRender = React.useRef(true);

  // Initialize selection states when component mounts with data
  React.useEffect(() => {
    if (!enableRowSelection && !enableColumnSelection) {
return;
}
    if (data.length === 0) {
return;
}

    // Initialize row selection
    if (enableRowSelection && Object.keys(rowSelection).length === 0) {
      const initialRowSelection: RowSelectionState = {};
      data.forEach((_, index) => {
        initialRowSelection[index] = true;
      });
      setRowSelection(initialRowSelection);
    }

    // Initialize column selection
    if (enableColumnSelection && Object.keys(columnSelection).length === 0) {
      const initialSelection: Record<string, boolean> = {};
      headers.forEach((_, index) => {
        initialSelection[`col${index}`] = true;
      });
      setColumnSelection(initialSelection);
    }
  }, [data.length, headers.length, enableRowSelection, enableColumnSelection]);

  // Update internal state when external selection changes (one-way sync)
  React.useEffect(() => {
    if (externalSelectedRows !== undefined && enableRowSelection) {
      const newSelection: RowSelectionState = {};
      data.forEach((_, index) => {
        newSelection[index] = externalSelectedRows.includes(index);
      });
      setRowSelection(newSelection);
    }
  }, [externalSelectedRows?.length]); // Only watch length to avoid deep comparison

  // Notify parent of row selection changes
  React.useEffect(() => {
    if (onRowSelectionChange && Object.keys(rowSelection).length > 0) {
      const selectedIndices = Object.keys(rowSelection)
        .filter(key => rowSelection[key])
        .map(key => parseInt(key));
      onRowSelectionChange(selectedIndices);
    }
  }, [rowSelection, onRowSelectionChange]);

  // Notify parent of column selection changes
  React.useEffect(() => {
    if (onColumnSelectionChange && Object.keys(columnSelection).length > 0) {
      const selectedIndices = headers
        .map((_, index) => index)
        .filter(index => columnSelection[`col${index}`] !== false);
      onColumnSelectionChange(selectedIndices);
    }
  }, [columnSelection, headers, onColumnSelectionChange]);

  // Create columns
  const columns: ColumnDef<TableRowData>[] = React.useMemo(() => {
    const cols: ColumnDef<TableRowData>[] = [];

    // Add selection column if enabled
    if (enableRowSelection) {
      cols.push({
        id: 'select',
        header: ({ table }) => <SelectAllCheckbox table={table} />,
        cell: ({ row }) => (
          <input
            type="checkbox"
            checked={row.getIsSelected()}
            onChange={row.getToggleSelectedHandler()}
            className="w-4 h-4 text-blue-600 bg-gray-700 border-gray-600 rounded focus:ring-blue-500"
          />
        ),
        size: 40
      });
    }

    if (hasRowNames) {
      cols.push({
        accessorKey: 'rowName',
        header: 'Sample_ID',
        cell: info => <div className="font-medium">{String(info.getValue())}</div>
      });
    }

    headers.forEach((header, index) => {
      const colId = `col${index}`;
      cols.push({
        accessorKey: colId,
        header: enableColumnSelection ? () => (
          <div className="flex flex-col items-center space-y-1">
            <input
              type="checkbox"
              checked={columnSelection[colId] !== false}
              onChange={(e) => {
                setColumnSelection(prev => ({
                  ...prev,
                  [colId]: e.target.checked
                }));
              }}
              className="w-4 h-4 text-blue-600 bg-white dark:bg-gray-700 border-gray-300 dark:border-gray-600 rounded focus:ring-blue-500"
            />
            <span>{header}</span>
          </div>
        ) : header,
        cell: info => {
          const value = info.getValue() as number;
          return <div className="text-right">{value?.toFixed(4) || ''}</div>;
        }
      });
    });

    return cols;
  }, [headers, hasRowNames, enableRowSelection, enableColumnSelection, columnSelection]);

  // Transform data for table
  const tableData = React.useMemo(() => {
    return data.map((row, rowIndex) => {
      const rowData: TableRowData = {};
      if (hasRowNames) {
        rowData.rowName = rowNames[rowIndex];
      }
      row.forEach((value, colIndex) => {
        rowData[`col${colIndex}`] = value;
      });
      return rowData;
    });
  }, [data, rowNames, hasRowNames]);

  const table = useReactTable({
    data: tableData,
    columns,
    getCoreRowModel: getCoreRowModel(),
    enableRowSelection: enableRowSelection,
    onRowSelectionChange: setRowSelection,
    state: {
      rowSelection
    }
  });

  // Calculate selected counts
  const selectedRowCount = Object.keys(rowSelection).filter(key => rowSelection[key]).length;
  const selectedColCount = headers.filter((_, index) => columnSelection[`col${index}`] !== false).length;

  return (
    <div className="flex flex-col space-y-2">
      {title && <h3 className="text-lg font-semibold text-gray-900 dark:text-white">{title}</h3>}
      <div className="overflow-auto max-h-96 border border-gray-300 dark:border-gray-600 rounded-lg">
        <table className="w-full text-sm text-left text-gray-700 dark:text-gray-300">
          <thead className="text-xs uppercase bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 sticky top-0">
            {table.getHeaderGroups().map(headerGroup => (
              <tr key={headerGroup.id}>
                {headerGroup.headers.map(header => (
                  <th key={header.id} className="px-4 py-2">
                    {header.isPlaceholder
                      ? null
                      : flexRender(
                          header.column.columnDef.header,
                          header.getContext()
                        )}
                  </th>
                ))}
              </tr>
            ))}
          </thead>
          <tbody>
            {table.getRowModel().rows.map(row => (
              <tr key={row.id} className="border-b border-gray-200 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-800">
                {row.getVisibleCells().map(cell => (
                  <td key={cell.id} className="px-4 py-2">
                    {flexRender(cell.column.columnDef.cell, cell.getContext())}
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      <div className="text-sm text-gray-600 dark:text-gray-400">
        {enableRowSelection || enableColumnSelection ? (
          <div className="flex justify-between">
            <span>Showing {data.length} rows × {headers.length} columns</span>
            <span>
              {enableRowSelection && `Selected: ${selectedRowCount} rows`}
              {enableRowSelection && enableColumnSelection && ', '}
              {enableColumnSelection && `${selectedColCount} columns`}
            </span>
          </div>
        ) : (
          <span>Showing {data.length} rows × {headers.length} columns</span>
        )}
      </div>
    </div>
  );
};