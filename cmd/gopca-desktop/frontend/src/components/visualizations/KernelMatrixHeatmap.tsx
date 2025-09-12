// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Kernel Matrix Heatmap - Visualizes pairwise similarities in kernel space

import React from 'react';
import Plot from 'react-plotly.js';
import { useTheme } from '@gopca/ui-components';
import { PCAResult } from '../../types';
import { usePalette } from '../../contexts/PaletteContext';

interface KernelMatrixHeatmapProps {
  pcaResult: PCAResult;
  rowNames?: string[];
  fontScale?: number;
  showValues?: boolean;
  colorScale?: 'Viridis' | 'Blues' | 'RdBu' | 'YlOrRd' | 'Jet';
}

export const KernelMatrixHeatmap: React.FC<KernelMatrixHeatmapProps> = ({
  pcaResult,
  rowNames = [],
  fontScale = 1.0,
  showValues = false,
  colorScale = 'Viridis'
}) => {
  const { theme } = useTheme();
  const isDark = theme === 'dark';

  // Check if kernel matrix is available
  if (!pcaResult.kernel_matrix || pcaResult.kernel_matrix.length === 0) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className={`text-center p-8 ${isDark ? 'text-gray-400' : 'text-gray-600'}`}>
          <p className="text-lg mb-2">Kernel matrix not available</p>
          <p className="text-sm">
            The kernel matrix is only computed for datasets with 1000 or fewer samples to conserve memory.
          </p>
        </div>
      </div>
    );
  }

  // Prepare data for Plotly heatmap
  const matrix = pcaResult.kernel_matrix;
  const n = matrix.length;

  // Create sample labels
  const sampleLabels = rowNames.length > 0
    ? rowNames.slice(0, n)
    : Array.from({ length: n }, (_, i) => `Sample ${i + 1}`);

  // Create the heatmap trace
  const trace = {
    z: matrix,
    x: sampleLabels,
    y: sampleLabels,
    type: 'heatmap' as const,
    colorscale: colorScale,
    showscale: true,
    hovertemplate:
      '<b>Sample %{y}</b> ↔ <b>Sample %{x}</b><br>' +
      'Similarity: %{z:.4f}<br>' +
      '<extra></extra>',
    colorbar: {
      title: {
        text: 'Kernel<br>Similarity',
        side: 'right' as const,
        font: {
          size: 12 * fontScale,
          color: isDark ? '#E5E7EB' : '#1F2937'
        }
      },
      tickfont: {
        size: 10 * fontScale,
        color: isDark ? '#E5E7EB' : '#1F2937'
      },
      thickness: 15,
      len: 0.8,
      x: 1.02
    }
  };

  // Prepare layout
  const layout: any = {
    title: {
      text: 'Kernel Matrix Heatmap',
      font: {
        size: 16 * fontScale,
        color: isDark ? '#E5E7EB' : '#1F2937'
      }
    },
    xaxis: {
      title: 'Samples',
      side: 'bottom' as const,
      tickangle: -45,
      tickfont: {
        size: Math.max(8, 10 * fontScale * (n > 50 ? 0.5 : 1)),
        color: isDark ? '#9CA3AF' : '#4B5563'
      },
      titlefont: {
        size: 12 * fontScale,
        color: isDark ? '#E5E7EB' : '#1F2937'
      },
      showticklabels: n <= 100, // Hide labels for large matrices
      gridcolor: isDark ? '#374151' : '#E5E7EB',
      zerolinecolor: isDark ? '#4B5563' : '#9CA3AF',
      // Kernel matrix is always square (n×n), so we enforce a 1:1 aspect ratio
      // to avoid misleading rectangular representation
      scaleanchor: 'y', // Link x and y axis scales
      scaleratio: 1 // Enforce 1:1 aspect ratio
    },
    yaxis: {
      title: 'Samples',
      tickfont: {
        size: Math.max(8, 10 * fontScale * (n > 50 ? 0.5 : 1)),
        color: isDark ? '#9CA3AF' : '#4B5563'
      },
      titlefont: {
        size: 12 * fontScale,
        color: isDark ? '#E5E7EB' : '#1F2937'
      },
      showticklabels: n <= 100, // Hide labels for large matrices
      gridcolor: isDark ? '#374151' : '#E5E7EB',
      zerolinecolor: isDark ? '#4B5563' : '#9CA3AF',
      autorange: 'reversed' as const,
      scaleanchor: 'x' // Link to x axis
    },
    paper_bgcolor: isDark ? '#1F2937' : '#FFFFFF',
    plot_bgcolor: isDark ? '#111827' : '#F9FAFB',
    margin: {
      l: n <= 100 ? 80 : 60,
      r: 80,
      t: 60,
      b: n <= 100 ? 80 : 60
    },
    width: undefined,
    height: undefined,
    autosize: true,
    annotations: showValues && n <= 20 ?
      matrix.flatMap((row, i) =>
        row.map((value, j) => ({
          x: sampleLabels[j],
          y: sampleLabels[i],
          text: value.toFixed(2),
          showarrow: false,
          font: {
            size: 8 * fontScale,
            color: value > 0.5 ? '#FFFFFF' : '#000000'
          }
        }))
      ) : []
  };

  const config: any = {
    responsive: true,
    displayModeBar: true,
    displaylogo: false,
    modeBarButtonsToRemove: ['select2d', 'lasso2d', 'autoScale2d'],
    toImageButtonOptions: {
      format: 'png' as const,
      filename: 'kernel_matrix_heatmap',
      height: 800,
      width: 800,
      scale: 2
    }
  };

  return (
    <div style={{ width: '100%', height: '100%' }}>
      <Plot
        data={[trace]}
        layout={layout}
        config={config}
        style={{ width: '100%', height: '100%' }}
        useResizeHandler={true}
      />
    </div>
  );
};