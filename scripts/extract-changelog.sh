#!/bin/bash

# Script to extract changelog content for a specific version
# Usage: ./scripts/extract-changelog.sh <version>
# Example: ./scripts/extract-changelog.sh v1.1.1

set -e

# Check if version argument is provided
if [ -z "$1" ]; then
    echo "Error: Version argument is required"
    echo "Usage: $0 <version>"
    echo "Example: $0 v1.1.1"
    exit 1
fi

VERSION=$1
VERSION_NO_V=${VERSION:1}  # Remove 'v' prefix

# Path to CHANGELOG.md
CHANGELOG="CHANGELOG.md"

if [ ! -f "$CHANGELOG" ]; then
    echo "Error: CHANGELOG.md not found"
    exit 1
fi

# Extract the content for the specific version
# This looks for a line starting with ## [version] and captures everything
# until the next ## section or end of file
awk -v ver="$VERSION_NO_V" '
    /^## \[/ {
        if (found) exit
        if (index($0, "[" ver "]") > 0) {
            found = 1
            next
        }
    }
    found && /^## \[/ { exit }
    found { print }
' "$CHANGELOG"

# If nothing found, return empty
if [ $? -ne 0 ]; then
    echo ""
fi