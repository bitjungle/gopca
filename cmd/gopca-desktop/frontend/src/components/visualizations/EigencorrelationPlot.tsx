// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Plotly-based Eigencorrelation Plot

import React from 'react';
import { PCAEigencorrelationPlot, useTheme } from '@gopca/ui-components';
import { PCAResult } from '../../types';
import {
  transformToEigencorrelationPlotData,
  createEigencorrelationPlotConfig
} from '../../utils/plotlyDataTransform';
import { usePalette } from '../../contexts/PaletteContext';
import { getSequentialPalette } from '../../utils/colorPalettes';

interface EigencorrelationPlotProps {
  pcaResult: PCAResult;
  maxComponents?: number;
  fontScale?: number;
}

export const EigencorrelationPlot: React.FC<EigencorrelationPlotProps> = ({
  pcaResult,
  maxComponents
}) => {
  const { theme } = useTheme();
  const { sequentialPalette } = usePalette();

  // Get the color scheme from the current sequential palette
  // Note: EigencorrelationPlot uses RdBu colorscale for correlation heatmap,
  // but we include this for consistency
  const colorScheme = getSequentialPalette(sequentialPalette);

  // Transform data to Plotly format
  const plotlyData = transformToEigencorrelationPlotData(pcaResult);

  // Check if eigencorrelations are available
  if (!plotlyData) {
    return (
      <div style={{ width: '100%', height: '100%', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
        <p style={{ color: theme === 'dark' ? '#9ca3af' : '#6b7280', textAlign: 'center' }}>
          No eigencorrelation data available. Please ensure metadata variables were included when calculating PCA.
        </p>
      </div>
    );
  }

  // Create config for Plotly component
  const plotlyConfig = createEigencorrelationPlotConfig(maxComponents, theme, colorScheme);

  return (
    <div style={{ width: '100%', height: '100%' }}>
      <PCAEigencorrelationPlot
        data={plotlyData}
        config={plotlyConfig}
      />
    </div>
  );
};