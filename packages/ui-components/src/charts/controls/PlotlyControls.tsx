// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

import React, { useState } from 'react';
import ReactDOM from 'react-dom';

interface PlotlyControlsProps {
  onToggleFullscreen: () => void;
  isFullscreen: boolean;
  className?: string;
}

interface TooltipState {
  show: boolean;
  text: string;
  x: number;
  y: number;
}

/**
 * Control buttons for Plotly visualizations
 * Provides fullscreen toggle and can be extended with additional controls
 */
export const PlotlyControls: React.FC<PlotlyControlsProps> = ({ 
  onToggleFullscreen,
  isFullscreen,
  className = '' 
}) => {
  const [tooltip, setTooltip] = useState<TooltipState>({ show: false, text: '', x: 0, y: 0 });

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
    <>
      <div className={`flex items-center gap-2 ${className}`}>
        <button
          onClick={onToggleFullscreen}
          className="px-3 py-1 text-sm rounded-lg bg-gray-200 dark:bg-gray-700 hover:bg-gray-300 dark:hover:bg-gray-600 text-gray-700 dark:text-gray-300 transition-colors flex items-center gap-2"
          onMouseEnter={(e) => handleMouseEnter(e, isFullscreen ? 'Exit fullscreen' : 'Enter fullscreen')}
          onMouseLeave={handleMouseLeave}
          aria-label={isFullscreen ? 'Exit fullscreen' : 'Enter fullscreen'}
        >
          {isFullscreen ? (
            // Exit fullscreen icon
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} 
                d="M9 9V4.5M9 9H4.5M9 9L3.75 3.75M9 15v4.5M9 15H4.5M9 15l-5.25 5.25M15 9h4.5M15 9V4.5M15 9l5.25-5.25M15 15h4.5M15 15v4.5m0-4.5l5.25 5.25" 
              />
            </svg>
          ) : (
            // Enter fullscreen icon
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} 
                d="M4 8V4m0 0h4M4 4l5 5m11-5h-4m4 0v4m0 0l-5-5M4 16v4m0 0h4m-4 0l5-5m11 5l-5-5m5 5v-4m0 4h-4" 
              />
            </svg>
          )}
          {isFullscreen ? 'Exit' : 'Fullscreen'}
        </button>
      </div>

      {/* Tooltip portal */}
      {tooltip.show && ReactDOM.createPortal(
        <div
          className="fixed z-[100000] px-2 py-1 text-xs text-white bg-gray-800 dark:bg-gray-900 rounded pointer-events-none whitespace-nowrap transform -translate-x-1/2 -translate-y-full"
          style={{
            left: `${tooltip.x}px`,
            top: `${tooltip.y}px`,
          }}
        >
          {tooltip.text}
        </div>,
        document.body
      )}
    </>
  );
};