// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

import React from 'react';
import { toPng, toSvg } from 'html-to-image';
import { saveAs } from 'file-saver';
import { SaveFile } from '../../wailsjs/go/main/App';
import { ExportButton as SharedExportButton, useTheme, type ExportConfig } from '@gopca/ui-components';

interface ChartExportButtonProps {
  chartRef: React.RefObject<HTMLDivElement>;
  fileName: string;
  className?: string;
}

export const ExportButton: React.FC<ChartExportButtonProps> = ({ chartRef, fileName, className = '' }) => {
  const { theme } = useTheme();

  const handleExportPNG = async () => {
    if (!chartRef.current) {
      throw new Error('Chart element not found');
    }

    // Wait for fonts to be fully loaded
    await document.fonts.ready;

    // Give Plotly time to fully render all text elements
    await new Promise(resolve => setTimeout(resolve, 300));

    // Get the chart element and its bounds
    const chartElement = chartRef.current;
    const bounds = chartElement.getBoundingClientRect();

    const dataUrl = await toPng(chartElement, {
      backgroundColor: theme === 'dark' ? '#1F2937' : '#FFFFFF',
      width: bounds.width,
      height: bounds.height,
      pixelRatio: 2,
      cacheBust: true,
      style: {
        fontFamily: '"Nunito", -apple-system, BlinkMacSystemFont, "Segoe UI", "Roboto", sans-serif'
      }
    });

    // Check if we're in Wails environment
    if (window.go && window.go.main && window.go.main.App && window.go.main.App.SaveFile) {
      // Use Wails backend for saving
      await SaveFile(`${fileName}.png`, dataUrl);
    } else {
      // Fallback to browser download for development
      saveAs(dataUrl, `${fileName}.png`);
    }
  };

  const handleExportSVG = async () => {
    if (!chartRef.current) {
      throw new Error('Chart element not found');
    }

    const dataUrl = await toSvg(chartRef.current, {
      backgroundColor: theme === 'dark' ? '#1F2937' : '#FFFFFF'
    });

    // Check if we're in Wails environment
    if (window.go && window.go.main && window.go.main.App && window.go.main.App.SaveFile) {
      // Use Wails backend for saving
      await SaveFile(`${fileName}.svg`, dataUrl);
    } else {
      // Fallback to browser download for development
      const blob = new Blob([dataUrl], { type: 'image/svg+xml' });
      saveAs(blob, `${fileName}.svg`);
    }
  };

  const exportFormats: ExportConfig[] = [
    {
      format: 'png',
      label: 'Export as PNG',
      handler: handleExportPNG
    },
    {
      format: 'svg',
      label: 'Export as SVG',
      handler: handleExportSVG
    }
  ];

  return (
    <SharedExportButton
      formats={exportFormats}
      label="Export"
      className={className}
      size="md"
    />
  );
};