import React from 'react';
import { usePalette } from '../contexts/PaletteContext';
import { 
  QUALITATIVE_PALETTES, 
  SEQUENTIAL_PALETTES,
  QualitativePaletteName,
  SequentialPaletteName
} from '../utils/colorPalettes';
import { Palette } from 'lucide-react';

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
  
  // Get current palette colors for preview
  const getPreviewColors = () => {
    if (mode === 'continuous') {
      const palette = SEQUENTIAL_PALETTES[sequentialPalette];
      // Show evenly spaced colors from the gradient
      return [0, 2, 4, 6, 8].map(i => palette[i]);
    } else if (mode === 'categorical') {
      return QUALITATIVE_PALETTES[qualitativePalette].slice(0, 5);
    }
    return [];
  };
  
  const previewColors = getPreviewColors();
  
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
      <Palette className="w-4 h-4 text-gray-600 dark:text-gray-400" />
      
      {/* Mode indicator */}
      <div className="text-xs text-gray-500 dark:text-gray-400 font-medium">
        {mode === 'none' && 'No Color'}
        {mode === 'categorical' && '🏷️ Categorical'}
        {mode === 'continuous' && '📊 Continuous'}
      </div>
      
      {/* Palette selector */}
      {mode !== 'none' && (
        <div className="relative">
          <select
            value={mode === 'continuous' ? sequentialPalette : qualitativePalette}
            onChange={mode === 'continuous' ? handleSequentialChange : handleQualitativeChange}
            className={`
              appearance-none bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 
              rounded px-3 py-1 pr-8 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 
              cursor-pointer
            `}
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
          <div className="absolute right-2 top-1/2 transform -translate-y-1/2 pointer-events-none">
            <svg className="w-4 h-4 text-gray-600 dark:text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
            </svg>
          </div>
        </div>
      )}
      
      {/* Color preview */}
      {mode !== 'none' && previewColors.length > 0 && (
        <div className="flex gap-1 ml-2">
          {previewColors.map((color, index) => (
            <div
              key={index}
              className="w-3 h-3 rounded-sm border border-gray-300 dark:border-gray-600"
              style={{ backgroundColor: color }}
              title={`Preview color ${index + 1}`}
            />
          ))}
        </div>
      )}
    </div>
  );
};