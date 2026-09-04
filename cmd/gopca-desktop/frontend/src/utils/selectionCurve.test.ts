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
import {
    selectionMetricOf,
    selectionCurveOf,
    metricLabel,
    otherMetricLabel
} from './selectionCurve';

// The two curves are deliberately different at every point. If the module ever
// returns the wrong one, no assertion below can pass by coincidence.
const report = {
    rmsecv: [1.0, 0.8, 0.7, 0.65],
    mae: [0.9, 0.5, 0.4, 0.3]
};

describe('selectionCurve', () => {
    it('reads the MAE curve when the report says the rule was applied to MAE', () => {
        const cv = { ...report, metric: 'mae' };
        expect(selectionMetricOf(cv)).toBe('mae');
        expect(selectionCurveOf(cv)).toEqual(report.mae);
    });

    it('reads the RMSECV curve when the report says RMSE', () => {
        const cv = { ...report, metric: 'rmse' };
        expect(selectionMetricOf(cv)).toBe('rmse');
        expect(selectionCurveOf(cv)).toEqual(report.rmsecv);
    });

    // The Go zero value for CVReport.Metric is the empty string, and a model
    // written before the field existed carries no metric at all. Both must mean
    // RMSE, which is what SelectionConfig documents, rather than rendering blank.
    it('treats a missing or empty metric as RMSE', () => {
        expect(selectionMetricOf({ ...report, metric: '' })).toBe('rmse');
        expect(selectionMetricOf(report)).toBe('rmse');
        expect(selectionCurveOf(report)).toEqual(report.rmsecv);
        expect(selectionMetricOf(undefined)).toBe('rmse');
        expect(selectionCurveOf(null)).toEqual([]);
    });

    it('labels each metric as the results panel names it', () => {
        expect(metricLabel('mae')).toBe('MAE');
        expect(metricLabel('rmse')).toBe('RMSECV');
        expect(otherMetricLabel('mae')).toBe('RMSECV');
        expect(otherMetricLabel('rmse')).toBe('MAE');
    });

    // The defect this module exists to prevent: `lowest_error` and `selected`
    // index the selected curve, so looking them up in the other one reports a
    // number that was never compared against anything. Here that mistake would
    // claim the rule gave up 0.65 - 0.7 rather than 0.3 - 0.4.
    it('indexes the same curve the rule chose from', () => {
        const cv = { ...report, metric: 'mae' };
        const curve = selectionCurveOf(cv);
        const candidates = [0, 1, 2, 3];
        const selected = 2;
        const lowest = 3;

        expect(curve[candidates.indexOf(selected)]).toBe(0.4);
        expect(curve[candidates.indexOf(lowest)]).toBe(0.3);
        expect(curve[candidates.indexOf(selected)]).not.toBe(report.rmsecv[selected]);
    });
});
