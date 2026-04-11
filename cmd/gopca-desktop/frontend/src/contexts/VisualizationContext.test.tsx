// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
import React from 'react';
import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { FileDataProvider } from './FileDataContext';
import { PCAProvider } from './PCAContext';
import { PaletteProvider } from './PaletteContext';
import { VisualizationProvider, useVisualizationContext } from './VisualizationContext';

function AllProviders({ children }: { children: React.ReactNode }) {
    return (
        <PaletteProvider>
            <FileDataProvider>
                <PCAProvider>
                    <VisualizationProvider>
                        {children}
                    </VisualizationProvider>
                </PCAProvider>
            </FileDataProvider>
        </PaletteProvider>
    );
}

function Consumer() {
    const {
        selectedPlot,
        selectedXComponent, selectedYComponent, selectedZComponent,
        selectedLoadingComponent,
        showEllipses, confidenceLevel,
        showRowLabels, maxLabelsToShow,
        loadingsPlotType, plotFontScale,
    } = useVisualizationContext();
    return (
        <div>
            <span data-testid="plot">{selectedPlot}</span>
            <span data-testid="x">{selectedXComponent}</span>
            <span data-testid="y">{selectedYComponent}</span>
            <span data-testid="z">{selectedZComponent}</span>
            <span data-testid="loading-comp">{selectedLoadingComponent}</span>
            <span data-testid="ellipses">{showEllipses ? 'on' : 'off'}</span>
            <span data-testid="confidence">{confidenceLevel}</span>
            <span data-testid="labels">{showRowLabels ? 'on' : 'off'}</span>
            <span data-testid="max-labels">{maxLabelsToShow}</span>
            <span data-testid="loadings-type">{loadingsPlotType ?? 'null'}</span>
            <span data-testid="font-scale">{plotFontScale}</span>
        </div>
    );
}

function ThrowingConsumer() {
    useVisualizationContext();
    return null;
}

describe('VisualizationContext', () => {
    it('provides correct initial state', () => {
        render(<AllProviders><Consumer /></AllProviders>);

        expect(screen.getByTestId('plot').textContent).toBe('scores');
        expect(screen.getByTestId('x').textContent).toBe('0');
        expect(screen.getByTestId('y').textContent).toBe('1');
        expect(screen.getByTestId('z').textContent).toBe('2');
        expect(screen.getByTestId('loading-comp').textContent).toBe('0');
        expect(screen.getByTestId('ellipses').textContent).toBe('off');
        expect(screen.getByTestId('confidence').textContent).toBe('0.95');
        expect(screen.getByTestId('labels').textContent).toBe('off');
        expect(screen.getByTestId('max-labels').textContent).toBe('10');
        expect(screen.getByTestId('loadings-type').textContent).toBe('null');
        expect(screen.getByTestId('font-scale').textContent).toBe('1');
    });

    it('exposes all required setters and handlers', () => {
        let ctx: ReturnType<typeof useVisualizationContext> | null = null;
        function Capture() { ctx = useVisualizationContext(); return null; }
        render(<AllProviders><Capture /></AllProviders>);

        expect(typeof ctx!.setSelectedPlot).toBe('function');
        expect(typeof ctx!.setSelectedXComponent).toBe('function');
        expect(typeof ctx!.setSelectedYComponent).toBe('function');
        expect(typeof ctx!.setSelectedZComponent).toBe('function');
        expect(typeof ctx!.setSelectedLoadingComponent).toBe('function');
        expect(typeof ctx!.setShowEllipses).toBe('function');
        expect(typeof ctx!.setConfidenceLevel).toBe('function');
        expect(typeof ctx!.setShowRowLabels).toBe('function');
        expect(typeof ctx!.setMaxLabelsToShow).toBe('function');
        expect(typeof ctx!.setLoadingsPlotType).toBe('function');
        expect(typeof ctx!.setPlotFontScale).toBe('function');
        expect(typeof ctx!.getColumnData).toBe('function');
        expect(typeof ctx!.handlePlotSelectionChange).toBe('function');
        expect(typeof ctx!.resetVisualizationSelections).toBe('function');
    });

    it('getColumnData returns empty object when no fileData', () => {
        let ctx: ReturnType<typeof useVisualizationContext> | null = null;
        function Capture() { ctx = useVisualizationContext(); return null; }
        render(<AllProviders><Capture /></AllProviders>);

        const result = ctx!.getColumnData('any-column');
        expect(result).toEqual({});
    });

    it('plotPaletteConfig contains all expected plot types', () => {
        let ctx: ReturnType<typeof useVisualizationContext> | null = null;
        function Capture() { ctx = useVisualizationContext(); return null; }
        render(<AllProviders><Capture /></AllProviders>);

        const keys = Object.keys(ctx!.plotPaletteConfig);
        expect(keys).toContain('scores');
        expect(keys).toContain('scree');
        expect(keys).toContain('loadings');
        expect(keys).toContain('biplot');
        expect(keys).toContain('kernel-matrix');
        expect(keys).toContain('eigencorrelation');
    });

    it('throws when consumed outside provider', () => {
        const consoleError = vi.spyOn(console, 'error').mockImplementation(() => {});
        expect(() => render(<ThrowingConsumer />)).toThrow(
            'useVisualizationContext must be used within VisualizationProvider'
        );
        consoleError.mockRestore();
    });
});
