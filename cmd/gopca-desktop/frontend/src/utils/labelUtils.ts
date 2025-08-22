// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

/**
 * Utility functions for label display in plots
 * Follows DRY principle - shared across ScoresPlot, Biplot, and DiagnosticScatterPlot
 */

/**
 * Calculate which points should have labels based on distance from origin
 * @param data Array of points with x, y coordinates and index
 * @param showLabels Whether labels should be shown at all
 * @param maxLabels Maximum number of labels to show
 * @returns Set of indices for points that should have labels
 */
export function calculateTopPoints(
  data: Array<{ x: number; y: number; index: number }>,
  showLabels: boolean,
  maxLabels: number
): Set<number> {
  if (!showLabels) {
return new Set<number>();
}

  const pointsWithDistance = data.map(point => ({
    index: point.index,
    distance: Math.sqrt(point.x ** 2 + point.y ** 2)
  }));

  // Sort by distance and get top N
  const topIndices = pointsWithDistance
    .sort((a, b) => b.distance - a.distance)
    .slice(0, maxLabels)
    .map(p => p.index);

  return new Set(topIndices);
}

/**
 * Determine label position based on point quadrant to avoid overlaps
 * @param x X coordinate of the point
 * @param y Y coordinate of the point
 * @returns Object with textAnchor, dx, and dy values for label positioning
 */
export function getLabelPosition(x: number, y: number): {
  textAnchor: 'start' | 'end' | 'middle';
  dx: number;
  dy: number;
} {
  let textAnchor: 'start' | 'end' | 'middle' = 'start';
  let dx = 8;
  let dy = 0;

  // Adjust horizontal position based on x coordinate
  if (x < 0) {
    textAnchor = 'end';
    dx = -8;
  }

  // Adjust vertical position based on y coordinate
  if (y < 0) {
    dy = 12;
  } else {
    dy = -5;
  }

  return { textAnchor, dx, dy };
}