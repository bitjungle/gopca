// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

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
        handleLogoClick, copyToClipboard,
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
        handleLogoClick, copyToClipboard,
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
