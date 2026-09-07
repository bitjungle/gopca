// GoPCA Suite
//
// Copyright © 2025-2026 Rune Mathisen <devel@bitjungle.com>
//
// This file is part of GoPCA Suite.
//
// See LICENSE for the full license terms.

import { describe, it, expect } from 'vitest';
import { asPercentages, asFractions } from './variancePercent';
import type { PCAResponse, PCAResult } from '../types';

// The backend reports fractions as of V2; the UI renders percentages. This
// module is the only place that bridges the two, so it is the only place a
// factor-of-100 error can enter — and a factor of 100 in a variance figure is
// the kind of wrong that gets believed rather than questioned.
//
// The fixture is a real PCAResult, not a partial object cast into place. A cast
// would compile whatever it was given, so it would keep compiling if PCAResult
// gained a required field the production code then read off these values —
// which is the one thing a type-checked fixture is here to catch. It is also
// what broke the build: the earlier `as never` made the helper's return
// unusable, since reading any property off `never` is an error.
const result = (ratio: number[], cumulative: number[]): PCAResult => ({
    scores: [[1, 2]],
    loadings: [[0.7, 0.7]],
    explained_variance: ratio.map(v => v * 10),
    explained_variance_ratio: ratio,
    cumulative_variance: cumulative,
    component_labels: ratio.map((_, i) => `PC${i + 1}`),
    components_computed: ratio.length,
    method: 'svd',
    preprocessing_applied: true
});

// Narrowed to a present result so the fixture needs no non-null assertion.
// asPercentages still returns a plain PCAResponse, whose result is genuinely
// optional, so reading through its return does.
const response = (
    ratio: number[],
    cumulative: number[]
): PCAResponse & { result: PCAResult } => ({
    success: true,
    result: result(ratio, cumulative)
});

describe('variancePercent', () => {
    it('scales the iris profile to the percentages the UI labels expect', () => {
        const out = asPercentages(response([0.729624, 0.228508], [0.729624, 0.958132]));
        expect(out.result!.explained_variance_ratio[0]).toBeCloseTo(72.9624, 4);
        expect(out.result!.cumulative_variance[1]).toBeCloseTo(95.8132, 4);
    });

    it('leaves everything else on the result untouched', () => {
        const out = asPercentages(response([0.5], [0.5]));
        expect(out.result!.scores).toEqual([[1, 2]]);
        expect(out.result!.loadings).toEqual([[0.7, 0.7]]);
        expect(out.result!.explained_variance).toEqual([5]);
        expect(out.result!.method).toBe('svd');
    });

    it('does not mutate the response it was given', () => {
        const original = response([0.5], [0.5]);
        const before = [...original.result.explained_variance_ratio];
        asPercentages(original);
        expect(original.result.explained_variance_ratio).toEqual(before);
    });

    // A failed run carries no result, and the export path can be reached before
    // one exists. Neither may throw.
    it('passes a resultless response through', () => {
        const failed: PCAResponse = { success: false, error: 'boom' };
        expect(asPercentages(failed)).toBe(failed);
    });

    // The export path sends the result back to Go, which writes the model file
    // with the same code the CLI uses. If this did not invert, the Desktop would
    // write percentages where the CLI writes fractions.
    it('round-trips back to fractions for export', () => {
        const fractions = [0.729624, 0.228508, 0.036689];
        const cumulative = [0.729624, 0.958132, 0.994821];
        const shown = asPercentages(response(fractions, cumulative));
        const exported = asFractions(shown.result!);

        exported.explained_variance_ratio.forEach((v: number, i: number) => {
            expect(v).toBeCloseTo(fractions[i], 12);
        });
        exported.cumulative_variance.forEach((v: number, i: number) => {
            expect(v).toBeCloseTo(cumulative[i], 12);
        });
    });

    // asFractions is generic over anything carrying the two fields, so the
    // extra one rides along with its type intact and needs no cast.
    it('inverts exactly the two fields it claims to, and no others', () => {
        const out = asFractions({
            explained_variance_ratio: [50],
            cumulative_variance: [50],
            rmsec: 12.5
        });
        expect(out.explained_variance_ratio[0]).toBeCloseTo(0.5, 12);
        expect(out.rmsec).toBe(12.5);
    });
});
