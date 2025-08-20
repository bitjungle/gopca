// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Biplot combining scores and loading vectors

import React, { useMemo } from 'react';
import Plot from 'react-plotly.js';
import { Data, Layout } from 'plotly.js';
import { 
  selectSmartLabels,
  calculateConfidenceEllipse,
  generateEllipsePath,
  Point2D
} from '../utils/plotlyMath';
import { getPlotlyTheme, mergeLayouts, ThemeMode } from '../utils/plotlyTheme';
import { getExportMenuItems } from '../utils/plotlyExport';

export interface BiplotData {
  scores: number[][];  // [n_samples][n_components]
  loadings: number[][];  // [n_components][n_variables]
  explainedVariance: number[];
  sampleNames?: string[];
  variableNames: string[];
  groups?: string[];
  groupValues?: number[]; // For continuous data
  groupType?: 'categorical' | 'continuous';
}

export interface BiplotConfig {
  pcX?: number;  // PC for X-axis (1-indexed)
  pcY?: number;  // PC for Y-axis (1-indexed)
  scalingType?: 'correlation' | 'symmetric' | 'pca';
  showScores?: boolean;
  showLoadings?: boolean;
  showLabels?: boolean;
  labelThreshold?: number;
  vectorScale?: number;
  colorScheme?: string[];
  pointSize?: number;
  arrowWidth?: number;
  theme?: ThemeMode;
  showEllipses?: boolean;
  ellipseConfidence?: number;
  maxVariables?: number;  // Maximum number of loading vectors to display
}

/**
 * Biplot visualization combining PCA scores and loading vectors
 * Reference: Gabriel (1971), "The biplot graphic display of matrices with application to principal component analysis"
 */
export class PlotlyBiplot {
  private data: BiplotData;
  private config: BiplotConfig;
  
  constructor(data: BiplotData, config?: BiplotConfig) {
    this.data = data;
    this.config = {
      pcX: 1,
      pcY: 2,
      scalingType: 'correlation',
      showScores: true,
      showLoadings: true,
      showLabels: true,
      labelThreshold: 20,
      vectorScale: 1.0,
      pointSize: 8,
      arrowWidth: 2,
      theme: 'light',
      showEllipses: false,
      ellipseConfidence: 0.95,
      maxVariables: 100,
      ...config
    };
  }
  
  private prepareData() {
    const { scores, loadings } = this.data;
    const pcX = (this.config.pcX || 1) - 1;
    const pcY = (this.config.pcY || 2) - 1;
    
    // Extract scores for selected PCs
    const scoresX = scores.map(row => row[pcX]);
    const scoresY = scores.map(row => row[pcY]);
    
    // Calculate maximum loading magnitude for the selected components
    const loadingMagnitudes: number[] = [];
    for (let i = 0; i < loadings[pcX].length; i++) {
      const x = loadings[pcX][i];
      const y = loadings[pcY][i];
      loadingMagnitudes.push(Math.sqrt(x * x + y * y));
    }
    const maxLoadingMagnitude = Math.max(...loadingMagnitudes);
    
    // Calculate score plot bounds
    const maxAbsScore = Math.max(
      ...scoresX.map(Math.abs),
      ...scoresY.map(Math.abs)
    );
    // Add 20% padding, but ensure minimum visibility
    const plotMax = Math.max(maxAbsScore * 1.2, 1.0);
    
    // Scale factor to make the largest loading vector reach 70% of plot bounds
    // This is the same approach as the working Recharts implementation
    const scaleFactor = maxLoadingMagnitude > 0 ? (plotMax * 0.7) / maxLoadingMagnitude : 1;
    
    // Apply scaling to loadings with optional manual adjustment
    const manualScale = this.config.vectorScale !== undefined ? this.config.vectorScale : 1.0;
    const totalScale = scaleFactor * manualScale;
    
    const loadingsX = loadings[pcX].map((v: number) => v * totalScale);
    const loadingsY = loadings[pcY].map((v: number) => v * totalScale);
    
    return { scoresX, scoresY, loadingsX, loadingsY, pcX, pcY };
  }
  
  getTraces(): Data[] {
    const traces: Data[] = [];
    const { scoresX, scoresY, loadingsX, loadingsY, pcX, pcY } = this.prepareData();
    const { groups, groupValues, groupType, sampleNames, variableNames } = this.data;
    
    // Add scores scatter plot
    if (this.config.showScores) {
      // Handle continuous vs categorical data
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
          return `<b>${label}</b><br>Value: ${valueStr}<br>PC${pcX + 1}: ${x.toFixed(2)}<br>PC${pcY + 1}: ${scoresY[i].toFixed(2)}`;
        });
        
