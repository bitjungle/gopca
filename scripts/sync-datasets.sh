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

# Array of "source_csv:target_gz_name" pairs
ENTRIES=(
    "testdata/iris/iris.csv:iris"
    "testdata/wine/wine.csv:wine"
    "testdata/corn/corn.csv:corn"
    "testdata/swiss_roll/circles.csv:swiss_roll"
    "testdata/eye_state/eeg_eye_state.csv:eeg_eye_state"
    "testdata/CSTR/cstr_temporal_pca.csv:cstr"
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

    # Recompress if the gz doesn't exist or the source is newer than the gz
    if [ ! -f "$dst" ] || [ "$src" -nt "$dst" ]; then
        gzip -c "$src" > "$dst"
        echo -e "${GREEN}✓${NC} ${name}: compressed $src → $dst"
        updated=$((updated + 1))
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
