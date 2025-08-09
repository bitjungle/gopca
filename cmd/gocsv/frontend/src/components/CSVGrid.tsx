// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

import React, { useCallback, useMemo, useRef, useState, useEffect, forwardRef, useImperativeHandle } from 'react';
import { AgGridReact } from 'ag-grid-react';
import { ColDef, GridReadyEvent, CellValueChangedEvent, GridApi, ColumnApi, CellClickedEvent, RowClickedEvent, ColumnResizedEvent } from 'ag-grid-community';
import 'ag-grid-community/styles/ag-grid.css';
import 'ag-grid-community/styles/ag-theme-quartz.css';
import { useTheme } from '@gopca/ui-components';
import { ExecuteDeleteRows, ExecuteDeleteColumns, ExecuteInsertRow, ExecuteInsertColumn, ExecuteToggleTargetColumn, ExecuteHeaderEdit, ExecuteDuplicateRows } from '../../wailsjs/go/main/App';
import { RenameDialog } from './RenameDialog';
import { ConfirmDialog } from './ConfirmDialog';
import { TargetColumnIcon, CategoryColumnIcon, TargetColumnMenuIcon } from './ColumnIcons';

interface CSVGridProps {
    data: string[][];
    headers: string[];
    rowNames?: string[];
    fileData: any; // The full FileData object for operations
    onDataChange?: (rowIndex: number, colIndex: number, newValue: string) => void;
    onHeaderChange?: (colIndex: number, newHeader: string) => void;
    onRowNameChange?: (rowIndex: number, newRowName: string) => void;
    onRefresh?: (updatedData?: any) => void; // Callback to refresh data after operations
}

// Context menu component
interface ContextMenuItem {
    label?: string;
    action?: () => void;
    icon?: string | React.ReactNode;
    separator?: boolean;
}

interface ContextMenuProps {
    x: number;
    y: number;
    items: ContextMenuItem[];
    onClose: () => void;
}

const ContextMenu: React.FC<ContextMenuProps> = ({ x, y, items, onClose }) => {
    const menuRef = useRef<HTMLDivElement>(null);
    
    useEffect(() => {
        const handleClickOutside = (event: MouseEvent) => {
            if (menuRef.current && !menuRef.current.contains(event.target as Node)) {
                onClose();
            }
        };
        
        document.addEventListener('mousedown', handleClickOutside);
        return () => document.removeEventListener('mousedown', handleClickOutside);
    }, [onClose]);
    
    return (
        <div
            ref={menuRef}
            className="fixed z-50 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg shadow-lg py-1 min-w-[200px]"
            style={{ left: x, top: y }}
        >
            {items.map((item, index) => {
                if (item.separator) {
                    return <div key={index} className="border-t border-gray-200 dark:border-gray-700 my-1" />;
                }
                return (
                    <button
                        key={index}
                        onClick={() => {
                            item.action?.();
                            onClose();
                        }}
                        className="w-full px-4 py-2 text-left text-sm hover:bg-gray-100 dark:hover:bg-gray-700 flex items-center gap-2"
                    >
                        {item.icon && (
                            typeof item.icon === 'string' ? 
                                <span dangerouslySetInnerHTML={{ __html: item.icon }} /> : 
                                item.icon
                        )}
                        <span>{item.label}</span>
                    </button>
                );
            })}
        </div>
    );
};

// Custom header component for AG-Grid to display icons
const CustomHeader = (props: any) => {
    const { displayName, isTargetColumn, isCategoricalColumn } = props;
    
    return (
        <div className="ag-header-cell-label" style={{ display: 'flex', alignItems: 'center' }}>
            <span className="ag-header-cell-text">{displayName}</span>
            {isTargetColumn && <TargetColumnIcon />}
            {isCategoricalColumn && <CategoryColumnIcon />}
        </div>
    );
};

