// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

// Circle of Correlations visualization for PCA

import React, { useMemo } from 'react';
import { Data, Layout } from 'plotly.js';
import { getPlotlyTheme, mergeLayouts, ThemeMode } from '../utils/plotlyTheme';
import { getExportMenuItems } from '../utils/plotlyExport';
import { PLOT_CONFIG, getScaledMarkerSize } from '../config/plotConfig';
import { PlotlyWithFullscreen } from '../utils/plotlyFullscreen';
import { getWatermarkDataUrlSync } from '../assets/watermark';

export interface CircleOfCorrelationsData {
  loadings: number[][];  // [n_components][n_variables]
  variableNames: string[];
  explainedVariance: number[];
}

export interface CircleOfCorrelationsConfig {
  pcX?: number;  // PC for X-axis (1-indexed)
  pcY?: number;  // PC for Y-axis (1-indexed)
  showCircle?: boolean;
  showGrid?: boolean;
  showLabels?: boolean;
  minVectorLength?: number;  // Minimum vector length to display
  colorByMagnitude?: boolean;
  arrowWidth?: number;
  labelSize?: number;
  theme?: ThemeMode;
  colorScheme?: string[];  // Color palette for visualization
  fontScale?: number;  // Scale factor for all font sizes (default: 1.0)
}

/**
 * Circle of Correlations showing variable correlations with principal components
 * Vectors represent correlations between original variables and PCs
 * Reference: Abdi & Williams (2010), "Principal component analysis"
 */
export class PlotlyCircleOfCorrelations {
  private data: CircleOfCorrelationsData;
  private config: CircleOfCorrelationsConfig;

  constructor(data: CircleOfCorrelationsData, config?: CircleOfCorrelationsConfig) {
    this.data = data;
    this.config = {
      pcX: 1,
      pcY: 2,
      showCircle: true,
      showGrid: true,
      showLabels: true,
      minVectorLength: 0.1,
      colorByMagnitude: false,  // Changed to false for consistency with Biplot
      arrowWidth: 2,
      labelSize: 10,
      ...config
    };
  }

  private prepareData() {
    const { loadings, variableNames } = this.data;
    const pcX = (this.config.pcX || 1) - 1;
    const pcY = (this.config.pcY || 2) - 1;

    // Extract correlations (loadings) for selected PCs
    const correlationsX = loadings[pcX];
    const correlationsY = loadings[pcY];

    // Calculate vector magnitudes
    const magnitudes = correlationsX.map((x, i) =>
      Math.sqrt(x * x + correlationsY[i] * correlationsY[i])
    );

    // Filter by minimum vector length
    const filteredIndices = magnitudes
      .map((mag, i) => mag >= this.config.minVectorLength! ? i : -1)
      .filter(i => i >= 0);

    return {
      correlationsX: filteredIndices.map(i => correlationsX[i]),
      correlationsY: filteredIndices.map(i => correlationsY[i]),
      filteredNames: filteredIndices.map(i => variableNames[i]),
      magnitudes: filteredIndices.map(i => magnitudes[i]),
      pcX,
      pcY
    };
  }

