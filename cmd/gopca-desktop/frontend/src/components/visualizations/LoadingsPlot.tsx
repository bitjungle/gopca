import React, { useState, useEffect, useRef, useCallback } from 'react';
import { 
  BarChart, 
  Bar, 
  LineChart,
  Line,
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
import { useChartTheme } from '../../hooks/useChartTheme';

interface LoadingsPlotProps {
  pcaResult: PCAResult;
  selectedComponent?: number; // 0-based index
  variableThreshold?: number; // Threshold for auto-switching between bar and line
}

export const LoadingsPlot: React.FC<LoadingsPlotProps> = ({ 
  pcaResult, 
  selectedComponent = 0,
  variableThreshold = 50
}) => {
  const chartRef = useRef<HTMLDivElement>(null);
  const fullscreenRef = useRef<HTMLDivElement>(null);
  const [isFullscreen, setIsFullscreen] = useState(false);
  const chartTheme = useChartTheme();
  
  // Check if loadings are available (not available for Kernel PCA)
  if (!pcaResult.loadings || pcaResult.loadings.length === 0) {
    return (
      <div className="w-full h-full flex items-center justify-center text-gray-400">
        <p>Loadings are not available for this PCA method</p>
      </div>
    );
  }
  
  // Auto-detect plot type based on number of variables
  const numVariables = pcaResult.loadings[0]?.length || 0;
  const autoPlotType = numVariables > variableThreshold ? 'line' : 'bar';
  
  const [plotType, setPlotType] = useState<'bar' | 'line'>(autoPlotType);

  // Update plot type when data changes
  useEffect(() => {
    setPlotType(numVariables > variableThreshold ? 'line' : 'bar');
  }, [numVariables, variableThreshold]);

  // Extract loadings for selected component
  const loadingsData = pcaResult.loadings.map((row, index) => {
    const loading = row[selectedComponent] || 0;
    return {
      variable: pcaResult.variable_labels?.[index] || `Var${index + 1}`,
      loading: loading,
      absLoading: Math.abs(loading),
      variableIndex: index
    };
  });

  // Sort by absolute loading value for bar chart
  const sortedData = plotType === 'bar' 
    ? [...loadingsData].sort((a, b) => b.absLoading - a.absLoading)
    : loadingsData;

  // Calculate min/max for axis domain
  const loadingValues = sortedData.map(d => d.loading);
  const minLoading = Math.min(...loadingValues);
  const maxLoading = Math.max(...loadingValues);
  const absMax = Math.max(Math.abs(minLoading), Math.abs(maxLoading));
  const padding = absMax * 0.1;

  // Get component label and variance
  const componentLabel = pcaResult.component_labels?.[selectedComponent] || `PC${selectedComponent + 1}`;
  const variance = pcaResult.explained_variance[selectedComponent]?.toFixed(1) || '0';

  // Handle edge cases
  if (!loadingsData || loadingsData.length === 0) {
    return (
      <div className="w-full h-full flex items-center justify-center text-gray-500 dark:text-gray-400">
        <p>No loadings data to display</p>
      </div>
    );
  }

  // Custom tooltip
  const CustomTooltip = ({ active, payload }: any) => {
    if (active && payload && payload.length) {
      const data = payload[0].payload;
      return (
        <div 
          className="p-3 rounded shadow-lg border"
          style={{ 
            backgroundColor: chartTheme.tooltipBackgroundColor,
            borderColor: chartTheme.tooltipBorderColor
          }}
        >
          <p className="font-semibold" style={{ color: chartTheme.tooltipTextColor }}>{data.variable}</p>
          <p className="text-blue-400">
            Loading: {data.loading.toFixed(4)}
          </p>
        </div>
      );
    }
    return null;
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
        {/* Header with plot type selector and export button */}
        <div className="flex items-center justify-between mb-4">
          <h4 className="text-md font-medium text-gray-700 dark:text-gray-300">
            {componentLabel} Loadings ({variance}% variance)
          </h4>
        <div className="flex items-center gap-4">
          <div className="flex items-center gap-2">
            <span className="text-sm text-gray-600 dark:text-gray-400">Plot type:</span>
            <select
              value={plotType}
              onChange={(e) => setPlotType(e.target.value as 'bar' | 'line')}
              className="px-2 py-1 bg-gray-100 dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded text-sm text-gray-900 dark:text-white"
            >
              <option value="bar">Bar Chart</option>
              <option value="line">Line Chart</option>
            </select>
            {plotType !== autoPlotType && (
              <span className="text-xs text-yellow-500">(manual)</span>
            )}
          </div>
          <div className="flex items-center gap-2">
            <PlotControls 
              onResetView={handleResetView}
              onToggleFullscreen={handleToggleFullscreen}
              isFullscreen={isFullscreen}
            />
            <ExportButton 
              chartRef={chartRef} 
              fileName={`loadings-plot-${componentLabel}`}
            />
          </div>
        </div>
      </div>

      <ResponsiveContainer width="100%" height="90%">
        {plotType === 'bar' ? (
          <BarChart 
            data={sortedData} 
            margin={{ top: 20, right: 20, bottom: 60, left: 80 }}
          >
            <CartesianGrid strokeDasharray="3 3" stroke={chartTheme.gridColor} />
            
            <XAxis 
              dataKey="variable" 
              stroke={chartTheme.axisColor}
              angle={-45}
              textAnchor="end"
              height={60}
              interval={numVariables <= 20 ? 0 : 'preserveStartEnd'}
            />
            
            <YAxis 
              stroke={chartTheme.axisColor}
              domain={[-absMax - padding, absMax + padding]}
              tickFormatter={(value) => value.toFixed(2)}
            />
            
            <ReferenceLine y={0} stroke={chartTheme.referenceLineColor} strokeWidth={2} />
            
            <Tooltip content={<CustomTooltip />} />
            
            <Bar 
              dataKey="loading" 
              radius={[4, 4, 0, 0]}
            >
              {sortedData.map((entry, index) => (
                <Cell key={`cell-${index}`} fill={entry.loading >= 0 ? '#3B82F6' : '#EF4444'} />
              ))}
            </Bar>
          </BarChart>
        ) : (
          <LineChart 
            data={sortedData} 
            margin={{ top: 20, right: 20, bottom: 60, left: 80 }}
          >
            <CartesianGrid strokeDasharray="3 3" stroke={chartTheme.gridColor} />
            
            <XAxis 
              dataKey="variableIndex" 
              stroke={chartTheme.axisColor}
              label={{ 
                value: 'Variable Index', 
                position: 'insideBottom', 
                offset: -10,
                style: { fill: chartTheme.textColor }
              }}
              domain={[0, numVariables - 1]}
            />
            
            <YAxis 
              stroke={chartTheme.axisColor}
              domain={[-absMax - padding, absMax + padding]}
              tickFormatter={(value) => value.toFixed(2)}
              label={{ 
                value: 'Loading Value', 
                angle: -90, 
                position: 'insideLeft',
                style: { fill: chartTheme.textColor }
              }}
            />
            
            <ReferenceLine y={0} stroke={chartTheme.referenceLineColor} strokeWidth={2} />
            
            <Tooltip content={<CustomTooltip />} />
            
            <Line 
              type="monotone" 
              dataKey="loading" 
              stroke="#3B82F6" 
              strokeWidth={2}
              dot={numVariables <= 50}
              activeDot={{ r: 6 }}
            />
          </LineChart>
        )}
      </ResponsiveContainer>
    </div>
    </div>
  );
};