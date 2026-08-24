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

import { useState, useCallback, useRef, useEffect } from 'react';
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
    const copyTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

    // Clear any pending timeout when the component unmounts.
    useEffect(() => {
        return () => {
            if (copyTimeoutRef.current !== null) clearTimeout(copyTimeoutRef.current);
        };
    }, []);

    const handleLogoClick = useCallback(() => {
        setShowAboutDialog(true);
    }, []);

    const copyToClipboard = useCallback(async (text: string) => {
        try {
            await navigator.clipboard.writeText(text);
            setShowCopied(true);
            // Cancel any previous timeout so rapid clicks don't stack.
            if (copyTimeoutRef.current !== null) clearTimeout(copyTimeoutRef.current);
            copyTimeoutRef.current = setTimeout(() => {
                setShowCopied(false);
                copyTimeoutRef.current = null;
            }, 2000);
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
        copyToClipboard
    };
}
