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
import {
    toRuns, describeRun, runToIndices, parseRangeSpec, columnMeans, namesFormOrderedAxis,
    columnExtents, profileFractions, sharedScaleIsReadable
} from './columnRanges';

// A 700-channel NIR axis, 1100–2498 nm at 2 nm steps, as in the Corn dataset.
const NIR = Array.from({ length: 700 }, (_, i) => String(1100 + i * 2));
const idx = (nm: number) => (nm - 1100) / 2;

describe('toRuns', () => {
    it('returns nothing for an empty selection', () => {
        expect(toRuns([])).toEqual([]);
    });

    it('collapses a contiguous block into one run', () => {
        expect(toRuns([3, 4, 5, 6])).toEqual([{ start: 3, end: 6 }]);
    });

    it('separates disjoint blocks', () => {
        expect(toRuns([1, 2, 7, 8, 9])).toEqual([{ start: 1, end: 2 }, { start: 7, end: 9 }]);
    });

    it('sorts and de-duplicates its input', () => {
        expect(toRuns([5, 3, 4, 3])).toEqual([{ start: 3, end: 5 }]);
    });

    it('turns the two Corn water bands into two runs, not 57 entries', () => {
        const water = [
            ...Array.from({ length: 26 }, (_, i) => idx(1400) + i),
            ...Array.from({ length: 31 }, (_, i) => idx(1900) + i)
        ];
        expect(water).toHaveLength(57);
        const runs = toRuns(water);
        expect(runs).toHaveLength(2);
        expect(describeRun(runs[0], NIR)).toBe('1400–1450');
        expect(describeRun(runs[1], NIR)).toBe('1900–1960');
    });
});

describe('describeRun', () => {
    it('names a single column without a dash', () => {
        expect(describeRun({ start: 0, end: 0 }, NIR)).toBe('1100');
    });
    it('falls back to a 1-based position when a name is missing', () => {
        expect(describeRun({ start: 1, end: 2 }, [])).toBe('2–3');
    });
});

describe('runToIndices', () => {
    it('expands inclusive bounds', () => {
        expect(runToIndices({ start: 2, end: 5 })).toEqual([2, 3, 4, 5]);
    });
});

describe('parseRangeSpec', () => {
    it('resolves a range written as wavelengths', () => {
        const { indices, errors } = parseRangeSpec('1400-1450', NIR);
        expect(errors).toEqual([]);
        expect(indices).toHaveLength(26);
        expect(NIR[indices[0]]).toBe('1400');
        expect(NIR[indices[25]]).toBe('1450');
    });

    it('handles the full Corn water-band expression', () => {
        const { indices, errors } = parseRangeSpec('1400-1450, 1900-1960', NIR);
        expect(errors).toEqual([]);
        expect(indices).toHaveLength(57);
    });

    it('prefers a column name over an index, matching the CLI', () => {
        // Columns named 5..12: "5" must mean the column called 5, not the fifth.
        const headers = ['5', '6', '7', '8', '9', '10', '11', '12'];
        expect(parseRangeSpec('5', headers).indices).toEqual([0]);
    });

    it('falls back to a 1-based index when no column carries that name', () => {
        expect(parseRangeSpec('3', ['a', 'b', 'c', 'd']).indices).toEqual([2]);
    });

    it('treats a hyphenated column name as a name, not a range', () => {
        expect(parseRangeSpec('col-1', ['col-1', 'col-2', 'x']).indices).toEqual([0]);
    });

    it('reports unresolvable tokens instead of dropping them', () => {
        const { indices, errors } = parseRangeSpec('1400, nope, 99999', NIR);
        expect(indices).toEqual([idx(1400)]);
        expect(errors).toHaveLength(2);
    });

    it('rejects an inverted or out-of-bounds index range', () => {
        expect(parseRangeSpec('9-2', ['a', 'b', 'c']).errors).toHaveLength(1);
        expect(parseRangeSpec('1-99', ['a', 'b', 'c']).errors).toHaveLength(1);
    });

    it('ignores empty tokens and surrounding whitespace', () => {
        expect(parseRangeSpec(' 1400 , , 1402 ', NIR).indices)
            .toEqual([idx(1400), idx(1402)]);
    });

    it('de-duplicates overlapping ranges', () => {
        const { indices } = parseRangeSpec('1100-1110, 1106-1120', NIR);
        expect(indices).toEqual([...new Set(indices)]);
        expect(indices).toHaveLength(11);
    });
});

describe('columnMeans', () => {
    it('averages each column', () => {
        expect(columnMeans([[1, 10], [3, 20]], 2)).toEqual([2, 15]);
    });
    it('skips non-finite values rather than poisoning the mean', () => {
        expect(columnMeans([[1, NaN], [3, 20]], 2)).toEqual([2, 20]);
    });
    it('returns zero for a column with no finite values', () => {
        expect(columnMeans([[NaN], [NaN]], 1)).toEqual([0]);
    });
});

