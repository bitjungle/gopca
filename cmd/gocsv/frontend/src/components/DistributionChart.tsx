// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

import React from 'react';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';
import { main } from '../../wailsjs/go/models';

interface DistributionChartProps {
    distribution: main.DistributionInfo;
    columnName: string;
}

export const DistributionChart: React.FC<DistributionChartProps> = ({ distribution, columnName }) => {
    if (!distribution.histogram || distribution.histogram.length === 0) {
        return (
            <div className="text-center py-8 text-gray-500 dark:text-gray-400">
                No distribution data available
            </div>
        );
    }

    // Prepare data for Recharts
    const chartData = distribution.histogram.map((bin, index) => ({
        bin: `${bin.min.toFixed(2)}-${bin.max.toFixed(2)}`,
        binIndex: index,
        count: bin.count,
        min: bin.min,
        max: bin.max,
    }));

    // Custom tooltip
    const CustomTooltip = ({ active, payload }: any) => {
        if (active && payload && payload.length) {
            const data = payload[0].payload;
            return (
                <div className="bg-white dark:bg-gray-800 p-2 border border-gray-200 dark:border-gray-700 rounded shadow-lg">
                    <p className="text-sm font-medium text-gray-800 dark:text-gray-200">
                        Range: {data.min.toFixed(3)} - {data.max.toFixed(3)}
                    </p>
                    <p className="text-sm text-gray-600 dark:text-gray-400">
                        Count: {data.count}
                    </p>
                </div>
            );
        }
        return null;
    };

    return (
        <div className="w-full h-64">
            <ResponsiveContainer width="100%" height="100%">
                <BarChart 
                    data={chartData} 
                    margin={{ top: 10, right: 10, left: 10, bottom: 40 }}
                >
                    <CartesianGrid 
                        strokeDasharray="3 3" 
                        stroke="#374151" 
                        opacity={0.3}
                    />
                    <XAxis 
                        dataKey="binIndex"
                        tick={{ fontSize: 12, fill: '#9CA3AF' }}
                        tickFormatter={(value) => {
                            // Show only every other tick for readability
                            return value % 2 === 0 ? value : '';
                        }}
                        label={{ 
                            value: 'Bins', 
                            position: 'insideBottom', 
                            offset: -20,
                            style: { fill: '#9CA3AF', fontSize: 12 }
                        }}
                    />
                    <YAxis 
                        tick={{ fontSize: 12, fill: '#9CA3AF' }}
                        label={{ 
                            value: 'Frequency', 
                            angle: -90, 
                            position: 'insideLeft',
                            style: { fill: '#9CA3AF', fontSize: 12 }
                        }}
                    />
                    <Tooltip 
                        content={<CustomTooltip />}
                        cursor={{ fill: '#374151', opacity: 0.1 }}
                    />
                    <Bar 
                        dataKey="count" 
                        fill="#3B82F6"
                        radius={[4, 4, 0, 0]}
                    />
                </BarChart>
            </ResponsiveContainer>
        </div>
    );
};