// GoPCA Suite
//
// Copyright © 2025-2026 Rune Mathisen <devel@bitjungle.com>
//
// This file is part of GoPCA Suite.
//
// See LICENSE for the full license terms.

import { describe, it, expect } from 'vitest';
import { placeCircleLabels, CircleLabel } from './circleLabelPlacement';

/**
 * Wine, standard-scaled, PC1/PC2 — the case reported in #830.
 *
 * Correlations of each variable with the first two components, computed
 * independently with scikit-learn rather than taken from GoPCA, so the fixture
 * does not agree with the code under test merely because both came from it.
 * Three arrows sit within about 7° of each other and their names collide.
 */
const WINE: CircleLabel[] = [
    { text: 'alcohol', x: 0.313093, y: 0.764257 },
    { text: 'malic_acid', x: -0.531885, y: 0.355432 },
    { text: 'ash', x: -0.004449, y: 0.499446 },
    { text: 'alcalinity_of_ash', x: -0.519157, y: -0.016735 },
    { text: 'magnesium', x: 0.308023, y: 0.473476 },
    { text: 'total_phenols', x: 0.856137, y: 0.102774 },
    { text: 'flavanoids', x: 0.917470, y: -0.005309 },
    { text: 'nonflavanoid_phenols', x: -0.647607, y: 0.045477 },
    { text: 'proanthocyanins', x: 0.679922, y: 0.062104 },
    { text: 'color_intensity', x: -0.192236, y: 0.837489 },
    { text: 'hue', x: 0.643662, y: -0.441242 },
    { text: 'od280/od315_of_diluted_wines', x: 0.816019, y: -0.259934 },
    { text: 'proline', x: 0.622051, y: 0.576613 }
];

const FONT = 10;
const PLOT = 500;
const SPAN = 2.4;
const PAD = 6;
const dataPerPixel = SPAN / PLOT;
const halfWidth = (text: string) => ((text.length * 0.55 * FONT) + 2 * PAD) * dataPerPixel / 2;
const halfHeight = ((1.2 * FONT) + 2 * PAD) * dataPerPixel / 2;

/** Pairs whose label boxes intersect, by name. */
function collisions(labels: CircleLabel[], positions: { x: number; y: number }[]): string[] {
    const found: string[] = [];
    for (let i = 0; i < labels.length; i++) {
        for (let j = i + 1; j < labels.length; j++) {
            const overlapping
                = Math.abs(positions[i].x - positions[j].x) < halfWidth(labels[i].text) + halfWidth(labels[j].text)
                && Math.abs(positions[i].y - positions[j].y) < halfHeight * 2;
            if (overlapping) {
                found.push(`${labels[i].text} <-> ${labels[j].text}`);
            }
        }
    }
    return found.sort();
}

/** What the code did before this existed: radial placement, no separation. */
const radialOnly = (labels: CircleLabel[]) =>
    labels.map(l => ({ x: l.x * 1.15, y: l.y * 1.15 }));