describe('namesFormOrderedAxis', () => {
    it('accepts a spectrum sampled at regular intervals', () => {
        expect(namesFormOrderedAxis(NIR)).toBe(true);
    });

    it('accepts a descending axis', () => {
        expect(namesFormOrderedAxis(['2498', '2496', '2494'])).toBe(true);
    });

    it('accepts an unevenly sampled axis, which is still an axis', () => {
        expect(namesFormOrderedAxis(['400', '410', '480', '1200'])).toBe(true);
    });

    it('rejects named variables — a line through them would invent continuity', () => {
        expect(namesFormOrderedAxis(['alcohol', 'malic_acid', 'ash'])).toBe(false);
    });

    it('rejects numeric names that carry no order', () => {
        expect(namesFormOrderedAxis(['1', '5', '3'])).toBe(false);
    });

    it('rejects a mix of numbers and names', () => {
        expect(namesFormOrderedAxis(['1100', '1102', 'Moisture'])).toBe(false);
    });

    it('rejects repeated values, which are not a strict ordering', () => {
        expect(namesFormOrderedAxis(['1100', '1100', '1102'])).toBe(false);
    });

    it('rejects blank names, since Number("") is 0', () => {
        expect(namesFormOrderedAxis(['', '1', '2'])).toBe(false);
    });

    it('rejects fewer than two columns, where the question is meaningless', () => {
        expect(namesFormOrderedAxis(['1100'])).toBe(false);
        expect(namesFormOrderedAxis([])).toBe(false);
    });
});

describe('columnExtents', () => {
    it('reports the smallest and largest value per column', () => {
        const data = [[1, 40], [3, 10], [2, 20]];
        expect(columnExtents(data, 2)).toEqual([
            { min: 1, max: 3, empty: false },
            { min: 10, max: 40, empty: false }
        ]);
    });

    it('ignores non-finite values rather than letting them win the comparison', () => {
        const data = [[NaN, 5], [2, Infinity], [8, 7]];
        expect(columnExtents(data, 2)).toEqual([
            { min: 2, max: 8, empty: false },
            { min: 5, max: 7, empty: false }
        ]);
    });

    it('marks a column with no finite values as empty', () => {
        expect(columnExtents([[NaN], [NaN]], 1)).toEqual([{ min: 0, max: 0, empty: true }]);
    });
});

describe('profileFractions', () => {
    it('places the smallest and largest column mean at the ends of the shared axis', () => {
        // Column means are 1, 5 and 9, so the middle one lands exactly halfway.
        const data = [[1, 5, 9]];
        expect(profileFractions(data, 3, 'shared')).toEqual([0, 0.5, 1]);
    });

    it('is the regression this change exists for: a large column no longer flattens the rest', () => {
        // Four spectral channels next to one target column two orders of magnitude
        // above them — the shape of testdata/corn/corn.csv.
        const data = [
            [0.10, 0.20, 0.30, 0.40, 64],
            [0.12, 0.22, 0.34, 0.44, 66]
        ];
        const shared = profileFractions(data, 5, 'shared');
        // Under one axis the four channels are indistinguishable from the floor.
        expect(shared.slice(0, 4).every(f => f < 0.01)).toBe(true);

        // Scaled per column they separate again: each mean sits mid-range, because
        // each channel here has two evenly spread observations.
        const independent = profileFractions(data, 5, 'independent');
        expect(independent.slice(0, 4).every(f => f > 0.4 && f < 0.6)).toBe(true);
    });

    it('puts the mean where it falls inside a skewed column', () => {
        // min 0, max 10, mean 1 -> one tenth of the way up.
        const data = [[0], [0], [0], [0], [0], [0], [0], [0], [0], [10]];
        expect(profileFractions(data, 1, 'independent')[0]).toBeCloseTo(0.1, 10);
    });

    it('gives a constant column half height, having no spread to place the mean in', () => {
        expect(profileFractions([[7], [7], [7]], 1, 'independent')).toEqual([0.5]);
    });

    it('draws nothing for a column with no finite values', () => {
        expect(profileFractions([[NaN], [NaN]], 1, 'independent')).toEqual([0]);
    });

    it('survives an empty dataset', () => {
        expect(profileFractions([], 0, 'shared')).toEqual([]);
        expect(profileFractions([], 0, 'independent')).toEqual([]);
    });
});

describe('sharedScaleIsReadable', () => {
    it('accepts columns of comparable magnitude, as in a spectrum', () => {
        const data = [Array.from({ length: 20 }, (_, i) => 0.5 + i * 0.01)];
        expect(sharedScaleIsReadable(data, 20)).toBe(true);
    });

    it('rejects a dataset where one column dwarfs the others', () => {
        // Wine: proline near 750 alongside hue near 1.
        const data = [[746, 0.96, 1.59, 0.36, 2.03]];
        expect(sharedScaleIsReadable(data, 5)).toBe(false);
    });

    it('treats no columns as trivially readable', () => {
        expect(sharedScaleIsReadable([], 0)).toBe(true);
    });
});
