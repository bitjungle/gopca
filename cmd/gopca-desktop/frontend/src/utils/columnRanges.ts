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

// Helpers for describing and parsing column ranges.
//
// Wide datasets are almost always an ordered axis — wavelengths, retention
// times, m/z values, genomic positions — so exclusions are naturally contiguous
// regions rather than scattered individual columns. These helpers turn a set of
// indices into the regions a user recognises, and back again.

/** A contiguous, inclusive run of column indices. */
export interface ColumnRun {
    start: number;
    end: number;
}

/**
 * Collapse column indices into contiguous runs, so 57 excluded channels can be
 * shown as two regions rather than 57 chips.
 */
export function toRuns(indices: number[]): ColumnRun[] {
    if (indices.length === 0) return [];
    const sorted = [...new Set(indices)].sort((a, b) => a - b);
    const runs: ColumnRun[] = [];
    let start = sorted[0];
    let prev = sorted[0];
    for (let i = 1; i < sorted.length; i++) {
        if (sorted[i] === prev + 1) {
            prev = sorted[i];
            continue;
        }
        runs.push({ start, end: prev });
        start = sorted[i];
        prev = sorted[i];
    }
    runs.push({ start, end: prev });
    return runs;
}

/** Label a run by its column names: "1400–1450", or just "1400" for one column. */
export function describeRun(run: ColumnRun, headers: string[]): string {
    const a = headers[run.start] ?? String(run.start + 1);
    const b = headers[run.end] ?? String(run.end + 1);
    return run.start === run.end ? a : `${a}–${b}`;
}

/** Expand a run to the individual indices it covers. */
export function runToIndices(run: ColumnRun): number[] {
    const out: number[] = [];
    for (let i = run.start; i <= run.end; i++) out.push(i);
    return out;
}

export interface ParsedSpec {
    indices: number[];
    errors: string[];
}

/**
 * Parse a comma-separated exclusion spec into column indices.
 *
 * Deliberately mirrors the `--exclude-columns` flag of the pca CLI, including
 * its resolution order: a token is matched against column names first, then as
 * a range, then as a 1-based index. Name-first matters for spectra, where a
 * column called "1400" is a wavelength rather than the 1400th column.
 *
 * A range whose endpoints are both column names — "1400-1450" on a wavelength
 * axis — is resolved by name, which is what someone typing it means. Unresolved
 * tokens are returned rather than dropped, so a typo is visible instead of
 * silently narrowing the selection.
 */
export function parseRangeSpec(spec: string, headers: string[]): ParsedSpec {
    const indices = new Set<number>();
    const errors: string[] = [];
    const nameIndex = new Map<string, number>();
    headers.forEach((h, i) => {
        if (!nameIndex.has(h)) nameIndex.set(h, i);
    });

    for (const raw of spec.split(',')) {
        const token = raw.trim();
        if (token === '') continue;

        // 1. Whole token as a column name (handles names containing a hyphen).
        const byName = nameIndex.get(token);
        if (byName !== undefined) {
            indices.add(byName);
            continue;
        }

        // 2. A range "a-b", by column name at both ends or by 1-based index.
        const dash = token.indexOf('-', 1);
        if (dash > 0) {
            const lhs = token.slice(0, dash).trim();
            const rhs = token.slice(dash + 1).trim();
            const lhsName = nameIndex.get(lhs);
            const rhsName = nameIndex.get(rhs);
            if (lhsName !== undefined && rhsName !== undefined) {
                const lo = Math.min(lhsName, rhsName);
                const hi = Math.max(lhsName, rhsName);
                for (let i = lo; i <= hi; i++) indices.add(i);
                continue;
            }
            const lhsNum = Number(lhs);
            const rhsNum = Number(rhs);
            if (Number.isInteger(lhsNum) && Number.isInteger(rhsNum)) {
                if (lhsNum < 1 || rhsNum > headers.length || lhsNum > rhsNum) {
                    errors.push(`${token} (outside 1–${headers.length})`);
                    continue;
                }
                for (let i = lhsNum; i <= rhsNum; i++) indices.add(i - 1);
                continue;
            }
            errors.push(`${token} (not a range of column names or indices)`);
            continue;
        }

        // 3. A single 1-based index.
        const num = Number(token);
        if (Number.isInteger(num)) {
            if (num < 1 || num > headers.length) {
                errors.push(`${token} (outside 1–${headers.length})`);
            } else {
                indices.add(num - 1);
            }
            continue;
        }

        errors.push(`${token} (no column with this name)`);
    }

    return { indices: [...indices].sort((a, b) => a - b), errors };
}

/** Mean of each column, for the overview sparkline. Ignores non-finite values. */
export function columnMeans(data: number[][], columnCount: number): number[] {
    const sums = new Array(columnCount).fill(0);
    const counts = new Array(columnCount).fill(0);
    for (const row of data) {
        for (let c = 0; c < columnCount; c++) {
            const v = row[c];
            if (Number.isFinite(v)) {
                sums[c] += v;
                counts[c] += 1;
            }
        }
    }
    return sums.map((s, i) => (counts[i] > 0 ? s / counts[i] : 0));
}
