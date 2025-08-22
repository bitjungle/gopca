// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

// GoPCA watermark logo data and utility

// Base64 encoded GoPCA icon - placeholder SVG
// This creates a simple text watermark until the actual logo can be properly encoded
export const GOPCA_LOGO_BASE64 = 'data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNDAiIGhlaWdodD0iNDAiIHZpZXdCb3g9IjAgMCA0MCA0MCIgZmlsbD0ibm9uZSIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj4KPGNpcmNsZSBjeD0iMjAiIGN5PSIyMCIgcj0iMTgiIHN0cm9rZT0iIzNiODJmNiIgc3Ryb2tlLXdpZHRoPSIyIiBmaWxsPSJub25lIiBvcGFjaXR5PSIwLjMiLz4KPHRleHQgeD0iMjAiIHk9IjI2IiBmb250LWZhbWlseT0iQXJpYWwsIHNhbnMtc2VyaWYiIGZvbnQtc2l6ZT0iMTIiIGZpbGw9IiMzYjgyZjYiIHRleHQtYW5jaG9yPSJtaWRkbGUiIG9wYWNpdHk9IjAuNSI+R29QQ0E8L3RleHQ+Cjwvc3ZnPg==';

let loader: Promise<string> | null = null;

/**
 * Get the watermark image data URL
 * Caches the result to avoid repeated loading
 */
export async function getWatermarkDataUrl(): Promise<string> {
  if (!loader) {
    loader = Promise.resolve(GOPCA_LOGO_BASE64);
  }
  return loader;
}

/**
 * Get the watermark image data URL (synchronous version)
 */
export function getWatermarkDataUrlSync(): string {
  return GOPCA_LOGO_BASE64;
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