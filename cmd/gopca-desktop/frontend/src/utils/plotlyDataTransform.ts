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

// Utility functions to transform GoPCA data to Plotly component formats

import { PCAResult, EllipseParams, SampleMetrics } from '../types';
import type {
  ScoresPlotData,
  ScoresPlotConfig,
  Scores3DPlotData,
  Scores3DPlotConfig,
  ScreePlotData,
  ScreePlotConfig,
  LoadingsPlotData,
  LoadingsPlotConfig,
  BiplotData,
  BiplotConfig,
  Biplot3DData,
  Biplot3DConfig,
  CircleOfCorrelationsData,
  CircleOfCorrelationsConfig,
  DiagnosticPlotData,
  DiagnosticPlotConfig,
  EigencorrelationPlotData,
  EigencorrelationPlotConfig,
  TemporalLoadingsPlotData,
  TemporalLoadingsPlotConfig,
  TemporalVariableImportanceData,
  TemporalVariableImportancePlotConfig
} from '@gopca/ui-components';

/**
 * Helper function to transpose a matrix
 * Converts [rows][cols] to [cols][rows]
 */
function transposeMatrix(matrix: number[][]): number[][] {
  if (!matrix || matrix.length === 0) {
return [];
}
  const rows = matrix.length;
  const cols = matrix[0].length;
  const transposed: number[][] = Array(cols).fill(null).map(() => Array(rows));

  for (let i = 0; i < rows; i++) {
    for (let j = 0; j < cols; j++) {
      transposed[j][i] = matrix[i][j];
    }
  }

  return transposed;
}

/**
 * Shared config builder for all Plotly visualizations
 * Provides common configuration properties to ensure consistency
 */
function createBaseVisualizationConfig(
  theme?: 'light' | 'dark',
  colorScheme?: string[],
  fontScale?: number
): { theme?: 'light' | 'dark'; colorScheme?: string[]; fontScale?: number } {
  return {
    theme,
    colorScheme,
    fontScale
  };
}

/**
 * Transform PCAResult to Plotly ScoresPlot data format
 */
export function transformToScoresPlotData(
  pcaResult: PCAResult,
  rowNames: string[],
  groupLabels?: string[],
  groupValues?: number[],
  groupType?: 'categorical' | 'continuous',
  xComponent: number = 0,
  yComponent: number = 1
): ScoresPlotData {
  return {
    scores: pcaResult.scores,
    sampleNames: rowNames,
    groups: groupLabels || [],
    groupValues,
    groupType,
    explainedVariance: pcaResult.explained_variance_ratio, // Already in percentages from backend
    pc1: xComponent,
    pc2: yComponent
  };
}

/**
 * Create ScoresPlot config from GoPCA props
 */
export function createScoresPlotConfig(
  xComponent: number = 0,
  yComponent: number = 1,
  showEllipses?: boolean,
  confidenceLevel?: 0.90 | 0.95 | 0.99,
  showRowLabels?: boolean,
  maxLabelsToShow?: number,
  theme?: 'light' | 'dark',
  colorScheme?: string[],
  fontScale?: number
): ScoresPlotConfig {
  return {
    showEllipses,
    ellipseConfidence: confidenceLevel,
    showSmartLabels: showRowLabels,
    maxLabels: maxLabelsToShow,
    theme,
    colorScheme,
    fontScale
  };
}

/**
 * Transform PCAResult to Plotly 3D ScoresPlot data format
 */
export function transformToScores3DPlotData(
  pcaResult: PCAResult,
  rowNames: string[],
  groupLabels?: string[],
  groupValues?: number[],
  groupType?: 'categorical' | 'continuous',
  xComponent: number = 0,
  yComponent: number = 1,
  zComponent: number = 2
): Scores3DPlotData {
  // Ensure we always have groups - if none provided, create a single default group
  const groups = groupLabels && groupLabels.length > 0
    ? groupLabels
    : Array(pcaResult.scores.length).fill('All samples');

  return {
    scores: pcaResult.scores,
    sampleNames: rowNames,
    groups: groups,
    groupValues,
    groupType,
    explainedVariance: pcaResult.explained_variance_ratio,
    pc1: xComponent,
    pc2: yComponent,
    pc3: zComponent
  };
}

/**
 * Create 3D ScoresPlot config from GoPCA props
 */
export function createScores3DPlotConfig(
  _xComponent: number = 0,
  _yComponent: number = 1,
  _zComponent: number = 2,
  showRowLabels?: boolean,
  maxLabelsToShow?: number,
  theme?: 'light' | 'dark',
  colorScheme?: string[],
  fontScale?: number
): Scores3DPlotConfig {
  return {
    colorScheme,
    markerSize: 5,
    opacity: 0.8,
    showProjections: false,
    theme,
    showLabels: showRowLabels || false,
    maxLabels: maxLabelsToShow || 10,
    fontScale
  };
}

