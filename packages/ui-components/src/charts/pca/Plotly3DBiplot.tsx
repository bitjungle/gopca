// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

// 3D Biplot combining scores and loading vectors in 3D space

import React, { useMemo } from 'react';
import { Data, Layout, Config } from 'plotly.js';
import { PCA_REFERENCES } from '../utils/plotlyMath';
import { getPlotlyTheme, mergeLayouts, ThemeMode } from '../utils/plotlyTheme';
import { getExportMenuItems } from '../utils/plotlyExport';
import { PLOT_CONFIG } from '../config/plotConfig';
import { PlotlyWithFullscreen } from '../utils/plotlyFullscreen';
import { getWatermarkDataUrlSync } from '../assets/watermark';

export interface Biplot3DData {
  scores: number[][];  // [n_samples][n_components]
  loadings: number[][];  // [n_components][n_variables]
  explainedVariance: number[];
  sampleNames?: string[];
  variableNames: string[];
  groups?: string[];
  groupValues?: number[]; // For continuous data
  groupType?: 'categorical' | 'continuous';
  pc1?: number;  // PC for X-axis (0-indexed)
  pc2?: number;  // PC for Y-axis (0-indexed)
  pc3?: number;  // PC for Z-axis (0-indexed)
}

export interface Biplot3DConfig {
  scalingType?: 'correlation' | 'symmetric' | 'pca';
  showScores?: boolean;
  showLoadings?: boolean;
  maxVariables?: number;  // Maximum number of loading vectors to display
  vectorScale?: number;  // Manual scaling adjustment
  colorScheme?: string[];
  markerSize?: number;
  opacity?: number;
  arrowSize?: number;  // Size of arrow cones
  arrowOpacity?: number;  // Opacity of arrow cones
  showProjections?: boolean;  // Show 2D projections on planes
  cameraPosition?: {
    eye: { x: number; y: number; z: number };
    center?: { x: number; y: number; z: number };
  };
  theme?: ThemeMode;
}

/**
 * 3D Biplot visualization combining PCA scores and loading vectors in 3D space
 * Shows relationships between samples and variables across three principal components
 * Reference: Gabriel (1971), extended to 3D space
 */
export class Plotly3DBiplot {
  private data: Biplot3DData;
  private config: Biplot3DConfig;
  private displayedVariables: number = 0;
  private totalVariables: number = 0;
  private needsFiltering: boolean = false;

  constructor(data: Biplot3DData, config?: Biplot3DConfig) {
    this.data = data;
    this.config = {
      scalingType: 'correlation',
      showScores: true,
      showLoadings: true,
      maxVariables: 50,  // Default to 50 for clarity in 3D
      vectorScale: 1.0,
      colorScheme: [
        '#1f77b4', '#ff7f0e', '#2ca02c', '#d62728', '#9467bd',
        '#8c564b', '#e377c2', '#7f7f7f', '#bcbd22', '#17becf'
      ],
      markerSize: 5,
      opacity: 0.8,
      arrowSize: 8,
      arrowOpacity: 0.7,
      showProjections: false,
      cameraPosition: {
        eye: { x: 1.5, y: 1.5, z: 1.5 },
        center: { x: 0, y: 0, z: 0 }
      },
      ...config
    };
  }

