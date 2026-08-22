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

import React from 'react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { ColumnRangeSelector } from './ColumnRangeSelector';

// 700 NIR channels, 1100–2498 nm at 2 nm steps, as in the Corn dataset.
const HEADERS = Array.from({ length: 700 }, (_, i) => String(1100 + i * 2));
const DATA = Array.from({ length: 5 }, (_, r) => HEADERS.map((_, c) => Math.sin(c / 40) + r * 0.01));
const idx = (nm: number) => (nm - 1100) / 2;

// jsdom gives every element a zero-sized rect, so the pointer-to-column mapping
// needs a real width to work against.
beforeEach(() => {
    vi.spyOn(SVGElement.prototype, 'getBoundingClientRect').mockReturnValue({
        left: 0, top: 0, width: 1000, height: 120, right: 1000, bottom: 120, x: 0, y: 0,
        toJSON: () => ({})
    } as DOMRect);
});

function plot() {
    // The SVG is the only graphics element in the panel.
    return document.querySelector('svg') as SVGSVGElement;
}

describe('ColumnRangeSelector', () => {
    it('reports how many columns are included', () => {
        render(<ColumnRangeSelector headers={HEADERS} data={DATA} excludedColumns={[]} onChange={() => {}} />);
        expect(screen.getByText('all 700 included')).toBeTruthy();
    });

    it('excludes the dragged region — no scrolling involved', () => {
        const onChange = vi.fn();
        render(<ColumnRangeSelector headers={HEADERS} data={DATA} excludedColumns={[]} onChange={onChange} />);
        const svg = plot();
        // Drag from 1400 nm to 1450 nm. x is the column's fraction of the width.
        const xAt = (nm: number) => (idx(nm) / (HEADERS.length - 1)) * 1000;
        fireEvent.pointerDown(svg, { clientX: xAt(1400) });
        fireEvent.pointerMove(svg, { clientX: xAt(1450) });
        fireEvent.pointerUp(svg, { clientX: xAt(1450) });

        expect(onChange).toHaveBeenCalledTimes(1);
        const excluded = onChange.mock.calls[0][0] as number[];
        expect(excluded).toHaveLength(26);
        expect(HEADERS[excluded[0]]).toBe('1400');
        expect(HEADERS[excluded[excluded.length - 1]]).toBe('1450');
    });

    it('dragging back over an excluded region puts it back', () => {
        const already = Array.from({ length: 26 }, (_, i) => idx(1400) + i);
        const onChange = vi.fn();
        render(<ColumnRangeSelector headers={HEADERS} data={DATA} excludedColumns={already} onChange={onChange} />);
        const svg = plot();
        const xAt = (nm: number) => (idx(nm) / (HEADERS.length - 1)) * 1000;
        fireEvent.pointerDown(svg, { clientX: xAt(1400) });
        fireEvent.pointerUp(svg, { clientX: xAt(1450) });
        expect(onChange.mock.calls[0][0]).toEqual([]);
    });

    it('accepts a typed range using the same syntax as the CLI', () => {
        const onChange = vi.fn();
        render(<ColumnRangeSelector headers={HEADERS} data={DATA} excludedColumns={[]} onChange={onChange} />);
        const input = screen.getByLabelText('Exclude columns by name or range');
        fireEvent.change(input, { target: { value: '1400-1450, 1900-1960' } });
        fireEvent.click(screen.getByText('Exclude'));
        expect((onChange.mock.calls[0][0] as number[])).toHaveLength(57);
    });

    it('shows unresolvable tokens instead of silently narrowing the selection', () => {
        render(<ColumnRangeSelector headers={HEADERS} data={DATA} excludedColumns={[]} onChange={() => {}} />);
        const input = screen.getByLabelText('Exclude columns by name or range');
        fireEvent.change(input, { target: { value: 'not-a-column' } });
        fireEvent.click(screen.getByText('Exclude'));
        expect(screen.getByText(/Could not resolve/)).toBeTruthy();
    });

    it('summarises 57 excluded channels as two chips, not 57', () => {
        const water = [
            ...Array.from({ length: 26 }, (_, i) => idx(1400) + i),
            ...Array.from({ length: 31 }, (_, i) => idx(1900) + i)
        ];
        render(<ColumnRangeSelector headers={HEADERS} data={DATA} excludedColumns={water} onChange={() => {}} />);
        expect(screen.getByText('1400–1450')).toBeTruthy();
        expect(screen.getByText('1900–1960')).toBeTruthy();
        expect(screen.getByText('643 of 700 included · 57 excluded')).toBeTruthy();
    });

    it('a chip puts its region back', () => {
        const onChange = vi.fn();
        const water = Array.from({ length: 26 }, (_, i) => idx(1400) + i);
        render(<ColumnRangeSelector headers={HEADERS} data={DATA} excludedColumns={water} onChange={onChange} />);
        fireEvent.click(screen.getByText('1400–1450'));
        expect(onChange.mock.calls[0][0]).toEqual([]);
    });
});

describe('ColumnRangeSelector profile rendering', () => {
    // 40 unrelated variables: numeric column *names* are what make an axis, not
    // numeric data. A line drawn through these would assert an order they lack.
    const NAMED = Array.from({ length: 40 }, (_, i) => `assay_${i}`);
    const NAMED_DATA = Array.from({ length: 5 }, () => NAMED.map((_, c) => c % 7));

    it('draws a connected line for a wavelength axis', () => {
        render(<ColumnRangeSelector headers={HEADERS} data={DATA} excludedColumns={[]} onChange={() => {}} />);
        expect(plot().querySelector('polyline')).toBeTruthy();
        expect(plot().querySelector('path')).toBeNull();
    });

    it('draws bars for unordered variables', () => {
        render(<ColumnRangeSelector headers={NAMED} data={NAMED_DATA} excludedColumns={[]} onChange={() => {}} />);
        expect(plot().querySelector('polyline')).toBeNull();
        const bars = plot().querySelector('path');
        expect(bars).toBeTruthy();
        // One move-and-vertical-line per variable.
        expect((bars!.getAttribute('d') || '').match(/M/g) || []).toHaveLength(NAMED.length);
    });

    it('drags identically in bar mode', () => {
        const onChange = vi.fn();
        render(<ColumnRangeSelector headers={NAMED} data={NAMED_DATA} excludedColumns={[]} onChange={onChange} />);
        const svg = plot();
        const xAt = (i: number) => (i / (NAMED.length - 1)) * 1000;
        fireEvent.pointerDown(svg, { clientX: xAt(10) });
        fireEvent.pointerMove(svg, { clientX: xAt(14) });
        fireEvent.pointerUp(svg, { clientX: xAt(14) });

        expect(onChange).toHaveBeenCalledTimes(1);
        expect(onChange.mock.calls[0][0]).toEqual([10, 11, 12, 13, 14]);
    });

    it('treats numeric names that are out of order as unordered', () => {
        const shuffled = ['30', '10', '20', '40'];
        const data = [[1, 2, 3, 4], [2, 3, 4, 5]];
        render(<ColumnRangeSelector headers={shuffled} data={data} excludedColumns={[]} onChange={() => {}} />);
        expect(plot().querySelector('polyline')).toBeNull();
        expect(plot().querySelector('path')).toBeTruthy();
    });
});
