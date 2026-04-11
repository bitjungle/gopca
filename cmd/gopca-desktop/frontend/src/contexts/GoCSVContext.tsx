// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

import React, { createContext, useContext } from 'react';
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
    const value = useGoCSVIntegration();
    return (
        <GoCSVContext.Provider value={value}>
            {children}
        </GoCSVContext.Provider>
    );
};
