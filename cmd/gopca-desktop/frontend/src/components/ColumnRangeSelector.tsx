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

import React, { useCallback, useMemo, useRef, useState } from 'react';
import { X } from 'lucide-react';
import {
    toRuns, describeRun, runToIndices, parseRangeSpec, columnMeans
} from '../utils/columnRanges';

interface ColumnRangeSelectorProps {
    headers: string[];
    data: number[][];
    /** 0-based indices currently held out of the analysis. */
    excludedColumns: number[];
    onChange: (excluded: number[]) => void;
}

const VIEW_W = 1000;
const VIEW_H = 120;

/**
 * Range selector for wide datasets.
 *
 * A 700-channel spectrum rendered as 700 checkboxes is a strip tens of thousands
 * of pixels wide, where excluding a region means finding both ends by scrolling.
 * The columns of such a dataset are not an unordered set though — they are an
 * ordered axis, so the natural gesture is to drag across the region you want
 * gone.
 *
 * The profile behind the axis is the mean of each column, which for spectroscopy
 * is the mean spectrum: the water bands appear as peaks, and the user drags over
 * the shape they can already see.
 */
export const ColumnRangeSelector: React.FC<ColumnRangeSelectorProps> = ({
    headers, data, excludedColumns, onChange
}) => {
    const svgRef = useRef<SVGSVGElement>(null);
    const [drag, setDrag] = useState<{ from: number; to: number } | null>(null);
    const [spec, setSpec] = useState('');
    const [specErrors, setSpecErrors] = useState<string[]>([]);

    const n = headers.length;
    const excludedSet = useMemo(() => new Set(excludedColumns), [excludedColumns]);
    const runs = useMemo(() => toRuns(excludedColumns), [excludedColumns]);

    // Mean profile, normalised into the viewBox. Recomputed only when the data
    // changes, not on every drag frame.
    const profile = useMemo(() => {
        const means = columnMeans(data, n);
        const lo = Math.min(...means);
        const hi = Math.max(...means);
        const span = hi - lo || 1;
        return means.map((m, i) => {
            const x = n === 1 ? 0 : (i / (n - 1)) * VIEW_W;
            const y = VIEW_H - 8 - ((m - lo) / span) * (VIEW_H - 24);
            return `${x.toFixed(1)},${y.toFixed(1)}`;
        }).join(' ');
    }, [data, n]);

    /** Map a pointer position to the column index under it. */
    const columnAt = useCallback((clientX: number): number => {
        const rect = svgRef.current?.getBoundingClientRect();
        if (!rect || rect.width === 0) return 0;
        const frac = (clientX - rect.left) / rect.width;
        return Math.max(0, Math.min(n - 1, Math.round(frac * (n - 1))));
    }, [n]);

    const commit = useCallback((from: number, to: number) => {
        const lo = Math.min(from, to);
        const hi = Math.max(from, to);
        const range = runToIndices({ start: lo, end: hi });
        // Dragging over a region that is already wholly excluded puts it back,
        // so the same gesture both removes and restores.
        const allExcluded = range.every(i => excludedSet.has(i));
        const next = new Set(excludedColumns);
        range.forEach(i => (allExcluded ? next.delete(i) : next.add(i)));
        onChange([...next].sort((a, b) => a - b));
    }, [excludedColumns, excludedSet, onChange]);

    const applySpec = () => {
        const { indices, errors } = parseRangeSpec(spec, headers);
        setSpecErrors(errors);
        if (indices.length > 0) {
            onChange([...new Set([...excludedColumns, ...indices])].sort((a, b) => a - b));
            if (errors.length === 0) setSpec('');
        }
    };

    const xOf = (i: number) => (n === 1 ? 0 : (i / (n - 1)) * VIEW_W);
    // Ticks at even fractions, labelled with the real column names.
    const ticks = useMemo(() => {
        const count = Math.min(6, n);
        return Array.from({ length: count }, (_, k) => {
            const i = Math.round((k / Math.max(1, count - 1)) * (n - 1));
            return { i, label: headers[i] ?? String(i + 1) };
        });
    }, [headers, n]);

    const included = n - excludedColumns.length;

    return (
        <div className="bg-white dark:bg-gray-800 rounded-lg border border-gray-300 dark:border-gray-600 p-4">
            <div className="flex items-center justify-between mb-3 gap-3 flex-wrap">
                <div className="flex items-baseline gap-2">
                    <h4 className="font-medium text-gray-900 dark:text-white">Variables</h4>
                    <span className="text-xs text-gray-500 dark:text-gray-400">
                        {excludedColumns.length === 0
                            ? `all ${n} included`
                            : `${included} of ${n} included · ${excludedColumns.length} excluded`}
                    </span>
                </div>
                <div className="flex items-center gap-2">
                    <span className="hidden md:inline text-xs text-gray-400 dark:text-gray-500">
                        drag across the plot to exclude a region
                    </span>
                    {excludedColumns.length > 0 && (
                        <button
                            onClick={() => onChange([])}
                            className="text-xs px-2 py-1 bg-gray-200 dark:bg-gray-700 hover:bg-gray-300 dark:hover:bg-gray-600 rounded"
                        >
                            Include all
                        </button>
                    )}
                </div>
            </div>

            <svg
                ref={svgRef}
                viewBox={`0 0 ${VIEW_W} ${VIEW_H}`}
                preserveAspectRatio="none"
                className="w-full h-28 select-none cursor-crosshair rounded border border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-900"
                onPointerDown={(e) => {
                    (e.target as Element).setPointerCapture?.(e.pointerId);
                    const i = columnAt(e.clientX);
                    setDrag({ from: i, to: i });
                }}
                onPointerMove={(e) => {
                    if (!drag) return;
                    setDrag({ from: drag.from, to: columnAt(e.clientX) });
                }}
                onPointerUp={(e) => {
                    // Take the end of the range from the release position rather
                    // than the last move: a quick drag can produce no intermediate
                    // pointermove at all, which would collapse it to one column.
                    if (drag) commit(drag.from, columnAt(e.clientX));
                    setDrag(null);
                }}
                onPointerLeave={(e) => {
                    if (drag) commit(drag.from, columnAt(e.clientX));
                    setDrag(null);
                }}
            >
                {/* Already-excluded regions */}
                {runs.map((r, k) => (
                    <rect
                        key={`ex-${k}`}
                        x={xOf(r.start)}
                        width={Math.max(1.5, xOf(r.end) - xOf(r.start))}
                        y={0} height={VIEW_H}
                        className="fill-red-400/25 dark:fill-red-500/25"
                    />
                ))}

                {/* Live drag preview */}
                {drag && (
                    <rect
                        x={xOf(Math.min(drag.from, drag.to))}
                        width={Math.max(1.5, xOf(Math.max(drag.from, drag.to)) - xOf(Math.min(drag.from, drag.to)))}
                        y={0} height={VIEW_H}
                        className="fill-blue-400/30 dark:fill-blue-400/25"
                    />
                )}

                <polyline
                    points={profile}
                    fill="none"
                    vectorEffect="non-scaling-stroke"
                    className="stroke-blue-600 dark:stroke-blue-400"
                    strokeWidth={1.5}
                />

                {ticks.map(t => (
                    <line
                        key={`t-${t.i}`}
                        x1={xOf(t.i)} x2={xOf(t.i)} y1={VIEW_H - 6} y2={VIEW_H}
                        vectorEffect="non-scaling-stroke"
                        className="stroke-gray-400 dark:stroke-gray-500"
                        strokeWidth={1}
                    />
                ))}
            </svg>

            {/* Axis labels sit outside the SVG so they are not distorted by
                preserveAspectRatio="none". */}
            <div className="relative h-4 mt-1">
                {ticks.map(t => (
                    <span
                        key={`l-${t.i}`}
                        style={{ left: `${(xOf(t.i) / VIEW_W) * 100}%` }}
                        className="absolute -translate-x-1/2 text-[10px] text-gray-500 dark:text-gray-400 whitespace-nowrap"
                    >
                        {t.label}
                    </span>
                ))}
            </div>

            <div className="flex items-center gap-2 mt-3 flex-wrap">
                <input
                    type="text"
                    value={spec}
                    onChange={(e) => { setSpec(e.target.value); setSpecErrors([]); }}
                    onKeyDown={(e) => { if (e.key === 'Enter') applySpec(); }}
                    placeholder="Exclude by name or range, e.g. 1400-1450, 1900-1960"
                    aria-label="Exclude columns by name or range"
                    className="flex-1 min-w-[16rem] px-3 py-1.5 text-sm bg-gray-100 dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded text-gray-900 dark:text-white"
                />
                <button
                    onClick={applySpec}
                    disabled={spec.trim() === ''}
                    className="text-sm px-3 py-1.5 bg-blue-600 hover:bg-blue-700 disabled:opacity-40 disabled:hover:bg-blue-600 text-white rounded"
                >
                    Exclude
                </button>
            </div>

            {specErrors.length > 0 && (
                <p className="mt-2 text-xs text-red-600 dark:text-red-400">
                    Could not resolve {specErrors.join(', ')}
                </p>
            )}

            {runs.length > 0 && (
                <div className="flex items-center gap-2 mt-3 flex-wrap">
                    <span className="text-xs text-gray-500 dark:text-gray-400">Excluded:</span>
                    {runs.map((r, k) => (
                        <button
                            key={`chip-${k}`}
                            onClick={() => {
                                const next = new Set(excludedColumns);
                                runToIndices(r).forEach(i => next.delete(i));
                                onChange([...next].sort((a, b) => a - b));
                            }}
                            title="Click to put these columns back"
                            className="inline-flex items-center gap-1 text-xs px-2 py-1 rounded bg-red-100 dark:bg-red-900/40 text-red-800 dark:text-red-200 hover:bg-red-200 dark:hover:bg-red-900/70"
                        >
                            {describeRun(r, headers)}
                            <X className="w-3 h-3" />
                        </button>
                    ))}
                </div>
            )}
        </div>
    );
};
