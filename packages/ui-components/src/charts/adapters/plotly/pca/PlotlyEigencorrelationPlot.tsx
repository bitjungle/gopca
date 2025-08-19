// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Eigencorrelation Plot - Heatmap of correlations between variables and PCs

import React, { useMemo } from 'react';
import Plot from 'react-plotly.js';
import { Data, Layout } from 'plotly.js';
import { getExportMenuItems } from '../utils/plotlyExport';
import { getPlotlyTheme, mergeLayouts, ThemeMode } from '../utils/plotlyTheme';

export interface EigencorrelationPlotData {
  correlations: number[][];  // [n_components][n_variables]
  variableNames: string[];
  explainedVariance: number[];
}

export interface EigencorrelationPlotConfig {
  maxComponents?: number;
  colorScale?: string | any[];
  showValues?: boolean;
  valueFormat?: string;  // Format string for values (e.g., '.2f')
  clusterVariables?: boolean;
  clusterComponents?: boolean;
  annotationThreshold?: number;  // Only show values above this threshold
  theme?: ThemeMode;
  colorScheme?: string[];  // Color palette for visualization
}

/**
 * Eigencorrelation Plot showing correlations between metadata variables and principal components
 * as a heatmap visualization. This helps identify confounding influences or meaningful associations
 * between external covariates (e.g., batch, treatment, clinical variables) and PCA components.
 * Reference: Friendly (2002), "Corrgrams: Exploratory displays for correlation matrices"
 */
export class PlotlyEigencorrelationPlot {
  private data: EigencorrelationPlotData;
  private config: EigencorrelationPlotConfig;
  
  constructor(data: EigencorrelationPlotData, config?: EigencorrelationPlotConfig) {
    this.data = data;
    this.config = {
      colorScale: 'RdBu',
      showValues: true,
      valueFormat: '.2f',
      clusterVariables: false,
      clusterComponents: false,
      annotationThreshold: 0.3,
      theme: 'light',
      ...config
    };
  }
  
  private prepareData() {
    const { correlations, variableNames, explainedVariance } = this.data;
    
    // Ensure we have valid data
    if (!correlations || correlations.length === 0 || !variableNames) {
      return { matrix: [], componentLabels: [], variableNames: [], numComponents: 0 };
    }
    
    // Limit components based on available data
    const maxAvailableComponents = Math.min(
      correlations.length,
      explainedVariance?.length || correlations.length
    );
    
    const numComponents = this.config.maxComponents 
      ? Math.min(this.config.maxComponents, maxAvailableComponents)
      : maxAvailableComponents;
    
    // Prepare correlation matrix for heatmap
    const matrix = correlations.slice(0, numComponents);
    
    // Create component labels with safe access to explainedVariance
    const componentLabels = Array.from({ length: numComponents }, (_, i) => {
      const variance = explainedVariance && i < explainedVariance.length 
        ? explainedVariance[i].toFixed(1) 
        : '0.0';
      return `PC${i + 1} (${variance}%)`;
    });
    
    // Optionally cluster variables by similarity
    let orderedVariableNames = [...variableNames];
    let orderedMatrix = matrix.map(row => [...row]);
    
    if (this.config.clusterVariables) {
      // Simple clustering by first PC loading
      const indices = Array.from({ length: variableNames.length }, (_, i) => i);
      indices.sort((a, b) => Math.abs(matrix[0][b]) - Math.abs(matrix[0][a]));
      
      orderedVariableNames = indices.map(i => variableNames[i]);
      orderedMatrix = matrix.map(row => indices.map(i => row[i]));
    }
    
    return {
      matrix: orderedMatrix,
      componentLabels,
      variableNames: orderedVariableNames,
      numComponents
    };
  }
  
