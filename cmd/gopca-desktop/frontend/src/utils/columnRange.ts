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

// Range selection for column pickers.
//
// Wide datasets are the motivating case. A 700-channel NIR spectrum renders as a
// 700-checkbox strip, so excluding a contiguous region — a water absorption band,
// a noisy detector edge — otherwise means one click per channel. Shift-click sets
// every column between the anchor and the clicked column to the same state.

/**
 * Apply a toggle to a column selection, extending to a range when shift is held.
 *
 * @param selection current map of column index to included/excluded
 * @param index     the column that was clicked
 * @param checked   the state to apply
 * @param anchor    the previously toggled column, or null if there is none
 * @param key       maps a column index to its key in the selection map
 * @returns the updated selection; the caller stores `index` as the next anchor
 */
export function applyColumnToggle<K extends string | number>(
    selection: Record<K, boolean>,
    index: number,
    checked: boolean,
    anchor: number | null,
    key: (i: number) => K
): Record<K, boolean> {
    const next = { ...selection };
    if (anchor !== null) {
        const lo = Math.min(anchor, index);
        const hi = Math.max(anchor, index);
        for (let i = lo; i <= hi; i++) {
            next[key(i)] = checked;
        }
    } else {
        next[key(index)] = checked;
    }
    return next;
}
