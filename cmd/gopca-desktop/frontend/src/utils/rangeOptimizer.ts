// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Utility for optimizing arrays of indices to range notation

/**
 * Optimizes an array of indices to range notation
 * e.g., [1,2,3,5,6] becomes "1-3,5-6"
 *
 * @param indices Array of numeric indices to optimize
 * @returns String in range notation (e.g., "1-3,5-6")
 */
export function optimizeToRanges(indices: number[]): string {
    if (indices.length === 0) {
return '';
}

    const sorted = [...indices].sort((a, b) => a - b);
    const ranges: string[] = [];
    let start = sorted[0];
    let end = sorted[0];

    for (let i = 1; i < sorted.length; i++) {
        if (sorted[i] === end + 1) {
            end = sorted[i];
        } else {
            ranges.push(start === end ? `${start}` : `${start}-${end}`);
            start = sorted[i];
            end = sorted[i];
        }
    }

    ranges.push(start === end ? `${start}` : `${start}-${end}`);
    return ranges.join(',');
}