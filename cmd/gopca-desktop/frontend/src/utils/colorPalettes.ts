// Color palette definitions for data visualization
// Based on seaborn color palettes

export type PaletteType = 'qualitative' | 'sequential';

// Seaborn default/deep palette - for categorical data
export const QUALITATIVE_PALETTE = [
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
];

// Rocket-inspired sequential palette - for continuous data
// Dark to light gradient
export const SEQUENTIAL_PALETTE = [
  '#000428', // very dark blue
  '#1a1f71', // dark blue
  '#3d3393', // blue-purple
  '#6b4c9a', // purple
  '#9c6591', // pink-purple
  '#c97d84', // pink
  '#eb9f7e', // orange-pink
  '#fdc086', // light orange
  '#fee8b6', // light yellow
  '#ffffd4', // very light yellow
];

/**
 * Get a color from the qualitative palette
 * @param index - Index of the color (will wrap around if > palette length)
 * @returns Hex color string
 */
export function getQualitativeColor(index: number): string {
  return QUALITATIVE_PALETTE[index % QUALITATIVE_PALETTE.length];
}

/**
 * Interpolate a color from the sequential palette based on a normalized value
 * @param value - Normalized value between 0 and 1
 * @returns Hex color string
 */
export function getSequentialColor(value: number): string {
  // Clamp value between 0 and 1
  const normalizedValue = Math.max(0, Math.min(1, value));
  
  // Calculate position in palette
  const paletteIndex = normalizedValue * (SEQUENTIAL_PALETTE.length - 1);
  const lowerIndex = Math.floor(paletteIndex);
  const upperIndex = Math.ceil(paletteIndex);
  
  // If we're exactly on a color, return it
  if (lowerIndex === upperIndex) {
    return SEQUENTIAL_PALETTE[lowerIndex];
  }
  
  // Otherwise, interpolate between two colors
  const ratio = paletteIndex - lowerIndex;
  return interpolateColors(
    SEQUENTIAL_PALETTE[lowerIndex],
    SEQUENTIAL_PALETTE[upperIndex],
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
        b: parseInt(result[3], 16),
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
 * @returns Hex color string
 */
export function getSequentialColorScale(value: number, min: number, max: number): string {
  if (max === min) {
    return getSequentialColor(0.5); // Middle color if no range
  }
  
  const normalized = (value - min) / (max - min);
  return getSequentialColor(normalized);
}

/**
 * Create a color map for groups using the qualitative palette
 * @param groupLabels - Array of group labels
 * @returns Map of group label to color
 */
export function createQualitativeColorMap(groupLabels: string[]): Map<string, string> {
  const uniqueGroups = [...new Set(groupLabels)].sort();
  const colorMap = new Map<string, string>();
  
  uniqueGroups.forEach((group, index) => {
    colorMap.set(group, getQualitativeColor(index));
  });
  
  return colorMap;
}

/**
 * Get color based on palette type
 * @param paletteType - Type of palette to use
 * @param index - Index or normalized value
 * @param min - Minimum value (for sequential)
 * @param max - Maximum value (for sequential)
 * @returns Hex color string
 */
export function getColorFromPalette(
  paletteType: PaletteType,
  index: number,
  min?: number,
  max?: number
): string {
  if (paletteType === 'qualitative') {
    return getQualitativeColor(Math.floor(index));
  } else {
    // For sequential, if min/max provided, use them for normalization
    if (min !== undefined && max !== undefined) {
      return getSequentialColorScale(index, min, max);
    }
    // Otherwise treat index as already normalized (0-1)
    return getSequentialColor(index);
  }
}