// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

import React, { useMemo } from 'react';
import Plot from 'react-plotly.js';
import { Layout, Config, Data } from 'plotly.js';
import { useChartTheme } from '../../../hooks/useChartTheme';
import { useTheme } from '../../../contexts/ThemeContext';

interface EllipseParams {
  centerX: number;
  centerY: number;
  semiMajorAxis: number;
  semiMinorAxis: number;
  angle: number; // in radians
  group: string;
  points?: number;
}

interface PlotlyScoresPlotProps {
  data: Array<{
    x: number;
    y: number;
    name: string;
    group?: string;
    color?: string;
    index?: number;
  }>;
  xLabel?: string;
  yLabel?: string;
  domain?: {
    x?: [number, number];
    y?: [number, number];
  };
  groupEllipses?: Record<string, EllipseParams>;
  showEllipses?: boolean;
  confidenceLevel?: 0.90 | 0.95 | 0.99;
  showRowLabels?: boolean;
  maxLabelsToShow?: number;
  groupColorMap?: Map<string, string>;
  className?: string;
}

// Generate ellipse path points
const generateEllipsePoints = (params: EllipseParams): { x: number[]; y: number[] } => {
  const { centerX, centerY, semiMajorAxis, semiMinorAxis, angle, points = 100 } = params;
  
  const x: number[] = [];
  const y: number[] = [];
  
  for (let i = 0; i <= points; i++) {
    const t = (i / points) * 2 * Math.PI;
    const ellipseX = semiMajorAxis * Math.cos(t);
    const ellipseY = semiMinorAxis * Math.sin(t);
    
    // Rotate and translate
    const rotatedX = ellipseX * Math.cos(angle) - ellipseY * Math.sin(angle);
    const rotatedY = ellipseX * Math.sin(angle) + ellipseY * Math.cos(angle);
    
    x.push(centerX + rotatedX);
    y.push(centerY + rotatedY);
  }
  
  return { x, y };
};

export const PlotlyScoresPlot: React.FC<PlotlyScoresPlotProps> = ({
  data,
  xLabel = 'PC1',
  yLabel = 'PC2',
  domain,
  groupEllipses,
  showEllipses = false,
  confidenceLevel = 0.95,
  showRowLabels = false,
  maxLabelsToShow: _maxLabelsToShow = 10,
  groupColorMap,
  className,
}) => {
  const chartTheme = useChartTheme();
  const { theme } = useTheme();
  const isDarkMode = theme === 'dark';

  const plotlyData: Data[] = useMemo(() => {
    const traces: Data[] = [];
    
    // Group data by group if available
    const groups = new Map<string, typeof data>();
    
    data.forEach(point => {
      const group = point.group || 'Default';
      if (!groups.has(group)) {
        groups.set(group, []);
      }
      groups.get(group)!.push(point);
    });
    
    // Create scatter trace for each group
    groups.forEach((groupData, groupName) => {
      const color = groupColorMap?.get(groupName) || groupData[0]?.color || '#3B82F6';
      
      traces.push({
        x: groupData.map(d => d.x),
        y: groupData.map(d => d.y),
        type: 'scatter',
        mode: showRowLabels ? 'markers+text' as any : 'markers',
        name: groupName,
        marker: {
          color: color,
          size: 8,
          line: {
            color: 'rgba(0, 0, 0, 0.2)',
            width: 1,
          },
        },
        text: showRowLabels ? groupData.map(d => d.name) : undefined,
        textposition: 'top center',
        textfont: {
          size: 10,
          color: chartTheme.axisColor,
        },
        hovertemplate: '%{text}<br>X: %{x:.3f}<br>Y: %{y:.3f}<extra></extra>',
      });
    });
    
    // Add confidence ellipses if enabled
    if (showEllipses && groupEllipses) {
      Object.entries(groupEllipses).forEach(([groupName, ellipse]) => {
        const color = groupColorMap?.get(groupName) || '#3B82F6';
        const ellipsePoints = generateEllipsePoints(ellipse);
        
        traces.push({
          x: ellipsePoints.x,
          y: ellipsePoints.y,
          type: 'scatter',
          mode: 'lines',
          name: `${groupName} (${confidenceLevel * 100}% CI)`,
          line: {
            color: color,
            width: 2,
            dash: 'dot',
          },
          fill: 'toself',
          fillcolor: `${color}20`, // 20% opacity
          hoverinfo: 'skip',
          showlegend: false,
        });
      });
    }
    
    return traces;
  }, [data, groupEllipses, showEllipses, confidenceLevel, showRowLabels, groupColorMap, chartTheme]);

  const layout: Partial<Layout> = useMemo(() => {
    const baseLayout: Partial<Layout> = {
      title: undefined,
      paper_bgcolor: isDarkMode ? '#1f2937' : '#ffffff',
      plot_bgcolor: isDarkMode ? '#1f2937' : '#ffffff',
      font: {
        color: chartTheme.axisColor,
      },
      xaxis: {
        title: { text: xLabel },
        gridcolor: chartTheme.gridColor,
        zerolinecolor: chartTheme.referenceLineColor,
        zerolinewidth: 2,
        tickcolor: chartTheme.axisColor,
        linecolor: chartTheme.axisColor,
        range: domain?.x,
        zeroline: true,
      },
      yaxis: {
        title: { text: yLabel },
        gridcolor: chartTheme.gridColor,
        zerolinecolor: chartTheme.referenceLineColor,
        zerolinewidth: 2,
        tickcolor: chartTheme.axisColor,
        linecolor: chartTheme.axisColor,
        range: domain?.y,
        zeroline: true,
      },
      hovermode: 'closest',
      dragmode: 'zoom',
      legend: {
        x: 1,
        y: 1,
        xanchor: 'right',
        yanchor: 'top',
        bgcolor: isDarkMode ? 'rgba(31, 41, 55, 0.8)' : 'rgba(255, 255, 255, 0.8)',
        bordercolor: chartTheme.axisColor,
        borderwidth: 1,
      },
      margin: {
        l: 80,
        r: 20,
        t: 40,
        b: 60,
      },
    };
    
    return baseLayout;
  }, [chartTheme, isDarkMode, xLabel, yLabel, domain]);

  const config: Partial<Config> = useMemo(() => ({
    displayModeBar: true,
    displaylogo: false,
    responsive: true,
    toImageButtonOptions: {
      format: 'png',
      width: 1200,
      height: 800,
      scale: 2,
    },
    modeBarButtonsToAdd: [
      'zoom2d',
      'pan2d',
      'zoomIn2d',
      'zoomOut2d',
      'autoScale2d',
      'resetScale2d',
    ],
    modeBarButtonsToRemove: ['lasso2d', 'select2d'],
  }), []);

  return (
    <div className={className} style={{ width: '100%', height: '100%' }}>
      <Plot
        data={plotlyData}
        layout={layout}
        config={config}
        style={{ width: '100%', height: '100%' }}
        useResizeHandler
      />
    </div>
  );
};