  getTraces(): Data[] {
    const traces: Data[] = [];
    const { correlationsX, correlationsY, filteredNames, magnitudes } = this.prepareData();

    // Add unit circle
    if (this.config.showCircle) {
      const theta = Array.from({ length: 101 }, (_, i) => (i * 2 * Math.PI) / 100);
      traces.push({
        type: 'scatter',
        mode: 'lines',
        x: theta.map(t => Math.cos(t)),
        y: theta.map(t => Math.sin(t)),
        line: {
          color: 'gray',
          width: 2,
          dash: 'dash'
        },
        showlegend: false,
        hoverinfo: 'skip'
      });

      // Add inner circles at 0.5
      traces.push({
        type: 'scatter',
        mode: 'lines',
        x: theta.map(t => 0.5 * Math.cos(t)),
        y: theta.map(t => 0.5 * Math.sin(t)),
        line: {
          color: 'lightgray',
          width: 1,
          dash: 'dot'
        },
        showlegend: false,
        hoverinfo: 'skip'
      });
    }

    // Add correlation vectors - use same color as Biplot (colorScheme[1])
    filteredNames.forEach((name, i) => {
      const color = this.config.colorByMagnitude
        ? `hsl(${240 - magnitudes[i] * 240}, 70%, 50%)`  // Blue to red gradient
        : (this.config.colorScheme?.[1] || '#ef4444');  // Use index 1 like Biplot

      // Vector line
      traces.push({
        type: 'scatter',
        mode: 'lines',
        x: [0, correlationsX[i]],
        y: [0, correlationsY[i]],
        line: {
          color: color,
          width: this.config.arrowWidth
        },
        showlegend: false,
        hovertemplate: `<b>${name}</b><br>` +
                       `PC${(this.config.pcX || 1)}: ${correlationsX[i].toFixed(3)}<br>` +
                       `PC${(this.config.pcY || 2)}: ${correlationsY[i].toFixed(3)}<br>` +
                       `Magnitude: ${magnitudes[i].toFixed(3)}<extra></extra>`
      });

      // Arrowhead marker
      traces.push({
        type: 'scatter',
        mode: 'markers',
        x: [correlationsX[i]],
        y: [correlationsY[i]],
        marker: {
          symbol: 'circle',
          size: getScaledMarkerSize(6, this.config.fontScale || 1.0),
          color: color
        },
        showlegend: false,
        hoverinfo: 'skip'
      });
    });

    // Add labels
    if (this.config.showLabels) {
      traces.push({
        type: 'scatter',
        mode: 'text',
        x: correlationsX.map((x) => x * 1.15),  // Slightly beyond vector tip
        y: correlationsY.map((y) => y * 1.15),
        text: filteredNames,
        textposition: 'middle center',
        textfont: {
          size: Math.round((this.config.labelSize || 10) * (this.config.fontScale || 1.0)),
          color: this.config.theme === 'dark' ? '#e5e7eb' : '#374151'
        },
        showlegend: false,
        hoverinfo: 'skip'
      });
    }

    // Add axes through origin
    traces.push({
      type: 'scatter',
      mode: 'lines',
      x: [-1.1, 1.1],
      y: [0, 0],
      line: {
        color: this.config.theme === 'dark' ? '#6b7280' : '#374151',
        width: 1
      },
      showlegend: false,
      hoverinfo: 'skip'
    });

    traces.push({
      type: 'scatter',
      mode: 'lines',
      x: [0, 0],
      y: [-1.1, 1.1],
      line: {
        color: this.config.theme === 'dark' ? '#6b7280' : '#374151',
        width: 1
      },
      showlegend: false,
      hoverinfo: 'skip'
    });

    return traces;
  }

  getEnhancedLayout(): Partial<Layout> {
    const baseLayout = this.getLayout();
    const themeLayout = getPlotlyTheme(this.config.theme || 'light', this.config.fontScale).layout;
    
    // Add watermark if enabled
    let watermarkImages: any[] = [];
    if (PLOT_CONFIG.watermark.enabled) {
      const watermarkUrl = getWatermarkDataUrlSync();
      watermarkImages = [{
        source: watermarkUrl,
        xref: PLOT_CONFIG.watermark.position.xref,
        yref: PLOT_CONFIG.watermark.position.yref,
        x: PLOT_CONFIG.watermark.position.x,
        y: PLOT_CONFIG.watermark.position.y,
        sizex: PLOT_CONFIG.watermark.size.width / 400,  // Normalize to plot units
        sizey: PLOT_CONFIG.watermark.size.height / 400, // Normalize to plot units
        xanchor: PLOT_CONFIG.watermark.position.xanchor,
        yanchor: PLOT_CONFIG.watermark.position.yanchor,
        sizing: 'contain',
        opacity: PLOT_CONFIG.watermark.opacity,
        layer: 'above'
      }];
    }
    
    return mergeLayouts(themeLayout, baseLayout, { images: watermarkImages });
  }

