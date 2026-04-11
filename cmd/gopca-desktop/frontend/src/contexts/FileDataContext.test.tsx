// Copyright 2025-2026 bitjungle - Rune Mathisen. All rights reserved.
import React from 'react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, act } from '@testing-library/react';
import { FileDataProvider, useFileDataContext } from './FileDataContext';

// ── Helpers ──────────────────────────────────────────────────────────────────

function Consumer() {
    const { fileData, fileName, loading } = useFileDataContext();
    return (
        <div>
            <span data-testid="filename">{fileName || 'none'}</span>
            <span data-testid="loading">{loading ? 'loading' : 'idle'}</span>
            <span data-testid="has-data">{fileData ? 'yes' : 'no'}</span>
        </div>
    );
}

function ThrowingConsumer() {
    // Calling the hook outside the provider should throw
    useFileDataContext();
    return null;
}

// ── Tests ─────────────────────────────────────────────────────────────────────

describe('FileDataContext', () => {
    it('provides initial state to consumers', () => {
        render(
            <FileDataProvider>
                <Consumer />
            </FileDataProvider>
        );

        expect(screen.getByTestId('filename').textContent).toBe('none');
        expect(screen.getByTestId('loading').textContent).toBe('idle');
        expect(screen.getByTestId('has-data').textContent).toBe('no');
    });

    it('throws when consumed outside provider', () => {
        // Suppress React's error boundary output in test logs
        const consoleError = vi.spyOn(console, 'error').mockImplementation(() => {});

        expect(() => render(<ThrowingConsumer />)).toThrow(
            'useFileDataContext must be used within FileDataProvider'
        );

        consoleError.mockRestore();
    });

    it('exposes loadDataset and handleNativeFileSelect functions', () => {
        let ctx: ReturnType<typeof useFileDataContext> | null = null;

        function Capture() {
            ctx = useFileDataContext();
            return null;
        }

        render(
            <FileDataProvider>
                <Capture />
            </FileDataProvider>
        );

        expect(typeof ctx!.loadDataset).toBe('function');
        expect(typeof ctx!.handleNativeFileSelect).toBe('function');
        expect(typeof ctx!.setFileDataDirect).toBe('function');
        expect(typeof ctx!.clearFileError).toBe('function');
    });
});
