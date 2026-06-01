// GoPCA Suite
//
// Copyright © 2025-2026 Rune Mathisen <devel@bitjungle.com>
//
// This file is part of GoPCA Suite. See LICENSE for the full license terms.

import React from 'react';

/** A single entry in the table of contents. */
export interface TocEntry {
    level: 2 | 3;
    text: string;
    id: string;
}

/**
 * Converts a heading string to a URL-safe slug used as an element ID.
 * "What is PCA? A Step-by-Step Guide" → "what-is-pca-a-step-by-step-guide"
 */
export function toSlug(text: string): string {
    return text
        .toLowerCase()
        .replace(/[^\w\s-]/g, '')
        .replace(/\s+/g, '-')
        .replace(/-+/g, '-')
        .trim();
}

/**
 * Recursively extracts plain text from React children.
 * Handles bold/italic headings such as "## **Key** Concept".
 */
export function extractTextContent(children: React.ReactNode): string {
    if (typeof children === 'string') return children;
    if (typeof children === 'number') return String(children);
    if (Array.isArray(children)) return children.map(extractTextContent).join('');
    if (React.isValidElement(children)) {
        return extractTextContent(
            (children.props as { children?: React.ReactNode }).children
        );
    }
    return '';
}

/**
 * Parses a markdown string and returns H2 and H3 headings as TocEntry[].
 * Uses the same slug logic as the MarkdownRenderer heading overrides so that
 * IDs generated at parse time match the IDs rendered in the DOM.
 */
export function extractHeadings(markdown: string): TocEntry[] {
    const entries: TocEntry[] = [];
    for (const line of markdown.split('\n')) {
        const trimmed = line.trim();
        const h2 = /^## (.+)$/.exec(trimmed);
        const h3 = /^### (.+)$/.exec(trimmed);
        if (h2) {
            const text = h2[1].trim();
            entries.push({ level: 2, text, id: toSlug(text) });
        } else if (h3) {
            const text = h3[1].trim();
            entries.push({ level: 3, text, id: toSlug(text) });
        }
    }
    return entries;
}
