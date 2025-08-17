// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

import { ChartAdapter } from '../../types';
import { PlotlyScatterChart } from './PlotlyScatterChart';
import { PlotlyBarChart } from './PlotlyBarChart';
import { PlotlyLineChart } from './PlotlyLineChart';
import { PlotlyComposedChart } from './PlotlyComposedChart';

const PlotlyAdapter: ChartAdapter = {
  ScatterChart: PlotlyScatterChart,
  BarChart: PlotlyBarChart,
  LineChart: PlotlyLineChart,
  ComposedChart: PlotlyComposedChart
};

export default PlotlyAdapter;