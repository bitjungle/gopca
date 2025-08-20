// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Diagnostic Plot for PCA outlier detection

import React, { useMemo } from 'react';
import Plot from 'react-plotly.js';
import { Data, Layout } from 'plotly.js';
import { getPlotlyTheme, mergeLayouts, ThemeMode } from '../utils/plotlyTheme';
import { getExportMenuItems } from '../utils/plotlyExport';

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
 * 
 * Statistical basis:
 * - X-axis: Mahalanobis distance (Hotelling's T²) - measures leverage in model space
 * - Y-axis: Residual Sum of Squares (Q-statistic/SPE) - measures distance from model space
 * 
 * Threshold calculations (performed in backend):
 * - T² limit: Based on F-distribution: T² = p(n-1)/(n-p) * F_{p,n-p}(α)
 *   Reference: Hotelling, H. (1931). The generalization of Student's ratio.
 * - Q limit: Based on Jackson & Mudholkar approximation for SPE distribution
 *   Reference: Jackson & Mudholkar (1979). Control procedures for residuals in PCA.
 */
export class PlotlyDiagnosticPlot {
  private data: DiagnosticPlotData;
  private config: DiagnosticPlotConfig;
  
  constructor(data: DiagnosticPlotData, config?: DiagnosticPlotConfig) {
    this.data = data;
    this.config = {
      showThresholds: true,
      confidenceLevel: 0.95,
      showLabels: false,  // Default to false as user prefers
      labelThreshold: 10,
      pointSize: 8,
      theme: 'light',
      ...config
    };
    
    // Validate that thresholds are provided when needed
    if (this.config.showThresholds && 
        (!this.config.mahalanobisThreshold || !this.config.rssThreshold)) {
      console.warn('Diagnostic plot: Thresholds not provided from backend, hiding threshold lines');
      this.config.showThresholds = false;
    }
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
    
    // Add labels for samples
    if (this.config.showLabels && sampleNames) {
      // Calculate normalized distance for all points to determine which to label
      // Use thresholds if available, otherwise use max values for normalization
      const maxMahalanobis = this.config.mahalanobisThreshold || Math.max(...mahalanobisDistances) || 1;
      const maxRSS = this.config.rssThreshold || Math.max(...residualSumOfSquares) || 1;
      
      // Map all points with their distances
      const allPoints = mahalanobisDistances.map((md, i) => ({
        index: i,
        x: md,
        y: residualSumOfSquares[i],
        // Calculate normalized distance from origin
        distance: Math.sqrt(
          Math.pow(md / maxMahalanobis, 2) +
          Math.pow(residualSumOfSquares[i] / maxRSS, 2)
        )
      }));
      
      // Sort by distance (furthest from origin first) and take top N
      allPoints.sort((a, b) => b.distance - a.distance);
      const topPoints = allPoints.slice(0, this.config.labelThreshold || 10);
      
      traces.push({
        type: 'scatter',
        mode: 'text',
        x: topPoints.map(p => p.x),
        y: topPoints.map(p => p.y),
        text: topPoints.map(p => sampleNames[p.index]),
        textposition: 'top center',
        textfont: {
          size: 10,
          color: this.config.theme === 'dark' ? '#e5e7eb' : '#374151'
        },
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
    const { mahalanobisDistances, residualSumOfSquares } = this.data;
    
    const layout: Partial<Layout> = {
      title: {
        text: 'PCA Diagnostic Plot'
      },
      xaxis: {
        title: {
          text: 'Mahalanobis Distance (Hotelling\'s T²)'
        },
        zeroline: false,
        showgrid: true,
        gridcolor: 'rgba(128, 128, 128, 0.2)',
        rangemode: 'tozero'
      },
      yaxis: {
        title: {
          text: 'Residual Sum of Squares (Q-statistic)'
        },
        zeroline: false,
        showgrid: true,
        gridcolor: 'rgba(128, 128, 128, 0.2)',
        rangemode: 'tozero'
      },
      hovermode: 'closest',
      showlegend: true,
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
    
    // Add threshold lines with proper labels
    if (this.config.showThresholds && this.config.mahalanobisThreshold && this.config.rssThreshold) {
      // T² threshold (vertical line) - represents Hotelling's T-squared limit
      layout.shapes!.push({
        type: 'line',
        x0: this.config.mahalanobisThreshold,
        x1: this.config.mahalanobisThreshold,
        y0: 0,
        y1: 1,
        yref: 'paper',
        line: {
          color: this.config.colorScheme?.[3] || '#C44E52',  // Use palette color (red from deep palette)
          width: 2,
          dash: 'dash'
        }
      });
      
      // Q threshold (horizontal line) - represents SPE/Q-statistic limit
      layout.shapes!.push({
        type: 'line',
        x0: 0,
        x1: 1,
        xref: 'paper',
        y0: this.config.rssThreshold,
        y1: this.config.rssThreshold,
        line: {
          color: this.config.colorScheme?.[3] || '#C44E52',  // Use palette color (red from deep palette)
          width: 2,
          dash: 'dash'
        }
      });
      
      // Add quadrant labels
      const maxMD = Math.max(...mahalanobisDistances) * 1.1;
      const maxRSS = Math.max(...residualSumOfSquares) * 1.1;
      
      // Calculate confidence percentage for labels
      const confidencePercent = Math.round((this.config.confidenceLevel || 0.95) * 100);
      
      layout.annotations = [
        // Threshold labels
        {
          text: `T²-limit (${confidencePercent}%)`,
          x: this.config.mahalanobisThreshold,
          y: maxRSS * 0.95,  // Position near top of plot
          xanchor: 'left',
          yanchor: 'bottom',
          showarrow: false,
          font: { size: 11, color: this.config.colorScheme?.[3] || '#C44E52' },
          textangle: '-90'
        },
        {
          text: `Q-limit (${confidencePercent}%)`,
          x: maxMD * 0.95,  // Position near right of plot
          y: this.config.rssThreshold,
          xanchor: 'right',
          yanchor: 'bottom',
          showarrow: false,
          font: { size: 11, color: this.config.colorScheme?.[3] || '#C44E52' }
        },
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
        format: 'png',
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