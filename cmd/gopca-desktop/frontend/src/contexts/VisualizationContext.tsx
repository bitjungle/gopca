// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

import React, { createContext, useContext } from 'react';
import { useVisualization, VisualizationResult } from '../hooks/useVisualization';
import { usePCAContext } from './PCAContext';
import { useFileDataContext } from './FileDataContext';

/**
 * VisualizationContext exposes all plot selection state and helpers.
 *
 * Must be nested inside <PCAProvider> and <FileDataProvider>.
 */
const VisualizationContext = createContext<VisualizationResult | undefined>(undefined);

/**
 * Consume VisualizationContext. Throws if called outside <VisualizationProvider>.
 */
export function useVisualizationContext(): VisualizationResult {
    const ctx = useContext(VisualizationContext);
    if (!ctx) {
        throw new Error('useVisualizationContext must be used within VisualizationProvider');
    }
    return ctx;
}

export const VisualizationProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
    const { pcaResponse, loading, selectedGroupColumn, setExcludedRows } = usePCAContext();
    const { fileData } = useFileDataContext();

    const value = useVisualization(
        pcaResponse,
        fileData,
        loading,
        selectedGroupColumn,
        setExcludedRows,
    );

    return (
        <VisualizationContext.Provider value={value}>
            {children}
        </VisualizationContext.Provider>
    );
};