  getLayout(): Partial<Layout> {
    const { explainedVariance } = this.data;
    const { pcX, pcY } = this.prepareData();

    const layout: Partial<Layout> = {
      title: {
        text: 'Circle of Correlations'
      },
      xaxis: {
        title: {
          text: `PC${pcX + 1} (${explainedVariance[pcX].toFixed(1)}%)`
        },
        range: [-1.2, 1.2],
        zeroline: false,
        showgrid: this.config.showGrid,
        gridcolor: 'rgba(128, 128, 128, 0.2)',
        scaleanchor: 'y',
        scaleratio: 1
      },
      yaxis: {
        title: {
          text: `PC${pcY + 1} (${explainedVariance[pcY].toFixed(1)}%)`
        },
        range: [-1.2, 1.2],
        zeroline: false,
        showgrid: this.config.showGrid,
        gridcolor: 'rgba(128, 128, 128, 0.2)'
      },
      hovermode: 'closest',
      showlegend: false,
      annotations: [
        // Add quadrant labels
        {
          text: '+/+',
          x: 1.05,
          y: 1.05,
          xref: 'x',
          yref: 'y',
          showarrow: false,
          font: { size: Math.round(12 * (this.config.fontScale || 1.0)), color: 'gray' }
        },
        {
          text: '-/+',
          x: -1.05,
          y: 1.05,
          xref: 'x',
          yref: 'y',
          showarrow: false,
          font: { size: Math.round(12 * (this.config.fontScale || 1.0)), color: 'gray' }
        },
        {
          text: '-/-',
          x: -1.05,
          y: -1.05,
          xref: 'x',
          yref: 'y',
          showarrow: false,
          font: { size: Math.round(12 * (this.config.fontScale || 1.0)), color: 'gray' }
        },
        {
          text: '+/-',
          x: 1.05,
          y: -1.05,
          xref: 'x',
          yref: 'y',
          showarrow: false,
          font: { size: Math.round(12 * (this.config.fontScale || 1.0)), color: 'gray' }
        }
      ]
    };

    // Add arrow annotations for vectors
    const { correlationsX, correlationsY, magnitudes } = this.prepareData();
    const arrowAnnotations = correlationsX.map((_x, i) => ({
      x: correlationsX[i],
      y: correlationsY[i],
      ax: 0,
      ay: 0,
      xref: 'x' as any,
      yref: 'y' as any,
      axref: 'x' as any,
      ayref: 'y' as any,
      showarrow: true,
      arrowhead: 2,
      arrowsize: 1,
      arrowwidth: this.config.arrowWidth,
      arrowcolor: this.config.colorByMagnitude
        ? `hsl(${240 - magnitudes[i] * 240}, 70%, 50%)`
        : '#3b82f6'
    }));

    layout.annotations = [...(layout.annotations || []), ...arrowAnnotations];

    return layout;
  }

  getConfig(): Partial<any> {
    return {
      responsive: true,
      displaylogo: false,
      modeBarButtonsToAdd: getExportMenuItems() as any,
      toImageButtonOptions: {
        ...PLOT_CONFIG.export.presentation,
        filename: 'circle-of-correlations'
      }
    };
  }
}

/**
 * React component wrapper for Circle of Correlations
 */
export const PCACircleOfCorrelations: React.FC<{
  data: CircleOfCorrelationsData;
  config?: CircleOfCorrelationsConfig;
}> = ({ data, config }) => {
  const plot = useMemo(() => new PlotlyCircleOfCorrelations(data, config), [data, config]);

  return (
    <PlotlyWithFullscreen
      data={plot.getTraces()}
      layout={plot.getEnhancedLayout()}
      config={plot.getConfig()}
      style={{ width: '100%', height: '100%' }}
    />
  );
};