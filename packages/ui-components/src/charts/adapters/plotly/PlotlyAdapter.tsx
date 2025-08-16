// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

import React, { useMemo } from 'react';
import Plot from 'react-plotly.js';
import { Layout, Config, Data } from 'plotly.js';
import {
  ScatterChartProps,
  BarChartProps,
  LineChartProps,
  ComposedChartProps,
  ChartAdapter,
} from '../../types';
import { useChartTheme } from '../../../hooks/useChartTheme';
import { useTheme } from '../../../contexts/ThemeContext';

// Common Plotly configuration
const getPlotlyConfig = (): Partial<Config> => ({
  displayModeBar: true,
  displaylogo: false,
  responsive: true,
  toImageButtonOptions: {
    format: 'png',
    width: 1200,
    height: 800,
    scale: 2,
  },
  modeBarButtonsToRemove: ['lasso2d', 'select2d'],
});

// Convert our theme to Plotly layout
const getPlotlyLayout = (
  theme: ReturnType<typeof useChartTheme>,
  isDarkMode: boolean,
  xLabel?: string,
  yLabel?: string,
  title?: string
): Partial<Layout> => ({
  title: title ? { text: title } : undefined,
  paper_bgcolor: isDarkMode ? '#1f2937' : '#ffffff',
  plot_bgcolor: isDarkMode ? '#1f2937' : '#ffffff',
  font: {
    color: theme.axisColor,
  },
  xaxis: {
    title: xLabel ? { text: xLabel } : undefined,
    gridcolor: theme.gridColor,
    zerolinecolor: theme.referenceLineColor,
    tickcolor: theme.axisColor,
    linecolor: theme.axisColor,
  },
  yaxis: {
    title: yLabel ? { text: yLabel } : undefined,
    gridcolor: theme.gridColor,
    zerolinecolor: theme.referenceLineColor,
    tickcolor: theme.axisColor,
    linecolor: theme.axisColor,
  },
  hovermode: 'closest',
  dragmode: 'zoom',
  showlegend: false,
  margin: {
    l: 80,
    r: 20,
    t: 40,
    b: 60,
  },
});

export const PlotlyScatterChart: React.FC<ScatterChartProps> = ({
  data,
  domain,
  xDataKey = 'x',
  yDataKey = 'y',
  xLabel,
  yLabel,
  showGrid = true,
  showReferenceLines = true,
  fill = '#3B82F6',
  stroke = '#1E40AF',
  children: _children,
  className,
}) => {
  const chartTheme = useChartTheme();
  const { theme } = useTheme();
  const isDarkMode = theme === 'dark';

  const plotlyData: Data[] = useMemo(() => {
    // Extract x and y values
    const xValues = data.map(d => d[xDataKey] as number);
    const yValues = data.map(d => d[yDataKey] as number);
    
    // Extract colors if available
    const colors = data.map(d => d.color || fill);
    
    return [{
      x: xValues,
      y: yValues,
      type: 'scatter',
      mode: 'markers',
      marker: {
        color: colors,
        size: 8,
        line: {
          color: stroke,
          width: 1,
        },
      },
      text: data.map(d => d.name || ''),
      hovertemplate: '%{text}<br>X: %{x:.3f}<br>Y: %{y:.3f}<extra></extra>',
    }];
  }, [data, xDataKey, yDataKey, fill, stroke]);

  const layout = useMemo(() => {
    const baseLayout = getPlotlyLayout(chartTheme, isDarkMode, xLabel, yLabel);
    
    if (domain?.x) {
      baseLayout.xaxis = {
        ...baseLayout.xaxis,
        range: domain.x,
      };
    }
    
    if (domain?.y) {
      baseLayout.yaxis = {
        ...baseLayout.yaxis,
        range: domain.y,
      };
    }

    if (!showGrid) {
      baseLayout.xaxis!.showgrid = false;
      baseLayout.yaxis!.showgrid = false;
    }

    if (showReferenceLines) {
      baseLayout.xaxis!.zeroline = true;
      baseLayout.yaxis!.zeroline = true;
    }

    return baseLayout;
  }, [chartTheme, isDarkMode, xLabel, yLabel, domain, showGrid, showReferenceLines]);

  return (
    <div className={className}>
      <Plot
        data={plotlyData}
        layout={layout}
        config={getPlotlyConfig()}
        style={{ width: '100%', height: '100%' }}
        useResizeHandler
      />
    </div>
  );
};

