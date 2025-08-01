import React, { useRef, useState, useCallback, useEffect, useMemo } from 'react';
import ReactDOM from 'react-dom';
import { PCAResult, FileData } from '../../types';
import { ExportButton } from '../ExportButton';
import { PlotControls } from '../PlotControls';
import { useChartTheme } from '../../hooks/useChartTheme';
import { CalculateEigencorrelations } from '../../../wailsjs/go/main/App';

interface EigencorrelationPlotProps {
  pcaResult: PCAResult;
  fileData: FileData;
  selectedComponents?: number[]; // Which PCs to show (0-based indices)
  correlationMethod?: 'pearson' | 'spearman';
  showSignificance?: boolean;
  significanceThreshold?: number;
}

interface CorrelationData {
  correlations: { [key: string]: number[] };
  pValues: { [key: string]: number[] };
  variables: string[];
  components: string[];
}

export const EigencorrelationPlot: React.FC<EigencorrelationPlotProps> = ({
  pcaResult,
  fileData,
  selectedComponents = [0, 1, 2, 3, 4], // Default to first 5 components
  correlationMethod = 'pearson',
  showSignificance = true,
  significanceThreshold = 0.05
}) => {
  const chartRef = useRef<HTMLDivElement>(null);
  const fullscreenRef = useRef<HTMLDivElement>(null);
  const [isFullscreen, setIsFullscreen] = useState(false);
  const [correlationData, setCorrelationData] = useState<CorrelationData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const chartTheme = useChartTheme();
  
  // Tooltip state
  const [tooltip, setTooltip] = useState<{ show: boolean; text: string; x: number; y: number }>({
    show: false, text: '', x: 0, y: 0
  });

  // Calculate correlations
  useEffect(() => {
    const calculateCorrelations = async () => {
      if (!pcaResult.scores || !fileData) {
        setLoading(false);
        return;
      }

      setLoading(true);
      setError(null);

      try {
        // Prepare metadata
        const metadataNumeric: { [key: string]: number[] } = {};
        const metadataCategorical: { [key: string]: string[] } = {};

        // Add numeric target columns
        if (fileData.numericTargetColumns) {
          Object.entries(fileData.numericTargetColumns).forEach(([name, values]) => {
            metadataNumeric[name] = values;
          });
        }

        // Add categorical columns
        if (fileData.categoricalColumns) {
          Object.entries(fileData.categoricalColumns).forEach(([name, values]) => {
            metadataCategorical[name] = values;
          });
        }

        // Ensure we have some metadata to correlate
        if (Object.keys(metadataNumeric).length === 0 && Object.keys(metadataCategorical).length === 0) {
          setError('No metadata variables available for correlation analysis');
          setLoading(false);
          return;
        }

        // Filter selected components
        const componentsToUse = selectedComponents.filter(i => i < pcaResult.scores[0].length);
        
        const response = await CalculateEigencorrelations({
          scores: pcaResult.scores,
          metadataNumeric,
          metadataCategorical,
          components: componentsToUse,
          method: correlationMethod
        });

        if (response.success) {
          setCorrelationData({
            correlations: response.correlations || {},
            pValues: response.pValues || {},
            variables: response.variables || [],
            components: response.components || []
          });
        } else {
          setError(response.error || 'Failed to calculate correlations');
        }
      } catch (err) {
        setError(`Error calculating correlations: ${err}`);
      } finally {
        setLoading(false);
      }
    };

    calculateCorrelations();
  }, [pcaResult, fileData, selectedComponents, correlationMethod]);

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

  // Loading state
  if (loading) {
    return (
      <div className="w-full h-full flex items-center justify-center text-gray-500 dark:text-gray-400">
        <p>Calculating correlations...</p>
      </div>
    );
  }

  // Error state
  if (error) {
    return (
      <div className="w-full h-full flex items-center justify-center text-red-500">
        <p>{error}</p>
      </div>
    );
  }

  // No data state
  if (!correlationData || correlationData.variables.length === 0) {
    return (
      <div className="w-full h-full flex items-center justify-center text-gray-500 dark:text-gray-400">
        <p>No correlation data to display</p>
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
            Eigencorrelation Plot: Component-Metadata Correlations ({correlationMethod})
          </h4>
          <div className="flex items-center gap-2">
            <PlotControls 
              onResetView={handleResetView}
              onToggleFullscreen={handleToggleFullscreen}
              isFullscreen={isFullscreen}
            />
            <ExportButton 
              chartRef={chartRef} 
              fileName={`eigencorrelations-${correlationMethod}`}
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