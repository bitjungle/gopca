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
import { useUIState, UIStateResult } from '../hooks/useUIState';

/**
 * UIContext exposes modal visibility, clipboard feedback, and the main scroll
 * ref so sub-components (header, config section) can access them without props.
 */
const UIContext = createContext<UIStateResult | undefined>(undefined);

/**
 * Consume UIContext. Throws if called outside <UIProvider>.
 */
export function useUIContext(): UIStateResult {
    const ctx = useContext(UIContext);
    if (!ctx) {
        throw new Error('useUIContext must be used within UIProvider');
    }
    return ctx;
}

export const UIProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
    const {
        showDocumentation, setShowDocumentation,
        showAboutDialog, setShowAboutDialog,
        showCopied, mainScrollRef,
        handleLogoClick, copyToClipboard
    } = useUIState();

    // Memoize the context value to avoid creating a new object reference on
    // every provider render (e.g. parent re-renders). This prevents all context
    // consumers from re-rendering when none of these deps have changed.
    // Note: any change to a dep still re-renders ALL consumers — React context
    // does not support per-field subscriptions without context splitting.
    const value = useMemo<UIStateResult>(() => ({
        showDocumentation, setShowDocumentation,
        showAboutDialog, setShowAboutDialog,
        showCopied, mainScrollRef,
        handleLogoClick, copyToClipboard
    }), [showDocumentation, setShowDocumentation,
        showAboutDialog, setShowAboutDialog,
        showCopied, mainScrollRef,
        handleLogoClick, copyToClipboard]);

    return (
        <UIContext.Provider value={value}>
            {children}
        </UIContext.Provider>
    );
};
