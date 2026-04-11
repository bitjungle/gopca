// Copyright 2025-2026 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

import { useState, useCallback, useRef } from 'react';
import { logger } from '../utils/logger';

export interface UIStateResult {
    showDocumentation: boolean;
    setShowDocumentation: (show: boolean) => void;
    showAboutDialog: boolean;
    setShowAboutDialog: (show: boolean) => void;
    showCopied: boolean;
    mainScrollRef: React.RefObject<HTMLDivElement>;
    handleLogoClick: () => void;
    copyToClipboard: (text: string) => Promise<void>;
}

/**
 * Manages pure UI state: modal visibility flags, clipboard feedback,
 * and the main scroll container ref.
 */
export function useUIState(): UIStateResult {
    const [showDocumentation, setShowDocumentation] = useState(false);
    const [showAboutDialog, setShowAboutDialog] = useState(false);
    const [showCopied, setShowCopied] = useState(false);

    const mainScrollRef = useRef<HTMLDivElement>(null);

    const handleLogoClick = useCallback(() => {
        setShowAboutDialog(true);
    }, []);

    const copyToClipboard = useCallback(async (text: string) => {
        try {
            await navigator.clipboard.writeText(text);
            setShowCopied(true);
            setTimeout(() => setShowCopied(false), 2000);
        } catch (err) {
            logger.error('Failed to copy to clipboard:', err);
        }
    }, []);

    return {
        showDocumentation,
        setShowDocumentation,
        showAboutDialog,
        setShowAboutDialog,
        showCopied,
        mainScrollRef,
        handleLogoClick,
        copyToClipboard,
    };
}