  getTraces(): Data[] {
    const traces: Data[] = [];
    const { matrix, componentLabels, variableNames } = this.prepareData();
    
    // Transpose matrix for heatmap (plotly expects [y][x])
    const transposedMatrix: number[][] = [];
    for (let i = 0; i < variableNames.length; i++) {
      transposedMatrix[i] = [];
      for (let j = 0; j < componentLabels.length; j++) {
        transposedMatrix[i][j] = matrix[j][i];
      }
    }
    
    // Main heatmap
    traces.push({
      type: 'heatmap',
      z: transposedMatrix,
      x: Array.from({ length: componentLabels.length }, (_, i) => i),
      y: Array.from({ length: variableNames.length }, (_, i) => i),
      colorscale: this.config.colorScale as any,
      zmin: -1,
      zmax: 1,
      colorbar: {
        title: {
          text: 'Correlation'
        } as any,
        tickmode: 'linear',
        tick0: -1,
        dtick: 0.5,
        len: 0.9,
        thickness: 15
      },
      hoverongaps: false,
      hovertemplate: '<b>%{y}</b><br>%{x}<br>Correlation: %{z:.3f}<extra></extra>'
    });
    
    // Add text annotations if enabled
    if (this.config.showValues) {
      const annotations: any[] = [];
      
      for (let i = 0; i < variableNames.length; i++) {
        for (let j = 0; j < componentLabels.length; j++) {
          const value = transposedMatrix[i][j];
          
          // Only show annotations for significant correlations
          if (Math.abs(value) >= this.config.annotationThreshold!) {
            annotations.push({
              x: j,
              y: i,
              text: value.toFixed(2),
              showarrow: false,
              font: {
                size: 10,
                color: Math.abs(value) > 0.5 ? 'white' : 'black'
              }
            });
          }
        }
      }
      
      // We'll add these annotations to the layout
      traces.push({
        type: 'scatter',
        x: [],
        y: [],
        mode: 'markers',
        showlegend: false,
        hoverinfo: 'skip'
      });
    }
    
    return traces;
  }
  
  getEnhancedLayout(): Partial<Layout> {
    const baseLayout = this.getLayout();
    const themeLayout = getPlotlyTheme(this.config.theme || 'light').layout;
    return mergeLayouts(themeLayout, baseLayout);
  }
  
  getLayout(): Partial<Layout> {
    const { variableNames, componentLabels } = this.prepareData();
    
    const layout: Partial<Layout> = {
      title: {
        text: 'Eigencorrelation Plot: Component-Metadata Correlations'
      },
      xaxis: {
        title: {
          text: 'Principal Components',
          standoff: 20
        },
        side: 'bottom',
        tickangle: variableNames.length > 10 ? -45 : 0,
        tickmode: 'array',
        tickvals: Array.from({ length: componentLabels.length }, (_, i) => i),
        ticktext: componentLabels
      },
      yaxis: {
        title: {
          text: 'Metadata Variables',
          standoff: 20
        },
        tickmode: 'array',
        tickvals: Array.from({ length: variableNames.length }, (_, i) => i),
        ticktext: variableNames,
        autorange: 'reversed'  // Put first variable at top
      },
      hovermode: 'closest',
      showlegend: false,
      annotations: []
    };
    
    // Add value annotations if enabled
    if (this.config.showValues) {
      const { matrix } = this.prepareData();
      
      for (let i = 0; i < variableNames.length; i++) {
        for (let j = 0; j < componentLabels.length; j++) {
          const value = matrix[j][i];
          
          if (Math.abs(value) >= this.config.annotationThreshold!) {
            layout.annotations!.push({
              x: j,
              y: i,
              xref: 'x',
              yref: 'y',
              text: value.toFixed(2),
              showarrow: false,
              font: {
                size: 10,
                color: Math.abs(value) > 0.5 ? 'white' : 'black'
              }
            });
          }
        }
      }
    }
    
    // Add interpretation guide
    layout.annotations!.push({
      text: 'Strong correlations (|r| > 0.7) indicate metadata variables that are associated with PC variance',
      xref: 'paper',
      yref: 'paper',
      x: 0.5,
      y: -0.08,
      xanchor: 'center',
      showarrow: false,
      font: { size: 10, color: 'gray' }
    });
    
    // Adjust margins to accommodate labels
    layout.margin = {
      l: Math.max(...variableNames.map(n => n.length)) * 6 + 50,
      r: 100,
      t: 50,
      b: 120
    };
    
    return layout;
  }
  
  getConfig(): Partial<any> {
    return {
      responsive: true,
      displaylogo: false,
      modeBarButtonsToAdd: getExportMenuItems() as any,
      toImageButtonOptions: {
        format: 'svg',
        filename: 'eigencorrelation-matrix',
        height: 1600,
        width: 1200,
        scale: 2
      }
    };
  }
}

/**
 * React component wrapper for Eigencorrelation Plot
 */
export const PCAEigencorrelationPlot: React.FC<{
  data: EigencorrelationPlotData;
  config?: EigencorrelationPlotConfig;
}> = ({ data, config }) => {
  const plot = useMemo(() => new PlotlyEigencorrelationPlot(data, config), [data, config]);
  
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