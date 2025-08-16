// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

import React from 'react';
import {
  ComposedChart as RechartComposedChart,
  Scatter as RechartsScatter,
  Bar as RechartsBar,
  Line as RechartsLine,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  ReferenceLine,
} from 'recharts';
import {
  ScatterChartProps,
  BarChartProps,
  LineChartProps,
  ComposedChartProps,
  ChartAdapter,
} from '../../types';
import { useChartTheme } from '../../../hooks/useChartTheme';

export const RechartsScatterChart: React.FC<ScatterChartProps> = ({
  data,
  domain,
  margin = { top: 20, right: 20, bottom: 60, left: 80 },
  width = '100%',
  height = '100%',
  xDataKey = 'x',
  yDataKey = 'y',
  xLabel,
  yLabel,
  showGrid = true,
  showReferenceLines = true,
  tooltip,
  dot,
  fill = '#3B82F6',
  stroke = '#1E40AF',
  children,
  onMouseMove,
  onMouseDown,
  onMouseUp,
  onMouseLeave,
  className,
}) => {
  const chartTheme = useChartTheme();

  return (
    <div 
      className={className}
      onMouseDown={onMouseDown}
      onMouseMove={onMouseMove}
      onMouseUp={onMouseUp}
      onMouseLeave={onMouseLeave}
    >
      <ResponsiveContainer width={width} height={height}>
        <RechartComposedChart data={data} margin={margin}>
          {showGrid && (
            <CartesianGrid strokeDasharray="3 3" stroke={chartTheme.gridColor} />
          )}
          <XAxis
            type="number"
            dataKey={xDataKey}
            name={xLabel}
            label={xLabel ? { value: xLabel, position: 'insideBottom', offset: -10 } : undefined}
            stroke={chartTheme.axisColor}
            domain={domain?.x || ['dataMin', 'dataMax']}
            axisLine={{ stroke: chartTheme.axisColor }}
            tickLine={{ stroke: chartTheme.axisColor }}
            tickFormatter={(value) => typeof value === 'number' ? value.toFixed(1) : value}
          />
          <YAxis
            type="number"
            dataKey={yDataKey}
            name={yLabel}
            label={yLabel ? { value: yLabel, angle: -90, position: 'insideLeft' } : undefined}
            stroke={chartTheme.axisColor}
            domain={domain?.y || ['dataMin', 'dataMax']}
            axisLine={{ stroke: chartTheme.axisColor }}
            tickLine={{ stroke: chartTheme.axisColor }}
            tickFormatter={(value) => typeof value === 'number' ? value.toFixed(1) : value}
          />
          {showReferenceLines && (
            <>
              <ReferenceLine x={0} stroke={chartTheme.referenceLineColor} strokeWidth={2} />
              <ReferenceLine y={0} stroke={chartTheme.referenceLineColor} strokeWidth={2} />
            </>
          )}
          {tooltip && (
            <Tooltip content={typeof tooltip === 'function' ? tooltip : undefined} />
          )}
          <RechartsScatter
            name="Data"
            fill={fill}
            fillOpacity={0.8}
            strokeWidth={1}
            stroke={stroke}
            shape={typeof dot === 'function' ? dot : undefined}
          />
          {children}
        </RechartComposedChart>
      </ResponsiveContainer>
    </div>
  );
};

export const RechartsBarChart: React.FC<BarChartProps> = ({
  data,
  dataKey,
  xDataKey = 'x',
  margin = { top: 20, right: 20, bottom: 60, left: 80 },
  width = '100%',
  height = '100%',
  xLabel,
  yLabel,
  showGrid = true,
  fill = '#3B82F6',
  children,
  className,
}) => {
  const chartTheme = useChartTheme();

  return (
    <div className={className}>
      <ResponsiveContainer width={width} height={height}>
        <RechartComposedChart data={data} margin={margin}>
          {showGrid && (
            <CartesianGrid strokeDasharray="3 3" stroke={chartTheme.gridColor} />
          )}
          <XAxis
            dataKey={xDataKey}
            stroke={chartTheme.axisColor}
            label={xLabel ? { value: xLabel, position: 'insideBottom', offset: -10 } : undefined}
            axisLine={{ stroke: chartTheme.axisColor }}
            tickLine={{ stroke: chartTheme.axisColor }}
          />
          <YAxis
            stroke={chartTheme.axisColor}
            label={yLabel ? { value: yLabel, angle: -90, position: 'insideLeft' } : undefined}
            axisLine={{ stroke: chartTheme.axisColor }}
            tickLine={{ stroke: chartTheme.axisColor }}
          />
          <Tooltip />
          <RechartsBar dataKey={dataKey} fill={fill} />
          {children}
        </RechartComposedChart>
      </ResponsiveContainer>
    </div>
  );
};

