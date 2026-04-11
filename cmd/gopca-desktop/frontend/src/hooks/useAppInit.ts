// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

import { useState, useEffect } from 'react';
import { GetVersion, GetGUIConfig, SaveFile, CheckGoCSVStatus, LoadCSVFile } from '../../wailsjs/go/main/App';
import { EventsOn } from '../../wailsjs/runtime/runtime';
import { setupPlotlyWailsIntegration } from '@gopca/ui-components';
import { config } from '../../wailsjs/go/models';
import { FileData } from '../types';
import { logger } from '../utils/logger';

export interface AppInitResult {
    version: string;
    guiConfig: config.GUIConfig | null;
}

/**
 * Handles one-time app initialisation: fetches version and GUI config from the
 * backend, wires up the Plotly–Wails export integration, and registers the
 * startup file-load event.
 *
 * @param onStartupFile - called when the app is launched with a file path argument
 */
export function useAppInit(
    onStartupFile: (data: FileData) => void
): AppInitResult {
    const [version, setVersion] = useState<string>('');
    const [guiConfig, setGuiConfig] = useState<config.GUIConfig | null>(null);

    useEffect(() => {
        // Make SaveFile available globally for Plotly export integration
        if (typeof SaveFile !== 'undefined') {
            (window as any).SaveFile = SaveFile;
        }

        setupPlotlyWailsIntegration();

        GetVersion()
            .then(setVersion)
            .catch((err) => logger.error('Failed to get version:', err));

        GetGUIConfig()
            .then(setGuiConfig)
            .catch((err) => logger.error('Failed to get GUI config:', err));

        // Check GoCSV on startup so the button label is correct immediately
        CheckGoCSVStatus().catch((err) =>
            logger.error('Failed to check GoCSV status on startup:', err)
        );

        // A file path may be passed when the app is opened via file association
        const unsubscribe = EventsOn('load-file-on-startup', async (filePath: string) => {
            try {
                const result = await LoadCSVFile(filePath);
                onStartupFile(result);
            } catch (err) {
                logger.error('Failed to load startup file:', err);
            }
        });

        return () => { unsubscribe(); };
        // onStartupFile is intentionally excluded: it changes on every render
        // because it closes over setState, but the handler only needs to be
        // registered once at mount.
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, []);

    return { version, guiConfig };
}
