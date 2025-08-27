// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Plotly-based PCA Scree Plot

import React from 'react';
import { PCAScreePlot, useTheme } from '@gopca/ui-components';
import { PCAResult } from '../../types';
import {
  transformToScreePlotData,
  createScreePlotConfig
} from '../../utils/plotlyDataTransform';
import { usePalette } from '../../contexts/PaletteContext';
import { getQualitativePalette } from '../../utils/colorPalettes';

interface ScreePlotProps {
  pcaResult: PCAResult;
  showCumulative?: boolean;
  elbowThreshold?: number; // Optional: highlight components explaining this % variance
  fontScale?: number;
}

export const ScreePlot: React.FC<ScreePlotProps> = ({
  pcaResult,
  showCumulative = true,
  elbowThreshold = 80,
  fontScale = 1.0
}) => {
  const { theme } = useTheme();
  const { qualitativePalette } = usePalette();

  // Get the color scheme from the current palette
  const colorScheme = getQualitativePalette(qualitativePalette);

  // Transform data to Plotly format
  const plotlyData = transformToScreePlotData(pcaResult);

  // Create config for Plotly component
  const plotlyConfig = createScreePlotConfig(
    showCumulative,
    elbowThreshold,
    theme,
    colorScheme,
    fontScale
  );

  return (
    <div style={{ width: '100%', height: '100%' }}>
      <PCAScreePlot
        data={plotlyData}
        config={plotlyConfig}
      />
    </div>
  );
};