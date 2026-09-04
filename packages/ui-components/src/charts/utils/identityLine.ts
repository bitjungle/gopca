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

/**
 * Endpoints for a y = x reference line spanning a scatter series.
 *
 * Extracted from PlotlyScatterChart so the empty and non-finite cases can be
 * tested without rendering Plotly. `Math.min()` of an empty list is `Infinity`
 * and `Math.max()` of one is `-Infinity`, so a series with no finite point used
 * to produce a trace running from Infinity to -Infinity — not a degenerate line
 * but an invalid one, which Plotly may reject outright.
 *
 * Returns null when there is nothing to span. An omitted diagonal is the honest
 * depiction of a series with no points in it; a line drawn anyway would suggest
 * agreement between measurements that were never made.
 */
export function identityLineEnds(
  points: ReadonlyArray<{ x: number; y: number }>,
  domainX?: readonly [number, number] | number[]
): [number, number] | null {
  // An explicit domain wins, because the caller has told us what the axis shows.
  // It is read per-endpoint rather than all-or-nothing so a half-specified
  // domain still falls back to the data for the end it did not fix.
  const values: number[] = [];
  for (const p of points) {
    if (Number.isFinite(p.x)) values.push(p.x);
    if (Number.isFinite(p.y)) values.push(p.y);
  }

  // Math.min() of an empty list is Infinity and Math.max() of one is -Infinity,
  // so the finite check below is what actually rejects an empty series; guarding
  // on values.length as well would be a second spelling of the same condition,
  // and one a test cannot distinguish from the first.
  const from = domainX?.[0] ?? Math.min(...values);
  const to = domainX?.[1] ?? Math.max(...values);

  if (!Number.isFinite(from) || !Number.isFinite(to)) return null;
  return [from, to];
}
