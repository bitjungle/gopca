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

export interface ChartTheme {
  gridColor: string;
  axisColor: string;
  textColor: string;
  backgroundColor: string;
  referenceLineColor: string;
  tooltipBackgroundColor: string;
  tooltipBorderColor: string;
  tooltipTextColor: string;
}

export const getChartTheme = (isDark: boolean): ChartTheme => {
  if (isDark) {
    return {
      gridColor: '#374151',
      axisColor: '#9CA3AF',
      textColor: '#E5E7EB',
      backgroundColor: '#1F2937',
      referenceLineColor: '#6B7280',
      tooltipBackgroundColor: '#1F2937',
      tooltipBorderColor: '#374151',
      tooltipTextColor: '#E5E7EB'
    };
  } else {
    return {
      gridColor: '#E5E7EB',
      axisColor: '#6B7280',
      textColor: '#374151',
      backgroundColor: '#FFFFFF',
      referenceLineColor: '#D1D5DB',
      tooltipBackgroundColor: '#FFFFFF',
      tooltipBorderColor: '#E5E7EB',
      tooltipTextColor: '#374151'
    };
  }
};