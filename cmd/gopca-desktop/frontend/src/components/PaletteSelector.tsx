import React from 'react';
import { usePalette } from '../contexts/PaletteContext';
import { QUALITATIVE_PALETTE, SEQUENTIAL_PALETTE } from '../utils/colorPalettes';
import { Palette } from 'lucide-react';

export const PaletteSelector: React.FC = () => {
  const { paletteType, setPaletteType } = usePalette();

  const handleChange = (event: React.ChangeEvent<HTMLSelectElement>) => {
    setPaletteType(event.target.value as 'qualitative' | 'sequential');
  };

  // Show first 5 colors as preview
  const previewColors = paletteType === 'qualitative' 
    ? QUALITATIVE_PALETTE.slice(0, 5)
    : SEQUENTIAL_PALETTE.filter((_, i) => i % 2 === 0).slice(0, 5);

  return (
    <div className="flex items-center gap-2">
      <Palette className="w-4 h-4 text-gray-600 dark:text-gray-400" />
      <div className="relative">
        <select
          value={paletteType}
          onChange={handleChange}
          className="appearance-none bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded px-3 py-1 pr-8 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 cursor-pointer"
        >
          <option value="qualitative">Categorical</option>
          <option value="sequential">Sequential</option>
        </select>
        <div className="absolute right-2 top-1/2 transform -translate-y-1/2 pointer-events-none">
          <svg className="w-4 h-4 text-gray-600 dark:text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
          </svg>
        </div>
      </div>
      
      {/* Color preview */}
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
    </div>
  );
};