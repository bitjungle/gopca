// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

import { useState, useEffect, useCallback } from 'react';
import { usePalette } from '../contexts/PaletteContext';
import { FileData, PCAResponse } from '../types';

export type PlotType =
    | 'scores' | 'scores3d' | 'scree' | 'loadings'
    | 'biplot' | 'biplot3d' | 'correlations'
    | 'diagnostics' | 'eigencorrelation'
    | 'temporal-loadings' | 'temporal-variable-importance'
    | 'kernel-matrix' | 'sample-contributions';

// Which palette mode each plot type uses.
// 'dynamic' means it follows the Color-by column selection.
type PaletteType = 'categorical' | 'continuous' | 'dynamic';

const PLOT_PALETTE_CONFIG: Record<string, { hasPalette: boolean; paletteType: PaletteType }> = {
    'scores':                       { hasPalette: true,  paletteType: 'dynamic' },
    'scores3d':                     { hasPalette: true,  paletteType: 'dynamic' },
    'scree':                        { hasPalette: true,  paletteType: 'categorical' },
    'loadings':                     { hasPalette: true,  paletteType: 'categorical' },
    'biplot':                       { hasPalette: true,  paletteType: 'dynamic' },
    'biplot3d':                     { hasPalette: true,  paletteType: 'dynamic' },
    'correlations':                 { hasPalette: true,  paletteType: 'categorical' },
    'diagnostics':                  { hasPalette: true,  paletteType: 'dynamic' },
    'eigencorrelation':             { hasPalette: true,  paletteType: 'continuous' },
    'temporal-loadings':            { hasPalette: true,  paletteType: 'categorical' },
    'temporal-variable-importance': { hasPalette: true,  paletteType: 'continuous' },
    'kernel-matrix':                { hasPalette: false, paletteType: 'continuous' },
    'sample-contributions':         { hasPalette: true,  paletteType: 'categorical' },
};

export interface VisualizationResult {
    selectedPlot: PlotType;
    setSelectedPlot: (plot: PlotType) => void;
    selectedXComponent: number;
    setSelectedXComponent: (i: number) => void;
    selectedYComponent: number;
    setSelectedYComponent: (i: number) => void;
    selectedZComponent: number;
    setSelectedZComponent: (i: number) => void;
    selectedLoadingComponent: number;
    setSelectedLoadingComponent: (i: number) => void;
    showEllipses: boolean;
    setShowEllipses: (show: boolean) => void;
    confidenceLevel: 0.90 | 0.95 | 0.99;
    setConfidenceLevel: (level: 0.90 | 0.95 | 0.99) => void;
    showRowLabels: boolean;
    setShowRowLabels: (show: boolean) => void;
    maxLabelsToShow: number;
    setMaxLabelsToShow: (n: number) => void;
    loadingsPlotType: 'bar' | 'line' | null;
    setLoadingsPlotType: (type: 'bar' | 'line' | null) => void;
    plotFontScale: number;
    setPlotFontScale: (scale: number) => void;
    getColumnData: (columnName: string | null) => { values?: string[] | number[]; type?: 'categorical' | 'continuous' };
    handlePlotSelectionChange: (indices: number[]) => void;
    resetVisualizationSelections: () => void;
    plotPaletteConfig: typeof PLOT_PALETTE_CONFIG;
}

/**
 * Manages visualization state: active plot, PC selectors, ellipses, labels,
 * and font scale.  Also owns the palette-mode synchronisation effect.
 *
 * selectedGroupColumn is intentionally NOT owned here — it is needed by
 * usePCARunner (to build requests) and by useVisualization (to drive palette
 * mode), so it is lifted into AppContent and passed to both hooks as a param.
 *
 * @param pcaResponse      - current PCA result (may be null)
 * @param fileData         - current file data (may be null)
 * @param loading          - whether a PCA run is in progress
 * @param selectedGroupColumn - lifted state from AppContent
 * @param setExcludedRows  - from usePCAConfig; used for plot-click row toggling
 */
