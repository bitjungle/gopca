import React, { useRef, useState, useCallback, useMemo } from 'react';
import { ScatterChart, Scatter, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, ReferenceLine, Cell } from 'recharts';
import { PCAResult } from '../../types';
import { ExportButton } from '../ExportButton';
import { PlotControls } from '../PlotControls';
import { useZoomPan } from '../../hooks/useZoomPan';
import { createGroupColorMap } from '../../utils/colors';

interface ScoresPlotProps {
  pcaResult: PCAResult;
  rowNames: string[];
  xComponent?: number; // 0-based index
  yComponent?: number; // 0-based index
  groupColumn?: string | null;
  groupLabels?: string[];
}

export const ScoresPlot: React.FC<ScoresPlotProps> = ({ 
  pcaResult, 
  rowNames,
  xComponent = 0, 
  yComponent = 1,
  groupColumn,
  groupLabels
}) => {
  const chartRef = useRef<HTMLDivElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  const fullscreenRef = useRef<HTMLDivElement>(null);
  
  const [isFullscreen, setIsFullscreen] = useState(false);
  
  // Create color map for groups
  const groupColorMap = useMemo(() => {
    if (groupLabels && groupColumn) {
      return createGroupColorMap(groupLabels);
    }
    return null;
  }, [groupLabels, groupColumn]);
  
  // Transform scores data for Recharts
  const data = pcaResult.scores.map((row, index) => {
    const xVal = row[xComponent] || 0;
    const yVal = row[yComponent] || 0;
    
    // Check for invalid values
    if (!isFinite(xVal) || !isFinite(yVal)) {
      console.warn(`Invalid values at index ${index}: x=${xVal}, y=${yVal}`);
      return null;
    }
    
    const group = groupLabels?.[index];
    
    return {
      x: xVal,
      y: yVal,
      name: rowNames[index] || `Sample ${index + 1}`,
      group: group || 'Unknown',
      color: group && groupColorMap ? groupColorMap.get(group) : '#3B82F6'
    };
  }).filter(point => point !== null);


  // Get variance percentages for axis labels
  const xVariance = pcaResult.explained_variance[xComponent]?.toFixed(1) || '0';
  const yVariance = pcaResult.explained_variance[yComponent]?.toFixed(1) || '0';

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
  const xPadding = (xMax - xMin) * 0.1 || 1;
  const yPadding = (yMax - yMin) * 0.1 || 1;
  
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

  // Handle case where there's no data
  if (data.length === 0) {
    return (
      <div className="w-full h-full flex items-center justify-center text-gray-400">
        <p>No data to display</p>
      </div>
    );
  }

  return (
    <div ref={fullscreenRef} className={`w-full h-full ${isFullscreen ? 'fixed inset-0 z-50 bg-gray-900 p-4' : ''}`}>
      <div className="flex justify-between items-center mb-2">
        <div className="flex items-center gap-4">
          {/* Group legend */}
          {groupColumn && groupColorMap && (
            <div className="flex items-center gap-3 text-sm">
              <span className="text-gray-400">{groupColumn}:</span>
              {Array.from(groupColorMap.entries()).map(([group, color]) => (
                <div key={group} className="flex items-center gap-1">
                  <div 
                    className="w-3 h-3 rounded-full" 
                    style={{ backgroundColor: color }}
                  />
                  <span className="text-gray-300">{group}</span>
                </div>
              ))}
            </div>
          )}
          {isZoomed && (
            <span className="text-sm text-gray-400">
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
        className={`w-full ${isZoomed ? (isPanning ? 'cursor-grabbing' : 'cursor-grab') : ''}`}
        style={{ height: isFullscreen ? 'calc(100vh - 80px)' : 'calc(100% - 40px)' }}
        onMouseDown={handlePanStart}
        onMouseMove={handlePanMove}
        onMouseUp={handlePanEnd}
        onMouseLeave={handlePanEnd}
      >
        <ResponsiveContainer width="100%" height="100%">
        <ScatterChart
          data={data}
          margin={{ top: 20, right: 20, bottom: 60, left: 80 }}
        >
          <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
          <XAxis 
            type="number" 
            dataKey="x" 
            name={xLabel}
            label={{ value: xLabel, position: 'insideBottom', offset: -10 }}
            stroke="#9CA3AF"
            domain={zoomDomain.x || defaultXDomain}
            axisLine={{ stroke: '#9CA3AF' }}
            tickLine={{ stroke: '#9CA3AF' }}
            tickFormatter={(value) => value.toFixed(1)}
          />
          <YAxis 
            type="number" 
            dataKey="y" 
            name={yLabel}
            label={{ value: yLabel, angle: -90, position: 'insideLeft' }}
            stroke="#9CA3AF"
            domain={zoomDomain.y || defaultYDomain}
            axisLine={{ stroke: '#9CA3AF' }}
            tickLine={{ stroke: '#9CA3AF' }}
            tickFormatter={(value) => value.toFixed(1)}
          />
          <ReferenceLine x={0} stroke="#6B7280" strokeWidth={2} />
          <ReferenceLine y={0} stroke="#6B7280" strokeWidth={2} />
          <Tooltip 
            cursor={{ strokeDasharray: '3 3' }}
            content={({ active, payload }) => {
              if (active && payload && payload.length) {
                const data = payload[0].payload;
                return (
                  <div className="bg-gray-800 p-2 rounded shadow-lg border border-gray-600">
                    <p className="text-white font-semibold">{data.name}</p>
                    {groupColumn && data.group !== 'Unknown' && (
                      <p className="text-gray-300">{groupColumn}: {data.group}</p>
                    )}
                    <p className="text-gray-300">{xLabel}: {data.x.toFixed(3)}</p>
                    <p className="text-gray-300">{yLabel}: {data.y.toFixed(3)}</p>
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
            {groupColumn && groupLabels ? (
              data.map((entry, index) => (
                <Cell key={`cell-${index}`} fill={entry!.color} />
              ))
            ) : null}
          </Scatter>
        </ScatterChart>
        </ResponsiveContainer>
      </div>
    </div>
  );
};