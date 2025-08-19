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
  plotType?: 'bar' | 'line'; // Optional plot type override from parent
}

export const LoadingsPlot: React.FC<LoadingsPlotProps> = ({ 
  pcaResult, 
  selectedComponent = 0,
  variableThreshold = 100,
  plotType: plotTypeProp
}) => {
  const { theme } = useTheme();
  const { qualitativePalette } = usePalette();
  
  // Get the color scheme from the current palette
  const colorScheme = getQualitativePalette(qualitativePalette);
  
  // Determine plot type based on number of variables
  const numVariables = pcaResult.loadings.length || 0;
  const autoPlotType = useMemo(() => {
    return numVariables > variableThreshold ? 'line' : 'bar';
  }, [numVariables, variableThreshold]);
  
  // Use plot type from prop or auto-determine
  const plotType = plotTypeProp || autoPlotType;

  // Transform data to Plotly format
  const plotlyData = transformToLoadingsPlotData(
    pcaResult,
    selectedComponent
  );

  // Create config for Plotly component
  const plotlyConfig = createLoadingsPlotConfig(
    plotType,
    false, // sortByMagnitude - could be made configurable
    theme,
    colorScheme,
    numVariables,
    variableThreshold
  );

  return (
    <div style={{ width: '100%', height: '100%' }}>
      <PCALoadingsPlot 
        data={plotlyData} 
        config={plotlyConfig} 
      />
    </div>
  );
};