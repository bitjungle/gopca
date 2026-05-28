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

// Plotly-based Temporal Variable Importance Visualization

import React from 'react';
import { PCATemporalVariableImportancePlot, useTheme } from '@gopca/ui-components';
import { PCAResult } from '../../types';
import {
  transformToTemporalVariableImportancePlotData,
  createTemporalVariableImportancePlotConfig
} from '../../utils/plotlyDataTransform';
import { usePalette } from '../../contexts/PaletteContext';
import { getSequentialPalette } from '../../utils/colorPalettes';

interface TemporalVariableImportancePlotProps {
  pcaResult: PCAResult;
  maxComponents?: number;
  fontScale?: number;
}

export const TemporalVariableImportancePlot: React.FC<TemporalVariableImportancePlotProps> = ({
  pcaResult,
  maxComponents = 10,
  fontScale = 1.0
}) => {
  const { theme } = useTheme();
  const { sequentialPalette } = usePalette();

  // Get the sequential palette colors
  const paletteColors = getSequentialPalette(sequentialPalette);

  // Transform data to Plotly format
  const plotlyData = transformToTemporalVariableImportancePlotData(pcaResult);

  // Handle case where temporal variable importance is not available
  if (!plotlyData) {
    return (
      <div className="flex items-center justify-center h-full text-gray-500 dark:text-gray-400">
        <p>Variable importance data not available. This visualization is only for Temporal PCA.</p>
      </div>
    );
  }

  // Create config for Plotly component
  const plotlyConfig = createTemporalVariableImportancePlotConfig(
    maxComponents,
    theme,
    paletteColors,
    fontScale
  );

  return (
    <div style={{ width: '100%', height: '100%' }}>
      <PCATemporalVariableImportancePlot
        data={plotlyData}
        config={plotlyConfig}
      />
    </div>
  );
};

export default TemporalVariableImportancePlot;