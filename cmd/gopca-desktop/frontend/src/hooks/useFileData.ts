// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

import { useState, useCallback } from 'react';
import { LoadDatasetFile, SelectCSVFile } from '../../wailsjs/go/main/App';
import { FileData } from '../types';
import { logger } from '../utils/logger';

export interface FileDataResult {
    fileData: FileData | null;
    fileName: string;
    filePath: string;
    fileError: string | null;
    datasetId: number;
    loading: boolean;
    setFileError: (error: string | null) => void;
    loadDataset: (filename: string, defaultGroupColumn?: string) => Promise<{ data: FileData; defaultGroupColumn?: string } | null>;
    handleNativeFileSelect: () => Promise<FileData | null>;
    setFileDataDirect: (data: FileData, name: string, path: string) => void;
    clearFileError: () => void;
}

/**
 * Manages CSV file loading from disk (native picker), built-in sample datasets,
 * and programmatic setting (e.g. from the startup event).
 *
 * Returns the loaded data plus helpers to trigger loads.  Callers are
 * responsible for resetting dependent state (exclusions, PCA results, etc.)
 * via the return values — load functions return the new FileData so callers
 * can react.
 */
export function useFileData(): FileDataResult {
    const [fileData, setFileData] = useState<FileData | null>(null);
    const [fileName, setFileName] = useState<string>('');
    const [filePath, setFilePath] = useState<string>('');
    const [fileError, setFileError] = useState<string | null>(null);
    const [datasetId, setDatasetId] = useState(0);
    const [loading, setLoading] = useState(false);

    const bumpDatasetId = () => setDatasetId((prev) => prev + 1);

    /** Load a built-in sample dataset by filename. */
    const loadDataset = useCallback(async (
        filename: string,
        defaultGroupColumn?: string
    ): Promise<{ data: FileData; defaultGroupColumn?: string } | null> => {
        setLoading(true);
        setFileError(null);

        try {
            const result = await LoadDatasetFile(filename);
            setFileData(result);
            setFileName(filename);
            setFilePath('');
            bumpDatasetId();
            return { data: result, defaultGroupColumn };
        } catch (err) {
            setFileError(`Failed to load ${filename}: ${err}`);
            return null;
        } finally {
            setLoading(false);
        }
    }, []);

    /** Open the native OS file picker and load the selected CSV. */
    const handleNativeFileSelect = useCallback(async (): Promise<FileData | null> => {
        setLoading(true);
        setFileError(null);

        try {
            const result = await SelectCSVFile();

            if (!result) {
                // User cancelled
                return null;
            }

            const { data, filePath: selectedFilePath } = result;

            if (!data) {
                throw new Error('No data returned from file selection');
            }

            setFileName('Selected File');
            setFilePath(selectedFilePath);
            setFileData(data);
            bumpDatasetId();
            return data;
        } catch (err) {
            logger.error('File selection failed:', err);
            setFileError(`Failed to load file: ${err}`);
            setFileData(null);
            return null;
        } finally {
            setLoading(false);
        }
    }, []);

    /** Set file data directly (used by the startup file-load event). */
    const setFileDataDirect = useCallback((data: FileData, name: string, path: string) => {
        setFileData(data);
        setFileName(name);
        setFilePath(path);
        setFileError(null);
        bumpDatasetId();
    }, []);

    const clearFileError = useCallback(() => setFileError(null), []);

    return {
        fileData,
        fileName,
        filePath,
        fileError,
        datasetId,
        loading,
        setFileError,
        loadDataset,
        handleNativeFileSelect,
        setFileDataDirect,
        clearFileError,
    };
}
