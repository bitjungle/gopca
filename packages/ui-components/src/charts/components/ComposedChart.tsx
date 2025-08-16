// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

import React from 'react';
import { useChartConfig } from '../ChartProvider';
import { ComposedChartProps } from '../types';
import RechartsAdapter from '../adapters/recharts/RechartsAdapter';
import PlotlyAdapter from '../adapters/plotly/PlotlyAdapter';

export const ComposedChart: React.FC<ComposedChartProps> = (props) => {
  const { config } = useChartConfig();
  
  switch (config.provider) {
    case 'recharts':
      return <RechartsAdapter.ComposedChart {...props} />;
    case 'plotly':
      return <PlotlyAdapter.ComposedChart {...props} />;
    case 'd3':
      // Future: return <D3Adapter.ComposedChart {...props} />;
      console.warn('D3 adapter not yet implemented, falling back to Recharts');
      return <RechartsAdapter.ComposedChart {...props} />;
    default:
      return <RechartsAdapter.ComposedChart {...props} />;
  }
};