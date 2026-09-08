// GoPCA Suite
//
// Copyright © 2025-2026 Rune Mathisen <devel@bitjungle.com>
//
// This file is part of GoPCA Suite.
//
// See LICENSE for the full license terms.

/**
 * Validation for the Savitzky-Golay controls.
 *
 * The engine enforces all of this and returns a clear message, so none of it is
 * load-bearing for correctness. It exists so the panel can say what is wrong
 * while the user is still looking at the control that caused it, rather than
 * after a run that had no chance of succeeding.
 *
 * Kept out of the component and exported so it can be tested directly: these
 * are arithmetic rules with edge cases, and reaching them through rendered
 * markup would test React rather than the rules.
 */

export interface SavGolSettings {
    savgolWindow: number;
    savgolPolyOrder: number;
    savgolDeriv: number;
}

/**
 * Returns a message describing the first problem with the settings, or null if
 * they are usable.
 *
 * `variableCount` is the number of variables the filter will run across, after
 * any exclusions — a window wider than the data has nothing to slide along.
 */
export function validateSavGol(
    settings: SavGolSettings,
    variableCount: number
): string | null {
    const { savgolWindow: window, savgolPolyOrder: order, savgolDeriv: deriv } = settings;

    // Zero is how the filter is switched off, so nothing else matters.
    if (window === 0) {
        return null;
    }

    if (!Number.isInteger(window) || window < 0) {
        return 'Window length must be a whole number of variables.';
    }
    if (window < 3) {
        return 'Window length must be at least 3.';
    }
    if (window % 2 === 0) {
        return 'Window length must be odd, so the window has a centre.';
    }
    if (!Number.isInteger(order) || order < 0) {
        return 'Polynomial order must be a whole number.';
    }
    if (order >= window) {
        return `Polynomial order (${order}) must be less than the window length (${window}).`;
    }
    if (deriv > order) {
        // Worth naming the consequence: the run would succeed and return zeros,
        // which looks like a filter that destroyed the data.
        return `A degree-${order} polynomial has no derivative of order ${deriv}, so the result would be zero everywhere. Raise the polynomial order or lower the derivative.`;
    }
    if (variableCount > 0 && window > variableCount) {
        return `Window length (${window}) is wider than the ${variableCount} variables available.`;
    }
    return null;
}

/**
 * Nudges a window length to the nearest valid odd value at or above 3.
 *
 * The number input steps by two from an odd start, but a user can still type an
 * even number. Snapping on commit is friendlier than refusing the keystroke,
 * which would make the field impossible to edit through an even intermediate
 * value such as the 1 in 11.
 */
export function snapWindowToOdd(value: number): number {
    if (!Number.isFinite(value) || value <= 0) {
        return 0;
    }
    const rounded = Math.round(value);
    if (rounded < 3) {
        return 3;
    }
    return rounded % 2 === 0 ? rounded + 1 : rounded;
}

/** The four states the panel's Savitzky-Golay selector can be in. */
export type SavGolSelection = 'none' | 'smooth' | 'deriv1' | 'deriv2';

/** Default window when the filter is first switched on. */
export const DEFAULT_SAVGOL_WINDOW = 11;

/**
 * Maps the stored settings back to the selector's value.
 *
 * A window of zero is the only thing that means "off"; the derivative order is
 * kept across a switch to None so that turning the filter back on returns the
 * user to what they had.
 */
export function savgolSelection(settings: SavGolSettings): SavGolSelection {
    if (settings.savgolWindow === 0) {
        return 'none';
    }
    if (settings.savgolDeriv === 1) {
        return 'deriv1';
    }
    if (settings.savgolDeriv === 2) {
        return 'deriv2';
    }
    return 'smooth';
}

/**
 * Applies a selector change to the settings.
 *
 * Two rules are worth stating because both are easy to get wrong and neither
 * fails loudly:
 *
 * 1. Switching on from None needs a window. Leaving it at zero would leave the
 *    selector showing a derivative while the filter stayed off.
 * 2. A derivative needs a polynomial that has one. Choosing a 2nd derivative
 *    with a linear fit is a configuration the engine refuses, and one that would
 *    return zeros everywhere if it did not — so raise the order to meet it
 *    rather than presenting a choice that cannot work.
 */
export function applySavGolSelection(
    selection: SavGolSelection,
    current: SavGolSettings
): SavGolSettings {
    if (selection === 'none') {
        return { ...current, savgolWindow: 0 };
    }
    const deriv = selection === 'deriv1' ? 1 : selection === 'deriv2' ? 2 : 0;
    return {
        savgolWindow: current.savgolWindow === 0 ? DEFAULT_SAVGOL_WINDOW : current.savgolWindow,
        savgolPolyOrder: Math.max(current.savgolPolyOrder, deriv),
        savgolDeriv: deriv
    };
}
