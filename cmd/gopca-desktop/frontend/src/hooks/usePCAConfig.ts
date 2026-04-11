// Copyright 2025-2026 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

import { useState, useCallback } from 'react';
import { FileData } from '../types';

export interface PCAConfigState {
    components: number;
    meanCenter: boolean;
    standardScale: boolean;
    robustScale: boolean;
    scaleOnly: boolean;
    snv: boolean;
    vectorNorm: boolean;
    method: string;
    missingStrategy: string;
    // Kernel PCA
    kernelType: string;
    kernelGamma: number;
    kernelDegree: number;
    kernelCoef0: number;
    // Temporal PCA
    temporalLags: number;
    varianceExplained: number;
}

export const DEFAULT_PCA_CONFIG: PCAConfigState = {
    components: 5,
    meanCenter: true,
    standardScale: false,
    robustScale: false,
    scaleOnly: false,
    snv: false,
    vectorNorm: false,
    method: 'SVD',
    missingStrategy: 'error',
    kernelType: 'rbf',
    kernelGamma: 1.0,
    kernelDegree: 3,
    kernelCoef0: 1.0,
    temporalLags: 10,
    varianceExplained: 0.0,
};

export interface PCAConfigResult {
    config: PCAConfigState;
    setConfig: React.Dispatch<React.SetStateAction<PCAConfigState>>;
    excludedRows: number[];
    excludedColumns: number[];
    setExcludedRows: React.Dispatch<React.SetStateAction<number[]>>;
    setExcludedColumns: React.Dispatch<React.SetStateAction<number[]>>;
    updateGammaForData: (data: FileData) => void;
    resetExclusions: () => void;
}

/**
 * Manages the PCA algorithm configuration and row/column exclusion state.
 *
 * Exclusions live here (not in useVisualization) because they are PCA inputs —
 * changing an exclusion set requires re-running PCA to take effect.
 *
 * The row/column selection callbacks are created in AppContent via useCallback
 * so they can close over both the setters from this hook and the fileData from
 * useFileData, avoiding any circular dependency between hooks.
 */
export function usePCAConfig(): PCAConfigResult {
    const [config, setConfig] = useState<PCAConfigState>(DEFAULT_PCA_CONFIG);
    const [excludedRows, setExcludedRows] = useState<number[]>([]);
    const [excludedColumns, setExcludedColumns] = useState<number[]>([]);

    /** Auto-set gamma and cap component count to data dimensions. */
    const updateGammaForData = useCallback((data: FileData) => {
        if (data?.data?.[0]) {
            const numFeatures = data.data[0].length;
            setConfig(prev => ({
                ...prev,
                kernelGamma: 1.0 / numFeatures,
                components: Math.min(5, numFeatures),
            }));
        }
    }, []);

    const resetExclusions = useCallback(() => {
        setExcludedRows([]);
        setExcludedColumns([]);
    }, []);

    return {
        config,
        setConfig,
        excludedRows,
        excludedColumns,
        setExcludedRows,
        setExcludedColumns,
        updateGammaForData,
        resetExclusions,
    };
}
