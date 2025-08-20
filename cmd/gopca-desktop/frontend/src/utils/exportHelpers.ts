// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

import { toPng, toSvg } from 'html-to-image';
import { saveAs } from 'file-saver';
import { SaveFile } from '../../wailsjs/go/main/App';
import type { ExportFormat } from '@gopca/ui-components';

export const createChartExportHandler = (
  chartRef: React.RefObject<HTMLDivElement>,
  fileName: string,
  theme: 'light' | 'dark'
) => {
  return async (format: ExportFormat) => {
    if (!chartRef.current) {
      throw new Error('Chart reference not available');
    }

    if (format !== 'png' && format !== 'svg') {
      throw new Error(`Unsupported format for chart export: ${format}`);
    }

    let dataUrl: string;

    if (format === 'png') {
      // Wait for fonts to be fully loaded
      await document.fonts.ready;

      // Give charts time to fully render all text elements
      await new Promise(resolve => setTimeout(resolve, 300));

      const chartElement = chartRef.current;
      const bounds = chartElement.getBoundingClientRect();

      // Clone the element to ensure all styles are computed
      const clonedElement = chartElement.cloneNode(true) as HTMLElement;

      dataUrl = await toPng(chartElement, {
        backgroundColor: theme === 'dark' ? '#1F2937' : '#FFFFFF',
        width: bounds.width,
        height: bounds.height,
        pixelRatio: 2,
        cacheBust: true,
        style: {
          fontFamily: '"Nunito", -apple-system, BlinkMacSystemFont, "Segoe UI", "Roboto", sans-serif'
        }
        // No filter needed for Plotly charts
      });
    } else {
      dataUrl = await toSvg(chartRef.current, {
        backgroundColor: theme === 'dark' ? '#1F2937' : '#FFFFFF'
      });
    }

    if (window.go && window.go.main && window.go.main.App && window.go.main.App.SaveFile) {
      const fullFileName = `${fileName}.${format}`;
      await SaveFile(fullFileName, dataUrl);
    } else {
      if (format === 'png') {
        saveAs(dataUrl, `${fileName}.png`);
      } else {
        const blob = new Blob([dataUrl], { type: 'image/svg+xml' });
        saveAs(blob, `${fileName}.svg`);
      }
    }
  };
};