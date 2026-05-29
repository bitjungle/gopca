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

/**
 * Utility functions for ellipse rendering in plots
 * Follows DRY principle - shared across ScoresPlot and Biplot
 */

import { EllipseParams } from '../types';

/**
 * Generate SVG path points for an ellipse
 * @param ellipse Ellipse parameters (center, axes, angle)
 * @param steps Number of points to generate (default 50)
 * @returns Array of {x, y} points forming the ellipse path
 */
export function generateEllipsePoints(ellipse: EllipseParams, steps: number = 50): Array<{ x: number; y: number }> {
  const { centerX, centerY, majorAxis, minorAxis, angle } = ellipse;
  const points = [];

  for (let i = 0; i <= steps; i++) {
    const t = (i / steps) * 2 * Math.PI;
    // Ellipse in local coordinates
    const x = majorAxis * Math.cos(t);
    const y = minorAxis * Math.sin(t);

    // Apply rotation
    const rotatedX = x * Math.cos(angle) - y * Math.sin(angle);
    const rotatedY = x * Math.sin(angle) + y * Math.cos(angle);

    // Translate to center
    points.push({
      x: centerX + rotatedX,
      y: centerY + rotatedY
    });
  }

  return points;
}

/**
 * Create scale functions to convert data coordinates to pixel coordinates
 * @param domain Current domain [min, max]
 * @param range Pixel range (width or height minus margins)
 * @param margin Leading margin (left for X, top for Y)
 * @param inverted Whether to invert the scale (true for Y axis in SVG)
 * @returns Scale function that converts data value to pixel coordinate
 */
export function createScale(
  domain: [number, number],
  range: number,
  margin: number,
  inverted: boolean = false
): (value: number) => number {
  return (value: number) => {
    const ratio = (value - domain[0]) / (domain[1] - domain[0]);
    if (inverted) {
      return margin + range - ratio * range;
    } else {
      return margin + ratio * range;
    }
  };
}

/**
 * Convert ellipse points to SVG path data string
 * @param points Array of {x, y} points
 * @param xScale Scale function for X coordinates
 * @param yScale Scale function for Y coordinates
 * @returns SVG path data string
 */
export function pointsToPath(
  points: Array<{ x: number; y: number }>,
  xScale: (value: number) => number,
  yScale: (value: number) => number
): string {
  return points
    .map((point, index) => {
      const x = xScale(point.x);
      const y = yScale(point.y);
      return index === 0 ? `M ${x} ${y}` : `L ${x} ${y}`;
    })
    .join(' ') + ' Z';
}