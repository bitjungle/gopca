#!/bin/bash
# Generate application icons from master image
# Usage: ./scripts/generate-icons.sh <app-name> <source-image>
#
# app-name: "gocsv" or "gopca"
# source-image: Path to 1024x1024 PNG master image
#
# Example: ./scripts/generate-icons.sh gocsv docs/images/GoCSV-icon-1024-black.png

set -e

# Color output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Check arguments
if [ $# -ne 2 ]; then
    echo "Usage: $0 <app-name> <source-image>"
    echo "  app-name: 'gocsv' or 'gopca'"
    echo "  source-image: Path to 1024x1024 PNG master image"
    exit 1
fi

APP_NAME="$1"
SOURCE_IMAGE="$2"

# Validate app name
if [ "$APP_NAME" != "gocsv" ] && [ "$APP_NAME" != "gopca" ]; then
    echo -e "${RED}Error: app-name must be 'gocsv' or 'gopca'${NC}"
    exit 1
fi

# Set paths based on app
if [ "$APP_NAME" = "gocsv" ]; then
    APP_DIR="cmd/gocsv"
    APP_DISPLAY="GoCSV"
else
    APP_DIR="cmd/gopca-desktop"
    APP_DISPLAY="GoPCA"
fi

# Check if source image exists
if [ ! -f "$SOURCE_IMAGE" ]; then
    echo -e "${RED}Error: Source image not found: $SOURCE_IMAGE${NC}"
    exit 1
fi

# Check for required tools
if ! command -v sips &> /dev/null; then
    echo -e "${RED}Error: sips not found (macOS tool required)${NC}"
    exit 1
fi

if ! command -v iconutil &> /dev/null; then
    echo -e "${RED}Error: iconutil not found (macOS tool required)${NC}"
    exit 1
fi

# Check for ImageMagick (for Windows .ico)
HAS_MAGICK=false
if command -v magick &> /dev/null || command -v convert &> /dev/null; then
    HAS_MAGICK=true
fi

echo "🎨 Generating icons for $APP_DISPLAY from $SOURCE_IMAGE"
echo ""

# Create directories if they don't exist
mkdir -p "$APP_DIR/build"
mkdir -p "$APP_DIR/build/darwin"
mkdir -p "$APP_DIR/build/windows"
mkdir -p "$APP_DIR/build/icons"

# Step 1: Create 512x512 master for Wails
echo "Creating 512x512 master icon..."
sips -z 512 512 "$SOURCE_IMAGE" --out "$APP_DIR/build/appicon.png" >/dev/null 2>&1
echo -e "${GREEN}✓${NC} Created $APP_DIR/build/appicon.png"

# Step 2: Generate macOS .icns
echo ""
echo "Generating macOS icon (.icns)..."

# Create temporary iconset
ICONSET_DIR=$(mktemp -d)/icon.iconset
mkdir -p "$ICONSET_DIR"

# Generate all required sizes for macOS
sips -z 16 16     "$SOURCE_IMAGE" --out "$ICONSET_DIR/icon_16x16.png" >/dev/null 2>&1
sips -z 32 32     "$SOURCE_IMAGE" --out "$ICONSET_DIR/icon_16x16@2x.png" >/dev/null 2>&1
sips -z 32 32     "$SOURCE_IMAGE" --out "$ICONSET_DIR/icon_32x32.png" >/dev/null 2>&1
sips -z 64 64     "$SOURCE_IMAGE" --out "$ICONSET_DIR/icon_32x32@2x.png" >/dev/null 2>&1
sips -z 128 128   "$SOURCE_IMAGE" --out "$ICONSET_DIR/icon_128x128.png" >/dev/null 2>&1
sips -z 256 256   "$SOURCE_IMAGE" --out "$ICONSET_DIR/icon_128x128@2x.png" >/dev/null 2>&1
sips -z 256 256   "$SOURCE_IMAGE" --out "$ICONSET_DIR/icon_256x256.png" >/dev/null 2>&1
sips -z 512 512   "$SOURCE_IMAGE" --out "$ICONSET_DIR/icon_256x256@2x.png" >/dev/null 2>&1
sips -z 512 512   "$SOURCE_IMAGE" --out "$ICONSET_DIR/icon_512x512.png" >/dev/null 2>&1
sips -z 1024 1024 "$SOURCE_IMAGE" --out "$ICONSET_DIR/icon_512x512@2x.png" >/dev/null 2>&1

# Convert to .icns
iconutil -c icns "$ICONSET_DIR" -o "$APP_DIR/build/darwin/icon.icns"
cp "$APP_DIR/build/darwin/icon.icns" "$APP_DIR/build/icons/icon.icns"
echo -e "${GREEN}✓${NC} Created $APP_DIR/build/darwin/icon.icns"

# Clean up iconset
rm -rf "$(dirname "$ICONSET_DIR")"

# Step 3: Generate Windows .ico
if [ "$HAS_MAGICK" = true ]; then
    echo ""
    echo "Generating Windows icon (.ico)..."
    
    # Create temporary directory for ico PNGs
    ICO_DIR=$(mktemp -d)
    
    # Generate sizes for Windows .ico
    sips -z 16 16   "$SOURCE_IMAGE" --out "$ICO_DIR/icon_16.png" >/dev/null 2>&1
    sips -z 32 32   "$SOURCE_IMAGE" --out "$ICO_DIR/icon_32.png" >/dev/null 2>&1
    sips -z 48 48   "$SOURCE_IMAGE" --out "$ICO_DIR/icon_48.png" >/dev/null 2>&1
    sips -z 64 64   "$SOURCE_IMAGE" --out "$ICO_DIR/icon_64.png" >/dev/null 2>&1
    sips -z 128 128 "$SOURCE_IMAGE" --out "$ICO_DIR/icon_128.png" >/dev/null 2>&1
    sips -z 256 256 "$SOURCE_IMAGE" --out "$ICO_DIR/icon_256.png" >/dev/null 2>&1
    
    # Create .ico file
    if command -v magick &> /dev/null; then
        magick "$ICO_DIR"/icon_*.png "$APP_DIR/build/windows/icon.ico"
    else
        convert "$ICO_DIR"/icon_*.png "$APP_DIR/build/windows/icon.ico"
    fi
    
    cp "$APP_DIR/build/windows/icon.ico" "$APP_DIR/build/icons/icon.ico"
    echo -e "${GREEN}✓${NC} Created $APP_DIR/build/windows/icon.ico"
    
    # Clean up
    rm -rf "$ICO_DIR"
else
    echo ""
    echo -e "${YELLOW}⚠️  ImageMagick not found - skipping Windows .ico generation${NC}"
    echo "   Install ImageMagick to generate Windows icons:"
    echo "   brew install imagemagick"
fi

# Step 4: Generate Linux PNG icons
echo ""
echo "Generating Linux PNG icons..."

cp "$SOURCE_IMAGE" "$APP_DIR/build/icons/icon.png"
sips -z 256 256 "$SOURCE_IMAGE" --out "$APP_DIR/build/icons/icon-256.png" >/dev/null 2>&1
sips -z 128 128 "$SOURCE_IMAGE" --out "$APP_DIR/build/icons/icon-128.png" >/dev/null 2>&1
sips -z 64 64   "$SOURCE_IMAGE" --out "$APP_DIR/build/icons/icon-64.png" >/dev/null 2>&1
sips -z 48 48   "$SOURCE_IMAGE" --out "$APP_DIR/build/icons/icon-48.png" >/dev/null 2>&1
sips -z 32 32   "$SOURCE_IMAGE" --out "$APP_DIR/build/icons/icon-32.png" >/dev/null 2>&1
sips -z 16 16   "$SOURCE_IMAGE" --out "$APP_DIR/build/icons/icon-16.png" >/dev/null 2>&1
echo -e "${GREEN}✓${NC} Created Linux PNG icons in $APP_DIR/build/icons/"

echo ""
echo -e "${GREEN}✨ Icon generation complete for $APP_DISPLAY!${NC}"
echo ""
echo "Generated files:"
echo "  • $APP_DIR/build/appicon.png (512x512 master)"
echo "  • $APP_DIR/build/darwin/icon.icns (macOS)"
if [ "$HAS_MAGICK" = true ]; then
    echo "  • $APP_DIR/build/windows/icon.ico (Windows)"
fi
echo "  • $APP_DIR/build/icons/*.png (Linux/general)"