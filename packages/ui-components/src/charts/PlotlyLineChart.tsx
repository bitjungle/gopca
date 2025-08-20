// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

import React, { useEffect, useRef } from 'react';
import Plotly from 'plotly.js-basic-dist-min';
import { LineChartProps } from './types';
import { getPlotlyTheme, mergeLayouts } from './utils';
import { useChartTheme } from '../hooks/useChartTheme';

export const PlotlyLineChart: React.FC<LineChartProps> = ({
  data,
  dataKey,
  xDataKey = 'x',
  xLabel,
  yLabel,
  domain,
  margin,
  width = '100%',
  height = 400,
  className,
  showGrid = true,
  stroke = '#3b82f6',
  strokeWidth = 2,
  dot = true
}) => {
  const containerRef = useRef<HTMLDivElement>(null);
  const { theme } = useChartTheme();
  
  useEffect(() => {
    if (!containerRef.current) return;
    
    const baseTheme = getPlotlyTheme(theme);
    
    // Prepare data
    const xValues = data.map(d => d[xDataKey]);
    const yValues = data.map(d => d[dataKey]);
    
    // Create trace
    const traces = [{
      x: xValues,
      y: yValues,
      type: 'scatter' as const,
      mode: (dot ? 'lines+markers' : 'lines') as any,
      line: {
        color: stroke,
        width: strokeWidth
      },
      marker: dot ? {
        color: stroke,
        size: 6
      } : undefined,
      hovertemplate: '%{x}<br>%{y:.3f}<extra></extra>'
    }];
    
    // Add zero reference line
    traces.push({
      x: [Math.min(...(xValues as number[])), Math.max(...(xValues as number[]))],
      y: [0, 0],
      mode: 'lines',
      type: 'scatter' as const,
      line: {
        color: theme === 'dark' ? '#6b7280' : '#9ca3af',
        width: 1,
        dash: 'solid'
      },
      showlegend: false,
      hoverinfo: 'skip'
    } as any);
    
    // Prepare layout
    const layout = mergeLayouts(
      baseTheme.layout,
      {
        xaxis: {
          title: xLabel ? { text: xLabel } : undefined,
          showgrid: showGrid,
          tickmode: 'linear',
          dtick: 1
        },
        yaxis: {
          title: yLabel ? { text: yLabel } : undefined,
          range: domain?.y,
          showgrid: showGrid,
          zeroline: true,
          zerolinecolor: theme === 'dark' ? '#6b7280' : '#9ca3af',
          zerolinewidth: 2
        },
        showlegend: false,
        hovermode: 'closest',
        margin: margin ? {
          l: margin.left ?? 60,
          r: margin.right ?? 30,
          t: margin.top ?? 30,
          b: margin.bottom ?? 60
        } : baseTheme.layout.margin,
        width: typeof width === 'number' ? width : undefined,
        height: typeof height === 'number' ? height : undefined,
        autosize: typeof width === 'string'
      }
    );
    
    // Create or update plot
    Plotly.react(containerRef.current, traces, layout, baseTheme.config);
    
    // Handle resize
    const handleResize = () => {
      if (containerRef.current) {
        Plotly.Plots.resize(containerRef.current);
      }
    };
    
    window.addEventListener('resize', handleResize);
    
    return () => {
      window.removeEventListener('resize', handleResize);
      if (containerRef.current) {
        Plotly.purge(containerRef.current);
      }
    };
  }, [data, dataKey, xDataKey, xLabel, yLabel, domain, margin, width, height, 
      theme, showGrid, stroke, strokeWidth, dot]);
  
  return (
    <div 
      ref={containerRef}
      className={className}
      style={{ width, height }}
    />
  );
};