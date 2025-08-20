// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

import React, { useState, useEffect } from 'react';
import { ApplyTransformation, GetTransformableColumns } from '../../wailsjs/go/main/App';
import { main } from '../../wailsjs/go/models';

type FileData = main.FileData;
type TransformationResult = main.TransformationResult;

interface DataTransformDialogProps {
    isOpen: boolean;
    onClose: () => void;
    fileData: FileData;
    onTransformComplete: (data: FileData) => void;
}

type TransformationType = 'log' | 'sqrt' | 'square' | 'standardize' | 'minmax' | 'bin' | 'onehot';

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
        description: 'Create binary columns for each category',
        category: 'encode',
        requiresNumeric: false,
        requiresCategorical: true
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

    // Load available columns when dialog opens or transform type changes
    useEffect(() => {
        if (isOpen && fileData) {
            loadAvailableColumns();
        }
    }, [isOpen, selectedTransform, fileData]);

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
                maxValue: selectedTransform === 'minmax' ? maxValue : undefined
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

    if (!isOpen) {
return null;
}

    const currentTransform = transformations.find(t => t.type === selectedTransform);

    return (
        <div className="fixed inset-0 z-50 overflow-y-auto">
            <div className="flex items-center justify-center min-h-screen px-4 pt-4 pb-20 text-center sm:block sm:p-0">
                {/* Background overlay */}
                <div
                    className="fixed inset-0 transition-opacity bg-gray-500 bg-opacity-75 dark:bg-gray-900 dark:bg-opacity-75"
                    onClick={onClose}
                />

                {/* Modal panel */}
                <div className="inline-block w-full max-w-2xl my-8 text-left align-middle transition-all transform bg-white dark:bg-gray-800 shadow-xl rounded-lg">
                    {/* Header */}
                    <div className="px-6 py-4 border-b border-gray-200 dark:border-gray-700">
                        <div className="flex items-center justify-between">
                            <h2 className="text-lg font-semibold text-gray-900 dark:text-white">
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
                </div>
            </div>
        </div>
    );
};