// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

// Plotly-based 3D PCA Scores Plot

import React from 'react';
import { PCA3DScoresPlot, useTheme } from '@gopca/ui-components';
import { PCAResult } from '../../types';
import {
  transformToScores3DPlotData,
  createScores3DPlotConfig
} from '../../utils/plotlyDataTransform';
import { usePalette } from '../../contexts/PaletteContext';
import { getQualitativePalette, getSequentialPalette } from '../../utils/colorPalettes';

interface Scores3DPlotProps {
  pcaResult: PCAResult;
  rowNames: string[];
  xComponent?: number; // 0-based index
  yComponent?: number; // 0-based index
  zComponent?: number; // 0-based index
  groupColumn?: string | null;
  groupLabels?: string[];
  groupValues?: number[]; // For continuous columns
  groupType?: 'categorical' | 'continuous';
  showRowLabels?: boolean;
  maxLabelsToShow?: number;
}

export const Scores3DPlot: React.FC<Scores3DPlotProps> = ({
  pcaResult,
  rowNames,
  xComponent = 0,
  yComponent = 1,
  zComponent = 2,
  groupColumn,
  groupLabels,
  groupValues,
  groupType = 'categorical',
  showRowLabels = false,
  maxLabelsToShow = 10
}) => {
  const { theme } = useTheme();
  const { qualitativePalette, sequentialPalette, mode } = usePalette();

  // Get the appropriate color scheme based on palette mode
  const colorScheme = groupType === 'continuous'
    ? getSequentialPalette(sequentialPalette)
    : getQualitativePalette(qualitativePalette);

  // Transform data to Plotly 3D format
  const plotlyData = transformToScores3DPlotData(
    pcaResult,
    rowNames,
    groupLabels,
    groupValues,
    groupType,
    xComponent,
    yComponent,
    zComponent
  );

  // Create config for Plotly 3D component
  const plotlyConfig = createScores3DPlotConfig(
    xComponent,
    yComponent,
    zComponent,
    showRowLabels,
    maxLabelsToShow,
    theme,
    colorScheme
  );

  return (
    <div style={{ width: '100%', height: '100%' }}>
      <PCA3DScoresPlot
        data={plotlyData}
        config={plotlyConfig}
      />
    </div>
  );
};