export function useVisualization(
    pcaResponse: PCAResponse | null,
    fileData: FileData | null,
    loading: boolean,
    selectedGroupColumn: string | null,
    setExcludedRows: React.Dispatch<React.SetStateAction<number[]>>
): VisualizationResult {
    const { setMode } = usePalette();

    const [selectedPlot, setSelectedPlot] = useState<PlotType>('scores');
    const [selectedXComponent, setSelectedXComponent] = useState(0);
    const [selectedYComponent, setSelectedYComponent] = useState(1);
    const [selectedZComponent, setSelectedZComponent] = useState(2);
    const [selectedLoadingComponent, setSelectedLoadingComponent] = useState(0);
    const [showEllipses, setShowEllipses] = useState(false);
    const [confidenceLevel, setConfidenceLevel] = useState<0.90 | 0.95 | 0.99>(0.95);
    const [showRowLabels, setShowRowLabels] = useState(false);
    const [maxLabelsToShow, setMaxLabelsToShow] = useState(10);
    const [loadingsPlotType, setLoadingsPlotType] = useState<'bar' | 'line' | null>(null);
    const [plotFontScale, setPlotFontScale] = useState(1.0);

    // Auto-switch away from plots that are incompatible with Kernel PCA
    useEffect(() => {
        if (pcaResponse?.result?.method === 'kernel') {
            if (
                selectedPlot === 'loadings' ||
                selectedPlot === 'diagnostics' ||
                selectedPlot === 'eigencorrelation'
            ) {
                setSelectedPlot('scores');
            }
        }
    }, [pcaResponse, selectedPlot]);

    // Synchronise palette mode with the active plot type
    useEffect(() => {
        if (!pcaResponse?.result) return;

        const plotConfig = PLOT_PALETTE_CONFIG[selectedPlot];
        if (!plotConfig?.hasPalette) {
            setMode('none');
            return;
        }

        if (plotConfig.paletteType === 'categorical') {
            setMode('categorical');
        } else if (plotConfig.paletteType === 'continuous') {
            setMode('continuous');
        } else {
            // Dynamic: follow Color-by selection
            if (!selectedGroupColumn) {
                setMode('categorical');
            }
            // Otherwise the Color-by control manages the mode
        }
    }, [selectedPlot, pcaResponse, setMode, selectedGroupColumn]);

    /** Return the values and type for a group/color column. */
    const getColumnData = useCallback((
        columnName: string | null
    ): { values?: string[] | number[]; type?: 'categorical' | 'continuous' } => {
        if (!columnName || !fileData) return {};

        if (columnName === 'Row Index') {
            const numSamples = pcaResponse?.result?.scores?.length || fileData.data.length;
            return {
                values: Array.from({ length: numSamples }, (_, i) => i + 1),
                type: 'continuous',
            };
        }

        if (pcaResponse?.filteredCategoricalColumns?.[columnName] !== undefined) {
            return { values: pcaResponse.filteredCategoricalColumns[columnName], type: 'categorical' };
        }

        if (pcaResponse?.filteredNumericTargetColumns?.[columnName] !== undefined) {
            return { values: pcaResponse.filteredNumericTargetColumns[columnName], type: 'continuous' };
        }

        if (fileData.categoricalColumns?.[columnName] !== undefined) {
            return { values: fileData.categoricalColumns[columnName], type: 'categorical' };
        }

        if (fileData.numericTargetColumns?.[columnName] !== undefined) {
            return { values: fileData.numericTargetColumns[columnName], type: 'continuous' };
        }

        return {};
    }, [fileData, pcaResponse]);

    /** Toggle row exclusion from a plot click/lasso selection. */
    const handlePlotSelectionChange = useCallback((indices: number[]) => {
        if (loading || !fileData || indices.length === 0) return;

        setExcludedRows(prev => {
            const next = new Set(prev);
            indices.forEach(idx => {
                if (next.has(idx)) { next.delete(idx); } else { next.add(idx); }
            });
            return Array.from(next);
        });
    }, [fileData, loading, setExcludedRows]);

    /** Reset PC selectors to defaults after a new PCA run. */
    const resetVisualizationSelections = useCallback(() => {
        setSelectedXComponent(0);
        setSelectedYComponent(1);
    }, []);

    return {
        selectedPlot,
        setSelectedPlot,
        selectedXComponent,
        setSelectedXComponent,
        selectedYComponent,
        setSelectedYComponent,
        selectedZComponent,
        setSelectedZComponent,
        selectedLoadingComponent,
        setSelectedLoadingComponent,
        showEllipses,
        setShowEllipses,
        confidenceLevel,
        setConfidenceLevel,
        showRowLabels,
        setShowRowLabels,
        maxLabelsToShow,
        setMaxLabelsToShow,
        loadingsPlotType,
        setLoadingsPlotType,
        plotFontScale,
        setPlotFontScale,
        getColumnData,
        handlePlotSelectionChange,
        resetVisualizationSelections,
        plotPaletteConfig: PLOT_PALETTE_CONFIG,
    };
}
