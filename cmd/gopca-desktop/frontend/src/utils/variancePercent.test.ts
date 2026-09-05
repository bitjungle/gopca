// GoPCA Suite
//
// Copyright © 2025-2026 Rune Mathisen <devel@bitjungle.com>
//
// This file is part of GoPCA Suite.
//
// See LICENSE for the full license terms.

import { describe, it, expect } from 'vitest';
import { asPercentages, asFractions } from './variancePercent';

// The backend reports fractions as of V2; the UI renders percentages. This
// module is the only place that bridges the two, so it is the only place a
// factor-of-100 error can enter — and a factor of 100 in a variance figure is
// the kind of wrong that gets believed rather than questioned.
const response = (ratio: number[], cumulative: number[]) =>
    ({
        success: true,
        result: {
            explained_variance_ratio: ratio,
            cumulative_variance: cumulative,
            scores: [[1, 2]],
            loadings: [[0.7, 0.7]]
        }
    }) as never;

describe('variancePercent', () => {
    it('scales the iris profile to the percentages the UI labels expect', () => {
        const out = asPercentages(response([0.729624, 0.228508], [0.729624, 0.958132]));
        expect(out.result!.explained_variance_ratio[0]).toBeCloseTo(72.9624, 4);
        expect(out.result!.cumulative_variance[1]).toBeCloseTo(95.8132, 4);
    });

    it('leaves everything else on the result untouched', () => {
        const out = asPercentages(response([0.5], [0.5]));
        expect((out.result as { scores: number[][] }).scores).toEqual([[1, 2]]);
        expect((out.result as { loadings: number[][] }).loadings).toEqual([[0.7, 0.7]]);
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
        const failed = { success: false, error: 'boom' } as never;
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

    it('inverts exactly the two fields it claims to, and no others', () => {
        const out = asFractions({
            explained_variance_ratio: [50],
            cumulative_variance: [50],
            rmsec: 12.5
        } as never) as unknown as { explained_variance_ratio: number[]; rmsec: number };
        expect(out.explained_variance_ratio[0]).toBeCloseTo(0.5, 12);
        expect(out.rmsec).toBe(12.5);
    });
});