/**
 * Transform PCAResult to Plotly ScreePlot data format
 */
export function transformToScreePlotData(pcaResult: PCAResult): ScreePlotData {
  return {
    explainedVariance: pcaResult.explained_variance_ratio, // Already in percentages from backend
    cumulativeVariance: pcaResult.cumulative_variance
    // eigenvalues could be calculated if needed
  };
}

/**
 * Create ScreePlot config
 */
export function createScreePlotConfig(
  showCumulative: boolean = true,
  elbowThreshold: number = 80,
  theme?: 'light' | 'dark',
  colorScheme?: string[],
  fontScale?: number
): ScreePlotConfig {
  return {
    showCumulativeLine: showCumulative,
    showThresholdLine: true,
    thresholdValue: elbowThreshold,
    theme,
    colorScheme,
    fontScale
  };
}

/**
 * Transform PCAResult to Plotly LoadingsPlot data format
 */
export function transformToLoadingsPlotData(
  pcaResult: PCAResult,
  selectedComponent: number = 0
): LoadingsPlotData {
  // Backend stores loadings as [variables][components], but frontend expects [components][variables]
  const transposedLoadings = transposeMatrix(pcaResult.loadings);

  return {
    loadings: transposedLoadings,
    variableNames: pcaResult.variable_labels ||
      Array.from({ length: pcaResult.loadings.length }, (_, i) => `Var${i + 1}`), // Use loadings.length for number of variables
    componentIndex: selectedComponent
  };
}

/**
 * Create LoadingsPlot config
 */
export function createLoadingsPlotConfig(
  plotType: 'bar' | 'line' = 'bar',
  sortByMagnitude: boolean = false,
  theme?: 'light' | 'dark',
  colorScheme?: string[],
  numVariables?: number,
  variableThreshold?: number,
  fontScale?: number
): LoadingsPlotConfig {
  // Determine whether to show markers in line mode
  // When we have many variables (above threshold), don't show markers for cleaner visualization
  let showMarkers = true; // Default to showing markers
  if (plotType === 'line' && numVariables !== undefined && variableThreshold !== undefined) {
    showMarkers = numVariables <= variableThreshold;
  }

  return {
    mode: plotType,
    sortByMagnitude,
    showThreshold: true,
    thresholdValue: 0.3,
    showMarkers,
    // Don't set maxVariables - show all by default
    theme,
    colorScheme,
    fontScale
  };
}

/**
 * Transform PCAResult to Plotly Biplot data format
 */
export function transformToBiplotData(
  pcaResult: PCAResult,
  rowNames: string[],
  groupLabels?: string[],
  groupValues?: number[],
  groupType?: 'categorical' | 'continuous'
): BiplotData {
  // Check if loadings exist (e.g., not available for Kernel PCA)
  if (!pcaResult.loadings || pcaResult.loadings.length === 0) {
    throw new Error('Biplot visualization requires loadings data, which is not available for this PCA method.');
  }

  // Backend stores loadings as [variables][components], but frontend expects [components][variables]
  const transposedLoadings = transposeMatrix(pcaResult.loadings);

  return {
    scores: pcaResult.scores,
    loadings: transposedLoadings,
    explainedVariance: pcaResult.explained_variance_ratio, // Already in percentages from backend
    sampleNames: rowNames,
    variableNames: pcaResult.variable_labels ||
      Array.from({ length: pcaResult.loadings.length }, (_, i) => `Var${i + 1}`), // Use loadings.length for number of variables
    groups: groupLabels,
    groupValues,
    groupType
  };
}

/**
 * Create Biplot config
 */
export function createBiplotConfig(
  xComponent: number = 0,
  yComponent: number = 1,
  showLabels: boolean = true,
  theme?: 'light' | 'dark',
  colorScheme?: string[],
  showEllipses: boolean = false,
  ellipseConfidence: number = 0.95,
  fontScale?: number
): BiplotConfig {
  return {
    pcX: xComponent + 1,
    pcY: yComponent + 1,
    scalingType: 'correlation',
    showScores: true,
    showLoadings: true,
    showLabels,
    labelThreshold: 20,
    vectorScale: 1.0,
    showEllipses,
    ellipseConfidence,
    ...createBaseVisualizationConfig(theme, colorScheme, fontScale)
  };
}

/**
 * Transform PCAResult to Circle of Correlations data
 */
