// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

import React, { useState } from 'react';
import { toPng, toSvg } from 'html-to-image';
import { saveAs } from 'file-saver';
import { SaveFile } from '../../wailsjs/go/main/App';
import { useTheme } from '@gopca/ui-components';

interface ExportButtonProps {
  chartRef: React.RefObject<HTMLDivElement>;
  fileName: string;
  className?: string;
}

export const ExportButton: React.FC<ExportButtonProps> = ({ chartRef, fileName, className = '' }) => {
  const [isExporting, setIsExporting] = useState(false);
  const [showMenu, setShowMenu] = useState(false);
  const { theme } = useTheme();

  const handleExport = async (format: 'png' | 'svg') => {
    if (!chartRef.current) return;
    
    setIsExporting(true);
    setShowMenu(false);
    
    try {
      let dataUrl: string;
      
      if (format === 'png') {
        // Wait for fonts to be fully loaded
        await document.fonts.ready;
        
        // Give Recharts time to fully render all text elements
        await new Promise(resolve => setTimeout(resolve, 300));
        
        // Get the chart element and its bounds
        const chartElement = chartRef.current;
        const bounds = chartElement.getBoundingClientRect();
        
        dataUrl = await toPng(chartElement, {
          backgroundColor: theme === 'dark' ? '#1F2937' : '#FFFFFF',
          width: bounds.width,
          height: bounds.height,
          pixelRatio: 2,
          cacheBust: true,
          style: {
            fontFamily: '"Nunito", -apple-system, BlinkMacSystemFont, "Segoe UI", "Roboto", sans-serif',
          },
          // No filter needed for Plotly charts
        });
      } else {
        // SVG export
        dataUrl = await toSvg(chartRef.current, {
          backgroundColor: theme === 'dark' ? '#1F2937' : '#FFFFFF',
        });
      }
      
      // Check if we're in Wails environment
      if (window.go && window.go.main && window.go.main.App && window.go.main.App.SaveFile) {
        // Use Wails backend for saving
        const fullFileName = `${fileName}.${format}`;
        await SaveFile(fullFileName, dataUrl);
      } else {
        // Fallback to browser download for development
        if (format === 'png') {
          saveAs(dataUrl, `${fileName}.png`);
        } else {
          const blob = new Blob([dataUrl], { type: 'image/svg+xml' });
          saveAs(blob, `${fileName}.svg`);
        }
      }
    } catch (error) {
      console.error('Failed to export chart:', error);
    } finally {
      setIsExporting(false);
    }
  };

  return (
    <div className={`relative ${className}`}>
      <button
        onClick={() => setShowMenu(!showMenu)}
        disabled={isExporting}
        className={`
          px-3 py-1 text-sm rounded-lg transition-colors
          ${isExporting 
            ? 'bg-gray-400 dark:bg-gray-600 text-gray-200 dark:text-gray-400 cursor-not-allowed' 
            : 'bg-gray-200 dark:bg-gray-700 hover:bg-gray-300 dark:hover:bg-gray-600 text-gray-700 dark:text-gray-300'
          }
        `}
      >
        {isExporting ? (
          <span className="flex items-center gap-2">
            <svg className="animate-spin h-4 w-4" viewBox="0 0 24 24">
              <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
              <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
            </svg>
            Exporting...
          </span>
        ) : (
          <span className="flex items-center gap-2">
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
            </svg>
            Export
          </span>
        )}
      </button>
      
      {showMenu && !isExporting && (
        <div className="absolute right-0 mt-2 w-32 bg-white dark:bg-gray-800 rounded-lg shadow-lg border border-gray-200 dark:border-gray-700 z-10">
          <button
            onClick={() => handleExport('png')}
            className="w-full text-left px-4 py-2 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-t-lg"
          >
            Export as PNG
          </button>
          <button
            onClick={() => handleExport('svg')}
            className="w-full text-left px-4 py-2 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-b-lg"
          >
            Export as SVG
          </button>
        </div>
      )}
    </div>
  );
};