  private prepareData() {
    const { scores, loadings } = this.data;
    const pc1 = this.data.pc1 ?? 0;
    const pc2 = this.data.pc2 ?? 1;
    const pc3 = this.data.pc3 ?? 2;

    // Extract scores for selected PCs
    const scoresX = scores.map(row => row[pc1]);
    const scoresY = scores.map(row => row[pc2]);
    const scoresZ = scores.map(row => row[pc3]);

    // Calculate maximum loading magnitude for the selected components
    const loadingMagnitudes: number[] = [];
    for (let i = 0; i < loadings[pc1].length; i++) {
      const x = loadings[pc1][i];
      const y = loadings[pc2][i];
      const z = loadings[pc3][i];
      loadingMagnitudes.push(Math.sqrt(x * x + y * y + z * z));
    }
    const maxLoadingMagnitude = Math.max(...loadingMagnitudes);

    // Calculate score plot bounds
    const maxAbsScore = Math.max(
      ...scoresX.map(Math.abs),
      ...scoresY.map(Math.abs),
      ...scoresZ.map(Math.abs)
    );
    // Add 20% padding, but ensure minimum visibility
    const plotMax = Math.max(maxAbsScore * 1.2, 1.0);

    // Scale factor to make the largest loading vector reach 60% of plot bounds (less than 2D for clarity)
    const scaleFactor = maxLoadingMagnitude > 0 ? (plotMax * 0.6) / maxLoadingMagnitude : 1;

    // Apply scaling to loadings with optional manual adjustment
    const manualScale = this.config.vectorScale !== undefined ? this.config.vectorScale : 1.0;
    const totalScale = scaleFactor * manualScale;

    const loadingsX = loadings[pc1].map((v: number) => v * totalScale);
    const loadingsY = loadings[pc2].map((v: number) => v * totalScale);
    const loadingsZ = loadings[pc3].map((v: number) => v * totalScale);

    // Store loading magnitudes and filtering info
    this.totalVariables = loadings[pc1].length;
    
    // Filter to top N vectors by magnitude if needed
    const allVectors = this.data.variableNames.map((name, i) => {
      const magnitude = Math.sqrt(
        loadingsX[i] ** 2 + loadingsY[i] ** 2 + loadingsZ[i] ** 2
      );
      return { name, i, magnitude };
    });

    this.needsFiltering = allVectors.length > (this.config.maxVariables || 50);
    
    let validVectors = allVectors;
    if (this.needsFiltering) {
      validVectors = [...allVectors]
        .sort((a, b) => b.magnitude - a.magnitude)
        .slice(0, this.config.maxVariables || 50);
    }

    // Filter out very small vectors
    const minMagnitude = 0.01;
    validVectors = validVectors.filter(v => v.magnitude >= minMagnitude);
    this.displayedVariables = validVectors.length;

    return { 
      scoresX, scoresY, scoresZ, 
      loadingsX, loadingsY, loadingsZ,
      pc1, pc2, pc3,
      validVectors
    };
  }

