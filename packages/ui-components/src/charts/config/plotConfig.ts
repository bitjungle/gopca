// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Central configuration for all Plotly visualizations

/**
 * Central configuration for all Plotly visualizations
 * This ensures consistency across all charts and reduces code duplication
 */
export const PLOT_CONFIG = {
  // Export configurations
  export: {
    presentation: {
      format: 'png' as const,
      width: 1920,
      height: 1080,
      scale: 2
    },
    publication: {
      format: 'svg' as const,
      width: 3200,
      height: 2400,
      scale: 4
    },
    web: {
      format: 'png' as const,
      width: 1200,
      height: 800,
      scale: 1
    }
  },
  
  // Color palettes
  colors: {
    // Primary colors (Tailwind palette)
    primary: '#3b82f6',   // Blue-500
    secondary: '#ef4444', // Red-500
    success: '#10b981',   // Green-500
    warning: '#f59e0b',   // Amber-500
    info: '#8b5cf6',      // Purple-500
    
    // Default categorical palette for multi-group visualizations
    categorical: ['#3b82f6', '#ef4444', '#10b981', '#f59e0b', '#8b5cf6'],
    
    // Diagnostic plot specific colors
    diagnostic: {
      normal: '#10b981',      // Green for normal points
      goodLeverage: '#3b82f6', // Blue for good leverage
      outlier: '#f59e0b',     // Amber for outliers
      badLeverage: '#ef4444', // Red for bad leverage
      unknown: '#8b5cf6'      // Purple for unknown
    },
    
    // Grid and axis colors for light/dark themes
    grid: {
      light: 'rgba(128, 128, 128, 0.2)',
      dark: 'rgba(200, 200, 200, 0.15)'
    },
    zeroline: {
      light: 'rgba(128, 128, 128, 0.5)',
      dark: 'rgba(200, 200, 200, 0.3)'
    }
  },
  
  // Visual properties
  visual: {
    markerSize: 10,
    opacity: {
      primary: 0.8,
      secondary: 0.5,
      overlay: 0.3,
      bars: 0.8,
      ellipse: 0.3
    },
    fontSize: {
      label: 10,
      title: 14,
      axis: 12,
      annotation: 10
    },
    line: {
      width: 2,
      dashArray: '5,5'
    }
  },
  
  // Performance thresholds
  performance: {
    webglThreshold: 1000,      // Switch to WebGL above this point count
    decimationThreshold: 10000, // Start decimating above this count
    densityThreshold: 100000,   // Use density plots above this count
    labelThreshold: 100         // Maximum labels to show by default
  }
};

/**
 * Get color from categorical palette by index
 */
export function getCategoricalColor(index: number): string {
  const colors = PLOT_CONFIG.colors.categorical;
  return colors[index % colors.length];
}

/**
 * Get grid color based on theme
 */
export function getGridColor(isDark: boolean = false): string {
  return isDark ? PLOT_CONFIG.colors.grid.dark : PLOT_CONFIG.colors.grid.light;
}

/**
 * Get zeroline color based on theme
 */
export function getZerolineColor(isDark: boolean = false): string {
  return isDark ? PLOT_CONFIG.colors.zeroline.dark : PLOT_CONFIG.colors.zeroline.light;
}