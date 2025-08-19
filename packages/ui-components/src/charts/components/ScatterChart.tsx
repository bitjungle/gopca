// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

import React from 'react';
import { ScatterChartProps } from '../types';
import PlotlyAdapter from '../adapters/plotly/PlotlyAdapter';

// ScatterChart now only uses Plotly - Recharts support removed
export const ScatterChart: React.FC<ScatterChartProps> = (props) => {
  return <PlotlyAdapter.ScatterChart {...props} />;
};