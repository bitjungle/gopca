// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

import { Layout, Config } from 'plotly.js';

export type ThemeMode = 'light' | 'dark';

export interface PlotlyTheme {
  layout: Partial<Layout>;
  config: Partial<Config>;
}

export const getPlotlyTheme = (mode: ThemeMode): PlotlyTheme => {
  const isDark = mode === 'dark';
  
  return {
    layout: {
      paper_bgcolor: isDark ? '#1f2937' : '#ffffff',
      plot_bgcolor: isDark ? '#374151' : '#f9fafb',
      font: {
        family: 'system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif',
        size: 12,
        color: isDark ? '#e5e7eb' : '#1f2937'
      },
      xaxis: {
        gridcolor: isDark ? '#4b5563' : '#e5e7eb',
        zerolinecolor: isDark ? '#6b7280' : '#9ca3af',
        linecolor: isDark ? '#6b7280' : '#9ca3af',
        tickfont: {
          color: isDark ? '#d1d5db' : '#4b5563'
        }
      },
      yaxis: {
        gridcolor: isDark ? '#4b5563' : '#e5e7eb',
        zerolinecolor: isDark ? '#6b7280' : '#9ca3af',
        linecolor: isDark ? '#6b7280' : '#9ca3af',
        tickfont: {
          color: isDark ? '#d1d5db' : '#4b5563'
        }
      },
      hoverlabel: {
        bgcolor: isDark ? '#374151' : '#ffffff',
        bordercolor: isDark ? '#6b7280' : '#e5e7eb',
        font: {
          color: isDark ? '#e5e7eb' : '#1f2937'
        }
      },
      legend: {
        bgcolor: isDark ? 'rgba(31, 41, 55, 0.8)' : 'rgba(255, 255, 255, 0.8)',
        bordercolor: isDark ? '#4b5563' : '#e5e7eb',
        borderwidth: 1,
        font: {
          color: isDark ? '#e5e7eb' : '#1f2937'
        }
      },
      margin: {
        l: 60,
        r: 30,
        t: 30,
        b: 60
      }
    },
    config: {
      displayModeBar: true,
      displaylogo: false,
      modeBarButtonsToRemove: ['sendDataToCloud', 'select2d', 'lasso2d'],
      toImageButtonOptions: {
        format: 'png',
        filename: 'gopca-plot',
        height: 800,
        width: 1200,
        scale: 2
      }
    }
  };
};

export const mergeLayouts = (
  base: Partial<Layout>,
  ...overrides: Partial<Layout>[]
): Partial<Layout> => {
  return overrides.reduce((acc, override) => {
    const merged = { ...acc, ...override };
    
    // Handle axis titles properly
    if (override.xaxis) {
      merged.xaxis = {
        ...acc.xaxis,
        ...override.xaxis,
        title: typeof override.xaxis.title === 'string' 
          ? { text: override.xaxis.title }
          : override.xaxis.title
      };
    }
    
    if (override.yaxis) {
      merged.yaxis = {
        ...acc.yaxis,
        ...override.yaxis,
        title: typeof override.yaxis.title === 'string'
          ? { text: override.yaxis.title }
          : override.yaxis.title
      };
    }
    
    if (override.xaxis2) {
      merged.xaxis2 = { ...acc.xaxis2, ...override.xaxis2 };
    }
    
    if (override.yaxis2) {
      merged.yaxis2 = { ...acc.yaxis2, ...override.yaxis2 };
    }
    
    if (override.font) {
      merged.font = { ...acc.font, ...override.font };
    }
    
    if (override.hoverlabel) {
      merged.hoverlabel = { ...acc.hoverlabel, ...override.hoverlabel };
    }
    
    if (override.legend) {
      merged.legend = { ...acc.legend, ...override.legend };
    }
    
    if (override.margin) {
      merged.margin = { ...acc.margin, ...override.margin };
    }
    
    return merged;
  }, base);
};