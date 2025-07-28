import React, { useState, useRef, useCallback, useMemo } from 'react';
import { 
  ScatterChart, 
  Scatter, 
  XAxis, 
  YAxis, 
  CartesianGrid, 
  Tooltip, 
  ResponsiveContainer, 
  ReferenceLine,
  Cell
} from 'recharts';
import { PCAResult } from '../../types';
import { ExportButton } from '../ExportButton';
import { PlotControls } from '../PlotControls';
import { useZoomPan } from '../../hooks/useZoomPan';
import { useChartTheme } from '../../hooks/useChartTheme';
import { usePalette } from '../../contexts/PaletteContext';
import { getQualitativeColor, getSequentialColor } from '../../utils/colorPalettes';

interface BiplotProps {
  pcaResult: PCAResult;
  rowNames: string[];
  xComponent?: number; // 0-based index
  yComponent?: number; // 0-based index
  showLoadingLabels?: boolean;
  groupColumn?: string | null;
  groupLabels?: string[];
  groupValues?: number[]; // For continuous columns
  groupType?: 'categorical' | 'continuous';
}

export const Biplot: React.FC<BiplotProps> = ({ 
  pcaResult, 
  rowNames,
  xComponent = 0, 
  yComponent = 1,
  showLoadingLabels = true,
  groupColumn,
  groupLabels,
  groupValues,
  groupType = 'categorical'
}) => {
  const [hoveredVariable, setHoveredVariable] = useState<number | null>(null);
  const chartRef = useRef<HTMLDivElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  const fullscreenRef = useRef<HTMLDivElement>(null);
  
  const [isFullscreen, setIsFullscreen] = useState(false);
  const chartTheme = useChartTheme();
  const { paletteType } = usePalette();
  
  // Check if loadings are available (not available for Kernel PCA)
  if (!pcaResult.loadings || pcaResult.loadings.length === 0) {
    return (
      <div className="w-full h-full flex items-center justify-center text-gray-400">
        <p>Biplot is not available for this PCA method (loadings required)</p>
      </div>
    );
  }

  // Create color map for groups based on selected palette type
  const groupColorMap = useMemo(() => {
    if (groupType === 'categorical' && groupLabels && groupColumn) {
      const uniqueGroups = [...new Set(groupLabels)].sort();
      const colorMap = new Map<string, string>();
      
      if (paletteType === 'sequential') {
        // For sequential palette, distribute colors across the gradient
        uniqueGroups.forEach((group, index) => {
          const normalizedValue = uniqueGroups.length > 1 
            ? index / (uniqueGroups.length - 1) 
            : 0.5;
          colorMap.set(group, getSequentialColor(normalizedValue));
        });
      } else {
        // For qualitative palette, use distinct colors
        uniqueGroups.forEach((group, index) => {
          colorMap.set(group, getQualitativeColor(index));
        });
      }
      
      return colorMap;
    }
    return null;
  }, [groupLabels, groupColumn, paletteType, groupType]);
  
  // Calculate min/max for continuous values
  const continuousRange = useMemo(() => {
    if (groupType === 'continuous' && groupValues) {
      const validValues = groupValues.filter(v => !isNaN(v) && isFinite(v));
      if (validValues.length > 0) {
        return {
          min: Math.min(...validValues),
          max: Math.max(...validValues)
        };
      }
    }
    return null;
  }, [groupValues, groupType]);

  // Transform scores data
  const scoresData = pcaResult.scores.map((row, index) => {
    let color = '#3B82F6'; // Default color
    let group = 'Unknown';
    let value: number | undefined;
    
    if (groupType === 'categorical') {
      group = groupLabels?.[index] || 'Unknown';
      if (group && groupColorMap) {
        color = groupColorMap.get(group) || color;
      }
    } else if (groupType === 'continuous' && groupValues && continuousRange) {
      const val = groupValues[index];
      value = val;
      if (!isNaN(val) && isFinite(val)) {
        const normalized = (val - continuousRange.min) / (continuousRange.max - continuousRange.min);
        color = getSequentialColor(normalized);
        group = val.toFixed(2); // For display purposes
      }
    }
    
    return {
      x: row[xComponent] || 0,
      y: row[yComponent] || 0,
      name: rowNames[index] || `Sample ${index + 1}`,
      type: 'score',
      group: group,
      color: color,
      value: value
    };
  });

  // Calculate scores range for plot bounds
  const scoreXValues = scoresData.map(d => d.x);
  const scoreYValues = scoresData.map(d => d.y);
  const scoreXMin = Math.min(...scoreXValues);
  const scoreXMax = Math.max(...scoreXValues);
  const scoreYMin = Math.min(...scoreYValues);
  const scoreYMax = Math.max(...scoreYValues);
  
  // Calculate plot bounds (including some padding)
  // For biplot, we need symmetric axes centered at origin
  const scoreXRange = scoreXMax - scoreXMin;
  const scoreYRange = scoreYMax - scoreYMin;
  const maxAbsScore = Math.max(Math.abs(scoreXMin), Math.abs(scoreXMax), Math.abs(scoreYMin), Math.abs(scoreYMax));
  
  // Use the maximum absolute value to ensure we capture all points
  // Add 20% padding, but ensure minimum visibility
  const plotMax = Math.max(maxAbsScore * 1.2, 1.0);

  // Calculate loading vectors and find the maximum
  const loadingVectors = pcaResult.loadings.map(row => {
    const x = row[xComponent] || 0;
    const y = row[yComponent] || 0;
    return {
      x,
      y,
      magnitude: Math.sqrt(x * x + y * y)
    };
  });
  
  const maxLoadingMagnitude = Math.max(...loadingVectors.map(v => v.magnitude));
  
  // Scale factor to make the largest loading vector reach 70% of plot bounds
  const scaleFactor = maxLoadingMagnitude > 0 ? (plotMax * 0.7) / maxLoadingMagnitude : 1;

  // Transform loadings data for display
  const loadingsData = pcaResult.loadings.map((row, index) => {
    const originalX = row[xComponent] || 0;
    const originalY = row[yComponent] || 0;
    
    // Scale loadings to be visible
    const scaledX = originalX * scaleFactor;
    const scaledY = originalY * scaleFactor;
    const magnitude = Math.sqrt(scaledX * scaledX + scaledY * scaledY);
    
    return {
      index,
      x: scaledX,
      y: scaledY,
      magnitude,
      label: pcaResult.variable_labels?.[index] || `Var${index + 1}`,
      originalX,
      originalY
    };
  });

  // Sort loadings by magnitude and get top N for labeling
  const topLoadings = [...loadingsData]
    .sort((a, b) => b.magnitude - a.magnitude)
    .slice(0, 8); // Show labels for top 8 variables

  // Set symmetric axis ranges based on plot bounds
  const axisRange = plotMax;
  const defaultDomain: [number, number] = [-axisRange, axisRange];
  
  // Use zoom/pan hook with maintain aspect ratio for biplot
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
    defaultXDomain: defaultDomain,
    defaultYDomain: defaultDomain,
    zoomFactor: 0.7,
    maintainAspectRatio: true // Important for biplot to keep symmetry
  });
  
  const handleToggleFullscreen = useCallback(() => {
    if (!fullscreenRef.current) return;
    
    if (!isFullscreen) {
      if (fullscreenRef.current.requestFullscreen) {
        fullscreenRef.current.requestFullscreen();
      } else if ((fullscreenRef.current as any).webkitRequestFullscreen) {
        (fullscreenRef.current as any).webkitRequestFullscreen();
      }
      setIsFullscreen(true);
    } else {
      if (document.exitFullscreen) {
        document.exitFullscreen();
      } else if ((document as any).webkitExitFullscreen) {
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

  // Get variance percentages for axis labels
  const xVariance = pcaResult.explained_variance[xComponent]?.toFixed(1) || '0';
  const yVariance = pcaResult.explained_variance[yComponent]?.toFixed(1) || '0';
  const xLabel = `PC${xComponent + 1} (${xVariance}%)`;
  const yLabel = `PC${yComponent + 1} (${yVariance}%)`;

  // Custom dot for loading endpoints
  const LoadingEndpoint = (props: any) => {
    const { cx, cy, payload } = props;
    if (!payload || !payload.isLoadingEnd) return null;
    
    const { index, label } = payload;
    const isHovered = hoveredVariable === index;
    const isTopLoading = topLoadings.some(tl => tl.index === index);
    const shouldShowLabel = showLoadingLabels && (isTopLoading || isHovered);
    
    return (
      <g>
        <circle
          cx={cx}
          cy={cy}
          r={4}
          fill={isHovered ? "#EF4444" : "#10B981"}
          onMouseEnter={() => setHoveredVariable(index)}
          onMouseLeave={() => setHoveredVariable(null)}
        />
        {shouldShowLabel && (
          <text
            x={cx + (payload.x > 0 ? 10 : -10)}
            y={cy}
            fill={chartTheme.textColor}
            fontSize="12"
            fontWeight="500"
            textAnchor={payload.x > 0 ? "start" : "end"}
            dominantBaseline="middle"
          >
            {label}
          </text>
        )}
      </g>
    );
  };

  // Create data for loading endpoints
  const loadingEndpoints = loadingsData.map(loading => ({
    ...loading,
    isLoadingEnd: true
  }));

  // Custom shape to draw loading arrows from origin
  const LoadingArrows = () => {
    return (
      <g>
        {loadingsData.map(loading => {
          const isHovered = hoveredVariable === loading.index;
          const angle = Math.atan2(loading.y, loading.x);
          const arrowLength = 0.3;
          const arrowAngle = 0.4;
          
          // Calculate arrow head points
          const headX1 = loading.x - arrowLength * Math.cos(angle - arrowAngle);
          const headY1 = loading.y - arrowLength * Math.sin(angle - arrowAngle);
          const headX2 = loading.x - arrowLength * Math.cos(angle + arrowAngle);
          const headY2 = loading.y - arrowLength * Math.sin(angle + arrowAngle);
          
          return (
            <g key={`arrow-${loading.index}`}>
              <line
                x1={0}
                y1={0}
                x2={loading.x}
                y2={loading.y}
                stroke={isHovered ? "#EF4444" : "#10B981"}
                strokeWidth={isHovered ? 3 : 2}
                onMouseEnter={() => setHoveredVariable(loading.index)}
                onMouseLeave={() => setHoveredVariable(null)}
              />
              <path
                d={`M ${loading.x} ${loading.y} L ${headX1} ${headY1} L ${headX2} ${headY2} Z`}
                fill={isHovered ? "#EF4444" : "#10B981"}
              />
            </g>
          );
        })}
      </g>
    );
  };

  // Custom tooltip
  const CustomTooltip = ({ active, payload }: any) => {
    if (active && payload && payload.length) {
      const data = payload[0].payload;
      
      if (data.type === 'score') {
        return (
          <div 
            className="p-3 rounded shadow-lg border"
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
      } else if (data.isLoadingEnd) {
        return (
          <div 
            className="p-3 rounded shadow-lg border"
            style={{ 
              backgroundColor: chartTheme.tooltipBackgroundColor,
              borderColor: chartTheme.tooltipBorderColor
            }}
          >
            <p className="font-semibold" style={{ color: chartTheme.tooltipTextColor }}>{data.label}</p>
            <p style={{ color: chartTheme.tooltipTextColor }}>Loading values:</p>
            <p style={{ color: chartTheme.tooltipTextColor }}>{xLabel}: {data.originalX.toFixed(4)}</p>
            <p style={{ color: chartTheme.tooltipTextColor }}>{yLabel}: {data.originalY.toFixed(4)}</p>
          </div>
        );
      }
    }
    return null;
  };

  return (
    <div ref={fullscreenRef} className={`w-full h-full ${isFullscreen ? 'fixed inset-0 z-50 bg-white dark:bg-gray-900 p-4' : ''}`}>
      <div className="flex items-center justify-between mb-2">
        <div className="flex items-center gap-4">
          <h4 className="text-md font-medium text-gray-700 dark:text-gray-300">
            Biplot: {xLabel} vs {yLabel}
          </h4>
          {/* Group legend */}
          {groupColumn && groupType === 'categorical' && groupColorMap ? (
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
          ) : groupColumn && groupType === 'continuous' && continuousRange ? (
            <div className="flex items-center gap-3 text-sm">
              <span className="text-gray-600 dark:text-gray-400">{groupColumn}:</span>
              <div className="flex items-center gap-2">
                <span className="text-gray-700 dark:text-gray-300">{continuousRange.min.toFixed(2)}</span>
                <div 
                  className="w-32 h-4 rounded"
                  style={{
                    background: `linear-gradient(to right, ${getSequentialColor(0)}, ${getSequentialColor(0.5)}, ${getSequentialColor(1)})`
                  }}
                />
                <span className="text-gray-700 dark:text-gray-300">{continuousRange.max.toFixed(2)}</span>
              </div>
            </div>
          ) : (
            <div className="flex items-center gap-4 text-sm text-gray-600 dark:text-gray-400">
              <span className="flex items-center gap-2">
                <span className="w-3 h-3 bg-blue-500 rounded-full"></span>
                Scores
              </span>
              <span className="flex items-center gap-2">
                <svg width="20" height="12">
                  <line x1="0" y1="6" x2="15" y2="6" stroke="#10B981" strokeWidth="2" />
                  <path d="M 15 6 L 12 3 L 12 9 Z" fill="#10B981" />
                </svg>
                Loadings
              </span>
            </div>
          )}
          {isZoomed && <span className="text-sm text-gray-600 dark:text-gray-400">Zoomed (drag to pan)</span>}
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
            fileName={`biplot-PC${xComponent + 1}-vs-PC${yComponent + 1}`}
          />
        </div>
      </div>
      
      <div 
        ref={containerRef} 
        className={`w-full ${isZoomed ? (isPanning ? 'cursor-grabbing' : 'cursor-grab') : ''}`}
        style={{ height: isFullscreen ? 'calc(100vh - 80px)' : 'calc(100% - 40px)' }}
        onMouseDown={handlePanStart}
        onMouseMove={handlePanMove}
        onMouseUp={handlePanEnd}
        onMouseLeave={handlePanEnd}
      >
        <ResponsiveContainer width="100%" height="100%">
        <ScatterChart
          margin={{ top: 20, right: 20, bottom: 60, left: 80 }}
        >
          <defs>
            <marker
              id="arrowhead"
              markerWidth="10"
              markerHeight="10"
              refX="9"
              refY="3"
              orient="auto"
              markerUnits="strokeWidth"
            >
              <path d="M0,0 L0,6 L9,3 z" fill="#10B981" />
            </marker>
          </defs>
          
          <CartesianGrid strokeDasharray="3 3" stroke={chartTheme.gridColor} />
          
          <XAxis 
            type="number" 
            dataKey="x" 
            name={xLabel}
            label={{ value: xLabel, position: 'insideBottom', offset: -10, style: { fill: chartTheme.textColor } }}
            stroke={chartTheme.axisColor}
            domain={zoomDomain.x || defaultDomain}
            tickFormatter={(value) => value.toFixed(1)}
            allowDataOverflow={false}
          />
          
          <YAxis 
            type="number" 
            dataKey="y" 
            name={yLabel}
            label={{ value: yLabel, angle: -90, position: 'insideLeft', style: { fill: chartTheme.textColor } }}
            stroke={chartTheme.axisColor}
            domain={zoomDomain.y || defaultDomain}
            tickFormatter={(value) => value.toFixed(1)}
            allowDataOverflow={false}
          />
          
          <ReferenceLine x={0} stroke={chartTheme.referenceLineColor} strokeWidth={2} />
          <ReferenceLine y={0} stroke={chartTheme.referenceLineColor} strokeWidth={2} />
          
          <Tooltip content={<CustomTooltip />} />
          
          {/* Loading vectors as reference lines */}
          {loadingsData.map(loading => {
            const isHovered = hoveredVariable === loading.index;
            return (
              <ReferenceLine
                key={`loading-line-${loading.index}`}
                segment={[{ x: 0, y: 0 }, { x: loading.x, y: loading.y }]}
                stroke={isHovered ? "#EF4444" : "#10B981"}
                strokeWidth={isHovered ? 3 : 2}
                ifOverflow="visible"
              />
            );
          })}
          
          {/* Scores (samples) */}
          <Scatter 
            name="Scores" 
            data={scoresData}
            fill="#3B82F6"
            fillOpacity={0.8}
            strokeWidth={1}
            stroke="#1E40AF"
          >
            {groupColumn ? (
              scoresData.map((entry, index) => (
                <Cell key={`cell-${index}`} fill={entry.color} stroke={entry.color} />
              ))
            ) : null}
          </Scatter>
          
          {/* Loading endpoints for interaction */}
          <Scatter
            name="LoadingEnds"
            data={loadingEndpoints}
            fill="transparent"
            shape={<LoadingEndpoint />}
          />
        </ScatterChart>
        </ResponsiveContainer>
      </div>
    </div>
  );
};