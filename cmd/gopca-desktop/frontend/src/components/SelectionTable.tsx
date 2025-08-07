// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

import React, { useRef, useMemo, useEffect } from 'react';
import { useVirtualizer } from '@tanstack/react-virtual';

interface SelectionTableProps {
  headers: string[];
  rowNames: string[];
  data: number[][];
  title?: string;
  onRowSelectionChange?: (selectedRows: number[]) => void;
  onColumnSelectionChange?: (selectedColumns: number[]) => void;
}

export const SelectionTable: React.FC<SelectionTableProps> = ({
  headers,
  rowNames,
  data,
  title,
  onRowSelectionChange,
  onColumnSelectionChange,
}) => {
  // Selection states
  const [rowSelection, setRowSelection] = React.useState<Record<number, boolean>>({});
  const [columnSelection, setColumnSelection] = React.useState<Record<number, boolean>>({});

  // Initialize selections as all selected
  useEffect(() => {
    if (data.length > 0 && Object.keys(rowSelection).length === 0) {
      const initialRowSelection: Record<number, boolean> = {};
      data.forEach((_, index) => {
        initialRowSelection[index] = true;
      });
      setRowSelection(initialRowSelection);
    }

    if (headers.length > 0 && Object.keys(columnSelection).length === 0) {
      const initialColSelection: Record<number, boolean> = {};
      headers.forEach((_, index) => {
        initialColSelection[index] = true;
      });
      setColumnSelection(initialColSelection);
    }
  }, [data.length, headers.length]);

  // Notify parent of selection changes
  useEffect(() => {
    if (Object.keys(rowSelection).length === 0) return;
    
    if (onRowSelectionChange) {
      const selectedIndices = Object.keys(rowSelection)
        .filter(key => rowSelection[Number(key)])
        .map(key => Number(key));
      onRowSelectionChange(selectedIndices);
    }
  }, [rowSelection, onRowSelectionChange]);

  useEffect(() => {
    if (Object.keys(columnSelection).length === 0) return;
    
    if (onColumnSelectionChange) {
      const selectedIndices = Object.keys(columnSelection)
        .filter(key => columnSelection[Number(key)])
        .map(key => Number(key));
      onColumnSelectionChange(selectedIndices);
    }
  }, [columnSelection, onColumnSelectionChange]);

  // Refs for virtualized lists
  const rowListRef = useRef<HTMLDivElement>(null);
  const colListRef = useRef<HTMLDivElement>(null);

  // Virtualizers for rows and columns
  const rowVirtualizer = useVirtualizer({
    count: rowNames.length,
    getScrollElement: () => rowListRef.current,
    estimateSize: () => 35,
    overscan: 5,
  });

  const colVirtualizer = useVirtualizer({
    count: headers.length,
    getScrollElement: () => colListRef.current,
    estimateSize: () => 120,
    horizontal: true,
    overscan: 5,
  });

  // Calculate selected counts
  const selectedRowCount = Object.values(rowSelection).filter(Boolean).length;
  const selectedColCount = Object.values(columnSelection).filter(Boolean).length;

  // Toggle all rows
  const toggleAllRows = () => {
    const allSelected = selectedRowCount === rowNames.length;
    const newSelection: Record<number, boolean> = {};
    rowNames.forEach((_, index) => {
      newSelection[index] = !allSelected;
    });
    setRowSelection(newSelection);
  };

  // Toggle all columns
  const toggleAllColumns = () => {
    const allSelected = selectedColCount === headers.length;
    const newSelection: Record<number, boolean> = {};
    headers.forEach((_, index) => {
      newSelection[index] = !allSelected;
    });
    setColumnSelection(newSelection);
  };

  // Get data preview (first 10x10)
  const previewData = useMemo(() => {
    const maxRows = Math.min(10, data.length);
    const maxCols = Math.min(10, headers.length);
    const preview: number[][] = [];
    
    for (let i = 0; i < maxRows; i++) {
      preview.push(data[i].slice(0, maxCols));
    }
    
    return preview;
  }, [data, headers]);

  return (
    <div className="flex flex-col h-full">
      {title && <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">{title}</h3>}
      
      {/* Info banner */}
      <div className="bg-blue-100 dark:bg-blue-900 border border-blue-300 dark:border-blue-700 rounded-lg p-3 mb-4">
        <p className="text-sm text-blue-800 dark:text-blue-200">
          <strong>Large dataset mode:</strong> Showing selection controls for {rowNames.length} rows × {headers.length} columns.
          Use the lists below to select/deselect rows and columns for PCA analysis.
        </p>
      </div>

      <div className="flex gap-4 h-[600px] overflow-hidden">
        {/* Left panel - Row selection */}
        <div className="flex-none w-64 bg-white dark:bg-gray-800 rounded-lg border border-gray-300 dark:border-gray-600 p-4">
          <div className="flex items-center justify-between mb-3">
            <h4 className="font-medium text-gray-900 dark:text-white">Rows</h4>
            <button
              onClick={toggleAllRows}
              className="text-xs px-2 py-1 bg-gray-200 dark:bg-gray-700 hover:bg-gray-300 dark:hover:bg-gray-600 rounded"
            >
              {selectedRowCount === rowNames.length ? 'Deselect All' : 'Select All'}
            </button>
          </div>
          
          <div
            ref={rowListRef}
            className="h-[calc(100%-40px)] overflow-auto border border-gray-200 dark:border-gray-700 rounded"
          >
            <div
              style={{
                height: `${rowVirtualizer.getTotalSize()}px`,
                width: '100%',
                position: 'relative',
              }}
            >
              {rowVirtualizer.getVirtualItems().map((virtualRow) => (
                <div
                  key={virtualRow.key}
                  style={{
                    position: 'absolute',
                    top: 0,
                    left: 0,
                    width: '100%',
                    height: `${virtualRow.size}px`,
                    transform: `translateY(${virtualRow.start}px)`,
                  }}
                  className="flex items-center px-2 hover:bg-gray-50 dark:hover:bg-gray-700"
                >
                  <input
                    type="checkbox"
                    checked={rowSelection[virtualRow.index] ?? true}
                    onChange={(e) => {
                      setRowSelection(prev => ({
                        ...prev,
                        [virtualRow.index]: e.target.checked
                      }));
                    }}
                    className="mr-2 w-4 h-4 text-blue-600 bg-white dark:bg-gray-700 border-gray-300 dark:border-gray-600 rounded focus:ring-blue-500"
                  />
                  <span className="text-sm text-gray-700 dark:text-gray-300 truncate">
                    {rowNames[virtualRow.index]}
                  </span>
                </div>
              ))}
            </div>
          </div>
        </div>

        {/* Center/Right panel */}
        <div className="flex-1 flex flex-col min-w-0">
          {/* Top panel - Column selection */}
          <div className="bg-white dark:bg-gray-800 rounded-lg border border-gray-300 dark:border-gray-600 p-4 mb-4 overflow-hidden">
            <div className="flex items-center justify-between mb-3">
              <h4 className="font-medium text-gray-900 dark:text-white">Columns</h4>
              <button
                onClick={toggleAllColumns}
                className="text-xs px-2 py-1 bg-gray-200 dark:bg-gray-700 hover:bg-gray-300 dark:hover:bg-gray-600 rounded"
              >
                {selectedColCount === headers.length ? 'Deselect All' : 'Select All'}
              </button>
            </div>
            
            <div
              ref={colListRef}
              className="h-20 overflow-x-auto overflow-y-hidden border border-gray-200 dark:border-gray-700 rounded"
              style={{ 
                overflowX: 'auto',
                overflowY: 'hidden',
                maxWidth: '100%'
              }}
            >
              <div
                style={{
                  width: `${colVirtualizer.getTotalSize()}px`,
                  height: '100%',
                  position: 'relative',
                }}
              >
                {colVirtualizer.getVirtualItems().map((virtualCol) => (
                  <div
                    key={virtualCol.key}
                    style={{
                      position: 'absolute',
                      top: 0,
                      left: 0,
                      height: '100%',
                      width: `${virtualCol.size}px`,
                      transform: `translateX(${virtualCol.start}px)`,
                    }}
                    className="flex flex-col items-center justify-center px-2 hover:bg-gray-50 dark:hover:bg-gray-700"
                  >
                    <input
                      type="checkbox"
                      checked={columnSelection[virtualCol.index] ?? true}
                      onChange={(e) => {
                        setColumnSelection(prev => ({
                          ...prev,
                          [virtualCol.index]: e.target.checked
                        }));
                      }}
                      className="mb-1 w-4 h-4 text-blue-600 bg-white dark:bg-gray-700 border-gray-300 dark:border-gray-600 rounded focus:ring-blue-500"
                    />
                    <span className="text-xs text-gray-700 dark:text-gray-300 truncate max-w-full">
                      {headers[virtualCol.index]}
                    </span>
                  </div>
                ))}
              </div>
            </div>
          </div>

          {/* Data preview */}
          <div className="flex-1 bg-white dark:bg-gray-800 rounded-lg border border-gray-300 dark:border-gray-600 p-4">
            <h4 className="font-medium text-gray-900 dark:text-white mb-3">Data Preview (first 10×10)</h4>
            <div className="overflow-auto">
              <table className="text-xs">
                <thead>
                  <tr>
                    <th className="px-2 py-1 text-left text-gray-600 dark:text-gray-400">Sample</th>
                    {headers.slice(0, 10).map((header, i) => (
                      <th key={i} className="px-2 py-1 text-right text-gray-600 dark:text-gray-400 truncate max-w-[100px]">
                        {header}
                      </th>
                    ))}
                    {headers.length > 10 && (
                      <th className="px-2 py-1 text-center text-gray-600 dark:text-gray-400">...</th>
                    )}
                  </tr>
                </thead>
                <tbody>
                  {previewData.map((row, rowIdx) => (
                    <tr key={rowIdx} className="border-t border-gray-200 dark:border-gray-700">
                      <td className="px-2 py-1 text-gray-700 dark:text-gray-300 truncate max-w-[100px]">
                        {rowNames[rowIdx]}
                      </td>
                      {row.map((value, colIdx) => (
                        <td key={colIdx} className="px-2 py-1 text-right text-gray-700 dark:text-gray-300">
                          {value != null && !isNaN(value) ? value.toFixed(2) : 'NaN'}
                        </td>
                      ))}
                      {headers.length > 10 && (
                        <td className="px-2 py-1 text-center text-gray-500 dark:text-gray-500">...</td>
                      )}
                    </tr>
                  ))}
                  {data.length > 10 && (
                    <tr className="border-t border-gray-200 dark:border-gray-700">
                      <td className="px-2 py-1 text-center text-gray-500 dark:text-gray-500">...</td>
                      {headers.slice(0, Math.min(10, headers.length)).map((_, i) => (
                        <td key={i} className="px-2 py-1 text-center text-gray-500 dark:text-gray-500">...</td>
                      ))}
                      {headers.length > 10 && (
                        <td className="px-2 py-1 text-center text-gray-500 dark:text-gray-500">...</td>
                      )}
                    </tr>
                  )}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </div>

      {/* Selection summary */}
      <div className="mt-4 text-sm text-gray-600 dark:text-gray-400">
        <div className="flex justify-between">
          <span>Matrix size: {rowNames.length} rows × {headers.length} columns ({(rowNames.length * headers.length).toLocaleString()} cells)</span>
          <span>Selected: {selectedRowCount} rows, {selectedColCount} columns</span>
        </div>
      </div>
    </div>
  );
};