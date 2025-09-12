#!/bin/bash

# Privacy Verification Script for GoPCA Suite
# This script verifies that no telemetry or data collection exists in the codebase

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "========================================="
echo "    GoPCA Privacy Verification Audit    "
echo "========================================="
echo ""

# Track if any issues are found
ISSUES_FOUND=0

# Function to print status
print_status() {
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✅ PASS${NC}: $2"
    else
        echo -e "${RED}❌ FAIL${NC}: $2"
        ISSUES_FOUND=1
    fi
}

# Function to print warning
print_warning() {
    echo -e "${YELLOW}⚠️  WARNING${NC}: $1"
}

echo "📋 Starting privacy audit..."
echo ""

# 1. Check for telemetry-related Go dependencies
echo "1. Checking Go dependencies for telemetry..."
cd "$PROJECT_ROOT"
if go list -m all 2>/dev/null | grep -iE "telemetry|analytics|tracking|metrics|sentry|rollbar|datadog|newrelic|amplitude|mixpanel|segment" > /dev/null 2>&1; then
    print_status 1 "Found potential telemetry-related Go dependencies:"
    go list -m all | grep -iE "telemetry|analytics|tracking|metrics|sentry|rollbar|datadog|newrelic|amplitude|mixpanel|segment"
else
    print_status 0 "No telemetry-related Go dependencies found"
fi
echo ""

# 2. Check for telemetry-related npm dependencies
echo "2. Checking npm dependencies for telemetry..."
if [ -f "$PROJECT_ROOT/package-lock.json" ]; then
    if npm ls 2>/dev/null | grep -iE "analytics|telemetry|tracking|metrics|gtag|google-analytics|sentry|rollbar|datadog|newrelic|amplitude|mixpanel|segment" > /dev/null 2>&1; then
        print_status 1 "Found potential telemetry-related npm dependencies:"
        npm ls 2>/dev/null | grep -iE "analytics|telemetry|tracking|metrics|gtag|google-analytics|sentry|rollbar|datadog|newrelic|amplitude|mixpanel|segment"
    else
        print_status 0 "No telemetry-related npm dependencies found"
    fi
else
    print_warning "No package-lock.json found, skipping npm check"
fi
echo ""

# 3. Check for external URLs in source code
echo "3. Checking for external URLs in source code..."
EXTERNAL_URLS=$(grep -r "https://" \
    --include="*.go" \
    --include="*.ts" \
    --include="*.tsx" \
    --include="*.js" \
    --include="*.jsx" \
    --exclude-dir="dist" \
    --exclude-dir="build" \
    --exclude-dir="vendor" \
    --exclude-dir="node_modules" \
    --exclude-dir=".venv" \
    --exclude-dir="testdata" \
    "$PROJECT_ROOT" 2>/dev/null | \
    grep -vE "github\.com|json-schema\.org|doi\.org|arxiv\.org|npmjs\.com|timestamp\.digicert\.com|wails\.io/docs|reactjs\.org/docs|localhost|127\.0\.0\.1|example\.com|cdn\.jsdelivr\.net/npm/ag-grid|bitjungle\.github\.io/gopca|heroicons\.com|vitejs\.dev/config" | \
    grep -vE "^\s*//" || true)

if [ -n "$EXTERNAL_URLS" ]; then
    print_status 1 "Found external URLs in source code (review needed):"
    echo "$EXTERNAL_URLS" | head -10
    echo "(Showing first 10 results)"
else
    print_status 0 "No concerning external URLs found in source code"
fi
echo ""

# 4. Check for HTTP client usage in Go code
echo "4. Checking for HTTP client usage in Go code..."
HTTP_USAGE=$(grep -r "http\.\(Get\|Post\|Client\)" \
    --include="*.go" \
    --exclude-dir="vendor" \
    "$PROJECT_ROOT" 2>/dev/null || true)

if [ -n "$HTTP_USAGE" ]; then
    print_status 1 "Found HTTP client usage in Go code:"
    echo "$HTTP_USAGE"
else
    print_status 0 "No HTTP client usage found in Go code"
fi
echo ""

# 5. Check for fetch/XMLHttpRequest in JavaScript/TypeScript
echo "5. Checking for network calls in JavaScript/TypeScript..."
FETCH_USAGE=$(grep -r "fetch\|XMLHttpRequest\|axios\|WebSocket" \
    --include="*.ts" \
    --include="*.tsx" \
    --include="*.js" \
    --include="*.jsx" \
    --exclude-dir="dist" \
    --exclude-dir="build" \
    --exclude-dir="node_modules" \
    --exclude-dir=".venv" \
    --exclude-dir="testdata" \
    "$PROJECT_ROOT" 2>/dev/null | \
    grep -v "// \|/\* \| \* \|fetchMetrics\|fetch(dataUrl)\|fetch(markdownPath)\|blobToDataURL\|fetchError\|fetchedBlob" || true)

