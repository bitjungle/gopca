// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

import React, { useState, useCallback } from 'react';

export type ExportFormat = 'png' | 'svg' | 'csv' | 'json' | 'xlsx';

export interface ExportConfig {
  format: ExportFormat;
  label?: string;
  handler: (format: ExportFormat) => Promise<void>;
}

export interface ExportButtonProps {
  formats: ExportConfig[];
  label?: string;
  className?: string;
  disabled?: boolean;
  size?: 'sm' | 'md' | 'lg';
}

export const ExportButton: React.FC<ExportButtonProps> = ({
  formats,
  label = 'Export',
  className = '',
  disabled = false,
  size = 'md'
}) => {
  const [isExporting, setIsExporting] = useState(false);
  const [showMenu, setShowMenu] = useState(false);

  const handleExport = useCallback(async (config: ExportConfig) => {
    if (isExporting || disabled) {
return;
}

    setIsExporting(true);
    setShowMenu(false);

    try {
      await config.handler(config.format);
    } catch (error) {
      console.error(`Failed to export as ${config.format}:`, error);
    } finally {
      setIsExporting(false);
    }
  }, [isExporting, disabled]);

  const sizeClasses = {
    sm: 'px-2 py-1 text-xs',
    md: 'px-3 py-1 text-sm',
    lg: 'px-4 py-2 text-base'
  };

  const iconSize = {
    sm: 'w-3 h-3',
    md: 'w-4 h-4',
    lg: 'w-5 h-5'
  };

  const formatLabels: Record<ExportFormat, string> = {
    png: 'Export as PNG',
    svg: 'Export as SVG',
    csv: 'Export as CSV',
    json: 'Export as JSON',
    xlsx: 'Export as Excel'
  };

  if (formats.length === 0) {
    return null;
  }

  const isDisabled = disabled || isExporting;

  if (formats.length === 1) {
    const config = formats[0];
    return (
      <button
        onClick={() => handleExport(config)}
        disabled={isDisabled}
        className={`
          ${sizeClasses[size]} rounded-lg transition-colors inline-flex items-center gap-2
          ${isDisabled
            ? 'bg-gray-400 dark:bg-gray-600 text-gray-200 dark:text-gray-400 cursor-not-allowed'
            : 'bg-gray-200 dark:bg-gray-700 hover:bg-gray-300 dark:hover:bg-gray-600 text-gray-700 dark:text-gray-300'
          }
          ${className}
        `}
      >
        {isExporting ? (
          <>
            <svg className={`animate-spin ${iconSize[size]}`} viewBox="0 0 24 24">
              <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
              <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
            </svg>
            Exporting...
          </>
        ) : (
          <>
            <svg className={iconSize[size]} fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
            </svg>
            {config.label || label}
          </>
        )}
      </button>
    );
  }

  return (
    <div className={`relative ${className}`}>
      <button
        onClick={() => setShowMenu(!showMenu)}
        disabled={isDisabled}
        className={`
          ${sizeClasses[size]} rounded-lg transition-colors inline-flex items-center gap-2
          ${isDisabled
            ? 'bg-gray-400 dark:bg-gray-600 text-gray-200 dark:text-gray-400 cursor-not-allowed'
            : 'bg-gray-200 dark:bg-gray-700 hover:bg-gray-300 dark:hover:bg-gray-600 text-gray-700 dark:text-gray-300'
          }
        `}
      >
        {isExporting ? (
          <>
            <svg className={`animate-spin ${iconSize[size]}`} viewBox="0 0 24 24">
              <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
              <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
            </svg>
            Exporting...
          </>
        ) : (
          <>
            <svg className={iconSize[size]} fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
            </svg>
            {label}
          </>
        )}
      </button>

      {showMenu && !isExporting && (
        <div className="absolute right-0 mt-2 w-40 bg-white dark:bg-gray-800 rounded-lg shadow-lg border border-gray-200 dark:border-gray-700 z-10">
          {formats.map((config, index) => (
            <button
              key={config.format}
              onClick={() => handleExport(config)}
              className={`
                w-full text-left px-4 py-2 text-sm text-gray-700 dark:text-gray-300 
                hover:bg-gray-100 dark:hover:bg-gray-700
                ${index === 0 ? 'rounded-t-lg' : ''}
                ${index === formats.length - 1 ? 'rounded-b-lg' : ''}
              `}
            >
              {config.label || formatLabels[config.format]}
            </button>
          ))}
        </div>
      )}
    </div>
  );
};