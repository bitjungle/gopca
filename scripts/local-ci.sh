#!/bin/bash
#
# local-ci.sh - Run CI checks locally
#
# This script simulates the CI environment locally, allowing developers
# to catch issues before pushing to GitHub.

set -e

echo "=== Running Local CI Checks ==="
echo ""

# Change to project root
cd "$(dirname "$0")/.."

# Setup environment
echo "1. Setting up environment..."
./scripts/ci/setup-environment.sh
echo ""

# Run tests
echo "2. Running core tests..."
if ./scripts/ci/test-core.sh; then
    echo "✓ Tests passed"
else
    echo "✗ Tests failed"
    exit 1
fi
echo ""

# Run linter checks
echo "3. Running linter checks..."
echo "  - gofmt..."
if [ -z "$(gofmt -l . | grep -v cmd/gopca-desktop)" ]; then
    echo "    ✓ Code is formatted"
else
    echo "    ✗ Code needs formatting"
    exit 1
fi

echo "  - go vet..."
if go vet ./internal/... ./pkg/... ./cmd/gopca-cli/...; then
    echo "    ✓ go vet passed"
else
    echo "    ✗ go vet failed"
    exit 1
fi
echo ""

# Build CLI
echo "4. Building CLI..."
if go build -o build/gopca-cli cmd/gopca-cli/main.go; then
    echo "✓ CLI build successful"
    echo "Binary: build/gopca-cli"
else
    echo "✗ CLI build failed"
    exit 1
fi
echo ""

# Build desktop (optional)
echo "5. Building desktop app (optional)..."
read -p "Do you want to build the desktop app? (y/N) " -n 1 -r
echo ""
if [[ $REPLY =~ ^[Yy]$ ]]; then
    if ./scripts/ci/build-desktop.sh; then
        echo "✓ Desktop build successful"
    else
        echo "✗ Desktop build failed"
        exit 1
    fi
else
    echo "⚠ Skipping desktop build"
fi

echo ""
echo "=== Local CI checks completed successfully ==="
echo ""
echo "You can now commit and push your changes with confidence!"