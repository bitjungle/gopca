import React from 'react';
import { usePalette } from '../contexts/PaletteContext';
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
          <select
            value={mode === 'continuous' ? sequentialPalette : qualitativePalette}
            onChange={mode === 'continuous' ? handleSequentialChange : handleQualitativeChange}
            className="px-2 py-1 bg-gray-100 dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded text-sm text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            {mode === 'categorical' && (
              <>
                <option value="deep">{paletteDisplayNames.deep}</option>
                <option value="pastel">{paletteDisplayNames.pastel}</option>
                <option value="dark">{paletteDisplayNames.dark}</option>
                <option value="colorblind">{paletteDisplayNames.colorblind}</option>
                <option value="husl">{paletteDisplayNames.husl}</option>
              </>
            )}
            {mode === 'continuous' && (
              <>
                <option value="rocket">{paletteDisplayNames.rocket}</option>
                <option value="viridis">{paletteDisplayNames.viridis}</option>
                <option value="blues">{paletteDisplayNames.blues}</option>
                <option value="reds">{paletteDisplayNames.reds}</option>
                <option value="crest">{paletteDisplayNames.crest}</option>
                <option value="mako">{paletteDisplayNames.mako}</option>
                <option value="flare">{paletteDisplayNames.flare}</option>
              </>
            )}
          </select>
        </div>
      )}
    </div>
  );
};