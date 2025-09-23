// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// CLI command generation for GoPCA Desktop

import { optimizeToRanges } from './rangeOptimizer';

export interface CLIConfig {
    fileName?: string;
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
}

/**
 * Generates a CLI command string based on the configuration
 * @param config The PCA configuration
 * @returns The CLI command string
 */
export function generateCLICommand(config: CLIConfig): string {
    let cmd = 'pca analyze';

    // Add file path (with quotes if it contains spaces)
    if (config.fileName) {
        if (config.fileName.includes(' ')) {
            cmd += ` "${config.fileName}"`;
        } else {
            cmd += ` ${config.fileName}`;
        }
    }

    // Add number of components
    cmd += ` --components ${config.components}`;

    // Add method
    cmd += ` --method ${config.method.toLowerCase()}`;

    // Add kernel parameters if using kernel PCA
    if (config.method === 'kernel') {
        cmd += ` --kernel-type ${config.kernelType}`;
        if (config.kernelType === 'rbf' || config.kernelType === 'laplacian' || config.kernelType === 'sigmoid') {
            cmd += ` --kernel-gamma ${config.kernelGamma}`;
        }
        if (config.kernelType === 'polynomial' || config.kernelType === 'sigmoid') {
            cmd += ` --kernel-degree ${config.kernelDegree}`;
            cmd += ` --kernel-coef0 ${config.kernelCoef0}`;
        }
    }

    // Add temporal parameters if using temporal PCA
    if (config.method === 'temporal') {
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
        cmd += ` --exclude-cols ${rangeStr}`;
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