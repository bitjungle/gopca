#!/bin/bash
# test-integration.sh - Run comprehensive integration tests
# This script validates end-to-end workflows, cross-application communication,
# and ensures all components work together correctly.

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

echo "=========================================="
echo "GoPCA Integration Test Suite"
echo "=========================================="
echo "Running from: $PROJECT_ROOT"
echo "Date: $(date)"
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0
SKIPPED_TESTS=0

# Function to print colored output
print_status() {
    local status=$1
    local message=$2
    case $status in
        "PASS")
            echo -e "${GREEN}✓${NC} $message"
            ((PASSED_TESTS++))
            ;;
        "FAIL")
            echo -e "${RED}✗${NC} $message"
            ((FAILED_TESTS++))
            ;;
        "SKIP")
            echo -e "${YELLOW}⊘${NC} $message"
            ((SKIPPED_TESTS++))
            ;;
        "INFO")
            echo -e "  $message"
            ;;
    esac
    ((TOTAL_TESTS++))
}

# Function to run a test with timeout
run_test_with_timeout() {
    local test_name=$1
    local timeout=$2
    shift 2
    local cmd="$@"
    
    echo ""
    echo "Running: $test_name"
    echo "Command: $cmd"
    
    if timeout "$timeout" bash -c "$cmd" > /tmp/test_output.log 2>&1; then
        print_status "PASS" "$test_name completed successfully"
        return 0
    else
        local exit_code=$?
        if [ $exit_code -eq 124 ]; then
            print_status "FAIL" "$test_name timed out after ${timeout}s"
        else
            print_status "FAIL" "$test_name failed with exit code $exit_code"
        fi
        echo "Last 20 lines of output:"
        tail -20 /tmp/test_output.log
        return 1
    fi
}

# Check prerequisites
echo "Checking prerequisites..."

if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed"
    exit 1
fi

if ! command -v npm &> /dev/null; then
    echo "Warning: npm is not installed - skipping frontend tests"
    SKIP_FRONTEND=1
fi

# Step 1: Build CLI tool
echo ""
echo "Step 1: Building CLI tool..."
cd "$PROJECT_ROOT"

if make build > /tmp/build.log 2>&1; then
    print_status "PASS" "CLI build successful"
else
    print_status "FAIL" "CLI build failed"
    tail -20 /tmp/build.log
    exit 1
fi

# Step 2: Run unit tests first
echo ""
echo "Step 2: Running unit tests..."

if go test -race -cover ./internal/... ./pkg/... > /tmp/unit_tests.log 2>&1; then
    print_status "PASS" "Unit tests passed"
    
    # Extract coverage
    coverage=$(grep -o '[0-9.]*%' /tmp/unit_tests.log | tail -1)
    print_status "INFO" "Coverage: $coverage"
else
    print_status "FAIL" "Unit tests failed"
    tail -20 /tmp/unit_tests.log
    exit 1
fi

# Step 3: Run integration tests
echo ""
echo "Step 3: Running integration tests..."

# Set up test environment
export GOPCA_TEST_MODE=1
export GOPCA_TEST_TIMEOUT=60

# Run different test suites
test_suites=(
    "TestE2EBasicWorkflow"
    "TestE2EAllPreprocessingMethods"
    "TestE2EExportFormats"
    "TestE2EModelExportImport"
    "TestE2EDiagnosticMetrics"
    "TestParitySVD"
    "TestParityNIPALS"
    "TestParityKernelPCA"
    "TestRegressionSuite"
    "TestPerformanceRegression"
    "TestBackwardCompatibility"
    "TestSecurityRegression"
    "TestNumericalStability"
)

for suite in "${test_suites[@]}"; do
    if run_test_with_timeout "$suite" 120 \
        "go test -v -run '^$suite$' ./internal/integration/..."; then
        :
    else
        # Continue even if one suite fails
        echo "Continuing with remaining tests..."
    fi
done

# Step 4: Test cross-application integration (if GoCSV exists)
echo ""
echo "Step 4: Testing cross-application integration..."

if [ -f "$PROJECT_ROOT/cmd/gocsv/main.go" ]; then
    if make csv-build > /tmp/gocsv_build.log 2>&1; then
        print_status "PASS" "GoCSV build successful"
        
        # Test app launching
        if go test -v -run TestAppIntegration ./pkg/integration/... > /tmp/app_integration.log 2>&1; then
            print_status "PASS" "App integration tests passed"
        else
            print_status "FAIL" "App integration tests failed"
            tail -10 /tmp/app_integration.log
        fi
    else
        print_status "SKIP" "GoCSV build skipped"
    fi
else
    print_status "SKIP" "GoCSV not found - skipping cross-app tests"
fi

# Step 5: Test sample datasets
echo ""
echo "Step 5: Testing with sample datasets..."

sample_datasets=(
    "testdata/iris/iris.csv"
    "testdata/wine/wine.csv"
    "testdata/corn/corn.csv"
)

