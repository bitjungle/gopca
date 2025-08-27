// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Plotly-based PCA Scores Plot

import React from 'react';
import { PCAScoresPlot, useTheme } from '@gopca/ui-components';
import { PCAResult, EllipseParams } from '../../types';
import {
  transformToScoresPlotData,
  createScoresPlotConfig
} from '../../utils/plotlyDataTransform';
import { usePalette } from '../../contexts/PaletteContext';
import { getQualitativePalette, getSequentialPalette } from '../../utils/colorPalettes';

interface ScoresPlotProps {
  pcaResult: PCAResult;
  rowNames: string[];
  xComponent?: number; // 0-based index
  yComponent?: number; // 0-based index
  groupColumn?: string | null;
  groupLabels?: string[];
  groupValues?: number[]; // For continuous columns
  groupType?: 'categorical' | 'continuous';
  groupEllipses?: Record<string, EllipseParams>;
  showEllipses?: boolean;
  confidenceLevel?: 0.90 | 0.95 | 0.99;
  showRowLabels?: boolean;
  maxLabelsToShow?: number;
  fontScale?: number;
}

export const ScoresPlot: React.FC<ScoresPlotProps> = ({
  pcaResult,
  rowNames,
  xComponent = 0,
  yComponent = 1,
  groupColumn,
  groupLabels,
  groupValues,
  groupType = 'categorical',
  groupEllipses,
  showEllipses = false,
  confidenceLevel = 0.95,
  showRowLabels = false,
  maxLabelsToShow = 10,
  fontScale = 1.0
}) => {
  const { theme } = useTheme();
  const { qualitativePalette, sequentialPalette, mode } = usePalette();

  // Get the appropriate color scheme based on palette mode
  const colorScheme = groupType === 'continuous'
    ? getSequentialPalette(sequentialPalette)
    : getQualitativePalette(qualitativePalette);

  // Transform data to Plotly format
  const plotlyData = transformToScoresPlotData(
    pcaResult,
    rowNames,
    groupLabels,
    groupValues,
    groupType,
    xComponent,
    yComponent
  );

  // Create config for Plotly component
  const plotlyConfig = createScoresPlotConfig(
    xComponent,
    yComponent,
    showEllipses,
    confidenceLevel,
    showRowLabels,
    maxLabelsToShow,
    theme,
    colorScheme,
    fontScale
  );

  return (
    <div style={{ width: '100%', height: '100%' }}>
      <PCAScoresPlot
        data={plotlyData}
        config={plotlyConfig}
      />
    </div>
  );
};