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

// Plotly-based PCA Loadings Plot

import React, { useMemo } from 'react';
import { PCALoadingsPlot, useTheme } from '@gopca/ui-components';
import { PCAResult } from '../../types';
import {
  transformToLoadingsPlotData,
  createLoadingsPlotConfig
} from '../../utils/plotlyDataTransform';
import { usePalette } from '../../contexts/PaletteContext';
import { getQualitativePalette } from '../../utils/colorPalettes';

interface LoadingsPlotProps {
  pcaResult: PCAResult;
  selectedComponent?: number; // 0-based index
  variableThreshold?: number; // Threshold for auto-switching between bar and line
  plotType?: 'bar' | 'line'; // Optional plot type override from parent
  fontScale?: number;
}

export const LoadingsPlot: React.FC<LoadingsPlotProps> = ({
  pcaResult,
  selectedComponent = 0,
  variableThreshold = 100,
  plotType: plotTypeProp,
  fontScale = 1.0
}) => {
  const { theme } = useTheme();
  const { qualitativePalette } = usePalette();

  // Get the color scheme from the current palette
  const colorScheme = getQualitativePalette(qualitativePalette);

  // Determine plot type based on number of variables
  const numVariables = pcaResult.loadings.length || 0;
  const autoPlotType = useMemo(() => {
    return numVariables > variableThreshold ? 'line' : 'bar';
  }, [numVariables, variableThreshold]);

  // Use plot type from prop or auto-determine
  const plotType = plotTypeProp || autoPlotType;

  // Transform data to Plotly format
  const plotlyData = transformToLoadingsPlotData(
    pcaResult,
    selectedComponent
  );

  // Create config for Plotly component
  const plotlyConfig = createLoadingsPlotConfig(
    plotType,
    false, // sortByMagnitude - could be made configurable
    theme,
    colorScheme,
    numVariables,
    variableThreshold,
    fontScale
  );

  return (
    <div style={{ width: '100%', height: '100%' }}>
      <PCALoadingsPlot
        data={plotlyData}
        config={plotlyConfig}
      />
    </div>
  );
};