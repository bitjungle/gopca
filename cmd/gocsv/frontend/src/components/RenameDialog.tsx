// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

import React, { useState, useEffect, useRef } from 'react';

interface RenameDialogProps {
    isOpen: boolean;
    onClose: () => void;
    onRename: (newName: string) => void;
    currentName: string;
    title?: string;
}

export const RenameDialog: React.FC<RenameDialogProps> = ({
    isOpen,
    onClose,
    onRename,
    currentName,
    title = 'Rename Column'
}) => {
    const [newName, setNewName] = useState(currentName);
    const inputRef = useRef<HTMLInputElement>(null);

    useEffect(() => {
        if (isOpen) {
            setNewName(currentName);
            // Focus input after a short delay to ensure modal is rendered
            setTimeout(() => {
                inputRef.current?.focus();
                inputRef.current?.select();
            }, 100);
        }
    }, [isOpen, currentName]);

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        if (newName.trim() && newName !== currentName) {
            onRename(newName.trim());
            onClose();
        }
    };

    const handleKeyDown = (e: React.KeyboardEvent) => {
        if (e.key === 'Escape') {
            onClose();
        }
    };

    if (!isOpen) {
return null;
}

    return (
        <div className="fixed inset-0 z-50 flex items-center justify-center">
            {/* Backdrop */}
            <div
                className="absolute inset-0 bg-black bg-opacity-50"
                onClick={onClose}
            />

            {/* Dialog */}
            <div className="relative bg-white dark:bg-gray-800 rounded-lg shadow-xl p-6 w-96 max-w-[90vw]">
                <h3 className="text-lg font-semibold mb-4 text-gray-900 dark:text-gray-100">
                    {title}
                </h3>

                <form onSubmit={handleSubmit}>
                    <input
                        ref={inputRef}
                        type="text"
                        value={newName}
                        onChange={(e) => setNewName(e.target.value)}
                        onKeyDown={handleKeyDown}
                        className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500"
                        placeholder="Enter new name"
                    />

                    <div className="flex justify-end gap-2 mt-4">
                        <button
                            type="button"
                            onClick={onClose}
                            className="px-4 py-2 text-gray-600 dark:text-gray-400 hover:text-gray-800 dark:hover:text-gray-200"
                        >
                            Cancel
                        </button>
                        <button
                            type="submit"
                            disabled={!newName.trim() || newName === currentName}
                            className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
                        >
                            Rename
                        </button>
                    </div>
                </form>
            </div>
        </div>
    );
};