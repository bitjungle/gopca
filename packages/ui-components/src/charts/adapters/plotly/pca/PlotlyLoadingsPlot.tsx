// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Loadings Plot with bar chart and line chart modes

import React, { useMemo } from 'react';
import Plot from 'react-plotly.js';
import { Data, Layout } from 'plotly.js';
import { getPlotlyTheme, mergeLayouts, ThemeMode } from '../utils/plotlyTheme';
import { getExportMenuItems } from '../utils/plotlyExport';

export interface LoadingsPlotData {
  loadings: number[][];  // [components][variables]
  variableNames: string[];
  componentIndex?: number;  // Which PC to show (0-based)
}

export interface LoadingsPlotConfig {
  mode?: 'bar' | 'line' | 'grouped';
  colorScheme?: string[];
  showThreshold?: boolean;
  thresholdValue?: number;
  maxVariables?: number;
  sortByMagnitude?: boolean;
  showGrid?: boolean;
  theme?: ThemeMode;
}

/**
 * Loadings Plot showing variable contributions to principal components
 * Supports bar chart, line chart, and grouped bar modes
 * Reference: Johnson & Wichern (2007), Applied Multivariate Statistical Analysis, Ch. 8
 */
export class PlotlyLoadingsPlot {
  private data: LoadingsPlotData;
  private config: LoadingsPlotConfig;
  
  constructor(data: LoadingsPlotData, config?: LoadingsPlotConfig) {
    this.data = data;
    this.config = {
      mode: 'bar',
      showThreshold: true,
      thresholdValue: 0.3,
      sortByMagnitude: false,
      showGrid: true,
      ...config
    };
  }
  
  private prepareData() {
    const { loadings, variableNames } = this.data;
    const componentIndex = this.data.componentIndex || 0;
    
    // Get loadings for selected component
    let componentLoadings = loadings[componentIndex];
    let sortedVariableNames = [...variableNames];
    
    // Sort by magnitude if requested
    if (this.config.sortByMagnitude) {
      const indices = Array.from({ length: variableNames.length }, (_, i) => i);
      indices.sort((a, b) => Math.abs(componentLoadings[b]) - Math.abs(componentLoadings[a]));
      
      componentLoadings = indices.map(i => componentLoadings[i]);
      sortedVariableNames = indices.map(i => variableNames[i]);
    }
    
    // Limit variables if specified
    if (this.config.maxVariables && this.config.maxVariables < variableNames.length) {
      componentLoadings = componentLoadings.slice(0, this.config.maxVariables);
      sortedVariableNames = sortedVariableNames.slice(0, this.config.maxVariables);
    }
    
    return { componentLoadings, sortedVariableNames };
  }
  
  getTraces(): Data[] {
    const traces: Data[] = [];
    const { componentLoadings, sortedVariableNames } = this.prepareData();
    const componentIndex = this.data.componentIndex || 0;
    
    if (this.config.mode === 'bar') {
      // Bar chart mode
      const colors = this.config.colorScheme || ['#3b82f6', '#ef4444'];
      traces.push({
        type: 'bar',
        x: sortedVariableNames,
        y: componentLoadings,
        name: `PC${componentIndex + 1} Loadings`,
        marker: {
          color: componentLoadings.map(v => v >= 0 ? colors[0] : colors[1]),
          opacity: 0.8
        },
        hovertemplate: '<b>%{x}</b><br>Loading: %{y:.3f}<extra></extra>'
      });
    } else if (this.config.mode === 'line') {
      // Line chart mode - use indices for x-axis
      const colors = this.config.colorScheme || ['#3b82f6', '#ef4444'];
      const xValues = Array.from({ length: sortedVariableNames.length }, (_, i) => i);
      
      traces.push({
        type: 'scatter',
        mode: 'lines+markers',
        x: xValues,
        y: componentLoadings,
        text: sortedVariableNames,
        name: `PC${componentIndex + 1} Loadings`,
        line: {
          color: colors[0],
          width: 2
        },
        marker: {
          size: 8,
          color: componentLoadings.map(v => v >= 0 ? colors[0] : colors[1])
        },
        hovertemplate: '<b>%{text}</b><br>Loading: %{y:.3f}<br>Index: %{x}<extra></extra>'
      });
    } else if (this.config.mode === 'grouped') {
      // Grouped bar chart for multiple components
      const numComponents = Math.min(3, this.data.loadings.length);
      for (let i = 0; i < numComponents; i++) {
        const { componentLoadings: loadings, sortedVariableNames: names } = 
          this.prepareDataForComponent(i);
        
        traces.push({
          type: 'bar',
          x: names,
          y: loadings,
          name: `PC${i + 1}`,
          marker: {
            color: this.config.colorScheme![i % this.config.colorScheme!.length],
            opacity: 0.8
          },
          hovertemplate: '<b>%{x}</b><br>PC' + (i + 1) + ': %{y:.3f}<extra></extra>'
        });
      }
    }
    
    return traces;
  }
  
