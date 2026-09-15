import { describe, expect, it } from 'vitest';
import { sampleLabel } from './sampleLabel';

describe('sampleLabel', () => {
    it('uses the row name when there is one', () => {
        expect(sampleLabel(['se_01', 'se_02'], 1)).toBe('se_02');
    });

    it('numbers from one when there are no row names', () => {
        // The first row of a file is row 1 in GoCSV's position gutter, so it
        // cannot be "Sample 0" here without the suite disagreeing with itself.
        expect(sampleLabel(undefined, 0)).toBe('Sample 1');
        expect(sampleLabel(undefined, 199)).toBe('Sample 200');
    });

    it('numbers from one when the array is short', () => {
        expect(sampleLabel(['only'], 1)).toBe('Sample 2');
    });

    it('treats an empty name as absent', () => {
        // An empty string is what a blank row-name cell produces, and it would
        // otherwise render as a point with no label at all.
        expect(sampleLabel(['', 'b'], 0)).toBe('Sample 1');
    });
});
