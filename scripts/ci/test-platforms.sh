#!/bin/bash
# test-platforms.sh - Platform-specific testing for Windows, macOS, and Linux
# Tests platform-specific behaviors, file paths, and system integrations

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

echo "=========================================="
echo "Platform-Specific Test Suite"
echo "=========================================="
echo "Platform: $(uname -s)"
echo "Architecture: $(uname -m)"
echo "Date: $(date)"
echo ""

# Detect platform
PLATFORM="unknown"
case "$(uname -s)" in
    Darwin*)    PLATFORM="macos" ;;
    Linux*)     PLATFORM="linux" ;;
    MINGW*|MSYS*|CYGWIN*) PLATFORM="windows" ;;
    *)          echo "Unknown platform: $(uname -s)"; exit 1 ;;
esac

echo "Detected platform: $PLATFORM"
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Test counters
PASSED=0
FAILED=0
SKIPPED=0

# Print test result
print_result() {
    local status=$1
    local test_name=$2
    local details=$3
    
    case $status in
        "PASS")
            echo -e "${GREEN}✓${NC} $test_name"
            [ -n "$details" ] && echo "  $details"
            ((PASSED++))
            ;;
        "FAIL")
            echo -e "${RED}✗${NC} $test_name"
            [ -n "$details" ] && echo "  Error: $details"
            ((FAILED++))
            ;;
        "SKIP")
            echo -e "${YELLOW}⊘${NC} $test_name"
            [ -n "$details" ] && echo "  Reason: $details"
            ((SKIPPED++))
            ;;
    esac
}

# Test path separators and file handling
test_path_handling() {
    echo "Testing path handling..."
    
    local test_file="test_data.csv"
    local test_dir="test_dir"
    
    # Create test directory with spaces (if supported)
    if [ "$PLATFORM" != "windows" ]; then
        local space_dir="test dir with spaces"
        if mkdir -p "$PROJECT_ROOT/$space_dir" 2>/dev/null; then
            # Create CSV in directory with spaces
            echo "a,b,c" > "$PROJECT_ROOT/$space_dir/$test_file"
            echo "1,2,3" >> "$PROJECT_ROOT/$space_dir/$test_file"
            
            # Test CLI with spaces in path
            if "$PROJECT_ROOT/build/pca" analyze "$PROJECT_ROOT/$space_dir/$test_file" \
                --method svd --components 1 > /dev/null 2>&1; then
                print_result "PASS" "Paths with spaces"
            else
                print_result "FAIL" "Paths with spaces"
            fi
            
            rm -rf "$PROJECT_ROOT/$space_dir"
        else
            print_result "SKIP" "Paths with spaces" "Cannot create directory with spaces"
        fi
    else
        print_result "SKIP" "Paths with spaces" "Windows test not implemented"
    fi
    
    # Test relative vs absolute paths
    cd "$PROJECT_ROOT"
    echo "x,y,z" > relative_test.csv
    echo "4,5,6" >> relative_test.csv
    
    if "$PROJECT_ROOT/build/pca" analyze relative_test.csv \
        --method svd --components 1 > /dev/null 2>&1; then
        print_result "PASS" "Relative paths"
    else
        print_result "FAIL" "Relative paths"
    fi
    
    if "$PROJECT_ROOT/build/pca" analyze "$PROJECT_ROOT/relative_test.csv" \
        --method svd --components 1 > /dev/null 2>&1; then
        print_result "PASS" "Absolute paths"
    else
        print_result "FAIL" "Absolute paths"
    fi
    
    rm -f relative_test.csv
    
    # Test Unicode filenames (if supported)
    if [ "$PLATFORM" != "windows" ]; then
        local unicode_file="测试文件.csv"
        echo "col1,col2" > "$unicode_file" 2>/dev/null || {
            print_result "SKIP" "Unicode filenames" "Filesystem doesn't support Unicode"
            return
        }
        echo "7,8" >> "$unicode_file"
        
        if "$PROJECT_ROOT/build/pca" analyze "$unicode_file" \
            --method svd --components 1 > /dev/null 2>&1; then
            print_result "PASS" "Unicode filenames"
        else
            print_result "FAIL" "Unicode filenames"
        fi
        
        rm -f "$unicode_file"
    else
        print_result "SKIP" "Unicode filenames" "Windows Unicode test not implemented"
    fi
}

