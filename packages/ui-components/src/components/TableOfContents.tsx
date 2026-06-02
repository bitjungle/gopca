// GoPCA Suite
//
// Copyright © 2025-2026 Rune Mathisen <devel@bitjungle.com>
//
// This file is part of GoPCA Suite. See LICENSE for the full license terms.

import React from 'react';
import { TocEntry } from '../utils/tocUtils';

export interface TableOfContentsProps {
    entries: TocEntry[];
    activeId: string | null;
    onEntryClick: (id: string) => void;
}

/**
 * Sticky sidebar table of contents for the DocumentationViewer.
 * Renders H2 and H3 entries; H3s are indented. The active section
 * (tracked by an IntersectionObserver in the parent) is highlighted
 * with a blue left-border accent.
 *
 * Returns null when there are fewer than 2 entries.
 */
export const TableOfContents: React.FC<TableOfContentsProps> = ({
    entries,
    activeId,
    onEntryClick,
}) => {
    if (entries.length < 2) return null;

    return (
        <aside
            className="w-56 shrink-0 overflow-y-auto border-r border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800/40"
            aria-label="Table of contents"
        >
            <div className="py-6 px-3">
                <p className="mb-3 px-3 text-xs font-semibold uppercase tracking-wider text-gray-400 dark:text-gray-500 select-none">
                    Contents
                </p>
                <nav>
                    <ul className="space-y-0.5">
                        {entries.map((entry) => {
                            const isActive = entry.id === activeId;
                            const isH3 = entry.level === 3;

                            return (
                                <li key={entry.id}>
                                    <button
                                        type="button"
                                        onClick={() => onEntryClick(entry.id)}
                                        className={[
                                            'w-full text-left leading-snug rounded-r transition-colors',
                                            'border-l-2 focus:outline-none focus-visible:ring-2 focus-visible:ring-blue-500',
                                            isH3
                                                ? 'py-1 pr-2 text-xs'
                                                : 'py-1.5 pr-2 text-sm',
                                            isH3 ? 'pl-6' : 'pl-3',
                                            isActive
                                                ? 'border-blue-500 text-blue-600 dark:text-blue-400 font-medium'
                                                : 'border-transparent text-gray-500 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-100 hover:border-gray-300 dark:hover:border-gray-500',
                                        ].join(' ')}
                                        aria-current={isActive ? 'location' : undefined}
                                    >
                                        {entry.text}
                                    </button>
                                </li>
                            );
                        })}
                    </ul>
                </nav>
            </div>
        </aside>
    );
};
