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

import { Dialog } from '@gopca/ui-components';
import React, { useState, useEffect } from 'react';
import { ApplyTransformation, GetTransformableColumns, SuggestCategoryOrder } from '../../wailsjs/go/main/App';
import { main } from '../../wailsjs/go/models';

type FileData = main.FileData;
type TransformationResult = main.TransformationResult;

interface DataTransformDialogProps {
    isOpen: boolean;
    onClose: () => void;
    fileData: FileData;
    onTransformComplete: (data: FileData) => void;
}

// Mirrors the TransformationType constants in cmd/gocsv/transforms.go and the
// Type constants in pkg/transform/types.go. All three are maintained by hand;
// adding a transformation means adding it to each.
type TransformationType = 'log' | 'sqrt' | 'square' | 'standardize' | 'minmax' | 'bin' | 'onehot' | 'ordinal';

interface TransformationInfo {
    type: TransformationType;
    name: string;
    description: string;
    category: 'math' | 'scale' | 'encode';
    requiresNumeric: boolean;
    requiresCategorical: boolean;
    hasOptions?: boolean;
}

const transformations: TransformationInfo[] = [
    {
        type: 'log',
        name: 'Log Transform',
        description: 'Apply natural logarithm (ln) to positive values',
        category: 'math',
        requiresNumeric: true,
        requiresCategorical: false
    },
    {
        type: 'sqrt',
        name: 'Square Root',
        description: 'Apply square root to non-negative values',
        category: 'math',
        requiresNumeric: true,
        requiresCategorical: false
    },
    {
        type: 'square',
        name: 'Square',
        description: 'Square the values (x²)',
        category: 'math',
        requiresNumeric: true,
        requiresCategorical: false
    },
    {
        type: 'standardize',
        name: 'Standardize (Z-score)',
        description: 'Scale to mean=0, std=1',
        category: 'scale',
        requiresNumeric: true,
        requiresCategorical: false
    },
    {
        type: 'minmax',
        name: 'Min-Max Scale',
        description: 'Scale to a specific range',
        category: 'scale',
        requiresNumeric: true,
        requiresCategorical: false,
        hasOptions: true
    },
    {
        type: 'bin',
        name: 'Binning',
        description: 'Convert numeric to categorical bins',
        category: 'encode',
        requiresNumeric: true,
        requiresCategorical: false,
        hasOptions: true
    },
    {
        type: 'onehot',
        name: 'One-Hot Encode',
        description: 'Create binary columns for each category, keeping the original',
        category: 'encode',
        requiresNumeric: false,
        requiresCategorical: true,
        hasOptions: true
    },
    {
        type: 'ordinal',
        name: 'Ordinal Encode',
        description: 'Number the categories in an order you choose (label encoding)',
        category: 'encode',
        requiresNumeric: false,
        requiresCategorical: true,
        hasOptions: true
    }
];

