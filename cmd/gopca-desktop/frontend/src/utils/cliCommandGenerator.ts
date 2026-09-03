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

// CLI command generation for GoPCA Desktop

import { optimizeToRanges } from './rangeOptimizer';

export interface CLIConfig {
    fileName?: string;
    filePath?: string;
    components: number;
    method: string;
    kernelType?: string;
    kernelGamma?: number;
    kernelDegree?: number;
    kernelCoef0?: number;
    temporalLags?: number;
    varianceExplained?: number;
    snv?: boolean;
    vectorNorm?: boolean;
    standardScale?: boolean;
    robustScale?: boolean;
    meanCenter?: boolean;
    scaleOnly?: boolean;
    missingStrategy?: string;
    excludedColumns?: number[];
    excludedRows?: number[];

    /**
     * Present when the command should fit a regression rather than run a plain
     * decomposition. The predictor-side settings above apply either way, since
     * both modes preprocess and decompose identically.
     */
    regression?: RegressionCLIConfig;
}

export interface RegressionCLIConfig {
    response: string;
    /** Zero means the count is chosen by cross-validation. */
    components: number;
    maxComponents: number;
    /** Zero means one fold per group, which is leave-one-out. */
    cvFolds: number;
    cvScheme: string;
    cvGroupColumn: string;
    cvSeed: number;
    selectRule: string;
    metric: string;
    tolerance: number;
    woldR: number;
}

/**
 * Generates a CLI command string based on the configuration
 * @param config The PCA configuration
 * @returns The CLI command string
 */
export function generateCLICommand(config: CLIConfig): string {
    const regression = config.regression;
    let cmd = regression ? 'pca regress' : 'pca analyze';

    // Use filePath if available (user file), otherwise fileName (built-in dataset)
    // For built-in datasets, fileName is just for illustration
    const pathToUse = config.filePath || config.fileName;

    // Add file path (with quotes if it contains spaces)
    if (pathToUse) {
        if (pathToUse.includes(' ')) {
            cmd += ` "${pathToUse}"`;
        } else {
            cmd += ` ${pathToUse}`;
        }
    }

    if (regression) {
        cmd += ` --response "${regression.response}"`;
        if (regression.components > 0) {
            cmd += ` --components ${regression.components}`;
        } else {
            cmd += ` --max-components ${regression.maxComponents}`;
            cmd += regression.cvFolds === 0 ? ' --cv loo' : ` --cv ${regression.cvFolds}`;
            if (regression.cvScheme && regression.cvScheme !== 'random') {
                cmd += ` --cv-scheme ${regression.cvScheme}`;
            }
            if (regression.cvGroupColumn) {
                cmd += ` --cv-group "${regression.cvGroupColumn}"`;
            }
            if (regression.cvSeed !== 42) {
                cmd += ` --cv-seed ${regression.cvSeed}`;
            }
            if (regression.selectRule && regression.selectRule !== 'one-se') {
                cmd += ` --select ${regression.selectRule}`;
            }
            if (regression.selectRule === 'tolerance') {
                cmd += ` --tolerance ${regression.tolerance}`;
            }
            if (regression.selectRule === 'wold' && regression.woldR !== 1) {
                cmd += ` --wold-r ${regression.woldR}`;
            }
            if (regression.metric && regression.metric !== 'rmse') {
                cmd += ` --metric ${regression.metric}`;
            }
        }
    } else {
        cmd += ` --components ${config.components}`;
    }

    // Add method
    cmd += ` --method ${config.method.toLowerCase()}`;

    // Add kernel parameters if using kernel PCA
    if (!regression && config.method === 'kernel') {
        cmd += ` --kernel-type ${config.kernelType}`;
        if (config.kernelType === 'rbf') {
            cmd += ` --kernel-gamma ${config.kernelGamma}`;
        }
        if (config.kernelType === 'polynomial' || config.kernelType === 'poly') {
            cmd += ` --kernel-gamma ${config.kernelGamma}`;
            cmd += ` --kernel-degree ${config.kernelDegree}`;
            cmd += ` --kernel-coef0 ${config.kernelCoef0}`;
        }
    }

    // Add temporal parameters if using temporal PCA
    if (!regression && config.method === 'temporal') {
        cmd += ` --temporal-lags ${config.temporalLags}`;
        if (config.varianceExplained && config.varianceExplained > 0) {
            cmd += ` --var-explained ${config.varianceExplained}`;
        }
    }

    // Add row preprocessing (Step 1)
    if (config.snv) {
        cmd += ' --snv';
    } else if (config.vectorNorm) {
        cmd += ' --vector-norm';
    }

    // Add column preprocessing (Step 2)
    if (config.standardScale) {
        cmd += ' --scale standard';
    } else if (config.robustScale) {
        cmd += ' --scale robust';
    } else if (!config.meanCenter) {
        // Mean centering is on by default in CLI, so we need to explicitly disable it
        cmd += ' --no-mean-centering';
    }
    // Note: if only meanCenter is true, we don't need any flag (it's the default)

    // Add scale-only flag if needed
    if (config.scaleOnly) {
        cmd += ' --scale-only';
    }

    // Add missing data strategy
    if (config.missingStrategy && config.missingStrategy !== 'error') {
        cmd += ` --missing-strategy ${config.missingStrategy}`;
    }

    // Add excluded columns if any
    if (config.excludedColumns && config.excludedColumns.length > 0) {
        // Convert 0-indexed to 1-indexed for CLI
        const columnIndices = config.excludedColumns.map(c => c + 1);
        const rangeStr = optimizeToRanges(columnIndices);
        cmd += ` --exclude-columns ${rangeStr}`;
    }

    // Add excluded rows if any
    if (config.excludedRows && config.excludedRows.length > 0) {
        // Convert 0-indexed to 1-indexed for CLI
        const rowIndices = config.excludedRows.map(r => r + 1);
        const rangeStr = optimizeToRanges(rowIndices);
        cmd += ` --exclude-rows ${rangeStr}`;
    }

    return cmd;
}