#!/bin/bash
#
# setup-environment.sh - Common environment setup for CI
#
# This script sets up the environment and provides information about
# the build environment for debugging purposes.

set -e

echo "=== GoPCA CI Environment Setup ==="
echo "Date: $(date)"
echo "Directory: $(pwd)"
echo ""

# OS Information
echo "=== Operating System ==="
case "$(uname -s)" in
    Linux*)
        echo "OS: Linux"
        if [ -f /etc/os-release ]; then
            . /etc/os-release
            echo "Distribution: $NAME $VERSION"
        fi
        ;;
    Darwin*)
        echo "OS: macOS"
        echo "Version: $(sw_vers -productVersion 2>/dev/null || echo 'unknown')"
        ;;
    MINGW*|MSYS*|CYGWIN*)
        echo "OS: Windows"
        echo "Shell: $SHELL"
        ;;
    *)
        echo "OS: Unknown ($(uname -s))"
        ;;
esac

# Go Environment
echo ""
echo "=== Go Environment ==="
if command -v go &> /dev/null; then
    go version
    echo "GOPATH: $(go env GOPATH)"
    echo "GOROOT: $(go env GOROOT)"
    echo "GO111MODULE: $(go env GO111MODULE)"
else
    echo "Go is not installed"
fi

# Node Environment
echo ""
echo "=== Node Environment ==="
if command -v node &> /dev/null; then
    echo "Node: $(node --version)"
    echo "NPM: $(npm --version 2>/dev/null || echo 'not found')"
else
    echo "Node.js is not installed"
fi

# Wails Environment
echo ""
echo "=== Wails Environment ==="
if command -v wails &> /dev/null; then
    wails version 2>/dev/null || echo "Wails found but version unknown"
elif [ -x "$(go env GOPATH)/bin/wails" ]; then
    echo "Wails found in GOPATH but not in PATH"
    echo "Location: $(go env GOPATH)/bin/wails"
else
    echo "Wails is not installed"
fi

# Git Information
echo ""
echo "=== Git Information ==="
if command -v git &> /dev/null; then
    echo "Git version: $(git --version)"
    echo "Current branch: $(git branch --show-current 2>/dev/null || echo 'unknown')"
    echo "Last commit: $(git log -1 --oneline 2>/dev/null || echo 'unknown')"
else
    echo "Git is not installed"
fi

echo ""
echo "=== Environment setup complete ===" 