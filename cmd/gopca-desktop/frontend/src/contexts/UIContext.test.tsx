// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
import React from 'react';
import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { UIProvider, useUIContext } from './UIContext';

function Consumer() {
    const { showDocumentation, showAboutDialog, showCopied } = useUIContext();
    return (
        <div>
            <span data-testid="docs">{showDocumentation ? 'open' : 'closed'}</span>
            <span data-testid="about">{showAboutDialog ? 'open' : 'closed'}</span>
            <span data-testid="copied">{showCopied ? 'copied' : 'idle'}</span>
        </div>
    );
}

function ThrowingConsumer() {
    useUIContext();
    return null;
}

describe('UIContext', () => {
    it('provides initial closed state', () => {
        render(
            <UIProvider>
                <Consumer />
            </UIProvider>
        );

        expect(screen.getByTestId('docs').textContent).toBe('closed');
        expect(screen.getByTestId('about').textContent).toBe('closed');
        expect(screen.getByTestId('copied').textContent).toBe('idle');
    });

    it('setShowDocumentation opens documentation', () => {
        function Toggle() {
            const { showDocumentation, setShowDocumentation } = useUIContext();
            return (
                <button onClick={() => setShowDocumentation(true)} data-testid="toggle">
                    {showDocumentation ? 'open' : 'closed'}
                </button>
            );
        }

        render(<UIProvider><Toggle /></UIProvider>);
        const btn = screen.getByTestId('toggle');
        expect(btn.textContent).toBe('closed');
        fireEvent.click(btn);
        expect(btn.textContent).toBe('open');
    });

    it('handleLogoClick sets showAboutDialog to true', () => {
        function LogoBtn() {
            const { showAboutDialog, handleLogoClick } = useUIContext();
            return (
                <button onClick={handleLogoClick} data-testid="logo">
                    {showAboutDialog ? 'open' : 'closed'}
                </button>
            );
        }

        render(<UIProvider><LogoBtn /></UIProvider>);
        const btn = screen.getByTestId('logo');
        expect(btn.textContent).toBe('closed');
        fireEvent.click(btn);
        expect(btn.textContent).toBe('open');
    });

    it('throws when consumed outside provider', () => {
        const consoleError = vi.spyOn(console, 'error').mockImplementation(() => {});
        expect(() => render(<ThrowingConsumer />)).toThrow(
            'useUIContext must be used within UIProvider'
        );
        consoleError.mockRestore();
    });

    it('exposes mainScrollRef', () => {
        let ctx: ReturnType<typeof useUIContext> | null = null;
        function Capture() { ctx = useUIContext(); return null; }
        render(<UIProvider><Capture /></UIProvider>);
        expect(ctx!.mainScrollRef).toBeDefined();
        // It is a React ref object
        expect(Object.prototype.hasOwnProperty.call(ctx!.mainScrollRef, 'current')).toBe(true);
    });
});
