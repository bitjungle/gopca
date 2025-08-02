import React, { useRef, useState, useCallback, useMemo } from 'react';
import { ComposedChart, Scatter, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, ReferenceLine, Cell } from 'recharts';
import { PCAResult, EllipseParams } from '../../types';
import { ExportButton } from '../ExportButton';
import { PlotControls } from '../PlotControls';
import { useZoomPan } from '../../hooks/useZoomPan';
import { useChartTheme } from '../../hooks/useChartTheme';
import { usePalette } from '../../contexts/PaletteContext';
import { getQualitativeColor, getSequentialColor, createQualitativeColorMap, getSequentialColorScale } from '../../utils/colorPalettes';
import { useEllipses } from '../../hooks/useEllipses';

interface ScoresPlotProps {
  pcaResult: PCAResult;
  rowNames: string[];
  xComponent?: number; // 0-based index
  yComponent?: number; // 0-based index
  groupColumn?: string | null;
  groupLabels?: string[];
  groupValues?: number[]; // For continuous columns
  groupType?: 'categorical' | 'continuous';
  groupEllipses?: Record<string, EllipseParams>;
  showEllipses?: boolean;
  confidenceLevel?: 0.90 | 0.95 | 0.99;
}

export const ScoresPlot: React.FC<ScoresPlotProps> = ({ 
  pcaResult, 
  rowNames,
  xComponent = 0, 
  yComponent = 1,
  groupColumn,
  groupLabels,
  groupValues,
  groupType = 'categorical',
  groupEllipses,
  showEllipses = false,
  confidenceLevel = 0.95
}) => {
  const containerRef = useRef<HTMLDivElement>(null);
  const fullscreenRef = useRef<HTMLDivElement>(null);
  
  const [isFullscreen, setIsFullscreen] = useState(false);
  const [containerSize, setContainerSize] = useState({ width: 0, height: 0 });
  const chartTheme = useChartTheme();
  const { mode, qualitativePalette, sequentialPalette } = usePalette();
  
  // Create color map for groups based on selected palette
  const groupColorMap = useMemo(() => {
    if (groupType === 'categorical' && groupLabels && groupColumn) {
      return createQualitativeColorMap(groupLabels, qualitativePalette);
    }
    return null;
  }, [groupLabels, groupColumn, qualitativePalette, groupType]);
  
  // Calculate ellipses dynamically if not provided
  const shouldCalculateEllipses = !!(showEllipses && groupType === 'categorical' && groupLabels && !groupEllipses);
  const { 
    ellipses90: dynamicEllipses90,
    ellipses95: dynamicEllipses95,
    ellipses99: dynamicEllipses99,
    isLoading: ellipsesLoading,
    error: ellipsesError
  } = useEllipses({
    scores: pcaResult.scores,
    groupLabels: groupLabels || [],
    xComponent,
    yComponent,
    enabled: shouldCalculateEllipses
  });
  
  // Use provided ellipses or dynamically calculated ones based on confidence level
  const effectiveGroupEllipses = groupEllipses || 
    (showEllipses && (
      confidenceLevel === 0.90 ? dynamicEllipses90 :
      confidenceLevel === 0.95 ? dynamicEllipses95 :
      dynamicEllipses99
    )) || 
    undefined;
  
  // Calculate min/max for continuous values
  const continuousRange = useMemo(() => {
    if (groupType === 'continuous' && groupValues) {
      const validValues = groupValues.filter(v => !isNaN(v) && isFinite(v));
      console.log(`Continuous values - Total: ${groupValues.length}, Valid: ${validValues.length}`);
      if (validValues.length > 0) {
        const range = {
          min: Math.min(...validValues),
          max: Math.max(...validValues)
        };
        console.log(`Range: min=${range.min}, max=${range.max}`);
        return range;
      }
      console.log('No valid values found for continuous range');
    }
    return null;
  }, [groupValues, groupType]);
  
  // Transform scores data for Recharts
  const data = pcaResult.scores.map((row, index) => {
    const xVal = row[xComponent] || 0;
    const yVal = row[yComponent] || 0;
    
    // Check for invalid values
    if (!isFinite(xVal) || !isFinite(yVal)) {
      console.warn(`Invalid values at index ${index}: x=${xVal}, y=${yVal}`);
      return null;
    }
    
    let color = '#3B82F6'; // Default color
    let group = 'Unknown';
    let value: number | undefined;
    const MISSING_VALUE_COLOR = '#9CA3AF'; // Gray color for missing values
    
    if (groupType === 'categorical') {
      const labelValue = groupLabels?.[index];
      if (!labelValue || labelValue === '') {
        group = 'Missing';
        color = MISSING_VALUE_COLOR;
      } else {
        group = labelValue;
        if (groupColorMap) {
          color = groupColorMap.get(group) || color;
        }
      }
    } else if (groupType === 'continuous' && groupValues) {
      const val = groupValues[index];
      value = val;
      if (!isNaN(val) && isFinite(val) && continuousRange) {
        color = getSequentialColorScale(val, continuousRange.min, continuousRange.max, sequentialPalette);
        group = val.toFixed(2); // For display purposes
      } else {
        // Handle missing values explicitly
        color = MISSING_VALUE_COLOR;
        group = 'Missing';
      }
    }
    
    return {
      x: xVal,
      y: yVal,
      name: rowNames[index] || `Sample ${index + 1}`,
      group: group,
      color: color,
      value: value
    };
  }).filter(point => point !== null);
  
  // Generate ellipse path points for Line rendering
  const generateEllipsePoints = useCallback((ellipse: EllipseParams) => {
    const { centerX, centerY, majorAxis, minorAxis, angle } = ellipse;
    const points = [];
    const steps = 50;
    
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
  }, []);

  // Get variance percentages for axis labels
  const xVariance = pcaResult.explained_variance_ratio[xComponent]?.toFixed(1) || '0';
  const yVariance = pcaResult.explained_variance_ratio[yComponent]?.toFixed(1) || '0';

  const xLabel = `PC${xComponent + 1} (${xVariance}%)`;
  const yLabel = `PC${yComponent + 1} (${yVariance}%)`;

  // Calculate data range to ensure 0 is included and axes cross at origin
  const xValues = data.map(d => d!.x);
  const yValues = data.map(d => d!.y);
  const xMin = Math.min(0, ...xValues);
  const xMax = Math.max(0, ...xValues);
  const yMin = Math.min(0, ...yValues);
  const yMax = Math.max(0, ...yValues);
  
  // Add padding to the range
  // Calculate padding that's proportional to the range but with reasonable limits
  const xRange = xMax - xMin;
  const yRange = yMax - yMin;
  
  // If range is very small, use a fixed small padding
  // Otherwise use 10% of the range
  const xPadding = xRange < 0.01 ? 0.1 : xRange * 0.1;
  const yPadding = yRange < 0.01 ? 0.1 : yRange * 0.1;
  
  // Default domain (full range)
  const defaultXDomain: [number, number] = [xMin - xPadding, xMax + xPadding];
  const defaultYDomain: [number, number] = [yMin - yPadding, yMax + yPadding];
  
  // Use zoom/pan hook
  const {
    zoomDomain,
    isPanning,
    handleZoomIn,
    handleZoomOut,
    handleResetView,
    handlePanStart,
    handlePanMove,
    handlePanEnd,
    isZoomed
  } = useZoomPan({
    defaultXDomain,
    defaultYDomain,
    zoomFactor: 0.7
  });
  
  const handleToggleFullscreen = useCallback(() => {
    if (!fullscreenRef.current) return;
    
    if (!isFullscreen) {
      if (fullscreenRef.current.requestFullscreen) {
        fullscreenRef.current.requestFullscreen();
      } else if ('webkitRequestFullscreen' in fullscreenRef.current) {
        (fullscreenRef.current as any).webkitRequestFullscreen();
      }
      setIsFullscreen(true);
    } else {
      if (document.exitFullscreen) {
        document.exitFullscreen();
      } else if ('webkitExitFullscreen' in document) {
        (document as any).webkitExitFullscreen();
      }
      setIsFullscreen(false);
    }
  }, [isFullscreen]);
  
  // Listen for fullscreen changes
  React.useEffect(() => {
    const handleFullscreenChange = () => {
      setIsFullscreen(!!document.fullscreenElement);
    };
    
    document.addEventListener('fullscreenchange', handleFullscreenChange);
    document.addEventListener('webkitfullscreenchange', handleFullscreenChange);
    
    return () => {
      document.removeEventListener('fullscreenchange', handleFullscreenChange);
      document.removeEventListener('webkitfullscreenchange', handleFullscreenChange);
    };
  }, []);
  
  // Update container size on resize
  React.useEffect(() => {
    if (!containerRef.current) return;
    
    const resizeObserver = new ResizeObserver((entries) => {
      for (const entry of entries) {
        setContainerSize({
          width: entry.contentRect.width,
          height: entry.contentRect.height
        });
      }
    });
    
    resizeObserver.observe(containerRef.current);
    
    return () => {
      resizeObserver.disconnect();
    };
  }, []);

  // Handle case where there's no data
  if (data.length === 0) {
    return (
      <div className="w-full h-full flex items-center justify-center text-gray-400">
        <p>No data to display</p>
      </div>
    );
  }

  return (
    <div ref={fullscreenRef} className={`w-full h-full ${isFullscreen ? 'fixed inset-0 z-50 bg-white dark:bg-gray-900 p-4' : ''}`}>
      <div className="flex justify-between items-center mb-2">
        <div className="flex items-center gap-4">
          {/* Group legend */}
          {groupColumn && groupType === 'categorical' && groupColorMap && (
            <div className="flex items-center gap-3 text-sm">
              <span className="text-gray-600 dark:text-gray-400">{groupColumn}:</span>
              {Array.from(groupColorMap.entries()).map(([group, color]) => (
                <div key={group} className="flex items-center gap-1">
                  <div 
                    className="w-3 h-3 rounded-full" 
                    style={{ backgroundColor: color }}
                  />
                  <span className="text-gray-700 dark:text-gray-300">{group}</span>
                </div>
              ))}
            </div>
          )}
          {/* Continuous legend */}
          {groupColumn && groupType === 'continuous' && continuousRange && (
            <div className="flex items-center gap-3 text-sm">
              <span className="text-gray-600 dark:text-gray-400">{groupColumn}:</span>
              <div className="flex items-center gap-2">
                <span className="text-gray-700 dark:text-gray-300">{continuousRange.min.toFixed(2)}</span>
                <div 
                  className="w-32 h-4 rounded"
                  style={{
                    background: `linear-gradient(to right, ${getSequentialColor(0, sequentialPalette)}, ${getSequentialColor(0.5, sequentialPalette)}, ${getSequentialColor(1, sequentialPalette)})`
                  }}
                />
                <span className="text-gray-700 dark:text-gray-300">{continuousRange.max.toFixed(2)}</span>
              </div>
            </div>
          )}
          {isZoomed && (
            <span className="text-sm text-gray-600 dark:text-gray-400">
              Zoomed (drag to pan)
            </span>
          )}
        </div>
        <div className="flex items-center gap-2">
          <PlotControls 
            onResetView={handleResetView}
            onToggleFullscreen={handleToggleFullscreen}
            onZoomIn={handleZoomIn}
            onZoomOut={handleZoomOut}
            isFullscreen={isFullscreen}
          />
          <ExportButton 
            chartRef={containerRef} 
            fileName={`scores-plot-PC${xComponent + 1}-vs-PC${yComponent + 1}`}
          />
        </div>
      </div>
      <div 
        ref={containerRef} 
        className={`w-full relative ${isZoomed ? (isPanning ? 'cursor-grabbing' : 'cursor-grab') : ''}`}
        style={{ height: isFullscreen ? 'calc(100vh - 80px)' : 'calc(100% - 40px)' }}
        onMouseDown={handlePanStart}
        onMouseMove={handlePanMove}
        onMouseUp={handlePanEnd}
        onMouseLeave={handlePanEnd}
      >
        {/* Show ellipse error if any */}
        {showEllipses && ellipsesError && (
          <div className="absolute top-2 left-2 bg-red-100 dark:bg-red-900 text-red-700 dark:text-red-200 px-3 py-1 rounded text-sm z-10">
            {ellipsesError}
          </div>
        )}
        
        {/* Show loading indicator for ellipses */}
        {showEllipses && ellipsesLoading && (
          <div className="absolute top-2 left-2 bg-blue-100 dark:bg-blue-900 text-blue-700 dark:text-blue-200 px-3 py-1 rounded text-sm z-10">
            Calculating ellipses...
          </div>
        )}
        
        {/* SVG Overlay for confidence ellipses */}
        {showEllipses && effectiveGroupEllipses && groupColorMap && !ellipsesError && (
          <svg 
            className="absolute inset-0 pointer-events-none" 
            style={{ width: '100%', height: '100%' }}
          >
            {Object.entries(effectiveGroupEllipses).map(([group, ellipse]) => {
              const color = groupColorMap.get(group) || '#888888';
              const points = generateEllipsePoints(ellipse);
              
              // Calculate scale based on current domain and chart dimensions
              const chartMargins = { top: 20, right: 20, bottom: 60, left: 80 };
              const currentXDomain = zoomDomain.x || defaultXDomain;
              const currentYDomain = zoomDomain.y || defaultYDomain;
              
              // Convert data coordinates to pixel coordinates
              const xScale = (value: number) => {
                const range = currentXDomain[1] - currentXDomain[0];
                const ratio = (value - currentXDomain[0]) / range;
                // Account for margins
                const plotWidth = containerSize.width - chartMargins.left - chartMargins.right;
                return chartMargins.left + ratio * plotWidth;
              };
              
              const yScale = (value: number) => {
                const range = currentYDomain[1] - currentYDomain[0];
                const ratio = (value - currentYDomain[0]) / range;
                // Y is inverted in SVG, account for margins
                const plotHeight = containerSize.height - chartMargins.top - chartMargins.bottom;
                return chartMargins.top + plotHeight - ratio * plotHeight;
              };
              
              // Convert points to SVG path
              const pathData = points
                .map((point, index) => {
                  const x = xScale(point.x);
                  const y = yScale(point.y);
                  return index === 0 ? `M ${x} ${y}` : `L ${x} ${y}`;
                })
                .join(' ') + ' Z';
              
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
        )}
        <ResponsiveContainer width="100%" height="100%">
        <ComposedChart
          data={data}
          margin={{ top: 20, right: 20, bottom: 60, left: 80 }}
        >
          <CartesianGrid strokeDasharray="3 3" stroke={chartTheme.gridColor} />
          <XAxis 
            type="number" 
            dataKey="x" 
            name={xLabel}
            label={{ value: xLabel, position: 'insideBottom', offset: -10 }}
            stroke={chartTheme.axisColor}
            domain={zoomDomain.x || defaultXDomain}
            axisLine={{ stroke: chartTheme.axisColor }}
            tickLine={{ stroke: chartTheme.axisColor }}
            tickFormatter={(value) => value.toFixed(1)}
          />
          <YAxis 
            type="number" 
            dataKey="y" 
            name={yLabel}
            label={{ value: yLabel, angle: -90, position: 'insideLeft' }}
            stroke={chartTheme.axisColor}
            domain={zoomDomain.y || defaultYDomain}
            axisLine={{ stroke: chartTheme.axisColor }}
            tickLine={{ stroke: chartTheme.axisColor }}
            tickFormatter={(value) => value.toFixed(1)}
          />
          <ReferenceLine x={0} stroke={chartTheme.referenceLineColor} strokeWidth={2} />
          <ReferenceLine y={0} stroke={chartTheme.referenceLineColor} strokeWidth={2} />
          
          <Tooltip 
            cursor={{ strokeDasharray: '3 3' }}
            content={({ active, payload }) => {
              if (active && payload && payload.length) {
                const data = payload[0].payload;
                return (
                  <div 
                    className="p-2 rounded shadow-lg border"
                    style={{ 
                      backgroundColor: chartTheme.tooltipBackgroundColor,
                      borderColor: chartTheme.tooltipBorderColor
                    }}
                  >
                    <p className="font-semibold" style={{ color: chartTheme.tooltipTextColor }}>{data.name}</p>
                    {groupColumn && data.group !== 'Unknown' && (
                      <p style={{ color: chartTheme.tooltipTextColor }}>{groupColumn}: {data.group}</p>
                    )}
                    <p style={{ color: chartTheme.tooltipTextColor }}>{xLabel}: {data.x.toFixed(3)}</p>
                    <p style={{ color: chartTheme.tooltipTextColor }}>{yLabel}: {data.y.toFixed(3)}</p>
                  </div>
                );
              }
              return null;
            }}
          />
          <Scatter 
            name="Scores"
            fill="#3B82F6"
            fillOpacity={0.8}
            strokeWidth={1}
            stroke="#1E40AF"
          >
            {groupColumn ? (
              data.map((entry, index) => {
                const fillColor = entry?.color || '#3B82F6';
                return (
                  <Cell key={`cell-${index}`} fill={fillColor} stroke={fillColor} />
                );
              })
            ) : null}
          </Scatter>
        </ComposedChart>
        </ResponsiveContainer>
      </div>
    </div>
  );
};