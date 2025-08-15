// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

package core

import (
	"testing"

	"github.com/bitjungle/gopca/pkg/profiling"
	"github.com/bitjungle/gopca/pkg/types"
	"gonum.org/v1/gonum/mat"
)

// TestMemoryLeaks checks for memory leaks in PCA operations
func TestMemoryLeaks(t *testing.T) {
	// Create test data
	data := mat.NewDense(100, 50, nil)
	for i := 0; i < 100; i++ {
		for j := 0; j < 50; j++ {
			data.Set(i, j, float64(i+j))
		}
	}

	t.Run("SVD_NoLeak", func(t *testing.T) {
		report := profiling.DetectLeaksInFunc("SVD_PCA", func() {
			engine := NewPCAEngine()
			matrix := make(types.Matrix, 100)
			for i := 0; i < 100; i++ {
				matrix[i] = make([]float64, 50)
				for j := 0; j < 50; j++ {
					matrix[i][j] = data.At(i, j)
				}
			}
			for i := 0; i < 10; i++ {
				_, _ = engine.FitTransform(matrix, types.PCAConfig{
					Components: 10,
					Method:     "svd",
				})
			}
		})

		if report.HasLeaks {
			t.Errorf("Memory leak detected in SVD PCA: %v", report.Leaks)
		}
	})

	t.Run("NIPALS_NoLeak", func(t *testing.T) {
		report := profiling.DetectLeaksInFunc("NIPALS_PCA", func() {
			engine := NewPCAEngine()
			matrix := make(types.Matrix, 100)
			for i := 0; i < 100; i++ {
				matrix[i] = make([]float64, 50)
				for j := 0; j < 50; j++ {
					matrix[i][j] = data.At(i, j)
				}
			}
			for i := 0; i < 5; i++ {
				_, _ = engine.FitTransform(matrix, types.PCAConfig{
					Components: 5,
					Method:     "nipals",
				})
			}
		})

		if report.HasLeaks {
			t.Errorf("Memory leak detected in NIPALS PCA: %v", report.Leaks)
		}
	})

	t.Run("Kernel_NoLeak", func(t *testing.T) {
		// Use smaller data for kernel PCA (it's O(n²))
		smallData := mat.NewDense(50, 20, nil)
		for i := 0; i < 50; i++ {
			for j := 0; j < 20; j++ {
				smallData.Set(i, j, float64(i+j))
			}
		}

		report := profiling.DetectLeaksInFunc("Kernel_PCA", func() {
			engine := NewPCAEngineForMethod("kernel")
			matrix := make(types.Matrix, 50)
			for i := 0; i < 50; i++ {
				matrix[i] = make([]float64, 20)
				for j := 0; j < 20; j++ {
					matrix[i][j] = smallData.At(i, j)
				}
			}
			for i := 0; i < 3; i++ {
				_, _ = engine.FitTransform(matrix, types.PCAConfig{
					Components: 5,
					Method:     "kernel",
					KernelType: "rbf",
				})
			}
		})

		if report.HasLeaks {
			t.Errorf("Memory leak detected in Kernel PCA: %v", report.Leaks)
		}
	})
}

// TestMemoryUsageWithLargeDataset tests memory usage with bronir2-like dataset
func TestMemoryUsageWithLargeDataset(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large dataset test in short mode")
	}

	// Create bronir2-sized dataset (881x1008)
	rows, cols := 881, 1008
	data := mat.NewDense(rows, cols, nil)
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			data.Set(i, j, float64(i*cols+j))
		}
	}

	profiler := profiling.NewMemoryProfiler()
	profiler.Start("Large_Dataset_PCA")

	engine := NewPCAEngine()
	matrix := make(types.Matrix, rows)
	for i := 0; i < rows; i++ {
		matrix[i] = make([]float64, cols)
		for j := 0; j < cols; j++ {
			matrix[i][j] = data.At(i, j)
		}
	}

	profiler.Checkpoint("Data_Created")

	result, err := engine.FitTransform(matrix, types.PCAConfig{
		Components:    20,
		Method:        "svd",
		MeanCenter:    true,
		StandardScale: true,
	})

	profiler.Checkpoint("PCA_Complete")

	if err != nil {
		t.Fatalf("PCA failed: %v", err)
	}

	if result == nil {
		t.Fatal("PCA returned nil result")
	}

	summary := profiler.Stop()

	// Check memory usage is reasonable
	expectedSize := profiling.EstimateMatrixMemory(rows, cols)
	// Allow up to 5x the data size for intermediate computations
	if summary.PeakAlloc > expectedSize*5 {
		t.Errorf("Excessive memory usage: peak %.2f MB for %.2f MB dataset",
			float64(summary.PeakAlloc)/1024/1024,
			float64(expectedSize)/1024/1024)
	}

	t.Logf("Memory usage for %dx%d dataset:", rows, cols)
	t.Logf("  Initial: %s", profiling.FormatBytes(summary.InitialAlloc))
	t.Logf("  Peak: %s", profiling.FormatBytes(summary.PeakAlloc))
	t.Logf("  Total allocated: %s", profiling.FormatBytes(summary.TotalAllocated))
	t.Logf("  GC runs: %d", summary.NumGCs)
}
