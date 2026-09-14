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
//
// Parsing source with regexes fails in two directions, and only one of them is
// safe. Matching too little is the dangerous one: an entry the patterns do not
// recognise is simply not checked, and the run stays green while the invariant
// stops being enforced. Review of #928 found both failure modes present, so the
// structure below defends against them explicitly -- string-aware bracket
// matching instead of "to the end of the file", and a cross-check that the
// number of entries found equals the number that exist.

import { readFileSync } from 'node:fs';

const FILE = 'cmd/gocsv/frontend/src/components/CSVGrid.tsx';
const src = readFileSync(FILE, 'utf8');
const failures = [];

// Skip string and template literals when scanning for brackets, so a paren
// inside a label or an error message cannot unbalance the count.
function endOfSpread(from) {
    let depth = 0;
    let quote = null;
    for (let i = from; i < src.length; i++) {
        const c = src[i];
        if (quote) {
            if (c === '\\') i++;
            else if (c === quote) quote = null;
            continue;
        }
        if (c === "'" || c === '"' || c === '`') { quote = c; continue; }
        if (c === '(') depth++;
        else if (c === ')') {
            depth--;
            if (depth === 0) return i;
        }
    }
    return -1;
}

// Tolerate both quote styles and any spacing. A label the pattern misses would
// go unchecked rather than fail, which the count cross-check below catches.
const LABEL = /label:\s*(?:[^,]*\?\s*)?['"]([^'"]+)['"]/g;
const labels = [...src.matchAll(LABEL)];
const icons = [...src.matchAll(/icon:\s*</g)].map(m => m.index);

// Every menu entry has both a label and an action. If the two counts disagree,
// the patterns above have stopped matching the source -- the entries are still
// there, but this check is no longer looking at all of them.
const declared = (src.match(/\blabel:/g) || []).length;
const actions = (src.match(/\baction:/g) || []).length;
if (labels.length !== declared) {
    failures.push(
        `matched ${labels.length} of ${declared} label declarations -- the patterns in ` +
        `this check no longer fit the source, so entries are going unchecked`);
}
if (declared !== actions) {
    failures.push(
        `${declared} labels but ${actions} actions -- if an entry legitimately has ` +
        `no action, update this cross-check; otherwise the menu is malformed`);
}
if (labels.length === 0) {
    failures.push(`no menu entries found in ${FILE} -- this check has gone stale`);
}

// Every entry carries an icon. The entries reserve horizontal space for a
// glyph, so one without an icon is visibly misaligned rather than merely plain.
for (const [i, m] of labels.entries()) {
    const end = i + 1 < labels.length ? labels[i + 1].index : src.length;
    if (!icons.some(ic => ic > m.index && ic < end)) {
        failures.push(`menu entry "${m[1]}" has no icon (#927)`);
    }
}

// The row-name entries are mutually exclusive and each must sit under the guard
// that makes it usable: the command behind each one refuses in the other state.
//
// The guard's extent is its own bracket range. Taking it to the end of the file
// when no later spread happens to follow -- as the first version of this check
// did -- means an entry moved out of the guard still counts as inside it,
// which is the exact regression #923 was.
function guarded(guard, label) {
    const open = src.indexOf(`...(${guard}`);
    if (open === -1) {
        return `no ${guard} guard -- either it was removed or inverted (#923)`;
    }
    const start = src.indexOf('(', open);
    const end = endOfSpread(start);
    if (end === -1) {
        return `the ${guard} guard is unterminated`;
    }
    const at = src.indexOf(`'${label}'`, start);
    if (at === -1 || at > end) {
        return `"${label}" is not inside the ${guard} guard (#923)`;
    }
    return null;
}

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
