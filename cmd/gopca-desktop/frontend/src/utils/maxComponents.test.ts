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
import { maxComponentsFor, clampComponentCount } from './maxComponents';

describe('maxComponentsFor', () => {
    describe('SVD and NIPALS', () => {
        it('is bounded by the variable count when variables < samples', () => {
            expect(maxComponentsFor('SVD', 3, 1000)).toBe(3);
            expect(maxComponentsFor('NIPALS', 13, 178)).toBe(13);
        });

        it('is bounded by the sample count when samples < variables', () => {
            // Corn NIR: 80 samples, 700 wavelengths
            expect(maxComponentsFor('SVD', 700, 80)).toBe(80);
        });
    });

    describe('Kernel PCA', () => {
        // Regression test for #767: the spinner capped Kernel PCA at the
        // variable count, making it impossible to request the components the
        // method can actually produce. Kernel PCA eigendecomposes the n×n
        // kernel matrix, so the bound is the sample count.
        it('is bounded by the sample count, not the variable count', () => {
            expect(maxComponentsFor('kernel', 3, 1000)).toBe(1000);
        });

        it('allows more components than there are variables', () => {
            const nVariables = 3;
            expect(maxComponentsFor('kernel', nVariables, 1000)).toBeGreaterThan(nVariables);
        });

        it('does not exceed the sample count', () => {
            expect(maxComponentsFor('kernel', 700, 80)).toBe(80);
        });
    });

    describe('Temporal PCA', () => {
        // Trajectory matrix is [(n - L + 1) x (m * L)]; bound is the smaller side.
        it('is bounded by m * L when that is the smaller dimension', () => {
            // EEG: 14980 samples, 14 channels, L = 32 -> min(14949, 448) = 448
            expect(maxComponentsFor('temporal', 14, 14980, 32)).toBe(448);
        });

        it('is bounded by n - L + 1 when that is the smaller dimension', () => {
            // 100 samples, 12 variables, L = 40 -> min(61, 480) = 61
            expect(maxComponentsFor('temporal', 12, 100, 40)).toBe(61);
        });

        it('allows more components than there are variables', () => {
            // CSTR: 801 samples, 12 variables, L = 40 -> min(762, 480) = 480
            expect(maxComponentsFor('temporal', 12, 801, 40)).toBe(480);
        });

        it('treats a missing or zero lag count as L = 1', () => {
            expect(maxComponentsFor('temporal', 12, 801)).toBe(12);
            expect(maxComponentsFor('temporal', 12, 801, 0)).toBe(12);
        });
    });

    describe('guards', () => {
        it('never returns less than 1', () => {
            expect(maxComponentsFor('SVD', 0, 0)).toBe(1);
            expect(maxComponentsFor('kernel', 3, 0)).toBe(1);
            expect(maxComponentsFor('temporal', 12, 10, 40)).toBe(1);
        });

        it('treats an unknown method like SVD', () => {
            expect(maxComponentsFor('something-else', 3, 1000)).toBe(3);
        });
    });
});

describe('clampComponentCount', () => {
    // Regression test for the review finding on #768: the old handler used
    // `parseInt(value) || 5`, so clearing the field set the count to 5 — which
    // exceeds the ceiling on any dataset with fewer than 5 variables. On the
    // shipped Iris set (4 variables) the backend then rejects the run with
    // "too many components requested: maximum 4, got 5".
    it('keeps the previous value when the field is cleared', () => {
        expect(clampComponentCount('', 3, 4)).toBe(3);
    });

    it('keeps the previous value for non-numeric input', () => {
        expect(clampComponentCount('abc', 2, 4)).toBe(2);
    });

    it('does not substitute a default that exceeds the ceiling', () => {
        // Iris: 4 variables, so max = 4. Clearing must not yield 5.
        expect(clampComponentCount('', 4, 4)).toBeLessThanOrEqual(4);
    });

    it('clamps a typed value above the ceiling', () => {
        // `max` on a number input only constrains the stepper arrows.
        expect(clampComponentCount('999', 2, 4)).toBe(4);
    });

    it('clamps zero and negatives up to 1', () => {
        expect(clampComponentCount('0', 2, 4)).toBe(1);
        expect(clampComponentCount('-3', 2, 4)).toBe(1);
    });

    it('passes through a valid value unchanged', () => {
        expect(clampComponentCount('3', 1, 10)).toBe(3);
    });

    it('allows the large ceilings Kernel PCA permits', () => {
        const max = maxComponentsFor('kernel', 3, 1000);
        expect(clampComponentCount('250', 5, max)).toBe(250);
    });
});
