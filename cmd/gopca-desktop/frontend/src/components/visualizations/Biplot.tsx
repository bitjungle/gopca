// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Plotly-based PCA Biplot

import React from 'react';
import { PCABiplot, useTheme } from '@gopca/ui-components';
import { PCAResult, EllipseParams } from '../../types';
import {
  transformToBiplotData,
  createBiplotConfig
} from '../../utils/plotlyDataTransform';
import { usePalette } from '../../contexts/PaletteContext';
import { getQualitativePalette, getSequentialPalette } from '../../utils/colorPalettes';

interface BiplotProps {
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
  showLoadings?: boolean;
  vectorScale?: number;
}

export const Biplot: React.FC<BiplotProps> = ({
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
  showLoadings = true,
  vectorScale = 1.0
}) => {
  const { theme } = useTheme();
  const { qualitativePalette, sequentialPalette } = usePalette();
  
  // Get the appropriate color scheme based on palette mode
  const colorScheme = groupType === 'continuous' 
    ? getSequentialPalette(sequentialPalette)
    : getQualitativePalette(qualitativePalette);
  
  // Transform data to Plotly format
  const plotlyData = transformToBiplotData(
    pcaResult,
    rowNames,
    groupLabels,
    groupValues,
    groupType
  );

  // Create config for Plotly component with additional settings
  const plotlyConfig = {
    ...createBiplotConfig(
      xComponent, 
      yComponent, 
      showRowLabels, 
      theme, 
      colorScheme,
      showEllipses,
      confidenceLevel
    ),
    showLoadings,
    vectorScale,
    labelThreshold: maxLabelsToShow
  };

  return (
    <div style={{ width: '100%', height: '100%' }}>
      <PCABiplot 
        data={plotlyData} 
        config={plotlyConfig} 
      />
    </div>
  );
};