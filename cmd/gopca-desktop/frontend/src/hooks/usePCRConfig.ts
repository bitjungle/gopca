// GoPCA Suite
//
// Copyright © 2025-2026 Rune Mathisen <devel@bitjungle.com>
// SPDX-License-Identifier: See LICENSE file for details.

import { useCallback, useState } from 'react';

/**
 * Selection rules for turning a cross-validated error curve into a component count.
 *
 * 'one-se' is the default rather than 'min' because the minimum of a noisy curve
 * is frequently reached by a model far more complex than the data supports.
 */
export type SelectionRule = 'min' | 'one-se' | 'tolerance' | 'wold';

/** Fold layouts. Grouping is orthogonal and set through cvGroupColumn. */
export type CVScheme = 'random' | 'contiguous' | 'forward-chaining';

export interface PCRConfigState {
    /** Numeric #target column to predict. Empty means nothing chosen yet. */
    response: string;

    /**
     * Fixed component count. Zero means choose it by cross-validation, which is
     * the default: retaining components by explained variance is not a valid rule
     * for prediction, because the decomposition never sees the response.
     */
    components: number;

    /** Ceiling for the cross-validation sweep. */
    maxComponents: number;

    /** Folds. Zero means one fold per group, which is leave-one-out. */
    cvFolds: number;
    cvScheme: CVScheme;

    /** Categorical column whose levels must not be split across folds. */
    cvGroupColumn: string;
    cvSeed: number;

    selectRule: SelectionRule;
    metric: 'rmse' | 'mae';
    tolerance: number;
    woldR: number;
}

export const DEFAULT_PCR_CONFIG: PCRConfigState = {
    response: '',
    components: 0,
    maxComponents: 20,
    cvFolds: 10,
    cvScheme: 'random',
    cvGroupColumn: '',
    cvSeed: 42,
    selectRule: 'one-se',
    metric: 'rmse',
    tolerance: 0,
    woldR: 1.0
};

export interface PCRConfigResult {
    config: PCRConfigState;
    setConfig: React.Dispatch<React.SetStateAction<PCRConfigState>>;
    updateConfig: <K extends keyof PCRConfigState>(key: K, value: PCRConfigState[K]) => void;
    resetConfig: () => void;
}

/** usePCRConfig holds the regression settings. */
export function usePCRConfig(): PCRConfigResult {
    const [config, setConfig] = useState<PCRConfigState>(DEFAULT_PCR_CONFIG);

    const updateConfig = useCallback(
        <K extends keyof PCRConfigState>(key: K, value: PCRConfigState[K]) => {
            setConfig(previous => ({ ...previous, [key]: value }));
        },
        []
    );

    const resetConfig = useCallback(() => setConfig(DEFAULT_PCR_CONFIG), []);

    return { config, setConfig, updateConfig, resetConfig };
}
