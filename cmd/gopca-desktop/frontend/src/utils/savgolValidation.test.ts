// GoPCA Suite
//
// Copyright © 2025-2026 Rune Mathisen <devel@bitjungle.com>
//
// This file is part of GoPCA Suite.
//
// See LICENSE for the full license terms.

import { describe, it, expect } from 'vitest';
import {
    validateSavGol,
    snapWindowToOdd,
    savgolSelection,
    applySavGolSelection,
    DEFAULT_SAVGOL_WINDOW
} from './savgolValidation';

const settings = (savgolWindow: number, savgolPolyOrder = 2, savgolDeriv = 0) =>
    ({ savgolWindow, savgolPolyOrder, savgolDeriv });

describe('validateSavGol', () => {
    it('accepts the filter being switched off, whatever the other values say', () => {
        // Order and derivative keep their values while the filter is off, so the
        // panel has something to show the moment it is switched on. They must
        // not produce a complaint in the meantime.
        expect(validateSavGol(settings(0, 0, 5), 100)).toBeNull();
    });

    it('accepts a usable configuration', () => {
        expect(validateSavGol(settings(11, 2, 1), 700)).toBeNull();
    });

    it('rejects an even window, which has no centre', () => {
        expect(validateSavGol(settings(10, 2, 1), 700)).toMatch(/odd/);
    });

    it('rejects a window below three', () => {
        expect(validateSavGol(settings(1, 0, 0), 700)).toMatch(/at least 3/);
    });

    it('rejects an order that is not below the window length', () => {
        expect(validateSavGol(settings(5, 5, 0), 700)).toMatch(/less than the window length/);
    });

    // The one that would otherwise succeed and return zeros everywhere, which
    // looks like a filter that destroyed the data rather than a bad request.
    it('rejects a derivative the polynomial cannot have, and says what it would produce', () => {
        const message = validateSavGol(settings(7, 2, 3), 700);
        expect(message).toMatch(/zero everywhere/);
        expect(message).toMatch(/Raise the polynomial order or lower the derivative/);
    });

    it('rejects a window wider than the data', () => {
        expect(validateSavGol(settings(101, 2, 1), 50)).toMatch(/wider than the 50 variables/);
    });

    it('does not judge the width when the variable count is unknown', () => {
        // Before a file is loaded there is nothing to compare against, and
        // complaining then would flag a perfectly good default as an error.
        expect(validateSavGol(settings(101, 2, 1), 0)).toBeNull();
    });

    it('blocks the run when a filter is left set on Temporal PCA', () => {
        // Reachable by configuring the filter and then switching method: the
        // selector goes disabled, but the window it left behind is still set.
        const message = validateSavGol(settings(11, 2, 1), 700, 'temporal');
        expect(message).toMatch(/not available for Temporal PCA/i);
        expect(message).toMatch(/Set it to None or choose another method/);
    });

    it('says nothing about Temporal PCA when no filter is set', () => {
        expect(validateSavGol(settings(0, 2, 0), 700, 'temporal')).toBeNull();
    });

    it('accepts the filter for every other method', () => {
        for (const method of ['SVD', 'NIPALS', 'kernel', undefined]) {
            expect(validateSavGol(settings(11, 2, 1), 700, method)).toBeNull();
        }
    });

    it('rejects a fractional window', () => {
        expect(validateSavGol(settings(10.5, 2, 0), 700)).toMatch(/whole number/);
    });
});

describe('snapWindowToOdd', () => {
    it.each([
        [10, 11],
        [11, 11],
        [4, 5],
        [3, 3],
        // Below the minimum, snap up rather than to a value the engine refuses.
        [1, 3],
        [2, 3],
        [10.4, 11]
    ])('snaps %d to %d', (input, expected) => {
        expect(snapWindowToOdd(input)).toBe(expected);
    });

    it('treats zero and nonsense as "off" rather than snapping them up', () => {
        // A cleared field arrives as NaN. Snapping that to 3 would switch the
        // filter on while the user was in the middle of retyping the number.
        expect(snapWindowToOdd(0)).toBe(0);
        expect(snapWindowToOdd(NaN)).toBe(0);
        expect(snapWindowToOdd(-5)).toBe(0);
    });
});

describe('savgolSelection', () => {
    it('reads a zero window as off, whatever the derivative says', () => {
        expect(savgolSelection(settings(0, 2, 2))).toBe('none');
    });

    it.each([
        [0, 'smooth'],
        [1, 'deriv1'],
        [2, 'deriv2']
    ])('reads derivative %d as %s', (deriv, expected) => {
        expect(savgolSelection(settings(11, 2, deriv))).toBe(expected);
    });
});

describe('applySavGolSelection', () => {
    it('switching off keeps the other settings, so switching back on restores them', () => {
        const off = applySavGolSelection('none', settings(15, 3, 2));
        expect(off.savgolWindow).toBe(0);
        expect(off.savgolPolyOrder).toBe(3);
        expect(off.savgolDeriv).toBe(2);
    });

    it('switching on from off supplies a window, or the filter would stay off', () => {
        const on = applySavGolSelection('deriv1', settings(0, 2, 0));
        expect(on.savgolWindow).toBe(DEFAULT_SAVGOL_WINDOW);
        expect(on.savgolDeriv).toBe(1);
    });

    it('keeps a window the user has already chosen', () => {
        expect(applySavGolSelection('deriv1', settings(21, 3, 0)).savgolWindow).toBe(21);
    });

    // The rule that stops the panel offering a configuration the engine refuses.
    it('raises the polynomial order to meet the derivative', () => {
        const result = applySavGolSelection('deriv2', settings(11, 1, 0));
        expect(result.savgolPolyOrder).toBe(2);
        expect(validateSavGol(result, 700)).toBeNull();
    });

    it('never lowers an order the user chose', () => {
        expect(applySavGolSelection('deriv1', settings(11, 4, 0)).savgolPolyOrder).toBe(4);
    });

    // The property that matters more than any individual case: no reachable
    // selection may produce settings the engine would reject.
    it('produces a valid configuration from every selection and starting point', () => {
        const starts = [
            settings(0, 2, 0), settings(0, 0, 0), settings(11, 2, 1),
            settings(5, 1, 0), settings(21, 4, 2)
        ];
        for (const start of starts) {
            for (const selection of ['none', 'smooth', 'deriv1', 'deriv2'] as const) {
                const result = applySavGolSelection(selection, start);
                expect(validateSavGol(result, 700)).toBeNull();
            }
        }
    });
});
