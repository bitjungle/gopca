// GoPCA Suite
//
// Copyright © 2025-2026 Rune Mathisen <devel@bitjungle.com>
// SPDX-License-Identifier: See LICENSE file for details.

import React, { createContext, useContext, useMemo } from 'react';
import { usePCRConfig, PCRConfigState, PCRConfigResult } from '../hooks/usePCRConfig';
import { usePCRRunner, PCRRunnerResult } from '../hooks/usePCRRunner';
import { useFileDataContext } from './FileDataContext';
import { usePCAContext } from './PCAContext';

/**
 * AnalysisMode selects what the configuration and results panels are for.
 *
 * 'explore' is the decomposition on its own, which is what the application has
 * always done. 'regress' adds a response and predicts it from the component
 * scores. The two share a decomposition and every preprocessing option, so the
 * switch changes what is asked of the data rather than how the data is treated.
 */
export type AnalysisMode = 'explore' | 'regress';

export interface PCRContextType extends PCRConfigResult, PCRRunnerResult {
    mode: AnalysisMode;
    setMode: (mode: AnalysisMode) => void;

    /** Numeric #target columns, sorted, which are the columns that can be predicted. */
    availableResponses: string[];

    /**
     * Columns marked as targets that cannot be predicted, so the interface can
     * explain the absence rather than leave the user hunting for a column they
     * deliberately marked.
     */
    categoricalTargets: string[];

    /** Categorical columns usable for keeping replicates together in a fold. */
    groupingColumns: string[];
}

const PCRContext = createContext<PCRContextType | undefined>(undefined);

export function usePCRContext(): PCRContextType {
    const context = useContext(PCRContext);
    if (!context) {
        throw new Error('usePCRContext must be used within a PCRProvider');
    }
    return context;
}

export const PCRProvider: React.FC<{
    mode: AnalysisMode;
    setMode: (mode: AnalysisMode) => void;
    children: React.ReactNode;
}> = ({ mode, setMode, children }) => {
    const { fileData } = useFileDataContext();
    // Exclusions and the predictor-side configuration live in PCAContext and are
    // reused verbatim, so a row excluded in Explore mode stays excluded here and
    // the two modes cannot disagree about which data they are looking at.
    const { config: pcaConfig, excludedRows, excludedColumns } = usePCAContext();
    const configState = usePCRConfig();

    const runner = usePCRRunner(
        fileData,
        pcaConfig as unknown as Record<string, unknown>,
        configState.config,
        excludedRows,
        excludedColumns
    );

    // Sorted, because object key order is not guaranteed and a picker that
    // reshuffles between loads is disorienting.
    const availableResponses = useMemo(
        () => Object.keys(fileData?.numericTargetColumns || {}).sort(),
        [fileData]
    );

    const categoricalTargets = useMemo(
        () =>
            Object.keys(fileData?.categoricalColumns || {})
                .filter(name => name.toLowerCase().endsWith('#target'))
                .sort(),
        [fileData]
    );

    const groupingColumns = useMemo(
        () => Object.keys(fileData?.categoricalColumns || {}).sort(),
        [fileData]
    );

    const value = useMemo<PCRContextType>(
        () => ({
            ...configState,
            ...runner,
            mode,
            setMode,
            availableResponses,
            categoricalTargets,
            groupingColumns
        }),
        [configState, runner, mode, setMode, availableResponses, categoricalTargets, groupingColumns]
    );

    return <PCRContext.Provider value={value}>{children}</PCRContext.Provider>;
};

export type { PCRConfigState };
