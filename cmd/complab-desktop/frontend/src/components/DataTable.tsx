import React from 'react';
import {
  useReactTable,
  getCoreRowModel,
  flexRender,
  ColumnDef,
} from '@tanstack/react-table';

interface DataTableProps {
  headers: string[];
  rowNames: string[];
  data: number[][];
  title?: string;
}

export const DataTable: React.FC<DataTableProps> = ({ headers, rowNames, data, title }) => {
  const hasRowNames = rowNames && rowNames.length > 0;
  
  // Create columns
  const columns: ColumnDef<any>[] = React.useMemo(() => {
    const cols: ColumnDef<any>[] = [];
    
    if (hasRowNames) {
      cols.push({
        accessorKey: 'rowName',
        header: 'Sample',
        cell: info => <div className="font-medium">{String(info.getValue())}</div>,
      });
    }
    
    headers.forEach((header, index) => {
      cols.push({
        accessorKey: `col${index}`,
        header: header,
        cell: info => {
          const value = info.getValue() as number;
          return <div className="text-right">{value?.toFixed(4) || ''}</div>;
        },
      });
    });
    
    return cols;
  }, [headers, hasRowNames]);
  
  // Transform data for table
  const tableData = React.useMemo(() => {
    return data.map((row, rowIndex) => {
      const rowData: any = {};
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
  });
  
  return (
    <div className="flex flex-col space-y-2">
      {title && <h3 className="text-lg font-semibold text-white">{title}</h3>}
      <div className="overflow-auto max-h-96 border border-gray-600 rounded-lg">
        <table className="w-full text-sm text-left text-gray-300">
          <thead className="text-xs uppercase bg-gray-700 text-gray-300 sticky top-0">
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
              <tr key={row.id} className="border-b border-gray-700 hover:bg-gray-800">
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
      <div className="text-sm text-gray-400">
        Showing {data.length} rows × {headers.length} columns
      </div>
    </div>
  );
};