// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

import { ReactNode } from 'react';

export interface ChartDataPoint {
  x: number;
  y: number;
  [key: string]: any;
}

export interface ChartDomain {
  x?: [number, number];
  y?: [number, number];
}

export interface ChartMargin {
  top?: number;
  right?: number;
  bottom?: number;
  left?: number;
}

export interface BaseChartProps {
  data: ChartDataPoint[];
  domain?: ChartDomain;
  margin?: ChartMargin;
  width?: number | string;
  height?: number | string;
  className?: string;
  onMouseMove?: (event: any) => void;
  onMouseDown?: (event: any) => void;
  onMouseUp?: (event: any) => void;
  onMouseLeave?: (event: any) => void;
}

export interface ScatterChartProps extends BaseChartProps {
  xDataKey?: string;
  yDataKey?: string;
  xLabel?: string;
  yLabel?: string;
  showGrid?: boolean;
  showReferenceLines?: boolean;
  tooltip?: ReactNode | ((props: any) => ReactNode);
  dot?: ReactNode | ((props: any) => React.ReactElement);
  fill?: string;
  stroke?: string;
  children?: ReactNode;
}

export interface BarChartProps extends BaseChartProps {
  dataKey: string;
  xDataKey?: string;
  xLabel?: string;
  yLabel?: string;
  showGrid?: boolean;
  fill?: string;
  children?: ReactNode;
}

export interface LineChartProps extends BaseChartProps {
  dataKey: string;
  xDataKey?: string;
  xLabel?: string;
  yLabel?: string;
  showGrid?: boolean;
  stroke?: string;
  strokeWidth?: number;
  dot?: boolean | ReactNode;
  children?: ReactNode;
}

export interface ComposedChartProps extends BaseChartProps {
  xDataKey?: string;
  xLabel?: string;
  yLabel?: string;
  showGrid?: boolean;
  children?: ReactNode;
}

export interface ChartAdapter {
  ScatterChart: React.FC<ScatterChartProps>;
  BarChart: React.FC<BarChartProps>;
  LineChart: React.FC<LineChartProps>;
  ComposedChart: React.FC<ComposedChartProps>;
}