// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Plotly-based PCA Diagnostic Plot

import React from 'react';
import { PCADiagnosticPlot, useTheme } from '@gopca/ui-components';
import { PCAResult } from '../../types';
import {
  transformToDiagnosticPlotData,
  createDiagnosticPlotConfig
} from '../../utils/plotlyDataTransform';
import { usePalette } from '../../contexts/PaletteContext';
import { getQualitativePalette } from '../../utils/colorPalettes';

interface DiagnosticScatterPlotProps {
  pcaResult: PCAResult;
  rowNames: string[];
  groupColumn?: string | null;
  groupLabels?: string[];
  showThresholds?: boolean;
  confidenceLevel?: number;
}

export const DiagnosticScatterPlot: React.FC<DiagnosticScatterPlotProps> = ({
  pcaResult,
  rowNames,
  groupColumn,
  groupLabels,
  showThresholds = true,
  confidenceLevel = 0.975
}) => {
  const { theme } = useTheme();
  const { qualitativePalette } = usePalette();
  
  // Get the color scheme from the current palette
  const colorScheme = getQualitativePalette(qualitativePalette);
  
  // Transform data to Plotly format
  const plotlyData = transformToDiagnosticPlotData(
    pcaResult,
    rowNames,
    groupLabels
  );

  // Create config for Plotly component
  const plotlyConfig = createDiagnosticPlotConfig(
    showThresholds,
    confidenceLevel,
    theme,
    colorScheme
  );

  return (
    <div style={{ width: '100%', height: '100%' }}>
      <PCADiagnosticPlot 
        data={plotlyData} 
        config={plotlyConfig} 
      />
    </div>
  );
};