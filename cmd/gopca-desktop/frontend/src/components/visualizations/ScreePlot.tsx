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