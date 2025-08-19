// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Utility functions to transform GoPCA data to Plotly component formats

import { PCAResult, EllipseParams, SampleMetrics } from '../types';
import type {
  ScoresPlotData,
  ScoresPlotConfig,
  ScreePlotData,
  ScreePlotConfig,
  LoadingsPlotData,
  LoadingsPlotConfig,
  BiplotData,
  BiplotConfig,
  CircleOfCorrelationsData,
  CircleOfCorrelationsConfig,
  DiagnosticPlotData,
  DiagnosticPlotConfig,
  EigencorrelationPlotData,
  EigencorrelationPlotConfig,
} from '@gopca/ui-components';

/**
 * Helper function to transpose a matrix
 * Converts [rows][cols] to [cols][rows]
 */
function transposeMatrix(matrix: number[][]): number[][] {
  if (!matrix || matrix.length === 0) return [];
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
  colorScheme?: string[]
): ScoresPlotConfig {
  return {
    showEllipses,
    ellipseConfidence: confidenceLevel,
    showSmartLabels: showRowLabels,
    maxLabels: maxLabelsToShow,
    theme,
    colorScheme
  };
}

/**
 * Transform PCAResult to Plotly ScreePlot data format
 */
export function transformToScreePlotData(pcaResult: PCAResult): ScreePlotData {
  return {
    explainedVariance: pcaResult.explained_variance_ratio, // Already in percentages from backend
    cumulativeVariance: pcaResult.cumulative_variance,
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
  colorScheme?: string[]
): ScreePlotConfig {
  return {
    showCumulativeLine: showCumulative,
    showThresholdLine: true,
    thresholdValue: elbowThreshold,
    theme,
    colorScheme
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
  variableThreshold?: number
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
    colorScheme
  };
}

/**
 * Transform PCAResult to Plotly Biplot data format
 */
export function transformToBiplotData(
  pcaResult: PCAResult,
  rowNames: string[],
  groupLabels?: string[]
): BiplotData {
  // Backend stores loadings as [variables][components], but frontend expects [components][variables]
  const transposedLoadings = transposeMatrix(pcaResult.loadings);
  
  return {
    scores: pcaResult.scores,
    loadings: transposedLoadings,
    explainedVariance: pcaResult.explained_variance_ratio, // Already in percentages from backend
    sampleNames: rowNames,
    variableNames: pcaResult.variable_labels || 
      Array.from({ length: pcaResult.loadings.length }, (_, i) => `Var${i + 1}`), // Use loadings.length for number of variables
    groups: groupLabels
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
  ellipseConfidence: number = 0.95
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
    theme,
    colorScheme,
    showEllipses,
    ellipseConfidence
  };
}

/**
 * Transform PCAResult to Circle of Correlations data
 */
export function transformToCircleOfCorrelationsData(
  pcaResult: PCAResult
): CircleOfCorrelationsData {
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
  colorScheme?: string[]
): CircleOfCorrelationsConfig {
  return {
    pcX: xComponent + 1,
    pcY: yComponent + 1,
    showCircle: true,
    showGrid: true,
    showLabels: true,
    minVectorLength: 0.1,
    colorByMagnitude: true,
    theme,
    colorScheme
  };
}

/**
 * Transform PCAResult to Diagnostic Plot data
 */
export function transformToDiagnosticPlotData(
  pcaResult: PCAResult,
  rowNames: string[],
  groupLabels?: string[]
): DiagnosticPlotData {
  // Extract Mahalanobis distances and RSS from metrics if available
  const metrics = pcaResult.metrics || [];
  
  return {
    mahalanobisDistances: metrics.map(m => m.mahalanobis || 0),
    residualSumOfSquares: metrics.map(m => m.rss || 0),
    sampleNames: rowNames,
    groups: groupLabels
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
  rssThreshold?: number
): DiagnosticPlotConfig {
  return {
    showThresholds,
    confidenceLevel,
    showLabels: false,  // Changed to false by default
    labelThreshold: 10,
    theme,
    colorScheme,
    mahalanobisThreshold,
    rssThreshold
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
  colorScheme?: string[]
): EigencorrelationPlotConfig {
  return {
    maxComponents,
    colorScale: 'RdBu',
    showValues: true,
    valueFormat: '.2f',
    clusterVariables: false,
    annotationThreshold: 0.3,
    theme,
    colorScheme
  };
}