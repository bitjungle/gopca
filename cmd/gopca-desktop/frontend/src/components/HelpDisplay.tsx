// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

import React from 'react';

interface HelpDisplayProps {
  helpKey: string | null;
  title: string;
  text: string;
}

export const HelpDisplay: React.FC<HelpDisplayProps> = ({ helpKey, title, text }) => {
  if (!helpKey) {
    return (
      <div className="h-10 flex items-center justify-center text-gray-500 dark:text-gray-400">
        <svg className="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8.228 9c.549-1.165 2.03-2 3.772-2 2.21 0 4 1.343 4 3 0 1.4-1.278 2.575-3.006 2.907-.542.104-.994.54-.994 1.093m0 3h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        <span className="text-sm">Hover over any element for help</span>
      </div>
    );
  }

  return (
    <div className="h-10 flex items-center justify-center max-w-2xl mx-auto animate-fadeIn">
      <div className="flex items-center gap-2 text-center">
        <span className="text-sm font-semibold text-gray-900 dark:text-gray-100">
          {title}:
        </span>
        <span className="text-sm text-gray-600 dark:text-gray-300 line-clamp-2">
          {text}
        </span>
      </div>
    </div>
  );
};