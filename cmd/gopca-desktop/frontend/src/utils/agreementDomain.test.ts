// GoPCA Suite
//
// Copyright © 2025-2026 Rune Mathisen <devel@bitjungle.com>
// SPDX-License-Identifier: See LICENSE file for details.

import { describe, expect, it } from 'vitest';
import { sharedDomainFor } from './agreementDomain';

describe('sharedDomainFor', () => {
    // The reported defect: corn Moisture runs from about 9.4 to 11.0, and the
    // plot was drawn from -10 to 12, putting every point in one corner. The
    // range has to follow the data.
    it('follows the data rather than a fixed span', () => {
        const domain = sharedDomainFor([9.38, 10.45, 11.0, 9.9]);
        expect(domain).toBeDefined();
        const [low, high] = domain!.x;
        expect(low).toBeGreaterThan(9);
        expect(high).toBeLessThan(11.5);
        // Nothing about this data justifies including zero.
        expect(low).toBeGreaterThan(0);
    });

    it('gives both axes the same range', () => {
        const domain = sharedDomainFor([1, 2, 3, 4, 5, 2.5, 3.5]);
        expect(domain!.x).toEqual(domain!.y);
    });

    it('leaves a margin so points do not sit on the frame', () => {
        const domain = sharedDomainFor([0, 10]);
        expect(domain!.x[0]).toBeCloseTo(-0.5, 10);
        expect(domain!.x[1]).toBeCloseTo(10.5, 10);
    });

    // A constant response has no spread, and a fraction of zero would collapse
    // the range to a single point that cannot be drawn.
    it('still produces a drawable range for a constant value', () => {
        const domain = sharedDomainFor([7, 7, 7]);
        expect(domain!.x[0]).toBeLessThan(domain!.x[1]);
    });

    it('produces a drawable range when every value is zero', () => {
        const domain = sharedDomainFor([0, 0]);
        expect(domain!.x[0]).toBeLessThan(domain!.x[1]);
    });

    it('ignores values that are not finite', () => {
        const domain = sharedDomainFor([1, NaN, 3, Infinity]);
        expect(domain!.x[0]).toBeCloseTo(0.9, 10);
        expect(domain!.x[1]).toBeCloseTo(3.1, 10);
    });

    // Nothing to scale: leave the chart to autoscale rather than invent a range.
    it('returns nothing when there is no usable data', () => {
        expect(sharedDomainFor([])).toBeUndefined();
        expect(sharedDomainFor([NaN, NaN])).toBeUndefined();
    });

    it('handles a response that spans zero', () => {
        const domain = sharedDomainFor([-4, 6]);
        expect(domain!.x[0]).toBeCloseTo(-4.5, 10);
        expect(domain!.x[1]).toBeCloseTo(6.5, 10);
    });
});
