// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

import React, { useRef, useState, useCallback, useMemo } from 'react';
import ReactDOM from 'react-dom';
import { PCAResult, FileData } from '../../types';
import { ExportButton } from '../ExportButton';
import { PlotControls } from '../PlotControls';
import { useChartTheme } from '../../hooks/useChartTheme';

interface EigencorrelationPlotProps {
  pcaResult: PCAResult;
  fileData: FileData;
  selectedComponents?: number[]; // Which PCs to show (0-based indices)
  showSignificance?: boolean;
  significanceThreshold?: number;
}

export const EigencorrelationPlot: React.FC<EigencorrelationPlotProps> = ({
  pcaResult,
  fileData,
  selectedComponents = [0, 1, 2, 3, 4], // Default to first 5 components
  showSignificance = true,
  significanceThreshold = 0.05
}) => {
  const chartRef = useRef<HTMLDivElement>(null);
  const fullscreenRef = useRef<HTMLDivElement>(null);
  const [isFullscreen, setIsFullscreen] = useState(false);
  const chartTheme = useChartTheme();
  
  // Tooltip state
  const [tooltip, setTooltip] = useState<{ show: boolean; text: string; x: number; y: number }>({
    show: false, text: '', x: 0, y: 0
  });

  // Get pre-calculated eigencorrelations from PCA result
  const correlationData = useMemo(() => {
    if (!pcaResult?.eigencorrelations) {
      return null;
    }

    const eigencorr = pcaResult.eigencorrelations;
    
    // Filter components based on selection
    const componentsToShow = selectedComponents.filter(i => i < eigencorr.components.length);
    
    // Filter data for selected components
    const filteredCorrelations: { [key: string]: number[] } = {};
    const filteredPValues: { [key: string]: number[] } = {};
    
    eigencorr.variables.forEach(variable => {
      filteredCorrelations[variable] = componentsToShow.map(i => eigencorr.correlations[variable][i]);
      filteredPValues[variable] = componentsToShow.map(i => eigencorr.pValues[variable][i]);
    });

    return {
      correlations: filteredCorrelations,
      pValues: filteredPValues,
      variables: eigencorr.variables,
      components: componentsToShow.map(i => eigencorr.components[i]),
      method: eigencorr.method
    };
  }, [pcaResult?.eigencorrelations, selectedComponents]);

  // Color scale function for correlation values
  const getColor = useCallback((value: number): string => {
    // Diverging color scale: blue (-1) -> white (0) -> red (+1)
    const absValue = Math.abs(value);
    const intensity = Math.floor(absValue * 255);
    
    if (value < 0) {
      // Blue for negative correlations
      return `rgb(${255 - intensity}, ${255 - intensity}, 255)`;
    } else {
      // Red for positive correlations
      return `rgb(255, ${255 - intensity}, ${255 - intensity})`;
    }
  }, []);

  // Format correlation value for display
  const formatCorrelation = (value: number): string => {
    return value.toFixed(2);
  };

  // Check if value is significant
  const isSignificant = (pValue: number): boolean => {
    return pValue < significanceThreshold;
  };

  // Fullscreen handler
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
    // No zoom functionality for heatmap
  }, []);

  // Calculate cell dimensions
  const cellSize = useMemo(() => {
    if (!correlationData) return { width: 50, height: 30 };
    
    const numVars = correlationData.variables.length;
    const numComps = correlationData.components.length;
    
    // Adjust cell size based on data dimensions
    const maxWidth = isFullscreen ? window.innerWidth - 300 : 800;
    const maxHeight = isFullscreen ? window.innerHeight - 200 : 600;
    
    const cellWidth = Math.min(80, Math.max(40, maxWidth / numComps));
    const cellHeight = Math.min(40, Math.max(25, maxHeight / numVars));
    
    return { width: cellWidth, height: cellHeight };
  }, [correlationData, isFullscreen]);

  // No data state
  if (!correlationData || correlationData.variables.length === 0) {
    return (
      <div className="w-full h-full flex items-center justify-center text-gray-500 dark:text-gray-400">
        <p>No eigencorrelation data available. Please ensure metadata variables were included when calculating PCA.</p>
      </div>
    );
  }

  // SVG dimensions
  const margin = { top: 100, right: 100, bottom: 100, left: 200 };
  const width = correlationData.components.length * cellSize.width + margin.left + margin.right;
  const height = correlationData.variables.length * cellSize.height + margin.top + margin.bottom;

  return (
    <div ref={fullscreenRef} className={`w-full h-full ${isFullscreen ? 'fixed inset-0 z-50 bg-white dark:bg-gray-900 p-4' : ''}`}>
      <div className="w-full h-full" ref={chartRef}>
        {/* Header with controls */}
        <div className="flex items-center justify-between mb-4">
          <h4 className="text-md font-medium text-gray-700 dark:text-gray-300">
            Eigencorrelation Plot: Component-Metadata Correlations ({correlationData.method})
          </h4>
          <div className="flex items-center gap-2">
            <PlotControls 
              onResetView={handleResetView}
              onToggleFullscreen={handleToggleFullscreen}
              isFullscreen={isFullscreen}
            />
            <ExportButton 
              chartRef={chartRef} 
              fileName={`eigencorrelations-${correlationData.method}`}
            />
          </div>
        </div>
        
        {/* Heatmap */}
        <div style={{ height: isFullscreen ? 'calc(100vh - 120px)' : 'calc(100% - 60px)', overflowY: 'auto', overflowX: 'auto' }}>
          <svg width={width} height={height}>
            {/* Column headers (Components) */}
            {correlationData.components.map((comp, i) => (
              <g key={`comp-${i}`}>
                <text
                  x={margin.left + i * cellSize.width + cellSize.width / 2}
                  y={margin.top - 10}
                  textAnchor="middle"
                  fontSize="12"
                  fill={chartTheme.textColor}
                  fontWeight="bold"
                >
                  {comp}
                </text>
              </g>
            ))}
            
            {/* Row headers (Variables) */}
            {correlationData.variables.map((variable, i) => (
              <text
                key={`var-${i}`}
                x={margin.left - 10}
                y={margin.top + i * cellSize.height + cellSize.height / 2}
                textAnchor="end"
                fontSize="11"
                fill={chartTheme.textColor}
                dominantBaseline="middle"
              >
                {variable}
              </text>
            ))}
            
            {/* Heatmap cells */}
            {correlationData.variables.map((variable, varIndex) => (
              correlationData.components.map((comp, compIndex) => {
                const correlation = correlationData.correlations[variable]?.[compIndex] || 0;
                const pValue = correlationData.pValues[variable]?.[compIndex] || 1;
                const significant = isSignificant(pValue);
                
                return (
                  <g key={`cell-${varIndex}-${compIndex}`}>
                    {/* Cell background */}
                    <rect
                      x={margin.left + compIndex * cellSize.width}
                      y={margin.top + varIndex * cellSize.height}
                      width={cellSize.width}
                      height={cellSize.height}
                      fill={getColor(correlation)}
                      stroke={chartTheme.gridColor}
                      strokeWidth="1"
                      onMouseEnter={(e) => {
                        const rect = e.currentTarget.getBoundingClientRect();
                        setTooltip({
                          show: true,
                          text: `${variable} × ${comp}\nCorrelation: ${correlation.toFixed(4)}\np-value: ${pValue.toFixed(4)}${significant ? ' (*)' : ''}`,
                          x: rect.left + rect.width / 2,
                          y: rect.top - 10
                        });
                      }}
                      onMouseLeave={() => setTooltip({ show: false, text: '', x: 0, y: 0 })}
                      style={{ cursor: 'pointer' }}
                    />
                    
                    {/* Cell value */}
                    <text
                      x={margin.left + compIndex * cellSize.width + cellSize.width / 2}
                      y={margin.top + varIndex * cellSize.height + cellSize.height / 2}
                      textAnchor="middle"
                      dominantBaseline="middle"
                      fontSize="11"
                      fill={Math.abs(correlation) > 0.5 ? 'white' : chartTheme.textColor}
                      fontWeight={significant && showSignificance ? 'bold' : 'normal'}
                    >
                      {formatCorrelation(correlation)}
                      {significant && showSignificance ? '*' : ''}
                    </text>
                  </g>
                );
              })
            ))}
            
            {/* Color legend */}
            <g transform={`translate(${width - margin.right + 20}, ${margin.top})`}>
              <text
                x="35"
                y="-10"
                textAnchor="middle"
                fontSize="12"
                fill={chartTheme.textColor}
                fontWeight="bold"
              >
                Correlation
              </text>
              
              {/* Gradient */}
              <defs>
                <linearGradient id="correlationGradient" x1="0%" y1="0%" x2="0%" y2="100%">
                  <stop offset="0%" stopColor="rgb(255, 0, 0)" />
                  <stop offset="50%" stopColor="rgb(255, 255, 255)" />
                  <stop offset="100%" stopColor="rgb(0, 0, 255)" />
                </linearGradient>
              </defs>
              
              <rect
                x="20"
                y="0"
                width="30"
                height="200"
                fill="url(#correlationGradient)"
                stroke={chartTheme.gridColor}
              />
              
              {/* Legend labels */}
              <text x="55" y="5" fontSize="11" fill={chartTheme.textColor}>+1.0</text>
              <text x="55" y="105" fontSize="11" fill={chartTheme.textColor}>0.0</text>
              <text x="55" y="205" fontSize="11" fill={chartTheme.textColor}>-1.0</text>
            </g>
            
            {/* Significance note */}
            {showSignificance && (
              <text
                x={margin.left}
                y={height - 20}
                fontSize="11"
                fill={chartTheme.textColor}
              >
                * p &lt; {significanceThreshold}
              </text>
            )}
          </svg>
        </div>
      </div>
      
      {/* Tooltip */}
      {tooltip.show && ReactDOM.createPortal(
        <div
          className="fixed z-50 px-3 py-2 text-xs rounded shadow-lg border pointer-events-none whitespace-pre-line"
          style={{
            backgroundColor: chartTheme.tooltipBackgroundColor,
            borderColor: chartTheme.tooltipBorderColor,
            color: chartTheme.tooltipTextColor,
            left: tooltip.x,
            top: tooltip.y - 30,
            transform: 'translateX(-50%)'
          }}
        >
          {tooltip.text}
        </div>,
        document.body
      )}
    </div>
  );
};