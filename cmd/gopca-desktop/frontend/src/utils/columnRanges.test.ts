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
import { toRuns, describeRun, runToIndices, parseRangeSpec, columnMeans } from './columnRanges';

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
