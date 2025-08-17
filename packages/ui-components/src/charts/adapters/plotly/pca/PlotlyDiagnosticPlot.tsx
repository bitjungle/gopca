// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Diagnostic Plot for PCA outlier detection

import React, { useMemo } from 'react';
import Plot from 'react-plotly.js';
import { Data, Layout } from 'plotly.js';
import { getExportMenuItems } from '../utils/plotlyExport';
import { getPlotlyTheme, mergeLayouts, ThemeMode } from '../utils/plotlyTheme';

export interface DiagnosticPlotData {
  mahalanobisDistances: number[];
  residualSumOfSquares: number[];
  sampleNames?: string[];
  groups?: string[];
}

export interface DiagnosticPlotConfig {
  showThresholds?: boolean;
  mahalanobisThreshold?: number;  // Chi-square based threshold
  rssThreshold?: number;
  confidenceLevel?: number;  // For Mahalanobis threshold calculation
  showLabels?: boolean;
  labelThreshold?: number;  // Number of outliers to label
  colorScheme?: string[];
  pointSize?: number;
  theme?: ThemeMode;
}

/**
 * Diagnostic Plot for identifying outliers in PCA
 * Combines Mahalanobis distance (leverage) and Residual Sum of Squares (RSS)
 * Reference: Hubert et al. (2005), "ROBPCA: A new approach to robust principal component analysis"
 */
export class PlotlyDiagnosticPlot {
  private data: DiagnosticPlotData;
  private config: DiagnosticPlotConfig;
  
  constructor(data: DiagnosticPlotData, config?: DiagnosticPlotConfig) {
    this.data = data;
    this.config = {
      showThresholds: true,
      confidenceLevel: 0.975,
      showLabels: false,  // Default to false as user prefers
      labelThreshold: 10,
      pointSize: 8,
      theme: 'light',
      ...config
    };
    
    // Calculate default thresholds if not provided
    if (!this.config.mahalanobisThreshold) {
      // Use chi-square distribution with appropriate degrees of freedom
      // For 97.5% confidence level with 2 PCs: ~7.38
      this.config.mahalanobisThreshold = this.calculateMahalanobisThreshold();
    }
    
    if (!this.config.rssThreshold) {
      // Use median + 3*MAD as robust threshold
      this.config.rssThreshold = this.calculateRSSThreshold();
    }
  }
  
  private calculateMahalanobisThreshold(): number {
    // Chi-square quantile for given confidence level
    // Approximation for 2 degrees of freedom
    const alpha = 1 - (this.config.confidenceLevel || 0.975);
    return -2 * Math.log(alpha);  // Simplified approximation
  }
  
  private calculateRSSThreshold(): number {
    const { residualSumOfSquares } = this.data;
    const sorted = [...residualSumOfSquares].sort((a, b) => a - b);
    const median = sorted[Math.floor(sorted.length / 2)];
    
    // Calculate MAD (Median Absolute Deviation)
    const deviations = residualSumOfSquares.map(v => Math.abs(v - median));
    const sortedDev = [...deviations].sort((a, b) => a - b);
    const mad = sortedDev[Math.floor(sortedDev.length / 2)];
    
    return median + 3 * mad * 1.4826;  // 1.4826 is consistency constant
  }
  
  private identifyOutliers() {
    const { mahalanobisDistances, residualSumOfSquares } = this.data;
    const outliers: number[] = [];
    const goodLeverage: number[] = [];
    const orthogonal: number[] = [];
    const regular: number[] = [];
    
    mahalanobisDistances.forEach((md, i) => {
      const rss = residualSumOfSquares[i];
      const isMahalanobisOutlier = md > this.config.mahalanobisThreshold!;
      const isRSSOutlier = rss > this.config.rssThreshold!;
      
      if (isMahalanobisOutlier && isRSSOutlier) {
        outliers.push(i);
      } else if (isMahalanobisOutlier && !isRSSOutlier) {
        goodLeverage.push(i);
      } else if (!isMahalanobisOutlier && isRSSOutlier) {
        orthogonal.push(i);
      } else {
        regular.push(i);
      }
    });
    
    return { outliers, goodLeverage, orthogonal, regular };
  }
  
