// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Plotly-based Temporal Loadings (U matrix) Visualization

import React from 'react';
import { PCATemporalLoadingsPlot, useTheme } from '@gopca/ui-components';
import { PCAResult } from '../../types';
import {
  transformToTemporalLoadingsPlotData,
  createTemporalLoadingsPlotConfig
} from '../../utils/plotlyDataTransform';
import { usePalette } from '../../contexts/PaletteContext';
import { getQualitativePalette } from '../../utils/colorPalettes';

interface TemporalLoadingsPlotProps {
  pcaResult: PCAResult;
  maxComponents?: number;
  fontScale?: number;
}

export const TemporalLoadingsPlot: React.FC<TemporalLoadingsPlotProps> = ({
  pcaResult,
  maxComponents = 5,
  fontScale = 1.0
}) => {
  const { theme } = useTheme();
  const { qualitativePalette } = usePalette();

  // Get the color scheme from the current palette
  const colorScheme = getQualitativePalette(qualitativePalette);

  // Transform data to Plotly format
  const plotlyData = transformToTemporalLoadingsPlotData(pcaResult);

  // Handle case where temporal eigenvectors are not available
  if (!plotlyData) {
    return (
      <div className="flex items-center justify-center h-full text-gray-500 dark:text-gray-400">
        <p>Temporal eigenvectors not available. This visualization is only for Temporal PCA.</p>
      </div>
    );
  }

  // Create config for Plotly component
  const plotlyConfig = createTemporalLoadingsPlotConfig(
    maxComponents,
    theme,
    colorScheme,
    fontScale
  );

  return (
    <div style={{ width: '100%', height: '100%' }}>
      <PCATemporalLoadingsPlot
        data={plotlyData}
        config={plotlyConfig}
      />
    </div>
  );
};

export default TemporalLoadingsPlot;