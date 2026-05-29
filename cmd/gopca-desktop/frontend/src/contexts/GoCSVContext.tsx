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
import { useGoCSVIntegration, GoCSVIntegrationResult } from '../hooks/useGoCSVIntegration';

/**
 * GoCSVContext exposes the companion app integration state.
 * DataLoadSection uses it to trigger GoCSV actions; AppContent uses it for
 * the download confirmation dialog.
 */
const GoCSVContext = createContext<GoCSVIntegrationResult | undefined>(undefined);

/**
 * Consume GoCSVContext. Throws if called outside <GoCSVProvider>.
 */
export function useGoCSVContext(): GoCSVIntegrationResult {
    const ctx = useContext(GoCSVContext);
    if (!ctx) {
        throw new Error('useGoCSVContext must be used within GoCSVProvider');
    }
    return ctx;
}

export const GoCSVProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
    const {
        goCSVStatus, isCheckingGoCSV, showGoCSVDownloadDialog,
        setShowGoCSVDownloadDialog, handleGoCSVAction, handleGoCSVDownload,
    } = useGoCSVIntegration();

    // Memoize the context value to avoid creating a new object reference on
    // every provider render (e.g. parent re-renders). This prevents all context
    // consumers from re-rendering when none of these deps have changed.
    // Note: any change to a dep still re-renders ALL consumers — React context
    // does not support per-field subscriptions without context splitting.
    const value = useMemo<GoCSVIntegrationResult>(() => ({
        goCSVStatus, isCheckingGoCSV, showGoCSVDownloadDialog,
        setShowGoCSVDownloadDialog, handleGoCSVAction, handleGoCSVDownload,
    }), [goCSVStatus, isCheckingGoCSV, showGoCSVDownloadDialog,
        setShowGoCSVDownloadDialog, handleGoCSVAction, handleGoCSVDownload]);

    return (
        <GoCSVContext.Provider value={value}>
            {children}
        </GoCSVContext.Provider>
    );
};
