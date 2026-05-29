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
