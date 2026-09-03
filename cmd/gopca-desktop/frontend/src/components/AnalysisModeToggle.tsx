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
 * It lives in the header, which is sticky, so the control keeps one position for
 * the whole session. Placed in the scrolling content it moved down the page as
 * the loaded-data panel grew, and a control that changes what the entire screen
 * is for should not have to be hunted for.
 *
 * The two options are stacked rather than side by side, which halves the width
 * the control takes from the header. Stacking constrains the height: the header's
 * content box is 56px, so the two buttons and the surrounding frame have to fit
 * inside that. Small text with a one-unit vertical padding gives 24px per button,
 * which is the smallest target WCAG 2.2 accepts at level AA, and leaves the
 * control no taller than the logo beside it.
 *
 * It remains a segmented control rather than an on/off switch. A switch expresses
 * one feature being enabled, with an implied default. Explore and Regress are
 * peers: neither is the absence of the other, so there is no honest way to label
 * a switch. "Regress: off" would suggest regression is disabled rather than that
 * a different question is being asked, and a screen reader would announce exactly
 * that. Two buttons carrying aria-pressed announce "Explore, pressed" and
 * "Regress, not pressed", which is what is actually true.
 *
 * It is also not icon-only like the theme and documentation buttons. Those use
 * icons because a sun and a book are universally understood; there is no
 * conventional glyph for principal component regression, and an unlabelled one
 * would trade a control that moved for a control nobody can identify.
 */
export const AnalysisModeToggle: React.FC<AnalysisModeToggleProps> = ({
    mode,
    onChange,
    disabled = false
}) => (
    <div
        className="inline-flex flex-col rounded-lg bg-gray-100 dark:bg-gray-900 p-0.5 border border-gray-300 dark:border-gray-600"
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
                        className={`w-full px-2.5 py-1 text-xs font-medium leading-4 rounded
                            transition-colors duration-150
                            focus:outline-none focus-visible:ring-2 focus-visible:ring-blue-500
                            disabled:opacity-50 disabled:cursor-not-allowed
                            ${active
                                ? 'bg-white dark:bg-gray-700 text-blue-700 dark:text-blue-300 shadow-sm'
                                : 'text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-200'
                            }`}
                    >
                        {option.label}
                    </button>
                </HelpWrapper>
            );
        })}
    </div>
);
