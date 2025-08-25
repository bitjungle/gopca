// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

import React from 'react';
import { usePalette } from '../contexts/PaletteContext';
import { CustomSelect } from '@gopca/ui-components';
import {
  QUALITATIVE_PALETTES,
  SEQUENTIAL_PALETTES,
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

  const handleQualitativeChange = (event: React.ChangeEvent<HTMLSelectElement>) => {
    setQualitativePalette(event.target.value as QualitativePaletteName);
  };

  const handleSequentialChange = (event: React.ChangeEvent<HTMLSelectElement>) => {
    setSequentialPalette(event.target.value as SequentialPaletteName);
  };

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
    crest: 'Crest',
    mako: 'Mako',
    flare: 'Flare'
  };

  return (
    <div className="flex items-center gap-2">
      <label className="text-sm text-gray-600 dark:text-gray-400">Palette:</label>
      {mode !== 'none' && (
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
                { value: 'crest', label: paletteDisplayNames.crest },
                { value: 'mako', label: paletteDisplayNames.mako },
                { value: 'flare', label: paletteDisplayNames.flare }
              ]
            }
            className="min-w-[150px]"
          />
        </div>
      )}
    </div>
  );
};