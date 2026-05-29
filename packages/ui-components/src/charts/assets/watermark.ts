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

// GoPCA watermark logo data and utility

import logoUrl from './GoPCA-icon-64-transp.png';

/**
 * Get the watermark image URL
 * Returns the URL to the GoPCA logo image
 */
export function getWatermarkDataUrlSync(): string {
  return logoUrl;
}

/**
 * Get the watermark image URL (async version for compatibility)
 */
export async function getWatermarkDataUrl(): Promise<string> {
  return logoUrl;
}

/**
 * Watermark configuration for Plotly layouts
 */
export interface WatermarkConfig {
  enabled: boolean;
  opacity: {
    light: number;
    dark: number;
  };
  position: {
    x: number;
    y: number;
    xanchor: 'left' | 'center' | 'right';
    yanchor: 'top' | 'middle' | 'bottom';
  };
  size: number;  // Relative size (0-1)
}