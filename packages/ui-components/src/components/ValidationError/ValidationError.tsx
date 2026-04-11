// Copyright 2025-2026 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

import React from 'react';

interface ValidationErrorProps {
  message: string;
  fieldName?: string;
  className?: string;
}

export const ValidationError: React.FC<ValidationErrorProps> = ({
  message,
  fieldName,
  className = ''
}) => {
  return (
    <div
      className={`mt-1 text-sm text-red-600 dark:text-red-400 ${className}`}
      role="alert"
      aria-live="polite"
      aria-atomic="true"
    >
      <div className="flex items-center">
        <svg
          className="h-4 w-4 mr-1 flex-shrink-0"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          aria-hidden="true"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
          />
        </svg>
        <span>
          {fieldName && <span className="font-medium">{fieldName}: </span>}
          {message}
        </span>
      </div>
    </div>
  );
};