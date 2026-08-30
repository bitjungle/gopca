#!/bin/bash
#
# check-dev-setup.sh - Check if development environment is properly set up
#

echo "Checking GoPCA development setup..."
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check Go version
echo "→ Checking Go version..."
if command -v go &> /dev/null; then
    GO_VERSION=$(go version | grep -oE '[0-9]+\.[0-9]+' | head -1)
    GO_MAJOR=$(echo $GO_VERSION | cut -d. -f1)
    GO_MINOR=$(echo $GO_VERSION | cut -d. -f2)
    
    if [ "$GO_MAJOR" -gt 1 ] || ([ "$GO_MAJOR" -eq 1 ] && [ "$GO_MINOR" -ge 26 ]); then
        echo -e "${GREEN}✓${NC} Go $GO_VERSION installed"
    else
        echo -e "${RED}✗${NC} Go version $GO_VERSION is too old. Need 1.26+"
    fi
else
    echo -e "${RED}✗${NC} Go not installed"
fi

# Check Node.js (for GUI development)
echo ""
echo "→ Checking Node.js version..."
if command -v node &> /dev/null; then
    NODE_VERSION=$(node --version | grep -oE '[0-9]+' | head -1)
    if [ "$NODE_VERSION" -ge 24 ]; then
        echo -e "${GREEN}✓${NC} Node.js $(node --version) installed"
    else
        echo -e "${YELLOW}⚠${NC} Node.js version $(node --version) is old. Recommend 24+"
    fi
else
    echo -e "${YELLOW}⚠${NC} Node.js not installed (only needed for GUI development)"
fi

# Check if Git hooks are installed
echo ""
echo "→ Checking Git hooks..."
if [ -f ".git/hooks/pre-commit" ]; then
    echo -e "${GREEN}✓${NC} Pre-commit hook installed"
else
    echo -e "${RED}✗${NC} Pre-commit hook not installed. Run: make install-hooks"
fi

# Check if dependencies are installed
echo ""
echo "→ Checking Go dependencies..."
if [ -f "go.sum" ] && [ -s "go.sum" ]; then
    echo -e "${GREEN}✓${NC} Go dependencies downloaded"
else
    echo -e "${YELLOW}⚠${NC} Go dependencies not downloaded. Run: make deps"
fi

# Check if golangci-lint is installed
echo ""
echo "→ Checking optional tools..."
if command -v golangci-lint &> /dev/null; then
    echo -e "${GREEN}✓${NC} golangci-lint installed"
else
    echo -e "${YELLOW}⚠${NC} golangci-lint not installed (optional but recommended)"
fi

# Check if wails is installed, and that it is new enough for the Go toolchain.
#
# The Wails CLI parses the project with its own bundled copy of golang.org/x/tools,
# and that pin decides which Go releases it can read. v2.12.0 bundles x/tools
# v0.30.0, which predates Go 1.27's export-data format and fails with
# "package \"context\" without types" rather than anything that names Wails.
# Reinstalling the same version does not help: the pin is in Wails' own go.mod.
WAILS_BIN=""
if command -v wails &> /dev/null; then
    WAILS_BIN="wails"
elif [ -x "$HOME/go/bin/wails" ]; then
    WAILS_BIN="$HOME/go/bin/wails"
fi

if [ -n "$WAILS_BIN" ]; then
    WAILS_VERSION=$("$WAILS_BIN" version 2>/dev/null | grep -oE 'v?[0-9]+\.[0-9]+\.[0-9]+' | head -1 | tr -d 'v')
    WAILS_MAJOR=$(echo "$WAILS_VERSION" | cut -d. -f1)
    WAILS_MINOR=$(echo "$WAILS_VERSION" | cut -d. -f2)
    if [ -z "$WAILS_VERSION" ]; then
        echo -e "${YELLOW}⚠${NC} Wails installed but its version could not be determined"
    elif [ "$WAILS_MAJOR" -gt 2 ] 2>/dev/null || { [ "$WAILS_MAJOR" -eq 2 ] && [ "$WAILS_MINOR" -ge 13 ]; } 2>/dev/null; then
        echo -e "${GREEN}✓${NC} Wails v$WAILS_VERSION installed (for GUI development)"
    else
        echo -e "${RED}✗${NC} Wails v$WAILS_VERSION is too old for Go 1.26+. Run: go install github.com/wailsapp/wails/v2/cmd/wails@v2.13.0"
    fi
else
    echo -e "${YELLOW}⚠${NC} Wails not installed (only needed for GUI development)"
fi

# Summary
echo ""
echo "================================"
echo "Summary:"
echo ""

ISSUES=0

if ! command -v go &> /dev/null || [ "$GO_MAJOR" -lt 1 ] || ([ "$GO_MAJOR" -eq 1 ] && [ "$GO_MINOR" -lt 21 ]); then
    echo -e "${RED}!${NC} Install Go 1.21+ from https://golang.org/dl/"
    ISSUES=$((ISSUES + 1))
fi

if [ ! -f ".git/hooks/pre-commit" ]; then
    echo -e "${RED}!${NC} Run 'make install-hooks' to install Git hooks"
    ISSUES=$((ISSUES + 1))
fi

if [ ! -f "go.sum" ] || [ ! -s "go.sum" ]; then
    echo -e "${YELLOW}!${NC} Run 'make deps' to download dependencies"
    ISSUES=$((ISSUES + 1))
fi

if [ $ISSUES -eq 0 ]; then
    echo -e "${GREEN}✅ Your development environment is properly set up!${NC}"
else
    echo -e "${YELLOW}⚠ Please address the above issues before contributing.${NC}"
fi