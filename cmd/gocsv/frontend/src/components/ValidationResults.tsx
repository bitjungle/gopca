// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

import React from 'react';

interface ValidationResultsProps {
    isValid: boolean;
    messages: string[];
    onClose: () => void;
}

export const ValidationResults: React.FC<ValidationResultsProps> = ({ isValid, messages, onClose }) => {
    const getMessageStyle = (message: string) => {
        if (message.startsWith('ERROR:')) {
            return 'text-red-600 dark:text-red-400';
        } else if (message.startsWith('WARNING:')) {
            return 'text-yellow-600 dark:text-yellow-400';
        } else if (message.startsWith('INFO:')) {
            return 'text-blue-600 dark:text-blue-400';
        }
        return 'text-gray-600 dark:text-gray-400';
    };

    const getMessageIcon = (message: string) => {
        if (message.startsWith('ERROR:')) {
            return (
                <svg className="w-5 h-5 text-red-600 dark:text-red-400 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
            );
        } else if (message.startsWith('WARNING:')) {
            return (
                <svg className="w-5 h-5 text-yellow-600 dark:text-yellow-400 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                </svg>
            );
        } else if (message.startsWith('INFO:')) {
            return (
                <svg className="w-5 h-5 text-blue-600 dark:text-blue-400 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
            );
        }
        return null;
    };

    const formatMessage = (message: string) => {
        // Remove the prefix for display
        return message.replace(/^(ERROR:|WARNING:|INFO:)\s*/, '');
    };

    return (
        <div className="mt-4 bg-gray-50 dark:bg-gray-700/50 rounded-lg p-4">
            <div className="flex items-center justify-between mb-3">
                <h4 className="text-sm font-semibold text-gray-700 dark:text-gray-300">
                    Validation Results
                </h4>
                <button
                    onClick={onClose}
                    className="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
                >
                    <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M6 18L18 6M6 6l12 12" />
                    </svg>
                </button>
            </div>

            {isValid ? (
                <div className="flex items-center gap-2 text-green-600 dark:text-green-400">
                    <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
                    <span className="text-sm font-medium">Data is valid for GoPCA analysis!</span>
                </div>
            ) : (
                <div className="flex items-center gap-2 text-red-600 dark:text-red-400 mb-3">
                    <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
                    <span className="text-sm font-medium">Data has validation errors that must be fixed before PCA analysis</span>
                </div>
            )}

            {messages.length > 0 && (
                <div className="space-y-2 mt-3 max-h-64 overflow-y-auto">
                    {messages.map((message, index) => (
                        <div key={index} className="flex items-start gap-2">
                            {getMessageIcon(message)}
                            <span className={`text-sm ${getMessageStyle(message)}`}>
                                {formatMessage(message)}
                            </span>
                        </div>
                    ))}
                </div>
            )}
        </div>
    );
};