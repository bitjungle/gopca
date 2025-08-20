// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Performance optimization utilities for Plotly visualizations

import { Data } from 'plotly.js';

export interface PerformanceConfig {
  webglThreshold: number;
  decimationThreshold: number;
  densityThreshold: number;
  targetFPS: number;
}

export const DEFAULT_PERFORMANCE_CONFIG: PerformanceConfig = {
  webglThreshold: 100,  // Use WebGL for datasets > 100 points for better performance
  decimationThreshold: 10000,
  densityThreshold: 100000,
  targetFPS: 60
};

/**
 * Optimize trace type based on data size
 * Uses WebGL (scattergl) for better performance with large datasets
 *
 * @param data - Data array or trace
 * @param threshold - Size threshold for WebGL (default 1000)
 * @returns Optimized trace type
 */
export function optimizeTraceType(data: any[], threshold: number = 100): 'scatter' | 'scattergl' {
  return data.length > threshold ? 'scattergl' : 'scatter';
}

/**
 * Convert scatter traces to WebGL (scattergl) for performance
 *
 * @param traces - Array of Plotly traces
 * @returns Traces with WebGL types where applicable
 */
export function convertToWebGL(traces: Data[]): Data[] {
  return traces.map(trace => {
    if (trace.type === 'scatter') {
      return { ...trace, type: 'scattergl' as any };
    }
    if (trace.type === 'bar' && 'x' in trace && Array.isArray(trace.x) && trace.x.length > 1000) {
      // For very large bar charts, consider using scatter with bar mode
      return {
        ...trace,
        type: 'scattergl' as any,
        mode: 'markers',
        marker: {
          ...((trace as any).marker || {}),
          symbol: 'square',
          size: 20
        }
      };
    }
    return trace;
  });
}

/**
 * Decimate data using uniform sampling
 * Reduces data points while preserving overall shape
 *
 * @param data - Input data array
 * @param targetSize - Target number of points
 * @returns Decimated data
 */
export function decimateData<T>(data: T[], targetSize: number): T[] {
  if (data.length <= targetSize) {
    return data;
  }

  const step = Math.ceil(data.length / targetSize);
  const decimated: T[] = [];

  // Always include first and last points
  decimated.push(data[0]);

  for (let i = step; i < data.length - 1; i += step) {
    decimated.push(data[i]);
  }

  decimated.push(data[data.length - 1]);

  return decimated;
}

/**
 * Decimate trace data while preserving structure
 *
 * @param trace - Plotly trace
 * @param targetSize - Target number of points
 * @returns Decimated trace
 */
export function decimateTrace(trace: Data, targetSize: number): Data {
  if (!('x' in trace) || !Array.isArray(trace.x)) {
    return trace;
  }

  const originalSize = trace.x.length;
  if (originalSize <= targetSize) {
    return trace;
  }

  const indices = getDecimationIndices(originalSize, targetSize);

  const decimatedTrace: any = { ...trace };

  // Decimate all array properties
  if (trace.x) {
decimatedTrace.x = indices.map(i => (trace.x as any[])[i]);
}
  if ('y' in trace && Array.isArray(trace.y)) {
    decimatedTrace.y = indices.map(i => (trace.y as any[])[i]);
  }
  if ('z' in trace && Array.isArray(trace.z)) {
    decimatedTrace.z = indices.map(i => (trace.z as any[])[i]);
  }
  if (trace.text) {
decimatedTrace.text = indices.map(i => (trace.text as any[])[i]);
}
  if ('customdata' in trace && Array.isArray(trace.customdata)) {
    decimatedTrace.customdata = indices.map(i => (trace.customdata as any[])[i]);
  }

  return decimatedTrace;
}

/**
 * Get indices for decimation that preserve data extremes
 *
 * @param originalSize - Original data size
 * @param targetSize - Target data size
 * @returns Array of indices to keep
 */
function getDecimationIndices(originalSize: number, targetSize: number): number[] {
  const step = (originalSize - 1) / (targetSize - 1);
  const indices: number[] = [];

  for (let i = 0; i < targetSize; i++) {
    indices.push(Math.round(i * step));
  }

  // Ensure last index is included
  if (indices[indices.length - 1] !== originalSize - 1) {
    indices[indices.length - 1] = originalSize - 1;
  }

  return indices;
}

