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

# --check mode: report drift without modifying any files; exit 1 if out of sync.
# Used by the pre-commit hook and CI to catch a stale mirror before it lands.
CHECK_ONLY=false
[ "${1:-}" = "--check" ] && CHECK_ONLY=true
DRIFT=0

# Source documentation files
GOPCA_DOC="docs/intro_to_pca.md"
GOCSV_DOC="docs/intro_to_data_prep.md"

# Source image directories
IMAGES_DIR="docs/images"

# Target directories
GOPCA_TARGET="cmd/gopca-desktop/frontend/public/docs"
GOCSV_TARGET="cmd/gocsv/frontend/public/docs"
GOPCA_IMAGES_TARGET="cmd/gopca-desktop/frontend/public/docs/images"
GOCSV_IMAGES_TARGET="cmd/gocsv/frontend/public/docs/images"

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
    
    # Already in sync?
    if [ -f "$target_file" ] && diff -q "$source" "$target_file" >/dev/null 2>&1; then
        echo -e "${GREEN}✓${NC} $app_name documentation already up to date"
        return 0
    fi

    # Source and mirror differ (or the mirror is missing).
    if [ "$CHECK_ONLY" = true ]; then
        echo -e "${RED}✗${NC} OUT OF SYNC: $target_file (differs from $source)"
        DRIFT=1
        return 0
    fi

    if [ -f "$target_file" ]; then
        echo -e "${YELLOW}⟳${NC} Updating $app_name documentation..."
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

# Function to sync a tutorial directory (markdown + images) to frontend public.
# Accepts an optional second argument to override the target directory name,
# for cases where the testdata folder name differs from the frontend path
# (e.g. testdata/eye_state → public/tutorials/eeg_eye_state).
sync_tutorial() {
    local dataset=$1
    local target_name="${2:-$1}"
    local source_dir="testdata/$dataset"
    local target_dir="cmd/gopca-desktop/frontend/public/tutorials/$target_name"

    if [ ! -d "$source_dir" ]; then
        echo -e "${RED}✗${NC} Tutorial source directory not found: $source_dir"
        return 1
    fi

    mkdir -p "$target_dir"

    local updated=0
    local count=0
    local drifted=0

    # Sync .md files and image files (.png, .jpg, .jpeg, .svg)
    for source_file in "$source_dir"/*.md "$source_dir"/*.png "$source_dir"/*.jpg "$source_dir"/*.jpeg "$source_dir"/*.svg; do
        [ -f "$source_file" ] || continue
        count=$((count + 1))
        filename=$(basename "$source_file")
        target_file="$target_dir/$filename"

        if [ -f "$target_file" ] && diff -q "$source_file" "$target_file" >/dev/null 2>&1; then
            continue  # already in sync
        fi
        if [ "$CHECK_ONLY" = true ]; then
            echo -e "${RED}✗${NC} OUT OF SYNC: $target_file"
            DRIFT=1
            drifted=$((drifted + 1))
        else
            cp -p "$source_file" "$target_file"
            updated=$((updated + 1))
        fi
    done

    if [ $count -eq 0 ]; then
        echo -e "${YELLOW}⚠${NC}  Tutorial '$dataset': no files found in $source_dir"
    elif [ "$CHECK_ONLY" = true ] && [ $drifted -gt 0 ]; then
        echo -e "${RED}✗${NC} Tutorial '$dataset': $drifted of $count file(s) out of sync"
    elif [ $updated -eq 0 ]; then
        echo -e "${GREEN}✓${NC} Tutorial '$dataset' already up to date ($count files)"
    else
        echo -e "${GREEN}✓${NC} Tutorial '$dataset': updated $updated of $count files"
    fi
}

# Function to sync image files
sync_images() {
    local pattern=$1
    local target_dir=$2
    local app_name=$3
    
    # Create target directory if it doesn't exist
    mkdir -p "$target_dir"
    
    # Count matching files
    local count=0
    local updated=0
    local drifted=0

    # Find and copy matching image files
    for source_file in $IMAGES_DIR/$pattern; do
        # Check if pattern matched any files
        if [ ! -f "$source_file" ]; then
            continue
        fi
        
        count=$((count + 1))
        filename=$(basename "$source_file")
        target_file="$target_dir/$filename"
        
        # Already in sync?
        if [ -f "$target_file" ] && diff -q "$source_file" "$target_file" >/dev/null 2>&1; then
            continue
        fi
        if [ "$CHECK_ONLY" = true ]; then
            echo -e "${RED}✗${NC} OUT OF SYNC: $target_file"
            DRIFT=1
            drifted=$((drifted + 1))
        else
            cp -p "$source_file" "$target_file"
            updated=$((updated + 1))
        fi
    done

    if [ $count -eq 0 ]; then
        echo -e "${YELLOW}⚠${NC}  $app_name: No images matching pattern $pattern"
        return 0
    elif [ "$CHECK_ONLY" = true ] && [ $drifted -gt 0 ]; then
        echo -e "${RED}✗${NC} $app_name: $drifted of $count image(s) out of sync"
        return 0
    elif [ $updated -eq 0 ]; then
        echo -e "${GREEN}✓${NC} $app_name images already up to date ($count files)"
        return 0
    else
        echo -e "${GREEN}✓${NC} $app_name: Updated $updated of $count image files"
        return 0
    fi
}

# Main execution
echo "Synchronizing documentation files..."
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Sync GoPCA documentation
sync_doc "$GOPCA_DOC" "$GOPCA_TARGET" "GoPCA Desktop"
gopca_result=$?

# Sync GoPCA images
sync_images "intro_to_pca_fig_*.jpg" "$GOPCA_IMAGES_TARGET" "GoPCA Desktop"
gopca_images_result=$?

# Sync GoCSV documentation
sync_doc "$GOCSV_DOC" "$GOCSV_TARGET" "GoCSV Desktop"
gocsv_result=$?

# Sync GoCSV images (if any exist for data prep guide)
sync_images "intro_to_data_prep_fig_*.jpg" "$GOCSV_IMAGES_TARGET" "GoCSV Desktop"
gocsv_images_result=$?

# Sync sample dataset tutorials
sync_tutorial "iris"
sync_tutorial "wine"
sync_tutorial "corn"
sync_tutorial "swiss_roll"
sync_tutorial "eye_state" "eeg_eye_state"
sync_tutorial "CSTR" "cstr"
sync_tutorial "nhanes" "body_measures"

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# In --check mode, report drift and exit without the "sync complete" messaging.
if [ "$CHECK_ONLY" = true ]; then
    if [ "$DRIFT" -ne 0 ]; then
        echo -e "${RED}✗${NC} Documentation mirrors are OUT OF SYNC — run 'make sync-docs' and commit the result."
        exit 1
    fi
    echo -e "${GREEN}✓${NC} Documentation mirrors are in sync"
    exit 0
fi

# Check overall result
if [ $gopca_result -eq 0 ] && [ $gocsv_result -eq 0 ] && [ $gopca_images_result -eq 0 ] && [ $gocsv_images_result -eq 0 ]; then
    echo -e "${GREEN}✓${NC} Documentation synchronization complete"
    exit 0
else
    echo -e "${RED}✗${NC} Documentation synchronization failed"
    exit 1
fi