// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Mathematical utilities for Plotly visualizations with academic references

import { MathReference } from '../core/PlotlyVisualization';

export interface Point2D {
  x: number;
  y: number;
}

export interface EllipseParams {
  centerX: number;
  centerY: number;
  majorAxis: number;
  minorAxis: number;
  angle: number; // in radians
}

export interface VectorTrace {
  x: number[];
  y: number[];
  text: string;
  color: string;
}

/**
 * Calculate confidence ellipse parameters using chi-square distribution
 * Reference: Johnson & Wichern (2007), Applied Multivariate Statistical Analysis, Ch. 4
 *
 * @param points - Bivariate data points
 * @param confidence - Confidence level (0.90, 0.95, 0.99)
 * @returns Ellipse parameters (center, axes, angle)
 *
 * Algorithm complexity: O(n) where n is the number of data points
 */
export function calculateConfidenceEllipse(
  points: Point2D[],
  confidence: number = 0.95
): EllipseParams {
  // Calculate mean
  const n = points.length;
  const meanX = points.reduce((sum, p) => sum + p.x, 0) / n;
  const meanY = points.reduce((sum, p) => sum + p.y, 0) / n;

  // Calculate covariance matrix elements
  let cov_xx = 0, cov_yy = 0, cov_xy = 0;
  for (const point of points) {
    const dx = point.x - meanX;
    const dy = point.y - meanY;
    cov_xx += dx * dx;
    cov_yy += dy * dy;
    cov_xy += dx * dy;
  }
  cov_xx /= (n - 1);
  cov_yy /= (n - 1);
  cov_xy /= (n - 1);

  // Eigendecomposition of 2x2 covariance matrix
  // Reference: Johnson & Wichern (2007), Equation 4.8
  const trace = cov_xx + cov_yy;
  const det = cov_xx * cov_yy - cov_xy * cov_xy;
  const discriminant = Math.sqrt(Math.max(0, trace * trace - 4 * det));

  const eigenvalue1 = (trace + discriminant) / 2;
  const eigenvalue2 = (trace - discriminant) / 2;

  // Angle of rotation (principal axis)
  const angle = Math.atan2(2 * cov_xy, cov_xx - cov_yy) / 2;

  // Chi-square critical value for 2 degrees of freedom
  const chiSquare = getChiSquareCritical(confidence, 2);

  // Ellipse axes lengths
  const majorAxis = 2 * Math.sqrt(chiSquare * Math.max(eigenvalue1, eigenvalue2));
  const minorAxis = 2 * Math.sqrt(chiSquare * Math.min(eigenvalue1, eigenvalue2));

  return {
    centerX: meanX,
    centerY: meanY,
    majorAxis,
    minorAxis,
    angle
  };
}

/**
 * Get chi-square critical value
 * Reference: Johnson & Wichern (2007), Table 3 in Appendix
 */
export function getChiSquareCritical(confidence: number, df: number): number {
  // Chi-square critical values for df=2
  const chiSquareTable: Record<number, number> = {
    0.90: 4.605,
    0.95: 5.991,
    0.99: 9.210
  };

  if (df !== 2) {
    console.warn('Chi-square values only implemented for df=2');
  }

  return chiSquareTable[confidence] || chiSquareTable[0.95];
}

/**
 * Generate ellipse path points for plotting
 * @param params - Ellipse parameters
 * @param numPoints - Number of points to generate (default 100)
 * @returns Array of points forming the ellipse
 */
export function generateEllipsePath(
  params: EllipseParams,
  numPoints: number = 100
): Point2D[] {
  const points: Point2D[] = [];
  const angleStep = (2 * Math.PI) / numPoints;

  for (let i = 0; i <= numPoints; i++) {
    const theta = i * angleStep;

    // Point on standard ellipse
    const x0 = (params.majorAxis / 2) * Math.cos(theta);
    const y0 = (params.minorAxis / 2) * Math.sin(theta);

    // Rotate by angle
    const cos_a = Math.cos(params.angle);
    const sin_a = Math.sin(params.angle);
    const x = x0 * cos_a - y0 * sin_a + params.centerX;
    const y = x0 * sin_a + y0 * cos_a + params.centerY;

    points.push({ x, y });
  }

  return points;
}

