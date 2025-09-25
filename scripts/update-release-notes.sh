#!/bin/bash

# Script to update existing GitHub release notes with proper changelog content
# Usage: ./scripts/update-release-notes.sh <version>
# Or: ./scripts/update-release-notes.sh all  (to update all recent releases)

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

update_release() {
    local VERSION=$1
    echo -e "${YELLOW}Updating release notes for ${VERSION}...${NC}"

    # Extract changelog content
    CHANGELOG_CONTENT=$(./scripts/extract-changelog.sh "$VERSION")

    if [ -z "$CHANGELOG_CONTENT" ]; then
        echo "No changelog content found for $VERSION, skipping..."
        return
    fi

    # Get the current release body
    CURRENT_BODY=$(gh release view "$VERSION" --json body --jq '.body')

    # Check if already updated (if it starts with "## What's Changed in")
    if echo "$CURRENT_BODY" | grep -q "^## What's Changed in ${VERSION}"; then
        echo "Release $VERSION already has updated format, skipping..."
        return
    fi

    # Extract the downloads section (everything after "## Downloads" or similar)
    DOWNLOADS_SECTION=$(echo "$CURRENT_BODY" | awk '/^## Downloads/,EOF')

    # If no Downloads section found, try to preserve everything after the auto-generated content
    if [ -z "$DOWNLOADS_SECTION" ]; then
        DOWNLOADS_SECTION=$(echo "$CURRENT_BODY" | awk '/^---/,EOF')
    fi

    # Create new body with changelog at the top
    cat > release_body_temp.md << EOF
## What's Changed in ${VERSION}

${CHANGELOG_CONTENT}

---

${DOWNLOADS_SECTION}
EOF

    # Update the release
    gh release edit "$VERSION" --notes-file release_body_temp.md

    echo -e "${GREEN}✓ Updated release notes for ${VERSION}${NC}"
    rm release_body_temp.md
}

# Main script
if [ -z "$1" ]; then
    echo "Usage: $0 <version|all>"
    echo "Example: $0 v1.1.1"
    echo "Example: $0 all"
    exit 1
fi

if [ "$1" = "all" ]; then
    # Update recent releases
    VERSIONS=$(gh release list --limit 10 --json tagName --jq '.[].tagName')
    for VERSION in $VERSIONS; do
        update_release "$VERSION"
        echo ""
    done
else
    # Update specific version
    update_release "$1"
fi

echo -e "${GREEN}✓ Release notes update complete!${NC}"