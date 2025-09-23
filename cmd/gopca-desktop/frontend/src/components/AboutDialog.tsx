// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

import React from 'react';
import { AboutDialog as SharedAboutDialog } from '@gopca/ui-components';
import { BrowserOpenURL } from '../../wailsjs/runtime/runtime';
import logo from '../assets/images/GoPCA-logo-1024-transp.png';

interface AboutDialogProps {
    isOpen: boolean;
    onClose: () => void;
    version: string;
}

export const AboutDialog: React.FC<AboutDialogProps> = ({ isOpen, onClose, version }) => {
    const handleHomepageClick = (e: React.MouseEvent) => {
        e.preventDefault();
        BrowserOpenURL('https://bitjungle.github.io/gopca/');
    };

    const handleLicenseClick = (e: React.MouseEvent) => {
        e.preventDefault();
        BrowserOpenURL('https://github.com/bitjungle/gopca/blob/main/LICENSE');
    };

    return (
        <SharedAboutDialog
            isOpen={isOpen}
            onClose={onClose}
            version={version}
            appName="GoPCA Desktop"
            tagline="The definitive Principal Component Analysis toolset"
            logoSrc={logo}
            logoAlt="GoPCA Logo"
            onHomepageClick={handleHomepageClick}
            onLicenseClick={handleLicenseClick}
        />
    );
};