/**
 * Scale loading vectors for biplot visualization
 * Reference: Gabriel (1971), "The biplot graphic display", Biometrika 58(3), 453-467
 *
 * @param loadings - Loading matrix (variables x components)
 * @param scores - Score matrix (observations x components)
 * @param scale - Scaling factor (0-1, where 0.5 is symmetric)
 * @param variableNames - Names of variables
 * @returns Scaled vector traces for plotting
 *
 * scale=0: row-metric preserving (emphasis on observations)
 * scale=1: column-metric preserving (emphasis on variables)
 * scale=0.5: symmetric (default, balanced emphasis)
 */
export function scaleBiplotVectors(
  loadings: number[][],
  scores: number[][],
  scale: number = 0.5,
  variableNames: string[]
): VectorTrace[] {
  const nObservations = scores.length;

  // Gabriel (1971), Equation 2
  const alpha = Math.pow(nObservations, scale);

  // Find maximum score value for scaling
  let maxScore = 0;
  for (const score of scores) {
    for (const val of score) {
      maxScore = Math.max(maxScore, Math.abs(val));
    }
  }

  // Scale loadings
  const scaledLoadings = loadings.map(loading =>
    loading.map(val => val * alpha)
  );

  // Find maximum loading for additional scaling
  let maxLoading = 0;
  for (const loading of scaledLoadings) {
    for (const val of loading) {
      maxLoading = Math.max(maxLoading, Math.abs(val));
    }
  }

  // Additional scaling to fit within plot
  const plotScale = (maxScore * 0.8) / maxLoading;

  return scaledLoadings.map((loading, i) => ({
    x: [0, loading[0] * plotScale],
    y: [0, loading[1] * plotScale],
    text: variableNames[i],
    color: 'red'
  }));
}

/**
 * Calculate smart labels - select points furthest from origin
 * This preserves the smart label selection feature from the previous implementation
 *
 * @param points - Data points
 * @param maxLabels - Maximum number of labels to show
 * @returns Indices of points to label
 */
export function calculateSmartLabels(
  points: Point2D[],
  maxLabels: number = 10
): number[] {
  // Calculate distances from origin
  const distances = points.map((p, i) => ({
    index: i,
    distance: Math.sqrt(p.x * p.x + p.y * p.y)
  }));

  // Sort by distance and take top N
  distances.sort((a, b) => b.distance - a.distance);

  return distances
    .slice(0, maxLabels)
    .map(d => d.index);
}

// Alias for consistency
export const selectSmartLabels = calculateSmartLabels;

/**
 * Calculate biplot scaling for loading vectors with adaptive visual scaling
 * Reference: Gabriel (1971), "The biplot graphic display of matrices with application to principal component analysis"
 *
 * @param loadings - Loading matrix [n_components][n_variables]
 * @param explainedVariance - Explained variance per component
 * @param scalingType - Type of scaling to apply
 * @param scores - Optional score matrix for adaptive scaling [n_samples][n_components]
 * @returns Scaled loadings for biplot visualization
 */
export function calculateBiplotScaling(
  loadings: number[][],
  explainedVariance: number[],
  scalingType: 'correlation' | 'symmetric' | 'pca' = 'correlation',
  scores?: number[][]
): { scaledLoadings: number[][]; adaptiveScale: number } {
  // First apply mathematical scaling based on biplot type
  const mathScaledLoadings = loadings.map((componentLoadings, i) => {
    let scale = 1;

    switch (scalingType) {
      case 'correlation':
        // Scale by sqrt of explained variance (preserves correlations)
        // This emphasizes variable relationships
        scale = Math.sqrt(explainedVariance[i] / 100);
        break;
      case 'symmetric':
        // Square root scaling for both scores and loadings
        // This provides a balanced representation
        scale = Math.pow(explainedVariance[i] / 100, 0.25);
        break;
      case 'pca':
        // Standard PCA scaling (loadings as-is)
        // Raw loadings without variance scaling
        scale = 1;
        break;
    }

    return componentLoadings.map(loading => loading * scale);
  });

  // Apply adaptive visual scaling if scores are provided
  let adaptiveScale = 1;
  if (scores && scores.length > 0) {
    // Calculate the range of scores
    const scoreValues = scores.flat();
    const scoreRange = Math.max(...scoreValues.map(Math.abs));

    // Calculate the range of mathematically scaled loadings
    const loadingValues = mathScaledLoadings.flat();
    const loadingRange = Math.max(...loadingValues.map(Math.abs));

    // Ensure loadings are not too small
    if (loadingRange > 0 && scoreRange > 0) {
      // Scale loadings to use 60-70% of the score range for good visibility
      // This ensures arrows are clearly visible without dominating the plot
      adaptiveScale = (scoreRange * 0.65) / loadingRange;

      // Apply reasonable limits to prevent extreme scaling
      adaptiveScale = Math.max(0.1, Math.min(adaptiveScale, 100));
    }
  }

  // Apply adaptive scaling to all loadings
  const scaledLoadings = mathScaledLoadings.map(componentLoadings =>
    componentLoadings.map(loading => loading * adaptiveScale)
  );

  return { scaledLoadings, adaptiveScale };
}

