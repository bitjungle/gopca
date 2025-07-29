# Windows CI Troubleshooting - Lessons Learned

## Problem Summary
The Windows CI tests were failing with exit code 1 even though all individual tests showed as passing. Multiple attempts to fix this issue revealed several important insights about Windows-specific CI behaviors.

## Failed Attempts and Analysis

### Attempt 1: Remove packages without test files
**What we tried**: Removed `./cmd/gopca-cli/...` from test commands since it has no test files.
**Result**: Failed - the issue persisted
**Why it failed**: This was only addressing a symptom, not the root cause

### Attempt 2: Change `-race` to `-cover` flag  
**What we tried**: Changed from `go test -v -race` to `go test -v -cover`
**Result**: Failed - the issue persisted
**Why it failed**: Removed an important testing flag without understanding why

### Attempt 3: Explicitly list packages with test files
**What we tried**: Listed specific packages instead of using wildcards: `./internal/cli ./internal/core ./internal/io ./internal/utils ./pkg/types`
**Result**: Failed - only `pkg/types` appeared in output
**Why it failed**: The command syntax for multiple packages may not work correctly on Windows, or packages were being skipped

### Attempt 4: Revert to wildcards with `-race`
**What we tried**: Used `./internal/... ./pkg/...` with `-race` flag
**Result**: Failed - hit GitHub Actions usage limit
**Why it failed**: Unknown - ran out of attempts before seeing results

## Key Observations from Output

1. **Truncated output**: The test output only showed results from `pkg/types` (the last package), suggesting earlier packages weren't being tested or their output was lost.

2. **Contradictory status**: Output showed:
   ```
   PASS
   coverage: 86.3% of statements
   ok      github.com/bitjungle/gopca/pkg/types    0.029s    coverage: 86.3% of statements
   FAIL
   Error: Process completed with exit code 1.
   ```
   This suggests the test framework itself passed but something else caused the failure.

3. **Missing packages**: When explicitly listing packages, only the last one appeared in output, suggesting:
   - Command parsing issues on Windows
   - Early termination
   - Output buffering problems

## Insights from PR 66 and PR 70

### PR 66: Platform-specific icons and CI fixes
- Windows CI needs to run Go commands directly, not through make/bash scripts
- Had issues with path separators (backslash vs forward slash)
- Eventually solved formatting check issues by ensuring LF line endings
- Key commit: `37002bd9` - "handle Windows in CI by avoiding bash scripts"

### PR 70: CI robustness improvements
- Added `-race` flag for race condition detection
- Set `fail-fast: false` to continue other matrix jobs
- Added 10-minute timeout to prevent hanging tests
- Used wildcard patterns (`./internal/... ./pkg/...`) successfully

## Potential Root Causes

1. **Line ending issues**: Windows CRLF vs Unix LF (solved in PR 66 with git config)
2. **Path separator issues**: Windows backslash vs Unix forward slash
3. **Command parsing**: Multi-package syntax might not work the same on Windows
4. **Shell differences**: Even with `shell: bash`, behavior might differ
5. **Exit code handling**: Windows might handle Go test exit codes differently
6. **Package discovery**: Windows might have issues with `./...` pattern in certain contexts

## Recommended Next Steps

1. **Isolate the issue**:
   ```yaml
   - name: Debug Windows test issue
     if: runner.os == 'Windows'
     run: |
       echo "=== Testing each package individually ==="
       go test -v -race ./internal/cli || echo "cli failed with $?"
       go test -v -race ./internal/core || echo "core failed with $?"
       go test -v -race ./internal/io || echo "io failed with $?"
       go test -v -race ./internal/utils || echo "utils failed with $?"
       go test -v -race ./pkg/types || echo "types failed with $?"
       echo "=== Testing with wildcards ==="
       go test -v -race ./internal/... ./pkg/... || echo "Wildcard test failed with $?"
   ```

2. **Check for Windows-specific test failures**:
   - Some tests might behave differently on Windows (file paths, permissions, etc.)
   - Race conditions might manifest differently on Windows

3. **Consider using the exact working configuration from main branch**:
   - Check what the current working Windows CI configuration uses
   - Don't change multiple variables at once

4. **Alternative approaches**:
   - Use `go test -json` for structured output
   - Run tests in separate steps to isolate failures
   - Use PowerShell instead of bash for Windows

## Common Windows CI Pitfalls

1. **Path handling**: Always use forward slashes or `filepath.Join()`
2. **Line endings**: Configure git to use LF: `git config core.autocrlf false`
3. **Shell selection**: Explicitly specify `shell: bash` or `shell: pwsh`
4. **Environment variables**: Syntax differs between shells
5. **File permissions**: Windows has different permission model
6. **Temporary files**: Different temp directory structure

## References

- PR #66: Fixed Windows CI issues related to formatting and paths
- PR #70: Implemented robust CI with race detection
- PR #106: Current attempt to fix Windows test failures

## Next Investigation Areas

1. Check if packages without tests cause issues even with wildcards
2. Verify race detector works correctly on Windows
3. Test if output buffering causes truncation
4. Check if timeout is being hit (though 10m should be plenty)
5. Investigate if certain tests hang on Windows

## Temporary Workaround

Until resolved, consider:
- Allowing Windows tests to fail without blocking PR (`continue-on-error: true`)
- Running Windows tests without `-race` flag
- Using a different test approach for Windows only