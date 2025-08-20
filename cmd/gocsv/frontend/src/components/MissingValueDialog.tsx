// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

import React, { useState } from 'react';

interface MissingValueDialogProps {
    isOpen: boolean;
    onClose: () => void;
    onFill: (strategy: string, column: string, value?: string) => void;
    columns: string[];
    columnTypes: Record<string, string>;
}

export const MissingValueDialog: React.FC<MissingValueDialogProps> = ({
    isOpen,
    onClose,
    onFill,
    columns,
    columnTypes
}) => {
    const [strategy, setStrategy] = useState('mean');
    const [selectedColumn, setSelectedColumn] = useState('');
    const [customValue, setCustomValue] = useState('');

    if (!isOpen) {
return null;
}

    const handleFill = () => {
        onFill(strategy, selectedColumn, strategy === 'custom' ? customValue : undefined);
        onClose();
    };

    const getAvailableStrategies = () => {
        if (!selectedColumn) {
            return ['mean', 'median', 'mode', 'forward', 'backward', 'custom'];
        }

        const colType = columnTypes[selectedColumn] || 'numeric';

        if (colType === 'numeric') {
            return ['mean', 'median', 'mode', 'forward', 'backward', 'custom'];
        } else {
            return ['mode', 'forward', 'backward', 'custom'];
        }
    };

    const strategies = getAvailableStrategies();

    return (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
            <div className="bg-white dark:bg-gray-800 rounded-lg shadow-xl w-96">
                <div className="flex items-center justify-between p-4 border-b border-gray-200 dark:border-gray-700">
                    <h2 className="text-lg font-semibold text-gray-800 dark:text-gray-200">
                        Fill Missing Values
                    </h2>
                    <button
                        onClick={onClose}
                        className="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
                    >
                        <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M6 18L18 6M6 6l12 12" />
                        </svg>
                    </button>
                </div>

                <div className="p-4 space-y-4">
                    {/* Column Selection */}
                    <div>
                        <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                            Apply to Column
                        </label>
                        <select
                            value={selectedColumn}
                            onChange={(e) => {
                                setSelectedColumn(e.target.value);
                                // Reset strategy if not available for new column type
                                const newType = columnTypes[e.target.value] || 'numeric';
                                if (newType !== 'numeric' && (strategy === 'mean' || strategy === 'median')) {
                                    setStrategy('mode');
                                }
                            }}
                            className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
                        >
                            <option value="">All Columns</option>
                            {columns.map(col => (
                                <option key={col} value={col}>
                                    {col} ({columnTypes[col] || 'numeric'})
                                </option>
                            ))}
                        </select>
                    </div>

                    {/* Strategy Selection */}
                    <div>
                        <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                            Fill Strategy
                        </label>
                        <select
                            value={strategy}
                            onChange={(e) => setStrategy(e.target.value)}
                            className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
                        >
                            {strategies.includes('mean') && <option value="mean">Mean (average)</option>}
                            {strategies.includes('median') && <option value="median">Median (middle value)</option>}
                            <option value="mode">Mode (most frequent)</option>
                            <option value="forward">Forward Fill</option>
                            <option value="backward">Backward Fill</option>
                            <option value="custom">Custom Value</option>
                        </select>
                    </div>

                    {/* Custom Value Input */}
                    {strategy === 'custom' && (
                        <div>
                            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                                Custom Value
                            </label>
                            <input
                                type="text"
                                value={customValue}
                                onChange={(e) => setCustomValue(e.target.value)}
                                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
                                placeholder="Enter value..."
                            />
                        </div>
                    )}

                    {/* Strategy Description */}
                    <div className="bg-gray-50 dark:bg-gray-700 rounded-md p-3 text-sm text-gray-600 dark:text-gray-400">
                        {strategy === 'mean' && "Replace missing values with the column's average."}
                        {strategy === 'median' && "Replace missing values with the column's middle value."}
                        {strategy === 'mode' && 'Replace missing values with the most frequent value.'}
                        {strategy === 'forward' && 'Replace missing values with the previous non-missing value.'}
                        {strategy === 'backward' && 'Replace missing values with the next non-missing value.'}
                        {strategy === 'custom' && 'Replace missing values with a specific value.'}
                    </div>
                </div>

                <div className="flex justify-end gap-2 p-4 border-t border-gray-200 dark:border-gray-700">
                    <button
                        onClick={onClose}
                        className="px-4 py-2 text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-md transition-colors"
                    >
                        Cancel
                    </button>
                    <button
                        onClick={handleFill}
                        disabled={strategy === 'custom' && !customValue}
                        className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                    >
                        Fill Values
                    </button>
                </div>
            </div>
        </div>
    );
};