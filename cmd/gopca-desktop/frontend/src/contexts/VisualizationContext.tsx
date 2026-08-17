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

import React, { createContext, useContext, useMemo } from 'react';
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

    const {
        selectedPlot, setSelectedPlot,
        selectedXComponent, setSelectedXComponent,
        selectedYComponent, setSelectedYComponent,
        selectedZComponent, setSelectedZComponent,
        selectedLoadingComponent, setSelectedLoadingComponent,
        showEllipses, setShowEllipses,
        confidenceLevel, setConfidenceLevel,
        showRowLabels, setShowRowLabels,
        maxLabelsToShow, setMaxLabelsToShow,
        loadingsPlotType, setLoadingsPlotType,
        plotFontScale, setPlotFontScale,
        getColumnData, handlePlotSelectionChange,
        resetVisualizationSelections, plotPaletteConfig
    } = useVisualization(pcaResponse, fileData, loading, selectedGroupColumn, setExcludedRows);

    // Memoize the context value to avoid creating a new object reference on
    // every provider render (e.g. parent re-renders). This prevents all context
    // consumers from re-rendering when none of these deps have changed.
    // Note: any change to a dep still re-renders ALL consumers — React context
    // does not support per-field subscriptions without context splitting.
    const value = useMemo<VisualizationResult>(() => ({
        selectedPlot, setSelectedPlot,
        selectedXComponent, setSelectedXComponent,
        selectedYComponent, setSelectedYComponent,
        selectedZComponent, setSelectedZComponent,
        selectedLoadingComponent, setSelectedLoadingComponent,
        showEllipses, setShowEllipses,
        confidenceLevel, setConfidenceLevel,
        showRowLabels, setShowRowLabels,
        maxLabelsToShow, setMaxLabelsToShow,
        loadingsPlotType, setLoadingsPlotType,
        plotFontScale, setPlotFontScale,
        getColumnData, handlePlotSelectionChange,
        resetVisualizationSelections, plotPaletteConfig
    }), [
        selectedPlot, setSelectedPlot,
        selectedXComponent, setSelectedXComponent,
        selectedYComponent, setSelectedYComponent,
        selectedZComponent, setSelectedZComponent,
        selectedLoadingComponent, setSelectedLoadingComponent,
        showEllipses, setShowEllipses,
        confidenceLevel, setConfidenceLevel,
        showRowLabels, setShowRowLabels,
        maxLabelsToShow, setMaxLabelsToShow,
        loadingsPlotType, setLoadingsPlotType,
        plotFontScale, setPlotFontScale,
        getColumnData, handlePlotSelectionChange,
        resetVisualizationSelections, plotPaletteConfig
    ]);

    return (
        <VisualizationContext.Provider value={value}>
            {children}
        </VisualizationContext.Provider>
    );
};
