// GoPCA Suite
//
// Copyright © 2025-2026 Rune Mathisen <devel@bitjungle.com>
//
// This file is part of GoPCA Suite.
//
// See LICENSE for the full license terms.

import React from 'react';
import { PlotlyWithFullscreen, useTheme } from '@gopca/ui-components';
import { PreprocessingPreview as PreviewData } from '../hooks/usePreprocessingPreview';
import { HelpWrapper } from './index';

interface PreprocessingPreviewProps {
    preview: PreviewData;
    /** The steps these curves have actually been through. */
    rowStage: string[];
    /** What happens after, and is deliberately not drawn. */
    columnStage: string[];
}

/**
 * A small plot of the spectra as the decomposition will see them.
 *
 * Placed beside the preprocessing controls rather than among the result plots
 * because the moment it answers a question is while a window and polynomial
 * order are being chosen — which is before any decomposition has run.
 *
 * Shows the row stage only: scatter correction and Savitzky-Golay, but not
 * column centring. Centred spectra are pulled toward zero by the mean spectrum
 * and are much harder to read, and "the preprocessed spectra" in the
 * spectroscopic sense means this stage.
 */
export function PreprocessingPreview({ preview, rowStage, columnStage }: PreprocessingPreviewProps) {
    const { theme } = useTheme();
    const [showRaw, setShowRaw] = React.useState(false);

    const isDark = theme === 'dark';
    const curves = showRaw ? preview.raw : preview.processed;
    const hasCurves = curves.length > 0 && curves[0]?.length > 0;

    const x = React.useMemo(() => {
        const width = curves[0]?.length ?? 0;
        if (preview.positions && preview.positions.length === width) {
            return preview.positions;
        }
        return Array.from({ length: width }, (_, i) => i);
    }, [curves, preview.positions]);

    const traces = React.useMemo(() => curves.map((row, i) => ({
        x,
        y: row,
        type: 'scattergl' as const,
        mode: 'lines' as const,
        // Thirty overlaid spectra are read as a band, not as individual curves,
        // so they share one muted colour and one legend entry rather than
        // thirty of each. Which sample a line belongs to is not the question
        // this plot answers.
        line: { width: 1, color: showRaw ? '#94a3b8' : '#3b82f6' },
        opacity: 0.55,
        hoverinfo: 'skip' as const,
        showlegend: false,
        name: `sample ${i}`
    })), [curves, x, showRaw]);

    const layout = React.useMemo(() => ({
        autosize: true,
        margin: { l: 44, r: 10, t: 6, b: 32 },
        paper_bgcolor: 'rgba(0,0,0,0)',
        plot_bgcolor: 'rgba(0,0,0,0)',
        font: { size: 10, color: isDark ? '#cbd5e1' : '#475569' },
        xaxis: {
            title: { text: preview.positions ? 'Variable' : 'Variable index', standoff: 6 },
            gridcolor: isDark ? '#334155' : '#e2e8f0',
            zeroline: false
        },
        yaxis: {
            gridcolor: isDark ? '#334155' : '#e2e8f0',
            zerolinecolor: isDark ? '#475569' : '#cbd5e1'
        },
        showlegend: false
    }), [isDark, preview.positions]);

    return (
        <HelpWrapper helpKey="preprocessing-preview" className="mt-2 block">
            <div className="flex items-center justify-between mb-1">
                <span className="text-xs font-medium text-gray-600 dark:text-gray-300">
                    {showRaw ? 'Raw spectra' : 'After preprocessing'}
                </span>
                <button
                    type="button"
                    onClick={() => setShowRaw(v => !v)}
                    className="text-xs px-2 py-0.5 rounded border border-gray-300 dark:border-gray-600 hover:bg-gray-100 dark:hover:bg-gray-700"
                >
                    {showRaw ? 'Show preprocessed' : 'Show raw'}
                </button>
            </div>

            <div style={{ height: 160 }}>
                {hasCurves
                    ? (
                        <PlotlyWithFullscreen
                            data={traces}
                            layout={layout}
                            config={{ displayModeBar: false, responsive: true }}
                            style={{ width: '100%', height: '100%' }}
                        />
                    )
                    : (
                        <div className="h-full flex items-center justify-center text-xs text-gray-400">
                            {preview.loading ? 'Computing…' : 'No preview available'}
                        </div>
                    )}
            </div>

            <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
                {preview.raw.length < preview.totalRows
                    ? `${preview.raw.length} of ${preview.totalRows} samples, spread across the file`
                    : `${preview.raw.length} samples`}
                {!showRaw && rowStage.length > 0 && ` · ${rowStage.join(' → ')}`}
            </p>

            {/* Naming what is not drawn matters as much as naming what is. The
                caption used to list the column step alongside the others, which
                described a picture the reader was not looking at: centred spectra
                are all pulled toward zero, and these are not. */}
            {!showRaw && columnStage.length > 0 && (
                <p className="text-xs text-gray-400 dark:text-gray-500">
                    {columnStage.join(' → ')} follows, and is not shown — it would pull every
                    curve toward zero and hide the shape you are judging.
                </p>
            )}

            {/* The previous curves stay on screen. The usual cause is a
                half-typed window, and blanking the plot on every intermediate
                value would make it flicker rather than inform. */}
            {preview.error && (
                <p className="text-xs text-amber-600 dark:text-amber-400 mt-1">
                    Showing the last valid preview — {preview.error}
                </p>
            )}
        </HelpWrapper>
    );
}
