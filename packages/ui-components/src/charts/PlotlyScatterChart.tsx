// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

import React, { useEffect, useRef, useMemo } from 'react';
import Plotly from 'plotly.js-dist-min';
import { ScatterChartProps } from './types';
import { getPlotlyTheme, mergeLayouts, calculatePlotlyLabels, getPlotlyTextPosition } from './utils';
import { useChartTheme } from '../hooks/useChartTheme';

export const PlotlyScatterChart: React.FC<ScatterChartProps> = ({
  data,
  domain,
  margin,
  width = '100%',
  height = 400,
  className,
  xDataKey = 'x',
  yDataKey = 'y',
  xLabel,
  yLabel,
  showGrid = true,
  showReferenceLines = true,
  fill = '#3b82f6',
  stroke
}) => {
  const containerRef = useRef<HTMLDivElement>(null);
  const { theme } = useChartTheme();

  // Extract data points
  const plotData = useMemo(() => {
    return data.map(d => ({
      x: d[xDataKey] as number,
      y: d[yDataKey] as number,
      label: d.label as string | undefined,
      index: d.index as number | undefined
    }));
  }, [data, xDataKey, yDataKey]);

  // Calculate smart labels if labels are provided
  const labels = useMemo(() => {
    const hasLabels = plotData.some(d => d.label);
    if (!hasLabels) {
return plotData.map(() => '');
}

    return calculatePlotlyLabels({
      showLabels: true,
      maxLabels: 10, // Default max labels
      labels: plotData.map(d => d.label || ''),
      data: plotData
    });
  }, [plotData]);

  // Calculate text positions for labels
  const textPositions = useMemo(() => {
    return plotData.map(d => getPlotlyTextPosition(d.x, d.y));
  }, [plotData]);

  useEffect(() => {
    if (!containerRef.current) {
return;
}

    const baseTheme = getPlotlyTheme(theme);

    // Prepare traces
    const traces = [{
      x: plotData.map(d => d.x),
      y: plotData.map(d => d.y),
      mode: (labels.some(l => l) ? 'markers+text' : 'markers') as any,
      type: 'scatter' as const,
      marker: {
        color: fill,
        size: 8,
        line: stroke ? {
          color: stroke,
          width: 1
        } : undefined
      },
      text: labels,
      textposition: textPositions as any,
      textfont: {
        size: 10,
        color: theme === 'dark' ? '#e5e7eb' : '#1f2937'
      },
      hovertemplate: '%{text}<br>X: %{x:.2f}<br>Y: %{y:.2f}<extra></extra>'
    }];

    // Add reference lines if requested
    if (showReferenceLines) {
      // Vertical line at x=0
      traces.push({
        x: [0, 0],
        y: [domain?.y?.[0] ?? -10, domain?.y?.[1] ?? 10],
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

      // Horizontal line at y=0
      traces.push({
        x: [domain?.x?.[0] ?? -10, domain?.x?.[1] ?? 10],
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
    }

    // Prepare layout
    const layout = mergeLayouts(
      baseTheme.layout,
      {
        xaxis: {
          title: xLabel ? { text: xLabel } : undefined,
          range: domain?.x,
          showgrid: showGrid,
          zeroline: showReferenceLines
        },
        yaxis: {
          title: yLabel ? { text: yLabel } : undefined,
          range: domain?.y,
          showgrid: showGrid,
          zeroline: showReferenceLines
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
  }, [data, plotData, labels, textPositions, theme, domain, margin, width, height,
      xLabel, yLabel, showGrid, showReferenceLines, fill, stroke]);

  return (
    <div
      ref={containerRef}
      className={className}
      style={{ width, height }}
    />
  );
};