export const DataTransformDialog: React.FC<DataTransformDialogProps> = ({
    isOpen,
    onClose,
    fileData,
    onTransformComplete
}) => {
    const [selectedTransform, setSelectedTransform] = useState<TransformationType>('log');
    const [selectedColumns, setSelectedColumns] = useState<string[]>([]);
    const [availableColumns, setAvailableColumns] = useState<string[]>([]);
    const [isLoading, setIsLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [result, setResult] = useState<TransformationResult | null>(null);

    // Transform-specific options
    const [binCount, setBinCount] = useState(5);
    const [minValue, setMinValue] = useState(0);
    const [maxValue, setMaxValue] = useState(1);
    // One-hot encoding used to discard the source column unconditionally.
    // Keeping it is the default: GoPCA colours plots by categorical columns,
    // so dropping e.g. "species" silently costs that.
    const [keepOriginal, setKeepOriginal] = useState(true);
    // Category order per column, for ordinal encoding. Seeded from the backend
    // suggestion when a column is picked, then reordered by the user.
    const [categoryOrder, setCategoryOrder] = useState<Record<string, string[]>>({});

    // Above this many distinct values, hand-ordering is not a usable control --
    // and a column with that many categories is almost certainly nominal, which
    // ordinal encoding is the wrong tool for. Alphabetical order is used, and
    // the dialog says so rather than silently offering nothing.
    const MAX_ORDERABLE_CATEGORIES = 50;

    // Load available columns when dialog opens or transform type changes
    useEffect(() => {
        if (isOpen && fileData) {
            loadAvailableColumns();
        }
    }, [isOpen, selectedTransform, fileData]);

    // Reset the transient dialog state whenever it is reopened.
    //
    // Closing renders null but does not unmount, so every useState above
    // survives. That made two choices sticky in a way nobody asked for:
    //
    //   keepOriginal  unticking it once carried the destructive choice into
    //                 every later transformation, which defeats the whole point
    //                 of the default being on.
    //   result        the footer shows Apply while there is no result and Close
    //                 once there is one, so a leftover result replaced the Apply
    //                 button and made the second transformation of a session
    //                 impossible to start.
    useEffect(() => {
        if (isOpen) {
            setKeepOriginal(true);
            setResult(null);
            setError(null);
        }
    }, [isOpen]);

    // Fetch a suggested order for each newly selected column. Existing entries
    // are kept, so a user who has reordered a column does not lose that work by
    // selecting a second one.
    useEffect(() => {
        if (selectedTransform !== 'ordinal' || !fileData) {
            return;
        }
        let cancelled = false;

        const load = async () => {
            const additions: Record<string, string[]> = {};
            for (const column of selectedColumns) {
                if (categoryOrder[column]) {
                    continue;
                }
                try {
                    additions[column] = await SuggestCategoryOrder(fileData, column);
                } catch (err) {
                    console.error(`Error suggesting order for ${column}:`, err);
                    additions[column] = [];
                }
            }
            if (!cancelled && Object.keys(additions).length > 0) {
                // current wins: a suggestion must never overwrite an order the
                // user has already arranged. Newly selected columns are absent
                // from current, so they still come through.
                setCategoryOrder((current) => ({ ...additions, ...current }));
            }
        };
        void load();

        return () => {
            cancelled = true;
        };
    }, [selectedTransform, selectedColumns, fileData]);

    // Move a category one place up or down in its column's order.
    const moveCategory = (column: string, index: number, delta: number) => {
        setCategoryOrder((current) => {
            const values = current[column];
            const target = index + delta;
            if (!values || target < 0 || target >= values.length) {
                return current;
            }
            const reordered = [...values];
            [reordered[index], reordered[target]] = [reordered[target], reordered[index]];
            return { ...current, [column]: reordered };
        });
    };

    // A previously arranged order belongs to the file it was arranged for.
    // Carrying it into another file would silently apply one column's scale to
    // a same-named column holding different categories.
    useEffect(() => {
        setCategoryOrder({});
    }, [fileData]);

    const loadAvailableColumns = async () => {
        try {
            const columns = await GetTransformableColumns(fileData, selectedTransform);
            setAvailableColumns(columns);
            setSelectedColumns([]);
        } catch (err) {
            console.error('Error loading columns:', err);
            setAvailableColumns([]);
        }
    };

    const handleApplyTransform = async () => {
        if (selectedColumns.length === 0) {
            setError('Please select at least one column');
            return;
        }

        setIsLoading(true);
        setError(null);
        setResult(null);

        try {
            const options = {
                type: selectedTransform,
                columns: selectedColumns,
                binCount: selectedTransform === 'bin' ? binCount : undefined,
                minValue: selectedTransform === 'minmax' ? minValue : undefined,
                maxValue: selectedTransform === 'minmax' ? maxValue : undefined,
                removeOriginal:
                    selectedTransform === 'onehot' || selectedTransform === 'ordinal'
                        ? !keepOriginal
                        : undefined,
                categoryOrder: selectedTransform === 'ordinal' ? categoryOrder : undefined
            };

            const transformResult = await ApplyTransformation(fileData, options);

            if (transformResult.success && transformResult.data) {
                setResult(transformResult);
                onTransformComplete(transformResult.data);
            } else {
                setError('Transformation failed');
            }
        } catch (err) {
            setError(`Error applying transformation: ${err}`);
        } finally {
            setIsLoading(false);
        }
    };

    const toggleColumn = (column: string) => {
        if (selectedColumns.includes(column)) {
            setSelectedColumns(selectedColumns.filter(c => c !== column));
        } else {
            setSelectedColumns([...selectedColumns, column]);
        }
    };

    const selectAllColumns = () => {
        setSelectedColumns([...availableColumns]);
    };

    const deselectAllColumns = () => {
        setSelectedColumns([]);
    };

    // Dialog owns the isOpen check; unmounting from here would skip its
    // focus-restore cleanup.

    const currentTransform = transformations.find(t => t.type === selectedTransform);

    return (
        <Dialog
            isOpen={isOpen}
            onClose={onClose}
            width="w-full max-w-2xl"
            padded={false}
            ariaLabelledBy="data-transform-title"
            className="max-h-[90vh] overflow-y-auto text-left"
        >
                    {/* Header */}
                    <div className="px-6 py-4 border-b border-gray-200 dark:border-gray-700">
                        <div className="flex items-center justify-between">
                            <h2
                                id="data-transform-title"
                                className="text-lg font-semibold text-gray-900 dark:text-white"
                            >
                                Data Transformations
                            </h2>
                            <button
                                onClick={onClose}
                                className="text-gray-400 hover:text-gray-500 dark:hover:text-gray-300"
                            >
                                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                                </svg>
                            </button>
                        </div>
                    </div>

                    {/* Content */}
                    <div className="px-6 py-4">
                        {/* Transformation type selection */}
                        <div className="mb-6">
                            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                                Transformation Type
                            </label>
                            <div className="grid grid-cols-1 gap-2">
                                {['math', 'scale', 'encode'].map(category => (
                                    <div key={category}>
                                        <div className="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase mb-1">
                                            {category === 'math' ? 'Mathematical' : category === 'scale' ? 'Scaling' : 'Encoding'}
                                        </div>
                                        <div className="space-y-1">
                                            {transformations
                                                .filter(t => t.category === category)
                                                .map(transform => (
                                                    <button
                                                        key={transform.type}
                                                        onClick={() => setSelectedTransform(transform.type)}
                                                        className={`w-full text-left px-3 py-2 rounded-lg border transition-colors ${
                                                            selectedTransform === transform.type
                                                                ? 'border-blue-500 bg-blue-50 dark:bg-blue-900/20 text-blue-700 dark:text-blue-300'
                                                                : 'border-gray-200 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-700/50'
                                                        }`}
                                                    >
                                                        <div className="font-medium text-sm">{transform.name}</div>
                                                        <div className="text-xs text-gray-600 dark:text-gray-400">
                                                            {transform.description}
                                                        </div>
                                                    </button>
                                                ))}
                                        </div>
                                    </div>
                                ))}
                            </div>
                        </div>

                        {/* Transform-specific options */}
                        {currentTransform?.hasOptions && (
                            <div className="mb-6 p-4 bg-gray-50 dark:bg-gray-700/50 rounded-lg">
                                {selectedTransform === 'minmax' && (
                                    <div className="grid grid-cols-2 gap-4">
                                        <div>
                                            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                                                Min Value
                                            </label>
                                            <input
                                                type="number"
                                                value={minValue}
                                                onChange={(e) => setMinValue(parseFloat(e.target.value) || 0)}
                                                step="0.1"
                                                className="w-full px-3 py-2 text-sm border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
                                            />
                                        </div>
                                        <div>
                                            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                                                Max Value
                                            </label>
                                            <input
                                                type="number"
                                                value={maxValue}
                                                onChange={(e) => setMaxValue(parseFloat(e.target.value) || 1)}
                                                step="0.1"
                                                className="w-full px-3 py-2 text-sm border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
                                            />
                                        </div>
                                    </div>
                                )}
                                {(selectedTransform === 'onehot' || selectedTransform === 'ordinal') && (
                                    <div>
                                        <label className="flex items-start gap-2 cursor-pointer">
                                            <input
                                                type="checkbox"
                                                checked={keepOriginal}
                                                onChange={(e) => setKeepOriginal(e.target.checked)}
                                                className="mt-0.5 h-4 w-4 rounded border-gray-300 dark:border-gray-600"
                                            />
                                            <span className="text-sm text-gray-700 dark:text-gray-300">
                                                Keep original column
                                                <span className="block text-xs text-gray-500 dark:text-gray-400">
                                                    Unchecking removes the source column once the new
                                                    {selectedTransform === 'ordinal' ? ' code column is' : ' binary columns are'} created.
                                                    Keeping it lets GoPCA still colour plots by this category.
                                                </span>
                                            </span>
                                        </label>
                                    </div>
                                )}
                                {selectedTransform === 'ordinal' && (
                                    <div className="mt-4">
                                        {selectedColumns.length === 0 && (
                                            <p className="text-sm text-gray-500 dark:text-gray-400">
                                                Select a column to set the order of its categories.
                                            </p>
                                        )}
                                        {selectedColumns.map((column) => {
                                            const values = categoryOrder[column] || [];
                                            const tooMany = values.length > MAX_ORDERABLE_CATEGORIES;
                                            return (
                                                <div key={column} className="mb-4 last:mb-0">
                                                    <div className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                                                        Category order for &lsquo;{column}&rsquo;
                                                    </div>
                                                    {tooMany ? (
                                                        <p className="text-xs text-gray-500 dark:text-gray-400">
                                                            {values.length} distinct values — too many to order by hand,
                                                            so codes will be assigned alphabetically. A column with this
                                                            many categories is usually unordered; One-Hot Encode is
                                                            likely the better fit.
                                                        </p>
                                                    ) : (
                                                        <>
                                                            <p className="text-xs text-gray-500 dark:text-gray-400 mb-2">
                                                                Codes are assigned by position. Reorder with the arrows.
                                                            </p>
                                                            <ul className="space-y-1">
                                                                {values.map((value, index) => (
                                                                    <li
                                                                        key={value}
                                                                        className="flex items-center gap-2 text-sm text-gray-800 dark:text-gray-200"
                                                                    >
                                                                        <span className="w-10 shrink-0 text-right font-mono text-xs text-gray-500 dark:text-gray-400">
                                                                            = {index}
                                                                        </span>
                                                                        <span className="flex-1 truncate">{value}</span>
                                                                        <button
                                                                            type="button"
                                                                            onClick={() => moveCategory(column, index, -1)}
                                                                            disabled={index === 0}
                                                                            aria-label={`Move ${value} up`}
                                                                            className="px-2 py-0.5 text-xs rounded border border-gray-300 dark:border-gray-600 disabled:opacity-30"
                                                                        >
                                                                            ↑
                                                                        </button>
                                                                        <button
                                                                            type="button"
                                                                            onClick={() => moveCategory(column, index, 1)}
                                                                            disabled={index === values.length - 1}
                                                                            aria-label={`Move ${value} down`}
                                                                            className="px-2 py-0.5 text-xs rounded border border-gray-300 dark:border-gray-600 disabled:opacity-30"
                                                                        >
                                                                            ↓
                                                                        </button>
                                                                    </li>
                                                                ))}
                                                            </ul>
                                                        </>
                                                    )}
                                                </div>
                                            );
                                        })}
                                        <p className="mt-3 text-xs text-gray-500 dark:text-gray-400">
                                            Only meaningful when the categories have a real order. For unordered
                                            categories such as species or site, use One-Hot Encode — numbering them
                                            would tell PCA that the gaps between them are real.
                                        </p>
                                    </div>
                                )}
                                {selectedTransform === 'bin' && (
                                    <div>
                                        <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                                            Number of Bins
                                        </label>
                                        <input
                                            type="number"
                                            value={binCount}
                                            onChange={(e) => setBinCount(Math.max(2, parseInt(e.target.value) || 5))}
                                            min="2"
                                            max="20"
                                            className="w-full px-3 py-2 text-sm border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
                                        />
                                    </div>
                                )}
                            </div>
                        )}

                        {/* Column selection */}
                        <div className="mb-6">
                            <div className="flex items-center justify-between mb-2">
                                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">
                                    Select Columns
                                </label>
                                <div className="flex gap-2">
                                    <button
                                        onClick={selectAllColumns}
                                        className="text-xs text-blue-600 dark:text-blue-400 hover:text-blue-700 dark:hover:text-blue-300"
                                    >
                                        Select All
                                    </button>
                                    <span className="text-gray-400">|</span>
                                    <button
                                        onClick={deselectAllColumns}
                                        className="text-xs text-blue-600 dark:text-blue-400 hover:text-blue-700 dark:hover:text-blue-300"
                                    >
                                        Deselect All
                                    </button>
                                </div>
                            </div>

                            {availableColumns.length === 0 ? (
                                <div className="text-sm text-gray-500 dark:text-gray-400 text-center py-4">
                                    No columns available for this transformation
                                </div>
                            ) : (
                                <div className="border border-gray-200 dark:border-gray-700 rounded-lg max-h-48 overflow-y-auto">
                                    {availableColumns.map(column => (
                                        <label
                                            key={column}
                                            className="flex items-center gap-3 px-3 py-2 hover:bg-gray-50 dark:hover:bg-gray-700/50 cursor-pointer border-b border-gray-100 dark:border-gray-800 last:border-b-0"
                                        >
                                            <input
                                                type="checkbox"
                                                checked={selectedColumns.includes(column)}
                                                onChange={() => toggleColumn(column)}
                                                className="rounded text-blue-600 focus:ring-blue-500"
                                            />
                                            <span className="text-sm text-gray-700 dark:text-gray-300">
                                                {column}
                                            </span>
                                        </label>
                                    ))}
                                </div>
                            )}
                        </div>

                        {/* Error message */}
                        {error && (
                            <div className="mb-4 p-3 bg-red-50 dark:bg-red-900/20 text-red-700 dark:text-red-300 rounded-lg text-sm">
                                {error}
                            </div>
                        )}

                        {/* Result messages */}
                        {result && result.messages && result.messages.length > 0 && (
                            <div className="mb-4 p-3 bg-blue-50 dark:bg-blue-900/20 text-blue-700 dark:text-blue-300 rounded-lg">
                                <div className="text-sm font-medium mb-1">Transformation Results:</div>
                                <ul className="list-disc list-inside text-sm space-y-1">
                                    {result.messages.map((msg, index) => (
                                        <li key={index}>{msg}</li>
                                    ))}
                                </ul>
                            </div>
                        )}
                    </div>

                    {/* Footer */}
                    <div className="px-6 py-4 border-t border-gray-200 dark:border-gray-700">
                        <div className="flex justify-end gap-2">
                            <button
                                onClick={onClose}
                                className="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-600"
                            >
                                {result ? 'Close' : 'Cancel'}
                            </button>
                            {!result && (
                                <button
                                    onClick={handleApplyTransform}
                                    disabled={isLoading || selectedColumns.length === 0}
                                    className="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
                                >
                                    {isLoading ? 'Applying...' : 'Apply Transform'}
                                </button>
                            )}
                        </div>
                    </div>
        </Dialog>
    );
};