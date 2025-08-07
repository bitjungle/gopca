// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

import React from 'react';

interface CorrelationMatrixProps {
    correlations: { [key: string]: { [key: string]: number } };
}

export const CorrelationMatrix: React.FC<CorrelationMatrixProps> = ({ correlations }) => {
    const columns = Object.keys(correlations);
    
    if (columns.length === 0) {
        return (
            <div className="text-center py-8 text-gray-500 dark:text-gray-400">
                No correlation data available
            </div>
        );
    }

    const getColorForCorrelation = (value: number) => {
        const absValue = Math.abs(value);
        if (absValue > 0.8) return value > 0 ? 'bg-red-500' : 'bg-blue-500';
        if (absValue > 0.6) return value > 0 ? 'bg-red-400' : 'bg-blue-400';
        if (absValue > 0.4) return value > 0 ? 'bg-red-300' : 'bg-blue-300';
        if (absValue > 0.2) return value > 0 ? 'bg-red-200' : 'bg-blue-200';
        return 'bg-gray-200 dark:bg-gray-700';
    };

    const getTextColorForCorrelation = (value: number) => {
        const absValue = Math.abs(value);
        return absValue > 0.4 ? 'text-white' : 'text-gray-700 dark:text-gray-300';
    };

    return (
        <div className="overflow-x-auto">
            <table className="min-w-full">
                <thead>
                    <tr>
                        <th className="p-2 text-xs text-gray-600 dark:text-gray-400"></th>
                        {columns.map((col) => (
                            <th key={col} className="p-2 text-xs text-gray-600 dark:text-gray-400 text-center">
                                <div className="transform -rotate-45 origin-center whitespace-nowrap">
                                    {col.length > 10 ? col.substring(0, 10) + '...' : col}
                                </div>
                            </th>
                        ))}
                    </tr>
                </thead>
                <tbody>
                    {columns.map((row) => (
                        <tr key={row}>
                            <td className="p-2 text-xs text-gray-600 dark:text-gray-400 font-medium">
                                {row.length > 15 ? row.substring(0, 15) + '...' : row}
                            </td>
                            {columns.map((col) => {
                                const value = correlations[row]?.[col] || 0;
                                return (
                                    <td key={col} className="p-1">
                                        <div
                                            className={`w-10 h-10 flex items-center justify-center rounded text-xs font-medium ${getColorForCorrelation(value)} ${getTextColorForCorrelation(value)}`}
                                            title={`${row} vs ${col}: ${value.toFixed(3)}`}
                                        >
                                            {value.toFixed(2)}
                                        </div>
                                    </td>
                                );
                            })}
                        </tr>
                    ))}
                </tbody>
            </table>
            
            {/* Legend */}
            <div className="mt-4 flex items-center justify-center gap-4 text-xs">
                <div className="flex items-center gap-2">
                    <div className="w-4 h-4 bg-blue-500 rounded"></div>
                    <span className="text-gray-600 dark:text-gray-400">Strong negative</span>
                </div>
                <div className="flex items-center gap-2">
                    <div className="w-4 h-4 bg-gray-200 dark:bg-gray-700 rounded"></div>
                    <span className="text-gray-600 dark:text-gray-400">Weak</span>
                </div>
                <div className="flex items-center gap-2">
                    <div className="w-4 h-4 bg-red-500 rounded"></div>
                    <span className="text-gray-600 dark:text-gray-400">Strong positive</span>
                </div>
            </div>
        </div>
    );
};