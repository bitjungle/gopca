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
    /** 1-based line in the markdown source; identifies the heading uniquely. */
    line: number;
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
 *
 * Duplicate heading texts are disambiguated with a numeric suffix so that
 * every entry has a unique ID — matching the scheme used in MarkdownRenderer.
 * Example: two "The Core Idea" headings → "the-core-idea", "the-core-idea-2".
 */
export function extractHeadings(markdown: string): TocEntry[] {
    const entries: TocEntry[] = [];
    const counts = new Map<string, number>();

    const lines = markdown.split('\n');
    for (let index = 0; index < lines.length; index++) {
        const trimmed = lines[index].trim();
        const h2 = /^## (.+)$/.exec(trimmed);
        const h3 = /^### (.+)$/.exec(trimmed);
        const match = h2 ?? h3;
        if (!match) continue;

        const level = h2 ? 2 : 3;
        const text = match[1].trim();
        const base = toSlug(text);
        const count = counts.get(base) ?? 0;
        counts.set(base, count + 1);
        const id = count === 0 ? base : `${base}-${count + 1}`;
        entries.push({ level: level as 2 | 3, text, id, line: index + 1 });
    }
    return entries;
}

/**
 * Maps each heading's source line to the id it must carry.
 *
 * The renderer used to number duplicate headings with a counter it mutated
 * while rendering. That is a side effect during render, and React StrictMode
 * deliberately invokes render twice to surface exactly this: on the second pass
 * every heading looked like a repeat of itself and took the "-2" id, while the
 * table of contents -- built in a single pass over the source -- kept asking for
 * the first. Every entry then pointed at nothing, and clicking one did nothing
 * at all (#436).
 *
 * Deriving the id from the heading's position in the source removes the state.
 * The same markdown yields the same ids however many times it is rendered, and
 * both sides get them from this one function rather than from two counters that
 * have to agree.
 */
export function headingIdsByLine(markdown: string): Map<number, string> {
    const byLine = new Map<number, string>();
    for (const entry of extractHeadings(markdown)) {
        byLine.set(entry.line, entry.id);
    }
    return byLine;
}
