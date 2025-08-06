#!/bin/bash
#
# sign-macos-ci.sh - Sign macOS binaries in CI environment
#
# This script handles certificate import, signing, and cleanup for GitHub Actions.
# It expects the following environment variables from GitHub secrets:
#   - APPLE_CERTIFICATE_BASE64: Base64-encoded Developer ID certificate
#   - APPLE_CERTIFICATE_PASSWORD: Password for the certificate
#   - APPLE_IDENTITY: Full certificate identity string
#
# Usage: ./scripts/ci/sign-macos-ci.sh <binary-path>

set -e

# Check if running in CI
if [ "$CI" != "true" ]; then
    echo "This script is intended for CI environments only"
    exit 1
fi

# Check required environment variables
if [ -z "$APPLE_CERTIFICATE_BASE64" ] || [ -z "$APPLE_CERTIFICATE_PASSWORD" ] || [ -z "$APPLE_IDENTITY" ]; then
    echo "WARNING: Apple signing credentials not available, skipping signing"
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

echo "=== macOS CI Code Signing ==="
echo "Binary: $BINARY_PATH"
echo "Identity: $APPLE_IDENTITY"

# Create a temporary directory for certificate
TEMP_DIR=$(mktemp -d)
CERT_PATH="$TEMP_DIR/cert.p12"
KEYCHAIN_NAME="temp-signing-$(date +%s).keychain-db"
KEYCHAIN_PASSWORD="$(openssl rand -base64 32)"

# Cleanup function
cleanup() {
    echo "Cleaning up..."
    # Delete temporary keychain
    if security list-keychains | grep -q "$KEYCHAIN_NAME"; then
        security delete-keychain "$KEYCHAIN_NAME" 2>/dev/null || true
    fi
    # Remove temporary files
    rm -rf "$TEMP_DIR"
}

# Set trap to cleanup on exit
trap cleanup EXIT

# Decode certificate from base64
echo "Decoding certificate..."
echo "$APPLE_CERTIFICATE_BASE64" | base64 --decode > "$CERT_PATH"

# Create temporary keychain
echo "Creating temporary keychain..."
security create-keychain -p "$KEYCHAIN_PASSWORD" "$KEYCHAIN_NAME"
security set-keychain-settings -lut 21600 "$KEYCHAIN_NAME"
security unlock-keychain -p "$KEYCHAIN_PASSWORD" "$KEYCHAIN_NAME"

# Add keychain to search list
ORIGINAL_KEYCHAINS=$(security list-keychains -d user)
security list-keychains -d user -s "$KEYCHAIN_NAME" $(security list-keychains -d user | sed 's/"//g')

# Import certificate
echo "Importing certificate..."
security import "$CERT_PATH" -P "$APPLE_CERTIFICATE_PASSWORD" -A -t cert -f pkcs12 -k "$KEYCHAIN_NAME"
security set-key-partition-list -S apple-tool:,apple:,codesign: -s -k "$KEYCHAIN_PASSWORD" "$KEYCHAIN_NAME"

# Sign the binary
echo "Signing binary..."
if [[ "$BINARY_PATH" == *.app ]]; then
    # For .app bundles, use deep signing
    codesign --force --deep \
        --sign "$APPLE_IDENTITY" \
        --options runtime \
        --timestamp \
        --verbose \
        "$BINARY_PATH"
else
    # For standalone binaries
    codesign --force \
        --sign "$APPLE_IDENTITY" \
        --options runtime \
        --timestamp \
        --verbose \
        "$BINARY_PATH"
fi

# Verify signature
echo "Verifying signature..."
if codesign --verify --verbose "$BINARY_PATH"; then
    echo "✅ Signature verified successfully"
else
    echo "❌ Signature verification failed"
    exit 1
fi

# Display signature info
echo "Signature details:"
codesign --display --verbose=2 "$BINARY_PATH" 2>&1 | grep -E "Authority|TeamIdentifier|Timestamp" || true

# Restore original keychain list
security list-keychains -d user -s $(echo "$ORIGINAL_KEYCHAINS" | sed 's/"//g')

echo "=== Signing completed successfully ==="