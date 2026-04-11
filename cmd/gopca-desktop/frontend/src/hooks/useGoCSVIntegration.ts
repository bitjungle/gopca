// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

import { useState, useCallback } from 'react';
import { CheckGoCSVStatus, OpenInGoCSV, LaunchGoCSV, DownloadGoCSV } from '../../wailsjs/go/main/App';
import { FileData } from '../types';
import { logger } from '../utils/logger';

export interface GoCSVStatus {
    installed: boolean;
    path?: string;
    error?: string;
}

export interface GoCSVIntegrationResult {
    goCSVStatus: GoCSVStatus | null;
    isCheckingGoCSV: boolean;
    showGoCSVDownloadDialog: boolean;
    setShowGoCSVDownloadDialog: (show: boolean) => void;
    handleGoCSVAction: (fileData: FileData | null) => Promise<void>;
    handleGoCSVDownload: () => Promise<void>;
}

/**
 * Manages the GoCSV companion app integration: status checks, launching,
 * opening files in GoCSV, and the download confirmation dialog.
 */
export function useGoCSVIntegration(): GoCSVIntegrationResult {
    const [goCSVStatus, setGoCSVStatus] = useState<GoCSVStatus | null>(null);
    const [isCheckingGoCSV, setIsCheckingGoCSV] = useState(false);
    const [showGoCSVDownloadDialog, setShowGoCSVDownloadDialog] = useState(false);

    const handleGoCSVAction = useCallback(async (fileData: FileData | null) => {
        setIsCheckingGoCSV(true);

        try {
            const status = await CheckGoCSVStatus();
            setGoCSVStatus(status);

            if (status.installed) {
                if (fileData) {
                    await OpenInGoCSV(fileData);
                } else {
                    await LaunchGoCSV();
                }
            } else {
                setShowGoCSVDownloadDialog(true);
            }
        } catch (err) {
            logger.error('GoCSV action failed:', err);
            alert(`Failed to perform GoCSV action: ${err}`);
        } finally {
            setIsCheckingGoCSV(false);
        }
    }, []);

    const handleGoCSVDownload = useCallback(async () => {
        setShowGoCSVDownloadDialog(false);
        try {
            await DownloadGoCSV();
        } catch (error) {
            logger.error('Error downloading GoCSV:', error);
            alert('Failed to open download page: ' + error);
        }
    }, []);

    return {
        goCSVStatus,
        isCheckingGoCSV,
        showGoCSVDownloadDialog,
        setShowGoCSVDownloadDialog,
        handleGoCSVAction,
        handleGoCSVDownload,
    };
}
