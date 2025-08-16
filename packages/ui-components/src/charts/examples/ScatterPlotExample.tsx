// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

/**
 * Example component demonstrating how to use the chart abstraction layer.
 * This shows how visualization components can be migrated to use the abstraction
 * instead of directly importing from Recharts.
 */

import React from 'react';
import { ScatterChart, Cell } from '../index';
import { useChartTheme } from '../../hooks/useChartTheme';

interface ExampleData {
  x: number;
  y: number;
  name: string;
  group: string;
  color: string;
}

interface ScatterPlotExampleProps {
  data: ExampleData[];
  xLabel?: string;
  yLabel?: string;
  showReferenceLines?: boolean;
}

export const ScatterPlotExample: React.FC<ScatterPlotExampleProps> = ({
  data,
  xLabel = 'X Axis',
  yLabel = 'Y Axis',
  showReferenceLines = true,
}) => {
  const chartTheme = useChartTheme();

  // Custom tooltip renderer
  const renderTooltip = ({ active, payload }: any) => {
    if (active && payload && payload.length) {
      const data = payload[0].payload;
      return (
        <div 
          className="p-2 rounded shadow-lg border"
          style={{ 
            backgroundColor: chartTheme.tooltipBackgroundColor,
            borderColor: chartTheme.tooltipBorderColor
          }}
        >
          <p className="font-semibold" style={{ color: chartTheme.tooltipTextColor }}>
            {data.name}
          </p>
          <p style={{ color: chartTheme.tooltipTextColor }}>
            Group: {data.group}
          </p>
          <p style={{ color: chartTheme.tooltipTextColor }}>
            X: {data.x.toFixed(3)}
          </p>
          <p style={{ color: chartTheme.tooltipTextColor }}>
            Y: {data.y.toFixed(3)}
          </p>
        </div>
      );
    }
    return null;
  };

  // Custom dot renderer
  const renderDot = (props: any) => {
    const { cx, cy, payload } = props;
    return (
      <circle
        cx={cx}
        cy={cy}
        r={4}
        fill={payload.color}
        stroke={payload.color}
        strokeWidth={1}
        fillOpacity={0.8}
      />
    );
  };

  return (
    <div className="w-full h-full">
      <ScatterChart
        data={data}
        xLabel={xLabel}
        yLabel={yLabel}
        showReferenceLines={showReferenceLines}
        tooltip={renderTooltip}
        dot={renderDot}
        className="w-full h-full"
      >
        {/* Additional Recharts components can be added as children */}
        {data.map((entry, index) => (
          <Cell key={`cell-${index}`} fill={entry.color} />
        ))}
      </ScatterChart>
    </div>
  );
};

/**
 * Migration Guide:
 * 
 * 1. Replace Recharts imports:
 *    Before: import { ComposedChart, Scatter, ... } from 'recharts';
 *    After:  import { ComposedChart, ScatterChart, ... } from '@gopca/ui-components';
 * 
 * 2. Use abstracted chart components where possible:
 *    - ScatterChart for scatter plots
 *    - BarChart for bar charts
 *    - LineChart for line charts
 *    - ComposedChart for complex compositions
 * 
 * 3. Native Recharts components (Cell, Legend, etc.) are re-exported from
 *    the abstraction layer for compatibility.
 * 
 * 4. The abstraction automatically handles:
 *    - Theme integration
 *    - Provider selection (Recharts, Plotly, D3)
 *    - Common chart configurations
 * 
 * 5. Complex visualizations may need gradual migration:
 *    - Start with simple charts
 *    - Keep complex interactions with Recharts for now
 *    - Migrate fully when alternative providers are ready
 */