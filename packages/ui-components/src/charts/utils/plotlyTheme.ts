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

import { Layout, Config } from 'plotly.js';

export type ThemeMode = 'light' | 'dark';

export interface PlotlyTheme {
  layout: Partial<Layout>;
  config: Partial<Config>;
}

export const getPlotlyTheme = (mode: ThemeMode, fontScale: number = 1.0): PlotlyTheme => {
  const isDark = mode === 'dark';
  const baseFontSize = 12;
  const scaledFontSize = Math.round(baseFontSize * fontScale);

  return {
    layout: {
      paper_bgcolor: isDark ? '#1f2937' : '#ffffff',
      plot_bgcolor: isDark ? '#374151' : '#f9fafb',
      font: {
        family: 'system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif',
        size: scaledFontSize,
        color: isDark ? '#e5e7eb' : '#1f2937'
      },
      xaxis: {
        gridcolor: isDark ? '#4b5563' : '#e5e7eb',
        zerolinecolor: isDark ? '#6b7280' : '#9ca3af',
        linecolor: isDark ? '#6b7280' : '#9ca3af',
        tickfont: {
          color: isDark ? '#d1d5db' : '#4b5563',
          size: scaledFontSize
        }
      },
      yaxis: {
        gridcolor: isDark ? '#4b5563' : '#e5e7eb',
        zerolinecolor: isDark ? '#6b7280' : '#9ca3af',
        linecolor: isDark ? '#6b7280' : '#9ca3af',
        tickfont: {
          color: isDark ? '#d1d5db' : '#4b5563',
          size: scaledFontSize
        }
      },
      hoverlabel: {
        bgcolor: isDark ? '#374151' : '#ffffff',
        bordercolor: isDark ? '#6b7280' : '#e5e7eb',
        font: {
          color: isDark ? '#e5e7eb' : '#1f2937',
          size: scaledFontSize
        }
      },
      legend: {
        bgcolor: isDark ? 'rgba(31, 41, 55, 0.8)' : 'rgba(255, 255, 255, 0.8)',
        bordercolor: isDark ? '#4b5563' : '#e5e7eb',
        borderwidth: 1,
        font: {
          color: isDark ? '#e5e7eb' : '#1f2937',
          size: scaledFontSize
        }
      },
      margin: {
        l: Math.round(60 * Math.max(fontScale, 1.0)),
        r: Math.round(30 * Math.max(fontScale, 1.0)),
        t: Math.round(30 * Math.max(fontScale, 1.0)),
        b: Math.round(60 * Math.max(fontScale, 1.0))
      }
    },
    config: {
      displayModeBar: true,
      displaylogo: false,
      modeBarButtonsToRemove: [
        'sendDataToCloud',
        'select2d',
        'lasso2d',
        'hoverClosestCartesian',
        'hoverCompareCartesian',
        'toggleSpikelines',
        'autoScale2d'
      ],
      modeBarButtonsToAdd: [],
      toImageButtonOptions: {
        format: 'png',
        filename: 'pca-plot',
        height: 1600,
        width: 1600,
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