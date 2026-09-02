// GoPCA Suite
//
// Copyright © 2025-2026 Rune Mathisen <devel@bitjungle.com>
// SPDX-License-Identifier: See LICENSE file for details.

import React from 'react';
import { HelpWrapper } from '@gopca/ui-components';
import { AnalysisMode } from '../contexts/PCRContext';

interface AnalysisModeToggleProps {
    mode: AnalysisMode;
    onChange: (mode: AnalysisMode) => void;
    disabled?: boolean;
}

const MODES: { value: AnalysisMode; label: string; helpKey: string }[] = [
    { value: 'explore', label: 'Explore', helpKey: 'analysis-mode-explore' },
    { value: 'regress', label: 'Regress', helpKey: 'analysis-mode-regress' }
];

/**
 * AnalysisModeToggle switches between exploring a decomposition and regressing a
 * response on it.
 *
 * It is a single switch rather than extra controls scattered through the existing
 * panel, so that nothing changes for someone who only wants PCA. The two modes
 * share the same preprocessing and the same decomposition; the switch changes
 * what is asked of the data, not how the data is treated.
 */
export const AnalysisModeToggle: React.FC<AnalysisModeToggleProps> = ({
    mode,
    onChange,
    disabled = false
}) => (
    <div
        className="inline-flex rounded-lg border border-gray-300 dark:border-gray-600 overflow-hidden"
        role="group"
        aria-label="Analysis mode"
    >
        {MODES.map(option => {
            const active = mode === option.value;
            return (
                <HelpWrapper key={option.value} helpKey={option.helpKey}>
                    <button
                        type="button"
                        onClick={() => onChange(option.value)}
                        disabled={disabled}
                        aria-pressed={active}
                        className={`px-4 py-1.5 text-sm font-medium transition-colors duration-150
                            disabled:opacity-50 disabled:cursor-not-allowed
                            ${active
                                ? 'bg-blue-600 text-white'
                                : 'bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700'
                            }`}
                    >
                        {option.label}
                    </button>
                </HelpWrapper>
            );
        })}
    </div>
);
