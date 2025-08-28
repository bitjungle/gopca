// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Plotly-based Circle of Correlations

import React from 'react';
import { PCACircleOfCorrelations, useTheme } from '@gopca/ui-components';
import { PCAResult } from '../../types';
import {
  transformToCircleOfCorrelationsData,
  createCircleOfCorrelationsConfig
} from '../../utils/plotlyDataTransform';
import { usePalette } from '../../contexts/PaletteContext';
import { getQualitativePalette } from '../../utils/colorPalettes';

interface CircleOfCorrelationsProps {
  pcaResult: PCAResult;
  xComponent?: number; // 0-based index
  yComponent?: number; // 0-based index
  fontScale?: number;
}

export const CircleOfCorrelations: React.FC<CircleOfCorrelationsProps> = ({
  pcaResult,
  xComponent = 0,
  yComponent = 1,
  fontScale
}) => {
  const { theme } = useTheme();
  const { qualitativePalette } = usePalette();

  // Get the color scheme from the current palette
  const colorScheme = getQualitativePalette(qualitativePalette);

  // Transform data to Plotly format
  const plotlyData = transformToCircleOfCorrelationsData(pcaResult);

  // Create config for Plotly component
  const plotlyConfig = createCircleOfCorrelationsConfig(
    xComponent,
    yComponent,
    theme,
    colorScheme,
    fontScale
  );

  return (
    <div style={{ width: '100%', height: '100%' }}>
      <PCACircleOfCorrelations
        data={plotlyData}
        config={plotlyConfig}
      />
    </div>
  );
};