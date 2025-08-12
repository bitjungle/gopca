// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

import React, { useRef, useState, useCallback } from 'react';
import { 
  ComposedChart, 
  Bar, 
  Line, 
  XAxis, 
  YAxis, 
  CartesianGrid, 
  Tooltip, 
  ResponsiveContainer,
  Legend,
  ReferenceLine,
  Cell
} from 'recharts';
import { PCAResult } from '../../types';
import { ExportButton } from '../ExportButton';
import { PlotControls } from '../PlotControls';
import { useChartTheme } from '../../hooks/useChartTheme';
import { usePalette } from '../../contexts/PaletteContext';
import { getQualitativeColor, getSequentialColorScale } from '../../utils/colorPalettes';

interface ScreePlotProps {
  pcaResult: PCAResult;
  showCumulative?: boolean;
  elbowThreshold?: number; // Optional: highlight components explaining this % variance
}

export const ScreePlot: React.FC<ScreePlotProps> = ({ 
  pcaResult, 
  showCumulative = true,
  elbowThreshold = 80 
}) => {
  const chartRef = useRef<HTMLDivElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  const fullscreenRef = useRef<HTMLDivElement>(null);
  const [isFullscreen, setIsFullscreen] = useState(false);
  const chartTheme = useChartTheme();
  const { qualitativePalette, sequentialPalette } = usePalette();
  
  // Transform variance data for Recharts
  const data = pcaResult.explained_variance.map((variance, index) => ({
    component: pcaResult.component_labels?.[index] || `PC${index + 1}`,
    variance: variance,
    cumulative: pcaResult.cumulative_variance[index],
    componentNumber: index + 1
  }));

  // Find elbow point (where cumulative variance crosses threshold)
  const elbowIndex = pcaResult.cumulative_variance.findIndex(cv => cv >= elbowThreshold);
  const elbowComponent = elbowIndex >= 0 ? elbowIndex + 1 : data.length;

  // Handle edge cases
  if (!data || data.length === 0) {
    return (
      <div className="w-full h-full flex items-center justify-center text-gray-500 dark:text-gray-400">
        <p>No variance data to display</p>
      </div>
    );
  }

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
      <div className="w-full h-full">
        <div className="flex justify-end mb-2 gap-2">
          <PlotControls 
            onResetView={handleResetView}
            onToggleFullscreen={handleToggleFullscreen}
            isFullscreen={isFullscreen}
          />
          <ExportButton 
            chartRef={containerRef} 
            fileName="scree-plot"
          />
        </div>
      <div ref={containerRef} className="w-full" style={{ height: isFullscreen ? 'calc(100vh - 80px)' : 'calc(100% - 40px)' }}>
        <ResponsiveContainer width="100%" height="100%">
        <ComposedChart 
          data={data} 
          margin={{ top: 20, right: 20, bottom: 60, left: 80 }}
        >
          <CartesianGrid strokeDasharray="3 3" stroke={chartTheme.gridColor} />
          
          <XAxis 
            dataKey="component" 
            stroke={chartTheme.axisColor}
            label={{ 
              value: 'Principal Component', 
              position: 'insideBottom', 
              offset: -10,
              style: { fill: chartTheme.textColor }
            }}
          />
          
          <YAxis 
            yAxisId="variance"
            stroke={chartTheme.axisColor}
            label={{ 
              value: 'Explained Variance (%)', 
              angle: -90, 
              position: 'insideLeft',
              style: { fill: chartTheme.textColor }
            }}
            domain={[0, 'auto']}
            tickFormatter={(value) => value.toFixed(0)}
          />
          
          {showCumulative && (
            <YAxis 
              yAxisId="cumulative"
              orientation="right"
              stroke="#10B981"
              label={{ 
                value: 'Cumulative Variance (%)', 
                angle: 90, 
                position: 'insideRight',
                style: { fill: '#10B981' }
              }}
              domain={[0, 100]}
              tickFormatter={(value) => value.toFixed(0)}
            />
          )}
          
          <Tooltip 
            content={({ active, payload }) => {
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
                    <p className="font-semibold" style={{ color: chartTheme.tooltipTextColor }}>{data.component}</p>
                    <p className="text-blue-400">
                      Variance: {data.variance.toFixed(2)}%
                    </p>
                    {showCumulative && (
                      <p className="text-green-400">
                        Cumulative: {data.cumulative.toFixed(2)}%
                      </p>
                    )}
                  </div>
                );
              }
              return null;
            }}
          />
          
          <Legend 
            wrapperStyle={{ paddingTop: '20px' }}
            iconType="rect"
          />
          
          {/* Reference line at elbow threshold */}
          {showCumulative && (
            <ReferenceLine 
              y={elbowThreshold} 
              yAxisId="cumulative"
              stroke="#EF4444" 
              strokeDasharray="5 5"
              label={{ 
                value: `${elbowThreshold}%`, 
                position: 'right',
                style: { fill: '#EF4444' }
              }}
            />
          )}
          
          {/* Vertical line at elbow component */}
          {elbowComponent <= data.length && (
            <ReferenceLine 
              x={`PC${elbowComponent}`} 
              stroke="#EF4444" 
              strokeDasharray="3 3"
              strokeOpacity={0.5}
            />
          )}
          
          <Bar 
            yAxisId="variance"
            dataKey="variance" 
            fill="#3B82F6"
            name="Explained Variance"
            radius={[4, 4, 0, 0]}
          >
            {data.map((entry, index) => {
              // Use qualitative palette with color cycling for each component
              const color = getQualitativeColor(index, qualitativePalette);
              return <Cell key={`cell-${index}`} fill={color} />;
            })}
          </Bar>
          
          {showCumulative && (
            <Line 
              yAxisId="cumulative"
              type="monotone" 
              dataKey="cumulative" 
              stroke="#10B981"
              strokeWidth={2}
              name="Cumulative Variance"
              dot={{ fill: '#10B981', r: 4 }}
            />
          )}
        </ComposedChart>
        </ResponsiveContainer>
      </div>
    </div>
    </div>
  );
};