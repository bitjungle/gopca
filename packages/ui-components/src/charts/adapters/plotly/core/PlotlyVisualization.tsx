// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// PlotlyVisualization Base Class - Foundation for all Plotly visualizations

import React from 'react';
import Plot from 'react-plotly.js';
import { Data, Layout, Config, PlotlyHTMLElement } from 'plotly.js';
import { getPlotlyTheme, mergeLayouts, ThemeMode } from '../utils/plotlyTheme';

export interface MathReference {
  authors: string;
  title: string;
  year: number;
  page?: string;
  equation?: string;
}

export interface PlotlyVisualizationConfig {
  useWebGL?: boolean;
  dataThreshold?: number;
  enableLasso?: boolean;
  enableCrosshair?: boolean;
  showDensity?: boolean;
  densityType?: 'contour' | 'heatmap' | 'kde';
  exportScale?: number;
  maintainAspectRatio?: boolean;
  theme?: ThemeMode;
}

export interface PlotlyButton {
  name: string;
  title: string;
  icon?: string;
  toggle?: boolean;
  click: (gd: PlotlyHTMLElement) => void;
}

/**
 * Base class for all Plotly visualizations in GoPCA
 * Provides automatic optimization, theming, and advanced features
 */
export abstract class PlotlyVisualization<T = any> {
  protected data: T;
  protected config: PlotlyVisualizationConfig;
  protected theme: ThemeMode;
  
  // Performance thresholds
  protected readonly WEBGL_THRESHOLD = 1000;
  protected readonly DECIMATION_THRESHOLD = 10000;
  protected readonly DENSITY_THRESHOLD = 100000;
  
  constructor(data: T, config?: PlotlyVisualizationConfig) {
    this.data = data;
    this.config = {
      useWebGL: true,
      dataThreshold: this.WEBGL_THRESHOLD,
      enableLasso: true,
      enableCrosshair: false,
      showDensity: false,
      densityType: 'contour',
      exportScale: 2,
      maintainAspectRatio: false,
      theme: 'light',
      ...config
    };
    this.theme = this.config.theme || 'light';
  }
  
  /**
   * Get optimized traces based on data size
   * Automatically switches between scatter, scattergl, and density representations
   */
  protected abstract getTraces(): Data[];
  
  /**
   * Get standard traces (SVG rendering)
   */
  protected abstract getStandardTraces(): Data[];
  
  /**
   * Get WebGL optimized traces
   */
  protected abstract getWebGLTraces(): Data[];
  
  /**
   * Get the plot layout configuration
   */
  protected abstract getLayout(): Partial<Layout>;
  
  /**
   * Optimize traces based on data size
   * Algorithm: Use WebGL for >1000 points, decimation for >10000, density for >100000
   */
  protected optimizeForPerformance(traces: Data[]): Data[] {
    const dataSize = this.getDataSize();
    
    if (!this.config.useWebGL || dataSize <= this.config.dataThreshold!) {
      return traces;
    }
    
    if (dataSize <= this.DECIMATION_THRESHOLD) {
      // Use WebGL rendering (scattergl)
      return this.convertToWebGL(traces);
    }
    
    if (dataSize <= this.DENSITY_THRESHOLD) {
      // Apply decimation
      return this.decimateData(traces, this.DECIMATION_THRESHOLD);
    }
    
    // Use density representation
    return this.convertToDensity(traces);
  }
  
  /**
   * Convert traces to WebGL (scattergl)
   */
  protected convertToWebGL(traces: Data[]): Data[] {
    return traces.map(trace => {
      if (trace.type === 'scatter') {
        return { ...trace, type: 'scattergl' as any };
      }
      return trace;
    });
  }
  
  /**
   * Decimate data for very large datasets
   * Uses uniform sampling to reduce data points
   */
  protected decimateData(traces: Data[], targetSize: number): Data[] {
    return traces.map(trace => {
      if ('x' in trace && Array.isArray(trace.x)) {
        const originalSize = trace.x.length;
        if (originalSize <= targetSize) return trace;
        
        const step = Math.ceil(originalSize / targetSize);
        const decimatedIndices = Array.from(
          { length: Math.floor(originalSize / step) },
          (_, i) => i * step
        );
        
        const decimatedTrace: any = {
          ...trace,
          x: decimatedIndices.map(i => (trace.x as any[])[i])
        };
        
        if ('y' in trace && Array.isArray(trace.y)) {
          decimatedTrace.y = decimatedIndices.map(i => (trace.y as any[])[i]);
        }
        
        if (trace.text) {
          decimatedTrace.text = decimatedIndices.map(i => (trace.text as any[])[i]);
        }
        
        return decimatedTrace;
      }
      return trace;
    });
  }
  
  /**
   * Convert to density representation for massive datasets
   * This should be overridden by specific visualizations
   */
  protected convertToDensity(traces: Data[]): Data[] {
    console.warn('Density conversion not implemented for this visualization type');
    return this.decimateData(traces, this.DECIMATION_THRESHOLD);
  }
  
  /**
   * Get the size of the dataset
   */
  protected abstract getDataSize(): number;
  
  /**
   * Get enhanced layout with all features
   */
  protected getEnhancedLayout(): Partial<Layout> {
    const baseLayout = this.getLayout();
    const themeLayout = getPlotlyTheme(this.theme).layout;
    
    return mergeLayouts(
      themeLayout,
      baseLayout,
      {
        dragmode: this.config.enableLasso ? 'lasso' : 'zoom',
        hovermode: this.config.enableCrosshair ? 'x unified' : 'closest',
        showlegend: true
      }
    );
  }
  
  /**
   * Get advanced configuration for the plot
   */
  protected getAdvancedConfig(): Partial<Config> {
    const config: Partial<Config> = {
      responsive: true,
      displaylogo: false,
      modeBarButtonsToAdd: [],
      toImageButtonOptions: {
        format: 'svg',
        filename: 'pca-plot',
        height: 1200 * (this.config.exportScale || 2),
        width: 1600 * (this.config.exportScale || 2),
        scale: this.config.exportScale || 2
      }
    };
    
    if (this.config.enableLasso) {
      config.modeBarButtonsToAdd = ['select2d', 'lasso2d'];
    }
    
    return config;
  }
  
  
  /**
   * Render the visualization
   */
  render(): React.ReactElement {
    const traces = this.optimizeForPerformance(this.getTraces());
    const layout = this.getEnhancedLayout();
    const themeConfig = getPlotlyTheme(this.theme).config;
    const config = { ...themeConfig, ...this.getAdvancedConfig() };
    
    return (
      <Plot
        data={traces}
        layout={layout}
        config={config}
        style={{ width: '100%', height: '100%' }}
        useResizeHandler={true}
      />
    );
  }
}