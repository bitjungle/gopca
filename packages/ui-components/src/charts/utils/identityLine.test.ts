// GoPCA Suite
//
// Copyright © 2025-2026 Rune Mathisen <devel@bitjungle.com>
//
// This file is part of GoPCA Suite.
//
// See LICENSE for the full license terms.

import { describe, it, expect } from 'vitest';
import { identityLineEnds } from './identityLine';

describe('identityLineEnds', () => {
    it('spans the finite extent of the data', () => {
        expect(identityLineEnds([
            { x: 2, y: 5 },
            { x: 9, y: 1 }
        ])).toEqual([1, 9]);
    });

    // The defect this guard exists for. Math.min() of an empty list is Infinity
    // and Math.max() is -Infinity, so the unguarded version handed Plotly a
    // trace running from Infinity to -Infinity rather than drawing nothing.
    it('returns null when there is no point to span', () => {
        expect(identityLineEnds([])).toBeNull();
    });

    it('returns null when every point is non-finite', () => {
        expect(identityLineEnds([
            { x: NaN, y: NaN },
            { x: Infinity, y: -Infinity }
        ])).toBeNull();
    });

    it('ignores non-finite points among finite ones', () => {
        expect(identityLineEnds([
            { x: 3, y: NaN },
            { x: NaN, y: 8 }
        ])).toEqual([3, 8]);
    });

    // An explicit domain is the caller stating what the axis shows, so it wins
    // over the data — including when there is no data at all, which is how a
    // fixed-range plot still gets its diagonal.
    it('prefers an explicit domain', () => {
        expect(identityLineEnds([{ x: 2, y: 5 }], [0, 10])).toEqual([0, 10]);
        expect(identityLineEnds([], [0, 10])).toEqual([0, 10]);
    });
});