  private prepareDataForComponent(componentIndex: number) {
    const { loadings, variableNames } = this.data;
    let componentLoadings = loadings[componentIndex];
    let sortedVariableNames = [...variableNames];
    
    if (this.config.sortByMagnitude) {
      const indices = Array.from({ length: variableNames.length }, (_, i) => i);
      indices.sort((a, b) => Math.abs(componentLoadings[b]) - Math.abs(componentLoadings[a]));
      
      componentLoadings = indices.map(i => componentLoadings[i]);
      sortedVariableNames = indices.map(i => variableNames[i]);
    }
    
    if (this.config.maxVariables && this.config.maxVariables < variableNames.length) {
      componentLoadings = componentLoadings.slice(0, this.config.maxVariables);
      sortedVariableNames = sortedVariableNames.slice(0, this.config.maxVariables);
    }
    
    return { componentLoadings, sortedVariableNames };
  }
  
  getEnhancedLayout(): Partial<Layout> {
    const baseLayout = this.getLayout();
    const themeLayout = getPlotlyTheme(this.config.theme || 'light').layout;
    return mergeLayouts(themeLayout, baseLayout);
  }
  
  getLayout(): Partial<Layout> {
    const { sortedVariableNames } = this.prepareData();
    const componentIndex = this.data.componentIndex || 0;
    
    const layout: Partial<Layout> = {
      title: {
        text: this.config.mode === 'grouped' 
          ? 'Loadings Comparison'
          : `Loadings Plot - PC${componentIndex + 1}`
      },
      xaxis: {
        title: {
          text: this.config.mode === 'line' ? 'Variable Index' : 'Variables'
        },
        type: this.config.mode === 'line' ? 'linear' : 'category',
        tickangle: this.config.mode === 'bar' && sortedVariableNames.length > 10 ? -45 : 0
      },
      yaxis: {
        title: {
          text: 'Loading Value'
        },
        zeroline: true,
        zerolinewidth: 2,
        zerolinecolor: 'black',
        showgrid: this.config.showGrid,
        gridcolor: 'rgba(128, 128, 128, 0.2)'
      },
      hovermode: 'x unified',
      showlegend: this.config.mode === 'grouped',
      legend: {
        x: 1.02,
        y: 1,
        xanchor: 'left',
        yanchor: 'top',
        borderwidth: 1
      },
      shapes: [],
      annotations: []
    };
    
    // Add threshold lines if enabled
    if (this.config.showThreshold) {
      layout.shapes = [
        {
          type: 'line',
          x0: 0,
          x1: 1,
          xref: 'paper',
          y0: this.config.thresholdValue,
          y1: this.config.thresholdValue,
          yref: 'y',
          line: {
            color: 'orange',
            width: 2,
            dash: 'dash'
          }
        },
        {
          type: 'line',
          x0: 0,
          x1: 1,
          xref: 'paper',
          y0: -this.config.thresholdValue!,
          y1: -this.config.thresholdValue!,
          yref: 'y',
          line: {
            color: 'orange',
            width: 2,
            dash: 'dash'
          }
        }
      ];
      
      layout.annotations = [
        {
          text: `Threshold: ±${this.config.thresholdValue}`,
          x: 1,
          xref: 'paper',
          y: this.config.thresholdValue!,
          yref: 'y',
          xanchor: 'left',
          yanchor: 'bottom',
          showarrow: false,
          font: {
            color: 'orange',
            size: 10
          }
        }
      ];
    }
    
    return layout;
  }
  
  getConfig(): Partial<any> {
    return {
      responsive: true,
      displaylogo: false,
      modeBarButtonsToAdd: getExportMenuItems() as any,
      toImageButtonOptions: {
        format: 'png',
        filename: 'loadings-plot',
        height: 1200,
        width: 1600,
        scale: 2
      }
    };
  }
}

/**
 * React component wrapper for Loadings Plot
 */
export const PCALoadingsPlot: React.FC<{
  data: LoadingsPlotData;
  config?: LoadingsPlotConfig;
}> = ({ data, config }) => {
  const plot = useMemo(() => new PlotlyLoadingsPlot(data, config), [data, config]);
  
  return (
    <Plot
      data={plot.getTraces()}
      layout={plot.getEnhancedLayout()}
      config={plot.getConfig()}
      style={{ width: '100%', height: '100%' }}
      useResizeHandler={true}
    />
  );
};