export const RechartsLineChart: React.FC<LineChartProps> = ({
  data,
  dataKey,
  xDataKey = 'x',
  margin = { top: 20, right: 20, bottom: 60, left: 80 },
  width = '100%',
  height = '100%',
  xLabel,
  yLabel,
  showGrid = true,
  stroke = '#3B82F6',
  strokeWidth = 2,
  dot = false,
  children,
  className,
}) => {
  const chartTheme = useChartTheme();

  return (
    <div className={className}>
      <ResponsiveContainer width={width} height={height}>
        <RechartComposedChart data={data} margin={margin}>
          {showGrid && (
            <CartesianGrid strokeDasharray="3 3" stroke={chartTheme.gridColor} />
          )}
          <XAxis
            dataKey={xDataKey}
            stroke={chartTheme.axisColor}
            label={xLabel ? { value: xLabel, position: 'insideBottom', offset: -10 } : undefined}
            axisLine={{ stroke: chartTheme.axisColor }}
            tickLine={{ stroke: chartTheme.axisColor }}
          />
          <YAxis
            stroke={chartTheme.axisColor}
            label={yLabel ? { value: yLabel, angle: -90, position: 'insideLeft' } : undefined}
            axisLine={{ stroke: chartTheme.axisColor }}
            tickLine={{ stroke: chartTheme.axisColor }}
          />
          <Tooltip />
          <RechartsLine
            type="monotone"
            dataKey={dataKey}
            stroke={stroke}
            strokeWidth={strokeWidth}
            dot={typeof dot === 'boolean' ? dot : (typeof dot === 'function' ? dot : false)}
          />
          {children}
        </RechartComposedChart>
      </ResponsiveContainer>
    </div>
  );
};

export const RechartsComposedChart: React.FC<ComposedChartProps> = ({
  data,
  domain,
  margin = { top: 20, right: 20, bottom: 60, left: 80 },
  width = '100%',
  height = '100%',
  xDataKey = 'x',
  xLabel,
  yLabel,
  showGrid = true,
  children,
  onMouseMove,
  onMouseDown,
  onMouseUp,
  onMouseLeave,
  className,
}) => {
  const chartTheme = useChartTheme();

  return (
    <div 
      className={className}
      onMouseDown={onMouseDown}
      onMouseMove={onMouseMove}
      onMouseUp={onMouseUp}
      onMouseLeave={onMouseLeave}
    >
      <ResponsiveContainer width={width} height={height}>
        <RechartComposedChart data={data} margin={margin}>
          {showGrid && (
            <CartesianGrid strokeDasharray="3 3" stroke={chartTheme.gridColor} />
          )}
          <XAxis
            type="number"
            dataKey={xDataKey}
            stroke={chartTheme.axisColor}
            domain={domain?.x || ['dataMin', 'dataMax']}
            label={xLabel ? { value: xLabel, position: 'insideBottom', offset: -10 } : undefined}
            axisLine={{ stroke: chartTheme.axisColor }}
            tickLine={{ stroke: chartTheme.axisColor }}
            tickFormatter={(value) => typeof value === 'number' ? value.toFixed(1) : value}
          />
          <YAxis
            type="number"
            stroke={chartTheme.axisColor}
            domain={domain?.y || ['dataMin', 'dataMax']}
            label={yLabel ? { value: yLabel, angle: -90, position: 'insideLeft' } : undefined}
            axisLine={{ stroke: chartTheme.axisColor }}
            tickLine={{ stroke: chartTheme.axisColor }}
            tickFormatter={(value) => typeof value === 'number' ? value.toFixed(1) : value}
          />
          <Tooltip />
          {children}
        </RechartComposedChart>
      </ResponsiveContainer>
    </div>
  );
};

const RechartsAdapter: ChartAdapter = {
  ScatterChart: RechartsScatterChart,
  BarChart: RechartsBarChart,
  LineChart: RechartsLineChart,
  ComposedChart: RechartsComposedChart,
};

export default RechartsAdapter;