export function transformToCircleOfCorrelationsData(
  pcaResult: PCAResult
): CircleOfCorrelationsData {
  // Check if loadings exist (e.g., not available for Kernel PCA)
  if (!pcaResult.loadings || pcaResult.loadings.length === 0) {
    throw new Error('Circle of Correlations visualization requires loadings data, which is not available for this PCA method.');
  }

  // Backend stores loadings as [variables][components], but frontend expects [components][variables]
  const transposedLoadings = transposeMatrix(pcaResult.loadings);

  return {
    loadings: transposedLoadings,
    variableNames: pcaResult.variable_labels ||
      Array.from({ length: pcaResult.loadings.length }, (_, i) => `Var${i + 1}`), // Use loadings.length for number of variables
    explainedVariance: pcaResult.explained_variance_ratio // Already in percentages from backend
  };
}

/**
 * Create Circle of Correlations config
 */
export function createCircleOfCorrelationsConfig(
  xComponent: number = 0,
  yComponent: number = 1,
  theme?: 'light' | 'dark',
  colorScheme?: string[],
  fontScale?: number
): CircleOfCorrelationsConfig {
  return {
    pcX: xComponent + 1,
    pcY: yComponent + 1,
    showCircle: true,
    showGrid: true,
    showLabels: true,
    minVectorLength: 0.1,
    colorByMagnitude: false,  // Use palette colors for each variable
    ...createBaseVisualizationConfig(theme, colorScheme, fontScale)
  };
}

/**
 * Transform PCAResult to Diagnostic Plot data
 */
export function transformToDiagnosticPlotData(
  pcaResult: PCAResult,
  rowNames: string[],
  groupLabels?: string[],
  groupValues?: number[],
  groupType?: 'categorical' | 'continuous'
): DiagnosticPlotData {
  // Extract Mahalanobis distances and RSS from metrics if available
  const metrics = pcaResult.metrics || [];

  return {
    mahalanobisDistances: metrics.map(m => m.mahalanobis || 0),
    residualSumOfSquares: metrics.map(m => m.rss || 0),
    sampleNames: rowNames,
    groups: groupLabels,
    groupValues,
    groupType
  };
}

/**
 * Create Diagnostic Plot config
 * Uses backend-calculated thresholds based on proper statistical distributions:
 * - T² limit: Hotelling's T-squared distribution (leverage in model space)
 * - Q limit: Jackson & Mudholkar SPE distribution (residuals orthogonal to model)
 */
export function createDiagnosticPlotConfig(
  showThresholds: boolean = true,
  confidenceLevel: number = 0.95,
  theme?: 'light' | 'dark',
  colorScheme?: string[],
  mahalanobisThreshold?: number,
  rssThreshold?: number,
  fontScale?: number
): DiagnosticPlotConfig {
  return {
    showThresholds,
    confidenceLevel,
    showLabels: false,  // Changed to false by default
    labelThreshold: 10,
    mahalanobisThreshold,
    rssThreshold,
    ...createBaseVisualizationConfig(theme, colorScheme, fontScale)
  };
}

/**
 * Transform PCAResult to Eigencorrelation Plot data
 */
export function transformToEigencorrelationPlotData(
  pcaResult: PCAResult
): EigencorrelationPlotData | null {
  // Check if eigencorrelations exist
  if (!pcaResult.eigencorrelations) {
    return null;
  }

  const eigencorr = pcaResult.eigencorrelations;

  // Transform from map format to 2D array format [components][variables]
  // Backend format: {variable: [correlations per component]}
  // Frontend expects: [[correlations per component for all variables]]
  const numComponents = eigencorr.components.length;
  const numVariables = eigencorr.variables.length;

  const correlationMatrix: number[][] = [];

  // Build the matrix with components as rows and variables as columns
  for (let compIdx = 0; compIdx < numComponents; compIdx++) {
    const row: number[] = [];
    for (const variable of eigencorr.variables) {
      row.push(eigencorr.correlations[variable][compIdx]);
    }
    correlationMatrix.push(row);
  }

  return {
    correlations: correlationMatrix,
    variableNames: eigencorr.variables, // Metadata variable names
    explainedVariance: pcaResult.explained_variance_ratio // Already in percentages from backend
  };
}

/**
 * Create Eigencorrelation Plot config
 */
export function createEigencorrelationPlotConfig(
  maxComponents?: number,
  theme?: 'light' | 'dark',
  colorScheme?: string[],
  fontScale?: number
): EigencorrelationPlotConfig {
  // Convert color array to Plotly colorscale format
  let colorScale: any = 'Reds'; // Default fallback
  if (colorScheme && colorScheme.length > 0) {
    // Create a Plotly colorscale from the palette colors
    colorScale = colorScheme.map((color, index) => [
      index / (colorScheme.length - 1),
      color
    ]);
  }

  return {
    maxComponents,
    colorScale,
    showValues: true,
    valueFormat: '.2f',
    clusterVariables: false,
    annotationThreshold: 0.3,
    ...createBaseVisualizationConfig(theme, colorScheme, fontScale)
  };
}

/**
 * Transform GoPCA results to 3D Biplot data format
 */
