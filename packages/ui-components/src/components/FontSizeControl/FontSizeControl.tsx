// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

import React from 'react';

interface FontSizeControlProps {
  value: number; // Scale factor (0.7 to 1.5)
  onChange: (value: number) => void;
}

export const FontSizeControl: React.FC<FontSizeControlProps> = ({ value, onChange }) => {
  const percentage = Math.round(value * 100);
  
  const handleSliderChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const newValue = parseFloat(event.target.value);
    onChange(newValue);
  };

  const handleReset = () => {
    onChange(1.0);
  };

  return (
    <div className="flex items-center gap-3">
      <label className="text-sm text-gray-600 dark:text-gray-400 whitespace-nowrap">
        Font Size:
      </label>
      <div className="flex items-center gap-2">
        <input
          type="range"
          min="0.7"
          max="1.5"
          step="0.05"
          value={value}
          onChange={handleSliderChange}
          className="w-24 accent-blue-500 cursor-pointer"
          title={`Font size: ${percentage}%`}
          aria-label="Font size adjustment"
        />
        <span className="text-sm text-gray-700 dark:text-gray-300 min-w-[45px] text-right">
          {percentage}%
        </span>
        {value !== 1.0 && (
          <button
            onClick={handleReset}
            className="text-xs text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200 px-1"
            title="Reset to 100%"
            aria-label="Reset font size to 100%"
          >
            Reset
          </button>
        )}
      </div>
    </div>
  );
};