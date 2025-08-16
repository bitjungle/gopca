# Integration Testing Guide

This guide covers the comprehensive integration testing framework for the GoPCA monorepo, including end-to-end workflows, cross-application communication, and platform-specific testing.

## Overview

The integration test suite validates that all components of GoPCA work together correctly:
- CLI tool functionality
- GUI applications (GoPCA Desktop and GoCSV)
- Cross-application communication
- Platform-specific behaviors
- Performance and security requirements

## Test Categories

### 1. End-to-End (E2E) Tests
Located in `internal/integration/e2e_test.go`

Tests complete workflows from data input to results export:
- CSV loading and parsing
- PCA analysis with all methods (SVD, NIPALS, Kernel)
- All preprocessing combinations
- Export in multiple formats (JSON, CSV, TSV)
- Model export and transformation
- Diagnostic metrics calculation

### 2. CLI/GUI Parity Tests
Located in `internal/integration/parity_test.go`

Ensures CLI and GUI produce identical results:
- Same input → same output validation
- Numerical precision comparison
- All PCA methods and preprocessing options
- Diagnostic metrics consistency

### 3. Regression Tests
Located in `internal/integration/regression_test.go`

Validates that existing functionality remains intact:
- Backward compatibility
- Edge case handling
- Security measures
- Numerical stability
- Performance baselines

### 4. Platform-Specific Tests
Script: `scripts/ci/test-platforms.sh`

Tests platform-specific behaviors:
- File path handling (spaces, Unicode)
- Line ending formats (LF, CRLF, CR)
- Security features (Windows reserved names, macOS quarantine)
- Build artifacts validation
- GUI application structure

## Running Tests

### Quick Start

```bash
# Run all integration tests
make test-integration

# Run specific test suite
go test -v -run TestE2EBasicWorkflow ./internal/integration/...

# Run with coverage
go test -cover ./internal/integration/...

# Run platform-specific tests
./scripts/ci/test-platforms.sh
```

### Test Scripts

#### `test-integration.sh`
Comprehensive integration test runner:
```bash
./scripts/ci/test-integration.sh

# Skip certain test categories
SKIP_FRONTEND=1 ./scripts/ci/test-integration.sh
SKIP_MEMORY_TESTS=1 ./scripts/ci/test-integration.sh
SKIP_BENCHMARKS=1 ./scripts/ci/test-integration.sh
```

#### `test-platforms.sh`
Platform-specific testing:
```bash
# Automatically detects platform (macOS/Linux/Windows)
./scripts/ci/test-platforms.sh
```

### Test Modes

#### Short Mode
Skip slow/intensive tests:
```bash
go test -short ./internal/integration/...
```

#### Verbose Mode
Detailed output:
```bash
go test -v ./internal/integration/...
```

#### Race Detection
Check for race conditions:
```bash
go test -race ./internal/integration/...
```

## Writing Integration Tests

### Test Structure

```go
func TestE2EWorkflow(t *testing.T) {
    // Skip in short mode
    SkipIfShort(t)
    
    // Create test configuration
    tc := NewTestConfig(t)
    tc.BuildCLI(t)
    
    // Create test data
    datasets := tc.CreateSampleDatasets(t)
    
    // Run test cases
    for _, dataset := range datasets {
        t.Run(dataset.Name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### Helper Functions

The `helpers.go` file provides utilities:

```go
// Create test CSV files
csvPath := tc.CreateTestCSV(t, "test.csv", data)

// Run CLI commands
output, err := tc.RunCLI(t, "analyze", csvPath, "--method", "svd")

// Load JSON results
results := tc.LoadJSONResult(t, "output.json")

// Compare matrices/vectors
err := CompareMatrices(a, b, tolerance)
err := CompareVectors(a, b, tolerance)

// Generate test data
data := GenerateTestMatrix(rows, cols, seed)
```

### Assertions

```go
// Basic assertions
AssertNoError(t, err, "Operation should succeed")
AssertError(t, err, "Operation should fail")
AssertContains(t, output, "expected", "Output check")

// File checks
CheckFileExists(t, path)

