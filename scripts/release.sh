#!/bin/bash

# Script to create a release after the release PR has been merged
# Usage: ./scripts/release.sh <version>
# Example: ./scripts/release.sh v0.9.0

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if version argument is provided
if [ -z "$1" ]; then
    echo -e "${RED}Error: Version argument is required${NC}"
    echo "Usage: $0 <version>"
    echo "Example: $0 v0.9.0"
    exit 1
fi

VERSION=$1

# Validate version format (should be vX.Y.Z)
if ! [[ "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo -e "${RED}Error: Invalid version format${NC}"
    echo "Version should be in format vX.Y.Z (e.g., v0.9.0)"
    exit 1
fi

echo -e "${GREEN}Creating release ${VERSION}${NC}"

# 1. Ensure we're on main branch
CURRENT_BRANCH=$(git branch --show-current)
if [ "$CURRENT_BRANCH" != "main" ]; then
    echo -e "${RED}Error: Must be on main branch to create a release${NC}"
    echo "Current branch: $CURRENT_BRANCH"
    exit 1
fi
echo "✓ On main branch"

# 2. Pull latest changes
echo -e "${YELLOW}Pulling latest changes...${NC}"
git pull origin main

# 3. Check that the version in wails.json matches
VERSION_NO_V=${VERSION:1}
WAILS_VERSION=$(grep '"productVersion"' cmd/gopca-desktop/wails.json | sed 's/.*"productVersion": "\(.*\)".*/\1/')
if [ "$WAILS_VERSION" != "$VERSION_NO_V" ]; then
    echo -e "${RED}Error: Version mismatch${NC}"
    echo "Expected version in wails.json: $VERSION_NO_V"
    echo "Found version in wails.json: $WAILS_VERSION"
    echo "Did you forget to merge the release PR?"
    exit 1
fi
echo "✓ Version in wails.json matches"

# 4. Check if tag already exists
if git rev-parse "$VERSION" >/dev/null 2>&1; then
    echo -e "${RED}Error: Tag $VERSION already exists${NC}"
    exit 1
fi

# 5. Create and push tag
echo -e "${YELLOW}Creating tag ${VERSION}...${NC}"
git tag -a "$VERSION" -m "Release $VERSION"
git push origin "$VERSION"
echo -e "${GREEN}✓ Tag created and pushed${NC}"

# 6. Build release artifacts
echo -e "${YELLOW}Building release artifacts...${NC}"
echo "This is where we would build all platform binaries"
echo "For now, we'll create the release without artifacts"

# 7. Create GitHub release
echo -e "${YELLOW}Creating GitHub release...${NC}"

# Determine if this is a pre-release
if [[ "$VERSION" =~ -rc\.|alpha|beta ]]; then
    PRERELEASE_FLAG="--prerelease"
    echo "Marking as pre-release"
else
    PRERELEASE_FLAG=""
fi

# Create the release using gh CLI
gh release create "$VERSION" \
    --title "GoPCA ${VERSION}" \
    --generate-notes \
    $PRERELEASE_FLAG

echo -e "${GREEN}✓ Release ${VERSION} created successfully!${NC}"
echo ""
echo "Next steps:"
echo "1. Review the release at: https://github.com/bitjungle/gopca/releases/tag/${VERSION}"
echo "2. Edit the release notes if needed"
echo "3. Upload release artifacts when ready"

# Optional: Open the release page
echo ""
read -p "Would you like to open the release page in your browser? (y/n) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    gh release view "$VERSION" --web
fi