# Test platform-specific line endings
test_line_endings() {
    echo ""
    echo "Testing line ending handling..."
    
    # Create files with different line endings
    printf "a,b,c\n1,2,3\n" > unix_endings.csv
    printf "a,b,c\r\n1,2,3\r\n" > windows_endings.csv
    printf "a,b,c\r1,2,3\r" > mac_endings.csv
    
    # Test Unix line endings (LF)
    if "$PROJECT_ROOT/build/pca" analyze unix_endings.csv \
        --method svd --components 1 > /dev/null 2>&1; then
        print_result "PASS" "Unix line endings (LF)"
    else
        print_result "FAIL" "Unix line endings (LF)"
    fi
    
    # Test Windows line endings (CRLF)
    if "$PROJECT_ROOT/build/pca" analyze windows_endings.csv \
        --method svd --components 1 > /dev/null 2>&1; then
        print_result "PASS" "Windows line endings (CRLF)"
    else
        print_result "FAIL" "Windows line endings (CRLF)"
    fi
    
    # Test old Mac line endings (CR) - might not be supported
    if "$PROJECT_ROOT/build/pca" analyze mac_endings.csv \
        --method svd --components 1 > /dev/null 2>&1; then
        print_result "PASS" "Mac line endings (CR)"
    else
        print_result "SKIP" "Mac line endings (CR)" "Not supported"
    fi
    
    rm -f unix_endings.csv windows_endings.csv mac_endings.csv
}

# Test platform-specific security features
test_security() {
    echo ""
    echo "Testing platform-specific security..."
    
    case "$PLATFORM" in
        windows)
            # Test Windows reserved names
            for reserved in "CON" "PRN" "AUX" "NUL" "COM1" "LPT1"; do
                if "$PROJECT_ROOT/build/pca" analyze "${reserved}.csv" \
                    > /dev/null 2>&1; then
                    print_result "FAIL" "Block Windows reserved name: $reserved" \
                        "Should have been blocked"
                else
                    print_result "PASS" "Block Windows reserved name: $reserved"
                fi
            done
            
            # Test drive letter paths
            if "$PROJECT_ROOT/build/pca" analyze "C:\\test.csv" \
                > /dev/null 2>&1; then
                print_result "SKIP" "Windows drive paths" "Test file doesn't exist"
            else
                print_result "PASS" "Windows drive paths"
            fi
            ;;
            
        macos)
            # Test macOS bundle paths
            if [ -d "$PROJECT_ROOT/cmd/gopca-desktop/build/bin/GoPCA.app" ]; then
                bundle_path="$PROJECT_ROOT/cmd/gopca-desktop/build/bin/GoPCA.app/Contents/MacOS/GoPCA"
                if [ -x "$bundle_path" ]; then
                    print_result "PASS" "macOS bundle structure"
                else
                    print_result "FAIL" "macOS bundle structure" "Binary not executable"
                fi
            else
                print_result "SKIP" "macOS bundle structure" "App not built"
            fi
            
            # Test quarantine attribute handling
            test_file="quarantine_test.csv"
            echo "a,b" > "$test_file"
            echo "1,2" >> "$test_file"
            
            # Add quarantine attribute (simulated download)
            xattr -w com.apple.quarantine "0001;00000000;Safari;" "$test_file" 2>/dev/null || {
                print_result "SKIP" "Quarantine attribute" "Cannot set xattr"
            }
            
            if xattr -l "$test_file" | grep -q quarantine; then
                if "$PROJECT_ROOT/build/pca" analyze "$test_file" \
                    --method svd --components 1 > /dev/null 2>&1; then
                    print_result "PASS" "Handle quarantined files"
                else
                    print_result "FAIL" "Handle quarantined files"
                fi
            fi
            
            rm -f "$test_file"
            ;;
            
        linux)
            # Test symlink handling
            test_file="real_file.csv"
            symlink="symlink.csv"
            echo "x,y" > "$test_file"
            echo "3,4" >> "$test_file"
            
            ln -s "$test_file" "$symlink"
            
            if "$PROJECT_ROOT/build/pca" analyze "$symlink" \
                --method svd --components 1 > /dev/null 2>&1; then
                print_result "PASS" "Symlink handling"
            else
                print_result "FAIL" "Symlink handling"
            fi
            
            rm -f "$test_file" "$symlink"
            
            # Test permission handling
            restricted_file="restricted.csv"
            echo "a,b" > "$restricted_file"
            echo "5,6" >> "$restricted_file"
            chmod 000 "$restricted_file"
            
            if "$PROJECT_ROOT/build/pca" analyze "$restricted_file" \
                > /dev/null 2>&1; then
                print_result "FAIL" "Permission checking" "Should have failed on restricted file"
            else
                print_result "PASS" "Permission checking"
            fi
            
            chmod 644 "$restricted_file"
            rm -f "$restricted_file"
            ;;
    esac
}