describe('placeCircleLabels', () => {
    it('reproduces the three collisions reported on Wine', () => {
        // Without this the next test would pass on a function that returned its
        // input, and prove nothing about Wine. Naming the pairs also pins the
        // calibration: if the box model drifts, this notices before the fix
        // silently starts under- or over-firing.
        expect(collisions(WINE, radialOnly(WINE))).toEqual([
            'alcalinity_of_ash <-> nonflavanoid_phenols',
            'flavanoids <-> proanthocyanins',
            'total_phenols <-> proanthocyanins'
        ]);
    });

    it('does not treat the next-closest pair as a collision', () => {
        // magnesium and alcohol are 10.8 degrees apart and read fine. A model
        // that flagged them would be separating labels that never needed it.
        expect(collisions(WINE, radialOnly(WINE)).join(' ')).not.toContain('magnesium');
    });

    it('separates every colliding pair on Wine', () => {
        const placed = placeCircleLabels(WINE, { fontSizePx: FONT, plotSizePx: PLOT, axisSpan: SPAN });
        expect(collisions(WINE, placed)).toEqual([]);
    });

    it('never lets a label cover its own arrow, unless it cannot fit outside', () => {
        // The defect a screenshot showed after the first fix: labels were
        // centred on a point beyond the tip, so the inner half of the text lay
        // back over the arrow it names. On Wine that covered up to a third of
        // some arrows, and proanthocyanins rendered as "oanthocyanins" with its
        // first letters lost under its own arrowhead.
        const LIMIT = 1.2;
        const placed = placeCircleLabels(WINE, { fontSizePx: FONT, plotSizePx: PLOT, axisSpan: SPAN });
        WINE.forEach((label, i) => {
            const tip = Math.hypot(label.x, label.y);
            const angle = Math.atan2(placed[i].y, placed[i].x);
            const radius = Math.hypot(placed[i].x, placed[i].y);
            const reach
                = halfWidth(label.text) * Math.abs(Math.cos(angle))
                + halfHeight * Math.abs(Math.sin(angle));

            const clears = radius - reach >= tip - 1e-9;
            // The only permitted exception: clearing the tip would push the
            // text off the axis, and a clipped label is worse than a crowded
            // one. Then it must be left exactly where it used to be.
            const cannotFit = tip + 2 * reach > LIMIT;
            if (!clears) {
                expect(cannotFit, `${label.text} overlaps its arrow but would have fitted outside`).toBe(true);
                // Pushed as far out as the axis allows rather than abandoned,
                // and never pulled inward -- a label among the arrows would be
                // worse than one that merely touches its own.
                expect(radius).toBeGreaterThanOrEqual(tip * 1.15 - 1e-9);
                // Not pushed out into being clipped either. A very long name on
                // a long arrow already extended past the axis under the old
                // placement, so the bound is "no further than it was", not "in
                // range" -- moving it out would make an existing overflow worse.
                expect(radius).toBeLessThanOrEqual(Math.max(tip * 1.15, 1.2 - reach) + 1e-9);
            }
        });
    });

    it('leaves a name too long to fit exactly where it was', () => {
        // od280/od315_of_diluted_wines is 28 characters on an arrow of 0.86.
        // Clearing its tip would put its far edge at 1.67 against an axis that
        // stops at 1.2, so it stays put rather than being cut off.
        const long = WINE.find(w => w.text.startsWith('od280'))!;
        const placed = placeCircleLabels([long], { fontSizePx: FONT, plotSizePx: PLOT, axisSpan: SPAN });
        expect(Math.hypot(placed[0].x, placed[0].y)).toBeCloseTo(Math.hypot(long.x, long.y) * 1.15, 10);
    });

    it('does that by pushing labels out, never by pulling them in', () => {
        // A label closer to the origin than the old placement would sit among
        // the arrows rather than outside them.
        const placed = placeCircleLabels(WINE, { fontSizePx: FONT, plotSizePx: PLOT, axisSpan: SPAN });
        WINE.forEach((label, i) => {
            const before = Math.hypot(label.x, label.y) * 1.15;
            const after = Math.hypot(placed[i].x, placed[i].y);
            expect(after).toBeGreaterThanOrEqual(before - 1e-12);
        });
    });

    it('leaves a vertical arrow almost where it was, since text is thin that way', () => {
        // Pushing a label clear of its tip costs half a line height for an arrow
        // pointing up, against half a name's width for one pointing sideways.
        // A rule that ignored direction would shove vertical labels far out.
        const name = 'phenols';
        const up = placeCircleLabels([{ text: name, x: 0, y: 0.7 }],
            { fontSizePx: FONT, plotSizePx: PLOT, axisSpan: SPAN });
        expect(Math.hypot(up[0].x, up[0].y)).toBeLessThan(0.7 * 1.15 + 0.03);

        const right = placeCircleLabels([{ text: name, x: 0.7, y: 0 }],
            { fontSizePx: FONT, plotSizePx: PLOT, axisSpan: SPAN });
        expect(Math.hypot(right[0].x, right[0].y)).toBeGreaterThan(0.7 * 1.15 + 0.03);
    });

    it('never rotates a label further than the cap from its own arrow', () => {
        // A label that has drifted far from its arrow is no longer obviously
        // that arrow's name, which is the problem rather than the fix.
        const cap = 14;
        const placed = placeCircleLabels(WINE, {
            fontSizePx: FONT, plotSizePx: PLOT, axisSpan: SPAN, maxShiftDegrees: cap
        });
        for (const p of placed) {
            expect(Math.abs(p.shiftedDegrees)).toBeLessThanOrEqual(cap + 1e-9);
        }
    });

    it('leaves labels that do not collide exactly where they were', () => {
        // Four arrows at the compass points cannot overlap, and moving them
        // would break the radial convention for no reason.
        const spread: CircleLabel[] = [
            { text: 'east', x: 0.9, y: 0 },
            { text: 'north', x: 0, y: 0.9 },
            { text: 'west', x: -0.9, y: 0 },
            { text: 'south', x: 0, y: -0.9 }
        ];
        const placed = placeCircleLabels(spread, { fontSizePx: FONT, plotSizePx: PLOT, axisSpan: SPAN });
        placed.forEach((p, i) => {
            expect(p.shiftedDegrees).toBe(0);
            expect(p.x).toBeCloseTo(spread[i].x * 1.15, 12);
            expect(p.y).toBeCloseTo(spread[i].y * 1.15, 12);
        });
    });

    it('is deterministic, so a plot does not rearrange itself between renders', () => {
        const a = placeCircleLabels(WINE, { fontSizePx: FONT, plotSizePx: PLOT, axisSpan: SPAN });
        const b = placeCircleLabels(WINE, { fontSizePx: FONT, plotSizePx: PLOT, axisSpan: SPAN });
        expect(a).toEqual(b);
    });

    it('separates a pair that straddles the angle wrap at 180 degrees', () => {
        // Naive angle arithmetic pushes these through each other instead of
        // apart, because their difference looks like nearly a full turn.
        const straddling: CircleLabel[] = [
            { text: 'just_below_pi', x: -0.9, y: 0.02 },
            { text: 'just_above_pi', x: -0.9, y: -0.02 }
        ];
        const placed = placeCircleLabels(straddling, { fontSizePx: FONT, plotSizePx: PLOT, axisSpan: SPAN });
        expect(collisions(straddling, placed)).toEqual([]);
    });

    it('handles degenerate input without throwing', () => {
        expect(placeCircleLabels([], {})).toEqual([]);
        expect(placeCircleLabels([{ text: 'only', x: 0.5, y: 0.5 }], {})).toHaveLength(1);
        // A variable uncorrelated with both components sits at the origin,
        // where there is no direction to rotate around.
        const atOrigin = placeCircleLabels([
            { text: 'a', x: 0, y: 0 },
            { text: 'b', x: 0, y: 0 }
        ], {});
        expect(atOrigin).toHaveLength(2);
        expect(atOrigin.every(p => Number.isFinite(p.x) && Number.isFinite(p.y))).toBe(true);
    });

    it('gives up rather than dragging a label away from its arrow', () => {
        // Ten identical long names at one angle cannot all be separated within
        // the cap. Leaving some crowded is the intended outcome; a label parked
        // beside the wrong arrow would be worse than one that is hard to read.
        const crowded: CircleLabel[] = Array.from({ length: 10 }, (_, i) => ({
            text: `a_very_long_variable_name_${i}`,
            x: 0.8,
            y: 0.001 * i
        }));
        const placed = placeCircleLabels(crowded, {
            fontSizePx: FONT, plotSizePx: PLOT, axisSpan: SPAN, maxShiftDegrees: 5
        });
        expect(placed).toHaveLength(10);
        for (const p of placed) {
            expect(Math.abs(p.shiftedDegrees)).toBeLessThanOrEqual(5 + 1e-9);
        }
    });
});

describe('option handling', () => {
    it('falls back to the default when an option is explicitly undefined', () => {
        // The chart passes plotSizePx from a measurement that does not exist on
        // the first render. Spreading undefined over the defaults would make
        // every pixel size NaN and every position NaN, and the plot would
        // silently render nothing.
        const withUndefined = placeCircleLabels(WINE, { fontSizePx: FONT, plotSizePx: undefined });
        expect(withUndefined.every(p => Number.isFinite(p.x) && Number.isFinite(p.y))).toBe(true);
        expect(withUndefined).toEqual(placeCircleLabels(WINE, { fontSizePx: FONT }));
    });

    it('separates more aggressively on a smaller plot', () => {
        // The same labels occupy more of the data range when the plot is small,
        // so more pairs collide and more of them have to move.
        const shifted = (size: number) => placeCircleLabels(WINE, { fontSizePx: FONT, plotSizePx: size })
            .filter(p => p.shiftedDegrees !== 0).length;
        expect(shifted(300)).toBeGreaterThan(shifted(900));
    });
});