export const CSVGrid = forwardRef<any, CSVGridProps>(({ 
    data, 
    headers,
    rowNames,
    fileData,
    onDataChange,
    onHeaderChange,
    onRowNameChange,
    onRefresh
}, ref) => {
    // Validate inputs
    if (!data || !headers || data.length === 0 || headers.length === 0) {
        return <div className="w-full h-full flex items-center justify-center text-gray-500">No data to display</div>;
    }
    const gridRef = useRef<AgGridReact>(null);
    const [gridApi, setGridApi] = useState<GridApi | null>(null);
    const [columnApi, setColumnApi] = useState<ColumnApi | null>(null);
    const { theme } = useTheme();
    const [hasUserResized, setHasUserResized] = useState(false);
    
    // Context menu state
    const [contextMenu, setContextMenu] = useState<{
        x: number;
        y: number;
        items: ContextMenuItem[];
    } | null>(null);
    
    // Rename dialog state
    const [renameDialog, setRenameDialog] = useState<{
        isOpen: boolean;
        colIndex: number;
        currentName: string;
    }>({ isOpen: false, colIndex: -1, currentName: '' });
    
    // Confirm dialog state
    const [confirmDialog, setConfirmDialog] = useState<{
        isOpen: boolean;
        title: string;
        message: string;
        onConfirm: () => void;
    }>({ isOpen: false, title: '', message: '', onConfirm: () => {} });
    
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
    
    // Declare context menu handlers early
    const handleHeaderContextMenu = useCallback((event: React.MouseEvent, colIndex: number) => {
        event.preventDefault();
        
        const header = headers[colIndex];
        const isTargetColumn = header.toLowerCase().endsWith('#target') || 
                              header.toLowerCase().endsWith('# target');
        
        const items: ContextMenuItem[] = [
            {
                label: isTargetColumn ? 'Remove Target Flag' : 'Mark as Target Column',
                action: async () => {
                    if (fileData) {
                        try {
                            const updatedData = await ExecuteToggleTargetColumn(fileData, colIndex);
                            onRefresh?.(updatedData);
                        } catch (error) {
                            console.error('Error toggling target column:', error);
                        }
                    }
                },
                icon: <TargetColumnMenuIcon />
            },
            {
                label: 'Rename Column',
                action: () => {
                    setRenameDialog({
                        isOpen: true,
                        colIndex: colIndex,
                        currentName: header
                    });
                },
                icon: '✏️'
            },
            {
                label: 'Insert Column Before',
                action: async () => {
                    if (fileData) {
                        const updatedData = await ExecuteInsertColumn(fileData, colIndex, '');
                        onRefresh?.(updatedData);
                    }
                },
                icon: '⬅️'
            },
            {
                label: 'Insert Column After',
                action: async () => {
                    if (fileData) {
                        const updatedData = await ExecuteInsertColumn(fileData, colIndex + 1, '');
                        onRefresh?.(updatedData);
                    }
                },
                icon: '➡️'
            },
            { separator: true },
            {
                label: 'Delete Column',
                action: () => {
                    if (fileData) {
                        setConfirmDialog({
                            isOpen: true,
                            title: 'Delete Column',
                            message: `Are you sure you want to delete column '${header}'?`,
                            onConfirm: async () => {
                                try {
                                    const updatedData = await ExecuteDeleteColumns(fileData, [colIndex]);
                                    onRefresh?.(updatedData);
                                } catch (error) {
                                    console.error('Error deleting column:', error);
                                }
                            }
                        });
                    }
                },
                icon: '🗑️'
            }
        ];
        
        setContextMenu({ x: event.clientX, y: event.clientY, items });
    }, [fileData, headers, onRefresh]);
    
    const handleRowContextMenu = useCallback((event: React.MouseEvent, rowIndex: number) => {
        event.preventDefault();
        
        const items: ContextMenuItem[] = [
            {
                label: 'Insert Row Above',
                action: async () => {
                    if (fileData) {
                        const updatedData = await ExecuteInsertRow(fileData, rowIndex);
                        onRefresh?.(updatedData);
                    }
                },
                icon: '⬆️'
            },
            {
                label: 'Insert Row Below',
                action: async () => {
                    if (fileData) {
                        const updatedData = await ExecuteInsertRow(fileData, rowIndex + 1);
                        onRefresh?.(updatedData);
                    }
                },
                icon: '⬇️'
            },
            { separator: true },
            {
                label: 'Duplicate Row',
                action: async () => {
                    if (fileData) {
                        const selectedRows = gridApi?.getSelectedRows() || [];
                        const rowIndices = selectedRows.length > 0 
                            ? selectedRows.map(row => row.id)
                            : [rowIndex];
                        
                        const updatedData = await ExecuteDuplicateRows(fileData, rowIndices);
                        onRefresh?.(updatedData);
                    }
                },
                icon: '📋'
            },
            {
                label: 'Delete Row',
                action: () => {
                    if (fileData) {
                        const selectedRows = gridApi?.getSelectedRows() || [];
                        const rowIndices = selectedRows.length > 0 
                            ? selectedRows.map(row => row.id)
                            : [rowIndex];
                        
                        const confirmMsg = rowIndices.length > 1 
                            ? `Are you sure you want to delete ${rowIndices.length} rows?`
                            : 'Are you sure you want to delete this row?';
                        
                        setConfirmDialog({
                            isOpen: true,
                            title: 'Delete Row',
                            message: confirmMsg,
                            onConfirm: async () => {
                                try {
                                    const updatedData = await ExecuteDeleteRows(fileData, rowIndices);
                                    onRefresh?.(updatedData);
                                } catch (error) {
                                    console.error('Error deleting rows:', error);
                                }
                            }
                        });
                    }
                },
                icon: '🗑️'
            }
        ];
        
        setContextMenu({ x: event.clientX, y: event.clientY, items });
    }, [fileData, gridApi, onRefresh]);
    
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
                maxWidth: 300,
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
            
            // Check if column is categorical (if fileData has categoricalColumns info)
            const isCategoricalColumn = fileData?.categoricalColumns && 
                                      Object.keys(fileData.categoricalColumns).includes(header);
            
            cols.push({
                field: `col${index}`,
                headerName: header,
                headerComponent: CustomHeader,
                headerComponentParams: {
                    displayName: header,
                    isTargetColumn,
                    isCategoricalColumn,
                    colIndex: index,
                    onContextMenu: handleHeaderContextMenu
                },
                editable: true,
                sortable: true,
                filter: true,
                resizable: true,
                minWidth: 80,
                maxWidth: 400,
                cellClass: (params) => {
                    const classes = [];
                    if (colType === 'numeric') classes.push('numeric-cell');
                    if (colType === 'mixed') classes.push('mixed-cell');
                    if (isTargetColumn) classes.push('target-column');
                    
                    // Check for missing values using same logic as backend
                    const value = params.value?.toString().trim() || '';
                    const lowerValue = value.toLowerCase();
                    const isMissing = !value || 
                        ['na', 'n/a', 'nan', 'null', 'none', 'missing', '-', '?'].includes(lowerValue);
                    
                    if (isMissing) classes.push('missing-value');
                    return classes.join(' ');
                },
                headerClass: () => {
                    const classes = [];
                    if (colType === 'numeric') classes.push('numeric-header');
                    if (colType === 'mixed') classes.push('mixed-header');
                    if (isTargetColumn) classes.push('target-header');
                    return classes.join(' ');
                },
                headerTooltip: isTargetColumn ? 'Target column - excluded from PCA (right-click to toggle)' :
                              isCategoricalColumn ? 'Categorical/grouping column' :
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
    }, [headers, detectColumnType, rowNames, theme, handleHeaderContextMenu]);
    
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
        
        // Auto-size columns based on content with a small delay to ensure data is loaded
        setTimeout(() => {
            params.columnApi.autoSizeAllColumns(false);
        }, 100);
    }, []);
    
    // Handle cell right-click
    const onCellContextMenu = useCallback((event: any) => {
        handleRowContextMenu(event.event, event.rowIndex);
    }, [handleRowContextMenu]);
    
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
        minWidth: 80,
        maxWidth: 400,
        editable: true,
        sortable: true,
        filter: true,
        resizable: true,
    }), []);
    
    
    // Handle keyboard shortcuts
    const handleKeyDown = useCallback((event: KeyboardEvent) => {
        if (!gridApi || !fileData) return;
        
        // Check if we're editing a cell
        const editingCells = gridApi.getEditingCells();
        if (editingCells && editingCells.length > 0) return;
        
        // Delete key - delete selected rows
        if (event.key === 'Delete' || event.key === 'Backspace') {
            const selectedRows = gridApi.getSelectedRows();
            if (selectedRows.length > 0) {
                event.preventDefault();
                const rowIndices = selectedRows.map(row => row.id);
                const confirmMsg = rowIndices.length > 1 
                    ? `Are you sure you want to delete ${rowIndices.length} rows?`
                    : 'Are you sure you want to delete this row?';
                
                setConfirmDialog({
                    isOpen: true,
                    title: 'Delete Row',
                    message: confirmMsg,
                    onConfirm: async () => {
                        try {
                            const updatedData = await ExecuteDeleteRows(fileData, rowIndices);
                            onRefresh?.(updatedData);
                        } catch (error) {
                            console.error('Error deleting rows:', error);
                        }
                    }
                });
            }
        }
        
        // Ctrl/Cmd+D - duplicate selected rows
        if ((event.ctrlKey || event.metaKey) && event.key === 'd') {
            event.preventDefault();
            const selectedRows = gridApi.getSelectedRows();
            if (selectedRows.length > 0) {
                const rowIndices = selectedRows.map(row => row.id);
                ExecuteDuplicateRows(fileData, rowIndices).then((updatedData) => {
                    onRefresh?.(updatedData);
                });
            }
        }
    }, [gridApi, fileData, onRefresh, setConfirmDialog]);
    
    useEffect(() => {
        window.addEventListener('keydown', handleKeyDown);
        return () => {
            window.removeEventListener('keydown', handleKeyDown);
        };
    }, [handleKeyDown]);
    
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
        rowMultiSelectWithClick: true,
        
        // Suppress default context menu
        suppressContextMenu: true,
        
        // Pagination for very large datasets
        pagination: data.length > 10000,
        paginationPageSize: 1000,
        paginationPageSizeSelector: [100, 500, 1000, 5000],
        
        // Other options
        enableCellTextSelection: true,
        ensureDomOrder: true,
    }), [data.length]);
    
    // Auto-size columns on window resize only if user hasn't manually resized
    useEffect(() => {
        const handleResize = () => {
            if (gridApi && columnApi && !hasUserResized) {
                // Maintain content-based sizing on window resize
                columnApi.autoSizeAllColumns(false);
            }
        };
        
        window.addEventListener('resize', handleResize);
        return () => window.removeEventListener('resize', handleResize);
    }, [gridApi, columnApi, hasUserResized]);
    
    // Expose auto-size function for external use
    useImperativeHandle(ref, () => ({
        autoSizeColumns: () => {
            if (columnApi) {
                columnApi.autoSizeAllColumns(false);
                setHasUserResized(false);
            }
        }
    }), [columnApi]);
    
    // Add header right-click handling after grid is ready
    useEffect(() => {
        if (!gridApi || !columnApi) return;
        
        // Add event listener to ag-grid header
        const headerContainer = document.querySelector('.ag-header-container');
        if (!headerContainer) return;
        
        const handleHeaderRightClick = (e: Event) => {
            const event = e as MouseEvent;
            event.preventDefault();
            
            // Find which column was clicked
            const target = event.target as HTMLElement;
            const headerCell = target.closest('.ag-header-cell');
            if (!headerCell) return;
            
            const colId = headerCell.getAttribute('col-id');
            if (colId && colId.startsWith('col')) {
                const colIndex = parseInt(colId.replace('col', ''));
                handleHeaderContextMenu(event as any as React.MouseEvent, colIndex);
            }
        };
        
        headerContainer.addEventListener('contextmenu', handleHeaderRightClick);
        
        return () => {
            headerContainer.removeEventListener('contextmenu', handleHeaderRightClick);
        };
    }, [gridApi, columnApi, handleHeaderContextMenu]);
    
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
                    onCellContextMenu={onCellContextMenu}
                    onColumnResized={(event: ColumnResizedEvent) => {
                        if (event.finished) {
                            setHasUserResized(true);
                        }
                    }}
                    {...gridOptions}
                />
            </div>
            
            {/* Custom context menu */}
            {contextMenu && (
                <ContextMenu
                    x={contextMenu.x}
                    y={contextMenu.y}
                    items={contextMenu.items}
                    onClose={() => setContextMenu(null)}
                />
            )}
            
            {/* Rename dialog */}
            <RenameDialog
                isOpen={renameDialog.isOpen}
                onClose={() => setRenameDialog({ isOpen: false, colIndex: -1, currentName: '' })}
                onRename={async (newName) => {
                    if (fileData && onHeaderChange) {
                        try {
                            await onHeaderChange(renameDialog.colIndex, newName);
                        } catch (error) {
                            console.error('Error renaming column:', error);
                        }
                    }
                }}
                currentName={renameDialog.currentName}
                title="Rename Column"
            />
            
            {/* Confirm dialog */}
            <ConfirmDialog
                isOpen={confirmDialog.isOpen}
                onClose={() => setConfirmDialog({ ...confirmDialog, isOpen: false })}
                onConfirm={confirmDialog.onConfirm}
                title={confirmDialog.title}
                message={confirmDialog.message}
                confirmText="Delete"
                destructive={true}
            />
        </div>
    );
});

CSVGrid.displayName = 'CSVGrid';