/**
 * Estimate render time based on data size and type
 *
 * @param dataSize - Number of data points
 * @param traceType - Type of trace
 * @returns Estimated render time in milliseconds
 */
export function estimateRenderTime(dataSize: number, traceType: string): number {
  // Based on empirical measurements from Python benchmarks
  const baseTime = 10; // Base overhead in ms

  if (traceType === 'scattergl') {
    // WebGL rendering: logarithmic complexity
    return baseTime + Math.log10(dataSize) * 5;
  } else if (traceType === 'scatter') {
    // SVG rendering: linear complexity
    return baseTime + dataSize * 0.01;
  } else if (traceType === 'heatmap' || traceType === 'contour') {
    // Heatmap/contour: quadratic for grid generation
    return baseTime + Math.sqrt(dataSize) * 2;
  }

  return baseTime + dataSize * 0.005;
}

/**
 * Check if WebGL is supported in the current browser
 */
export function isWebGLSupported(): boolean {
  try {
    const canvas = document.createElement('canvas');
    return !!(
      window.WebGLRenderingContext &&
      (canvas.getContext('webgl') || canvas.getContext('experimental-webgl'))
    );
  } catch (e) {
    return false;
  }
}

/**
 * Get optimal configuration based on data characteristics
 *
 * @param dataSize - Number of data points
 * @param hasGroups - Whether data has multiple groups
 * @param isInteractive - Whether interactivity is needed
 * @returns Optimal performance configuration
 */
export function getOptimalConfig(
  dataSize: number,
  hasGroups: boolean = false,
  isInteractive: boolean = true
): {
  useWebGL: boolean;
  decimation: boolean;
  targetSize?: number;
  hovermode: string;
  dragmode: string;
} {
  const config: any = {
    useWebGL: false,
    decimation: false,
    hovermode: 'closest',
    dragmode: 'zoom'
  };

  if (!isWebGLSupported()) {
    // Fallback to aggressive decimation without WebGL
    if (dataSize > 5000) {
      config.decimation = true;
      config.targetSize = 5000;
    }
  } else {
    if (dataSize > DEFAULT_PERFORMANCE_CONFIG.webglThreshold) {
      config.useWebGL = true;
    }

    if (dataSize > DEFAULT_PERFORMANCE_CONFIG.decimationThreshold) {
      config.decimation = true;
      config.targetSize = DEFAULT_PERFORMANCE_CONFIG.decimationThreshold;
    }
  }

  // Adjust interactivity for large datasets
  if (dataSize > 10000) {
    config.hovermode = false; // Disable hover for very large datasets
  } else if (dataSize > 5000 && hasGroups) {
    config.hovermode = 'x unified'; // More efficient for grouped data
  }

  if (isInteractive && dataSize < 5000) {
    config.dragmode = 'lasso'; // Enable lasso selection for smaller datasets
  }

  return config;
}

/**
 * Performance monitoring utility
 */
export class PerformanceMonitor {
  private frameTimestamps: number[] = [];
  private readonly maxSamples = 60;

  /**
   * Record a frame render
   */
  recordFrame(): void {
    const now = performance.now();
    this.frameTimestamps.push(now);

    // Keep only recent samples
    if (this.frameTimestamps.length > this.maxSamples) {
      this.frameTimestamps.shift();
    }
  }

  /**
   * Calculate current FPS
   */
  getFPS(): number {
    if (this.frameTimestamps.length < 2) {
      return 60; // Default
    }

    const timeSpan = this.frameTimestamps[this.frameTimestamps.length - 1] - this.frameTimestamps[0];
    const frameCount = this.frameTimestamps.length - 1;

    return (frameCount / timeSpan) * 1000;
  }

  /**
   * Check if performance is degraded
   */
  isPerformanceDegraded(targetFPS: number = 30): boolean {
    return this.getFPS() < targetFPS;
  }

  /**
   * Reset monitoring
   */
  reset(): void {
    this.frameTimestamps = [];
  }
}