// Platform checks
SkipOnPlatform(t, "windows")
RequirePlatform(t, "darwin")
```

## Test Data

### Sample Datasets

Test datasets are created programmatically or loaded from `testdata/`:
- Small (10×5): Quick tests
- Medium (50×10): Standard tests
- Large (1000×100): Performance tests
- Missing data: NIPALS testing
- Edge cases: Numerical stability

### Expected Outputs

Golden files for comparison are stored in `testdata/expected/`.

## CI/CD Integration

### GitHub Actions

The integration tests run in CI:

```yaml
- name: Run Integration Tests
  run: |
    ./scripts/ci/test-integration.sh
  env:
    GOPCA_TEST_MODE: 1
    GOPCA_TEST_TIMEOUT: 120
```

### Platform Matrix

Tests run on multiple platforms:
- Ubuntu (latest)
- macOS (latest)
- Windows (latest)

## Performance Baselines

### Benchmark Targets

Based on Phase 5 optimization:

| Dataset Size | Max Time | Max Memory |
|-------------|----------|------------|
| 100×50 | 5s | 10MB |
| 500×100 | 10s | 50MB |
| 1000×200 | 30s | 200MB |
| 5000×500 | 2min | 1GB |

### Running Benchmarks

```bash
# Run all benchmarks
go test -bench=. ./internal/integration/...

# Run specific benchmark
go test -bench=BenchmarkLargeDataset ./internal/integration/...

# With memory profiling
go test -bench=. -benchmem ./internal/integration/...
```

## Debugging Failed Tests

### Verbose Output

```bash
# Run single test with verbose output
go test -v -run TestE2EBasicWorkflow/SVD_with_standard_scaling ./internal/integration/...
```

### Test Artifacts

Failed tests leave artifacts in temp directories:
```bash
# Check test output
ls -la /tmp/gopca_test_*
```

### Debug Mode

Enable debug output:
```bash
GOPCA_DEBUG=1 go test -v ./internal/integration/...
```

## Coverage Reports

### Generate Coverage

```bash
# Generate coverage profile
go test -coverprofile=coverage.out ./internal/integration/...

# View coverage report
go tool cover -html=coverage.out
```

### Coverage Targets

- Integration tests: >80%
- Core functionality: >90%
- Overall project: >75%

## Troubleshooting

### Common Issues

1. **Tests timeout**
   - Increase timeout: `GOPCA_TEST_TIMEOUT=300`
   - Run in non-short mode

2. **Platform-specific failures**
   - Check platform detection
   - Verify build artifacts exist
   - Check file permissions

3. **Numerical differences**
   - Adjust tolerance in comparisons
   - Check preprocessing consistency
   - Verify platform-specific math libraries

4. **Missing dependencies**
   - Run `make deps-all`
   - Check Go version (1.24+)
   - Verify npm for frontend tests

### Debug Commands

```bash
# Check test environment
go env

# List available tests
go test -list . ./internal/integration/...

# Run with custom timeout
timeout 300 go test ./internal/integration/...

# Check for race conditions
go test -race -count=10 ./internal/integration/...
```

## Best Practices

1. **Isolation**: Each test should be independent
2. **Cleanup**: Use `t.TempDir()` for automatic cleanup
3. **Deterministic**: Use seeded random data
4. **Fast**: Use `SkipIfShort()` for slow tests
5. **Descriptive**: Clear test names and error messages
6. **Platform-aware**: Handle platform differences gracefully

## Adding New Tests

1. Choose appropriate test file:
   - `e2e_test.go`: Workflow tests
   - `parity_test.go`: CLI/GUI comparison
   - `regression_test.go`: Bug prevention

2. Follow naming convention:
   - `TestE2E*`: End-to-end tests
   - `TestParity*`: Comparison tests
   - `TestRegression*`: Regression tests

3. Use helper functions from `helpers.go`

4. Add to CI if needed:
   - Update `test-integration.sh`
   - Add to GitHub Actions workflow

## Maintenance

### Regular Tasks

- Update performance baselines quarterly
- Review and remove obsolete tests
- Add tests for new features
- Update golden files when output format changes

### Test Metrics

Track:
- Test execution time
- Coverage percentage
- Failure rate
- Platform-specific issues

## Related Documentation

- [Code Audit Report](../../docs_tmp/CODE_AUDIT_v1.0.md)
- [CI/CD Workflows](../../.github/workflows/README.md)
- [Performance Benchmarks](../devel/performance.md)
- [Security Testing](../devel/security.md)