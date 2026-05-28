// GoPCA Suite
//
// Copyright © 2025-2026 Rune Mathisen <devel@bitjungle.com>
//
// This file is part of GoPCA Suite.
//
// GoPCA Suite is source-available software with free binary redistribution.
// Official compiled binary releases may be used and redistributed free of charge
// under the GoPCA Suite Source-Available Freeware License.
//
// The source code is provided for viewing, review, education, security analysis,
// research, interoperability analysis, and evaluation only.
//
// Modification, redistribution, publication, sublicensing, reuse, incorporation
// into another project, or creation of derivative works based on the source code
// is not permitted without prior written permission from the copyright holder.
//
// Usage Restriction: GoPCA Suite may not be used, directly or indirectly, for
// military, warfare, weapons, intelligence, surveillance, targeting, or
// law-enforcement surveillance applications.
//
// See LICENSE for the full license terms.

// Temporal Loadings Plot for SSA/Temporal PCA - displays U matrix columns

import React, { useMemo } from 'react';
import { Data, Layout } from 'plotly.js';
import { getPlotlyTheme, mergeLayouts } from '../utils/plotlyTheme';
import { PlotlyVisualizationConfig } from '../core/PlotlyVisualization';
import { getExportMenuItems } from '../utils/plotlyExport';
import { PLOT_CONFIG } from '../config/plotConfig';
import { PlotlyWithFullscreen } from '../utils/plotlyFullscreen';
import { getWatermarkDataUrlSync } from '../assets/watermark';

export interface TemporalLoadingsPlotData {
  temporalEigenvectors: number[][];  // U matrix [lags × components]
  explainedVariance?: number[];      // For labeling with explained variance
}

export interface TemporalLoadingsPlotConfig extends PlotlyVisualizationConfig {
  colorScheme?: string[];
  maxComponents?: number;
}

/**
 * Temporal Loadings Plot showing U matrix columns from SSA/Temporal PCA
 * Displays temporal patterns as line plots against lag indices
 */
export class PlotlyTemporalLoadings {
  private data: TemporalLoadingsPlotData;
  private config: TemporalLoadingsPlotConfig;

  constructor(data: TemporalLoadingsPlotData, config?: TemporalLoadingsPlotConfig) {
    this.data = data;
    this.config = {
      maxComponents: 5,
      ...config
    };
  }

  getTraces(): Data[] {
    const traces: Data[] = [];
    const { temporalEigenvectors, explainedVariance } = this.data;

    if (!temporalEigenvectors || temporalEigenvectors.length === 0) {
      return traces;
    }

    // Get dimensions
    const numLags = temporalEigenvectors.length;
    const numComponents = Math.min(
      temporalEigenvectors[0].length,
      this.config.maxComponents || 5
    );

    // Create lag indices
    const lagIndices = Array.from({ length: numLags }, (_, i) => i);

    // Default color scheme
    const colors = this.config.colorScheme || [
      '#3b82f6', '#ef4444', '#10b981', '#f59e0b', '#8b5cf6',
      '#ec4899', '#14b8a6', '#f97316', '#a855f7', '#06b6d4'
    ];

    // Create a trace for each component
    for (let comp = 0; comp < numComponents; comp++) {
      // Extract column comp from U matrix
      const componentValues = temporalEigenvectors.map(row => row[comp]);

      // Create user-friendly label with explained variance if available
      let legendName = `Component ${comp + 1}`;
      if (explainedVariance && explainedVariance[comp] !== undefined) {
        legendName += ` (${explainedVariance[comp].toFixed(1)}%)`;
      }

      traces.push({
        type: 'scatter',
        mode: 'lines',
        x: lagIndices,
        y: componentValues,
        name: legendName,
        line: {
          color: colors[comp % colors.length],
          width: 2.5
        },
        hovertemplate:
          `<b>Component ${comp + 1}</b><br>` +
          'Lag: %{x}<br>' +
          'Loading: %{y:.4f}<br>' +
          '<extra></extra>'
      });
    }

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
    const layout: Partial<Layout> = {
      title: {
        text: 'Temporal Loadings Pattern'
      },
      xaxis: {
        title: {
          text: 'Lag'
        },
        showgrid: true,
        gridcolor: 'rgba(128, 128, 128, 0.2)',
        zeroline: true,
        zerolinecolor: 'rgba(128, 128, 128, 0.4)'
      },
      yaxis: {
        title: {
          text: 'Loading'
        },
        showgrid: true,
        gridcolor: 'rgba(128, 128, 128, 0.2)',
        zeroline: true,
        zerolinecolor: 'rgba(128, 128, 128, 0.4)'
      },
      hovermode: 'closest',
      showlegend: true,
      legend: {
        x: 1,
        xanchor: 'left',
        y: 1,
        yanchor: 'top',
        borderwidth: 1
      }
    };

    return layout;
  }

  getConfig(): Partial<any> {
    return {
      responsive: true,
      displaylogo: false,
      modeBarButtonsToAdd: getExportMenuItems() as any,
      toImageButtonOptions: {
        ...PLOT_CONFIG.export.presentation,
        filename: 'temporal-loadings'
      }
    };
  }
}

/**
 * React component wrapper for Temporal Loadings Plot
 */
export const PCATemporalLoadingsPlot: React.FC<{
  data: TemporalLoadingsPlotData;
  config?: TemporalLoadingsPlotConfig;
}> = ({ data, config }) => {
  const plot = useMemo(() => new PlotlyTemporalLoadings(data, config), [data, config]);

  return (
    <PlotlyWithFullscreen
      data={plot.getTraces()}
      layout={plot.getEnhancedLayout()}
      config={plot.getConfig()}
      style={{ width: '100%', height: '100%' }}
    />
  );
};