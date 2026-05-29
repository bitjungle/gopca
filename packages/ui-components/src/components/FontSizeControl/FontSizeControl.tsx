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