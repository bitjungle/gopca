// Copyright 2025-2026 bitjungle - Rune Mathisen. All rights reserved.
import React from 'react';
import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { FileDataProvider } from './FileDataContext';
import { PCAProvider, usePCAContext } from './PCAContext';
import { PaletteProvider } from './PaletteContext';

// PCAProvider requires FileDataProvider + PaletteProvider above it.
function AllProviders({ children }: { children: React.ReactNode }) {
    return (
        <PaletteProvider>
            <FileDataProvider>
                <PCAProvider>
                    {children}
                </PCAProvider>
            </FileDataProvider>
        </PaletteProvider>
    );
}

function Consumer() {
    const {
        config, excludedRows, excludedColumns,
        selectedGroupColumn, pcaResponse, pcaError,
        pcaLoading, loading, pcaHasExclusions,
    } = usePCAContext();
    return (
        <div>
            <span data-testid="method">{config.method}</span>
            <span data-testid="components">{config.components}</span>
            <span data-testid="excluded-rows">{excludedRows.length}</span>
            <span data-testid="excluded-cols">{excludedColumns.length}</span>
            <span data-testid="group-col">{selectedGroupColumn ?? 'null'}</span>
            <span data-testid="pca-response">{pcaResponse ? 'has-result' : 'no-result'}</span>
            <span data-testid="pca-error">{pcaError ?? 'no-error'}</span>
            <span data-testid="loading">{loading ? 'loading' : 'idle'}</span>
            <span data-testid="pca-loading">{pcaLoading ? 'running' : 'idle'}</span>
            <span data-testid="exclusions">{pcaHasExclusions ? 'yes' : 'no'}</span>
        </div>
    );
}

function ThrowingConsumer() {
    usePCAContext();
    return null;
}

describe('PCAContext', () => {
    it('provides correct initial state', () => {
        render(<AllProviders><Consumer /></AllProviders>);

        // Default config values from DEFAULT_PCA_CONFIG
        expect(screen.getByTestId('method').textContent).toBe('SVD');
        expect(screen.getByTestId('components').textContent).toBe('5');
        expect(screen.getByTestId('excluded-rows').textContent).toBe('0');
        expect(screen.getByTestId('excluded-cols').textContent).toBe('0');
        expect(screen.getByTestId('group-col').textContent).toBe('null');
        expect(screen.getByTestId('pca-response').textContent).toBe('no-result');
        expect(screen.getByTestId('pca-error').textContent).toBe('no-error');
        expect(screen.getByTestId('loading').textContent).toBe('idle');
        expect(screen.getByTestId('pca-loading').textContent).toBe('idle');
        expect(screen.getByTestId('exclusions').textContent).toBe('no');
    });

    it('exposes all required action handlers', () => {
        let ctx: ReturnType<typeof usePCAContext> | null = null;
        function Capture() { ctx = usePCAContext(); return null; }
        render(<AllProviders><Capture /></AllProviders>);

        expect(typeof ctx!.setConfig).toBe('function');
        expect(typeof ctx!.setExcludedRows).toBe('function');
        expect(typeof ctx!.setExcludedColumns).toBe('function');
        expect(typeof ctx!.resetExclusions).toBe('function');
        expect(typeof ctx!.setSelectedGroupColumn).toBe('function');
        expect(typeof ctx!.runPCA).toBe('function');
        expect(typeof ctx!.clearPcaError).toBe('function');
        expect(typeof ctx!.clearPcaResponse).toBe('function');
        expect(typeof ctx!.handleLoadDataset).toBe('function');
        expect(typeof ctx!.handleNativeFileSelectWithReset).toBe('function');
        expect(typeof ctx!.handleExportModel).toBe('function');
        expect(typeof ctx!.generateCLICommand).toBe('function');
        expect(typeof ctx!.handleRowSelectionChange).toBe('function');
        expect(typeof ctx!.handleColumnSelectionChange).toBe('function');
        expect(typeof ctx!.handleStartupFile).toBe('function');
    });

    it('pcaResultsRef and pcaErrorRef are React refs', () => {
        let ctx: ReturnType<typeof usePCAContext> | null = null;
        function Capture() { ctx = usePCAContext(); return null; }
        render(<AllProviders><Capture /></AllProviders>);

        expect(Object.prototype.hasOwnProperty.call(ctx!.pcaResultsRef, 'current')).toBe(true);
        expect(Object.prototype.hasOwnProperty.call(ctx!.pcaErrorRef, 'current')).toBe(true);
    });

    it('throws when consumed outside provider', () => {
        const consoleError = vi.spyOn(console, 'error').mockImplementation(() => {});
        expect(() => render(<ThrowingConsumer />)).toThrow(
            'usePCAContext must be used within PCAProvider'
        );
        consoleError.mockRestore();
    });

    it('generateCLICommand returns a string', () => {
        let ctx: ReturnType<typeof usePCAContext> | null = null;
        function Capture() { ctx = usePCAContext(); return null; }
        render(<AllProviders><Capture /></AllProviders>);

        const cmd = ctx!.generateCLICommand();
        expect(typeof cmd).toBe('string');
    });
});
