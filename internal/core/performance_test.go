// Copyright 2025-2026 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package core

import (
	"fmt"
	"math/rand"
	"runtime"
	"testing"
	"time"

	"github.com/bitjungle/gopca/pkg/types"
)

// generateRandomMatrix creates a random matrix for benchmarking
func generateRandomMatrix(rows, cols int) types.Matrix {
	data := make(types.Matrix, rows)
	for i := range data {
		data[i] = make([]float64, cols)
		for j := range data[i] {
			data[i][j] = rand.NormFloat64()
		}
	}
	return data
}

// BenchmarkPCASVD benchmarks SVD method with various matrix sizes
func BenchmarkPCASVD(b *testing.B) {
	sizes := []struct {
		rows int
		cols int
	}{
		{100, 10},
		{1000, 10},
		{1000, 100},
		{5000, 100},
		{10000, 100},
	}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("%dx%d", size.rows, size.cols), func(b *testing.B) {
			data := generateRandomMatrix(size.rows, size.cols)
			config := types.PCAConfig{
				Components:    min(10, size.cols),
				MeanCenter:    true,
				StandardScale: false,
				Method:        "svd",
			}

			engine := NewPCAEngine()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_, err := engine.Fit(data, config)
				if err != nil {
					b.Fatalf("PCA failed: %v", err)
				}
			}
		})
	}
}

// BenchmarkPCANIPALS benchmarks NIPALS method with various matrix sizes
func BenchmarkPCANIPALS(b *testing.B) {
	sizes := []struct {
		rows int
		cols int
	}{
		{100, 10},
		{1000, 10},
		{1000, 100},
		{5000, 100},
	}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("%dx%d", size.rows, size.cols), func(b *testing.B) {
			data := generateRandomMatrix(size.rows, size.cols)
			config := types.PCAConfig{
				Components:    min(10, size.cols),
				MeanCenter:    true,
				StandardScale: false,
				Method:        "nipals",
			}

			engine := NewPCAEngine()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_, err := engine.Fit(data, config)
				if err != nil {
					b.Fatalf("PCA failed: %v", err)
				}
			}
		})
	}
}

// BenchmarkPCAPreprocessing benchmarks different preprocessing options
func BenchmarkPCAPreprocessing(b *testing.B) {
	data := generateRandomMatrix(1000, 50)

	preprocessingOptions := []struct {
		name          string
		meanCenter    bool
		standardScale bool
	}{
		{"NoPreprocessing", false, false},
		{"MeanCenter", true, false},
		{"Standardize", true, true},
	}

	for _, opt := range preprocessingOptions {
		b.Run(opt.name, func(b *testing.B) {
			config := types.PCAConfig{
				Components:    10,
				MeanCenter:    opt.meanCenter,
				StandardScale: opt.standardScale,
				Method:        "svd",
			}

			engine := NewPCAEngine()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_, err := engine.Fit(data, config)
				if err != nil {
					b.Fatalf("PCA failed: %v", err)
				}
			}
		})
	}
}

// TestPCAMemoryUsage tests memory usage for large datasets
func TestPCAMemoryUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory usage test in short mode")
	}

	testCases := []struct {
		name string
		rows int
		cols int
	}{
		{"Small", 100, 10},
		{"Medium", 1000, 100},
		{"Large", 5000, 100},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Get initial memory stats
			var m1 runtime.MemStats
			runtime.GC()
			runtime.ReadMemStats(&m1)

			// Generate data and run PCA
			data := generateRandomMatrix(tc.rows, tc.cols)
			config := types.PCAConfig{
				Components:    min(10, tc.cols),
				MeanCenter:    true,
				StandardScale: false,
				Method:        "svd",
			}

			engine := NewPCAEngine()
			result, err := engine.Fit(data, config)
			if err != nil {
				t.Fatalf("PCA failed: %v", err)
			}

			// Get final memory stats
			var m2 runtime.MemStats
			runtime.ReadMemStats(&m2)

			// Calculate memory usage
			memUsed := m2.Alloc - m1.Alloc
			memUsedMB := float64(memUsed) / 1024 / 1024

			// Estimate expected memory usage (rough estimate)
			// Data matrix + scores + loadings + working memory
			dataSize := tc.rows * tc.cols * 8               // 8 bytes per float64
			expectedMB := float64(dataSize*4) / 1024 / 1024 // 4x for all matrices

			t.Logf("Matrix %dx%d: Used %.2f MB (expected ~%.2f MB)",
				tc.rows, tc.cols, memUsedMB, expectedMB)

			// Memory should scale roughly linearly
			if memUsedMB > expectedMB*10 {
				t.Errorf("Excessive memory usage: %.2f MB (expected < %.2f MB)",
					memUsedMB, expectedMB*10)
			}

			// Cleanup
			_ = result
			runtime.GC()
		})
	}
}

