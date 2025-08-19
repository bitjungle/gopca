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
  showRowLabels?: boolean;
  maxLabelsToShow?: number;
}

export const DiagnosticScatterPlot: React.FC<DiagnosticScatterPlotProps> = ({
  pcaResult,
  rowNames,
  groupColumn,
  groupLabels,
  showThresholds = true,
  confidenceLevel = 0.95,
  showRowLabels = false,
  maxLabelsToShow = 10
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

  // Select appropriate thresholds based on confidence level
  // T² limit represents Hotelling's T-squared (leverage in model space)
  // Q limit represents Squared Prediction Error (residuals orthogonal to model)
  const mahalanobisThreshold = confidenceLevel === 0.99 ? 
    pcaResult.t2_limit_99 : pcaResult.t2_limit_95;
  const rssThreshold = confidenceLevel === 0.99 ? 
    pcaResult.q_limit_99 : pcaResult.q_limit_95;

  // Create config for Plotly component with label settings
  const plotlyConfig = {
    ...createDiagnosticPlotConfig(
      showThresholds,
      confidenceLevel,
      theme,
      colorScheme,
      mahalanobisThreshold,
      rssThreshold
    ),
    showLabels: showRowLabels,
    labelThreshold: maxLabelsToShow
  };

  return (
    <div style={{ width: '100%', height: '100%' }}>
      <PCADiagnosticPlot 
        data={plotlyData} 
        config={plotlyConfig} 
      />
    </div>
  );
};