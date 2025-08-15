#!/bin/bash
# Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
# Measures build times for all targets in the GoPCA monorepo

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to measure build time
measure_build() {
    local target=$1
    local description=$2
    
    echo -e "${YELLOW}Building: ${description}${NC}"
    
    # Clean before build
    make clean >/dev/null 2>&1 || true
    
    # Measure time
    start_time=$(date +%s.%N)
    make $target >/dev/null 2>&1
    end_time=$(date +%s.%N)
    
    # Calculate duration
    duration=$(echo "$end_time - $start_time" | bc)
    
    echo -e "${GREEN}✓ ${description}: ${duration}s${NC}"
    echo "$target,$description,$duration" >> build_times.csv
}

# Function to measure parallel build
measure_parallel_build() {
    local target=$1
    local description=$2
    local jobs=$3
    
    echo -e "${YELLOW}Building (parallel -j$jobs): ${description}${NC}"
    
    # Clean before build
    make clean >/dev/null 2>&1 || true
    
    # Measure time
    start_time=$(date +%s.%N)
    make -j$jobs $target >/dev/null 2>&1
    end_time=$(date +%s.%N)
    
    # Calculate duration
    duration=$(echo "$end_time - $start_time" | bc)
    
    echo -e "${GREEN}✓ ${description} (parallel): ${duration}s${NC}"
    echo "$target-parallel,$description (parallel),$duration" >> build_times.csv
}

# Main execution
echo "========================================="
echo "GoPCA Build Time Measurement"
echo "========================================="
echo ""

# Create CSV header
echo "Target,Description,Time(s)" > build_times.csv

# Measure individual targets
echo -e "${YELLOW}Measuring individual build targets...${NC}"
echo ""

measure_build "build" "CLI (current platform)"
measure_build "pca-build" "GoPCA Desktop GUI"
measure_build "csv-build" "GoCSV GUI"
measure_build "build-all" "CLI (all platforms)"

echo ""
echo -e "${YELLOW}Measuring parallel builds...${NC}"
echo ""

# Measure parallel builds
cores=$(nproc 2>/dev/null || sysctl -n hw.ncpu 2>/dev/null || echo 4)
measure_parallel_build "build-all" "CLI (all platforms)" $cores

echo ""
echo "========================================="
echo "Build Time Summary"
echo "========================================="
echo ""

# Display summary
total_sequential=0
total_parallel=0

while IFS=, read -r target description time; do
    if [[ $target != "Target" ]]; then
        if [[ $target == *"-parallel" ]]; then
            total_parallel=$(echo "$total_parallel + $time" | bc)
        else
            total_sequential=$(echo "$total_sequential + $time" | bc)
        fi
        printf "%-30s %10.2fs\n" "$description:" "$time"
    fi
done < build_times.csv

echo ""
echo "----------------------------------------"
printf "%-30s %10.2fs\n" "Total Sequential Time:" "$total_sequential"
printf "%-30s %10.2fs\n" "Total Parallel Time:" "$total_parallel"

if (( $(echo "$total_sequential > 0" | bc -l) )); then
    speedup=$(echo "scale=2; $total_sequential / $total_parallel" | bc)
    improvement=$(echo "scale=1; (($total_sequential - $total_parallel) / $total_sequential) * 100" | bc)
    echo ""
    echo -e "${GREEN}Parallel build speedup: ${speedup}x${NC}"
    echo -e "${GREEN}Time saved: ${improvement}%${NC}"
fi

echo ""
echo "Results saved to: build_times.csv"