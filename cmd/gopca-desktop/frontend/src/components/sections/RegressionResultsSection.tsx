// GoPCA Suite
//
// Copyright © 2025-2026 Rune Mathisen <devel@bitjungle.com>
// SPDX-License-Identifier: See LICENSE file for details.

import React, { useMemo } from 'react';
import { PlotlyScatterChart, PlotlyLineChart } from '@gopca/ui-components';
import { usePCRContext } from '../../contexts/PCRContext';
import { useFileDataContext } from '../../contexts/FileDataContext';

/** Formats a number for display, showing a dash where there is nothing to show. */
function show(value: number | undefined | null, digits = 4): string {
    if (value === undefined || value === null || Number.isNaN(value)) {
        return '—';
    }
    return Number(value).toPrecision(digits);
}

/**
 * RegressionResultsSection presents the fitted model.
 *
 * The three error figures are shown under their own names with a note on what
 * each is for. RMSEC is the smallest and most flattering number on any
 * well-fitted model, and a reader who takes it for a performance estimate will
 * overstate what the model can do, so it is labelled at the point of display
 * rather than in documentation the reader may never open.
 */
export const RegressionResultsSection: React.FC = () => {
    const { pcrResponse, pcrResultsRef } = usePCRContext();
    const { fileData } = useFileDataContext();

    const result = pcrResponse?.result;

    const selectionCurve = useMemo(() => {
        if (!result?.cv) return [];
        return result.cv.candidates.map((k, i) => ({
            x: k,
            y: Number(result.cv!.rmsecv[i])
        }));
    }, [result]);

    const predictedVsMeasured = useMemo(() => {
        if (!result) return [];
        const heldOut = result.cv?.out_of_fold_predictions;
        return result.fitted.map((fitted, i) => {
            const measured = Number(fitted) + Number(result.residuals[i]);
            return {
                x: measured,
                y: heldOut ? Number(heldOut[i]) : Number(fitted),
                name: fileData?.rowNames?.[result.labelled_rows?.[i] ?? i] ?? `Row ${i + 1}`
            };
        });
    }, [result, fileData]);

    const coefficients = useMemo(() => {
        if (!result?.original_scale_valid || !result.coefficients) return [];
        return result.coefficients.map((value, i) => ({
            x: i + 1,
            y: Number(value),
            name: fileData?.headers?.[i] ?? `Variable ${i + 1}`
        }));
    }, [result, fileData]);

    if (!result) {
        return null;
    }

    const selectedIndex = result.cv
        ? result.cv.candidates.indexOf(result.cv.selected)
        : -1;
    const hitCeiling =
        result.cv !== undefined &&
        result.cv !== null &&
        result.cv.candidates.length > 1 &&
        result.cv.selected === result.cv.candidates[result.cv.candidates.length - 1];

    return (
        <div ref={pcrResultsRef} className="space-y-6">
            <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
                <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-4">
                    Regression: {result.response}
                </h2>

                <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
                    <div>
                        <div className="text-gray-500 dark:text-gray-400">Components</div>
                        <div className="text-xl font-semibold text-gray-900 dark:text-gray-100">
                            {result.components}
                        </div>
                    </div>
                    <div>
                        <div className="text-gray-500 dark:text-gray-400">RMSEC</div>
                        <div className="text-xl font-semibold text-gray-900 dark:text-gray-100">
                            {show(Number(result.rmsec))}
                        </div>
                        <div className="text-xs text-gray-500 dark:text-gray-400">training fit</div>
                    </div>
                    <div>
                        <div className="text-gray-500 dark:text-gray-400">RMSECV</div>
                        <div className="text-xl font-semibold text-gray-900 dark:text-gray-100">
                            {selectedIndex >= 0 ? show(Number(result.cv!.rmsecv[selectedIndex])) : '—'}
                        </div>
                        <div className="text-xs text-gray-500 dark:text-gray-400">held out</div>
                    </div>
                    <div>
                        <div className="text-gray-500 dark:text-gray-400">Q²</div>
                        <div className="text-xl font-semibold text-gray-900 dark:text-gray-100">
                            {selectedIndex >= 0 ? show(Number(result.cv!.q2[selectedIndex])) : '—'}
                        </div>
                    </div>
                </div>

                {selectedIndex >= 0 && (
                    <div className="mt-4 grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
                        <div>
                            <div className="text-gray-500 dark:text-gray-400">Bias</div>
                            <div className="text-gray-900 dark:text-gray-100">
                                {show(Number(result.cv!.bias[selectedIndex]))}
                            </div>
                        </div>
                        <div>
                            <div className="text-gray-500 dark:text-gray-400">SEP</div>
                            <div className="text-gray-900 dark:text-gray-100">
                                {show(Number(result.cv!.sep[selectedIndex]))}
                            </div>
                        </div>
                        <div className="col-span-2">
                            <div className="text-gray-500 dark:text-gray-400">Design</div>
                            <div className="text-gray-900 dark:text-gray-100">
                                {result.cv!.design}
                                {result.cv!.group_by ? `, grouped by ${result.cv!.group_by}` : ''}
                            </div>
                        </div>
                    </div>
                )}

                <div className="mt-4 space-y-2 text-xs text-gray-600 dark:text-gray-400">
                    <p>
                        <strong>RMSEC</strong> describes the fit: the model has seen every row it
                        is scored on there, so it is not an estimate of future performance.
                        <strong> RMSECV</strong> is measured on held-out rows and is what chose the
                        component count. <strong>RMSEP</strong> would need a test set kept out of
                        model development entirely, which this screen does not create.
                    </p>
                    {result.excluded_rows && result.excluded_rows.length > 0 && (
                        <p>
                            {result.excluded_rows.length} rows had no measured response. They were
                            left out of the regression but still informed the decomposition, since
                            it does not use the response.
                        </p>
                    )}
                    {hitCeiling && (
                        <p className="text-amber-700 dark:text-amber-400">
                            The search stopped at its ceiling of {result.cv!.selected} components
                            and the error was still falling. Raise the maximum to see whether it
                            keeps improving; this is the end of the range, not a minimum within it.
                        </p>
                    )}
                    {result.cv && result.cv.selected_by_alternate_metric !== result.cv.selected && (
                        <p>
                            The other error measure would have chosen{' '}
                            {result.cv.selected_by_alternate_metric} components. The two disagree
                            when a few large residuals drive the choice.
                        </p>
                    )}
                </div>
            </div>

            {selectionCurve.length > 0 && (
                <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
                    <h3 className="text-base font-semibold text-gray-900 dark:text-gray-100 mb-1">
                        Cross-validated error by component count
                    </h3>
                    <p className="text-xs text-gray-500 dark:text-gray-400 mb-3">
                        {result.components} components selected by the{' '}
                        {result.cv!.rule} rule. Zero components is the intercept-only baseline,
                        which predicts the training mean.
                    </p>
                    <PlotlyLineChart
                        data={selectionCurve}
                        dataKey="y"
                        xDataKey="x"
                        xLabel="Components"
                        yLabel="RMSECV"
                        height={320}
                    />
                </div>
            )}

            <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
                <h3 className="text-base font-semibold text-gray-900 dark:text-gray-100 mb-1">
                    Predicted against measured
                </h3>
                <p className="text-xs text-gray-500 dark:text-gray-400 mb-3">
                    {result.cv
                        ? 'Held-out predictions, so each point comes from a model that did not train on it.'
                        : 'Fitted values. No cross-validation was run, so these come from a model that saw every point.'}
                </p>
                <PlotlyScatterChart
                    data={predictedVsMeasured}
                    xDataKey="x"
                    yDataKey="y"
                    xLabel={`Measured ${result.response}`}
                    yLabel="Predicted"
                    height={360}
                />
            </div>

            <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
                <h3 className="text-base font-semibold text-gray-900 dark:text-gray-100 mb-1">
                    Regression coefficients
                </h3>
                {result.original_scale_valid ? (
                    <>
                        <p className="text-xs text-gray-500 dark:text-gray-400 mb-3">
                            On the original variable scale. With strongly correlated predictors the
                            pattern is one of many near-equivalent solutions and shifts with the
                            component count, so read it for which regions drive the model rather
                            than as an attribution to particular variables.
                        </p>
                        <PlotlyScatterChart
                            data={coefficients}
                            xDataKey="x"
                            yDataKey="y"
                            xLabel="Variable"
                            yLabel="Coefficient"
                            height={320}
                        />
                    </>
                ) : (
                    <p className="text-sm text-gray-600 dark:text-gray-400">
                        Not available. Row-wise preprocessing (SNV or vector normalisation) scales
                        each sample by a statistic of that same sample, so no fixed set of
                        per-variable coefficients reproduces this model&apos;s predictions. The model
                        still predicts correctly through the full pipeline.
                    </p>
                )}
            </div>
        </div>
    );
};
