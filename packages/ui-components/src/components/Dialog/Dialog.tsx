// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

import React, { useEffect, useRef } from 'react';
import { useFocusManagement, useFocusRestore } from '../../hooks/useFocusManagement';
import { useEscapeKey } from '../../hooks/useKeyboardShortcuts';

export interface DialogProps {
    isOpen: boolean;
    onClose: () => void;
    title?: string;
    children: React.ReactNode;
    className?: string;
    width?: string;
    showCloseButton?: boolean;
    closeOnBackdropClick?: boolean;
    closeOnEscape?: boolean;
}

/**
 * Generic Dialog component for modal overlays
 * Provides a reusable dialog structure with backdrop, title, and content area
 */
export const Dialog: React.FC<DialogProps> = ({
    isOpen,
    onClose,
    title,
    children,
    className = '',
    width = 'w-96',
    showCloseButton = false,
    closeOnBackdropClick = true,
    closeOnEscape = true
}) => {
    const dialogRef = useRef<HTMLDivElement>(null);
    const { trapFocus, focusFirst } = useFocusManagement();

    // Save and restore focus when dialog opens/closes
    useFocusRestore();

    // Handle Escape key
    useEscapeKey(() => {
        if (closeOnEscape) {
onClose();
}
    }, isOpen && closeOnEscape);

    useEffect(() => {
        if (isOpen && dialogRef.current) {
            // Focus the dialog and trap focus within it
            focusFirst(dialogRef.current);
            const cleanup = trapFocus(dialogRef.current);

            // Prevent body scroll when dialog is open
            document.body.style.overflow = 'hidden';

            return () => {
                cleanup();
                document.body.style.overflow = 'unset';
            };
        }
    }, [isOpen, trapFocus, focusFirst]);

    if (!isOpen) {
return null;
}

    return (
        <div className="fixed inset-0 z-50 flex items-center justify-center">
            {/* Backdrop */}
            <div
                className="absolute inset-0 bg-black bg-opacity-50"
                onClick={closeOnBackdropClick ? onClose : undefined}
                aria-hidden="true"
            />

            {/* Dialog */}
            <div
                ref={dialogRef}
                className={`relative bg-white dark:bg-gray-800 rounded-lg shadow-xl p-6 ${width} max-w-[90vw] ${className}`}
                role="dialog"
                aria-modal="true"
                aria-labelledby={title ? 'dialog-title' : undefined}
            >
                {/* Header */}
                {(title || showCloseButton) && (
                    <div className="flex justify-between items-center mb-4">
                        {title && (
                            <h3 id="dialog-title" className="text-lg font-semibold text-gray-900 dark:text-gray-100">
                                {title}
                            </h3>
                        )}
                        {showCloseButton && (
                            <button
                                onClick={onClose}
                                className="text-gray-400 hover:text-gray-600 dark:hover:text-gray-200"
                                aria-label="Close dialog"
                            >
                                <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
                                    <path fillRule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clipRule="evenodd" />
                                </svg>
                            </button>
                        )}
                    </div>
                )}

                {/* Content */}
                {children}
            </div>
        </div>
    );
};

/**
 * Dialog footer component for action buttons
 */
export const DialogFooter: React.FC<{ children: React.ReactNode; className?: string }> = ({
    children,
    className = ''
}) => {
    return (
        <div className={`flex justify-end gap-2 mt-4 ${className}`}>
            {children}
        </div>
    );
};

/**
 * Dialog body component for main content
 */
export const DialogBody: React.FC<{ children: React.ReactNode; className?: string }> = ({
    children,
    className = ''
}) => {
    return (
        <div className={`text-gray-700 dark:text-gray-300 ${className}`}>
            {children}
        </div>
    );
};