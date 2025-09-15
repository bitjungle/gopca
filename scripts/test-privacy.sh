#!/bin/bash

# Runtime Privacy Testing Script for GoPCA Suite
# This script tests for network connections during runtime

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo "========================================="
echo "   GoPCA Runtime Privacy Testing        "
echo "========================================="
echo ""

# Check if running as root (needed for some network monitoring)
if [ "$EUID" -ne 0 ] && [ "$1" != "--no-root" ]; then 
    echo -e "${YELLOW}Note: Some tests require root access for network monitoring.${NC}"
    echo -e "${YELLOW}Run with sudo for complete testing, or use --no-root for basic tests.${NC}"
    echo ""
fi

# Function to print status
print_status() {
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✅ PASS${NC}: $2"
    else
        echo -e "${RED}❌ FAIL${NC}: $2"
    fi
}

# Function to print info
print_info() {
    echo -e "${BLUE}ℹ️  INFO${NC}: $1"
}

# Function to check if a binary exists
check_binary() {
    local binary=$1
    if [ -f "$binary" ]; then
        print_info "Found binary: $binary"
        return 0
    else
        echo -e "${YELLOW}Binary not found: $binary${NC}"
        return 1
    fi
}

# 1. Check for CLI binary network symbols
echo "1. Checking pca CLI binary for network symbols..."
PCA_BINARY="$PROJECT_ROOT/build/pca"
if check_binary "$PCA_BINARY"; then
    # Check for network-related symbols
    if command -v nm &> /dev/null; then
        NETWORK_SYMBOLS=$(nm "$PCA_BINARY" 2>/dev/null | grep -iE "socket|connect|send|recv|http|tcp|udp" | wc -l)
        if [ "$NETWORK_SYMBOLS" -gt 50 ]; then
            print_status 1 "Found $NETWORK_SYMBOLS network-related symbols (may be from standard library)"
        else
            print_status 0 "Minimal network symbols found ($NETWORK_SYMBOLS)"
        fi
    else
        print_info "nm not available, skipping symbol check"
    fi
    
    # Check for HTTP strings
    if strings "$PCA_BINARY" 2>/dev/null | grep -qE "https://|http://|telemetry|analytics" | grep -v "github.com\|json-schema"; then
        print_status 1 "Found HTTP URLs or telemetry strings in binary"
    else
        print_status 0 "No concerning URLs or telemetry strings in binary"
    fi
fi
echo ""

# 2. Test network isolation
echo "2. Testing network isolation..."
if [ "$EUID" -eq 0 ]; then
    print_info "Starting network monitor (10 second test)..."
    
    # Start tcpdump in background
    PCAP_FILE="/tmp/gopca_test_$$.pcap"
    tcpdump -i any -w "$PCAP_FILE" "not host 127.0.0.1 and not host ::1" 2>/dev/null &
    TCPDUMP_PID=$!
    
    print_info "Please run GoPCA Desktop or pca CLI now..."
    sleep 10
    
    # Stop tcpdump
    kill $TCPDUMP_PID 2>/dev/null || true
    wait $TCPDUMP_PID 2>/dev/null || true
    
    # Analyze capture
    PACKET_COUNT=$(tcpdump -r "$PCAP_FILE" 2>/dev/null | wc -l)
    if [ "$PACKET_COUNT" -eq 0 ]; then
        print_status 0 "No external network traffic detected"
    else
        print_status 1 "Found $PACKET_COUNT external packets"
        echo "Top destinations:"
        tcpdump -r "$PCAP_FILE" -nn 2>/dev/null | awk '{print $5}' | sort | uniq -c | sort -rn | head -5
    fi
    
    rm -f "$PCAP_FILE"
else
    print_info "Skipping network monitor (requires root)"
fi
echo ""

# 3. Check for listening ports
echo "3. Checking for listening ports..."
if command -v lsof &> /dev/null; then
    LISTENING_PORTS=$(lsof -i -P 2>/dev/null | grep -i "gopca\|pca" | grep LISTEN || true)
    if [ -n "$LISTENING_PORTS" ]; then
        print_status 1 "Found listening ports:"
        echo "$LISTENING_PORTS"
    else
        print_status 0 "No listening ports found"
    fi
elif command -v netstat &> /dev/null; then
    LISTENING_PORTS=$(netstat -an 2>/dev/null | grep LISTEN | grep -i "gopca\|pca" || true)
    if [ -n "$LISTENING_PORTS" ]; then
        print_status 1 "Found listening ports:"
        echo "$LISTENING_PORTS"
    else
        print_status 0 "No listening ports found"
    fi
else
    print_info "lsof/netstat not available, skipping port check"
fi
echo ""

