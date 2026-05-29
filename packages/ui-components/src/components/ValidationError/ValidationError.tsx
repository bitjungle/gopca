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