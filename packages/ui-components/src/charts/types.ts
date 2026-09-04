// GoPCA Suite
//
// Copyright © 2025-2026 Rune Mathisen <devel@bitjungle.com>
//
// This file is part of GoPCA Suite.
//
// GoPCA Suite is source-available software with free binary redistribution.
// Official compiled binary releases may be used and redistributed free of charge
// under the GoPCA Suite Source-Available Freeware License.
//
// The source code is provided for viewing, review, education, security analysis,
// research, interoperability analysis, and evaluation only.
//
// Modification, redistribution, publication, sublicensing, reuse, incorporation
// into another project, or creation of derivative works based on the source code
// is not permitted without prior written permission from the copyright holder.
//
// Usage Restriction: GoPCA Suite may not be used, directly or indirectly, for
// military, warfare, weapons, intelligence, surveillance, targeting, or
// law-enforcement surveillance applications.
//
// See LICENSE for the full license terms.

// Type definitions for Plotly chart components
// These are the minimal types needed for the chart components

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
  /**
   * Draws a dashed y = x line. Meaningful only when both axes carry the same
   * quantity on the same scale, as in a predicted-against-measured plot, where
   * agreement is read as closeness to that diagonal.
   */
  identityLine?: boolean;
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