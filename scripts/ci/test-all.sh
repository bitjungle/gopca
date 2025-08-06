#!/bin/bash
#
# test-all.sh - Comprehensive test runner for all GoPCA packages
#
# This script runs tests for all packages including Wails applications,
# handling platform-specific issues and providing clear error reporting.

set -e

echo "=== Running Comprehensive GoPCA Tests ==="

# Show environment info
echo "Go version: $(go version)"
echo "OS: $(uname -s)"
echo "Architecture: $(uname -m)"

# Download dependencies if needed
echo ""
echo "Downloading dependencies..."
go mod download

# Run tests for core packages
echo ""
echo "=== Testing Core Packages ==="
CORE_PACKAGES="./internal/cli ./internal/core ./internal/io ./internal/utils ./pkg/types"

echo "Testing: $CORE_PACKAGES"
if go test -v -race -cover $CORE_PACKAGES; then
    echo "✓ Core packages tests passed"
else
    echo "✗ Core packages tests failed"
    exit 1
fi

# Run tests for GoCSV (excluding app_test.go which requires Wails context)
echo ""
echo "=== Testing GoCSV ==="
echo "Note: Testing command logic only (app tests require Wails runtime)"

# Test only commands_test.go which doesn't require Wails context
cd cmd/gocsv
if go test -v -cover -run "TestMultiStepUndoRedo|TestUndoRedoState" .; then
    echo "✓ GoCSV command tests passed"
else
    echo "✗ GoCSV command tests failed"
    exit 1
fi
cd ../..

# Run tests for GoPCA Desktop
echo ""
echo "=== Testing GoPCA Desktop ==="
if go test -v -cover ./cmd/gopca-desktop; then
    echo "✓ GoPCA Desktop tests passed"
else
    echo "✗ GoPCA Desktop tests failed"
    exit 1
fi

# Show overall coverage summary
echo ""
echo "=== Overall Test Summary ==="
echo "Core packages: ✓"
echo "GoCSV commands: ✓"
echo "GoPCA Desktop: ✓"

# Calculate combined coverage (optional)
echo ""
echo "=== Coverage Report ==="
go test -cover $CORE_PACKAGES ./cmd/gocsv ./cmd/gopca-desktop 2>/dev/null | grep -E "coverage:|ok" || true

echo ""
echo "=== All tests completed successfully ===" 