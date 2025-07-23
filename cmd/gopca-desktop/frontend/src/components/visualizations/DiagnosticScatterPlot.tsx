import React, { useRef, useState } from 'react';
import { 
  ScatterChart, 
  Scatter, 
  XAxis, 
  YAxis, 
  CartesianGrid, 
  Tooltip, 
  ResponsiveContainer,
  ReferenceLine,
  Cell,
  Label
} from 'recharts';
import { PCAResult } from '../../types';
import { ExportButton } from '../ExportButton';
import { PlotControls } from '../PlotControls';
import { useZoomPan } from '../../hooks/useZoomPan';
import { useChartTheme } from '../../hooks/useChartTheme';

interface DiagnosticScatterPlotProps {
  pcaResult: PCAResult;
  rowNames?: string[];
  mahalanobisThreshold?: number;
  rssThreshold?: number;
}

export const DiagnosticScatterPlot: React.FC<DiagnosticScatterPlotProps> = ({ 
  pcaResult,
  rowNames = [],
  mahalanobisThreshold = 3.0,  // Default threshold for Mahalanobis distance
  rssThreshold = 0.03           // Default threshold for RSS (adjusted for typical RSS scale)
}) => {
  const chartRef = useRef<HTMLDivElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  const [isFullscreen, setIsFullscreen] = useState(false);
  const chartTheme = useChartTheme();
  
  // Check if metrics are available
  if (!pcaResult.metrics || pcaResult.metrics.length === 0) {
    return (
      <div className="w-full h-full flex items-center justify-center text-gray-400">
        <p>Diagnostic metrics are not available. Enable metrics calculation in PCA configuration.</p>
      </div>
    );
  }

  // Prepare data for scatter plot
  const data = pcaResult.metrics.map((metric, index) => {
    const mahalanobis = metric.mahalanobis || 0;
    const rss = metric.rss || 0;
    
    // Determine outlier type based on thresholds
    let outlierType = 'normal';
    if (mahalanobis > mahalanobisThreshold && rss > rssThreshold) {
      outlierType = 'strong-outlier';
    } else if (mahalanobis > mahalanobisThreshold) {
      outlierType = 'leverage';
    } else if (rss > rssThreshold) {
      outlierType = 'poor-fit';
    }
    
    return {
      index,
      name: rowNames[index] || `Sample ${index + 1}`,
      mahalanobis,
      rss,
      outlierType
    };
  });

  // Calculate axis domains with padding
  const mahalanobisValues = data.map(d => d.mahalanobis);
  const rssValues = data.map(d => d.rss);
  const maxMahalanobis = Math.max(...mahalanobisValues, mahalanobisThreshold * 1.2);
  const maxRSS = Math.max(...rssValues, rssThreshold * 1.2);

  // Initialize zoom/pan with calculated domains
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
    defaultXDomain: [0, maxRSS],       // X-axis is RSS
    defaultYDomain: [0, maxMahalanobis], // Y-axis is Mahalanobis
    maintainAspectRatio: false
  });

  const xDomain = zoomDomain.x || [0, maxRSS];
  const yDomain = zoomDomain.y || [0, maxMahalanobis];

  // Color mapping for outlier types
  const getColor = (outlierType: string) => {
    switch (outlierType) {
      case 'normal':
        return '#10B981'; // Green
      case 'leverage':
        return '#F59E0B'; // Amber
      case 'poor-fit':
        return '#3B82F6'; // Blue
      case 'strong-outlier':
        return '#EF4444'; // Red
      default:
        return '#6B7280'; // Gray
    }
  };

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
          <p className="font-semibold" style={{ color: chartTheme.tooltipTextColor }}>{data.name}</p>
          <p className="text-sm" style={{ color: chartTheme.tooltipTextColor }}>
            RSS: {data.rss.toFixed(3)}
          </p>
          <p className="text-sm" style={{ color: chartTheme.tooltipTextColor }}>
            Mahalanobis: {data.mahalanobis.toFixed(3)}
          </p>
          <p className="text-sm font-medium" style={{ color: getColor(data.outlierType) }}>
            Type: {data.outlierType.replace('-', ' ')}
          </p>
        </div>
      );
    }
    return null;
  };

  return (
    <div className={`w-full h-full ${isFullscreen ? 'fixed inset-0 z-50 bg-white dark:bg-gray-900 p-4' : ''}`} ref={containerRef}>
      <div className="h-full flex flex-col">
        {/* Header */}
        <div className="flex items-center justify-between mb-4">
          <h4 className="text-md font-medium text-gray-700 dark:text-gray-300">
            Diagnostic Plot: Mahalanobis Distance vs Residual Sum of Squares
          </h4>
          <div className="flex items-center gap-2">
            <ExportButton 
              chartRef={chartRef} 
              fileName="diagnostic-plot-mahalanobis-rss"
            />
            <PlotControls
              onZoomIn={handleZoomIn}
              onZoomOut={handleZoomOut}
              onResetView={handleResetView}
              onToggleFullscreen={() => setIsFullscreen(!isFullscreen)}
              isFullscreen={isFullscreen}
            />
          </div>
        </div>

        {/* Chart */}
        <div className="flex-1" ref={chartRef}>
          <ResponsiveContainer width="100%" height="100%">
            <ScatterChart
              margin={{ top: 20, right: 20, bottom: 60, left: 80 }}
            >
              <CartesianGrid strokeDasharray="3 3" stroke={chartTheme.gridColor} />
              
              <XAxis 
                type="number"
                dataKey="rss"
                domain={xDomain}
                stroke={chartTheme.axisColor}
                tickFormatter={(value) => value.toFixed(2)}
                label={{ 
                  value: 'Residual Sum of Squares (RSS)', 
                  position: 'insideBottom', 
                  offset: -10,
                  style: { fill: chartTheme.textColor }
                }}
              />
              
              <YAxis 
                type="number"
                dataKey="mahalanobis"
                domain={yDomain}
                stroke={chartTheme.axisColor}
                tickFormatter={(value) => value.toFixed(2)}
                label={{ 
                  value: 'Mahalanobis Distance', 
                  angle: -90, 
                  position: 'insideLeft',
                  style: { fill: chartTheme.textColor }
                }}
              />
              
              {/* Reference lines for thresholds */}
              <ReferenceLine 
                x={rssThreshold} 
                stroke="#EF4444" 
                strokeDasharray="5 5"
                strokeWidth={2}
              >
                <Label 
                  value="RSS Threshold" 
                  position="top"
                  style={{ fill: '#EF4444', fontSize: 12 }}
                />
              </ReferenceLine>
              
              <ReferenceLine 
                y={mahalanobisThreshold} 
                stroke="#EF4444" 
                strokeDasharray="5 5"
                strokeWidth={2}
              >
                <Label 
                  value="Mahalanobis Threshold" 
                  position="right"
                  angle={-90}
                  style={{ fill: '#EF4444', fontSize: 12 }}
                />
              </ReferenceLine>
              
              <Tooltip content={<CustomTooltip />} />
              
              <Scatter 
                data={data} 
                fill="#8884d8"
              >
                {data.map((entry, index) => (
                  <Cell key={`cell-${index}`} fill={getColor(entry.outlierType)} />
                ))}
              </Scatter>
            </ScatterChart>
          </ResponsiveContainer>
        </div>

        {/* Legend */}
        <div className="mt-4 flex justify-center gap-6 text-sm">
          <div className="flex items-center gap-2">
            <div className="w-3 h-3 rounded-full" style={{ backgroundColor: '#10B981' }}></div>
            <span className="text-gray-600 dark:text-gray-400">Normal</span>
          </div>
          <div className="flex items-center gap-2">
            <div className="w-3 h-3 rounded-full" style={{ backgroundColor: '#F59E0B' }}></div>
            <span className="text-gray-600 dark:text-gray-400">Leverage Point</span>
          </div>
          <div className="flex items-center gap-2">
            <div className="w-3 h-3 rounded-full" style={{ backgroundColor: '#3B82F6' }}></div>
            <span className="text-gray-600 dark:text-gray-400">Poor Model Fit</span>
          </div>
          <div className="flex items-center gap-2">
            <div className="w-3 h-3 rounded-full" style={{ backgroundColor: '#EF4444' }}></div>
            <span className="text-gray-600 dark:text-gray-400">Strong Outlier</span>
          </div>
        </div>

        {/* Info text */}
        <div className="mt-2 text-xs text-gray-500 dark:text-gray-400 text-center">
          Points in different quadrants indicate different types of outliers. 
          Top-right quadrant contains samples that are both outliers in the model space and have poor reconstruction.
        </div>
      </div>
    </div>
  );
};