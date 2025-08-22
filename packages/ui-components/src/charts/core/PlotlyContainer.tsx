// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

import React, { useState, useRef, useEffect, useCallback } from 'react';
import { PlotlyControls } from '../controls/PlotlyControls';

interface PlotlyContainerProps {
  children: React.ReactElement;
  enableFullscreen?: boolean;
  className?: string;
  onResize?: () => void;
}

/**
 * Container component for Plotly visualizations with fullscreen support
 * Manages fullscreen state and provides resize notifications to child components
 */
export const PlotlyContainer: React.FC<PlotlyContainerProps> = ({
  children,
  enableFullscreen = true,
  className = '',
  onResize
}) => {
  const [isFullscreen, setIsFullscreen] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);
  const resizeTimeoutRef = useRef<ReturnType<typeof setTimeout>>();

  // Handle fullscreen toggle
  const handleToggleFullscreen = useCallback(async () => {
    if (!containerRef.current) return;

    try {
      if (!document.fullscreenElement) {
        await containerRef.current.requestFullscreen();
      } else {
        await document.exitFullscreen();
      }
    } catch (error) {
      console.error('Fullscreen API error:', error);
    }
  }, []);

  // Handle fullscreen change events
  useEffect(() => {
    const handleFullscreenChange = () => {
      const isNowFullscreen = !!document.fullscreenElement;
      setIsFullscreen(isNowFullscreen);

      // Trigger resize after fullscreen change
      if (resizeTimeoutRef.current) {
        clearTimeout(resizeTimeoutRef.current);
      }

      resizeTimeoutRef.current = setTimeout(() => {
        // Trigger Plotly resize
        if (containerRef.current) {
          const plotlyDiv = containerRef.current.querySelector('.js-plotly-plot') as any;
          if (plotlyDiv && window.Plotly) {
            window.Plotly.Plots.resize(plotlyDiv);
          }
        }

        // Call optional resize handler
        if (onResize) {
          onResize();
        }
      }, 100); // Small delay to ensure DOM has updated
    };

    document.addEventListener('fullscreenchange', handleFullscreenChange);
    document.addEventListener('webkitfullscreenchange', handleFullscreenChange);
    document.addEventListener('mozfullscreenchange', handleFullscreenChange);
    document.addEventListener('MSFullscreenChange', handleFullscreenChange);

    return () => {
      document.removeEventListener('fullscreenchange', handleFullscreenChange);
      document.removeEventListener('webkitfullscreenchange', handleFullscreenChange);
      document.removeEventListener('mozfullscreenchange', handleFullscreenChange);
      document.removeEventListener('MSFullscreenChange', handleFullscreenChange);

      if (resizeTimeoutRef.current) {
        clearTimeout(resizeTimeoutRef.current);
      }
    };
  }, [onResize]);

  // Handle window resize in fullscreen mode
  useEffect(() => {
    if (!isFullscreen) return;

    const handleResize = () => {
      if (resizeTimeoutRef.current) {
        clearTimeout(resizeTimeoutRef.current);
      }

      resizeTimeoutRef.current = setTimeout(() => {
        if (containerRef.current) {
          const plotlyDiv = containerRef.current.querySelector('.js-plotly-plot') as any;
          if (plotlyDiv && window.Plotly) {
            window.Plotly.Plots.resize(plotlyDiv);
          }
        }

        if (onResize) {
          onResize();
        }
      }, 250); // Debounce resize events
    };

    window.addEventListener('resize', handleResize);
    return () => {
      window.removeEventListener('resize', handleResize);
    };
  }, [isFullscreen, onResize]);

  const containerClasses = [
    className,
    isFullscreen ? 'fixed inset-0 z-[9999] bg-white dark:bg-gray-900' : 'relative w-full h-full'
  ].filter(Boolean).join(' ');

  const contentClasses = isFullscreen 
    ? 'w-full h-full p-4' 
    : 'w-full h-full';

  return (
    <div ref={containerRef} className={containerClasses}>
      {enableFullscreen && (
        <div className={`absolute top-2 right-2 z-10 ${isFullscreen ? '' : 'opacity-75 hover:opacity-100'} transition-opacity`}>
          <PlotlyControls
            onToggleFullscreen={handleToggleFullscreen}
            isFullscreen={isFullscreen}
          />
        </div>
      )}
      <div className={contentClasses}>
        {children}
      </div>
    </div>
  );
};

// Type declaration for Plotly global
declare global {
  interface Window {
    Plotly: any;
  }
}