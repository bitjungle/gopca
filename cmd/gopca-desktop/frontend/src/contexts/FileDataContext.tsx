// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

import React, { createContext, useContext, useMemo } from 'react';
import { useFileData, FileDataResult } from '../hooks/useFileData';

/**
 * FileDataContext exposes all file-loading state and operations to the
 * component tree, eliminating prop drilling for fileData and related helpers.
 *
 * Provider: wrap the subtree that needs file data with <FileDataProvider>.
 * Consumer: call useFileDataContext() in any descendant component.
 */
const FileDataContext = createContext<FileDataResult | undefined>(undefined);

/**
 * Consume FileDataContext. Throws if called outside <FileDataProvider>.
 */
export function useFileDataContext(): FileDataResult {
    const ctx = useContext(FileDataContext);
    if (!ctx) {
        throw new Error('useFileDataContext must be used within FileDataProvider');
    }
    return ctx;
}

export const FileDataProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
    const {
        fileData, fileName, filePath, fileError, datasetId, loading,
        setFileError, loadDataset, handleNativeFileSelect, setFileDataDirect, clearFileError,
    } = useFileData();

    // Memoize the context value so consumers only re-render when state they
    // care about actually changes. Callbacks are stable (useCallback in hook).
    const value = useMemo<FileDataResult>(() => ({
        fileData, fileName, filePath, fileError, datasetId, loading,
        setFileError, loadDataset, handleNativeFileSelect, setFileDataDirect, clearFileError,
    }), [fileData, fileName, filePath, fileError, datasetId, loading,
        setFileError, loadDataset, handleNativeFileSelect, setFileDataDirect, clearFileError]);

    return (
        <FileDataContext.Provider value={value}>
            {children}
        </FileDataContext.Provider>
    );
};