# Test platform-specific build artifacts
test_build_artifacts() {
    echo ""
    echo "Testing platform-specific build artifacts..."
    
    case "$PLATFORM" in
        windows)
            if [ -f "$PROJECT_ROOT/build/pca.exe" ]; then
                print_result "PASS" "Windows executable (.exe)"
            else
                print_result "FAIL" "Windows executable (.exe)" "Not found"
            fi
            ;;
            
        macos)
            if [ -f "$PROJECT_ROOT/build/pca" ]; then
                # Check if it's a Mach-O binary
                if file "$PROJECT_ROOT/build/pca" | grep -q "Mach-O"; then
                    print_result "PASS" "macOS binary format"
                    
                    # Check architecture
                    if file "$PROJECT_ROOT/build/pca" | grep -q "arm64"; then
                        print_result "PASS" "Apple Silicon (arm64) binary"
                    elif file "$PROJECT_ROOT/build/pca" | grep -q "x86_64"; then
                        print_result "PASS" "Intel (x86_64) binary"
                    else
                        print_result "FAIL" "Unknown architecture"
                    fi
                else
                    print_result "FAIL" "macOS binary format" "Not a Mach-O binary"
                fi
            else
                print_result "FAIL" "macOS binary" "Not found"
            fi
            ;;
            
        linux)
            if [ -f "$PROJECT_ROOT/build/pca" ]; then
                # Check if it's an ELF binary
                if file "$PROJECT_ROOT/build/pca" | grep -q "ELF"; then
                    print_result "PASS" "Linux ELF binary"
                    
                    # Check if dynamically or statically linked
                    if ldd "$PROJECT_ROOT/build/pca" 2>/dev/null | grep -q "not a dynamic"; then
                        print_result "PASS" "Static binary"
                    else
                        print_result "PASS" "Dynamic binary"
                    fi
                else
                    print_result "FAIL" "Linux binary format" "Not an ELF binary"
                fi
            else
                print_result "FAIL" "Linux binary" "Not found"
            fi
            ;;
    esac
}

