// GoPCA Suite
//
// Copyright © 2025-2026 Rune Mathisen <devel@bitjungle.com>
//
// This file is part of GoPCA Suite.
//
// GoPCA Suite is source-available software with free binary redistribution.
// Official compiled binary releases may be used and redistributed free of charge
// under the GoPCA Suite Source-Available Freeware License.
//
// The source code is provided for viewing, review, education, security analysis,
// research, interoperability analysis, and evaluation only.
//
// Modification, redistribution, publication, sublicensing, reuse, incorporation
// into another project, or creation of derivative works based on the source code
// is not permitted without prior written permission from the copyright holder.
//
// Usage Restriction: GoPCA Suite may not be used, directly or indirectly, for
// military, warfare, weapons, intelligence, surveillance, targeting, or
// law-enforcement surveillance applications.
//
// See LICENSE for the full license terms.

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
  fontScale?: number;
}

export const Scores3DPlot: React.FC<Scores3DPlotProps> = ({
  pcaResult,
  rowNames,
  xComponent = 0,
  yComponent = 1,
  zComponent = 2,
  groupColumn: _groupColumn,
  groupLabels,
  groupValues,
  groupType = 'categorical',
  showRowLabels = false,
  maxLabelsToShow = 10,
  fontScale = 1.0
}) => {
  const { theme } = useTheme();
  const { qualitativePalette, sequentialPalette } = usePalette();

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
    colorScheme,
    fontScale
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