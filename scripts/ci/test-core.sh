#!/bin/bash
#
# test-core.sh - Run tests for core GoPCA packages
#
# This script runs tests excluding the desktop package which requires
# frontend build artifacts that may not be available during testing.

set -e

echo "=== Running Core GoPCA Tests ==="

# Show Go version
echo "Go version: $(go version)"

# Download dependencies if needed
echo "Downloading dependencies..."
go mod download

# Run tests with proper exclusions
echo "Running tests (excluding desktop package)..."

# Method 1: Explicitly list packages to test
# This is the most reliable method as it doesn't require Go to parse the desktop package
# We only test packages that have test files to avoid Windows CI issues
# Note: GoCSV app tests require Wails context and should be run separately

# First run core packages and GoPCA Desktop tests
if ! go test -v -cover ./internal/cli ./internal/core ./internal/io ./internal/utils ./pkg/types ./cmd/gopca-desktop; then
    echo "✗ Core tests failed"
    exit 1
fi

# Then run GoCSV tests that don't require Wails context
cd cmd/gocsv
if ! go test -v -cover -run "TestMultiStepUndoRedo|TestUndoRedoState" .; then
    echo "✗ GoCSV tests failed"
    cd ../..
    exit 1
fi
cd ../..

echo "✓ All core tests passed"

# Show coverage summary
echo ""
echo "=== Coverage Summary ==="
go test -cover ./internal/cli ./internal/core ./internal/io ./internal/utils ./pkg/types ./cmd/gocsv ./cmd/gopca-desktop 2>/dev/null | grep -E "coverage:|ok" || true

echo ""
echo "=== Core tests completed successfully ===" 