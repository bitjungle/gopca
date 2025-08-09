// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

import React from 'react';

/**
 * Icon components for column type indicators.
 * Using Heroicons (https://heroicons.com/) for consistency with the rest of the application.
 * All icons use outline style with strokeWidth 1.5.
 */

// Target column icon - Flag icon to indicate this is the target/goal column for analysis
export const TargetColumnIcon: React.FC<{ className?: string }> = ({ className = "w-4 h-4 inline-block ml-1" }) => (
    <span title="Target column for analysis">
        <svg
            xmlns="http://www.w3.org/2000/svg"
            fill="none"
            viewBox="0 0 24 24"
            strokeWidth={1.5}
            stroke="currentColor"
            className={className}
            aria-label="Target column"
        >
            <path
                strokeLinecap="round"
                strokeLinejoin="round"
                d="M3 3v1.5M3 21v-6m0 0l2.77-.693a9 9 0 016.208.682l.108.054a9 9 0 006.086.71l3.114-.732a48.524 48.524 0 01-.005-10.499l-3.11.732a9 9 0 01-6.085-.711l-.108-.054a9 9 0 00-6.208-.682L3 4.5M3 15V4.5"
            />
        </svg>
    </span>
);

// Category column icon - Tag icon to indicate categorical/grouping columns
export const CategoryColumnIcon: React.FC<{ className?: string }> = ({ className = "w-4 h-4 inline-block ml-1" }) => (
    <span title="Categorical/grouping column">
        <svg
            xmlns="http://www.w3.org/2000/svg"
            fill="none"
            viewBox="0 0 24 24"
            strokeWidth={1.5}
            stroke="currentColor"
            className={className}
            aria-label="Category column"
        >
            <path
                strokeLinecap="round"
                strokeLinejoin="round"
                d="M9.568 3H5.25A2.25 2.25 0 003 5.25v4.318c0 .597.237 1.17.659 1.591l9.581 9.581c.699.699 1.78.872 2.607.33a18.095 18.095 0 005.223-5.223c.542-.827.369-1.908-.33-2.607L11.16 3.66A2.25 2.25 0 009.568 3z"
            />
            <path
                strokeLinecap="round"
                strokeLinejoin="round"
                d="M6 6h.008v.008H6V6z"
            />
        </svg>
    </span>
);

// For context menu usage - returns the icon without default styling
export const TargetColumnMenuIcon: React.FC = () => (
    <span className="inline-flex items-center">
        <TargetColumnIcon className="w-4 h-4 mr-2" />
    </span>
);

export const CategoryColumnMenuIcon: React.FC = () => (
    <span className="inline-flex items-center">
        <CategoryColumnIcon className="w-4 h-4 mr-2" />
    </span>
);