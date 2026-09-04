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

interface HelpDisplayProps {
  helpKey: string | null;
  title: string;
  text: string;
}

export const HelpDisplay: React.FC<HelpDisplayProps> = ({ helpKey, title, text }) => {
  if (!helpKey) {
    return (
      <div
        role="status"
        className="h-10 flex items-center justify-center text-gray-500 dark:text-gray-400"
      >
        <svg
          aria-hidden="true"
          className="w-5 h-5 mr-2"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M8.228 9c.549-1.165 2.03-2 3.772-2 2.21 0 4 1.343 4 3 0 1.4-1.278 2.575-3.006 2.907-.542.104-.994.54-.994 1.093m0 3h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
          />
        </svg>
        <span className="text-sm">Hover over any element for help</span>
      </div>
    );
  }

  return (
    // The width cap, not the surrounding layout, is what decides how much help
    // text is visible: the header gives this element more room than 2xl already,
    // so widening the header does nothing until this changes. 4xl fits the median
    // entry on one line and the 90th percentile comfortably on two.
    <div role="status" className="h-10 flex items-center justify-center max-w-4xl mx-auto">
      {/*
        Title and text flow as one paragraph rather than sitting in two flex
        items. As siblings in a flex row the title was a shrinkable item, so a
        title of three or four words was broken across three lines while the text
        beside it kept its own two-line clamp and the pair overflowed the fixed
        h-10 box. Inline, the title only wraps if it alone exceeds a line, which
        at a longest observed 28 characters it never does, and the clamp now
        governs the whole message instead of the text alone.
      */}
      <p className="text-sm text-center line-clamp-2">
        <span className="font-semibold text-gray-900 dark:text-gray-100">{title}:</span>{' '}
        <span className="text-gray-600 dark:text-gray-300">{text}</span>
      </p>
    </div>
  );
};
