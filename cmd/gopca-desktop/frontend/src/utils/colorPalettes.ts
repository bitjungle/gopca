// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

// Color palette definitions for data visualization
// Based on seaborn color palettes and scientific visualization best practices

// Palette names for user selection
export type QualitativePaletteName = 'deep' | 'pastel' | 'dark' | 'colorblind' | 'husl';
export type SequentialPaletteName = 'rocket' | 'viridis' | 'blues' | 'reds' | 'crest' | 'mako' | 'flare';

// Qualitative palettes for categorical data
export const QUALITATIVE_PALETTES: Record<QualitativePaletteName, string[]> = {
  // Seaborn deep/default palette - extended to 25 colors
  deep: [
    // Original 10 colors (unchanged for backward compatibility)
    '#4C72B0', // blue
    '#DD8452', // orange
    '#55A868', // green
    '#C44E52', // red
    '#8172B3', // purple
    '#937860', // brown
    '#DA8BC3', // pink
    '#8C8C8C', // gray
    '#CCB974', // tan
    '#64B5CD', // light blue
    // Additional 15 colors maintaining deep/saturated character
    '#2E5985', // deep blue
    '#E67E22', // deep orange
    '#27AE60', // forest green
    '#9B59B6', // deep purple
    '#16A085', // teal
    '#D35400', // burnt orange
    '#8E44AD', // violet
    '#2C3E50', // dark blue-gray
    '#E74C3C', // crimson
    '#F39C12', // amber
    '#1ABC9C', // turquoise
    '#34495E', // slate
    '#7F8C8D', // medium gray
    '#C0392B', // dark red
    '#D68910'  // gold
  ],

  // Seaborn pastel palette (lighter version) - extended to 25 colors
  pastel: [
    // Original 10 colors (unchanged for backward compatibility)
    '#A1C9F4', // light blue
    '#FFB482', // light orange
    '#8DE5A1', // light green
    '#FF9F9B', // light red
    '#D0BBFF', // light purple
    '#DEBB9B', // light brown
    '#FAB0E4', // light pink
    '#CFCFCF', // light gray
    '#FFFEA3', // light yellow
    '#B9F2F0', // light cyan
    // Additional 15 pastel colors (soft and desaturated)
    '#C5CAE9', // lavender blue
    '#FFCCBC', // peach
    '#C8E6C9', // mint green
    '#F8BBD0', // blush pink
    '#D1C4E9', // lilac
    '#FFE0B2', // cream
    '#B2EBF2', // sky blue
    '#DCEDC8', // pale green
    '#F0F4C3', // pale yellow
    '#FFCDD2', // rose
    '#E1BEE7', // orchid
    '#C5E1A5', // sage
    '#FFD4E5', // cotton candy
    '#D7CCC8', // taupe
    '#B2DFDB'  // seafoam
  ],

  // Seaborn dark palette (darker version) - extended to 25 colors
  dark: [
    // Original 10 colors (unchanged for backward compatibility)
    '#023EFF', // dark blue
    '#FF7C00', // dark orange
    '#1AC938', // dark green
    '#E8000B', // dark red
    '#8B2BE2', // dark purple
    '#9F4800', // dark brown
    '#F14CC1', // dark pink
    '#4D4D4D', // dark gray
    '#FFC400', // dark yellow
    '#00D7FF', // dark cyan
    // Additional 15 dark but vibrant colors
    '#0000CD', // medium blue
    '#FF4500', // orange red
    '#228B22', // forest green
    '#8B008B', // dark magenta
    '#008B8B', // dark cyan
    '#B22222', // fire brick
    '#483D8B', // dark slate blue
    '#2F4F4F', // dark slate gray
    '#FF1493', // deep pink
    '#00CED1', // dark turquoise
    '#9400D3', // violet
    '#FF8C00', // dark orange
    '#8B0000', // dark red
    '#006400', // dark green
    '#4B0082'  // indigo
  ],

  // Colorblind safe palette (Paul Tol's palette) - extended to 25 colors
  colorblind: [
    // Original 10 colors (unchanged for backward compatibility)
    '#0173B2', // blue
    '#DE8F05', // orange
    '#029E73', // green
    '#CC78BC', // pink
    '#CA9161', // brown
    '#FBAFE4', // light pink
    '#949494', // gray
    '#ECE133', // yellow
    '#56B4E9', // light blue
    '#208B3A', // dark green
    // Additional 15 colorblind-safe colors (based on Paul Tol's research)
    '#882E72', // purple
    '#B178A6', // light purple
    '#D6C1DE', // very light purple
    '#1965B0', // strong blue
    '#5289C7', // medium blue
    '#7BAFDE', // light blue
    '#4EB265', // green
    '#90C987', // light green
    '#CAE0AB', // pale green
    '#F7F056', // bright yellow
    '#F6C141', // orange yellow
    '#F1932D', // orange
    '#E8601C', // red orange
    '#DC050C', // red
    '#777777'  // medium gray
  ],

  // HUSL palette - perceptually uniform colors - extended to 25 colors
  husl: [
    // Original 10 colors (unchanged for backward compatibility)
    '#F77189', // red
    '#BB9832', // yellow
    '#50B131', // green
    '#36ADA4', // cyan
    '#3BA3EC', // blue
    '#8B7AA8', // purple
    '#E85B7A', // pink
    '#9C9C9C', // gray
    '#C29D4F', // tan
    '#5FBCD3', // light blue
    // Additional 15 perceptually uniform colors (HUSL color space)
    '#F7914D', // orange
    '#E8A838', // golden
    '#A4C61A', // lime
    '#5DC963', // light green
    '#00C9A7', // mint
    '#00B4D8', // sky blue
    '#0096C7', // ocean blue
    '#7678ED', // periwinkle
    '#9D4EDD', // violet
    '#C77DFF', // lavender
    '#E0AAFF', // mauve
    '#F48C8C', // coral
    '#F9C74F', // saffron
    '#80B918', // chartreuse
    '#52B788'  // emerald
  ]
};

