// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

import React, { useState } from 'react';
import ReactDOM from 'react-dom';
import { useChartTheme } from '../hooks/useChartTheme';

interface PlotControlsProps {
  onResetView: () => void;
  onToggleFullscreen: () => void;
  onZoomIn?: () => void;
  onZoomOut?: () => void;
  isFullscreen: boolean;
  className?: string;
}

interface TooltipState {
  show: boolean;
  text: string;
  x: number;
  y: number;
}

export const PlotControls: React.FC<PlotControlsProps> = ({
  onResetView,
  onToggleFullscreen,
  onZoomIn,
  onZoomOut,
  isFullscreen,
  className = ''
}) => {
  const [tooltip, setTooltip] = useState<TooltipState>({ show: false, text: '', x: 0, y: 0 });
  const chartTheme = useChartTheme();

  const handleMouseEnter = (e: React.MouseEvent<HTMLButtonElement>, text: string) => {
    const rect = e.currentTarget.getBoundingClientRect();
    setTooltip({
      show: true,
      text,
      x: rect.left + rect.width / 2,
      y: rect.top - 10
    });
  };

  const handleMouseLeave = () => {
    setTooltip({ show: false, text: '', x: 0, y: 0 });
  };

  return (
    <div className={`flex items-center gap-2 ${className}`}>
      {onZoomIn && onZoomOut && (
        <>
          <button
            onClick={onZoomIn}
            className="px-2 py-1 text-sm rounded-lg bg-gray-200 dark:bg-gray-700 hover:bg-gray-300 dark:hover:bg-gray-600 text-gray-700 dark:text-gray-300 transition-colors"
            onMouseEnter={(e) => handleMouseEnter(e, 'Zoom in')}
            onMouseLeave={handleMouseLeave}
          >
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
                d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0zM10 7v6m3-3H7"
              />
            </svg>
          </button>

          <button
            onClick={onZoomOut}
            className="px-2 py-1 text-sm rounded-lg bg-gray-200 dark:bg-gray-700 hover:bg-gray-300 dark:hover:bg-gray-600 text-gray-700 dark:text-gray-300 transition-colors"
            onMouseEnter={(e) => handleMouseEnter(e, 'Zoom out')}
            onMouseLeave={handleMouseLeave}
          >
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
                d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0zM13 10H7"
              />
            </svg>
          </button>
        </>
      )}

      <button
        onClick={onResetView}
        className="px-3 py-1 text-sm rounded-lg bg-gray-200 dark:bg-gray-700 hover:bg-gray-300 dark:hover:bg-gray-600 text-gray-700 dark:text-gray-300 transition-colors flex items-center gap-2"
        onMouseEnter={(e) => handleMouseEnter(e, 'Reset zoom')}
        onMouseLeave={handleMouseLeave}
      >
        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
            d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
          />
        </svg>
        Reset
      </button>

      <button
        onClick={onToggleFullscreen}
        className="px-3 py-1 text-sm rounded-lg bg-gray-200 dark:bg-gray-700 hover:bg-gray-300 dark:hover:bg-gray-600 text-gray-700 dark:text-gray-300 transition-colors flex items-center gap-2"
        onMouseEnter={(e) => handleMouseEnter(e, isFullscreen ? 'Exit fullscreen' : 'Enter fullscreen')}
        onMouseLeave={handleMouseLeave}
      >
        {isFullscreen ? (
          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
              d="M6 18L18 6M6 6l12 12"
            />
          </svg>
        ) : (
          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
              d="M4 8V4m0 0h4M4 4l5 5m11-1V4m0 0h-4m4 0l-5 5M4 16v4m0 0h4m-4 0l5-5m11 5l-5-5m5 5v-4m0 4h-4"
            />
          </svg>
        )}
        {isFullscreen ? 'Exit' : 'Fullscreen'}
      </button>

      {/* Tooltip Portal */}
      {tooltip.show && ReactDOM.createPortal(
        <div
          className="fixed z-50 px-2 py-1 text-xs rounded shadow-lg border pointer-events-none"
          style={{
            backgroundColor: chartTheme.tooltipBackgroundColor,
            borderColor: chartTheme.tooltipBorderColor,
            color: chartTheme.tooltipTextColor,
            left: tooltip.x,
            top: tooltip.y - 30,
            transform: 'translateX(-50%)'
          }}
        >
          {tooltip.text}
        </div>,
        document.body
      )}
    </div>
  );
};