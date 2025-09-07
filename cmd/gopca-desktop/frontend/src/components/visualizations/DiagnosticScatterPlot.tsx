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
import { getQualitativePalette, getSequentialPalette } from '../../utils/colorPalettes';

interface DiagnosticScatterPlotProps {
  pcaResult: PCAResult;
  rowNames: string[];
  groupColumn?: string | null;
  groupLabels?: string[];
  groupValues?: number[]; // For continuous columns
  groupType?: 'categorical' | 'continuous';
  showThresholds?: boolean;
  confidenceLevel?: number;
  showRowLabels?: boolean;
  maxLabelsToShow?: number;
  fontScale?: number;
  onSelectionChange?: (indices: number[]) => void;
  excludedRows?: number[]; // Indices of rows excluded from PCA
}

export const DiagnosticScatterPlot: React.FC<DiagnosticScatterPlotProps> = ({
  pcaResult,
  rowNames,
  groupColumn,
  groupLabels,
  groupValues,
  groupType = 'categorical',
  showThresholds = true,
  confidenceLevel = 0.95,
  showRowLabels = false,
  maxLabelsToShow = 10,
  fontScale,
  onSelectionChange,
  excludedRows = []
}) => {
  const { theme } = useTheme();
  const { qualitativePalette, sequentialPalette } = usePalette();

  // Get the appropriate color scheme based on group type
  const colorScheme = groupType === 'continuous'
    ? getSequentialPalette(sequentialPalette)
    : getQualitativePalette(qualitativePalette);

  // Transform data to Plotly format
  const plotlyData = transformToDiagnosticPlotData(
    pcaResult,
    rowNames,
    groupLabels,
    groupValues,
    groupType
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
      rssThreshold,
      fontScale
    ),
    showLabels: showRowLabels,
    labelThreshold: maxLabelsToShow,
    groupColumn,
    groupType
  };

  // Handle selection callback
  const handleSelection = React.useCallback((indices: number[]) => {
    console.log('DiagnosticScatterPlot wrapper: Selection received', indices);
    if (onSelectionChange) {
      onSelectionChange(indices);
    }
  }, [onSelectionChange]);

  return (
    <div style={{ width: '100%', height: '100%' }}>
      <PCADiagnosticPlot
        data={plotlyData}
        config={plotlyConfig}
        onSelection={handleSelection}
        excludedRows={excludedRows}
      />
    </div>
  );
};