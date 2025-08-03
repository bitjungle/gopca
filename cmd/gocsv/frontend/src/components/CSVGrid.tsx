import React, { useCallback, useMemo, useRef, useState, useEffect } from 'react';
import { AgGridReact } from 'ag-grid-react';
import { ColDef, GridReadyEvent, CellValueChangedEvent, GridApi, ColumnApi } from 'ag-grid-community';
import 'ag-grid-community/styles/ag-grid.css';
import 'ag-grid-community/styles/ag-theme-quartz.css';
import { useTheme } from '../contexts/ThemeContext';

interface CSVGridProps {
    data: string[][];
    headers: string[];
    rowNames?: string[];
    onDataChange?: (rowIndex: number, colIndex: number, newValue: string) => void;
    onHeaderChange?: (colIndex: number, newHeader: string) => void;
    onRowNameChange?: (rowIndex: number, newRowName: string) => void;
}

export const CSVGrid: React.FC<CSVGridProps> = ({ 
    data, 
    headers,
    rowNames,
    onDataChange,
    onHeaderChange,
    onRowNameChange
}) => {
    // Validate inputs
    if (!data || !headers || data.length === 0 || headers.length === 0) {
        return <div className="w-full h-full flex items-center justify-center text-gray-500">No data to display</div>;
    }
    const gridRef = useRef<AgGridReact>(null);
    const [gridApi, setGridApi] = useState<GridApi | null>(null);
    const [columnApi, setColumnApi] = useState<ColumnApi | null>(null);
    const { theme } = useTheme();
    
    // Detect column types
    const detectColumnType = useCallback((colIndex: number): 'numeric' | 'text' | 'mixed' => {
        let hasNumeric = false;
        let hasText = false;
        
        for (let i = 0; i < Math.min(data.length, 100); i++) { // Sample first 100 rows
            const value = data[i]?.[colIndex];
            if (value && value.trim()) {
                if (!isNaN(Number(value))) {
                    hasNumeric = true;
                } else {
                    hasText = true;
                }
            }
        }
        
        if (hasNumeric && !hasText) return 'numeric';
        if (hasText && !hasNumeric) return 'text';
        return 'mixed';
    }, [data]);
    
    // Create column definitions
    const columnDefs = useMemo<ColDef[]>(() => {
        const cols: ColDef[] = [];
        
        // Add row name column if present
        if (rowNames && rowNames.length > 0) {
            cols.push({
                field: 'rowName',
                headerName: '',
                editable: true,
                sortable: true,
                filter: true,
                resizable: true,
                minWidth: 100,
                cellClass: 'row-name-cell',
                headerClass: 'row-name-header',
                pinned: 'left',
                lockPinned: true,
                cellStyle: {
                    backgroundColor: theme === 'dark' ? '#374151' : '#f3f4f6',
                    fontWeight: 'bold'
                }
            });
        }
        
        // Add data columns
        headers.forEach((header, index) => {
            const colType = detectColumnType(index);
            const isTargetColumn = header.toLowerCase().endsWith('#target') || 
                                 header.toLowerCase().endsWith('# target');
            
            cols.push({
                field: `col${index}`,
                headerName: header,
                editable: true,
                sortable: true,
                filter: true,
                resizable: true,
                minWidth: 100,
                cellClass: (params) => {
                    const classes = [];
                    if (colType === 'numeric') classes.push('numeric-cell');
                    if (colType === 'mixed') classes.push('mixed-cell');
                    if (isTargetColumn) classes.push('target-column');
                    if (!params.value || params.value.trim() === '') classes.push('missing-value');
                    return classes.join(' ');
                },
                headerClass: () => {
                    const classes = [];
                    if (colType === 'numeric') classes.push('numeric-header');
                    if (colType === 'mixed') classes.push('mixed-header');
                    if (isTargetColumn) classes.push('target-header');
                    return classes.join(' ');
                },
                headerTooltip: isTargetColumn ? 'Target column - excluded from PCA' : 
                              colType === 'numeric' ? 'Numeric column' : 
                              colType === 'mixed' ? 'Mixed column' : 'Text column',
                valueFormatter: (params: any) => {
                    if (colType === 'numeric' && params.value) {
                        const num = Number(params.value);
                        if (!isNaN(num)) {
                            return num.toFixed(4);
                        }
                    }
                    return params.value;
                },
            });
        });
        
        return cols;
    }, [headers, detectColumnType, rowNames, theme]);
    
    // Convert data to row format for ag-Grid
    const rowData = useMemo(() => {
        return data.map((row, rowIndex) => {
            const rowObj: any = { id: rowIndex };
            
            // Add row name if present
            if (rowNames && rowIndex < rowNames.length) {
                rowObj.rowName = rowNames[rowIndex];
            }
            
            // Add data columns
            headers.forEach((_, colIndex) => {
                rowObj[`col${colIndex}`] = row[colIndex] || '';
            });
            return rowObj;
        });
    }, [data, headers, rowNames]);
    
    // Grid ready event
    const onGridReady = useCallback((params: GridReadyEvent) => {
        setGridApi(params.api);
        setColumnApi(params.columnApi);
        
        // Size columns to fit available space
        params.api.sizeColumnsToFit();
    }, []);
    
    // Cell value changed event
    const onCellValueChanged = useCallback((event: CellValueChangedEvent) => {
        if (event.colDef?.field) {
            const rowIndex = event.node.data.id;
            
            if (event.colDef.field === 'rowName' && onRowNameChange) {
                onRowNameChange(rowIndex, event.newValue);
            } else if (onDataChange) {
                const colIndex = parseInt(event.colDef.field.replace('col', ''));
                onDataChange(rowIndex, colIndex, event.newValue);
            }
        }
    }, [onDataChange, onRowNameChange]);
    
    // Default column definition
    const defaultColDef = useMemo<ColDef>(() => ({
        flex: 1,
        minWidth: 100,
        editable: true,
        sortable: true,
        filter: true,
        resizable: true,
    }), []);
    
    // Grid options for performance
    const gridOptions = useMemo(() => ({
        // Performance optimizations
        animateRows: false,
        suppressColumnVirtualisation: false, // Enable column virtualization
        suppressRowVirtualisation: false, // Enable row virtualization (default)
        rowBuffer: 20, // Render 20 rows outside visible area
        debounceVerticalScrollbar: true, // Smoother scrolling
        
        // Editing
        singleClickEdit: true,
        stopEditingWhenCellsLoseFocus: true,
        
        // Selection
        rowSelection: 'multiple' as const,
        
        // Pagination for very large datasets
        pagination: data.length > 10000,
        paginationPageSize: 1000,
        paginationPageSizeSelector: [100, 500, 1000, 5000],
        
        // Other options
        enableCellTextSelection: true,
        ensureDomOrder: true,
    }), [data.length]);
    
    // Auto-size columns on window resize
    useEffect(() => {
        const handleResize = () => {
            if (gridApi) {
                gridApi.sizeColumnsToFit();
            }
        };
        
        window.addEventListener('resize', handleResize);
        return () => window.removeEventListener('resize', handleResize);
    }, [gridApi]);
    
    return (
        <div className="w-full h-full">
            <div 
                className={`${theme === 'dark' ? 'ag-theme-quartz-dark' : 'ag-theme-quartz'} w-full h-full`}
                style={{
                    '--ag-header-background-color': theme === 'dark' ? '#374151' : '#f3f4f6',
                    '--ag-header-foreground-color': theme === 'dark' ? '#e5e7eb' : '#111827',
                    '--ag-background-color': theme === 'dark' ? '#1f2937' : '#ffffff',
                    '--ag-foreground-color': theme === 'dark' ? '#e5e7eb' : '#111827',
                    '--ag-row-hover-color': theme === 'dark' ? '#374151' : '#f3f4f6',
                    '--ag-selected-row-background-color': theme === 'dark' ? '#4338ca' : '#6366f1',
                    '--ag-border-color': theme === 'dark' ? '#4b5563' : '#e5e7eb',
                } as React.CSSProperties}
            >
                <style>{`
                    .numeric-header {
                        background-color: ${theme === 'dark' ? '#065f46' : '#d1fae5'} !important;
                    }
                    .mixed-header {
                        background-color: ${theme === 'dark' ? '#7c2d12' : '#fed7aa'} !important;
                    }
                    .target-header {
                        background-color: ${theme === 'dark' ? '#1e3a8a' : '#dbeafe'} !important;
                    }
                    .numeric-cell {
                        text-align: right;
                        font-family: monospace;
                    }
                    .missing-value {
                        background-color: ${theme === 'dark' ? '#7f1d1d' : '#fee2e2'} !important;
                        opacity: 0.7;
                    }
                    .target-column {
                        background-color: ${theme === 'dark' ? '#1e293b' : '#f1f5f9'} !important;
                    }
                `}</style>
                
                <AgGridReact
                    ref={gridRef}
                    rowData={rowData}
                    columnDefs={columnDefs}
                    defaultColDef={defaultColDef}
                    onGridReady={onGridReady}
                    onCellValueChanged={onCellValueChanged}
                    {...gridOptions}
                />
            </div>
        </div>
    );
};