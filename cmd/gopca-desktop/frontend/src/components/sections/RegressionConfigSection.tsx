// GoPCA Suite
//
// Copyright © 2025-2026 Rune Mathisen <devel@bitjungle.com>
// SPDX-License-Identifier: See LICENSE file for details.

import React from 'react';
import { HelpWrapper } from '@gopca/ui-components';
import { usePCRContext } from '../../contexts/PCRContext';
import { useFileDataContext } from '../../contexts/FileDataContext';

/**
 * RegressionConfigSection collects everything specific to predicting a response.
 *
 * Preprocessing and the PCA method are deliberately absent: they are set once in
 * the shared configuration panel and apply to both modes, so there is no second
 * set of controls to drift out of step with the first.
 */
export const RegressionConfigSection: React.FC = () => {
    const {
        config, updateConfig, runPCR, loading, pcrError,
        availableResponses, categoricalTargets, groupingColumns
    } = usePCRContext();
    const { fileData } = useFileDataContext();

    if (!fileData) {
        return null;
    }

    const noResponses = availableResponses.length === 0;
    const usingCrossValidation = config.components === 0;

    return (
        <div className="bg-white dark:bg-gray-800 rounded-lg p-6 shadow-lg border border-gray-200 dark:border-gray-700 space-y-4">
            <h2 className="text-xl font-semibold mb-6">Step 3: Regression</h2>

            {noResponses ? (
                <div className="text-sm text-gray-600 dark:text-gray-400 space-y-2">
                    <p>
                        This file has no numeric column marked with <code>#target</code>, so
                        there is nothing to predict. Mark the response column with a
                        <code> #target</code> suffix in its header.
                    </p>
                    {categoricalTargets.length > 0 && (
                        <p>
                            {categoricalTargets.join(', ')}{' '}
                            {categoricalTargets.length === 1 ? 'is' : 'are'} marked as a target
                            but {categoricalTargets.length === 1 ? 'holds' : 'hold'} categories
                            rather than numbers. Predicting a category is classification, which
                            GoPCA does not do.
                        </p>
                    )}
                </div>
            ) : (
                <>
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                        <HelpWrapper helpKey="pcr-response">
                            <label className="block">
                                <span className="text-sm font-medium text-gray-700 dark:text-gray-300">
                                    Response
                                </span>
                                <select
                                    value={config.response}
                                    onChange={e => updateConfig('response', e.target.value)}
                                    className="mt-1 block w-full rounded-md border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-gray-100 text-sm"
                                >
                                    <option value="">Choose a column…</option>
                                    {availableResponses.map(name => (
                                        <option key={name} value={name}>{name}</option>
                                    ))}
                                </select>
                            </label>
                        </HelpWrapper>

                        <HelpWrapper helpKey="pcr-components">
                            <label className="block">
                                <span className="text-sm font-medium text-gray-700 dark:text-gray-300">
                                    Components
                                </span>
                                <div className="mt-1 flex items-center gap-2">
                                    <select
                                        value={usingCrossValidation ? 'cv' : 'fixed'}
                                        onChange={e =>
                                            updateConfig('components', e.target.value === 'cv' ? 0 : 5)
                                        }
                                        className="rounded-md border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-gray-100 text-sm"
                                    >
                                        <option value="cv">Chosen by cross-validation</option>
                                        <option value="fixed">Fixed</option>
                                    </select>
                                    <input
                                        type="number"
                                        min={usingCrossValidation ? 1 : 0}
                                        value={usingCrossValidation ? config.maxComponents : config.components}
                                        onChange={e =>
                                            updateConfig(
                                                usingCrossValidation ? 'maxComponents' : 'components',
                                                Math.max(0, parseInt(e.target.value, 10) || 0)
                                            )
                                        }
                                        className="w-24 rounded-md border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-gray-100 text-sm"
                                        aria-label={usingCrossValidation ? 'Maximum components' : 'Components'}
                                    />
                                </div>
                            </label>
                        </HelpWrapper>
                    </div>

                    {usingCrossValidation && (
                        <div className="grid grid-cols-1 md:grid-cols-3 gap-4 pt-2 border-t border-gray-200 dark:border-gray-700">
                            <HelpWrapper helpKey="pcr-cv-folds">
                                <label className="block">
                                    <span className="text-sm font-medium text-gray-700 dark:text-gray-300">
                                        Folds
                                    </span>
                                    <select
                                        value={config.cvFolds}
                                        onChange={e => updateConfig('cvFolds', parseInt(e.target.value, 10))}
                                        className="mt-1 block w-full rounded-md border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-gray-100 text-sm"
                                    >
                                        <option value={5}>5</option>
                                        <option value={10}>10</option>
                                        <option value={20}>20</option>
                                        <option value={0}>Leave one out</option>
                                    </select>
                                </label>
                            </HelpWrapper>

                            <HelpWrapper helpKey="pcr-cv-scheme">
                                <label className="block">
                                    <span className="text-sm font-medium text-gray-700 dark:text-gray-300">
                                        Fold layout
                                    </span>
                                    <select
                                        value={config.cvScheme}
                                        onChange={e => updateConfig('cvScheme', e.target.value as never)}
                                        className="mt-1 block w-full rounded-md border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-gray-100 text-sm"
                                    >
                                        <option value="random">Random</option>
                                        <option value="contiguous">Contiguous blocks</option>
                                        <option value="forward-chaining">Forward chaining</option>
                                    </select>
                                </label>
                            </HelpWrapper>

                            <HelpWrapper helpKey="pcr-cv-group">
                                <label className="block">
                                    <span className="text-sm font-medium text-gray-700 dark:text-gray-300">
                                        Keep together
                                    </span>
                                    <select
                                        value={config.cvGroupColumn}
                                        onChange={e => updateConfig('cvGroupColumn', e.target.value)}
                                        disabled={groupingColumns.length === 0}
                                        className="mt-1 block w-full rounded-md border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-gray-100 text-sm disabled:opacity-50"
                                    >
                                        <option value="">Nothing (split by row)</option>
                                        {groupingColumns.map(name => (
                                            <option key={name} value={name}>{name}</option>
                                        ))}
                                    </select>
                                </label>
                            </HelpWrapper>

                            <HelpWrapper helpKey="pcr-select-rule">
                                <label className="block">
                                    <span className="text-sm font-medium text-gray-700 dark:text-gray-300">
                                        Selection rule
                                    </span>
                                    <select
                                        value={config.selectRule}
                                        onChange={e => updateConfig('selectRule', e.target.value as never)}
                                        className="mt-1 block w-full rounded-md border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-gray-100 text-sm"
                                    >
                                        <option value="one-se">One standard error</option>
                                        <option value="first-min">First minimum</option>
                                        <option value="min">Lowest error</option>
                                        <option value="tolerance">Within a tolerance</option>
                                        <option value="wold">Wold&apos;s R</option>
                                    </select>
                                </label>
                            </HelpWrapper>

                            <HelpWrapper helpKey="pcr-metric">
                                <label className="block">
                                    <span className="text-sm font-medium text-gray-700 dark:text-gray-300">
                                        Metric
                                    </span>
                                    <select
                                        value={config.metric}
                                        onChange={e => updateConfig('metric', e.target.value as never)}
                                        className="mt-1 block w-full rounded-md border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-gray-100 text-sm"
                                    >
                                        <option value="rmse">RMSE</option>
                                        <option value="mae">MAE</option>
                                    </select>
                                </label>
                            </HelpWrapper>

                            {config.selectRule === 'tolerance' && (
                                <label className="block">
                                    <span className="text-sm font-medium text-gray-700 dark:text-gray-300">
                                        Tolerance
                                    </span>
                                    <input
                                        type="number"
                                        step="any"
                                        min={0}
                                        value={config.tolerance}
                                        onChange={e =>
                                            updateConfig('tolerance', parseFloat(e.target.value) || 0)
                                        }
                                        className="mt-1 block w-full rounded-md border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-gray-100 text-sm"
                                    />
                                </label>
                            )}
                        </div>
                    )}

                    {pcrError && (
                        <div className="rounded-md bg-red-50 dark:bg-red-900/30 p-3 text-sm text-red-800 dark:text-red-200">
                            {pcrError}
                        </div>
                    )}

                    <button
                        type="button"
                        onClick={runPCR}
                        disabled={loading || !config.response}
                        className="px-4 py-2 rounded-md bg-blue-600 text-white text-sm font-medium hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                    >
                        {loading ? 'Fitting…' : 'Fit regression'}
                    </button>
                </>
            )}
        </div>
    );
};
