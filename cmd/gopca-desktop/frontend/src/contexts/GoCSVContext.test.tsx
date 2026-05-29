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
import { render, screen } from '@testing-library/react';
import { GoCSVProvider, useGoCSVContext } from './GoCSVContext';

function Consumer() {
    const { goCSVStatus, isCheckingGoCSV, showGoCSVDownloadDialog } = useGoCSVContext();
    return (
        <div>
            <span data-testid="status">{goCSVStatus ? 'has-status' : 'no-status'}</span>
            <span data-testid="checking">{isCheckingGoCSV ? 'checking' : 'idle'}</span>
            <span data-testid="dialog">{showGoCSVDownloadDialog ? 'open' : 'closed'}</span>
        </div>
    );
}

function ThrowingConsumer() {
    useGoCSVContext();
    return null;
}

describe('GoCSVContext', () => {
    it('provides initial idle state', () => {
        render(
            <GoCSVProvider>
                <Consumer />
            </GoCSVProvider>
        );

        expect(screen.getByTestId('status').textContent).toBe('no-status');
        expect(screen.getByTestId('checking').textContent).toBe('idle');
        expect(screen.getByTestId('dialog').textContent).toBe('closed');
    });

    it('exposes action handlers', () => {
        let ctx: ReturnType<typeof useGoCSVContext> | null = null;
        function Capture() { ctx = useGoCSVContext(); return null; }

        render(<GoCSVProvider><Capture /></GoCSVProvider>);

        expect(typeof ctx!.handleGoCSVAction).toBe('function');
        expect(typeof ctx!.handleGoCSVDownload).toBe('function');
        expect(typeof ctx!.setShowGoCSVDownloadDialog).toBe('function');
    });

    it('throws when consumed outside provider', () => {
        const consoleError = vi.spyOn(console, 'error').mockImplementation(() => {});
        expect(() => render(<ThrowingConsumer />)).toThrow(
            'useGoCSVContext must be used within GoCSVProvider'
        );
        consoleError.mockRestore();
    });
});
