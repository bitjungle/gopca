#!/bin/bash
# Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
# MSIX Package Builder for GoPCA Desktop
#
# This script creates an MSIX package for Microsoft Store distribution.
# It can run on any platform for structure validation, but MakeAppx.exe
# is only available on Windows.

set -e  # Exit on error

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Default values
VERSION=""
PUBLISHER=""
EXE_PATH=""
OUTPUT_DIR=""
SKIP_MAKEAPPX=false

# Usage information
usage() {
    cat << EOF
Usage: $0 --version VERSION --publisher PUBLISHER --exe EXE_PATH --output OUTPUT_DIR [--skip-makeappx]

Build an MSIX package for GoPCA Desktop.

Options:
    --version VERSION       Version in X.X.X format (will be converted to X.X.X.0)
    --publisher PUBLISHER   Publisher ID with CN= prefix (e.g., "CN=12345678-...")
    --exe EXE_PATH         Path to GoPCA.exe binary
    --output OUTPUT_DIR    Output directory for MSIX package
    --skip-makeappx        Skip MakeAppx.exe step (for non-Windows testing)
    -h, --help             Show this help message

Example:
    $0 --version 1.1.4 \\
       --publisher "CN=12345678-90AB-CDEF-1234-567890ABCDEF" \\
       --exe build/GoPCA.exe \\
       --output build/msix
EOF
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --version)
            VERSION="$2"
            shift 2
            ;;
        --publisher)
            PUBLISHER="$2"
            shift 2
            ;;
        --exe)
            EXE_PATH="$2"
            shift 2
            ;;
        --output)
            OUTPUT_DIR="$2"
            shift 2
            ;;
        --skip-makeappx)
            SKIP_MAKEAPPX=true
            shift
            ;;
        -h|--help)
            usage
            exit 0
            ;;
        *)
            echo -e "${RED}Error: Unknown option $1${NC}"
            usage
            exit 1
            ;;
    esac
done

# Validate required arguments
if [ -z "$VERSION" ] || [ -z "$PUBLISHER" ] || [ -z "$EXE_PATH" ] || [ -z "$OUTPUT_DIR" ]; then
    echo -e "${RED}Error: Missing required arguments${NC}"
    usage
    exit 1
fi

# Validate version format (should be X.X.X or X.X.X.X)
if ! [[ "$VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+(\.[0-9]+)?$ ]]; then
    echo -e "${RED}Error: Invalid version format: $VERSION${NC}"
    echo "Expected format: X.X.X or X.X.X.X"
    exit 1
fi

# Convert version to X.X.X.0 format if needed
if ! [[ "$VERSION" =~ \.[0-9]+$ ]]; then
    MSIX_VERSION="${VERSION}.0"
else
    MSIX_VERSION="$VERSION"
fi

# Validate publisher format (should start with CN=)
if ! [[ "$PUBLISHER" =~ ^CN= ]]; then
    echo -e "${RED}Error: Publisher must start with 'CN='${NC}"
    echo "Got: $PUBLISHER"
    exit 1
fi

# Validate EXE exists
if [ ! -f "$EXE_PATH" ]; then
    echo -e "${RED}Error: GoPCA.exe not found at: $EXE_PATH${NC}"
    exit 1
fi

echo -e "${GREEN}=== MSIX Package Builder ===${NC}"
echo "Version: $VERSION → $MSIX_VERSION"
echo "Publisher: $PUBLISHER"
echo "Executable: $EXE_PATH"
echo "Output: $OUTPUT_DIR"
echo

# Find repository root (script is in scripts/windows/)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Paths
TEMPLATE_PATH="$REPO_ROOT/cmd/gopca-desktop/AppxManifest.template.xml"
ASSETS_DIR="$REPO_ROOT/cmd/gopca-desktop/Assets"
PACKAGE_DIR="$OUTPUT_DIR/package"
MANIFEST_PATH="$PACKAGE_DIR/AppxManifest.xml"
MSIX_PATH="$OUTPUT_DIR/GoPCA_${MSIX_VERSION}_x64.msix"

# Validate template exists
if [ ! -f "$TEMPLATE_PATH" ]; then
    echo -e "${RED}Error: AppxManifest template not found: $TEMPLATE_PATH${NC}"
    exit 1
fi

# Validate assets directory exists
if [ ! -d "$ASSETS_DIR" ]; then
    echo -e "${RED}Error: Assets directory not found: $ASSETS_DIR${NC}"
    exit 1
fi

# Create output directory structure
echo -e "${YELLOW}[1/5] Creating package structure...${NC}"
rm -rf "$PACKAGE_DIR"
mkdir -p "$PACKAGE_DIR/Assets"

# Copy Assets
echo -e "${YELLOW}[2/5] Copying assets...${NC}"
cp -v "$ASSETS_DIR"/*.png "$PACKAGE_DIR/Assets/"

# Copy GoPCA.exe
echo -e "${YELLOW}[3/5] Copying GoPCA.exe...${NC}"
cp -v "$EXE_PATH" "$PACKAGE_DIR/GoPCA.exe"

# Generate AppxManifest.xml from template
echo -e "${YELLOW}[4/5] Generating AppxManifest.xml...${NC}"
sed -e "s|{{VERSION}}|${MSIX_VERSION}|g" \
    -e "s|{{PUBLISHER}}|${PUBLISHER}|g" \
    "$TEMPLATE_PATH" > "$MANIFEST_PATH"

echo "Manifest generated: $MANIFEST_PATH"
echo "  Version: $MSIX_VERSION"
echo "  Publisher: $PUBLISHER"

# Validate manifest was generated correctly
if ! grep -q "$MSIX_VERSION" "$MANIFEST_PATH"; then
    echo -e "${RED}Error: Version not substituted in manifest${NC}"
    exit 1
fi

if ! grep -q "$PUBLISHER" "$MANIFEST_PATH"; then
    echo -e "${RED}Error: Publisher not substituted in manifest${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Package structure created successfully${NC}"
echo

# Create MSIX package (Windows only)
if [ "$SKIP_MAKEAPPX" = true ]; then
    echo -e "${YELLOW}[5/5] Skipping MakeAppx.exe (--skip-makeappx flag)${NC}"
    echo -e "${GREEN}✓ Package structure ready at: $PACKAGE_DIR${NC}"
    echo
    echo "To build MSIX on Windows, run:"
    echo "  MakeAppx.exe pack /v /h SHA256 /d \"$PACKAGE_DIR\" /p \"$MSIX_PATH\""
else
    echo -e "${YELLOW}[5/5] Building MSIX package with MakeAppx.exe...${NC}"

    # Check if we're on Windows and MakeAppx is available
    if command -v MakeAppx.exe &> /dev/null; then
        MakeAppx.exe pack /v /h SHA256 /d "$PACKAGE_DIR" /p "$MSIX_PATH"
        echo -e "${GREEN}✓ MSIX package created: $MSIX_PATH${NC}"
        ls -lh "$MSIX_PATH"
    else
        echo -e "${YELLOW}Warning: MakeAppx.exe not found${NC}"
        echo "This is expected on non-Windows systems."
        echo "Package structure is ready at: $PACKAGE_DIR"
        echo
        echo "To build MSIX on Windows, run:"
        echo "  MakeAppx.exe pack /v /h SHA256 /d \"$PACKAGE_DIR\" /p \"$MSIX_PATH\""
    fi
fi

echo
echo -e "${GREEN}=== Build Complete ===${NC}"
echo "Package directory: $PACKAGE_DIR"
echo "MSIX path (when built): $MSIX_PATH"
