// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

import React from 'react';
import { getLabelPosition } from '../utils/labelUtils';

/**
 * Reusable custom dot component that can render labels
 * Used in ScoresPlot, Biplot, and DiagnosticScatterPlot
 * Follows DRY principle - single implementation for all plots
 */
interface CustomPointWithLabelProps {
  cx?: number;
  cy?: number;
  payload?: any;
  topPoints: Set<number>;
  hoveredPoint: number | null;
  showLabels: boolean;
  onMouseEnter: (index: number) => void;
  onMouseLeave: () => void;
  chartTheme: any;
  fontSize?: number;
}

export const CustomPointWithLabel: React.FC<CustomPointWithLabelProps> = ({
  cx = 0,
  cy = 0,
  payload,
  topPoints,
  hoveredPoint,
  showLabels,
  onMouseEnter,
  onMouseLeave,
  chartTheme,
  fontSize = 11
}) => {
  if (!payload) return null;
  
  const isTopPoint = topPoints.has(payload.index);
  const isHovered = hoveredPoint === payload.index;
  const shouldShowLabel = showLabels && (isTopPoint || isHovered);
  
  // Get label position based on quadrant
  const { textAnchor, dx, dy } = getLabelPosition(payload.x, payload.y);
  
  return (
    <g>
      <circle
        cx={cx}
        cy={cy}
        r={4}
        fill={payload.color || '#3B82F6'}
        fillOpacity={0.8}
        stroke={payload.color || '#1E40AF'}
        strokeWidth={1}
        onMouseEnter={() => onMouseEnter(payload.index)}
        onMouseLeave={onMouseLeave}
        style={{ cursor: 'pointer' }}
      />
      {shouldShowLabel && (
        <text
          x={cx + dx}
          y={cy + dy}
          fill={chartTheme.textColor}
          fontSize={fontSize}
          fontWeight={isHovered ? '600' : '400'}
          textAnchor={textAnchor}
          dominantBaseline="middle"
          style={{ pointerEvents: 'none', userSelect: 'none' }}
        >
          {payload.name}
        </text>
      )}
    </g>
  );
};