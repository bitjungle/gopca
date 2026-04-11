// Copyright 2025-2026 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

import { useState, useCallback, useRef } from 'react';
import { RunPCA } from '../../wailsjs/go/main/App';
import { FileData, PCARequest, PCAResponse } from '../types';
import { PCAConfigState } from './usePCAConfig';

export interface PCARunnerResult {
    pcaResponse: PCAResponse | null;
    pcaError: string | null;
    loading: boolean;
    pcaHasExclusions: boolean;
    pcaResultsRef: React.RefObject<HTMLDivElement>;
    pcaErrorRef: React.RefObject<HTMLDivElement>;
    runPCA: () => Promise<void>;
    clearPcaError: () => void;
    clearPcaResponse: () => void;
}

/**
 * Manages the PCA execution lifecycle: building the request, calling the
 * backend, storing the result or error, and smooth-scrolling to the outcome.
 *
 * Plot auto-switching after a kernel PCA run is handled in AppContent via a
 * useEffect that watches pcaResponse, keeping this hook free of UI concerns.
 */
export function usePCARunner(
    fileData: FileData | null,
    config: PCAConfigState,
    excludedRows: number[],
    excludedColumns: number[],
    selectedGroupColumn: string | null
): PCARunnerResult {
    const [pcaResponse, setPcaResponse] = useState<PCAResponse | null>(null);
    const [pcaError, setPcaError] = useState<string | null>(null);
    const [loading, setLoading] = useState(false);
    const [pcaHasExclusions, setPcaHasExclusions] = useState(false);

    const pcaResultsRef = useRef<HTMLDivElement>(null);
    const pcaErrorRef = useRef<HTMLDivElement>(null);

    const runPCA = useCallback(async () => {
        if (!fileData) return;

        setLoading(true);
        setPcaError(null);

        try {
            const request: PCARequest = {
                data: fileData.data,
                missingMask: fileData.missingMask,
                headers: fileData.headers,
                rowNames: fileData.rowNames,
                ...config,
                excludedRows,
                excludedColumns,
                ...(selectedGroupColumn && fileData.categoricalColumns && {
                    groupColumn: selectedGroupColumn,
                    groupLabels: fileData.categoricalColumns[selectedGroupColumn],
                }),
                metadataNumeric: fileData.numericTargetColumns || {},
                metadataCategorical: fileData.categoricalColumns || {},
                calculateEigencorrelations:
                    (fileData.numericTargetColumns && Object.keys(fileData.numericTargetColumns).length > 0) ||
                    (fileData.categoricalColumns && Object.keys(fileData.categoricalColumns).length > 0),
            };

            const result = await RunPCA(request);

            if (result.success) {
                setPcaResponse(result);
                setPcaError(null);
                setPcaHasExclusions(excludedRows.length > 0);
                setTimeout(() => {
                    pcaResultsRef.current?.scrollIntoView({ behavior: 'smooth', block: 'start' });
                }, 100);
            } else {
                setPcaError(result.error || 'PCA analysis failed');
                setPcaResponse(null);
                setTimeout(() => {
                    pcaErrorRef.current?.scrollIntoView({ behavior: 'smooth', block: 'start' });
                }, 100);
            }
        } catch (err) {
            setPcaError(`Failed to run PCA: ${err}`);
        } finally {
            setLoading(false);
        }
    }, [fileData, config, excludedRows, excludedColumns, selectedGroupColumn]);

    const clearPcaError = useCallback(() => setPcaError(null), []);
    const clearPcaResponse = useCallback(() => setPcaResponse(null), []);

    return {
        pcaResponse,
        pcaError,
        loading,
        pcaHasExclusions,
        pcaResultsRef,
        pcaErrorRef,
        runPCA,
        clearPcaError,
        clearPcaResponse,
    };
}
