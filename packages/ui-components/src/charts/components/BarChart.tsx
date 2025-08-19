// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

import React from 'react';
import { BarChartProps } from '../types';
import PlotlyAdapter from '../adapters/plotly/PlotlyAdapter';

// BarChart now only uses Plotly - Recharts support removed
export const BarChart: React.FC<BarChartProps> = (props) => {
  return <PlotlyAdapter.BarChart {...props} />;
};