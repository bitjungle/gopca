// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

import React from 'react';

export const MatrixIllustration: React.FC = () => {
  return (
    <div className="flex flex-col items-center justify-center h-full">
      <svg
        width="380"
        height="200"
        viewBox="0 0 380 200"
        className="w-full max-w-[380px] h-auto"
      >
        {/* Background grid */}
        <rect x="70" y="40" width="250" height="120" fill="none" stroke="#e5e7eb" strokeWidth="1" className="dark:stroke-gray-600"/>

        {/* Column headers */}
        <text x="115" y="30" textAnchor="middle" className="text-xs fill-gray-600 dark:fill-gray-400 font-medium">Variable 1</text>
        <text x="185" y="30" textAnchor="middle" className="text-xs fill-gray-600 dark:fill-gray-400 font-medium">Variable 2</text>
        <text x="255" y="30" textAnchor="middle" className="text-xs fill-gray-600 dark:fill-gray-400 font-medium">Variable 3</text>
        <text x="295" y="30" textAnchor="middle" className="text-xs fill-gray-500 dark:fill-gray-500">...</text>

        {/* Row headers */}
        <text x="65" y="65" textAnchor="end" className="text-xs fill-gray-600 dark:fill-gray-400 font-medium">Sample 1</text>
        <text x="65" y="95" textAnchor="end" className="text-xs fill-gray-600 dark:fill-gray-400 font-medium">Sample 2</text>
        <text x="65" y="125" textAnchor="end" className="text-xs fill-gray-600 dark:fill-gray-400 font-medium">Sample 3</text>
        <text x="65" y="155" textAnchor="end" className="text-xs fill-gray-500 dark:fill-gray-500">...</text>

        {/* Grid lines */}
        <line x1="70" y1="70" x2="320" y2="70" stroke="#e5e7eb" strokeWidth="1" className="dark:stroke-gray-600"/>
        <line x1="70" y1="100" x2="320" y2="100" stroke="#e5e7eb" strokeWidth="1" className="dark:stroke-gray-600"/>
        <line x1="70" y1="130" x2="320" y2="130" stroke="#e5e7eb" strokeWidth="1" className="dark:stroke-gray-600"/>

        <line x1="150" y1="40" x2="150" y2="160" stroke="#e5e7eb" strokeWidth="1" className="dark:stroke-gray-600"/>
        <line x1="220" y1="40" x2="220" y2="160" stroke="#e5e7eb" strokeWidth="1" className="dark:stroke-gray-600"/>
        <line x1="290" y1="40" x2="290" y2="160" stroke="#e5e7eb" strokeWidth="1" className="dark:stroke-gray-600"/>

        {/* Data values */}
        <text x="115" y="60" textAnchor="middle" className="text-xs fill-gray-700 dark:fill-gray-300">5.1</text>
        <text x="185" y="60" textAnchor="middle" className="text-xs fill-gray-700 dark:fill-gray-300">3.5</text>
        <text x="255" y="60" textAnchor="middle" className="text-xs fill-gray-700 dark:fill-gray-300">1.4</text>

        <text x="115" y="90" textAnchor="middle" className="text-xs fill-gray-700 dark:fill-gray-300">4.9</text>
        <text x="185" y="90" textAnchor="middle" className="text-xs fill-gray-700 dark:fill-gray-300">3.0</text>
        <text x="255" y="90" textAnchor="middle" className="text-xs fill-gray-700 dark:fill-gray-300">1.4</text>

        <text x="115" y="120" textAnchor="middle" className="text-xs fill-gray-700 dark:fill-gray-300">4.7</text>
        <text x="185" y="120" textAnchor="middle" className="text-xs fill-gray-700 dark:fill-gray-300">3.2</text>
        <text x="255" y="120" textAnchor="middle" className="text-xs fill-gray-700 dark:fill-gray-300">1.3</text>

        {/* Arrow indicators */}
        <path d="M 330 100 L 340 100 L 335 95 M 340 100 L 335 105" stroke="#3b82f6" strokeWidth="2" fill="none" className="dark:stroke-blue-400"/>
        <text x="345" y="103" className="text-xs fill-blue-600 dark:fill-blue-400 font-medium">Rows</text>

        <path d="M 195 170 L 195 180 L 190 175 M 195 180 L 200 175" stroke="#3b82f6" strokeWidth="2" fill="none" className="dark:stroke-blue-400"/>
        <text x="195" y="195" textAnchor="middle" className="text-xs fill-blue-600 dark:fill-blue-400 font-medium">Columns</text>
      </svg>

      <div className="mt-4 text-center">
        <p className="text-sm text-gray-600 dark:text-gray-400">
          CSV format: first row contains variable names,<br/>
          first column contains sample names
        </p>
      </div>
    </div>
  );
};