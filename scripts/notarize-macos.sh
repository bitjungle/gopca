#!/bin/bash
# Notarize macOS binaries with Apple
# Requires: xcrun notarytool (Xcode command line tools)
# 
# Prerequisites:
# 1. Create an app-specific password at https://appleid.apple.com
# 2. Add to .env file: APPLE_APP_SPECIFIC_PASSWORD=xxxx-xxxx-xxxx-xxxx
# 3. Binaries must be signed before notarization (use 'make sign' first)

set -e

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Load environment variables from .env if it exists
if [ -f ".env" ]; then
    echo -e "${BLUE}📋 Loading credentials from .env file${NC}"
    # Use a safer method to load .env that handles quotes and special characters
    set -a
    source .env
    set +a
elif [ -f "../.env" ]; then
    # If run from scripts directory
    echo -e "${BLUE}📋 Loading credentials from ../.env file${NC}"
    set -a
    source ../.env
    set +a
fi

# Apple credentials - can be overridden by environment
APPLE_ID="${APPLE_ID:-mathrune@icloud.com}"
APPLE_TEAM_ID="${APPLE_TEAM_ID:-LV599Q54BU}"
APPLE_APP_SPECIFIC_PASSWORD="${APPLE_APP_SPECIFIC_PASSWORD}"

echo "🍎 macOS Notarization Script"
echo "Apple ID: $APPLE_ID"
echo "Team ID: $APPLE_TEAM_ID"
echo ""

# Check for required password
if [ -z "$APPLE_APP_SPECIFIC_PASSWORD" ]; then
    echo -e "${RED}❌ Error: APPLE_APP_SPECIFIC_PASSWORD not found${NC}"
    echo ""
    echo "Please set your app-specific password using one of these methods:"
    echo ""
    echo "1. Add to .env file (recommended):"
    echo "   echo 'APPLE_APP_SPECIFIC_PASSWORD=xxxx-xxxx-xxxx-xxxx' >> .env"
    echo ""
    echo "2. Export in your shell:"
    echo "   export APPLE_APP_SPECIFIC_PASSWORD='xxxx-xxxx-xxxx-xxxx'"
    echo ""
    echo "3. Pass inline (note the space before command to avoid history):"
    echo "    APPLE_APP_SPECIFIC_PASSWORD='xxxx-xxxx-xxxx-xxxx' $0"
    echo ""
    echo "Create an app-specific password at: https://appleid.apple.com"
    exit 1
fi

# Function to notarize a binary
notarize_binary() {
    local binary="$1"
    local name="$2"
    local is_app="$3"
    
    if [ ! -e "$binary" ]; then
        echo -e "${YELLOW}⚠️  $name not found at $binary - skipping${NC}"
        return
    fi
    
    echo -e "${BLUE}📦 Notarizing $name...${NC}"
    
    # For CLI binaries, we need to zip them first
    if [ "$is_app" = "false" ]; then
        local zip_file="${binary}.zip"
        echo "  Creating zip archive..."
        
        # Remove old zip if exists
        rm -f "$zip_file"
        
        # Use ditto to create zip while preserving attributes
        ditto -c -k --keepParent "$binary" "$zip_file"
        
        echo "  Submitting for notarization..."
        xcrun notarytool submit "$zip_file" \
            --apple-id "$APPLE_ID" \
            --password "$APPLE_APP_SPECIFIC_PASSWORD" \
            --team-id "$APPLE_TEAM_ID" \
            --wait
        
        # Clean up zip
        rm -f "$zip_file"
        
        # Note: Cannot staple to standalone executables, only to .app, .dmg, or .pkg
        # The notarization ticket will be checked online when the binary runs
        echo "  Note: Standalone executables cannot be stapled, but notarization is complete."
        echo "  The notarization will be verified online when the binary runs."
    else
        # For .app bundles, we need to zip them first too
        local zip_file="${binary}.zip"
        echo "  Creating zip archive of app bundle..."
        
        # Remove old zip if exists
        rm -f "$zip_file"
        
        # Use ditto to create zip while preserving attributes
        ditto -c -k --keepParent "$binary" "$zip_file"
        
        echo "  Submitting app bundle for notarization..."
        xcrun notarytool submit "$zip_file" \
            --apple-id "$APPLE_ID" \
            --password "$APPLE_APP_SPECIFIC_PASSWORD" \
            --team-id "$APPLE_TEAM_ID" \
            --wait
        
        # Clean up zip
        rm -f "$zip_file"
        
        # Staple to app bundle
        echo "  Stapling ticket to app bundle..."
        xcrun stapler staple "$binary"
    fi
    
    # Verify notarization
    echo "  Verifying notarization..."
    if [ "$is_app" = "true" ]; then
        # For .app bundles, we can verify with spctl
        if spctl -a -vvv -t install "$binary" 2>&1 | grep -q "accepted"; then
            echo -e "${GREEN}✅ $name notarized and stapled successfully${NC}"
        else
            echo -e "${YELLOW}⚠️  $name notarization may have issues - check with: spctl -a -vvv -t install $binary${NC}"
        fi
    else
        # For standalone binaries, notarization is complete but can't be verified locally
        echo -e "${GREEN}✅ $name notarized successfully${NC}"
        echo "  (Verification will happen online when the binary runs)"
    fi
    echo ""
}

# Notarize CLI binary
echo "🔧 Processing CLI binary..."
notarize_binary "build/gopca-cli" "GoPCA CLI" false

# Notarize GoPCA Desktop app
echo "🖥️  Processing GoPCA Desktop app..."
notarize_binary "cmd/gopca-desktop/build/bin/gopca-desktop.app" "GoPCA Desktop" true

# Notarize GoCSV app
echo "📊 Processing GoCSV app..."
notarize_binary "cmd/gocsv/build/bin/gocsv.app" "GoCSV" true

echo -e "${GREEN}✨ Notarization complete!${NC}"
echo ""
echo "To test Gatekeeper acceptance, you can simulate quarantine:"
echo "  xattr -w com.apple.quarantine \"0081;00000000;Safari;|\" <binary>"
echo "  ./<binary>"
echo ""
echo "Your binaries are now ready for distribution!"