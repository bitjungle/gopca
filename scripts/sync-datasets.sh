#!/bin/bash
# GoPCA Suite
#
# Copyright © 2025-2026 Rune Mathisen <devel@bitjungle.com>
#
# This file is part of GoPCA Suite. See LICENSE for the full license terms.
#
# sync-datasets.sh — Compress source CSVs into internal/datasets/*.csv.gz
#
# The embedded dataset files (internal/datasets/*.csv.gz) are compiled into
# the GoPCA binary at build time via //go:embed directives.  This script is
# the single authoritative step that keeps them in sync with their sources in
# testdata/.  It must be run (and is run automatically) before every build.
#
# Mapping: source CSV → embedded gz
#   testdata/iris/iris.csv                   → internal/datasets/iris.csv.gz
#   testdata/wine/wine.csv                   → internal/datasets/wine.csv.gz
#   testdata/corn/corn.csv                   → internal/datasets/corn.csv.gz
#   testdata/swiss_roll/circles.csv          → internal/datasets/swiss_roll.csv.gz
#   testdata/eye_state/eeg_eye_state.csv     → internal/datasets/eeg_eye_state.csv.gz
#   testdata/CSTR/cstr_temporal_pca.csv      → internal/datasets/cstr.csv.gz

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$( cd "$SCRIPT_DIR/.." && pwd )"
cd "$PROJECT_ROOT"

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

DATASETS_DIR="internal/datasets"

# --check reports differences without writing, so CI can fail on a stale
# embedded dataset instead of the difference being noticed months later.
CHECK_ONLY=0
if [ "${1:-}" = "--check" ]; then
    CHECK_ONLY=1
fi

# Array of "source_csv:target_gz_name" pairs
# Every embedded dataset must appear here, and the source named must be the file
# the tutorial actually ships. Three entries were wrong until #915: swiss_roll
# pointed at circles.csv (a different dataset entirely), body_measures had no
# entry at all, and iris pointed at a test fixture that carries an extra
# class-coded column the tutorial does not want.
ENTRIES=(
    "testdata/iris/iris_tutorial.csv:iris"
    "testdata/wine/wine.csv:wine"
    "testdata/corn/corn.csv:corn"
    "testdata/swiss_roll/swiss_roll.csv:swiss_roll"
    "testdata/eye_state/eeg_eye_state.csv:eeg_eye_state"
    "testdata/CSTR/cstr_temporal_pca.csv:cstr"
    "testdata/nhanes/body_measures.csv:body_measures"
)

echo "Synchronizing embedded datasets..."
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

errors=0
updated=0
skipped=0

for entry in "${ENTRIES[@]}"; do
    src="${entry%%:*}"
    name="${entry##*:}"
    dst="$DATASETS_DIR/${name}.csv.gz"

    if [ ! -f "$src" ]; then
        echo -e "${RED}✗${NC} Source not found: $src"
        errors=$((errors + 1))
        continue
    fi

    # Compare content, not timestamps. The guard used to be "$src" -nt "$dst",
    # which depends on filesystem mtimes: a git checkout reorders them
    # arbitrarily, so a needed rebuild could be skipped and a stale embedded
    # dataset could sit undetected indefinitely. That is exactly what happened
    # to three of these entries (#915). Comparing the decompressed bytes makes
    # the answer a property of the files themselves.
    if [ ! -f "$dst" ] || ! gzip -dc "$dst" 2>/dev/null | cmp -s - "$src"; then
        if [ "$CHECK_ONLY" = "1" ]; then
            # Missing and stale need different remedies -- one is a dataset that
            # was never built, the other one that drifted -- so say which.
            if [ ! -f "$dst" ]; then
                echo -e "${RED}✗${NC} ${name}: embedded copy $dst is missing"
            else
                echo -e "${RED}✗${NC} ${name}: embedded copy differs from $src"
            fi
            errors=$((errors + 1))
        else
            # -n omits the modification time from the gzip header, so the same
            # input always produces the same bytes. Without it every rebuild
            # shows as a binary diff even when the data is unchanged.
            gzip -n -c "$src" > "$dst"
            echo -e "${GREEN}✓${NC} ${name}: compressed $src → $dst"
            updated=$((updated + 1))
        fi
    else
        echo -e "${GREEN}✓${NC} ${name}: up to date"
        skipped=$((skipped + 1))
    fi
done

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if [ $errors -gt 0 ]; then
    echo -e "${RED}✗${NC} Dataset sync failed ($errors errors)"
    exit 1
fi

echo -e "${GREEN}✓${NC} Dataset sync complete — $updated updated, $skipped already current"