export const PlotlyBarChart: React.FC<BarChartProps> = ({
  data,
  dataKey,
  xDataKey = 'x',
  xLabel,
  yLabel,
  showGrid = true,
  fill = '#3B82F6',
  className,
}) => {
  const chartTheme = useChartTheme();
  const { theme } = useTheme();
  const isDarkMode = theme === 'dark';

  const plotlyData: Data[] = useMemo(() => {
    const xValues = data.map(d => d[xDataKey]);
    const yValues = data.map(d => d[dataKey] as number);
    
    return [{
      x: xValues,
      y: yValues,
      type: 'bar',
      marker: {
        color: fill,
      },
    }];
  }, [data, xDataKey, dataKey, fill]);

  const layout = useMemo(() => {
    const baseLayout = getPlotlyLayout(chartTheme, isDarkMode, xLabel, yLabel);
    
    if (!showGrid) {
      baseLayout.xaxis!.showgrid = false;
      baseLayout.yaxis!.showgrid = false;
    }

    return baseLayout;
  }, [chartTheme, isDarkMode, xLabel, yLabel, showGrid]);

  return (
    <div className={className}>
      <Plot
        data={plotlyData}
        layout={layout}
        config={getPlotlyConfig()}
        style={{ width: '100%', height: '100%' }}
        useResizeHandler
      />
    </div>
  );
};

export const PlotlyLineChart: React.FC<LineChartProps> = ({
  data,
  dataKey,
  xDataKey = 'x',
  xLabel,
  yLabel,
  showGrid = true,
  stroke = '#3B82F6',
  strokeWidth = 2,
  dot = false,
  className,
}) => {
  const chartTheme = useChartTheme();
  const { theme } = useTheme();
  const isDarkMode = theme === 'dark';

  const plotlyData: Data[] = useMemo(() => {
    const xValues = data.map(d => d[xDataKey]);
    const yValues = data.map(d => d[dataKey] as number);
    
    return [{
      x: xValues,
      y: yValues,
      type: 'scatter',
      mode: dot ? 'lines+markers' : 'lines',
      line: {
        color: stroke,
        width: strokeWidth,
      },
      marker: dot ? {
        color: stroke,
        size: 6,
      } : undefined,
    }];
  }, [data, xDataKey, dataKey, stroke, strokeWidth, dot]);

  const layout = useMemo(() => {
    const baseLayout = getPlotlyLayout(chartTheme, isDarkMode, xLabel, yLabel);
    
    if (!showGrid) {
      baseLayout.xaxis!.showgrid = false;
      baseLayout.yaxis!.showgrid = false;
    }

    return baseLayout;
  }, [chartTheme, isDarkMode, xLabel, yLabel, showGrid]);

  return (
    <div className={className}>
      <Plot
        data={plotlyData}
        layout={layout}
        config={getPlotlyConfig()}
        style={{ width: '100%', height: '100%' }}
        useResizeHandler
      />
    </div>
  );
};

export const PlotlyComposedChart: React.FC<ComposedChartProps> = ({
  data,
  domain,
  xDataKey = 'x',
  xLabel,
  yLabel,
  showGrid = true,
  children: _children,
  className,
}) => {
  const chartTheme = useChartTheme();
  const { theme } = useTheme();
  const isDarkMode = theme === 'dark';

  // For composed charts, we'll need to process children to extract traces
  // This is a simplified version - in practice, we'd need to map React children
  // to Plotly traces based on their types and props
  const plotlyData: Data[] = useMemo(() => {
    // Default to scatter plot if no specific traces defined
    const xValues = data.map(d => d[xDataKey] as number);
    const yValues = data.map(d => d.y as number);
    
    return [{
      x: xValues,
      y: yValues,
      type: 'scatter',
      mode: 'markers',
      marker: {
        color: '#3B82F6',
        size: 8,
      },
    }];
  }, [data, xDataKey]);

  const layout = useMemo(() => {
    const baseLayout = getPlotlyLayout(chartTheme, isDarkMode, xLabel, yLabel);
    
    if (domain?.x) {
      baseLayout.xaxis = {
        ...baseLayout.xaxis,
        range: domain.x,
      };
    }
    
    if (domain?.y) {
      baseLayout.yaxis = {
        ...baseLayout.yaxis,
        range: domain.y,
      };
    }

    if (!showGrid) {
      baseLayout.xaxis!.showgrid = false;
      baseLayout.yaxis!.showgrid = false;
    }

    return baseLayout;
  }, [chartTheme, isDarkMode, xLabel, yLabel, domain, showGrid]);

  return (
    <div className={className}>
      <Plot
        data={plotlyData}
        layout={layout}
        config={getPlotlyConfig()}
        style={{ width: '100%', height: '100%' }}
        useResizeHandler
      />
    </div>
  );
};

const PlotlyAdapter: ChartAdapter = {
  ScatterChart: PlotlyScatterChart,
  BarChart: PlotlyBarChart,
  LineChart: PlotlyLineChart,
  ComposedChart: PlotlyComposedChart,
};

export default PlotlyAdapter;