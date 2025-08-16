// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

// Provider
export { ChartProvider, useChartConfig } from './ChartProvider';
export type { ChartLibrary } from './ChartProvider';

// Chart Components
export { ScatterChart } from './components/ScatterChart';
export { BarChart } from './components/BarChart';
export { LineChart } from './components/LineChart';
export { ComposedChart } from './components/ComposedChart';

// Types
export type {
  ChartDataPoint,
  ChartDomain,
  ChartMargin,
  BaseChartProps,
  ScatterChartProps,
  BarChartProps,
  LineChartProps,
  ComposedChartProps,
} from './types';

// Native components for use within ComposedChart
export { 
  Scatter,
  Bar,
  Line,
  Cell,
  Legend,
  ReferenceLine,
  Tooltip as RechartsTooltip,
} from 'recharts';