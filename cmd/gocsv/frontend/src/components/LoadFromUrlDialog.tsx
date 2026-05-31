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

import React, { useState, useRef } from 'react';
import { PeekRemoteURL, LoadCSV } from '../../wailsjs/go/main/App';
import { main } from '../../wailsjs/go/models';

type FileData = main.FileData;
type URLPeekResult = main.URLPeekResult;

interface LoadFromUrlDialogProps {
    isOpen: boolean;
    onClose: () => void;
    onDataLoaded: (data: FileData) => void;
}

function formatFileSize(bytes: number): string {
    if (bytes < 0) return '';
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(0)} KB`;
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

const LARGE_FILE_THRESHOLD = 10 * 1024 * 1024; // 10 MB

export const LoadFromUrlDialog: React.FC<LoadFromUrlDialogProps> = ({
    isOpen,
    onClose,
    onDataLoaded,
}) => {
    const [url, setUrl] = useState('');
    const [peekResult, setPeekResult] = useState<URLPeekResult | null>(null);
    const [isPeeking, setIsPeeking] = useState(false);
    const [isDownloading, setIsDownloading] = useState(false);
    const inputRef = useRef<HTMLInputElement>(null);

    const handleClose = () => {
        if (isDownloading) return;
        setUrl('');
        setPeekResult(null);
        onClose();
    };

    const handleCheck = async () => {
        const trimmed = url.trim();
        if (!trimmed) return;
        setIsPeeking(true);
        setPeekResult(null);
        try {
            const result = await PeekRemoteURL(trimmed);
            setPeekResult(result);
        } catch (e) {
            setPeekResult({
                url: trimmed,
                fileFormat: '',
                fileSizeBytes: -1,
                accessible: false,
                error: `Unexpected error: ${e}`,
            });
        } finally {
            setIsPeeking(false);
        }
    };

    const handleKeyDown = (e: React.KeyboardEvent) => {
        if (e.key === 'Enter' && !isPeeking && !isDownloading) {
            handleCheck();
        }
    };

    const handleDownload = async () => {
        if (!peekResult?.accessible || !peekResult.url) return;
        setIsDownloading(true);
        try {
            const data = await LoadCSV(peekResult.url);
            onDataLoaded(data);
            handleClose();
        } catch (e) {
            setPeekResult(prev => prev ? {
                ...prev,
                accessible: false,
                error: `Download failed: ${e}`,
            } : null);
        } finally {
            setIsDownloading(false);
        }
    };

    const canDownload = peekResult?.accessible === true &&
        !!peekResult.fileFormat &&
        !isPeeking &&
        !isDownloading;

    const isLargeFile = (peekResult?.fileSizeBytes ?? -1) > LARGE_FILE_THRESHOLD;

    if (!isOpen) return null;

    return (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
            <div className="bg-white dark:bg-gray-800 rounded-lg shadow-xl w-full max-w-lg mx-4">

                {/* Header */}
                <div className="flex items-center justify-between p-4 border-b border-gray-200 dark:border-gray-700">
                    <h2 className="text-lg font-semibold text-gray-900 dark:text-white">
                        Load from URL
                    </h2>
                    <button
                        onClick={handleClose}
                        disabled={isDownloading}
                        className="text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 disabled:opacity-50"
                        aria-label="Close"
                    >
                        <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                        </svg>
                    </button>
                </div>

                {/* Body */}
                <div className="p-4 space-y-4">

                    {/* URL input + Check button */}
                    <div className="flex gap-2">
                        <input
                            ref={inputRef}
                            type="url"
                            value={url}
                            onChange={e => { setUrl(e.target.value); setPeekResult(null); }}
                            onKeyDown={handleKeyDown}
                            placeholder="https://example.com/data.csv"
                            disabled={isDownloading}
                            autoFocus
                            className="flex-1 px-3 py-2 text-sm border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-violet-500 disabled:opacity-50"
                        />
                        <button
                            onClick={handleCheck}
                            disabled={!url.trim() || isPeeking || isDownloading}
                            className="px-4 py-2 text-sm bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-200 border border-gray-300 dark:border-gray-600 rounded-lg hover:bg-gray-200 dark:hover:bg-gray-600 disabled:opacity-50 disabled:cursor-not-allowed transition-colors whitespace-nowrap"
                        >
                            {isPeeking ? 'Checking…' : 'Check URL'}
                        </button>
                    </div>

                    {/* Peek result */}
                    {peekResult && (
                        <div className={`rounded-lg px-4 py-3 text-sm ${
                            peekResult.accessible
                                ? 'bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800'
                                : 'bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800'
                        }`}>
                            {peekResult.accessible ? (
                                <div className="space-y-1">
                                    <div className="flex items-center gap-2 font-medium text-green-800 dark:text-green-200">
                                        <svg className="w-4 h-4 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                                        </svg>
                                        <span className="uppercase tracking-wide">{peekResult.fileFormat}</span>
                                        {peekResult.fileSizeBytes > 0 && (
                                            <span className="font-normal text-green-700 dark:text-green-300">
                                                · {formatFileSize(peekResult.fileSizeBytes)}
                                            </span>
                                        )}
                                    </div>
                                    {isLargeFile && (
                                        <p className="text-yellow-700 dark:text-yellow-300 text-xs">
                                            ⚠ Large file — download may take a moment.
                                        </p>
                                    )}
                                    {peekResult.url !== url.trim() && (
                                        <p className="text-green-600 dark:text-green-400 text-xs">
                                            GitHub link rewritten to raw URL automatically.
                                        </p>
                                    )}
                                </div>
                            ) : (
                                <div className="flex items-start gap-2 text-red-700 dark:text-red-300">
                                    <svg className="w-4 h-4 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                                    </svg>
                                    <span>{peekResult.error}</span>
                                </div>
                            )}
                        </div>
                    )}

                    {/* Download spinner */}
                    {isDownloading && (
                        <div className="flex items-center gap-2 text-sm text-gray-500 dark:text-gray-400">
                            <svg className="animate-spin w-4 h-4 text-violet-500" fill="none" viewBox="0 0 24 24">
                                <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                                <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v8z" />
                            </svg>
                            Downloading…
                        </div>
                    )}

                    {/* Help text */}
                    <p className="text-xs text-gray-500 dark:text-gray-400">
                        Enter a direct file download URL (CSV, TSV, Excel, Parquet).
                        GitHub links are rewritten to raw content URLs automatically.
                    </p>

                </div>

                {/* Footer */}
                <div className="flex justify-end gap-2 p-4 border-t border-gray-200 dark:border-gray-700">
                    <button
                        onClick={handleClose}
                        disabled={isDownloading}
                        className="px-4 py-2 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg disabled:opacity-50 transition-colors"
                    >
                        Cancel
                    </button>
                    <button
                        onClick={handleDownload}
                        disabled={!canDownload}
                        className="px-4 py-2 text-sm text-white bg-violet-600 rounded-lg hover:bg-violet-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                    >
                        {isDownloading ? 'Downloading…' : 'Download and Import'}
                    </button>
                </div>

            </div>
        </div>
    );
};
