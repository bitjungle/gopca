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
  maxComponents,
  fontScale
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
  const plotlyConfig = createEigencorrelationPlotConfig(maxComponents, theme, colorScheme, fontScale);

  return (
    <div style={{ width: '100%', height: '100%' }}>
      <PCAEigencorrelationPlot
        data={plotlyData}
        config={plotlyConfig}
      />
    </div>
  );
};