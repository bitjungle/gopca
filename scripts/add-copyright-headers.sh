#!/bin/bash

# Script to add copyright headers to all source files
# Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.

# The copyright header to add
HEADER="// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications."

# Function to add header to a file
add_header() {
    local file="$1"
    local temp_file="${file}.tmp"
    
    # Check if file already has copyright header
    if head -n 1 "$file" | grep -q "Copyright"; then
        echo "Skipping $file (already has copyright)"
        return
    fi
    
    # Add header to the file
    echo "$HEADER" > "$temp_file"
    echo "" >> "$temp_file"
    cat "$file" >> "$temp_file"
    mv "$temp_file" "$file"
    echo "Added header to $file"
}

echo "Adding copyright headers to Go files..."
# Find all Go files, excluding vendor and generated files
find . -name "*.go" -type f \
    -not -path "./vendor/*" \
    -not -path "*/wailsjs/*" \
    -not -path "./node_modules/*" | while read -r file; do
    add_header "$file"
done

echo ""
echo "Adding copyright headers to TypeScript/JavaScript files..."
# Find all TypeScript and JavaScript files
find . \( -name "*.ts" -o -name "*.tsx" -o -name "*.js" -o -name "*.jsx" \) -type f \
    -not -path "./vendor/*" \
    -not -path "*/wailsjs/*" \
    -not -path "./node_modules/*" \
    -not -path "./packages/*/node_modules/*" \
    -not -path "./cmd/*/frontend/node_modules/*" \
    -not -path "./cmd/*/frontend/dist/*" \
    -not -path "./packages/*/dist/*" | while read -r file; do
    add_header "$file"
done

echo ""
echo "Copyright headers added successfully!"