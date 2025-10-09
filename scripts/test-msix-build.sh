#!/bin/bash
# Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
# Local MSIX Build Test Script
#
# Tests the MSIX build process locally using secrets from .env file.
# This script is for local development and testing only.

set -e

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Find repository root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

echo -e "${GREEN}=== MSIX Build Test ===${NC}"
echo "Repository: $REPO_ROOT"
echo

# Load .env file
ENV_FILE="$REPO_ROOT/.env"
if [ ! -f "$ENV_FILE" ]; then
    echo -e "${RED}Error: .env file not found${NC}"
    echo "Please create .env from .env.example and configure Microsoft Partner Center secrets"
    exit 1
fi

echo -e "${YELLOW}Loading secrets from .env...${NC}"
source "$ENV_FILE"

# Validate required secrets
if [ -z "$MS_WINDOWS_PUBLISHER_ID" ]; then
    echo -e "${RED}Error: MS_WINDOWS_PUBLISHER_ID not set in .env${NC}"
    echo "Please configure this secret in your .env file"
    exit 1
fi

# Construct full publisher string (add CN= prefix)
PUBLISHER="CN=${MS_WINDOWS_PUBLISHER_ID}"
echo "Publisher: $PUBLISHER"
echo

# Test version
VERSION="1.1.4-test"
echo "Test version: $VERSION"
echo

# Check if GoPCA.exe exists (for real testing)
# If not, we'll run in skip-makeappx mode
EXE_PATH="$REPO_ROOT/cmd/gopca-desktop/build/bin/GoPCA.exe"
SKIP_MAKEAPPX=""

if [ ! -f "$EXE_PATH" ]; then
    echo -e "${YELLOW}Warning: GoPCA.exe not found at $EXE_PATH${NC}"
    echo "Will test package structure only (no MSIX build)"
    echo
    # Create a dummy executable for structure testing
    EXE_PATH="$REPO_ROOT/build/GoPCA-test.exe"
    mkdir -p "$(dirname "$EXE_PATH")"
    echo "Test executable" > "$EXE_PATH"
    SKIP_MAKEAPPX="--skip-makeappx"
else
    echo -e "${GREEN}✓ Found GoPCA.exe at $EXE_PATH${NC}"
    echo
fi

# Output directory
OUTPUT_DIR="$REPO_ROOT/build/msix-test"

# Run the build script
echo -e "${YELLOW}Running build-msix.sh...${NC}"
echo

"$REPO_ROOT/scripts/windows/build-msix.sh" \
    --version "$VERSION" \
    --publisher "$PUBLISHER" \
    --exe "$EXE_PATH" \
    --output "$OUTPUT_DIR" \
    $SKIP_MAKEAPPX

echo
echo -e "${GREEN}=== Test Complete ===${NC}"
echo
echo "Package structure created at: $OUTPUT_DIR/package"
echo "You can inspect the package contents and manifest:"
echo "  cat $OUTPUT_DIR/package/AppxManifest.xml"
echo "  ls -la $OUTPUT_DIR/package/"
echo
echo "To build MSIX on Windows, run:"
echo "  MakeAppx.exe pack /v /h SHA256 /d \"$OUTPUT_DIR/package\" /p \"$OUTPUT_DIR/GoPCA_${VERSION}.0_x64.msix\""
