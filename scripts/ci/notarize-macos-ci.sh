#!/bin/bash
#
# notarize-macos-ci.sh - Notarize macOS binaries in CI environment
#
# This script submits signed binaries to Apple for notarization.
# It expects the following environment variables from GitHub secrets:
#   - APPLE_ID: Apple Developer account email
#   - APPLE_APP_SPECIFIC_PASSWORD: App-specific password for notarization
#   - APPLE_TEAM_ID: Apple Developer Team ID
#
# Usage: ./scripts/ci/notarize-macos-ci.sh <binary-path>

set -e

# Check if running in CI
if [ "$CI" != "true" ]; then
    echo "This script is intended for CI environments only"
    exit 1
fi

# Check required environment variables
if [ -z "$APPLE_ID" ] || [ -z "$APPLE_APP_SPECIFIC_PASSWORD" ] || [ -z "$APPLE_TEAM_ID" ]; then
    echo "WARNING: Apple notarization credentials not available, skipping notarization"
    exit 0
fi

BINARY_PATH="$1"
if [ -z "$BINARY_PATH" ]; then
    echo "ERROR: Binary path not provided"
    echo "Usage: $0 <binary-path>"
    exit 1
fi

if [ ! -e "$BINARY_PATH" ]; then
    echo "ERROR: Binary not found at: $BINARY_PATH"
    exit 1
fi

echo "=== macOS CI Notarization ==="
echo "Binary: $BINARY_PATH"
echo "Apple ID: $APPLE_ID"
echo "Team ID: $APPLE_TEAM_ID"

# Create temporary directory for working files
TEMP_DIR=$(mktemp -d)
ZIP_PATH="$TEMP_DIR/$(basename "$BINARY_PATH").zip"

# Cleanup function
cleanup() {
    echo "Cleaning up temporary files..."
    rm -rf "$TEMP_DIR"
}

# Set trap to cleanup on exit
trap cleanup EXIT

# Check if binary is signed
echo "Checking signature..."
if ! codesign --verify "$BINARY_PATH" 2>/dev/null; then
    echo "ERROR: Binary is not signed. Please sign it first."
    exit 1
fi

# Create ZIP archive for notarization
echo "Creating ZIP archive for notarization..."
if [[ "$BINARY_PATH" == *.app ]]; then
    # For .app bundles, use ditto to preserve structure
    ditto -c -k --keepParent "$BINARY_PATH" "$ZIP_PATH"
else
    # For standalone binaries, create a simple zip
    # We need to preserve the executable permissions
    (cd "$(dirname "$BINARY_PATH")" && zip -r "$ZIP_PATH" "$(basename "$BINARY_PATH")")
fi

echo "Archive created: $(du -h "$ZIP_PATH" | cut -f1)"

# Submit for notarization
echo "Submitting to Apple for notarization..."
echo "This may take several minutes..."

# Create a notarization profile to avoid passing credentials directly
# This is more secure as it doesn't expose credentials in process lists
xcrun notarytool store-credentials "ci-notarization" \
    --apple-id "$APPLE_ID" \
    --password "$APPLE_APP_SPECIFIC_PASSWORD" \
    --team-id "$APPLE_TEAM_ID" \
    --validate \
    2>&1 | grep -v "password" || true

# Submit and wait for notarization
SUBMISSION_ID=""
if SUBMISSION_OUTPUT=$(xcrun notarytool submit "$ZIP_PATH" \
    --keychain-profile "ci-notarization" \
    --wait \
    --timeout 30m \
    --verbose 2>&1); then
    
    echo "✅ Notarization succeeded"
    
    # Extract submission ID for logging
    SUBMISSION_ID=$(echo "$SUBMISSION_OUTPUT" | grep -E "id: [a-f0-9-]+" | head -1 | awk '{print $2}')
    if [ -n "$SUBMISSION_ID" ]; then
        echo "Submission ID: $SUBMISSION_ID"
    fi
    
    # For .app bundles, staple the ticket
    if [[ "$BINARY_PATH" == *.app ]]; then
        echo "Stapling notarization ticket to app bundle..."
        if xcrun stapler staple "$BINARY_PATH"; then
            echo "✅ Ticket stapled successfully"
            
            # Verify stapling
            echo "Verifying notarization..."
            if spctl -a -vvv -t install "$BINARY_PATH" 2>&1 | grep -q "accepted"; then
                echo "✅ App bundle is properly notarized and will run without Gatekeeper warnings"
            else
                echo "⚠️ Notarization verification had warnings. The app should still run."
            fi
        else
            echo "⚠️ Failed to staple ticket, but notarization is complete"
            echo "The app will still run, but will need online verification"
        fi
    else
        echo "ℹ️ Standalone executables cannot be stapled, but notarization is complete"
        echo "The binary will be verified online when first run"
    fi
else
    echo "❌ Notarization failed"
    echo "Error output:"
    echo "$SUBMISSION_OUTPUT" | grep -v "password" || true
    
    # Try to get more details about the failure
    if [ -n "$SUBMISSION_ID" ]; then
        echo "Attempting to get notarization log..."
        xcrun notarytool log "$SUBMISSION_ID" \
            --keychain-profile "ci-notarization" \
            2>&1 | grep -v "password" || true
    fi
    
    # Non-fatal for CI builds
    echo "WARNING: Notarization failed, but continuing build"
    exit 0
fi

# Clean up stored credentials
xcrun notarytool delete-credentials "ci-notarization" 2>/dev/null || true

echo "=== Notarization completed successfully ==="