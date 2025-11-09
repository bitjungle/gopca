#!/bin/bash
# Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
# MSIX Package Builder for GoPCA Desktop and GoCSV
#
# This script creates a bundled MSIX package containing both GoPCA Desktop
# and GoCSV Desktop for Microsoft Store distribution.
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
GOPCA_EXE=""
GOCSV_EXE=""
OUTPUT_DIR=""
SKIP_MAKEAPPX=false

# Usage information
usage() {
    cat << EOF
Usage: $0 --version VERSION --publisher PUBLISHER --gopca-exe GOPCA_PATH --gocsv-exe GOCSV_PATH --output OUTPUT_DIR [--skip-makeappx]

Build a bundled MSIX package containing both GoPCA Desktop and GoCSV Desktop.

Options:
    --version VERSION       Version in X.X.X format (will be converted to X.X.X.0)
    --publisher PUBLISHER   Publisher ID with CN= prefix (e.g., "CN=12345678-...")
    --gopca-exe GOPCA_PATH Path to GoPCA.exe binary
    --gocsv-exe GOCSV_PATH Path to GoCSV.exe binary
    --output OUTPUT_DIR    Output directory for MSIX package
    --skip-makeappx        Skip MakeAppx.exe step (for non-Windows testing)
    -h, --help             Show this help message

Example:
    $0 --version 1.1.4 \\
       --publisher "CN=12345678-90AB-CDEF-1234-567890ABCDEF" \\
       --gopca-exe cmd/gopca-desktop/build/bin/GoPCA.exe \\
       --gocsv-exe cmd/gocsv/build/bin/GoCSV.exe \\
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
        --gopca-exe)
            GOPCA_EXE="$2"
            shift 2
            ;;
        --gocsv-exe)
            GOCSV_EXE="$2"
            shift 2
            ;;
        --exe)
            # Legacy support: --exe sets GoPCA only
            echo -e "${YELLOW}Warning: --exe is deprecated, use --gopca-exe instead${NC}"
            GOPCA_EXE="$2"
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
if [ -z "$VERSION" ] || [ -z "$PUBLISHER" ] || [ -z "$GOPCA_EXE" ] || [ -z "$GOCSV_EXE" ] || [ -z "$OUTPUT_DIR" ]; then
    echo -e "${RED}Error: Missing required arguments${NC}"
    [ -z "$VERSION" ] && echo "  - Missing: --version"
    [ -z "$PUBLISHER" ] && echo "  - Missing: --publisher"
    [ -z "$GOPCA_EXE" ] && echo "  - Missing: --gopca-exe"
    [ -z "$GOCSV_EXE" ] && echo "  - Missing: --gocsv-exe"
    [ -z "$OUTPUT_DIR" ] && echo "  - Missing: --output"
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

# Validate executables exist
if [ ! -f "$GOPCA_EXE" ]; then
    echo -e "${RED}Error: GoPCA.exe not found at: $GOPCA_EXE${NC}"
    exit 1
fi

if [ ! -f "$GOCSV_EXE" ]; then
    echo -e "${RED}Error: GoCSV.exe not found at: $GOCSV_EXE${NC}"
    echo -e "${YELLOW}Note: GoCSV Desktop must be built before creating MSIX package${NC}"
    echo "  Build command: make csv-build"
    exit 1
fi

echo -e "${GREEN}=== MSIX Package Builder (Bundled) ===${NC}"
echo "Version: $VERSION → $MSIX_VERSION"
echo "Publisher: $PUBLISHER"
echo "GoPCA Executable: $GOPCA_EXE"
echo "GoCSV Executable: $GOCSV_EXE"
echo "Output: $OUTPUT_DIR"
echo

# Find repository root (script is in scripts/windows/)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Paths
TEMPLATE_PATH="$REPO_ROOT/cmd/gopca-desktop/AppxManifest.template.xml"
GOPCA_ASSETS_DIR="$REPO_ROOT/cmd/gopca-desktop/Assets"
GOCSV_ASSETS_DIR="$REPO_ROOT/cmd/gocsv/Assets"
PACKAGE_DIR="$OUTPUT_DIR/package"
MANIFEST_PATH="$PACKAGE_DIR/AppxManifest.xml"
MSIX_PATH="$OUTPUT_DIR/GoPCA_${MSIX_VERSION}_x64.msix"

# Validate template exists
if [ ! -f "$TEMPLATE_PATH" ]; then
    echo -e "${RED}Error: AppxManifest template not found: $TEMPLATE_PATH${NC}"
    exit 1
fi

# Validate assets directories exist
if [ ! -d "$GOPCA_ASSETS_DIR" ]; then
    echo -e "${RED}Error: GoPCA Assets directory not found: $GOPCA_ASSETS_DIR${NC}"
    exit 1
fi

if [ ! -d "$GOCSV_ASSETS_DIR" ]; then
    echo -e "${RED}Error: GoCSV Assets directory not found: $GOCSV_ASSETS_DIR${NC}"
    echo -e "${YELLOW}Note: GoCSV MSIX assets are required for bundled package${NC}"
    echo "  Expected location: cmd/gocsv/Assets/"
    echo "  See cmd/gocsv/Assets/README.md for asset specifications"
    exit 1
fi

# Validate GoCSV assets exist
if [ ! -f "$GOCSV_ASSETS_DIR/GoCSV_Square150x150Logo.png" ] || [ ! -f "$GOCSV_ASSETS_DIR/GoCSV_Square44x44Logo.png" ]; then
    echo -e "${RED}Error: GoCSV MSIX assets are missing${NC}"
    echo "  Required files:"
    echo "    - cmd/gocsv/Assets/GoCSV_Square150x150Logo.png"
    echo "    - cmd/gocsv/Assets/GoCSV_Square44x44Logo.png"
    echo "  See cmd/gocsv/Assets/README.md for specifications"
    exit 1
fi

# Create output directory structure
echo -e "${YELLOW}[1/5] Creating package structure...${NC}"
rm -rf "$PACKAGE_DIR"
mkdir -p "$PACKAGE_DIR/Assets"

# Copy Assets (both GoPCA and GoCSV)
echo -e "${YELLOW}[2/5] Copying assets...${NC}"
echo "  GoPCA assets..."
cp -v "$GOPCA_ASSETS_DIR"/*.png "$PACKAGE_DIR/Assets/"
echo "  GoCSV assets..."
cp -v "$GOCSV_ASSETS_DIR"/*.png "$PACKAGE_DIR/Assets/"

# Copy executables (both GoPCA.exe and GoCSV.exe)
echo -e "${YELLOW}[3/5] Copying executables...${NC}"
echo "  GoPCA.exe..."
cp -v "$GOPCA_EXE" "$PACKAGE_DIR/GoPCA.exe"
echo "  GoCSV.exe..."
cp -v "$GOCSV_EXE" "$PACKAGE_DIR/GoCSV.exe"

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
