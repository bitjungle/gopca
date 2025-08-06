#!/bin/bash
# Sign macOS binaries with Apple Developer ID
# Usage: ./scripts/sign-macos.sh [identity]
#
# If identity is not provided, uses the APPLE_DEVELOPER_ID environment variable
# or defaults to "Developer ID Application: Rune Mathisen (LV599Q54BU)"

set -e

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Get signing identity from argument, environment, or default
IDENTITY="${1:-${APPLE_DEVELOPER_ID:-Developer ID Application: Rune Mathisen (LV599Q54BU)}}"

echo "🔐 macOS Code Signing Script"
echo "Identity: $IDENTITY"
echo ""

# Function to sign a binary
sign_binary() {
    local binary="$1"
    local name="$2"
    
    if [ ! -e "$binary" ]; then
        echo -e "${YELLOW}⚠️  $name not found at $binary - skipping${NC}"
        return
    fi
    
    echo -e "Signing $name..."
    
    # Sign with hardened runtime, timestamp, and deep signing for .app bundles
    if [[ "$binary" == *.app ]]; then
        codesign --force --deep --sign "$IDENTITY" \
            --options runtime \
            --timestamp \
            "$binary"
    else
        codesign --force --sign "$IDENTITY" \
            --options runtime \
            --timestamp \
            "$binary"
    fi
    
    # Verify signature
    if codesign --verify --verbose "$binary" 2>&1 | grep -q "valid on disk"; then
        echo -e "${GREEN}✅ $name signed successfully${NC}"
    else
        echo -e "${RED}❌ Failed to sign $name${NC}"
        exit 1
    fi
    echo ""
}

# Sign CLI binary
echo "📦 Signing CLI binary..."
sign_binary "build/gopca-cli" "GoPCA CLI"

# Sign GoPCA Desktop app
echo "📦 Signing GoPCA Desktop app..."
sign_binary "cmd/gopca-desktop/build/bin/gopca-desktop.app" "GoPCA Desktop"

# Sign GoCSV app
echo "📦 Signing GoCSV app..."
sign_binary "cmd/gocsv/build/bin/gocsv.app" "GoCSV"

echo -e "${GREEN}✨ All binaries signed successfully!${NC}"
echo ""
echo "To verify signatures manually, run:"
echo "  codesign --verify --verbose <binary>"
echo ""
echo "To check signature details, run:"
echo "  codesign --display --verbose <binary>"