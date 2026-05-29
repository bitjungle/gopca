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

// Plotly-based 3D PCA Biplot

import React from 'react';
import { PCA3DBiplot, useTheme } from '@gopca/ui-components';
import { PCAResult, EllipseParams } from '../../types';
import {
  transformToBiplot3DData,
  createBiplot3DConfig
} from '../../utils/plotlyDataTransform';
import { usePalette } from '../../contexts/PaletteContext';
import { getQualitativePalette, getSequentialPalette } from '../../utils/colorPalettes';

interface Biplot3DProps {
  pcaResult: PCAResult;
  rowNames: string[];
  xComponent?: number; // 0-based index
  yComponent?: number; // 0-based index
  zComponent?: number; // 0-based index
  groupColumn?: string | null;
  groupLabels?: string[];
  groupValues?: number[]; // For continuous columns
  groupType?: 'categorical' | 'continuous';
  groupEllipses?: Record<string, EllipseParams>;
  showEllipses?: boolean;
  confidenceLevel?: 0.90 | 0.95 | 0.99;
  showRowLabels?: boolean;
  maxLabelsToShow?: number;
  showLoadings?: boolean;
  vectorScale?: number;
  maxVariables?: number; // Maximum number of loading vectors to display
  fontScale?: number;
}

/**
 * 3D Biplot visualization combining scores and loading vectors in 3D space
 */
export const Biplot3D: React.FC<Biplot3DProps> = ({
  pcaResult,
  rowNames,
  xComponent = 0,
  yComponent = 1,
  zComponent = 2,
  groupColumn: _groupColumn,
  groupLabels,
  groupValues,
  groupType,
  showRowLabels = false,
  maxLabelsToShow = 10,
  showLoadings = true,
  vectorScale = 1.0,
  maxVariables = 50,
  fontScale = 1.0
}) => {
  const theme = useTheme();
  const { qualitativePalette, sequentialPalette } = usePalette();

  // Get the appropriate color palette
  const colorScheme = groupType === 'continuous'
    ? getSequentialPalette(sequentialPalette)
    : getQualitativePalette(qualitativePalette);

  // Transform data for 3D biplot
  const biplotData = transformToBiplot3DData(
    pcaResult,
    rowNames,
    groupLabels,
    groupValues,
    groupType,
    xComponent,
    yComponent,
    zComponent
  );

  // Create configuration
  const config = createBiplot3DConfig({
    theme: theme.theme,
    colorScheme,
    showScores: true,
    showLoadings,
    showLabels: showRowLabels,
    maxLabels: maxLabelsToShow,
    vectorScale,
    maxVariables,
    fontScale
  });

  return (
    <PCA3DBiplot
      data={biplotData}
      config={config}
    />
  );
};