if [ -n "$FETCH_USAGE" ]; then
    print_status 1 "Found potential network calls in frontend code:"
    echo "$FETCH_USAGE" | head -10
    echo "(Showing first 10 results)"
else
    print_status 0 "No network calls found in frontend source code"
fi
echo ""

# 6. Check for analytics/tracking scripts
echo "6. Checking for analytics/tracking scripts..."
ANALYTICS=$(grep -r "gtag\|ga(\|_gaq\|analytics\|pixel\|beacon" \
    --include="*.html" \
    --include="*.ts" \
    --include="*.tsx" \
    --include="*.js" \
    --include="*.jsx" \
    --exclude-dir="dist" \
    --exclude-dir="build" \
    --exclude-dir="node_modules" \
    --exclude-dir=".venv" \
    --exclude-dir="testdata" \
    "$PROJECT_ROOT" 2>/dev/null | \
    grep -v "// \|/\* \| \* \|pixelRatio\|No analytics" || true)

if [ -n "$ANALYTICS" ]; then
    print_status 1 "Found potential analytics/tracking code:"
    echo "$ANALYTICS"
else
    print_status 0 "No analytics/tracking scripts found"
fi
echo ""

# 7. Check for CDN usage
echo "7. Checking for CDN usage..."
CDN_USAGE=$(grep -r "cdn\.\|unpkg\.com\|jsdelivr\|cloudflare" \
    --include="*.html" \
    --include="*.ts" \
    --include="*.tsx" \
    --include="*.js" \
    --include="*.jsx" \
    --exclude-dir="dist" \
    --exclude-dir="build" \
    --exclude-dir="node_modules" \
    "$PROJECT_ROOT" 2>/dev/null | \
    grep -v "// \|/\* \| \* \|ag-grid" || true)

if [ -n "$CDN_USAGE" ]; then
    print_warning "Found CDN references (verify they're not used in production):"
    echo "$CDN_USAGE" | head -5
else
    print_status 0 "No CDN usage found"
fi
echo ""

# 8. Check for auto-update mechanisms
echo "8. Checking for auto-update mechanisms..."
UPDATE_CHECK=$(grep -r "auto.?update\|self.?update\|check.?for.?update\|version.?check" \
    --include="*.go" \
    --include="*.ts" \
    --include="*.tsx" \
    --exclude-dir="vendor" \
    --exclude-dir="node_modules" \
    "$PROJECT_ROOT" 2>/dev/null | \
    grep -v "// \|/\* \| \* \|README\|docs/" || true)

if [ -n "$UPDATE_CHECK" ]; then
    print_warning "Found potential update check code (verify it's disabled):"
    echo "$UPDATE_CHECK" | head -5
else
    print_status 0 "No auto-update mechanisms found"
fi
echo ""

# 9. Verify Go telemetry is disabled
echo "9. Checking Go telemetry configuration..."
if command -v go &> /dev/null; then
    GO_TELEMETRY=$(go env GOTELEMETRY 2>/dev/null || echo "not set")
    if [ "$GO_TELEMETRY" = "off" ] || [ "$GO_TELEMETRY" = "not set" ] || [ "$GO_TELEMETRY" = "" ] || [ "$GO_TELEMETRY" = "local" ]; then
        if [ "$GO_TELEMETRY" = "local" ]; then
            print_status 0 "Go telemetry is set to local only (not uploading)"
        else
            print_status 0 "Go telemetry is disabled or not configured"
        fi
    else
        print_status 1 "Go telemetry is set to: $GO_TELEMETRY (should be 'off' or 'local')"
    fi
else
    print_warning "Go not installed, skipping telemetry check"
fi
echo ""

# 10. Check for sensitive data logging
echo "10. Checking for sensitive data logging..."
SENSITIVE_LOG=$(grep -r "console\.log\|fmt\.Print\|log\." \
    --include="*.go" \
    --include="*.ts" \
    --include="*.tsx" \
    --exclude-dir="vendor" \
    --exclude-dir="node_modules" \
    --exclude-dir="test" \
    "$PROJECT_ROOT" 2>/dev/null | \
    grep -iE "password|token|key|secret|credential|auth" | \
    grep -v "// \|/\* \| \* \|_test\.go" || true)

if [ -n "$SENSITIVE_LOG" ]; then
    print_warning "Found potential sensitive data logging (review needed):"
    echo "$SENSITIVE_LOG" | head -5
else
    print_status 0 "No obvious sensitive data logging found"
fi
echo ""

# Summary
echo "========================================="
echo "           AUDIT SUMMARY                "
echo "========================================="
echo ""

if [ $ISSUES_FOUND -eq 0 ]; then
    echo -e "${GREEN}✅ Privacy audit PASSED${NC}"
    echo "No privacy or telemetry concerns found."
else
    echo -e "${RED}❌ Privacy audit FAILED${NC}"
    echo "Issues found that need review."
    echo "Please check the details above."
fi

echo ""
echo "Audit completed at: $(date)"
echo ""

# Exit with appropriate code
exit $ISSUES_FOUND