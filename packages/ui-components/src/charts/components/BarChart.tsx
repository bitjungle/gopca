// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

import React from 'react';
import { useChartConfig } from '../ChartProvider';
import { BarChartProps } from '../types';
import RechartsAdapter from '../adapters/recharts/RechartsAdapter';

export const BarChart: React.FC<BarChartProps> = (props) => {
  const { config } = useChartConfig();
  
  switch (config.provider) {
    case 'recharts':
      return <RechartsAdapter.BarChart {...props} />;
    case 'plotly':
      // Future: return <PlotlyAdapter.BarChart {...props} />;
      console.warn('Plotly adapter not yet implemented, falling back to Recharts');
      return <RechartsAdapter.BarChart {...props} />;
    case 'd3':
      // Future: return <D3Adapter.BarChart {...props} />;
      console.warn('D3 adapter not yet implemented, falling back to Recharts');
      return <RechartsAdapter.BarChart {...props} />;
    default:
      return <RechartsAdapter.BarChart {...props} />;
  }
};