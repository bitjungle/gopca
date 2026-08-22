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
import { applyColumnToggle } from './columnRange';

const numKey = (i: number) => i;
const strKey = (i: number) => `col${i}`;

describe('applyColumnToggle', () => {
    it('toggles a single column when there is no anchor', () => {
        const out = applyColumnToggle({ 0: true, 1: true, 2: true }, 1, false, null, numKey);
        expect(out).toEqual({ 0: true, 1: false, 2: true });
    });

    it('applies to the whole range when an anchor is set', () => {
        const start = { 0: true, 1: true, 2: true, 3: true, 4: true };
        const out = applyColumnToggle(start, 3, false, 1, numKey);
        expect(out).toEqual({ 0: true, 1: false, 2: false, 3: false, 4: true });
    });

    it('works when the range is dragged backwards', () => {
        const start = { 0: true, 1: true, 2: true, 3: true };
        expect(applyColumnToggle(start, 1, false, 3, numKey))
            .toEqual({ 0: true, 1: false, 2: false, 3: false });
    });

    it('re-including a range works the same way', () => {
        const start = { 0: false, 1: false, 2: false };
        expect(applyColumnToggle(start, 2, true, 0, numKey))
            .toEqual({ 0: true, 1: true, 2: true });
    });

    it('handles an anchor equal to the clicked column', () => {
        expect(applyColumnToggle({ 0: true, 1: true }, 1, false, 1, numKey))
            .toEqual({ 0: true, 1: false });
    });

    it('does not mutate the input', () => {
        const start = { 0: true, 1: true };
        applyColumnToggle(start, 0, false, null, numKey);
        expect(start).toEqual({ 0: true, 1: true });
    });

    it('supports string keys, as used by the small-dataset table', () => {
        const start = { col0: true, col1: true, col2: true };
        expect(applyColumnToggle(start, 2, false, 1, strKey))
            .toEqual({ col0: true, col1: false, col2: false });
    });

    // The case that motivated the feature: 26 channels of a water band, in one action.
    it('excludes a wide contiguous band in a single shift-click', () => {
        const sel: Record<number, boolean> = {};
        for (let i = 0; i < 700; i++) sel[i] = true;
        const out = applyColumnToggle(sel, 175, false, 150, numKey);
        const excluded = Object.keys(out).filter(k => !out[Number(k)]).map(Number);
        expect(excluded).toHaveLength(26);
        expect(excluded[0]).toBe(150);
        expect(excluded[25]).toBe(175);
    });
});