  getTraces(): Data[] {
    const traces: Data[] = [];
    const { mahalanobisDistances, residualSumOfSquares, sampleNames } = this.data;
    const { outliers, goodLeverage, orthogonal, regular } = this.identifyOutliers();
    
    // Use colorScheme from config, fallback to defaults if not provided
    const colors = this.config.colorScheme || [
      '#10b981', '#3b82f6', '#f59e0b', '#ef4444', '#8b5cf6',
      '#ec4899', '#14b8a6', '#f97316', '#6366f1', '#84cc16'
    ];
    
    // Define point categories and their properties
    const categories = [
      { name: 'Regular', indices: regular, color: colors[0] || '#10b981', symbol: 'circle' },
      { name: 'Good Leverage', indices: goodLeverage, color: colors[1] || '#3b82f6', symbol: 'square' },
      { name: 'Orthogonal Outliers', indices: orthogonal, color: colors[2] || '#f59e0b', symbol: 'diamond' },
      { name: 'Bad Outliers', indices: outliers, color: colors[3] || '#ef4444', symbol: 'x' }
    ];
    
    // Add traces for each category
    categories.forEach(cat => {
      if (cat.indices.length === 0) return;
      
      traces.push({
        type: 'scatter',
        mode: 'markers',
        x: cat.indices.map(i => mahalanobisDistances[i]),
        y: cat.indices.map(i => residualSumOfSquares[i]),
        name: cat.name,
        marker: {
          color: cat.color,
          size: this.config.pointSize,
          symbol: cat.symbol,
          opacity: 0.7
        },
        text: sampleNames ? cat.indices.map(i => sampleNames[i]) : undefined,
        hovertemplate: '<b>%{text}</b><br>' +
                      'Mahalanobis: %{x:.2f}<br>' +
                      'RSS: %{y:.2f}<extra></extra>'
      });
    });
    
    // Add labels for outliers
    if (this.config.showLabels && sampleNames) {
      // Combine all potential outliers
      const outlierIndices = [...outliers, ...goodLeverage, ...orthogonal];
      
      if (outlierIndices.length > 0) {
        // Sort by combined distance from origin
        const outlierPoints = outlierIndices.map(i => ({
          index: i,
          x: mahalanobisDistances[i],
          y: residualSumOfSquares[i],
          distance: Math.sqrt(
            Math.pow(mahalanobisDistances[i] / this.config.mahalanobisThreshold!, 2) +
            Math.pow(residualSumOfSquares[i] / this.config.rssThreshold!, 2)
          )
        }));
        
        // Sort by distance and take top N
        outlierPoints.sort((a, b) => b.distance - a.distance);
        const topOutliers = outlierPoints.slice(0, this.config.labelThreshold);
        
        traces.push({
          type: 'scatter',
          mode: 'text',
          x: topOutliers.map(p => p.x),
          y: topOutliers.map(p => p.y),
          text: topOutliers.map(p => sampleNames[p.index]),
          textposition: 'top center',
          textfont: {
            size: 10,
            color: 'black'
          },
          showlegend: false,
          hoverinfo: 'skip'
        });
      }
    }
    
    return traces;
  }
  
  getEnhancedLayout(): Partial<Layout> {
    const baseLayout = this.getLayout();
    const themeLayout = getPlotlyTheme(this.config.theme || 'light').layout;
    return mergeLayouts(themeLayout, baseLayout);
  }
  
  getLayout(): Partial<Layout> {
    const { mahalanobisDistances, residualSumOfSquares } = this.data;
    
    const layout: Partial<Layout> = {
      title: {
        text: 'PCA Diagnostic Plot'
      },
      xaxis: {
        title: {
          text: 'Mahalanobis Distance (Leverage)'
        },
        zeroline: false,
        showgrid: true,
        gridcolor: 'rgba(128, 128, 128, 0.2)',
        rangemode: 'tozero'
      },
      yaxis: {
        title: {
          text: 'Residual Sum of Squares (RSS)'
        },
        zeroline: false,
        showgrid: true,
        gridcolor: 'rgba(128, 128, 128, 0.2)',
        rangemode: 'tozero'
      },
      hovermode: 'closest',
      showlegend: true,
      legend: {
        x: 1,
        y: 1,
        xanchor: 'right',
        yanchor: 'top',
        bgcolor: 'rgba(255, 255, 255, 0.9)',
        bordercolor: 'black',
        borderwidth: 1
      },
      shapes: [],
      annotations: []
    };
    
    // Add threshold lines
    if (this.config.showThresholds) {
      // Mahalanobis threshold (vertical line)
      layout.shapes!.push({
        type: 'line',
        x0: this.config.mahalanobisThreshold,
        x1: this.config.mahalanobisThreshold,
        y0: 0,
        y1: 1,
        yref: 'paper',
        line: {
          color: 'red',
          width: 2,
          dash: 'dash'
        }
      });
      
      // RSS threshold (horizontal line)
      layout.shapes!.push({
        type: 'line',
        x0: 0,
        x1: 1,
        xref: 'paper',
        y0: this.config.rssThreshold,
        y1: this.config.rssThreshold,
        line: {
          color: 'red',
          width: 2,
          dash: 'dash'
        }
      });
      
      // Add quadrant labels
      const maxMD = Math.max(...mahalanobisDistances) * 1.1;
      const maxRSS = Math.max(...residualSumOfSquares) * 1.1;
      
      layout.annotations = [
        {
          text: 'Regular',
          x: this.config.mahalanobisThreshold! / 2,
          y: this.config.rssThreshold! / 2,
          showarrow: false,
          font: { size: 12, color: 'gray' },
          opacity: 0.5
        },
        {
          text: 'Good Leverage',
          x: (this.config.mahalanobisThreshold! + maxMD) / 2,
          y: this.config.rssThreshold! / 2,
          showarrow: false,
          font: { size: 12, color: 'gray' },
          opacity: 0.5
        },
        {
          text: 'Orthogonal',
          x: this.config.mahalanobisThreshold! / 2,
          y: (this.config.rssThreshold! + maxRSS) / 2,
          showarrow: false,
          font: { size: 12, color: 'gray' },
          opacity: 0.5
        },
        {
          text: 'Bad Outliers',
          x: (this.config.mahalanobisThreshold! + maxMD) / 2,
          y: (this.config.rssThreshold! + maxRSS) / 2,
          showarrow: false,
          font: { size: 12, color: 'gray' },
          opacity: 0.5
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
        format: 'svg',
        filename: 'diagnostic-plot',
        height: 1200,
        width: 1600,
        scale: 2
      }
    };
  }
}

/**
 * React component wrapper for Diagnostic Plot
 */
export const PCADiagnosticPlot: React.FC<{
  data: DiagnosticPlotData;
  config?: DiagnosticPlotConfig;
}> = ({ data, config }) => {
  const plot = useMemo(() => new PlotlyDiagnosticPlot(data, config), [data, config]);
  
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