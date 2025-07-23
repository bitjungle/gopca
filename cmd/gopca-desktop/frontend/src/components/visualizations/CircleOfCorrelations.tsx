import React, { useRef } from 'react';
import { PCAResult } from '../../types';
import { ExportButton } from '../ExportButton';
import { useChartTheme } from '../../hooks/useChartTheme';

interface CircleOfCorrelationsProps {
  pcaResult: PCAResult;
  xComponent?: number; // 0-based index
  yComponent?: number; // 0-based index
  threshold?: number; // Minimum loading magnitude to display label
}

export const CircleOfCorrelations: React.FC<CircleOfCorrelationsProps> = ({ 
  pcaResult, 
  xComponent = 0,
  yComponent = 1,
  threshold = 0.3
}) => {
  const chartRef = useRef<HTMLDivElement>(null);
  const chartTheme = useChartTheme();
  
  // Check if loadings are available (not available for Kernel PCA)
  if (!pcaResult.loadings || pcaResult.loadings.length === 0) {
    return (
      <div className="w-full h-full flex items-center justify-center text-gray-400">
        <p>Loadings are not available for this PCA method</p>
      </div>
    );
  }

  // Extract loadings for selected components
  const loadings = pcaResult.loadings.map((row, index) => {
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

  // Get component labels and variance
  const xLabel = pcaResult.component_labels?.[xComponent] || `PC${xComponent + 1}`;
  const yLabel = pcaResult.component_labels?.[yComponent] || `PC${yComponent + 1}`;
  const xVariance = pcaResult.explained_variance[xComponent]?.toFixed(1) || '0';
  const yVariance = pcaResult.explained_variance[yComponent]?.toFixed(1) || '0';

  // SVG dimensions
  const width = 500;
  const height = 500;
  const padding = 60;
  const radius = (Math.min(width, height) - 2 * padding) / 2;
  const centerX = width / 2;
  const centerY = height / 2;

  // Color scale based on magnitude
  const getColor = (magnitude: number) => {
    // Use a gradient from blue (low) to red (high)
    const hue = (1 - magnitude) * 240; // 240 is blue, 0 is red
    return `hsl(${hue}, 70%, 50%)`;
  };

  return (
    <div className="w-full h-full" ref={chartRef}>
      {/* Header with export button */}
      <div className="flex items-center justify-between mb-4">
        <h4 className="text-md font-medium text-gray-700 dark:text-gray-300">
          Circle of Correlations: {xLabel} ({xVariance}%) vs {yLabel} ({yVariance}%)
        </h4>
        <ExportButton 
          chartRef={chartRef} 
          fileName={`circle-of-correlations-${xLabel}-${yLabel}`}
        />
      </div>

      {/* SVG visualization */}
      <div className="flex justify-center items-center h-full">
        <svg width={width} height={height} className="bg-white dark:bg-gray-800 rounded-lg shadow-lg">
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
            const color = getColor(loading.magnitude);
            
            return (
              <g key={index}>
                {/* Vector line */}
                <line
                  x1={centerX}
                  y1={centerY}
                  x2={endX}
                  y2={endY}
                  stroke={color}
                  strokeWidth="2"
                  markerEnd={`url(#arrowhead-${index})`}
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
                
                {/* Variable label (only for significant loadings) */}
                {loading.magnitude > threshold && (
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
                
                {/* Tooltip on hover */}
                <circle
                  cx={endX}
                  cy={endY}
                  r="5"
                  fill={color}
                  opacity="0"
                  className="cursor-pointer"
                >
                  <title>
                    {loading.variable}&#10;
                    {xLabel}: {loading.x.toFixed(3)}&#10;
                    {yLabel}: {loading.y.toFixed(3)}&#10;
                    Length: {loading.magnitude.toFixed(3)}
                  </title>
                </circle>
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
      </div>
      
      {/* Legend */}
      <div className="mt-4 text-sm text-gray-600 dark:text-gray-400 text-center">
        <p>Arrow length indicates correlation strength with the PCs</p>
        <p>Color: <span style={{ color: 'hsl(240, 70%, 50%)' }}>■</span> Low correlation → 
           <span style={{ color: 'hsl(120, 70%, 50%)' }}> ■</span> Medium → 
           <span style={{ color: 'hsl(0, 70%, 50%)' }}> ■</span> High correlation</p>
      </div>
    </div>
  );
};