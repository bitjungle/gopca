// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

import React, { useMemo } from 'react';
import { EllipseParams } from '../types';
import { generateEllipsePoints, createScale, pointsToPath } from '../utils/ellipseUtils';

/**
 * Reusable component for rendering confidence ellipses as an SVG overlay
 * Used in ScoresPlot and Biplot
 * Follows DRY principle - single implementation for all plots with ellipses
 */
interface EllipseOverlayProps {
  groupEllipses: Record<string, EllipseParams>;
  groupColorMap: Map<string, string>;
  xDomain: [number, number];
  yDomain: [number, number];
  containerSize: { width: number; height: number };
  margins?: { top: number; right: number; bottom: number; left: number };
}

export const EllipseOverlay: React.FC<EllipseOverlayProps> = ({
  groupEllipses,
  groupColorMap,
  xDomain,
  yDomain,
  containerSize,
  margins = { top: 20, right: 20, bottom: 60, left: 80 }
}) => {
  // Calculate scale functions
  const { xScale, yScale } = useMemo(() => {
    const plotWidth = containerSize.width - margins.left - margins.right;
    const plotHeight = containerSize.height - margins.top - margins.bottom;
    
    return {
      xScale: createScale(xDomain, plotWidth, margins.left, false),
      yScale: createScale(yDomain, plotHeight, margins.top, true) // Y is inverted in SVG
    };
  }, [xDomain, yDomain, containerSize, margins]);
  
  return (
    <svg 
      className="absolute inset-0 pointer-events-none" 
      style={{ width: '100%', height: '100%' }}
    >
      {Object.entries(groupEllipses).map(([group, ellipse]) => {
        const color = groupColorMap.get(group) || '#888888';
        const points = generateEllipsePoints(ellipse);
        const pathData = pointsToPath(points, xScale, yScale);
        
        return (
          <path
            key={`ellipse-${group}`}
            d={pathData}
            fill={color}
            fillOpacity={0.1}
            stroke={color}
            strokeWidth={2}
            strokeOpacity={0.8}
            strokeDasharray="5,5"
          />
        );
      })}
    </svg>
  );
};