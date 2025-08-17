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
    explainedVariance: pcaResult.explained_variance,
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
    explainedVariance: pcaResult.explained_variance,
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
  return {
    loadings: pcaResult.loadings,
    variableNames: pcaResult.variable_labels || 
      Array.from({ length: pcaResult.loadings[0]?.length || 0 }, (_, i) => `Var${i + 1}`),
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
  colorScheme?: string[]
): LoadingsPlotConfig {
  return {
    mode: plotType,
    sortByMagnitude,
    showThreshold: true,
    thresholdValue: 0.3,
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
  return {
    scores: pcaResult.scores,
    loadings: pcaResult.loadings,
    explainedVariance: pcaResult.explained_variance,
    sampleNames: rowNames,
    variableNames: pcaResult.variable_labels || 
      Array.from({ length: pcaResult.loadings[0]?.length || 0 }, (_, i) => `Var${i + 1}`),
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
  colorScheme?: string[]
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
    colorScheme
  };
}

/**
 * Transform PCAResult to Circle of Correlations data
 */
export function transformToCircleOfCorrelationsData(
  pcaResult: PCAResult
): CircleOfCorrelationsData {
  return {
    loadings: pcaResult.loadings,
    variableNames: pcaResult.variable_labels || 
      Array.from({ length: pcaResult.loadings[0]?.length || 0 }, (_, i) => `Var${i + 1}`),
    explainedVariance: pcaResult.explained_variance
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
 */
export function createDiagnosticPlotConfig(
  showThresholds: boolean = true,
  confidenceLevel: number = 0.975,
  theme?: 'light' | 'dark',
  colorScheme?: string[]
): DiagnosticPlotConfig {
  return {
    showThresholds,
    confidenceLevel,
    showLabels: false,  // Changed to false by default
    labelThreshold: 10,
    theme,
    colorScheme
  };
}

/**
 * Transform PCAResult to Eigencorrelation Plot data
 */
export function transformToEigencorrelationPlotData(
  pcaResult: PCAResult
): EigencorrelationPlotData {
  // Loadings are already correlations in standardized PCA
  return {
    correlations: pcaResult.loadings,
    variableNames: pcaResult.variable_labels || 
      Array.from({ length: pcaResult.loadings[0]?.length || 0 }, (_, i) => `Var${i + 1}`),
    explainedVariance: pcaResult.explained_variance
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