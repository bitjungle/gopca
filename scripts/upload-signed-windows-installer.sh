#!/bin/bash
#
# Upload signed Windows installer to GitHub release
# Usage: ./upload-signed-windows-installer.sh <version> <signed-installer-path>
# Example: ./upload-signed-windows-installer.sh v1.0.2 ~/Downloads/GoPCA-Setup-v1.0.2-signed.exe
#
# This script:
# 1. Uploads the signed installer to the specified release
# 2. Provides instructions for updating the release notes
#
# Prerequisites:
# - GitHub CLI (gh) installed and authenticated
# - The release must already exist

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check arguments
if [ $# -ne 2 ]; then
    echo -e "${RED}Error: Incorrect number of arguments${NC}"
    echo "Usage: $0 <version> <signed-installer-path>"
    echo "Example: $0 v1.0.2 ~/Downloads/GoPCA-Setup-v1.0.2-signed.exe"
    exit 1
fi

VERSION="$1"
SIGNED_INSTALLER="$2"

# Validate version format
if [[ ! "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo -e "${RED}Error: Invalid version format${NC}"
    echo "Version must be in format: vX.X.X (e.g., v1.0.2)"
    exit 1
fi

# Check if signed installer exists
if [ ! -f "$SIGNED_INSTALLER" ]; then
    echo -e "${RED}Error: Signed installer file not found: $SIGNED_INSTALLER${NC}"
    exit 1
fi

# Check if gh is installed
if ! command -v gh &> /dev/null; then
    echo -e "${RED}Error: GitHub CLI (gh) is not installed${NC}"
    echo "Install it from: https://cli.github.com/"
    exit 1
fi

# Check if authenticated
if ! gh auth status &> /dev/null; then
    echo -e "${RED}Error: Not authenticated with GitHub${NC}"
    echo "Run: gh auth login"
    exit 1
fi

echo -e "${GREEN}Uploading signed Windows installer for release $VERSION${NC}"
echo "File: $SIGNED_INSTALLER"

# Get the expected filename
EXPECTED_NAME="GoPCA-Setup-${VERSION}-signed.exe"
UPLOAD_FILE="$SIGNED_INSTALLER"

# If the file doesn't have the expected name, copy it
if [ "$(basename "$SIGNED_INSTALLER")" != "$EXPECTED_NAME" ]; then
    echo -e "${YELLOW}Renaming file to: $EXPECTED_NAME${NC}"
    UPLOAD_FILE="/tmp/$EXPECTED_NAME"
    cp "$SIGNED_INSTALLER" "$UPLOAD_FILE"
fi

# Check if release exists
echo "Checking if release $VERSION exists..."
if ! gh release view "$VERSION" &> /dev/null; then
    echo -e "${RED}Error: Release $VERSION does not exist${NC}"
    echo "Available releases:"
    gh release list --limit 5
    exit 1
fi

# Upload the signed installer
echo "Uploading signed installer to release..."
if gh release upload "$VERSION" "$UPLOAD_FILE" --clobber; then
    echo -e "${GREEN}✓ Successfully uploaded signed installer${NC}"
else
    echo -e "${RED}Error: Failed to upload signed installer${NC}"
    exit 1
fi

# Clean up temp file if we created one
if [ "$UPLOAD_FILE" != "$SIGNED_INSTALLER" ]; then
    rm "$UPLOAD_FILE"
fi

# Get the release URL
RELEASE_URL=$(gh release view "$VERSION" --json url -q .url)

echo ""
echo -e "${GREEN}=== Next Steps ===${NC}"
echo ""
echo "1. Edit the release notes at:"
echo "   $RELEASE_URL"
echo ""
echo "2. Update the Windows installer download link:"
echo "   Change: GoPCA-Setup-${VERSION}.exe"
echo "   To:     GoPCA-Setup-${VERSION}-signed.exe"
echo ""
echo "3. Add '(digitally signed)' note next to the Windows installer"
echo ""
echo -e "${YELLOW}Tip: You can also edit directly with:${NC}"
echo "   gh release edit $VERSION --notes-file <updated-notes.md>"
echo ""
echo -e "${GREEN}✓ Signed installer upload complete!${NC}"