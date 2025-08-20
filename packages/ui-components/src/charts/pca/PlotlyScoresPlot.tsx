// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// PCA Scores Plot with confidence ellipses, smart labels, and density overlays

import React, { useMemo } from 'react';
import Plot from 'react-plotly.js';
import { Data, Layout, Config } from 'plotly.js';
import { PlotlyVisualization, PlotlyVisualizationConfig } from '../core/PlotlyVisualization';
import {
  calculateConfidenceEllipse,
  generateEllipsePath,
  calculateSmartLabels,
  kernelDensityEstimate2D,
  Point2D
} from '../utils/plotlyMath';
import { optimizeTraceType, getOptimalConfig } from '../utils/plotlyPerformance';
import { getExportMenuItems } from '../utils/plotlyExport';

export interface ScoresPlotData {
  scores: number[][];
  groups: string[];
  groupValues?: number[]; // For continuous data
  groupType?: 'categorical' | 'continuous';
  sampleNames?: string[];
  explainedVariance: number[];
  pc1?: number;
  pc2?: number;
}

export interface ScoresPlotConfig extends PlotlyVisualizationConfig {
  showEllipses?: boolean;
  ellipseConfidence?: number;
  showSmartLabels?: boolean;
  maxLabels?: number;
  showDensity?: boolean;
  colorScheme?: string[];
}

/**
 * PCA Scores Plot with advanced features
 * Implements smart labels, confidence ellipses, and optional density overlays
 */
export class PlotlyScoresPlot extends PlotlyVisualization<ScoresPlotData> {
  protected scoresConfig: ScoresPlotConfig;

  constructor(data: ScoresPlotData, config?: ScoresPlotConfig) {
    super(data, config);
    this.scoresConfig = {
      showEllipses: true,
      ellipseConfidence: 0.95,
      showSmartLabels: true,
      maxLabels: 10,
      showDensity: false,
      colorScheme: [
        '#1f77b4', '#ff7f0e', '#2ca02c', '#d62728', '#9467bd',
        '#8c564b', '#e377c2', '#7f7f7f', '#bcbd22', '#17becf'
      ],
      ...config
    };
  }

  protected getDataSize(): number {
    return this.data.scores.length;
  }

