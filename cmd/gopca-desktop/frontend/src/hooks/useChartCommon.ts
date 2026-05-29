// GoPCA Suite
//
// Copyright © 2025-2026 Rune Mathisen <devel@bitjungle.com>
//
// This file is part of GoPCA Suite.
//
// GoPCA Suite is source-available software with free binary redistribution.
// Official compiled binary releases may be used and redistributed free of charge
// under the GoPCA Suite Source-Available Freeware License.
//
// The source code is provided for viewing, review, education, security analysis,
// research, interoperability analysis, and evaluation only.
//
// Modification, redistribution, publication, sublicensing, reuse, incorporation
// into another project, or creation of derivative works based on the source code
// is not permitted without prior written permission from the copyright holder.
//
// Usage Restriction: GoPCA Suite may not be used, directly or indirectly, for
// military, warfare, weapons, intelligence, surveillance, targeting, or
// law-enforcement surveillance applications.
//
// See LICENSE for the full license terms.

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