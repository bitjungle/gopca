/*
 * GoPCA Suite
 *
 * Copyright © 2025-2026 Rune Mathisen <devel@bitjungle.com>
 *
 * This file is part of GoPCA Suite.
 *
 * See LICENSE for the full license terms.
 */

// Structural checks on the GoCSV column context menu.
//
// Two defects reached a build in two days, and no existing gate could have gone
// red for either. #923 gated `Number the Rows` on `hasRowNames`, hiding it in
// the only situation it exists for. #927 shipped `Mark as Category Column`
// with no icon, so it rendered unindented against its twelve neighbours.
//
// Both are valid TypeScript, so `tsc` passes. Both live in a React callback the
// Go tests never reach, so `make ci-test` passes. The menu has no test runner
// of its own, and adding one to check a property this shallow is not worth the
// dependency. Parsing the source is crude, but it is the check that would have
// failed for both bugs.

import { readFileSync } from 'node:fs';

const FILE = 'cmd/gocsv/frontend/src/components/CSVGrid.tsx';
const src = readFileSync(FILE, 'utf8');
const failures = [];

// Every menu entry carries an icon. The entries reserve horizontal space for a
// glyph, so one without an icon is visibly misaligned rather than merely plain.
const labels = [...src.matchAll(/label: (?:[^,]*\?\s*)?'([^']+)'/g)];
const icons = [...src.matchAll(/icon: </g)].map(m => m.index);

if (labels.length === 0) {
    failures.push(`no menu labels found in ${FILE} -- this check has gone stale`);
}

for (const [i, m] of labels.entries()) {
    const end = i + 1 < labels.length ? labels[i + 1].index : src.length;
    if (!icons.some(ic => ic > m.index && ic < end)) {
        failures.push(`menu entry "${m[1]}" has no icon (#927)`);
    }
}

// The row-name entries are mutually exclusive and each must sit under the guard
// that makes it usable: the command behind each one refuses in the other state.
const guarded = (guard, label) => {
    const g = src.indexOf(`...(${guard}`);
    if (g === -1) return `no ${guard} guard -- either it was removed or inverted (#923)`;
    const l = src.indexOf(`label: '${label}'`, g);
    // The guard's spread ends at the next spread or the following separator.
    const next = src.indexOf('...(', g + 4);
    const end = next === -1 ? src.length : next;
    if (l === -1 || l > end) return `"${label}" is not inside the ${guard} guard (#923)`;
    return null;
};

for (const f of [
    guarded('!hasRowNames', 'Number the Rows'),
    guarded('hasRowNames', 'Move Row Names into Table'),
]) {
    if (f) failures.push(f);
}

if (failures.length > 0) {
    console.error(`\nContext menu checks FAILED in ${FILE}:\n`);
    for (const f of failures) console.error(`  - ${f}`);
    console.error('');
    process.exit(1);
}

console.log(`Context menu invariants hold (${labels.length} entries, all with icons).`);
