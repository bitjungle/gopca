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

import React from 'react';
import { usePalette } from '../contexts/PaletteContext';
import { CustomSelect } from '@gopca/ui-components';
import {
  QualitativePaletteName,
  SequentialPaletteName
} from '../utils/colorPalettes';

export const PaletteSelector: React.FC = () => {
  const {
    mode,
    qualitativePalette,
    setQualitativePalette,
    sequentialPalette,
    setSequentialPalette
  } = usePalette();

  // Get display names for palettes
  const paletteDisplayNames = {
    // Qualitative
    deep: 'Deep',
    pastel: 'Pastel',
    dark: 'Dark',
    colorblind: 'Colorblind Safe',
    husl: 'HUSL',
    // Sequential
    rocket: 'Rocket',
    viridis: 'Viridis',
    blues: 'Blues',
    reds: 'Reds',
    mako: 'Mako',
    flare: 'Flare'
  };

  // Only render if palette control should be visible
  if (mode === 'none') {
    return null;
  }

  return (
    <div className="flex items-center gap-2">
      <label className="text-sm text-gray-600 dark:text-gray-400">Palette:</label>
      <div className="relative">
        <CustomSelect
          value={mode === 'continuous' ? sequentialPalette : qualitativePalette}
          onChange={mode === 'continuous' ? (value) => setSequentialPalette(value as SequentialPaletteName) : (value) => setQualitativePalette(value as QualitativePaletteName)}
          options={
            mode === 'categorical' ? [
              { value: 'deep', label: paletteDisplayNames.deep },
              { value: 'pastel', label: paletteDisplayNames.pastel },
              { value: 'dark', label: paletteDisplayNames.dark },
              { value: 'colorblind', label: paletteDisplayNames.colorblind },
              { value: 'husl', label: paletteDisplayNames.husl }
            ] : [
              { value: 'rocket', label: paletteDisplayNames.rocket },
              { value: 'viridis', label: paletteDisplayNames.viridis },
              { value: 'blues', label: paletteDisplayNames.blues },
              { value: 'reds', label: paletteDisplayNames.reds },
              { value: 'mako', label: paletteDisplayNames.mako },
              { value: 'flare', label: paletteDisplayNames.flare }
            ]
          }
          className="min-w-[150px]"
        />
      </div>
    </div>
  );
};