/**
 * 2D Kernel Density Estimation using Gaussian kernel
 * Reference: Scott (1992), "Multivariate Density Estimation", Wiley
 *
 * @param points - Data points
 * @param bandwidth - Bandwidth parameter ('scott', 'silverman', or numeric)
 * @param gridSize - Grid resolution (default 50x50)
 * @returns Density grid for contour plotting
 */
export function kernelDensityEstimate2D(
  points: Point2D[],
  bandwidth: 'scott' | 'silverman' | number = 'scott',
  gridSize: number = 50
): { x: number[], y: number[], z: number[][] } {
  const n = points.length;

  // Calculate bandwidth using Scott's or Silverman's rule
  let h: number;
  if (bandwidth === 'scott') {
    // Scott's rule: h = n^(-1/6) * std
    const stdX = calculateStandardDeviation(points.map(p => p.x));
    const stdY = calculateStandardDeviation(points.map(p => p.y));
    h = Math.pow(n, -1/6) * Math.sqrt(stdX * stdY);
  } else if (bandwidth === 'silverman') {
    // Silverman's rule: h = 0.9 * min(std, IQR/1.34) * n^(-1/5)
    const stdX = calculateStandardDeviation(points.map(p => p.x));
    const stdY = calculateStandardDeviation(points.map(p => p.y));
    h = 0.9 * Math.sqrt(stdX * stdY) * Math.pow(n, -1/5);
  } else {
    h = bandwidth;
  }

  // Create grid
  const xMin = Math.min(...points.map(p => p.x)) - 3 * h;
  const xMax = Math.max(...points.map(p => p.x)) + 3 * h;
  const yMin = Math.min(...points.map(p => p.y)) - 3 * h;
  const yMax = Math.max(...points.map(p => p.y)) + 3 * h;

  const xStep = (xMax - xMin) / gridSize;
  const yStep = (yMax - yMin) / gridSize;

  const x = Array.from({ length: gridSize }, (_, i) => xMin + i * xStep);
  const y = Array.from({ length: gridSize }, (_, i) => yMin + i * yStep);
  const z: number[][] = Array(gridSize).fill(0).map(() => Array(gridSize).fill(0));

  // Calculate density at each grid point
  const norm = 1 / (2 * Math.PI * h * h * n);

  for (let i = 0; i < gridSize; i++) {
    for (let j = 0; j < gridSize; j++) {
      let density = 0;
      for (const point of points) {
        const dx = (x[j] - point.x) / h;
        const dy = (y[i] - point.y) / h;
        const dist2 = dx * dx + dy * dy;
        density += Math.exp(-0.5 * dist2);
      }
      z[i][j] = density * norm;
    }
  }

  return { x, y, z };
}

/**
 * Calculate standard deviation
 */
function calculateStandardDeviation(values: number[]): number {
  const n = values.length;
  const mean = values.reduce((sum, v) => sum + v, 0) / n;
  const variance = values.reduce((sum, v) => sum + (v - mean) ** 2, 0) / (n - 1);
  return Math.sqrt(variance);
}

/**
 * Mathematical references for PCA visualizations
 */
export const PCA_REFERENCES: MathReference[] = [
  {
    authors: 'Johnson, R. A., & Wichern, D. W.',
    title: 'Applied Multivariate Statistical Analysis',
    year: 2007,
    page: 'Ch. 4'
  },
  {
    authors: 'Gabriel, K. R.',
    title: 'The biplot graphic display of matrices with application to principal component analysis',
    year: 1971,
    page: 'Biometrika 58(3), 453-467'
  },
  {
    authors: 'Scott, D. W.',
    title: 'Multivariate Density Estimation',
    year: 1992
  },
  {
    authors: 'Golub, G. H., & Van Loan, C. F.',
    title: 'Matrix Computations',
    year: 2013,
    page: 'Ch. 8'
  }
];