// Sequential palettes for continuous data
export const SEQUENTIAL_PALETTES: Record<SequentialPaletteName, string[]> = {
  // Rocket palette - dark to light (seaborn's rocket)
  rocket: [
    '#000428', // very dark blue
    '#1a1f71', // dark blue
    '#3d3393', // blue-purple
    '#6b4c9a', // purple
    '#9c6591', // pink-purple
    '#c97d84', // pink
    '#eb9f7e', // orange-pink
    '#fdc086', // light orange
    '#fee8b6', // light yellow
    '#ffffd4' // very light yellow
  ],

  // Viridis palette - scientific standard
  viridis: [
    '#440154', // dark purple
    '#482878', // purple
    '#3e4989', // blue-purple
    '#31688e', // blue
    '#26828e', // blue-green
    '#1f9e89', // green-blue
    '#35b779', // green
    '#6ece58', // light green
    '#b5de2b', // yellow-green
    '#fde725' // yellow
  ],

  // Blues palette - single hue
  blues: [
    '#f7fbff', // very light blue
    '#deebf7', // light blue
    '#c6dbef', //
    '#9ecae1', //
    '#6baed6', //
    '#4292c6', // medium blue
    '#2171b5', //
    '#08519c', //
    '#08306b' // dark blue
  ],

  // Reds palette - single hue
  reds: [
    '#fff5f0', // very light red
    '#fee0d2', // light red
    '#fcbba1', //
    '#fc9272', //
    '#fb6a4a', //
    '#ef3b2c', // medium red
    '#cb181d', //
    '#a50f15', //
    '#67000d' // dark red
  ],

  // Crest palette - blue to purple (seaborn's crest)
  crest: [
    '#f0f9ff', // very light blue
    '#d0e7f7', // light blue
    '#a8d5e2', //
    '#7dc0d4', //
    '#4fa8c5', //
    '#2e8ab5', // medium blue
    '#236ba3', //
    '#22508c', // blue-purple
    '#1e3670', // dark blue-purple
    '#071e58' // very dark purple
  ],

  // Mako palette - blue to green (seaborn's mako)
  mako: [
    '#0B0405', // very dark
    '#1A1A2D', // dark blue
    '#233447', //
    '#1F5061', //
    '#166B7D', // blue-green
    '#0C8B8C', //
    '#14A789', // green-blue
    '#3DBC74', // green
    '#85CE58', // light green
    '#DEF5E5' // very light green
  ],

  // Flare palette - yellow to purple (seaborn's flare)
  flare: [
    '#E3F2FD', // very light blue
    '#E8D4F1', // light purple
    '#F0B9D2', // light pink
    '#F6A192', // light orange
    '#F68C57', // orange
    '#F47F17', // orange-yellow
    '#F8870E', // yellow-orange
    '#FA9B0E', // yellow
    '#FDB417', // bright yellow
    '#FFD125' // light yellow
  ]
};