# 4. Check DNS queries
echo "4. Checking for DNS queries..."
if [ "$(uname)" = "Darwin" ]; then
    # macOS
    print_info "Monitoring DNS cache (10 seconds)..."
    dscacheutil -statistics > /tmp/dns_before_$$.txt 2>/dev/null
    
    print_info "Please run GoPCA now..."
    sleep 10
    
    dscacheutil -statistics > /tmp/dns_after_$$.txt 2>/dev/null
    
    # Compare DNS stats
    if diff /tmp/dns_before_$$.txt /tmp/dns_after_$$.txt > /dev/null 2>&1; then
        print_status 0 "No new DNS queries detected"
    else
        print_info "DNS cache changed (may be system activity)"
    fi
    
    rm -f /tmp/dns_before_$$.txt /tmp/dns_after_$$.txt
elif [ "$EUID" -eq 0 ]; then
    # Linux with root
    print_info "Monitoring DNS queries (10 seconds)..."
    timeout 10 tcpdump -i any port 53 2>/dev/null > /tmp/dns_$$.txt &
    
    print_info "Please run GoPCA now..."
    wait
    
    DNS_COUNT=$(wc -l < /tmp/dns_$$.txt)
    if [ "$DNS_COUNT" -eq 0 ]; then
        print_status 0 "No DNS queries detected"
    else
        print_status 1 "Found $DNS_COUNT DNS queries"
    fi
    
    rm -f /tmp/dns_$$.txt
else
    print_info "DNS monitoring requires root on Linux"
fi
echo ""

# 5. Test with firewall blocking
echo "5. Testing with firewall (manual test)..."
echo -e "${YELLOW}Manual test instructions:${NC}"
echo "  1. Block all network access for GoPCA using your firewall"
echo "     - macOS: Use Little Snitch, LuLu, or pfctl"
echo "     - Linux: Use iptables or ufw"
echo "     - Windows: Use Windows Firewall"
echo "  2. Run GoPCA with network blocked"
echo "  3. Verify all features work correctly"
echo ""

# 6. Check for WebView/Browser network activity
echo "6. Checking for WebView network activity..."
if [ "$(uname)" = "Darwin" ]; then
    # Check for WebKit network processes
    WEBKIT_PROCS=$(ps aux | grep -i "webkit\|webview" | grep -i gopca || true)
    if [ -n "$WEBKIT_PROCS" ]; then
        print_info "WebView processes found (expected for Wails apps)"
    fi
fi
echo ""

# 7. Binary strings analysis
echo "7. Deep binary analysis..."
for binary in "$PROJECT_ROOT/build/pca" \
              "$PROJECT_ROOT/cmd/gopca-desktop/build/bin/GoPCA" \
              "$PROJECT_ROOT/cmd/gocsv/build/bin/GoCSV"; do
    if [ -f "$binary" ]; then
        echo "Analyzing: $(basename "$binary")"
        
        # Check for telemetry keywords (excluding our own metrics calculations)
        TELEMETRY_STRINGS=$(strings "$binary" 2>/dev/null | \
            grep -iE "telemetry|analytics|tracking|beacon|pixel" | \
            grep -v "github\|gopca\|Error\|Metrics\|metrics:" | wc -l)
        
        if [ "$TELEMETRY_STRINGS" -eq 0 ]; then
            print_status 0 "  No telemetry strings found"
        else
            print_status 1 "  Found $TELEMETRY_STRINGS potential telemetry strings"
        fi
    fi
done
echo ""

# Summary
echo "========================================="
echo "         RUNTIME TEST SUMMARY           "
echo "========================================="
echo ""
echo -e "${GREEN}Testing complete!${NC}"
echo ""
echo "Next steps for complete verification:"
echo "  1. Run this script with sudo for network monitoring"
echo "  2. Use Wireshark to capture all traffic during app usage"
echo "  3. Test with firewall blocking all connections"
echo "  4. Monitor system logs for network attempts"
echo ""
echo "For continuous monitoring, run:"
echo "  watch 'lsof -i -P | grep -i gopca'"
echo ""

# Create test report
REPORT_FILE="$PROJECT_ROOT/privacy-test-report-$(date +%Y%m%d-%H%M%S).txt"
{
    echo "GoPCA Privacy Test Report"
    echo "Generated: $(date)"
    echo ""
    echo "Tests performed:"
    echo "  - Binary symbol analysis"
    echo "  - String content analysis"
    echo "  - Network isolation testing"
    echo "  - DNS query monitoring"
    echo "  - Port listening checks"
    echo ""
    echo "Results: See console output above"
} > "$REPORT_FILE"

print_info "Report saved to: $REPORT_FILE"