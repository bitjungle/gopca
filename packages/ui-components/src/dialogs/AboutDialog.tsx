// Copyright 2025-2026 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

import React from 'react';
import { Dialog, DialogBody, DialogFooter } from '../components/Dialog';

interface AboutDialogProps {
    isOpen: boolean;
    onClose: () => void;
    version: string;
    appName: string;
    tagline: string;
    logoSrc: string;
    logoAlt: string;
    onHomepageClick: (e: React.MouseEvent) => void;
    onLicenseClick: (e: React.MouseEvent) => void;
}

export const AboutDialog: React.FC<AboutDialogProps> = ({
    isOpen,
    onClose,
    version,
    appName,
    tagline,
    logoSrc,
    logoAlt,
    onHomepageClick,
    onLicenseClick
}) => {
    return (
        <Dialog
            isOpen={isOpen}
            onClose={onClose}
            title={`About ${appName}`}
            width="w-[500px]"
            showCloseButton={true}
        >
            <DialogBody>
                <div className="flex flex-col items-center text-center space-y-4">
                    <img
                        src={logoSrc}
                        alt={logoAlt}
                        className="h-24 w-auto"
                    />

                    <div className="space-y-2">
                        <h2 className="text-2xl font-bold text-gray-900 dark:text-gray-100">
                            {appName}
                        </h2>
                        <p className="text-sm text-gray-600 dark:text-gray-400">
                            Version {version}
                        </p>
                        <p className="text-sm text-gray-700 dark:text-gray-300 italic">
                            {tagline}
                        </p>
                    </div>

                    <div className="border-t border-gray-200 dark:border-gray-700 pt-4 w-full">
                        <p className="text-sm text-gray-700 dark:text-gray-300">
                            © 2025 bitjungle - Rune Mathisen
                        </p>
                        <p className="text-sm text-gray-700 dark:text-gray-300">
                            All rights reserved.
                        </p>
                    </div>

                    <div className="border-t border-gray-200 dark:border-gray-700 pt-4 w-full space-y-2">
                        <p className="text-sm text-gray-700 dark:text-gray-300">
                            Licensed under the{' '}
                            <a
                                href="#"
                                onClick={onLicenseClick}
                                className="text-blue-600 dark:text-blue-400 hover:underline"
                            >
                                MIT License
                            </a>
                        </p>
                        <p className="text-xs text-gray-600 dark:text-gray-400 italic">
                            The author respectfully requests that it not be used for
                            military, warfare, or surveillance applications.
                        </p>
                    </div>

                    <div className="border-t border-gray-200 dark:border-gray-700 pt-4 w-full">
                        <a
                            href="#"
                            onClick={onHomepageClick}
                            className="text-blue-600 dark:text-blue-400 hover:underline text-sm"
                        >
                            https://bitjungle.github.io/gopca/
                        </a>
                    </div>
                </div>
            </DialogBody>

            <DialogFooter>
                <button
                    onClick={onClose}
                    className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 dark:bg-blue-500 dark:hover:bg-blue-600 transition-colors"
                >
                    Close
                </button>
            </DialogFooter>
        </Dialog>
    );
};