  getTraces(): Data[] {
    const traces: Data[] = [];
    const { 
      scoresX, scoresY, scoresZ, 
      loadingsX, loadingsY, loadingsZ,
      pc1, pc2, pc3,
      validVectors
    } = this.prepareData();
    const { groups, groupValues, groupType, sampleNames } = this.data;

    // Add scores scatter plot
    if (this.config.showScores) {
      if (groupType === 'continuous' && groupValues) {
        // Continuous coloring
        const validValues = groupValues.filter(v => v !== null && v !== undefined && !isNaN(v) && isFinite(v));
        const min = Math.min(...validValues);
        const max = Math.max(...validValues);

        // Create a custom colorscale from the palette
        const palette = this.config.colorScheme || ['#440154', '#31688e', '#35b779', '#fde725'];
        const colorscale: [number, string][] = palette.map((color, i) => [
          i / (palette.length - 1),
          color
        ]);

        // Prepare hover text
        const hovertext = scoresX.map((x, i) => {
          const label = sampleNames?.[i] || `Sample ${i}`;
          const value = groupValues[i];
          const valueStr = value !== null && value !== undefined && !isNaN(value) && isFinite(value)
            ? value.toFixed(2)
            : 'Missing';
          return `<b>${label}</b><br>Value: ${valueStr}<br>PC${pc1 + 1}: ${x.toFixed(2)}<br>PC${pc2 + 1}: ${scoresY[i].toFixed(2)}<br>PC${pc3 + 1}: ${scoresZ[i].toFixed(2)}`;
        });

        traces.push({
          type: 'scatter3d',
          mode: 'markers',
          x: scoresX,
          y: scoresY,
          z: scoresZ,
          name: 'Scores',
          hovertext: hovertext,
          hovertemplate: '%{hovertext}<extra></extra>',
          marker: {
            size: this.config.markerSize,
            color: groupValues,
            colorscale: colorscale,
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
            opacity: this.config.opacity
          }
        });
      } else if (groups) {
        // Group by categories
        const uniqueGroups = Array.from(new Set(groups));
        uniqueGroups.forEach((group, groupIndex) => {
          const indices = groups.map((g, idx) => g === group ? idx : -1).filter(idx => idx >= 0);

          const groupX = indices.map((idx: number) => scoresX[idx]);
          const groupY = indices.map((idx: number) => scoresY[idx]);
          const groupZ = indices.map((idx: number) => scoresZ[idx]);

          // Prepare hover text
          const hovertext = indices.map(i => {
            const label = sampleNames?.[i] || `Sample ${i}`;
            return `<b>${label}</b><br>Group: ${group}<br>PC${pc1 + 1}: ${scoresX[i].toFixed(2)}<br>PC${pc2 + 1}: ${scoresY[i].toFixed(2)}<br>PC${pc3 + 1}: ${scoresZ[i].toFixed(2)}`;
          });

          traces.push({
            type: 'scatter3d',
            mode: 'markers',
            name: group,
            x: groupX,
            y: groupY,
            z: groupZ,
            hovertext: hovertext,
            hovertemplate: '%{hovertext}<extra></extra>',
            marker: {
              size: this.config.markerSize,
              color: this.config.colorScheme![groupIndex % this.config.colorScheme!.length],
              opacity: this.config.opacity
            }
          });
        });
      } else {
        // Single group
        const hovertext = scoresX.map((x, i) => {
          const label = sampleNames?.[i] || `Sample ${i}`;
          return `<b>${label}</b><br>PC${pc1 + 1}: ${x.toFixed(2)}<br>PC${pc2 + 1}: ${scoresY[i].toFixed(2)}<br>PC${pc3 + 1}: ${scoresZ[i].toFixed(2)}`;
        });

        traces.push({
          type: 'scatter3d',
          mode: 'markers',
          x: scoresX,
          y: scoresY,
          z: scoresZ,
          name: 'Scores',
          hovertext: hovertext,
          hovertemplate: '%{hovertext}<extra></extra>',
          marker: {
            size: this.config.markerSize,
            color: this.config.colorScheme![0],
            opacity: this.config.opacity
          }
        });
      }

      // Add projection traces if enabled
      if (this.config.showProjections) {
        traces.push(...this.getProjectionTraces(scoresX, scoresY, scoresZ, groups));
      }
    }

    // Add loading vectors
    if (this.config.showLoadings && validVectors.length > 0) {
      // Add origin point
      traces.push({
        type: 'scatter3d',
        mode: 'markers',
        x: [0],
        y: [0],
        z: [0],
        marker: {
          symbol: 'circle',
          size: 8,
          color: this.config.colorScheme![1] || '#ef4444',
          opacity: 0.5
        },
        name: 'Origin',
        showlegend: false,
        hoverinfo: 'skip'
      });

      // Add loading vectors as individual lines for better coloring
      validVectors.forEach(v => {
        const vectorColor = this.config.colorScheme![1] || '#ef4444';
        
        traces.push({
          type: 'scatter3d',
          mode: 'lines',
          x: [0, loadingsX[v.i]],
          y: [0, loadingsY[v.i]],
          z: [0, loadingsZ[v.i]],
          line: {
            color: vectorColor,
            width: 4
          },
          showlegend: false,
          hoverinfo: 'skip'
        });
      });

      // Add endpoints as markers (arrowheads)
      const endpointX: number[] = [];
      const endpointY: number[] = [];
      const endpointZ: number[] = [];
      const hovertext: string[] = [];

      validVectors.forEach(v => {
        endpointX.push(loadingsX[v.i]);
        endpointY.push(loadingsY[v.i]);
        endpointZ.push(loadingsZ[v.i]);
        
        hovertext.push(
          `<b>${v.name}</b><br>` +
          `PC${pc1 + 1}: ${loadingsX[v.i].toFixed(3)}<br>` +
          `PC${pc2 + 1}: ${loadingsY[v.i].toFixed(3)}<br>` +
          `PC${pc3 + 1}: ${loadingsZ[v.i].toFixed(3)}<br>` +
          `Magnitude: ${v.magnitude.toFixed(3)}`
        );
      });

      traces.push({
        type: 'scatter3d',
        mode: 'markers',
        x: endpointX,
        y: endpointY,
        z: endpointZ,
        marker: {
          symbol: 'diamond',
          size: 6,
          color: this.config.colorScheme![1],
          opacity: 0.9
        },
        name: 'Loading Vectors',
        showlegend: false,
        hovertext: hovertext,
        hovertemplate: '%{hovertext}<extra></extra>'
      });

      // Add text labels for variables
      const labelPositions = validVectors.map(v => {
        // Position labels slightly beyond the arrow tip
        const scale = 1.1;
        return {
          x: loadingsX[v.i] * scale,
          y: loadingsY[v.i] * scale,
          z: loadingsZ[v.i] * scale,
          text: v.name
        };
      });

      traces.push({
        type: 'scatter3d',
        mode: 'text',
        x: labelPositions.map(p => p.x),
        y: labelPositions.map(p => p.y),
        z: labelPositions.map(p => p.z),
        text: labelPositions.map(p => p.text),
        textfont: {
          size: 10,
          color: this.config.colorScheme![1]
        },
        showlegend: false,
        hoverinfo: 'skip'
      });
    }

    return traces;
  }

