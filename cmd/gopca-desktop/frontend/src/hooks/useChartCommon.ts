// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

import { useRef, useState, useCallback } from 'react';
import { useChartTheme } from './useChartTheme';
import { usePalette } from '../contexts/PaletteContext';

export const useChartCommon = () => {
  const chartRef = useRef<HTMLDivElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  const fullscreenRef = useRef<HTMLDivElement>(null);

  const [isFullscreen, setIsFullscreen] = useState(false);
  const chartTheme = useChartTheme();
  const { mode, qualitativePalette, sequentialPalette } = usePalette();

  const handleToggleFullscreen = useCallback(() => {
    if (!fullscreenRef.current) {
return;
}

    if (!isFullscreen) {
      if (fullscreenRef.current.requestFullscreen) {
        fullscreenRef.current.requestFullscreen();
      }
    } else {
      if (document.exitFullscreen) {
        document.exitFullscreen();
      }
    }

    setIsFullscreen(!isFullscreen);
  }, [isFullscreen]);

  return {
    chartRef,
    containerRef,
    fullscreenRef,
    isFullscreen,
    setIsFullscreen,
    chartTheme,
    mode,
    qualitativePalette,
    sequentialPalette,
    handleToggleFullscreen
  };
};