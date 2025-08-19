// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

// Charts module - Legacy chart components removed as part of Plotly migration
// Use the Plotly PCA visualizations from '../charts/adapters/plotly/pca' instead

// Provider (kept for potential future use)
export { ChartProvider, useChartConfig } from './ChartProvider';
export type { ChartLibrary } from './ChartProvider';

// Types (kept for backward compatibility)
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