        traces.push({
          type: 'scatter',
          mode: 'markers',
          x: scoresX,
          y: scoresY,
          name: 'Scores',
          hovertext: hovertext,
          hovertemplate: '%{hovertext}<extra></extra>',
          marker: {
            size: this.config.pointSize,
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
            opacity: 0.7
          }
        });
      } else if (groups) {
        // Group by categories
        const uniqueGroups = Array.from(new Set(groups));
        uniqueGroups.forEach((group, i) => {
          const indices = groups.map((g, idx) => g === group ? idx : -1).filter(idx => idx >= 0);
          
          const groupX = indices.map((idx: number) => scoresX[idx]);
          const groupY = indices.map((idx: number) => scoresY[idx]);
          
          traces.push({
            type: 'scatter',
            mode: 'markers',
            x: groupX,
            y: groupY,
            name: group,
            marker: {
              color: this.config.colorScheme
                ? this.config.colorScheme[i % this.config.colorScheme.length]
                : ['#3b82f6', '#ef4444', '#10b981', '#f59e0b', '#8b5cf6'][i % 5],
              size: this.config.pointSize,
              opacity: 0.7
            },
            text: sampleNames ? indices.map((idx: number) => sampleNames[idx]) : undefined,
            hovertemplate: '<b>%{text}</b><br>PC' + (pcX + 1) + ': %{x:.2f}<br>PC' + 
                          (pcY + 1) + ': %{y:.2f}<extra></extra>'
          });
          
          // Add confidence ellipse if enabled
          if (this.config.showEllipses && groupX.length > 2) {
            const points: Point2D[] = groupX.map((x, idx) => ({ x, y: groupY[idx] }));
            try {
              const ellipseParams = calculateConfidenceEllipse(points, this.config.ellipseConfidence || 0.95);
              const ellipsePath = generateEllipsePath(ellipseParams);
              
              traces.push({
                type: 'scatter',
                mode: 'lines',
                x: ellipsePath.map(p => p.x),
                y: ellipsePath.map(p => p.y),
                line: {
                  color: this.config.colorScheme
                    ? this.config.colorScheme[i % this.config.colorScheme.length]
                    : ['#3b82f6', '#ef4444', '#10b981', '#f59e0b', '#8b5cf6'][i % 5],
                  width: 2,
                  dash: 'dash'
                },
                showlegend: false,
                hoverinfo: 'skip',
                name: `${group} (${(this.config.ellipseConfidence! * 100).toFixed(0)}% CI)`
              });
            } catch (error) {
              console.warn(`Failed to calculate ellipse for group ${group}:`, error);
            }
          }
        });
      } else {
        // Single group
        traces.push({
          type: 'scatter',
          mode: 'markers',
          x: scoresX,
          y: scoresY,
          name: 'Scores',
          marker: {
            color: this.config.colorScheme?.[0] || '#3b82f6',
            size: this.config.pointSize,
            opacity: 0.7
          },
          text: sampleNames,
          hovertemplate: '<b>%{text}</b><br>PC' + (pcX + 1) + ': %{x:.2f}<br>PC' + 
                        (pcY + 1) + ': %{y:.2f}<extra></extra>'
        });
      }
      
      // Add smart labels for scores
      if (this.config.showLabels && sampleNames) {
        const scorePoints = scoresX.map((x, i) => ({ x, y: scoresY[i] }));
        const selectedIndices = selectSmartLabels(
          scorePoints,
          this.config.labelThreshold || 20
        );
        
        traces.push({
          type: 'scatter',
          mode: 'text',
          x: selectedIndices.map(i => scoresX[i]),
          y: selectedIndices.map(i => scoresY[i]),
          text: selectedIndices.map(i => sampleNames[i]),
          textposition: 'top center',
          textfont: {
            size: 10,
            color: this.config.theme === 'dark' ? '#e5e7eb' : '#374151'
          },
          showlegend: false,
          hoverinfo: 'skip'
        });
      }
    }
    
    // Add loading vectors
    if (this.config.showLoadings) {
      // Calculate magnitude for all vectors
      const allVectors = variableNames.map((name, i) => {
        const magnitude = Math.sqrt(loadingsX[i]**2 + loadingsY[i]**2);
        return { name, i, magnitude };
      });
      
      // Check if we need to filter based on maxVariables
      const needsFiltering = allVectors.length > (this.config.maxVariables || 100);
      
      // Filter to top N vectors by magnitude if needed
      let validVectors = allVectors;
      if (needsFiltering) {
        validVectors = [...allVectors]
          .sort((a, b) => b.magnitude - a.magnitude)
          .slice(0, this.config.maxVariables || 100);
      }
      
      // Filter out very small vectors
      const minMagnitude = 0.01;
      validVectors = validVectors.filter(v => v.magnitude >= minMagnitude);
      
      // Store for later use in title
      (this as any)._needsFiltering = needsFiltering;
      (this as any)._totalVariables = allVectors.length;
      (this as any)._displayedVariables = validVectors.length;
      
      // Add all loading vectors as a single trace for better performance
      if (validVectors.length > 0) {
        const vectorX: number[] = [];
        const vectorY: number[] = [];
        const vectorText: string[] = [];
        
        validVectors.forEach(v => {
          if (!v) return;
          // Add line from origin to loading point
          vectorX.push(0, loadingsX[v.i]);
          vectorX.push(null as any);  // null creates line break
          vectorY.push(0, loadingsY[v.i]);
          vectorY.push(null as any);
          vectorText.push('', v.name, '');
        });
        
        // Add loading vectors as lines
        traces.push({
          type: 'scatter',
          mode: 'lines',
          x: vectorX,
          y: vectorY,
          line: {
            color: this.config.colorScheme?.[1] || '#ef4444',
            width: this.config.arrowWidth || 2
          },
          name: 'Loadings',
          showlegend: false,
          hoverinfo: 'skip'
        });
        
        // Add arrowheads as markers at the end of vectors
        traces.push({
          type: 'scatter',
          mode: 'markers',
          x: validVectors.map(v => v ? loadingsX[v.i] : 0),
          y: validVectors.map(v => v ? loadingsY[v.i] : 0),
          marker: {
            symbol: 'arrow',
            size: 12,
            color: this.config.colorScheme?.[1] || '#ef4444'
          } as any,
          showlegend: false,
          hovertemplate: validVectors.map(v => 
            `<b>${v?.name}</b><br>Loading X: %{x:.3f}<br>Loading Y: %{y:.3f}<extra></extra>`
          )
        });
        
        // Add text labels for variables
        const labelPositions = validVectors.map(v => {
          if (!v) return { x: 0, y: 0 };
          // Position labels slightly beyond the arrow tip
          const scale = 1.15;
          return {
            x: loadingsX[v.i] * scale,
            y: loadingsY[v.i] * scale
          };
        });
        
        traces.push({
          type: 'scatter',
          mode: 'text',
          x: labelPositions.map(p => p.x),
          y: labelPositions.map(p => p.y),
          text: validVectors.map(v => v?.name || ''),
          textposition: 'middle center',
          textfont: {
            size: 10,
            color: this.config.colorScheme?.[1] || '#ef4444'
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
    const { pcX, pcY, scoresX, scoresY, loadingsX, loadingsY } = this.prepareData();
    const { explainedVariance } = this.data;
    
    // Calculate axis ranges to accommodate both scores and loadings
    const allX = [...scoresX, ...loadingsX, 0];
    const allY = [...scoresY, ...loadingsY, 0];
    const xRange = [Math.min(...allX) * 1.2, Math.max(...allX) * 1.2];
    const yRange = [Math.min(...allY) * 1.2, Math.max(...allY) * 1.2];
    
    // Create title with filtering indicator if needed
    let titleText = `Biplot (${this.config.scalingType} scaling)`;
    if ((this as any)._needsFiltering) {
      titleText += `<br><span style="font-size: 12px; color: #f59e0b;">Showing top ${(this as any)._displayedVariables} of ${(this as any)._totalVariables} variables</span>`;
    }
    
    const layout: Partial<Layout> = {
      title: {
        text: titleText
      },
      xaxis: {
        title: {
          text: `PC${pcX + 1} (${explainedVariance[pcX].toFixed(1)}%)`
        },
        zeroline: true,
        zerolinewidth: 1,
        zerolinecolor: 'gray',
        showgrid: true,
        gridcolor: 'rgba(128, 128, 128, 0.2)',
        range: xRange
      },
      yaxis: {
        title: {
          text: `PC${pcY + 1} (${explainedVariance[pcY].toFixed(1)}%)`
        },
        zeroline: true,
        zerolinewidth: 1,
        zerolinecolor: 'gray',
        showgrid: true,
        gridcolor: 'rgba(128, 128, 128, 0.2)',
        range: yRange,
        scaleanchor: 'x',
        scaleratio: 1
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
      annotations: []
    };
    
    return layout;
  }
  
  getConfig(): Partial<any> {
    return {
      responsive: true,
      displaylogo: false,
      modeBarButtonsToAdd: getExportMenuItems() as any,
      toImageButtonOptions: {
        format: 'png',
        filename: 'biplot',
        height: 1600,
        width: 1600,
        scale: 2
      }
    };
  }
}

/**
 * React component wrapper for Biplot
 */
export const PCABiplot: React.FC<{
  data: BiplotData;
  config?: BiplotConfig;
}> = ({ data, config }) => {
  const plot = useMemo(() => new PlotlyBiplot(data, config), [data, config]);
  
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