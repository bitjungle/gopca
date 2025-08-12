// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

import React, { useRef, useState, useCallback } from 'react';
import ReactDOM from 'react-dom';
import { PCAResult } from '../../types';
import { ExportButton } from '../ExportButton';
import { PlotControls } from '../PlotControls';
import { useChartTheme } from '../../hooks/useChartTheme';
import { usePalette } from '../../contexts/PaletteContext';
import { getSequentialColorScale } from '../../utils/colorPalettes';

interface CircleOfCorrelationsProps {
  pcaResult: PCAResult;
  xComponent?: number; // 0-based index
  yComponent?: number; // 0-based index
  threshold?: number; // Minimum loading magnitude to display label
  maxVariables?: number; // Maximum number of variables to display
}

export const CircleOfCorrelations: React.FC<CircleOfCorrelationsProps> = ({ 
  pcaResult, 
  xComponent = 0,
  yComponent = 1,
  threshold = 0.3,
  maxVariables = 100
}) => {
  const chartRef = useRef<HTMLDivElement>(null);
  const fullscreenRef = useRef<HTMLDivElement>(null);
  const svgRef = useRef<SVGSVGElement>(null);
  const [isFullscreen, setIsFullscreen] = useState(false);
  const [hoveredVariable, setHoveredVariable] = useState<number | null>(null);
  const [tooltipPosition, setTooltipPosition] = useState({ x: 0, y: 0 });
  const chartTheme = useChartTheme();
  const { sequentialPalette } = usePalette();
  
  // Check if loadings are available (not available for Kernel PCA)
  if (!pcaResult.loadings || pcaResult.loadings.length === 0) {
    return (
      <div className="w-full h-full flex items-center justify-center text-gray-400">
        <p>Loadings are not available for this PCA method</p>
      </div>
    );
  }

  // Extract loadings for selected components
  const allLoadings = pcaResult.loadings.map((row, index) => {
    const x = row[xComponent] || 0;
    const y = row[yComponent] || 0;
    const magnitude = Math.sqrt(x * x + y * y);
    
    return {
      variable: pcaResult.variable_labels?.[index] || `Var${index + 1}`,
      x,
      y,
      magnitude,
      angle: Math.atan2(y, x) * 180 / Math.PI
    };
  });

  // Check if filtering is needed
  const needsFiltering = pcaResult.loadings.length > maxVariables;
  
  // Filter loadings if needed
  const filteredLoadings = needsFiltering
    ? [...allLoadings]
        .sort((a, b) => b.magnitude - a.magnitude)
        .slice(0, maxVariables)
    : allLoadings;

  // Find maximum magnitude for scaling
  const maxMagnitude = Math.max(...filteredLoadings.map(l => l.magnitude));
  
  // Scale factor to ensure visibility (scale so max magnitude reaches 90% of circle)
  const scaleFactor = maxMagnitude > 0 ? 0.9 / maxMagnitude : 1;
  
  // Apply scaling to filtered loadings
  const loadings = filteredLoadings.map((loading, idx) => ({
    ...loading,
    idx,  // Add index for hover tracking
    originalX: loading.x,  // Preserve original values for tooltip
    originalY: loading.y,
    x: loading.x * scaleFactor,
    y: loading.y * scaleFactor,
    scaledMagnitude: loading.magnitude * scaleFactor
  }));

  // Get component labels and variance
  const xLabel = pcaResult.component_labels?.[xComponent] || `PC${xComponent + 1}`;
  const yLabel = pcaResult.component_labels?.[yComponent] || `PC${yComponent + 1}`;
  const xVariance = pcaResult.explained_variance_ratio[xComponent]?.toFixed(1) || '0';
  const yVariance = pcaResult.explained_variance_ratio[yComponent]?.toFixed(1) || '0';

  // SVG dimensions
  const width = 500;
  const height = 500;
  const padding = 60;
  const radius = (Math.min(width, height) - 2 * padding) / 2;
  const centerX = width / 2;
  const centerY = height / 2;

  // Color scale based on magnitude using sequential palette
  const getColor = (scaledMagnitude: number) => {
    // Use sequential palette for correlation intensity (0 to 1)
    // Clamp to [0, 1] range for safety
    const normalizedMag = Math.min(1, Math.max(0, scaledMagnitude));
    return getSequentialColorScale(normalizedMag, 0, 1, sequentialPalette);
  };

  const handleToggleFullscreen = useCallback(() => {
    if (!fullscreenRef.current) return;
    
    if (!isFullscreen) {
      if (fullscreenRef.current.requestFullscreen) {
        fullscreenRef.current.requestFullscreen();
      }
    } else {
      if (document.exitFullscreen) {
        document.exitFullscreen();
      }
    }
    
    setIsFullscreen(!isFullscreen);
  }, [isFullscreen]);

  const handleResetView = useCallback(() => {
    // No zoom functionality for this plot, but keeping for consistency
  }, []);

  return (
    <div ref={fullscreenRef} className={`w-full h-full ${isFullscreen ? 'fixed inset-0 z-50 bg-white dark:bg-gray-900 p-4' : ''}`}>
      <div className="w-full h-full" ref={chartRef}>
        {/* Header with export button */}
        <div className="flex items-center justify-between mb-4">
          <h4 className="text-md font-medium text-gray-700 dark:text-gray-300">
            Circle of Correlations: {xLabel} ({xVariance}%) vs {yLabel} ({yVariance}%)
            {needsFiltering && (
              <span className="ml-2 text-sm text-amber-600 dark:text-amber-400">
                (showing top {maxVariables} of {pcaResult.loadings.length} variables)
              </span>
            )}
          </h4>
          <div className="flex items-center gap-2">
            <PlotControls 
              onResetView={handleResetView}
              onToggleFullscreen={handleToggleFullscreen}
              isFullscreen={isFullscreen}
            />
            <ExportButton 
              chartRef={chartRef} 
              fileName={`circle-of-correlations-${xLabel}-${yLabel}`}
            />
          </div>
        </div>

      {/* SVG visualization */}
      <div className="flex justify-center items-center h-full relative">
        <svg ref={svgRef} width={width} height={height} className="bg-white dark:bg-gray-800 rounded-lg shadow-lg">
          {/* Background circle */}
          <circle
            cx={centerX}
            cy={centerY}
            r={radius}
            fill="none"
            stroke={chartTheme.gridColor}
            strokeWidth="2"
          />
          
          {/* Grid lines */}
          <line
            x1={padding}
            y1={centerY}
            x2={width - padding}
            y2={centerY}
            stroke={chartTheme.gridColor}
            strokeWidth="1"
            strokeDasharray="3,3"
          />
          <line
            x1={centerX}
            y1={padding}
            x2={centerX}
            y2={height - padding}
            stroke={chartTheme.gridColor}
            strokeWidth="1"
            strokeDasharray="3,3"
          />
          
          {/* Inner circles at 0.5 and 0.75 */}
          <circle
            cx={centerX}
            cy={centerY}
            r={radius * 0.5}
            fill="none"
            stroke={chartTheme.gridColor}
            strokeWidth="1"
            strokeDasharray="2,2"
            opacity="0.5"
          />
          <circle
            cx={centerX}
            cy={centerY}
            r={radius * 0.75}
            fill="none"
            stroke={chartTheme.gridColor}
            strokeWidth="1"
            strokeDasharray="2,2"
            opacity="0.5"
          />
          
          {/* Axis labels */}
          <text
            x={width - padding + 10}
            y={centerY + 5}
            fontSize="14"
            fill={chartTheme.textColor}
            textAnchor="start"
          >
            {xLabel}
          </text>
          <text
            x={centerX - 5}
            y={padding - 10}
            fontSize="14"
            fill={chartTheme.textColor}
            textAnchor="middle"
          >
            {yLabel}
          </text>
          
          {/* Loading vectors */}
          {loadings.map((loading, index) => {
            const endX = centerX + loading.x * radius;
            const endY = centerY - loading.y * radius; // Invert Y for standard orientation
            const isHovered = hoveredVariable === loading.idx;
            const color = isHovered ? '#EF4444' : getColor(loading.scaledMagnitude);
            
            return (
              <g key={index}>
                {/* Vector line */}
                <line
                  x1={centerX}
                  y1={centerY}
                  x2={endX}
                  y2={endY}
                  stroke={color}
                  strokeWidth={isHovered ? "3" : "2"}
                  markerEnd={`url(#arrowhead-${index})`}
                  style={{ transition: 'stroke-width 0.2s, stroke 0.2s' }}
                />
                
                {/* Arrowhead marker */}
                <defs>
                  <marker
                    id={`arrowhead-${index}`}
                    markerWidth="10"
                    markerHeight="10"
                    refX="9"
                    refY="3"
                    orient="auto"
                    markerUnits="strokeWidth"
                  >
                    <path
                      d="M0,0 L0,6 L9,3 z"
                      fill={color}
                    />
                  </marker>
                </defs>
                
                {/* Variable label - show all for small datasets, filter by threshold for large ones */}
                {(loadings.length <= 20 || loading.magnitude > threshold) && (
                  <text
                    x={endX}
                    y={endY}
                    dx={loading.x > 0 ? 5 : -5}
                    dy={loading.y > 0 ? -5 : 10}
                    fontSize="12"
                    fill={chartTheme.textColor}
                    textAnchor={loading.x > 0 ? "start" : "end"}
                    className="select-none"
                  >
                    {loading.variable}
                  </text>
                )}
                
                {/* Hover target - larger invisible circle for easier hovering */}
                <circle
                  cx={endX}
                  cy={endY}
                  r="10"
                  fill="transparent"
                  className="cursor-pointer"
                  onMouseEnter={(e) => {
                    setHoveredVariable(loading.idx);
                    // Get mouse position relative to the page
                    const rect = e.currentTarget.getBoundingClientRect();
                    setTooltipPosition({
                      x: rect.left + rect.width / 2,
                      y: rect.top
                    });
                  }}
                  onMouseMove={(e) => {
                    if (hoveredVariable === loading.idx) {
                      const rect = e.currentTarget.getBoundingClientRect();
                      setTooltipPosition({
                        x: rect.left + rect.width / 2,
                        y: rect.top
                      });
                    }
                  }}
                  onMouseLeave={() => setHoveredVariable(null)}
                />
                
                {/* Visible dot at arrow end */}
                <circle
                  cx={endX}
                  cy={endY}
                  r={isHovered ? "4" : "2"}
                  fill={color}
                  style={{ transition: 'r 0.2s' }}
                  pointerEvents="none"
                />
              </g>
            );
          })}
          
          {/* Scale indicators */}
          <text
            x={centerX + radius * 0.5}
            y={centerY + 15}
            fontSize="10"
            fill={chartTheme.textColor}
            textAnchor="middle"
            opacity="0.6"
          >
            0.5
          </text>
          <text
            x={centerX + radius * 0.75}
            y={centerY + 15}
            fontSize="10"
            fill={chartTheme.textColor}
            textAnchor="middle"
            opacity="0.6"
          >
            0.75
          </text>
          <text
            x={centerX + radius}
            y={centerY + 15}
            fontSize="10"
            fill={chartTheme.textColor}
            textAnchor="middle"
            opacity="0.6"
          >
            1.0
          </text>
        </svg>
        
        {/* Tooltip Portal */}
        {hoveredVariable !== null && (() => {
          const hoveredLoading = loadings.find(l => l.idx === hoveredVariable);
          if (!hoveredLoading) return null;
          
          // Calculate tooltip position
          const tooltipHeight = 120;
          const tooltipWidth = 200;
          const padding = 10;
          
          let x = tooltipPosition.x;
          let y = tooltipPosition.y;
          
          // Adjust horizontal position
          if (x + tooltipWidth / 2 > window.innerWidth - padding) {
            x = window.innerWidth - tooltipWidth - padding;
          } else if (x - tooltipWidth / 2 < padding) {
            x = tooltipWidth / 2 + padding;
          }
          
          // Always show above the point to avoid covering the arrow
          y = y - tooltipHeight - 20;
          
          // If it would go off the top, show below instead
          if (y < padding) {
            y = tooltipPosition.y + 20;
          }
          
          return ReactDOM.createPortal(
            <div
              className="fixed z-50 p-3 rounded shadow-lg border pointer-events-none"
              style={{
                backgroundColor: chartTheme.tooltipBackgroundColor,
                borderColor: chartTheme.tooltipBorderColor,
                left: x,
                top: y,
                transform: 'translateX(-50%)',
                minWidth: '180px'
              }}
            >
              <p className="font-semibold mb-1" style={{ color: chartTheme.tooltipTextColor }}>
                {hoveredLoading.variable}
              </p>
              <div className="text-sm space-y-1">
                <p style={{ color: chartTheme.tooltipTextColor }}>
                  {xLabel}: {hoveredLoading.originalX.toFixed(4)}
                </p>
                <p style={{ color: chartTheme.tooltipTextColor }}>
                  {yLabel}: {hoveredLoading.originalY.toFixed(4)}
                </p>
                <p style={{ color: chartTheme.tooltipTextColor }}>
                  Magnitude: {hoveredLoading.magnitude.toFixed(4)}
                </p>
              </div>
            </div>,
            document.body
          );
        })()}
      </div>
      
      {/* Legend */}
      <div className="mt-4 text-sm text-gray-600 dark:text-gray-400 text-center">
        <p>
          Arrow length indicates correlation strength with the PCs
          {scaleFactor !== 1 && scaleFactor > 1.1 && (
            <span className="ml-2 text-xs text-amber-600 dark:text-amber-400">
              (scaled {scaleFactor.toFixed(1)}× for visibility)
            </span>
          )}
        </p>
        <p>Color gradient indicates correlation strength (low to high)</p>
      </div>
    </div>
    </div>
  );
};