  private getProjectionTraces(scoresX: number[], scoresY: number[], scoresZ: number[], groups?: string[]): Data[] {
    const projectionTraces: Data[] = [];
    const uniqueGroups = groups ? Array.from(new Set(groups)) : ['All'];

    uniqueGroups.forEach((group, groupIndex) => {
      let groupX: number[], groupY: number[], groupZ: number[];
      
      if (groups) {
        const indices = groups.map((g, idx) => g === group ? idx : -1).filter(idx => idx >= 0);
        groupX = indices.map(i => scoresX[i]);
        groupY = indices.map(i => scoresY[i]);
        groupZ = indices.map(i => scoresZ[i]);
      } else {
        groupX = scoresX;
        groupY = scoresY;
        groupZ = scoresZ;
      }

      const color = this.config.colorScheme![groupIndex % this.config.colorScheme!.length];

      // XY plane projection (z=min)
      const minZ = Math.min(...scoresZ);
      projectionTraces.push({
        type: 'scatter3d',
        mode: 'markers',
        x: groupX,
        y: groupY,
        z: groupX.map(() => minZ),
        marker: {
          size: 2,
          color: color,
          opacity: 0.3
        },
        showlegend: false,
        hoverinfo: 'skip'
      });

      // XZ plane projection (y=min)
      const minY = Math.min(...scoresY);
      projectionTraces.push({
        type: 'scatter3d',
        mode: 'markers',
        x: groupX,
        y: groupX.map(() => minY),
        z: groupZ,
        marker: {
          size: 2,
          color: color,
          opacity: 0.3
        },
        showlegend: false,
        hoverinfo: 'skip'
      });

      // YZ plane projection (x=min)
      const minX = Math.min(...scoresX);
      projectionTraces.push({
        type: 'scatter3d',
        mode: 'markers',
        x: groupX.map(() => minX),
        y: groupY,
        z: groupZ,
        marker: {
          size: 2,
          color: color,
          opacity: 0.3
        },
        showlegend: false,
        hoverinfo: 'skip'
      });
    });

    return projectionTraces;
  }