  protected getTraces(): Data[] {
    const { scores, groups, groupValues, groupType = 'categorical', sampleNames, pc1 = 0, pc2 = 1 } = this.data;
    const traces: Data[] = [];

    // Handle continuous vs categorical data
    if (groupType === 'continuous' && groupValues) {
      return this.getContinuousTraces();
    }

    // Handle empty groups - default to single group
    const effectiveGroups = groups && groups.length > 0 ? groups : scores.map(() => 'All Samples');

    // Get unique groups
    const uniqueGroups = Array.from(new Set(effectiveGroups));

    // Calculate smart labels globally
    const allPoints: Point2D[] = scores.map(s => ({ x: s[pc1], y: s[pc2] }));
    const smartLabelIndices = this.scoresConfig.showSmartLabels
      ? calculateSmartLabels(allPoints, this.scoresConfig.maxLabels!)
      : [];

    // Add density overlay if enabled
    if (this.scoresConfig.showDensity && scores.length > 20) {
      traces.push(...this.getDensityTraces(uniqueGroups, pc1, pc2));
    }

    // Create traces for each group
    uniqueGroups.forEach((group, groupIndex) => {
      const groupIndices = effectiveGroups.map((g, i) => g === group ? i : -1).filter(i => i >= 0);
      const groupScores = groupIndices.map(i => scores[i]);
      const groupPoints = groupScores.map(s => ({ x: s[pc1], y: s[pc2] }));

      // Prepare hover text
      const hovertext = groupIndices.map(i => {
        const label = sampleNames?.[i] || `Sample ${i}`;
        return `<b>${label}</b><br>Group: ${group}<br>PC${pc1 + 1}: ${scores[i][pc1].toFixed(2)}<br>PC${pc2 + 1}: ${scores[i][pc2].toFixed(2)}`;
      });

      // Determine trace type based on performance
      const traceType = optimizeTraceType(groupScores, this.config.dataThreshold!);

      // Main scatter trace - markers only (WebGL compatible)
      traces.push({
        type: traceType as any,
        mode: 'markers',
        name: group,
        x: groupScores.map(s => s[pc1]),
        y: groupScores.map(s => s[pc2]),
        hovertext: hovertext,
        hovertemplate: '%{hovertext}<extra></extra>',
        marker: {
          size: 8,
          color: this.scoresConfig.colorScheme![groupIndex % this.scoresConfig.colorScheme!.length],
          opacity: 0.8
        },
        selectedpoints: undefined,
        selected: {
          marker: {
            size: 12,
            opacity: 1
          }
        } as any,
        unselected: {
          marker: {
            opacity: 0.3
          }
        } as any
      });

      // Add confidence ellipse if enabled
      if (this.scoresConfig.showEllipses && groupScores.length > 2) {
        const ellipseTrace = this.getEllipseTrace(groupPoints, group, groupIndex);
        if (ellipseTrace) {
          traces.push(ellipseTrace);
        }
      }
    });

    // Add text labels as a separate trace (if enabled)
    // This ensures text renders properly with both scatter and scattergl
    if (this.scoresConfig.showSmartLabels && smartLabelIndices.length > 0) {
      const labelX: number[] = [];
      const labelY: number[] = [];
      const labelText: string[] = [];

      smartLabelIndices.forEach(i => {
        labelX.push(scores[i][pc1]);
        labelY.push(scores[i][pc2]);
        labelText.push(sampleNames?.[i] || `Sample ${i}`);
      });

      traces.push({
        type: 'scatter',
        mode: 'text',
        x: labelX,
        y: labelY,
        text: labelText,
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

  protected getStandardTraces(): Data[] {
    return this.getTraces();
  }

  protected getWebGLTraces(): Data[] {
    const traces = this.getTraces();
    return traces.map(trace => {
      if (trace.type === 'scatter' && trace.mode?.includes('markers')) {
        return { ...trace, type: 'scattergl' as any };
      }
      return trace;
    });
  }

  private getContinuousTraces(): Data[] {
    const { scores, groupValues, sampleNames, pc1 = 0, pc2 = 1 } = this.data;
    const traces: Data[] = [];

    if (!groupValues || groupValues.length === 0) {
      return this.getTraces(); // Fall back to categorical
    }

    // Calculate min/max for continuous values
    const validValues = groupValues.filter(v => v !== null && v !== undefined && !isNaN(v) && isFinite(v));
    const min = Math.min(...validValues);
    const max = Math.max(...validValues);

    // Calculate smart labels globally
    const allPoints: Point2D[] = scores.map(s => ({ x: s[pc1], y: s[pc2] }));
    const smartLabelIndices = this.scoresConfig.showSmartLabels
      ? calculateSmartLabels(allPoints, this.scoresConfig.maxLabels!)
      : [];

    // Create a custom colorscale from the palette
    const palette = this.scoresConfig.colorScheme || ['#440154', '#31688e', '#35b779', '#fde725'];
    const colorscale: [number, string][] = palette.map((color, i) => [
      i / (palette.length - 1),
      color
    ]);

    // Prepare hover text
    const hovertext = scores.map((score, i) => {
      const label = sampleNames?.[i] || `Sample ${i}`;
      const value = groupValues[i];
      const valueStr = value !== null && value !== undefined && !isNaN(value) && isFinite(value)
        ? value.toFixed(2)
        : 'Missing';
      return `<b>${label}</b><br>Value: ${valueStr}<br>PC${pc1 + 1}: ${score[pc1].toFixed(2)}<br>PC${pc2 + 1}: ${score[pc2].toFixed(2)}`;
    });

    // Prepare text labels
    const text = scores.map((_, i) => {
      const label = sampleNames?.[i] || `Sample ${i}`;
      return smartLabelIndices.includes(i) ? label : '';
    });

    // Check if we have any actual labels to display
    const hasLabels = smartLabelIndices.length > 0;

    // Create single trace with gradient colors
    // Use the actual numeric values for colors, not interpolated hex colors
    traces.push({
      type: 'scatter',
      mode: hasLabels ? ('markers+text' as any) : 'markers',
      name: 'Samples',
      x: scores.map(s => s[pc1]),
      y: scores.map(s => s[pc2]),
      text: hasLabels ? text : undefined,
      hovertext: hovertext,
      hovertemplate: '%{hovertext}<extra></extra>',
      textposition: hasLabels ? 'top center' : undefined,
      textfont: hasLabels ? {
        size: 10,
        color: this.config.theme === 'dark' ? '#e5e7eb' : '#374151'
      } : undefined,
      marker: {
        size: 8,
        color: groupValues, // Use raw numeric values
        colorscale: colorscale, // Use custom colorscale from palette
        cmin: min,
        cmax: max,
        showscale: true,
        colorbar: {
          title: {
            text: 'Value'
          } as any,
          thickness: 15,
          len: 0.9
        },
        opacity: 0.8
      }
    });

    return traces;
  }

  private getEllipseTrace(points: Point2D[], groupName: string, groupIndex: number): Data | null {
    if (points.length < 3) {
return null;
}

    try {
      const ellipseParams = calculateConfidenceEllipse(points, this.scoresConfig.ellipseConfidence!);
      const ellipsePath = generateEllipsePath(ellipseParams);

      return {
        type: 'scatter',
        mode: 'lines',
        x: ellipsePath.map(p => p.x),
        y: ellipsePath.map(p => p.y),
        line: {
          color: this.scoresConfig.colorScheme![groupIndex % this.scoresConfig.colorScheme!.length],
          width: 2,
          dash: 'dash'
        },
        showlegend: false,
        hoverinfo: 'skip',
        name: `${groupName} (${(this.scoresConfig.ellipseConfidence! * 100).toFixed(0)}% CI)`
      };
    } catch (error) {
      console.warn(`Failed to calculate ellipse for group ${groupName}:`, error);
      return null;
    }
  }

  private getDensityTraces(groups: string[], pc1: number, pc2: number): Data[] {
    const traces: Data[] = [];
    const uniqueGroups = Array.from(new Set(groups));

    uniqueGroups.forEach((group, groupIndex) => {
      const groupIndices = groups.map((g, i) => g === group ? i : -1).filter(i => i >= 0);
      const groupPoints: Point2D[] = groupIndices.map(i => ({
        x: this.data.scores[i][pc1],
        y: this.data.scores[i][pc2]
      }));

      if (groupPoints.length < 5) {
return;
} // Need enough points for KDE

      try {
        const kde = kernelDensityEstimate2D(groupPoints, 'scott', 30);

        traces.push({
          type: 'contour',
          x: kde.x,
          y: kde.y,
          z: kde.z,
          showscale: false,
          colorscale: [
            [0, 'rgba(255,255,255,0)'],
            [1, this.scoresConfig.colorScheme![groupIndex % this.scoresConfig.colorScheme!.length]]
          ],
          opacity: 0.2,
          contours: {
            coloring: 'heatmap',
            showlines: false
          },
          hoverinfo: 'skip',
          showlegend: false
        });
      } catch (error) {
        console.warn(`Failed to calculate density for group ${group}:`, error);
      }
    });

    return traces;
  }

  protected getLayout(): Partial<Layout> {
    const { explainedVariance, pc1 = 0, pc2 = 1 } = this.data;

    return {
      title: {
        text: 'PCA Scores Plot'
      },
      xaxis: {
        title: {
          text: `PC${pc1 + 1} (${explainedVariance[pc1].toFixed(1)}%)`
        },
        zeroline: true,
        zerolinecolor: 'rgba(128, 128, 128, 0.5)',
        gridcolor: 'rgba(128, 128, 128, 0.2)'
      },
      yaxis: {
        title: {
          text: `PC${pc2 + 1} (${explainedVariance[pc2].toFixed(1)}%)`
        },
        zeroline: true,
        zerolinecolor: 'rgba(128, 128, 128, 0.5)',
        gridcolor: 'rgba(128, 128, 128, 0.2)',
        scaleanchor: this.config.maintainAspectRatio ? 'x' : undefined,
        scaleratio: this.config.maintainAspectRatio ? 1 : undefined
      },
      hovermode: 'closest',
      dragmode: this.config.enableLasso ? 'lasso' : 'zoom'
    };
  }

  /**
   * Public method to get optimized traces based on data size
   */
  public getOptimizedTraces(): Data[] {
    const dataSize = this.getDataSize();
    const perfConfig = getOptimalConfig(
      dataSize,
      Array.from(new Set(this.data.groups)).length > 1,
      true
    );

    return perfConfig.useWebGL ? this.getWebGLTraces() : this.getStandardTraces();
  }

  /**
   * Public method to get the layout
   */
  public getPlotLayout(): Partial<Layout> {
    return this.getEnhancedLayout();
  }

  /**
   * Public method to get the config
   */
  public getPlotConfig(): Partial<Config> {
    const baseConfig = this.getAdvancedConfig();
    return {
      ...baseConfig,
      modeBarButtonsToAdd: getExportMenuItems() as any
    };
  }
}

/**
 * React component wrapper for PlotlyScoresPlot
 */
export const PCAScoresPlot: React.FC<{
  data: ScoresPlotData;
  config?: ScoresPlotConfig;
  onSelection?: (indices: number[]) => void;
}> = ({ data, config, onSelection }) => {
  const plot = useMemo(() => new PlotlyScoresPlot(data, config), [data, config]);

  const handleSelected = (event: any) => {
    if (onSelection && event?.points) {
      const indices = event.points.map((p: any) => p.pointIndex);
      onSelection(indices);
    }
  };

  return (
    <Plot
      data={plot.getOptimizedTraces()}
      layout={plot.getPlotLayout()}
      config={plot.getPlotConfig()}
      style={{ width: '100%', height: '100%' }}
      useResizeHandler={true}
      onSelected={handleSelected}
    />
  );
};