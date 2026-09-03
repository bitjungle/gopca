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

import { describe, it, expect } from 'vitest';
import helpContent from './help-content.json';

// Vite's own glob rather than node's fs: the frontend has no @types/node, so a
// filesystem walk here would run fine under vitest and then fail the tsc pass
// that `npm run build` does. Raw imports keep this test inside the bundler's
// world, which is the same world the app is built in.
const SOURCES = import.meta.glob('../**/*.{ts,tsx}', {
    query: '?raw',
    eager: true,
    import: 'default'
}) as Record<string, string>;

describe('help content', () => {
    const defined = new Set(Object.keys(helpContent.help));

    // A helpKey is a hand-copied string joining a component to help-content.json,
    // and nothing checks the two agree: a key with a typo renders an empty help
    // area rather than failing, so the header quietly stops keeping its promise
    // that hovering any element explains it.
    it('defines every key that a component asks for', () => {
        const missing: string[] = [];
        for (const [file, text] of Object.entries(SOURCES)) {
            if (file.includes('.test.')) continue;
            for (const m of text.matchAll(/helpKey=["']([a-zA-Z0-9_-]+)["']/g)) {
                if (!defined.has(m[1])) missing.push(`${m[1]} (${file})`);
            }
        }
        expect(missing).toEqual([]);
    });

    it('gives every entry a title and a text', () => {
        for (const [key, entry] of Object.entries(helpContent.help)) {
            expect(entry.title, `${key} has no title`).toBeTruthy();
            expect(entry.text, `${key} has no text`).toBeTruthy();
        }
    });

    // HelpDisplay renders into a fixed h-10 box with line-clamp-2, so a long
    // entry is silently cut off rather than wrapping. Where exactly the clamp
    // falls depends on the rendered font and has not been measured, so this is a
    // ratchet rather than a real limit: it holds the line at the longest entry we
    // already ship, so nothing new makes the problem worse.
    //
    // The box was widened from max-w-2xl to max-w-4xl, which roughly doubles what
    // two lines hold, so the entries near this bound are likelier to fit now than
    // when the number was first set. Lower it again whenever a long entry is
    // rewritten; the value is the longest title plus text currently shipped.
    const LONGEST_SHIPPED = 281;
    it('adds no entry longer than the longest one already shipped', () => {
        const tooLong = Object.entries(helpContent.help)
            .map(([key, e]) => [key, e.title.length + e.text.length] as const)
            .filter(([, len]) => len > LONGEST_SHIPPED)
            .map(([key, len]) => `${key} (${len})`);
        expect(tooLong).toEqual([]);
    });
});
