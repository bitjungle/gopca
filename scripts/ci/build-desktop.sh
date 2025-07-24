#!/bin/bash
#
# build-desktop.sh - Build GoPCA desktop application
#
# This script ensures all prerequisites are met before building the desktop app
# and provides clear error messages when something is missing.

set -e

echo "=== Building GoPCA Desktop Application ==="

# Check for required tools
echo "Checking prerequisites..."

if ! command -v go &> /dev/null; then
    echo "ERROR: Go is not installed"
    exit 1
fi

if ! command -v node &> /dev/null; then
    echo "ERROR: Node.js is not installed"
    exit 1
fi

if ! command -v wails &> /dev/null; then
    echo "WARNING: Wails CLI not found in PATH, checking GOPATH..."
    # Try to find wails in GOPATH
    WAILS_BIN="$(go env GOPATH)/bin/wails"
    if [ -x "$WAILS_BIN" ]; then
        echo "Found wails at: $WAILS_BIN"
        # Add GOPATH/bin to PATH for this script
        export PATH="$PATH:$(go env GOPATH)/bin"
    else
        echo "ERROR: Wails CLI is not installed"
        echo "Install with: go install github.com/wailsapp/wails/v2/cmd/wails@latest"
        exit 1
    fi
fi

# Show versions
echo ""
echo "Tool versions:"
echo "- Go: $(go version)"
echo "- Node: $(node --version)"
echo "- Wails: $(wails version 2>/dev/null || echo 'version unknown')"

# Change to desktop directory
cd cmd/gopca-desktop

# Install frontend dependencies (wails build needs this)
echo ""
echo "Installing frontend dependencies..."
cd frontend
if [ -f "package.json" ]; then
    npm ci || npm install
else
    echo "ERROR: frontend/package.json not found"
    exit 1
fi

# Return to desktop directory
cd ..

# Build the desktop app
echo ""
echo "Building desktop application..."

# Detect platform if not set
if [ -z "$PLATFORM" ]; then
    case "$(uname -s)" in
        Linux*)   PLATFORM=linux ;;
        Darwin*)  PLATFORM=darwin ;;
        MINGW*|MSYS*|CYGWIN*)   PLATFORM=windows ;;
        *)        
            echo "WARNING: Unknown platform $(uname -s), defaulting to linux"
            PLATFORM=linux ;;
    esac
fi

echo "Building for platform: $PLATFORM"

# Build with wails (includes generating bindings and building frontend)
if wails build -platform "$PLATFORM"; then
    echo ""
    echo "=== Desktop application built successfully ==="
    echo "Binary location: build/bin/"
    ls -la build/bin/ 2>/dev/null || true
else
    echo ""
    echo "ERROR: Desktop build failed"
    exit 1
fi