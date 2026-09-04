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

/**
 * Which cross-validated error curve a PCR model's component count was chosen
 * from, and what to call it on screen.
 *
 * The engine selects on RMSECV or on MAE according to the Metric control, and
 * `selected`, `lowest_error` and `curve_still_falling` all describe that curve
 * and no other. The results panel originally read `rmsecv` unconditionally, so
 * choosing MAE gave a screen that plotted one curve, stated that it had chosen
 * the component count, and compared the selected count against its lowest point
 * — while all three numbers had come from the other curve. Nothing was wrong in
 * the model; the display simply reported the wrong one.
 *
 * This lives apart from the component so the choice can be tested directly. A
 * test that renders the panel would have to assert on prose to see the bug.
 */

export type SelectionMetric = 'rmse' | 'mae';

/** The parts of a CV report this module needs; a structural subset of CVReportJSON. */
export interface SelectionCurveSource {
    metric?: string;
    rmsecv: readonly number[];
    mae: readonly number[];
}

/**
 * The metric the rule was applied to.
 *
 * An absent or unrecognised value means RMSE, matching `types.MetricRMSE`, whose
 * zero value is the empty string. A model file written before the field existed
 * therefore renders as it always did rather than as a blank plot.
 */
export function selectionMetricOf(cv: SelectionCurveSource | null | undefined): SelectionMetric {
    return cv?.metric === 'mae' ? 'mae' : 'rmse';
}

/** The label a reader sees for that metric. */
export function metricLabel(metric: SelectionMetric): string {
    return metric === 'mae' ? 'MAE' : 'RMSECV';
}

/** The label for the measure that was *not* selected on. */
export function otherMetricLabel(metric: SelectionMetric): string {
    return metric === 'mae' ? 'RMSECV' : 'MAE';
}

/**
 * The error values the rule read, parallel to `candidates`.
 *
 * Returns an empty array rather than throwing when the report is absent, so a
 * caller can render an empty plot instead of failing.
 */
export function selectionCurveOf(cv: SelectionCurveSource | null | undefined): readonly number[] {
    if (!cv) return [];
    return selectionMetricOf(cv) === 'mae' ? cv.mae : cv.rmsecv;
}