  getEnhancedLayout(): Partial<Layout> {
    const baseLayout = this.getLayout();
    const themeLayout = getPlotlyTheme(this.config.theme || 'light').layout;
    
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
    const pc1 = this.data.pc1 ?? 0;
    const pc2 = this.data.pc2 ?? 1;
    const pc3 = this.data.pc3 ?? 2;
    
    // Theme-aware colors for 3D scene
    const isDark = this.config.theme === 'dark';
    const sceneColors = {
      backgroundcolor: isDark ? 'rgba(31, 41, 55, 0.5)' : 'rgba(230, 230, 230, 0.5)',
      gridcolor: isDark ? 'rgba(75, 85, 99, 0.5)' : 'rgba(200, 200, 200, 0.5)',
      zerolinecolor: isDark ? 'rgba(107, 114, 128, 0.8)' : 'rgba(128, 128, 128, 0.8)'
    };

    // Create title with filtering indicator if needed
    let titleText = `3D Biplot (${this.config.scalingType} scaling)`;
    if (this.needsFiltering) {
      titleText += `<br><span style="font-size: 12px; color: #f59e0b;">Showing top ${this.displayedVariables} of ${this.totalVariables} variables</span>`;
    }

    return {
      title: {
        text: titleText
      },
      scene: {
        xaxis: {
          title: {
            text: `PC${pc1 + 1} (${explainedVariance[pc1].toFixed(1)}%)`
          },
          backgroundcolor: sceneColors.backgroundcolor,
          gridcolor: sceneColors.gridcolor,
          showbackground: true,
          zerolinecolor: sceneColors.zerolinecolor,
          zeroline: true
        },
        yaxis: {
          title: {
            text: `PC${pc2 + 1} (${explainedVariance[pc2].toFixed(1)}%)`
          },
          backgroundcolor: sceneColors.backgroundcolor,
          gridcolor: sceneColors.gridcolor,
          showbackground: true,
          zerolinecolor: sceneColors.zerolinecolor,
          zeroline: true
        },
        zaxis: {
          title: {
            text: `PC${pc3 + 1} (${explainedVariance[pc3].toFixed(1)}%)`
          },
          backgroundcolor: sceneColors.backgroundcolor,
          gridcolor: sceneColors.gridcolor,
          showbackground: true,
          zerolinecolor: sceneColors.zerolinecolor,
          zeroline: true
        },
        camera: this.config.cameraPosition,
        aspectmode: 'cube',
        hovermode: 'closest'
      },
      showlegend: true,
      legend: {
        borderwidth: 1,
        font: { size: 12 },
        x: 1.02,
        y: 1,
        xanchor: 'left',
        yanchor: 'top'
      },
      annotations: [
        {
          text: `References: ${PCA_REFERENCES.map(r => `${r.authors} (${r.year})`).join(', ')}`,
          xref: 'paper',
          yref: 'paper',
          x: 0,
          y: -0.1,
          showarrow: false,
          font: { size: 10, color: 'gray' }
        }
      ]
    };
  }

  getConfig(): Partial<Config> {
    return {
      responsive: true,
      displaylogo: false,
      modeBarButtonsToAdd: getExportMenuItems() as any,
      toImageButtonOptions: {
        ...PLOT_CONFIG.export.presentation,
        filename: 'pca-3d-biplot'
      }
    };
  }
}

/**
 * React component wrapper for 3D Biplot
 */
export const PCA3DBiplot: React.FC<{
  data: Biplot3DData;
  config?: Biplot3DConfig;
}> = ({ data, config }) => {
  const plot = useMemo(() => new Plotly3DBiplot(data, config), [data, config]);

  return (
    <PlotlyWithFullscreen
      data={plot.getTraces()}
      layout={plot.getEnhancedLayout()}
      config={plot.getConfig()}
      style={{ width: '100%', height: '100%' }}
    />
  );
};