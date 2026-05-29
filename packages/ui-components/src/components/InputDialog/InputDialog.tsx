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

import React, { useState, useEffect, useRef } from 'react';
import { Dialog, DialogFooter } from '../Dialog';

export interface InputDialogProps {
    isOpen: boolean;
    onClose: () => void;
    onSubmit: (value: string) => void;
    title: string;
    placeholder?: string;
    initialValue?: string;
    submitLabel?: string;
    cancelLabel?: string;
    validator?: (value: string) => boolean;
    inputType?: 'text' | 'number' | 'email' | 'url';
}

/**
 * Input dialog for getting a single text input from the user
 * Commonly used for renaming, creating new items, etc.
 */
export const InputDialog: React.FC<InputDialogProps> = ({
    isOpen,
    onClose,
    onSubmit,
    title,
    placeholder = '',
    initialValue = '',
    submitLabel = 'OK',
    cancelLabel = 'Cancel',
    validator,
    inputType = 'text'
}) => {
    const [value, setValue] = useState(initialValue);
    const inputRef = useRef<HTMLInputElement>(null);

    useEffect(() => {
        if (isOpen) {
            setValue(initialValue);
            // Focus and select input after a short delay to ensure modal is rendered
            setTimeout(() => {
                inputRef.current?.focus();
                inputRef.current?.select();
            }, 100);
        }
    }, [isOpen, initialValue]);

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        const trimmedValue = value.trim();

        if (trimmedValue && (!validator || validator(trimmedValue))) {
            onSubmit(trimmedValue);
            onClose();
        }
    };

    const isValid = () => {
        const trimmedValue = value.trim();
        if (!trimmedValue) {
return false;
}
        if (validator) {
return validator(trimmedValue);
}
        return true;
    };

    return (
        <Dialog isOpen={isOpen} onClose={onClose} title={title}>
            <form onSubmit={handleSubmit}>
                <input
                    ref={inputRef}
                    type={inputType}
                    value={value}
                    onChange={(e) => setValue(e.target.value)}
                    className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500"
                    placeholder={placeholder}
                />

                <DialogFooter>
                    <button
                        type="button"
                        onClick={onClose}
                        className="px-4 py-2 text-gray-600 dark:text-gray-400 hover:text-gray-800 dark:hover:text-gray-200"
                    >
                        {cancelLabel}
                    </button>
                    <button
                        type="submit"
                        disabled={!isValid()}
                        className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
                    >
                        {submitLabel}
                    </button>
                </DialogFooter>
            </form>
        </Dialog>
    );
};