export function transformToBiplot3DData(
  pcaResult: PCAResult,
  rowNames: string[],
  groupLabels?: string[],
  groupValues?: number[],
  groupType?: 'categorical' | 'continuous',
  pc1?: number,
  pc2?: number,
  pc3?: number
): Biplot3DData {
  // Check if loadings exist (e.g., not available for Kernel PCA)
  if (!pcaResult.loadings || pcaResult.loadings.length === 0) {
    throw new Error('3D Biplot visualization requires loadings data, which is not available for this PCA method.');
  }

  // Backend stores loadings as [variables][components], but frontend expects [components][variables]
  const transposedLoadings = transposeMatrix(pcaResult.loadings);

  return {
    scores: pcaResult.scores,
    loadings: transposedLoadings,
    explainedVariance: pcaResult.explained_variance_ratio, // Already in percentages from backend
    sampleNames: rowNames,
    variableNames: pcaResult.variable_labels ||
      Array.from({ length: pcaResult.loadings.length }, (_, i) => `Var${i + 1}`), // Use loadings.length for number of variables
    groups: groupLabels,
    groupValues,
    groupType,
    pc1,
    pc2,
    pc3
  };
}

/**
 * Create configuration for 3D Biplot
 */
export function createBiplot3DConfig(options: {
  theme: 'light' | 'dark';
  colorScheme?: string[];
  showScores?: boolean;
  showLoadings?: boolean;
  showLabels?: boolean;
  maxLabels?: number;
  vectorScale?: number;
  maxVariables?: number;
  fontScale?: number;
}): Biplot3DConfig {
  const {
    theme,
    colorScheme,
    showScores = true,
    showLoadings = true,
    showLabels = false,
    maxLabels = 10,
    vectorScale = 1.0,
    maxVariables = 50,
    fontScale
  } = options;

  return {
    scalingType: 'correlation',
    showScores,
    showLoadings,
    showLabels,
    maxLabels,
    maxVariables,
    vectorScale,
    colorScheme,
    markerSize: 5,
    opacity: 0.8,
    arrowSize: 8,
    arrowOpacity: 0.7,
    showProjections: false,
    cameraPosition: {
      eye: { x: 1.5, y: 1.5, z: 1.5 },
      center: { x: 0, y: 0, z: 0 }
    },
    theme,
    fontScale
  };
}

/**
 * Transform temporal eigenvectors (U matrix) to TemporalLoadingsPlot data format
 */
export function transformToTemporalLoadingsPlotData(
  pcaResult: PCAResult
): TemporalLoadingsPlotData | null {
  if (!pcaResult.temporal_eigenvectors || pcaResult.temporal_eigenvectors.length === 0) {
    return null;
  }

  return {
    temporalEigenvectors: pcaResult.temporal_eigenvectors,
    explainedVariance: pcaResult.explained_variance_ratio
  };
}

/**
 * Create configuration for Temporal Loadings plot
 */
export function createTemporalLoadingsPlotConfig(
  maxComponents: number = 5,
  theme?: 'light' | 'dark',
  colorScheme?: string[],
  fontScale?: number
): TemporalLoadingsPlotConfig {
  return {
    maxComponents,
    theme,
    colorScheme,
    fontScale
  };
}

/**
 * Transform temporal variable importance data to plot format
 */
export function transformToTemporalVariableImportancePlotData(
  pcaResult: PCAResult
): TemporalVariableImportanceData | null {
  if (!pcaResult.temporal_variable_importance || pcaResult.temporal_variable_importance.length === 0) {
    return null;
  }

  // Get variable names or generate default ones
  const variableNames = pcaResult.variable_labels ||
    Array.from({ length: pcaResult.temporal_variable_importance[0].length }, (_, i) => `Var${i + 1}`);

  return {
    importance: pcaResult.temporal_variable_importance,
    variableNames,
    explainedVariance: pcaResult.explained_variance_ratio
  };
}

/**
 * Create configuration for Temporal Variable Importance plot
 */
export function createTemporalVariableImportancePlotConfig(
  maxComponents: number = 10,
  theme?: 'light' | 'dark',
  colorScheme?: string[],
  fontScale?: number
): TemporalVariableImportancePlotConfig {
  // Convert color array to Plotly colorscale format
  let colorScale: any = 'Blues'; // Default fallback
  if (colorScheme && colorScheme.length > 0) {
    // Create a Plotly colorscale from the palette colors
    colorScale = colorScheme.map((color, index) => [
      index / (colorScheme.length - 1),
      color
    ]);
  }

  return {
    maxComponents,
    theme,
    colorScheme,
    fontScale,
    showValues: true,
    valueFormat: '.3f',
    annotationThreshold: 0.01,
    colorScale,
    showWatermark: true  // Enable watermark for consistency with other plots
  };
}