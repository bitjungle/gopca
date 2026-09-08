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

import React from 'react';
import { Copy, Check } from 'lucide-react';
import { CustomSelect } from '@gopca/ui-components';
import { HelpWrapper } from '../index';
import { useFileDataContext } from '../../contexts/FileDataContext';
import { usePCAContext } from '../../contexts/PCAContext';
import { usePCRContext } from '../../contexts/PCRContext';
import { useUIContext } from '../../contexts/UIContext';
import { maxComponentsFor, clampComponentCount } from '../../utils/maxComponents';
import {
    preprocessingPipeline,
    rowStagePipeline,
    columnStagePipeline,
    shouldShowPreview
} from '../../utils/preprocessingPipeline';
import { usePreprocessingPreview } from '../../hooks/usePreprocessingPreview';
import { PreprocessingPreview } from '../PreprocessingPreview';
import {
    validateSavGol,
    snapWindowToOdd,
    savgolSelection,
    applySavGolSelection,
    type SavGolSelection
} from '../../utils/savgolValidation';

interface PCAConfigSectionProps {
    /** Called after runPCA() to reset PC component selectors to 0,1. */
    onRunPCA: () => Promise<void>;
}

/**
 * Step 2: Configure PCA — method, kernel options, preprocessing, Go PCA button,
 * and CLI command preview.
 *
 * `onRunPCA` is passed from AppContent because it must coordinate both PCAContext
 * (runPCA) and VisualizationContext (resetVisualizationSelections), which are
 * sibling consumers — neither can call the other's setters directly.
 */
