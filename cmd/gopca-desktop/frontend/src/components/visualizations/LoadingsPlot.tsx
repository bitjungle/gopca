// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Plotly-based PCA Loadings Plot

import React, { useMemo, useState } from 'react';
import { PCALoadingsPlot, useTheme } from '@gopca/ui-components';
import { PCAResult } from '../../types';
import {
  transformToLoadingsPlotData,
  createLoadingsPlotConfig
} from '../../utils/plotlyDataTransform';
import { usePalette } from '../../contexts/PaletteContext';
import { getQualitativePalette } from '../../utils/colorPalettes';

interface LoadingsPlotProps {
  pcaResult: PCAResult;
  selectedComponent?: number; // 0-based index
  variableThreshold?: number; // Threshold for auto-switching between bar and line
}

export const LoadingsPlot: React.FC<LoadingsPlotProps> = ({ 
  pcaResult, 
  selectedComponent = 0,
  variableThreshold = 100
}) => {
  const { theme } = useTheme();
  const { qualitativePalette } = usePalette();
  
  // Get the color scheme from the current palette
  const colorScheme = getQualitativePalette(qualitativePalette);
  
  // Determine plot type based on number of variables
  const numVariables = pcaResult.loadings[0]?.length || 0;
  const autoPlotType = useMemo(() => {
    return numVariables > variableThreshold ? 'line' : 'bar';
  }, [numVariables, variableThreshold]);
  
  // State for manual plot type override
  const [manualPlotType, setManualPlotType] = useState<'bar' | 'line' | null>(null);
  const plotType = manualPlotType || autoPlotType;
  const isManual = manualPlotType !== null && manualPlotType !== autoPlotType;

  // Transform data to Plotly format
  const plotlyData = transformToLoadingsPlotData(
    pcaResult,
    selectedComponent
  );

  // Create config for Plotly component
  const plotlyConfig = createLoadingsPlotConfig(
    plotType as 'bar' | 'line',
    false, // sortByMagnitude - could be made configurable
    theme,
    colorScheme
  );

  // Get component label and variance for display
  const componentLabel = `PC${selectedComponent + 1}`;
  const variance = pcaResult.explained_variance[selectedComponent]?.toFixed(1) || '0';
  
  return (
    <div style={{ width: '100%', height: '100%', display: 'flex', flexDirection: 'column' }}>
      {/* Header with plot type selector */}
      <div style={{ 
        display: 'flex', 
        alignItems: 'center', 
        justifyContent: 'space-between', 
        marginBottom: '8px',
        padding: '0 8px'
      }}>
        <h4 style={{ 
          fontSize: '14px', 
          fontWeight: 500, 
          color: theme === 'dark' ? '#d1d5db' : '#374151',
          margin: 0
        }}>
          {componentLabel} Loadings ({variance}% variance)
        </h4>
        <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
          <span style={{ 
            fontSize: '12px', 
            color: theme === 'dark' ? '#9ca3af' : '#6b7280' 
          }}>
            Plot type:
          </span>
          <select
            value={plotType}
            onChange={(e) => {
              const value = e.target.value as 'bar' | 'line';
              setManualPlotType(value === autoPlotType ? null : value);
            }}
            style={{
              padding: '4px 8px',
              backgroundColor: theme === 'dark' ? '#374151' : '#f3f4f6',
              border: `1px solid ${theme === 'dark' ? '#4b5563' : '#d1d5db'}`,
              borderRadius: '4px',
              fontSize: '12px',
              color: theme === 'dark' ? '#ffffff' : '#111827',
              cursor: 'pointer'
            }}
          >
            <option value="bar">Bar Chart</option>
            <option value="line">Line Chart</option>
          </select>
          {isManual && (
            <span style={{ 
              fontSize: '10px', 
              color: '#eab308' 
            }}>
              (manual)
            </span>
          )}
        </div>
      </div>
      
      {/* Plot container */}
      <div style={{ flex: 1, width: '100%' }}>
        <PCALoadingsPlot 
          data={plotlyData} 
          config={plotlyConfig} 
        />
      </div>
    </div>
  );
};