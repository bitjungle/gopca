// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

import React, { useState, useEffect, useRef, useCallback } from 'react';
import ReactDOM from 'react-dom';
import Plot from 'react-plotly.js';

/**
 * Fullscreen overlay container for Plotly visualizations
 * Fills the entire GoPCA application window
 */
export const PlotlyFullscreenModal: React.FC<{
  isOpen: boolean;
  onClose: () => void;
  plotData: any;
  plotLayout: any;
  plotConfig: any;
}> = ({ isOpen, onClose, plotData, plotLayout, plotConfig }) => {
  const modalRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && isOpen) {
        onClose();
      }
    };

    if (isOpen) {
      document.addEventListener('keydown', handleEscape);
      document.body.style.overflow = 'hidden';
    }

    return () => {
      document.removeEventListener('keydown', handleEscape);
      document.body.style.overflow = '';
    };
  }, [isOpen, onClose]);

  // Force Plotly to resize when modal opens
  useEffect(() => {
    if (isOpen && modalRef.current) {
      const plotDiv = modalRef.current.querySelector('.js-plotly-plot') as any;
      if (plotDiv && window.Plotly) {
        setTimeout(() => {
          window.Plotly.Plots.resize(plotDiv);
        }, 100);
      }
    }
  }, [isOpen]);

  if (!isOpen) return null;

  // Enhanced config with exit fullscreen button
  // Ensure modebar is visible and add the fullscreen toggle button
  const enhancedConfig = {
    ...plotConfig,
    displayModeBar: true, // Ensure modebar is visible
    displaylogo: false,
    modeBarButtonsToAdd: [
      ...(plotConfig.modeBarButtonsToAdd || []).filter(
        (btn: any) => btn.name !== 'fullscreen' // Remove existing fullscreen button
      ),
      createFullscreenButton(onClose) // Add toggle button that exits fullscreen
    ]
  };

  return ReactDOM.createPortal(
    <div className="fixed inset-0 z-[999999] bg-white dark:bg-gray-900" style={{ zIndex: 999999 }}>
      {/* Plot container with padding to ensure modebar is visible */}
      <div ref={modalRef} className="w-full h-full relative pt-12">
        <Plot
          data={plotData}
          layout={{
            ...plotLayout,
            autosize: true,
          }}
          config={enhancedConfig}
          style={{ width: '100%', height: 'calc(100% - 3rem)' }}
          useResizeHandler={true}
        />
        
        {/* Exit button overlaid on top-right */}
        <button
          onClick={onClose}
          className="absolute top-4 right-4 p-2 rounded-lg bg-gray-200 dark:bg-gray-700 hover:bg-gray-300 dark:hover:bg-gray-600 transition-colors z-10"
          aria-label="Exit fullscreen"
          title="Exit fullscreen (ESC)"
        >
          <svg className="w-5 h-5 text-gray-700 dark:text-gray-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} 
              d="M9 9V4.5M9 9H4.5M9 9L3.75 3.75M9 15v4.5M9 15H4.5M9 15l-5.25 5.25M15 9h4.5M15 9V4.5M15 9l5.25-5.25M15 15h4.5M15 15v4.5m0-4.5l5.25 5.25" 
            />
          </svg>
        </button>
      </div>
    </div>,
    document.body
  );
};

/**
 * Hook to add fullscreen functionality to Plotly plots
 */
export const usePlotlyFullscreen = (data: any, layout: any, config: any) => {
  const [isFullscreen, setIsFullscreen] = useState(false);

  const openFullscreen = useCallback(() => {
    setIsFullscreen(true);
  }, []);

  const closeFullscreen = useCallback(() => {
    setIsFullscreen(false);
  }, []);

  const fullscreenModal = (
    <PlotlyFullscreenModal
      isOpen={isFullscreen}
      onClose={closeFullscreen}
      plotData={data}
      plotLayout={layout}
      plotConfig={config}
    />
  );

  return {
    isFullscreen,
    openFullscreen,
    closeFullscreen,
    fullscreenModal
  };
};

/**
 * Create a custom fullscreen button for Plotly's modebar
 */
export const createFullscreenButton = (onClick: () => void) => {
  return {
    name: 'fullscreen',
    title: 'Toggle fullscreen',
    icon: {
      width: 1000,
      height: 1000,
      path: 'M250 250 L250 500 L500 500 M750 250 L750 500 L500 500 M250 750 L250 500 L500 500 M750 750 L750 500 L500 500',
      transform: 'matrix(1 0 0 1 0 0)'
    },
    click: onClick
  };
};

/**
 * Wrapper component that adds fullscreen capability to Plotly plots
 */
export const PlotlyWithFullscreen: React.FC<{
  data: any;
  layout: any;
  config: any;
  style?: React.CSSProperties;
  onSelected?: (event: any) => void;
}> = ({ data, layout, config, style, onSelected }) => {
  const { openFullscreen, fullscreenModal } = usePlotlyFullscreen(data, layout, config);

  // Add fullscreen button to modebar
  const enhancedConfig = {
    ...config,
    modeBarButtonsToAdd: [
      ...(config.modeBarButtonsToAdd || []),
      createFullscreenButton(openFullscreen)
    ]
  };

  return (
    <>
      <Plot
        data={data}
        layout={layout}
        config={enhancedConfig}
        style={style || { width: '100%', height: '100%' }}
        useResizeHandler={true}
        onSelected={onSelected}
      />
      {fullscreenModal}
    </>
  );
};

// Type declaration for Plotly global
declare global {
  interface Window {
    Plotly: any;
  }
}