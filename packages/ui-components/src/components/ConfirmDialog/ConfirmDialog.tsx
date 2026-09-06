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

import React, { useEffect, useCallback, useRef } from 'react';
import { useFocusManagement, useFocusRestore } from '../../hooks/useFocusManagement';

export interface ConfirmDialogProps {
  isOpen: boolean;
  onClose: () => void;
  onConfirm: () => void;
  title?: string;
  message: string;
  confirmText?: string;
  cancelText?: string;
  destructive?: boolean;
  confirmButtonClassName?: string;
  cancelButtonClassName?: string;
  containerClassName?: string;
}

export const ConfirmDialog: React.FC<ConfirmDialogProps> = ({
  isOpen,
  onClose,
  onConfirm,
  title = 'Confirm',
  message,
  confirmText = 'Confirm',
  cancelText = 'Cancel',
  destructive = false,
  confirmButtonClassName,
  cancelButtonClassName,
  containerClassName
}) => {
  const handleConfirm = useCallback(() => {
    onConfirm();
    onClose();
  }, [onConfirm, onClose]);

  const handleKeyDown = useCallback((e: KeyboardEvent) => {
    if (!isOpen) {
return;
}

    if (e.key === 'Escape') {
      onClose();
    } else if (e.key === 'Enter' && (e.metaKey || e.ctrlKey)) {
      handleConfirm();
    }
  }, [isOpen, onClose, handleConfirm]);

  useEffect(() => {
    window.addEventListener('keydown', handleKeyDown);
    return () => {
      window.removeEventListener('keydown', handleKeyDown);
    };
  }, [handleKeyDown]);

  if (!isOpen) {
return null;
}

  // Trap focus inside the dialog while it is open, and restore it on close.
  // Without this a keyboard user can Tab straight out of the dialog into the
  // page behind it, which is the same gap #874 covers for GoCSV's dialogs.
  const panelRef = useRef<HTMLDivElement>(null);
  const { trapFocus, focusFirst } = useFocusManagement();
  useFocusRestore();

  useEffect(() => {
    if (!isOpen || !panelRef.current) {
      return;
    }
    focusFirst(panelRef.current);
    const release = trapFocus(panelRef.current);
    return release;
  }, [isOpen, trapFocus, focusFirst]);

  const defaultConfirmClass = destructive
    ? 'bg-red-600 hover:bg-red-700 text-white'
    : 'bg-blue-600 hover:bg-blue-700 text-white';

  const defaultCancelClass = 'text-gray-600 dark:text-gray-400 hover:text-gray-800 dark:hover:text-gray-200';

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center" role="dialog" aria-modal="true">
      <div
        className="absolute inset-0 bg-black bg-opacity-50"
        onClick={onClose}
        aria-hidden="true"
      />

      <div
        ref={panelRef}
        className={containerClassName || 'relative bg-white dark:bg-gray-800 rounded-lg shadow-xl p-6 w-96 max-w-[90vw]'}
      >
        <h3 className="text-lg font-semibold mb-4 text-gray-900 dark:text-gray-100">
          {title}
        </h3>

        <p className="text-gray-700 dark:text-gray-300 mb-6">
          {message}
        </p>

        <div className="flex justify-end gap-2">
          <button
            type="button"
            onClick={onClose}
            className={cancelButtonClassName || `px-4 py-2 ${defaultCancelClass} transition-colors`}
            aria-label={cancelText}
          >
            {cancelText}
          </button>
          <button
            type="button"
            onClick={handleConfirm}
            className={confirmButtonClassName || `px-4 py-2 rounded-md ${defaultConfirmClass} transition-colors`}
            aria-label={confirmText}
          >
            {confirmText}
          </button>
        </div>

        <div className="mt-4 text-xs text-gray-500 dark:text-gray-400">
          <kbd className="px-1 py-0.5 bg-gray-100 dark:bg-gray-700 rounded">Esc</kbd> to cancel,{' '}
          <kbd className="px-1 py-0.5 bg-gray-100 dark:bg-gray-700 rounded">⌘</kbd>+
          <kbd className="px-1 py-0.5 bg-gray-100 dark:bg-gray-700 rounded">Enter</kbd> to confirm
        </div>
      </div>
    </div>
  );
};