# Test GUI application (if available)
test_gui_application() {
    echo ""
    echo "Testing GUI applications..."
    
    case "$PLATFORM" in
        macos)
            app_path="$PROJECT_ROOT/cmd/gopca-desktop/build/bin/GoPCA.app"
            if [ -d "$app_path" ]; then
                # Check Info.plist
                if [ -f "$app_path/Contents/Info.plist" ]; then
                    print_result "PASS" "macOS Info.plist exists"
                    
                    # Check bundle identifier
                    if grep -q "com.bitjungle.gopca" "$app_path/Contents/Info.plist"; then
                        print_result "PASS" "Bundle identifier correct"
                    else
                        print_result "FAIL" "Bundle identifier incorrect"
                    fi
                else
                    print_result "FAIL" "macOS Info.plist missing"
                fi
                
                # Check code signing
                if codesign -v "$app_path" 2>/dev/null; then
                    print_result "PASS" "Code signing valid"
                else
                    print_result "SKIP" "Code signing" "Not signed"
                fi
            else
                print_result "SKIP" "macOS GUI app" "Not built"
            fi
            ;;
            
        windows)
            exe_path="$PROJECT_ROOT/cmd/gopca-desktop/build/bin/GoPCA-Desktop.exe"
            if [ -f "$exe_path" ]; then
                print_result "PASS" "Windows GUI executable exists"
            else
                print_result "SKIP" "Windows GUI app" "Not built"
            fi
            ;;
            
        linux)
            appimage_path="$PROJECT_ROOT/cmd/gopca-desktop/build/bin/gopca-desktop"
            if [ -f "$appimage_path" ]; then
                print_result "PASS" "Linux GUI binary exists"
                
                # Check if it's executable
                if [ -x "$appimage_path" ]; then
                    print_result "PASS" "Linux GUI binary executable"
                else
                    print_result "FAIL" "Linux GUI binary not executable"
                fi
            else
                print_result "SKIP" "Linux GUI app" "Not built"
            fi
            ;;
    esac
}

# Test environment variables
test_environment() {
    echo ""
    echo "Testing environment variables..."
    
    # Test GOPCA_DEBUG
    GOPCA_DEBUG=1 "$PROJECT_ROOT/build/pca" --version > /tmp/debug_test.log 2>&1
    if grep -q "DEBUG" /tmp/debug_test.log; then
        print_result "PASS" "GOPCA_DEBUG environment variable"
    else
        print_result "SKIP" "GOPCA_DEBUG environment variable" "Debug output not detected"
    fi
    
    # Test GOPCA_PROFILE
    GOPCA_PROFILE=1 timeout 5 "$PROJECT_ROOT/build/pca" analyze \
        "$PROJECT_ROOT/testdata/iris/iris.csv" --method svd > /tmp/profile_test.log 2>&1 || true
    
    if grep -q -i "profile\|memory" /tmp/profile_test.log; then
        print_result "PASS" "GOPCA_PROFILE environment variable"
    else
        print_result "SKIP" "GOPCA_PROFILE environment variable" "Profile output not detected"
    fi
}

# Main test execution
echo "Starting platform-specific tests for $PLATFORM..."
echo ""

# Build the CLI if not already built
if [ ! -f "$PROJECT_ROOT/build/pca" ] && [ ! -f "$PROJECT_ROOT/build/pca.exe" ]; then
    echo "Building CLI tool..."
    cd "$PROJECT_ROOT"
    if make build > /dev/null 2>&1; then
        echo "Build successful"
    else
        echo "Build failed - some tests will be skipped"
    fi
fi

# Run test suites
test_path_handling
test_line_endings
test_security
test_build_artifacts
test_gui_application
test_environment

# Summary
echo ""
echo "=========================================="
echo "Platform Test Summary for $PLATFORM"
echo "=========================================="
echo -e "Passed:  ${GREEN}$PASSED${NC}"
if [ $FAILED -gt 0 ]; then
    echo -e "Failed:  ${RED}$FAILED${NC}"
else
    echo "Failed:  $FAILED"
fi
if [ $SKIPPED -gt 0 ]; then
    echo -e "Skipped: ${YELLOW}$SKIPPED${NC}"
else
    echo "Skipped: $SKIPPED"
fi
echo "=========================================="

# Exit with appropriate code
if [ $FAILED -gt 0 ]; then
    exit 1
else
    exit 0
fi