export function PCAConfigSection({ onRunPCA }: PCAConfigSectionProps) {
    const { fileData } = useFileDataContext();
    const { config, setConfig, loading, generateCLICommand, excludedRows, excludedColumns, variableAxis } = usePCAContext();

    // In Regress mode this panel configures the decomposition the regression is
    // built on, so the preprocessing controls still apply. The Go PCA button does
    // not: it runs a decomposition whose results Regress mode never displays, so
    // pressing it spent the time and discarded the answer. The command preview is
    // kept but rebuilt as `pca regress`, since `pca analyze` would not reproduce
    // what is on screen.
    const { mode, regressionCLIConfig } = usePCRContext();
    const regressing = mode === 'regress';

    // Reported next to the control that caused it, rather than after a run that
    // never had a chance of succeeding. The engine enforces the same rules, so
    // this is about when the user hears, not whether the rule holds.
    const pipeline = preprocessingPipeline(config);

    // Savitzky-Golay fits a polynomial across neighbouring variables, which only
    // means anything if neighbouring variables measure nearly the same thing.
    // The check is offered as a default, not a verdict: `savgolOverride` lets a
    // user who knows their axis is ordered proceed anyway. Withholding the
    // control outright would be indistinguishable, to them, from the feature not
    // existing -- and a null report (no file yet, or the call failed) must not
    // read as a negative answer.
    const [savgolOverride, setSavgolOverride] = React.useState(false);

    // A decision made about one dataset says nothing about the next one.
    React.useEffect(() => { setSavgolOverride(false); }, [fileData]);
    // measurable guards the difference between "these variables are not a
    // continuum" and "nothing could be measured". Only the first is a finding;
    // the second must not disable anything.
    const axisUnsuitable = variableAxis !== null && variableAxis.measurable && !variableAxis.isContinuous;
    const savgolBlocked = axisUnsuitable && !savgolOverride;

    // Shown only when there is a transformation to inspect and an axis to draw it
    // along. On data that is not a continuum a line plot across the variables
    // would be a shape with no meaning, and for plain PCA there is nothing to
    // compare against.
    const previewWorthShowing = shouldShowPreview(Boolean(variableAxis?.isContinuous), config, config.method);
    const preview = usePreprocessingPreview(
        fileData, config, excludedRows, excludedColumns, previewWorthShowing
    );

    const savgolError = validateSavGol(
        config,
        fileData ? fileData.headers.length - excludedColumns.length : 0,
        config.method,
        savgolBlocked
    );
    const commandLine = regressing
        ? generateCLICommand(regressionCLIConfig)
        : generateCLICommand();
    const { showCopied, copyToClipboard } = useUIContext();

    if (!fileData) return null;

    return (
        <div className="bg-white dark:bg-gray-800 rounded-lg p-6 shadow-lg border border-gray-200 dark:border-gray-700">
            <HelpWrapper helpKey={regressing ? 'configure-pca-for-regression' : 'configure-pca'}>
                <h2 className="text-xl font-semibold mb-6">
                    {regressing ? 'Step 2: Configure the PCA Decomposition' : 'Step 2: Configure PCA'}
                </h2>
            </HelpWrapper>

            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
                {/* Left Column - Core PCA Configuration */}
                <div className="space-y-4">
                    <h3 className="text-sm font-semibold text-gray-700 dark:text-gray-300">
                        {regressing ? 'Decomposition Options' : 'PCA Options'}
                    </h3>

                    <div className="p-3 bg-gray-50 dark:bg-gray-700/50 rounded-lg space-y-4">
                        {regressing ? (
                            <p className="text-sm text-gray-500 dark:text-gray-400">
                                The number of components is set in step 3, where it can also be
                                chosen by cross-validation.
                            </p>
                        ) : (
                            <HelpWrapper helpKey="num-components">
                                <label className="block text-sm font-medium mb-2">
                                    Number of Components
                                </label>
                                <input
                                    type="number"
                                    min="1"
                                    max={maxComponentsFor(config.method, fileData.headers.length, fileData.data.length, config.temporalLags)}
                                    value={config.components}
                                    onChange={(e) => setConfig(prev => ({
                                        ...prev,
                                        components: clampComponentCount(
                                            e.target.value,
                                            prev.components,
                                            maxComponentsFor(prev.method, fileData.headers.length, fileData.data.length, prev.temporalLags)
                                        )
                                    }))}
                                    className="w-full px-3 py-2 bg-gray-100 dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-lg text-gray-900 dark:text-white"
                                />
                            </HelpWrapper>
                        )}

                        <HelpWrapper helpKey="pca-method">
                            <label className="block text-sm font-medium mb-2">
                                Method
                            </label>
                            <CustomSelect
                                value={config.method}
                                onChange={(value: string) => {
                                    setConfig(prev => {
                                        const newMethod = value;
                                        const next = { ...prev, method: newMethod };
                                        if (newMethod === 'kernel') {
                                            if (next.meanCenter || next.standardScale || next.robustScale) {
                                                next.meanCenter = false;
                                                next.standardScale = false;
                                                next.robustScale = false;
                                                next.scaleOnly = false;
                                            }
                                        } else if (prev.method === 'kernel' && newMethod !== 'kernel') {
                                            next.meanCenter = true;
                                            next.standardScale = false;
                                            next.robustScale = false;
                                            next.scaleOnly = false;
                                        }
                                        // Methods have different component ceilings (Kernel PCA is
                                        // bounded by samples, the others by variables), so a count
                                        // that was valid before the switch may not be after it.
                                        next.components = Math.min(
                                            next.components,
                                            maxComponentsFor(newMethod, fileData.headers.length, fileData.data.length, next.temporalLags)
                                        );
                                        return next;
                                    });
                                }}
                                options={[
                                    { value: 'SVD', label: 'SVD' },
                                    { value: 'NIPALS', label: 'NIPALS' },
                                    { value: 'kernel', label: 'Kernel PCA' },
                                    { value: 'temporal', label: 'Temporal PCA' }
                                ]}
                                className="w-full"
                            />
                        </HelpWrapper>
                    </div>

                    {config.method === 'SVD' && (
                        <div className="p-4 bg-blue-50 dark:bg-blue-900/20 rounded-lg space-y-3">
                            <h4 className="font-medium text-sm text-blue-900 dark:text-blue-100">SVD Method</h4>
                            <div className="space-y-2 text-sm text-blue-800 dark:text-blue-200">
                                <p className="flex items-start"><span className="mr-2">•</span><span>Gold standard for PCA using Singular Value Decomposition</span></p>
                                <p className="flex items-start"><span className="mr-2">•</span><span>Fast and numerically stable for complete datasets</span></p>
                                <p className="flex items-start"><span className="mr-2">•</span><span>Computes all components simultaneously</span></p>
                                <p className="flex items-start"><span className="mr-2">•</span><span>Best choice for most applications</span></p>
                            </div>
                        </div>
                    )}

                    {config.method === 'NIPALS' && (
                        <div className="p-4 bg-green-50 dark:bg-green-900/20 rounded-lg space-y-3">
                            <h4 className="font-medium text-sm text-green-900 dark:text-green-100">NIPALS Method</h4>
                            <div className="space-y-2 text-sm text-green-800 dark:text-green-200">
                                <p className="flex items-start"><span className="mr-2">•</span><span>Nonlinear Iterative Partial Least Squares algorithm</span></p>
                                <p className="flex items-start"><span className="mr-2">•</span><span>Handles missing data gracefully</span></p>
                                <p className="flex items-start"><span className="mr-2">•</span><span>Computes components sequentially</span></p>
                                <p className="flex items-start"><span className="mr-2">•</span><span>Ideal for large datasets when only few components needed</span></p>
                            </div>
                        </div>
                    )}

                    {config.method === 'kernel' && (
                        <div className="p-4 bg-gray-50 dark:bg-gray-700/50 rounded-lg space-y-4">
                            <h4 className="font-medium text-sm">Kernel PCA Options</h4>

                            {fileData.data.length > 5000 && (
                                <div className="p-3 bg-yellow-100 dark:bg-yellow-900/50 border border-yellow-300 dark:border-yellow-700 rounded-lg">
                                    <p className="text-sm text-yellow-800 dark:text-yellow-200">
                                        <strong>⚠️ Warning:</strong> Kernel PCA with {fileData.data.length.toLocaleString()} samples
                                        may require significant memory (~{Math.round(fileData.data.length * fileData.data.length * 24 / 1024 / 1024 / 1024 * 10) / 10} GB).
                                        {fileData.data.length > 10000 && (
                                            <span className="block mt-1 font-semibold">
                                                Maximum supported: 10,000 samples. Consider using SVD or NIPALS instead.
                                            </span>
                                        )}
                                    </p>
                                </div>
                            )}

                            <div className="space-y-4">
                                <HelpWrapper helpKey="kernel-type">
                                    <label className="block text-sm font-medium mb-1">Kernel Type</label>
                                    <CustomSelect
                                        value={config.kernelType}
                                        onChange={(value: string) => setConfig(prev => ({ ...prev, kernelType: value }))}
                                        options={[
                                            { value: 'rbf', label: 'RBF (Gaussian)' },
                                            { value: 'linear', label: 'Linear' },
                                            { value: 'poly', label: 'Polynomial' }
                                        ]}
                                        className="w-full"
                                    />
                                </HelpWrapper>
                                <HelpWrapper helpKey="kernel-gamma">
                                    <label className="block text-sm font-medium mb-1">Gamma</label>
                                    <input
                                        type="number"
                                        value={config.kernelGamma}
                                        step="0.01"
                                        min="0.001"
                                        onChange={(e) => {
                                            const value = parseFloat(e.target.value);
                                            setConfig(prev => ({ ...prev, kernelGamma: isNaN(value) ? 1.0 : value }));
                                        }}
                                        className="w-full px-3 py-2 bg-gray-100 dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-lg text-gray-900 dark:text-white"
                                    />
                                </HelpWrapper>
                                {config.kernelType === 'poly' && (
                                    <>
                                        <HelpWrapper helpKey="kernel-degree">
                                            <label className="block text-sm font-medium mb-1">Degree</label>
                                            <input
                                                type="number"
                                                value={config.kernelDegree}
                                                min="1"
                                                max="10"
                                                onChange={(e) => setConfig(prev => ({ ...prev, kernelDegree: parseInt(e.target.value) || 3 }))}
                                                className="w-full px-3 py-2 bg-gray-100 dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-lg text-gray-900 dark:text-white"
                                            />
                                        </HelpWrapper>
                                        <HelpWrapper helpKey="kernel-coef0">
                                            <label className="block text-sm font-medium mb-1">Coef0</label>
                                            <input
                                                type="number"
                                                value={config.kernelCoef0}
                                                step="0.1"
                                                onChange={(e) => {
                                                    const value = parseFloat(e.target.value);
                                                    setConfig(prev => ({ ...prev, kernelCoef0: isNaN(value) ? 1.0 : value }));
                                                }}
                                                className="w-full px-3 py-2 bg-gray-100 dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-lg text-gray-900 dark:text-white"
                                            />
                                        </HelpWrapper>
                                    </>
                                )}
                            </div>
                            <p className="text-xs text-gray-500 dark:text-gray-400 mt-2">
                                Note: Kernel PCA uses its own centering in kernel space.
                            </p>
                        </div>
                    )}

                    {config.method === 'temporal' && (
                        <div className="p-4 bg-purple-50 dark:bg-purple-900/20 rounded-lg space-y-4">
                            <h4 className="font-medium text-sm text-purple-900 dark:text-purple-100">Temporal PCA Options</h4>
                            <div className="space-y-2 text-sm text-purple-800 dark:text-purple-200">
                                <p className="flex items-start"><span className="mr-2">•</span><span>Time-Delay PCA for time-series analysis</span></p>
                                <p className="flex items-start"><span className="mr-2">•</span><span>Captures temporal dynamics and dependencies</span></p>
                                <p className="flex items-start"><span className="mr-2">•</span><span>Based on SSA (Singular Spectrum Analysis) methodology</span></p>
                            </div>
                            <div className="space-y-4">
                                <HelpWrapper helpKey="temporal-lags">
                                    <label className="block text-sm font-medium mb-1">Number of Time Lags</label>
                                    <input
                                        type="number"
                                        value={config.temporalLags}
                                        min="2"
                                        max="100"
                                        onChange={(e) => {
                                            const value = parseInt(e.target.value);
                                            setConfig(prev => {
                                                const lags = isNaN(value) || value < 2 ? 2 : value;
                                                return {
                                                    ...prev,
                                                    temporalLags: lags,
                                                    components: Math.min(
                                                        prev.components,
                                                        maxComponentsFor(prev.method, fileData.headers.length, fileData.data.length, lags)
                                                    )
                                                };
                                            });
                                        }}
                                        className="w-full px-3 py-2 bg-gray-100 dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-lg text-gray-900 dark:text-white"
                                    />
                                    <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
                                        Number of time points to include in lag matrix. Use 24 for daily cycles in hourly data, 7 for weekly patterns in daily data.
                                    </p>
                                </HelpWrapper>
                                <div className="p-3 bg-purple-100 dark:bg-purple-900/30 rounded-lg">
                                    <p className="text-xs text-purple-800 dark:text-purple-200">
                                        <strong>Lag Selection Guidelines:</strong><br/>
                                        • Hourly data with daily patterns: L = 24<br/>
                                        • Daily data with weekly patterns: L = 7<br/>
                                        • Monthly data with annual patterns: L = 12<br/>
                                        • General exploration: Start with T/4 (where T = number of samples)
                                    </p>
                                </div>
                                {config.temporalLags >= fileData.data.length && (
                                    <div className="p-3 bg-yellow-100 dark:bg-yellow-900/50 border border-yellow-300 dark:border-yellow-700 rounded-lg">
                                        <p className="text-sm text-yellow-800 dark:text-yellow-200">
                                            <strong>⚠️ Warning:</strong> Number of lags ({config.temporalLags}) should be less than the number of samples ({fileData.data.length}).
                                            Recommended: {Math.floor(fileData.data.length / 4)} lags or less.
                                        </p>
                                    </div>
                                )}
                            </div>
                            <p className="text-xs text-gray-500 dark:text-gray-400 mt-2">
                                Creates a lag matrix where each row contains L consecutive observations, enabling capture of temporal patterns.
                            </p>
                        </div>
                    )}
                </div>

                {/* Right Column - Preprocessing Options */}
                <div className="space-y-4">
                    <h3 className="text-sm font-semibold text-gray-700 dark:text-gray-300">Preprocessing Options</h3>

                    <HelpWrapper helpKey="row-preprocessing" className="p-3 bg-gray-50 dark:bg-gray-700/50 rounded-lg">
                        <label className="block text-sm font-medium mb-2">
                            Row-wise Normalization (optional)
                        </label>
                        <CustomSelect
                            value={config.snv ? 'snv' : config.vectorNorm ? 'vector-norm' : 'none'}
                            onChange={(value: string) => {
                                setConfig(prev => ({ ...prev, snv: value === 'snv', vectorNorm: value === 'vector-norm' }));
                            }}
                            options={[
                                { value: 'none', label: 'None' },
                                { value: 'snv', label: 'SNV (Standard Normal Variate)' },
                                { value: 'vector-norm', label: 'L2 Vector Normalization' }
                            ]}
                            className="w-full"
                        />
                        <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
                            Per sample: removes scatter and overall intensity differences between rows
                        </p>
                    </HelpWrapper>

                    {/* A separate control, not an entry in the row-wise list, because
                        the two are complementary rather than alternatives. A derivative
                        removes any constant, so SNV's mean-subtraction vanishes and
                        SNV-then-derivative reduces exactly to the derivative divided by
                        each sample's spectral spread -- the multiplicative scatter
                        correction a derivative cannot perform. Folding it into the
                        row-wise selector would make the two mutually exclusive and
                        remove the combination that measured best in #843.

                        The controls were once labelled "Step 1/2/3" to convey the order
                        they are applied in. That read as a sequence the user had to work
                        through, so the order is now stated by the summary below, where it
                        describes what is actually selected. */}
                    <HelpWrapper helpKey="savitzky-golay" className="p-3 bg-gray-50 dark:bg-gray-700/50 rounded-lg">
                        <label className="block text-sm font-medium mb-2">
                            Smoothing &amp; Derivatives (optional)
                        </label>
                        <CustomSelect
                            value={savgolSelection(config)}
                            onChange={(value: string) => {
                                setConfig(prev => ({
                                    ...prev,
                                    ...applySavGolSelection(value as SavGolSelection, prev)
                                }));
                            }}
                            options={[
                                { value: 'none', label: 'None' },
                                { value: 'smooth', label: 'Smoothing only' },
                                { value: 'deriv1', label: '1st Derivative', disabled: config.method === 'temporal' },
                                { value: 'deriv2', label: '2nd Derivative', disabled: config.method === 'temporal' }
                            ]}
                            disabled={config.method === 'temporal' || savgolBlocked}
                            className="w-full"
                        />

                        {config.savgolWindow > 0 && (
                            <div className="grid grid-cols-2 gap-2 mt-2">
                                <div>
                                    <label className="block text-xs font-medium mb-1">Window length</label>
                                    <input
                                        type="number"
                                        value={config.savgolWindow}
                                        min={3}
                                        step={2}
                                        onChange={(e) => setConfig(prev => ({
                                            ...prev,
                                            savgolWindow: parseInt(e.target.value, 10) || 0
                                        }))}
                                        onBlur={(e) => setConfig(prev => ({
                                            ...prev,
                                            savgolWindow: snapWindowToOdd(parseInt(e.target.value, 10))
                                        }))}
                                        className="w-full px-2 py-1 text-sm border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-800"
                                    />
                                </div>
                                <div>
                                    <label className="block text-xs font-medium mb-1">Polynomial order</label>
                                    <input
                                        type="number"
                                        value={config.savgolPolyOrder}
                                        min={0}
                                        step={1}
                                        onChange={(e) => setConfig(prev => ({
                                            ...prev,
                                            savgolPolyOrder: parseInt(e.target.value, 10) || 0
                                        }))}
                                        className="w-full px-2 py-1 text-sm border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-800"
                                    />
                                </div>
                            </div>
                        )}

                        {savgolError && (
                            <p className="text-xs text-red-600 dark:text-red-400 mt-1">{savgolError}</p>
                        )}

                        {!savgolError && config.method === 'temporal' && (
                            <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
                                Not available for Temporal PCA, which applies no transform along the variable axis
                            </p>
                        )}

                        {/* The measured reason, not a bare refusal. A scientist can
                            check a number; "unavailable" teaches nothing and cannot
                            be argued with. */}
                        {!savgolError && config.method !== 'temporal' && axisUnsuitable && (
                            <div className="mt-1">
                                <p className="text-xs text-gray-500 dark:text-gray-400">
                                    These {variableAxis.variables} variables do not form a continuum — adjacent
                                    values differ only {variableAxis.smoothnessFactor.toFixed(1)}× less than a
                                    random ordering would. A derivative here would mostly amplify noise.
                                </p>
                                <label className="mt-1 flex items-center gap-1 text-xs text-gray-500 dark:text-gray-400">
                                    <input
                                        type="checkbox"
                                        checked={savgolOverride}
                                        onChange={e => setSavgolOverride(e.target.checked)}
                                    />
                                    Enable anyway — my variables are in a measured order
                                </label>
                            </div>
                        )}

                        {/* Continuous, but the axis may still be broken. The continuity
                            ratio cannot see this: removing a band from the middle of a
                            spectrum leaves the data every bit as smooth. */}
                        {!savgolError && config.method !== 'temporal' && !axisUnsuitable
                            && variableAxis?.namesNumeric && !variableAxis.spacingUniform && (
                            <p className="text-xs text-amber-600 dark:text-amber-400 mt-1">
                                ⚠ The variables are not evenly spaced ({variableAxis.distinctSteps} different step
                                sizes). Savitzky-Golay treats them as equally spaced, so where a gap falls it
                                combines variables that are not really adjacent — excluding columns from the middle
                                of a spectrum does this.
                            </p>
                        )}

                        {!savgolError && config.method !== 'temporal' && !axisUnsuitable
                            && !(variableAxis?.namesNumeric && !variableAxis.spacingUniform) && (
                            <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
                                Smooths along the variables; derivatives remove a baseline offset (1st) or slope (2nd)
                                {variableAxis?.isContinuous && ` — variables look continuous (${variableAxis.smoothnessFactor.toFixed(0)}× smoother than a random ordering)`}
                            </p>
                        )}

                    </HelpWrapper>

                    <HelpWrapper helpKey="column-preprocessing" className="p-3 bg-gray-50 dark:bg-gray-700/50 rounded-lg">
                        <label className="block text-sm font-medium mb-2">
                            Column-wise Scaling
                        </label>
                        <CustomSelect
                            value={
                                config.scaleOnly ? 'scale-only' :
                                config.robustScale ? 'robust' :
                                config.standardScale ? 'standard' :
                                config.meanCenter ? 'center' : 'none'
                            }
                            onChange={(value: string) => {
                                setConfig(prev => {
                                    if (prev.method === 'kernel' && !['none', 'scale-only'].includes(value)) return prev;
                                    return {
                                        ...prev,
                                        meanCenter: value === 'center' || value === 'standard',
                                        standardScale: value === 'standard',
                                        robustScale: value === 'robust',
                                        scaleOnly: value === 'scale-only'
                                    };
                                });
                            }}
                            options={[
                                { value: 'none', label: 'None (Raw Data)' },
                                { value: 'center', label: 'Mean Center Only', disabled: config.method === 'kernel' },
                                { value: 'standard', label: 'Standard Scale (Mean + Std Dev)', disabled: config.method === 'kernel' },
                                { value: 'robust', label: 'Robust Scale (Median + MAD)', disabled: config.method === 'kernel' },
                                { value: 'scale-only', label: 'Variance Scale (Std Dev Only)' }
                            ]}
                            className="w-full"
                        />
                        <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
                            {config.method === 'kernel'
                                ? config.scaleOnly
                                    ? 'Variance scaling divides by standard deviation without centering - suitable for Kernel PCA'
                                    : 'Kernel PCA performs centering in kernel space. Consider Variance Scale if features have different scales.'
                                : 'Normalizes each column/feature across all samples'}
                        </p>
                    </HelpWrapper>

                    {/* States the order the three controls above are applied in, and
                        does it by describing what is selected rather than what could
                        be. Shown always, including when nothing is selected: feeding
                        raw data straight into the decomposition is a choice, and one
                        worth seeing before pressing the button. */}
                    <HelpWrapper helpKey="preprocessing-pipeline" className="p-3 bg-gray-50 dark:bg-gray-700/50 rounded-lg">
                        <div className="text-xs text-gray-600 dark:text-gray-300">
                            <span className="font-medium">Applied in order: </span>
                            {pipeline.length === 0
                                ? <span className="italic">none — the raw data is used as it is</span>
                                : pipeline.map((step, i) => (
                                    <span key={step}>
                                        {i > 0 && <span className="text-gray-400 dark:text-gray-500"> → </span>}
                                        {step}
                                    </span>
                                ))}
                        </div>

                        {/* Belongs here rather than inside the Savitzky-Golay card,
                            which was where it first went. It previews the whole row
                            stage -- scatter correction as well as the filter -- so
                            sitting inside one of the two steps misrepresented what it
                            shows, and made it look as though the filter had to be on
                            for the plot to mean anything. */}
                        {/* Deliberately not gated on savgolError. Most of those errors
                            are a window in the middle of being typed, and unmounting
                            the plot for each intermediate value would blank it exactly
                            while someone is adjusting the number -- which is the
                            opposite of what a live preview is for. The component keeps
                            the last valid curves and says so. */}
                        {preview && (
                            <PreprocessingPreview
                                preview={preview}
                                rowStage={rowStagePipeline(config)}
                                columnStage={columnStagePipeline(config)}
                            />
                        )}
                    </HelpWrapper>

                    <HelpWrapper helpKey="missing-strategy" className="p-3 bg-gray-50 dark:bg-gray-700/50 rounded-lg">
                        <label className="block text-sm font-medium mb-2">
                            Missing Data Strategy
                        </label>
                        <CustomSelect
                            value={config.missingStrategy}
                            onChange={(value: string) => setConfig(prev => ({ ...prev, missingStrategy: value }))}
                            options={[
                                { value: 'error', label: 'Show Error (default)' },
                                { value: 'drop', label: 'Drop Rows with Missing Values' },
                                { value: 'mean', label: 'Impute with Column Mean' },
                                { value: 'median', label: 'Impute with Column Median' },
                                { value: 'zero', label: 'Impute with Zero' },
                                { value: 'native', label: 'Native NIPALS Handling (NIPALS only)' }
                            ]}
                            className="w-full"
                        />
                        <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
                            Choose how to handle missing values (NaN) in your data
                        </p>
                    </HelpWrapper>
                </div>
            </div>

            {/* Go PCA! button, hidden while regressing; Fit regression is the action there */}
            {!regressing && (
                <div className="mt-6 flex justify-center">
                    <HelpWrapper helpKey="go-pca-button">
                        <button
                            onClick={onRunPCA}
                            // Also blocked on a bad filter. Leaving the button live
                            // would make the message beside the control advisory
                            // only, and the run would fail with the same complaint
                            // a few seconds later.
                            disabled={loading || savgolError !== null}
                            title={savgolError ?? undefined}
                            className="px-6 py-2 bg-blue-600 hover:bg-blue-700 disabled:bg-gray-400 dark:disabled:bg-gray-600 rounded-lg font-medium text-white"
                        >
                            {loading ? 'Running...' : 'Go PCA!'}
                        </button>
                    </HelpWrapper>
                </div>
            )}

            {/* CLI Command Preview */}
            <div className="mt-4 bg-gray-900 dark:bg-gray-950 rounded-lg p-4 border border-gray-700">
                <div className="flex items-center justify-between gap-3">
                    <div className="flex items-center gap-3 flex-1">
                        <span className="text-sm font-medium text-gray-300">Command line:</span>
                        <HelpWrapper helpKey="cli-command-preview">
                            <div className="flex-1 bg-black rounded px-3 py-2 font-mono text-xs text-green-400 overflow-x-auto">
                                {commandLine}
                            </div>
                        </HelpWrapper>
                    </div>
                    <button
                        onClick={() => copyToClipboard(commandLine)}
                        className="px-2 py-1 bg-gray-700 hover:bg-gray-600 rounded text-white transition-colors flex-shrink-0"
                        title="Copy command"
                    >
                        {showCopied ? <Check className="w-4 h-4" /> : <Copy className="w-4 h-4" />}
                    </button>
                </div>
            </div>
        </div>
    );
}
