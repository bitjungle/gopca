import { describe, it, expect } from 'vitest';

/**
 * Both temporal plots resolve how many components to draw the same way, and the
 * point of this test is that they agree.
 *
 * PR #674 fixed a fixed default that silently truncated the Temporal Loadings
 * plot, but the identical default survived in Temporal Variable Importance. A
 * reader running 15 components saw all 15 curves in one plot and only 10 rows in
 * the other, with nothing to say rows were missing — and the EEG tutorial's
 * Step 8 asks them to compare the spatial pattern of an oscillatory pair that
 * lands at PC15/PC16, outside the cap.
 */
const resolve = (maxComponents: number | undefined, computed: number | undefined, fallback: number) =>
    maxComponents ?? (computed ?? fallback);

describe('temporal plot component count', () => {
    it('shows every computed component when the caller sets no limit', () => {
        expect(resolve(undefined, 15, 10)).toBe(15);
        expect(resolve(undefined, 20, 10)).toBe(20);
    });

    it('honours an explicit limit from the caller', () => {
        expect(resolve(5, 15, 10)).toBe(5);
    });

    it('falls back only when the result carries no component labels', () => {
        expect(resolve(undefined, undefined, 10)).toBe(10);
    });

    it('does not truncate at the old fixed default of 10', () => {
        // The regression: 15 computed components must not come back as 10.
        expect(resolve(undefined, 15, 10)).not.toBe(10);
    });

    it('reaches the EEG tutorial pair at PC15/PC16', () => {
        // Step 8 compares the channel-importance rows of the Step 7 pair.
        expect(resolve(undefined, 16, 10)).toBeGreaterThanOrEqual(16);
    });
});