// TestPCAPerformanceScaling tests that performance scales appropriately
func TestPCAPerformanceScaling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance scaling test in short mode")
	}

	// Test that doubling the size roughly doubles the time (or less with optimizations)
	sizes := []struct {
		rows int
		cols int
	}{
		{500, 50},
		{1000, 50},
		{2000, 50},
	}

	times := make([]time.Duration, len(sizes))

	for i, size := range sizes {
		data := generateRandomMatrix(size.rows, size.cols)
		config := types.PCAConfig{
			Components:    10,
			MeanCenter:    true,
			StandardScale: false,
			Method:        "svd",
		}

		engine := NewPCAEngine()

		start := time.Now()
		_, err := engine.Fit(data, config)
		if err != nil {
			t.Fatalf("PCA failed for size %dx%d: %v", size.rows, size.cols, err)
		}
		times[i] = time.Since(start)

		t.Logf("Size %dx%d took %v", size.rows, size.cols, times[i])
	}

	// Check that time scaling is reasonable (not exponential)
	// Time should increase less than quadratically with size
	for i := 1; i < len(times); i++ {
		ratio := float64(times[i]) / float64(times[i-1])
		sizeRatio := float64(sizes[i].rows) / float64(sizes[i-1].rows)

		// Allow up to cubic scaling (generous for safety)
		maxRatio := sizeRatio * sizeRatio * sizeRatio
		if ratio > maxRatio {
			t.Errorf("Performance scaling too poor: time increased by %.2fx for %.2fx size increase",
				ratio, sizeRatio)
		}
	}
}

// BenchmarkPCAWithWideData benchmarks PCA with wide matrices (more columns than rows)
func BenchmarkPCAWithWideData(b *testing.B) {
	sizes := []struct {
		rows int
		cols int
	}{
		{10, 100},
		{50, 500},
		{100, 1000},
	}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("%dx%d", size.rows, size.cols), func(b *testing.B) {
			data := generateRandomMatrix(size.rows, size.cols)
			config := types.PCAConfig{
				Components:    min(size.rows-1, 10),
				MeanCenter:    true,
				StandardScale: false,
				Method:        "svd",
			}

			engine := NewPCAEngine()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_, err := engine.Fit(data, config)
				if err != nil {
					b.Fatalf("PCA failed: %v", err)
				}
			}
		})
	}
}

// TestLargeScaleStress performs stress testing with large matrices
func TestLargeScaleStress(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large scale stress test in short mode")
	}

	testCases := []struct {
		name   string
		rows   int
		cols   int
		method string
	}{
		{"10000x100_SVD", 10000, 100, "svd"},
		{"1000x1000_SVD", 1000, 1000, "svd"},
		{"100x10000_SVD", 100, 10000, "svd"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set timeout for this test
			timeout := time.After(60 * time.Second)
			done := make(chan bool)

			go func() {
				data := generateRandomMatrix(tc.rows, tc.cols)
				config := types.PCAConfig{
					Components:    min(10, min(tc.rows-1, tc.cols)),
					MeanCenter:    true,
					StandardScale: false,
					Method:        tc.method,
				}

				engine := NewPCAEngine()
				start := time.Now()

				result, err := engine.Fit(data, config)
				elapsed := time.Since(start)

				if err != nil {
					t.Errorf("PCA failed for %s: %v", tc.name, err)
				} else {
					t.Logf("%s completed in %v", tc.name, elapsed)
					// Basic sanity check
					if result.ComponentsComputed == 0 {
						t.Errorf("No components computed for %s", tc.name)
					}
				}

				done <- true
			}()

			select {
			case <-done:
				// Test completed successfully
			case <-timeout:
				t.Errorf("%s timed out after 60 seconds", tc.name)
			}
		})
	}
}