// Helper function to get a qualitative palette
export function getQualitativePalette(name: QualitativePaletteName = 'deep'): string[] {
  return QUALITATIVE_PALETTES[name] || QUALITATIVE_PALETTES.deep;
}

// Helper function to get a sequential palette
export function getSequentialPalette(name: SequentialPaletteName = 'rocket'): string[] {
  return SEQUENTIAL_PALETTES[name] || SEQUENTIAL_PALETTES.rocket;
}

/**
 * Get a color from the specified qualitative palette
 * @param index - Index of the color (will wrap around if > palette length)
 * @param paletteName - Name of the qualitative palette to use
 * @returns Hex color string
 */
export function getQualitativeColor(index: number, paletteName: QualitativePaletteName = 'deep'): string {
  const palette = getQualitativePalette(paletteName);
  return palette[index % palette.length];
}

/**
 * Interpolate a color from the sequential palette based on a normalized value
 * @param value - Normalized value between 0 and 1
 * @param paletteName - Name of the sequential palette to use
 * @returns Hex color string
 */
export function getSequentialColor(value: number, paletteName: SequentialPaletteName = 'rocket'): string {
  // Clamp value between 0 and 1
  const normalizedValue = Math.max(0, Math.min(1, value));
  const palette = getSequentialPalette(paletteName);

  // Calculate position in palette
  const paletteIndex = normalizedValue * (palette.length - 1);
  const lowerIndex = Math.floor(paletteIndex);
  const upperIndex = Math.ceil(paletteIndex);

  // If we're exactly on a color, return it
  if (lowerIndex === upperIndex) {
    return palette[lowerIndex];
  }

  // Otherwise, interpolate between two colors
  const ratio = paletteIndex - lowerIndex;
  return interpolateColors(
    palette[lowerIndex],
    palette[upperIndex],
    ratio
  );
}

/**
 * Interpolate between two hex colors
 * @param color1 - First hex color
 * @param color2 - Second hex color
 * @param ratio - Interpolation ratio (0-1)
 * @returns Interpolated hex color
 */
function interpolateColors(color1: string, color2: string, ratio: number): string {
  // Convert hex to RGB
  const rgb1 = hexToRgb(color1);
  const rgb2 = hexToRgb(color2);

  if (!rgb1 || !rgb2) {
    return color1; // Fallback
  }

  // Interpolate RGB values
  const r = Math.round(rgb1.r + (rgb2.r - rgb1.r) * ratio);
  const g = Math.round(rgb1.g + (rgb2.g - rgb1.g) * ratio);
  const b = Math.round(rgb1.b + (rgb2.b - rgb1.b) * ratio);

  // Convert back to hex
  return rgbToHex(r, g, b);
}

/**
 * Convert hex color to RGB
 */
function hexToRgb(hex: string): { r: number; g: number; b: number } | null {
  const result = /^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i.exec(hex);
  return result
    ? {
        r: parseInt(result[1], 16),
        g: parseInt(result[2], 16),
        b: parseInt(result[3], 16)
      }
    : null;
}

/**
 * Convert RGB to hex color
 */
function rgbToHex(r: number, g: number, b: number): string {
  return '#' + ((1 << 24) + (r << 16) + (g << 8) + b).toString(16).slice(1);
}

/**
 * Get a color scale for continuous data
 * @param value - The data value
 * @param min - Minimum value in the data range
 * @param max - Maximum value in the data range
 * @param paletteName - Name of the sequential palette to use
 * @returns Hex color string
 */
export function getSequentialColorScale(
  value: number,
  min: number,
  max: number,
  paletteName: SequentialPaletteName = 'rocket'
): string {
  if (max === min) {
    return getSequentialColor(0.5, paletteName); // Middle color if no range
  }

  const normalized = (value - min) / (max - min);
  return getSequentialColor(normalized, paletteName);
}

/**
 * Create a color map for groups using the specified qualitative palette
 * @param groupLabels - Array of group labels
 * @param paletteName - Name of the qualitative palette to use
 * @returns Map of group label to color
 */
export function createQualitativeColorMap(
  groupLabels: string[],
  paletteName: QualitativePaletteName = 'deep'
): Map<string, string> {
  const uniqueGroups = [...new Set(groupLabels)].sort();
  const colorMap = new Map<string, string>();

  uniqueGroups.forEach((group, index) => {
    colorMap.set(group, getQualitativeColor(index, paletteName));
  });

  return colorMap;
}

// For backward compatibility - keep old function signatures but use default palettes
export const QUALITATIVE_PALETTE = QUALITATIVE_PALETTES.deep;