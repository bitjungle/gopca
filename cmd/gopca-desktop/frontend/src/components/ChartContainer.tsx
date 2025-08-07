// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

import React, { ReactNode } from 'react';
import { PlotControls } from './PlotControls';
import { ExportButton } from './ExportButton';

interface ChartContainerProps {
  title: string;
  children: ReactNode;
  chartRef: React.RefObject<HTMLDivElement>;
  fullscreenRef: React.RefObject<HTMLDivElement>;
  isFullscreen: boolean;
  onToggleFullscreen: () => void;
  onResetView?: () => void;
  onZoomIn?: () => void;
  onZoomOut?: () => void;
  exportFileName: string;
  additionalControls?: ReactNode;
}

export const ChartContainer: React.FC<ChartContainerProps> = ({
  title,
  children,
  chartRef,
  fullscreenRef,
  isFullscreen,
  onToggleFullscreen,
  onResetView,
  onZoomIn,
  onZoomOut,
  exportFileName,
  additionalControls
}) => {
  return (
    <div ref={fullscreenRef} className={`w-full h-full ${isFullscreen ? 'fixed inset-0 z-50 bg-white dark:bg-gray-900 p-4' : ''}`}>
      <div className="w-full h-full" ref={chartRef}>
        {/* Header with title and controls */}
        <div className="flex items-center justify-between mb-4">
          <h4 className="text-md font-medium text-gray-700 dark:text-gray-300">
            {title}
          </h4>
          <div className="flex items-center gap-2">
            {additionalControls}
            <PlotControls 
              onResetView={onResetView || (() => {})}
              onToggleFullscreen={onToggleFullscreen}
              isFullscreen={isFullscreen}
              onZoomIn={onZoomIn}
              onZoomOut={onZoomOut}
            />
            <ExportButton 
              chartRef={chartRef} 
              fileName={exportFileName}
            />
          </div>
        </div>
        
        {/* Chart content */}
        <div style={{ height: isFullscreen ? 'calc(100vh - 80px)' : 'calc(100% - 40px)' }}>
          {children}
        </div>
      </div>
    </div>
  );
};