#!/bin/bash
# Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
# Use of this source code is governed by the MIT license
# that can be found in the LICENSE file.
#
# sync-docs.sh - Synchronize documentation files to frontend public directories
#
# This script ensures that the master documentation files in the docs/ directory
# are copied to the appropriate frontend public directories for serving via HTTP.

set -e

# Get the script directory and project root
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$( cd "$SCRIPT_DIR/.." && pwd )"

# Change to project root
cd "$PROJECT_ROOT"

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Source documentation files
GOPCA_DOC="docs/intro_to_pca.md"
GOCSV_DOC="docs/intro_to_data_prep.md"

# Target directories
GOPCA_TARGET="cmd/gopca-desktop/frontend/public/docs"
GOCSV_TARGET="cmd/gocsv/frontend/public/docs"

# Function to sync a documentation file
sync_doc() {
    local source=$1
    local target_dir=$2
    local app_name=$3
    
    if [ ! -f "$source" ]; then
        echo -e "${RED}✗${NC} Source file not found: $source"
        return 1
    fi
    
    # Create target directory if it doesn't exist
    mkdir -p "$target_dir"
    
    # Get the filename
    filename=$(basename "$source")
    target_file="$target_dir/$filename"
    
    # Check if files are different
    if [ -f "$target_file" ]; then
        if diff -q "$source" "$target_file" >/dev/null 2>&1; then
            echo -e "${GREEN}✓${NC} $app_name documentation already up to date"
            return 0
        else
            echo -e "${YELLOW}⟳${NC} Updating $app_name documentation..."
        fi
    else
        echo -e "${YELLOW}+${NC} Creating $app_name documentation..."
    fi
    
    # Copy with timestamp preservation
    cp -p "$source" "$target_file"
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓${NC} $app_name: $source → $target_file"
        return 0
    else
        echo -e "${RED}✗${NC} Failed to copy $source to $target_file"
        return 1
    fi
}

# Main execution
echo "Synchronizing documentation files..."
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Sync GoPCA documentation
sync_doc "$GOPCA_DOC" "$GOPCA_TARGET" "GoPCA Desktop"
gopca_result=$?

# Sync GoCSV documentation
sync_doc "$GOCSV_DOC" "$GOCSV_TARGET" "GoCSV Desktop"
gocsv_result=$?

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Check overall result
if [ $gopca_result -eq 0 ] && [ $gocsv_result -eq 0 ]; then
    echo -e "${GREEN}✓${NC} Documentation synchronization complete"
    exit 0
else
    echo -e "${RED}✗${NC} Documentation synchronization failed"
    exit 1
fi