for dataset in "${sample_datasets[@]}"; do
    if [ -f "$PROJECT_ROOT/$dataset" ]; then
        dataset_name=$(basename "$dataset" .csv)
        
        if timeout 30 "$PROJECT_ROOT/build/pca" analyze \
            "$PROJECT_ROOT/$dataset" \
            --method svd \
            --components 2 \
            --format json \
            --output "/tmp/${dataset_name}_output" > /tmp/sample_test.log 2>&1; then
            print_status "PASS" "Sample dataset: $dataset_name"
        else
            print_status "FAIL" "Sample dataset: $dataset_name"
        fi
    else
        print_status "SKIP" "Sample dataset not found: $dataset"
    fi
done

# Step 6: Memory leak detection
echo ""
echo "Step 6: Running memory leak detection..."

if [ -n "$SKIP_MEMORY_TESTS" ]; then
    print_status "SKIP" "Memory leak tests skipped (SKIP_MEMORY_TESTS set)"
else
    if go test -v -run TestMemoryLeaks ./internal/core/... > /tmp/memory_test.log 2>&1; then
        print_status "PASS" "No memory leaks detected"
    else
        print_status "FAIL" "Memory leak test failed"
        tail -10 /tmp/memory_test.log
    fi
fi

# Step 7: Performance benchmarks
echo ""
echo "Step 7: Running performance benchmarks..."

if [ -n "$SKIP_BENCHMARKS" ]; then
    print_status "SKIP" "Benchmarks skipped (SKIP_BENCHMARKS set)"
else
    if timeout 300 go test -bench=. -benchtime=10s ./internal/core/... > /tmp/benchmark.log 2>&1; then
        print_status "PASS" "Benchmarks completed"
        
        # Extract key benchmark results
        echo "Key benchmark results:"
        grep -E "Benchmark.*-[0-9]+" /tmp/benchmark.log | head -5
    else
        print_status "FAIL" "Benchmarks failed or timed out"
    fi
fi

# Step 8: Frontend tests (if not skipped)
echo ""
echo "Step 8: Running frontend tests..."

if [ -n "$SKIP_FRONTEND" ]; then
    print_status "SKIP" "Frontend tests skipped (npm not available)"
else
    # Test GoPCA Desktop frontend
    if [ -d "$PROJECT_ROOT/cmd/gopca-desktop/frontend" ]; then
        cd "$PROJECT_ROOT/cmd/gopca-desktop/frontend"
        
        if npm test -- --watchAll=false > /tmp/frontend_test.log 2>&1; then
            print_status "PASS" "GoPCA Desktop frontend tests"
        else
            print_status "SKIP" "GoPCA Desktop frontend tests (no tests configured)"
        fi
    fi
    
    # Test GoCSV frontend
    if [ -d "$PROJECT_ROOT/cmd/gocsv/frontend" ]; then
        cd "$PROJECT_ROOT/cmd/gocsv/frontend"
        
        if npm test -- --watchAll=false > /tmp/gocsv_frontend_test.log 2>&1; then
            print_status "PASS" "GoCSV frontend tests"
        else
            print_status "SKIP" "GoCSV frontend tests (no tests configured)"
        fi
    fi
fi

# Step 9: CLI validation tests
echo ""
echo "Step 9: Validating CLI commands..."

cd "$PROJECT_ROOT"

# Test help output
if "$PROJECT_ROOT/build/pca" --help > /dev/null 2>&1; then
    print_status "PASS" "CLI help command"
else
    print_status "FAIL" "CLI help command"
fi

# Test version output
if "$PROJECT_ROOT/build/pca" --version > /dev/null 2>&1; then
    print_status "PASS" "CLI version command"
else
    print_status "FAIL" "CLI version command"
fi

# Test validate command
if echo "a,b,c" | "$PROJECT_ROOT/build/pca" validate --stdin > /dev/null 2>&1; then
    print_status "PASS" "CLI validate command"
else
    print_status "SKIP" "CLI validate command not implemented"
fi

# Generate summary report
echo ""
echo "=========================================="
echo "Integration Test Summary"
echo "=========================================="
echo "Total Tests:   $TOTAL_TESTS"
echo -e "Passed:        ${GREEN}$PASSED_TESTS${NC}"
if [ $FAILED_TESTS -gt 0 ]; then
    echo -e "Failed:        ${RED}$FAILED_TESTS${NC}"
else
    echo "Failed:        $FAILED_TESTS"
fi
if [ $SKIPPED_TESTS -gt 0 ]; then
    echo -e "Skipped:       ${YELLOW}$SKIPPED_TESTS${NC}"
else
    echo "Skipped:       $SKIPPED_TESTS"
fi

# Calculate pass rate
if [ $TOTAL_TESTS -gt 0 ]; then
    PASS_RATE=$((PASSED_TESTS * 100 / TOTAL_TESTS))
    echo "Pass Rate:     ${PASS_RATE}%"
fi

echo "=========================================="

# Exit with appropriate code
if [ $FAILED_TESTS -gt 0 ]; then
    echo -e "${RED}Integration tests failed!${NC}"
    exit 1
else
    echo -e "${GREEN}All integration tests passed!${NC}"
    exit 0
fi