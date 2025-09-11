// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Plotly-based Temporal Variable Importance Visualization

import React from 'react';
import { PCATemporalVariableImportancePlot, useTheme } from '@gopca/ui-components';
import { PCAResult } from '../../types';
import {
  transformToTemporalVariableImportancePlotData,
  createTemporalVariableImportancePlotConfig
} from '../../utils/plotlyDataTransform';
import { usePalette } from '../../contexts/PaletteContext';
import { getQualitativePalette } from '../../utils/colorPalettes';

interface TemporalVariableImportancePlotProps {
  pcaResult: PCAResult;
  maxComponents?: number;
  fontScale?: number;
}

export const TemporalVariableImportancePlot: React.FC<TemporalVariableImportancePlotProps> = ({
  pcaResult,
  maxComponents = 10,
  fontScale = 1.0
}) => {
  const { theme } = useTheme();
  const { qualitativePalette } = usePalette();

  // Get the color scheme from the current palette
  const colorScheme = getQualitativePalette(qualitativePalette);

  // Transform data to Plotly format
  const plotlyData = transformToTemporalVariableImportancePlotData(pcaResult);

  // Handle case where temporal variable importance is not available
  if (!plotlyData) {
    return (
      <div className="flex items-center justify-center h-full text-gray-500 dark:text-gray-400">
        <p>Variable importance data not available. This visualization is only for Temporal PCA.</p>
      </div>
    );
  }

  // Create config for Plotly component
  const plotlyConfig = createTemporalVariableImportancePlotConfig(
    maxComponents,
    theme,
    colorScheme,
    fontScale
  );

  return (
    <div style={{ width: '100%', height: '100%' }}>
      <PCATemporalVariableImportancePlot
        data={plotlyData}
        config={plotlyConfig}
      />
    </div>
  );
};

export default TemporalVariableImportancePlot;