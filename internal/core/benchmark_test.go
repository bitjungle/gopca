// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

package core

import (
	"fmt"
	"math"
	"math/rand"
	"testing"

	"github.com/bitjungle/gopca/internal/utils"
	"github.com/bitjungle/gopca/pkg/profiling"
	"github.com/bitjungle/gopca/pkg/types"
	"gonum.org/v1/gonum/mat"
)

// generateBenchmarkData creates a synthetic dataset for benchmarking
func generateBenchmarkData(rows, cols int) *mat.Dense {
	data := make([]float64, rows*cols)
	src := rand.NewSource(42) // Fixed seed for reproducibility
	rng := rand.New(src)

	for i := range data {
		data[i] = rng.NormFloat64()
	}

	return mat.NewDense(rows, cols, data)
}

// Benchmark SVD PCA with various dataset sizes
func BenchmarkSVD_PCA(b *testing.B) {
	sizes := []struct {
		rows, cols int
	}{
		{100, 50},
		{500, 100},
		{1000, 200},
		{5000, 500},
	}

	for _, size := range sizes {
		name := fmt.Sprintf("%dx%d", size.rows, size.cols)
		b.Run(name, func(b *testing.B) {
			data := generateBenchmarkData(size.rows, size.cols)
			engine := NewPCAEngine()

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				matrix := utils.DenseToMatrix(data)
				_, err := engine.FitTransform(matrix, types.PCAConfig{
					Components: min(10, size.cols/2),
					Method:     "svd",
				})
				if err != nil {
					b.Fatal(err)
				}
			}

			// Report memory usage
			b.ReportMetric(float64(size.rows*size.cols*8)/1024/1024, "MB/dataset")
		})
	}
}

// Benchmark NIPALS PCA
func BenchmarkNIPALS_PCA(b *testing.B) {
	sizes := []struct {
		rows, cols int
	}{
		{100, 50},
		{500, 100},
		{1000, 200},
	}

	for _, size := range sizes {
		name := fmt.Sprintf("%dx%d", size.rows, size.cols)
		b.Run(name, func(b *testing.B) {
			data := generateBenchmarkData(size.rows, size.cols)
			engine := NewPCAEngine()

			// Add some missing values for NIPALS testing
			dataCopy := mat.DenseCopyOf(data)
			for i := 0; i < size.rows*size.cols/20; i++ { // 5% missing
				row := rand.Intn(size.rows)
				col := rand.Intn(size.cols)
				dataCopy.Set(row, col, math.NaN())
			}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				matrix := utils.DenseToMatrix(dataCopy)
				_, err := engine.FitTransform(matrix, types.PCAConfig{
					Components: min(5, size.cols/2),
					Method:     "nipals",
				})
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// Benchmark Kernel PCA
func BenchmarkKernel_PCA(b *testing.B) {
	sizes := []struct {
		rows, cols int
	}{
		{100, 50},
		{300, 100},
		{500, 200},
	}

	kernelTypes := []string{"rbf", "linear", "polynomial"}

	for _, size := range sizes {
		for _, kernel := range kernelTypes {
			name := fmt.Sprintf("%dx%d_%s", size.rows, size.cols, kernel)
			b.Run(name, func(b *testing.B) {
				data := generateBenchmarkData(size.rows, size.cols)
				engine := NewPCAEngineForMethod("kernel")

				b.ResetTimer()
				b.ReportAllocs()

				for i := 0; i < b.N; i++ {
					matrix := utils.DenseToMatrix(data)
					_, err := engine.FitTransform(matrix, types.PCAConfig{
						Components:  min(5, size.cols/2),
						Method:      "kernel",
						KernelType:  kernel,
						KernelGamma: 0.01,
					})
					if err != nil {
						b.Fatal(err)
					}
				}
			})
		}
	}
}

// Benchmark preprocessing operations
func BenchmarkPreprocessing(b *testing.B) {
	sizes := []struct {
		rows, cols int
	}{
		{1000, 100},
		{5000, 500},
	}

	preprocessingTypes := []struct {
		name   string
		config types.PCAConfig
	}{
		{"MeanCenter", types.PCAConfig{MeanCenter: true}},
		{"StandardScale", types.PCAConfig{StandardScale: true}},
		{"RobustScale", types.PCAConfig{RobustScale: true}},
		{"SNV", types.PCAConfig{SNV: true}},
		{"All", types.PCAConfig{MeanCenter: true, StandardScale: true, SNV: true}},
	}

	for _, size := range sizes {
		for _, prep := range preprocessingTypes {
			name := fmt.Sprintf("%dx%d_%s", size.rows, size.cols, prep.name)
			b.Run(name, func(b *testing.B) {
				data := generateBenchmarkData(size.rows, size.cols)

				b.ResetTimer()
				b.ReportAllocs()

				for i := 0; i < b.N; i++ {
					matrix := make(types.Matrix, size.rows)
					for row := 0; row < size.rows; row++ {
						matrix[row] = make([]float64, size.cols)
						for col := 0; col < size.cols; col++ {
							matrix[row][col] = data.At(row, col)
						}
					}
					preprocessor := NewPreprocessorWithScaleOnly(
						prep.config.MeanCenter,
						prep.config.StandardScale,
						prep.config.RobustScale,
						prep.config.ScaleOnly,
						prep.config.SNV,
						prep.config.VectorNorm,
					)
					_, _ = preprocessor.FitTransform(matrix)
				}
			})
		}
	}
}

// Benchmark matrix operations
func BenchmarkMatrixOperations(b *testing.B) {
	sizes := []int{100, 500, 1000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("Transpose_%d", size), func(b *testing.B) {
			data := generateBenchmarkData(size, size)
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_ = data.T()
			}
		})

		b.Run(fmt.Sprintf("Multiply_%d", size), func(b *testing.B) {
			a := generateBenchmarkData(size, size)
			bt := generateBenchmarkData(size, size).T()
			c := mat.NewDense(size, size, nil)

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				c.Mul(a, bt)
			}
		})

		b.Run(fmt.Sprintf("Copy_%d", size), func(b *testing.B) {
			data := generateBenchmarkData(size, size)
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_ = mat.DenseCopyOf(data)
			}
		})
	}
}

// Benchmark with real-world dataset characteristics
func BenchmarkRealWorldScenario(b *testing.B) {
	// Simulate bronir2-like dataset (881x1008)
	b.Run("NIR_Dataset", func(b *testing.B) {
		data := generateBenchmarkData(881, 1008)
		engine := NewPCAEngine()

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			matrix := utils.DenseToMatrix(data)
			_, err := engine.FitTransform(matrix, types.PCAConfig{
				Components:    10,
				Method:        "svd",
				MeanCenter:    true,
				StandardScale: true,
				SNV:           true,
			})
			if err != nil {
				b.Fatal(err)
			}
		}

		// Report estimated memory usage
		memEstimate := profiling.EstimateMatrixMemory(881, 1008)
		b.ReportMetric(float64(memEstimate)/1024/1024, "MB/estimated")
	})

	// Simulate large dataset
	b.Run("Large_Dataset", func(b *testing.B) {
		// Skip if short testing
		if testing.Short() {
			b.Skip("Skipping large dataset benchmark in short mode")
		}

		data := generateBenchmarkData(5000, 2000)
		engine := NewPCAEngine()

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			matrix := utils.DenseToMatrix(data)
			_, err := engine.FitTransform(matrix, types.PCAConfig{
				Components:    20,
				Method:        "svd",
				MeanCenter:    true,
				StandardScale: true,
			})
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// min returns the smaller of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
