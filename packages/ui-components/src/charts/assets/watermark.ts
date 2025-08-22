// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

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