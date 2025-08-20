// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

import React, { useMemo } from 'react';
import { PlotlyBarChart, useTheme } from '@gopca/ui-components';
import { main } from '../../wailsjs/go/models';

interface PlotlyDistributionChartProps {
    distribution: main.DistributionInfo;
    columnName: string;
}

export const PlotlyDistributionChart: React.FC<PlotlyDistributionChartProps> = ({ 
    distribution, 
    columnName 
}) => {
    const { theme } = useTheme();
    
    // Transform histogram data for PlotlyBarChart
    const chartData = useMemo(() => {
        if (!distribution.histogram || distribution.histogram.length === 0) {
            return [];
        }
        
        return distribution.histogram.map((bin, index) => ({
            x: index, // x as numeric index
            y: bin.count, // y is required by ChartDataPoint interface
            binIndex: index.toString(), // String version for display
            binLabel: `${bin.min.toFixed(2)}-${bin.max.toFixed(2)}`,
            count: bin.count,
            min: bin.min,
            max: bin.max
        }));
    }, [distribution.histogram]);
    
    if (chartData.length === 0) {
        return (
            <div className="text-center py-8 text-gray-500 dark:text-gray-400">
                No distribution data available
            </div>
        );
    }
    
    return (
        <div className="w-full h-64">
            <PlotlyBarChart
                data={chartData}
                dataKey="count"
                xDataKey="binIndex"
                xLabel="Bins"
                yLabel="Frequency"
                margin={{ top: 10, right: 10, bottom: 60, left: 50 }}
                height={256} // h-64 = 16rem = 256px
                fill="#3B82F6"
                showGrid={true}
            />
        </div>
    );
};